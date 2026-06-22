<script setup>
import { onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'

const router = useRouter()
const route = useRoute()

onMounted(() => {
  const token = route.query.token
  const username = route.query.username
  const userId = route.query.user_id

  if (token) {
    localStorage.setItem('token', token)
    localStorage.setItem('username', username)
    localStorage.setItem('user_id', userId)
    router.push('/') // Переход в чат
  } else {
    router.push('/login') // Ошибка, возвращаем к логину
  }
})
</script>

<template>
  <div>Загрузка авторизации...</div>
</template>