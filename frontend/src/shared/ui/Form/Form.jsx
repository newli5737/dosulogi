import './form.css'

export function Field({ label, children, required }) {
  return (
    <label className="field">
      <span className="field-label">{label}{required && ' *'}</span>
      {children}
    </label>
  )
}

export function Input(props) {
  return <input className="field-input" {...props} />
}

export function Select({ children, ...props }) {
  return <select className="field-input" {...props}>{children}</select>
}

export function Textarea(props) {
  return <textarea className="field-input field-textarea" rows={3} {...props} />
}
