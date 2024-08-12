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
import { components } from "../api/schema";
import { Grid } from "antd";
import PipelineInfo from "./PipelineInfo";
import Elements from "./elements/Elements";
import { getMode, MyCodeMirror } from "./MyCodeMirror";
import { StopOutlined } from "@ant-design/icons";
import { selectClusterInfo } from "../store/reducers/cluster";
import { useSelector } from "react-redux";
import styled from "@emotion/styled";
import TimeCost from "./TimeCost";
import { toSlug } from "../utils/slug";
import LogOutput from "./LogOutput";
import useDeploy from "../contexts/useDeploy";
import ConfigHistory from "./ConfigHistory";
import { ReactDiffViewerStylesOverride } from "react-diff-viewer";
import DiffViewer from "./DiffViewer";

const { useBreakpoint } = Grid;

interface FormTypes {
  extraValues: components["schemas"]["websocket.ExtraValue"][];
  config: string;
  branch: string;
  commit: string;
  repoId: number;
}

const defaultCurr = {
  slug: "",
  appName: "",
  needGitRepo: false,
  gitProjectId: 0,
  projectId: 0,
};

const DeployProjectForm: React.FC<{
  namespaceId: number;
  onSuccess?: () => void;
  project?: components["schemas"]["types.ProjectModel"];
  isEdit?: boolean;
}> = ({ namespaceId, project, isEdit, onSuccess }) => {
  const [form] = Form.useForm();

  const initValues = project
    ? {
        extraValues: project.extraValues,
        config: project.config,
        branch: project.gitBranch,
        commit: project.gitCommit,
        repoId: String(project.repoId),
      }
    : { extraValues: [], config: "", branch: "", commit: "", repoId: "" };

  const formRepoId = Form.useWatch("repoId", form);
  const branch = Form.useWatch("branch", form);
  const commit = Form.useWatch("commit", form);
  const config = Form.useWatch("config", form);

  const [repoId, setRepoId] = useState(project ? project.repoId : 0);
  useEffect(() => {
    setRepoId(formRepoId);
  }, [formRepoId]);

  const [options, setOptions] = useState<{
    project: components["schemas"]["git.Option"][];
    branch: components["schemas"]["git.Option"][];
    commit: components["schemas"]["git.Option"][];
  }>({
    project: [],
    branch: [],
    commit: [],
  });

  const curr = useMemo((): {
    slug: string;
    appName: string;
    needGitRepo: boolean;
    gitProjectId: number;
    projectId: number;
  } => {
    let found = options.project.find((v) => Number(v.value) === Number(repoId));
    return found
      ? {
          slug: toSlug(namespaceId, found.label),
          appName: found.label,
          needGitRepo: found.needGitRepo,
          gitProjectId: found.gitProjectId,
          projectId: project ? project.id : 0,
        }
      : defaultCurr;
  }, [options.project, repoId, namespaceId, project]);

  const [loading, setLoading] = useState({
    project: false,
    branch: false,
    commit: false,
  });
  const screens = useBreakpoint();

  const [needGitRepo, setNeedGitRepo] = useState(true);

  const onProjectVisibleChange = useCallback((open: boolean) => {
    if (!open) {
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
  }, []);

  useEffect(() => {
    if (isEdit) {
      onProjectVisibleChange(true);
    }
  }, [isEdit, onProjectVisibleChange]);

  useEffect(() => {
    setNeedGitRepo(curr.needGitRepo);
  }, [repoId, options.project, curr]);

  useEffect(() => {
    if (curr.gitProjectId > 0 && branch) {
      setLoading((v) => ({ ...v, commit: true }));
      form.setFieldValue("commit", "");
      ajax
        .GET(
          "/api/git/projects/{gitProjectId}/branches/{branch}/commit_options",
          {
            params: {
              path: { gitProjectId: curr.gitProjectId, branch: branch },
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
  }, [branch, curr.gitProjectId, curr.slug, form]);

  useEffect(() => {
    if (curr.gitProjectId && repoId) {
      setLoading((v) => ({ ...v, branch: true }));
      form.setFieldValue("branch", "");
      form.setFieldValue("commit", "");
      ajax
        .GET("/api/git/projects/{gitProjectId}/branch_options", {
          params: {
            path: { gitProjectId: curr.gitProjectId },
            query: { repoId: repoId },
          },
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
  }, [curr, form, repoId]);

  const isBiggerScreen = useMemo(() => {
    return screens.md || screens.lg || screens.xl || screens.xxl;
  }, [screens]);

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
            if (!isEdit) {
              form.setFieldsValue({
                config: data.item.marsConfig.configFileValues,
              });
              form.setFieldValue("extraValues", data.item.marsConfig.elements);
            }
            setElements(data.item.marsConfig.elements);
            setMode(getMode(data.item.marsConfig.configFileType));
          }
        });
    },
    [form, isEdit]
  );

  useEffect(() => {
    if (repoId) {
      !isEdit && form.setFieldValue("extraValues", []);
      loadConfigFile(repoId);
    }
  }, [repoId, loadConfigFile, form, isEdit]);

  const info = useSelector(selectClusterInfo);
  const [focusIdx, setFocusIdx] = useState<number | null>(null);

  const {
    hasLog,
    isLoading,
    cancelDeploy,
    createProject,
    updateProject,
    processPercent,
    isSuccess,
    clearProject,
  } = useDeploy({
    namespaceID: namespaceId,
    slug: curr.slug,
  });
  const [showLog, setShowLog] = useState(false);

  useEffect(() => {
    if (isSuccess) {
      onSuccess?.();
      clearProject();
      setShowLog(false);
    }
  }, [isSuccess, onSuccess, clearProject]);

  const onFinish = useCallback(
    (values: FormTypes) => {
      if (curr.needGitRepo && (!values.branch || !values.commit)) {
        message.warning("请先选择分支/commit");
        return;
      }
      if (isEdit) {
        project?.version &&
          updateProject({
            projectId: curr.projectId,
            version: project.version,
            branch: values.branch,
            commit: values.commit,
            extraValues: values.extraValues,
            config: values.config,
          });
      } else {
        createProject({
          repoId: values.repoId,
          branch: values.branch,
          commit: values.commit,
          extraValues: values.extraValues,
          config: values.config,
        });
      }

      setShowLog(true);
      return;
    },
    [
      createProject,
      curr.needGitRepo,
      isEdit,
      curr.projectId,
      project?.version,
      updateProject,
    ]
  );
  // console.log("curr", curr, repoId);
  // console.log("initval", initValues);
  return (
    <Form
      layout="horizontal"
      form={form}
      labelWrap
      autoComplete="off"
      onFinish={onFinish}
      clearOnDestroy
      initialValues={project ? initValues : {}}
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
                        projectId={curr.gitProjectId}
                        branch={branch}
                        commit={commit}
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
                          needGitRepo && isBiggerScreen
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
                          label: (
                            <div>
                              {v.label}
                              <span
                                style={{
                                  color: "gray",
                                  marginLeft: 10,
                                  fontSize: 10,
                                }}
                              >
                                {v.description}
                              </span>
                            </div>
                          ),
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
                              needGitRepo && isBiggerScreen
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
                              needGitRepo && isBiggerScreen
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
                  <Space size={"small"}>
                    <Button
                      onClick={() => form.submit()}
                      style={{ fontSize: 12 }}
                      size="small"
                      danger={info.status === "bad"}
                      type={"primary"}
                      loading={isLoading}
                    >
                      {info.status === "bad" ? "集群资源不足" : "部署"}
                    </Button>
                    {isEdit && !isLoading && (
                      <Button
                        size="small"
                        style={{ fontSize: 12 }}
                        disabled={isLoading}
                        onClick={() => form.resetFields()}
                      >
                        重置
                      </Button>
                    )}
                    {isLoading && (
                      <Button
                        style={{ fontSize: 12 }}
                        size="small"
                        danger
                        icon={<StopOutlined />}
                        type="dashed"
                        onClick={() => {
                          cancelDeploy(curr.appName);
                        }}
                      >
                        取消
                      </Button>
                    )}
                    {hasLog && (
                      <Button
                        type="dashed"
                        style={{ fontSize: 12 }}
                        size="small"
                        onClick={() => setShowLog((show) => !show)}
                      >
                        {showLog ? "隐藏" : "查看"}日志
                      </Button>
                    )}
                    {!isLoading && isEdit && (
                      <ConfigHistory
                        projectID={curr.projectId}
                        configType={
                          project?.repo.marsConfig.configFileType || ""
                        }
                      />
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
                <LogOutput
                  pending={<TimeCost done={!isLoading} />}
                  slug={curr.slug}
                />
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
                  <Col
                    span={
                      isEdit && curr.appName && config !== project?.config
                        ? 12
                        : 24
                    }
                  >
                    <Form.Item name="config" noStyle>
                      <MyCodeMirror mode={mode} />
                    </Form.Item>
                  </Col>
                  {isEdit && curr.appName && (
                    <Col
                      className={css`
                        pre {
                          line-height: 20px !important;
                        }
                      `}
                      span={isEdit && config !== project?.config ? 12 : 0}
                      style={{ fontSize: 13 }}
                    >
                      <DiffViewer
                        styles={diffViewerStyles}
                        mode={project?.repo.marsConfig.configFileType || ""}
                        showDiffOnly={false}
                        oldValue={project?.config || ""}
                        newValue={config}
                        splitView={false}
                      />
                    </Col>
                  )}
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

const diffViewerStyles: ReactDiffViewerStylesOverride = {
  gutter: { padding: "0 5px", minWidth: 25 },
  marker: { padding: "0 6px" },
  diffContainer: {
    height: "100%",
    display: "block",
    width: "100%",
    overflowX: "auto",
  },
};
