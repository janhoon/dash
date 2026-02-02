<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { Play, AlertCircle, History, X } from 'lucide-vue-next'
import QueryBuilder from '../components/QueryBuilder.vue'
import TimeRangePicker from '../components/TimeRangePicker.vue'
import LineChart from '../components/LineChart.vue'
import { useTimeRange } from '../composables/useTimeRange'
import { queryPrometheus, transformToChartData, type PrometheusQueryResult } from '../composables/useProm'
import type { ChartSeries } from '../components/LineChart.vue'

const { timeRange, onRefresh } = useTimeRange()

// Query state
const query = ref('')
const loading = ref(false)
const error = ref<string | null>(null)
const result = ref<PrometheusQueryResult | null>(null)
const chartSeries = ref<ChartSeries[]>([])

// Query history (session storage)
const HISTORY_KEY = 'explore_query_history'
const MAX_HISTORY = 10
const queryHistory = ref<string[]>([])
const showHistory = ref(false)

// Load history from session storage
onMounted(() => {
  const stored = sessionStorage.getItem(HISTORY_KEY)
  if (stored) {
    try {
      queryHistory.value = JSON.parse(stored)
    } catch {
      queryHistory.value = []
    }
  }
})

// Save query to history
function addToHistory(q: string) {
  if (!q.trim()) return

  // Remove duplicate if exists
  const filtered = queryHistory.value.filter(h => h !== q)

  // Add to beginning
  queryHistory.value = [q, ...filtered].slice(0, MAX_HISTORY)

  // Save to session storage
  sessionStorage.setItem(HISTORY_KEY, JSON.stringify(queryHistory.value))
}

// Run the query
async function runQuery() {
  if (!query.value.trim()) {
    error.value = 'Query is required'
    return
  }

  loading.value = true
  error.value = null
  result.value = null
  chartSeries.value = []

  try {
    // Convert time range from milliseconds to seconds
    const start = Math.floor(timeRange.value.start / 1000)
    const end = Math.floor(timeRange.value.end / 1000)

    // Calculate step based on time range (aim for ~200 data points)
    const duration = end - start
    const step = Math.max(15, Math.floor(duration / 200))

    const response = await queryPrometheus(query.value, start, end, step)
    result.value = response

    if (response.status === 'error') {
      error.value = response.error || 'Query failed'
    } else {
      const chartData = transformToChartData(response)
      chartSeries.value = chartData.series.map(s => ({
        name: s.name,
        data: s.data
      }))

      // Add to history on successful query
      addToHistory(query.value)
    }
  } catch (e) {
    error.value = e instanceof Error ? e.message : 'Failed to execute query'
  } finally {
    loading.value = false
  }
}

// Handle keyboard shortcut
function handleKeydown(e: KeyboardEvent) {
  if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
    e.preventDefault()
    runQuery()
  }
}

// Select query from history
function selectHistoryQuery(q: string) {
  query.value = q
  showHistory.value = false
}

// Clear history
function clearHistory() {
  queryHistory.value = []
  sessionStorage.removeItem(HISTORY_KEY)
}

// Subscribe to refresh events
let unsubscribeRefresh: (() => void) | null = null

onMounted(() => {
  unsubscribeRefresh = onRefresh(() => {
    if (query.value.trim() && result.value?.status === 'success') {
      runQuery()
    }
  })
})

onUnmounted(() => {
  if (unsubscribeRefresh) {
    unsubscribeRefresh()
  }
})

// Computed properties
const hasResults = computed(() => result.value?.status === 'success' && chartSeries.value.length > 0)
const seriesCount = computed(() => chartSeries.value.length)
</script>

