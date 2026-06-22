<script setup>
import { ref, onMounted } from 'vue';
import api from '@/services/api'; // Импортируем ваш Axios-сервис

const emit = defineEmits(['close', 'chat-created']);
const users = ref([]);
const myId = localStorage.getItem('user_id');

const fetchUsers = async () => {
  try {
    // Axios возвращает данные напрямую, не нужно делать await response.json()
    const data = await api.get(`/api/users?exclude_id=${myId}`);
    users.value = data;
  } catch (err) {
    console.error("Ошибка при загрузке пользователей:", err);
  }
};

const createChat = async (partnerId) => {
  try {
    // Axios автоматически добавит Content-Type: application/json
    // И сам выполнит JSON.stringify для объекта в body
    const data = await api.post('/api/chats/create', { 
      creator_id: parseInt(myId), 
      partner_id: partnerId 
    });
    
    emit('chat-created', data.chat_id);
  } catch (err) {
    console.error("Ошибка при создании чата:", err);
  }
};

onMounted(fetchUsers);
</script>

<template>
  <div class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center p-4">
    <div class="bg-white rounded-2xl w-full max-w-sm p-6">
      <h3 class="font-bold text-lg mb-4">Выберите собеседника</h3>
      <div class="space-y-2">
        <button v-for="user in users" :key="user.id" @click="createChat(user.id)"
          class="w-full text-left p-3 hover:bg-slate-100 rounded-lg transition">
          {{ user.username }}
        </button>
      </div>
      <button @click="$emit('close')" class="mt-4 w-full text-slate-500">Отмена</button>
    </div>
  </div>
</template>