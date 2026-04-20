// Client-side authentication
export const isAuthenticated = () => {
    return localStorage.getItem('marubot_auth') === 'true';
};

export const logout = () => {
    localStorage.removeItem('marubot_auth');
    window.location.href = '/login';
};

export const authenticatedFetch = async (url: string, options?: RequestInit) => {
    const response = await fetch(url, options);
    if (response.status === 401) {
        logout();
        throw new Error('Unauthorized');
    }
    return response;
};

export const signIn = async (password: string) => {
    const response = await fetch('/api/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ password }),
    });
    if (response.ok) {
        localStorage.setItem('marubot_auth', 'true');
        return true;
    }
    return false;
};
