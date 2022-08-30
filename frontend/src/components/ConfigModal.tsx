import React, { useCallback, useEffect, memo } from "react";
import { MyCodeMirror as CodeMirror, getMode } from "./MyCodeMirror";
import { CopyOutlined, CloseOutlined } from "@ant-design/icons";
import CopyToClipboard from "./CopyToClipboard";
import DynamicElement from "./elements/DynamicElement";
import SelectFileType from "./SelectFileType";
import { useAsyncState } from "../utils/async";

import pb from "../api/compiled";
import { get, debounce } from "lodash";
import yaml from "js-yaml";

import {
  Tooltip,
  Popover,
  Switch,
  Select,
  Button,
  message,
  Modal,
  Skeleton,
  Row,
  Badge,
  Input,
  Col,
  Form,
} from "antd";
import { QuestionCircleOutlined, EditOutlined } from "@ant-design/icons";
import {
  globalConfig as globalConfigApi,
  marsConfig,
  toggleGlobalEnabled as toggleGlobalEnabledApi,
  updateGlobalConfig,
  getDefaultValues,
} from "../api/mars";
import { branchOptions as branches } from "../api/git";
import MarsExample from "./MarsExample";
import { PrismLight as SyntaxHighlighter } from "react-syntax-highlighter";
import { materialDark } from "react-syntax-highlighter/dist/esm/styles/prism";
import pyaml from "react-syntax-highlighter/dist/esm/languages/prism/yaml";

SyntaxHighlighter.registerLanguage("yaml", pyaml);

interface Config extends pb.mars.Config {}

const { Option } = Select;

const initConfig = {
  display_name: "",
  config_file: "",
  config_file_values: "",
  config_field: "",
  is_simple_env: false,
  config_file_type: "yaml",
  local_chart_path: "",
  branches: [],
  values_yaml: "",
  elements: [],
};

interface WatchData {
  config_field: string;
  config_file_values: string;
  config_file_type: string;
}

const initDefaultValues = "# 没找到对应的 values.yaml";

