package main

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/gomail.v2"
	"io" // <-- ДОБАВИЛИ ДЛЯ КОПИРОВАНИЯ ФАЙЛОВ
	"log"
	"math/rand"
	"net/http"
	"os" // <-- ДОБАВИЛИ ДЛЯ РАБОТЫ С ПАПКАМИ
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/jackc/pgx/v5/pgxpool"
)

// --- СТРУКТУРЫ ДАННЫХ ---

type User struct {
	ID          int    `json:"id"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"username"`
}

type Chat struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	LastMessage string `json:"last_message"`
	Time        string `json:"time"`
	UnreadCount int    `json:"unread_count"`
}

type Message struct {
	ID           int       `json:"id"`
	ChatID       int       `json:"chat_id"`
	UserID       int       `json:"user_id"`
	Username     string    `json:"username"`
	Text         string    `json:"text"`
	Type         string    `json:"type"` // "text" или "read_receipt"
	IsRead       bool      `json:"is_read"`
	IsTyping     bool      `json:"is_typing"`
	Filename     string    `json:"filename"` // Было sql.NullString
	OriginalName string    `json:"original_name"`
	CreatedAt    time.Time `json:"created_at"`
}

// --- УПРАВЛЕНИЕ ВЕБ-СОКЕТАМИ (HUB) ---

type Client struct {
	UserID int
	Conn   *websocket.Conn
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
	mu         sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run(db *pgxpool.Pool) {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("Пользователь %d подключил сокет", client.UserID)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Conn.Close(websocket.StatusNormalClosure, "соединение закрыто")
				log.Printf("Пользователь %d отключил сокет", client.UserID)
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			// Сериализуем сообщение в JSON один раз для всех
			payload, err := json.Marshal(msg)
			if err != nil {
				log.Printf("Ошибка маршалинга: %v", err)
				continue
			}

			h.mu.Lock()
			// Ищем по пулу, кому принадлежит этот чат (для приватных чатов)
			// Достаем участников чата напрямую из базы данных, чтобы знать, кому слать
			rows, err := db.Query(context.Background(), "SELECT user_id FROM chat_members WHERE chat_id = $1", msg.ChatID)
			if err == nil {
				members := make(map[int]bool)
				for rows.Next() {
					var uid int
					rows.Scan(&uid)
					members[uid] = true
				}
				rows.Close()

				// Рассылаем только тем, кто состоит в этой комнате и сейчас онлайн
				for client := range h.clients {
					if members[client.UserID] {
						err := client.Conn.Write(context.Background(), websocket.MessageText, payload)
						if err != nil {
							log.Printf("Ошибка отправки клиенту %d: %v", client.UserID, err)
						}
					}
				}
			}
			h.mu.Unlock()
		}
	}
}

// --- НАСТРОЙКА И ИНИЦИАЛИЗАЦИЯ ИСПРАВЛЕННОЙ БД ---

