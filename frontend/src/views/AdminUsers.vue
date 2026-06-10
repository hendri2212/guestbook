<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import AdminLayout from '@/components/AdminLayout.vue'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'
const AUTH_STORAGE_KEY = 'guestbook_admin_auth'
const roleOptions = [
  { value: 'owner', label: 'Owner' },
  { value: 'admin', label: 'Admin' },
  { value: 'staff', label: 'Staff' },
]

const router = useRouter()
const auth = ref(readStoredAuth())
const users = ref([])
const isLoading = ref(false)
const isSaving = ref(false)
const isDeleting = ref('')
const pageError = ref('')
const successMessage = ref('')
const editingID = ref(null)

const form = reactive({
  company_id: '',
  name: '',
  email: '',
  password: '',
  role: 'staff',
  is_active: true,
})

const filters = reactive({
  q: '',
  role: '',
  is_active: '',
})

const user = computed(() => auth.value?.user || null)
const company = computed(() => auth.value?.company || null)
const isEditing = computed(() => Boolean(editingID.value))
const endpoint = computed(() => `${API_BASE_URL}/api/admin/users`)
const companyOptions = computed(() => {
  const map = new Map()

  if (company.value?.id) {
    map.set(company.value.id, company.value)
  }

  users.value.forEach((adminUser) => {
    if (adminUser.company?.id) {
      map.set(adminUser.company.id, adminUser.company)
    }
  })

  return [...map.values()]
})

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

function isAuthError(message) {
  const normalized = message.toLowerCase()
  return normalized.includes('unauthorized') || normalized.includes('invalid') || normalized.includes('expired')
}

function resetForm() {
  editingID.value = null
  Object.assign(form, {
    company_id: company.value?.id || '',
    name: '',
    email: '',
    password: '',
    role: 'staff',
    is_active: true,
  })
}

function editUser(adminUser) {
  editingID.value = adminUser.id
  Object.assign(form, {
    company_id: adminUser.company_id || adminUser.company?.id || company.value?.id || '',
    name: adminUser.name || '',
    email: adminUser.email || '',
    password: '',
    role: adminUser.role || 'staff',
    is_active: Boolean(adminUser.is_active),
  })
  pageError.value = ''
  successMessage.value = ''
}

function buildQuery() {
  const params = new URLSearchParams()
  if (filters.q.trim() !== '') {
    params.set('q', filters.q.trim())
  }
  if (filters.role !== '') {
    params.set('role', filters.role)
  }
  if (filters.is_active !== '') {
    params.set('is_active', filters.is_active)
  }

  const query = params.toString()
  return query ? `?${query}` : ''
}

function formatDateTime(value) {
  if (!value) {
    return '-'
  }

  return new Intl.DateTimeFormat('id-ID', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(value))
}

async function loadProfile() {
  const response = await fetch(`${API_BASE_URL}/api/admin/me`, {
    headers: authHeaders(),
  })
  const data = await response.json().catch(() => ({}))

  if (response.status === 401) {
    redirectToLogin()
    return false
  }
  if (!response.ok) {
    throw new Error(data.error || data.message || 'Sesi tidak valid.')
  }

  auth.value = {
    ...auth.value,
    user: data.user,
    company: data.company,
  }
  localStorage.setItem(AUTH_STORAGE_KEY, JSON.stringify(auth.value))

  if (data.user?.role !== 'owner') {
    router.replace({ name: 'admin-dashboard' })
    return false
  }

  if (!form.company_id) {
    form.company_id = data.company?.id || ''
  }

  return true
}

async function loadUsers() {
  if (!auth.value?.token) {
    router.replace({ name: 'admin-login' })
    return
  }

  isLoading.value = true
  pageError.value = ''

  try {
    const isAllowed = await loadProfile()
    if (!isAllowed) {
      return
    }

    const response = await fetch(`${endpoint.value}${buildQuery()}`, {
      headers: authHeaders(),
    })
    const data = await response.json().catch(() => ({}))

    if (response.status === 401) {
      redirectToLogin()
      return
    }
    if (!response.ok) {
      throw new Error(data.error || data.message || 'Gagal memuat user.')
    }

    users.value = data
  } catch (error) {
    pageError.value = error.message || 'Gagal memuat user.'
    if (isAuthError(pageError.value)) {
      redirectToLogin()
    }
  } finally {
    isLoading.value = false
  }
}

