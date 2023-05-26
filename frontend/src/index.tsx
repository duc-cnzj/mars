import React, { Suspense, lazy } from "react";
import reportWebVitals from "./reportWebVitals";
import { Provider } from "react-redux";
import store from "./store";
import { disableReactDevTools } from "@fvilers/disable-react-devtools";
import { Route, BrowserRouter as Router, Routes } from "react-router-dom";
import { GuestRoute, ProvideAuth } from "./contexts/auth";
import { createRoot } from "react-dom/client";
import { ConfigProvider } from "antd";
import theme from "./styles/theme";
import "./styles/index.css";
import "antd/dist/reset.css";
import "prism-themes/themes/prism-material-dark.css";

const Login = lazy(() => import("./components/Login"));
const Callback = lazy(() => import("./components/AuthCallback"));
const App = lazy(() => import("./App"));

if (process.env.NODE_ENV === "production") {
  disableReactDevTools();
}

const container = document.getElementById("root");
const root = createRoot(container!);

root.render(
  <Provider store={store}>
    <Suspense fallback={null}>
      <Router>
        <ProvideAuth>
          <ConfigProvider theme={{ token: theme }}>
            <Routes>
              <Route
                path="/auth/callback"
                element={
                  <GuestRoute>
                    <Callback />
                  </GuestRoute>
                }
              />
              <Route
                path="/login"
                element={
                  <GuestRoute>
                    <Login />
                  </GuestRoute>
                }
              />
              <Route path="/*" element={<App />} />
            </Routes>
          </ConfigProvider>
        </ProvideAuth>
      </Router>
    </Suspense>
  </Provider>
);

reportWebVitals();
