import { WorkspaceProvider } from './context/WorkspaceContext';
import IDELayout from './components/layout/IDELayout';
import ErrorBoundary from './components/common/ErrorBoundary';

export default function App() {
  return (
    <ErrorBoundary>
      <WorkspaceProvider>
        <IDELayout />
      </WorkspaceProvider>
    </ErrorBoundary>
  );
}
