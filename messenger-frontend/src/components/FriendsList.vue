<script setup>
import { ref, onMounted } from 'vue';
import api from '@/services/api';

const props = defineProps(['isOpen']);
const emit = defineEmits(['close']);

const myId = localStorage.getItem("user_id");
const friends = ref([]);

const fetchFriends = async () => {
  try {
    const data = await api.get(`/api/friends?user_id=${myId}`);
    friends.value = data || [];
  } catch (err) {
    console.error("Ошибка загрузки друзей:", err);
  }
};

const acceptFriend = async (friendId) => {
  try {
    await api.post('/api/friends/accept', {
      user_id: parseInt(myId),
      friend_id: friendId
    });
    fetchFriends();
  } catch (err) {
    console.error("Ошибка при принятии запроса:", err);
  }
};

onMounted(fetchFriends);
</script>

<template>
  <div :class="[
    props.isOpen ? 'fixed inset-0 z-50 bg-white md:static md:bg-transparent' : 'hidden',
    'w-full md:w-[320px] h-full border-l border-slate-200 flex flex-col flex-shrink-0'
  ]">
    
    <div class="h-[76px] px-6 border-b border-slate-200 flex items-center justify-between bg-white">
      <h2 class="font-black text-slate-900 text-sm uppercase tracking-tight">Друзья</h2>
      <button @click="$emit('close')" class="p-2 text-slate-500 hover:text-slate-800">✕</button>
    </div>

    <div class="flex-1 overflow-y-auto p-4 space-y-1">
      <div v-for="f in friends" :key="f.id" 
           class="flex items-center gap-3 p-3 rounded-2xl hover:bg-slate-50 transition-all">
        
        <div class="w-12 h-12 bg-slate-100 text-slate-600 rounded-full flex items-center justify-center font-bold text-sm">
          {{ f.username.substring(0, 2).toUpperCase() }}
        </div>

        <div class="flex-1 min-w-0">
          <h4 class="font-bold text-sm text-slate-900 truncate">{{ f.username }}</h4>
          <p class="text-xs font-medium" :class="f.status === 'pending' ? 'text-amber-600' : 'text-slate-500'">
            {{ f.status === 'pending' ? 'Ожидает подтверждения' : 'Друг' }}
          </p>
        </div>

        <button v-if="f.status === 'pending'" 
                @click="acceptFriend(f.id)"
                class="px-3 py-1.5 bg-[#6344F2] hover:bg-[#5538d6] text-white text-[10px] font-bold rounded-lg transition-colors">
          Принять
        </button>
      </div>
      
      <div v-if="friends.length === 0" class="text-center p-10 text-slate-400 text-sm">
        Пока никого нет...
      </div>
    </div>
  </div>
</template>