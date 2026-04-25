/**
 * Sovereign Intelligence Core - Production-Grade Auth Client
 */

let accessToken: string | null = null;
let refreshTimer: number | null = null;

async function login(credentials: object) {
    try {
        const response = await fetch('/api/auth/login', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(credentials)
        });

        if (!response.ok) throw new Error('Login failed');

        const { token, expiresIn } = await response.json();
        accessToken = token;

        // Start refresh loop 60s before expiry
        scheduleRefresh(expiresIn - 60);
        
        window.location.href = '/dashboard';
    } catch (err) {
        console.error(err);
        alert('Authentication Error');
    }
}

function scheduleRefresh(delaySeconds: number) {
    if (refreshTimer) clearTimeout(refreshTimer);
    
    refreshTimer = window.setTimeout(async () => {
        const success = await refreshToken();
        if (!success) {
            window.location.href = '/login';
        }
    }, delaySeconds * 1000);
}

async function refreshToken(): Promise<boolean> {
    try {
        const response = await fetch('/api/auth/refresh', {
            method: 'POST',
            // Refresh token is in an httpOnly cookie, so it's sent automatically
        });

        if (!response.ok) return false;

        const { token, expiresIn } = await response.json();
        accessToken = token;
        scheduleRefresh(expiresIn - 60);
        return true;
    } catch (err) {
        return false;
    }
}

/**
 * Wrapper for authenticated fetch calls.
 */
async function authFetch(url: string, options: RequestInit = {}): Promise<Response> {
    const headers = new Headers(options.headers || {});
    if (accessToken) {
        headers.set('Authorization', `Bearer ${accessToken}`);
    }

    // CSRF Protection
    const csrfToken = getCookie('csrf_token');
    if (csrfToken) {
        headers.set('X-CSRF-Token', csrfToken);
    }

    options.headers = headers;
    let response = await fetch(url, options);

    // If unauthorized, try one refresh
    if (response.status === 401) {
        const success = await refreshToken();
        if (success) {
            headers.set('Authorization', `Bearer ${accessToken}`);
            response = await fetch(url, options);
        } else {
            window.location.href = '/login';
        }
    }

    return response;
}

function getCookie(name: string): string | null {
    const value = `; ${document.cookie}`;
    const parts = value.split(`; ${name}=`);
    if (parts.length === 2) return parts.pop()?.split(';').shift() || null;
    return null;
}
