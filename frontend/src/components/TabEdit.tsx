import React, { useMemo, useState, useEffect, useCallback, memo } from "react";
import { MyCodeMirror as CodeMirror, getMode } from "./MyCodeMirror";
import Elements from "./elements/Elements";
import PipelineInfo from "./PipelineInfo";
import ConfigHistory from "./ConfigHistory";
import {
  DeployStatus as DeployStatusEnum,
  selectList,
} from "../store/reducers/createProject";
import { Button, Row, Col, Form, Progress } from "antd";
import { useSelector, useDispatch } from "react-redux";
import {
  clearCreateProjectLog,
  setCreateProjectLoading,
  setDeployStatus,
  setStart as dispatchSetStart,
  setStartAt as dispatchSetStartAt,
} from "../store/actions";
import { toSlug } from "../utils/slug";
import { useWs, useWsReady } from "../contexts/useWebsocket";
import { StopOutlined } from "@ant-design/icons";
import classNames from "classnames";
import LogOutput from "./LogOutput";
import ProjectSelector from "./ProjectSelector";
import DebugModeSwitch from "./DebugModeSwitch";
import pb from "../api/compiled";
import TimeCost from "./TimeCost";
import { selectTimer } from "../store/reducers/deployTimer";
import DiffViewer from "./DiffViewer";

interface WatchData {
  gitProjectId: number;
  gitBranch: string;
  gitCommit: string;
  config: string;
}

