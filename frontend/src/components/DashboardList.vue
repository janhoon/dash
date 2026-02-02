<script setup lang="ts">
import { ref, onMounted } from 'vue'
import type { Dashboard } from '../types/dashboard'
import { listDashboards, deleteDashboard } from '../api/dashboards'
import CreateDashboardModal from './CreateDashboardModal.vue'
import EditDashboardModal from './EditDashboardModal.vue'

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

onMounted(fetchDashboards)
</script>

<template>
  <div class="dashboard-list">
    <header class="header">
      <h1>Dashboards</h1>
      <button class="btn btn-primary" @click="openCreateModal">
        + New Dashboard
      </button>
    </header>

    <div v-if="loading" class="loading">Loading dashboards...</div>

    <div v-else-if="error" class="error">{{ error }}</div>

    <div v-else-if="dashboards.length === 0" class="empty">
      <p>No dashboards yet</p>
      <button class="btn btn-primary" @click="openCreateModal">
        Create your first dashboard
      </button>
    </div>

    <div v-else class="dashboard-grid">
      <div
        v-for="dashboard in dashboards"
        :key="dashboard.id"
        class="dashboard-card"
      >
        <div class="card-header">
          <h3>{{ dashboard.title }}</h3>
          <div class="card-actions">
            <button class="btn btn-icon" @click="openEditModal(dashboard)" title="Edit">
              ✏️
            </button>
            <button class="btn btn-icon btn-danger" @click="confirmDelete(dashboard)" title="Delete">
              🗑️
            </button>
          </div>
        </div>
        <p v-if="dashboard.description" class="card-description">
          {{ dashboard.description }}
        </p>
        <div class="card-meta">
          <span>Created: {{ formatDate(dashboard.created_at) }}</span>
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
        <h2>Delete Dashboard</h2>
        <p>Are you sure you want to delete "{{ deletingDashboard?.title }}"?</p>
        <p class="warning">This action cannot be undone.</p>
        <div class="modal-actions">
          <button class="btn" @click="cancelDelete">Cancel</button>
          <button class="btn btn-danger" @click="handleDelete">Delete</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.dashboard-list {
  padding: 2rem;
  max-width: 1200px;
  margin: 0 auto;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
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

.btn:hover {
  background: #f5f5f5;
}

.btn-primary {
  background: #3498db;
  border-color: #3498db;
  color: white;
}

.btn-primary:hover {
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

.btn-icon {
  padding: 0.25rem 0.5rem;
  background: transparent;
  border: none;
}

.loading, .error, .empty {
  text-align: center;
  padding: 3rem;
  color: #666;
}

.error {
  color: #e74c3c;
}

.dashboard-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 1.5rem;
}

.dashboard-card {
  background: white;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  padding: 1.5rem;
  transition: box-shadow 0.2s;
}

.dashboard-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 0.5rem;
}

.card-header h3 {
  margin: 0;
  color: #2c3e50;
  font-size: 1.125rem;
}

.card-actions {
  display: flex;
  gap: 0.25rem;
}

.card-description {
  color: #666;
  font-size: 0.875rem;
  margin: 0 0 1rem 0;
}

.card-meta {
  font-size: 0.75rem;
  color: #999;
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
