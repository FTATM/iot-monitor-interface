<template>
  <div class="flex justify-center items-center min-h-screen bg-slate-100">
    <div class="bg-white w-full max-w-[400px] p-10 rounded-xl shadow-[0_4px_20px_rgba(0,0,0,0.08)]">

      <div class="text-center mb-[30px]">
        <h2 class="m-0 text-slate-800 text-[1.75rem]  font-bold">System Login</h2>
        <p class="text-slate-500 mt-2 text-[0.95rem]">Enter your credentials to access the dashboard</p>
      </div>

      <form @submit.prevent="handleLogin" class="block">
        <!-- Error Message Banner -->
        <div v-if="messageLogin"
          class="bg-red-50 text-red-500 p-3 rounded-md mb-5 text-[0.9rem] text-center border border-red-300">
          {{ messageLogin }}
        </div>

        <div class="mb-5">
          <label for="username" class="block mb-2 text-slate-700 font-medium text-[0.95rem]">
            Username
          </label>
          <input type="text" id="username" v-model="username" placeholder="Enter your username" required
            :disabled="isUserLoginLoading"
            class="w-full p-3 border border-slate-300 rounded-md text-black text-base bg-slate-50 box-border transition-colors duration-200 focus:outline-none focus:border-blue-500 focus:ring-[3px] focus:ring-blue-500/10 disabled:bg-slate-200 disabled:text-slate-400 disabled:cursor-not-allowed" />
        </div>

        <div class="mb-5">
          <label for="password" class="block mb-2 text-slate-700 font-medium text-[0.95rem]">
            Password
          </label>
          <input type="password" id="password" v-model="password" placeholder="Enter your password" required
            :disabled="isUserLoginLoading"
            class="w-full p-3 border border-slate-300 rounded-md text-base bg-slate-50 box-border transition-colors duration-200 focus:outline-none focus:border-blue-500 focus:ring-[3px] focus:ring-blue-500/10 disabled:bg-slate-200 disabled:text-slate-400 disabled:cursor-not-allowed" />
        </div>

        <button type="submit" :disabled="isUserLoginLoading"
          class="w-full p-3 bg-blue-500 text-white border-none rounded-md text-base font-semibold cursor-pointer transition-colors mt-2.5 hover:bg-blue-600 disabled:bg-slate-400 disabled:cursor-not-allowed">
          {{ isUserLoginLoading ? 'Authenticating...' : 'Sign In' }}
        </button>
      </form>

      <div class="text-center mt-6 text-[0.8rem] text-slate-400">
        Accounts are provisioned by the system administrator.
      </div>

    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { useMutation } from '@/composables/useMutation'
import { useUserStore } from '@/stores/useUserStore';

const router = useRouter();


// --- STORE ---
const userStore = useUserStore();
const { setUser } = userStore;

const messageLogin = ref("");
const username = ref('');
const password = ref('');

const {
  data: userLogin,
  res: userLoginRes,
  isLoading: isUserLoginLoading,
  error: userLoginError,
  execute: userLoginApi
} = useMutation();

const handleLogin = async () => {
  messageLogin.value = ""
  await userLoginApi('/user/login', { username: username.value, password: password.value }, "POST")

  if (!userLoginRes.value.ok) {
    if (userLoginRes.value.status >= 500 || !userLoginError.value?.message) {
      messageLogin.value = "Cannot reach the server. Please check your internet connection or try again later"
    } else {
      messageLogin.value = userLoginError.value.message
    }
    return
  }

  setUser(
    {
      id: userLogin.value.data.userId,
      firstName: userLogin.value.data.firstName,
      lastName: userLogin.value.data.lastName,
      fullName: `${userLogin.value.data.firstName} ${userLogin.value.data.lastName}`
    }
  )
  router.push('/dashboard');
};
</script>