async function saveUser() {
  pageError.value = ''
  successMessage.value = ''
  isSaving.value = true

  try {
    const payload = {
      company_id: form.company_id,
      name: form.name.trim(),
      email: form.email.trim(),
      role: form.role,
      is_active: form.is_active,
    }

    if (!isEditing.value || form.password.trim() !== '') {
      payload.password = form.password.trim()
    }

    const response = await fetch(isEditing.value ? `${endpoint.value}/${editingID.value}` : endpoint.value, {
      method: isEditing.value ? 'PUT' : 'POST',
      headers: authHeaders(true),
      body: JSON.stringify(payload),
    })
    const data = await response.json().catch(() => ({}))

    if (response.status === 401) {
      redirectToLogin()
      return
    }
    if (!response.ok) {
      throw new Error(data.error || data.message || 'User gagal disimpan.')
    }

    successMessage.value = isEditing.value ? 'User berhasil diperbarui.' : 'User berhasil dibuat.'
    resetForm()
    await loadUsers()
  } catch (error) {
    pageError.value = error.message || 'User gagal disimpan.'
    if (isAuthError(pageError.value)) {
      redirectToLogin()
    }
  } finally {
    isSaving.value = false
  }
}

async function deleteUser(adminUser) {
  const confirmed = window.confirm(`Hapus user "${adminUser.name}"?`)
  if (!confirmed) {
    return
  }

  pageError.value = ''
  successMessage.value = ''
  isDeleting.value = adminUser.id

  try {
    const response = await fetch(`${endpoint.value}/${adminUser.id}`, {
      method: 'DELETE',
      headers: authHeaders(),
    })

    if (!response.ok) {
      const data = await response.json().catch(() => ({}))
      if (response.status === 401) {
        redirectToLogin()
        return
      }
      throw new Error(data.error || data.message || 'User gagal dihapus.')
    }

    successMessage.value = 'User berhasil dihapus.'
    if (editingID.value === adminUser.id) {
      resetForm()
    }
    await loadUsers()
  } catch (error) {
    pageError.value = error.message || 'User gagal dihapus.'
    if (isAuthError(pageError.value)) {
      redirectToLogin()
    }
  } finally {
    isDeleting.value = ''
  }
}

onMounted(loadUsers)
</script>

