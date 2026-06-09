<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'
const AUTH_STORAGE_KEY = 'guestbook_admin_auth'
const MANAGED_FIELDS = [
  { name: 'guest_email', label: 'Email', type: 'email', stateKey: 'enable_email' },
  { name: 'person_to_meet', label: 'Bertemu Dengan', type: 'text', stateKey: 'enable_person_to_meet' },
  { name: 'identity_number', label: 'Nomor Identitas', type: 'text', stateKey: 'enable_identity_number' },
  { name: 'department', label: 'Departemen Tujuan', type: 'text', stateKey: 'enable_department' },
]

const router = useRouter()
const auth = ref(readStoredAuth())
const guestForms = ref([])
const isLoading = ref(false)
const isSaving = ref(false)
const isDeleting = ref('')
const pageError = ref('')
const successMessage = ref('')
const editingID = ref(null)

const form = reactive({
  name: '',
  public_slug: '',
  title: '',
  description: '',
  is_active: true,
  require_photo: false,
  require_signature: false,
  enable_email: true,
  enable_person_to_meet: true,
  enable_identity_number: true,
  enable_department: true,
  extra_fields: [],
})

const user = computed(() => auth.value?.user || null)
const company = computed(() => auth.value?.company || null)
const canManageGuestForms = computed(() => user.value?.role === 'admin')
const isEditing = computed(() => Boolean(editingID.value))
const endpoint = computed(() => `${API_BASE_URL}/api/admin/guest-forms`)

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
    'Content-Type': 'application/json',
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

function resetForm() {
  editingID.value = null
  Object.assign(form, {
    name: '',
    public_slug: '',
    title: '',
    description: '',
    is_active: true,
    require_photo: false,
    require_signature: false,
    enable_email: true,
    enable_person_to_meet: true,
    enable_identity_number: true,
    enable_department: true,
    extra_fields: [],
  })
}

function editGuestForm(guestForm) {
  const fields = normalizeFields(guestForm.fields)
  const fieldState = fieldStateFromFields(fields)

  editingID.value = guestForm.id
  Object.assign(form, {
    name: guestForm.name || '',
    public_slug: guestForm.public_slug || '',
    title: guestForm.title || '',
    description: guestForm.description || '',
    is_active: Boolean(guestForm.is_active),
    require_photo: Boolean(guestForm.require_photo),
    require_signature: Boolean(guestForm.require_signature),
    ...fieldState,
    extra_fields: fields.filter((field) => !MANAGED_FIELDS.some((managed) => managed.name === field.name)),
  })
  successMessage.value = ''
  pageError.value = ''
}

function publicFormPath(publicSlug) {
  return `/forms/${publicSlug}`
}

function normalizeFields(rawFields) {
  if (Array.isArray(rawFields)) {
    return rawFields
  }

  if (typeof rawFields === 'string' && rawFields.trim() !== '') {
    try {
      const parsed = JSON.parse(rawFields)
      return Array.isArray(parsed) ? parsed : []
    } catch {
      return []
    }
  }

  return []
}

function fieldStateFromFields(fields) {
  return MANAGED_FIELDS.reduce((state, fieldDefinition) => {
    const existingField = fields.find((field) => field.name === fieldDefinition.name)
    state[fieldDefinition.stateKey] = existingField?.enabled !== false
    return state
  }, {})
}

function buildFields() {
  const managedFields = MANAGED_FIELDS.map((fieldDefinition) => ({
    name: fieldDefinition.name,
    label: fieldDefinition.label,
    type: fieldDefinition.type,
    required: false,
    enabled: Boolean(form[fieldDefinition.stateKey]),
  }))

  return [...managedFields, ...form.extra_fields]
}

async function loadGuestForms() {
  if (!auth.value?.token) {
    router.replace({ name: 'admin-login' })
    return
  }

  if (!canManageGuestForms.value) {
    router.replace({ name: 'admin-dashboard' })
    return
  }

  isLoading.value = true
  pageError.value = ''

  try {
    const response = await fetch(endpoint.value, {
      headers: {
        Authorization: `Bearer ${auth.value.token}`,
      },
    })
    const data = await response.json().catch(() => ({}))

    if (!response.ok) {
      throw new Error(data.error || data.message || 'Gagal memuat guest form.')
    }

    guestForms.value = data
  } catch (error) {
    pageError.value = error.message || 'Gagal memuat guest form.'
    if (pageError.value.toLowerCase().includes('unauthorized')) {
      clearSession()
      router.replace({ name: 'admin-login' })
    }
  } finally {
    isLoading.value = false
  }
}

