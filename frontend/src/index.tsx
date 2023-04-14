import React, { Suspense, lazy } from "react";
import reportWebVitals from "./reportWebVitals";
import { Provider } from "react-redux";
import store from "./store";
import { disableReactDevTools } from "@fvilers/disable-react-devtools";
import { BrowserRouter as Router, Switch } from "react-router-dom";
import { PrivateRoute, GuestRoute, ProvideAuth } from "./contexts/auth";
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
            <Switch>
              <GuestRoute path="/auth/callback">
                <Callback />
              </GuestRoute>
              <GuestRoute path="/login">
                <Login />
              </GuestRoute>
              <PrivateRoute path="/">
                <App />
              </PrivateRoute>
            </Switch>
          </ConfigProvider>
        </ProvideAuth>
      </Router>
    </Suspense>
  </Provider>
);

reportWebVitals();
