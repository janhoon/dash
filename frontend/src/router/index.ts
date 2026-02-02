import { createRouter, createWebHistory } from 'vue-router'
import DashboardsView from '../views/DashboardsView.vue'

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
    }
  ]
})

export default router
