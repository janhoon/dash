<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { GridLayout, GridItem } from 'vue-grid-layout'
import type { Dashboard } from '../types/dashboard'
import type { Panel as PanelType, GridPos } from '../types/panel'
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
    <header class="header">
      <div class="header-left">
        <button class="btn btn-back" @click="goBack">
          &larr; Back
        </button>
        <h1 v-if="dashboard">{{ dashboard.title }}</h1>
      </div>
      <div class="header-right">
        <TimeRangePicker />
        <button class="btn btn-primary" @click="openAddPanel" :disabled="loading">
          + Add Panel
        </button>
      </div>
    </header>

    <div v-if="loading" class="loading">Loading dashboard...</div>

    <div v-else-if="error" class="error">{{ error }}</div>

    <template v-else>
      <p v-if="dashboard?.description" class="dashboard-description">
        {{ dashboard.description }}
      </p>

      <div v-if="panels.length === 0" class="empty">
        <p>No panels yet</p>
        <button class="btn btn-primary" @click="openAddPanel">
          Add your first panel
        </button>
      </div>

      <GridLayout
        v-else
        :layout="layout"
        :col-num="colNum"
        :row-height="rowHeight"
        :margin="[10, 10]"
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
        <h2>Delete Panel</h2>
        <p>Are you sure you want to delete "{{ deletingPanel?.title }}"?</p>
        <p class="warning">This action cannot be undone.</p>
        <div class="modal-actions">
          <button class="btn" @click="cancelDelete">Cancel</button>
          <button class="btn btn-danger" @click="handleDeletePanel">Delete</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.dashboard-detail {
  padding: 2rem;
  max-width: 1400px;
  margin: 0 auto;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.header h1 {
  margin: 0;
  color: #2c3e50;
}

.btn {
  padding: 0.5rem 1rem;
  border: 1px solid #ddd;
  border-radius: 4px;
  background: white;
  cursor: pointer;
  font-size: 0.875rem;
}

.btn:hover:not(:disabled) {
  background: #f5f5f5;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-back {
  padding: 0.5rem 0.75rem;
}

.btn-primary {
  background: #3498db;
  border-color: #3498db;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: #2980b9;
}

.btn-danger {
  background: #e74c3c;
  border-color: #e74c3c;
  color: white;
}

.btn-danger:hover {
  background: #c0392b;
}

.dashboard-description {
  color: #666;
  margin: 0 0 1.5rem 0;
}

.loading, .error, .empty {
  text-align: center;
  padding: 3rem;
  color: #666;
}

.error {
  color: #e74c3c;
}

.grid-layout {
  min-height: 300px;
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
}

.modal {
  background: white;
  border-radius: 8px;
  padding: 2rem;
  width: 100%;
  max-width: 400px;
}

.delete-modal h2 {
  margin: 0 0 1rem 0;
  color: #e74c3c;
}

.warning {
  color: #e74c3c;
  font-size: 0.875rem;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
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
  background: rgba(52, 152, 219, 0.2);
  border: 2px dashed #3498db;
  border-radius: 4px;
}

.vue-grid-item > .vue-resizable-handle {
  position: absolute;
  width: 20px;
  height: 20px;
  bottom: 0;
  right: 0;
  cursor: se-resize;
  background: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 6 6' fill='%23999'%3E%3Cpolygon points='6 0 0 6 6 6'/%3E%3C/svg%3E") no-repeat;
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
</style>
