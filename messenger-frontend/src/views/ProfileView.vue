<script setup>
import { ref, onMounted } from 'vue';

// Добавляем defineEmits, чтобы сообщать родителю о желании вернуться
const emit = defineEmits(['show-chat']);

const username = ref('');
const userId = ref('');

onMounted(() => {
  username.value = localStorage.getItem('username') || 'Гость';
  userId.value = localStorage.getItem('user_id') || 'Неизвестен';
});

const logout = () => {
  localStorage.clear();
  window.location.href = '/login';
};
</script>

<template>
  <div class="flex-1 h-screen bg-[#F8FAFC] p-6 md:p-8">
    <button 
      @click="$emit('show-chat')" 
      class="flex items-center gap-2 text-slate-500 hover:text-[#6344F2] mb-6 transition-colors"
    >
      <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
      </svg>
      <span class="font-bold">К чатам</span>
    </button>

    <h2 class="text-2xl font-bold text-slate-800 mb-6">Профиль пользователя</h2>
    
    <div class="bg-white p-6 rounded-2xl shadow-sm border border-slate-200 max-w-md">
      <div class="flex items-center gap-4 mb-8">
        <div class="w-16 h-16 bg-indigo-100 text-indigo-600 rounded-full flex items-center justify-center text-xl font-bold">
          {{ username ? username.substring(0, 2).toUpperCase() : '??' }}
        </div>
        <div>
          <h3 class="font-bold text-lg">{{ username }}</h3>
          <p class="text-slate-400 text-sm">ID: {{ userId }}</p>
        </div>
      </div>

      <button 
        @click="logout" 
        class="w-full py-3 px-4 bg-red-50 hover:bg-red-100 text-red-600 font-bold rounded-xl transition-colors flex items-center justify-center gap-2"
      >
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"/></svg>
        Выйти из аккаунта
      </button>
    </div>
  </div>
</template>