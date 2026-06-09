import React, { useEffect } from 'react';
import {
  BrowserRouter as Router,
  Routes,
  Route,
  Navigate,
} from 'react-router-dom';

import { Toaster } from 'react-hot-toast';

import useAuthStore from './store/authStore';

import LoginPage from './pages/LoginPage';
import RegisterPage from './pages/RegisterPage';
import DashboardPage from './pages/DashboardPage';
import ProfilePage from './pages/ProfilePage';
import AdminPage from './pages/AdminPage';

import ProtectedRoute from './components/ProtectedRoute';
import Layout from './components/Layout';

function App() {

  // Zustand store values
  const user = useAuthStore(
    (state) => state.user
  );

  const token = useAuthStore(
    (state) => state.token
  );

  const isAuthenticated = useAuthStore(
    (state) => state.isAuthenticated
  );

  // Verify token on app load
  useEffect(() => {
    if (token) {
      console.log('User authenticated');
    }
  }, [token]);

  return (
    <Router>

      <Toaster position="top-right" />

      <Routes>

        {/* Public Routes */}
        <Route
          path="/login"
          element={
            isAuthenticated()
              ? <Navigate to="/dashboard" replace />
              : <LoginPage />
          }
        />

        <Route
          path="/register"
          element={<RegisterPage />}
        />

        {/* Protected Routes */}
        <Route element={<ProtectedRoute />}>

          <Route element={<Layout />}>

            <Route
              path="/"
              element={<DashboardPage />}
            />

            <Route
              path="/dashboard"
              element={<DashboardPage />}
            />

            <Route
              path="/profile"
              element={<ProfilePage />}
            />

            {/* Admin Route */}
            {user?.role === 'admin' && (
              <Route
                path="/admin"
                element={<AdminPage />}
              />
            )}

          </Route>

        </Route>

        {/* Redirect Unknown Routes */}
        <Route
          path="*"
          element={<Navigate to="/" replace />}
        />

      </Routes>

    </Router>
  );
}

export default App;