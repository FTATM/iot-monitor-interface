<template>
  <div class="flex flex-col items-center justify-center h-full w-full p-8 text-center bg-base-200">
    
    <!-- 1. LOADING STATE: While pinging the server -->
    <template v-if="isChecking">
      <span class="loading loading-spinner loading-lg text-primary"></span>
      <p class="mt-4 text-base-content/70 font-medium">Verifying connection...</p>
    </template>

    <!-- 2. SERVER DOWN STATE: Network error or 500 status -->
    <template v-else-if="isServerDown">
      <div class="text-error mb-4 animate-bounce">
        <Icon icon="lucide:server-crash" class="w-24 h-24 mx-auto drop-shadow-sm" />
      </div>
      <h2 class="text-3xl font-extrabold text-base-content mb-3 tracking-tight">Connection Lost</h2>
      <p class="text-base-content/70 max-w-md mx-auto text-lg">
        We cannot reach the server right now. Please check your internet connection or try again later.
      </p>
      <button @click="checkConnection" class="btn btn-outline btn-error mt-6">
        <Icon icon="lucide:refresh-cw" class="w-4 h-4 mr-2" /> Try Again
      </button>
    </template>

    <!-- 3. NO ACCESS STATE: Server is alive, user just lacks permissions -->
    <template v-else>
      <div class="text-warning mb-4">
        <Icon icon="lucide:shield-alert" class="w-24 h-24 mx-auto drop-shadow-sm" />
      </div>
      <h2 class="text-3xl font-extrabold text-base-content mb-3 tracking-tight">Access Denied</h2>
      <p class="text-base-content/70 max-w-md mx-auto text-lg">
        You do not have permission to view this page. If you believe this is a mistake, contact your administrator.
      </p>
      <router-link :to="{ name: 'dashboard' }" class="btn btn-primary mt-6 text-white shadow-sm hover:scale-105 transition-transform">
        Return to Dashboard
      </router-link>
    </template>

  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import { Icon } from '@iconify/vue';

// --- STATE ---
const isChecking = ref(true);
const isServerDown = ref(false);

// --- METHODS ---
const checkConnection = async () => {
  isChecking.value = true;
  isServerDown.value = false;

  try {
    const response = await fetch(import.meta.env.VITE_API_BASE_URL + '/ping', { cache: 'no-store' });
    
    if (!response.ok && response.status >= 500) {
      // The server responded, but it is throwing a fatal error
      isServerDown.value = true;
    } else {
      // Even if the ping returns 401 Unauthorized or 403 Forbidden, 
      // it means the server is ALIVE. Therefore, it is a true "No Access" scenario.
      isServerDown.value = false;
    }
  } catch (error) {
    // A catch block here means the network is disconnected completely
    isServerDown.value = true;
  } finally {
    isChecking.value = false;
  }
};

onMounted(() => {
  // Fire the ping exactly when the No Access component tries to render
  checkConnection();
});
</script>