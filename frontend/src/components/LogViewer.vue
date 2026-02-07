<script setup lang="ts">
import { computed } from 'vue'
import type { LogEntry } from '../types/datasource'

const props = defineProps<{
  logs: LogEntry[]
}>()

function getLevelClass(level?: string): string {
  switch (level) {
    case 'error':
      return 'level-error'
    case 'warning':
    case 'warn':
      return 'level-warning'
    case 'info':
      return 'level-info'
    case 'debug':
      return 'level-debug'
    default:
      return ''
  }
}

function formatTimestamp(ts: string): string {
  try {
    const date = new Date(ts)
    return date.toLocaleTimeString('en-US', {
      hour12: false,
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
      fractionalSecondDigits: 3,
    })
  } catch {
    return ts
  }
}

const displayLogs = computed(() => props.logs.slice(0, 1000))
</script>

<template>
  <div class="log-viewer">
    <div class="log-header">
      <span class="log-count">{{ logs.length }} log entries</span>
    </div>
    <div class="log-table-wrapper">
      <table class="log-table">
        <thead>
          <tr>
            <th class="col-time">Timestamp</th>
            <th class="col-level">Level</th>
            <th class="col-message">Message</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="(log, i) in displayLogs"
            :key="i"
            :class="getLevelClass(log.level)"
          >
            <td class="col-time">
              <span class="timestamp">{{ formatTimestamp(log.timestamp) }}</span>
            </td>
            <td class="col-level">
              <span v-if="log.level" class="level-badge" :class="getLevelClass(log.level)">
                {{ log.level }}
              </span>
            </td>
            <td class="col-message">
              <span class="log-line">{{ log.line }}</span>
              <div v-if="log.labels && Object.keys(log.labels).length > 0" class="log-labels">
                <span
                  v-for="(value, key) in log.labels"
                  :key="String(key)"
                  class="label-tag"
                >
                  {{ key }}={{ value }}
                </span>
              </div>
            </td>
          </tr>
          <tr v-if="logs.length === 0">
            <td colspan="3" class="empty-row">No log entries</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<style scoped>
.log-viewer {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

.log-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.5rem 0;
  flex-shrink: 0;
}

.log-count {
  font-size: 0.75rem;
  color: var(--text-tertiary);
}

.log-table-wrapper {
  flex: 1;
  overflow: auto;
  border: 1px solid var(--border-primary);
  border-radius: 6px;
}

.log-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.8rem;
  font-family: 'JetBrains Mono', 'Fira Code', 'Cascadia Code', 'Consolas', monospace;
}

.log-table thead {
  position: sticky;
  top: 0;
  z-index: 1;
}

.log-table th {
  background: var(--bg-tertiary);
  padding: 0.5rem 0.75rem;
  text-align: left;
  font-weight: 600;
  color: var(--text-secondary);
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  border-bottom: 1px solid var(--border-primary);
}

.log-table td {
  padding: 0.375rem 0.75rem;
  border-bottom: 1px solid var(--border-primary);
  vertical-align: top;
}

.log-table tr:hover td {
  background: var(--bg-hover);
}

.col-time {
  width: 110px;
  white-space: nowrap;
}

.col-level {
  width: 70px;
}

.col-message {
  word-break: break-word;
}

.timestamp {
  color: var(--text-tertiary);
}

.level-badge {
  display: inline-block;
  padding: 0.125rem 0.375rem;
  border-radius: 3px;
  font-size: 0.7rem;
  font-weight: 600;
  text-transform: uppercase;
}

.level-error .level-badge {
  background: rgba(255, 107, 107, 0.15);
  color: #ff6b6b;
}

.level-warning .level-badge {
  background: rgba(254, 202, 87, 0.15);
  color: #feca57;
}

.level-info .level-badge {
  background: rgba(56, 189, 248, 0.15);
  color: var(--accent-primary);
}

.level-debug .level-badge {
  background: rgba(160, 160, 160, 0.15);
  color: #a0a0a0;
}

tr.level-error td {
  border-left: 2px solid #ff6b6b;
}

tr.level-warning td {
  border-left: 2px solid #feca57;
}

.log-line {
  color: var(--text-primary);
  white-space: pre-wrap;
}

.log-labels {
  margin-top: 0.25rem;
  display: flex;
  flex-wrap: wrap;
  gap: 0.25rem;
}

.label-tag {
  display: inline-block;
  padding: 0.1rem 0.375rem;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-primary);
  border-radius: 3px;
  font-size: 0.7rem;
  color: var(--text-secondary);
}

.empty-row {
  text-align: center;
  color: var(--text-tertiary);
  padding: 2rem 1rem !important;
}
</style>
