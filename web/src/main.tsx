/**
 * Application entry point.
 *
 * Mounts the React application to the DOM with:
 * - React.StrictMode for development warnings
 * - Global styles imported
 */

import React from 'react';
import ReactDOM from 'react-dom/client';
import App from './App.tsx';
import './styles/index.css';

const rootElement = document.getElementById('root');

if (!rootElement) {
  throw new Error('Failed to find the root element');
}

ReactDOM.createRoot(rootElement).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
);
