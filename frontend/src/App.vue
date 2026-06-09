<script setup>
import { computed, onMounted, reactive, ref } from 'vue'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

const publicSlug = computed(() => {
  const segments = window.location.pathname.split('/').filter(Boolean)
  const formIndex = segments.findIndex((segment) => segment === 'forms')

  if (formIndex >= 0) {
    return segments[formIndex + 1] || null
  }

  return segments[0] || null
})

const hasPublicSlug = computed(() => Boolean(publicSlug.value))

const form = reactive({
  guest_name: '',
  guest_email: '',
  guest_phone: '',
  guest_company: '',
  purpose: '',
  person_to_meet: '',
  identity_number: '',
  department: '',
})

const isSubmitting = ref(false)
const isLoadingForm = ref(false)
const formLoadError = ref('')
const guestForm = ref(null)
const submitError = ref('')
const submitResult = ref(null)

const formEndpoint = computed(
  () => `${API_BASE_URL}/api/public/forms/${encodeURIComponent(publicSlug.value)}`,
)

const endpoint = computed(
  () => `${API_BASE_URL}/api/public/forms/${encodeURIComponent(publicSlug.value)}/visits`,
)

const canShowForm = computed(() => Boolean(hasPublicSlug.value && guestForm.value && !formLoadError.value))

function optionalString(value) {
  const trimmed = value.trim()
  return trimmed === '' ? null : trimmed
}

function resetForm() {
  Object.assign(form, {
    guest_name: '',
    guest_email: '',
    guest_phone: '',
    guest_company: '',
    purpose: '',
    person_to_meet: '',
    identity_number: '',
    department: '',
  })
}

async function submitGuestVisit() {
  if (!canShowForm.value) {
    submitError.value = 'Kode form tidak ditemukan pada URL.'
    return
  }

  submitError.value = ''
  submitResult.value = null
  isSubmitting.value = true

  const payload = {
    guest_name: form.guest_name.trim(),
    guest_email: optionalString(form.guest_email),
    guest_phone: form.guest_phone.trim(),
    guest_company: optionalString(form.guest_company),
    purpose: form.purpose.trim(),
    person_to_meet: optionalString(form.person_to_meet),
    metadata: {
      identity_number: optionalString(form.identity_number),
      department: optionalString(form.department),
    },
  }

  try {
    const response = await fetch(endpoint.value, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(payload),
    })

    const data = await response.json().catch(() => ({}))

    if (!response.ok) {
      throw new Error(data.error || data.message || 'Data kunjungan gagal disimpan.')
    }

    submitResult.value = data
    resetForm()
  } catch (error) {
    submitError.value = error.message || 'Terjadi kesalahan saat mengirim data.'
  } finally {
    isSubmitting.value = false
  }
}

async function loadGuestForm() {
  if (!hasPublicSlug.value) {
    return
  }

  isLoadingForm.value = true
  formLoadError.value = ''
  guestForm.value = null

  try {
    const response = await fetch(formEndpoint.value)
    const data = await response.json().catch(() => ({}))

    if (!response.ok) {
      throw new Error(data.error || data.message || 'Kode form tidak valid.')
    }

    guestForm.value = data
  } catch (error) {
    formLoadError.value = error.message || 'Kode form tidak valid.'
  } finally {
    isLoadingForm.value = false
  }
}

onMounted(loadGuestForm)
</script>

