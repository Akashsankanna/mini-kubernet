import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import api from '../api';

const useAuthStore = create(
  persist(
    (set, get) => ({
      user: null,
      token: null,
      refreshToken: null,
      isLoading: false,
      error: null,

      // Login
      login: async (username, password) => {
        set({ isLoading: true, error: null });

        try {
          const response = await api.post('/auth/login', {
            username,
            password,
          });

          const { access_token, refresh_token, user } =
            response.data;

          set({
            user,
            token: access_token,
            refreshToken: refresh_token,
            isLoading: false,
          });

          return response.data;

        } catch (error) {
          const errorMsg =
            error.response?.data?.message ||
            'Login failed';

          set({
            error: errorMsg,
            isLoading: false,
          });

          throw error;
        }
      },

      // Register
      register: async (data) => {
        set({ isLoading: true, error: null });

        try {
          const response = await api.post(
            '/auth/register',
            data
          );

          set({ isLoading: false });

          return response.data;

        } catch (error) {
          const errorMsg =
            error.response?.data?.message ||
            'Registration failed';

          set({
            error: errorMsg,
            isLoading: false,
          });

          throw error;
        }
      },

      // Request OTP
      requestOTP: async (email) => {
        set({ isLoading: true, error: null });

        try {
          const response = await api.post(
            '/auth/login/otp/request',
            { email }
          );

          set({ isLoading: false });

          return response.data;

        } catch (error) {
          const errorMsg =
            error.response?.data?.message ||
            'OTP request failed';

          set({
            error: errorMsg,
            isLoading: false,
          });

          throw error;
        }
      },

      // Verify OTP
      verifyOTP: async (email, otpCode) => {
        set({ isLoading: true, error: null });

        try {
          const response = await api.post(
            '/auth/login/otp/verify',
            {
              email,
              otp_code: otpCode,
            }
          );

          const { access_token, refresh_token, user } =
            response.data;

          set({
            user,
            token: access_token,
            refreshToken: refresh_token,
            isLoading: false,
          });

          return response.data;

        } catch (error) {
          const errorMsg =
            error.response?.data?.message ||
            'OTP verification failed';

          set({
            error: errorMsg,
            isLoading: false,
          });

          throw error;
        }
      },

      // Google Login
      googleLogin: async (token) => {
        set({ isLoading: true, error: null });

        try {
          const response = await api.post(
            '/auth/login/google',
            { token }
          );

          const { access_token, refresh_token, user } =
            response.data;

          set({
            user,
            token: access_token,
            refreshToken: refresh_token,
            isLoading: false,
          });

          return response.data;

        } catch (error) {
          const errorMsg =
            error.response?.data?.message ||
            'Google login failed';

          set({
            error: errorMsg,
            isLoading: false,
          });

          throw error;
        }
      },

      // Refresh Access Token
      refreshAccessToken: async () => {
        try {
          const { refreshToken } = get();

          if (!refreshToken) {
            throw new Error('No refresh token');
          }

          const response = await api.post(
            '/auth/token/refresh',
            {
              refresh_token: refreshToken,
            }
          );

          set({
            token: response.data.access_token,
          });

          return response.data;

        } catch (error) {
          get().logout();
          throw error;
        }
      },

      // Logout
      logout: async () => {
        try {
          await api.post('/auth/logout');

        } catch (error) {
          console.error(
            'Logout error:',
            error
          );

        } finally {
          set({
            user: null,
            token: null,
            refreshToken: null,
            error: null,
          });
        }
      },

      // Error handlers
      setError: (error) =>
        set({ error }),

      clearError: () =>
        set({ error: null }),

      // Auth check
      isAuthenticated: () =>
        !!get().token,
    }),

    {
      name: 'auth-storage',

      partialize: (state) => ({
        user: state.user,
        token: state.token,
        refreshToken: state.refreshToken,
      }),
    }
  )
);

export default useAuthStore;