<template>
  <AdminLayout :user="user" :company="company" active-menu="users" @logout="logout">
    <template #header>
      <div>
        <p class="text-uppercase small fw-semibold text-primary mb-2">Management User</p>
        <h2 class="h3 mb-1">Kelola user admin</h2>
        <p class="text-secondary mb-0">Tambah, edit, dan nonaktifkan user yang dapat mengakses admin panel.</p>
      </div>
    </template>

    <div v-if="pageError" class="alert alert-danger" role="alert">{{ pageError }}</div>
    <div v-if="successMessage" class="alert alert-success" role="alert">{{ successMessage }}</div>

    <div class="row g-4">
      <div class="col-12 col-xl-7">
        <section class="content-panel">
          <div class="d-flex align-items-center justify-content-between gap-3 mb-3">
            <div>
              <h3 class="h5 mb-1">Daftar user</h3>
              <p class="text-secondary mb-0">User owner, admin, dan staff yang terdaftar.</p>
            </div>
            <button type="button" class="btn btn-outline-primary btn-sm" @click="loadUsers">
              <i class="bi bi-arrow-clockwise me-1"></i>
              Refresh
            </button>
          </div>

          <div class="filter-grid mb-3">
            <input v-model="filters.q" type="search" class="form-control" placeholder="Cari nama atau email" />
            <select v-model="filters.role" class="form-select">
              <option value="">Semua role</option>
              <option v-for="role in roleOptions" :key="role.value" :value="role.value">{{ role.label }}</option>
            </select>
            <select v-model="filters.is_active" class="form-select">
              <option value="">Semua status</option>
              <option value="true">Aktif</option>
              <option value="false">Nonaktif</option>
            </select>
            <button type="button" class="btn btn-primary" @click="loadUsers">
              <i class="bi bi-funnel me-1"></i>
              Terapkan
            </button>
          </div>

          <div v-if="isLoading" class="alert alert-light border mb-0" role="status">
            Memuat user...
          </div>

          <div v-else-if="users.length === 0" class="empty-state">
            <h4 class="h6 mb-2">Belum ada user</h4>
            <p class="text-secondary mb-0">Tambahkan user pertama menggunakan form di samping.</p>
          </div>

          <div v-else class="table-responsive">
            <table class="table align-middle mb-0">
              <thead>
                <tr>
                  <th>User</th>
                  <th>Instansi</th>
                  <th>Role</th>
                  <th>Status</th>
                  <th>Login Terakhir</th>
                  <th class="text-end">Aksi</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="adminUser in users" :key="adminUser.id">
                  <td>
                    <p class="fw-semibold mb-1">{{ adminUser.name }}</p>
                    <p class="small text-secondary mb-0">{{ adminUser.email }}</p>
                  </td>
                  <td>{{ adminUser.company?.name || '-' }}</td>
                  <td>{{ roleOptions.find((role) => role.value === adminUser.role)?.label || adminUser.role }}</td>
                  <td>
                    <span
                      class="badge rounded-pill"
                      :class="adminUser.is_active ? 'text-bg-success' : 'text-bg-secondary'"
                    >
                      {{ adminUser.is_active ? 'Aktif' : 'Nonaktif' }}
                    </span>
                  </td>
                  <td>{{ formatDateTime(adminUser.last_login_at) }}</td>
                  <td class="text-end">
                    <div class="d-inline-flex gap-2">
                      <button
                        type="button"
                        class="btn btn-outline-primary btn-sm icon-btn"
                        title="Edit user"
                        aria-label="Edit user"
                        @click="editUser(adminUser)"
                      >
                        <i class="bi bi-pencil-square"></i>
                      </button>
                      <button
                        type="button"
                        class="btn btn-outline-danger btn-sm icon-btn"
                        :disabled="isDeleting === adminUser.id || adminUser.id === user?.id"
                        title="Hapus user"
                        aria-label="Hapus user"
                        @click="deleteUser(adminUser)"
                      >
                        <span
                          v-if="isDeleting === adminUser.id"
                          class="spinner-border spinner-border-sm"
                          aria-hidden="true"
                        ></span>
                        <i v-else class="bi bi-trash"></i>
                      </button>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>
      </div>

      <div class="col-12 col-xl-5">
        <form class="content-panel" @submit.prevent="saveUser">
          <div class="d-flex align-items-start justify-content-between gap-3 mb-3">
            <div>
              <h3 class="h5 mb-1">{{ isEditing ? 'Edit user' : 'Tambah user' }}</h3>
              <p class="text-secondary mb-0">Owner dapat mengatur role, status, dan instansi user.</p>
            </div>
            <button v-if="isEditing" type="button" class="btn btn-light btn-sm" @click="resetForm">
              <i class="bi bi-x-lg me-1"></i>
              Batal
            </button>
          </div>

          <div class="row g-3">
            <div class="col-12">
              <label for="companyID" class="form-label">Instansi *</label>
              <select id="companyID" v-model="form.company_id" class="form-select" required>
                <option value="" disabled>Pilih instansi</option>
                <option v-for="item in companyOptions" :key="item.id" :value="item.id">
                  {{ item.name }}
                </option>
              </select>
            </div>

            <div class="col-12 col-md-6">
              <label for="userName" class="form-label">Nama *</label>
              <input id="userName" v-model="form.name" type="text" class="form-control" required />
            </div>

            <div class="col-12 col-md-6">
              <label for="userEmail" class="form-label">Email *</label>
              <input id="userEmail" v-model="form.email" type="email" class="form-control" required />
            </div>

            <div class="col-12 col-md-6">
              <label for="userPassword" class="form-label">Password {{ isEditing ? 'Baru' : '*' }}</label>
              <input
                id="userPassword"
                v-model="form.password"
                type="password"
                class="form-control"
                autocomplete="new-password"
                :required="!isEditing"
                placeholder="Kosongkan jika tidak diganti"
              />
            </div>

            <div class="col-12 col-md-6">
              <label for="userRole" class="form-label">Role *</label>
              <select id="userRole" v-model="form.role" class="form-select" required>
                <option v-for="role in roleOptions" :key="role.value" :value="role.value">{{ role.label }}</option>
              </select>
            </div>

            <div class="col-12">
              <div class="field-settings">
                <div class="form-check form-switch">
                  <input id="userActive" v-model="form.is_active" class="form-check-input" type="checkbox" />
                  <label class="form-check-label" for="userActive">User aktif</label>
                </div>
              </div>
            </div>
          </div>

          <div class="d-flex justify-content-end mt-4">
            <button type="submit" class="btn btn-primary px-3" :disabled="isSaving">
              <span v-if="isSaving" class="spinner-border spinner-border-sm me-2" aria-hidden="true"></span>
              <i v-else class="bi bi-save me-2"></i>
              {{ isSaving ? 'Menyimpan...' : isEditing ? 'Simpan Perubahan' : 'Tambah User' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </AdminLayout>
</template>

<style scoped>
.empty-state {
  border: 1px dashed #cbd5e1;
  border-radius: 16px;
  background: #f8fafc;
  padding: 24px;
}

.filter-grid {
  display: grid;
  gap: 10px;
  grid-template-columns: minmax(180px, 1fr) 150px 150px auto;
}

.table th {
  color: #64748b;
  font-size: 0.78rem;
  text-transform: uppercase;
}

.form-label,
.form-check-label {
  color: #334155;
  font-weight: 650;
}

.form-control,
.form-select {
  border-color: #dbe3ef;
  border-radius: 12px;
  padding: 0.76rem 0.9rem;
}

.form-control:focus,
.form-select:focus,
.form-check-input:focus {
  border-color: #86b7fe;
  box-shadow: 0 0 0 0.25rem rgba(13, 110, 253, 0.12);
}

.field-settings {
  border: 1px solid #e2e8f0;
  border-radius: 16px;
  background: #f8fafc;
  padding: 16px;
}

.icon-btn {
  display: inline-grid;
  width: 34px;
  height: 34px;
  place-items: center;
  padding: 0;
}

@media (max-width: 767.98px) {
  .filter-grid {
    grid-template-columns: 1fr;
  }
}
</style>
