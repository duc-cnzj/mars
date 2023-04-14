import React, { FC, lazy, Suspense } from "react";
import { Layout } from "antd";
import AppContent from "./components/AppContent";
import { ProvideWebsocket } from "./contexts/useWebsocket";
import { Switch, Route } from "react-router-dom";
import AppHeader from "./components/AppHeader";
import AppFooter from "./components/AppFooter";
import { PrivateRoute } from "./contexts/auth";
import { GlobalScrollbar } from "mac-scrollbar";
import "mac-scrollbar/dist/mac-scrollbar.css";
import { css } from "@emotion/css";
import appTheme from "./styles/theme";

const { Header, Content, Footer } = Layout;

const GitProjectManager = lazy(() => import("./components/GitProjectManager"));
const Events = lazy(() => import("./components/Events"));
const AccessTokenManager = lazy(
  () => import("./components/AccessTokenManager")
);

const App: FC = () => {
  return (
    <ProvideWebsocket>
      <GlobalScrollbar />
      <Layout className="app">
        <Header
          className={css`
            position: "fixed";
            z-index: 1;
            width: "100%";
            overflow: "hidden";
            background-image: ${appTheme.mainLinear};
          `}
          style={{
            position: "fixed",
            zIndex: 1,
            width: "100%",
            overflow: "hidden",
          }}
        >
          <AppHeader />
        </Header>
        <Content
          className={css`
            margin-top: 64px;
            padding: 0 50px;
            min-height: calc(100vh - 122px) !important;
          `}
        >
          <Switch>
            <PrivateRoute path={`/git_project_manager`}>
              <Suspense fallback={null}>
                <GitProjectManager />
              </Suspense>
            </PrivateRoute>
            <PrivateRoute path={`/events`}>
              <Suspense fallback={null}>
                <Events />
              </Suspense>
            </PrivateRoute>
            <PrivateRoute path={`/access_token_manager`}>
              <Suspense fallback={null}>
                <AccessTokenManager />
              </Suspense>
            </PrivateRoute>
            <PrivateRoute path={`/`} exact>
              <AppContent />
            </PrivateRoute>
            <Route path="*" exact>
              404
            </Route>
          </Switch>
        </Content>
        <Footer
          className={css`
            background-image: ${appTheme.footerLinear};
            padding: 8px 50px 2px !important;
            text-align: center;
            .copyright {
              font-family: "Gill Sans", "Gill Sans MT", Calibri, "Trebuchet MS",
                sans-serif;
              color: ${appTheme.mainFontColor};
            }
          `}
        >
          <AppFooter />
        </Footer>
      </Layout>
    </ProvideWebsocket>
  );
};

export default App;
