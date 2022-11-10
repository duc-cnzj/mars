import { configureStore } from '@reduxjs/toolkit'
import rootReducer from "./reducers";

const store = configureStore({
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware({
      serializableCheck: false,
      immutableCheck: false,
    }),
  devTools: process.env.NODE_ENV !== "production",
  reducer: rootReducer,
})
export default store;