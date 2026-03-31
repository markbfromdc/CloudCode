import { useState, useCallback, useEffect } from 'react';
import TitleBar from './TitleBar';
import ActivityBar from './ActivityBar';
import Sidebar from '../sidebar/Sidebar';
import EditorTabs from '../editor/EditorTabs';
import CodeEditor from '../editor/CodeEditor';
import BottomPanel from './BottomPanel';
import StatusBar from './StatusBar';
import CommandPalette from '../common/CommandPalette';
import { useKeyboardShortcuts } from '../../hooks/useKeyboardShortcuts';
import { useWorkspace } from '../../context/WorkspaceContext';
import type { FileNode } from '../../types';

/** Demo file tree showing realistic workspace content. */
const DEMO_FILES: FileNode[] = [
  {
    name: 'src',
    path: '/workspace/src',
    type: 'directory',
    isExpanded: true,
    children: [
      {
        name: 'components',
        path: '/workspace/src/components',
        type: 'directory',
        isExpanded: false,
        children: [
          { name: 'App.tsx', path: '/workspace/src/components/App.tsx', type: 'file' },
          { name: 'Header.tsx', path: '/workspace/src/components/Header.tsx', type: 'file' },
          { name: 'Sidebar.tsx', path: '/workspace/src/components/Sidebar.tsx', type: 'file' },
          { name: 'Dashboard.tsx', path: '/workspace/src/components/Dashboard.tsx', type: 'file' },
        ],
      },
      {
        name: 'hooks',
        path: '/workspace/src/hooks',
        type: 'directory',
        isExpanded: false,
        children: [
          { name: 'useAuth.ts', path: '/workspace/src/hooks/useAuth.ts', type: 'file' },
          { name: 'useApi.ts', path: '/workspace/src/hooks/useApi.ts', type: 'file' },
          { name: 'useWebSocket.ts', path: '/workspace/src/hooks/useWebSocket.ts', type: 'file' },
        ],
      },
      {
        name: 'services',
        path: '/workspace/src/services',
        type: 'directory',
        isExpanded: false,
        children: [
          { name: 'api.ts', path: '/workspace/src/services/api.ts', type: 'file' },
          { name: 'auth.ts', path: '/workspace/src/services/auth.ts', type: 'file' },
        ],
      },
      { name: 'main.tsx', path: '/workspace/src/main.tsx', type: 'file' },
      { name: 'index.css', path: '/workspace/src/index.css', type: 'file' },
      { name: 'types.ts', path: '/workspace/src/types.ts', type: 'file' },
    ],
  },
  {
    name: 'server',
    path: '/workspace/server',
    type: 'directory',
    isExpanded: false,
    children: [
      { name: 'main.go', path: '/workspace/server/main.go', type: 'file' },
      { name: 'handler.go', path: '/workspace/server/handler.go', type: 'file' },
      { name: 'middleware.go', path: '/workspace/server/middleware.go', type: 'file' },
      { name: 'config.go', path: '/workspace/server/config.go', type: 'file' },
      { name: 'go.mod', path: '/workspace/server/go.mod', type: 'file' },
    ],
  },
  {
    name: 'tests',
    path: '/workspace/tests',
    type: 'directory',
    isExpanded: false,
    children: [
      { name: 'api.test.ts', path: '/workspace/tests/api.test.ts', type: 'file' },
      { name: 'auth.test.ts', path: '/workspace/tests/auth.test.ts', type: 'file' },
    ],
  },
  { name: 'package.json', path: '/workspace/package.json', type: 'file' },
  { name: 'tsconfig.json', path: '/workspace/tsconfig.json', type: 'file' },
  { name: 'Dockerfile', path: '/workspace/Dockerfile', type: 'file' },
  { name: 'docker-compose.yml', path: '/workspace/docker-compose.yml', type: 'file' },
  { name: 'README.md', path: '/workspace/README.md', type: 'file' },
  { name: '.gitignore', path: '/workspace/.gitignore', type: 'file' },
  { name: '.env.example', path: '/workspace/.env.example', type: 'file' },
];

export default function IDELayout() {
  const { dispatch } = useWorkspace();
  const [paletteOpen, setPaletteOpen] = useState(false);

  const togglePalette = useCallback(() => {
    setPaletteOpen((prev) => !prev);
  }, []);

  useKeyboardShortcuts(togglePalette);

  useEffect(() => {
    dispatch({ type: 'SET_FILES', files: DEMO_FILES });
  }, [dispatch]);

  return (
    <div className="flex flex-col h-screen overflow-hidden">
      <TitleBar />
      <div className="flex flex-1 overflow-hidden">
        <ActivityBar />
        <Sidebar />
        <div className="flex flex-col flex-1 overflow-hidden">
          <EditorTabs />
          <CodeEditor />
          <BottomPanel />
        </div>
      </div>
      <StatusBar />
      <CommandPalette isOpen={paletteOpen} onClose={() => setPaletteOpen(false)} />
    </div>
  );
}
