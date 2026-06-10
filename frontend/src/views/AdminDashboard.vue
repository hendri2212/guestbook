<script setup>
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'
const AUTH_STORAGE_KEY = 'guestbook_admin_auth'

const router = useRouter()
const auth = ref(readStoredAuth())
const isLoading = ref(false)
const loadError = ref('')
const guestForms = ref([])
const guestVisits = ref([])

const user = computed(() => auth.value?.user || null)
const company = computed(() => auth.value?.company || null)
const canManageGuestForms = computed(() => user.value?.role === 'admin')
const activeGuestForms = computed(() => guestForms.value.filter((guestForm) => guestForm.is_active))
const todayDate = computed(() => new Date().toISOString().slice(0, 10))
const todayVisits = computed(() => guestVisits.value.filter((visit) => visit.visit_date === todayDate.value))
const checkedInVisits = computed(() => guestVisits.value.filter((visit) => visit.status === 'checked_in'))
const companiesCount = computed(() => {
  return new Set(
    guestVisits.value
      .map((visit) => (visit.guest_company || '').trim())
      .filter((companyName) => companyName !== ''),
  ).size
})
const stats = computed(() => [
  { label: 'Tamu hari ini', value: todayVisits.value.length, tone: 'primary' },
  { label: 'Sedang berkunjung', value: checkedInVisits.value.length, tone: 'success' },
  { label: 'Form aktif', value: activeGuestForms.value.length, tone: 'info' },
  { label: 'Instansi tercatat', value: companiesCount.value, tone: 'warning' },
])
const recentVisits = computed(() => guestVisits.value.slice(0, 5))

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

function statusLabel(status) {
  if (status === 'checked_in') {
    return 'Checked in'
  }
  if (status === 'checked_out') {
    return 'Checked out'
  }
  if (status === 'cancelled') {
    return 'Cancelled'
  }
  return status || '-'
}

async function loadDashboard() {
  if (!auth.value?.token) {
    router.replace({ name: 'admin-login' })
    return
  }

  isLoading.value = true
  loadError.value = ''

  try {
    const [profileResponse, formsResponse, visitsResponse] = await Promise.all([
      fetch(`${API_BASE_URL}/api/admin/me`, { headers: authHeaders() }),
      fetch(`${API_BASE_URL}/api/admin/guest-forms`, { headers: authHeaders() }),
      fetch(`${API_BASE_URL}/api/admin/guest-visits`, { headers: authHeaders() }),
    ])
    const profileData = await profileResponse.json().catch(() => ({}))
    const formsData = await formsResponse.json().catch(() => ({}))
    const visitsData = await visitsResponse.json().catch(() => ({}))

    if ([profileResponse, formsResponse, visitsResponse].some((response) => response.status === 401)) {
      redirectToLogin()
      return
    }

    if (!profileResponse.ok) {
      throw new Error(profileData.error || profileData.message || 'Sesi tidak valid.')
    }
    if (!formsResponse.ok) {
      throw new Error(formsData.error || formsData.message || 'Gagal memuat form public.')
    }
    if (!visitsResponse.ok) {
      throw new Error(visitsData.error || visitsData.message || 'Gagal memuat kunjungan.')
    }

    auth.value = {
      ...auth.value,
      user: profileData.user,
      company: profileData.company,
    }
    localStorage.setItem(AUTH_STORAGE_KEY, JSON.stringify(auth.value))
    guestForms.value = formsData
    guestVisits.value = visitsData
  } catch (error) {
    loadError.value = error.message || 'Gagal memuat dashboard.'
    if (isAuthError(loadError.value)) {
      redirectToLogin()
    }
  } finally {
    isLoading.value = false
  }
}

function isAuthError(message) {
  const normalized = message.toLowerCase()
  return normalized.includes('unauthorized') || normalized.includes('invalid') || normalized.includes('expired') || normalized.includes('sesi')
}

onMounted(loadDashboard)
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
        <RouterLink class="nav-link active" to="/admin">
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
        <a class="nav-link disabled" href="#" aria-disabled="true">
          <i class="bi bi-gear"></i>
          Pengaturan
        </a>
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
          <p class="text-uppercase small fw-semibold text-primary mb-2">Dashboard</p>
          <h2 class="h3 mb-1">Ringkasan buku tamu</h2>
          <p class="text-secondary mb-0">
            {{ company?.name || 'Memuat instansi' }} · {{ user?.role || 'admin' }}
          </p>
        </div>
        <div class="user-chip">
          <span class="avatar">{{ (user?.name || 'A').slice(0, 1) }}</span>
          <div>
            <p class="fw-semibold mb-0">{{ user?.name || 'Admin' }}</p>
            <p class="small text-secondary mb-0">{{ user?.email || 'admin' }}</p>
          </div>
        </div>
      </header>

      <div v-if="isLoading" class="alert alert-light border" role="status">
        Memuat dashboard...
      </div>

      <div v-if="loadError" class="alert alert-warning" role="alert">
        {{ loadError }} Silakan login ulang.
      </div>

      <div class="row g-3 mb-4">
        <div v-for="item in stats" :key="item.label" class="col-12 col-sm-6 col-xl-3">
          <div class="metric-card">
            <p class="text-secondary mb-2">{{ item.label }}</p>
            <div class="d-flex align-items-end justify-content-between">
              <strong class="metric-value">{{ item.value }}</strong>
              <span :class="`badge text-bg-${item.tone}`">
                <i class="bi bi-activity me-1"></i>
                Live
              </span>
            </div>
          </div>
        </div>
      </div>

      <div class="row g-4">
        <div class="col-12">
          <section class="content-panel">
            <div class="d-flex align-items-center justify-content-between gap-3 mb-3">
              <div>
                <h3 class="h5 mb-1">Kunjungan terbaru</h3>
                <p class="text-secondary mb-0">Data terbaru dari seluruh form public instansi.</p>
              </div>
              <RouterLink class="btn btn-outline-primary btn-sm" to="/admin/guest-visits">
                <i class="bi bi-list-check me-1"></i>
                Kelola kunjungan
              </RouterLink>
            </div>

            <div v-if="recentVisits.length === 0" class="empty-state">
              <h4 class="h6 mb-2">Belum ada kunjungan</h4>
              <p class="text-secondary mb-0">Data akan tampil setelah tamu mengisi form public.</p>
            </div>

            <div v-else class="table-responsive">
              <table class="table align-middle mb-0">
                <thead>
                  <tr>
                    <th>Nama</th>
                    <th>Instansi</th>
                    <th>Keperluan</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="visit in recentVisits" :key="visit.id">
                    <td class="fw-semibold">{{ visit.guest_name }}</td>
                    <td>{{ visit.guest_company || '-' }}</td>
                    <td>{{ visit.purpose }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </section>
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
.metric-card,
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

.metric-card,
.content-panel {
  padding: 20px;
}

.metric-value {
  color: #111827;
  font-size: 2rem;
  line-height: 1;
}

.empty-state {
  border: 1px dashed #cbd5e1;
  border-radius: 16px;
  background: #f8fafc;
  padding: 24px;
}

.table th {
  color: #64748b;
  font-size: 0.78rem;
  text-transform: uppercase;
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
