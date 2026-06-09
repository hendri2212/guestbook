<script setup>
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'
const AUTH_STORAGE_KEY = 'guestbook_admin_auth'

const router = useRouter()
const auth = ref(readStoredAuth())
const isLoading = ref(false)
const loadError = ref('')

const user = computed(() => auth.value?.user || null)
const company = computed(() => auth.value?.company || null)
const canManageGuestForms = computed(() => user.value?.role === 'admin')
const publicFormUrl = computed(() => {
  const slug = company.value?.slug ? `buku-tamu-${company.value.slug}` : 'buku-tamu-instansi-demo'
  return `/forms/${slug}`
})

const stats = [
  { label: 'Tamu hari ini', value: '12', tone: 'primary' },
  { label: 'Sedang berkunjung', value: '4', tone: 'success' },
  { label: 'Form aktif', value: '1', tone: 'info' },
  { label: 'Perlu follow-up', value: '3', tone: 'warning' },
]

const recentVisits = [
  { name: 'Budi Santoso', company: 'PT Contoh Sejahtera', purpose: 'Meeting administrasi', status: 'Checked in' },
  { name: 'Sari Dewi', company: 'CV Nusantara', purpose: 'Pengiriman dokumen', status: 'Checked out' },
  { name: 'Andi Pratama', company: 'Mandiri Vendor', purpose: 'Koordinasi proyek', status: 'Checked in' },
]

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

async function loadProfile() {
  if (!auth.value?.token) {
    router.replace({ name: 'admin-login' })
    return
  }

  isLoading.value = true
  loadError.value = ''

  try {
    const response = await fetch(`${API_BASE_URL}/api/admin/me`, {
      headers: {
        Authorization: `Bearer ${auth.value.token}`,
      },
    })

    const data = await response.json().catch(() => ({}))

    if (!response.ok) {
      throw new Error(data.error || data.message || 'Sesi tidak valid.')
    }

    auth.value = {
      ...auth.value,
      user: data.user,
      company: data.company,
    }
    localStorage.setItem(AUTH_STORAGE_KEY, JSON.stringify(auth.value))
  } catch (error) {
    loadError.value = error.message || 'Gagal memuat profil admin.'
    clearSession()
  } finally {
    isLoading.value = false
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
        <RouterLink class="nav-link active" to="/admin">Dashboard</RouterLink>
        <a class="nav-link disabled" href="#" aria-disabled="true">Kunjungan</a>
        <RouterLink v-if="canManageGuestForms" class="nav-link" to="/admin/guest-forms">
          Form Public
        </RouterLink>
        <a class="nav-link disabled" href="#" aria-disabled="true">Pengaturan</a>
      </nav>

      <div class="sidebar-footer mt-auto">
        <p class="small text-secondary mb-1">Instansi</p>
        <p class="fw-semibold mb-3">{{ company?.name || 'Memuat...' }}</p>
        <button type="button" class="btn btn-outline-secondary w-100" @click="logout">Keluar</button>
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
        Memuat profil admin...
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
              <span :class="`badge text-bg-${item.tone}`">Live</span>
            </div>
          </div>
        </div>
      </div>

      <div class="row g-4">
        <div class="col-12 col-xl-8">
          <section class="content-panel">
            <div class="d-flex align-items-center justify-content-between gap-3 mb-3">
              <div>
                <h3 class="h5 mb-1">Kunjungan terbaru</h3>
                <p class="text-secondary mb-0">Data contoh untuk template awal dashboard.</p>
              </div>
              <RouterLink v-if="canManageGuestForms" class="btn btn-outline-primary btn-sm" to="/admin/guest-forms">
                Kelola form
              </RouterLink>
            </div>

            <div class="table-responsive">
              <table class="table align-middle mb-0">
                <thead>
                  <tr>
                    <th>Nama</th>
                    <th>Instansi</th>
                    <th>Keperluan</th>
                    <th>Status</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="visit in recentVisits" :key="visit.name">
                    <td class="fw-semibold">{{ visit.name }}</td>
                    <td>{{ visit.company }}</td>
                    <td>{{ visit.purpose }}</td>
                    <td>
                      <span class="badge rounded-pill text-bg-light border">{{ visit.status }}</span>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </section>
        </div>

        <div class="col-12 col-xl-4">
          <section class="content-panel">
            <h3 class="h5 mb-3">Akses form public</h3>
            <p class="text-secondary">
              Gunakan tautan ini untuk membuka halaman check-in tamu milik instansi.
            </p>
            <RouterLink class="btn btn-primary w-100" :to="publicFormUrl">Buka Form Public</RouterLink>
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
