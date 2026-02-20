// Client-side authentication
export const isAuthenticated = () => {
    return localStorage.getItem('marubot_auth') === 'true';
};

export const logout = () => {
    localStorage.removeItem('marubot_auth');
    window.location.href = '/login';
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
