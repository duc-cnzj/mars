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
  REMOVE_SHELL,
  SET_OPENED_MODALS,
  CLEAN_PROJECT,
} from "./actionTypes";
import { DeployStatus } from "./reducers/createProject";
import { Dispatch } from "redux";
import { message } from "antd";
import { setUid } from "../utils/uid";
import pb from "../api/websocket";
import { debounce } from "lodash";
import { components } from "../api/schema";

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
  containers?: pb.websocket.Container[],
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
export const cleanProject = (id: string) => ({
  type: CLEAN_PROJECT,
  data: {
    id,
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
export const setShellSessionId = (id: string) => ({
  type: SET_SHELL_SESSION_ID,
  data: {
    id: id,
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

export const setClusterInfo = (
  info: components["schemas"]["websocket.ClusterInfo"],
) => ({
  type: SET_CLUSTER_INFO,
  info: info,
});

export const removeShell = (id: string) => ({
  type: REMOVE_SHELL,
  data: {
    id: id,
  },
});

export const setPodEventPID = (pid: number) => ({
  type: PROJECT_POD_EVENT,
  projectIDWithTimestamp: `${new Date().getTime()}-${pid}`,
});

export const setOpenedModals = (modals: { [key: number]: boolean }) => ({
  type: SET_OPENED_MODALS,
  data: {
    modals,
  },
});

const debounceLoadNamespace = debounce((dispatch: Dispatch, nsID: number) => {
  dispatch(setNamespaceReload(true, nsID));
}, 500);

export const handleEvents = (
  id: string,
  data: pb.websocket.Metadata,
  input: any,
) => {
  return function (dispatch: Dispatch) {
    console.log(data);
    switch (data.type.valueOf()) {
      case pb.websocket.Type.ProjectPodEvent:
        let nsEvent = pb.websocket.WsProjectPodEventResponse.decode(input);
        dispatch(setPodEventPID(nsEvent.projectId));
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
        debounceLoadNamespace(dispatch, nsReload.namespaceId);
        break;
      case pb.websocket.Type.UpdateProject:
        let containers: pb.websocket.Container[] = [];
        if (data.result === pb.websocket.ResultType.LogWithContainers) {
          containers =
            pb.websocket.WsWithContainerMessageResponse.decode(
              input,
            ).containers;
        }

        dispatch(
          appendCreateProjectLog(id, data.message, data.result, containers),
        );

        if (data.end) {
          switch (data.result) {
            case pb.websocket.ResultType.Deployed:
              dispatch(setDeployStatus(id, DeployStatus.DeploySuccess));
              dispatch(clearCreateProjectLog(id));
              break;
            case pb.websocket.ResultType.DeployedCanceled:
              dispatch(setDeployStatus(id, DeployStatus.DeployCanceled));
              dispatch(
                appendCreateProjectLog(
                  id,
                  "部署已取消",
                  data.result,
                  containers,
                ),
              );
              message.warning("部署已取消");
              break;
            case pb.websocket.ResultType.DeployedFailed:
            default:
              dispatch(setDeployStatus(id, DeployStatus.DeployFailed));
              dispatch(
                appendCreateProjectLog(id, "部署失败", data.result, containers),
              );
              message.error("部署失败");
              break;
          }
          dispatch(setCreateProjectLoading(id, false));
        }
        break;
      case pb.websocket.Type.CreateProject:
        let createContainers: pb.websocket.Container[] = [];
        if (data.result === pb.websocket.ResultType.LogWithContainers) {
          createContainers =
            pb.websocket.WsWithContainerMessageResponse.decode(
              input,
            ).containers;
        }
        dispatch(
          appendCreateProjectLog(
            id,
            data.message,
            data.result,
            createContainers,
          ),
        );

        if (data.end) {
          dispatch(setCreateProjectLoading(id, false));
          switch (data.result) {
            case pb.websocket.ResultType.Deployed:
              dispatch(setProcessPercent(id, 100));
              dispatch(setDeployStatus(id, DeployStatus.DeploySuccess));
              setTimeout(() => {
                dispatch(clearCreateProjectLog(id));
              }, 1000);
              break;
            case pb.websocket.ResultType.DeployedCanceled:
              dispatch(setDeployStatus(id, DeployStatus.DeployCanceled));
              dispatch(
                appendCreateProjectLog(
                  id,
                  "部署已取消",
                  data.result,
                  createContainers,
                ),
              );
              message.warning("部署已取消");
              break;
            case pb.websocket.ResultType.DeployedFailed:
            default:
              dispatch(setDeployStatus(id, DeployStatus.DeployFailed));
              dispatch(
                appendCreateProjectLog(
                  id,
                  "部署失败",
                  data.result,
                  createContainers,
                ),
              );
              message.error("部署失败");
              break;
          }
        }
        break;
      case pb.websocket.Type.ProcessPercent:
        dispatch(setProcessPercent(id, data.percent));
        break;
      case pb.websocket.Type.HandleCloseShell:
      case pb.websocket.Type.HandleExecShell:
        if (data.result === pb.websocket.ResultType.Error) {
          message.error(data.message);
          break;
        }
        let res = pb.websocket.WsHandleShellResponse.decode(input);

        if (res.container && res.terminalMessage) {
          dispatch(setShellSessionId(res.terminalMessage.sessionId));
          if (data.type.valueOf() === pb.websocket.Type.HandleCloseShell) {
            dispatch(removeShell(res.terminalMessage.sessionId));
          }
        }
        break;
      case pb.websocket.Type.HandleExecShellMsg:
        let logRes = pb.websocket.WsHandleShellResponse.decode(input);
        logRes.container &&
          logRes.terminalMessage &&
          dispatch(
            setShellLog(
              logRes.terminalMessage.sessionId,
              logRes.terminalMessage,
            ),
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
