import { render } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { MonacoQueryEditor } from '@/components/MonacoQueryEditor'

const { create, updateOptions } = vi.hoisted(() => {
  const updateOptions = vi.fn()
  const create = vi.fn((_container: unknown, _options?: Record<string, unknown>) => ({
    updateOptions,
    getValue: vi.fn(() => ''),
    setValue: vi.fn(),
    getModel: vi.fn(() => null),
    layout: vi.fn(),
    dispose: vi.fn(),
    onDidChangeModelContent: vi.fn(() => ({ dispose: vi.fn() })),
    onDidFocusEditorText: vi.fn(() => ({ dispose: vi.fn() })),
    onDidBlurEditorText: vi.fn(() => ({ dispose: vi.fn() })),
    addCommand: vi.fn(),
  }))
  return { create, updateOptions }
})

vi.mock('monaco-editor', () => ({
  editor: {
    create,
    setTheme: vi.fn(),
    setModelLanguage: vi.fn(),
  },
  KeyMod: { CtrlCmd: 2048 },
  KeyCode: { Enter: 3 },
}))

vi.mock('@/monaco/setupWorkers', () => ({}))
vi.mock('@/promql/completionProvider', () => ({ registerCompletionProvider: vi.fn() }))
vi.mock('@/promql/hoverProvider', () => ({ registerHoverProvider: vi.fn() }))
vi.mock('@/promql/language', () => ({
  PROMQL_LANGUAGE_ID: 'promql',
  registerPromQLLanguage: vi.fn(),
  definePromQLTheme: vi.fn(),
  definePromQLLightTheme: vi.fn(),
}))

describe('MonacoQueryEditor', () => {
  beforeEach(() => {
    create.mockClear()
    updateOptions.mockClear()
  })

  it('syncs compact monaco options when the compact prop changes', () => {
    const { rerender } = render(
      <MonacoQueryEditor value="up" onChange={() => undefined} compact height={40} />,
    )

    expect(create).toHaveBeenCalledTimes(1)
    expect(create).toHaveBeenCalledWith(
      expect.anything(),
      expect.objectContaining({
        lineNumbers: 'off',
        renderLineHighlight: 'none',
        lineDecorationsWidth: 0,
        lineNumbersMinChars: 0,
        padding: { top: 0, bottom: 0 },
      }),
    )

    rerender(
      <MonacoQueryEditor value="up" onChange={() => undefined} compact={false} height={160} />,
    )

    expect(create).toHaveBeenCalledTimes(1)
    expect(updateOptions).toHaveBeenCalledWith({
      lineNumbers: 'on',
      renderLineHighlight: 'line',
      lineDecorationsWidth: 8,
      lineNumbersMinChars: 3,
      padding: { top: 8, bottom: 8 },
    })
  })
})
