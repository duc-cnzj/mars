import React, { useState, useCallback, useEffect, useMemo, memo } from "react";
import { selectClusterInfo } from "../store/reducers/cluster";
import Elements from "./elements/Elements";
import { DraggableModal } from "../pkg/DraggableModal/DraggableModal";
import { MyCodeMirror as CodeMirror, getMode } from "./MyCodeMirror";
import pb from "../api/websocket";
import { useAsyncState } from "../utils/async";
import { selectTimer } from "../store/reducers/deployTimer";
import {
  DeployStatus as DeployStatusEnum,
  selectList,
} from "../store/reducers/createProject";
import { useWs, useWsReady } from "../contexts/useWebsocket";
import { message, Button, Form, Progress, Affix } from "antd";
import { PlusOutlined, StopOutlined } from "@ant-design/icons";
import { useDispatch, useSelector } from "react-redux";
import {
  clearCreateProjectLog,
  setCreateProjectLoading,
  setDeployStatus,
  setStart as dispatchSetStart,
  setStartAt as dispatchSetStartAt,
} from "../store/actions";
import { toSlug } from "../utils/slug";
import LogOutput from "./LogOutput";
import ProjectSelector from "./ProjectSelectorV2";
import DebugModeSwitch from "./DebugModeSwitch";
import TimeCost from "./TimeCost";
import { css } from "@emotion/css";
import ajax from "../api/ajax";
import { components } from "../api/schema";

const initFormValues = {
  debug: true,
  extra_values: [],
  config: "",
};

