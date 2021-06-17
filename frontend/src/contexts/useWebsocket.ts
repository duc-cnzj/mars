import React, { useContext } from "react";

export const WsContext = React.createContext<WebSocket | null>(null);

export function useWs(): WebSocket| null {
  let ws = useContext(WsContext);
  if (ws) {
    return ws;
  }

  return null;
}
