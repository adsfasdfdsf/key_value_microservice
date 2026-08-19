import ApiClient from "./ApiClient"
import 


class UserApi {
    #api

    constructor(){
        this.#api = ApiClient
    }

    LogIn(email, password){
        this.#api.post()
    }

    SignUp(email, password){
        this.#api.post()
    }

    GetUserValues(){}
}