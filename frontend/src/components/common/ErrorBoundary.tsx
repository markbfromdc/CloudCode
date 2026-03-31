import { Component, type ReactNode } from 'react';

interface ErrorBoundaryProps {
  children: ReactNode;
}

interface ErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
}

export default class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, error };
  }

  handleReload = () => {
    window.location.reload();
  };

  render() {
    if (this.state.hasError) {
      return (
        <div className="flex items-center justify-center h-screen bg-[var(--bg-primary)]">
          <div className="text-center p-8 max-w-md">
            <div className="text-4xl mb-4 text-[var(--error)]">Something went wrong</div>
            <p className="text-sm text-[var(--text-secondary)] mb-2">
              An unexpected error occurred in the application.
            </p>
            {this.state.error && (
              <pre className="text-xs text-[var(--text-secondary)] bg-[var(--bg-tertiary)] p-3 rounded mb-4 overflow-auto max-h-40 text-left">
                {this.state.error.message}
              </pre>
            )}
            <button
              onClick={this.handleReload}
              className="px-4 py-2 bg-[var(--accent)] hover:bg-[var(--accent-hover)] text-white rounded text-sm transition-colors"
            >
              Reload
            </button>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}
