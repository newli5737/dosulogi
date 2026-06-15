import { createPortal } from 'react-dom'
import type { ReactNode } from 'react'
import './modal.css'

interface ModalProps {
  open: boolean
  onClose: () => void
  title: string
  children: ReactNode
  wide?: boolean
}

export function Modal({ open, onClose, title, children, wide }: ModalProps) {
  if (!open) return null
  return createPortal(
    <div className="modal-backdrop" onClick={onClose} role="presentation">
      <div className={`modal-panel ${wide ? 'modal-panel--wide' : ''}`} onClick={(e) => e.stopPropagation()} role="dialog" aria-modal="true">
        <header className="modal-header">
          <h3>{title}</h3>
          <button type="button" className="modal-close" onClick={onClose} aria-label="Đóng">×</button>
        </header>
        <div className="modal-body">{children}</div>
      </div>
    </div>,
    document.body,
  )
}
