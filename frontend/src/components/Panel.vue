<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { Panel } from '../types/panel'
import { useTimeRange } from '../composables/useTimeRange'
import { useProm, type ChartSeries } from '../composables/useProm'
import LineChart from './LineChart.vue'

const props = defineProps<{
  panel: Panel
}>()

defineEmits<{
  edit: [panel: Panel]
  delete: [panel: Panel]
}>()

const { timeRange, onRefresh } = useTimeRange()

// Extract PromQL query from panel config
const promqlQuery = computed(() => {
  if (props.panel.query && typeof props.panel.query.promql === 'string') {
    return props.panel.query.promql
  }
  return ''
})

// Create refs for useProm inputs
const queryRef = ref(promqlQuery.value)
const startRef = computed(() => Math.floor(timeRange.value.start / 1000))
const endRef = computed(() => Math.floor(timeRange.value.end / 1000))

// Watch for query changes
watch(promqlQuery, (newQuery) => {
  queryRef.value = newQuery
}, { immediate: true })

const { chartData, loading, error, fetch: fetchData } = useProm({
  query: queryRef,
  start: startRef,
  end: endRef,
  autoFetch: true,
})

// Transform chartData to LineChart series format
const chartSeries = computed<ChartSeries[]>(() => {
  return chartData.value.series.map((s) => ({
    name: s.name,
    data: s.data,
  }))
})

// Subscribe to time range refresh
onRefresh(() => {
  if (promqlQuery.value) {
    fetchData()
  }
})

const hasQuery = computed(() => !!promqlQuery.value)
const isLineChart = computed(() => props.panel.type === 'line_chart')
</script>

<template>
  <div class="panel">
    <div class="panel-header">
      <h3 class="panel-title">{{ panel.title }}</h3>
      <div class="panel-actions">
        <button class="btn btn-icon" @click="$emit('edit', panel)" title="Edit">
          Edit
        </button>
        <button class="btn btn-icon btn-danger" @click="$emit('delete', panel)" title="Delete">
          X
        </button>
      </div>
    </div>
    <div class="panel-content">
      <div v-if="loading" class="panel-loading">
        Loading...
      </div>
      <div v-else-if="error" class="panel-error">
        {{ error }}
      </div>
      <div v-else-if="!hasQuery" class="panel-placeholder">
        <span class="panel-type">{{ panel.type }}</span>
        <p>No query configured. Edit this panel to add a PromQL query.</p>
      </div>
      <div v-else-if="isLineChart && chartSeries.length > 0" class="chart-container">
        <LineChart :series="chartSeries" />
      </div>
      <div v-else-if="chartSeries.length === 0" class="panel-no-data">
        <p>No data returned for the selected time range.</p>
      </div>
      <div v-else class="panel-placeholder">
        <span class="panel-type">{{ panel.type }}</span>
        <p>Visualization type not supported yet.</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.panel {
  background: #1e1e1e;
  border: 1px solid #333;
  border-radius: 4px;
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 150px;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem 1rem;
  border-bottom: 1px solid #333;
  background: #252526;
  cursor: move;
}

.panel-title {
  margin: 0;
  font-size: 0.875rem;
  font-weight: 600;
  color: #ccc;
}

.panel-actions {
  display: flex;
  gap: 0.25rem;
}

.btn {
  padding: 0.25rem 0.5rem;
  border: 1px solid #555;
  border-radius: 4px;
  background: #333;
  color: #ccc;
  cursor: pointer;
  font-size: 0.75rem;
}

.btn:hover {
  background: #444;
}

.btn-icon {
  padding: 0.25rem 0.5rem;
}

.btn-danger {
  color: #e74c3c;
  border-color: #e74c3c;
}

.btn-danger:hover {
  background: #4a2020;
}

.panel-content {
  flex: 1;
  padding: 0.5rem;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.chart-container {
  flex: 1;
  min-height: 0;
}

.panel-placeholder,
.panel-loading,
.panel-error,
.panel-no-data {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: #999;
  text-align: center;
}

.panel-error {
  color: #e74c3c;
}

.panel-type {
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  background: #333;
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  margin-bottom: 0.5rem;
}

.panel-placeholder p,
.panel-no-data p {
  margin: 0;
  font-size: 0.875rem;
}
</style>
