import { useState } from 'react';
import { Search } from 'lucide-react';

interface Setting {
  id: string;
  label: string;
  category: string;
  type: 'toggle' | 'select' | 'number';
  value: string | boolean | number;
  options?: string[];
  description: string;
}

const DEFAULT_SETTINGS: Setting[] = [
  {
    id: 'editor.fontSize',
    label: 'Font Size',
    category: 'Editor',
    type: 'number',
    value: 14,
    description: 'Controls the font size in pixels for the editor.',
  },
  {
    id: 'editor.tabSize',
    label: 'Tab Size',
    category: 'Editor',
    type: 'number',
    value: 2,
    description: 'The number of spaces a tab is equal to.',
  },
  {
    id: 'editor.wordWrap',
    label: 'Word Wrap',
    category: 'Editor',
    type: 'select',
    value: 'off',
    options: ['off', 'on', 'wordWrapColumn', 'bounded'],
    description: 'Controls how lines should wrap.',
  },
  {
    id: 'editor.minimap',
    label: 'Minimap Enabled',
    category: 'Editor',
    type: 'toggle',
    value: true,
    description: 'Controls whether the minimap is shown.',
  },
  {
    id: 'editor.bracketPairColorization',
    label: 'Bracket Pair Colorization',
    category: 'Editor',
    type: 'toggle',
    value: true,
    description: 'Controls whether bracket pair colorization is enabled.',
  },
  {
    id: 'editor.formatOnSave',
    label: 'Format on Save',
    category: 'Editor',
    type: 'toggle',
    value: true,
    description: 'Format the file on save.',
  },
  {
    id: 'editor.autoSave',
    label: 'Auto Save',
    category: 'Files',
    type: 'select',
    value: 'afterDelay',
    options: ['off', 'afterDelay', 'onFocusChange', 'onWindowChange'],
    description: 'Controls auto save of editors.',
  },
  {
    id: 'terminal.fontSize',
    label: 'Terminal Font Size',
    category: 'Terminal',
    type: 'number',
    value: 13,
    description: 'Controls the font size for the integrated terminal.',
  },
  {
    id: 'terminal.cursorStyle',
    label: 'Terminal Cursor Style',
    category: 'Terminal',
    type: 'select',
    value: 'bar',
    options: ['bar', 'block', 'underline'],
    description: 'Controls the cursor style for the terminal.',
  },
  {
    id: 'workbench.colorTheme',
    label: 'Color Theme',
    category: 'Workbench',
    type: 'select',
    value: 'Dark (CloudCode)',
    options: ['Dark (CloudCode)', 'Light', 'Monokai', 'Solarized Dark', 'One Dark Pro'],
    description: 'Specifies the color theme used in the workbench.',
  },
];

export default function SettingsPanel() {
  const [query, setQuery] = useState('');
  const [settings, setSettings] = useState(DEFAULT_SETTINGS);

  const filtered = settings.filter(
    (s) =>
      s.label.toLowerCase().includes(query.toLowerCase()) ||
      s.id.toLowerCase().includes(query.toLowerCase()) ||
      s.category.toLowerCase().includes(query.toLowerCase())
  );

  const categories = [...new Set(filtered.map((s) => s.category))];

  const updateSetting = (id: string, newValue: string | boolean | number) => {
    setSettings((prev) => prev.map((s) => (s.id === id ? { ...s, value: newValue } : s)));
  };

  return (
    <div className="flex flex-col h-full">
      <div className="px-4 py-2 text-xs uppercase tracking-wider text-[var(--text-secondary)] font-semibold">
        Settings
      </div>
      <div className="px-3 pb-2">
        <div className="flex items-center bg-[var(--bg-tertiary)] border border-[var(--border)] rounded">
          <Search size={14} className="ml-2 text-[var(--text-secondary)]" />
          <input
            type="text"
            placeholder="Search settings..."
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            className="flex-1 bg-transparent px-2 py-1 text-sm text-[var(--text-primary)] outline-none"
          />
        </div>
      </div>
      <div className="flex-1 overflow-y-auto">
        {categories.map((cat) => (
          <div key={cat}>
            <div className="px-4 py-1 text-xs uppercase tracking-wider text-[var(--text-secondary)] font-semibold sticky top-0 bg-[var(--bg-sidebar)]">
              {cat}
            </div>
            {filtered
              .filter((s) => s.category === cat)
              .map((setting) => (
                <div key={setting.id} className="px-4 py-2 hover:bg-[var(--bg-hover)]">
                  <div className="text-sm text-[var(--text-primary)]">{setting.label}</div>
                  <div className="text-xs text-[var(--text-secondary)] mb-1.5">{setting.description}</div>
                  {setting.type === 'toggle' && (
                    <button
                      onClick={() => updateSetting(setting.id, !setting.value)}
                      className={`w-9 h-5 rounded-full relative transition-colors ${
                        setting.value ? 'bg-[var(--accent)]' : 'bg-[var(--bg-tertiary)]'
                      }`}
                    >
                      <span
                        className={`absolute top-0.5 w-4 h-4 rounded-full bg-white transition-transform ${
                          setting.value ? 'left-[18px]' : 'left-0.5'
                        }`}
                      />
                    </button>
                  )}
                  {setting.type === 'select' && (
                    <select
                      value={String(setting.value)}
                      onChange={(e) => updateSetting(setting.id, e.target.value)}
                      className="bg-[var(--bg-tertiary)] border border-[var(--border)] text-[var(--text-primary)] text-xs px-2 py-1 rounded outline-none"
                    >
                      {setting.options?.map((opt) => (
                        <option key={opt} value={opt}>{opt}</option>
                      ))}
                    </select>
                  )}
                  {setting.type === 'number' && (
                    <input
                      type="number"
                      value={Number(setting.value)}
                      onChange={(e) => updateSetting(setting.id, parseInt(e.target.value) || 0)}
                      className="w-20 bg-[var(--bg-tertiary)] border border-[var(--border)] text-[var(--text-primary)] text-xs px-2 py-1 rounded outline-none"
                    />
                  )}
                </div>
              ))}
          </div>
        ))}
      </div>
    </div>
  );
}
