import { useCallback, useMemo } from "react";
import { websocket } from "../api/websocket";
import {
  selectList,
  DeployStatus as DeployStatusEnum,
} from "../store/reducers/createProject";
import { useSelector } from "react-redux";
import { useWs } from "./useWebsocket";
import {
  clearCreateProjectLog,
  setCreateProjectLoading,
  setDeployStatus,
  cleanProject,
} from "../store/actions";
import { useDispatch } from "react-redux";

interface DeployProps {
  slug: string;
  namespaceID: number;
}
interface createProjectProps {
  repoId: number;
  branch: string;
  commit: string;
  extraValues: websocket.ExtraValue[];
  config: string;
}
interface updateProjectProps {
  projectId: number;
  version: number;
  branch: string;
  commit: string;
  extraValues: websocket.ExtraValue[];
  config: string;
}

const defaultState = {
  isLoading: false,
  deployStatus: DeployStatusEnum.DeployUnknown,
  output: [],
  processPercent: 0,
};

export default function useDeploy({ namespaceID, slug }: DeployProps) {
  const ws = useWs();
  const list = useSelector(selectList);
  const state = useMemo(() => {
    if (slug && list[slug]) {
      return list[slug];
    }
    return defaultState;
  }, [list, slug]);
  const dispatch = useDispatch();
  const deployStatus = state.deployStatus;
  const processPercent = state.processPercent;
  const isLoading = state.isLoading;

  const createProject = useCallback(
    ({ repoId, branch, commit, extraValues, config }: createProjectProps) => {
      let createParams = websocket.CreateProjectInput.encode({
        type: websocket.Type.CreateProject,
        namespaceId: namespaceID,
        repoId: repoId,
        gitBranch: branch,
        gitCommit: commit,
        extraValues: extraValues,
        config: config,
      }).finish();
      dispatch(setDeployStatus(slug, DeployStatusEnum.DeployUnknown));
      dispatch(clearCreateProjectLog(slug));
      dispatch(setCreateProjectLoading(slug, true));
      //   dispatch(dispatchSetStart(slug, true));

      ws?.send(createParams);
    },
    [ws, namespaceID, dispatch, slug],
  );

  const updateProject = useCallback(
    ({
      projectId,
      branch,
      commit,
      extraValues,
      config,
      version,
    }: updateProjectProps) => {
      let createParams = websocket.UpdateProjectInput.encode({
        type: websocket.Type.UpdateProject,
        projectId: projectId,
        gitBranch: branch,
        gitCommit: commit,
        extraValues: extraValues,
        config: config,
        version: version,
      }).finish();
      dispatch(setDeployStatus(slug, DeployStatusEnum.DeployUnknown));
      dispatch(clearCreateProjectLog(slug));
      dispatch(setCreateProjectLoading(slug, true));
      //   dispatch(dispatchSetStart(slug, true));

      ws?.send(createParams);
    },
    [ws, dispatch, slug],
  );

  const cancelDeploy = useCallback(
    (name: string) => {
      let s = websocket.CancelInput.encode({
        type: websocket.Type.CancelProject,
        namespaceId: namespaceID,
        name: name,
      }).finish();
      ws?.send(s);
    },
    [ws, namespaceID],
  );

  const clearProject = useCallback(() => {
    dispatch(cleanProject(slug));
  }, [dispatch, slug]);

  const isSuccess = state.deployStatus === DeployStatusEnum.DeploySuccess;
  const hasLog = state.output?.length > 0;
  if (!slug) {
    return defaultReturn;
  }

  return {
    isSuccess,
    processPercent,
    deployStatus,
    hasLog,
    isLoading,
    createProject,
    updateProject,
    cancelDeploy,
    clearProject,
  };
}

const defaultReturn = {
  processPercent: 0,
  isSuccess: false,
  deployStatus: DeployStatusEnum.DeployUnknown,
  hasLog: false,
  isLoading: false,
  updateProject: (v: updateProjectProps) => {},
  createProject: (v: createProjectProps) => {},
  cancelDeploy: (name: string) => {},
  clearProject: () => {},
};
