import React, { useState, memo, useCallback, useEffect } from "react";
import { Affix, Col, Form, Row, Select, Space } from "antd";
import { css } from "@emotion/css";
import ajax from "../api/ajax";
import { components } from "../api/schema";
import { Grid } from "antd";
import PipelineInfo from "./PipelineInfo";
import Elements from "./elements/Elements";
import { getMode, MyCodeMirror } from "./MyCodeMirror";

const { useBreakpoint } = Grid;

const DeployProjectForm: React.FC = () => {
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
        .then(({ data }) => {
          data && setOptions((opt) => ({ ...opt, branch: data.items }));
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

  return (
    <Form
      layout="horizontal"
      form={form}
      labelWrap
      autoComplete="off"
      onFinish={(values) => {
        console.log(values);
      }}
      // initialValues={initFormValues}
      style={{
        height: "100%",
        // display: "flex",
        // flexDirection: "column",
      }}
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
                <Col md={needGitRepo ? 8 : 24} xs={24} sm={24}>
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
                </Col>
                {needGitRepo && (
                  <>
                    <Col md={8} xs={24} sm={24}>
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
                    </Col>
                    <Col md={8} xs={24} sm={24}>
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
                    </Col>
                  </>
                )}
              </Row>
            </Affix>
            <Row>
              <Col span={24}>
                <Form.Item name="extra_values" noStyle>
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
          </Space>
        </Row>
      </div>
    </Form>
  );
};

export default memo(DeployProjectForm);
