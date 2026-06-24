<script setup>
import { ref, onMounted } from 'vue'
import Sidebar from '../components/Sidebar.vue'
import ChatList from '../components/ChatList.vue'
import ChatWindow from '../components/ChatWindow.vue'
import ProfileView from './ProfileView.vue'
import NotificationPanel from '../components/NotificationPanel.vue'
import UserSelector from '../components/UserSelector.vue'
import FriendsList from '../components/FriendsList.vue'

const isNotificationOpen = ref(false)
const activeChatId = ref(null)
const activeChatName = ref('')
const activeView = ref('chat') // 'chat' или 'profile'
const isUserSelectorOpen = ref(false);
const chatListRef = ref(null)
const isFriendsOpen = ref(false);

// Метод для выбора чата
const selectChat = (chat) => {
  activeChatId.value = chat.id
  activeChatName.value = chat.name
  activeView.value = 'chat' // При выборе чата переключаемся на вид чата
}

const handleNewChat = () => {
  isUserSelectorOpen.value = false;
  if (chatListRef.value) {
    chatListRef.value.fetchChats()
  }
}

const closeSidePanels = () => {
  isNotificationOpen.value = false;
  isFriendsOpen.value = false;
};

const toggleNotifications = () => {
  if (!isNotificationOpen.value) { // Если открываем
    closeSidePanels(); // Сначала закрываем всё
    isNotificationOpen.value = true;
  } else {
    isNotificationOpen.value = false;
  }
  console.log(isNotificationOpen.value);
};

const toggleFriends = () => {
  if (!isFriendsOpen.value) { // Если открываем
    closeSidePanels(); // Сначала закрываем всё
    isFriendsOpen.value = true;
  } else {
    isFriendsOpen.value = false;
  }
};
</script>

<template>
  <div class="flex w-full h-screen bg-slate-100 overflow-hidden font-sans antialiased">

    <Sidebar :activeTab="activeView" @click-friends="toggleFriends" @toggle-notifications="toggleNotifications"
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

    <ProfileView v-if="activeView === 'profile'" class="flex-1 md:w-[280px]" @show-chat="activeView = 'chat'" />



    <NotificationPanel v-if="isNotificationOpen" :is-open="isNotificationOpen" @close="isNotificationOpen = false" />

    <FriendsList v-if="isFriendsOpen" :is-open="isFriendsOpen" @close="isFriendsOpen = false" />
  </div>
</template>