import { Button } from './button'

interface ErrorNoticeProps {
  message: string
  onRetry?: () => void
}

export function ErrorNotice({ message, onRetry }: ErrorNoticeProps) {
  return (
    <div className="flex flex-wrap items-center gap-4 text-sm" role="alert">
      <p>{message}</p>
      {onRetry && (
        <Button variant="outline" size="sm" onClick={onRetry}>
          Повторить
        </Button>
      )}
    </div>
  )
}
