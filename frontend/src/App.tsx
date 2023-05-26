import React, { FC, lazy, Suspense } from "react";
import AppContent from "./components/AppContent";
import { Route, Routes, useNavigate } from "react-router-dom";
import { PrivateRoute } from "./contexts/auth";
import "mac-scrollbar/dist/mac-scrollbar.css";
import AppLayout from "./components/AppLayout";
import { Button, Result } from "antd";

const GitProjectManager = lazy(() => import("./components/GitProjectManager"));
const Events = lazy(() => import("./components/Events"));
const AccessTokenManager = lazy(
  () => import("./components/AccessTokenManager")
);

const App: FC = () => {
  const navigate = useNavigate();
  return (
    <Routes>
      <Route path="/" element={<AppLayout />}>
        <Route
          index
          element={
            <PrivateRoute>
              <AppContent />
            </PrivateRoute>
          }
        />
        <Route
          path="git_project_manager"
          element={
            <PrivateRoute>
              <Suspense fallback={null}>
                <GitProjectManager />
              </Suspense>
            </PrivateRoute>
          }
        />
        <Route
          path="events"
          element={
            <PrivateRoute>
              <Suspense fallback={null}>
                <Events />
              </Suspense>
            </PrivateRoute>
          }
        />
        <Route
          path="access_token_manager"
          element={
            <PrivateRoute>
              <Suspense fallback={null}>
                <AccessTokenManager />
              </Suspense>
            </PrivateRoute>
          }
        />

        <Route
          path="*"
          element={
            <Result
              status="404"
              title="404"
              subTitle="页面不存在~"
              extra={
                <Button type="primary" onClick={() => navigate("/")}>
                  返回主页
                </Button>
              }
            />
          }
        ></Route>
      </Route>
    </Routes>
  );
};

export default App;
