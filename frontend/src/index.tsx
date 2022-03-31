import React, { Suspense, lazy } from "react";
import ReactDOM from "react-dom";
import "./styles/index.less";
import reportWebVitals from "./reportWebVitals";
import { Provider } from "react-redux";
import store from "./store";
import { disableReactDevTools } from "@fvilers/disable-react-devtools";
import { BrowserRouter as Router, Switch } from "react-router-dom";
import { PrivateRoute, GuestRoute, ProvideAuth } from "./contexts/auth";

const Login = lazy(() => import("./components/Login"));
const Callback = lazy(() => import("./components/AuthCallback"));
const App = lazy(() => import("./App"));

if (process.env.NODE_ENV === "production") {
  disableReactDevTools();
}

ReactDOM.render(
  <React.StrictMode>
    <Provider store={store}>
      <Suspense fallback={null}>
        <Router>
          <ProvideAuth>
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
          </ProvideAuth>
        </Router>
      </Suspense>
    </Provider>
  </React.StrictMode>,
  document.getElementById("root")
);

reportWebVitals();
