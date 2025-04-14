// src/hooks/useAuth.tsx
import React, { createContext, useContext, useEffect, useState } from 'react';
import { AuthState, User } from '../types';
import apiService from '../services/api';

// Create the context
interface AuthContextType {
  auth: AuthState;
  login: () => void;
  logout: () => Promise<void>;
  loginWithToken: (token: string) => Promise<void>;
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
      const token = localStorage.getItem('token');
      
      if (!token) {
        setAuth({
          isAuthenticated: false,
          user: null,
          loading: false,
          error: null,
        });
        return;
      }
      
      try {
        const user = await apiService.auth.getCurrentUser();
        setAuth({
          isAuthenticated: true,
          user,
          loading: false,
          error: null,
        });
      } catch (error) {
        // Token might be invalid or expired
        localStorage.removeItem('token');
        setAuth({
          isAuthenticated: false,
          user: null,
          loading: false,
          error: "Authentication failed",
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
      localStorage.removeItem('token');
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
  
  // Login with token (for auth callback)
  const loginWithToken = async (token: string) => {
    try {
      // Set the token in localStorage (should already be done in callback)
      localStorage.setItem('token', token);
      
      // Get user with the token
      const user = await apiService.auth.getCurrentUser();
      
      setAuth({
        isAuthenticated: true,
        user,
        loading: false,
        error: null,
      });
    } catch (error) {
      setAuth({
        isAuthenticated: false,
        user: null,
        loading: false,
        error: "Failed to authenticate with token",
      });
    }
  };

  return (
    <AuthContext.Provider value={{ auth, login, logout, loginWithToken }}>
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