import { configureStore } from "@reduxjs/toolkit";
import userSlice from "./Slices/User/User.js"


export const store = configureStore({
    reducer: {
        user: userSlice
    }
})

