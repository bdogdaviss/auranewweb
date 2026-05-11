// Separator — Radix-Themes-style divider. Plain CSS class composition;
// see theme.css under "Separator" for the actual styling. Defaults to a
// horizontal, full-width, faintly-tinted gray line.
//
// Props mirror the Radix component:
//   orientation: 'horizontal' | 'vertical'
//   size:        1 | 2 | 3 | 4   (25 / 50 / 75 / 100% of container)
//   color:       'gray' | 'green' | 'cyan' | 'orange' | 'crimson' | 'indigo'

type Orientation = 'horizontal' | 'vertical';
type Size = 1 | 2 | 3 | 4;
type Color = 'gray' | 'green' | 'cyan' | 'orange' | 'crimson' | 'indigo';

export default function Separator({
  orientation = 'horizontal',
  size = 4,
  color = 'gray',
  className = '',
}: {
  orientation?: Orientation;
  size?: Size;
  color?: Color;
  className?: string;
}) {
  return (
    <div
      role="separator"
      aria-orientation={orientation}
      className={`sep sep-${orientation} sep-size-${size} sep-${color} ${className}`.trim()}
    />
  );
}
