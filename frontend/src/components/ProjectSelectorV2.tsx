import React, { useState, memo, useCallback, useEffect, useMemo } from "react";
import {
  Affix,
  Button,
  Col,
  Form,
  message,
  Progress,
  Row,
  Select,
  Space,
} from "antd";
import { css } from "@emotion/css";
import ajax from "../api/ajax";
import { components } from "../api/schema.d";
import { Grid } from "antd";
import PipelineInfo from "./PipelineInfo";
import Elements from "./elements/Elements";
import { getMode, MyCodeMirror } from "./MyCodeMirror";
import { StopOutlined } from "@ant-design/icons";
import { selectClusterInfo } from "../store/reducers/cluster";
import { useSelector } from "react-redux";
import styled from "@emotion/styled";
import { useWs, useWsReady } from "../contexts/useWebsocket";
import { useDispatch } from "react-redux";
import TimeCost from "./TimeCost";
import {
  clearCreateProjectLog,
  setCreateProjectLoading,
  setDeployStatus,
  setStart as dispatchSetStart,
  setStartAt as dispatchSetStartAt,
} from "../store/actions";
import { websocket } from "../api/websocket";
import { toSlug } from "../utils/slug";
import { DeployStatus, selectList } from "../store/reducers/createProject";
import { selectTimer } from "../store/reducers/deployTimer";
import LogOutput from "./LogOutput";

const { useBreakpoint } = Grid;

interface FormTypes {
  extraValues: components["schemas"]["websocket.ExtraValue"][];
  config: string;
  branch: string;
  commit: string;
  repoId: number;
}