<template>
  <main class="public-page min-vh-100">
    <section class="container py-4 py-lg-5">
      <div class="row justify-content-center">
        <div class="col-12 col-xl-11">
          <div class="topbar d-flex align-items-center justify-content-between gap-3 mb-4">
            <div class="brand d-flex align-items-center gap-3">
              <div class="brand-mark" aria-hidden="true">BT</div>
              <div>
                <p class="brand-label mb-1">Buku Tamu Digital</p>
                <h1 class="h4 mb-0">Form Kunjungan Tamu</h1>
              </div>
            </div>
            <span class="badge rounded-pill text-bg-light border px-3 py-2">Public Form</span>
          </div>

          <div v-if="!hasPublicSlug" class="warning-panel">
            <p class="text-uppercase small fw-semibold text-danger mb-3">Kode form tidak ditemukan</p>
            <h2 class="warning-title mb-3">Halaman buku tamu tidak dapat dibuka dari domain utama.</h2>
            <p class="text-secondary mb-0">
              Gunakan tautan buku tamu yang sudah memuat kode form instansi. Data kunjungan tidak
              akan disimpan sebelum kode form valid tersedia pada URL.
            </p>
          </div>

          <div v-else-if="isLoadingForm" class="warning-panel">
            <p class="text-uppercase small fw-semibold text-primary mb-3">Memeriksa kode form</p>
            <h2 class="warning-title mb-3">Sebentar, kami sedang memastikan tautan buku tamu ini valid.</h2>
            <p class="text-secondary mb-0">Form akan ditampilkan setelah kode form ditemukan.</p>
          </div>

          <div v-else-if="formLoadError" class="warning-panel">
            <p class="text-uppercase small fw-semibold text-danger mb-3">Kode form tidak valid</p>
            <h2 class="warning-title mb-3">Tautan buku tamu ini tidak dapat digunakan.</h2>
            <p class="text-secondary mb-0">
              Kode form <code>{{ publicSlug }}</code> tidak ditemukan atau sudah tidak aktif. Hubungi
              petugas instansi untuk mendapatkan tautan buku tamu yang benar.
            </p>
          </div>

          <div v-else-if="canShowForm" class="row g-4 align-items-stretch">
            <div class="col-12 col-lg-5">
              <div class="intro-panel h-100">
                <p class="text-uppercase small fw-semibold text-primary mb-3">Check-in tamu</p>
                <h2 class="display-title mb-3">Silakan lengkapi data kunjungan Anda.</h2>
                <p class="text-secondary mb-4">
                  Informasi ini akan tersimpan sebagai catatan kunjungan resmi dan membantu tim
                  penerima menindaklanjuti kedatangan Anda dengan lebih cepat.
                </p>

                <div class="info-list">
                  <div class="info-item">
                    <span class="info-number">01</span>
                    <div>
                      <p class="fw-semibold mb-1">Isi identitas</p>
                      <p class="text-secondary mb-0">Nama dan tujuan kunjungan wajib diisi.</p>
                    </div>
                  </div>
                  <div class="info-item">
                    <span class="info-number">02</span>
                    <div>
                      <p class="fw-semibold mb-1">Kirim form</p>
                      <p class="text-secondary mb-0">Data langsung dikirim ke sistem buku tamu.</p>
                    </div>
                  </div>
                </div>

                <div class="slug-box mt-4">
                  <span class="text-secondary small d-block mb-1">Kode form aktif</span>
                  <code>{{ publicSlug }}</code>
                </div>
              </div>
            </div>

            <div class="col-12 col-lg-7">
              <form class="form-panel" @submit.prevent="submitGuestVisit">
                <div class="d-flex align-items-start justify-content-between gap-3 mb-4">
                  <div>
                    <h2 class="h5 mb-1">Data Tamu</h2>
                    <p class="text-secondary mb-0">Pastikan data yang dimasukkan sudah benar.</p>
                  </div>
                  <span class="required-note">* Wajib</span>
                </div>

                <div v-if="submitResult" class="alert alert-success" role="alert">
                  Kunjungan berhasil disimpan. Status:
                  <strong>{{ submitResult.status }}</strong>
                </div>

                <div v-if="submitError" class="alert alert-danger" role="alert">
                  {{ submitError }}
                </div>

                <div class="row g-3">
                  <div class="col-12">
                    <label for="guestName" class="form-label">Nama Lengkap *</label>
                    <input
                      id="guestName"
                      v-model="form.guest_name"
                      type="text"
                      class="form-control form-control-lg"
                      placeholder="Masukkan nama lengkap"
                      maxlength="140"
                      required
                    />
                  </div>

                  <div class="col-12 col-md-6">
                    <label for="guestEmail" class="form-label">Email</label>
                    <input
                      id="guestEmail"
                      v-model="form.guest_email"
                      type="email"
                      class="form-control"
                      placeholder="nama@email.com"
                      maxlength="160"
                    />
                  </div>

                  <div class="col-12 col-md-6">
                    <label for="guestPhone" class="form-label">Nomor Telepon *</label>
                    <input
                      id="guestPhone"
                      v-model="form.guest_phone"
                      type="tel"
                      class="form-control"
                      placeholder="08xxxxxxxxxx"
                      maxlength="40"
                      required
                    />
                  </div>

                  <div class="col-12 col-md-6">
                    <label for="guestCompany" class="form-label">Instansi / Perusahaan</label>
                    <input
                      id="guestCompany"
                      v-model="form.guest_company"
                      type="text"
                      class="form-control"
                      placeholder="Nama instansi"
                      maxlength="160"
                    />
                  </div>

                  <div class="col-12 col-md-6">
                    <label for="personToMeet" class="form-label">Bertemu Dengan</label>
                    <input
                      id="personToMeet"
                      v-model="form.person_to_meet"
                      type="text"
                      class="form-control"
                      placeholder="Nama penerima"
                      maxlength="140"
                    />
                  </div>

                  <div class="col-12 col-md-6">
                    <label for="identityNumber" class="form-label">Nomor Identitas</label>
                    <input
                      id="identityNumber"
                      v-model="form.identity_number"
                      type="text"
                      class="form-control"
                      placeholder="KTP/SIM/Paspor"
                    />
                  </div>

                  <div class="col-12 col-md-6">
                    <label for="department" class="form-label">Departemen Tujuan</label>
                    <input
                      id="department"
                      v-model="form.department"
                      type="text"
                      class="form-control"
                      placeholder="Contoh: Administrasi"
                    />
                  </div>

                  <div class="col-12">
                    <label for="purpose" class="form-label">Keperluan *</label>
                    <textarea
                      id="purpose"
                      v-model="form.purpose"
                      class="form-control"
                      rows="4"
                      placeholder="Jelaskan tujuan kunjungan"
                      required
                    ></textarea>
                  </div>
                </div>

                <div class="d-grid d-sm-flex justify-content-sm-end mt-4">
                  <button type="submit" class="btn btn-primary btn-lg px-4" :disabled="isSubmitting">
                    <span
                      v-if="isSubmitting"
                      class="spinner-border spinner-border-sm me-2"
                      aria-hidden="true"
                    ></span>
                    {{ isSubmitting ? 'Menyimpan...' : 'Simpan Kunjungan' }}
                  </button>
                </div>
              </form>
            </div>
          </div>
        </div>
      </div>
    </section>
  </main>
