import React, { useState, useCallback, useEffect, useMemo, memo } from "react";
import { selectClusterInfo } from "../store/reducers/cluster";
import PipelineInfo from "./PipelineInfo";
import Elements from "./elements/Elements";
import { DraggableModal } from "../pkg/DraggableModal/DraggableModal";
import { MyCodeMirror as CodeMirror, getMode } from "./MyCodeMirror";
import pb from "../api/compiled";
import { orderBy } from "lodash";
import { useAsyncState } from "../utils/async";

import { configFile } from "../api/git";
import {
  DeployStatus as DeployStatusEnum,
  selectList,
} from "../store/reducers/createProject";
import { useWs, useWsReady } from "../contexts/useWebsocket";
import { message, Progress, Button, Form } from "antd";
import { PlusOutlined, StopOutlined } from "@ant-design/icons";
import { useDispatch, useSelector } from "react-redux";
import {
  clearCreateProjectLog,
  setCreateProjectLoading,
  setDeployStatus,
} from "../store/actions";
import classNames from "classnames";
import { toSlug } from "../utils/slug";
import LogOutput from "./LogOutput";
import ProjectSelector from "./ProjectSelector";
import DebugModeSwitch from "./DebugModeSwitch";
import TimeCost from "./TimeCost";

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

  const [elements, setElements] = useState<pb.mars.Element[]>();

  const [start, setStart] = useState(false);
  const [startAt, setStartAt] = useState(0);
  const [showLog, setShowLog] = useState(false);

  const isLoading = useMemo(() => list[slug]?.isLoading ?? false, [list, slug]);
  const deployStatus = useMemo(() => list[slug]?.deployStatus, [list, slug]);
  const processPercent = useMemo(
    () => list[slug]?.processPercent,
    [list, slug]
  );

  const resetTimeCost = useCallback(() => {
    setStart(false);
    setStartAt(0);
  }, []);

  const onCancel = useCallback(() => {
    setElements([]);
    form.resetFields();
    resetTimeCost();
    setData(undefined);
    setVisible(false);
    dispatch(clearCreateProjectLog(slug));
  }, [dispatch, slug, form, setVisible, resetTimeCost]);

  const deploy = useCallback(
    (values: any) => {
      if (!wsReady) {
        message.error("连接断开了");
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
          namespace_id: Number(namespaceId),
          name: data.projectName,
          git_project_id: Number(data.gitProjectId),
          git_branch: data.gitBranch,
          git_commit: data.gitCommit,
          config: values.config,
          atomic: !values.debug,
          extra_values: values.extra_values,
        }).finish();

        dispatch(setDeployStatus(slug, DeployStatusEnum.DeployUnknown));

        dispatch(clearCreateProjectLog(slug));
        dispatch(setCreateProjectLoading(slug, true));
        setShowLog(true);
        setStart(true);
        setStartAt(Date.now());
        ws?.send(s);
        return;
      }

      message.error("项目id, 分支，提交必填");
    },
    [dispatch, slug, ws, namespaceId, wsReady, data, elements]
  );

  const onRemove = useCallback(() => {
    if (!wsReady) {
      message.error("连接断开了");
      return;
    }
    if (data && data.gitProjectId && data.gitBranch && data.gitCommit) {
      let s = pb.websocket.CancelInput.encode({
        type: pb.websocket.Type.CancelProject,
        namespace_id: namespaceId,
        name: data.projectName,
      }).finish();
      ws?.send(s);
      return;
    }
  }, [wsReady, ws, namespaceId, data]);

  const loadConfigFile = useCallback(
    (gitProjectId: string, gitBranch: string) => {
      configFile({
        git_project_id: gitProjectId,
        branch: gitBranch,
      }).then((res) => {
        if (!form.getFieldValue("config")) {
          form.setFieldsValue({ config: res.data.data });
        }
        setMode(getMode(res.data.type));
        setElements(res.data.elements);
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
        visible={visible}
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
        className="draggable-modal drag-item-modal"
      >
        <div
          className="create-project-modal"
          style={{ display: "flex", flexDirection: "column", height: "100%" }}
        >
          <div style={{ height: "100%" }}>
            <Form
              layout="horizontal"
              form={form}
              labelWrap
              autoComplete="off"
              onFinish={deploy}
              initialValues={initFormValues}
              style={{
                height: "100%",
                display: "flex",
                flexDirection: "column",
              }}
            >
              {data?.gitCommit && (
                <PipelineInfo
                  projectId={data.gitProjectId}
                  branch={data.gitBranch}
                  commit={data.gitCommit}
                />
              )}
              <Form.Item
                name="selectors"
                style={{ width: "100%", marginBottom: 10 }}
                rules={[{ required: true, message: "项目必选" }]}
              >
                <ProjectSelector isCreate onChange={onChange} />
              </Form.Item>

              <div
                style={{
                  display: "flex",
                  justifyContent: "space-between",
                  alignItems: "center",
                  paddingBottom: 10,
                }}
              >
                <div style={{ display: "flex" }}>
                  <Button
                    htmlType="submit"
                    style={{ fontSize: 12, marginRight: 5 }}
                    size="small"
                    danger={info.status === "bad"}
                    type={"primary"}
                    loading={isLoading}
                  >
                    {info.status === "bad" ? "集群资源不足" : "部署"}
                  </Button>
                  <Button
                    style={{ fontSize: 12, marginRight: 5 }}
                    size="small"
                    hidden={!isLoading}
                    danger
                    icon={<StopOutlined />}
                    type="dashed"
                    onClick={onRemove}
                  >
                    取消
                  </Button>
                  {list[slug]?.output?.length > 0 && (
                    <Button
                      type="dashed"
                      style={{ fontSize: 12, marginRight: 5 }}
                      size="small"
                      onClick={() => setShowLog((show) => !show)}
                    >
                      {showLog ? "隐藏" : "查看"}日志
                    </Button>
                  )}
                </div>
                <Form.Item noStyle name={"debug"}>
                  <DebugModeSwitch />
                </Form.Item>
              </div>
              <div
                style={{ marginTop: 10 }}
                className={classNames({ "display-none": !showLog })}
              >
                <Progress
                  strokeColor={{
                    from: "#108ee9",
                    to: "#87d068",
                  }}
                  style={{ padding: "0 3px", marginBottom: 5 }}
                  percent={processPercent}
                  status="active"
                />
                <LogOutput
                  pending={<TimeCost start={start} startAt={startAt} />}
                  slug={slug}
                />
              </div>
              <div
                className={classNames({ "display-none": showLog })}
                style={{
                  minWidth: 200,
                  marginBottom: 20,
                  height: "100%",
                }}
              >
                <Form.Item name="extra_values" noStyle>
                  <Elements
                    elements={orderBy(elements, ["type"], ["asc"])}
                    style={{
                      inputNumber: { fontSize: 10, width: "100%" },
                      input: { fontSize: 10 },
                      label: { fontSize: 10 },
                      formItem: {
                        marginBottom: 5,
                        display: "inline-block",
                        width: "calc(33% - 8px)",
                        marginRight: 8,
                      },
                    }}
                  />
                </Form.Item>
                <Form.Item name="config" style={{ height: "100%" }} noStyle>
                  <CodeMirror
                    options={{
                      mode: mode,
                      theme: "dracula",
                      lineNumbers: true,
                    }}
                  />
                </Form.Item>
              </div>
            </Form>
          </div>
        </div>
      </DraggableModal>
    </div>
  );
};

export default memo(CreateProjectModal);
