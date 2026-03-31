import { useState } from 'react';
import { Search, Replace, CaseSensitive, Regex, ArrowDownUp } from 'lucide-react';

export default function SearchPanel() {
  const [query, setQuery] = useState('');
  const [replaceText, setReplaceText] = useState('');
  const [showReplace, setShowReplace] = useState(false);
  const [caseSensitive, setCaseSensitive] = useState(false);
  const [useRegex, setUseRegex] = useState(false);

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center justify-between px-4 py-2 text-xs uppercase tracking-wider text-[var(--text-secondary)] font-semibold">
        Search
      </div>
      <div className="px-3 space-y-2">
        <div className="flex items-center gap-1">
          <button
            onClick={() => setShowReplace(!showReplace)}
            className="text-[var(--text-secondary)] hover:text-[var(--text-primary)] p-0.5"
            title="Toggle Replace"
          >
            <ArrowDownUp size={14} />
          </button>
          <div className="flex-1 flex items-center bg-[var(--bg-tertiary)] border border-[var(--border)] rounded">
            <Search size={14} className="ml-2 text-[var(--text-secondary)]" />
            <input
              type="text"
              placeholder="Search"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              className="flex-1 bg-transparent px-2 py-1 text-sm text-[var(--text-primary)] outline-none"
            />
            <button
              onClick={() => setCaseSensitive(!caseSensitive)}
              className={`p-1 ${caseSensitive ? 'text-[var(--text-active)] bg-[var(--bg-active)]' : 'text-[var(--text-secondary)]'}`}
              title="Match Case"
            >
              <CaseSensitive size={14} />
            </button>
            <button
              onClick={() => setUseRegex(!useRegex)}
              className={`p-1 ${useRegex ? 'text-[var(--text-active)] bg-[var(--bg-active)]' : 'text-[var(--text-secondary)]'}`}
              title="Use Regex"
            >
              <Regex size={14} />
            </button>
          </div>
        </div>
        {showReplace && (
          <div className="ml-5">
            <div className="flex items-center bg-[var(--bg-tertiary)] border border-[var(--border)] rounded">
              <Replace size={14} className="ml-2 text-[var(--text-secondary)]" />
              <input
                type="text"
                placeholder="Replace"
                value={replaceText}
                onChange={(e) => setReplaceText(e.target.value)}
                className="flex-1 bg-transparent px-2 py-1 text-sm text-[var(--text-primary)] outline-none"
              />
            </div>
          </div>
        )}
        {query && (
          <div className="text-xs text-[var(--text-secondary)] px-1">
            No results found.
          </div>
        )}
      </div>
    </div>
  );
}
