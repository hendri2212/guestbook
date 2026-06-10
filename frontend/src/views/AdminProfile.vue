<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import AdminLayout from '@/components/AdminLayout.vue'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'
const AUTH_STORAGE_KEY = 'guestbook_admin_auth'

const router = useRouter()
const auth = ref(readStoredAuth())
const isLoading = ref(false)
const isSaving = ref(false)
const loadError = ref('')
const successMessage = ref('')

const profileForm = reactive({
  name: '',
  email: '',
  password: '',
})

const user = computed(() => auth.value?.user || null)
const company = computed(() => auth.value?.company || null)
const canEditProfile = computed(() => Boolean(user.value?.id))

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

function authHeaders(includeContentType = false) {
  const headers = {
    Authorization: `Bearer ${auth.value?.token}`,
  }

  if (includeContentType) {
    headers['Content-Type'] = 'application/json'
  }

  return headers
}

function fillProfileForm(userData) {
  Object.assign(profileForm, {
    name: userData?.name || '',
    email: userData?.email || '',
    password: '',
  })
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
    fillProfileForm(data.user)
  } catch (error) {
    loadError.value = error.message || 'Gagal memuat profile.'
  } finally {
    isLoading.value = false
  }
}

async function saveProfile() {
  if (!user.value?.id) {
    loadError.value = 'Data user belum tersedia.'
    return
  }

  isSaving.value = true
  loadError.value = ''
  successMessage.value = ''

  try {
    const payload = {
      name: profileForm.name.trim(),
      email: profileForm.email.trim(),
    }

    if (profileForm.password.trim() !== '') {
      payload.password = profileForm.password.trim()
    }

    const response = await fetch(`${API_BASE_URL}/api/admin/users/${user.value.id}`, {
      method: 'PUT',
      headers: authHeaders(true),
      body: JSON.stringify(payload),
    })
    const data = await response.json().catch(() => ({}))

    if (response.status === 401) {
      redirectToLogin()
      return
    }
    if (!response.ok) {
      throw new Error(data.error || data.message || 'Gagal menyimpan profile.')
    }

    auth.value = {
      ...auth.value,
      user: {
        id: data.id,
        name: data.name,
        email: data.email,
        role: data.role,
      },
    }
    localStorage.setItem(AUTH_STORAGE_KEY, JSON.stringify(auth.value))
    fillProfileForm(data)
    successMessage.value = 'Profile berhasil diperbarui.'
  } catch (error) {
    loadError.value = error.message || 'Gagal menyimpan profile.'
  } finally {
    isSaving.value = false
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
    <div v-if="successMessage" class="alert alert-success" role="alert">{{ successMessage }}</div>

    <div class="row g-4">
      <div class="col-12">
        <form class="content-panel" @submit.prevent="saveProfile">
          <div class="profile-heading">
            <span class="profile-icon"><i class="bi bi-person-badge"></i></span>
            <div>
              <h3 class="h5 mb-1">Akun admin</h3>
              <p class="text-secondary mb-0">
                {{ canEditProfile ? 'Perbarui data profile pengguna.' : 'Anda tidak memiliki akses untuk mengubah profile.' }}
              </p>
            </div>
          </div>

          <div class="row g-3">
            <div class="col-12 col-md-6">
              <label for="profileName" class="form-label">Nama</label>
              <input id="profileName" v-model="profileForm.name" type="text" class="form-control"
                :disabled="!canEditProfile" required />
            </div>

            <div class="col-12 col-md-6">
              <label for="profileEmail" class="form-label">Email</label>
              <input id="profileEmail" v-model="profileForm.email" type="email" class="form-control"
                :disabled="!canEditProfile" required />
            </div>

            <div class="col-12 col-md-6">
              <label for="profilePassword" class="form-label">Password Baru</label>
              <input id="profilePassword" v-model="profileForm.password" type="password" class="form-control"
                :disabled="!canEditProfile" autocomplete="new-password" placeholder="Kosongkan jika tidak diganti" />
            </div>

            <div class="col-12 col-md-3">
              <label class="form-label">Role</label>
              <div class="readonly-field">{{ user?.role || '-' }}</div>
            </div>

            <div class="col-12 col-md-3">
              <label class="form-label">Instansi</label>
              <div class="readonly-field">{{ company?.name || '-' }}</div>
            </div>
          </div>

          <div class="d-flex justify-content-end mt-4">
            <button type="submit" class="btn btn-primary px-3" :disabled="!canEditProfile || isSaving">
              <span v-if="isSaving" class="spinner-border spinner-border-sm me-2" aria-hidden="true"></span>
              <i v-else class="bi bi-save me-2"></i>
              {{ isSaving ? 'Menyimpan...' : 'Simpan Profile' }}
            </button>
          </div>
        </form>
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

.form-label {
  color: #334155;
  font-weight: 650;
}

.form-control {
  border-color: #dbe3ef;
  border-radius: 12px;
  padding: 0.76rem 0.9rem;
}

.readonly-field {
  border: 1px solid #dbe3ef;
  border-radius: 12px;
  background: #f8fafc;
  color: #334155;
  font-weight: 650;
  min-height: 50px;
  overflow-wrap: anywhere;
  padding: 0.76rem 0.9rem;
}

.form-control:focus {
  border-color: #86b7fe;
  box-shadow: 0 0 0 0.25rem rgba(13, 110, 253, 0.12);
}
</style>