const DeployProjectForm: React.FC<{
  namespaceId: number;
  projectId?: number;
  edit?: boolean;
}> = ({ namespaceId, projectId, edit }) => {
  const [form] = Form.useForm();
  const repoId = Form.useWatch("repoId", form);
  const branch = Form.useWatch("branch", form);
  const commit = Form.useWatch("commit", form);

  const [loading, setLoading] = useState({
    project: false,
    branch: false,
    commit: false,
  });
  const screens = useBreakpoint();

  const [needGitRepo, setNeedGitRepo] = useState(true);

  const [options, setOptions] = useState<{
    project: components["schemas"]["git.Option"][];
    branch: components["schemas"]["git.Option"][];
    commit: components["schemas"]["git.Option"][];
  }>({
    project: [],
    branch: [],
    commit: [],
  });

  const findProject = useCallback(
    (repoID: string | number) => {
      let found = options.project.find(
        (v) => Number(v.value) === Number(repoID)
      );
      return found;
    },
    [options.project]
  );

  const onProjectVisibleChange = useCallback(
    (open: boolean) => {
      if (!open || options.project.length > 0) {
        return;
      }
      setLoading((l) => ({ ...l, project: true }));
      ajax
        .GET("/api/git/project_options")
        .then(
          ({ data }) =>
            data && setOptions({ project: data.items, branch: [], commit: [] })
        )
        .finally(() => setLoading((l) => ({ ...l, project: false })));
    },
    [options.project]
  );

  useEffect(() => {
    if (repoId > 0 && options.project.length > 0) {
      let found = findProject(repoId);
      found && setNeedGitRepo(found.needGitRepo);
    }
  }, [repoId, options.project, findProject]);

  useEffect(() => {
    let found = findProject(repoId);
    if (found && found.needGitRepo && found.gitProjectId > 0 && branch) {
      setLoading((v) => ({ ...v, commit: true }));
      form.setFieldValue("commit", "");
      ajax
        .GET(
          "/api/git/projects/{gitProjectId}/branches/{branch}/commit_options",
          {
            params: {
              path: { gitProjectId: found.gitProjectId, branch: branch },
            },
          }
        )
        .then(
          ({ data }) =>
            data && setOptions((opt) => ({ ...opt, commit: data.items }))
        )
        .finally(() => {
          setLoading((v) => ({ ...v, commit: false }));
        });
    }
  }, [branch, repoId, findProject, form]);

  useEffect(() => {
    let found = findProject(Number(repoId));
    if (found && found.needGitRepo && found.gitProjectId > 0) {
      setLoading((v) => ({ ...v, branch: true }));
      form.setFieldValue("branch", "");
      form.setFieldValue("commit", "");
      ajax
        .GET("/api/git/projects/{gitProjectId}/branch_options", {
          params: { path: { gitProjectId: found.gitProjectId } },
        })
        .then(({ data, error }) => {
          if (error) {
            message.error(error.message);
            return;
          }
          setOptions((opt) => ({ ...opt, branch: data.items }));
        })
        .finally(() => {
          setLoading((v) => ({ ...v, branch: false }));
        });
    }
  }, [findProject, form, repoId]);

  const isBiggerScreen = () => {
    return screens.md || screens.lg || screens.xl || screens.xxl;
  };

  const [pipeline, setPipeline] = useState({
    projectID: 0,
    branch: "",
    commit: "",
  });

  useEffect(() => {
    if (repoId > 0 && commit && branch) {
      let found = findProject(repoId);
      if (found) {
        setPipeline({ projectID: found.gitProjectId, branch, commit });
      }
    }
  }, [commit, branch, repoId, findProject]);

  const [container, setContainer] = useState<HTMLDivElement | null>(null);

  const [elements, setElements] = useState<
    components["schemas"]["mars.Element"][]
  >([]);

  const [mode, setMode] = useState<string>("text/x-yaml");
  const loadConfigFile = useCallback(
    (repoId: number) => {
      ajax
        .GET("/api/repos/{id}", {
          params: {
            path: {
              id: repoId,
            },
          },
        })
        .then(({ data }) => {
          if (data) {
            if (!form.getFieldValue("config")) {
              form.setFieldsValue({
                config: data.item.marsConfig.configFileValues,
              });
            }
            setMode(getMode(data.item.marsConfig.configFileType));
            console.log(data.item);
            setElements(data.item.marsConfig.elements);
          }
        });
    },
    [form]
  );
  useEffect(() => {
    console.log(repoId);
    if (repoId) {
      loadConfigFile(repoId);
    }
  }, [repoId, loadConfigFile]);

  useEffect(() => {
    if (repoId) {
      form.setFieldValue("extraValues", {});
    }
  }, [repoId, form]);

  const info = useSelector(selectClusterInfo);
  const [focusIdx, setFocusIdx] = useState<number | null>(null);
  let slug = useMemo(() => {
    let found = findProject(repoId);
    if (found) {
      return toSlug(namespaceId, found.label);
    }
  }, [namespaceId, findProject, repoId]);

  const ws = useWs();
  const wsReady = useWsReady();
  const dispatch = useDispatch();
  const list = useSelector(selectList);
  const isLoading = useMemo(
    () => (slug ? list[slug]?.isLoading : false),
    [list, slug]
  );
  const deployStatus = useMemo(
    () => (slug ? list[slug]?.deployStatus : DeployStatus.DeployUnknown),
    [list, slug]
  );
  const processPercent = useMemo(
    () => (slug ? list[slug]?.processPercent : 0),
    [list, slug]
  );
  const timer = useSelector(selectTimer);
  const start = useMemo(
    () => (slug ? timer[slug]?.start : false),
    [timer, slug]
  );
  const startAt = useMemo(
    () => (slug ? timer[slug]?.startAt : 0),
    [timer, slug]
  );
  const [deployStarted, setDeployStarted] = useState(false);
  const setStart = useCallback(
    (start: boolean) => {
      slug && dispatch(dispatchSetStart(slug, start));
    },
    [dispatch, slug]
  );
  const [showLog, setShowLog] = useState(start);
  const setStartAt = useCallback(
    (startAt: number) => {
      slug && dispatch(dispatchSetStartAt(slug, startAt));
    },
    [dispatch, slug]
  );

  const onFinish = useCallback(
    (values: FormTypes) => {
      console.log(values);
      if (!wsReady) {
        message.error("连接断开了");
        return;
      }
      if (!slug) {
        return;
      }
      let createParams = websocket.CreateProjectInput.encode({
        type: websocket.Type.CreateProject,
        namespaceId: namespaceId,
        repoId: values.repoId,
        gitBranch: values.branch,
        gitCommit: values.commit,
        extraValues: values.extraValues,
        config: values.config,
      }).finish();

      dispatch(setDeployStatus(slug, DeployStatus.DeployUnknown));
      dispatch(clearCreateProjectLog(slug));
      dispatch(setCreateProjectLoading(slug, true));
      setShowLog(true);
      setStart(true);
      setStartAt(Date.now());
      setDeployStarted(true);
      ws?.send(createParams);
      return;
    },
    [dispatch, namespaceId, setStart, setStartAt, slug, ws, wsReady]
  );

  const onRemove = useCallback(() => {
    if (!wsReady) {
      // message.error("连接断开了");
      return;
    }
    let found = findProject(repoId);
    if (found) {
      let s = websocket.CancelInput.encode({
        type: websocket.Type.CancelProject,
        namespaceId: namespaceId,
        name: found.label,
      }).finish();
      ws?.send(s);
    }
  }, [findProject, namespaceId, repoId, ws, wsReady]);

  return (
    <Form
      layout="horizontal"
      form={form}
      labelWrap
      autoComplete="off"
      onFinish={onFinish}
      // initialValues={}
      style={{ height: "100%" }}
    >
      <div
        ref={setContainer}
        className={css`
          overflow-y: auto;
        `}
        style={{ display: "flex", flexDirection: "column", height: "100%" }}
      >
        <Row style={{ height: "100%" }}>
          <Space
            className={css`
              & > .ant-space-item:last-of-type {
                height: 100%;
              }
            `}
            direction="vertical"
            style={{ width: "100%" }}
          >
            <Affix
              className={css`
                & > div {
                  height: 100%;
                }
              `}
              target={() => container}
              style={{ zIndex: 18, width: "100%" }}
            >
              <div>
                <Row style={{ backgroundColor: "white" }}>
                  <Col span={24}>
                    {needGitRepo && (
                      <PipelineInfo
                        projectId={pipeline.projectID}
                        branch={pipeline.branch}
                        commit={pipeline.commit}
                      />
                    )}
                  </Col>
                  <MyCol
                    md={needGitRepo ? 8 : 24}
                    xs={24}
                    sm={24}
                    onFocus={() => setFocusIdx(1)}
                    onBlur={() => setFocusIdx(null)}
                    focus={focusIdx === 1 ? 1 : 0}
                  >
                    <Form.Item name={"repoId"}>
                      <Select
                        loading={loading.project}
                        showSearch
                        className={
                          needGitRepo && isBiggerScreen()
                            ? css`
                                .ant-select-selector {
                                  border-top-right-radius: 0 !important;
                                  border-bottom-right-radius: 0 !important;
                                }
                              `
                            : ""
                        }
                        placeholder="选择项目"
                        optionFilterProp="label"
                        defaultActiveFirstOption={false}
                        onDropdownVisibleChange={onProjectVisibleChange}
                        options={options.project.map((v) => ({
                          label: v.label,
                          value: v.value,
                        }))}
                      />
                    </Form.Item>
                  </MyCol>
                  {needGitRepo && (
                    <>
                      <MyCol
                        md={8}
                        xs={24}
                        sm={24}
                        onFocus={() => setFocusIdx(2)}
                        onBlur={() => setFocusIdx(null)}
                        focus={focusIdx === 2 ? 1 : 0}
                      >
                        <Form.Item name={"branch"}>
                          <Select
                            loading={loading.branch}
                            showSearch
                            className={
                              needGitRepo && isBiggerScreen()
                                ? css`
                                    .ant-select-selector {
                                      border-radius: 0 !important;
                                    }
                                  `
                                : ""
                            }
                            placeholder="选择分支"
                            optionFilterProp="label"
                            defaultActiveFirstOption={false}
                            options={options.branch.map((v) => ({
                              label: v.label,
                              value: v.value,
                            }))}
                          />
                        </Form.Item>
                      </MyCol>
                      <MyCol
                        md={8}
                        xs={24}
                        sm={24}
                        onFocus={() => setFocusIdx(3)}
                        onBlur={() => setFocusIdx(null)}
                        focus={focusIdx === 3 ? 1 : 0}
                      >
                        <Form.Item name={"commit"}>
                          <Select
                            showSearch
                            loading={loading.commit}
                            className={
                              needGitRepo && isBiggerScreen()
                                ? css`
                                    .ant-select-selector {
                                      border-top-left-radius: 0 !important;
                                      border-bottom-left-radius: 0 !important;
                                    }
                                  `
                                : ""
                            }
                            placeholder="选择 commit"
                            optionFilterProp="label"
                            defaultActiveFirstOption={false}
                            options={options.commit.map((v) => ({
                              label: v.label,
                              value: v.value,
                            }))}
                          />
                        </Form.Item>
                      </MyCol>
                    </>
                  )}
                </Row>
                <Row>
                  <Space>
                    <Button
                      onClick={() => form.submit()}
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
                    {slug && list[slug] && list[slug].output?.length > 0 && (
                      <Button
                        type="dashed"
                        style={{ fontSize: 12, marginRight: 5 }}
                        size="small"
                        onClick={() => setShowLog((show) => !show)}
                      >
                        {showLog ? "隐藏" : "查看"}日志
                      </Button>
                    )}
                  </Space>
                </Row>
              </div>
            </Affix>
            <Row>
              <Col span={showLog ? 24 : 0}>
                <Progress
                  strokeColor={{
                    from: "#108ee9",
                    to: "#87d068",
                  }}
                  style={{ padding: "0 3px", marginBottom: 5 }}
                  percent={processPercent}
                  status="active"
                />
                {slug && (
                  <LogOutput
                    pending={<TimeCost start={start} startAt={startAt} />}
                    slug={slug}
                  />
                )}
              </Col>
            </Row>
            {!showLog && (
              <>
                <Row>
                  <Col span={24}>
                    <Form.Item name="extraValues" noStyle>
                      <Elements
                        elements={elements}
                        style={{
                          inputNumber: { fontSize: 10, width: "100%" },
                          input: { fontSize: 10 },
                          label: { fontSize: 10 },
                          textarea: { fontSize: 10 },
                          formItem: {
                            marginBottom: 2,
                            marginTop: 0,
                            display: "inline-block",
                            width: "calc(33.3% - 8px)",
                            marginRight: 8,
                          },
                        }}
                      />
                    </Form.Item>
                  </Col>
                </Row>
                <Row style={{ height: "100%" }}>
                  <Col span={24}>
                    <Form.Item name="config" noStyle>
                      <MyCodeMirror mode={mode} />
                    </Form.Item>
                  </Col>
                </Row>
              </>
            )}
          </Space>
        </Row>
      </div>
    </Form>
  );
};

export default memo(DeployProjectForm);

const MyCol = styled(Col)<{ focus: number }>`
  margin-right: -1px;
  &:hover {
    z-index: 100;
  }
  ${(p) =>
    p.focus &&
    `
    z-index: 100;
  `}
`;
