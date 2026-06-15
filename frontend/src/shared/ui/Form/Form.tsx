import type { ReactNode, InputHTMLAttributes, SelectHTMLAttributes, TextareaHTMLAttributes } from 'react'
import './form.css'

interface FieldProps {
  label: string
  children: ReactNode
  required?: boolean
}

export function Field({ label, children, required }: FieldProps) {
  return (
    <label className="field">
      <span className="field-label">{label}{required && ' *'}</span>
      {children}
    </label>
  )
}

export function Input(props: InputHTMLAttributes<HTMLInputElement>) {
  return <input className="field-input" {...props} />
}

export function Select({ children, ...props }: SelectHTMLAttributes<HTMLSelectElement>) {
  return <select className="field-input" {...props}>{children}</select>
}

export function Textarea(props: TextareaHTMLAttributes<HTMLTextAreaElement>) {
  return <textarea className="field-input field-textarea" rows={3} {...props} />
}
