// src/pages/AuthCallback.tsx
import React, { useEffect } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';

const AuthCallback: React.FC = () => {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const { loginWithToken } = useAuth();
  
  useEffect(() => {
    const token = searchParams.get('token');
    if (token) {
      // Save token to localStorage
      localStorage.setItem('token', token);
      
      // Update auth context
      loginWithToken(token);
      
      // Redirect to home page
      navigate('/', { replace: true });
    } else {
      navigate('/');
    }
  }, [searchParams, navigate, loginWithToken]);
  
  return (
    <div className="auth-callback">
      <p>Logging you in...</p>
    </div>
  );
};

export default AuthCallback;