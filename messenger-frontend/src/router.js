import { createRouter, createWebHistory } from 'vue-router'
import LoginView from './views/LoginView.vue'
import ChatView from './views/ChatView.vue'

const routes = [
  { path: '/login', component: LoginView },
  { path: '/', component: ChatView },
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// Защита роутов (Navigation Guard)
router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('token')
  if (to.path !== '/login' && !token) {
    next('/login')
  } else if (to.path === '/login' && token) {
    next('/')
  } else {
    next()
  }
})

export default router