<script setup>
import { ref, onMounted } from 'vue'
import Sidebar from '../components/Sidebar.vue'
import ChatList from '../components/ChatList.vue'
import ChatWindow from '../components/ChatWindow.vue'
import ProfileView from './ProfileView.vue'
import NotificationPanel from '../components/NotificationPanel.vue'
import UserSelector from '../components/UserSelector.vue'

const isNotificationOpen = ref(false)
const activeChatId = ref(null)
const activeChatName = ref('')
const activeView = ref('chat') // 'chat' или 'profile'
const isUserSelectorOpen = ref(false);
const chatListRef = ref(null)

// Метод для выбора чата
const selectChat = (chat) => {
  activeChatId.value = chat.id
  activeChatName.value = chat.name
  activeView.value = 'chat' // При выборе чата переключаемся на вид чата
}

const handleNewChat = () => {
  isUserSelectorOpen.value=false;
  if (chatListRef.value) {
    chatListRef.value.fetchChats() 
  }
}
</script>

<template>
  <div class="flex w-full h-screen bg-slate-100 overflow-hidden font-sans antialiased">

    <Sidebar @toggle-notifications="isNotificationOpen = !isNotificationOpen"
      @show-profile="activeView = 'profile'; isNotificationOpen = false" @show-chat="activeView = 'chat'"
      @open-user-selector="isUserSelectorOpen = true" :class="activeChatId ? 'hidden md:flex' : 'flex'"
      class="flex-shrink-0 z-30" />

    <ChatList ref="chatListRef" v-if="activeView === 'chat'" class="flex-1 md:flex-none md:w-[320px]"
      :class="activeChatId ? 'hidden md:flex' : 'flex'" @chat-selected="selectChat"
      @open-user-selector="isUserSelectorOpen = true" :active-chat-id="activeChatId" />

    <UserSelector v-if="isUserSelectorOpen" @close="isUserSelectorOpen = false" @chat-created="handleNewChat" />

    <template v-if="activeView === 'chat'">
      <ChatWindow v-if="activeChatId" :chat-id="activeChatId" :chat-name="activeChatName"
        class="flex-1 absolute md:relative inset-0 z-20 md:z-0 bg-white" @back="activeChatId = null" />
      <div v-else class="hidden md:flex flex-1 items-center justify-center text-slate-400">
        Выберите чат
      </div>
    </template>

    <ProfileView v-else class="flex-1 md:w-[280px]" @show-chat="activeView = 'chat'" />

    <NotificationPanel :is-open="isNotificationOpen" @close="isNotificationOpen = false" />
  </div>
</template>