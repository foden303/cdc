import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Check, Copy } from 'lucide-react';

interface JsonViewerProps {
  data: unknown;
  className?: string;
}

export function JsonViewer({ data, className = '' }: JsonViewerProps) {
  const { t } = useTranslation();
  const [copied, setCopied] = useState(false);

  const formattedJson = typeof data === 'string' 
    ? data 
    : JSON.stringify(data, null, 2);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(formattedJson);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.error('Failed to copy text: ', err);
    }
  };

  // Helper to apply syntax highlighting classes using Tailwind
  const highlightSyntax = (jsonStr: string) => {
    if (!jsonStr) return '';
    
    // Escape HTML tags
    let safeStr = jsonStr
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;');

    // Regular expression for JSON tokens
    const jsonRegex = /("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+-]?\d+)?)/g;

    safeStr = safeStr.replace(jsonRegex, (match) => {
      let cls = 'text-amber-400'; // Default numeric
      if (match.startsWith('"')) {
        if (match.endsWith(':')) {
          cls = 'text-sky-400 font-medium'; // Key
        } else {
          cls = 'text-emerald-400'; // String value
        }
      } else if (/true|false/.test(match)) {
        cls = 'text-purple-400'; // Boolean
      } else if (/null/.test(match)) {
        cls = 'text-gray-500 italic'; // Null
      }
      
      // If it's a key, only style the key name, not the colon
      if (cls.includes('text-sky-400')) {
        const keyPart = match.substring(0, match.length - 1);
        return `<span class="${cls}">${keyPart}</span>:`;
      }
      return `<span class="${cls}">${match}</span>`;
    });

    return safeStr;
  };

  const highlightedHtml = highlightSyntax(formattedJson);

  return (
    <div className={`relative group rounded-lg border border-slate-800 bg-slate-950 font-mono text-xs text-slate-300 overflow-hidden ${className}`}>
      {/* Header bar with controls */}
      <div className="flex items-center justify-between px-4 py-2 border-b border-slate-900 bg-slate-900/60 select-none">
        <span className="text-[10px] uppercase tracking-wider text-slate-500 font-semibold">{t('explorer.jsonData')}</span>
        <button
          onClick={handleCopy}
          className="p-1.5 rounded-md hover:bg-slate-800 text-slate-400 hover:text-slate-200 transition-colors cursor-pointer"
          title={t('explorer.copyTooltip')}
        >
          {copied ? (
            <Check className="h-3.5 w-3.5 text-emerald-500 animate-in fade-in zoom-in-50" />
          ) : (
            <Copy className="h-3.5 w-3.5" />
          )}
        </button>
      </div>

      {/* Scroller container */}
      <div className="p-4 overflow-auto max-h-[450px] leading-relaxed">
        <pre 
          className="whitespace-pre-wrap break-all"
          dangerouslySetInnerHTML={{ __html: highlightedHtml }}
        />
      </div>
    </div>
  );
}
