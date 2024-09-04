import React, { useEffect, memo, useMemo } from "react";
import yaml from "js-yaml";
import { MyCodeMirror as CodeMirror, getMode } from "./MyCodeMirror";
import { CloseOutlined, CopyOutlined } from "@ant-design/icons";
import CopyToClipboard from "./CopyToClipboard";
import _ from "lodash";
import DynamicElement from "./elements/DynamicElement";
import SelectFileType from "./SelectFileType";
import { useAsyncState } from "../utils/async";
import { css } from "@emotion/css";
import {
  Drawer,
  Switch,
  Select,
  Button,
  Row,
  Badge,
  Input,
  Col,
  Form,
  message,
  Skeleton,
  Popover,
} from "antd";
import { PrismLight as SyntaxHighlighter } from "react-syntax-highlighter";
import { materialDark } from "react-syntax-highlighter/dist/esm/styles/prism";
import pyaml from "react-syntax-highlighter/dist/esm/languages/prism/yaml";
import { components } from "../api/schema";
import ajax from "../api/ajax";
import { FormProps } from "antd/lib";
import TextArea from "antd/es/input/TextArea";

SyntaxHighlighter.registerLanguage("yaml", pyaml);

type FieldType =
  | components["schemas"]["repo.CreateRequest"]
  | components["schemas"]["repo.UpdateRequest"];

const onFinishFailed: FormProps<FieldType>["onFinishFailed"] = (errorInfo) => {
  console.log("Failed:", errorInfo);
};

