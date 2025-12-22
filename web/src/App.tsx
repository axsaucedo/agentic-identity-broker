/**
 * Main App component with React Router setup.
 *
 * Configures routes for the consent management application:
 * - / -> Consent overview (list of agent delegations)
 * - /agent/:agentId -> Agent grant detail page
 * - * -> 404 error page
 *
 * Performance Optimizations:
 * - Code splitting with React.lazy() for routes
 * - Suspense with loading fallback
 * - Global error boundary
 * - Toast provider for notifications
 */

import React, { Suspense, lazy } from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { GlobalErrorBoundary } from '@components/ui/GlobalErrorBoundary';
import { ToastProvider } from '@components/ui/Toast';

// Code-split route components using React.lazy()
// This creates separate bundles for each route, reducing initial load time
const ConsentOverviewPage = lazy(() => import('./pages/ConsentOverviewPage'));
const AgentGrantDetailPage = lazy(() => import('./pages/AgentGrantDetailPage'));
const ErrorPage = lazy(() => import('./pages/ErrorPage'));

/**
 * Loading fallback component shown during code splitting lazy load.
 */
function LoadingFallback() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-neutral-50">
      <div className="text-center">
        <div className="inline-block animate-spin rounded-full h-12 w-12 border-4 border-blue-500 border-t-transparent"></div>
        <p className="mt-4 text-neutral-600">Loading...</p>
      </div>
    </div>
  );
}

function App() {
  return (
    <GlobalErrorBoundary>
      <ToastProvider>
        <Router basename="/consent">
          <Suspense fallback={<LoadingFallback />}>
            <Routes>
              <Route path="/" element={<ConsentOverviewPage />} />
              <Route path="/agent/:agentId" element={<AgentGrantDetailPage />} />
              <Route path="*" element={<ErrorPage />} />
            </Routes>
          </Suspense>
        </Router>
      </ToastProvider>
    </GlobalErrorBoundary>
  );
}

export default App;
