import React, { useContext, useState, useEffect, useCallback } from "react";
import { useDispatch } from "react-redux";
import { handleEvents } from "../store/actions";
import { isJsonString } from "../utils/json";
import { getUid } from "../utils/uid";

export const WsContext = React.createContext<WebSocket | null>(null);

export function useWs(): WebSocket | null {
  return useContext(WsContext);
}

export const ProvideWebsocket: React.FC = ({ children }) => {
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
    let uid = getUid();
    if (uid) {
      url += "?uid=" + uid;
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
      let data: API.WsResponse = JSON.parse(evt.data);
      dispatch(handleEvents(data.slug, data));
    };
  }, [dispatch]);

  useEffect(() => {
    if (!ws) {
      connectWs();
    }
  }, [connectWs, ws]);

  return <WsContext.Provider value={ws}>{children}</WsContext.Provider>;
};
