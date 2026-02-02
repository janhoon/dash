<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { LayoutDashboard, Settings, Activity, ChevronLeft, ChevronRight, Compass } from 'lucide-vue-next'
import OrganizationDropdown from './OrganizationDropdown.vue'
import CreateOrganizationModal from './CreateOrganizationModal.vue'
import { useOrganization } from '../composables/useOrganization'

const route = useRoute()
const router = useRouter()
const { fetchOrganizations } = useOrganization()

const isExpanded = ref(true)
const showCreateOrgModal = ref(false)

interface NavItem {
  icon: typeof LayoutDashboard
  label: string
  path: string
}

const navItems: NavItem[] = [
  { icon: LayoutDashboard, label: 'Dashboards', path: '/dashboards' },
  { icon: Compass, label: 'Explore', path: '/explore' },
]

const bottomNavItems: NavItem[] = [
  { icon: Settings, label: 'Settings', path: '/settings' },
]

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

defineExpose({ isExpanded })
</script>

<template>
  <aside class="sidebar" :class="{ expanded: isExpanded }">
    <div class="sidebar-header">
      <div class="sidebar-logo">
        <Activity class="logo-icon" :size="24" />
        <span v-if="isExpanded" class="logo-text">Dash</span>
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
          v-for="item in bottomNavItems"
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
  width: 56px;
  min-height: 100vh;
  background: var(--bg-secondary);
  border-right: 1px solid var(--border-primary);
  display: flex;
  flex-direction: column;
  position: fixed;
  left: 0;
  top: 0;
  bottom: 0;
  z-index: 50;
  transition: width 0.2s ease;
}

.sidebar.expanded {
  width: 200px;
}

.sidebar-header {
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 0.5rem;
  border-bottom: 1px solid var(--border-primary);
}

.sidebar-logo {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding-left: 0.25rem;
}

.logo-icon {
  color: var(--accent-primary);
  flex-shrink: 0;
}

.logo-text {
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
}

.toggle-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  background: transparent;
  border: none;
  border-radius: 6px;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.2s;
  flex-shrink: 0;
}

.toggle-btn:hover {
  background: var(--bg-hover);
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
  padding: 0.75rem 0;
}

.nav-main,
.nav-bottom {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.nav-item {
  position: relative;
  height: 40px;
  margin: 0 0.5rem;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0 0.75rem;
  background: transparent;
  border: none;
  border-radius: 8px;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.2s;
}

.sidebar:not(.expanded) .nav-item {
  width: 40px;
  margin: 0 auto;
  padding: 0;
  justify-content: center;
}

.nav-item:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.nav-item.active {
  background: rgba(102, 126, 234, 0.15);
  color: var(--accent-primary);
}

.nav-item.active::before {
  content: '';
  position: absolute;
  left: -8px;
  top: 50%;
  transform: translateY(-50%);
  width: 3px;
  height: 20px;
  background: var(--accent-primary);
  border-radius: 0 3px 3px 0;
}

.sidebar:not(.expanded) .nav-item.active::before {
  left: -4px;
}

.nav-label {
  font-size: 0.875rem;
  font-weight: 500;
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
  background: var(--bg-tertiary);
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
</style>
