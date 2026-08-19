let accessToken = null;

export function getAccessToken(){
    accessToken = localStorage.getItem("accessToken");
}

export function setAccessToken(token) {
    localStorage.setItem("accessToken", token);
}


class ApiClient {
    #baseurl

    constructor(baseurl) {
        this.#baseurl = baseurl;
    }

    async request(endpoint, options = {}, retry = true) {
        const headers = {
            "Content-Type": "application/json",
            ...options.headers
        }
        if (accessToken){
            headers.Authorization = `Bearer ${getAccessToken()}`;
        }
        const response = await fetch(this.#baseurl + endpoint,
            {
                ...options,
                headers,
                credentials: "include",
            }
        )

        if ((response.status == 401) && retry){
            refreshed = await this.refresh();

            if (refreshed){
                return this.request(endpoint, options, false);
            }

            throw new Error("Unauthorized");
        }

        if (!response.ok) {
            throw new Error(`API error: ${response.status}`);
        }

        return response.json();
    }

    async refresh() {
        const response = await fetch(this.#baseurl + "/auth/refresh", {
            method: "POST",
            credentials: "include"
        });

        if (!response.ok) {
            return false;
        } 

        const data = await response.json();

        setAccessToken(data["accessToken"]);

        return true;
    }

    get(endpoint){
        return this.request(endpoint, {method: "GET"}, true)
    }

    post(endpoint, body){
        return this.request(endpoint, {
            method: "POST", 
            body: JSON.stringify(body)}, true);
    }
}

export default new ApiClient("link"); //todo link