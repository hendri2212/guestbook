<script setup>
import QRCode from 'qrcode'
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import AdminLayout from '@/components/AdminLayout.vue'
import { API_BASE_URL } from '@/config/api'

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
const barcodeForm = ref(null)
const qrCodes = ref({})

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
const canManageGuestForms = computed(() => ['owner', 'admin'].includes(user.value?.role))
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

function redirectToLogin() {
  clearSession()
  router.replace({ name: 'admin-login' })
}

function isAuthError(message) {
  const normalized = message.toLowerCase()
  return normalized.includes('unauthorized') || normalized.includes('invalid') || normalized.includes('expired')
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

function publicFormUrl(publicSlug) {
  return new URL(publicFormPath(publicSlug), window.location.origin).href
}

async function generateQrCode(guestForm) {
  if (!guestForm?.public_slug || qrCodes.value[guestForm.id]) {
    return
  }

  qrCodes.value = {
    ...qrCodes.value,
    [guestForm.id]: await QRCode.toDataURL(publicFormUrl(guestForm.public_slug), {
      errorCorrectionLevel: 'M',
      margin: 2,
      width: 320,
      color: {
        dark: '#0f172a',
        light: '#ffffff',
      },
    }),
  }
}

async function generateQrCodes(forms) {
  await Promise.all(forms.map((guestForm) => generateQrCode(guestForm)))
}

async function openBarcodeModal(guestForm) {
  barcodeForm.value = guestForm
  await generateQrCode(guestForm)
}

function closeBarcodeModal() {
  barcodeForm.value = null
}

function printBarcode() {
  if (!barcodeForm.value || !qrCodes.value[barcodeForm.value.id]) {
    return
  }

  const printWindow = window.open('', '_blank', 'width=720,height=900')
  if (!printWindow) {
    window.print()
    return
  }

  const formTitle = escapeHtml(barcodeForm.value.title || 'Form Public')
  const companyName = escapeHtml(company.value?.name || 'Guestbook')
  const publicUrl = escapeHtml(publicFormUrl(barcodeForm.value.public_slug))
  const qrCode = qrCodes.value[barcodeForm.value.id]

  printWindow.document.write(`
    <!doctype html>
    <html>
      <head>
        <title>Barcode ${formTitle}</title>
        <style>
          @page {
            size: A4 portrait;
            margin: 16mm;
          }

          * {
            box-sizing: border-box;
          }

          html,
          body {
            width: 100%;
            height: 100%;
            margin: 0;
          }

          body {
            display: grid;
            place-items: center;
            color: #111827;
            font-family: Arial, sans-serif;
          }

          .sheet {
            display: grid;
            width: 100%;
            max-width: 150mm;
            break-inside: avoid;
            gap: 7mm;
            justify-items: center;
            page-break-inside: avoid;
            text-align: center;
          }

          .company {
            color: #64748b;
            font-size: 12pt;
            margin: 0;
          }

          h1 {
            font-size: 20pt;
            margin: 0;
          }

          .hint {
            color: #64748b;
            font-size: 12pt;
            margin: 0;
          }

          .qr {
            width: 90mm;
            height: 90mm;
          }

          .url {
            color: #334155;
            font-size: 10pt;
            font-weight: 700;
            margin: 0;
            overflow-wrap: anywhere;
          }
        </style>
      </head>
      <body>
        <main class="sheet">
          <p class="company">${companyName}</p>
          <h1>${formTitle}</h1>
          <p class="hint">Scan barcode untuk mengisi form buku tamu.</p>
          <img class="qr" src="${qrCode}" alt="Barcode ${formTitle}" />
          <p class="url">${publicUrl}</p>
        </main>
        <script>
          window.addEventListener('load', () => {
            window.print();
            window.close();
          });
        <\/script>
      </body>
    </html>
  `)
  printWindow.document.close()
}

function escapeHtml(value) {
  return String(value)
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
    .replaceAll("'", '&#039;')
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

    if (response.status === 401) {
      redirectToLogin()
      return
    }

    if (!response.ok) {
      throw new Error(data.error || data.message || 'Gagal memuat guest form.')
    }

    guestForms.value = data
    await generateQrCodes(data)
  } catch (error) {
    pageError.value = error.message || 'Gagal memuat guest form.'
    if (isAuthError(pageError.value)) {
      redirectToLogin()
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

    if (response.status === 401) {
      redirectToLogin()
      return
    }

    if (!response.ok) {
      throw new Error(data.error || data.message || 'Guest form gagal disimpan.')
    }

    successMessage.value = isEditing.value ? 'Guest form berhasil diperbarui.' : 'Guest form berhasil dibuat.'
    resetForm()
    await loadGuestForms()
  } catch (error) {
    pageError.value = error.message || 'Guest form gagal disimpan.'
    if (isAuthError(pageError.value)) {
      redirectToLogin()
    }
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
      if (response.status === 401) {
        redirectToLogin()
        return
      }
      throw new Error(data.error || data.message || 'Guest form gagal dihapus.')
    }

    successMessage.value = 'Guest form berhasil dihapus.'
    if (editingID.value === guestForm.id) {
      resetForm()
    }
    await loadGuestForms()
  } catch (error) {
    pageError.value = error.message || 'Guest form gagal dihapus.'
    if (isAuthError(pageError.value)) {
      redirectToLogin()
    }
  } finally {
    isDeleting.value = ''
  }
}

