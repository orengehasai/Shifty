import type { ButtonHTMLAttributes } from 'react';

interface Props extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'danger';
  size?: 'sm' | 'md';
}

const base: React.CSSProperties = {
  border: 'none',
  borderRadius: 6,
  cursor: 'pointer',
  fontWeight: 500,
  fontSize: 14,
  transition: 'opacity 0.15s',
};

const variants: Record<string, React.CSSProperties> = {
  primary: { background: '#4F46E5', color: '#fff', padding: '8px 16px' },
  secondary: { background: '#E5E7EB', color: '#374151', padding: '8px 16px' },
  danger: { background: '#EF4444', color: '#fff', padding: '8px 16px' },
};

const sizes: Record<string, React.CSSProperties> = {
  sm: { padding: '4px 12px', fontSize: 13 },
  md: {},
};

export default function Button({ variant = 'primary', size = 'md', style, disabled, ...rest }: Props) {
  return (
    <button
      style={{
        ...base,
        ...variants[variant],
        ...sizes[size],
        ...(disabled ? { opacity: 0.5, cursor: 'not-allowed' } : {}),
        ...style,
      }}
      disabled={disabled}
      {...rest}
    />
  );
}
