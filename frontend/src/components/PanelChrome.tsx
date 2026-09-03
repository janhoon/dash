import { Pencil, Trash2 } from 'lucide-react'
import type { ReactNode } from 'react'

export const WIDE_PANEL_COLS = 8

export function panelHeaderMeta({
  gridWidth,
  queryExpr,
  selectedPreset,
  isCustomRange,
}: {
  gridWidth: number
  queryExpr: string
  selectedPreset: string
  isCustomRange: boolean
}): string {
  if (gridWidth >= WIDE_PANEL_COLS && queryExpr) return queryExpr
  if (isCustomRange) return ''
  return selectedPreset
}

type PanelChromeProps = {
  panelId: string
  title: string
  wide: boolean
  headerMeta: string
  onEdit?: () => void
  onDelete?: () => void
  children: ReactNode
}

export function PanelChrome({
  panelId,
  title,
  wide,
  headerMeta,
  onEdit,
  onDelete,
  children,
}: PanelChromeProps) {
  return (
    <div
      className="relative flex h-full flex-col gap-2 overflow-hidden rounded-lg p-[var(--panel-padding)]"
      style={{
        backgroundColor: 'var(--color-surface-container-low)',
        borderWidth: '1px',
        borderStyle: 'solid',
        borderColor: 'var(--color-stroke-subtle)',
      }}
      data-testid={`dashboard-panel-${panelId}`}
    >
      <div className="panel-header flex items-center justify-between gap-2">
        <h3
          className={`min-w-0 truncate font-medium ${wide ? 'text-[13px]' : 'text-xs'}`}
          style={{
            color: wide ? 'var(--color-on-surface)' : 'var(--color-on-surface-variant)',
          }}
          data-testid="panel-title"
        >
          {title}
        </h3>
        <div className="flex min-w-0 shrink-0 items-center gap-1">
          {headerMeta ? (
            <span
              className={`max-w-[220px] truncate ${wide ? 'font-mono text-[11px]' : 'text-[11px]'}`}
              style={{ color: 'var(--color-on-surface-variant)' }}
              data-testid="panel-header-meta"
              title={headerMeta}
            >
              {headerMeta}
            </span>
          ) : null}
          {(onEdit || onDelete) && (
            <div className="panel-actions flex gap-1">
              {onEdit ? (
                <button
                  type="button"
                  className="flex h-7 w-7 cursor-pointer items-center justify-center rounded-md border-0 bg-transparent transition hover:opacity-80"
                  style={{ color: 'var(--color-outline)' }}
                  data-testid="panel-edit-btn"
                  title="Edit panel"
                  onClick={onEdit}
                >
                  <Pencil size={16} />
                </button>
              ) : null}
              {onDelete ? (
                <button
                  type="button"
                  className="flex h-7 w-7 cursor-pointer items-center justify-center rounded-md border-0 bg-transparent transition hover:opacity-80"
                  style={{ color: 'var(--color-outline)' }}
                  data-testid="panel-delete-btn"
                  title="Delete panel"
                  onClick={onDelete}
                >
                  <Trash2 size={16} />
                </button>
              ) : null}
            </div>
          )}
        </div>
      </div>

      <div className="flex min-h-0 flex-1 flex-col overflow-hidden">{children}</div>
    </div>
  )
}