</template>

<style scoped>
:global(body) {
  margin: 0;
  background: #eef2f6;
  color: #1f2937;
  font-family:
    Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
}

.public-page {
  background:
    linear-gradient(135deg, rgba(255, 255, 255, 0.92), rgba(238, 242, 246, 0.76)),
    radial-gradient(circle at top left, rgba(13, 110, 253, 0.12), transparent 34%),
    #eef2f6;
}

.topbar,
.intro-panel,
.form-panel,
.warning-panel {
  border: 1px solid rgba(148, 163, 184, 0.22);
  background: rgba(255, 255, 255, 0.9);
  box-shadow: 0 24px 70px rgba(15, 23, 42, 0.08);
}

.topbar {
  border-radius: 18px;
  padding: 18px;
}

.brand-mark {
  display: grid;
  width: 48px;
  height: 48px;
  place-items: center;
  border-radius: 14px;
  background: #0d6efd;
  color: #ffffff;
  font-weight: 800;
}

.brand-label,
.required-note {
  color: #64748b;
  font-size: 0.82rem;
}

.intro-panel,
.form-panel,
.warning-panel {
  border-radius: 22px;
  padding: clamp(24px, 4vw, 40px);
}

.warning-panel {
  max-width: 760px;
  margin: 0 auto;
}

.warning-title {
  max-width: 16ch;
  color: #111827;
  font-size: clamp(2rem, 4vw, 3rem);
  font-weight: 750;
  line-height: 1.05;
}

.display-title {
  max-width: 12ch;
  color: #111827;
  font-size: clamp(2rem, 5vw, 3.8rem);
  font-weight: 750;
  line-height: 1;
}

.info-list {
  display: grid;
  gap: 16px;
}

.info-item {
  display: flex;
  gap: 14px;
  padding-top: 16px;
  border-top: 1px solid #e2e8f0;
}

.info-number {
  color: #0d6efd;
  font-weight: 800;
}

.slug-box {
  border-radius: 14px;
  background: #f8fafc;
  padding: 14px 16px;
}

.slug-box code {
  color: #334155;
  overflow-wrap: anywhere;
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

.form-control:focus {
  border-color: #86b7fe;
  box-shadow: 0 0 0 0.25rem rgba(13, 110, 253, 0.12);
}

.btn {
  border-radius: 12px;
  font-weight: 700;
}

@media (max-width: 575.98px) {
  .topbar {
    align-items: flex-start !important;
  }

  .topbar .badge {
    display: none;
  }
}
</style>