const ConfigModal: React.FC<{
  visible: boolean;
  item: pb.git.ProjectItem;
  onCancel: () => void;
}> = ({ visible, item, onCancel }) => {
  const [watch, setWatch] = useAsyncState<WatchData>({
    config_field: initConfig.config_field,
    config_file_values: initConfig.config_file_values,
    config_file_type: initConfig.config_file_type,
  });
  const [editMode, setEditMode] = useAsyncState(true);
  const [globalEnabled, setGlobalEnabled] = useAsyncState(false);
  const [config, setConfig] = useAsyncState<Config>(initConfig);
  const [modalBranch, setModalBranch] = useAsyncState("");
  const [configVisible, setConfigVisible] = useAsyncState(visible);
  const [mbranches, setMbranches] = useAsyncState<string[]>([]);
  const [defaultValues, setDefaultValues] =
    useAsyncState<string>(initDefaultValues);
  const [loading, setLoading] = useAsyncState(false);
  const [mode, setMode] = useAsyncState("");
  const [form] = Form.useForm();
  const [configFileContent, setConfigFileContent] = useAsyncState("");
  const [singleConfig, setSingleConfig] = useAsyncState(true);
  const [configFileTip, setConfigFileTip] = useAsyncState(false);
  const loadDefaultValues = useCallback(
    (projectId: number, branch: string) => {
      if (projectId) {
        getDefaultValues({ git_project_id: projectId, branch: branch })
          .then((res) => {
            setDefaultValues(res.data.value);
          })
          .catch((e) => {
            setDefaultValues(initDefaultValues);
          });
      }
    },
    [setDefaultValues]
  );

  const loadConfig = useCallback(
    (id: number, branch = "") => {
      setLoading(true);
      marsConfig({ branch, git_project_id: id })
        .then(({ data }) => {
          if (data.config) {
            setConfig(data.config);
          }
          setModalBranch(data.branch);
          setLoading(false);
          loadDefaultValues(id, data.branch);
        })
        .catch((e) => {
          message.error(e.response.data.message);
          setLoading(false);
        });
    },
    [loadDefaultValues, setConfig, setLoading, setModalBranch]
  );

  const loadGlobalConfig = useCallback(
    (id: number) => {
      setLoading(true);
      globalConfigApi({ git_project_id: id })
        .then(({ data }) => {
          if (data.config) {
            setConfig(data.config);
          }
          loadDefaultValues(id, "");
        })
        .catch((e) => {
          message.error(e.response.data.message);
        })
        .finally(() => {
          setLoading(false);
        });
    },
    [loadDefaultValues, setConfig, setLoading]
  );

  const resetModal = useCallback(() => {
    setMbranches([]);
    setLoading(false);
    setConfig({ ...initConfig });
    setConfigVisible(false);
    setEditMode(true);
    setConfigFileContent("");
    setConfigFileTip(false);
    onCancel();
  }, [
    onCancel,
    setConfig,
    setConfigFileContent,
    setConfigFileTip,
    setConfigVisible,
    setEditMode,
    setLoading,
    setMbranches,
  ]);

  const selectBranch = useCallback(
    (value: string) => {
      if (item) {
        loadConfig(item.id, value);
      }
    },
    [item, loadConfig]
  );

  const toggleGlobalEnabled = useCallback(
    (enabled: boolean) => {
      setConfigFileContent("");
      if (!enabled) {
        setEditMode(false);
      }
      toggleGlobalEnabledApi({ git_project_id: item.id, enabled })
        .then(() => {
          message.success("操作成功");
          setGlobalEnabled(enabled);
          if (enabled) {
            loadGlobalConfig(item.id);
          } else {
            loadConfig(item.id, "");
          }
        })
        .catch((e) => {
          message.error(e.message);
        });
    },
    [
      loadGlobalConfig,
      loadConfig,
      item,
      setConfigFileContent,
      setEditMode,
      setGlobalEnabled,
    ]
  );

  const onSave = useCallback(
    (values: any) => {
      updateGlobalConfig({
        git_project_id: item.id,
        config: values,
      })
        .then((res) => {
          message.success("保存成功");
          res.data.config &&
            setConfig((c) => {
              return { ...c, ...res.data.config };
            });

          loadDefaultValues(item.id, "");
          setEditMode(false);
        })
        .catch((e) => {
          message.error(e.response.data.message);
          globalConfigApi({ git_project_id: item.id }).then(({ data }) => {
            setGlobalEnabled(data.enabled);
          });
        });
    },
    [item.id, loadDefaultValues, setConfig, setEditMode, setGlobalEnabled]
  );

  useEffect(() => {
    if (visible && watch.config_file_type) {
      setMode(getMode(watch.config_file_type));
    }
  }, [watch, visible, setMode]);

  useEffect(() => {
    setConfigVisible(visible);
    if (visible) {
      setLoading(true);
      branches({ git_project_id: String(item.id), all: true }).then((res) => {
        setMbranches(res.data.items.map((op) => op.value));
      });
      globalConfigApi({ git_project_id: item.id })
        .then(({ data }) => {
          setGlobalEnabled(data.enabled);
          if (!data.enabled) {
            loadConfig(item.id);
          } else {
            if (data.config) {
              setConfig(data.config);
            }
            loadDefaultValues(item.id, "");
            setLoading(false);
          }
        })
        .catch((e) => {
          message.error(e.response.data.message);
        });
    }
  }, [
    item,
    loadConfig,
    visible,
    loadDefaultValues,
    setConfig,
    setConfigVisible,
    setGlobalEnabled,
    setLoading,
    setMbranches,
  ]);

  useEffect(() => {
    if (editMode && visible) {
      let d = debounce(() => {
        if (watch.config_field && defaultValues) {
          let data = get(
            yaml.load(defaultValues),
            watch.config_field.split("->"),
            ""
          );
          setSingleConfig(typeof data === "object" ? false : true);
          if (typeof data === "object") {
            data = yaml.dump(data);
          }
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
    editMode,
    watch.config_field,
    setSingleConfig,
    defaultValues,
    visible,
    setConfigFileContent,
  ]);

  useEffect(() => {
    setConfigFileTip(configFileContent.length > 0 && !watch.config_file_values);
  }, [configFileContent, watch.config_file_values, setConfigFileTip]);

  useEffect(() => {
    setWatch((w) => ({ ...w, ...config }));
    form.setFieldsValue(config);
  }, [config, form, setWatch]);

  return (
    <Modal
      keyboard={false}
      title={
        <div>
          <MarsExample />
          &nbsp;&nbsp;{item?.name}
        </div>
      }
      className="config-modal"
      visible={configVisible}
      footer={null}
      width={"100%"}
      onCancel={resetModal}
    >
      <Skeleton active loading={loading}>
        <Form
          form={form}
          initialValues={config}
          name="basic"
          layout="vertical"
          autoComplete="off"
          onFinish={onSave}
        >
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              alignItems: "start",
              height: 40,
            }}
          >
            {!globalEnabled ? (
              <Select
                placeholder="请选择"
                value={modalBranch}
                style={{ width: 250 }}
                loading={loading}
                onChange={selectBranch}
              >
                {mbranches.map((item) => (
                  <Option value={item} key={item}>
                    {item}
                  </Option>
                ))}
              </Select>
            ) : (
              <div>
                <Button
                  style={{ marginRight: 10 }}
                  type="ghost"
                  icon={!editMode ? <EditOutlined /> : null}
                  onClick={() => {
                    setEditMode((editMode) => {
                      if (editMode) {
                        setConfigFileTip(false);
                        setConfigFileContent("");
                        form.resetFields();
                        setWatch({
                          config_field: config.config_field,
                          config_file_values: config.config_file_values,
                          config_file_type: config.config_file_type,
                        });
                      }
                      return !editMode;
                    });
                  }}
                >
                  {!editMode ? "编辑" : "取消"}
                </Button>
                <Button hidden={!editMode} type="primary" htmlType="submit">
                  保存
                </Button>
              </div>
            )}
            <div>
              <span style={{ marginRight: 10 }}>
                使用全局配置&nbsp;
                <Tooltip
                  overlayStyle={{ fontSize: "12px" }}
                  placement="top"
                  title="全局配置优先级最高，会覆盖所有分支的配置"
                >
                  <QuestionCircleOutlined />
                </Tooltip>
              </span>
              <Switch
                checkedChildren="开启"
                unCheckedChildren="关闭"
                checked={globalEnabled}
                onChange={toggleGlobalEnabled}
              />
            </div>
          </div>
          <Row gutter={[3, 12]} className="config-modal__content">
            <Col span={24}>
              <Skeleton active loading={loading}>
                <Row
                  gutter={[16, 16]}
                  style={{ height: "100%", overflowY: "auto" }}
                >
                  <Col
                    span={10}
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
                            display: !editMode ? "none" : "block",
                            position: "absolute",
                            top: 40,
                            right: 20,
                            zIndex: 99999,
                            color: "white",
                          }}
                        >
                          <CopyToClipboard
                            text={defaultValues}
                            successText="已复制！"
                          >
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
                          {defaultValues}
                        </SyntaxHighlighter>
                      </div>
                    </Badge.Ribbon>
                  </Col>
                  <Col
                    span={14}
                    style={{
                      maxHeight: "800px",
                      overflowY: "scroll",
                      position: "relative",
                    }}
                  >
                    <Form.Item
                      label="应用名称"
                      rules={[
                        {
                          pattern: new RegExp("^[A-Za-z]([A-Z-_a-z]*[^_-])*$"),
                          message:
                            "名称必须符合表达式 '^[A-Za-z]([A-Z-_a-z]*[^_-])*$', 格式例如 my_app, my-app, app",
                        },
                      ]}
                      name={"display_name"}
                    >
                      <Input disabled={!editMode || !globalEnabled} />
                    </Form.Item>

                    <Form.Item
                      label="charts 的目录(需要第一个设置并且保存)"
                      name={"local_chart_path"}
                      rules={[{ required: true, message: "charts 路径必填" }]}
                      tooltip="charts 文件在项目中存放的目录(必填), 也可以是别的项目的文件，格式为 'pid|branch|path'"
                    >
                      <Input disabled={!editMode || !globalEnabled} />
                    </Form.Item>

                    <Row style={{ marginBottom: 0 }}>
                      <Form.Item
                        label="用户输入配置字段"
                        tooltip={`用户在部署时使用的自定义配置字段, 比如 "conf->config"`}
                        style={{
                          display: "inline-block",
                          width: "calc(50% - 8px)",
                          marginRight: 8,
                        }}
                        name={"config_field"}
                      >
                        <Input
                          disabled={!editMode || !globalEnabled}
                          onChange={(e) => {
                            form.setFieldsValue({
                              config_field: e.target.value,
                            });
                            setWatch((w) => ({
                              ...w,
                              config_field: e.target.value,
                            }));
                          }}
                        />
                      </Form.Item>

                      <Form.Item
                        label="配置文件类型"
                        style={{
                          display: "inline-block",
                          width: "calc(50% - 8px)",
                        }}
                        name={"config_file_type"}
                      >
                        <SelectFileType
                          onChange={(v) => {
                            setWatch((w) => ({ ...w, config_file_type: v }));
                            form.setFieldsValue({ config_file_type: v });
                          }}
                          showArrow={editMode}
                          disabled={!editMode || !globalEnabled}
                        />
                      </Form.Item>
                    </Row>

                    <Form.Item style={{ marginBottom: 0 }}>
                      <Form.Item
                        label="单字段"
                        tooltip="配置文件是不是一个整体的value值"
                        style={{
                          display: "inline-block",
                          width: "calc(15% - 8px)",
                        }}
                        name={"is_simple_env"}
                        valuePropName="checked"
                      >
                        <Switch
                          disabled={!editMode || !globalEnabled}
                          defaultChecked
                        />
                      </Form.Item>
                      <Form.Item
                        label="启用的分支"
                        style={{
                          display: "inline-block",
                          width: "calc(85% - 8px)",
                        }}
                        name={"branches"}
                      >
                        <Select
                          disabled={!editMode || !globalEnabled}
                          mode="multiple"
                          style={{ width: "100%" }}
                        >
                          <Option value="*">全部</Option>
                          {mbranches.map((v, k) => (
                            <Select.Option key={k} value={v}>
                              {v}
                            </Select.Option>
                          ))}
                        </Select>
                      </Form.Item>
                    </Form.Item>
                    <div style={{ maxHeight: "400px", overflowY: "auto" }}>
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
                                fontFamily:
                                  '"Fira code", "Fira Mono", monospace',
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
                                form.setFieldsValue({
                                  config_file_values: String(configFileContent),
                                  is_simple_env: singleConfig,
                                });
                                setWatch((w) => ({
                                  ...w,
                                  config_file_values: String(configFileContent),
                                }));
                                setConfigFileTip(false);
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
                        trigger="focus"
                        visible={configFileTip && editMode}
                        onVisibleChange={(v) => setConfigFileTip(v)}
                      ></Popover>
                      <Form.Item
                        label="全局配置文件"
                        tooltip="全局默认配置文件，如果没有设置 config_file 则使用这个"
                        name={"config_file_values"}
                      >
                        <CodeMirror
                          mode={mode}
                          onChange={(v) => {
                            form.setFieldsValue({ config_file_values: v });
                            setWatch((w) => ({
                              ...w,
                              config_file_values: v,
                            }));
                          }}
                        />
                      </Form.Item>
                    </div>

                    <DynamicElement
                      form={form}
                      disabled={!editMode || !globalEnabled}
                    />
                    <div
                      style={{
                        maxHeight: "800px",
                        overflowY: "scroll",
                        position: "relative",
                        fontSize: 13,
                      }}
                    >
                      <Form.Item
                        name={"values_yaml"}
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
                        <CodeMirror mode={getMode("yaml")} />
                      </Form.Item>
                    </div>
                  </Col>
                </Row>
              </Skeleton>
            </Col>
          </Row>
        </Form>
      </Skeleton>
    </Modal>
  );
};

export default memo(ConfigModal);
