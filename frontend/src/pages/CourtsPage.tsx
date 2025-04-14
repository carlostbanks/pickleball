// pickle/frontend/src/pages/CourtsPage.tsx
import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import apiService from '../services/api';
import { Court } from '../types';
import './CourtsPage.css';

const CourtsPage: React.FC = () => {
  const [courts, setCourts] = useState<Court[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchCity, setSearchCity] = useState('');
  const [radius, setRadius] = useState(10);
  const [useLocation, setUseLocation] = useState(false);

  useEffect(() => {
    // Load courts on initial render
    loadCourts();
  }, []);

  const loadCourts = async (city?: string, lat?: number, lng?: number, radiusKm?: number) => {
    setLoading(true);
    setError(null);

    try {
      const response = await apiService.courts.getCourts({
        city,
        latitude: lat,
        longitude: lng,
        radiusKm,
      });

      setCourts(response.courts);
    } catch (err) {
      setError('Failed to load courts. Please try again later.');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();

    if (useLocation) {
      // Use browser geolocation
      if (navigator.geolocation) {
        navigator.geolocation.getCurrentPosition(
          (position) => {
            loadCourts(
              '',
              position.coords.latitude,
              position.coords.longitude,
              radius
            );
          },
          (err) => {
            setError('Error getting your location. Please enter a city instead.');
            console.error(err);
          }
        );
      } else {
        setError('Geolocation is not supported by your browser. Please enter a city instead.');
      }
    } else {
      // Search by city
      loadCourts(searchCity);
    }
  };

  return (
    <div className="courts-page">
      <div className="search-section">
        <h1>Find Padel Courts</h1>
        <form onSubmit={handleSearch} className="search-form">
          <div className="search-options">
            <div className="search-method">
              <label>
                <input
                  type="radio"
                  checked={!useLocation}
                  onChange={() => setUseLocation(false)}
                />
                Search by city
              </label>
              <label>
                <input
                  type="radio"
                  checked={useLocation}
                  onChange={() => setUseLocation(true)}
                />
                Use my location
              </label>
            </div>

            {!useLocation ? (
              <div className="search-city">
                <input
                  type="text"
                  placeholder="Enter city name"
                  value={searchCity}
                  onChange={(e) => setSearchCity(e.target.value)}
                  required={!useLocation}
                />
              </div>
            ) : (
              <div className="search-radius">
                <label>
                  Search radius: {radius} km
                  <input
                    type="range"
                    min="1"
                    max="50"
                    value={radius}
                    onChange={(e) => setRadius(parseInt(e.target.value))}
                  />
                </label>
              </div>
            )}
          </div>

          <button type="submit" className="btn btn-primary">
            Search
          </button>
        </form>
      </div>

      <div className="courts-results">
        {loading ? (
          <div className="loading">Loading courts...</div>
        ) : error ? (
          <div className="error">{error}</div>
        ) : courts.length === 0 ? (
          <div className="no-results">
            No courts found. Try a different search or location.
          </div>
        ) : (
          <div className="court-grid">
            {courts.map((court) => (
              <Link 
                to={`/courts/${court.id}`} 
                className="court-card" 
                key={court.id}
              >
                <div className="court-image">
                  <img
                    src={court.imageUrl || 'https://via.placeholder.com/300x200?text=Padel+Court'}
                    alt={court.name}
                    onError={(e) => {
                      e.currentTarget.src = 'https://via.placeholder.com/300x200?text=Padel+Court';
                    }}
                  />
                </div>
                <div className="court-details">
                  <h3>{court.name}</h3>
                  <p className="court-address">{court.address}</p>
                  <p className="court-info">
                    {court.numberOfCourts} {court.numberOfCourts === 1 ? 'court' : 'courts'}
                  </p>
                  <div className="court-amenities">
                    {court.amenities.slice(0, 3).map((amenity, index) => (
                      <span key={index} className="amenity-tag">
                        {amenity}
                      </span>
                    ))}
                    {court.amenities.length > 3 && (
                      <span className="amenity-tag more">
                        +{court.amenities.length - 3}
                      </span>
                    )}
                  </div>
                </div>
              </Link>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

export default CourtsPage;