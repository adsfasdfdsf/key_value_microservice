let accessToken = null;
let sessionStatus = "checking";
let refreshPromise = null;
let expiryTimer = null;
const listeners = new Set();

export function getAccessToken() { return accessToken; }
export function getSessionStatus() { return sessionStatus; }
export function subscribeSession(listener) {
    listeners.add(listener);
    return () => listeners.delete(listener);
}

function updateStatus(status) {
    sessionStatus = status;
    listeners.forEach(listener => listener());
}

export function setAccessToken(token) {
    accessToken = token;
    clearTimeout(expiryTimer);
    updateStatus(token ? "authenticated" : "guest");
    if (token) {
        const payload = token.split(".")[1].replace(/-/g, "+").replace(/_/g, "/");
        const { exp } = JSON.parse(atob(payload));
        expiryTimer = setTimeout(() => {
            sessionApi.refresh().catch(() => updateStatus("error"));
        }, Math.max(0, exp * 1000 - Date.now()));
    }
}

export class ApiClient {
    #baseurl;
    constructor(baseurl) { this.#baseurl = baseurl; }

    async request(endpoint, options = {}, retry = true) {
        const usedToken = accessToken;
        const headers = { "Content-Type": "application/json", ...options.headers };
        if (usedToken) headers.Authorization = `Bearer ${usedToken}`;
        const response = await fetch(this.#baseurl + endpoint, {
            credentials: "include", ...options, headers,
        });
        const isAuthRequest = endpoint === "/api/v1/login" || endpoint === "/api/v1/signup";
        if (response.status === 401 && !isAuthRequest) {
            if (retry) {
                if ((accessToken && accessToken !== usedToken) || await this.refresh()) {
                    return this.request(endpoint, options, false);
                }
            }
            setAccessToken(null);
            throw new Error("Unauthorized");
        }
        if (!response.ok) throw new Error(`API error: ${response.status}`);
        return response;
    }

    async refresh() {
        if (!refreshPromise) {
            refreshPromise = this.#refresh().finally(() => { refreshPromise = null; });
        }
        return refreshPromise;
    }

    async #refresh() {
        const response = await fetch(this.#baseurl + "/api/v1/auth/refreshTokens", {
            method: "GET", credentials: "include",
            signal: AbortSignal.timeout(8000),
        });
        if (response.status === 401) {
            setAccessToken(null);
            return false;
        }
        if (!response.ok) throw new Error(`API error: ${response.status}`);
        const data = await response.json();
        setAccessToken(data.accessToken);
        return true;
    }

    get(endpoint) { return this.request(endpoint, { method: "GET" }); }
    post(endpoint, body) {
        return this.request(endpoint, { method: "POST", body: JSON.stringify(body) });
    }
}

export const sessionApi = new ApiClient("http://localhost:1128");
export async function restoreSession() {
    updateStatus("checking");
    try { await sessionApi.refresh(); }
    catch { updateStatus("error"); }
}
