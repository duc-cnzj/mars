import {
  SET_CREATE_PROJECT_LOADING,
  APPEND_CREATE_PROJECT_LOG,
  SET_DEPLOY_STATUS,
  CLEAR_CREATE_PROJECT_LOG,
  SET_NAMESPACE_RELOAD,
  SET_PROCESS_PERCENT,
  SET_CLUSTER_INFO,
  SET_SHELL_SESSION_ID,
  SET_SHELL_LOG,
} from "./actionTypes";
import { DeployStatus } from "./reducers/createProject";
import { Dispatch } from "redux";
import { message } from "antd";
import { setUid } from "../utils/uid";
import pb from "../api/compiled";
import { debounce } from "lodash";

export const setCreateProjectLoading = (id: string, loading: boolean) => ({
  type: SET_CREATE_PROJECT_LOADING,
  data: {
    id,
    isLoading: loading,
  },
});
export const appendCreateProjectLog = (id: string, log: string) => ({
  type: APPEND_CREATE_PROJECT_LOG,
  data: {
    id,
    output: log,
  },
});

export const clearCreateProjectLog = (id: string) => ({
  type: CLEAR_CREATE_PROJECT_LOG,
  data: {
    id,
  },
});

export const setDeployStatus = (id: string, status: DeployStatus) => ({
  type: SET_DEPLOY_STATUS,
  data: {
    id: id,
    deployStatus: status,
  },
});

export const setNamespaceReload = (reload: boolean) => ({
  type: SET_NAMESPACE_RELOAD,
  data: {
    reload: reload,
  },
});

export const setProcessPercent = (id: string, percent: number) => ({
  type: SET_PROCESS_PERCENT,
  data: {
    id: id,
    processPercent: percent,
  },
});
export const setShellSessionId = (id: string, sessionID: string) => ({
  type: SET_SHELL_SESSION_ID,
  data: {
    id: id,
    sessionID: sessionID,
  },
});
export const setShellLog = (id: string, log: pb.websocket.TerminalMessage) => ({
  type: SET_SHELL_LOG,
  data: {
    id: id,
    log: log,
  },
});

export const setClusterInfo = (info: pb.cluster.InfoResponse) => ({
  type: SET_CLUSTER_INFO,
  info: info,
});

const debounceLoadNamespace = debounce((dispatch: Dispatch) => {
  dispatch(setNamespaceReload(true));
}, 500);

export const handleEvents = (
  id: string,
  data: pb.websocket.Metadata,
  input: any
) => {
  return function (dispatch: Dispatch) {
    switch (data.type.valueOf()) {
      case pb.websocket.Type.SetUid:
        setUid(data.message);
        break;
      case pb.websocket.Type.ClusterInfoSync:
        let info = pb.websocket.WsHandleClusterResponse.decode(input);
        info.info && dispatch(setClusterInfo(info.info));
        break;
      case pb.websocket.Type.ReloadProjects:
        debounceLoadNamespace(dispatch);
        break;
      case pb.websocket.Type.UpdateProject:
        dispatch(appendCreateProjectLog(id, data.message ? data.message : ""));

        if (data.end) {
          switch (data.result) {
            case pb.websocket.ResultType.Deployed:
              dispatch(setDeployStatus(id, DeployStatus.DeployUpdateSuccess));
              message.success("部署成功");
              dispatch(clearCreateProjectLog(id));
              break;
            case pb.websocket.ResultType.DeployedCanceled:
              dispatch(setDeployStatus(id, DeployStatus.DeployCanceled));
              dispatch(appendCreateProjectLog(id, "部署已取消"));
              message.warn("部署已取消");
              break;
            case pb.websocket.ResultType.DeployedFailed:
            default:
              dispatch(setDeployStatus(id, DeployStatus.DeployFailed));
              dispatch(appendCreateProjectLog(id, "部署失败"));
              message.error("部署失败");
              break;
          }
          dispatch(setCreateProjectLoading(id, false));
        }
        break;
      case pb.websocket.Type.CreateProject:
        dispatch(appendCreateProjectLog(id, data.message ? data.message : ""));

        if (data.end) {
          switch (data.result) {
            case pb.websocket.ResultType.Deployed:
              dispatch(setDeployStatus(id, DeployStatus.DeploySuccess));
              message.success("部署成功");
              dispatch(clearCreateProjectLog(id));
              break;
            case pb.websocket.ResultType.DeployedCanceled:
              dispatch(setDeployStatus(id, DeployStatus.DeployCanceled));
              dispatch(appendCreateProjectLog(id, "部署已取消"));
              message.warn("部署已取消");
              break;
            case pb.websocket.ResultType.DeployedFailed:
            default:
              dispatch(setDeployStatus(id, DeployStatus.DeployFailed));
              dispatch(appendCreateProjectLog(id, "部署失败"));
              message.error("部署失败");
              break;
          }
          dispatch(setCreateProjectLoading(id, false));
          debounceLoadNamespace(dispatch);
        }
        break;
      case pb.websocket.Type.ProcessPercent:
        dispatch(setProcessPercent(id, Number(data.message)));
        break;
      case pb.websocket.Type.HandleExecShell:
        if (data.result === pb.websocket.ResultType.Error) {
          message.error(data.message);
          break;
        }
        let res = pb.websocket.WsHandleShellResponse.decode(input);

        res.container &&
          res.terminal_message &&
          dispatch(
            setShellSessionId(
              `${res.container.namespace}|${res.container.pod}|${res.container.container}`,
              res.terminal_message.session_id
            )
          );
        break;
      case pb.websocket.Type.HandleExecShellMsg:
        let logRes = pb.websocket.WsHandleShellResponse.decode(input);
        logRes.container &&
          logRes.terminal_message &&
          dispatch(
            setShellLog(
              `${logRes.container.namespace}|${logRes.container.pod}|${logRes.container.container}`,
              logRes.terminal_message
            )
          );
        break;
      case pb.websocket.Type.HandleAuthorize:
        if (data.result === pb.websocket.ResultType.Error) {
          message.error(data.message);
        }
        break;
      default:
        console.log("unknown event: ", data.type);
    }
  };
};