async function saveGuestForm() {
  pageError.value = ''
  successMessage.value = ''
  isSaving.value = true

  try {
    const payload = {
      name: form.name.trim() || null,
      public_slug: form.public_slug.trim(),
      title: form.title.trim(),
      description: form.description.trim() || null,
      is_active: form.is_active,
      require_photo: form.require_photo,
      require_signature: form.require_signature,
      fields: buildFields(),
    }

    const response = await fetch(isEditing.value ? `${endpoint.value}/${editingID.value}` : endpoint.value, {
      method: isEditing.value ? 'PUT' : 'POST',
      headers: authHeaders(),
      body: JSON.stringify(payload),
    })
    const data = await response.json().catch(() => ({}))

    if (!response.ok) {
      throw new Error(data.error || data.message || 'Guest form gagal disimpan.')
    }

    successMessage.value = isEditing.value ? 'Guest form berhasil diperbarui.' : 'Guest form berhasil dibuat.'
    resetForm()
    await loadGuestForms()
  } catch (error) {
    pageError.value = error.message || 'Guest form gagal disimpan.'
  } finally {
    isSaving.value = false
  }
}

async function deleteGuestForm(guestForm) {
  const confirmed = window.confirm(`Hapus form "${guestForm.title}"?`)
  if (!confirmed) {
    return
  }

  pageError.value = ''
  successMessage.value = ''
  isDeleting.value = guestForm.id

  try {
    const response = await fetch(`${endpoint.value}/${guestForm.id}`, {
      method: 'DELETE',
      headers: {
        Authorization: `Bearer ${auth.value?.token}`,
      },
    })

    if (!response.ok) {
      const data = await response.json().catch(() => ({}))
      throw new Error(data.error || data.message || 'Guest form gagal dihapus.')
    }

    successMessage.value = 'Guest form berhasil dihapus.'
    if (editingID.value === guestForm.id) {
      resetForm()
    }
    await loadGuestForms()
  } catch (error) {
    pageError.value = error.message || 'Guest form gagal dihapus.'
  } finally {
    isDeleting.value = ''
  }
}

