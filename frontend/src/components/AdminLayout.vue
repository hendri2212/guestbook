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
          v-if="['owner', 'admin'].includes(user?.role)"
          class="nav-link"
          :class="{ active: activeMenu === 'guest-forms' }"
          to="/admin/guest-forms"
        >
          <i class="bi bi-ui-checks-grid"></i>
          Form Public
        </RouterLink>
        <RouterLink
          v-if="user?.role === 'owner'"
          class="nav-link"
          :class="{ active: activeMenu === 'users' }"
          to="/admin/users"
        >
          <i class="bi bi-person-gear"></i>
          Management User
        </RouterLink>
        <RouterLink
          v-if="user?.role === 'owner'"
          class="nav-link"
          :class="{ active: activeMenu === 'companies' }"
          to="/admin/companies"
        >
          <i class="bi bi-buildings"></i>
          Management Company
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
        <details class="user-menu">
          <summary class="user-chip">
            <span class="avatar">{{ (user?.name || 'A').slice(0, 1) }}</span>
            <span class="user-chip-body">
              <span class="fw-semibold">{{ user?.name || 'Admin' }}</span>
              <span class="small text-secondary">{{ user?.email || 'admin' }}</span>
            </span>
            <i class="bi bi-chevron-down small text-secondary"></i>
          </summary>
          <div class="user-dropdown">
            <RouterLink class="user-dropdown-item" to="/admin/profile">
              <i class="bi bi-person-circle"></i>
              Profile
            </RouterLink>
            <button type="button" class="user-dropdown-item text-danger" @click="$emit('logout')">
              <i class="bi bi-box-arrow-right"></i>
              Keluar
            </button>
          </div>
        </details>
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

.user-menu,
.user-chip,
.content-panel {
  border: 1px solid rgba(148, 163, 184, 0.24);
  border-radius: 18px;
  background: #ffffff;
  box-shadow: 0 16px 46px rgba(15, 23, 42, 0.06);
}

.user-menu {
  position: relative;
}

.user-chip {
  display: flex;
  align-items: center;
  gap: 12px;
  border: 0;
  border-radius: 18px;
  cursor: pointer;
  list-style: none;
  padding: 12px 16px;
  box-shadow: none;
}

.user-chip::-webkit-details-marker {
  display: none;
}

.user-chip-body {
  display: grid;
  line-height: 1.2;
}

.user-dropdown {
  position: absolute;
  top: calc(100% + 8px);
  right: 0;
  z-index: 20;
  display: grid;
  min-width: 210px;
  border: 1px solid #e2e8f0;
  border-radius: 14px;
  background: #ffffff;
  box-shadow: 0 18px 42px rgba(15, 23, 42, 0.14);
  padding: 8px;
}

.user-dropdown-item {
  display: flex;
  align-items: center;
  gap: 10px;
  border: 0;
  border-radius: 10px;
  background: transparent;
  color: #334155;
  font-weight: 650;
  padding: 10px 12px;
  text-align: left;
  text-decoration: none;
}

.user-dropdown-item:hover {
  background: #f1f5f9;
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

  .user-menu {
    width: 100%;
  }

  .user-dropdown {
    left: 0;
    right: auto;
  }
}

@media (max-width: 575.98px) {
  .admin-content,
  .sidebar {
    padding: 18px;
  }
}
</style>
