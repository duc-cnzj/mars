import React, { useState, useCallback, useEffect, memo } from "react";
import { MyCodeMirror as CodeMirror, getMode } from "./MyCodeMirror";
import { CopyOutlined, CloseOutlined } from "@ant-design/icons";
import { CopyToClipboard } from "react-copy-to-clipboard";
// import Elements from "./elements/Elements";
import DynamicElement from "./elements/DynamicElement";
import SelectFileType from "./SelectFileType";
import {cloneDeep} from 'lodash';

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
import { branchOptions as branches } from "../api/gitlab";
import MarsExample from "./MarsExample";
import { PrismLight as SyntaxHighlighter } from "react-syntax-highlighter";
import { materialDark } from "react-syntax-highlighter/dist/esm/styles/prism";
import pyaml from "react-syntax-highlighter/dist/esm/languages/prism/yaml";

SyntaxHighlighter.registerLanguage("yaml", pyaml);

interface Config extends pb.MarsConfig {}

const { Option } = Select;

const initConfig = {
  config_file: "",
  config_file_values: "",
  config_field: "",
  is_simple_env: false,
  config_file_type: "",
  local_chart_path: "",
  branches: [],
  values_yaml: "",
  elements: [],
};

const initDefaultValues = "# 没找到对应的 values.yaml";

