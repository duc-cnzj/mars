import { WsResponse } from "./../App";
import {
  SET_CREATE_PROJECT_LOADING,
  APPEND_CREATE_PROJECT_LOG,
  SET_DEPLOY_STATUS,
  CLEAR_CREATE_PROJECT_LOG,
  SET_NAMESPACE_RELOAD,
  SET_PROCESS_PERCENT,
  SET_CLUSTER_INFO,
} from "./actionTypes";
import { DeployStatus } from "./reducers/createProject";
import { Dispatch } from "redux";
import { message } from "antd";
import { setUid } from "../utils/uid";

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

export const setClusterInfo = (info: API.ClusterInfo) => ({
  type: SET_CLUSTER_INFO,
  info: info,
});

export const handleEvents = (id: string, data: WsResponse) => {
  return function (dispatch: Dispatch) {
    switch (data.type) {
      case "set_uid":
        setUid(data.data);
        break;
      case "reload_projects":
        dispatch(setNamespaceReload(true));
        break;
      case "update_project":
        dispatch(appendCreateProjectLog(id, data.data ? data.data : ""));

        if (data.end) {
          switch (data.result) {
            case "deployed":
              dispatch(setDeployStatus(id, DeployStatus.DeployUpdateSuccess));
              message.success("部署成功");
              dispatch(clearCreateProjectLog(id));
              break;
            case "deployed_canceled":
              dispatch(setDeployStatus(id, DeployStatus.DeployCanceled));
              dispatch(appendCreateProjectLog(id, "部署已取消"));
              message.error("部署已取消");
              break;
            case "deployed_failed":
            default:
              dispatch(setDeployStatus(id, DeployStatus.DeployFailed));
              dispatch(appendCreateProjectLog(id, "部署失败"));
              message.error("部署失败");
              break;
          }
          dispatch(setCreateProjectLoading(id, false));
        }
        break;
      case "create_project":
        dispatch(appendCreateProjectLog(id, data.data ? data.data : ""));

        if (data.end) {
          switch (data.result) {
            case "deployed":
              dispatch(setDeployStatus(id, DeployStatus.DeploySuccess));
              message.success("部署成功");
              dispatch(clearCreateProjectLog(id));
              break;
            case "deployed_canceled":
              dispatch(setDeployStatus(id, DeployStatus.DeployCanceled));
              dispatch(appendCreateProjectLog(id, "部署已取消"));
              message.error("部署已取消");
              break;
            case "deployed_failed":
            default:
              dispatch(setDeployStatus(id, DeployStatus.DeployFailed));
              dispatch(appendCreateProjectLog(id, "部署失败"));
              message.error("部署失败");
              break;
          }
          dispatch(setCreateProjectLoading(id, false));
          setTimeout(() => {
            dispatch(setNamespaceReload(true));
          }, 1000);
        }
        break;
      case "process_percent":
        dispatch(setProcessPercent(id, Number(data.data)));
        break;
      default:
        console.log("unknown event: ", data.type);
    }
  };
};
