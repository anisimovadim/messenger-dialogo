<script setup>
import { ref, onMounted, watch, nextTick, onUnmounted } from "vue";
import VoicePlayer from "./VoicePlayer.vue";
import api, { WS_URL } from "@/services/api";

const props = defineProps(["chatId", "chatName"]);
const emit = defineEmits(["back"]);

let pingInterval = null;

const messages = ref([]);
const messageText = ref("");
const myId = parseInt(localStorage.getItem("user_id"));
let socket = null;

const messagesContainer = ref(null);
const fileInput = ref(null);

const isPartnerTyping = ref(false);
const partnerName = ref("");
let typingTimeout = null;


const recordingTime = ref(0);
const isRecording = ref(false);
const isCancelled = ref(false);
let mediaRecorder = null;
let audioChunks = [];
let timerInterval = null;

const playerRefs = ref([]);

const formatTime = (seconds) => {
  const mins = Math.floor(seconds / 60);
  const secs = seconds % 60;
  return `${mins}:${secs.toString().padStart(2, '0')}`;
};

const onPlay = (currentAudioInstance) => {
  playerRefs.value.forEach((player) => {
    if (player && player.audio !== currentAudioInstance) {
      player.stop();
    }
  });
};

const scrollToBottom = async () => {
  await nextTick();
  if (messagesContainer.value) {
    messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight;
  }
};

// --- ФУНКЦИИ ---
const deleteMessage = (id) => {
  socket.send(
    JSON.stringify({ type: "delete_message", id: id, chat_id: props.chatId }),
  );
};

const editMessage = (msg) => {
  const newText = prompt("Редактировать сообщение:", msg.text);
  if (newText && newText !== msg.text) {
    socket.send(
      JSON.stringify({
        type: "edit_message",
        id: msg.id,
        text: newText,
        chat_id: props.chatId,
      }),
    );
  }
};

const sendReadReceipt = () => {
  if (socket && socket.readyState === WebSocket.OPEN) {
    socket.send(
      JSON.stringify({
        chat_id: props.chatId,
        user_id: myId,
        type: "read_receipt",
      }),
    );
  }
};

const sendTypingStatus = (isTyping) => {
  if (socket && socket.readyState === WebSocket.OPEN) {
    lastTypingStatus = isTyping;
    socket.send(
      JSON.stringify({
        chat_id: props.chatId,
        user_id: myId, // Обязательно передаем ID отправителя
        username: localStorage.getItem("username") || "Пользователь",
        type: "typing",
        is_typing: isTyping,
      }),
    );
  }
};

let lastTypingStatus = false;
const handleInputChange = () => {
  sendTypingStatus(true);
  clearTimeout(typingTimeout);
  typingTimeout = setTimeout(() => {
    sendTypingStatus(false);
  }, 1500);
};

const triggerFileInput = () => {
  fileInput.value.click();
};
const handleFileUpload = async (event) => {
  const file = event.target.files[0];
  if (!file) return;

  const formData = new FormData();
  formData.append("file", file);

  try {
    // Теперь результат вызова - это сразу данные (JSON с сервера)
    const data = await api.post("/api/upload", formData);
    
    console.log("Данные от сервера:", data); // Для проверки в консоли

    // Обращаемся к свойствам напрямую (data.url, data.filename и т.д.)
    socket.send(JSON.stringify({
      chat_id: props.chatId,
      user_id: myId,
      text: data.url, 
      type: "file",
      filename: data.filename,
      original_name: data.original_name
    }));

    event.target.value = '';
  } catch (err) {
    console.error("Ошибка загрузки:", err);
    alert("Не удалось загрузить файл: " + err.message);
  }
};

const startRecording = async () => {
  try {
    const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
    mediaRecorder = new MediaRecorder(stream);
    audioChunks = [];

    mediaRecorder.ondataavailable = (event) => {
      if (event.data.size > 0) audioChunks.push(event.data);
    };

    mediaRecorder.onstop = async () => {
      if (isCancelled.value) {
        isCancelled.value = false; // Сбрасываем флаг для будущих записей
        stream.getTracks().forEach((track) => track.stop());
        return;
      }
      const audioBlob = new Blob(audioChunks, { type: "audio/webm" });
      const formData = new FormData();
      formData.append("file", audioBlob, "voice.webm");

      try {
        // ИСПОЛЬЗУЕМ AXIOS ВМЕСТО FETCH
        const data = await api.post("/api/upload", formData);

        socket.send(
          JSON.stringify({
            chat_id: props.chatId,
            user_id: myId,
            text: data.url,
            type: "voice",
          }),
        );
      } catch (err) {
        console.error("Ошибка отправки голосового:", err);
      }
      stream.getTracks().forEach((track) => track.stop());
    };
    mediaRecorder.start();
    recordingTime.value = 0;
    timerInterval = setInterval(() => recordingTime.value++, 1000);
    isRecording.value = true;
  } catch (err) {
    console.error("Ошибка микрофона:", err);
    alert("Не удалось получить доступ к микрофону");
  }
};

