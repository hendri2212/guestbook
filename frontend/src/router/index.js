import { createRouter, createWebHistory } from 'vue-router'

const AUTH_STORAGE_KEY = 'guestbook_admin_auth'

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
      path: '/admin/guest-forms',
      name: 'admin-guest-forms',
      component: () => import('@/views/AdminGuestForms.vue'),
      meta: {
        requiredRole: 'admin',
      },
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

function readStoredAuth() {
  try {
    const raw = localStorage.getItem(AUTH_STORAGE_KEY)
    return raw ? JSON.parse(raw) : null
  } catch {
    return null
  }
}

router.beforeEach((to) => {
  if (!to.meta.requiredRole) {
    return true
  }

  const auth = readStoredAuth()
  if (!auth?.token) {
    return { name: 'admin-login' }
  }

  if (auth.user?.role !== to.meta.requiredRole) {
    return { name: 'admin-dashboard' }
  }

  return true
})

export default router
