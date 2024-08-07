import { useState, useEffect } from "react";
import pb from "../api/websocket";

export default function useProjectRoom(
  namespaceID: number,
  projectID: number,
  ws: WebSocket | null
) {
  const [online, setOnline] = useState(false);
  useEffect(() => {
    console.log(namespaceID, projectID, "ducxxx---");
    let s = pb.websocket.ProjectPodEventJoinInput.encode({
      type: pb.websocket.Type.ProjectPodEvent,
      join: true,
      namespaceId: namespaceID,
      projectId: projectID,
    }).finish();
    ws?.send(s);
    setOnline(true);
    return () => {
      let leave = pb.websocket.ProjectPodEventJoinInput.encode({
        type: pb.websocket.Type.ProjectPodEvent,
        join: false,
        namespaceId: namespaceID,
        projectId: projectID,
      }).finish();
      ws?.send(leave);
      setOnline(false);
    };
  }, [ws, namespaceID, projectID]);

  return online;
}
