<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { X, Plus, Trash2 } from 'lucide-vue-next'
import type { Panel } from '../types/panel'
import { createPanel, updatePanel } from '../api/panels'
import QueryBuilder from './QueryBuilder.vue'

interface Threshold {
  value: number
  color: string
}

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

// Gauge-specific options
const gaugeMin = ref(
  typeof props.panel?.query?.min === 'number' ? props.panel.query.min : 0
)
const gaugeMax = ref(
  typeof props.panel?.query?.max === 'number' ? props.panel.query.max : 100
)
const gaugeUnit = ref(
  typeof props.panel?.query?.unit === 'string' ? props.panel.query.unit : ''
)
const gaugeDecimals = ref(
  typeof props.panel?.query?.decimals === 'number' ? props.panel.query.decimals : 2
)
const gaugeThresholds = ref<Threshold[]>(
  Array.isArray(props.panel?.query?.thresholds)
    ? (props.panel.query.thresholds as Threshold[])
    : [{ value: 80, color: '#ff6b6b' }]
)

const loading = ref(false)
const error = ref<string | null>(null)

const isGaugeType = computed(() => panelType.value === 'gauge')

function addThreshold() {
  const lastValue = gaugeThresholds.value.length > 0
    ? gaugeThresholds.value[gaugeThresholds.value.length - 1].value + 10
    : 50
  gaugeThresholds.value.push({ value: lastValue, color: '#feca57' })
}

function removeThreshold(index: number) {
  gaugeThresholds.value.splice(index, 1)
}

async function handleSubmit() {
  if (!title.value.trim()) {
    error.value = 'Title is required'
    return
  }

  // Build query config
  const query: Record<string, unknown> = {}

  if (promqlQuery.value.trim()) {
    query.promql = promqlQuery.value.trim()
  }

  // Add gauge-specific config if gauge type is selected
  if (isGaugeType.value) {
    query.min = gaugeMin.value
    query.max = gaugeMax.value
    query.unit = gaugeUnit.value
    query.decimals = gaugeDecimals.value
    query.thresholds = gaugeThresholds.value
  }

  const finalQuery = Object.keys(query).length > 0 ? query : undefined

  loading.value = true
  error.value = null

  try {
    if (isEditing.value && props.panel) {
      await updatePanel(props.panel.id, {
        title: title.value.trim(),
        type: panelType.value,
        query: finalQuery
      })
    } else {
      await createPanel(props.dashboardId, {
        title: title.value.trim(),
        type: panelType.value,
        grid_pos: { x: 0, y: 0, w: 6, h: 4 },
        query: finalQuery
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

        <!-- Gauge Configuration -->
        <div v-if="isGaugeType" class="gauge-config">
          <div class="config-header">
            <h4>Gauge Options</h4>
          </div>

          <div class="form-row form-row-4">
            <div class="form-group">
              <label for="gauge-min">Min</label>
              <input
                id="gauge-min"
                v-model.number="gaugeMin"
                type="number"
                :disabled="loading"
              />
            </div>
            <div class="form-group">
              <label for="gauge-max">Max</label>
              <input
                id="gauge-max"
                v-model.number="gaugeMax"
                type="number"
                :disabled="loading"
              />
            </div>
            <div class="form-group">
              <label for="gauge-unit">Unit</label>
              <input
                id="gauge-unit"
                v-model="gaugeUnit"
                type="text"
                placeholder="%"
                :disabled="loading"
              />
            </div>
            <div class="form-group">
              <label for="gauge-decimals">Decimals</label>
              <input
                id="gauge-decimals"
                v-model.number="gaugeDecimals"
                type="number"
                min="0"
                max="10"
                :disabled="loading"
              />
            </div>
          </div>

          <div class="thresholds-section">
            <div class="thresholds-header">
              <label>Thresholds</label>
              <button type="button" class="btn btn-sm" @click="addThreshold" :disabled="loading">
                <Plus :size="14" />
                Add
              </button>
            </div>
            <div class="thresholds-list">
              <div v-for="(threshold, index) in gaugeThresholds" :key="index" class="threshold-row">
                <input
                  v-model.number="threshold.value"
                  type="number"
                  placeholder="Value"
                  :disabled="loading"
                  class="threshold-value"
                />
                <input
                  v-model="threshold.color"
                  type="color"
                  :disabled="loading"
                  class="threshold-color"
                />
                <button
                  type="button"
                  class="btn-icon btn-icon-danger"
                  @click="removeThreshold(index)"
                  :disabled="loading"
                  title="Remove threshold"
                >
                  <Trash2 :size="14" />
                </button>
              </div>
              <p v-if="gaugeThresholds.length === 0" class="thresholds-empty">
                No thresholds configured. Values below any threshold will show green.
              </p>
            </div>
          </div>
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

/* Gauge configuration styles */
.gauge-config {
  border-top: 1px solid var(--border-primary);
  padding-top: 1.25rem;
  margin-bottom: 1.25rem;
}

.config-header h4 {
  margin: 0 0 1rem 0;
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--text-primary);
}

.form-row-4 {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 0.75rem;
}

.form-row-4 .form-group {
  margin-bottom: 0.75rem;
}

.thresholds-section {
  margin-top: 1rem;
}

.thresholds-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
}

.thresholds-header label {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--text-primary);
}

.btn-sm {
  padding: 0.375rem 0.625rem;
  font-size: 0.75rem;
  gap: 0.25rem;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-primary);
  color: var(--text-primary);
}

.btn-sm:hover:not(:disabled) {
  background: var(--bg-hover);
}

.thresholds-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.threshold-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.threshold-value {
  flex: 1;
  padding: 0.5rem 0.75rem;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-primary);
  border-radius: 6px;
  font-size: 0.875rem;
  color: var(--text-primary);
}

.threshold-value:focus {
  outline: none;
  border-color: var(--accent-primary);
}

.threshold-color {
  width: 40px;
  height: 36px;
  padding: 2px;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-primary);
  border-radius: 6px;
  cursor: pointer;
}

.threshold-color::-webkit-color-swatch-wrapper {
  padding: 2px;
}

.threshold-color::-webkit-color-swatch {
  border: none;
  border-radius: 4px;
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

.thresholds-empty {
  font-size: 0.75rem;
  color: var(--text-tertiary);
  margin: 0;
  padding: 0.5rem;
  text-align: center;
}
</style>
