import { createAsyncThunk } from "@reduxjs/toolkit";

export const setUserThunk = createAsyncThunk(`user/update`,
    async (values) => {
        //api request for change User
    }, 
)

export const userBuilder = (builder) => {
    builder.addCase(setUserThunk.pending, (state, action) => {
        state.isLoading = true;
    }).addCase(setUserThunk.fulfilled, (state, action) => {
        state.isLoading = false;
        state.user = action.payload;
        state.error = null;
    }).addCase(setUserThunk.rejected, (state, action) => {
        state.isLoading = false;
        state.error = action.payload;
    })
}