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
  SET_TIMER_START_AT,
  SET_TIMER_START,
  PROJECT_POD_EVENT,
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

export const appendCreateProjectLog = (
  id: string,
  log: string,
  type: pb.websocket.ResultType,
  containers?: pb.types.Container[]
) => ({
  type: APPEND_CREATE_PROJECT_LOG,
  data: {
    id,
    output: {
      type,
      log,
      containers,
    },
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

export const setNamespaceReload = (reload: boolean, nsID: number) => ({
  type: SET_NAMESPACE_RELOAD,
  data: {
    reload: reload,
    nsID: nsID,
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
export const setStart = (id: string, start: boolean) => ({
  type: SET_TIMER_START,
  data: {
    id: id,
    start: start,
  },
});
export const setStartAt = (id: string, startAt: number) => ({
  type: SET_TIMER_START_AT,
  data: {
    id: id,
    startAt: startAt,
  },
});

export const setClusterInfo = (info: pb.cluster.InfoResponse) => ({
  type: SET_CLUSTER_INFO,
  info: info,
});
export const setPodEventPID = (pid: number) => ({
  type: PROJECT_POD_EVENT,
  projectIDWithTimestamp: `${new Date().getTime()}-${pid}`,
});

const debounceLoadNamespace = debounce((dispatch: Dispatch, nsID: number) => {
  dispatch(setNamespaceReload(true, nsID));
}, 500);

export const handleEvents = (
  id: string,
  data: pb.websocket.Metadata,
  input: any
) => {
  return function (dispatch: Dispatch) {
    switch (data.type.valueOf()) {
      case pb.websocket.Type.ProjectPodEvent:
        let nsEvent = pb.websocket.WsProjectPodEventResponse.decode(input);
        dispatch(setPodEventPID(nsEvent.project_id));
        break;
      case pb.websocket.Type.SetUid:
        setUid(data.message);
        break;
      case pb.websocket.Type.ClusterInfoSync:
        let info = pb.websocket.WsHandleClusterResponse.decode(input);
        info.info && dispatch(setClusterInfo(info.info));
        break;
      case pb.websocket.Type.ReloadProjects:
        let nsReload = pb.websocket.WsReloadProjectsResponse.decode(input);
        debounceLoadNamespace(dispatch, nsReload.namespace_id);
        break;
      case pb.websocket.Type.UpdateProject:
        let containers: pb.types.Container[] = [];
        if (data.result === pb.websocket.ResultType.LogWithContainers) {
          containers =
            pb.websocket.WsWithContainerMessageResponse.decode(
              input
            ).containers;
        }

        dispatch(
          appendCreateProjectLog(id, data.message, data.result, containers)
        );

        if (data.end) {
          switch (data.result) {
            case pb.websocket.ResultType.Deployed:
              dispatch(setDeployStatus(id, DeployStatus.DeploySuccess));
              message.success("部署成功");
              dispatch(clearCreateProjectLog(id));
              break;
            case pb.websocket.ResultType.DeployedCanceled:
              dispatch(setDeployStatus(id, DeployStatus.DeployCanceled));
              dispatch(
                appendCreateProjectLog(
                  id,
                  "部署已取消",
                  data.result,
                  containers
                )
              );
              message.warn("部署已取消");
              break;
            case pb.websocket.ResultType.DeployedFailed:
            default:
              dispatch(setDeployStatus(id, DeployStatus.DeployFailed));
              dispatch(
                appendCreateProjectLog(id, "部署失败", data.result, containers)
              );
              message.error("部署失败");
              break;
          }
          dispatch(setCreateProjectLoading(id, false));
        }
        break;
      case pb.websocket.Type.CreateProject:
        let createContainers: pb.types.Container[] = [];
        if (data.result === pb.websocket.ResultType.LogWithContainers) {
          createContainers =
            pb.websocket.WsWithContainerMessageResponse.decode(
              input
            ).containers;
        }
        dispatch(
          appendCreateProjectLog(
            id,
            data.message,
            data.result,
            createContainers
          )
        );

        if (data.end) {
          switch (data.result) {
            case pb.websocket.ResultType.Deployed:
              dispatch(setDeployStatus(id, DeployStatus.DeploySuccess));
              message.success("部署成功");
              dispatch(clearCreateProjectLog(id));
              break;
            case pb.websocket.ResultType.DeployedCanceled:
              dispatch(setDeployStatus(id, DeployStatus.DeployCanceled));
              dispatch(
                appendCreateProjectLog(
                  id,
                  "部署已取消",
                  data.result,
                  createContainers
                )
              );
              message.warn("部署已取消");
              break;
            case pb.websocket.ResultType.DeployedFailed:
            default:
              dispatch(setDeployStatus(id, DeployStatus.DeployFailed));
              dispatch(
                appendCreateProjectLog(
                  id,
                  "部署失败",
                  data.result,
                  createContainers
                )
              );
              message.error("部署失败");
              break;
          }
          dispatch(setCreateProjectLoading(id, false));
          debounceLoadNamespace(dispatch, 0);
        }
        break;
      case pb.websocket.Type.ProcessPercent:
        dispatch(setProcessPercent(id, data.percent));
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
        if (data.result === pb.websocket.ResultType.Success) {
          message.info(data.message);
        }
        break;
      default:
        console.log("unknown event: ", data.type);
    }
  };
};
