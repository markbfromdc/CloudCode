import { WorkspaceProvider } from './context/WorkspaceContext';
import IDELayout from './components/layout/IDELayout';

export default function App() {
  return (
    <WorkspaceProvider>
      <IDELayout />
    </WorkspaceProvider>
  );
}
