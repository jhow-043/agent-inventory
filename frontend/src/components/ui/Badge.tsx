import type { ReactNode } from 'react';

type BadgeVariant = 'accent' | 'success' | 'danger' | 'warning' | 'info' | 'neutral';

interface BadgeProps {
  variant?: BadgeVariant;
  children: ReactNode;
  dot?: boolean;
  pulseDot?: boolean;
  className?: string;
}

const variantClasses: Record<BadgeVariant, { bg: string; dot: string }> = {
  accent: { bg: 'bg-accent/15 text-accent', dot: 'bg-accent' },
  success: { bg: 'bg-success/15 text-success', dot: 'bg-success' },
  danger: { bg: 'bg-danger/15 text-danger', dot: 'bg-danger' },
  warning: { bg: 'bg-warning/15 text-warning', dot: 'bg-warning' },
  info: { bg: 'bg-info/15 text-info', dot: 'bg-info' },
  neutral: { bg: 'bg-bg-tertiary text-text-muted', dot: 'bg-text-muted' },
};

export default function Badge({ variant = 'neutral', children, dot, pulseDot, className = '' }: BadgeProps) {
  const style = variantClasses[variant];

  return (
    <span
      className={`inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium ${style.bg} ${className}`}
    >
      {dot && (
        <span
          className={`w-1.5 h-1.5 rounded-full ${style.dot} ${pulseDot ? 'animate-[pulse-dot_2s_ease-in-out_infinite]' : ''}`}
        />
      )}
      {children}
    </span>
  );
}