<template>
  <div class="explore-page" @keydown="handleKeydown">
    <header class="explore-header">
      <h1>Explore</h1>
      <div class="header-actions">
        <TimeRangePicker />
      </div>
    </header>

    <div class="explore-content">
      <div class="query-section">
        <div class="query-builder-wrapper">
          <QueryBuilder v-model="query" :disabled="loading" />

          <!-- History button -->
          <div v-if="queryHistory.length > 0" class="history-container">
            <button
              class="history-btn"
              :class="{ active: showHistory }"
              @click="showHistory = !showHistory"
              title="Query history"
            >
              <History :size="16" />
              <span>History</span>
            </button>

            <!-- Query history dropdown -->
            <div v-if="showHistory" class="history-dropdown">
              <div class="history-header">
                <span>Recent Queries</span>
                <button class="clear-history-btn" @click="clearHistory" title="Clear history">
                  <X :size="14" />
                </button>
              </div>
              <button
                v-for="(q, index) in queryHistory"
                :key="index"
                class="history-item"
                @click="selectHistoryQuery(q)"
              >
                <code>{{ q }}</code>
              </button>
            </div>
          </div>
        </div>

        <div class="query-actions">
          <button
            class="btn btn-run"
            :disabled="loading || !query.trim()"
            @click="runQuery"
          >
            <Play :size="16" />
            <span>{{ loading ? 'Running...' : 'Run Query' }}</span>
          </button>
          <span class="hint">Ctrl+Enter to run</span>
        </div>

        <!-- Error display -->
        <div v-if="error" class="query-error">
          <AlertCircle :size="16" />
          <span>{{ error }}</span>
        </div>
      </div>

      <!-- Results section -->
      <div class="results-section">
        <div v-if="loading" class="loading-state">
          <div class="loading-spinner"></div>
          <span>Executing query...</span>
        </div>

        <div v-else-if="hasResults" class="results-container">
          <div class="results-header">
            <span class="result-count">{{ seriesCount }} {{ seriesCount === 1 ? 'series' : 'series' }}</span>
          </div>
          <div class="chart-container">
            <LineChart :series="chartSeries" :height="400" />
          </div>
        </div>

        <div v-else-if="result?.status === 'success' && chartSeries.length === 0" class="empty-state">
          <p>No data returned for the selected time range.</p>
        </div>

        <div v-else class="empty-state">
          <p>Write a PromQL query and click "Run Query" to visualize your metrics.</p>
          <p class="hint-text">Examples: <code>up</code>, <code>rate(http_requests_total[5m])</code>, <code>node_cpu_seconds_total</code></p>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.explore-page {
  display: flex;
  flex-direction: column;
  min-height: 100%;
  padding: 1.5rem 2rem;
}

.explore-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
}

.explore-header h1 {
  font-size: 1.5rem;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.explore-content {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  flex: 1;
}

.query-section {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding: 1.5rem;
  background: var(--bg-secondary);
  border: 1px solid var(--border-primary);
  border-radius: 12px;
}

.query-builder-wrapper {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.history-container {
  position: relative;
}

.history-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-primary);
  border-radius: 6px;
  color: var(--text-secondary);
  font-size: 0.8125rem;
  cursor: pointer;
  transition: all 0.2s;
}

.history-btn:hover,
.history-btn.active {
  background: var(--bg-hover);
  color: var(--text-primary);
  border-color: var(--border-secondary);
}

.history-dropdown {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  width: 350px;
  max-height: 300px;
  overflow-y: auto;
  background: var(--bg-secondary);
  border: 1px solid var(--border-primary);
  border-radius: 8px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.4);
  z-index: 100;
}

.history-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem 1rem;
  border-bottom: 1px solid var(--border-primary);
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.02em;
}

.clear-history-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  background: transparent;
  border: none;
  border-radius: 4px;
  color: var(--text-tertiary);
  cursor: pointer;
  transition: all 0.2s;
}

.clear-history-btn:hover {
  background: var(--bg-hover);
  color: var(--accent-danger);
}

.history-item {
  display: block;
  width: 100%;
  padding: 0.625rem 1rem;
  background: transparent;
  border: none;
  text-align: left;
  cursor: pointer;
  transition: background-color 0.15s;
}

.history-item:hover {
  background: var(--bg-hover);
}

.history-item code {
  display: block;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 0.8125rem;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.query-actions {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.btn-run {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.625rem 1.25rem;
  background: var(--accent-success);
  border: 1px solid var(--accent-success);
  border-radius: 8px;
  color: white;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}

.btn-run:hover:not(:disabled) {
  background: #00c49a;
  border-color: #00c49a;
}

.btn-run:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.hint {
  font-size: 0.75rem;
  color: var(--text-tertiary);
}

.query-error {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 1rem;
  background: rgba(255, 107, 107, 0.1);
  border: 1px solid rgba(255, 107, 107, 0.3);
  border-radius: 8px;
  color: var(--accent-danger);
  font-size: 0.875rem;
}

.results-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: var(--bg-secondary);
  border: 1px solid var(--border-primary);
  border-radius: 12px;
  overflow: hidden;
  min-height: 400px;
}

.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  padding: 3rem;
  color: var(--text-secondary);
  flex: 1;
}

.loading-spinner {
  width: 32px;
  height: 32px;
  border: 3px solid var(--border-primary);
  border-top-color: var(--accent-primary);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.results-container {
  display: flex;
  flex-direction: column;
  flex: 1;
}

.results-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem 1rem;
  border-bottom: 1px solid var(--border-primary);
  background: var(--bg-tertiary);
}

.result-count {
  font-size: 0.8125rem;
  color: var(--text-secondary);
}

.chart-container {
  flex: 1;
  padding: 1rem;
  min-height: 400px;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 3rem;
  text-align: center;
  flex: 1;
}

.empty-state p {
  margin: 0;
  color: var(--text-secondary);
  font-size: 0.9375rem;
}

.empty-state .hint-text {
  font-size: 0.8125rem;
  color: var(--text-tertiary);
}

.empty-state code {
  padding: 0.125rem 0.375rem;
  background: var(--bg-tertiary);
  border-radius: 4px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 0.8125rem;
  color: var(--accent-primary);
}
</style>
