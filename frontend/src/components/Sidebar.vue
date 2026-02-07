<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { LayoutDashboard, Settings, Activity, ChevronLeft, ChevronRight, Compass, LogOut, Database } from 'lucide-vue-next'
import OrganizationDropdown from './OrganizationDropdown.vue'
import CreateOrganizationModal from './CreateOrganizationModal.vue'
import { useOrganization } from '../composables/useOrganization'
import { useAuth } from '../composables/useAuth'

const route = useRoute()
const router = useRouter()
const { fetchOrganizations, clearOrganizations, currentOrg } = useOrganization()
const { logout, user } = useAuth()

const isExpanded = ref(typeof window !== 'undefined' ? window.innerWidth > 1100 : true)
const showCreateOrgModal = ref(false)

interface NavItem {
  icon: typeof LayoutDashboard
  label: string
  path: string
}

const navItems: NavItem[] = [
  { icon: LayoutDashboard, label: 'Dashboards', path: '/dashboards' },
  { icon: Compass, label: 'Explore', path: '/explore' },
  { icon: Database, label: 'Data Sources', path: '/datasources' },
]

// Settings path is dynamic based on current organization
const settingsPath = computed(() => {
  if (currentOrg.value) {
    return `/settings/org/${currentOrg.value.id}`
  }
  return null
})

function isActive(path: string): boolean {
  return route.path.startsWith(path)
}

function navigate(path: string) {
  router.push(path)
}

function toggleSidebar() {
  isExpanded.value = !isExpanded.value
}

function handleOrgCreated() {
  showCreateOrgModal.value = false
  fetchOrganizations()
}

async function handleLogout() {
  await logout()
  clearOrganizations()
  router.push('/login')
}

defineExpose({ isExpanded })
</script>

<template>
  <aside class="sidebar" :class="{ expanded: isExpanded }">
    <div class="sidebar-header">
      <div class="sidebar-logo">
        <Activity class="logo-icon" :size="24" />
        <div v-if="isExpanded" class="logo-copy">
          <span class="logo-text">Dash</span>
          <span class="logo-subtext">developer cockpit</span>
        </div>
      </div>
      <button class="toggle-btn" @click="toggleSidebar" :title="isExpanded ? 'Collapse' : 'Expand'">
        <component :is="isExpanded ? ChevronLeft : ChevronRight" :size="16" />
      </button>
    </div>

    <OrganizationDropdown :expanded="isExpanded" @createOrg="showCreateOrgModal = true" />

    <nav class="sidebar-nav">
      <div class="nav-main">
        <button
          v-for="item in navItems"
          :key="item.path"
          class="nav-item"
          :class="{ active: isActive(item.path) }"
          @click="navigate(item.path)"
          :title="isExpanded ? undefined : item.label"
        >
          <component :is="item.icon" :size="20" />
          <span v-if="isExpanded" class="nav-label">{{ item.label }}</span>
          <span v-else class="nav-tooltip">{{ item.label }}</span>
        </button>
      </div>

      <div class="nav-bottom">
        <button
          v-if="settingsPath"
          class="nav-item"
          :class="{ active: isActive('/settings') }"
          @click="navigate(settingsPath)"
          :title="isExpanded ? undefined : 'Settings'"
        >
          <Settings :size="20" />
          <span v-if="isExpanded" class="nav-label">Settings</span>
          <span v-else class="nav-tooltip">Settings</span>
        </button>
        <div v-if="isExpanded && user" class="user-info">
          <span class="user-email">{{ user.email }}</span>
        </div>
        <button
          class="nav-item logout-btn"
          @click="handleLogout"
          :title="isExpanded ? undefined : 'Log out'"
        >
          <LogOut :size="20" />
          <span v-if="isExpanded" class="nav-label">Log out</span>
          <span v-else class="nav-tooltip">Log out</span>
        </button>
      </div>
    </nav>

    <CreateOrganizationModal
      v-if="showCreateOrgModal"
      @close="showCreateOrgModal = false"
      @created="handleOrgCreated"
    />
  </aside>
</template>

<style scoped>
.sidebar {
  width: 64px;
  min-height: 100vh;
  background: linear-gradient(180deg, rgba(12, 21, 34, 0.95), rgba(10, 17, 28, 0.92));
  border-right: 1px solid var(--border-primary);
  display: flex;
  flex-direction: column;
  position: fixed;
  left: 0;
  top: 0;
  bottom: 0;
  z-index: 50;
  transition: width 0.24s ease;
  backdrop-filter: blur(10px);
}

.sidebar.expanded {
  width: 232px;
}

