// pickle/frontend/src/App.tsx
import React from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider, useAuth } from './hooks/useAuth';
import Header from './components/Header';
import Home from './pages/Home';
import CourtsPage from './pages/CourtsPage';
import CourtDetailPage from './pages/CourtDetailPage';
import BookingsPage from './pages/BookingsPage';
import NotFound from './pages/NotFound';
import './App.css';

// Protected route component
const ProtectedRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { auth } = useAuth();

  if (auth.loading) {
    return <div>Loading...</div>;
  }

  if (!auth.isAuthenticated) {
    return <Navigate to="/" />;
  }

  return <>{children}</>;
};

const App: React.FC = () => {
  return (
    <AuthProvider>
      <Header />
      <div className="container">
        <Routes>
          <Route path="/" element={<Home />} />
          <Route path="/courts" element={<CourtsPage />} />
          <Route path="/courts/:id" element={<CourtDetailPage />} />
          <Route
            path="/bookings"
            element={
              <ProtectedRoute>
                <BookingsPage />
              </ProtectedRoute>
            }
          />
          <Route path="*" element={<NotFound />} />
        </Routes>
      </div>
    </AuthProvider>
  );
};

export default App;