const AddRepoModal: React.FC<{
  visible: boolean;
  editItem?: components["schemas"]["types.RepoModel"];
  onCancel: () => void;
  onSuccess?: () => void;
}> = ({ visible, editItem, onSuccess, onCancel }) => {
  const [form] = Form.useForm();
  const needGitRepo = Form.useWatch("needGitRepo", form);
  const gitProjectId = Form.useWatch("gitProjectId", form);
  const configFileType = Form.useWatch(["marsConfig", "configFileType"], form);
  const localChartPath = Form.useWatch(["marsConfig", "localChartPath"], form);
  const configField = Form.useWatch(["marsConfig", "configField"], form);
  const configFileValues = Form.useWatch(
    ["marsConfig", "configFileValues"],
    form,
  );

  let isEdit = !!editItem && editItem.id > 0;
  const [projects, setProjects] = useAsyncState<
    components["schemas"]["git.AllReposResponse_Item"][]
  >([]);

  const [branches, setBranches] = useAsyncState<
    components["schemas"]["git.Option"][]
  >([]);

  const [loading, setLoading] = useAsyncState({
    project: false,
    branch: false,
    submit: false,
  });
  const [configFileContent, setConfigFileContent] = useAsyncState("");
  const [configFileTip, setConfigFileTip] = useAsyncState(false);
  const [valuesYaml, setValuesYaml] = useAsyncState("");

  const onDestroy = () => {
    setLoading({ project: false, branch: false, submit: false });
    setConfigFileContent("");
    setValuesYaml("");
    setConfigFileTip(false);
    form.resetFields();
  };

  const onFinish: FormProps<FieldType>["onFinish"] = async (values) => {
    console.log(values);
    setLoading((v) => ({ ...v, submit: true }));
    if (editItem && editItem.id > 0) {
      const { error } = await ajax.PUT("/api/repos/{id}", {
        body: { ...values, id: editItem.id },
        params: { path: { id: editItem.id } },
      });
      setLoading((v) => ({ ...v, submit: false }));
      if (error) {
        message.error(error.message);
        return;
      }
      message.success("更新成功！");
      onDestroy();
      onSuccess?.();
      return;
    }
    const { error } = await ajax.POST("/api/repos", { body: values });
    setLoading((v) => ({ ...v, submit: false }));
    if (error) {
      message.error(error.message);
      return;
    }
    onDestroy();
    message.success("创建成功！");
    onSuccess?.();
  };

  const getChartValues = useMemo(
    () =>
      _.debounce((input: string) => {
        ajax
          .POST("/api/git/get_chart_values_yaml", {
            body: { input: input },
          })
          .then(({ data, error }) => {
            if (!error) {
              setValuesYaml(data.values);
            }
          });
      }, 2000),
    [setValuesYaml],
  );

  useEffect(() => {
    if (localChartPath) {
      getChartValues(localChartPath);
    }
  }, [localChartPath, getChartValues]);

  useEffect(() => {
    if (gitProjectId > 0 && visible) {
      if (isEdit && gitProjectId !== editItem?.gitProjectId) {
        form.setFieldValue(["marsConfig", "branches"], []);
      }
      setBranches([]);
      setLoading((item) => ({ ...item, branch: true }));
      ajax
        .GET("/api/git/projects/{gitProjectId}/branch_options", {
          params: {
            path: {
              gitProjectId: gitProjectId,
            },
          },
        })
        .then(({ data }) => {
          data && setBranches(data.items);
        })
        .finally(() => {
          setLoading((item) => ({ ...item, branch: false }));
        });
    } else {
      setBranches([]);
    }
  }, [
    gitProjectId,
    form,
    isEdit,
    editItem?.gitProjectId,
    setBranches,
    visible,
    setLoading,
  ]);

  useEffect(() => {
    if (visible) {
      setLoading((item) => ({ ...item, project: true }));
      ajax.GET("/api/git/all_repos").then(({ data, error }) => {
        setLoading((item) => ({ ...item, project: false }));
        if (error) {
          return;
        }
        console.log(data);
        data && setProjects(data.items);
      });
    }
  }, [setLoading, setProjects, visible]);

  const allBranches = () => {
    let allBranches = [{ value: "*", label: "全部" }];
    if (branches) {
      allBranches = [
        ...allBranches,
        ...branches.map((item) => ({
          value: item.branch,
          label: item.label,
        })),
      ];
    }
    return allBranches;
  };
  const getProjectOptions = () => {
    let opt: { value: number; label: React.ReactNode; search: string }[] = [
      { value: 0, label: "未选择", search: "" },
    ];
    if (projects) {
      opt = [
        ...opt,
        ...projects.map((item) => ({
          value: item.id,
          search: item.name,
          label: (
            <div>
              <span>{item.name}</span>
              <span
                className={css`
                  margin-left: 10px;
                  font-size: 10;
                  color: #a3a3a3;
                `}
              >
                {item.description}
              </span>
            </div>
          ),
        })),
      ];
    }
    return opt;
  };
  useEffect(() => {
    if (visible) {
      let d = _.debounce(() => {
        if (configField && valuesYaml) {
          let data = _.get(yaml.load(valuesYaml), configField.split("->"), "");
          form.setFieldValue(
            ["marsConfig", "isSimpleEnv"],
            typeof data === "object" ? false : true,
          );
          if (typeof data === "object") {
            data = yaml.dump(data);
          }
          setConfigFileTip(true);
          setConfigFileContent(String(data));
        } else {
          setConfigFileContent("");
        }
      }, 1000);
      d();
      return () => {
        d.cancel();
      };
    }
  }, [
    configField,
    isEdit,
    setConfigFileTip,
    form,
    setConfigFileContent,
    valuesYaml,
    visible,
  ]);

  return (
    <Drawer
      keyboard={false}
      title={
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignContent: "center",
          }}
        >
          <span>{isEdit ? "更新" : "添加"} repo</span>
          <Button
            loading={loading.submit}
            size="small"
            type="primary"
            onClick={() => form.submit()}
          >
            {isEdit && editItem ? "更新: " + editItem.name : "创建"}
          </Button>
        </div>
      }
      className={css`
        .ant-input[disabled],
        .ant-select-disabled.ant-select:not(.ant-select-customize-input)
          .ant-select-selector,
        .ant-select-disabled.ant-select-multiple .ant-select-selection-item,
        .ant-switch-disabled {
          color: rgba(0, 0, 0, 1);
          opacity: 1;
        }
      `}
      open={visible}
      footer={null}
      width={"100%"}
      destroyOnClose
      onClose={() => {
        onDestroy();
        onCancel();
      }}
    >
      <Row gutter={[20, 12]}>
        <Col
          span={12}
          style={{
            maxHeight: "800px",
            overflowY: "scroll",
            position: "relative",
          }}
        >
          <Badge.Ribbon color="purple" text="charts 默认值">
            <div>
              <div
                style={{
                  position: "absolute",
                  top: 40,
                  right: 20,
                  zIndex: 99999,
                  color: "white",
                }}
              >
                <CopyToClipboard text={valuesYaml} successText="已复制！">
                  <CopyOutlined />
                </CopyToClipboard>
              </div>
              <SyntaxHighlighter
                language="yaml"
                style={materialDark}
                customStyle={{
                  minHeight: 200,
                  lineHeight: 1.2,
                  padding: "10px",
                  fontFamily: '"Fira code", "Fira Mono", monospace',
                  fontSize: 13,
                  margin: 0,
                  height: "100%",
                }}
              >
                {valuesYaml}
              </SyntaxHighlighter>
            </div>
          </Badge.Ribbon>
        </Col>
        <Col
          span={12}
          style={{
            maxHeight: "800px",
            overflowY: "scroll",
            position: "relative",
          }}
        >
          <Form
            form={form}
            name="basic"
            clearOnDestroy
            layout="vertical"
            autoComplete="off"
            initialValues={isEdit ? editItem : {}}
            onFinish={onFinish}
            onFinishFailed={onFinishFailed}
          >
            <Row gutter={[8, 8]}>
              <Col span={8}>
                <Form.Item<FieldType>
                  label="应用名称"
                  name="name"
                  rules={[{ required: true, message: "请输入名称" }]}
                >
                  <Input />
                </Form.Item>
              </Col>
              <Col span={16}>
                <Form.Item<FieldType> label="项目描述" name="description">
                  <TextArea rows={1} />
                </Form.Item>
              </Col>
            </Row>
            <Form.Item<FieldType> label="是否关联 git 仓库" name="needGitRepo">
              <Switch />
            </Form.Item>
            {needGitRepo && (
              <Row gutter={[8, 8]}>
                <Col span={12}>
                  <Form.Item<FieldType> label="git 仓库" name="gitProjectId">
                    {loading.project ? (
                      <Skeleton.Input block style={{ width: "100%" }} active />
                    ) : (
                      <Select
                        showSearch
                        allowClear
                        optionFilterProp="search"
                        options={getProjectOptions()}
                      />
                    )}
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item
                    label="启用的分支"
                    name={["marsConfig", "branches"]}
                  >
                    {loading.branch ? (
                      <Skeleton.Input block style={{ width: "100%" }} active />
                    ) : (
                      <Select mode="multiple" options={allBranches()} />
                    )}
                  </Form.Item>
                </Col>
              </Row>
            )}

            <Form.Item
              label="charts 地址, 格式为 'pid|branch|path'"
              name={["marsConfig", "localChartPath"]}
              rules={[{ required: true, message: "charts 路径必填" }]}
            >
              <Input />
            </Form.Item>

            <Row gutter={[8, 8]}>
              <Col span={12}>
                <Form.Item
                  label="用户输入配置字段"
                  tooltip={`用户在部署时使用的自定义配置字段, 比如 "conf->config"`}
                  name={["marsConfig", "configField"]}
                >
                  <Input />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item
                  label="配置文件类型"
                  name={["marsConfig", "configFileType"]}
                >
                  <SelectFileType />
                </Form.Item>
              </Col>
            </Row>

            <Col>
              <Form.Item
                label="单字段"
                tooltip="配置文件是不是一个整体的value值"
                name={["marsConfig", "isSimpleEnv"]}
                valuePropName="checked"
              >
                <Switch defaultChecked />
              </Form.Item>
            </Col>

            <Popover
              overlayInnerStyle={{
                maxHeight: 400,
                maxWidth: 600,
                overflowY: "scroll",
              }}
              placement="left"
              content={
                <div>
                  <SyntaxHighlighter
                    language="yaml"
                    style={materialDark}
                    customStyle={{
                      lineHeight: 1.2,
                      padding: "10px",
                      fontFamily: '"Fira code", "Fira Mono", monospace',
                      fontSize: 12,
                      margin: 0,
                      height: "100%",
                    }}
                  >
                    {String(configFileContent)}
                  </SyntaxHighlighter>
                  <Button
                    size="small"
                    onClick={() => {
                      form.setFieldValue(
                        ["marsConfig", "configFileValues"],
                        String(configFileContent),
                      );
                      setConfigFileTip(false);
                      setConfigFileContent("");
                    }}
                    type="dashed"
                    style={{ marginTop: 3, fontSize: 12 }}
                  >
                    使用该配置
                  </Button>
                </div>
              }
              title={
                <div
                  style={{
                    display: "flex",
                    justifyContent: "space-between",
                    alignItems: "center",
                  }}
                >
                  <div>检测到可用配置</div>
                  <Button
                    size="small"
                    type="link"
                    onClick={() => setConfigFileTip(false)}
                    icon={<CloseOutlined />}
                  ></Button>
                </div>
              }
              trigger="hover"
              open={
                configFileTip &&
                configFileValues !== configFileContent &&
                !configFileValues
              }
              onOpenChange={(v) => setConfigFileTip(v)}
            >
              <Form.Item
                label="全局配置文件"
                tooltip="全局默认配置文件，如果没有设置 config_file 则使用这个"
                name={["marsConfig", "configFileValues"]}
              >
                <CodeMirror mode={getMode(configFileType)} />
              </Form.Item>
            </Popover>
            <DynamicElement form={form} />

            <div
              style={{
                maxHeight: "800px",
                overflowY: "scroll",
                position: "relative",
                fontSize: 13,
              }}
            >
              <Form.Item
                name={["marsConfig", "valuesYaml"]}
                style={{
                  maxHeight: "800px",
                  overflowY: "scroll",
                  position: "relative",
                  fontSize: 13,
                }}
                label={
                  <div>
                    <div>
                      values.yaml &nbsp;&nbsp;&nbsp;
                      <span style={{ fontSize: 12 }}>
                        自动补全: 'alt+enter'
                      </span>
                    </div>
                  </div>
                }
                tooltip="等同于 helm 的 values.yaml, 特别注意: 不能出现特殊的用 '<>' 包裹的变量, go 模板会解析失败!"
              >
                <CodeMirror completionValues mode={getMode("yaml")} />
              </Form.Item>
            </div>
          </Form>
        </Col>
      </Row>
    </Drawer>
  );
};

export default memo(AddRepoModal);
