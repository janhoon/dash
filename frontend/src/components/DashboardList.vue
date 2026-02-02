<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Plus, Pencil, Trash2, LayoutDashboard, AlertCircle } from 'lucide-vue-next'
import type { Dashboard } from '../types/dashboard'
import { listDashboards, deleteDashboard } from '../api/dashboards'
import CreateDashboardModal from './CreateDashboardModal.vue'
import EditDashboardModal from './EditDashboardModal.vue'

const router = useRouter()

const dashboards = ref<Dashboard[]>([])
const loading = ref(true)
const error = ref<string | null>(null)
const showCreateModal = ref(false)
const showEditModal = ref(false)
const editingDashboard = ref<Dashboard | null>(null)
const showDeleteConfirm = ref(false)
const deletingDashboard = ref<Dashboard | null>(null)

async function fetchDashboards() {
  loading.value = true
  error.value = null
  try {
    dashboards.value = await listDashboards()
  } catch (e) {
    error.value = 'Failed to load dashboards'
  } finally {
    loading.value = false
  }
}

function openCreateModal() {
  showCreateModal.value = true
}

function closeCreateModal() {
  showCreateModal.value = false
}

function onDashboardCreated() {
  closeCreateModal()
  fetchDashboards()
}

function openEditModal(dashboard: Dashboard) {
  editingDashboard.value = dashboard
  showEditModal.value = true
}

function closeEditModal() {
  showEditModal.value = false
  editingDashboard.value = null
}

function onDashboardUpdated() {
  closeEditModal()
  fetchDashboards()
}

function confirmDelete(dashboard: Dashboard) {
  deletingDashboard.value = dashboard
  showDeleteConfirm.value = true
}

function cancelDelete() {
  showDeleteConfirm.value = false
  deletingDashboard.value = null
}