.sidebar::before {
  content: '';
  position: absolute;
  top: 0;
  right: -1px;
  width: 1px;
  height: 100%;
  background: linear-gradient(180deg, transparent, rgba(56, 189, 248, 0.4), transparent);
  pointer-events: none;
}

.sidebar-header {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 0.75rem;
  border-bottom: 1px solid var(--border-primary);
}

.sidebar-logo {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  padding-left: 0.1rem;
}

.logo-icon {
  color: var(--accent-primary);
  flex-shrink: 0;
  padding: 0.35rem;
  border-radius: 10px;
  background: linear-gradient(140deg, rgba(56, 189, 248, 0.24), rgba(52, 211, 153, 0.2));
  box-shadow: inset 0 0 0 1px rgba(56, 189, 248, 0.3);
}

.logo-copy {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.logo-text {
  font-size: 0.95rem;
  font-weight: 700;
  letter-spacing: 0.02em;
  text-transform: uppercase;
  font-family: var(--font-mono);
  color: var(--text-primary);
}

.logo-subtext {
  font-size: 0.64rem;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--text-tertiary);
  white-space: nowrap;
}

.toggle-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 30px;
  background: rgba(20, 35, 54, 0.9);
  border: 1px solid var(--border-primary);
  border-radius: 8px;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.2s ease;
  flex-shrink: 0;
}

.toggle-btn:hover {
  background: var(--bg-hover);
  border-color: var(--border-secondary);
  color: var(--text-primary);
}

.sidebar:not(.expanded) .toggle-btn {
  margin: 0 auto;
}

.sidebar-nav {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  padding: 0.9rem 0;
}

.nav-main,
.nav-bottom {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.nav-item {
  position: relative;
  height: 42px;
  margin: 0 0.6rem;
  display: flex;
  align-items: center;
  gap: 0.7rem;
  padding: 0 0.9rem;
  background: transparent;
  border: 1px solid transparent;
  border-radius: 10px;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.2s ease;
}

.sidebar:not(.expanded) .nav-item {
  width: 44px;
  margin: 0 auto;
  padding: 0;
  justify-content: center;
}

.nav-item:hover {
  background: rgba(31, 49, 73, 0.74);
  border-color: rgba(125, 211, 252, 0.22);
  color: var(--text-primary);
}

.nav-item.active {
  background: linear-gradient(90deg, rgba(56, 189, 248, 0.18), rgba(52, 211, 153, 0.1));
  border-color: rgba(56, 189, 248, 0.34);
  color: #bde9ff;
}

.nav-item.active::before {
  content: '';
  position: absolute;
  left: -5px;
  top: 50%;
  transform: translateY(-50%);
  width: 6px;
  height: 6px;
  background: var(--accent-primary);
  border-radius: 999px;
  box-shadow: 0 0 14px rgba(56, 189, 248, 0.7);
}

.sidebar:not(.expanded) .nav-item.active::before {
  left: -3px;
}

.nav-label {
  font-size: 0.82rem;
  font-weight: 500;
  letter-spacing: 0.01em;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.nav-tooltip {
  position: absolute;
  left: calc(100% + 12px);
  top: 50%;
  transform: translateY(-50%);
  padding: 0.5rem 0.75rem;
  background: rgba(11, 20, 31, 0.96);
  border: 1px solid var(--border-secondary);
  border-radius: 6px;
  font-size: 0.75rem;
  font-weight: 500;
  color: var(--text-primary);
  white-space: nowrap;
  opacity: 0;
  visibility: hidden;
  transition: opacity 0.2s, visibility 0.2s;
  pointer-events: none;
  z-index: 100;
}

.nav-tooltip::before {
  content: '';
  position: absolute;
  right: 100%;
  top: 50%;
  transform: translateY(-50%);
  border: 5px solid transparent;
  border-right-color: var(--border-secondary);
}

.sidebar:not(.expanded) .nav-item:hover .nav-tooltip {
  opacity: 1;
  visibility: visible;
}

.user-info {
  padding: 0.65rem 0.9rem;
  margin: 0.5rem 0.5rem 0;
  border-top: 1px solid var(--border-primary);
  background: rgba(19, 32, 50, 0.5);
  border-radius: 10px;
}

.user-email {
  font-size: 0.72rem;
  font-family: var(--font-mono);
  color: var(--text-tertiary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  display: block;
}

.logout-btn:hover {
  background: rgba(251, 113, 133, 0.15);
  border-color: rgba(251, 113, 133, 0.34);
  color: var(--accent-danger);
}

@media (max-width: 900px) {
  .sidebar.expanded {
    width: 210px;
  }
}
</style>
