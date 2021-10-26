import React, { FC } from "react";
import { Layout } from "antd";
import AppContent from "./components/AppContent";
import { ProvideWebsocket } from "./contexts/useWebsocket";
import { Switch, Route } from "react-router-dom";
import GitlabProjectManager from "./components/GitlabProjectManager";
import AppHeader from "./components/AppHeader";
import { PrivateRoute } from "./contexts/auth";
const { Header, Content, Footer } = Layout;

const App: FC = () => {
  return (
    <ProvideWebsocket>
      <Layout className="app">
        <Header style={{ position: "fixed", zIndex: 1, width: "100%" }}>
          <AppHeader />
        </Header>
        <Content className="app-content">
          <Switch>
            <PrivateRoute path={`/gitlab_project_manager`}>
              <GitlabProjectManager />
            </PrivateRoute>
            <PrivateRoute path={`/`} exact>
              <AppContent />
            </PrivateRoute>
            <Route path="*" exact>
              404
            </Route>
          </Switch>
        </Content>
        <Footer className="app-footer">
          <div className="copyright">created by duc@2021.</div>
        </Footer>
      </Layout>
    </ProvideWebsocket>
  );
};

export default App;
