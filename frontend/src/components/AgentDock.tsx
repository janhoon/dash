import { useEffect, useState } from 'react'
import { Button } from '@/components/ui/button'
import { useKeyboardShortcutsStore } from '@/lib/keyboardShortcuts'
import { useAiSidebarStore } from '@/stores/aiSidebarStore'

export const AGENT_DOCK_WIDTH_PX = 380

export type AgentDockTurn =
  | { role: 'user'; id: string; body: string }
  | { role: 'ace'; id: string; body: string; pinLabel?: string }

type AgentDockProps = {
  thread?: AgentDockTurn[]
  overlayOpen?: boolean
}

export const PLACEHOLDER_THREAD: AgentDockTurn[] = [
  { role: 'user', id: 'you-p99', body: 'Why did p99 move after 16:00?' },
  {
    role: 'ace',
    id: 'ace-pay',
    body: 'Checkout deploy at 16:12. Latency by route shows /pay jumped 80ms. Error ratio flat. Pin that series — do not make a new board.',
    pinLabel: 'Pin /pay p99',
  },
]

function AgentDockTurnView({ turn }: { turn: AgentDockTurn }) {
  switch (turn.role) {
    case 'user':
      return (
        <div
          className="flex w-full shrink-0 flex-col gap-1 rounded-lg"
          style={{
            backgroundColor: 'var(--color-surface-container-high)',
            padding: '10px',
          }}
        >
          <p
            className="font-medium"
            style={{ fontSize: '11px', color: 'var(--color-on-surface-variant)' }}
          >
            You
          </p>
          <p style={{ fontSize: '13px', color: 'var(--color-on-surface)' }}>{turn.body}</p>
        </div>
      )
    case 'ace':
      return (
        <div
          className="flex w-full shrink-0 flex-col gap-1 rounded-lg"
          style={{
            backgroundColor: 'var(--color-surface)',
            border: '1px solid var(--color-surface-container-high)',
            padding: '10px',
          }}
        >
          <p className="font-medium" style={{ fontSize: '11px', color: 'var(--color-primary)' }}>
            Ace
          </p>
          <p style={{ fontSize: '13px', color: 'var(--color-on-surface)' }}>{turn.body}</p>
          {turn.pinLabel ? (
            <Button
              type="button"
              className="h-auto self-start rounded-lg px-2 py-2 text-xs font-semibold"
            >
              {turn.pinLabel}
            </Button>
          ) : null}
        </div>
      )
    default: {
      const _exhaustive: never = turn
      return _exhaustive
    }
  }
}

export function AgentDock({ thread = [], overlayOpen = false }: AgentDockProps) {
  const isOpen = useAiSidebarStore((state) => state.isOpen)
  const pendingContext = useAiSidebarStore((state) => state.pendingContext)
  const consumePendingContext = useAiSidebarStore((state) => state.consumePendingContext)
  const registerShortcut = useKeyboardShortcutsStore((state) => state.register)
  const showHelp = useKeyboardShortcutsStore((state) => state.showHelp)
  const [pendingPreview, setPendingPreview] = useState<string | null>(null)

  useEffect(() => {
    if (!isOpen) {
      setPendingPreview(null)
      return
    }
    if (!pendingContext) return
    const pending = consumePendingContext()
    if (pending?.message) setPendingPreview(pending.message)
  }, [isOpen, pendingContext, consumePendingContext])

  useEffect(() => {
    if (!isOpen || overlayOpen || showHelp) return
    return registerShortcut(
      'Escape',
      () => {
        useAiSidebarStore.getState().close()
      },
      'Close agent dock',
      'General',
    )
  }, [isOpen, overlayOpen, showHelp, registerShortcut])

  if (!isOpen) return null

  const previewMessage = pendingPreview ?? pendingContext?.message ?? null

  return (
    <aside
      data-testid="agent-dock"
      aria-label="Agent"
      className="fixed top-0 right-0 bottom-0 z-40 flex flex-col"
      style={{
        width: AGENT_DOCK_WIDTH_PX,
        backgroundColor: 'var(--color-surface-container-low)',
        padding: 'var(--panel-padding)',
        gap: '12px',
        fontFamily: 'var(--font-sans)',
      }}
    >
      <div className="flex w-full shrink-0 items-start justify-between">
        <p className="font-semibold" style={{ fontSize: '14px', color: 'var(--color-on-surface)' }}>
          Agent
        </p>
        <p
          className="font-medium"
          style={{
            fontSize: '11px',
            color: 'var(--color-on-surface-variant)',
            fontFamily: 'var(--font-mono)',
          }}
        >
          ⌘J
        </p>
      </div>

      <div
        data-testid="agent-dock-thread"
        className="flex min-h-0 flex-1 flex-col gap-3 overflow-y-auto"
      >
        {previewMessage ? (
          <AgentDockTurnView turn={{ role: 'user', id: 'pending-context', body: previewMessage }} />
        ) : null}
        {thread.map((turn) => (
          <AgentDockTurnView key={turn.id} turn={turn} />
        ))}
      </div>

      <input
        type="text"
        readOnly
        aria-label="Ask about this board"
        placeholder="Ask about this board…"
        className="w-full shrink-0 rounded-lg border-0 outline-none placeholder:text-on-surface-variant"
        style={{
          backgroundColor: 'var(--color-surface-container-high)',
          padding: '10px',
          fontSize: '12px',
          color: 'var(--color-on-surface)',
          fontFamily: 'var(--font-sans)',
        }}
      />
    </aside>
  )
}
