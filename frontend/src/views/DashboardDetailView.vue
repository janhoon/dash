<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import type { Dashboard } from '../types/dashboard'
import type { Panel as PanelType } from '../types/panel'
import { getDashboard } from '../api/dashboards'
import { listPanels, deletePanel } from '../api/panels'
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

// Time range composable for panel data refresh
const { timeRange, onRefresh, cleanup: cleanupTimeRange } = useTimeRange()

// Register refresh callback to refetch panel data when time range changes or auto-refresh triggers
let unsubscribeRefresh: (() => void) | null = null

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
}

function openEditPanel(panel: PanelType) {
  editingPanel.value = panel
  showPanelModal.value = true
}

function closePanelModal() {
  showPanelModal.value = false
  editingPanel.value = null
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

      <div v-else class="panel-grid">
        <div
          v-for="panel in panels"
          :key="panel.id"
          class="panel-wrapper"
          :style="{
            gridColumn: `span ${panel.grid_pos.w}`,
            gridRow: `span ${panel.grid_pos.h}`
          }"
        >
          <Panel
            :panel="panel"
            @edit="openEditPanel"
            @delete="confirmDeletePanel"
          />
        </div>
      </div>
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

.panel-grid {
  display: grid;
  grid-template-columns: repeat(12, 1fr);
  gap: 1rem;
  grid-auto-rows: minmax(100px, auto);
}

.panel-wrapper {
  min-height: 200px;
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
