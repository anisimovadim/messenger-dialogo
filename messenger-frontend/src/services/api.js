import axios from 'axios';

// Используем BASE_URL, чтобы избежать ошибки ReferenceError
const BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';

// Превращаем http://... в ws://... для WebSocket
export const WS_URL = BASE_URL.replace(/^http/, 'ws');

const api = axios.create({
  baseURL: BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Добавляем перехватчик (interceptor) запросов
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Добавляем перехватчик ответов
api.interceptors.response.use(
  (response) => response.data,
  (error) => {
    // Проверяем, есть ли ответ от сервера
    if (error.response) {
      // Если сервер вернул текст, а не JSON
      if (typeof error.response.data === 'string') {
        return Promise.reject(new Error(error.response.data));
      }
      
      // Если сервер вернул JSON объект (стандартный случай)
      const message = error.response.data.error || error.response.data.message || 'Ошибка запроса';
      return Promise.reject(new Error(message));
    }
    
    // Если ошибки сети или нет ответа
    return Promise.reject(new Error('Ошибка сети или сервер недоступен'));
  }
);

export default api;