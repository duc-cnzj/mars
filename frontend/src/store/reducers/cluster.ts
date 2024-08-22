import pb from "./../../api/websocket";
import { SET_CLUSTER_INFO } from "../actionTypes";
const initialState: pb.websocket.ClusterInfo = {
  status: "",
  freeMemory: "",
  freeCpu: "",
  totalMemory: "",
  freeRequestCpu: "",
  freeRequestMemory: "",
  totalCpu: "",
  usageMemoryRate: "",
  usageCpuRate: "",
  requestMemoryRate: "",
  requestCpuRate: "",
};

export const selectClusterInfo = (state: {
  cluster: pb.websocket.ClusterInfo;
}) => state.cluster;

export default function cluster(
  state = initialState,
  action: { type: string; info?: pb.websocket.ClusterInfo },
) {
  switch (action.type) {
    case SET_CLUSTER_INFO:
      return { ...state, ...action.info };
    default:
      return state;
  }
}