const ModalSub: React.FC<{
  namespaceId: number;
  detail: pb.types.ProjectModel;
  onSuccess: () => void;
  elements: pb.mars.Element[];
  updatedAt: any;
}> = ({ detail, onSuccess, updatedAt, namespaceId, elements }) => {
  const ws = useWs();
  const wsReady = useWsReady();
  const [form] = Form.useForm();
  const list = useSelector(selectList);
  const dispatch = useDispatch();

  let slug = useMemo(
    () => toSlug(namespaceId, detail.name),
    [namespaceId, detail.name]
  );
  const isLoading = useMemo(() => list[slug]?.isLoading ?? false, [list, slug]);
  const deployStatus = useMemo(() => list[slug]?.deployStatus, [list, slug]);
  const processPercent = useMemo(
    () => list[slug]?.processPercent,
    [list, slug]
  );

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

  const [showLog, setShowLog] = useState(start);

  const [data, setData] = useState<WatchData>({
    gitProjectId: Number(detail.git_project_id),
    gitBranch: detail.git_branch,
    gitCommit: detail.git_commit,
    config: detail.config,
  });

  const formInitData = useMemo(
    () => ({
      selectors: {
        projectName: detail.name,
        gitProjectId: Number(detail.git_project_id),
        gitBranch: detail.git_branch,
        gitCommit: detail.git_commit,
      },
      name: detail.name,
      gitProjectId: Number(detail.git_project_id),
      gitBranch: detail.git_branch,
      gitCommit: detail.git_commit,
      config: detail.config,
      config_type: detail.config_type,
      debug: !detail.atomic,
      extra_values: detail.extra_values,
    }),
    [detail]
  );

  const onChange = useCallback(
    (v: any) => {
      setData((d) => ({ ...d, ...v }));
      form.setFieldsValue({ selectors: v });
    },
    [form]
  );
  const updateDeploy = (values: any) => {
    console.log(values);
    if (!wsReady) {
      // message.error("连接断开了");
      return;
    }
    if (values.extra_values) {
      values.extra_values = values.extra_values.map((i: any) => ({
        ...i,
        value: String(i.value),
      }));
    }
    if (values.selectors.gitCommit && values.selectors.gitBranch) {
      let s = pb.websocket.UpdateProjectInput.encode({
        type: pb.websocket.Type.UpdateProject,

        extra_values: values.extra_values,
        project_id: Number(detail.id),
        git_branch: values.selectors.gitBranch,
        git_commit: values.selectors.gitCommit,
        config: values.config,
        atomic: !values.debug,
      }).finish();
      dispatch(setDeployStatus(slug, DeployStatusEnum.DeployUnknown));

      dispatch(clearCreateProjectLog(slug));
      dispatch(setCreateProjectLoading(slug, true));
      setShowLog(true);
      setStart(true);
      setStartAt(Date.now());
      ws?.send(s);
    }
  };

  const resetTimeCost = useCallback(() => {
    setStart(false);
    setStartAt(0);
  }, [setStartAt, setStart]);

  const onCancel = useCallback(() => {
    if (!wsReady) {
      // message.error("连接断开了");
      return;
    }
    if (Number(namespaceId) > 0 && detail.name.length > 0) {
      let s = pb.websocket.CancelInput.encode({
        type: pb.websocket.Type.CancelProject,
        namespace_id: Number(namespaceId),
        name: detail.name,
      }).finish();
      ws?.send(s);
    }
  }, [ws, namespaceId, wsReady, detail.name]);

  const onReset = useCallback(() => {
    setShowLog(false);
    form.resetFields();
    setData(formInitData);
  }, [form, formInitData]);

  // 更新成功，触发 onSuccess
  useEffect(() => {
    if (deployStatus !== DeployStatusEnum.DeployUnknown) {
      resetTimeCost();
    }
    if (deployStatus === DeployStatusEnum.DeploySuccess) {
      resetTimeCost();
      dispatch(setDeployStatus(slug, DeployStatusEnum.DeployUnknown));
      setShowLog(false);
      onSuccess();
    }
  }, [deployStatus, dispatch, slug, onSuccess, resetTimeCost]);

  useEffect(() => {
    if (!wsReady) {
      resetTimeCost();
      dispatch(setCreateProjectLoading(slug, false));
    }
  }, [wsReady, dispatch, slug, resetTimeCost]);

  return (
    <div className="edit-project">
      <Form
        style={{ height: "100%" }}
        initialValues={formInitData}
        layout="horizontal"
        form={form}
        labelWrap
        autoComplete="off"
        onFinish={updateDeploy}
      >
        <div
          style={{ height: "100%", display: "flex", flexDirection: "column" }}
        >
          <PipelineInfo
            projectId={data.gitProjectId}
            branch={data.gitBranch}
            commit={data.gitCommit}
          />

          <Form.Item
            name="selectors"
            style={{ width: "100%", marginBottom: 10 }}
            rules={[{ required: true, message: "项目必选" }]}
          >
            <ProjectSelector
              disabled={isLoading}
              isCreate={false}
              onChange={onChange}
            />
          </Form.Item>

          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              alignItems: "center",
            }}
          >
            <div className={classNames("edit-project__footer")}>
              <Button
                htmlType="submit"
                style={{ fontSize: 12, marginRight: 5 }}
                size="small"
                type="primary"
                loading={isLoading}
              >
                部署
              </Button>

              <Button
                size="small"
                hidden={isLoading}
                style={{ marginRight: 5, fontSize: 12 }}
                disabled={isLoading}
                onClick={onReset}
              >
                重置
              </Button>
              <Button
                style={{ fontSize: 12, marginRight: 5 }}
                size="small"
                hidden={!isLoading}
                danger
                icon={<StopOutlined />}
                type="dashed"
                onClick={onCancel}
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

              {!isLoading && (
                <ConfigHistory
                  onDataChange={(s: string) => {
                    form.setFieldsValue({ config: s });
                    setData((d) => ({ ...d, config: s }));
                  }}
                  projectID={detail.id}
                  configType={detail.config_type}
                  currentConfig={data.config}
                  updatedAt={detail.updated_at}
                />
              )}
            </div>
            <Form.Item noStyle name={"debug"}>
              <DebugModeSwitch disabled={isLoading} />
            </Form.Item>
          </div>
          <div style={{ marginTop: 10, display: showLog ? "block" : "none" }}>
            <Progress
              percent={processPercent}
              strokeColor={{
                from: "#108ee9",
                to: "#87d068",
              }}
              style={{ padding: "0 3px", marginBottom: 5 }}
              status="active"
            />
            <LogOutput
              pending={<TimeCost start={start} startAt={startAt} />}
              slug={slug}
            />
          </div>
          <div
            style={{
              display: !showLog ? "flex" : "none",
              minWidth: 200,
              marginBottom: 20,
              height: "100%",
              flexDirection: "column",
            }}
          >
            <Form.Item name="extra_values" noStyle>
              <Elements
                elements={elements}
                style={{
                  inputNumber: { fontSize: 10, width: "100%" },
                  input: { fontSize: 10 },
                  label: { fontSize: 10 },
                  textarea: { fontSize: 10 },
                  formItem: {
                    marginBottom: 0,
                    marginTop: 0,
                    display: "inline-block",
                    width: "calc(33.3% - 8px)",
                    marginRight: 8,
                  },
                }}
              />
            </Form.Item>
            <Row style={{ height: "100%", marginTop: 3 }}>
              <Col span={detail.config === data.config ? 24 : 12}>
                <Form.Item name={"config"} noStyle>
                  <CodeMirror
                    mode={getMode(detail.config_type)}
                    onChange={(v) => {
                      form.setFieldsValue({ config: v });
                      setData((d) => {
                        return { ...d, config: v };
                      });
                    }}
                  />
                </Form.Item>
              </Col>
              <Col
                className="diff-viewer"
                span={detail.config === data.config ? 0 : 12}
                style={{ fontSize: 13 }}
              >
                <DiffViewer
                  styles={{
                    gutter: { padding: "0 5px", minWidth: 25 },
                    marker: { padding: "0 6px" },
                    diffContainer: {
                      height: "100%",
                      display: "block",
                      width: "100%",
                      overflowX: "auto",
                    },
                  }}
                  mode={detail.config_type}
                  showDiffOnly={false}
                  oldValue={detail.config}
                  newValue={data.config}
                  splitView={false}
                />
              </Col>
            </Row>
          </div>
        </div>
      </Form>
    </div>
  );
};

export default memo(ModalSub);
