<script setup>
import { ref, onMounted } from "vue";

defineProps({
  activeTab: {
    type: String,
    default: 'chat' // 'chat', 'friends', 'notifications', 'profile'
  }
});

const currentUserName = ref("");
const emit = defineEmits(["show-profile", "show-chat", "click-friends", "toggle-notifications"]);

onMounted(() => {
  currentUserName.value = localStorage.getItem("username") || "Пользователь";
});

// const toggleUser = () => {
//   const currentId = localStorage.getItem('user_id')

//   if (currentId === '1') {
//     localStorage.setItem('user_id', '2')
//     localStorage.setItem('username', 'Олег Олегов')
//   } else {
//     localStorage.setItem('user_id', '1')
//     localStorage.setItem('username', 'Маша Машева')
//   }

//   window.location.reload()
// }
</script>

<template>
  <div
    class="fixed bottom-0 left-0 w-full h-16 bg-[#F1F3F7] border-t border-slate-200 flex flex-row items-center justify-around px-2 z-50 md:static md:w-[72px] md:h-screen md:border-r md:border-slate-200 md:flex-col md:justify-between md:py-6 md:px-0">

    <div class="flex flex-row gap-2 md:flex-col md:gap-4 md:w-full md:items-center">

      <button
        class="w-12 h-12 md:w-11 md:h-11 bg-[#1E1B4B] text-white rounded-xl flex items-center justify-center shadow-md transition-colors cursor-pointer">
        <svg xmlns="http://www.w3.org/2000/svg" fill="currentColor" viewBox="0 0 24 24"
          class="w-6 h-6 md:w-5 md:h-5 text-indigo-400">
          <path
            d="M21.752 15.002A9.72 9.72 0 0 1 18 15.75c-5.385 0-9.75-4.365-9.75-9.75 0-1.33.266-2.597.748-3.752A9.753 9.753 0 0 0 3 11.25C3 16.635 7.365 21 12.75 21a9.753 9.753 0 0 0 9.002-5.998Z" />
        </svg>
      </button>

      <button title="Друзья" @click="$emit('click-friends')"
        class="w-12 h-12 md:w-11 md:h-11 rounded-xl flex items-center justify-center shadow-sm border transition-all cursor-pointer"
        :class="activeTab === 'friends' ? 'bg-indigo-50 border-indigo-600 text-indigo-600' : 'bg-white border-slate-100 text-slate-600 hover:bg-slate-50'">
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"
          class="w-6 h-6 md:w-5 md:h-5">
          <path stroke-linecap="round" stroke-linejoin="round"
            d="M18 18.72a9.094 9.094 0 0 0 3.741-.479 3 3 0 0 0-4.682-2.72m.94 3.198.001.031c0 .225-.012.447-.037.666A11.944 11.944 0 0 1 12 21c-2.17 0-4.207-.576-5.963-1.584A6.062 6.062 0 0 1 6 18.719m12 0a5.971 5.971 0 0 0-.941-3.197m0 0A5.995 5.995 0 0 0 12 12.75a5.995 5.995 0 0 0-5.058 2.772m0 0a3 3 0 0 0-4.681 2.72 8.986 8.986 0 0 0 3.74.477m.94-3.197a5.971 5.971 0 0 0-.94 3.197M15 6.75a3 3 0 1 1-6 0 3 3 0 0 1 6 0Zm6 3a2.25 2.25 0 1 1-4.5 0 2.25 2.25 0 0 1 4.5 0Zm-13.5 0a2.25 2.25 0 1 1-4.5 0 2.25 2.25 0 0 1 4.5 0Z" />
        </svg>
      </button>

      <button @click="$emit('show-chat')" title="Сообщения"
        class="w-12 h-12 md:w-11 md:h-11 rounded-xl flex items-center justify-center shadow-sm border transition-all cursor-pointer"
        :class="activeTab === 'chat' ? 'bg-indigo-50 border-indigo-600 text-indigo-600' : 'bg-white border-slate-100 text-slate-600 hover:bg-indigo-50 hover:text-indigo-600'">
        <svg width="22" height="22" viewBox="0 0 24 24" fill="currentColor" xmlns="http://www.w3.org/2000/svg"
          class="opacity-80">
          <path
            d="M21.1599 0.28816L7.15986 4.81656L2.15986 6.32597C-1.83992 7.83543 0.660011 9.3449 2.15986 9.84805L9.15986 12.3638L17.1599 6.82912L11.6599 14.8796L14.1599 21.9238C15.6602 25.949 17.1144 23.1893 17.6599 21.4206C19.1599 16.5568 22.4599 6.02413 23.6599 2.80394C24.8599 -0.416257 22.6602 -0.215091 21.1599 0.28816Z" />
        </svg>
      </button>

      <button @click="$emit('toggle-notifications')" title="Уведомления"
        class="w-12 h-12 md:w-11 md:h-11 rounded-xl flex items-center justify-center shadow-sm border transition-all cursor-pointer"
        :class="activeTab === 'notifications' ? 'bg-indigo-50 border-indigo-600 text-indigo-600' : 'bg-slate-200 border-slate-200 text-slate-600 hover:bg-slate-300'">
        <svg class="w-6 h-6 md:w-5 md:h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
            d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
        </svg>
      </button>
    </div>

    <button @click="$emit('show-profile')" class="flex items-center cursor-pointer">
      <div title="Текущий пользователь"
        class="w-12 h-12 md:w-10 md:h-10 rounded-xl text-sm md:text-xs flex items-center justify-center shadow-sm border-2 transition-all"
        :class="activeTab === 'profile' ? 'bg-indigo-600 border-white text-white' : 'bg-indigo-100 border-indigo-200 text-indigo-700 font-black'">
        {{ currentUserName.substring(0, 2).toUpperCase() }}
      </div>
    </button>
  </div>
</template>