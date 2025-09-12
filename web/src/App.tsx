import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ThemeProvider } from './components/theme-provider.tsx';
import { AuthProvider } from './contexts/auth-context.tsx';
import { Toaster } from './components/ui/toaster.tsx';
import Layout from './components/layout/layout.tsx';
import LoginPage from './pages/auth/login.tsx';
import RegisterPage from './pages/auth/register.tsx';
import HomePage from './pages/home.tsx';
import DashboardPage from './pages/dashboard.tsx';
import PackageDetailPage from './pages/package/[name].tsx';
import CreatePackagePage from './pages/package/create.tsx';
import PushPackagePage from './pages/package/push.tsx';
import ProtectedRoute from './components/protected-route.tsx';

const queryClient = new QueryClient();

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <ThemeProvider defaultTheme="system" storageKey="protodex-ui-theme">
        <AuthProvider>
          <Router>
            <Layout>
              <Routes>
                <Route path="/login" element={<LoginPage />} />
                <Route path="/register" element={<RegisterPage />} />
                <Route path="/" element={<HomePage />} />
                <Route path="/package/:name" element={<PackageDetailPage />} />
                <Route
                  path="/dashboard"
                  element={
                    <ProtectedRoute>
                      <DashboardPage />
                    </ProtectedRoute>
                  }
                />
                <Route
                  path="/package/create"
                  element={
                    <ProtectedRoute>
                      <CreatePackagePage />
                    </ProtectedRoute>
                  }
                />
                <Route
                  path="/package/push"
                  element={
                    <ProtectedRoute>
                      <PushPackagePage />
                    </ProtectedRoute>
                  }
                />
              </Routes>
            </Layout>
          </Router>
          <Toaster />
        </AuthProvider>
      </ThemeProvider>
    </QueryClientProvider>
  );
}

export default App;