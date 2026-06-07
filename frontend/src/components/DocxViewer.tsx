import React, { useEffect, useState } from 'react';
import mammoth from 'mammoth';

export function DocxViewer({ fileUrl }: { fileUrl: string }) {
  const [htmlContent, setHtmlContent] = useState<string>('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let isMounted = true;
    
    async function loadDocx() {
      try {
        setLoading(true);
        setError(null);
        
        // Fetch the file as an array buffer
        const response = await fetch(fileUrl);
        if (!response.ok) throw new Error('Failed to load document');
        const arrayBuffer = await response.arrayBuffer();

        // Convert to HTML using mammoth
        const result = await mammoth.convertToHtml({ arrayBuffer });
        
        if (isMounted) {
          setHtmlContent(result.value);
        }
      } catch (err: any) {
        if (isMounted) {
          setError(err.message || 'Error processing document');
        }
      } finally {
        if (isMounted) {
          setLoading(false);
        }
      }
    }

    loadDocx();

    return () => {
      isMounted = false;
    };
  }, [fileUrl]);

  if (loading) {
    return <div className="py-12 text-center text-sm text-zinc-500 flex flex-col items-center gap-3">
       <div className="w-6 h-6 border-2 border-primary/20 border-t-primary rounded-full animate-spin"></div>
       正在解析 Word 文档...
    </div>;
  }

  if (error) {
    return <div className="py-12 text-center text-sm text-rose-500 bg-rose-50 rounded-xl">
      无法预览该文档：{error}
    </div>;
  }

  return (
    <div 
      className="prose prose-sm md:prose-base prose-zinc max-w-none 
                 prose-headings:font-serif prose-headings:text-primary 
                 prose-table:border-collapse prose-table:border prose-table:border-zinc-300
                 prose-td:border prose-td:border-zinc-300 prose-td:p-2
                 prose-th:border prose-th:border-zinc-300 prose-th:p-2 prose-th:bg-zinc-100"
      dangerouslySetInnerHTML={{ __html: htmlContent }} 
    />
  );
}
