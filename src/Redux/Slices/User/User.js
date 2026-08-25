import { createSlice } from "@reduxjs/toolkit";
import { userBuilder } from "./Thunk";


const initialState = {
    user: null,
    isLoading: false,
    error: null
}

const userSlice = createSlice({
    name: `user`,
    initialState,
    extraReducers: userBuilder,
    reducers: {
        setUser: function(state, action) {
            state.user = action.payload;
        }
    }
    
})



export default userSlice.reducer;