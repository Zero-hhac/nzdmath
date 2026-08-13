import React from 'react';
import { Formula } from '@/src/components/Formula';

const backgroundFormulas = [
  { expression: String.raw`\nabla \cdot \vec{F} = 0`, className: 'top-[9%] left-[3%] lg:left-[6%] rotate-[-3deg] bg-pastel-blue/30 border-pastel-blue/50', delay: '0s' },
  { expression: String.raw`f(x)=\int_{0}^{x}\sin(t^2)\,dt`, className: 'top-[14%] right-[3%] lg:right-[6%] rotate-[3deg] bg-pastel-green/25 border-pastel-green/50', delay: '1.6s' },
  { expression: String.raw`\sum_{n=1}^{\infty}\frac{1}{n^2}=\frac{\pi^2}{6}`, className: 'bottom-[12%] left-[3%] lg:left-[6%] rotate-[2deg] bg-pastel-red/20 border-pastel-red/40', delay: '3.2s' },
  { expression: String.raw`\mathbb{P}(A\mid B)=\frac{\mathbb{P}(A\cap B)}{\mathbb{P}(B)}`, className: 'bottom-[10%] right-[3%] lg:right-[6%] rotate-[-3deg] bg-kinpaku/15 border-kinpaku/40', delay: '4.8s' },
];

export const MathBackground: React.FC = () => {
  return (
    <div className="pointer-events-none fixed inset-0 -z-10 overflow-hidden bg-canvas">
      {/* Warm daylight and cool corner washes for a softer, less flat canvas */}
      <div className="absolute inset-0 bg-[linear-gradient(180deg,rgba(255,235,205,0.16),transparent_42%),linear-gradient(90deg,rgba(225,243,254,0.20),transparent_58%)]" />

      {/* Minimalist geometric lines */}
      <svg
        className="absolute inset-0 h-full w-full opacity-[0.035]"
        viewBox="0 0 1440 1200"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        preserveAspectRatio="none"
      >
        <path d="M-40 240C80 210 170 130 270 150C370 170 430 310 540 328C650 346 760 238 890 226C1010 216 1110 282 1218 314C1310 342 1386 330 1480 278" stroke="#1f2a44" strokeWidth="1" />
        <path d="M122 1002C246 938 368 900 492 924C616 948 694 1034 802 1048C918 1064 1030 1000 1138 954C1246 908 1350 884 1480 920" stroke="#1f2a44" strokeWidth="1" strokeDasharray="4 8" />
      </svg>

      {/* Background Formulas as subtle pills, kept in viewport edges so they never collide with content */}
      <div className="absolute inset-0 hidden md:block">
        {backgroundFormulas.map((item) => (
          <div
            key={item.expression}
            className={`absolute rounded-full border border-border/70 backdrop-blur-sm px-3.5 py-1.5 text-xs text-text-muted opacity-70 math-float ${item.className}`}
            style={{ animationDelay: item.delay }}
          >
            <Formula expression={item.expression} />
          </div>
        ))}
      </div>

      {/* Oversized chalk-like symbols add editorial playfulness without distracting from content */}
      <div className="absolute -right-8 top-[38%] hidden select-none font-serif text-[11rem] leading-none text-pastel-blue/25 lg:block math-float" style={{ animationDelay: '2.2s' }}>
        π
      </div>
      <div className="absolute -left-10 bottom-[24%] hidden select-none font-serif text-[13rem] leading-none text-pastel-green/20 lg:block math-float" style={{ animationDelay: '4s' }}>
        Σ
      </div>
    </div>
  );
};
