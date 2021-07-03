import React from "react";
import ReactDOM from "react-dom";
import "normalize.css";
import "./styles/index.less";
import App from "./App";
import reportWebVitals from "./reportWebVitals";
import { Provider } from "react-redux";
import store from "./store";
import { disableReactDevTools } from '@fvilers/disable-react-devtools';
import { BrowserRouter as Router } from "react-router-dom";

if (process.env.NODE_ENV === 'production') {
  disableReactDevTools();
}

ReactDOM.render(
  <React.StrictMode>
    <Provider store={store}>
      <Router>
        <App />
      </Router>
    </Provider>
  </React.StrictMode>,
  document.getElementById("root")
);

reportWebVitals();
