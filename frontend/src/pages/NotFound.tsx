// pickle/frontend/src/pages/NotFound.tsx
import React from 'react';
import { Link } from 'react-router-dom';
import './NotFound.css';

const NotFound: React.FC = () => {
  return (
    <div className="not-found-page">
      <div className="not-found-content">
        <h1>404</h1>
        <h2>Page Not Found</h2>
        <p>The page you're looking for doesn't exist or has been moved.</p>
        <div className="not-found-actions">
          <Link to="/" className="btn btn-primary">
            Go Home
          </Link>
          <Link to="/courts" className="btn btn-secondary">
            Find Courts
          </Link>
        </div>
      </div>
    </div>
  );
};

export default NotFound;