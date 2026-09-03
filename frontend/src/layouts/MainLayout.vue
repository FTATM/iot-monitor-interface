<template>
  <div class="flex flex-col h-screen overflow-hidden bg-base-200">

    <!-- Header -->
    <header class="navbar bg-neutral text-neutral-content sticky top-0 z-[100] px-4 shadow-md h-[60px]">

      <div class="flex-1 flex items-center justify-start gap-2">
        <label for="main-drawer" class="btn btn-square btn-ghost cursor-pointer text-white hover:bg-white/20">
          <Icon icon="lucide:menu" class="w-6 h-6" />
        </label>

        <router-link to="/"
          class="btn btn-ghost text-2xl font-bold normal-case text-white hover:bg-white/20 flex items-center gap-3">
          <img src="/LOGO_FT.png" alt="Logo" class="w-8 h-8 rounded bg-white/10 p-1 shadow-sm" />
          <span>{{ $t('header.brand') }}</span>
        </router-link>
      </div>

      <div class="flex-none flex items-center gap-4">
        <div class="text-lg font-medium flex items-center gap-2">
          <Icon icon="lucide:user-circle" class="w-6 h-6 text-white/80" />
          <span>{{ userStore.user?.fullName || $t('header.defaultUser') }}</span>
        </div>
        <button @click="handleLogout" class="btn btn-outline btn-error btn-sm transition-transform hover:scale-105">
          <Icon icon="lucide:log-out" class="w-4 h-4 mr-1" /> {{ $t('header.logout') }}
        </button>
      </div>

    </header>

    <div class="drawer flex-1 relative overflow-hidden">
      <input id="main-drawer" type="checkbox" class="drawer-toggle" v-model="isSidebarOpen" />

      <main class="drawer-content flex flex-col h-full bg-base-200 overflow-y-auto w-full">
        <slot />
      </main>

      <!-- Sidebar Navigation -->
      <aside class="drawer-side absolute z-50 h-full">
        <label for="main-drawer" aria-label="close sidebar" class="drawer-overlay"></label>

        <div class="flex flex-col w-[280px] min-h-full bg-base-100 border-r border-base-200">

          <!-- Menu Items -->
          <ul class="menu p-4 text-base-content text-lg font-medium gap-2 flex-1">

            <li>
              <router-link :to="{ name: 'dashboard' }" active-class="!text-primary !bg-primary/10 font-bold"
                @click="isSidebarOpen = false" class="py-3 px-4 rounded-xl flex items-center gap-3">
                <Icon icon="lucide:layout-dashboard" class="w-5 h-5 opacity-70" />
                {{ $t('menu.dashboard') }}
              </router-link>
            </li>

            <li
              v-if="hasPermission('Canvas Design', 'Display') || hasPermission('Canvas Access', 'Display') || hasPermission('Canvas', 'Display')">
              <details>
                <summary class="py-3 px-4 rounded-xl flex items-center gap-3">
                  <Icon icon="lucide:layout-list" class="w-5 h-5 opacity-70" />
                  {{ $t('menu.canvas') }}
                </summary>
                <ul class="mt-2 gap-1 border-l-2 border-base-200 ml-4 pl-2">
                  <li v-if="hasPermission('Canvas', 'Display')">
                    <router-link :to="{ name: 'canvas' }" active-class="!text-primary !bg-primary/10 font-bold"
                      @click="isSidebarOpen = false" class="py-2.5 px-4 rounded-lg">
                      {{ $t('menu.canvas') }}
                    </router-link>
                  </li>
                  <li v-if="hasPermission('Canvas Design', 'Display')">
                    <router-link :to="{ name: 'canvasDesign' }" active-class="!text-primary !bg-primary/10 font-bold"
                      @click="isSidebarOpen = false" class="py-2.5 px-4 rounded-lg">
                      {{ $t('menu.canvasDesign') }}
                    </router-link>
                  </li>
                  <li v-if="hasPermission('Canvas Access', 'Display')">
                    <router-link :to="{ name: 'canvasAccess' }" active-class="!text-primary !bg-primary/10 font-bold"
                      @click="isSidebarOpen = false" class="py-2.5 px-4 rounded-lg">
                      {{ $t('menu.canvasAccess') }}
                    </router-link>
                  </li>
                </ul>
              </details>
            </li>

            <li v-if="hasPermission('Scheduler', 'Display')">
              <router-link :to="{ name: 'scheduler' }" active-class="!text-primary !bg-primary/10 font-bold"
                @click="isSidebarOpen = false" class="py-3 px-4 rounded-xl flex items-center gap-3">
                <Icon icon="lucide:calendar-clock" class="w-5 h-5 opacity-70" />
                {{ $t('menu.scheduler') }}
              </router-link>
            </li>

            <li v-if="hasPermission('Notification User', 'Display') || hasPermission('Notification Device', 'Display')">
              <details>
                <summary class="py-3 px-4 rounded-xl flex items-center gap-3">
                  <Icon icon="lucide:bell" class="w-5 h-5 opacity-70" />
                  {{ $t('menu.notification') }}
                </summary>
                <ul class="mt-2 gap-1 border-l-2 border-base-200 ml-4 pl-2">
                  <li v-if="hasPermission('Notification User', 'Display')">
                    <router-link :to="{ name: 'notifUser' }" active-class="!text-primary !bg-primary/10 font-bold"
                      @click="isSidebarOpen = false" class="py-2.5 px-4 rounded-lg">
                      {{ $t('menu.notifUser') }}
                    </router-link>
                  </li>
                  <li v-if="hasPermission('Notification Device', 'Display')">
                    <router-link :to="{ name: 'notifDevicerule' }" active-class="!text-primary !bg-primary/10 font-bold"
                      @click="isSidebarOpen = false" class="py-2.5 px-4 rounded-lg">
                      {{ $t('menu.notifDeviceRule') }}
                    </router-link>
                  </li>
                </ul>
              </details>
            </li>

            <li v-if="hasPermission('Log Report', 'Display')">
              <router-link :to="{ name: 'logReport' }" active-class="!text-primary !bg-primary/10 font-bold"
                @click="isSidebarOpen = false" class="py-3 px-4 rounded-xl flex items-center gap-3">
                <Icon icon="lucide:file-text" class="w-5 h-5 opacity-70" />
                {{ $t('menu.logReport') }}
              </router-link>
            </li>

            <li
              v-if="hasPermission('User', 'Display') || hasPermission('Role', 'Display') || hasPermission('Device', 'Display') || hasPermission('Device Group', 'Display')">
              <details>
                <summary class="py-3 px-4 rounded-xl flex items-center gap-3">
                  <Icon icon="lucide:settings" class="w-5 h-5 opacity-70" />
                  {{ $t('menu.management') }}
                </summary>
                <ul class="mt-2 gap-1 border-l-2 border-base-200 ml-4 pl-2">
                  <li v-if="hasPermission('User', 'Display')">
                    <router-link :to="{ name: 'user' }" active-class="!text-primary !bg-primary/10 font-bold"
                      @click="isSidebarOpen = false" class="py-2.5 px-4 rounded-lg">
                      {{ $t('menu.user') }}
                    </router-link>
                  </li>
                  <li v-if="hasPermission('Role', 'Display')">
                    <router-link :to="{ name: 'role' }" active-class="!text-primary !bg-primary/10 font-bold"
                      @click="isSidebarOpen = false" class="py-2.5 px-4 rounded-lg">
                      {{ $t('menu.role') }}
                    </router-link>
                  </li>
                  <li v-if="hasPermission('Device', 'Display')">
                    <router-link :to="{ name: 'device' }" active-class="!text-primary !bg-primary/10 font-bold"
                      @click="isSidebarOpen = false" class="py-2.5 px-4 rounded-lg">
                      {{ $t('menu.device') }}
                    </router-link>
                  </li>
                  <li v-if="hasPermission('Device Group', 'Display')">
                    <router-link :to="{ name: 'deviceGroup' }" active-class="!text-primary !bg-primary/10 font-bold"
                      @click="isSidebarOpen = false" class="py-2.5 px-4 rounded-lg">
                      {{ $t('menu.deviceGroup') }}
                    </router-link>
                  </li>
                </ul>
              </details>
            </li>
          </ul>

          <!-- Sidebar Footer (Theme & Language Toggles) -->
          <div class="p-4 border-t border-base-200 bg-base-100/50 flex flex-col gap-4">

            <!-- Language Toggle -->
            <div class="flex items-center justify-between px-2">
              <span class="text-sm font-semibold text-base-content/70 flex items-center gap-2">
                <Icon icon="lucide:globe" class="w-4 h-4" /> {{ $t('settings.language') }}
              </span>
              <div class="join bg-base-200 border border-base-300">
                <button type="button" class="btn btn-sm join-item border-none"
                  :class="locale === 'en' ? 'btn-primary' : 'btn-ghost'" @click="setLanguage('en')">
                  EN
                </button>
                <button type="button" class="btn btn-sm join-item border-none"
                  :class="locale === 'th' ? 'btn-primary' : 'btn-ghost'" @click="setLanguage('th')">
                  TH
                </button>
              </div>
            </div>

            <!-- Theme Toggle -->
            <div class="flex items-center justify-between px-2">
              <span class="text-sm font-semibold text-base-content/70 flex items-center gap-2">
                <Icon :icon="themeStore.isDarkTheme ? 'lucide:moon' : 'lucide:sun'" class="w-4 h-4" /> {{ $t('settings.theme') }}
              </span>
              <label class="swap swap-rotate btn btn-sm btn-circle btn-ghost bg-base-200 border border-base-300">
                <!-- อัปเดตให้ผูกกับ Store โดยตรง -->
                <input type="checkbox" v-model="themeStore.isDarkTheme" />
                <Icon icon="lucide:sun" class="swap-off w-4 h-4 text-warning" />
                <Icon icon="lucide:moon" class="swap-on w-4 h-4 text-info" />
              </label>
            </div>

          </div>

        </div>
      </aside>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { Icon } from '@iconify/vue';
