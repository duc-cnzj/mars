import React, { useContext, useState, useEffect, useCallback } from "react";
import { useDispatch } from "react-redux";
import { handleEvents } from "../store/actions";
import { isJsonString } from "../utils/json";
import { getUid } from "../utils/uid";

interface State {
  ws: WebSocket | null;
  ready: boolean;
}
export const WsContext = React.createContext<State | null>({
  ws: null,
  ready: false,
});

export function useWs(): WebSocket | null {
  let ctx = useContext(WsContext);
  if (ctx) {
    return ctx.ws;
  }
  return null;
}

export function useWsReady(): boolean {
  let ctx = useContext(WsContext);
  if (ctx) {
    return ctx.ready;
  }

  return false;
}

export const ProvideWebsocket: React.FC = ({ children }) => {
  const dispatch = useDispatch();
  const [ws, setWs] = useState<any>();
  const [ready, setReady] = useState(false)

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
    setWs({ws: conn});
    conn.onopen = function (evt) {
      setReady(true)
      console.log("ws onopen");
    };
    conn.onclose = function (evt) {
      setWs(null);
      setReady(false)
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
    setWs((ws: any)=>({...ws, ready: ready}));
  }, [ready])

  useEffect(() => {
    if (!ws) {
      setTimeout(() => {
        connectWs();
      }, 2000);
    }
  }, [connectWs, ws]);

  return <WsContext.Provider value={ws}>{children}</WsContext.Provider>;
};
