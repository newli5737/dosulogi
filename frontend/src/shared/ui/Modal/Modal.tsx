import { createPortal } from 'react-dom'
import type { LucideIcon } from 'lucide-react'
import { X } from 'lucide-react'
import type { ReactNode } from 'react'
import './modal.css'

interface ModalProps {
  open: boolean
  onClose: () => void
  title: string
  children: ReactNode
  wide?: boolean
  xlarge?: boolean
  icon?: LucideIcon
  tone?: 'blue' | 'green' | 'amber' | 'rose' | 'violet' | 'cyan'
}

export function Modal({ open, onClose, title, children, wide, xlarge, icon: Icon, tone = 'blue' }: ModalProps) {
  if (!open) return null
  const sizeClass = xlarge ? 'modal-panel--xlarge' : wide ? 'modal-panel--wide' : ''
  return createPortal(
    <div className="modal-backdrop" onClick={onClose} role="presentation">
      <div className={`modal-panel ${sizeClass}`} onClick={(e) => e.stopPropagation()} role="dialog" aria-modal="true">
        <header className="modal-header">
          <div className="modal-header__title">
            {Icon && (
              <span className={`modal-icon modal-icon--${tone}`} aria-hidden>
                <Icon size={18} strokeWidth={2.2} />
              </span>
            )}
            <h3>{title}</h3>
          </div>
          <button type="button" className="modal-close" onClick={onClose} aria-label="Đóng">
            <X size={18} />
          </button>
        </header>
        <div className="modal-body">{children}</div>
      </div>
    </div>,
    document.body,
  )
}
