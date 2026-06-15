import * as React from 'react';
import { cva, type VariantProps } from 'class-variance-authority';
import { cn } from '../../lib/utils';

// shadcn/ui-style Button, themed to Aura (emerald/dark). Export buttonVariants
// so it can also style <a> links (the shadcn pattern for link-buttons).
const buttonVariants = cva(
  'inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-xl font-semibold transition-all duration-200 will-change-transform focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-aura-400/60 focus-visible:ring-offset-2 focus-visible:ring-offset-[#0a0a0f] disabled:pointer-events-none disabled:opacity-50 active:translate-y-0',
  {
    variants: {
      variant: {
        default:
          'bg-gradient-to-b from-aura-400 to-aura-600 text-white shadow-[0_6px_20px_-4px_rgba(16,185,129,0.5)] hover:from-aura-300 hover:to-aura-500 hover:-translate-y-0.5 hover:shadow-[0_12px_28px_-6px_rgba(16,185,129,0.65)]',
        secondary:
          'bg-white/[0.06] text-white border border-white/10 hover:bg-white/[0.12] hover:border-white/20',
        outline:
          'border border-aura-500/50 text-aura-300 hover:bg-aura-500/10 hover:border-aura-400 hover:text-aura-200',
        ghost: 'text-gray-300 hover:bg-white/[0.06] hover:text-white',
        destructive:
          'bg-gradient-to-b from-red-500 to-red-600 text-white shadow-[0_6px_20px_-4px_rgba(239,68,68,0.5)] hover:-translate-y-0.5',
        discord:
          'bg-[#5865F2] text-white shadow-[0_6px_20px_-4px_rgba(88,101,242,0.5)] hover:bg-[#4752C4] hover:-translate-y-0.5',
      },
      size: {
        default: 'h-11 px-6 text-sm',
        sm: 'h-9 px-4 text-xs',
        lg: 'h-12 px-8 text-base',
        icon: 'h-11 w-11',
      },
    },
    defaultVariants: { variant: 'default', size: 'default' },
  },
);

export interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof buttonVariants> {}

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant, size, ...props }, ref) => (
    <button ref={ref} className={cn(buttonVariants({ variant, size }), className)} {...props} />
  ),
);
Button.displayName = 'Button';

export { Button, buttonVariants };
