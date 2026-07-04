<script setup>
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import AdminLayout from '@/components/AdminLayout.vue'
import { API_BASE_URL } from '@/config/api'

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
  <AdminLayout :user="user" :company="company" active-menu="guest-visits" @logout="logout">
      <template #header>
        <div>
          <p class="text-uppercase small fw-semibold text-primary mb-2">Kunjungan</p>
          <h2 class="h3 mb-1">Pilih form kunjungan</h2>
          <p class="text-secondary mb-0">Pilih form untuk melihat daftar tamu dan ringkasan kehadiran.</p>
        </div>
      </template>

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
  </AdminLayout>
</template>

<style scoped>
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

@media (max-width: 575.98px) {
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
