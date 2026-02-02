<script setup lang="ts">
import { ref, computed, onUnmounted } from 'vue'
import { useTimeRange } from '../composables/useTimeRange'

const {
  displayText,
  selectedPreset,
  isCustomRange,
  refreshInterval,
  refreshIntervalValue,
  presets,
  refreshIntervals,
  setPreset,
  setCustomRange,
  setRefreshInterval,
  refresh,
} = useTimeRange()

const isOpen = ref(false)
const showCustomRange = ref(false)
const customFrom = ref('')
const customTo = ref('')
const customRangeError = ref<string | null>(null)

const currentDisplayText = computed(() => displayText.value)

function toggleDropdown() {
  isOpen.value = !isOpen.value
  if (!isOpen.value) {
    showCustomRange.value = false
    customRangeError.value = null
  }
}

function closeDropdown() {
  isOpen.value = false
  showCustomRange.value = false
  customRangeError.value = null
}

function selectPreset(presetValue: string) {
  setPreset(presetValue)
  closeDropdown()
}

function openCustomRange() {
  showCustomRange.value = true
  // Initialize with current date/time values
  const now = new Date()
  const oneHourAgo = new Date(now.getTime() - 60 * 60 * 1000)
  customFrom.value = formatDateTimeLocal(oneHourAgo)
  customTo.value = formatDateTimeLocal(now)
}

function applyCustomRange() {
  const fromDate = new Date(customFrom.value)
  const toDate = new Date(customTo.value)

  if (isNaN(fromDate.getTime()) || isNaN(toDate.getTime())) {
    customRangeError.value = 'Please enter valid dates'
    return
  }

  if (fromDate >= toDate) {
    customRangeError.value = 'Start time must be before end time'
    return
  }

  setCustomRange(fromDate.getTime(), toDate.getTime())
  closeDropdown()
}

function cancelCustomRange() {
  showCustomRange.value = false
  customRangeError.value = null
}

function selectRefreshInterval(intervalValue: string) {
  setRefreshInterval(intervalValue)
}

function handleRefresh() {
  refresh()
}

function formatDateTimeLocal(date: Date): string {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  return `${year}-${month}-${day}T${hours}:${minutes}`
}

// Close dropdown when clicking outside
function handleClickOutside(event: MouseEvent) {
  const target = event.target as HTMLElement
  if (!target.closest('.time-range-picker')) {
    closeDropdown()
  }
}

// Add/remove click listener
if (typeof window !== 'undefined') {
  window.addEventListener('click', handleClickOutside)
}

onUnmounted(() => {
  if (typeof window !== 'undefined') {
    window.removeEventListener('click', handleClickOutside)
  }
})
</script>

<template>
  <div class="time-range-picker">
    <div class="picker-controls">
      <button
        class="time-display"
        @click.stop="toggleDropdown"
        :class="{ active: isOpen }"
      >
        <span class="clock-icon">&#128337;</span>
        <span class="display-text">{{ currentDisplayText }}</span>
        <span class="dropdown-arrow">{{ isOpen ? '&#9650;' : '&#9660;' }}</span>
      </button>

      <button
        class="refresh-btn"
        @click="handleRefresh"
        title="Refresh now"
      >
        &#8635;
      </button>

      <div class="refresh-interval-selector">
        <select
          :value="refreshIntervalValue"
          @change="selectRefreshInterval(($event.target as HTMLSelectElement).value)"
          title="Auto-refresh interval"
        >
          <option
            v-for="interval in refreshIntervals"
            :key="interval.value"
            :value="interval.value"
          >
            {{ interval.label }}
          </option>
        </select>
      </div>
    </div>

    <div v-if="isOpen" class="dropdown" @click.stop>
      <div v-if="!showCustomRange" class="preset-list">
        <div class="dropdown-section">
          <div class="section-title">Quick ranges</div>
          <button
            v-for="preset in presets"
            :key="preset.value"
            class="preset-item"
            :class="{ selected: !isCustomRange && selectedPreset === preset.value }"
            @click="selectPreset(preset.value)"
          >
            {{ preset.label }}
          </button>
        </div>

        <div class="dropdown-divider"></div>

        <button class="preset-item custom-range-btn" @click="openCustomRange">
          Custom range...
        </button>
      </div>

      <div v-else class="custom-range-form">
        <div class="section-title">Custom time range</div>

        <div class="form-group">
          <label for="custom-from">From</label>
          <input
            id="custom-from"
            type="datetime-local"
            v-model="customFrom"
          />
        </div>

        <div class="form-group">
          <label for="custom-to">To</label>
          <input
            id="custom-to"
            type="datetime-local"
            v-model="customTo"
          />
        </div>

        <div v-if="customRangeError" class="error-message">
          {{ customRangeError }}
        </div>

        <div class="form-actions">
          <button class="btn-cancel" @click="cancelCustomRange">Cancel</button>
          <button class="btn-apply" @click="applyCustomRange">Apply</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.time-range-picker {
  position: relative;
  display: inline-block;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
}

