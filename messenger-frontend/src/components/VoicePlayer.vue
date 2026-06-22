<script setup>
import { ref, onMounted, onUnmounted, watch } from 'vue'

const props = defineProps(['src'])
const emit = defineEmits(['play'])

const isPlaying = ref(false)
const currentTime = ref(0)
const duration = ref(0)
const audio = new Audio(props.src)

const heights = [12, 16, 20, 24, 28, 24, 20, 16, 12, 16, 20, 24, 28, 24, 20, 16, 12, 16, 12]

// Форматирование времени
const formatTime = (time) => {
  const mins = Math.floor(time / 60)
  const secs = Math.floor(time % 60)
  return `${mins}:${secs.toString().padStart(2, '0')}`
}

onMounted(() => {
  audio.onloadedmetadata = () => { duration.value = audio.duration }
  audio.ontimeupdate = () => { currentTime.value = audio.currentTime }
  audio.onended = () => { isPlaying.value = false }
})

const togglePlay = () => {
  // Эмитим событие, чтобы родитель знал: "этот плеер нажат"
  if (!isPlaying.value) {
    emit('play', audio) 
  }
  
  if (isPlaying.value) {
    audio.pause()
  } else {
    audio.play()
  }
  isPlaying.value = !isPlaying.value
}

// Внешний метод для принудительной остановки
const stop = () => {
  audio.pause()
  audio.currentTime = 0
  isPlaying.value = false
}

defineExpose({ stop })

onUnmounted(() => { audio.pause() })
</script>

<template>
  <div class="flex items-center gap-3">
    <button @click="togglePlay" class="w-10 h-10 rounded-full bg-[#6344F2] flex items-center justify-center text-white shrink-0 hover:scale-105 transition-transform">
      <svg v-if="!isPlaying" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" class="w-5 h-5 ml-0.5"><path d="M8 5v14l11-7z"/></svg>
      <svg v-else xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" class="w-5 h-5"><path d="M6 19h4V5H6v14zm8-14v14h4V5h-4z"/></svg>
    </button>

    <div class="flex items-center gap-0.5 h-8">
      <div v-for="(h, index) in heights" :key="index" 
           class="w-1 rounded-full transition-all duration-300"
           :class="isPlaying ? 'bg-[#6344F2] animate-bounce' : 'bg-slate-400'"
           :style="{ height: h + 'px', animationDelay: (index * 0.1) + 's' }">
      </div>
    </div>

    <span class="text-xs font-bold text-slate-700 tabular-nums">
      {{ formatTime(isPlaying ? currentTime : duration) }}
    </span>
  </div>
</template>