import { createContext, useContext, useEffect, useState, ReactNode } from 'react';
import { useToast } from '@/hooks/use-toast.ts';

interface User {
    id: number;
    username: string;
}

interface AuthContextType {
    user: User | null;
    login: (username: string, password: string) => Promise<void>;
    register: (username: string, password: string) => Promise<void>;
    logout: () => Promise<void>;
    loading: boolean;
    authenticatedFetch: (url: string, options?: RequestInit) => Promise<Response>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
    const [user, setUser] = useState<User | null>(null);
    const [loading, setLoading] = useState(true);
    const [token, setToken] = useState<string | null>(localStorage.getItem('auth_token'));
    const { toast } = useToast();

    const authenticatedFetch = (url: string, options: RequestInit = {}) => {
        return fetch(url, {
            ...options,
            headers: {
                ...options.headers,
                ...(token && { Authorization: `Bearer ${token}` }),
            },
        });
    };

    useEffect(() => {
        checkAuth();
    }, []);

    const checkAuth = async () => {
        if (!token) {
            setLoading(false);
            return;
        }

        try {
            const response = await authenticatedFetch('/api/auth/me');

            if (response.ok) {
                const userData = await response.json();
                setUser(userData);
            } else {
                localStorage.removeItem('auth_token');
                setToken(null);
            }
        } catch (error) {
            console.error('Auth check failed:', error);
        } finally {
            setLoading(false);
        }
    };

    const login = async (username: string, password: string) => {
        try {
            const response = await fetch('/api/auth/login', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ username, password }),
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || 'Login failed');
            }

            const data = await response.json();
            setToken(data.token);
            setUser(data.user);
            localStorage.setItem('auth_token', data.token);

            toast({
                title: "Success",
                description: "Logged in successfully!",
            });
        } catch (error) {
            toast({
                title: "Error",
                description: error instanceof Error ? error.message : 'Login failed',
                variant: "destructive",
            });
            throw error;
        }
    };

    const register = async (username: string, password: string) => {
        try {
            const response = await fetch('/api/auth/register', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ username, password }),
            });

            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData.error || 'Registration failed');
            }

            const data = await response.json();
            setToken(data.token);
            setUser(data.user);
            localStorage.setItem('auth_token', data.token);

            toast({
                title: "Success",
                description: "Account created successfully!",
            });
        } catch (error) {
            toast({
                title: "Error",
                description: error instanceof Error ? error.message : 'Registration failed',
                variant: "destructive",
            });
            throw error;
        }
    };

    const logout = async () => {
        localStorage.removeItem('auth_token');
        setToken(null);
        setUser(null);
        toast({
            title: "Success",
            description: "Logged out successfully!",
        });
    };

    const value = {
        user,
        login,
        register,
        logout,
        loading,
        authenticatedFetch,
    };

    return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export const useAuth = () => {
    const context = useContext(AuthContext);
    if (context === undefined) {
        throw new Error('useAuth must be used within an AuthProvider');
    }
    return context;
};