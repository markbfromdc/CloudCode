import { useState } from 'react';
import { Search, Download, Check, Star } from 'lucide-react';

interface Extension {
  id: string;
  name: string;
  publisher: string;
  description: string;
  installed: boolean;
  rating: number;
  downloads: string;
}

const EXTENSIONS: Extension[] = [
  {
    id: 'prettier',
    name: 'Prettier',
    publisher: 'Prettier',
    description: 'Code formatter using prettier',
    installed: true,
    rating: 4.8,
    downloads: '32M',
  },
  {
    id: 'eslint',
    name: 'ESLint',
    publisher: 'Microsoft',
    description: 'Integrates ESLint into the editor',
    installed: true,
    rating: 4.7,
    downloads: '28M',
  },
  {
    id: 'go',
    name: 'Go',
    publisher: 'Go Team at Google',
    description: 'Rich Go language support with IntelliSense',
    installed: true,
    rating: 4.9,
    downloads: '12M',
  },
  {
    id: 'python',
    name: 'Python',
    publisher: 'Microsoft',
    description: 'Linting, debugging, IntelliSense for Python',
    installed: false,
    rating: 4.8,
    downloads: '45M',
  },
  {
    id: 'docker',
    name: 'Docker',
    publisher: 'Microsoft',
    description: 'Docker file support and container management',
    installed: false,
    rating: 4.6,
    downloads: '18M',
  },
  {
    id: 'gitlens',
    name: 'GitLens',
    publisher: 'GitKraken',
    description: 'Git supercharged — visualize code authorship',
    installed: false,
    rating: 4.7,
    downloads: '22M',
  },
  {
    id: 'tailwind',
    name: 'Tailwind CSS IntelliSense',
    publisher: 'Tailwind Labs',
    description: 'Intelligent Tailwind CSS tooling',
    installed: false,
    rating: 4.9,
    downloads: '10M',
  },
];

export default function ExtensionsPanel() {
  const [query, setQuery] = useState('');
  const [extensions, setExtensions] = useState(EXTENSIONS);

  const filtered = extensions.filter(
    (ext) =>
      ext.name.toLowerCase().includes(query.toLowerCase()) ||
      ext.description.toLowerCase().includes(query.toLowerCase())
  );

  const installed = filtered.filter((e) => e.installed);
  const available = filtered.filter((e) => !e.installed);

  const toggleInstall = (id: string) => {
    setExtensions((exts) =>
      exts.map((e) => (e.id === id ? { ...e, installed: !e.installed } : e))
    );
  };

  return (
    <div className="flex flex-col h-full">
      <div className="px-4 py-2 text-xs uppercase tracking-wider text-[var(--text-secondary)] font-semibold">
        Extensions
      </div>
      <div className="px-3 pb-2">
        <div className="flex items-center bg-[var(--bg-tertiary)] border border-[var(--border)] rounded">
          <Search size={14} className="ml-2 text-[var(--text-secondary)]" />
          <input
            type="text"
            placeholder="Search extensions..."
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            className="flex-1 bg-transparent px-2 py-1 text-sm text-[var(--text-primary)] outline-none"
          />
        </div>
      </div>

      <div className="flex-1 overflow-y-auto">
        {installed.length > 0 && (
          <>
            <div className="px-4 py-1 text-xs uppercase tracking-wider text-[var(--text-secondary)] font-semibold">
              Installed ({installed.length})
            </div>
            {installed.map((ext) => (
              <ExtensionItem key={ext.id} ext={ext} onToggle={toggleInstall} />
            ))}
          </>
        )}
        {available.length > 0 && (
          <>
            <div className="px-4 py-1 mt-2 text-xs uppercase tracking-wider text-[var(--text-secondary)] font-semibold">
              Recommended ({available.length})
            </div>
            {available.map((ext) => (
              <ExtensionItem key={ext.id} ext={ext} onToggle={toggleInstall} />
            ))}
          </>
        )}
      </div>
    </div>
  );
}

function ExtensionItem({ ext, onToggle }: { ext: Extension; onToggle: (id: string) => void }) {
  return (
    <div className="flex items-start gap-2 px-3 py-2 hover:bg-[var(--bg-hover)] cursor-pointer">
      <div className="w-8 h-8 rounded bg-[var(--bg-tertiary)] flex items-center justify-center text-xs font-bold text-[var(--accent)] shrink-0 mt-0.5">
        {ext.name[0]}
      </div>
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-1">
          <span className="text-sm text-[var(--text-primary)] font-medium truncate">{ext.name}</span>
        </div>
        <div className="text-xs text-[var(--text-secondary)] truncate">{ext.description}</div>
        <div className="flex items-center gap-2 mt-0.5 text-xs text-[var(--text-secondary)]">
          <span>{ext.publisher}</span>
          <span className="flex items-center gap-0.5">
            <Star size={10} className="fill-yellow-400 text-yellow-400" />
            {ext.rating}
          </span>
          <span>{ext.downloads}</span>
        </div>
      </div>
      <button
        onClick={() => onToggle(ext.id)}
        className={`px-2 py-0.5 text-xs rounded shrink-0 mt-1 ${
          ext.installed
            ? 'bg-[var(--bg-active)] text-[var(--success)]'
            : 'bg-[var(--accent)] text-white hover:bg-[var(--accent-hover)]'
        }`}
      >
        {ext.installed ? (
          <span className="flex items-center gap-1"><Check size={12} /> Installed</span>
        ) : (
          <span className="flex items-center gap-1"><Download size={12} /> Install</span>
        )}
      </button>
    </div>
  );
}