onMounted(loadGuestForms)
</script>

<template>
  <AdminLayout :user="user" :company="company" active-menu="guest-forms" @logout="logout">
      <template #header>
        <div>
          <p class="text-uppercase small fw-semibold text-primary mb-2">Form Public</p>
          <h2 class="h3 mb-1">Management guest form</h2>
          <p class="text-secondary mb-0">
            Kelola form check-in tamu untuk {{ company?.name || 'instansi' }}.
          </p>
        </div>
      </template>

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
                <i class="bi bi-arrow-clockwise me-1"></i>
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
                        <a
                          class="btn btn-outline-secondary btn-sm icon-btn"
                          :href="publicFormUrl(guestForm.public_slug)"
                          target="_blank"
                          rel="noopener noreferrer"
                          title="Buka form public"
                          aria-label="Buka form public"
                        >
                          <i class="bi bi-box-arrow-up-right"></i>
                        </a>
                        <button
                          type="button"
                          class="btn btn-outline-success btn-sm icon-btn"
                          title="Barcode form"
                          aria-label="Barcode form"
                          @click="openBarcodeModal(guestForm)"
                        >
                          <i class="bi bi-qr-code"></i>
                        </button>
                        <button
                          type="button"
                          class="btn btn-outline-primary btn-sm icon-btn"
                          title="Edit form"
                          aria-label="Edit form"
                          @click="editGuestForm(guestForm)"
                        >
                          <i class="bi bi-pencil-square"></i>
                        </button>
                        <button
                          type="button"
                          class="btn btn-outline-danger btn-sm icon-btn"
                          :disabled="isDeleting === guestForm.id"
                          title="Hapus form"
                          aria-label="Hapus form"
                          @click="deleteGuestForm(guestForm)"
                        >
                          <span
                            v-if="isDeleting === guestForm.id"
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
          <form class="content-panel" @submit.prevent="saveGuestForm">
            <div class="d-flex align-items-start justify-content-between gap-3 mb-3">
              <div>
                <h3 class="h5 mb-1">{{ isEditing ? 'Edit guest form' : 'Buat guest form' }}</h3>
                <p class="text-secondary mb-0">Slug harus huruf kecil, angka, dan tanda hubung.</p>
              </div>
              <button v-if="isEditing" type="button" class="btn btn-light btn-sm" @click="resetForm">
                <i class="bi bi-x-lg me-1"></i>
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
                <i v-else class="bi bi-save me-2"></i>
                {{ isSaving ? 'Menyimpan...' : isEditing ? 'Simpan Perubahan' : 'Buat Guest Form' }}
              </button>
            </div>
          </form>
        </div>
      </div>

      <div
        v-if="barcodeForm"
        class="modal fade show barcode-modal"
        role="dialog"
        aria-modal="true"
        aria-labelledby="barcodeTitle"
        tabindex="-1"
        @click.self="closeBarcodeModal"
      >
        <div class="modal-dialog modal-dialog-centered">
          <div class="modal-content">
            <div class="modal-header">
              <div>
                <p class="text-uppercase small fw-semibold text-primary mb-1">Barcode Form Public</p>
                <h3 id="barcodeTitle" class="modal-title h5">{{ barcodeForm.title }}</h3>
              </div>
              <button type="button" class="btn-close no-print" aria-label="Tutup" @click="closeBarcodeModal"></button>
            </div>
            <div class="modal-body">
              <div class="barcode-print-area">
                <div class="print-brand">
                  <p class="small text-secondary mb-1">{{ company?.name || 'Guestbook' }}</p>
                  <h4 class="h5 mb-1">{{ barcodeForm.title }}</h4>
                  <p class="text-secondary mb-0">Scan barcode untuk mengisi form buku tamu.</p>
                </div>

                <div class="qr-frame">
                  <img
                    v-if="qrCodes[barcodeForm.id]"
                    :src="qrCodes[barcodeForm.id]"
                    :alt="`Barcode ${barcodeForm.title}`"
                  />
                  <div v-else class="alert alert-light border mb-0" role="status">Membuat barcode...</div>
                </div>

                <p class="public-url mb-0">{{ publicFormUrl(barcodeForm.public_slug) }}</p>
              </div>
            </div>
            <div class="modal-footer no-print">
              <button type="button" class="btn btn-light" @click="closeBarcodeModal">Tutup</button>
              <button type="button" class="btn btn-primary" @click="printBarcode">
                <i class="bi bi-printer me-2"></i>
                Cetak Barcode
              </button>
            </div>
          </div>
        </div>
      </div>
      <div v-if="barcodeForm" class="modal-backdrop fade show no-print"></div>
  </AdminLayout>
