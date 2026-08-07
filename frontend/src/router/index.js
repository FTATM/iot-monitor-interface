import { createRouter, createWebHistory } from 'vue-router';
import LoginView from '@/views/LoginView.vue';
import DashboardView from '@/views/DashboardView.vue';
import AboutView from '@/views/AboutView.vue';
import UserView from '@/views/management/UserView.vue';
import DeviceView from '@/views/management/DeviceView.vue';
import RoleView from '@/views/management/RoleView.vue';
import SchedulerView from '@/views/SchedulerView.vue';
import CanvasDesignView from '@/views/canvasManagement/CanvasDesignView.vue';
import CanvasAccessView from '@/views/canvasManagement/CanvasAccessView.vue';
import Test from '@/views/Test.vue';
import { useFetch } from '@/composables/useFetch';
import { toast } from 'vue3-toastify';
import { usePermissionStore } from '@/stores/usePermissionStore';
import { useUserStore } from '@/stores/useUserStore';

const { data: userPermissionData, error: userPermissionError, execute: userPermissionApi } = useFetch();


const router = createRouter({
  history: createWebHistory(),
  routes: [
    // redirect any unknown routes to dashboard (or login)
    { path: '/:pathMatch(.*)*', redirect: '/dashboard' },
    {
      path: '/login',
      component: LoginView,
      name: 'login',
      meta: { hideLayout: true }
    },
    {
      path: '/dashboard',
      component: DashboardView,
      name: 'dashboard',
      meta: { requiresAuth: true }
    },
    {
      path: '/about',
      name: 'about',
      component: AboutView
    },
    {
      path: '/canvasManagement',
      name: 'canvasManagement',
      meta: { requiresAuth: true },
      children: [
        {
          path: 'design',
          name: 'canvasDesign',
          component: CanvasDesignView,
        },
        {
          path: 'access',
          name: 'canvasAccess',
          component: CanvasAccessView,
        },
      ]
    },
    {
      path: '/management',
      name: 'management',
      meta: { requiresAuth: true },
      children: [
        {
          path: 'user',
          name: 'user',
          component: UserView,
        },
        {
          path: 'role',
          name: 'role',
          component: RoleView,
        },
        {
          path: 'device',
          name: 'device',
          component: DeviceView,
        },
      ]
    },
    {
      path: '/scheduler',
      name: 'scheduler',
      component: SchedulerView,
      meta: { requiresAuth: true }
    },
    {
      path: '/Test',
      name: 'Test',
      component: Test
    },
  ]
});

// GLOBAL ROUTE GUARD
router.beforeEach(async (to, from) => {
  // Check if the route needs authentication
  if (to.meta.requiresAuth) {
    const userStore = useUserStore();

    if (!userStore.user.id) {
      // No user data found, kick them back to login
      return { name: 'login' };
    } else {
      const permissionStore = usePermissionStore();

      await userPermissionApi('/user/permission');

      if (!userPermissionError.value && userPermissionData.value) {
        // Pass the permissions data from the Go API into the composable
        permissionStore.setPermissions(userPermissionData.value.data);
      } else {
        toast.error(userPermissionError.value?.message || "Failed to load permissions");
      }
      return true;
    }
  } else {
    // Route doesn't require auth, let them pass
    return true;
  }
});

export default router;