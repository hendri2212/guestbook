<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import AdminLayout from '@/components/AdminLayout.vue'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'
const AUTH_STORAGE_KEY = 'guestbook_admin_auth'

const router = useRouter()
const auth = ref(readStoredAuth())
const isLoading = ref(false)
const loadError = ref('')

const user = computed(() => auth.value?.user || null)
const company = computed(() => auth.value?.company || null)

function readStoredAuth() {
  try {
    const raw = localStorage.getItem(AUTH_STORAGE_KEY)
    return raw ? JSON.parse(raw) : null
  } catch {
    return null
  }
}

function clearSession() {
  localStorage.removeItem(AUTH_STORAGE_KEY)
  auth.value = null
}

function logout() {
  clearSession()
  router.push({ name: 'admin-login' })
}

function redirectToLogin() {
  clearSession()
  router.replace({ name: 'admin-login' })
}

function authHeaders() {
  return {
    Authorization: `Bearer ${auth.value?.token}`,
  }
}

async function loadProfile() {
  if (!auth.value?.token) {
    router.replace({ name: 'admin-login' })
    return
  }

  isLoading.value = true
  loadError.value = ''

  try {
    const response = await fetch(`${API_BASE_URL}/api/admin/me`, {
      headers: authHeaders(),
    })
    const data = await response.json().catch(() => ({}))

    if (response.status === 401) {
      redirectToLogin()
      return
    }
    if (!response.ok) {
      throw new Error(data.error || data.message || 'Gagal memuat profile.')
    }

    auth.value = {
      ...auth.value,
      user: data.user,
      company: data.company,
    }
    localStorage.setItem(AUTH_STORAGE_KEY, JSON.stringify(auth.value))
  } catch (error) {
    loadError.value = error.message || 'Gagal memuat profile.'
  } finally {
    isLoading.value = false
  }
}

onMounted(loadProfile)
</script>

<template>
  <AdminLayout :user="user" :company="company" active-menu="profile" @logout="logout">
    <template #header>
      <div>
        <p class="text-uppercase small fw-semibold text-primary mb-2">Profile</p>
        <h2 class="h3 mb-1">Profile pengguna</h2>
        <p class="text-secondary mb-0">Informasi akun admin yang sedang login.</p>
      </div>
    </template>

    <div v-if="isLoading" class="alert alert-light border" role="status">Memuat profile...</div>
    <div v-if="loadError" class="alert alert-warning" role="alert">{{ loadError }}</div>

    <div class="row g-4">
      <div class="col-12">
        <section class="content-panel">
          <div class="profile-heading">
            <span class="profile-icon"><i class="bi bi-person-badge"></i></span>
            <div>
              <h3 class="h5 mb-1">Akun admin</h3>
              <p class="text-secondary mb-0">Data pengguna yang sedang login.</p>
            </div>
          </div>

          <dl class="profile-list mb-0">
            <div>
              <dt>Nama</dt>
              <dd>{{ user?.name || '-' }}</dd>
            </div>
            <div>
              <dt>Email</dt>
              <dd>{{ user?.email || '-' }}</dd>
            </div>
            <div>
              <dt>Role</dt>
              <dd>{{ user?.role || '-' }}</dd>
            </div>
            <div>
              <dt>Instansi</dt>
              <dd>{{ company?.name || '-' }}</dd>
            </div>
          </dl>
        </section>
      </div>
    </div>
  </AdminLayout>
</template>

<style scoped>
.profile-heading {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 20px;
}

.profile-icon {
  display: grid;
  width: 46px;
  height: 46px;
  flex: 0 0 auto;
  place-items: center;
  border-radius: 14px;
  background: #0d6efd;
  color: #ffffff;
  font-weight: 800;
}

.profile-list {
  display: grid;
  gap: 12px;
}

.profile-list div {
  border-top: 1px solid #e2e8f0;
  padding-top: 12px;
}

.profile-list dt {
  color: #64748b;
  font-size: 0.78rem;
  text-transform: uppercase;
}

.profile-list dd {
  margin-bottom: 0;
  color: #111827;
  font-weight: 650;
  overflow-wrap: anywhere;
}
</style>