onMounted(loadGuestForms)
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
        <RouterLink class="nav-link" to="/admin">Dashboard</RouterLink>
        <a class="nav-link disabled" href="#" aria-disabled="true">Kunjungan</a>
        <RouterLink class="nav-link active" to="/admin/guest-forms">Form Public</RouterLink>
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
          <p class="text-uppercase small fw-semibold text-primary mb-2">Form Public</p>
          <h2 class="h3 mb-1">Management guest form</h2>
          <p class="text-secondary mb-0">
            Kelola form check-in tamu untuk {{ company?.name || 'instansi' }}.
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

      <div v-if="pageError" class="alert alert-danger" role="alert">{{ pageError }}</div>
      <div v-if="successMessage" class="alert alert-success" role="alert">{{ successMessage }}</div>

      <div class="row g-4">
        <div class="col-12 col-xl-7">
          <section class="content-panel">
            <div class="d-flex align-items-center justify-content-between gap-3 mb-3">
              <div>
                <h3 class="h5 mb-1">Daftar guest form</h3>
                <p class="text-secondary mb-0">Form yang aktif dapat diakses oleh tamu melalui URL public.</p>
              </div>
              <button type="button" class="btn btn-outline-primary btn-sm" @click="loadGuestForms">
                Refresh
              </button>
            </div>

            <div v-if="isLoading" class="alert alert-light border mb-0" role="status">
              Memuat guest form...
            </div>

            <div v-else-if="guestForms.length === 0" class="empty-state">
              <h4 class="h6 mb-2">Belum ada guest form</h4>
              <p class="text-secondary mb-0">Buat form pertama menggunakan panel di samping.</p>
            </div>

            <div v-else class="table-responsive">
              <table class="table align-middle mb-0">
                <thead>
                  <tr>
                    <th>Form</th>
                    <th>Slug</th>
                    <th>Status</th>
                    <th class="text-end">Aksi</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="guestForm in guestForms" :key="guestForm.id">
                    <td>
                      <p class="fw-semibold mb-1">{{ guestForm.title }}</p>
                      <p class="small text-secondary mb-0">{{ guestForm.name }}</p>
                    </td>
                    <td>
                      <code>{{ guestForm.public_slug }}</code>
                    </td>
                    <td>
                      <span
                        class="badge rounded-pill"
                        :class="guestForm.is_active ? 'text-bg-success' : 'text-bg-secondary'"
                      >
                        {{ guestForm.is_active ? 'Aktif' : 'Nonaktif' }}
                      </span>
                    </td>
                    <td class="text-end">
                      <div class="d-inline-flex gap-2">
                        <RouterLink
                          class="btn btn-outline-secondary btn-sm"
                          :to="publicFormPath(guestForm.public_slug)"
                        >
                          Buka
                        </RouterLink>
                        <button
                          type="button"
                          class="btn btn-outline-primary btn-sm"
                          @click="editGuestForm(guestForm)"
                        >
                          Edit
                        </button>
                        <button
                          type="button"
                          class="btn btn-outline-danger btn-sm"
                          :disabled="isDeleting === guestForm.id"
                          @click="deleteGuestForm(guestForm)"
                        >
                          {{ isDeleting === guestForm.id ? 'Hapus...' : 'Hapus' }}
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
          <form class="content-panel" @submit.prevent="saveGuestForm">
            <div class="d-flex align-items-start justify-content-between gap-3 mb-3">
              <div>
                <h3 class="h5 mb-1">{{ isEditing ? 'Edit guest form' : 'Buat guest form' }}</h3>
                <p class="text-secondary mb-0">Slug harus huruf kecil, angka, dan tanda hubung.</p>
              </div>
              <button v-if="isEditing" type="button" class="btn btn-light btn-sm" @click="resetForm">
                Batal
              </button>
            </div>

            <div class="row g-3">
              <div class="col-12 col-md-6">
                <label for="formName" class="form-label">Nama Internal</label>
                <input id="formName" v-model="form.name" type="text" class="form-control" placeholder="Form utama" />
              </div>

              <div class="col-12 col-md-6">
                <label for="publicSlug" class="form-label">Public Slug *</label>
                <input
                  id="publicSlug"
                  v-model="form.public_slug"
                  type="text"
                  class="form-control"
                  placeholder="buku-tamu-instansi"
                  required
                />
              </div>

              <div class="col-12">
                <label for="formTitle" class="form-label">Judul Form *</label>
                <input
                  id="formTitle"
                  v-model="form.title"
                  type="text"
                  class="form-control"
                  placeholder="Buku Tamu Instansi"
                  required
                />
              </div>

              <div class="col-12">
                <label for="description" class="form-label">Deskripsi</label>
                <textarea
                  id="description"
                  v-model="form.description"
                  class="form-control"
                  rows="3"
                  placeholder="Deskripsi singkat untuk form public"
                ></textarea>
              </div>

              <div class="col-12">
                <div class="field-settings">
                  <div class="mb-3">
                    <h4 class="h6 mb-1">Field yang ditampilkan</h4>
                    <p class="text-secondary mb-0">Pilih data tambahan yang perlu diisi oleh tamu.</p>
                  </div>

                  <div class="option-grid">
                    <div class="form-check form-switch">
                      <input id="enableEmail" v-model="form.enable_email" class="form-check-input" type="checkbox" />
                      <label class="form-check-label" for="enableEmail">Email</label>
                    </div>
                    <div class="form-check form-switch">
                      <input
                        id="enablePersonToMeet"
                        v-model="form.enable_person_to_meet"
                        class="form-check-input"
                        type="checkbox"
                      />
                      <label class="form-check-label" for="enablePersonToMeet">Bertemu Dengan</label>
                    </div>
                    <div class="form-check form-switch">
                      <input
                        id="enableIdentityNumber"
                        v-model="form.enable_identity_number"
                        class="form-check-input"
                        type="checkbox"
                      />
                      <label class="form-check-label" for="enableIdentityNumber">Nomor Identitas</label>
                    </div>
                    <div class="form-check form-switch">
                      <input
                        id="enableDepartment"
                        v-model="form.enable_department"
                        class="form-check-input"
                        type="checkbox"
                      />
                      <label class="form-check-label" for="enableDepartment">Departemen Tujuan</label>
                    </div>
                    <div class="form-check form-switch">
                      <input id="requirePhoto" v-model="form.require_photo" class="form-check-input" type="checkbox" />
                      <label class="form-check-label" for="requirePhoto">Wajib foto</label>
                    </div>
                    <div class="form-check form-switch">
                      <input
                        id="requireSignature"
                        v-model="form.require_signature"
                        class="form-check-input"
                        type="checkbox"
                      />
                      <label class="form-check-label" for="requireSignature">Wajib tanda tangan</label>
                    </div>
                  </div>
                </div>
              </div>

              <div class="col-12">
                <div class="field-settings">
                  <div class="mb-3">
                    <h4 class="h6 mb-1">Status form</h4>
                    <p class="text-secondary mb-0">Nonaktifkan jika form belum siap digunakan oleh tamu.</p>
                  </div>

                  <div class="option-grid">
                    <div class="form-check form-switch">
                      <input id="isActive" v-model="form.is_active" class="form-check-input" type="checkbox" />
                      <label class="form-check-label" for="isActive">Aktif</label>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <div class="d-grid mt-4">
              <button type="submit" class="btn btn-primary py-3" :disabled="isSaving">
                <span v-if="isSaving" class="spinner-border spinner-border-sm me-2" aria-hidden="true"></span>
                {{ isSaving ? 'Menyimpan...' : isEditing ? 'Simpan Perubahan' : 'Buat Guest Form' }}
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

.option-grid {
  display: grid;
  gap: 12px;
  grid-template-columns: repeat(3, minmax(0, 1fr));
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

  .option-grid {
    grid-template-columns: 1fr;
  }
}
</style>
