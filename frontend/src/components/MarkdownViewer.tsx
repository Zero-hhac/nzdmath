import React from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import remarkMath from 'remark-math';
import remarkBreaks from 'remark-breaks';
import rehypeKatex from 'rehype-katex';
import rehypeRaw from 'rehype-raw';
import rehypeSanitize, { defaultSchema } from 'rehype-sanitize';

const markdownSchema = {
  ...defaultSchema,
  attributes: {
    ...defaultSchema.attributes,
    '*': [...(defaultSchema.attributes?.['*'] || []), 'className', 'style'],
  },
};

export function MarkdownViewer({ content }: { content?: string }) {
  if (!content) return null;

  return (
    <div className="prose prose-sm md:prose-base prose-zinc max-w-none 
                    prose-headings:font-serif prose-headings:text-primary prose-headings:tracking-normal
                    prose-a:text-blue-600 prose-a:no-underline hover:prose-a:underline 
                    prose-blockquote:border-l-4 prose-blockquote:border-zinc-300 prose-blockquote:bg-zinc-50 prose-blockquote:px-4 prose-blockquote:py-1 prose-blockquote:rounded-r-lg prose-blockquote:not-italic
                    prose-code:text-zinc-800 prose-code:bg-zinc-100 prose-code:px-1.5 prose-code:py-0.5 prose-code:rounded-md prose-code:before:content-none prose-code:after:content-none
                    prose-pre:bg-zinc-50 prose-pre:text-zinc-800 prose-pre:rounded-xl prose-pre:border prose-pre:border-zinc-200/80
                    prose-img:rounded-2xl prose-img:shadow-md
                    prose-table:border-collapse prose-th:bg-zinc-100 prose-th:p-2 prose-td:p-2">
      <ReactMarkdown 
        remarkPlugins={[remarkGfm, remarkMath, remarkBreaks]}
        rehypePlugins={[rehypeRaw, [rehypeSanitize, markdownSchema], rehypeKatex]}
      >
        {content}
      </ReactMarkdown>
    </div>
  );
}
