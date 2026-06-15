import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

// cn merges class strings and de-dupes conflicting Tailwind utilities — the
// standard shadcn/ui helper.
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}
