let accessToken = null;

export function getAccessToken(){
    return accessToken
}

export function setAccessToken(token) {
    accessToken = token
}


export class ApiClient {
    #baseurl

    constructor(baseurl) {
        this.#baseurl = baseurl;
    }

    async request(endpoint, options = {}, retry = true) {
        const headers = {
            "Content-Type": "application/json",
            ...options.headers,
            
        }
        if (accessToken){
            headers.Authorization = `Bearer ${getAccessToken()}`;
        }
        
        const response = await fetch(this.#baseurl + endpoint,
            {
                credentials: "include", //TODO убрать в проде без Cors

                ...options,
                headers,
            }
        )

        if ((response.status == 401) && retry){
            const refreshed = await this.refresh();
            console.log("here")
            if (refreshed){
                return this.request(endpoint, options, false);
            }

            throw new Error("Unauthorized");
        }

        if (!response.ok) {
            throw new Error(`API error: ${response.status}`);
        }
        console.log(`foolfiled ${endpoint}`)
        return response;
    }

    async refresh() {
        const response = await fetch(this.#baseurl + "/api/v1/auth/refreshTokens", {
            method: "GET",
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
