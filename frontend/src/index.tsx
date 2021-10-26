import React from "react";
import ReactDOM from "react-dom";
import "normalize.css";
import "./styles/index.less";
import App from "./App";
import reportWebVitals from "./reportWebVitals";
import { Provider } from "react-redux";
import store from "./store";
import { disableReactDevTools } from "@fvilers/disable-react-devtools";
import { BrowserRouter as Router, Switch } from "react-router-dom";
import { PrivateRoute, GuestRoute, ProvideAuth } from "./contexts/auth";
import Callback from "./components/AuthCallback";
import Login from "./components/Login";

if (process.env.NODE_ENV === "production") {
  disableReactDevTools();
}

ReactDOM.render(
  <React.StrictMode>
    <Provider store={store}>
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
    </Provider>
  </React.StrictMode>,
  document.getElementById("root")
);

reportWebVitals();
