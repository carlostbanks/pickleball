// pickle/frontend/src/hooks/useAuth.tsx
import React, { createContext, useContext, useEffect, useState } from 'react';
import { AuthState, User } from '../types';
import apiService from '../services/api';

// Create the context
interface AuthContextType {
  auth: AuthState;
  login: () => void;
  logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

// Auth provider component
export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [auth, setAuth] = useState<AuthState>({
    isAuthenticated: false,
    user: null,
    loading: true,
    error: null,
  });

  // Check if user is already authenticated on mount
  useEffect(() => {
    const checkAuth = async () => {
      try {
        // Temporarily comment out API call during development
        // const user = await apiService.auth.getCurrentUser();
        setAuth({
          isAuthenticated: false,
          user: null,
          loading: false,
          error: null,
        });
      } catch (error) {
        setAuth({
          isAuthenticated: false,
          user: null,
          loading: false,
          error: null,
        });
      }
    };
  
    checkAuth();
  }, []);

  // Redirect to Google OAuth login
  const login = () => {
    window.location.href = `${import.meta.env.VITE_API_URL || 'http://localhost:8080'}/auth/google/login`;
  };

  // Log out
  const logout = async () => {
    try {
      await apiService.auth.logout();
      setAuth({
        isAuthenticated: false,
        user: null,
        loading: false,
        error: null,
      });
    } catch (error) {
      setAuth({
        ...auth,
        error: 'Failed to log out',
      });
    }
  };

  return (
    <AuthContext.Provider value={{ auth, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
};

// Custom hook to use the auth context
export const useAuth = (): AuthContextType => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};

export default useAuth;