func initDB() *pgxpool.Pool {
	connStr := "postgres://chat_user:chat_password@postgres:5432/chat_database"

	// Используем пул подключений вместо одиночного коннекта
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatalf("Не удалось создать пул подключений к БД: %v", err)
	}

	queries := []string{
		`DROP TABLE IF EXISTS messages CASCADE;`,
		`DROP TABLE IF EXISTS chat_members CASCADE;`,
		`DROP TABLE IF EXISTS chats CASCADE;`,
		`DROP TABLE IF EXISTS users CASCADE;`,

		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			email VARCHAR(100) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			display_name VARCHAR(50) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS chats (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100),
			type VARCHAR(20) NOT NULL DEFAULT 'private',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS chat_members (
			chat_id INT REFERENCES chats(id) ON DELETE CASCADE,
			user_id INT REFERENCES users(id) ON DELETE CASCADE,
			PRIMARY KEY (chat_id, user_id)
		);`,
		`CREATE TABLE IF NOT EXISTS messages (
			id SERIAL PRIMARY KEY,
			chat_id INT REFERENCES chats(id) ON DELETE CASCADE,
			user_id INT REFERENCES users(id) ON DELETE CASCADE,
			text TEXT NOT NULL,
			type VARCHAR(20) NOT NULL DEFAULT 'text',
			filename TEXT,
			original_name TEXT,
			is_read BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS pending_users (
			email VARCHAR(100) PRIMARY KEY,
			password_hash VARCHAR(255) NOT NULL,
			display_name VARCHAR(50) NOT NULL,
			verification_code VARCHAR(6) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,
	}

	for _, q := range queries {
		if _, err := pool.Exec(context.Background(), q); err != nil {
			log.Fatalf("Ошибка накатывания таблиц: %v", err)
		}
	}

	// seedDemoData(pool)
	log.Println("Успешная ПЕРЕСБОРКА PostgreSQL базы данных с ПУЛОМ подключений!")
	return pool
}

func seedDemoData(pool *pgxpool.Pool) {
	// 1. Создаем Машу и Олега (пароль у обоих: password123)
	// Для тестов зашиты хэши
	passHash := "$2a$10$8Kz6U1S9zM.PZ8WjWjEKe.L6bXwY2B8xLhL2H5vW7V1v1v1v1v1v1"

	pool.Exec(context.Background(),
		"INSERT INTO users (id, email, password_hash, display_name) VALUES (1, 'test@gmail.com', $1, 'Маша Машева') ON CONFLICT DO NOTHING", passHash)
	pool.Exec(context.Background(),
		"INSERT INTO users (id, email, password_hash, display_name) VALUES (2, 'oleg@gmail.com', $1, 'Олег Олегов') ON CONFLICT DO NOTHING", passHash)

	// 2. Создаем их общий чат
	var chatID int
	err := pool.QueryRow(context.Background(),
		"INSERT INTO chats (id, name, type) VALUES (1, 'Маша и Олег', 'private') ON CONFLICT (id) DO UPDATE SET name=EXCLUDED.name RETURNING id").Scan(&chatID)

	if err == nil {
		pool.Exec(context.Background(), "INSERT INTO chat_members (chat_id, user_id) VALUES ($1, 1) ON CONFLICT DO NOTHING", chatID)
		pool.Exec(context.Background(), "INSERT INTO chat_members (chat_id, user_id) VALUES ($1, 2) ON CONFLICT DO NOTHING", chatID)

		// 3. Закидываем стартовые сообщения
		pool.Exec(context.Background(), "INSERT INTO messages (chat_id, user_id, text, is_read) VALUES ($1, 2, 'Привет! Проект Dialogo готов к тестированию сокетов?', true)", chatID)
		pool.Exec(context.Background(), "INSERT INTO messages (chat_id, user_id, text, is_read) VALUES ($1, 1, 'Привет! Да, запускай бэкенд на Go, проверим статусы.', true)", chatID)
	}
}

// Вспомогательная функция для CORS-политики
func enableCORS(w http.ResponseWriter, r *http.Request) bool { // Добавили bool
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return true // Возвращаем true, если это OPTIONS
	}
	return false // Возвращаем false, если нет
}

// Функция для генерации случайного имени
func generateRandomName(original string) string {
	ext := filepath.Ext(original)
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("%x%s", b, ext)
}

// --- ОСНОВНАЯ ФУНКЦИЯ MAIN ---

