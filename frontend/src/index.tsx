import React, { Suspense, lazy, useEffect } from "react";
import { Provider } from "react-redux";
import store from "./store";
import { disableReactDevTools } from "@fvilers/disable-react-devtools";
import {
  Route,
  BrowserRouter as Router,
  Routes,
  useLocation,
} from "react-router-dom";
import { GuestRoute, PrivateRoute, ProvideAuth } from "./contexts/auth";
import { createRoot } from "react-dom/client";
import { ConfigProvider } from "antd";
import theme from "./styles/theme";
import "./styles/index.css";
import "antd/dist/reset.css";
import "prism-themes/themes/prism-material-dark.css";
import { useDispatch } from "react-redux";
import { setOpenedModals } from "./store/actions";

const Login = lazy(() => import("./components/Login"));
const Callback = lazy(() => import("./components/AuthCallback"));
const App = lazy(() => import("./App"));

if (process.env.NODE_ENV === "production") {
  disableReactDevTools();
}

const container = document.getElementById("root");
const root = createRoot(container!);

const RootElement: React.FC = () => {
  const lo = useLocation();
  const dispath = useDispatch();

  useEffect(() => {
    if (lo.pathname !== "/") {
      dispath(setOpenedModals({}));
    }
  }, [lo, dispath]);

  return (
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
          <Route
            path="/*"
            element={
              <PrivateRoute>
                <App />
              </PrivateRoute>
            }
          />
        </Routes>
      </ConfigProvider>
    </ProvideAuth>
  );
};

root.render(
  <Provider store={store}>
    <Suspense fallback={null}>
      <Router>
        <RootElement />
      </Router>
    </Suspense>
  </Provider>,
);
