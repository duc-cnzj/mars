import React, { useMemo, useState, useEffect, useCallback, memo } from "react";
import { MyCodeMirror as CodeMirror, getMode } from "./MyCodeMirror";
import ReactDiffViewer from "react-diff-viewer";
import Elements from "./elements/Elements";
import PipelineInfo from "./PipelineInfo";
import ConfigHistory from "./ConfigHistory";
import { getHighlightSyntax } from "../utils/highlight";
import { orderBy } from "lodash";
import {
  DeployStatus as DeployStatusEnum,
  selectList,
} from "../store/reducers/createProject";
import { Button, Progress, message, Row, Col, Form } from "antd";
import { useSelector, useDispatch } from "react-redux";
import {
  clearCreateProjectLog,
  setCreateProjectLoading,
  setDeployStatus,
} from "../store/actions";
import { toSlug } from "../utils/slug";
import { useWs, useWsReady } from "../contexts/useWebsocket";
import {
  ArrowLeftOutlined,
  StopOutlined,
  ArrowRightOutlined,
} from "@ant-design/icons";
import classNames from "classnames";
import LogOutput from "./LogOutput";
import ProjectSelector from "./ProjectSelector";
import TimeCost from "./TimeCost";
import DebugModeSwitch from "./DebugModeSwitch";
import pb from "../api/compiled";

interface WatchData {
  projectName: string;
  gitProjectId: number;
  gitBranch: string;
  gitCommit: string;
  config: string;
}