</template>

<style scoped>
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

.icon-btn {
  display: inline-grid;
  width: 34px;
  height: 34px;
  place-items: center;
  padding: 0;
}

.barcode-modal {
  display: block;
}

.barcode-print-area {
  display: grid;
  gap: 18px;
  justify-items: center;
  text-align: center;
}

.qr-frame {
  display: grid;
  width: min(100%, 340px);
  min-height: 340px;
  place-items: center;
  border: 1px solid #e2e8f0;
  border-radius: 18px;
  background: #ffffff;
  padding: 10px;
}

.qr-frame img {
  width: 100%;
  max-width: 320px;
  height: auto;
}

.public-url {
  color: #334155;
  font-size: 0.9rem;
  font-weight: 650;
  overflow-wrap: anywhere;
}

@media (max-width: 575.98px) {
  .option-grid {
    grid-template-columns: 1fr;
  }
}

@media print {
  @page {
    size: A4 portrait;
    margin: 16mm;
  }

  body * {
    visibility: hidden;
  }

  .barcode-print-area,
  .barcode-print-area * {
    visibility: visible;
  }

  .barcode-print-area {
    position: absolute;
    top: 50%;
    left: 50%;
    align-content: center;
    width: 150mm;
    background: #ffffff;
    color: #111827;
    break-inside: avoid;
    page-break-inside: avoid;
    padding: 0;
    transform: translate(-50%, -50%);
  }

  .qr-frame {
    width: 90mm;
    min-height: 90mm;
  }

  .qr-frame img {
    max-width: 86mm;
  }

  .no-print {
    display: none !important;
  }
}
</style>
