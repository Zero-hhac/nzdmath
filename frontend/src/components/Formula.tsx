import React from 'react';
import katex from 'katex';

interface FormulaProps {
  expression: string;
  block?: boolean;
  className?: string;
}

export const Formula: React.FC<FormulaProps> = ({ expression, block = false, className }) => {
  const html = katex.renderToString(expression, {
    throwOnError: false,
    displayMode: block,
    strict: 'ignore',
  });

  return <span className={className} dangerouslySetInnerHTML={{ __html: html }} />;
};
