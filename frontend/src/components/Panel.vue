<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { Pencil, Trash2, AlertCircle, BarChart3 } from 'lucide-vue-next'
import type { Panel } from '../types/panel'
import { useTimeRange } from '../composables/useTimeRange'
import { useProm } from '../composables/useProm'
import LineChart, { type ChartSeries } from './LineChart.vue'
import GaugeChart, { type Threshold } from './GaugeChart.vue'
import PieChart, { type PieDataItem } from './PieChart.vue'
import StatPanel, { type DataPoint } from './StatPanel.vue'

const props = defineProps<{
  panel: Panel
}>()

defineEmits<{
  edit: [panel: Panel]
  delete: [panel: Panel]
}>()

const { timeRange, onRefresh } = useTimeRange()

// Setup Prometheus query
const promqlQuery = computed(() => props.panel.query?.promql || '')

// Create refs for useProm
const queryRef = ref(promqlQuery.value)
const startRef = computed(() => Math.floor(timeRange.value.start / 1000))
const endRef = computed(() => Math.floor(timeRange.value.end / 1000))

// Watch for query changes
watch(promqlQuery, (newQuery) => {
  queryRef.value = newQuery
}, { immediate: true })

const { chartData: promChartData, loading, error, fetch: refetch } = useProm({
  query: queryRef,
  start: startRef,
  end: endRef,
  autoFetch: true,
})

// Use chart data from useProm
const chartData = promChartData

// Transform to chart series format
const chartSeries = computed(() => {
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

// Transform chartData to PieChart data format (use latest value from each series)
const pieData = computed<PieDataItem[]>(() => {
  return chartData.value.series.map((s) => ({
    name: s.name,
    value: s.data.length > 0 ? s.data[s.data.length - 1].value : 0,
  }))
})

// Extract pie chart config from panel query
const pieConfig = computed(() => {
  const query = props.panel.query || {}
  return {
    displayAs: (query.displayAs === 'donut' ? 'donut' : 'pie') as 'pie' | 'donut',
    showLegend: query.showLegend !== false,
    showLabels: query.showLabels !== false,
  }
})

// Auto-refresh on time range change
watch([timeRange, onRefresh], () => {
  if (hasQuery.value) {
    refetch()
  }
})

// Transform data to StatPanel format
const statData = computed<DataPoint[]>(() => {
  if (chartData.value.series.length === 0) return []
  const firstSeries = chartData.value.series[0]
  return firstSeries.data.map((d) => ({
    timestamp: d.timestamp,
    value: d.value,
  }))
})

// Extract stat panel config
const statConfig = computed(() => {
  const query = props.panel.query || {}
  return {
    unit: typeof query.unit === 'string' ? query.unit : '',
    decimals: typeof query.decimals === 'number' ? query.decimals : 2,
    showTrend: query.showTrend !== false,
    showSparkline: query.showSparkline !== false,
    thresholds: Array.isArray(query.thresholds) ? (query.thresholds as Threshold[]) : [],
  }
})

const hasQuery = computed(() => !!promqlQuery.value)
const isLineChart = computed(() => props.panel.type === 'line_chart')
const isGaugeChart = computed(() => props.panel.type === 'gauge')
const isPieChart = computed(() => props.panel.type === 'pie')
const isStatPanel = computed(() => props.panel.type === 'stat')
</script>

<template>
  <div class="panel">
    <div class="panel-header">
      <h3 class="panel-title">{{ panel.title }}</h3>
      <div class="panel-actions">
        <button class="panel-action-btn" @click="$emit('edit', panel)" title="Edit panel">
          <Pencil :size="16" />
        </button>
        <button class="panel-action-btn" @click="$emit('delete', panel)" title="Delete panel">
          <Trash2 :size="16" />
        </button>
      </div>
    </div>
    <div class="panel-body">
      <div v-if="!hasQuery" class="panel-state panel-no-query">
        <BarChart3 :size="48" class="icon-muted" />
        <p class="text-muted">No query configured</p>
        <button class="btn-primary" @click="$emit('edit', panel)">Configure Panel</button>
      </div>
      <div v-else-if="loading" class="panel-state">
        <div class="spinner"></div>
        <p class="text-muted">Loading data...</p>
      </div>
      <div v-else-if="error" class="panel-state panel-error">
        <AlertCircle :size="48" class="icon-error" />
        <p class="error-text">{{ error }}</p>
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
      <div v-else-if="isPieChart && pieData.length > 0" class="chart-container">
        <PieChart
          :data="pieData"
          :display-as="pieConfig.displayAs"
          :show-legend="pieConfig.showLegend"
          :show-labels="pieConfig.showLabels"
        />
      </div>
      <div v-else-if="isStatPanel && statData.length > 0" class="chart-container">
        <StatPanel
          :data="statData"
          :label="panel.title"
          :unit="statConfig.unit"
          :decimals="statConfig.decimals"
          :thresholds="statConfig.thresholds"
          :show-trend="statConfig.showTrend"
          :show-sparkline="statConfig.showSparkline"
        />
      </div>
      <div v-else-if="chartSeries.length === 0" class="panel-state panel-no-data">
        <AlertCircle :size="48" class="icon-warning" />
        <p class="text-muted">No data available</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.panel {
  background: var(--bg-secondary);
  border: 1px solid var(--border-primary);
  border-radius: 8px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  height: 100%;
  transition: border-color 0.2s, box-shadow 0.2s;
}

.panel:hover {
  border-color: var(--border-secondary);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border-primary);
  background: var(--bg-tertiary);
}

.panel-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.panel-actions {
  display: flex;
  gap: 4px;
}

.panel-action-btn {
  padding: 4px;
  background: transparent;
  border: none;
  color: var(--text-tertiary);
  cursor: pointer;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.panel-action-btn:hover {
  background: var(--bg-hover);
  color: var(--text-primary);
}

.panel-body {
  flex: 1;
  padding: 16px;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

.chart-container {
  flex: 1;
  min-height: 0;
}

.panel-state {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
}

.panel-no-query,
.panel-no-data {
  color: var(--text-tertiary);
}

.panel-error {
  color: var(--accent-danger);
}

.icon-muted {
  color: var(--text-tertiary);
}

.icon-warning {
  color: var(--accent-warning);
}

.icon-error {
  color: var(--accent-danger);
}

.text-muted {
  color: var(--text-secondary);
  font-size: 14px;
  margin: 0;
}

.error-text {
  color: var(--accent-danger);
  font-size: 14px;
  margin: 0;
}

.spinner {
  width: 32px;
  height: 32px;
  border: 3px solid var(--border-primary);
  border-top-color: var(--accent-primary);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.btn-primary {
  padding: 8px 16px;
  background: var(--accent-primary);
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.2s;
}

.btn-primary:hover {
  background: var(--accent-primary-hover);
}
</style>
