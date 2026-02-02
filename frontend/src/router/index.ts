import { createRouter, createWebHistory } from 'vue-router'
import DashboardsView from '../views/DashboardsView.vue'
import DashboardDetailView from '../views/DashboardDetailView.vue'
import Explore from '../views/Explore.vue'
import OrganizationSettings from '../views/OrganizationSettings.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
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

export default router