.picker-controls {
  display: flex;
  align-items: center;
  gap: 8px;
}

.time-display {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: #fff;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  color: #333;
  transition: border-color 0.2s, box-shadow 0.2s;
}

.time-display:hover {
  border-color: #3498db;
}

.time-display.active {
  border-color: #3498db;
  box-shadow: 0 0 0 2px rgba(52, 152, 219, 0.2);
}

.clock-icon {
  font-size: 16px;
}

.display-text {
  min-width: 120px;
}

.dropdown-arrow {
  font-size: 10px;
  color: #666;
}

.refresh-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  background: #fff;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  cursor: pointer;
  font-size: 18px;
  color: #333;
  transition: border-color 0.2s, background-color 0.2s;
}

.refresh-btn:hover {
  border-color: #3498db;
  background: #f5f5f5;
}

.refresh-btn:active {
  background: #e8e8e8;
}

.refresh-interval-selector select {
  padding: 8px 12px;
  background: #fff;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  font-size: 14px;
  color: #333;
  cursor: pointer;
  transition: border-color 0.2s;
}

.refresh-interval-selector select:hover {
  border-color: #3498db;
}

.refresh-interval-selector select:focus {
  outline: none;
  border-color: #3498db;
  box-shadow: 0 0 0 2px rgba(52, 152, 219, 0.2);
}

.dropdown {
  position: absolute;
  top: calc(100% + 4px);
  left: 0;
  min-width: 220px;
  background: #fff;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  z-index: 1000;
}

.dropdown-section {
  padding: 8px 0;
}

.section-title {
  padding: 8px 16px;
  font-size: 12px;
  font-weight: 600;
  color: #666;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.preset-item {
  display: block;
  width: 100%;
  padding: 10px 16px;
  background: none;
  border: none;
  text-align: left;
  font-size: 14px;
  color: #333;
  cursor: pointer;
  transition: background-color 0.15s;
}

.preset-item:hover {
  background: #f5f5f5;
}

.preset-item.selected {
  background: #e6f4ff;
  color: #3498db;
  font-weight: 500;
}

.dropdown-divider {
  height: 1px;
  background: #e8e8e8;
  margin: 4px 0;
}

.custom-range-btn {
  color: #3498db;
}

.custom-range-form {
  padding: 16px;
}

.form-group {
  margin-bottom: 12px;
}

.form-group label {
  display: block;
  margin-bottom: 4px;
  font-size: 12px;
  font-weight: 500;
  color: #666;
}

.form-group input {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  font-size: 14px;
  box-sizing: border-box;
}

.form-group input:focus {
  outline: none;
  border-color: #3498db;
  box-shadow: 0 0 0 2px rgba(52, 152, 219, 0.2);
}

.error-message {
  padding: 8px 12px;
  margin-bottom: 12px;
  background: #fff2f0;
  border: 1px solid #ffccc7;
  border-radius: 4px;
  color: #e74c3c;
  font-size: 13px;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.btn-cancel,
.btn-apply {
  padding: 8px 16px;
  border-radius: 4px;
  font-size: 14px;
  cursor: pointer;
  transition: background-color 0.2s, border-color 0.2s;
}

.btn-cancel {
  background: #fff;
  border: 1px solid #d9d9d9;
  color: #333;
}

.btn-cancel:hover {
  border-color: #999;
}

.btn-apply {
  background: #3498db;
  border: 1px solid #3498db;
  color: #fff;
}

.btn-apply:hover {
  background: #2980b9;
  border-color: #2980b9;
}
</style>
