# Используем один большой образ для всего
FROM golang:1.24

WORKDIR /app

RUN mkdir -p /app/uploads && chmod 777 /app/uploads

# Копируем всё содержимое папки messenger-backend в контейнер
COPY messenger-backend/ .

# Устанавливаем зависимости
RUN GOTOOLCHAIN=auto go mod download

# Собираем
RUN GOTOOLCHAIN=auto go build -o main .

# Запускаем
CMD ["./main"]