const stopRecording = () => {
  if (mediaRecorder && mediaRecorder.state !== "inactive") {
    mediaRecorder.stop();
    isRecording.value = false;
    clearInterval(timerInterval);
  }
};

const cancelRecording = () => {
  if (mediaRecorder && mediaRecorder.state !== "inactive") {
    isCancelled.value = true; // Устанавливаем флаг отмены
    mediaRecorder.stop();
  }
  isRecording.value = false;
  clearInterval(timerInterval);
  audioChunks = [];
};

const loadHistory = async (id) => {
  try {
    // Используем Axios вместо fetch
    const data = await api.get(`/api/messages?chat_id=${id}`);
    messages.value = data;
    scrollToBottom();
    sendReadReceipt();
  } catch (err) {
    console.error("Ошибка загрузки истории:", err);
  }
};

const connectWebSocket = () => {
  if (socket) socket.close();
  isPartnerTyping.value = false;

  socket = new WebSocket(`${WS_URL}/ws?user_id=${myId}`);
  socket.onopen = () => {
    sendReadReceipt();
    // Запускаем отправку пинга каждые 30 секунд
    pingInterval = setInterval(() => {
      if (socket.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify({ type: "ping" }));
      }
    }, 30000);
  };

  socket.onclose = () => {
    clearInterval(pingInterval); // Очищаем интервал при закрытии
    console.log("Сокет закрыт, переподключаемся...");
    setTimeout(connectWebSocket, 3000); // Авто-переподключение через 3 секунды
  };

  socket.onmessage = (event) => {
    const msg = JSON.parse(event.data);
    if (msg.chat_id !== props.chatId) return;

    if (msg.type === "typing") {
      if (msg.user_id !== myId) {
        isPartnerTyping.value = msg.is_typing;
        partnerName.value = msg.username;
      }
    } else if (msg.type === "read_receipt") {
      // Сейчас: помечает все сообщения отправителя как прочитанные
      messages.value.forEach((m) => {
        if (m.user_id === myId) m.is_read = true;
      });
    } else if (msg.type === "edit_message") {
      const index = messages.value.findIndex((m) => m.id === msg.id);
      if (index !== -1) messages.value[index].text = msg.text;
    } else if (msg.type === "delete_message") {
      messages.value = messages.value.filter((m) => m.id !== msg.id);
    } else {
      messages.value.push(msg);
      isPartnerTyping.value = false;
      scrollToBottom();
      if (msg.user_id !== myId) sendReadReceipt();
    }
  };
};

const sendMessage = () => {
  if (!messageText.value.trim() || !socket) return;
  clearTimeout(typingTimeout);
  sendTypingStatus(false);
  socket.send(
    JSON.stringify({
      chat_id: props.chatId,
      user_id: myId,
      text: messageText.value,
      type: "text",
    }),
  );
  messageText.value = "";
};

watch(
  () => props.chatId,
  (newId) => {
    if (newId) {
      loadHistory(newId);
      connectWebSocket();
    }
  },
  { immediate: true },
);
onMounted(() => scrollToBottom());
onUnmounted(() => {
  if (socket) socket.close();
  clearInterval(pingInterval);
  clearTimeout(typingTimeout);
});
</script>

