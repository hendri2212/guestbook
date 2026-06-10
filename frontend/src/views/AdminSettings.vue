<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import AdminLayout from '@/components/AdminLayout.vue'

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
  <AdminLayout :user="user" :company="company" active-menu="settings" @logout="logout">
      <template #header>
        <div>
          <p class="text-uppercase small fw-semibold text-primary mb-2">Pengaturan</p>
          <h2 class="h3 mb-1">Pengaturan instansi</h2>
          <p class="text-secondary mb-0">Perbarui informasi dasar instansi yang sedang aktif.</p>
        </div>
      </template>

      <div v-if="isLoading" class="alert alert-light border" role="status">Memuat pengaturan...</div>
      <div v-if="loadError" class="alert alert-warning" role="alert">{{ loadError }}</div>
      <div v-if="successMessage" class="alert alert-success" role="alert">{{ successMessage }}</div>

      <div class="row g-4">
        <div class="col-12">
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

            <div class="d-flex justify-content-end mt-4">
              <button type="submit" class="btn btn-primary px-3" :disabled="!canEditCompany || isSavingCompany">
                <span v-if="isSavingCompany" class="spinner-border spinner-border-sm me-2" aria-hidden="true"></span>
                <i v-else class="bi bi-save me-2"></i>
                {{ isSavingCompany ? 'Menyimpan...' : 'Simpan Data Instansi' }}
              </button>
            </div>
          </form>
        </div>
      </div>
  </AdminLayout>
</template>

<style scoped>
.settings-heading {
  display: flex;
  align-items: center;
  gap: 12px;
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

</style>
