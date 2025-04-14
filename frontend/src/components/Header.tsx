// pickle/frontend/src/components/Header.tsx
import React from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import './Header.css';

const Header: React.FC = () => {
  const { auth, login, logout } = useAuth();

  return (
    <header className="header">
      <div className="container">
        <div className="logo">
          <Link to="/">PadelPro</Link>
        </div>
        <nav className="nav">
          <ul>
            <li>
              <Link to="/courts">Find Courts</Link>
            </li>
            {auth.isAuthenticated && (
              <li>
                <Link to="/bookings">My Bookings</Link>
              </li>
            )}
          </ul>
        </nav>
        <div className="auth">
          {auth.loading ? (
            <span>Loading...</span>
          ) : auth.isAuthenticated ? (
            <div className="user-menu">
              <div className="user-info">
                {auth.user?.picture && (
                  <img 
                    src={auth.user.picture} 
                    alt={auth.user.name} 
                    className="user-avatar" 
                  />
                )}
                <span className="user-name">{auth.user?.name}</span>
              </div>
              <button className="btn btn-logout" onClick={() => logout()}>
                Logout
              </button>
            </div>
          ) : (
            <button className="btn btn-login" onClick={() => login()}>
              Login with Google
            </button>
          )}
        </div>
      </div>
    </header>
  );
};

export default Header;