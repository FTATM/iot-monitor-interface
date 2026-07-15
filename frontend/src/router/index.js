import { createRouter, createWebHistory } from 'vue-router';
import LoginView from '../views/LoginView.vue';
import DashboardView from '../views/DashboardView.vue';
import AboutView from '../views/AboutView.vue';

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/:pathMatch(.*)*', redirect: '/dashboard' },
    { path: '/login', component: LoginView },
    { 
      path: '/dashboard', 
      component: DashboardView,
      meta: { requiresAuth: true } // Mark this route as protected
    },
    {
      path: '/about',
      name: 'about',
      component: AboutView
    },
    // redirect any unknown routes to dashboard (or login)
  ]
});

// GLOBAL ROUTE GUARD
router.beforeEach((to, from, next) => {
  // Check if the route needs authentication
  if (to.meta.requiresAuth) {
    
    // We cannot read the HttpOnly cookie, so we check for the user profile data instead
    const isAuthenticated = !!localStorage.getItem('userId');
    
    if (!isAuthenticated) {
      // No user data found, kick them back to login
      next('/login');
    } else {
      // UI data exists, assume they have a valid cookie and let them proceed
      next();
    }
  } else {
    // Route doesn't require auth, let them pass
    next();
  }
});

export default router;