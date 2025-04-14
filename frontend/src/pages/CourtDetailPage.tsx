// pickle/frontend/src/pages/CourtDetailPage.tsx
import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { format, addDays, parseISO } from 'date-fns';
import apiService from '../services/api';
import { useAuth } from '../hooks/useAuth';
import { Court, Booking, BookingStatus } from '../types';
import './CourtDetailPage.css';

interface BookingFormData {
  date: string;
  startTime: string;
  endTime: string;
  numberOfPlayers: number;
  playerEmails: string[];
}

const CourtDetailPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { auth } = useAuth();
  
  const [court, setCourt] = useState<Court | null>(null);
  const [bookings, setBookings] = useState<Booking[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showBookingForm, setShowBookingForm] = useState(false);
  const [bookingFormData, setBookingFormData] = useState<BookingFormData>({
    date: format(new Date(), 'yyyy-MM-dd'),
    startTime: '10:00',
    endTime: '11:00',
    numberOfPlayers: 2,
    playerEmails: [''],
  });
  const [bookingSuccess, setBookingSuccess] = useState(false);
  const [bookingError, setBookingError] = useState<string | null>(null);
  const [selectedDate, setSelectedDate] = useState<string>(format(new Date(), 'yyyy-MM-dd'));

  useEffect(() => {
    if (!id) return;
    
    const fetchCourtAndBookings = async () => {
      setLoading(true);
      setError(null);
      
      try {
        // Fetch court details
        const courtData = await apiService.courts.getCourt({ courtId: id });
        setCourt(courtData);
        
        // Fetch bookings for the court on the selected date
        await fetchBookingsForDate(selectedDate);
      } catch (err) {
        setError('Failed to load court details. Please try again later.');
        console.error(err);
      } finally {
        setLoading(false);
      }
    };
    
    fetchCourtAndBookings();
  }, [id]);
  
  const fetchBookingsForDate = async (date: string) => {
    if (!id) return;
    
    try {
      const response = await apiService.bookings.getBookings({
        courtId: id,
        date,
      });
      setBookings(response.bookings);
    } catch (err) {
      console.error('Error fetching bookings:', err);
    }
  };
  
  const handleDateChange = async (date: string) => {
    setSelectedDate(date);
    await fetchBookingsForDate(date);
  };
  
  const handleBookingFormChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    const { name, value } = e.target;
    
    if (name.startsWith('playerEmail')) {
      const index = parseInt(name.replace('playerEmail', ''));
      const updatedEmails = [...bookingFormData.playerEmails];
      
      if (index >= updatedEmails.length) {
        updatedEmails.push(value);
      } else {
        updatedEmails[index] = value;
      }
      
      setBookingFormData({
        ...bookingFormData,
        playerEmails: updatedEmails,
      });
    } else if (name === 'numberOfPlayers') {
      const players = parseInt(value);
      let emails = [...bookingFormData.playerEmails];
      
      // Adjust player emails array based on number of players
      if (players > emails.length) {
        // Add empty emails if needed
        while (emails.length < players - 1) {
          emails.push('');
        }
      } else if (players < emails.length + 1) {
        // Remove extra emails if needed
        emails = emails.slice(0, players - 1);
      }
      
      setBookingFormData({
        ...bookingFormData,
        numberOfPlayers: players,
        playerEmails: emails,
      });
    } else {
      setBookingFormData({
        ...bookingFormData,
        [name]: value,
      });
    }
  };
  
  const handleSubmitBooking = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!auth.isAuthenticated) {
      setBookingError('You must be logged in to book a court.');
      return;
    }
    
    if (!id) return;
    
    setBookingError(null);
    
    try {
      await apiService.bookings.createBooking({
        courtId: id,
        date: bookingFormData.date,
        startTime: bookingFormData.startTime,
        endTime: bookingFormData.endTime,
        numberOfPlayers: bookingFormData.numberOfPlayers,
        playerEmails: bookingFormData.playerEmails.filter(email => email.trim() !== ''),
      });
      
      setBookingSuccess(true);
      setShowBookingForm(false);
      
      // Refresh bookings
      await fetchBookingsForDate(selectedDate);
      
      // Reset form
      setBookingFormData({
        date: format(new Date(), 'yyyy-MM-dd'),
        startTime: '10:00',
        endTime: '11:00',
        numberOfPlayers: 2,
        playerEmails: [''],
      });
      
      // Hide success message after 5 seconds
      setTimeout(() => {
        setBookingSuccess(false);
      }, 5000);
    } catch (err) {
      setBookingError('Failed to create booking. The time slot might already be booked.');
      console.error(err);
    }
  };
  
  // Generate an array of the next 7 days
  const nextSevenDays = Array.from({ length: 7 }, (_, i) => {
    const date = addDays(new Date(), i);
    return {
      date: format(date, 'yyyy-MM-dd'),
      display: format(date, 'E, MMM d'), // e.g., "Mon, Jan 1"
    };
  });
  
  if (loading) {
    return <div className="loading">Loading court details...</div>;
  }
  
  if (error || !court) {
    return <div className="error">{error || 'Court not found.'}</div>;
  }
  
  return (
    <div className="court-detail-page">
      <div className="court-header">
        <div className="court-image">
          <img
            src={court.imageUrl || 'https://via.placeholder.com/800x400?text=Padel+Court'}
            alt={court.name}
            onError={(e) => {
              e.currentTarget.src = 'https://via.placeholder.com/800x400?text=Padel+Court';
            }}
          />
        </div>
        <div className="court-info">
          <h1>{court.name}</h1>
          <p className="court-address">{court.address}</p>
          <div className="court-details">
            <div className="detail-item">
              <span className="detail-label">Number of courts:</span>
              <span className="detail-value">{court.numberOfCourts}</span>
            </div>
            <div className="detail-item">
              <span className="detail-label">Amenities:</span>
              <div className="amenities-list">
                {court.amenities.map((amenity, index) => (
                  <span key={index} className="amenity-tag">
                    {amenity}
                  </span>
                ))}
              </div>
            </div>
          </div>
          {auth.isAuthenticated ? (
            <button 
              className="btn btn-primary" 
              onClick={() => setShowBookingForm(!showBookingForm)}
            >
              {showBookingForm ? 'Cancel' : 'Book This Court'}
            </button>
          ) : (
            <p className="login-prompt">
              Please log in to book this court.
            </p>
          )}
        </div>
      </div>
      
      {bookingSuccess && (
        <div className="booking-success">
          Booking created successfully! Check your bookings for details.
        </div>
      )}
      
      {showBookingForm && (
        <div className="booking-form-container">
          <h2>Book a Court</h2>
          <form onSubmit={handleSubmitBooking} className="booking-form">
            <div className="form-group">
              <label htmlFor="date">Date</label>
              <input
                type="date"
                id="date"
                name="date"
                value={bookingFormData.date}
                onChange={handleBookingFormChange}
                min={format(new Date(), 'yyyy-MM-dd')}
                max={format(addDays(new Date(), 30), 'yyyy-MM-dd')}
                required
              />
            </div>
            
            <div className="form-row">
              <div className="form-group">
                <label htmlFor="startTime">Start Time</label>
                <select
                  id="startTime"
                  name="startTime"
                  value={bookingFormData.startTime}
                  onChange={handleBookingFormChange}
                  required
                >
                  {Array.from({ length: 13 }, (_, i) => i + 8).map((hour) => (
                    <option key={hour} value={`${hour}:00`}>
                      {hour}:00 {hour < 12 ? 'AM' : 'PM'}
                    </option>
                  ))}
                </select>
              </div>
              
              <div className="form-group">
                <label htmlFor="endTime">End Time</label>
                <select
                  id="endTime"
                  name="endTime"
                  value={bookingFormData.endTime}
                  onChange={handleBookingFormChange}
                  required
                >
                  {Array.from({ length: 13 }, (_, i) => i + 9).map((hour) => (
                    <option key={hour} value={`${hour}:00`}>
                      {hour}:00 {hour < 12 ? 'AM' : 'PM'}
                    </option>
                  ))}
                </select>
              </div>
            </div>
            
            <div className="form-group">
              <label htmlFor="numberOfPlayers">Number of Players</label>
              <select
                id="numberOfPlayers"
                name="numberOfPlayers"
                value={bookingFormData.numberOfPlayers}
                onChange={handleBookingFormChange}
                required
              >
                <option value={2}>2 Players</option>
                <option value={4}>4 Players</option>
              </select>
            </div>
            
            {bookingFormData.numberOfPlayers > 1 && (
              <div className="form-group">
                <label>Player Emails (optional)</label>
                <div className="player-emails">
                  {Array.from(
                    { length: bookingFormData.numberOfPlayers - 1 },
                    (_, i) => (
                      <input
                        key={i}
                        type="email"
                        name={`playerEmail${i}`}
                        placeholder={`Player ${i + 2} email`}
                        value={bookingFormData.playerEmails[i] || ''}
                        onChange={handleBookingFormChange}
                      />
                    )
                  )}
                </div>
                <small className="form-hint">
                  Players will receive a notification email.
                </small>
              </div>
            )}
            
            {bookingError && <div className="form-error">{bookingError}</div>}
            
            <button type="submit" className="btn btn-primary">
              Confirm Booking
            </button>
          </form>
        </div>
      )}
      
      <div className="availability-section">
        <h2>Availability</h2>
        
        <div className="date-selector">
          {nextSevenDays.map((day) => (
            <button
              key={day.date}
              className={`date-button ${selectedDate === day.date ? 'active' : ''}`}
              onClick={() => handleDateChange(day.date)}
            >
              {day.display}
            </button>
          ))}
        </div>
        
        <div className="availability-grid">
          <div className="time-labels">
            {Array.from({ length: 13 }, (_, i) => i + 8).map((hour) => (
              <div key={hour} className="time-slot">
                {hour}:00 {hour < 12 ? 'AM' : 'PM'}
              </div>
            ))}
          </div>
          
          <div className="court-slots">
            {Array.from({ length: court.numberOfCourts }, (_, courtIndex) => (
              <div key={courtIndex} className="court-row">
                <div className="court-label">Court {courtIndex + 1}</div>
                <div className="time-slots">
                  {Array.from({ length: 13 }, (_, timeIndex) => {
                    const hour = timeIndex + 8;
                    const timeSlot = `${hour}:00`;
                    
                    // Check if this time slot is booked
                    const isBooked = bookings.some(
                      (booking) => 
                        booking.startTime <= timeSlot && 
                        booking.endTime > timeSlot && 
                        booking.status !== BookingStatus.CANCELLED
                    );
                    
                    return (
                      <div 
                        key={timeIndex} 
                        className={`time-slot ${isBooked ? 'booked' : 'available'}`}
                      >
                        {isBooked ? 'Booked' : 'Available'}
                      </div>
                    );
                  })}
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
};

export default CourtDetailPage;