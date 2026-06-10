<script setup>
import { RouterLink } from 'vue-router'

defineProps({
  user: {
    type: Object,
    default: null,
  },
  company: {
    type: Object,
    default: null,
  },
  activeMenu: {
    type: String,
    default: 'dashboard',
  },
})

defineEmits(['logout'])
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
        <RouterLink class="nav-link" :class="{ active: activeMenu === 'dashboard' }" to="/admin">
          <i class="bi bi-speedometer2"></i>
          Dashboard
        </RouterLink>
        <RouterLink class="nav-link" :class="{ active: activeMenu === 'guest-visits' }" to="/admin/guest-visits">
          <i class="bi bi-people"></i>
          Kunjungan
        </RouterLink>
        <RouterLink
          v-if="user?.role === 'admin'"
          class="nav-link"
          :class="{ active: activeMenu === 'guest-forms' }"
          to="/admin/guest-forms"
        >
          <i class="bi bi-ui-checks-grid"></i>
          Form Public
        </RouterLink>
        <RouterLink class="nav-link" :class="{ active: activeMenu === 'settings' }" to="/admin/settings">
          <i class="bi bi-gear"></i>
          Pengaturan
        </RouterLink>
      </nav>

      <div class="sidebar-footer mt-auto">
        <p class="small text-secondary mb-1">Instansi</p>
        <p class="fw-semibold mb-3">{{ company?.name || 'Memuat...' }}</p>
        <button type="button" class="btn btn-outline-secondary w-100" @click="$emit('logout')">
          <i class="bi bi-box-arrow-right me-2"></i>
          Keluar
        </button>
      </div>
    </aside>

    <section class="admin-content">
      <header class="admin-header">
        <slot name="header"></slot>
        <div class="user-chip">
          <span class="avatar">{{ (user?.name || 'A').slice(0, 1) }}</span>
          <div>
            <p class="fw-semibold mb-0">{{ user?.name || 'Admin' }}</p>
            <p class="small text-secondary mb-0">{{ user?.email || 'admin' }}</p>
          </div>
        </div>
      </header>

      <slot></slot>
    </section>
  </main>
</template>

<style>
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
.avatar,
.settings-icon {
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
