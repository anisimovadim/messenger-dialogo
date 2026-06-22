<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import api from '@/services/api';

const router = useRouter()
const isRegistering = ref(false)
const step = ref('form');
const isLoading = ref(false); // Новая переменная для прелоадера

const email = ref('')
const password = ref('')
const confirmPassword = ref('')
const username = ref('')
const errorMessage = ref('')
const verificationCode = ref('');

// Функция валидации
const validateForm = () => {
  if (password.value.length < 6) {
    errorMessage.value = 'Пароль должен содержать минимум 6 символов'
    return false
  }
  if (isRegistering.value && step.value === 'form') {
    if (!username.value.trim()) {
      errorMessage.value = 'Введите имя пользователя'
      return false
    }
    if (password.value !== confirmPassword.value) {
      errorMessage.value = 'Пароли не совпадают'
      return false
    }
  }
  return true
}

const handleSubmit = async () => {
  errorMessage.value = '';
  if (!validateForm()) return;

  isLoading.value = true; // Включаем прелоадер

  try {
    if (!isRegistering.value) {
      // ОБЫЧНЫЙ ВХОД
      const data = await api.post('/api/login', {
        email: email.value,
        password: password.value
      });
      localStorage.setItem('token', data.token);
      localStorage.setItem('username', data.username);
      localStorage.setItem('user_id', data.user_id);
      router.push('/');
    }
    else if (step.value === 'form') {
      // ЭТАП 1: Запрос кода подтверждения
      await api.post('/api/send-code', {
        email: email.value,
        password: password.value,
        username: username.value
      });
      step.value = 'verify'; // Переключаем UI на ввод кода
    }
    else if (step.value === 'verify') {
      // ЭТАП 2: Финальная регистрация
      await api.post('/api/verify-and-register', {
        email: email.value,
        code: verificationCode.value
      });
      alert('Регистрация успешна!');
      isRegistering.value = false;
      step.value = 'form';
    }
  } catch (err) {
    errorMessage.value = err.message || 'Произошла ошибка';
  } finally {
    isLoading.value = false; // Выключаем прелоадер в любом случае
  }
};
</script>

<template>
  <div class="h-screen w-screen bg-white text-slate-800 flex flex-col justify-center items-center p-4">
    <div class="w-full max-w-[400px]">

      <div class="mb-8 text-center">
        <h1 class="text-3xl font-black text-[#1E1B4B] mb-2">
          {{ isRegistering ? 'Создать аккаунт' : 'Добро пожаловать' }}
        </h1>
        <p class="text-slate-400 text-sm">
          {{ isRegistering ? 'Заполните данные для регистрации' : 'Введите данные для входа в Dialogo' }}
        </p>
      </div>

      <form @submit.prevent="handleSubmit" class="space-y-4">
        <div v-if="errorMessage" class="p-3 bg-rose-50 text-rose-600 rounded-xl text-xs text-center font-medium">
          {{ errorMessage }}
        </div>
        <div v-if="step==='form'">
          <div v-if="isRegistering">
            <label class="block text-xs font-bold text-slate-700 mb-1.5">Имя пользователя</label>
            <input v-model="username" type="text" required placeholder="Ваше имя"
              class="w-full px-4 py-3 bg-slate-50 border border-slate-200 rounded-xl focus:border-[#6344F2] outline-none transition-all" />
          </div>

          <div>
            <label class="block text-xs font-bold text-slate-700 mb-1.5">Email</label>
            <input v-model="email" type="email" required placeholder="example@mail.com"
              class="w-full px-4 py-3 bg-slate-50 border border-slate-200 rounded-xl focus:border-[#6344F2] outline-none transition-all" />
          </div>

          <div>
            <label class="block text-xs font-bold text-slate-700 mb-1.5">Пароль</label>
            <input v-model="password" type="password" required placeholder="••••••••"
              class="w-full px-4 py-3 bg-slate-50 border border-slate-200 rounded-xl focus:border-[#6344F2] outline-none transition-all" />
          </div>

          <div v-if="isRegistering">
            <label class="block text-xs font-bold text-slate-700 mb-1.5">Подтверждение пароля</label>
            <input v-model="confirmPassword" type="password" required placeholder="••••••••"
              class="w-full px-4 py-3 bg-slate-50 border border-slate-200 rounded-xl focus:border-[#6344F2] outline-none transition-all" />
          </div>
        </div>

        <div v-if="isRegistering && step === 'verify'">
          <label class="block text-xs font-bold text-slate-700 mb-1.5">Введите код подтверждения</label>
          <input v-model="verificationCode" type="text" required placeholder="000000"
            class="w-full px-4 py-3 bg-slate-50 border border-slate-200 rounded-xl outline-none" />
        </div>

        <button type="submit" :disabled="isLoading"
          class="w-full py-3.5 bg-[#6344F2] hover:bg-[#5538d6] text-white font-bold rounded-xl transition-all shadow-lg shadow-[#6344F2]/20 disabled:opacity-70 disabled:cursor-not-allowed flex justify-center items-center">
          <span v-if="isLoading" class="animate-spin mr-2 h-4 w-4 border-2 border-white border-t-transparent rounded-full"></span>
          {{ isLoading ? 'Загрузка...' : (isRegistering ? (step === 'verify' ? 'Подтвердить' : 'Зарегистрироваться') : 'Войти') }}
        </button>
      </form>

      <div class="mt-6 text-center text-sm">
        <span class="text-slate-500">
          {{ isRegistering ? 'Уже есть аккаунт?' : 'Еще нет аккаунта?' }}
        </span>
        <button @click="isRegistering = !isRegistering; errorMessage = ''; step = 'form'"
          class="text-[#6344F2] font-bold ml-1 hover:underline cursor-pointer">
          {{ isRegistering ? 'Войти' : 'Создать аккаунт' }}
        </button>
      </div>
    </div>
  </div>
</template>