import type { ReactNode } from 'react'

interface FormFieldProps {
  label: string
  htmlFor: string
  children: ReactNode
  hint?: string
  error?: string
}

export function FormField({
  label,
  htmlFor,
  children,
  hint,
  error,
}: FormFieldProps) {
  return (
    <div className="min-w-0 space-y-2">
      <label className="text-sm leading-none font-medium" htmlFor={htmlFor}>
        {label}
      </label>
      {children}
      {hint && (
        <p id={`${htmlFor}-hint`} className="text-sm text-muted-foreground">
          {hint}
        </p>
      )}
      {error && (
        <p id={`${htmlFor}-error`} className="text-sm text-destructive">
          {error}
        </p>
      )}
    </div>
  )
}
