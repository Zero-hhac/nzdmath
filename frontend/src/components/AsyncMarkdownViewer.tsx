import React, { useEffect, useState } from 'react';
import { MarkdownViewer } from './MarkdownViewer';
import { api } from '@/src/lib/api';

type Props = {
  type: 'events' | 'news' | 'showcases';
  id: number;
  initialContent?: string;
};

export function AsyncMarkdownViewer({ type, id, initialContent }: Props) {
  const [content, setContent] = useState(initialContent || '');
  const [loading, setLoading] = useState(!initialContent);

  useEffect(() => {
    if (initialContent) return;
    let isMounted = true;
    
    const fetcher = 
      type === 'events' ? api.getEvent : 
      type === 'news' ? api.getNewsDetail : 
      api.getShowcase;
      
    fetcher(id)
      .then((res: any) => {
        if (isMounted) {
          setContent(res.data?.content || '');
          setLoading(false);
        }
      })
      .catch(() => {
        if (isMounted) setLoading(false);
      });
      
    return () => {
      isMounted = false;
    };
  }, [type, id, initialContent]);

  if (loading) {
    return (
      <div className="py-12 flex flex-col items-center justify-center gap-3 text-zinc-400">
        <div className="w-5 h-5 border-2 border-primary/20 border-t-primary rounded-full animate-spin"></div>
        <span className="text-xs tracking-wider">LOADING CONTENT...</span>
      </div>
    );
  }

  if (!content) {
    return (
      <div className="py-12 text-center text-sm text-zinc-400">
        暂无详细内容
      </div>
    );
  }

  return <MarkdownViewer content={content} />;
}
