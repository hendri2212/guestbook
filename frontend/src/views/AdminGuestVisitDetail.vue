<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'
const AUTH_STORAGE_KEY = 'guestbook_admin_auth'

const router = useRouter()
const route = useRoute()
const auth = ref(readStoredAuth())
const guestVisits = ref([])
const allGuestVisits = ref([])
const guestForms = ref([])
const selectedVisit = ref(null)
const isLoading = ref(false)
const isSaving = ref(false)
const isDeleting = ref('')
const isDetailModalOpen = ref(false)
const pageError = ref('')
const successMessage = ref('')
const editingID = ref(null)

const filters = reactive({
  q: '',
  status: '',
  form_id: String(route.params.formId || ''),
  date_from: '',
  date_to: '',
})

const form = reactive({
  form_id: '',
  guest_name: '',
  guest_email: '',
  guest_phone: '',
  guest_company: '',
  purpose: '',
  person_to_meet: '',
  visit_date: new Date().toISOString().slice(0, 10),
  status: 'checked_in',
  photo_url: '',
  signature_url: '',
})

const user = computed(() => auth.value?.user || null)
const company = computed(() => auth.value?.company || null)
const isEditing = computed(() => Boolean(editingID.value))
const endpoint = computed(() => `${API_BASE_URL}/api/admin/guest-visits`)
const selectedForm = computed(() => guestForms.value.find((guestForm) => guestForm.id === filters.form_id) || null)
const formSummaries = computed(() => {
  return guestForms.value.map((guestForm) => {
    const visits = allGuestVisits.value.filter((visit) => visit.form_id === guestForm.id)
    const companies = uniqueCompanies(visits)

    return {
      form: guestForm,
      total: visits.length,
      companies,
    }
  })
})
const selectedFormSummary = computed(() => {
  return formSummaries.value.find((summary) => summary.form.id === filters.form_id) || null
})
const statusOptions = [
  { value: '', label: 'Semua status' },
  { value: 'checked_in', label: 'Checked in' },
  { value: 'checked_out', label: 'Checked out' },
  { value: 'cancelled', label: 'Cancelled' },
]

