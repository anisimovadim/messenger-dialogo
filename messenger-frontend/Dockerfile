# Этап 1: Сборка фронтенда
FROM node:18-alpine AS build
WORKDIR /app
COPY package*.json ./
RUN npm install
COPY . .
# Важно: пробрасываем переменную окружения при сборке
ARG VITE_API_BASE_URL
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL
RUN npm run build

# Этап 2: Раздача через Nginx
FROM nginx:alpine
COPY --from=build /app/dist /usr/share/nginx/html
# Копируем конфиг nginx, если нужно (обычно достаточно стандартного)
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]