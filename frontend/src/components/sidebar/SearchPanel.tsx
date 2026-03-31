import { useState, useMemo, useCallback } from 'react';
import { Search, Replace, CaseSensitive, Regex, ArrowDownUp, FileText } from 'lucide-react';
import { useWorkspace } from '../../context/WorkspaceContext';
import type { EditorTab } from '../../types';

interface SearchMatch {
  tabId: string;
  tabName: string;
  lineNumber: number;
  lineContent: string;
  matchStart: number;
  matchEnd: number;
}

export default function SearchPanel() {
  const { state, dispatch } = useWorkspace();
  const [query, setQuery] = useState('');
  const [replaceText, setReplaceText] = useState('');
  const [showReplace, setShowReplace] = useState(false);
  const [caseSensitive, setCaseSensitive] = useState(false);
  const [useRegex, setUseRegex] = useState(false);

  const results = useMemo<SearchMatch[]>(() => {
    if (!query) return [];

    const matches: SearchMatch[] = [];

    let regex: RegExp;
    try {
      const flags = caseSensitive ? 'g' : 'gi';
      regex = useRegex ? new RegExp(query, flags) : new RegExp(query.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'), flags);
    } catch {
      return [];
    }

    for (const tab of state.openTabs) {
      const lines = tab.content.split('\n');
      for (let i = 0; i < lines.length; i++) {
        const line = lines[i];
        let match: RegExpExecArray | null;
        // Reset lastIndex for each line
        regex.lastIndex = 0;
        while ((match = regex.exec(line)) !== null) {
          matches.push({
            tabId: tab.id,
            tabName: tab.name,
            lineNumber: i + 1,
            lineContent: line,
            matchStart: match.index,
            matchEnd: match.index + match[0].length,
          });
          // Avoid infinite loop on zero-length matches
          if (match[0].length === 0) {
            regex.lastIndex++;
          }
        }
      }
    }

    return matches;
  }, [query, caseSensitive, useRegex, state.openTabs]);

  // Group results by file
  const grouped = useMemo(() => {
    const map = new Map<string, SearchMatch[]>();
    for (const m of results) {
      const existing = map.get(m.tabId);
      if (existing) {
        existing.push(m);
      } else {
        map.set(m.tabId, [m]);
      }
    }
    return map;
  }, [results]);

  const handleResultClick = useCallback((match: SearchMatch) => {
    dispatch({ type: 'SET_ACTIVE_TAB', tabId: match.tabId });
  }, [dispatch]);

  const handleReplaceInActive = useCallback(() => {
    if (!query || !state.activeTabId) return;
    const activeTab = state.openTabs.find((t) => t.id === state.activeTabId);
    if (!activeTab) return;

    let regex: RegExp;
    try {
      const flags = caseSensitive ? 'g' : 'gi';
      regex = useRegex ? new RegExp(query, flags) : new RegExp(query.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'), flags);
    } catch {
      return;
    }

    const newContent = activeTab.content.replace(regex, replaceText);
    if (newContent !== activeTab.content) {
      dispatch({ type: 'UPDATE_TAB_CONTENT', tabId: activeTab.id, content: newContent });
    }
  }, [query, replaceText, caseSensitive, useRegex, state.activeTabId, state.openTabs, dispatch]);

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
          <div className="ml-5 flex items-center gap-1">
            <div className="flex-1 flex items-center bg-[var(--bg-tertiary)] border border-[var(--border)] rounded">
              <Replace size={14} className="ml-2 text-[var(--text-secondary)]" />
              <input
                type="text"
                placeholder="Replace"
                value={replaceText}
                onChange={(e) => setReplaceText(e.target.value)}
                className="flex-1 bg-transparent px-2 py-1 text-sm text-[var(--text-primary)] outline-none"
              />
            </div>
            <button
              onClick={handleReplaceInActive}
              className="text-xs text-[var(--text-secondary)] hover:text-[var(--text-primary)] px-1.5 py-1 border border-[var(--border)] rounded"
              title="Replace all in active file"
            >
              Replace
            </button>
          </div>
        )}
        {query && (
          <div className="text-xs text-[var(--text-secondary)] px-1">
            {results.length} result{results.length !== 1 ? 's' : ''} in {grouped.size} file{grouped.size !== 1 ? 's' : ''}
          </div>
        )}
      </div>
      {query && results.length > 0 && (
        <div className="flex-1 overflow-y-auto mt-2">
          {Array.from(grouped.entries()).map(([tabId, matches]) => (
            <div key={tabId}>
              <div className="flex items-center gap-1.5 px-3 py-1 text-xs font-semibold text-[var(--text-primary)] bg-[var(--bg-sidebar)]">
                <FileText size={14} className="shrink-0" />
                <span className="truncate">{matches[0].tabName}</span>
                <span className="text-[var(--text-secondary)] ml-auto">{matches.length}</span>
              </div>
              {matches.slice(0, 50).map((match, i) => (
                <button
                  key={`${match.tabId}-${match.lineNumber}-${match.matchStart}-${i}`}
                  onClick={() => handleResultClick(match)}
                  className="w-full text-left px-5 py-0.5 text-xs hover:bg-[var(--bg-hover)] flex items-baseline gap-2"
                >
                  <span className="text-[var(--text-secondary)] shrink-0 w-6 text-right">{match.lineNumber}</span>
                  <span className="truncate text-[var(--text-primary)]">
                    {match.lineContent.slice(0, match.matchStart)}
                    <span className="bg-[var(--warning)] text-black rounded-sm px-px">
                      {match.lineContent.slice(match.matchStart, match.matchEnd)}
                    </span>
                    {match.lineContent.slice(match.matchEnd)}
                  </span>
                </button>
              ))}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