async function handleDelete() {
  if (!deletingDashboard.value) return

  try {
    await deleteDashboard(deletingDashboard.value.id)
    cancelDelete()
    fetchDashboards()
  } catch (e) {
    error.value = 'Failed to delete dashboard'
  }
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

function openDashboard(dashboard: Dashboard) {
  router.push(`/dashboards/${dashboard.id}`)
}

onMounted(fetchDashboards)
</script>

<template>
  <div class="dashboard-list">
    <header class="page-header">
      <div class="header-content">
        <h1>Dashboards</h1>
        <p class="header-subtitle">Monitor your metrics and visualize data</p>
      </div>
      <button class="btn btn-primary" @click="openCreateModal">
        <Plus :size="18" />
        <span>New Dashboard</span>
      </button>
    </header>

    <div v-if="loading" class="state-container">
      <div class="loading-spinner"></div>
      <p>Loading dashboards...</p>
    </div>

    <div v-else-if="error" class="state-container error">
      <AlertCircle :size="48" />
      <p>{{ error }}</p>
      <button class="btn btn-secondary" @click="fetchDashboards">Try Again</button>
    </div>

    <div v-else-if="dashboards.length === 0" class="state-container empty">
      <div class="empty-icon">
        <LayoutDashboard :size="64" />
      </div>
      <h2>No dashboards yet</h2>
      <p>Create your first dashboard to start monitoring your metrics</p>
      <button class="btn btn-primary" @click="openCreateModal">
        <Plus :size="18" />
        <span>Create Dashboard</span>
      </button>
    </div>

    <div v-else class="dashboard-grid">
      <div
        v-for="dashboard in dashboards"
        :key="dashboard.id"
        class="dashboard-card"
        @click="openDashboard(dashboard)"
      >
        <div class="card-header">
          <h3>{{ dashboard.title }}</h3>
          <div class="card-actions" @click.stop>
            <button class="btn-icon" @click="openEditModal(dashboard)" title="Edit">
              <Pencil :size="16" />
            </button>
            <button class="btn-icon btn-icon-danger" @click="confirmDelete(dashboard)" title="Delete">
              <Trash2 :size="16" />
            </button>
          </div>
        </div>
        <p v-if="dashboard.description" class="card-description">
          {{ dashboard.description }}
        </p>
        <div class="card-meta">
          <span>Created {{ formatDate(dashboard.created_at) }}</span>
        </div>
      </div>
    </div>

    <CreateDashboardModal
      v-if="showCreateModal"
      @close="closeCreateModal"
      @created="onDashboardCreated"
    />

    <EditDashboardModal
      v-if="showEditModal && editingDashboard"
      :dashboard="editingDashboard"
      @close="closeEditModal"
      @updated="onDashboardUpdated"
    />

    <div v-if="showDeleteConfirm" class="modal-overlay" @click.self="cancelDelete">
      <div class="modal delete-modal">
        <div class="modal-icon">
          <Trash2 :size="24" />
        </div>
        <h2>Delete Dashboard</h2>
        <p>Are you sure you want to delete "{{ deletingDashboard?.title }}"?</p>
        <p class="warning">This action cannot be undone.</p>
        <div class="modal-actions">
          <button class="btn btn-secondary" @click="cancelDelete">Cancel</button>
          <button class="btn btn-danger" @click="handleDelete">Delete</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.dashboard-list {
  padding: 2rem;
  max-width: 1400px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 2rem;
  padding-bottom: 1.5rem;
  border-bottom: 1px solid var(--border-primary);
}

.header-content h1 {
  margin-bottom: 0.25rem;
}

.header-subtitle {
  color: var(--text-secondary);
  font-size: 0.875rem;
}

/* Buttons */
.btn {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.625rem 1rem;
  border: 1px solid transparent;
  border-radius: 6px;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary {
  background: var(--accent-primary);
  color: white;
}

.btn-primary:hover {
  background: var(--accent-primary-hover);
}

.btn-secondary {
  background: var(--bg-tertiary);
  border-color: var(--border-primary);
  color: var(--text-primary);
}

.btn-secondary:hover {
  background: var(--bg-hover);
  border-color: var(--border-secondary);
}

.btn-danger {
  background: var(--accent-danger);
  color: white;
}

.btn-danger:hover {
  background: var(--accent-danger-hover);
}

.btn-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  padding: 0;
  background: transparent;
  border: none;
  border-radius: 6px;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.2s;
}

.btn-icon:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.btn-icon-danger:hover {
  background: rgba(255, 107, 107, 0.15);
  color: var(--accent-danger);
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
  border: 3px solid var(--border-primary);
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
  background: var(--bg-secondary);
  border: 1px solid var(--border-primary);
  border-radius: 16px;
  color: var(--text-tertiary);
  margin-bottom: 1rem;
}

/* Dashboard Grid */
.dashboard-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 1.5rem;
}

.dashboard-card {
  background: var(--bg-secondary);
  border: 1px solid var(--border-primary);
  border-radius: 12px;
  padding: 1.5rem;
  cursor: pointer;
  transition: all 0.2s;
}

.dashboard-card:hover {
  border-color: var(--accent-primary);
  box-shadow: 0 0 0 1px var(--accent-primary), 0 8px 24px rgba(0, 0, 0, 0.3);
  transform: translateY(-2px);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 0.75rem;
}

.card-header h3 {
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-primary);
}

.card-actions {
  display: flex;
  gap: 0.25rem;
  opacity: 0;
  transition: opacity 0.2s;
}

.dashboard-card:hover .card-actions {
  opacity: 1;
}

.card-description {
  color: var(--text-secondary);
  font-size: 0.875rem;
  margin-bottom: 1rem;
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.card-meta {
  font-size: 0.75rem;
  color: var(--text-tertiary);
}

/* Modal Styles */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  backdrop-filter: blur(4px);
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
  background: var(--bg-secondary);
  border: 1px solid var(--border-primary);
  border-radius: 12px;
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
  background: rgba(255, 107, 107, 0.15);
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
