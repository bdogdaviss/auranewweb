import { useState } from 'react';

type Props = {
  slides: string[]; // image URLs
  alt: string;
};

export default function Carousel({ slides, alt }: Props) {
  const [idx, setIdx] = useState(0);
  const total = slides.length;

  const go = (delta: number) => {
    setIdx((cur) => (cur + delta + total) % total);
  };

  return (
    <div className="card-image-wrap carousel-wrap">
      <div className="carousel-track-wrap">
        <div
          className="carousel-track"
          style={{ transform: `translateX(-${idx * 100}%)` }}
        >
          {slides.map((src, i) => (
            <div className="carousel-slide" key={i}>
              <img
                src={src}
                alt={`${alt} ${i + 1}`}
                onError={(e) => {
                  (e.currentTarget as HTMLImageElement).style.visibility = 'hidden';
                }}
              />
            </div>
          ))}
        </div>
      </div>
      {total > 1 && (
        <>
          <button
            className="carousel-btn carousel-prev"
            onClick={() => go(-1)}
            aria-label="Previous slide"
            type="button"
          >
            ‹
          </button>
          <button
            className="carousel-btn carousel-next"
            onClick={() => go(1)}
            aria-label="Next slide"
            type="button"
          >
            ›
          </button>
          <div className="carousel-dots" role="tablist">
            {slides.map((_, i) => (
              <span
                key={i}
                className={'carousel-dot' + (i === idx ? ' active' : '')}
                onClick={() => setIdx(i)}
                role="tab"
                aria-selected={i === idx}
              />
            ))}
          </div>
        </>
      )}
    </div>
  );
}
