// pickle/frontend/src/pages/BookingsPage.tsx
import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { format, parseISO, isBefore } from 'date-fns';
import apiService from '../services/api';
import { useAuth } from '../hooks/useAuth';
import { Booking, BookingStatus, Court } from '../types';
import './BookingsPage.css';

const BookingsPage: React.FC = () => {
  const { auth } = useAuth();
  const [bookings, setBookings] = useState<Booking[]>([]);
  const [courts, setCourts] = useState<Record<string, Court>>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<'upcoming' | 'past'>('upcoming');
  const [cancellingBookingId, setCancellingBookingId] = useState<string | null>(null);
  const [cancelSuccess, setCancelSuccess] = useState(false);

  useEffect(() => {
    if (!auth.isAuthenticated || !auth.user) return;

    const fetchBookings = async () => {
      setLoading(true);
      setError(null);

      try {
        const response = await apiService.bookings.getBookings({
          userId: auth.user?.id,
        });

        setBookings(response.bookings);

        // Fetch court details for each booking
        const courtIds = [...new Set(response.bookings.map((booking) => booking.courtId))];
        const courtsData: Record<string, Court> = {};

        for (const courtId of courtIds) {
          try {
            const court = await apiService.courts.getCourt({ courtId });
            courtsData[courtId] = court;
          } catch (err) {
            console.error(`Failed to fetch court ${courtId}:`, err);
          }
        }

        setCourts(courtsData);
      } catch (err) {
        setError('Failed to load bookings. Please try again later.');
        console.error(err);
      } finally {
        setLoading(false);
      }
    };

    fetchBookings();
  }, [auth.isAuthenticated, auth.user]);

  const handleCancelBooking = async (bookingId: string) => {
    setCancellingBookingId(bookingId);
    
    try {
      await apiService.bookings.cancelBooking({ bookingId });
      
      // Update the booking status in the local state
      setBookings((prevBookings) =>
        prevBookings.map((booking) =>
          booking.id === bookingId
            ? { ...booking, status: BookingStatus.CANCELLED }
            : booking
        )
      );
      
      setCancelSuccess(true);
      
      // Hide success message after 3 seconds
      setTimeout(() => {
        setCancelSuccess(false);
      }, 3000);
    } catch (err) {
      setError('Failed to cancel booking. Please try again.');
      console.error(err);
    } finally {
      setCancellingBookingId(null);
    }
  };

  const filteredBookings = bookings.filter((booking) => {
    const bookingDate = parseISO(`${booking.date}T${booking.endTime}`);
    const isPast = isBefore(bookingDate, new Date());
    
    return activeTab === 'upcoming' 
      ? !isPast && booking.status !== BookingStatus.CANCELLED
      : isPast || booking.status === BookingStatus.CANCELLED;
  });

  // Sort bookings by date and time
  const sortedBookings = [...filteredBookings].sort((a, b) => {
    const dateA = new Date(`${a.date}T${a.startTime}`);
    const dateB = new Date(`${b.date}T${b.startTime}`);
    return dateA.getTime() - dateB.getTime();
  });

  if (loading) {
    return <div className="loading">Loading your bookings...</div>;
  }

  return (
    <div className="bookings-page">
      <h1>My Bookings</h1>
      
      {cancelSuccess && (
        <div className="success-message">
          Booking cancelled successfully.
        </div>
      )}
      
      {error && <div className="error-message">{error}</div>}
      
      <div className="tabs">
        <button
          className={`tab ${activeTab === 'upcoming' ? 'active' : ''}`}
          onClick={() => setActiveTab('upcoming')}
        >
          Upcoming
        </button>
        <button
          className={`tab ${activeTab === 'past' ? 'active' : ''}`}
          onClick={() => setActiveTab('past')}
        >
          Past & Cancelled
        </button>
      </div>
      
      {sortedBookings.length === 0 ? (
        <div className="no-bookings">
          {activeTab === 'upcoming' 
            ? "You don't have any upcoming bookings."
            : "You don't have any past or cancelled bookings."}
          <br />
          <Link to="/courts" className="btn btn-primary">
            Find Courts to Book
          </Link>
        </div>
      ) : (
        <div className="bookings-list">
          {sortedBookings.map((booking) => {
            const court = courts[booking.courtId];
            
            return (
              <div key={booking.id} className={`booking-card status-${booking.status.toLowerCase()}`}>
                <div className="booking-left">
                  <div className="booking-date">
                    {format(parseISO(booking.date), 'EEEE, MMMM d, yyyy')}
                  </div>
                  <div className="booking-time">
                    {booking.startTime} - {booking.endTime}
                  </div>
                  <div className="booking-status">
                    {booking.status === BookingStatus.CONFIRMED && (
                      <span className="status confirmed">Confirmed</span>
                    )}
                    {booking.status === BookingStatus.PENDING && (
                      <span className="status pending">Pending</span>
                    )}
                    {booking.status === BookingStatus.CANCELLED && (
                      <span className="status cancelled">Cancelled</span>
                    )}
                  </div>
                </div>
                
                <div className="booking-center">
                  <h3 className="court-name">
                    {court ? (
                      <Link to={`/courts/${court.id}`}>{court.name}</Link>
                    ) : (
                      'Loading court details...'
                    )}
                  </h3>
                  {court && <div className="court-address">{court.address}</div>}
                  <div className="booking-details">
                    <span>{booking.numberOfPlayers} Players</span>
                    {booking.playerEmails.length > 0 && (
                      <div className="player-emails">
                        Invited: {booking.playerEmails.join(', ')}
                      </div>
                    )}
                  </div>
                </div>
                
                <div className="booking-right">
                  {activeTab === 'upcoming' && booking.status !== BookingStatus.CANCELLED && (
                    <>
                      <button
                        className="btn btn-cancel"
                        onClick={() => handleCancelBooking(booking.id)}
                        disabled={cancellingBookingId === booking.id}
                      >
                        {cancellingBookingId === booking.id ? 'Cancelling...' : 'Cancel Booking'}
                      </button>
                      {court && (
                        <Link to={`/courts/${court.id}`} className="btn btn-view">
                          View Court
                        </Link>
                      )}
                    </>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
};

export default BookingsPage;