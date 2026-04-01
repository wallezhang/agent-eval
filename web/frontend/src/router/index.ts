import { createRouter, createWebHistory } from 'vue-router'
import AppLayout from '@/components/AppLayout.vue'

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      component: AppLayout,
      children: [
        { path: '', redirect: '/dashboard' },
        {
          path: 'dashboard',
          name: 'dashboard',
          component: () => import('@/views/DashboardView.vue'),
        },
        {
          path: 'configs',
          name: 'configs',
          component: () => import('@/views/ConfigListView.vue'),
        },
        {
          path: 'configs/:filename',
          name: 'config-edit',
          component: () => import('@/views/ConfigEditView.vue'),
          props: true,
        },
        {
          path: 'runs',
          name: 'runs',
          component: () => import('@/views/RunListView.vue'),
        },
        {
          path: 'runs/:id',
          name: 'run-detail',
          component: () => import('@/views/RunDetailView.vue'),
          props: true,
        },
        {
          path: 'results/:id',
          name: 'result-detail',
          component: () => import('@/views/ResultDetailView.vue'),
          props: true,
        },
        {
          path: 'settings',
          name: 'settings',
          component: () => import('@/views/SettingsView.vue'),
        },
      ],
    },
  ],
})
