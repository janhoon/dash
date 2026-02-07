<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { GridLayout, GridItem } from 'vue3-grid-layout-next'
import { ArrowLeft, Plus, Trash2, LayoutGrid, AlertCircle } from 'lucide-vue-next'
import type { Dashboard } from '../types/dashboard'
import type { Panel as PanelType } from '../types/panel'
import { getDashboard } from '../api/dashboards'
import { listPanels, deletePanel, updatePanel } from '../api/panels'
import Panel from '../components/Panel.vue'
import PanelEditModal from '../components/PanelEditModal.vue'
import TimeRangePicker from '../components/TimeRangePicker.vue'
import { useTimeRange } from '../composables/useTimeRange'

const route = useRoute()
const router = useRouter()

const dashboard = ref<Dashboard | null>(null)
const panels = ref<PanelType[]>([])
const loading = ref(true)
const error = ref<string | null>(null)

const showPanelModal = ref(false)
const editingPanel = ref<PanelType | null>(null)
const showDeleteConfirm = ref(false)
const deletingPanel = ref<PanelType | null>(null)

const dashboardId = route.params.id as string

// Grid layout configuration
const colNum = 12
const rowHeight = 100

// Time range composable for panel data refresh
const { timeRange, onRefresh, cleanup: cleanupTimeRange, pauseAutoRefresh, resumeAutoRefresh } = useTimeRange()

// Register refresh callback to refetch panel data when time range changes or auto-refresh triggers
let unsubscribeRefresh: (() => void) | null = null

// Convert panels to grid layout format
interface LayoutItem {
  i: string
  x: number
  y: number
  w: number
  h: number
}

const layout = computed<LayoutItem[]>(() => {
  return panels.value.map(panel => ({
    i: panel.id,
    x: panel.grid_pos.x,
    y: panel.grid_pos.y,
    w: panel.grid_pos.w,
    h: panel.grid_pos.h,
  }))
})

// Debounce timer for saving layout changes
let saveLayoutTimeout: number | null = null

async function fetchDashboard() {
  try {
    dashboard.value = await getDashboard(dashboardId)
  } catch {
    error.value = 'Dashboard not found'
    return
  }
}

async function fetchPanels() {
  try {
    panels.value = await listPanels(dashboardId)
  } catch {
    error.value = 'Failed to load panels'
  }
}

async function loadData() {
  loading.value = true
  error.value = null
  await fetchDashboard()
  if (!error.value) {
    await fetchPanels()
  }
  loading.value = false
}

function openAddPanel() {
  editingPanel.value = null
  showPanelModal.value = true
  pauseAutoRefresh()
}

function openEditPanel(panel: PanelType) {
  editingPanel.value = panel
  showPanelModal.value = true
  pauseAutoRefresh()
}

function closePanelModal() {
  showPanelModal.value = false
  editingPanel.value = null
  resumeAutoRefresh()
}

function onPanelSaved() {
  closePanelModal()
  fetchPanels()
}

function confirmDeletePanel(panel: PanelType) {
  deletingPanel.value = panel
  showDeleteConfirm.value = true
}

function cancelDelete() {
  showDeleteConfirm.value = false
  deletingPanel.value = null
}

async function handleDeletePanel() {
  if (!deletingPanel.value) return

  try {
    await deletePanel(deletingPanel.value.id)
    cancelDelete()
    fetchPanels()
  } catch {
    error.value = 'Failed to delete panel'
  }
}

function goBack() {
  router.push('/dashboards')
}

// Handle layout changes (drag/resize)
function onLayoutUpdated(newLayout: LayoutItem[]) {
  // Update local panels state with new positions
  for (const item of newLayout) {
    const panel = panels.value.find(p => p.id === item.i)
    if (panel) {
      const changed =
        panel.grid_pos.x !== item.x ||
        panel.grid_pos.y !== item.y ||
        panel.grid_pos.w !== item.w ||
        panel.grid_pos.h !== item.h

      if (changed) {
        panel.grid_pos.x = item.x
        panel.grid_pos.y = item.y
        panel.grid_pos.w = item.w
        panel.grid_pos.h = item.h
      }
    }
  }

  // Debounce database save
  if (saveLayoutTimeout) {
    clearTimeout(saveLayoutTimeout)
  }
  saveLayoutTimeout = window.setTimeout(() => {
    saveLayoutToDatabase(newLayout)
  }, 500)
}

async function saveLayoutToDatabase(newLayout: LayoutItem[]) {
  for (const item of newLayout) {
    const panel = panels.value.find(p => p.id === item.i)
    if (panel) {
      try {
        await updatePanel(panel.id, {
          grid_pos: {
            x: item.x,
            y: item.y,
            w: item.w,
            h: item.h,
          },
        })
      } catch (e) {
        console.error('Failed to save panel position:', e)
      }
    }
  }
}