const CreateProjectModal: React.FC<{
  namespaceId: number;
}> = ({ namespaceId }) => {
  const ws = useWs();
  const wsReady = useWsReady();
  const [form] = Form.useForm();
  const list = useSelector(selectList);
  const dispatch = useDispatch();
  const [data, setData] = useState<{
    projectName: string;
    gitProjectId: number;
    gitBranch: string;
    gitCommit: string;
  }>();
  let slug = useMemo(
    () => toSlug(namespaceId, data?.projectName ? data.projectName : ""),
    [namespaceId, data]
  );

  const [mode, setMode] = useState<string>("text/x-yaml");
  const [visible, setVisible] = useAsyncState<boolean>(false);
  const info = useSelector(selectClusterInfo);

  const [elements, setElements] = useState<
    components["schemas"]["mars.Element"][]
  >([]);

  const timer = useSelector(selectTimer);
  const start = useMemo(() => timer[slug]?.start || false, [timer, slug]);
  const startAt = useMemo(() => timer[slug]?.startAt || 0, [timer, slug]);
  const setStart = useCallback(
    (start: boolean) => {
      dispatch(dispatchSetStart(slug, start));
    },
    [dispatch, slug]
  );

  const setStartAt = useCallback(
    (startAt: number) => {
      dispatch(dispatchSetStartAt(slug, startAt));
    },
    [dispatch, slug]
  );

  const [deployStarted, setDeployStarted] = useState(false);
  const [showLog, setShowLog] = useState(start);

  const isLoading = useMemo(() => list[slug]?.isLoading ?? false, [list, slug]);
  const deployStatus = useMemo(() => list[slug]?.deployStatus, [list, slug]);
  const processPercent = useMemo(
    () => list[slug]?.processPercent,
    [list, slug]
  );

  const resetTimeCost = useCallback(() => {
    setStart(false);
    setStartAt(0);
  }, [setStartAt, setStart]);

  const onCancel = useCallback(() => {
    setElements([]);
    form.resetFields();
    resetTimeCost();
    setShowLog(false);
    setData(undefined);
    setVisible(false);
    dispatch(clearCreateProjectLog(slug));
  }, [dispatch, slug, form, setVisible, resetTimeCost]);

  const deploy = useCallback(
    (values: any) => {
      if (!wsReady) {
        // message.error("连接断开了");
        return;
      }
      if (values.extra_values) {
        let defaults = elements
          ? elements.map((i) => ({
              path: i.path,
              value: i.default,
            }))
          : [];
        values.extra_values =
          values.extra_values.length > 0
            ? values.extra_values.map((i: any) => ({
                ...i,
                value: String(i.value),
              }))
            : defaults;
      }
      if (data && data.gitProjectId && data.gitBranch && data.gitCommit) {
        // todo ws connected!
        let s = pb.websocket.CreateProjectInput.encode({
          type: pb.websocket.Type.CreateProject,
          namespaceId: Number(namespaceId),
          name: data.projectName,
          // TODO
          repoId: 0,
          // git_project_id: Number(data.gitProjectId),
          gitBranch: data.gitBranch,
          gitCommit: data.gitCommit,
          config: values.config,
          atomic: !values.debug,
          extraValues: values.extra_values,
        }).finish();

        dispatch(setDeployStatus(slug, DeployStatusEnum.DeployUnknown));

        dispatch(clearCreateProjectLog(slug));
        dispatch(setCreateProjectLoading(slug, true));
        setShowLog(true);
        setStart(true);
        setStartAt(Date.now());
        setDeployStarted(true);
        ws?.send(s);
        return;
      }

      message.error("项目id, 分支，提交必填");
    },
    [
      dispatch,
      setStart,
      setStartAt,
      slug,
      ws,
      namespaceId,
      wsReady,
      data,
      elements,
    ]
  );

  const onRemove = useCallback(() => {
    if (!wsReady) {
      // message.error("连接断开了");
      return;
    }
    if (data && data.gitProjectId && data.gitBranch && data.gitCommit) {
      let s = pb.websocket.CancelInput.encode({
        type: pb.websocket.Type.CancelProject,
        namespaceId: namespaceId,
        name: data.projectName,
      }).finish();
      ws?.send(s);
      return;
    }
  }, [wsReady, ws, namespaceId, data]);

  const loadConfigFile = useCallback(
    (gitProjectId: string, gitBranch: string) => {
      ajax
        .GET("/api/git/projects/{gitProjectId}/branches/{branch}/config_file", {
          params: {
            path: {
              gitProjectId: gitProjectId,
              branch: gitBranch,
            },
          },
        })
        .then(({ data }) => {
          if (data) {
            if (!form.getFieldValue("config")) {
              form.setFieldsValue({ config: data.data });
            }
            setMode(getMode(data.type));
            setElements(data.elements);
          }
        });
    },
    [form]
  );

  const onChange = useCallback(
    (v: any) => {
      setData(v);
      form.setFieldsValue({ selectors: v });
      if (v.gitCommit !== "") {
        loadConfigFile(String(v.gitProjectId), v.gitBranch);
      }
    },
    [form, loadConfigFile]
  );

  useEffect(() => {
    if (!deployStarted) {
      return;
    }
    if (deployStatus !== DeployStatusEnum.DeployUnknown) {
      resetTimeCost();
    }
    if (deployStatus === DeployStatusEnum.DeploySuccess) {
      resetTimeCost();
      setData(undefined);
      setElements([]);
      form.resetFields();
      dispatch(setDeployStatus(slug, DeployStatusEnum.DeployUnknown));
      setShowLog(false);
      setTimeout(() => {
        setVisible(false);
      }, 500);
    }
  }, [
    list,
    dispatch,
    deployStarted,
    slug,
    onCancel,
    form,
    setVisible,
    deployStatus,
    resetTimeCost,
  ]);

  useEffect(() => {
    if (!wsReady) {
      resetTimeCost();
      dispatch(setCreateProjectLoading(slug, false));
    }
  }, [wsReady, dispatch, slug, resetTimeCost]);
  const [container, setContainer] = useState<HTMLDivElement | null>(null);

  return (
    <div>
      <Button
        onClick={() => setVisible(true)}
        style={{ width: "100%" }}
        type="dashed"
        icon={<PlusOutlined />}
      ></Button>
      <DraggableModal
        destroyOnClose
        open={visible}
        okButtonProps={{
          loading: isLoading,
          danger: info.status === "bad",
        }}
        onCancel={onCancel}
        cancelButtonProps={{ disabled: isLoading }}
        closable={!isLoading}
        footer={null}
        initialWidth={900}
        initialHeight={600}
        title={<div style={{ textAlign: "center" }}>创建项目</div>}
        className="draggable-modal"
      >
        <ProjectSelector namespaceId={namespaceId} />
      </DraggableModal>
    </div>
  );
};

export default memo(CreateProjectModal);
