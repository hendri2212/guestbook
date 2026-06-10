<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import AdminLayout from '@/components/AdminLayout.vue'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'
const AUTH_STORAGE_KEY = 'guestbook_admin_auth'

const router = useRouter()
const auth = ref(readStoredAuth())
const companies = ref([])
const isLoading = ref(false)
const isSaving = ref(false)
const isDeleting = ref('')
const pageError = ref('')
const successMessage = ref('')
const editingID = ref(null)

const form = reactive({
  name: '',
  slug: '',
  email: '',
  phone: '',
  address: '',
  is_active: true,
})

const user = computed(() => auth.value?.user || null)
const company = computed(() => auth.value?.company || null)
const isEditing = computed(() => Boolean(editingID.value))
const endpoint = computed(() => `${API_BASE_URL}/api/admin/companies`)

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

function optionalString(value) {
  const trimmed = value.trim()
  return trimmed === '' ? null : trimmed
}

function isAuthError(message) {
  const normalized = message.toLowerCase()
  return normalized.includes('unauthorized') || normalized.includes('invalid') || normalized.includes('expired')
}

function resetForm() {
  editingID.value = null
  Object.assign(form, {
    name: '',
    slug: '',
    email: '',
    phone: '',
    address: '',
    is_active: true,
  })
}

function editCompany(item) {
  editingID.value = item.id
  Object.assign(form, {
    name: item.name || '',
    slug: item.slug || '',
    email: item.email || '',
    phone: item.phone || '',
    address: item.address || '',
    is_active: Boolean(item.is_active),
  })
  pageError.value = ''
  successMessage.value = ''
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

  return true
}

async function loadCompanies() {
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

    const response = await fetch(endpoint.value, {
      headers: authHeaders(),
    })
    const data = await response.json().catch(() => ({}))

    if (response.status === 401) {
      redirectToLogin()
      return
    }
    if (!response.ok) {
      throw new Error(data.error || data.message || 'Gagal memuat company.')
    }

    companies.value = data
  } catch (error) {
    pageError.value = error.message || 'Gagal memuat company.'
    if (isAuthError(pageError.value)) {
      redirectToLogin()
    }
  } finally {
    isLoading.value = false
  }
}

async function saveCompany() {
  pageError.value = ''
  successMessage.value = ''
  isSaving.value = true

  try {
    const response = await fetch(isEditing.value ? `${endpoint.value}/${editingID.value}` : endpoint.value, {
      method: isEditing.value ? 'PUT' : 'POST',
      headers: authHeaders(true),
      body: JSON.stringify({
        name: form.name.trim(),
        slug: form.slug.trim(),
        email: optionalString(form.email),
        phone: optionalString(form.phone),
        address: optionalString(form.address),
        is_active: form.is_active,
      }),
    })
    const data = await response.json().catch(() => ({}))

    if (response.status === 401) {
      redirectToLogin()
      return
    }
    if (!response.ok) {
      throw new Error(data.error || data.message || 'Company gagal disimpan.')
    }

    successMessage.value = isEditing.value ? 'Company berhasil diperbarui.' : 'Company berhasil dibuat.'
    if (company.value?.id === data.id) {
      auth.value = {
        ...auth.value,
        company: {
          id: data.id,
          name: data.name,
          slug: data.slug,
        },
      }
      localStorage.setItem(AUTH_STORAGE_KEY, JSON.stringify(auth.value))
    }
    resetForm()
    await loadCompanies()
  } catch (error) {
    pageError.value = error.message || 'Company gagal disimpan.'
    if (isAuthError(pageError.value)) {
      redirectToLogin()
    }
  } finally {
    isSaving.value = false
  }
}

async function deleteCompany(item) {
  const confirmed = window.confirm(`Hapus company "${item.name}"?`)
  if (!confirmed) {
    return
  }

  pageError.value = ''
  successMessage.value = ''
  isDeleting.value = item.id

  try {
    const response = await fetch(`${endpoint.value}/${item.id}`, {
      method: 'DELETE',
      headers: authHeaders(),
    })

    if (!response.ok) {
      const data = await response.json().catch(() => ({}))
      if (response.status === 401) {
        redirectToLogin()
        return
      }
      throw new Error(data.error || data.message || 'Company gagal dihapus.')
    }

    successMessage.value = 'Company berhasil dihapus.'
    if (editingID.value === item.id) {
      resetForm()
    }
    await loadCompanies()
  } catch (error) {
    pageError.value = error.message || 'Company gagal dihapus.'
    if (isAuthError(pageError.value)) {
      redirectToLogin()
    }
  } finally {
    isDeleting.value = ''
  }
}

onMounted(loadCompanies)
</script>

