import React, { FC, lazy, Suspense } from "react";
import AppContent from "./components/AppContent";
import { Route, Routes, useNavigate } from "react-router-dom";
import "mac-scrollbar/dist/mac-scrollbar.css";
import AppLayout from "./components/AppLayout";
import { Button, Result } from "antd";

const Events = lazy(() => import("./pages/Events"));
const RepoPage = lazy(() => import("./pages/Repo"));
const AccessTokenManager = lazy(() => import("./pages/AccessTokenManager"));

const App: FC = () => {
  const navigate = useNavigate();

  return (
    <Routes>
      <Route path="/" element={<AppLayout />}>
        <Route index element={<AppContent />} />
        <Route
          path="repos"
          element={
            <Suspense fallback={null}>
              <RepoPage />
            </Suspense>
          }
        />
        <Route
          path="events"
          element={
            <Suspense fallback={null}>
              <Events />
            </Suspense>
          }
        />
        <Route
          path="access_token_manager"
          element={
            <Suspense fallback={null}>
              <AccessTokenManager />
            </Suspense>
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
