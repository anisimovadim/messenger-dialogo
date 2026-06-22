import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite' // <-- Импортируем новый плагин
import path from 'path'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    tailwindcss(), // <-- Добавляем его в список плагинов
  ],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'), // Вот эта строка связывает @ с папкой src
    },
  },
})