function readStoredAuth() {
  try {
    const raw = localStorage.getItem(AUTH_STORAGE_KEY)
    return raw ? JSON.parse(raw) : null
  } catch {
    return null
  }
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

function clearSession() {
  localStorage.removeItem(AUTH_STORAGE_KEY)
  auth.value = null
}

function logout() {
  clearSession()
  router.push({ name: 'admin-login' })
}

function optionalString(value) {
  const trimmed = value.trim()
  return trimmed === '' ? null : trimmed
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

function statusLabel(status) {
  return statusOptions.find((option) => option.value === status)?.label || status
}

function statusClass(status) {
  if (status === 'checked_in') {
    return 'text-bg-success'
  }
  if (status === 'checked_out') {
    return 'text-bg-secondary'
  }
  return 'text-bg-warning'
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

function resetForm() {
  editingID.value = null
  Object.assign(form, {
    form_id: filters.form_id || guestForms.value[0]?.id || '',
    guest_name: '',
    guest_email: '',
    guest_phone: '',
    guest_company: '',
    purpose: '',
    person_to_meet: '',
    visit_date: new Date().toISOString().slice(0, 10),
    status: 'checked_in',
    photo_url: '',
    signature_url: '',
  })
}

function editGuestVisit(guestVisit) {
  editingID.value = guestVisit.id
  Object.assign(form, {
    form_id: filters.form_id,
    guest_name: guestVisit.guest_name || '',
    guest_email: guestVisit.guest_email || '',
    guest_phone: guestVisit.guest_phone || '',
    guest_company: guestVisit.guest_company || '',
    purpose: guestVisit.purpose || '',
    person_to_meet: guestVisit.person_to_meet || '',
    visit_date: guestVisit.visit_date || new Date().toISOString().slice(0, 10),
    status: guestVisit.status || 'checked_in',
    photo_url: guestVisit.photo_url || '',
    signature_url: guestVisit.signature_url || '',
  })
  successMessage.value = ''
  pageError.value = ''
}

function openDetailModal(guestVisit) {
  selectedVisit.value = guestVisit
  isDetailModalOpen.value = true
}

function closeDetailModal() {
  isDetailModalOpen.value = false
}

function visitQueryString() {
  const params = new URLSearchParams()
  Object.entries(filters).forEach(([key, value]) => {
    const trimmed = value.trim()
    if (trimmed !== '') {
      params.set(key, trimmed)
    }
  })
  return params.toString()
}

function resetFilters() {
  Object.assign(filters, {
    q: '',
    status: '',
    form_id: String(route.params.formId || filters.form_id),
    date_from: '',
    date_to: '',
  })
  loadGuestVisits()
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

async function loadGuestForms() {
  const response = await fetch(`${API_BASE_URL}/api/admin/guest-forms`, {
    headers: authHeaders(),
  })
  const data = await response.json().catch(() => ({}))

  if (response.status === 401) {
    throw new Error('invalid or expired token')
  }

  if (!response.ok) {
    throw new Error(data.error || data.message || 'Gagal memuat daftar form.')
  }

  guestForms.value = data
  if (!filters.form_id) {
    await router.replace({ name: 'admin-guest-visits' })
    return
  }
  if (!data.some((guestForm) => guestForm.id === filters.form_id)) {
    throw new Error('Form kunjungan tidak ditemukan.')
  }
  if (!form.form_id) {
    form.form_id = filters.form_id
  }
}

async function loadGuestVisitSummary() {
  const response = await fetch(endpoint.value, {
    headers: authHeaders(),
  })
  const data = await response.json().catch(() => ({}))

  if (response.status === 401) {
    throw new Error('invalid or expired token')
  }

  if (!response.ok) {
    throw new Error(data.error || data.message || 'Gagal memuat ringkasan kunjungan.')
  }

  allGuestVisits.value = data
}

async function loadGuestVisits() {
  if (!auth.value?.token) {
    router.replace({ name: 'admin-login' })
    return
  }

  isLoading.value = true
  pageError.value = ''

  try {
    if (guestForms.value.length === 0) {
      await loadGuestForms()
    }
    await loadGuestVisitSummary()

    const queryString = visitQueryString()
    const response = await fetch(`${endpoint.value}${queryString ? `?${queryString}` : ''}`, {
      headers: authHeaders(),
    })
    const data = await response.json().catch(() => ({}))

    if (response.status === 401) {
      throw new Error('invalid or expired token')
    }

    if (!response.ok) {
      throw new Error(data.error || data.message || 'Gagal memuat kunjungan.')
    }

    guestVisits.value = data
    if (selectedVisit.value) {
      selectedVisit.value = data.find((visit) => visit.id === selectedVisit.value.id) || null
    }
  } catch (error) {
    await handleUnauthorized(error.message || 'Gagal memuat kunjungan.')
  } finally {
    isLoading.value = false
  }
}

async function saveGuestVisit() {
  pageError.value = ''
  successMessage.value = ''
  isSaving.value = true

  try {
    const payload = {
      form_id: filters.form_id,
      guest_name: form.guest_name.trim(),
      guest_email: optionalString(form.guest_email),
      guest_phone: form.guest_phone.trim(),
      guest_company: optionalString(form.guest_company),
      purpose: form.purpose.trim(),
      person_to_meet: optionalString(form.person_to_meet),
      visit_date: form.visit_date,
      status: form.status,
      photo_url: optionalString(form.photo_url),
      signature_url: optionalString(form.signature_url),
      metadata: {},
    }

    const response = await fetch(isEditing.value ? `${endpoint.value}/${editingID.value}` : endpoint.value, {
      method: isEditing.value ? 'PUT' : 'POST',
      headers: authHeaders(true),
      body: JSON.stringify(payload),
    })
    const data = await response.json().catch(() => ({}))

    if (!response.ok) {
      throw new Error(data.error || data.message || 'Kunjungan gagal disimpan.')
    }

    successMessage.value = isEditing.value ? 'Kunjungan berhasil diperbarui.' : 'Kunjungan berhasil dibuat.'
    selectedVisit.value = data
    resetForm()
    await loadGuestVisits()
  } catch (error) {
    pageError.value = error.message || 'Kunjungan gagal disimpan.'
    if (isAuthError(pageError.value)) {
      await handleUnauthorized(pageError.value)
    }
  } finally {
    isSaving.value = false
  }
}

async function deleteGuestVisit(guestVisit) {
  const confirmed = window.confirm(`Hapus kunjungan "${guestVisit.guest_name}"?`)
  if (!confirmed) {
    return
  }

  pageError.value = ''
  successMessage.value = ''
  isDeleting.value = guestVisit.id

  try {
    const response = await fetch(`${endpoint.value}/${guestVisit.id}`, {
      method: 'DELETE',
      headers: authHeaders(),
    })

    if (!response.ok) {
      const data = await response.json().catch(() => ({}))
      if (response.status === 401) {
        throw new Error('invalid or expired token')
      }
      throw new Error(data.error || data.message || 'Kunjungan gagal dihapus.')
    }

    successMessage.value = 'Kunjungan berhasil dihapus.'
    if (selectedVisit.value?.id === guestVisit.id) {
      selectedVisit.value = null
    }
    if (editingID.value === guestVisit.id) {
      resetForm()
    }
    await loadGuestVisits()
  } catch (error) {
    pageError.value = error.message || 'Kunjungan gagal dihapus.'
    if (isAuthError(pageError.value)) {
      await handleUnauthorized(pageError.value)
    }
  } finally {
    isDeleting.value = ''
  }
}

onMounted(loadGuestVisits)
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
          <p class="text-uppercase small fw-semibold text-primary mb-2">Kunjungan</p>
          <h2 class="h3 mb-1">{{ selectedForm?.title || 'Detail kunjungan' }}</h2>
          <p class="text-secondary mb-0">Pantau, tambah, dan perbarui data tamu untuk form ini.</p>
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
      <div v-if="successMessage" class="alert alert-success" role="alert">{{ successMessage }}</div>

      <div class="mb-4">
        <RouterLink class="btn btn-light" to="/admin/guest-visits">
          <i class="bi bi-arrow-left me-2"></i>
          Pilih form lain
        </RouterLink>
      </div>

      <section v-if="selectedFormSummary" class="summary-panel mb-4">
        <div class="summary-metrics">
          <div class="metric-card">
            <span class="metric-icon"><i class="bi bi-person-check"></i></span>
            <strong class="metric-value">{{ selectedFormSummary.total }}</strong>
          </div>
          <div class="metric-card">
            <span class="metric-icon"><i class="bi bi-buildings"></i></span>
            <strong class="metric-value">{{ selectedFormSummary.companies.length }}</strong>
          </div>
        </div>

        <div class="summary-companies">
          <h3 class="h6 mb-3">Instansi pada form ini</h3>
          <p v-if="!selectedFormSummary.companies.length" class="text-secondary mb-0">
            Belum ada instansi yang tercatat.
          </p>
          <span v-for="companyName in selectedFormSummary.companies" :key="companyName"
            class="badge text-bg-light border">
            {{ companyName }}
          </span>
        </div>
      </section>

      <section class="content-panel mb-4">
        <div class="row g-3 align-items-end">
          <div class="col-12 col-lg-3">
            <label for="search" class="form-label">Cari tamu</label>
            <input id="search" v-model="filters.q" type="search" class="form-control"
              placeholder="Nama, email, telepon" />
          </div>
          <div class="col-12 col-sm-6 col-lg-2">
            <label for="status" class="form-label">Status</label>
            <select id="status" v-model="filters.status" class="form-select">
              <option v-for="option in statusOptions" :key="option.value" :value="option.value">
                {{ option.label }}
              </option>
            </select>
          </div>
          <div class="col-6 col-lg-2">
            <label for="dateFrom" class="form-label">Dari</label>
            <input id="dateFrom" v-model="filters.date_from" type="date" class="form-control" />
          </div>
          <div class="col-6 col-lg-2">
            <label for="dateTo" class="form-label">Sampai</label>
            <input id="dateTo" v-model="filters.date_to" type="date" class="form-control" />
          </div>
          <div class="col-12 col-lg-3 d-flex justify-content-lg-end gap-2">
            <button type="button" class="btn btn-light" @click="resetFilters">
              <i class="bi bi-x-circle me-1"></i>
              Reset
            </button>
            <button type="button" class="btn btn-primary" @click="loadGuestVisits">
              <i class="bi bi-funnel me-1"></i>
              Terapkan filter
            </button>
          </div>
        </div>
      </section>

      <div class="row g-4">
        <div class="col-12 col-xl-7">
          <section class="content-panel">
            <div class="d-flex align-items-center justify-content-between gap-3 mb-3">
              <div>
                <h3 class="h5 mb-1">Daftar kunjungan</h3>
                <p class="text-secondary mb-0">
                  {{ selectedForm ? selectedForm.title : 'Pilih form untuk melihat kunjungan.' }}
                </p>
              </div>
              <button type="button" class="btn btn-outline-primary btn-sm" @click="loadGuestVisits">
                <i class="bi bi-arrow-clockwise me-1"></i>
                Refresh
              </button>
            </div>

            <div v-if="isLoading" class="alert alert-light border mb-0" role="status">
              Memuat kunjungan...
            </div>

            <div v-else-if="guestVisits.length === 0" class="empty-state">
              <h4 class="h6 mb-2">Belum ada kunjungan</h4>
              <p class="text-secondary mb-0">Data akan tampil setelah tamu mengisi form publik.</p>
            </div>

            <div v-else class="table-responsive">
              <table class="table align-middle mb-0">
                <thead>
                  <tr>
                    <th>Tamu</th>
                    <th>Form</th>
                    <th>Check-in</th>
                    <th class="text-end">Aksi</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="visit in guestVisits" :key="visit.id">
                    <td>
                      <p class="fw-semibold mb-1">{{ visit.guest_name }}</p>
                      <p class="small text-secondary mb-0">{{ visit.guest_phone }}</p>
                    </td>
                    <td>{{ visit.form?.title || '-' }}</td>
                    <td>{{ formatDateTime(visit.check_in_at) }}</td>
                    <td class="text-end">
                      <div class="d-inline-flex gap-2">
                        <button type="button" class="btn btn-outline-secondary btn-sm icon-btn" title="Lihat detail"
                          aria-label="Lihat detail" @click="openDetailModal(visit)">
                          <i class="bi bi-eye"></i>
                        </button>
                        <button type="button" class="btn btn-outline-primary btn-sm icon-btn" title="Edit kunjungan"
                          aria-label="Edit kunjungan" @click="editGuestVisit(visit)">
                          <i class="bi bi-pencil-square"></i>
                        </button>
                        <button type="button" class="btn btn-outline-danger btn-sm icon-btn"
                          :disabled="isDeleting === visit.id" title="Hapus kunjungan" aria-label="Hapus kunjungan"
                          @click="deleteGuestVisit(visit)">
                          <span v-if="isDeleting === visit.id" class="spinner-border spinner-border-sm"
                            aria-hidden="true"></span>
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
          <form class="content-panel mb-4" @submit.prevent="saveGuestVisit">
            <div class="d-flex align-items-start justify-content-between gap-3 mb-3">
              <div>
                <h3 class="h5 mb-1">{{ isEditing ? 'Edit kunjungan' : 'Tambah kunjungan' }}</h3>
                <p class="text-secondary mb-0">Nama, nomor telepon, dan keperluan wajib diisi.</p>
              </div>
              <button v-if="isEditing" type="button" class="btn btn-light btn-sm" @click="resetForm">
                <i class="bi bi-x-lg me-1"></i>
                Batal
              </button>
            </div>

            <div class="row g-3">
              <div class="col-12">
                <label class="form-label">Form</label>
                <div class="readonly-field">{{ selectedForm?.title || 'Memuat form...' }}</div>
              </div>

              <div class="col-12 col-md-6">
                <label for="guestName" class="form-label">Nama Tamu *</label>
                <input id="guestName" v-model="form.guest_name" type="text" class="form-control" required />
              </div>
              <div class="col-12 col-md-6">
                <label for="guestPhone" class="form-label">Nomor Telepon *</label>
                <input id="guestPhone" v-model="form.guest_phone" type="tel" class="form-control" required />
              </div>
              <div class="col-12 col-md-6">
                <label for="guestEmail" class="form-label">Email</label>
                <input id="guestEmail" v-model="form.guest_email" type="email" class="form-control" />
              </div>
              <div class="col-12 col-md-6">
                <label for="guestCompany" class="form-label">Instansi</label>
                <input id="guestCompany" v-model="form.guest_company" type="text" class="form-control" />
              </div>
              <div class="col-12 col-md-6">
                <label for="personToMeet" class="form-label">Bertemu Dengan</label>
                <input id="personToMeet" v-model="form.person_to_meet" type="text" class="form-control" />
              </div>
              <div class="col-12 col-md-6">
                <label for="visitDate" class="form-label">Tanggal Kunjungan</label>
                <input id="visitDate" v-model="form.visit_date" type="date" class="form-control" />
              </div>
              <div class="col-12 col-md-6">
                <label for="visitStatus" class="form-label">Status</label>
                <select id="visitStatus" v-model="form.status" class="form-select">
                  <option value="checked_in">Checked in</option>
                  <option value="checked_out">Checked out</option>
                  <option value="cancelled">Cancelled</option>
                </select>
              </div>
              <div class="col-12 col-md-6">
                <label for="photoUrl" class="form-label">Foto URL</label>
                <input id="photoUrl" v-model="form.photo_url" type="url" class="form-control" />
              </div>
              <div class="col-12">
                <label for="signatureUrl" class="form-label">Tanda Tangan URL</label>
                <input id="signatureUrl" v-model="form.signature_url" type="url" class="form-control" />
              </div>
              <div class="col-12">
                <label for="purpose" class="form-label">Keperluan *</label>
                <textarea id="purpose" v-model="form.purpose" class="form-control" rows="3" required></textarea>
              </div>
            </div>

            <div class="d-grid mt-4">
              <button type="submit" class="btn btn-primary py-3" :disabled="isSaving || guestForms.length === 0">
                <span v-if="isSaving" class="spinner-border spinner-border-sm me-2" aria-hidden="true"></span>
                <i v-else class="bi bi-save me-2"></i>
                {{ isSaving ? 'Menyimpan...' : isEditing ? 'Simpan Perubahan' : 'Tambah Kunjungan' }}
              </button>
            </div>
          </form>

        </div>
      </div>
    </section>

    <div v-if="isDetailModalOpen && selectedVisit" class="modal fade show detail-modal" role="dialog" aria-modal="true"
      aria-labelledby="visitDetailTitle" tabindex="-1" @click.self="closeDetailModal">
      <div class="modal-dialog modal-dialog-centered modal-lg">
        <div class="modal-content">
          <div class="modal-header">
            <div>
              <p class="text-uppercase small fw-semibold text-primary mb-1">Detail kunjungan</p>
              <h3 id="visitDetailTitle" class="modal-title h5">{{ selectedVisit.guest_name }}</h3>
            </div>
            <button type="button" class="btn-close" aria-label="Tutup" @click="closeDetailModal"></button>
          </div>
          <div class="modal-body">
            <div class="row g-3">
              <div class="col-12 col-md-6">
                <div class="detail-box">
                  <span>Form</span>
                  <strong>{{ selectedVisit.form?.title || '-' }}</strong>
                </div>
              </div>
              <div class="col-12 col-md-6">
                <div class="detail-box">
                  <span>Status</span>
                  <strong>{{ statusLabel(selectedVisit.status) }}</strong>
                </div>
              </div>
              <div class="col-12 col-md-6">
                <div class="detail-box">
                  <span>Telepon</span>
                  <strong>{{ selectedVisit.guest_phone || '-' }}</strong>
                </div>
              </div>
              <div class="col-12 col-md-6">
                <div class="detail-box">
                  <span>Email</span>
                  <strong>{{ selectedVisit.guest_email || '-' }}</strong>
                </div>
              </div>
              <div class="col-12 col-md-6">
                <div class="detail-box">
                  <span>Instansi</span>
                  <strong>{{ selectedVisit.guest_company || '-' }}</strong>
                </div>
              </div>
              <div class="col-12 col-md-6">
                <div class="detail-box">
                  <span>Bertemu Dengan</span>
                  <strong>{{ selectedVisit.person_to_meet || '-' }}</strong>
                </div>
              </div>
              <div class="col-12 col-md-6">
                <div class="detail-box">
                  <span>Check-in</span>
                  <strong>{{ formatDateTime(selectedVisit.check_in_at) }}</strong>
                </div>
              </div>
              <div class="col-12 col-md-6">
                <div class="detail-box">
                  <span>Check-out</span>
                  <strong>{{ formatDateTime(selectedVisit.check_out_at) }}</strong>
                </div>
              </div>
              <div class="col-12">
                <div class="detail-box">
                  <span>Keperluan</span>
                  <strong>{{ selectedVisit.purpose }}</strong>
                </div>
              </div>
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-light" @click="closeDetailModal">Tutup</button>
            <button type="button" class="btn btn-primary" @click="editGuestVisit(selectedVisit); closeDetailModal()">
              <i class="bi bi-pencil-square me-2"></i>
              Edit kunjungan
            </button>
          </div>
        </div>
      </div>
    </div>
    <div v-if="isDetailModalOpen && selectedVisit" class="modal-backdrop fade show"></div>
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
.content-panel,
.summary-panel {
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

.summary-panel {
  display: grid;
  gap: 18px;
  grid-template-columns: minmax(280px, 420px) 1fr;
  padding: 20px;
}

.summary-metrics {
  display: grid;
  gap: 12px;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.metric-card {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 14px;
  border: 1px solid #e2e8f0;
  border-radius: 16px;
  background: #f8fafc;
  padding: 18px;
}

.metric-icon {
  display: grid;
  width: 46px;
  height: 46px;
  flex: 0 0 auto;
  place-items: center;
  border-radius: 14px;
  background: #eaf2ff;
  color: #0d6efd;
  font-size: 1.35rem;
}

.metric-value {
  color: #111827;
  font-size: 2rem;
  line-height: 1;
}

.summary-companies {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-content: flex-start;
  border-left: 1px solid #e2e8f0;
  padding-left: 18px;
}

.summary-companies h3,
.summary-companies p {
  flex: 0 0 100%;
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

.form-label {
  color: #334155;
  font-weight: 650;
}

.form-control,
.form-select {
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
  padding: 0.76rem 0.9rem;
}

.form-control:focus,
.form-select:focus {
  border-color: #86b7fe;
  box-shadow: 0 0 0 0.25rem rgba(13, 110, 253, 0.12);
}

.btn {
  border-radius: 12px;
  font-weight: 700;
}

.icon-btn {
  display: inline-grid;
  width: 34px;
  height: 34px;
  place-items: center;
  padding: 0;
}

.detail-modal {
  display: block;
}

.detail-box {
  border: 1px solid #e2e8f0;
  border-radius: 14px;
  background: #f8fafc;
  padding: 14px 16px;
}

.detail-box span {
  display: block;
  color: #64748b;
  font-size: 0.78rem;
  font-weight: 700;
  margin-bottom: 4px;
  text-transform: uppercase;
}

.detail-box strong {
  color: #111827;
  overflow-wrap: anywhere;
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

  .summary-panel {
    grid-template-columns: 1fr;
  }

  .summary-companies {
    border-left: 0;
    border-top: 1px solid #e2e8f0;
    padding-left: 0;
    padding-top: 18px;
  }
}

@media (max-width: 575.98px) {

  .admin-content,
  .sidebar {
    padding: 18px;
  }

  .summary-metrics {
    grid-template-columns: 1fr;
  }
}
</style>
