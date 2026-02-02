import { createRouter, createWebHistory } from 'vue-router'
import { useAuth } from '../composables/useAuth'
import DashboardsView from '../views/DashboardsView.vue'
import DashboardDetailView from '../views/DashboardDetailView.vue'
import Explore from '../views/Explore.vue'
import OrganizationSettings from '../views/OrganizationSettings.vue'
import LoginView from '../views/LoginView.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: LoginView,
      meta: { public: true }
    },
    {
      path: '/',
      redirect: '/dashboards'
    },
    {
      path: '/dashboards',
      name: 'dashboards',
      component: DashboardsView
    },
    {
      path: '/dashboards/:id',
      name: 'dashboard-detail',
      component: DashboardDetailView
    },
    {
      path: '/explore',
      name: 'explore',
      component: Explore
    },
    {
      path: '/settings/org/:id',
      name: 'org-settings',
      component: OrganizationSettings
    }
  ]
})

// Navigation guard for authentication
router.beforeEach(async (to, _from, next) => {
  const { isAuthenticated, initialized, initialize } = useAuth()

  // Initialize auth state if not already done
  if (!initialized.value) {
    await initialize()
  }

  // Allow access to public routes
  if (to.meta.public) {
    // If authenticated and going to login, redirect to dashboards
    if (isAuthenticated.value && to.name === 'login') {
      next('/dashboards')
      return
    }
    next()
    return
  }

  // Protected routes require authentication
  if (!isAuthenticated.value) {
    next({ name: 'login', query: { redirect: to.fullPath } })
    return
  }

  next()
})

export default router