func main() {
	dbConn := initDB()
	defer dbConn.Close()

	hub := NewHub()
	go hub.Run(dbConn)

	// 1. Эндпоинт авторизации (Login)
	http.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		if enableCORS(w, r) {
			return
		}
		var creds struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		json.NewDecoder(r.Body).Decode(&creds)

		var u User
		var passHash string
		err := dbConn.QueryRow(context.Background(),
			"SELECT id, display_name, password_hash FROM users WHERE email = $1",
			creds.Email).Scan(&u.ID, &u.DisplayName, &passHash)

		if err != nil {
			http.Error(w, "Пользователь не найден", http.StatusUnauthorized)
			return
		}

		// Сверяем пароль
		if err := bcrypt.CompareHashAndPassword([]byte(passHash), []byte(creds.Password)); err != nil {
			http.Error(w, "Неверный пароль", http.StatusUnauthorized)
			return
		}

		// Успех
		response := map[string]interface{}{
			"token":    "mock-jwt-token",
			"username": u.DisplayName,
			"user_id":  u.ID,
		}
		json.NewEncoder(w).Encode(response)
	})

	http.HandleFunc("/api/send-code", func(w http.ResponseWriter, r *http.Request) {
		if enableCORS(w, r) {
			return
		}

		var u struct {
			Email, Password, Username string
		}
		json.NewDecoder(r.Body).Decode(&u)

		// 1. Генерируем 6-значный код
		code := fmt.Sprintf("%06d", rand.Intn(1000000))

		// 2. Хэшируем пароль
		hash, _ := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)

		// 3. Сохраняем в pending_users (upsert на случай повторной попытки)
		_, err := dbConn.Exec(context.Background(),
			"INSERT INTO pending_users (email, password_hash, display_name, verification_code) VALUES ($1, $2, $3, $4) ON CONFLICT (email) DO UPDATE SET verification_code = $4",
			u.Email, string(hash), u.Username, code)

		if err != nil {
			http.Error(w, "Ошибка сохранения", http.StatusInternalServerError)
			return
		}

		// 4. ТУТ ДОЛЖНА БЫТЬ ОТПРАВКА EMAIL (через gomail или smtp)
		// log.Printf("КОД ДЛЯ %s: %s", u.Email, code) тест

		// Добавьте в импорт: "gopkg.in/gomail.v2"

		m := gomail.NewMessage()
		m.SetHeader("From", "anisimov.vadim.alexeevich@gmail.com")
		m.SetHeader("To", u.Email)
		m.SetHeader("Subject", "Код подтверждения Dialogo.")
		m.SetBody("text/html", fmt.Sprintf("Ваш код подтверждения: <b>%s</b>", code))

		d := gomail.NewDialer("smtp.gmail.com", 587, "anisimov.vadim.alexeevich@gmail.com", "tnsz kmvz foen kdqg")

		if err := d.DialAndSend(m); err != nil {
			log.Printf("Ошибка отправки email: %v", err)
			http.Error(w, "Ошибка отправки письма", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/api/verify-and-register", func(w http.ResponseWriter, r *http.Request) {
		if enableCORS(w, r) {
			return
		}

		var req struct {
			Email string `json:"email"`
			Code  string `json:"code"`
		}
		json.NewDecoder(r.Body).Decode(&req)

		// 1. Ищем временную запись
		var u struct {
			PassHash, Name string
		}
		var code string
		err := dbConn.QueryRow(context.Background(),
			"SELECT password_hash, display_name, verification_code FROM pending_users WHERE email = $1",
			req.Email).Scan(&u.PassHash, &u.Name, &code)

		if err != nil || code != req.Code {
			http.Error(w, "Неверный код или почта", http.StatusBadRequest)
			return
		}

		// 2. Переносим в основную таблицу
		_, err = dbConn.Exec(context.Background(),
			"INSERT INTO users (email, password_hash, display_name) VALUES ($1, $2, $3)",
			req.Email, u.PassHash, u.Name)

		// 3. Удаляем из временной
		dbConn.Exec(context.Background(), "DELETE FROM pending_users WHERE email = $1", req.Email)

		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// 1.1 Эндпоинт регистрации
	http.HandleFunc("/api/register", func(w http.ResponseWriter, r *http.Request) {
		if enableCORS(w, r) {
			return
		}
		if r.Method != http.MethodPost {
			return
		}

		var u User

		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			http.Error(w, "Неверный формат данных", http.StatusBadRequest)
			return
		}

		if u.Email == "" || u.Password == "" || u.DisplayName == "" {
			http.Error(w, "Все поля должны быть заполнены", http.StatusBadRequest)
			return
		}

		if len(u.Password) < 6 {
			http.Error(w, "Пароль слишком короткий", http.StatusBadRequest)
			return
		}

		// Хэшируем пароль
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Ошибка сервера", http.StatusInternalServerError)
			return
		}

		// Сохраняем в БД
		_, err = dbConn.Exec(context.Background(),
			"INSERT INTO users (email, password_hash, display_name) VALUES ($1, $2, $3)",
			u.Email, string(hashedPassword), u.DisplayName)

		if err != nil {
			http.Error(w, "Email уже занят или ошибка БД", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"message": "Регистрация успешна"})
	})

	// 2. Эндпоинт получения списка чатов пользователя
	http.HandleFunc("/api/chats", func(w http.ResponseWriter, r *http.Request) {
		if enableCORS(w, r) {
			return
		}
		userIDStr := r.URL.Query().Get("user_id")
		userID, _ := strconv.Atoi(userIDStr)

		query := `
			SELECT c.id, c.type, 
				(SELECT u.display_name 
					FROM users u 
					JOIN chat_members cm2 ON u.id = cm2.user_id 
					WHERE cm2.chat_id = c.id AND u.id != $1 LIMIT 1) as partner_name,
				COALESCE((SELECT text FROM messages WHERE chat_id = c.id ORDER BY created_at DESC LIMIT 1), 'Нет сообщений') as last_msg
			FROM chats c
			JOIN chat_members cm ON c.id = cm.chat_id
			WHERE cm.user_id = $1 
			ORDER BY c.id DESC`

		// Обновленный запрос с поддержкой картинок и голосовых
		rows, err := dbConn.Query(context.Background(), query, userID)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var chatList []Chat
		for rows.Next() {
			var c Chat
			rows.Scan(&c.ID, &c.Type, &c.Name, &c.LastMessage)
			c.Time = "Сейчас"
			chatList = append(chatList, c)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(chatList)
	})

	// 3. Эндпоинт истории сообщений конкретного чата
	http.HandleFunc("/api/messages", func(w http.ResponseWriter, r *http.Request) {
		if enableCORS(w, r) {
			return
		}
		chatID := r.URL.Query().Get("chat_id")

		rows, err := dbConn.Query(context.Background(), `
			SELECT m.id, m.chat_id, m.user_id, u.display_name, m.text, m.type, m.is_read, m.created_at, 
				COALESCE(m.filename, ''), 
				COALESCE(m.original_name, '')
			FROM messages m
			JOIN users u ON m.user_id = u.id
			WHERE m.chat_id = $1 ORDER BY m.created_at ASC`, chatID)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var messages []Message = []Message{}
		for rows.Next() {
			var m Message
			// Теперь здесь обычные строки, которые успешно примут пустую строку от COALESCE
			err := rows.Scan(&m.ID, &m.ChatID, &m.UserID, &m.Username, &m.Text, &m.Type, &m.IsRead, &m.CreatedAt, &m.Filename, &m.OriginalName)
			if err != nil {
				log.Println("Ошибка сканирования:", err)
				continue
			}
			messages = append(messages, m)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(messages)
	})

	// 4. Получение списка всех пользователей для создания нового чата
	http.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		if enableCORS(w, r) {
			return
		}
		excludeID := r.URL.Query().Get("exclude_id")

		rows, err := dbConn.Query(context.Background(),
			"SELECT id, email, display_name FROM users WHERE id != $1", excludeID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var userList []User = []User{}
		for rows.Next() {
			var u User
			rows.Scan(&u.ID, &u.Email, &u.DisplayName)
			userList = append(userList, u)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(userList)
	})

	// 5. Создание нового приватного чата
	http.HandleFunc("/api/chats/create", func(w http.ResponseWriter, r *http.Request) {
		if enableCORS(w, r) {
			return
		}
		if r.Method != http.MethodPost {
			return
		}

		var req struct {
			CreatorID int `json:"creator_id"`
			PartnerID int `json:"partner_id"`
		}
		json.NewDecoder(r.Body).Decode(&req)

		var existingChatID int
		err := dbConn.QueryRow(context.Background(), `
			SELECT cm1.chat_id FROM chat_members cm1
			JOIN chat_members cm2 ON cm1.chat_id = cm2.chat_id
			JOIN chats c ON c.id = cm1.chat_id
			WHERE cm1.user_id = $1 AND cm2.user_id = $2 AND c.type = 'private'`,
			req.CreatorID, req.PartnerID).Scan(&existingChatID)

		if err == nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]int{"chat_id": existingChatID})
			return
		}

		var newChatID int
		tx, _ := dbConn.Begin(context.Background())
		defer tx.Rollback(context.Background())

		tx.QueryRow(context.Background(),
			"INSERT INTO chats (name, type) VALUES ($1, $2) RETURNING id",
			"Приватный чат", "private").Scan(&newChatID)

		tx.Exec(context.Background(),
			"INSERT INTO chat_members (chat_id, user_id) VALUES ($1, $2), ($1, $3)",
			newChatID, req.CreatorID, req.PartnerID)

		tx.Commit(context.Background())

		hub.broadcast <- Message{
			Type:   "new_chat",
			ChatID: newChatID,
			// Указываем ID участников, чтобы фильтровать на клиенте
			Text: fmt.Sprintf("%d,%d", req.CreatorID, req.PartnerID),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]int{"chat_id": newChatID})
	})

	// 6. Хэндлер WebSocket Соединений
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		userIDStr := r.URL.Query().Get("user_id")
		userID, _ := strconv.Atoi(userIDStr)

		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			OriginPatterns: []string{"*"},
		})
		if err != nil {
			log.Printf("Ошибка accept вебсокета: %v", err)
			return
		}

		client := &Client{UserID: userID, Conn: conn}
		hub.register <- client

		defer func() {
			hub.unregister <- client
		}()

		for {
			_, payload, err := conn.Read(context.Background())
			if err != nil {
				break
			}

			var msg Message
			if err := json.Unmarshal(payload, &msg); err != nil {
				continue
			}
			msg.CreatedAt = time.Now()

			// Если это отчет о прочтении, обновляем БД и пересылаем статус в хаб
			if msg.Type == "read_receipt" {
				// 1. Обновляем БД
				_, err = dbConn.Exec(context.Background(),
					"UPDATE messages SET is_read = TRUE WHERE chat_id = $1 AND user_id != $2",
					msg.ChatID, msg.UserID)

				// 2. Получаем актуальный список участников из базы
				rows, err := dbConn.Query(context.Background(), "SELECT user_id FROM chat_members WHERE chat_id = $1", msg.ChatID)
				if err == nil {
					members := make(map[int]bool)
					for rows.Next() {
						var uid int
						rows.Scan(&uid)
						members[uid] = true
					}
					rows.Close()

					// 3. Рассылаем уведомление только собеседнику (используем hub вместо undefined h)
					hub.mu.Lock()
					payload, _ := json.Marshal(msg)

					for client := range hub.clients {
						// Условие: если это участник чата И это не отправитель
						if members[client.UserID] && client.UserID != msg.UserID {
							client.Conn.Write(context.Background(), websocket.MessageText, payload)
						}
					}
					hub.mu.Unlock()
				}
				continue
			}

			if msg.Type == "typing" {
				hub.broadcast <- msg
				continue
			}

			if msg.Type == "edit_message" {
				// Обновляем текст в базе
				_, err = dbConn.Exec(context.Background(),
					"UPDATE messages SET text = $1 WHERE id = $2", msg.Text, msg.ID)
				if err == nil {
					hub.broadcast <- msg // Рассылаем всем обновленное сообщение
				}
				continue
			}

			if msg.Type == "delete_message" {
				log.Printf("Удаляем сообщение с ID: %d", msg.ID) // Добавь это
				res, err := dbConn.Exec(context.Background(), "DELETE FROM messages WHERE id = $1", msg.ID)
				if err != nil {
					log.Printf("Ошибка SQL удаления: %v", err)
				} else {
					rowsAffected := res.RowsAffected()
					log.Printf("Удалено строк: %d", rowsAffected)
				}
				hub.broadcast <- msg
				continue
			}

			// Стандартная запись сообщения (текст, файл или голос)
			dbConn.QueryRow(context.Background(),
				"INSERT INTO messages (chat_id, user_id, text, type, filename, original_name) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id",
				msg.ChatID, msg.UserID, msg.Text, msg.Type, msg.Filename, msg.OriginalName).Scan(&msg.ID)

			dbConn.QueryRow(context.Background(), "SELECT display_name FROM users WHERE id = $1", msg.UserID).Scan(&msg.Username)

			hub.broadcast <- msg
		}
	})

	// Создаем папку для загрузок, если её ещё нет
	_ = os.Mkdir("uploads", os.ModePerm)

	// Раздача файлов из папки uploads (чтобы картинки открывались по ссылкам)
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))

	// Эндпоинт загрузки файлов
	http.HandleFunc("/api/upload", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
		if r.Method == "OPTIONS" {
			return
		}
		if r.Method != http.MethodPost {
			return
		}

		r.ParseMultipartForm(20 << 20)
		file, handler, err := r.FormFile("file")
		if err != nil { /*...*/
		}
		defer file.Close()

		// Используем исходное имя файла
		randomName := generateRandomName(handler.Filename)
		filePath := "uploads/" + randomName

		dst, err := os.Create(filePath)

		if err != nil {
			http.Error(w, "Ошибка сохранения", http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		io.Copy(dst, file)

		// Возвращаем URL и ИМЯ файла
		// Возвращаем фронтенду и случайное имя, и оригинальное
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"url":           "/app/uploads/" + randomName,
			"filename":      randomName,
			"original_name": handler.Filename,
		})
	})

	log.Println("Go Комнатный-сервер запущен на :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