import { useMutation } from '@/composables/useMutation';
import { usePermissionStore } from '@/stores/usePermissionStore';
import { useUserStore } from '@/stores/useUserStore';
import { useThemeStore } from '@/stores/useThemeStore';

// --- ROUTER & I18N ---
const router = useRouter();
const { locale } = useI18n();

// --- STATE ---
const isSidebarOpen = ref(false);

// --- STORES ---
const permissionStore = usePermissionStore();
const { hasPermission, setPermissions } = permissionStore;
const userStore = useUserStore();
const { setUser } = userStore;
const themeStore = useThemeStore();

const { res: logoutRes, execute: logoutApi } = useMutation();

// --- INITIALIZATION ---
onMounted(() => {

  // Load Language
  const savedLang = localStorage.getItem('lang') || 'en';
  locale.value = savedLang;
});

// --- METHODS ---

const setLanguage = (lang) => {
  locale.value = lang;
  localStorage.setItem('lang', lang);
};

const handleLogout = async () => {
  try {
    await logoutApi('/user/logout', null, 'POST');
    if (!logoutRes.value.ok) {
      console.warn("Server-side logout failed (Response not OK), but proceeding with local logout.");
    }
  } catch (error) {
    console.error("Network error during logout:", error);
  } finally {
    setUser(null);
    setPermissions(null);
    router.push('/login');
  }
};
</script>