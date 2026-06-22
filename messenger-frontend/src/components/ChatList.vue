<script setup>
import { ref, onMounted, onUnmounted } from "vue";
import api from "@/services/api";
import { WS_URL } from "@/services/api";

const props = defineProps(["activeChatId"]);
const emit = defineEmits(["chat-selected", "open-user-selector"]);

const isCreatingChat = ref(false);

const currentTab = ref("all");
const chats = ref([]);
let listSocket = null;

const myId = localStorage.getItem("user_id");

const fetchChats = async () => {
  if (!myId) return;
  try {
    const data = await api.get(`/api/chats?user_id=${myId}`);

    const chatData = Array.isArray(data) ? data : [];

    chats.value = chatData.map((chat) => ({
      ...chat,
      name: chat.name || "Без имени",
      unread_count: chat.unread_count || 0,
      lastMessage: chat.last_message || "Нет сообщений"
    }));
  } catch (err) {
    console.error("Ошибка загрузки чатов:", err);
    chats.value = [];
  }
};

// Подключаем сокет в списке чатов для отслеживания непрочитанных сообщений
const connectListSocket = () => {
  if (listSocket) listSocket.close();
  listSocket = new WebSocket(`${WS_URL}/ws?user_id=${myId}`);

  listSocket.onmessage = (event) => {
    const msg = JSON.parse(event.data);

    if (msg.type === "new_chat") {
      const participants = msg.text.split(',').map(Number);
      // Если текущий пользователь есть среди участников, обновляем список
      if (participants.includes(parseInt(myId))) {
        fetchChats(); // Перезагружаем список чатов с сервера
      }
      return;
    }

    // СТРОГАЯ ПРОВЕРКА: Если это статус набора текста, список чатов его игнорирует!
    if (msg.type === "typing" || msg.type === "read_receipt") {
      return;
    }

    const targetChat = chats.value.find((c) => c.id === msg.chat_id);
    if (targetChat) {
      // Подставляем красивый текст в зависимости от типа медиа
      if (msg.type === "image") {
        targetChat.lastMessage = "🖼 Фотография";
      } else if (msg.type === "voice") {
        targetChat.lastMessage = "🎤 Голосовое сообщение";
      } else if (msg.type === "file") {
        // Здесь мы проверяем, картинка это или другой файл
        const isImage = msg.original_name && msg.original_name.match(/\.(jpg|jpeg|png|gif|webp)$/i);
        targetChat.lastMessage = isImage ? "🖼 Фотография" : `📁 ${msg.original_name || 'Файл'}`;
      } else {
        targetChat.lastMessage = msg.text;
      }

      if (msg.chat_id !== props.activeChatId) {
        targetChat.unread_count++;

        const notificationEvent = new CustomEvent("new-notification", {
          detail: {
            id: Date.now(),
            user: `@${msg.username}`,
            action: `прислал(а) сообщение в чат`,
            chatName: targetChat.name,
            time: new Date().toLocaleTimeString([], {
              hour: "2-digit",
              minute: "2-digit",
            }),
            avatarColor: "bg-indigo-100 text-indigo-800",
            type: "mention",
          },
        });
        window.dispatchEvent(notificationEvent);
      }
    }
  };
};

const handleChatClick = (chat) => {
  chat.unread_count = 0; // Сбрасываем счетчик при клике
  emit("chat-selected", chat);
};


onMounted(() => {
  fetchChats();
  connectListSocket();
});

onUnmounted(() => {
  if (listSocket) listSocket.close();
});

defineExpose({
  fetchChats
})
</script>

<template>
  <div :class="[
    'h-full bg-white border-r border-slate-200 flex flex-col flex-shrink-0 relative transition-all duration-300',
    props.activeChatId ? 'hidden md:flex w-full md:w-[320px]' : 'w-full md:w-[320px]'
  ]">
    <div class="p-6 pb-4 flex justify-between items-center">
      <h2 class="text-2xl font-black text-slate-900 tracking-tight">Чаты</h2>

      <button @click="$emit('open-user-selector')"
        class="w-8 h-8 flex items-center justify-center bg-slate-100 hover:bg-slate-200 text-slate-600 rounded-lg transition-colors"
        title="Новый чат">
        <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
        </svg>
      </button>
    </div>

    <div class="px-6 mb-4 flex gap-4 text-sm font-bold text-slate-400 border-b border-slate-100 pb-2">
      <button @click="currentTab = 'all'" :class="currentTab === 'all' ? 'text-slate-900' : ''"
        class="pb-1 relative cursor-pointer">
        Все
        <div v-if="currentTab === 'all'"
          class="absolute bottom-[-9px] left-0 right-0 h-[3px] bg-[#6344F2] rounded-full"></div>
      </button>
    </div>

    <div class="flex-1 overflow-y-auto px-4 space-y-1">
      <div v-for="chat in chats" :key="chat.id" @click="handleChatClick(chat)" :class="chat.id === props.activeChatId
        ? 'bg-[#1E1B4B] text-white'
        : 'hover:bg-slate-50 text-slate-800'
        " class="flex items-center gap-3 p-3 rounded-2xl cursor-pointer transition-all group">
        <div
          class="w-12 h-12 bg-indigo-100 text-indigo-700 rounded-full flex items-center justify-center font-bold text-sm flex-shrink-0">
          {{ chat.name.substring(0, 2).toUpperCase() }}
        </div>

        <div class="flex-1 min-w-0">
          <div class="flex justify-between items-baseline mb-0.5">
            <h4 :class="chat.id === props.activeChatId ? 'text-white' : 'text-slate-900'
              " class="font-bold text-sm truncate">
              {{ chat.name }}
            </h4>

            <span v-if="chat.unread_count > 0"
              class="bg-[#6344F2] text-white text-[10px] font-bold px-1.5 py-0.5 rounded-full min-w-[18px] text-center">
              {{ chat.unread_count }}
            </span>
          </div>
          <p :class="chat.id === props.activeChatId
            ? 'text-slate-300'
            : 'text-slate-500'
            " class="text-xs truncate font-medium">
            {{ chat.lastMessage }}
          </p>
        </div>
      </div>
    </div>
  </div>
</template>
