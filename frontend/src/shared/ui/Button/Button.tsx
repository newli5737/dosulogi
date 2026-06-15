import type { ButtonHTMLAttributes, ReactNode } from 'react'
import './button.css'

type ButtonVariant = 'primary' | 'secondary'

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant
  children: ReactNode
}

export function Button({ variant = 'primary', type = 'button', children, ...props }: ButtonProps) {
  return <button type={type} className={`btn btn--${variant}`} {...props}>{children}</button>
}
