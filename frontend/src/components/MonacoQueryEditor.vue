<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted, shallowRef } from 'vue'
import * as monaco from 'monaco-editor'
import { registerPromQLLanguage, definePromQLTheme, PROMQL_LANGUAGE_ID } from '../promql/language'
import { registerCompletionProvider } from '../promql/completionProvider'
import { registerHoverProvider } from '../promql/hoverProvider'

// Initialize Monaco PromQL support (only once)
let initialized = false
function initializeMonaco() {
  if (initialized) return
  initialized = true

  registerPromQLLanguage(monaco)
  definePromQLTheme(monaco)
  registerCompletionProvider(monaco)
  registerHoverProvider(monaco)
}

const props = withDefaults(defineProps<{
  modelValue: string
  disabled?: boolean
  height?: number
  placeholder?: string
}>(), {
  height: 100,
  placeholder: 'Enter PromQL query...'
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
  'submit': []
}>()

const containerRef = ref<HTMLElement | null>(null)
const editorInstance = shallowRef<monaco.editor.IStandaloneCodeEditor | null>(null)

// Track if component is mounted
const isMounted = ref(false)

// Create editor on mount
onMounted(() => {
  isMounted.value = true

  if (!containerRef.value) return

  // Initialize Monaco once
  initializeMonaco()

  // Create editor
  const editor = monaco.editor.create(containerRef.value, {
    value: props.modelValue,
    language: PROMQL_LANGUAGE_ID,
    theme: 'promql-dark',
    minimap: { enabled: false },
    lineNumbers: 'on',
    wordWrap: 'on',
    scrollBeyondLastLine: false,
    automaticLayout: true,
    fontSize: 13,
    fontFamily: "'Monaco', 'Menlo', 'Ubuntu Mono', monospace",
    padding: { top: 8, bottom: 8 },
    renderLineHighlight: 'line',
    lineHeight: 20,
    folding: false,
    glyphMargin: false,
    lineDecorationsWidth: 8,
    lineNumbersMinChars: 3,
    overviewRulerBorder: false,
    hideCursorInOverviewRuler: true,
    // Fix autocomplete dropdown being clipped by container
    fixedOverflowWidgets: true,
    scrollbar: {
      vertical: 'auto',
      horizontal: 'auto',
      verticalScrollbarSize: 8,
      horizontalScrollbarSize: 8
    },
    suggest: {
      showIcons: true,
      showStatusBar: true,
      preview: true,
      previewMode: 'prefix'
    },
    quickSuggestions: {
      other: true,
      comments: false,
      strings: true
    },
    acceptSuggestionOnEnter: 'on',
    tabCompletion: 'on',
    readOnly: props.disabled
  })

  editorInstance.value = editor

  // Listen for content changes
  editor.onDidChangeModelContent(() => {
    const value = editor.getValue()
    if (value !== props.modelValue) {
      emit('update:modelValue', value)
    }
  })

  // Handle Ctrl+Enter to submit
  editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.Enter, () => {
    emit('submit')
  })
})

// Clean up on unmount
onUnmounted(() => {
  isMounted.value = false
  if (editorInstance.value) {
    editorInstance.value.dispose()
    editorInstance.value = null
  }
})

// Sync external value changes to editor
watch(() => props.modelValue, (newValue) => {
  if (editorInstance.value && editorInstance.value.getValue() !== newValue) {
    editorInstance.value.setValue(newValue)
  }
})

// Handle disabled state
watch(() => props.disabled, (disabled) => {
  if (editorInstance.value) {
    editorInstance.value.updateOptions({ readOnly: disabled })
  }
})

// Handle height changes
watch(() => props.height, () => {
  if (editorInstance.value) {
    editorInstance.value.layout()
  }
})

// Focus the editor
function focus() {
  editorInstance.value?.focus()
}

// Expose methods
defineExpose({ focus })
</script>

<template>
  <div class="monaco-query-editor" :class="{ disabled }">
    <div
      ref="containerRef"
      class="editor-container"
      :style="{ height: `${height}px` }"
    ></div>
    <div v-if="!modelValue && !editorInstance" class="placeholder">
      {{ placeholder }}
    </div>
  </div>
</template>

<style scoped>
.monaco-query-editor {
  position: relative;
  border: 1px solid var(--border-primary);
  border-radius: 6px;
  background: var(--bg-secondary);
  transition: border-color 0.2s, box-shadow 0.2s;
}

.monaco-query-editor:focus-within {
  border-color: var(--accent-primary);
  box-shadow: var(--focus-ring);
}

.monaco-query-editor.disabled {
  opacity: 0.6;
  pointer-events: none;
}

.editor-container {
  width: 100%;
  min-height: 60px;
}

.placeholder {
  position: absolute;
  top: 8px;
  left: 48px;
  color: var(--text-tertiary);
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 13px;
  pointer-events: none;
}

/* Monaco editor styling overrides */
:deep(.monaco-editor) {
  border-radius: 6px;
}

:deep(.monaco-editor .margin) {
  background: var(--bg-secondary) !important;
}

:deep(.monaco-editor .monaco-scrollable-element > .scrollbar > .slider) {
  background: var(--border-secondary) !important;
  border-radius: 4px;
}

:deep(.monaco-editor .monaco-scrollable-element > .scrollbar > .slider:hover) {
  background: #444444 !important;
}

/* Autocomplete widget styling */
:deep(.monaco-editor .suggest-widget) {
  border-radius: 6px !important;
}

:deep(.monaco-editor .suggest-widget .monaco-list-row.focused) {
  background-color: var(--bg-hover) !important;
}

/* Hover widget styling */
:deep(.monaco-editor .monaco-hover) {
  border-radius: 6px !important;
}
</style>

<!-- Global styles for Monaco overflow widgets (rendered at body level) -->
<style>
.monaco-editor .overflow-guard > .overflowingContentWidgets,
body > .monaco-editor-overlaymessage,
body > .monaco-aria-container {
  z-index: 9999 !important;
}

/* Style the fixed overflow widgets */
.overflowingContentWidgets .suggest-widget {
  background: var(--bg-tertiary, #1a1a1a) !important;
  border: 1px solid var(--border-primary, #2a2a2a) !important;
  border-radius: 6px !important;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.4) !important;
}

.overflowingContentWidgets .suggest-widget .monaco-list-row.focused {
  background-color: var(--bg-hover, #242424) !important;
}

.overflowingContentWidgets .suggest-widget .monaco-list-row:hover {
  background-color: var(--bg-hover, #242424) !important;
}

.overflowingContentWidgets .monaco-hover {
  background: var(--bg-tertiary, #1a1a1a) !important;
  border: 1px solid var(--border-primary, #2a2a2a) !important;
  border-radius: 6px !important;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.4) !important;
}

.overflowingContentWidgets .monaco-hover-content {
  padding: 8px 12px !important;
}
</style>
