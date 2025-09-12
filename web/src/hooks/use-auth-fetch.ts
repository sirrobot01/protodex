import { useAuth } from '@/contexts/auth-context.tsx';

export const useAuthenticatedFetch = () => {
    const { authenticatedFetch } = useAuth();
    return authenticatedFetch;
};