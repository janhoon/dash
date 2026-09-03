import * as monaco from 'monaco-editor'
import { useEffect, useId, useRef, useState } from 'react'
import '@/monaco/setupWorkers'
import { registerCompletionProvider } from '@/promql/completionProvider'
import { registerHoverProvider } from '@/promql/hoverProvider'
import {
  definePromQLLightTheme,
  definePromQLTheme,
  PROMQL_LANGUAGE_ID,
  registerPromQLLanguage,
} from '@/promql/language'
import { useThemeStore } from '@/stores/themeStore'
import '@/components/monaco-query-editor.css'

type QueryLanguage = 'promql' | 'logql' | 'logsql'

function getMonacoLanguageId(language: QueryLanguage): string {
  if (language === 'logql' || language === 'logsql') {
    return language
  }
  return PROMQL_LANGUAGE_ID
}

function getCompactEditorOptions(compact: boolean) {
  return {
    lineNumbers: compact ? ('off' as const) : ('on' as const),
    padding: { top: compact ? 0 : 8, bottom: compact ? 0 : 8 },
    renderLineHighlight: compact ? ('none' as const) : ('line' as const),
    lineDecorationsWidth: compact ? 0 : 8,
    lineNumbersMinChars: compact ? 0 : 3,
  }
}

let initialized = false
function initializeMonaco() {
  if (initialized) return
  initialized = true

  registerPromQLLanguage(monaco)
  definePromQLTheme(monaco)
  definePromQLLightTheme(monaco)
  registerCompletionProvider(monaco)
  registerHoverProvider(monaco)
}

type MonacoQueryEditorProps = {
  value: string
  onChange: (value: string) => void
  onSubmit?: () => void
  disabled?: boolean
  height?: number
  placeholder?: string
  language?: QueryLanguage
  compact?: boolean
}

export function MonacoQueryEditor({
  value,
  onChange,
  onSubmit,
  disabled = false,
  height = 100,
  placeholder = 'Enter PromQL query...',
  language = 'promql',
  compact = false,
}: MonacoQueryEditorProps) {
  const containerRef = useRef<HTMLDivElement | null>(null)
  const editorRef = useRef<monaco.editor.IStandaloneCodeEditor | null>(null)
  const onChangeRef = useRef(onChange)
  const onSubmitRef = useRef(onSubmit)
  const isDark = useThemeStore((state) => state.isDark)

  useEffect(() => {
    onChangeRef.current = onChange
  }, [onChange])

  useEffect(() => {
    onSubmitRef.current = onSubmit
  }, [onSubmit])
  const [isFocused, setIsFocused] = useState(false)
  const placeholderId = useId()
  const showPlaceholder = !value && !isFocused

  // biome-ignore lint/correctness/useExhaustiveDependencies: Monaco editor is created once; theme/language/value sync in separate effects
  useEffect(() => {
    if (!containerRef.current) return

    initializeMonaco()

    const editor = monaco.editor.create(containerRef.current, {
      value,
      language: getMonacoLanguageId(language),
      theme: isDark ? 'promql-dark' : 'promql-light',
      minimap: { enabled: false },
      ...getCompactEditorOptions(compact),
      wordWrap: 'on',
      scrollBeyondLastLine: false,
      automaticLayout: true,
      fontSize: 13,
      fontFamily: "'JetBrains Mono', 'Menlo', 'Ubuntu Mono', monospace",
      lineHeight: 20,
      folding: false,
      glyphMargin: false,
      overviewRulerBorder: false,
      hideCursorInOverviewRuler: true,
      fixedOverflowWidgets: true,
      scrollbar: {
        vertical: 'auto',
        horizontal: 'auto',
        verticalScrollbarSize: 8,
        horizontalScrollbarSize: 8,
      },
      suggest: {
        showIcons: true,
        showStatusBar: true,
        preview: true,
        previewMode: 'prefix',
      },
      quickSuggestions: {
        other: true,
        comments: false,
        strings: true,
      },
      acceptSuggestionOnEnter: 'on',
      tabCompletion: 'on',
      readOnly: disabled,
    })

    editorRef.current = editor

    const contentDisposable = editor.onDidChangeModelContent(() => {
      const nextValue = editor.getValue()
      onChangeRef.current(nextValue)
    })

    const focusDisposable = editor.onDidFocusEditorText(() => {
      setIsFocused(true)
    })

    const blurDisposable = editor.onDidBlurEditorText(() => {
      setIsFocused(false)
    })

    editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.Enter, () => {
      onSubmitRef.current?.()
    })

    return () => {
      contentDisposable.dispose()
      focusDisposable.dispose()
      blurDisposable.dispose()
      editor.dispose()
      editorRef.current = null
    }
  }, [])

  useEffect(() => {
    editorRef.current?.updateOptions({ readOnly: disabled })
  }, [disabled])

  useEffect(() => {
    editorRef.current?.updateOptions(getCompactEditorOptions(compact))
  }, [compact])

  useEffect(() => {
    if (editorRef.current && editorRef.current.getValue() !== value) {
      editorRef.current.setValue(value)
    }
  }, [value])

  useEffect(() => {
    monaco.editor.setTheme(isDark ? 'promql-dark' : 'promql-light')
  }, [isDark])

  useEffect(() => {
    const model = editorRef.current?.getModel()
    if (model) {
      monaco.editor.setModelLanguage(model, getMonacoLanguageId(language))
    }
  }, [language])

  // biome-ignore lint/correctness/useExhaustiveDependencies: relayout when editor height prop changes
  useEffect(() => {
    editorRef.current?.layout()
  }, [height])

  return (
    <div
      className={`relative overflow-hidden rounded-md bg-[var(--color-surface-container-high)] transition-colors duration-[var(--motion-fast)] ease-[var(--ease-standard)] focus-within:shadow-[var(--shadow-focus)] ${compact ? 'px-3 py-3' : ''} ${disabled ? 'pointer-events-none opacity-60' : ''}`}
      data-testid="monaco-query-editor"
    >
      <div
        ref={containerRef}
        className={`w-full ${compact ? 'min-h-[20px]' : 'min-h-[60px]'}`}
        style={{ height: `${height}px` }}
      />
      {showPlaceholder ? (
        <div
          id={placeholderId}
          className={`pointer-events-none absolute font-mono text-[13px] text-[var(--color-outline)] ${compact ? 'top-3 left-3' : 'top-2 left-12'}`}
        >
          {placeholder}
        </div>
      ) : null}
    </div>
  )
}
