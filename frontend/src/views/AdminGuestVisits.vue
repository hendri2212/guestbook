<script setup>
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'
const AUTH_STORAGE_KEY = 'guestbook_admin_auth'

const router = useRouter()
const auth = ref(readStoredAuth())
const guestForms = ref([])
const guestVisits = ref([])
const isLoading = ref(false)
const pageError = ref('')

const user = computed(() => auth.value?.user || null)
const company = computed(() => auth.value?.company || null)
const formSummaries = computed(() => {
  return guestForms.value.map((guestForm) => {
    const visits = guestVisits.value.filter((visit) => visit.form_id === guestForm.id)
    const companies = uniqueCompanies(visits)

    return {
      form: guestForm,
      total: visits.length,
      companies,
    }
  })
})

function readStoredAuth() {
  try {
    const raw = localStorage.getItem(AUTH_STORAGE_KEY)
    return raw ? JSON.parse(raw) : null
  } catch {
    return null
  }
}

function authHeaders() {
  return {
    Authorization: `Bearer ${auth.value?.token}`,
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

function uniqueCompanies(visits) {
  return [
    ...new Set(
      visits
        .map((visit) => (visit.guest_company || '').trim())
        .filter((companyName) => companyName !== ''),
    ),
  ]
}

async function handleUnauthorized(errorMessage) {
  pageError.value = errorMessage
  if (isAuthError(errorMessage)) {
    clearSession()
    await router.replace({ name: 'admin-login' })
  }
}

function isAuthError(message) {
  const normalized = message.toLowerCase()
  return normalized.includes('unauthorized') || normalized.includes('invalid') || normalized.includes('expired')
}

async function loadData() {
  if (!auth.value?.token) {
    router.replace({ name: 'admin-login' })
    return
  }

  isLoading.value = true
  pageError.value = ''

  try {
    const [formsResponse, visitsResponse] = await Promise.all([
      fetch(`${API_BASE_URL}/api/admin/guest-forms`, { headers: authHeaders() }),
      fetch(`${API_BASE_URL}/api/admin/guest-visits`, { headers: authHeaders() }),
    ])
    const formsData = await formsResponse.json().catch(() => ({}))
    const visitsData = await visitsResponse.json().catch(() => ({}))

    if (formsResponse.status === 401 || visitsResponse.status === 401) {
      throw new Error('invalid or expired token')
    }

    if (!formsResponse.ok) {
      throw new Error(formsData.error || formsData.message || 'Gagal memuat daftar form.')
    }
    if (!visitsResponse.ok) {
      throw new Error(visitsData.error || visitsData.message || 'Gagal memuat ringkasan kunjungan.')
    }

    guestForms.value = formsData
    guestVisits.value = visitsData
  } catch (error) {
    await handleUnauthorized(error.message || 'Gagal memuat data kunjungan.')
  } finally {
    isLoading.value = false
  }
}

onMounted(loadData)
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
        <RouterLink class="nav-link active" to="/admin/guest-visits">
          <i class="bi bi-people"></i>
          Kunjungan
        </RouterLink>
        <RouterLink v-if="user?.role === 'admin'" class="nav-link" to="/admin/guest-forms">
          <i class="bi bi-ui-checks-grid"></i>
          Form Public
        </RouterLink>
        <RouterLink class="nav-link" to="/admin/settings">
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
          <p class="text-uppercase small fw-semibold text-primary mb-2">Kunjungan</p>
          <h2 class="h3 mb-1">Pilih form kunjungan</h2>
          <p class="text-secondary mb-0">Pilih form untuk melihat daftar tamu dan ringkasan kehadiran.</p>
        </div>
        <div class="user-chip">
          <span class="avatar">{{ (user?.name || 'A').slice(0, 1) }}</span>
          <div>
            <p class="fw-semibold mb-0">{{ user?.name || 'Admin' }}</p>
            <p class="small text-secondary mb-0">{{ user?.email || 'admin' }}</p>
          </div>
        </div>
      </header>

      <div v-if="pageError" class="alert alert-danger" role="alert">{{ pageError }}</div>

      <section class="content-panel">
        <div class="d-flex align-items-center justify-content-between gap-3 mb-4">
          <div>
            <h3 class="h5 mb-1">Form public</h3>
            <p class="text-secondary mb-0">Setiap form memiliki daftar kunjungan dan ringkasan sendiri.</p>
          </div>
          <button type="button" class="btn btn-outline-primary btn-sm" @click="loadData">
            <i class="bi bi-arrow-clockwise me-1"></i>
            Refresh
          </button>
        </div>

        <div v-if="isLoading" class="alert alert-light border mb-0" role="status">
          Memuat form kunjungan...
        </div>

        <div v-else-if="formSummaries.length === 0" class="empty-state">
          <h4 class="h6 mb-2">Belum ada form public</h4>
          <p class="text-secondary mb-0">Buat form public terlebih dahulu sebelum melihat kunjungan.</p>
        </div>

        <div v-else class="form-grid">
          <RouterLink
            v-for="summary in formSummaries"
            :key="summary.form.id"
            class="form-card"
            :to="{ name: 'admin-guest-visit-detail', params: { formId: summary.form.id } }"
          >
            <div class="form-card-row">
              <div class="form-card-title d-flex align-items-center gap-3">
                <span class="title-icon">
                  <i class="bi bi-clipboard2-check"></i>
                </span>
                <div class="min-w-0">
                  <h4 class="h6 mb-1">{{ summary.form.title }}</h4>
                  <p class="text-secondary mb-0">{{ summary.form.public_slug }}</p>
                </div>
              </div>
              <div class="form-card-meta">
                <div class="mini-metric">
                  <span><i class="bi bi-person-check"></i></span>
                  <strong>{{ summary.total }}</strong>
                </div>
                <div class="mini-metric">
                  <span><i class="bi bi-buildings"></i></span>
                  <strong>{{ summary.companies.length }}</strong>
                </div>
                <span class="form-icon"><i class="bi bi-arrow-right"></i></span>
              </div>
            </div>
          </RouterLink>
        </div>
      </section>
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
.avatar {
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

.user-chip {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
}

.content-panel {
  padding: 20px;
}

.form-grid {
  display: grid;
  gap: 16px;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
}

.form-card {
  border: 1px solid #dbe3ef;
  border-radius: 18px;
  background: #f8fafc;
  color: inherit;
  padding: 18px;
  text-decoration: none;
  transition:
    border-color 0.15s ease,
    box-shadow 0.15s ease,
    transform 0.15s ease;
}

.form-card:hover {
  border-color: #0d6efd;
  box-shadow: 0 14px 32px rgba(13, 110, 253, 0.14);
  transform: translateY(-1px);
}

.form-card-row,
.form-card-meta {
  display: flex;
  align-items: center;
  gap: 14px;
}

.form-card-row {
  justify-content: space-between;
}

.form-card-title {
  min-width: 0;
}

.min-w-0 {
  min-width: 0;
}

.form-card-title h4,
.form-card-title p {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.title-icon {
  display: grid;
  width: 42px;
  height: 42px;
  flex: 0 0 auto;
  place-items: center;
  border-radius: 14px;
  background: #eaf2ff;
  color: #0d6efd;
  font-size: 1.25rem;
}

.form-card-meta {
  flex: 0 0 auto;
}

.form-icon {
  display: grid;
  width: 34px;
  height: 34px;
  flex: 0 0 auto;
  place-items: center;
  border-radius: 50%;
  background: #eaf2ff;
  color: #0d6efd;
}

.mini-metric {
  border: 1px solid #e2e8f0;
  border-radius: 14px;
  background: #ffffff;
  display: flex;
  align-items: center;
  justify-content: flex-start;
  gap: 10px;
  min-width: 74px;
  padding: 10px 12px;
  text-align: left;
}

.mini-metric span {
  display: block;
  color: #0d6efd;
  font-size: 1.35rem;
  line-height: 1;
}

.mini-metric strong {
  color: #111827;
  font-size: 1.25rem;
}

.empty-state {
  border: 1px dashed #cbd5e1;
  border-radius: 16px;
  background: #f8fafc;
  padding: 24px;
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

  .form-card-row {
    align-items: flex-start;
    flex-direction: column;
  }

  .form-card-meta {
    width: 100%;
    justify-content: space-between;
  }
}
</style>
