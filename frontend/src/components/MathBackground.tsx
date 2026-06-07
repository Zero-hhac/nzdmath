import React from 'react';
import { Formula } from '@/src/components/Formula';

const backgroundFormulas = [
  { expression: String.raw`\nabla \cdot \vec{F} = 0`, className: 'top-[12%] left-[8%] rotate-[-2deg]' },
  { expression: String.raw`f(x)=\int_{0}^{x}\sin(t^2)\,dt`, className: 'top-[18%] right-[10%] rotate-[3deg]' },
  { expression: String.raw`e^{i\pi}+1=0`, className: 'top-[54%] right-[12%] rotate-[-4deg]' },
  { expression: String.raw`\sum_{n=1}^{\infty}\frac{1}{n^2}=\frac{\pi^2}{6}`, className: 'bottom-[16%] left-[9%] rotate-[2deg]' },
  { expression: String.raw`\mathbb{P}(A\mid B)=\frac{\mathbb{P}(A\cap B)}{\mathbb{P}(B)}`, className: 'bottom-[10%] right-[16%] rotate-[-3deg]' },
];

export const MathBackground: React.FC = () => {
  return (
    <div className="pointer-events-none fixed inset-0 -z-10 overflow-hidden bg-canvas">
      {/* Subtle radial light spot for depth without breaking minimal aesthetic */}
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_top_left,_rgba(212,175,55,0.03),_transparent_40%),radial-gradient(circle_at_80%_80%,_rgba(0,0,0,0.02),_transparent_50%)]" />

      {/* Minimalist geometric lines */}
      <svg
        className="absolute inset-0 h-full w-full opacity-[0.04]"
        viewBox="0 0 1440 1200"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        preserveAspectRatio="none"
      >
        <path d="M-40 240C80 210 170 130 270 150C370 170 430 310 540 328C650 346 760 238 890 226C1010 216 1110 282 1218 314C1310 342 1386 330 1480 278" stroke="#111111" strokeWidth="1" />
        <path d="M122 1002C246 938 368 900 492 924C616 948 694 1034 802 1048C918 1064 1030 1000 1138 954C1246 908 1350 884 1480 920" stroke="#111111" strokeWidth="1" strokeDasharray="4 8" />
      </svg>

      {/* Background Formulas as subtle pills */}
      <div className="absolute inset-0">
        {backgroundFormulas.map((item) => (
          <div
            key={item.expression}
            className={`absolute rounded-full border border-border bg-surface px-4 py-2 text-sm text-text-faint shadow-[0_2px_8px_rgba(0,0,0,0.02)] ${item.className}`}
          >
            <Formula expression={item.expression} />
          </div>
        ))}
      </div>
    </div>
  );
};
