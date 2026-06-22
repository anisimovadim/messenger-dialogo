<script setup>
import { ref, onMounted, onUnmounted } from 'vue'

const props = defineProps(['isOpen'])
const emit = defineEmits(['close'])

const notifications = ref([
  {
    id: 1,
    user: '@system',
    action: 'Добро пожаловать в систему Dialogo! Все каналы стабильны.',
    chatName: '',
    time: '12:00',
    avatarColor: 'bg-slate-200 text-slate-700',
    type: 'system'
  }
])

// Обработчик системного события нового уведомления
const handleNewNotification = (event) => {
  // Добавляем новое уведомление в начало списка
  notifications.value.unshift(event.detail)
  
  // Ограничим лог максимум 10 уведомлениями, чтобы не забивать память
  if (notifications.value.length > 10) {
    notifications.value.pop()
  }
}

onMounted(() => {
  window.addEventListener('new-notification', handleNewNotification)
})

onUnmounted(() => {
  window.removeEventListener('new-notification', handleNewNotification)
})
</script>

<template>
  <div :class="[
    props.isOpen ? 'fixed inset-0 z-50 bg-white md:static md:bg-transparent' : 'hidden lg:flex',
    'w-full md:w-[280px] h-full border-l border-slate-200 flex flex-col flex-shrink-0'
  ]">
    
    <div class="h-[76px] px-6 border-b border-slate-200 flex items-center justify-between bg-white">
      <h3 class="font-black text-slate-900 text-sm uppercase">Уведомления</h3>
      <button @click="$emit('close')" class="lg:hidden p-2 text-slate-500">✕ Закрыть</button>
    </div>

    <div class="flex-1 overflow-y-auto p-4 space-y-3">
      <div 
        v-for="notif in notifications" 
        :key="notif.id"
        class="bg-white p-4 rounded-2xl border border-slate-100 shadow-xs flex flex-col gap-2 transition-all hover:shadow-md"
      >
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-2">
            <div :class="notif.avatarColor" class="w-6 h-6 rounded-lg flex items-center justify-center text-[10px] font-bold">
              {{ notif.user[1]?.toUpperCase() || '?' }}
            </div>
            <span class="text-xs font-bold text-slate-800 truncate max-w-[120px]">{{ notif.user }}</span>
          </div>
          <span class="text-[10px] font-semibold text-slate-400">{{ notif.time }}</span>
        </div>
        <p class="text-xs text-slate-600 font-medium leading-normal">
          {{ notif.action }} 
          <span v-if="notif.chatName" class="font-bold text-[#6344F2]">«{{ notif.chatName }}»</span>
        </p>
      </div>
    </div>
  </div>
</template>