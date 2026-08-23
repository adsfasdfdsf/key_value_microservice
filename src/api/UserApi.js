import {setAccessToken, ApiClient} from "./ApiClient.js"


class UserApi {
    #api

    constructor(){
        this.#api = new ApiClient("http://localhost:1128")
    }

    async LogIn(email, password){
        const data = await (await this.#api.post("/api/v1/login", {email: email, password: password})).json();
        console.log(data);
        setAccessToken(data["accessToken"]);
    }

    async SignUp(email, password){
        const data = await (await this.#api.post("/api/v1/signup", {email: email, password: password})).json();
        console.log(data);
        setAccessToken(data["accessToken"]);
    }

    async GetUserValues(){
        const resp = await this.#api.get("/api/v1/getUserKeys")
        const data = await resp.json()
        console.log(data)
        return data
    }

    async AddKey(key, value){
        await this.#api.post("/api/v1/addKey", {key: key, value: value})
    }
}

export default new UserApi();

