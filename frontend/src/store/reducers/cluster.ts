import { SET_CLUSTER_INFO } from "../actionTypes";
const initialState: API.ClusterInfo = {
  status: "",
  free_memory: "",
  free_cpu: "",
  total_memory: "",
  free_request_cpu: "",
  free_request_memory: "",
  total_cpu: "",
  usage_memory_rate: "",
  usage_cpu_rate: "",
  request_memory_rate: "",
  request_cpu_rate: "",
};

export const selectClusterInfo = (state: { cluster: API.ClusterInfo }) =>
  state.cluster;

export default function cluster(
  state = initialState,
  action: { type: string; info?: API.ClusterInfo }
) {
  switch (action.type) {
    case SET_CLUSTER_INFO:
      return { ...state, ...action.info };
    default:
      return state;
  }
}
