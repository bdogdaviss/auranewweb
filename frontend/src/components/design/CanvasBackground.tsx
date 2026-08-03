
import { useEffect, useRef } from "react";

const DOT_SPACING = 53;
const DOT_SIZE = 1.5;
const DOT_MAX_ALPHA = 0.3;
const GRAIN_PROBABILITY = 0.5;

function paintGrain(ctx: CanvasRenderingContext2D, w: number, h: number) {
  const image = ctx.createImageData(w, h);
  const data = image.data;
  for (let i = 0; i < data.length; i += 4) {
    if (Math.random() < GRAIN_PROBABILITY) {
      const v = Math.floor(Math.random() * 256);
      data[i] = v;
      data[i + 1] = v;
      data[i + 2] = v;
      data[i + 3] = 1 + Math.floor(Math.random() * 12);
    }
  }
  ctx.putImageData(image, 0, 0);
}

function paintGlows(ctx: CanvasRenderingContext2D, w: number, h: number) {
  // Center vertical glow: ellipse via scaled radial gradient
  ctx.save();
  ctx.translate(0.52 * w, 0.45 * h);
  ctx.scale(1, 2.4);
  const centerRadius = 0.14 * w;
  const centerGlow = ctx.createRadialGradient(0, 0, 0, 0, 0, centerRadius);
  centerGlow.addColorStop(0, "rgba(255,255,255,0.05)");
  centerGlow.addColorStop(1, "rgba(255,255,255,0)");
  ctx.fillStyle = centerGlow;
  ctx.fillRect(-centerRadius, -centerRadius, centerRadius * 2, centerRadius * 2);
  ctx.restore();

  // Left round glow
  const leftX = 0.27 * w;
  const leftY = 0.5 * h;
  const leftRadius = 0.22 * w;
  const leftGlow = ctx.createRadialGradient(leftX, leftY, 0, leftX, leftY, leftRadius);
  leftGlow.addColorStop(0, "rgba(255,255,255,0.04)");
  leftGlow.addColorStop(1, "rgba(255,255,255,0)");
  ctx.fillStyle = leftGlow;
  ctx.fillRect(leftX - leftRadius, leftY - leftRadius, leftRadius * 2, leftRadius * 2);
}

function paintDotGrid(ctx: CanvasRenderingContext2D, w: number, h: number) {
  for (let x = 0; x < w; x += DOT_SPACING) {
    for (let y = 0; y < h; y += DOT_SPACING) {
      ctx.fillStyle = `rgba(255,255,255,${Math.random() * DOT_MAX_ALPHA})`;
      ctx.fillRect(x, y, DOT_SIZE, DOT_SIZE);
    }
  }
}

export function CanvasBackground() {
  const ref = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    const canvas = ref.current;
    if (!canvas) return;

    const paint = () => {
      const w = window.innerWidth;
      const h = window.innerHeight;
      canvas.width = w;
      canvas.height = h;
      const ctx = canvas.getContext("2d");
      if (!ctx) return;
      paintGrain(ctx, w, h);
      paintGlows(ctx, w, h);
      paintDotGrid(ctx, w, h);
    };

    paint();
    window.addEventListener("resize", paint);
    return () => window.removeEventListener("resize", paint);
  }, []);

  return <canvas ref={ref} className="fixed inset-0 z-0 pointer-events-none" />;
}
