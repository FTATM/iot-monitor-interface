<template>
  <div class="flex flex-col h-screen overflow-hidden bg-base-200">

    <!-- Header (Always visible, always clickable) -->
    <!-- Removed drawer-button from here, it's just a standard label now -->
    <header class="navbar bg-neutral text-neutral-content sticky top-0 z-[100] px-4 shadow-md h-[60px]">
      <div class="flex-1 gap-2">
        <!-- Hamburger Button -->
        <label for="main-drawer" class="btn btn-square btn-ghost cursor-pointer text-white hover:bg-white/20">
          <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24"
            class="inline-block w-6 h-6 stroke-current">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16"></path>
          </svg>
        </label>
        <router-link to="/" class="btn btn-ghost text-2xl font-bold normal-case text-white hover:bg-white/20">
          DashBoard
        </router-link>
      </div>

      <!-- User Profile and Logout -->
      <div class="flex-none flex items-center gap-4">
        <div class="text-lg font-medium flex items-center gap-2">
          <span>👤</span>
          <span>{{ userStore.user.fullName || "User" }}</span>
        </div>
        <button @click="handleLogout" class="btn btn-outline btn-error btn-base transition-transform hover:scale-105">
          Log Out
        </button>
      </div>
    </header>

    <!-- Container for Drawer (Locks the sidebar to this area only) -->
    <!-- Added 'relative' so the absolute sidebar stays contained here -->
    <div class="drawer flex-1 relative overflow-hidden">

      <input id="main-drawer" type="checkbox" class="drawer-toggle" v-model="isSidebarOpen" />

      <main class="drawer-content flex flex-col h-full bg-base-200 p-6 overflow-y-auto w-full">
        <!-- Dynamic Content Area -->
        <slot />
      </main>

      <!-- Sidebar Navigation -->
      <!-- FIXED: Added 'absolute' to override DaisyUI's fixed fullscreen layout -->
      <aside class="drawer-side absolute z-50 h-full">
        <label for="main-drawer" aria-label="close sidebar" class="drawer-overlay"></label>

        <!-- DaisyUI Menu -->
        <ul
          class="menu p-4 w-[250px] min-h-full bg-base-100 text-base-content border-r border-base-200 text-lg font-medium">
          <li>
            <router-link :to="{ name: 'dashboard' }" active-class="!text-orange-600 !bg-orange-100 font-bold"
              @click="isSidebarOpen = false">
              Dashboard
            </router-link>
          </li>
          <li v-if="hasPermission('Canvas Design', 'Display') || hasPermission('Canvas Access', 'Display') || hasPermission('Canvas', 'Display')">
            <details>
              <summary>Canvas Management</summary>
              <ul>
                <li v-if="hasPermission('Canvas', 'Display')">
                  <router-link :to="{ name: 'canvas' }" active-class="!text-orange-600 !bg-orange-100 font-bold"
                    @click="isSidebarOpen = false">
                    Canvas
                  </router-link>
                </li>
                <li v-if="hasPermission('Canvas Design', 'Display')">
                  <router-link :to="{ name: 'canvasDesign' }" active-class="!text-orange-600 !bg-orange-100 font-bold"
                    @click="isSidebarOpen = false">
                    Design
                  </router-link>
                </li>
                <li v-if="hasPermission('Canvas Access', 'Display')">
                  <router-link :to="{ name: 'canvasAccess' }" active-class="!text-orange-600 !bg-orange-100 font-bold"
                    @click="isSidebarOpen = false">
                    Access
                  </router-link>
                </li>
              </ul>
            </details>
          </li>
          <li>
            <router-link :to="{ name: 'about' }" active-class="!text-orange-600 !bg-orange-100 font-bold"
              @click="isSidebarOpen = false">
              About
            </router-link>
          </li>
          <li v-if="hasPermission('Scheduler', 'Display')">
            <router-link :to="{ name: 'scheduler' }" active-class="!text-orange-600 !bg-orange-100 font-bold"
              @click="isSidebarOpen = false">
              Scheduler
            </router-link>
          </li>

          <!-- Setting Parent -->
          <li
            v-if="hasPermission('User', 'Display') || hasPermission('Role', 'Display') || hasPermission('Device', 'Display')">
            <details>
              <summary>Management</summary>
              <ul>
                <li v-if="hasPermission('User', 'Display')">
                  <router-link :to="{ name: 'user' }" active-class="!text-orange-600 !bg-orange-100 font-bold"
                    @click="isSidebarOpen = false">
                    User
                  </router-link>
                </li>
                <li v-if="hasPermission('Role', 'Display')">
                  <router-link :to="{ name: 'role' }" active-class="!text-orange-600 !bg-orange-100 font-bold"
                    @click="isSidebarOpen = false">
                    Role
                  </router-link>
                </li>
                <li v-if="hasPermission('Device', 'Display')">
                  <router-link :to="{ name: 'device' }" active-class="!text-orange-600 !bg-orange-100 font-bold"
                    @click="isSidebarOpen = false">
                    Device
                  </router-link>
                </li>
              </ul>
            </details>
          </li>
        </ul>
      </aside>
    </div>

  </div>
</template>

<script setup>
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { useMutation } from '@/composables/useMutation';
import { usePermissionStore } from '@/stores/usePermissionStore';
import { useUserStore } from '@/stores/useUserStore';

// --- STATE ---
const isSidebarOpen = ref(false);

const router = useRouter();

// --- STORE ---
const permissionStore = usePermissionStore();
const { hasPermission,setPermissions } = permissionStore;
const userStore = useUserStore();
const { setUser } = userStore;

const { res: logoutRes, execute: logoutApi } = useMutation();

const handleLogout = async () => {
  try {
    // 1. Attempt to notify the server
    await logoutApi('/user/logout', null, 'POST');

    if (!logoutRes.value.ok) {
      console.warn("Server-side logout failed (Response not OK), but proceeding with local logout.");
    }
  } catch (error) {
    // Catch actual network errors (e.g., server is shut down, no internet)
    console.error("Network error during logout:", error);
  } finally {
    // 2. ALWAYS clear local state and redirect, regardless of server status
    setUser(null);
    setPermissions(null);
    router.push('/login');
  }
};
</script>