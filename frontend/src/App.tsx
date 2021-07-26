import React, { FC, useEffect, useState, useCallback } from "react";
import { Layout } from "antd";
import AppContent from "./components/AppContent";
import { WsContext } from "./contexts/useWebsocket";
import { handleEvents } from "./store/actions";
import { isJsonString } from "./utils/json";
import { useDispatch } from "react-redux";
import { Switch, Route, Link } from "react-router-dom";
import GitlabProjectManager from "./components/GitlabProjectManager";
import { getUid } from "./utils/uid";

const { Header, Content, Footer } = Layout;

export interface WsResponse {
  type: string;
  slug: string;
  result: string;
  data: string;
  end: boolean;
}

const App: FC = () => {
  const dispatch = useDispatch();
  const [ws, setWs] = useState<any>();

  const connectWs = useCallback(() => {
    console.log("ws init");
    let url: string = process.env.REACT_APP_WS_URL
      ? process.env.REACT_APP_WS_URL
      : "";
    if (url === "") {
      let isHttps = "https:" === window.location.protocol ? true : false;
      url = `${isHttps ? "wss" : "ws"}://${window.location.host}/ws`;
    }
    let uid = getUid()
    if (uid) {
      url+="?uid="+uid
    }
    let conn = new WebSocket(url);
    setWs(conn);
    conn.onopen = function (evt) {
      console.log("ws onopen");
    };
    conn.onclose = function (evt) {
      setWs(null);
      console.log("ws closed");
    };
    conn.onmessage = function (evt) {
      if (!isJsonString(evt.data)) {
        return;
      }
      let data: WsResponse = JSON.parse(evt.data);
      console.log(data, data.type);
      dispatch(handleEvents(data.slug, data));
      console.log("onmessage", evt.data);
    };
  }, [dispatch]);

  useEffect(() => {
    if (!ws) {
      connectWs();
    }
  }, [connectWs, ws]);

  return (
    <WsContext.Provider value={ws}>
      <Layout className="app">
        <Header style={{ position: "fixed", zIndex: 1, width: "100%" }}>
          <Link to="/" className="app-title">
            Mars
          </Link>
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
    </WsContext.Provider>
  );
};

export default App;
