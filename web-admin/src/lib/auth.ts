// Client-side authentication mock
export const auth = {
    user: { name: 'Guest' },
    session: null
};

export const signIn = async () => {
    console.warn('Auth currently disabled in client-side build');
    return Promise.resolve();
};

export const signOut = async () => {
    console.warn('Auth currently disabled in client-side build');
    return Promise.resolve();
};

export const useSession = () => {
    return { data: null, status: 'unauthenticated' };
};
