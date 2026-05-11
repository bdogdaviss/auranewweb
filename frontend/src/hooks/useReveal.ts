import { useEffect, useRef } from 'react';

// useReveal returns a ref + adds an `.is-visible` class to the element the
// first time it intersects the viewport. Combine with the `.reveal-on-scroll`
// CSS class for a fade-in-from-below entrance. Respects prefers-reduced-motion
// (immediately marks visible, no animation).
//
//   const ref = useReveal<HTMLDivElement>();
//   return <div ref={ref} className="reveal-on-scroll">...</div>;
//
// The IntersectionObserver is disconnected on unmount so we don't leak
// observers across unmounts/HMR.
export function useReveal<T extends Element = HTMLDivElement>(
  options: { threshold?: number; rootMargin?: string } = {}
) {
  const ref = useRef<T | null>(null);

  useEffect(() => {
    const el = ref.current;
    if (!el) return;

    if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) {
      el.classList.add('is-visible');
      return;
    }

    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            entry.target.classList.add('is-visible');
            observer.unobserve(entry.target);
          }
        });
      },
      {
        threshold: options.threshold ?? 0.1,
        rootMargin: options.rootMargin ?? '0px 0px -40px 0px',
      }
    );

    observer.observe(el);
    return () => observer.disconnect();
  }, [options.threshold, options.rootMargin]);

  return ref;
}
