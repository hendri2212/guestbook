<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'
const AUTH_STORAGE_KEY = 'guestbook_admin_auth'

const router = useRouter()
const auth = ref(readStoredAuth())
const isLoading = ref(false)
const isSavingCompany = ref(false)
const loadError = ref('')
const successMessage = ref('')
const companyDetail = ref(null)

const companyForm = reactive({
  name: '',
  slug: '',
  email: '',
  phone: '',
  address: '',
  is_active: true,
})

const user = computed(() => auth.value?.user || null)
const company = computed(() => companyDetail.value || auth.value?.company || null)
const canManageGuestForms = computed(() => user.value?.role === 'admin')
const canEditCompany = computed(() => ['owner', 'admin'].includes(user.value?.role))

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

function fillCompanyForm(companyData) {
  Object.assign(companyForm, {
    name: companyData?.name || '',
    slug: companyData?.slug || '',
    email: companyData?.email || '',
    phone: companyData?.phone || '',
    address: companyData?.address || '',
    is_active: Boolean(companyData?.is_active),
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
    const profileResponse = await fetch(`${API_BASE_URL}/api/admin/me`, {
      headers: authHeaders(),
    })
    const profileData = await profileResponse.json().catch(() => ({}))

    if (profileResponse.status === 401) {
      redirectToLogin()
      return
    }
    if (!profileResponse.ok) {
      throw new Error(profileData.error || profileData.message || 'Gagal memuat pengaturan.')
    }

    auth.value = {
      ...auth.value,
      user: profileData.user,
      company: profileData.company,
    }
    localStorage.setItem(AUTH_STORAGE_KEY, JSON.stringify(auth.value))

    const companyResponse = await fetch(`${API_BASE_URL}/api/admin/companies/${profileData.company.id}`, {
      headers: authHeaders(),
    })
    const companyData = await companyResponse.json().catch(() => ({}))

    if (companyResponse.status === 401) {
      redirectToLogin()
      return
    }
    if (!companyResponse.ok) {
      throw new Error(companyData.error || companyData.message || 'Gagal memuat data instansi.')
    }

    companyDetail.value = companyData
    fillCompanyForm(companyData)
  } catch (error) {
    loadError.value = error.message || 'Gagal memuat pengaturan.'
  } finally {
    isLoading.value = false
  }
}

async function saveCompany() {
  if (!company.value?.id) {
    loadError.value = 'Data instansi belum tersedia.'
    return
  }

  isSavingCompany.value = true
  loadError.value = ''
  successMessage.value = ''

  try {
    const response = await fetch(`${API_BASE_URL}/api/admin/companies/${company.value.id}`, {
      method: 'PUT',
      headers: authHeaders(true),
      body: JSON.stringify({
        name: companyForm.name.trim(),
        slug: companyForm.slug.trim(),
        email: optionalString(companyForm.email),
        phone: optionalString(companyForm.phone),
        address: optionalString(companyForm.address),
        is_active: companyForm.is_active,
      }),
    })
    const data = await response.json().catch(() => ({}))

    if (response.status === 401) {
      redirectToLogin()
      return
    }
    if (!response.ok) {
      throw new Error(data.error || data.message || 'Gagal menyimpan data instansi.')
    }

    companyDetail.value = data
    fillCompanyForm(data)
    auth.value = {
      ...auth.value,
      company: {
        id: data.id,
        name: data.name,
        slug: data.slug,
      },
    }
    localStorage.setItem(AUTH_STORAGE_KEY, JSON.stringify(auth.value))
    successMessage.value = 'Data instansi berhasil diperbarui.'
  } catch (error) {
    loadError.value = error.message || 'Gagal menyimpan data instansi.'
  } finally {
    isSavingCompany.value = false
  }
}

onMounted(loadProfile)
</script>

<template>
  <main class="admin-page min-vh-100">
    <aside class="sidebar">
      <div class="d-flex align-items-center gap-3 mb-4">
        <div class="brand-mark">BT</div>
        <div>
          <p class="small text-secondary mb-0">Guestbook</p>
          <h1 class="h6 mb-0">Admin Panel</h1>
        </div>
      </div>

      <nav class="nav flex-column gap-1">
        <RouterLink class="nav-link" to="/admin">
          <i class="bi bi-speedometer2"></i>
          Dashboard
        </RouterLink>
        <RouterLink class="nav-link" to="/admin/guest-visits">
          <i class="bi bi-people"></i>
          Kunjungan
        </RouterLink>
        <RouterLink v-if="canManageGuestForms" class="nav-link" to="/admin/guest-forms">
          <i class="bi bi-ui-checks-grid"></i>
          Form Public
        </RouterLink>
        <RouterLink class="nav-link active" to="/admin/settings">
          <i class="bi bi-gear"></i>
          Pengaturan
        </RouterLink>
      </nav>

      <div class="sidebar-footer mt-auto">
        <p class="small text-secondary mb-1">Instansi</p>
        <p class="fw-semibold mb-3">{{ company?.name || 'Memuat...' }}</p>
        <button type="button" class="btn btn-outline-secondary w-100" @click="logout">
          <i class="bi bi-box-arrow-right me-2"></i>
          Keluar
        </button>
      </div>
    </aside>

    <section class="admin-content">
      <header class="admin-header">
        <div>
          <p class="text-uppercase small fw-semibold text-primary mb-2">Pengaturan</p>
          <h2 class="h3 mb-1">Pengaturan akun dan instansi</h2>
          <p class="text-secondary mb-0">Informasi dasar dari sesi admin yang sedang aktif.</p>
        </div>
        <div class="user-chip">
          <span class="avatar">{{ (user?.name || 'A').slice(0, 1) }}</span>
          <div>
            <p class="fw-semibold mb-0">{{ user?.name || 'Admin' }}</p>
            <p class="small text-secondary mb-0">{{ user?.email || 'admin' }}</p>
          </div>
        </div>
      </header>

      <div v-if="isLoading" class="alert alert-light border" role="status">Memuat pengaturan...</div>
      <div v-if="loadError" class="alert alert-warning" role="alert">{{ loadError }}</div>
      <div v-if="successMessage" class="alert alert-success" role="alert">{{ successMessage }}</div>

      <div class="row g-4">
        <div class="col-12 col-xl-6">
          <section class="content-panel">
            <div class="settings-heading">
              <span class="settings-icon"><i class="bi bi-person-badge"></i></span>
              <div>
                <h3 class="h5 mb-1">Akun admin</h3>
                <p class="text-secondary mb-0">Data pengguna yang sedang login.</p>
              </div>
            </div>

            <dl class="settings-list mb-0">
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
            </dl>
          </section>
        </div>

        <div class="col-12 col-xl-6">
          <form class="content-panel" @submit.prevent="saveCompany">
            <div class="settings-heading">
              <span class="settings-icon"><i class="bi bi-building"></i></span>
              <div>
                <h3 class="h5 mb-1">Instansi</h3>
                <p class="text-secondary mb-0">
                  {{ canEditCompany ? 'Perbarui profil instansi.' : 'Anda tidak memiliki akses untuk mengubah data instansi.' }}
                </p>
              </div>
            </div>

            <div class="row g-3">
              <div class="col-12 col-md-6">
                <label for="companyName" class="form-label">Nama Instansi</label>
                <input
                  id="companyName"
                  v-model="companyForm.name"
                  type="text"
                  class="form-control"
                  :disabled="!canEditCompany"
                  required
                />
              </div>

              <div class="col-12 col-md-6">
                <label for="companySlug" class="form-label">Slug</label>
                <input
                  id="companySlug"
                  v-model="companyForm.slug"
                  type="text"
                  class="form-control"
                  :disabled="!canEditCompany"
                  required
                />
              </div>

              <div class="col-12 col-md-6">
                <label for="companyEmail" class="form-label">Email</label>
                <input
                  id="companyEmail"
                  v-model="companyForm.email"
                  type="email"
                  class="form-control"
                  :disabled="!canEditCompany"
                />
              </div>

              <div class="col-12 col-md-6">
                <label for="companyPhone" class="form-label">Telepon</label>
                <input
                  id="companyPhone"
                  v-model="companyForm.phone"
                  type="tel"
                  class="form-control"
                  :disabled="!canEditCompany"
                />
              </div>

              <div class="col-12">
                <label for="companyAddress" class="form-label">Alamat</label>
                <textarea
                  id="companyAddress"
                  v-model="companyForm.address"
                  class="form-control"
                  rows="3"
                  :disabled="!canEditCompany"
                ></textarea>
              </div>

              <div class="col-12">
                <div class="form-check form-switch">
                  <input
                    id="companyActive"
                    v-model="companyForm.is_active"
                    class="form-check-input"
                    type="checkbox"
                    :disabled="!canEditCompany"
                  />
                  <label class="form-check-label" for="companyActive">Instansi aktif</label>
                </div>
              </div>
            </div>

            <div class="d-grid mt-4">
              <button type="submit" class="btn btn-primary py-3" :disabled="!canEditCompany || isSavingCompany">
                <span v-if="isSavingCompany" class="spinner-border spinner-border-sm me-2" aria-hidden="true"></span>
                <i v-else class="bi bi-save me-2"></i>
                {{ isSavingCompany ? 'Menyimpan...' : 'Simpan Data Instansi' }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </section>
  </main>
</template>

<style scoped>
.admin-page {
  display: grid;
  grid-template-columns: 280px 1fr;
  background: #eef2f6;
}

.sidebar {
  position: sticky;
  top: 0;
  display: flex;
  flex-direction: column;
  height: 100vh;
  border-right: 1px solid #dbe3ef;
  background: #ffffff;
  padding: 24px;
}

.brand-mark,
.avatar,
.settings-icon {
  display: grid;
  place-items: center;
  background: #0d6efd;
  color: #ffffff;
  font-weight: 800;
}

.brand-mark {
  width: 46px;
  height: 46px;
  border-radius: 14px;
}

.avatar {
  width: 42px;
  height: 42px;
  border-radius: 50%;
  text-transform: uppercase;
}

.nav-link {
  display: flex;
  align-items: center;
  gap: 10px;
  border-radius: 12px;
  color: #475569;
  font-weight: 650;
  padding: 0.78rem 0.9rem;
}

.nav-link.active {
  background: #eaf2ff;
  color: #0d6efd;
}

.sidebar-footer {
  border-top: 1px solid #e2e8f0;
  padding-top: 18px;
}

.admin-content {
  min-width: 0;
  padding: 28px;
}

.admin-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  margin-bottom: 24px;
}

.user-chip,
.content-panel {
  border: 1px solid rgba(148, 163, 184, 0.24);
  border-radius: 18px;
  background: #ffffff;
  box-shadow: 0 16px 46px rgba(15, 23, 42, 0.06);
}

.user-chip,
.settings-heading {
  display: flex;
  align-items: center;
  gap: 12px;
}

.user-chip {
  padding: 12px 16px;
}

.content-panel {
  padding: 20px;
}

.settings-heading {
  margin-bottom: 20px;
}

.settings-icon {
  width: 46px;
  height: 46px;
  flex: 0 0 auto;
  border-radius: 14px;
}

.settings-list {
  display: grid;
  gap: 12px;
}

.settings-list div {
  border-top: 1px solid #e2e8f0;
  padding-top: 12px;
}

.settings-list dt {
  color: #64748b;
  font-size: 0.78rem;
  text-transform: uppercase;
}

.settings-list dd {
  margin-bottom: 0;
  color: #111827;
  font-weight: 650;
  overflow-wrap: anywhere;
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

.btn {
  border-radius: 12px;
  font-weight: 700;
}

@media (max-width: 991.98px) {
  .admin-page {
    grid-template-columns: 1fr;
  }

  .sidebar {
    position: static;
    height: auto;
  }

  .admin-header {
    align-items: flex-start;
    flex-direction: column;
  }

  .user-chip {
    width: 100%;
  }
}

@media (max-width: 575.98px) {
  .admin-content,
  .sidebar {
    padding: 18px;
  }
}
</style>
