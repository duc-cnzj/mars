import { useState, useEffect } from "react";
import pb from "../api/compiled";

export default function useProjectRoom(
  namespaceID: number,
  projectID: number,
  ws: WebSocket | null
) {
  const [online, setOnline] = useState(false);
  useEffect(() => {
    let s = pb.websocket.ProjectPodEventJoinInput.encode({
      type: pb.websocket.Type.ProjectPodEvent,
      join: true,
      namespace_id: namespaceID,
      project_id: projectID,
    }).finish();
    ws?.send(s);
    setOnline(true);
    return () => {
      let leave = pb.websocket.ProjectPodEventJoinInput.encode({
        type: pb.websocket.Type.ProjectPodEvent,
        join: false,
        namespace_id: namespaceID,
        project_id: projectID,
      }).finish();
      ws?.send(leave);
      setOnline(false);
    };
  }, [ws, namespaceID, projectID]);

  return online;
}
