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
      path: '/admin/guest-visits',
      name: 'admin-guest-visits',
      component: () => import('@/views/AdminGuestVisits.vue'),
    },
    {
      path: '/admin/guest-visits/:formId',
      name: 'admin-guest-visit-detail',
      component: () => import('@/views/AdminGuestVisitDetail.vue'),
    },
    {
      path: '/admin/guest-forms',
      name: 'admin-guest-forms',
      component: () => import('@/views/AdminGuestForms.vue'),
      meta: {
        requiredRole: ['owner', 'admin'],
      },
    },
    {
      path: '/admin/users',
      name: 'admin-users',
      component: () => import('@/views/AdminUsers.vue'),
      meta: {
        requiredRole: 'owner',
      },
    },
    {
      path: '/admin/companies',
      name: 'admin-companies',
      component: () => import('@/views/AdminCompanies.vue'),
      meta: {
        requiredRole: 'owner',
      },
    },
    {
      path: '/admin/settings',
      name: 'admin-settings',
      component: () => import('@/views/AdminSettings.vue'),
    },
    {
      path: '/admin/profile',
      name: 'admin-profile',
      component: () => import('@/views/AdminProfile.vue'),
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
  const isAdminRoute = to.path.startsWith('/admin') && to.name !== 'admin-login'
  const auth = readStoredAuth()

  if (isAdminRoute && !auth?.token) {
    return { name: 'admin-login' }
  }

  if (!to.meta.requiredRole) {
    return true
  }

  const requiredRoles = Array.isArray(to.meta.requiredRole) ? to.meta.requiredRole : [to.meta.requiredRole]
  if (!requiredRoles.includes(auth.user?.role)) {
    return { name: 'admin-dashboard' }
  }

  return true
})

export default router
