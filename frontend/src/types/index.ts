// pickle/frontend/src/types/index.ts

// User-related types
export interface User {
    id: string;
    email: string;
    name: string;
    picture: string;
    createdAt: string;
  }
  
  // Court-related types
  export interface Court {
    id: string;
    name: string;
    address: string;
    latitude: number;
    longitude: number;
    numberOfCourts: number;
    amenities: string[];
    imageUrl: string;
  }
  
  export interface GetCourtsRequest {
    city?: string;
    latitude?: number;
    longitude?: number;
    radiusKm?: number;
  }
  
  export interface GetCourtsResponse {
    courts: Court[];
  }
  
  export interface GetCourtRequest {
    courtId: string;
  }
  
  // Booking-related types
  export enum BookingStatus {
    PENDING = 'PENDING',
    CONFIRMED = 'CONFIRMED',
    CANCELLED = 'CANCELLED',
  }
  
  export interface Booking {
    id: string;
    courtId: string;
    userId: string;
    date: string;
    startTime: string;
    endTime: string;
    numberOfPlayers: number;
    playerEmails: string[];
    status: BookingStatus;
    createdAt: string;
    updatedAt: string;
  }
  
  export interface CreateBookingRequest {
    courtId: string;
    date: string;
    startTime: string;
    endTime: string;
    numberOfPlayers: number;
    playerEmails: string[];
  }
  
  export interface GetBookingsRequest {
    userId?: string;
    courtId?: string;
    date?: string;
  }
  
  export interface GetBookingsResponse {
    bookings: Booking[];
  }
  
  export interface UpdateBookingRequest {
    bookingId: string;
    startTime?: string;
    endTime?: string;
    numberOfPlayers?: number;
    playerEmails?: string[];
  }
  
  export interface CancelBookingRequest {
    bookingId: string;
  }
  
  export interface CancelBookingResponse {
    success: boolean;
    message: string;
  }
  
  // Authentication-related types
  export interface LoginResponse {
    token: string;
    user: User;
  }
  
  export interface AuthState {
    isAuthenticated: boolean;
    user: User | null;
    loading: boolean;
    error: string | null;
  }