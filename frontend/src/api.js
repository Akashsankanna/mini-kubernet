import axios from 'axios';

const API_BASE_URL =
  import.meta.env.VITE_API_URL || 'http://localhost:8081/api/v1';

const api = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Helper function to get auth data safely
const getAuthStorage = () => {
  try {
    const authData = localStorage.getItem('auth-storage');

    if (!authData) return null;

    return JSON.parse(authData);
  } catch (error) {
    console.error('Failed to parse auth storage:', error);
    return null;
  }
};

// Request interceptor
api.interceptors.request.use(
  (config) => {
    const authStorage = getAuthStorage();
    const token = authStorage?.state?.token;

    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }

    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor
api.interceptors.response.use(
  (response) => response,

  async (error) => {
    const originalRequest = error.config;

    // Handle Unauthorized
    if (
      error.response?.status === 401 &&
      !originalRequest._retry
    ) {
      originalRequest._retry = true;

      try {
        const authStorage = getAuthStorage();
        const refreshToken = authStorage?.state?.refreshToken;

        if (!refreshToken) {
          window.location.href = '/login';
          return Promise.reject(error);
        }

        // Refresh token request
        const response = await axios.post(
          `${API_BASE_URL}/auth/token/refresh`,
          {
            refresh_token: refreshToken,
          }
        );

        const accessToken = response.data.access_token;

        // Update local storage
        if (authStorage) {
          authStorage.state.token = accessToken;

          localStorage.setItem(
            'auth-storage',
            JSON.stringify(authStorage)
          );
        }

        // Retry original request
        originalRequest.headers.Authorization = `Bearer ${accessToken}`;

        return api(originalRequest);

      } catch (refreshError) {
        console.error('Token refresh failed:', refreshError);

        localStorage.removeItem('auth-storage');

        window.location.href = '/login';

        return Promise.reject(refreshError);
      }
    }

    return Promise.reject(error);
  }
);

export default api;