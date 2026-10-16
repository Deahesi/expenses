import { Dialog } from '@base-ui/react/dialog'
import type { ReactNode } from 'react'
import { Button } from './button'

interface ModalProps {
  title: string
  description?: string
  children: ReactNode
  onClose: () => void
  busy?: boolean
}

export function Modal({
  title,
  description,
  children,
  onClose,
  busy = false,
}: ModalProps) {
  return (
    <Dialog.Root
      open
      onOpenChange={(open) => {
        if (!open && !busy) onClose()
      }}
    >
      <Dialog.Portal>
        <Dialog.Backdrop className="fixed inset-0 z-50 bg-black/50" />
        <Dialog.Popup className="fixed top-1/2 left-1/2 z-50 grid max-h-[calc(100dvh-2rem)] w-[calc(100%-2rem)] max-w-lg -translate-x-1/2 -translate-y-1/2 gap-6 overflow-y-auto rounded-lg border bg-background p-6 shadow-lg outline-none">
          <div className="flex items-start justify-between gap-4">
            <div className="space-y-2">
              <Dialog.Title className="text-lg leading-none font-semibold">
                {title}
              </Dialog.Title>
              {description && (
                <Dialog.Description className="text-sm text-muted-foreground">
                  {description}
                </Dialog.Description>
              )}
            </div>
            <Button variant="ghost" size="sm" onClick={onClose} disabled={busy}>
              Закрыть
            </Button>
          </div>
          {children}
        </Dialog.Popup>
      </Dialog.Portal>
    </Dialog.Root>
  )
}
