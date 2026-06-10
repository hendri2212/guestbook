<script setup>
import { computed, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'
const DEFAULT_COMPANY_SLUG = import.meta.env.VITE_COMPANY_SLUG || 'instansi-demo'
const AUTH_STORAGE_KEY = 'guestbook_admin_auth'

const router = useRouter()
const isSubmitting = ref(false)
const loginError = ref('')

const form = reactive({
  email: 'admin@instansi-demo.test',
  password: '',
})

const endpoint = computed(() => `${API_BASE_URL}/api/auth/login`)

async function submitLogin() {
  loginError.value = ''
  isSubmitting.value = true

  try {
    const response = await fetch(endpoint.value, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        company_slug: DEFAULT_COMPANY_SLUG,
        email: form.email.trim(),
        password: form.password,
      }),
    })

    const data = await response.json().catch(() => ({}))

    if (!response.ok) {
      throw new Error(data.error || data.message || 'Login gagal. Periksa kembali data Anda.')
    }

    localStorage.setItem(
      AUTH_STORAGE_KEY,
      JSON.stringify({
        token: data.token,
        expires_at: data.expires_at,
        user: data.user,
        company: data.company,
      }),
    )

    router.push({ name: 'admin-dashboard' })
  } catch (error) {
    loginError.value = error.message || 'Login gagal. Periksa kembali data Anda.'
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <main class="login-page min-vh-100">
    <section class="container min-vh-100 d-flex align-items-center py-5">
      <div class="row justify-content-center w-100 g-4">
        <div class="col-12 col-lg-5">
          <div class="login-intro h-100">
            <div class="brand-mark mb-4"><i class="bi bi-journal-check"></i></div>
            <p class="text-uppercase small fw-semibold text-primary mb-3">Admin Panel</p>
            <h1 class="login-title mb-3">Kelola buku tamu dari satu dashboard.</h1>
            <p class="text-secondary mb-0">
              Masuk menggunakan akun admin instansi untuk memantau aktivitas kunjungan, form publik,
              dan operasional resepsionis.
            </p>
          </div>
        </div>

        <div class="col-12 col-lg-5">
          <form class="login-card" @submit.prevent="submitLogin">
            <div class="mb-4">
              <h2 class="h4 mb-1">Masuk Admin</h2>
              <p class="text-secondary mb-0">Gunakan kredensial yang sudah terdaftar.</p>
            </div>

            <div v-if="loginError" class="alert alert-danger" role="alert">
              {{ loginError }}
            </div>

            <div class="mb-3">
              <label for="email" class="form-label">Email</label>
              <input
                id="email"
                v-model="form.email"
                type="email"
                class="form-control"
                placeholder="admin@instansi-demo.test"
                required
              />
            </div>

            <div class="mb-4">
              <label for="password" class="form-label">Password</label>
              <input
                id="password"
                v-model="form.password"
                type="password"
                class="form-control"
                placeholder="Masukkan password"
                autocomplete="current-password"
                required
              />
            </div>

            <button type="submit" class="btn btn-primary w-100 py-3" :disabled="isSubmitting">
              <span
                v-if="isSubmitting"
                class="spinner-border spinner-border-sm me-2"
                aria-hidden="true"
              ></span>
              <i v-else class="bi bi-box-arrow-in-right me-2"></i>
              {{ isSubmitting ? 'Memproses...' : 'Masuk ke Dashboard' }}
            </button>
          </form>
        </div>
      </div>
    </section>
  </main>
</template>

<style scoped>
.login-page {
  background:
    linear-gradient(135deg, rgba(248, 250, 252, 0.96), rgba(226, 232, 240, 0.86)),
    #eef2f6;
}

.login-intro,
.login-card {
  border: 1px solid rgba(148, 163, 184, 0.24);
  border-radius: 22px;
  background: rgba(255, 255, 255, 0.92);
  box-shadow: 0 24px 70px rgba(15, 23, 42, 0.08);
  padding: clamp(28px, 4vw, 44px);
}

.brand-mark {
  display: grid;
  width: 56px;
  height: 56px;
  place-items: center;
  border-radius: 16px;
  background: #0d6efd;
  color: #ffffff;
  font-weight: 800;
}

.login-title {
  max-width: 11ch;
  color: #111827;
  font-size: clamp(2.2rem, 5vw, 4rem);
  font-weight: 760;
  line-height: 1;
}

.form-label {
  color: #334155;
  font-weight: 650;
}

.form-control {
  border-color: #dbe3ef;
  border-radius: 12px;
  padding: 0.82rem 0.95rem;
}

.form-control:focus {
  border-color: #86b7fe;
  box-shadow: 0 0 0 0.25rem rgba(13, 110, 253, 0.12);
}

.btn {
  border-radius: 12px;
  font-weight: 750;
}
</style>
