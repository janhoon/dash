<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { Pencil, Trash2, AlertCircle, BarChart3 } from 'lucide-vue-next'
import type { Panel } from '../types/panel'
import { useTimeRange } from '../composables/useTimeRange'
import { useProm } from '../composables/useProm'
import LineChart, { type ChartSeries } from './LineChart.vue'
import GaugeChart, { type Threshold } from './GaugeChart.vue'

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

// Get the latest value for gauge chart (from first series)
const gaugeValue = computed(() => {
  if (chartData.value.series.length === 0) return 0
  const firstSeries = chartData.value.series[0]
  if (firstSeries.data.length === 0) return 0
  return firstSeries.data[firstSeries.data.length - 1].value
})

// Extract gauge config from panel query
const gaugeConfig = computed(() => {
  const query = props.panel.query || {}
  return {
    min: typeof query.min === 'number' ? query.min : 0,
    max: typeof query.max === 'number' ? query.max : 100,
    unit: typeof query.unit === 'string' ? query.unit : '',
    decimals: typeof query.decimals === 'number' ? query.decimals : 2,
    thresholds: Array.isArray(query.thresholds)
      ? (query.thresholds as Threshold[])
      : [],
  }
})

// Subscribe to time range refresh
onRefresh(() => {
  if (promqlQuery.value) {
    fetchData()
  }
})

const hasQuery = computed(() => !!promqlQuery.value)
const isLineChart = computed(() => props.panel.type === 'line_chart')
const isGaugeChart = computed(() => props.panel.type === 'gauge')
</script>

<template>
  <div class="panel">
    <div class="panel-header">
      <h3 class="panel-title">{{ panel.title }}</h3>
      <div class="panel-actions">
        <button class="btn-icon" @click="$emit('edit', panel)" title="Edit">
          <Pencil :size="14" />
        </button>
        <button class="btn-icon btn-icon-danger" @click="$emit('delete', panel)" title="Delete">
          <Trash2 :size="14" />
        </button>
      </div>
    </div>
    <div class="panel-content">
      <div v-if="loading" class="panel-state">
        <div class="loading-spinner"></div>
        <span>Loading...</span>
      </div>
      <div v-else-if="error" class="panel-state panel-error">
        <AlertCircle :size="24" />
        <span>{{ error }}</span>
      </div>
      <div v-else-if="!hasQuery" class="panel-state panel-placeholder">
        <BarChart3 :size="32" />
        <span class="panel-type">{{ panel.type }}</span>
        <p>No query configured</p>
      </div>
      <div v-else-if="isLineChart && chartSeries.length > 0" class="chart-container">
        <LineChart :series="chartSeries" />
      </div>
      <div v-else-if="isGaugeChart && chartSeries.length > 0" class="chart-container">
        <GaugeChart
          :value="gaugeValue"
          :min="gaugeConfig.min"
          :max="gaugeConfig.max"
          :unit="gaugeConfig.unit"
          :decimals="gaugeConfig.decimals"
          :thresholds="gaugeConfig.thresholds"
        />
      </div>
      <div v-else-if="chartSeries.length === 0" class="panel-state panel-no-data">
        <BarChart3 :size="24" />
        <p>No data for selected range</p>
      </div>
      <div v-else class="panel-state panel-placeholder">
        <span class="panel-type">{{ panel.type }}</span>
        <p>Visualization not supported</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.panel {
  background: var(--bg-secondary);
  border: 1px solid var(--border-primary);
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 150px;
  transition: border-color 0.2s, box-shadow 0.2s;
}

.panel:hover {
  border-color: var(--border-secondary);
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem 1rem;
  border-bottom: 1px solid var(--border-primary);
  background: var(--bg-tertiary);
  border-radius: 8px 8px 0 0;
  cursor: move;
}

.panel-title {
  margin: 0;
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.panel-actions {
  display: flex;
  gap: 0.25rem;
  opacity: 0;
  transition: opacity 0.2s;
}

.panel:hover .panel-actions {
  opacity: 1;
}

.btn-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  padding: 0;
  background: transparent;
  border: none;
  border-radius: 4px;
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

.panel-content {
  flex: 1;
  padding: 0.75rem;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.chart-container {
  flex: 1;
  min-height: 0;
}

.panel-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  gap: 0.5rem;
  color: var(--text-tertiary);
  text-align: center;
}

.panel-state p,
.panel-state span {
  font-size: 0.75rem;
  margin: 0;
}

.panel-error {
  color: var(--accent-danger);
}

.panel-type {
  font-size: 0.625rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  background: var(--bg-tertiary);
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  color: var(--text-secondary);
}

.loading-spinner {
  width: 24px;
  height: 24px;
  border: 2px solid var(--border-primary);
  border-top-color: var(--accent-primary);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}
</style>