function getPanelById(id: string): PanelType | undefined {
  return panels.value.find(p => p.id === id)
}

onMounted(() => {
  loadData()
  // Subscribe to time range changes to refetch panels
  unsubscribeRefresh = onRefresh(() => {
    // In the future, this will refetch panel data with the new time range
    // For now, we log the time range for debugging
    console.log('Time range updated:', timeRange.value)
  })
})

onUnmounted(() => {
  if (unsubscribeRefresh) {
    unsubscribeRefresh()
  }
  if (saveLayoutTimeout) {
    clearTimeout(saveLayoutTimeout)
  }
  cleanupTimeRange()
})
</script>

<template>
  <div class="dashboard-detail">
    <header class="page-header">
      <div class="header-left">
        <button class="btn-back" @click="goBack" title="Back to Dashboards">
          <ArrowLeft :size="20" />
        </button>
        <div class="header-title" v-if="dashboard">
          <h1>{{ dashboard.title }}</h1>
          <p v-if="dashboard.description" class="header-description">
            {{ dashboard.description }}
          </p>
        </div>
      </div>
      <div class="header-right">
        <TimeRangePicker />
        <button class="btn btn-primary" @click="openAddPanel" :disabled="loading">
          <Plus :size="18" />
          <span>Add Panel</span>
        </button>
      </div>
    </header>

    <div v-if="loading" class="state-container">
      <div class="loading-spinner"></div>
      <p>Loading dashboard...</p>
    </div>

    <div v-else-if="error" class="state-container error">
      <AlertCircle :size="48" />
      <p>{{ error }}</p>
      <button class="btn btn-secondary" @click="goBack">Back to Dashboards</button>
    </div>

    <template v-else>
      <div v-if="panels.length === 0" class="state-container empty">
        <div class="empty-icon">
          <LayoutGrid :size="64" />
        </div>
        <h2>No panels yet</h2>
        <p>Add your first panel to start visualizing data</p>
        <button class="btn btn-primary" @click="openAddPanel">
          <Plus :size="18" />
          <span>Add Panel</span>
        </button>
      </div>

      <GridLayout
        v-else
        :layout="layout"
        :col-num="colNum"
        :row-height="rowHeight"
        :margin="[12, 12]"
        :is-draggable="true"
        :is-resizable="true"
        :vertical-compact="true"
        :use-css-transforms="true"
        :responsive="true"
        :breakpoints="{ lg: 1200, md: 996, sm: 768, xs: 480, xxs: 0 }"
        :cols="{ lg: 12, md: 10, sm: 6, xs: 4, xxs: 2 }"
        @layout-updated="onLayoutUpdated"
        class="grid-layout"
      >
        <GridItem
          v-for="item in layout"
          :key="item.i"
          :i="item.i"
          :x="item.x"
          :y="item.y"
          :w="item.w"
          :h="item.h"
          :min-w="2"
          :min-h="2"
          drag-allow-from=".panel-header"
          drag-ignore-from=".panel-actions"
        >
          <Panel
            :panel="getPanelById(item.i)!"
            @edit="openEditPanel"
            @delete="confirmDeletePanel"
          />
        </GridItem>
      </GridLayout>
    </template>

    <PanelEditModal
      v-if="showPanelModal"
      :dashboard-id="dashboardId"
      :panel="editingPanel || undefined"
      @close="closePanelModal"
      @saved="onPanelSaved"
    />

    <div v-if="showDeleteConfirm" class="modal-overlay" @click.self="cancelDelete">
      <div class="modal delete-modal">
        <div class="modal-icon">
          <Trash2 :size="24" />
        </div>
        <h2>Delete Panel</h2>
        <p>Are you sure you want to delete "{{ deletingPanel?.title }}"?</p>
        <p class="warning">This action cannot be undone.</p>
        <div class="modal-actions">
          <button class="btn btn-secondary" @click="cancelDelete">Cancel</button>
          <button class="btn btn-danger" @click="handleDeletePanel">Delete</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.dashboard-detail {
  padding: 1.35rem 1.8rem;
  max-width: 1600px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.15rem;
  padding: 1rem 1.2rem;
  border: 1px solid var(--border-primary);
  border-radius: 14px;
  background: var(--surface-1);
  backdrop-filter: blur(8px);
  box-shadow: var(--shadow-sm);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.btn-back {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 38px;
  height: 38px;
  background: var(--surface-2);
  border: 1px solid var(--border-primary);
  border-radius: 10px;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.2s;
}

.btn-back:hover {
  background: var(--bg-hover);
  border-color: var(--border-secondary);
  color: var(--text-primary);
}

.header-title h1 {
  margin-bottom: 0.25rem;
  font-family: var(--font-mono);
  font-size: 1.05rem;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.header-description {
  color: var(--text-secondary);
  font-size: 0.875rem;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 1rem;
}

/* Buttons */
.btn {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.625rem 1rem;
  border: 1px solid transparent;
  border-radius: 10px;
  font-size: 0.84rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-primary {
  background: linear-gradient(135deg, var(--accent-primary), var(--accent-secondary));
  border-color: rgba(125, 211, 252, 0.4);
  color: white;
}

.btn-primary:hover:not(:disabled) {
  transform: translateY(-1px);
}

.btn-secondary {
  background: var(--bg-tertiary);
  border-color: var(--border-primary);
  color: var(--text-primary);
}

.btn-secondary:hover:not(:disabled) {
  background: var(--bg-hover);
  border-color: var(--border-secondary);
}

.btn-danger {
  background: var(--accent-danger);
  color: white;
}

.btn-danger:hover:not(:disabled) {
  background: var(--accent-danger-hover);
}

/* State Containers */
.state-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 4rem 2rem;
  text-align: center;
  color: var(--text-secondary);
  border: 1px solid var(--border-primary);
  border-radius: 14px;
  background: var(--surface-1);
  min-height: 320px;
}

.state-container.error {
  color: var(--accent-danger);
}

.state-container h2 {
  margin: 1rem 0 0.5rem;
  color: var(--text-primary);
}

.state-container p {
  margin-bottom: 1.5rem;
}

.loading-spinner {
  width: 40px;
  height: 40px;
  border: 3px solid rgba(50, 81, 115, 0.65);
  border-top-color: var(--accent-primary);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin-bottom: 1rem;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.empty-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 120px;
  height: 120px;
  background: linear-gradient(160deg, rgba(56, 189, 248, 0.14), rgba(52, 211, 153, 0.08));
  border: 1px solid var(--border-primary);
  border-radius: 16px;
  color: var(--text-tertiary);
  margin-bottom: 1rem;
}

/* Grid Layout */
.grid-layout {
  min-height: 400px;
  padding-bottom: 0.7rem;
}

/* Modal Styles */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(3, 10, 18, 0.76);
  backdrop-filter: blur(8px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
  animation: fadeIn 0.2s ease-out;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.modal {
  background: var(--surface-1);
  border: 1px solid var(--border-primary);
  border-radius: 14px;
  padding: 2rem;
  width: 100%;
  max-width: 400px;
  animation: slideUp 0.3s ease-out;
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.delete-modal {
  text-align: center;
}

.modal-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  background: rgba(251, 113, 133, 0.15);
  border-radius: 50%;
  color: var(--accent-danger);
  margin-bottom: 1rem;
}

.delete-modal h2 {
  margin-bottom: 0.5rem;
  color: var(--text-primary);
}

.delete-modal p {
  color: var(--text-secondary);
  margin-bottom: 0.5rem;
}

.warning {
  color: var(--accent-danger);
  font-size: 0.875rem;
}

.modal-actions {
  display: flex;
  justify-content: center;
  gap: 0.75rem;
  margin-top: 1.5rem;
}
</style>

<style>
/* vue-grid-layout global styles */
.vue-grid-layout {
  background: transparent;
}

.vue-grid-item {
  touch-action: none;
}

.vue-grid-item.vue-grid-placeholder {
  background: rgba(56, 189, 248, 0.18);
  border: 2px dashed var(--accent-primary);
  border-radius: 8px;
}

.vue-grid-item > .vue-resizable-handle {
  position: absolute;
  width: 20px;
  height: 20px;
  bottom: 0;
  right: 0;
  cursor: se-resize;
  background: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 6 6' fill='%239eb0ca'%3E%3Cpolygon points='6 0 0 6 6 6'/%3E%3C/svg%3E") no-repeat;
  background-position: bottom right;
  padding: 0 3px 3px 0;
  background-repeat: no-repeat;
  background-origin: content-box;
  box-sizing: border-box;
  z-index: 10;
}

.vue-grid-item.vue-draggable-dragging {
  z-index: 100;
  opacity: 0.9;
}

.vue-grid-item.vue-resizable-resizing {
  z-index: 100;
}

@media (max-width: 900px) {
  .dashboard-detail {
    padding: 0.9rem;
  }

  .page-header {
    flex-direction: column;
    align-items: stretch;
    gap: 0.85rem;
  }

  .header-right {
    justify-content: space-between;
    flex-wrap: wrap;
  }
}
</style>
