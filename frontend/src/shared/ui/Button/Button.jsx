import './button.css'

export function Button({ variant = 'primary', type = 'button', children, ...props }) {
  return <button type={type} className={`btn btn--${variant}`} {...props}>{children}</button>
}