const ModalSub: React.FC<{
  detail: pb.ProjectShowResponse;
  onSuccess: () => void;
  updatedAt: any;
}> = ({ detail, onSuccess, updatedAt }) => {
  let id = detail.id;
  let namespaceId = detail.namespace?.id;
  const ws = useWs();
  const wsReady = useWsReady();
  const [form] = Form.useForm();
  const [editVisible, setEditVisible] = useState<boolean>(true);
  const [timelineVisible, setTimelineVisible] = useState<boolean>(false);
  const list = useSelector(selectList);
  const dispatch = useDispatch();
  const [data, setData] = useState<WatchData>({
    projectName: detail.name,
    gitProjectId: Number(detail.git_project_id),
    gitBranch: detail.git_branch,
    gitCommit: detail.git_commit,
    config: detail.config,
  });
  const [start, setStart] = useState(false);

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

  let slug = useMemo(
    () => toSlug(namespaceId || 0, detail.name),
    [namespaceId, detail.name]
  );

  const onChange = useCallback(
    (v: {
      projectName: string;
      gitProjectId: number;
      gitBranch: string;
      gitCommit: string;
    }) => {
      setData((d) => ({ ...d, ...v }));
      form.setFieldsValue({ selectors: v });
    },
    [form]
  );
  const updateDeploy = (values: any) => {
    if (values.extra_values) {
      values.extra_values = values.extra_values.map((i: any) => ({
        ...i,
        value: String(i.value),
      }));
    }
    if (!wsReady) {
      message.error("连接断开了");
      return;
    }
    if (data.gitCommit && data.gitBranch) {
      setStart(true);
      setEditVisible(false);
      setTimelineVisible(true);

      let s = pb.UpdateProjectInput.encode({
        type: pb.Type.UpdateProject,

        extra_values: values.extra_values,
        project_id: Number(id),
        git_branch: data.gitBranch,
        git_commit: data.gitCommit,
        config: values.config,
        atomic: !values.debug,
      }).finish();
      dispatch(setDeployStatus(slug, DeployStatusEnum.DeployUnknown));

      dispatch(clearCreateProjectLog(slug));
      dispatch(setCreateProjectLoading(slug, true));
      ws?.send(s);
    }
  };

  const onCancel = useCallback(() => {
    if (!wsReady) {
      message.error("连接断开了");
      return;
    }
    if (data.gitProjectId && data.gitBranch && data.gitCommit) {
      let s = pb.CancelInput.encode({
        type: pb.Type.CancelProject,
        namespace_id: Number(namespaceId),
        name: data.projectName,
      }).finish();
      ws?.send(s);
      return;
    }
  }, [data, ws, namespaceId, wsReady]);

  const highlightSyntax = useCallback(
    (str: string) => (
      <code
        dangerouslySetInnerHTML={{
          __html: getHighlightSyntax(str, detail.config_type),
        }}
      />
    ),
    [detail.config_type]
  );

  const onReset = useCallback(() => {
    form.resetFields();
    setData((d) => ({ ...d, ...formInitData }));
  }, [form, formInitData]);

  // 更新成功，触发 onSuccess
  useEffect(() => {
    if (list[slug]?.deployStatus !== DeployStatusEnum.DeployUnknown) {
      setStart(false);
    }
    if (list[slug]?.deployStatus === DeployStatusEnum.DeployUpdateSuccess) {
      setStart(false);
      setTimelineVisible(false);
      setEditVisible(true);
      dispatch(setDeployStatus(slug, DeployStatusEnum.DeployUnknown));
      onSuccess();
    }
  }, [list, dispatch, slug, onSuccess]);

  useEffect(() => {
    if (!wsReady) {
      setStart(false);
      dispatch(setCreateProjectLoading(slug, false));
    }
  }, [wsReady, dispatch, slug]);

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
          className={classNames({ "display-none": !editVisible })}
          style={{ height: "100%", display: "flex", flexDirection: "column" }}
        >
          <PipelineInfo
            projectId={data.gitProjectId}
            branch={data.gitBranch}
            commit={data.gitCommit}
          />

          <div
            style={{
              width: "100%",
              display: "flex",
              alignItems: "center",
              marginBottom: 10,
            }}
          >
            {list[slug]?.output?.length > 0 && (
              <Button
                type="dashed"
                style={{ marginRight: 5 }}
                disabled={list[slug]?.isLoading}
                onClick={() => {
                  setEditVisible(false);
                  setTimelineVisible(true);
                }}
                icon={<ArrowRightOutlined />}
              />
            )}

            <Form.Item
              name="selectors"
              style={{ width: "100%", margin: 0 }}
              rules={[{ required: true, message: "项目必选" }]}
            >
              <ProjectSelector isCreate={false} onChange={onChange} />
            </Form.Item>
          </div>

          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              alignItems: "center",
            }}
          >
            <div
              className={classNames("edit-project__footer", {
                "edit-project--hidden": list[slug]?.isLoading,
              })}
            >
              <span style={{ marginRight: 5 }}>编辑配置:</span>

              <Button
                size="small"
                style={{ marginRight: 5, fontSize: 12 }}
                disabled={list[slug]?.isLoading}
                onClick={onReset}
              >
                重置
              </Button>
              <Button
                htmlType="submit"
                style={{ fontSize: 12, marginRight: 5 }}
                size="small"
                type="primary"
                loading={list[slug]?.isLoading}
              >
                部署
              </Button>
              <ConfigHistory
                show={editVisible}
                onDataChange={(s: string) => {
                  form.setFieldsValue({ config: s });
                  setData((d) => ({ ...d, config: s }));
                }}
                projectID={detail.id}
                configType={detail.config_type}
                currentConfig={data.config}
                updatedAt={detail.updated_at}
              />
            </div>
            <Form.Item noStyle name={"debug"}>
              <DebugModeSwitch />
            </Form.Item>
          </div>
          <div
            style={{
              minWidth: 200,
              marginBottom: 20,
              height: "100%",
              display: "flex",
              flexDirection: "column",
            }}
          >
            <Form.Item name="extra_values" noStyle>
              <Elements
                elements={orderBy(detail.elements, ["type"], ["asc"])}
                style={{
                  inputNumber: { fontSize: 10, width: "100%" },
                  input: { fontSize: 10 },
                  label: { fontSize: 10 },
                  formItem: {
                    marginBottom: 0,
                    marginTop: 0,
                    display: "inline-block",
                    width: "calc(30% - 8px)",
                    marginRight: 8,
                  },
                }}
              />
            </Form.Item>
            <Row style={{ height: "100%", marginTop: 3 }}>
              <Col span={detail.config === data.config ? 24 : 12}>
                <Form.Item name={"config"} noStyle>
                  <CodeMirror
                    options={{
                      mode: getMode(detail.config_type),
                      theme: "dracula",
                      lineNumbers: true,
                    }}
                    onChange={(v) => {
                      form.setFieldsValue(["config", v]);
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
                <ReactDiffViewer
                  styles={{
                    gutter: { padding: "0 5px", minWidth: 25 },
                    marker: { padding: "0 6px" },
                    diffContainer: {
                      display: "block",
                      width: "100%",
                      overflowX: "auto",
                    },
                  }}
                  useDarkTheme
                  disableWordDiff
                  renderContent={highlightSyntax}
                  showDiffOnly={false}
                  oldValue={detail.config}
                  newValue={data.config}
                  splitView={false}
                />
              </Col>
            </Row>
          </div>
        </div>
        <div
          id="preview"
          style={{ height: "100%", overflow: "auto" }}
          className={classNames("preview", {
            "display-none": !timelineVisible,
          })}
        >
          <div
            style={{ display: "flex", alignItems: "center", marginBottom: 20 }}
          >
            <Button
              type="dashed"
              disabled={list[slug]?.isLoading}
              onClick={() => {
                setEditVisible(true);
                setTimelineVisible(false);
              }}
              icon={<ArrowLeftOutlined />}
            />
            <Progress
              strokeColor={{
                from: "#108ee9",
                to: "#87d068",
              }}
              style={{ padding: "0 10px" }}
              percent={list[slug]?.processPercent}
              status="active"
            />
          </div>
          <div
            style={{ display: "flex", alignItems: "center", marginBottom: 10 }}
          >
            <TimeCost start={start} />

            <Button
              size="small"
              type="primary"
              loading={list[slug]?.isLoading}
              htmlType="submit"
              style={{ marginRight: 10, marginLeft: 10, fontSize: 12 }}
            >
              部署
            </Button>
            <Button
              style={{ fontSize: 12 }}
              size="small"
              hidden={
                list[slug]?.deployStatus === DeployStatusEnum.DeployCanceled
              }
              danger
              icon={<StopOutlined />}
              type="dashed"
              onClick={onCancel}
            >
              取消
            </Button>
          </div>
          <LogOutput slug={slug} />
        </div>
      </Form>
    </div>
  );
};

export default memo(ModalSub);
