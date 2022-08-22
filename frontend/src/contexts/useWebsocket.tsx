import React, { useContext, useState, useEffect, useCallback } from "react";
import { useDispatch } from "react-redux";
import { handleEvents } from "../store/actions";
import { getUid } from "../utils/uid";
import { getToken } from "../utils/token";
import { message } from "antd";
import pb from "../api/compiled";

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

export const ProvideWebsocket: React.FC<{children: React.ReactNode;}> = ({ children }) => {
  const dispatch = useDispatch();
  const [ws, setWs] = useState<any>();

  const connectWs = useCallback(() => {
    let token = getToken();
    if (!token) {
      message.error("用户未登录");
      return;
    }
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
    conn.binaryType = "arraybuffer"
    conn.onopen = function (evt) {
      setWs({ ws: conn, ready: true });
      conn.send(
        pb.websocket.AuthorizeTokenInput.encode({
          token: getToken(),
          type: pb.websocket.Type.HandleAuthorize,
        }).finish()
      );
    };
    conn.onclose = function (evt) {
      setWs({ ws: null, ready: false });
      console.log("ws closed");
    };
    conn.onmessage = function (evt) {
      let data: pb.websocket.WsMetadataResponse = pb.websocket.WsMetadataResponse.decode(new Uint8Array(evt.data))
      data.metadata && dispatch(handleEvents(data.metadata.slug, data.metadata, new Uint8Array(evt.data)));
    };
  }, [dispatch]);

  useEffect(() => {
    if (!ws?.ws) {
      connectWs();
    }
  }, [connectWs, ws]);

  return <WsContext.Provider value={ws}>{children}</WsContext.Provider>;
};