<template>
  <div class="h-screen flex flex-col bg-[#F8FAFC] overflow-hidden">
    <div class="h-[76px] px-6 border-b border-slate-200 flex justify-between items-center bg-white flex-shrink-0">
      <div class="flex items-center gap-3">
        <button @click="$emit('back')" class="md:hidden p-1 -ml-2 text-slate-500 hover:text-[#6344F2]">
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M15 19l-7-7 7-7" />
          </svg>
        </button>
        <div
          class="w-10 h-10 bg-slate-200 text-slate-700 rounded-full flex items-center justify-center font-bold text-sm">
          {{ props.chatName.substring(0, 2).toUpperCase() }}
        </div>
        <div>
          <h3 class="font-bold text-slate-900 text-sm leading-tight">
            {{ props.chatName }}
          </h3>
          <span v-if="isPartnerTyping" class="text-xs text-[#6344F2] font-bold italic">печатает...</span>
          <span v-else class="text-xs text-slate-400 font-medium">в сети</span>
        </div>
      </div>
    </div>

    <div ref="messagesContainer" class="flex-1 overflow-y-auto p-6 space-y-4 bg-[#F1F5F9]/40 flex flex-col">
      <div v-for="msg in messages" :key="msg.id" :class="msg.user_id === myId ? 'justify-end' : 'justify-start'"
        class="flex w-full group">
        <div class="flex flex-col max-w-[70%]">
          <div v-if="msg.user_id === myId"
            class="opacity-0 group-hover:opacity-100 flex gap-2 mb-1 justify-end transition-opacity">
            <button v-if="msg.type === 'text'" @click="editMessage(msg)"
              class="text-[10px] text-slate-400 hover:text-blue-600 font-bold">
              ред.
            </button>
            <button @click="deleteMessage(msg.id)" class="text-[10px] text-slate-400 hover:text-red-600 font-bold">
              уд.
            </button>
          </div>

          <div :class="msg.user_id === myId
            ? 'bg-[#E1EFFE] text-slate-900 rounded-2xl rounded-tr-none'
            : 'bg-white text-slate-900 rounded-2xl rounded-tl-none border border-slate-100 shadow-sm'
            " class="px-4 py-2.5 text-sm font-medium leading-relaxed relative">
            <div v-if="msg.user_id !== myId" class="text-[11px] font-bold text-[#6344F2] mb-1">
              {{ msg.username }}
            </div>

            <div v-if="msg.type === 'image'" class="rounded-xl overflow-hidden my-1 max-w-[320px]">
              <img :src="msg.text" class="w-full h-auto object-cover border border-slate-200/60 rounded-xl"
                @load="scrollToBottom" />
            </div>

            <div v-else-if="msg.type === 'voice'"
              class="my-1.5 min-w-[260px] bg-white/50 border border-slate-200/50 rounded-2xl p-2 flex items-center gap-3">
              <VoicePlayer :ref="(el) => playerRefs.push(el)" :src="msg.text" @play="onPlay" />
            </div>

            <div v-else-if="msg.type === 'file'" class="my-1.5">
              <div v-if="msg.original_name.match(/\.(jpg|jpeg|png|gif|webp)$/i)"
                class="rounded-xl overflow-hidden max-w-[240px]">
                <img :src="msg.text" class="w-full h-auto object-cover border border-slate-200/60 rounded-xl" />
              </div>

              <a v-else :href="msg.text" target="_blank"
                class="flex items-center gap-3 bg-white border border-slate-200 rounded-xl p-3 hover:bg-slate-50 transition-colors">
                <div class="bg-indigo-100 p-2 rounded-lg text-indigo-600">
                  <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                      d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                  </svg>
                </div>
                <div class="text-sm font-bold text-slate-800 truncate">
                  {{ msg.original_name || 'Файл' }}
                </div>
              </a>
            </div>

            <div v-else>{{ msg.text }}</div>

            <div class="flex justify-end items-center gap-1.5 text-[9px] text-slate-400 font-semibold mt-1">
              <span>{{ new Date(msg.created_at).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" }) }}</span>

              <span v-if="msg.user_id === myId" class="inline-flex">
                <svg v-if="msg.is_read" class="w-4 h-4 text-blue-500" viewBox="0 0 24 24" fill="none"
                  stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M4 12l4 4 6-8" />
                  <path d="M10 12l4 4 6-8" />
                </svg>

                <svg v-else class="w-4 h-4 text-slate-400" viewBox="0 0 24 24" fill="none" stroke="currentColor"
                  stroke-width="3" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M7 12l4 4 6-8" />
                </svg>
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="p-5 bg-white border-t border-slate-200 flex-shrink-0">
      <form @submit.prevent="sendMessage"
        class="flex items-center gap-3 bg-[#F1F3F7] px-4 py-2.5 rounded-2xl transition-colors"
        :class="isRecording ? 'border-2 border-red-500 bg-red-50' : ''">

        <input type="file" ref="fileInput" class="hidden text-base" @change="handleFileUpload" />

        <button type="button" @click="isRecording ? cancelRecording() : triggerFileInput()"
          class="text-slate-400 transition-colors" :class="isRecording ? 'hover:text-red-500' : 'hover:text-[#6344F2]'">
          <svg v-if="isRecording" class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
          <svg v-else class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5"
              d="m18.375 12.739-7.693 7.693a4.5 4.5 0 0 1-6.364-6.364l10.94-10.94A3 3 0 1 1 19.5 7.372L8.552 18.32a1.5 1.5 0 0 1-2.12-2.121L16.202 7.5" />
          </svg>
        </button>

        <div v-if="isRecording" class="flex-1 text-sm font-bold text-red-500 flex items-center gap-2">
          <span class="relative flex h-3 w-3">
            <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-red-400 opacity-75"></span>
            <span class="relative inline-flex rounded-full h-3 w-3 bg-red-500"></span>
          </span>
          {{ formatTime(recordingTime) }}
        </div>
        <input v-else v-model="messageText" @input="handleInputChange" type="text" placeholder="Введите сообщение..."
          class="flex-1 bg-transparent text-base font-medium text-slate-800 focus:outline-none" />

        <button type="button" @click="isRecording ? stopRecording() : startRecording()"
          class="p-2 rounded-full transition-all"
          :class="isRecording ? 'bg-green-500 text-white scale-110' : 'text-slate-400 hover:text-[#6344F2]'">
          <svg v-if="isRecording" class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7" />
          </svg>
          <svg v-else class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5"
              d="M12 18.75a6 6 0 0 0 6-6v-1.5m-6 7.5a6 6 0 0 1-6-6v-1.5m6 7.5v3.75m-3.75 0h7.5M12 15.75a3 3 0 0 1-3-3v-6a3 3 0 0 1 6 0v6a3 3 0 0 1-3 3z" />
          </svg>
        </button>
      </form>
    </div>
  </div>
</template>