const ConfigModal: React.FC<{
  visible: boolean;
  item: undefined | pb.GitProjectItem;
  onCancel: () => void;
}> = ({ visible, item, onCancel }) => {
  const [editMode, setEditMode] = useState(true);
  const [globalEnabled, setGlobalEnabled] = useState(false);
  const [config, setConfig] = useState<Config>(initConfig);
  const [modalBranch, setModalBranch] = useState("");
  const [configVisible, setConfigVisible] = useState(visible);
  const [mbranches, setMbranches] = useState<string[]>([]);
  const [defaultValues, setDefaultValues] = useState<string>(initDefaultValues);
  const [loading, setLoading] = useState(false);
  const [mode, setMode] = useState("");
  const [old, setOld] = useState<Config>();

  const loadDefaultValues = useCallback((projectId: number, branch: string) => {
    if (projectId) {
      getDefaultValues({ project_id: projectId, branch: branch })
        .then((res) => {
          setDefaultValues(res.data.value);
        })
        .catch((e) => {
          setDefaultValues(initDefaultValues);
        });
    }
  }, []);

  const loadConfig = useCallback(
    (id: number, branch = "") => {
      setLoading(true);
      marsConfig({ branch, project_id: id })
        .then(({ data }) => {
          if (data.config) {
            setConfig(cloneDeep(data.config));
            console.log("oldoldoldold", data.config);
            setOld(cloneDeep(data.config));
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
    [loadDefaultValues]
  );

  useEffect(() => {
    if (visible && config) {
      setMode(getMode(config.config_file_type));
    }
  }, [config, visible]);

  const loadGlobalConfig = useCallback(
    (id: number) => {
      setLoading(true);
      globalConfigApi({ project_id: id })
        .then(({ data }) => {
          if (data.config) {
            setConfig(cloneDeep(data.config));
            console.log("oldoldoldold", data.config);
            setOld(cloneDeep(data.config));
          }
          loadDefaultValues(id, "");
          setLoading(false);
        })
        .catch((e) => {
          message.error(e.response.data.message);
          setLoading(false);
        });
    },
    [loadDefaultValues]
  );

  useEffect(() => {
    setConfigVisible(visible);
    if (item && visible) {
      setLoading(true);
      branches({ project_id: String(item.id), all: true }).then((res) => {
        setMbranches(res.data.data.map((op) => op.value));
      });
      globalConfigApi({ project_id: item.id })
        .then(({ data }) => {
          setGlobalEnabled(data.enabled);
          if (!data.enabled) {
            loadConfig(item.id);
          } else {
            console.log(data.config, !!data.config);
            if (data.config) {
              console.log("oldoldoldold", data.config);
              setOld(cloneDeep(data.config));
              setConfig(cloneDeep(data.config));
            }
            loadDefaultValues(item.id, "");
            setLoading(false);
          }
        })
        .catch((e) => {
          message.error(e.response.data.message);
        });
    }
  }, [item, loadConfig, visible, loadDefaultValues]);

  const resetModal = useCallback(() => {
    setMbranches([]);
    setLoading(false);
    setConfig({...initConfig});
    setConfigVisible(false);
    setEditMode(true);
    setConfigFileContent("");
    setConfigFileTip(false);
    onCancel();
  }, [onCancel]);

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
      item &&
        toggleGlobalEnabledApi({ project_id: item.id, enabled })
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
    [loadGlobalConfig, loadConfig, item]
  );

  useEffect(() => {
    if (editMode && visible) {
      let d = debounce(() => {
        console.log("debounce called");
        if (config.config_field && defaultValues) {
          let data = get(
            yaml.load(defaultValues),
            config.config_field.split("->"),
            ""
          );
          if (typeof data === "object") {
            data = yaml.dump(data);
          }
          setConfigFileContent(data);
        } else {
          setConfigFileContent("");
        }
      }, 1000);
      d();
      return () => {
        d.cancel();
        console.log("cancel called");
      };
    }
  }, [editMode, config.config_field, defaultValues, visible]);

  const onSave = () => {
    item &&
      updateGlobalConfig({
        project_id: item.id,
        config: config,
      })
        .then((res) => {
          message.success("保存成功");
          res.data.config &&
            setConfig((c) => {
              let a = { ...c, ...res.data.config };
              setOld(cloneDeep(a));

              return a;
            });
          loadDefaultValues(item.id, "");
          setEditMode(false);
        })
        .catch((e) => {
          message.error(e.response.data.message);
          globalConfigApi({ project_id: item.id }).then(({ data }) => {
            setGlobalEnabled(data.enabled);
            data.config && setOld(cloneDeep(data.config));
            console.log(data.config);
            // setConfig(res.config);
          });
        });
  };

  const [configFileContent, setConfigFileContent] = useState("");
  const [configFileTip, setConfigFileTip] = useState(false);
  useEffect(() => {
    setConfigFileTip(!!configFileContent && !config.config_file_values);
  }, [configFileContent, config.config_file_values]);

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
      {item ? (
        <>
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
                        console.log("old", old);
                        old && setConfig(cloneDeep(old));
                      }
                      console.log(old, config);
                      return !editMode;
                    });
                  }}
                >
                  {!editMode ? "编辑" : "取消"}
                </Button>
                {editMode ? (
                  <Button type="primary" onClick={onSave}>
                    保存
                  </Button>
                ) : (
                  <></>
                )}
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
              {loading ? (
                <Skeleton active />
              ) : (
                <Row
                  gutter={[16, 16]}
                  style={{ height: "100%", overflowY: "auto" }}
                >
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
                            onCopy={() => message.success("已复制！")}
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
                    span={12}
                    style={{
                      maxHeight: "800px",
                      overflowY: "scroll",
                      position: "relative",
                    }}
                  >
                    <Form name="basic" layout="vertical" autoComplete="off">
                      <Form.Item
                        label="charts 的目录"
                        tooltip="charts 文件在项目中存放的目录(必填), 也可以是别的项目的文件，格式为 'pid|branch|path'"
                      >
                        <Input
                          disabled={!editMode || !globalEnabled}
                          value={config.local_chart_path}
                          onChange={(
                            v: React.ChangeEvent<HTMLInputElement>
                          ) => {
                            setConfig((c) => ({
                              ...c,
                              local_chart_path: v.target.value,
                            }));
                          }}
                        />
                      </Form.Item>

                      <Form.Item style={{ marginBottom: 0 }}>
                        <Form.Item
                          label="用户输入配置字段"
                          tooltip={`用户在部署时使用的自定义配置字段, 比如 "conf->config"`}
                          style={{
                            display: "inline-block",
                            width: "calc(50% - 8px)",
                            marginRight: 8,
                          }}
                        >
                          <Popover
                            overlayInnerStyle={{
                              maxHeight: 600,
                              overflowY: "scroll",
                            }}
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
                                    setConfig((c) => ({
                                      ...c,
                                      config_file_values: configFileContent,
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
                          <Input
                            disabled={!editMode || !globalEnabled}
                            value={config.config_field}
                            onChange={(
                              v: React.ChangeEvent<HTMLInputElement>
                            ) => {
                              setConfig((c) => ({
                                ...c,
                                config_field: v.target.value,
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
                        >
                          <SelectFileType
                            showArrow={editMode}
                            disabled={!editMode || !globalEnabled}
                            value={config.config_file_type}
                            onChange={(value: any) => {
                              setConfig((c) => ({
                                ...c,
                                config_file_type: value,
                              }));
                            }}
                          />
                        </Form.Item>
                      </Form.Item>

                      <Form.Item
                        label="动态配置"
                        style={{
                          width: "100%",
                        }}
                      >
                        <DynamicElement
                          disabled={!editMode || !globalEnabled}
                          value={config.elements ? config.elements : []}
                          onChange={(v: pb.Element[]) => {
                            console.log("duccc", v);
                            setConfig((c) => ({
                              ...c,
                              elements: v,
                            }));
                          }}
                        />
                      </Form.Item>

                      <Form.Item style={{ marginBottom: 0 }}>
                        <Form.Item
                          label="单字段"
                          tooltip="配置文件是不是一个整体的value值"
                          style={{
                            display: "inline-block",
                            width: "calc(15% - 8px)",
                          }}
                        >
                          <Switch
                            disabled={!editMode || !globalEnabled}
                            defaultChecked
                            checked={config.is_simple_env}
                            onChange={(checked: boolean, event: MouseEvent) => {
                              setConfig((c) => ({
                                ...c,
                                is_simple_env: checked,
                              }));
                            }}
                          />
                        </Form.Item>
                        <Form.Item
                          label="启用的分支"
                          style={{
                            display: "inline-block",
                            width: "calc(85% - 8px)",
                          }}
                        >
                          <Select
                            disabled={!editMode || !globalEnabled}
                            mode="multiple"
                            style={{ width: "100%" }}
                            value={config.branches}
                            onChange={(v: string[]) => {
                              console.log(v);
                              setConfig((c) => ({ ...c, branches: v }));
                            }}
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
                      <Form.Item
                        label="全局配置文件"
                        tooltip="全局默认配置文件，如果没有设置 config_file 则使用这个"
                      >
                        <div style={{ maxHeight: "400px", overflowY: "auto" }}>
                          <CodeMirror
                            options={{
                              readOnly:
                                !editMode || !globalEnabled
                                  ? "nocursor"
                                  : false,
                              mode: mode,
                              theme: "dracula",
                            }}
                            value={config.config_file_values || ""}
                            onBeforeChange={(editor, d, value) => {
                              setConfig((c) => ({
                                ...c,
                                config_file_values: value,
                              }));
                            }}
                          />
                        </div>
                      </Form.Item>

                      <Form.Item
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
                        tooltip="等同于 helm 的 values.yaml"
                      >
                        <div
                          style={{
                            maxHeight: "800px",
                            overflowY: "scroll",
                            position: "relative",
                            fontSize: 13,
                          }}
                        >
                          <CodeMirror
                            options={{
                              readOnly:
                                !editMode || !globalEnabled
                                  ? "nocursor"
                                  : false,
                              mode: getMode("yaml"),
                              theme: "dracula",
                            }}
                            value={config.values_yaml}
                            onBeforeChange={(editor, d, value) => {
                              setConfig((c) => ({
                                ...c,
                                values_yaml: value,
                              }));
                            }}
                          />
                        </div>
                      </Form.Item>
                    </Form>
                  </Col>
                </Row>
              )}
            </Col>
          </Row>
        </>
      ) : (
        <Skeleton active />
      )}
    </Modal>
  );
};

export default memo(ConfigModal);
