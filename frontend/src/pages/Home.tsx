// pickle/frontend/src/pages/Home.tsx
import React from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import './Home.css';

const Home: React.FC = () => {
  const { auth, login } = useAuth();

  return (
    <div className="home-page">
      <div className="hero">
        <div className="hero-content">
          <h1>Find and Book Padel Courts Near You</h1>
          <p>
            Discover the best padel courts in your area and book your next game
            in seconds. Join the fastest growing racket sport in the world!
          </p>
          <div className="hero-buttons">
            <Link to="/courts" className="btn btn-primary">
              Find Courts
            </Link>
            {!auth.isAuthenticated && (
              <button className="btn btn-secondary" onClick={() => login()}>
                Login with Google
              </button>
            )}
          </div>
        </div>
        <div className="hero-image">
          <img
            src="/images/padel-hero.jpg"
            alt="Padel Court"
            onError={(e) => {
              e.currentTarget.src = 'https://via.placeholder.com/600x400?text=Padel+Courts';
            }}
          />
        </div>
      </div>

      <div className="features">
        <div className="feature">
          <div className="feature-icon">🔍</div>
          <h3>Find Courts</h3>
          <p>
            Search for padel courts near your location or in a specific city.
            Filter by availability, facilities, and more.
          </p>
        </div>
        <div className="feature">
          <div className="feature-icon">📅</div>
          <h3>Book Online</h3>
          <p>
            Book your court in seconds, view your upcoming games, and manage your
            reservations with ease.
          </p>
        </div>
        <div className="feature">
          <div className="feature-icon">👥</div>
          <h3>Invite Friends</h3>
          <p>
            Invite your friends to join you for a game and keep track of who's coming.
          </p>
        </div>
      </div>

      <div className="cta-section">
        <h2>Ready to Play?</h2>
        <p>
          Join thousands of padel enthusiasts who are already using our platform
          to find and book courts.
        </p>
        <Link to="/courts" className="btn btn-primary">
          Find a Court
        </Link>
      </div>
    </div>
  );
};

export default Home;