<script setup lang="ts">
import { ref, computed } from 'vue'
import { X } from 'lucide-vue-next'
import type { Panel } from '../types/panel'
import { createPanel, updatePanel } from '../api/panels'
import QueryBuilder from './QueryBuilder.vue'

const props = defineProps<{
  panel?: Panel
  dashboardId: string
}>()

const emit = defineEmits<{
  close: []
  saved: []
}>()

const isEditing = computed(() => !!props.panel)

const title = ref(props.panel?.title || '')
const panelType = ref(props.panel?.type || 'line_chart')
// Extract promql from query config, or use empty string
const promqlQuery = ref(
  typeof props.panel?.query?.promql === 'string' ? props.panel.query.promql : ''
)
const loading = ref(false)
const error = ref<string | null>(null)

async function handleSubmit() {
  if (!title.value.trim()) {
    error.value = 'Title is required'
    return
  }

  // Build query config with promql
  const query: Record<string, unknown> | undefined = promqlQuery.value.trim()
    ? { promql: promqlQuery.value.trim() }
    : undefined

  loading.value = true
  error.value = null

  try {
    if (isEditing.value && props.panel) {
      await updatePanel(props.panel.id, {
        title: title.value.trim(),
        type: panelType.value,
        query
      })
    } else {
      await createPanel(props.dashboardId, {
        title: title.value.trim(),
        type: panelType.value,
        grid_pos: { x: 0, y: 0, w: 6, h: 4 },
        query
      })
    }
    emit('saved')
  } catch {
    error.value = isEditing.value ? 'Failed to update panel' : 'Failed to create panel'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="modal-overlay" @click.self="emit('close')">
    <div class="modal">
      <header class="modal-header">
        <h2>{{ isEditing ? 'Edit Panel' : 'Add Panel' }}</h2>
        <button class="btn-close" @click="emit('close')">
          <X :size="20" />
        </button>
      </header>

      <form @submit.prevent="handleSubmit">
        <div class="form-row">
          <div class="form-group">
            <label for="title">Title <span class="required">*</span></label>
            <input
              id="title"
              v-model="title"
              type="text"
              placeholder="Panel title"
              :disabled="loading"
              autocomplete="off"
            />
          </div>

          <div class="form-group form-group-small">
            <label for="type">Panel Type</label>
            <select id="type" v-model="panelType" :disabled="loading">
              <option value="line_chart">Line Chart</option>
              <option value="bar_chart">Bar Chart</option>
              <option value="gauge">Gauge</option>
              <option value="stat">Stat</option>
              <option value="table">Table</option>
            </select>
          </div>
        </div>

        <div class="form-group query-builder-group">
          <label>Query</label>
          <QueryBuilder
            v-model="promqlQuery"
            :disabled="loading"
          />
        </div>

        <div v-if="error" class="error-message">{{ error }}</div>

        <div class="modal-actions">
          <button type="button" class="btn btn-secondary" @click="emit('close')" :disabled="loading">
            Cancel
          </button>
          <button type="submit" class="btn btn-primary" :disabled="loading">
            {{ loading ? 'Saving...' : (isEditing ? 'Save Changes' : 'Add Panel') }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<style scoped>
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
  width: 100%;
  max-width: 640px;
  max-height: 90vh;
  overflow-y: auto;
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

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.25rem 1.5rem;
  border-bottom: 1px solid var(--border-primary);
  position: sticky;
  top: 0;
  background: var(--bg-secondary);
  z-index: 1;
}

.modal-header h2 {
  margin: 0;
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--text-primary);
}

.btn-close {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  background: transparent;
  border: none;
  border-radius: 6px;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.2s;
}

.btn-close:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

form {
  padding: 1.5rem;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 1rem;
}

.form-group {
  margin-bottom: 1.25rem;
}

.form-group-small {
  min-width: 160px;
}

.query-builder-group {
  border-top: 1px solid var(--border-primary);
  padding-top: 1.25rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--text-primary);
}

.required {
  color: var(--accent-danger);
}

.form-group input,
.form-group textarea,
.form-group select {
  width: 100%;
  padding: 0.75rem 1rem;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-primary);
  border-radius: 6px;
  font-size: 0.875rem;
  color: var(--text-primary);
  transition: border-color 0.2s, box-shadow 0.2s;
}

.form-group input::placeholder,
.form-group textarea::placeholder {
  color: var(--text-tertiary);
}

.form-group input:focus,
.form-group textarea:focus,
.form-group select:focus {
  outline: none;
  border-color: var(--accent-primary);
  box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.15);
}

.form-group input:disabled,
.form-group textarea:disabled,
.form-group select:disabled {
  background: var(--bg-primary);
  color: var(--text-tertiary);
  cursor: not-allowed;
}

.form-group select {
  cursor: pointer;
  appearance: none;
  background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 24 24' fill='none' stroke='%23a0a0a0' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpath d='m6 9 6 6 6-6'/%3E%3C/svg%3E");
  background-repeat: no-repeat;
  background-position: right 0.75rem center;
  padding-right: 2.5rem;
}

.error-message {
  padding: 0.75rem 1rem;
  background: rgba(255, 107, 107, 0.1);
  border: 1px solid rgba(255, 107, 107, 0.3);
  border-radius: 6px;
  color: var(--accent-danger);
  font-size: 0.875rem;
  margin-bottom: 1.25rem;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  margin-top: 0.5rem;
}

.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0.625rem 1.25rem;
  border: 1px solid transparent;
  border-radius: 6px;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
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

.btn-primary {
  background: var(--accent-primary);
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: var(--accent-primary-hover);
}
</style>
