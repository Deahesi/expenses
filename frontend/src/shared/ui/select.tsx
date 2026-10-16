import type { ComponentProps } from 'react'
import { cn } from '@/shared/lib/cn'

type SelectProps = ComponentProps<'select'>

export function Select({ className, ...props }: SelectProps) {
  return (
    <select
      className={cn(
        'h-8 w-full min-w-0 appearance-none rounded-lg border border-input bg-transparent px-2.5 py-1 text-sm outline-none transition-colors focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 disabled:opacity-50 aria-invalid:border-destructive',
        className,
      )}
      {...props}
    />
  )
}
