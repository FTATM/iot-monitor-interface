<template>
  <div class="login-wrapper">
    <div class="login-card">
      <div class="login-header">
        <h2>System Login</h2>
        <p>Enter your credentials to access the dashboard</p>
      </div>

      <form @submit.prevent="handleLogin" class="login-form">

        <!-- Error Message Banner -->
        <div v-if="userLoginError" class="error-banner">
          {{ userLoginError }}
        </div>

        <div class="form-group">
          <label for="username">Username</label>
          <input type="text" id="username" v-model="username" placeholder="Enter your username" required
            :disabled="isUserLoginLoading" />
        </div>

        <div class="form-group">
          <label for="password">Password</label>
          <input type="password" id="password" v-model="password" placeholder="Enter your password" required
            :disabled="isUserLoginLoading" />
        </div>

        <button type="submit" class="submit-btn" :disabled="isUserLoginLoading">
          {{ isUserLoginLoading ? 'Authenticating...' : 'Sign In' }}
        </button>
      </form>

      <div class="admin-notice">
        Accounts are provisioned by the system administrator.
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { useMutation } from '../composables/useMutation'

const router = useRouter();
const baseUrl = import.meta.env.VITE_API_BASE_URL;

const username = ref('');
const password = ref('');
const errorMessage = ref('');
const isLoading = ref(false);
const {
  data: userLogin,
  res: userLoginRes,
  isLoading: isUserLoginLoading,
  error: userLoginError,
  execute: userLoginApi
} = useMutation();

const handleLogin = async () => {
  await userLoginApi(`${baseUrl}/user/login`, { username: username.value, password: password.value }, "POST")

  if (!userLoginRes.value.ok) {
    throw new Error(userLogin.value.message || 'Authentication failed');
  }

  // const data = await response.json();


  // Persist profile info for the UI layer
  localStorage.setItem('userId', userLogin.value.data.user.id);
  localStorage.setItem('userName', `${userLogin.value.data.user.firstName} ${userLogin.value.data.user.lastName}`);

  // Routinely move the user onto the secure canvas environment
  router.push('/dashboard');

};
</script>

<style scoped>
.login-wrapper {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  background-color: #f1f5f9;
}

.login-card {
  background: white;
  width: 100%;
  max-width: 400px;
  padding: 40px;
  border-radius: 12px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
}

.login-header {
  text-align: center;
  margin-bottom: 30px;
}

.login-header h2 {
  margin: 0;
  color: #1e293b;
  font-size: 1.75rem;
}

.login-header p {
  color: #64748b;
  margin-top: 8px;
  font-size: 0.95rem;
}

.error-banner {
  background-color: #fef2f2;
  color: #ef4444;
  padding: 12px;
  border-radius: 6px;
  margin-bottom: 20px;
  font-size: 0.9rem;
  text-align: center;
  border: 1px solid #fca5a5;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  color: #334155;
  font-weight: 500;
  font-size: 0.95rem;
}

.form-group input {
  width: 100%;
  padding: 12px;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  font-size: 1rem;
  background-color: #f8fafc;
  box-sizing: border-box;
  transition: border-color 0.2s;
}

.form-group input:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.form-group input:disabled {
  background-color: #e2e8f0;
  color: #94a3b8;
  cursor: not-allowed;
}

.submit-btn {
  width: 100%;
  padding: 12px;
  background-color: #3b82f6;
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: background-color 0.2s;
  margin-top: 10px;
}

.submit-btn:hover:not(:disabled) {
  background-color: #2563eb;
}

.submit-btn:disabled {
  background-color: #94a3b8;
  cursor: not-allowed;
}

.admin-notice {
  text-align: center;
  margin-top: 24px;
  font-size: 0.8rem;
  color: #94a3b8;
}
</style>