<template>
  <AdminLayout :user="user" :company="company" active-menu="companies" @logout="logout">
    <template #header>
      <div>
        <p class="text-uppercase small fw-semibold text-primary mb-2">Management Company</p>
        <h2 class="h3 mb-1">Kelola company</h2>
        <p class="text-secondary mb-0">Kelola data company dan status operasionalnya.</p>
      </div>
    </template>

    <div v-if="pageError" class="alert alert-danger" role="alert">{{ pageError }}</div>
    <div v-if="successMessage" class="alert alert-success" role="alert">{{ successMessage }}</div>

    <div class="row g-4">
      <div class="col-12 col-xl-7">
        <section class="content-panel">
          <div class="d-flex align-items-center justify-content-between gap-3 mb-3">
            <div>
              <h3 class="h5 mb-1">Daftar company</h3>
              <p class="text-secondary mb-0">Data company yang tersedia dan dapat dikelola owner.</p>
            </div>
            <button type="button" class="btn btn-outline-primary btn-sm" @click="loadCompanies">
              <i class="bi bi-arrow-clockwise me-1"></i>
              Refresh
            </button>
          </div>

          <div v-if="isLoading" class="alert alert-light border mb-0" role="status">
            Memuat company...
          </div>

          <div v-else-if="companies.length === 0" class="empty-state">
            <h4 class="h6 mb-2">Belum ada company</h4>
            <p class="text-secondary mb-0">Data company akan tampil setelah backend mengembalikan data.</p>
          </div>

          <div v-else class="table-responsive">
            <table class="table align-middle mb-0">
              <thead>
                <tr>
                  <th>Company</th>
                  <th>Kontak</th>
                  <th>Statistik</th>
                  <th>Status</th>
                  <th class="text-end">Aksi</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="item in companies" :key="item.id">
                  <td>
                    <p class="fw-semibold mb-1">{{ item.name }}</p>
                    <code>{{ item.slug }}</code>
                  </td>
                  <td>
                    <p class="mb-1">{{ item.email || '-' }}</p>
                    <p class="small text-secondary mb-0">{{ item.phone || '-' }}</p>
                  </td>
                  <td>
                    <div class="stats-row">
                      <span><i class="bi bi-person-gear"></i> {{ item.stats?.admin_users || 0 }}</span>
                      <span><i class="bi bi-ui-checks-grid"></i> {{ item.stats?.guest_forms || 0 }}</span>
                      <span><i class="bi bi-people"></i> {{ item.stats?.guest_visits || 0 }}</span>
                    </div>
                  </td>
                  <td>
                    <span
                      class="badge rounded-pill"
                      :class="item.is_active ? 'text-bg-success' : 'text-bg-secondary'"
                    >
                      {{ item.is_active ? 'Aktif' : 'Nonaktif' }}
                    </span>
                  </td>
                  <td class="text-end">
                    <div class="d-inline-flex gap-2">
                      <button
                        type="button"
                        class="btn btn-outline-primary btn-sm icon-btn"
                        title="Edit company"
                        aria-label="Edit company"
                        @click="editCompany(item)"
                      >
                        <i class="bi bi-pencil-square"></i>
                      </button>
                      <button
                        type="button"
                        class="btn btn-outline-danger btn-sm icon-btn"
                        :disabled="isDeleting === item.id"
                        title="Hapus company"
                        aria-label="Hapus company"
                        @click="deleteCompany(item)"
                      >
                        <span
                          v-if="isDeleting === item.id"
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
        <form class="content-panel" @submit.prevent="saveCompany">
          <div class="d-flex align-items-start justify-content-between gap-3 mb-3">
            <div>
              <h3 class="h5 mb-1">{{ isEditing ? 'Edit company' : 'Tambah company' }}</h3>
              <p class="text-secondary mb-0">
                {{ isEditing ? 'Perbarui data company yang dipilih.' : 'Buat company baru untuk buku tamu.' }}
              </p>
            </div>
            <button v-if="isEditing" type="button" class="btn btn-light btn-sm" @click="resetForm">
              <i class="bi bi-x-lg me-1"></i>
              Batal
            </button>
          </div>

          <div class="row g-3">
            <div class="col-12 col-md-6">
              <label for="companyName" class="form-label">Nama Company *</label>
              <input id="companyName" v-model="form.name" type="text" class="form-control" required />
            </div>

            <div class="col-12 col-md-6">
              <label for="companySlug" class="form-label">Slug *</label>
              <input id="companySlug" v-model="form.slug" type="text" class="form-control" required />
            </div>

            <div class="col-12 col-md-6">
              <label for="companyEmail" class="form-label">Email</label>
              <input id="companyEmail" v-model="form.email" type="email" class="form-control" />
            </div>

            <div class="col-12 col-md-6">
              <label for="companyPhone" class="form-label">Telepon</label>
              <input id="companyPhone" v-model="form.phone" type="tel" class="form-control" />
            </div>

            <div class="col-12">
              <label for="companyAddress" class="form-label">Alamat</label>
              <textarea
                id="companyAddress"
                v-model="form.address"
                class="form-control"
                rows="3"
              ></textarea>
            </div>

            <div class="col-12">
              <div class="field-settings">
                <div class="form-check form-switch">
                  <input id="companyActive" v-model="form.is_active" class="form-check-input" type="checkbox" />
                  <label class="form-check-label" for="companyActive">Company aktif</label>
                </div>
              </div>
            </div>
          </div>

          <div class="d-flex justify-content-end mt-4">
            <button type="submit" class="btn btn-primary px-3" :disabled="isSaving">
              <span v-if="isSaving" class="spinner-border spinner-border-sm me-2" aria-hidden="true"></span>
              <i v-else class="bi bi-save me-2"></i>
              {{ isSaving ? 'Menyimpan...' : isEditing ? 'Simpan Company' : 'Tambah Company' }}
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

.stats-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.stats-row span {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  border-radius: 999px;
  background: #f1f5f9;
  color: #334155;
  font-size: 0.82rem;
  font-weight: 700;
  padding: 5px 9px;
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

.form-control {
  border-color: #dbe3ef;
  border-radius: 12px;
  padding: 0.76rem 0.9rem;
}

.form-control:focus,
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
</style>
