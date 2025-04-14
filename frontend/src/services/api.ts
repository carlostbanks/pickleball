// pickle/frontend/src/services/api.ts
import axios, { AxiosInstance, AxiosRequestConfig, InternalAxiosRequestConfig } from 'axios';
import {
  Booking,
  CancelBookingRequest,
  CancelBookingResponse,
  Court,
  CreateBookingRequest,
  GetBookingsRequest,
  GetBookingsResponse,
  GetCourtRequest,
  GetCourtsRequest,
  GetCourtsResponse,
  UpdateBookingRequest,
  User,
} from '../types';

// API base URL
const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

// Create Axios instance
const api: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
  withCredentials: true, // Include cookies in requests
});

// Add request interceptor to include token if available
api.interceptors.request.use(
    (config: InternalAxiosRequestConfig) => {
    const token = localStorage.getItem('token');
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// API service
export const apiService = {
  // Auth endpoints
  auth: {
    // Get the current user
    getCurrentUser: async (): Promise<User> => {
      const response = await api.get('/api/users/me');
      return response.data;
    },

    // Log out
    logout: async (): Promise<void> => {
      await api.post('/auth/logout');
      localStorage.removeItem('token');
    },
  },

  // Court endpoints
  courts: {
    // Get courts based on search criteria
    getCourts: async (request: GetCourtsRequest): Promise<GetCourtsResponse> => {
      const response = await api.get('/api/courts', { params: request });
      return response.data;
    },

    // Get a specific court by ID
    getCourt: async (request: GetCourtRequest): Promise<Court> => {
      const response = await api.get(`/api/courts/${request.courtId}`);
      return response.data;
    },
  },

  // Booking endpoints
  bookings: {
    // Create a new booking
    createBooking: async (request: CreateBookingRequest): Promise<Booking> => {
      const response = await api.post('/api/bookings', request);
      return response.data;
    },

    // Get bookings based on filter criteria
    getBookings: async (request: GetBookingsRequest): Promise<GetBookingsResponse> => {
      const response = await api.get('/api/bookings', { params: request });
      return response.data;
    },

    // Update an existing booking
    updateBooking: async (request: UpdateBookingRequest): Promise<Booking> => {
      const response = await api.put(`/api/bookings/${request.bookingId}`, request);
      return response.data;
    },

    // Cancel a booking
    cancelBooking: async (request: CancelBookingRequest): Promise<CancelBookingResponse> => {
      const response = await api.delete(`/api/bookings/${request.bookingId}`);
      return response.data;
    },
  },
};

export default apiService;