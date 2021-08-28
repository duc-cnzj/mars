import React, { FC } from "react";
import { Layout } from "antd";
import AppContent from "./components/AppContent";
import { ProvideWebsocket } from "./contexts/useWebsocket";
import { Switch, Route, Link } from "react-router-dom";
import GitlabProjectManager from "./components/GitlabProjectManager";
import ClusterInfo from "./components/ClusterInfo";
const { Header, Content, Footer } = Layout;

const App: FC = () => {
  return (
    <ProvideWebsocket>
      <Layout className="app">
        <Header style={{ position: "fixed", zIndex: 1, width: "100%" }}>
          <div
            style={{
              display: "flex",
              alignItems: "center",
              justifyContent: "space-between",
            }}
          >
            <Link to="/" className="app-title">
              Mars
            </Link>
            <ClusterInfo />
          </div>
        </Header>
        <Content className="app-content">
          <Switch>
            <Route
              path="/web/gitlab_project_manager"
              component={GitlabProjectManager}
            />
            <Route path="*">
              <AppContent />
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
