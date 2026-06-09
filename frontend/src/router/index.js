import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/admin/login',
      name: 'admin-login',
      component: () => import('@/views/AdminLogin.vue'),
    },
    {
      path: '/admin',
      name: 'admin-dashboard',
      component: () => import('@/views/AdminDashboard.vue'),
    },
    {
      path: '/forms/:publicSlug',
      name: 'public-form-prefixed',
      component: () => import('@/views/PublicGuestForm.vue'),
    },
    {
      path: '/:publicSlug?',
      name: 'public-form',
      component: () => import('@/views/PublicGuestForm.vue'),
    },
  ],
})

export default router
