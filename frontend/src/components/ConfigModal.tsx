import React, { useState, useCallback, useEffect, useRef } from "react";
import { Controlled as CodeMirror } from "react-codemirror2";

import pb from '../api/compiled'

import "codemirror/lib/codemirror.css";
import {Prism as SyntaxHighlighter} from "react-syntax-highlighter";
import { materialDark } from "react-syntax-highlighter/dist/esm/styles/prism";
import {
  Tooltip,
  Switch,
  Select,
  Button,
  message,
  Modal,
  Skeleton,
  Spin,
  Empty,
} from "antd";
import { QuestionCircleOutlined, EditOutlined } from "@ant-design/icons";
import {
  globalConfig as globalConfigApi,
  marsConfig,
  toggleGlobalEnabled as toggleGlobalEnabledApi,
  updateGlobalConfig,
} from "../api/mars";
import { branches } from "../api/gitlab";

import "codemirror/theme/mdn-like.css";
import MarsExample from "./MarsExample";
require("codemirror/mode/yaml/yaml");

const { Option } = Select;

const ConfigModal: React.FC<{
  visible: boolean;
  item: undefined | pb.GitlabProjectInfo;
  onCancel: () => void;
  onChange?: () => void;
}> = ({ visible, item, onCancel, onChange }) => {
  const [editMode, setEditMode] = useState(false);
  const [globalEnabled, setGlobalEnabled] = useState(false);
  const [globalConfig, setGlobalConfig] = useState<string>();
  const [modalBranch, setModalBranch] = useState("");
  const [currentItem, setCurrentItem] = useState<pb.GitlabProjectInfo | undefined>(item);
  const [title, setTitle] = useState("");
  const [config, setConfig] = useState<string>();
  const [configVisible, setConfigVisible] = useState(visible);
  const [mbranches, setMbranches] = useState<string[]>([]);
  const [loading, setLoading] = useState(false);

  const loadConfig = useCallback((id: number, branch = "") => {
    setLoading(true);
    marsConfig({branch, project_id: id})
      .then((res) => {
        setConfig(res.config);
        setModalBranch(res.branch);
        setLoading(false);
      })
      .catch((e) => {
        message.error(e.response.data.message);
        setLoading(false);
      });
  }, []);

  useEffect(() => {
    setConfigVisible(visible);
    if (visible && item) {
      console.log(item);
      setCurrentItem(item);
      branches({project_id: String(item.id)}).then((res) =>
        setMbranches(res.data.data.map((op) => op.value))
      );
      globalConfigApi({project_id: item.id}).then((res) => {
        setGlobalEnabled(res.enabled);
        console.log(res.config);
        setGlobalConfig(res.config);
      });
      setTitle(item.name);
      loadConfig(item.id);
    }
  }, [visible, item, loadConfig]);

  const resetModal = useCallback(() => {
    setTitle("");
    setModalBranch("");
    setCurrentItem(undefined);
    setMbranches([]);
    setLoading(false);
    setConfig(undefined);
    setConfigVisible(false);
    onCancel();
  }, [onCancel]);

  const selectBranch = (value: string) => {
    if (currentItem) {
      loadConfig(currentItem.id, value);
    }
  };

  const toggleGlobalEnabled = (enabled: boolean) => {
    setLoading(true);
    if (!enabled) {
      setEditMode(false);
    }
    currentItem &&
      toggleGlobalEnabledApi({project_id: currentItem.id, enabled}).then(() => {
        message.success("操作成功");
        onChange?.();
        setGlobalEnabled(enabled);
        globalConfigApi({project_id: currentItem.id}).then((res) => {
          setGlobalEnabled(res.enabled);
          console.log(res.config);
          setGlobalConfig(res.config);
        });
        branches({project_id: String(currentItem.id)}).then((res) =>
          setMbranches(res.data.data.map((op) => op.value))
        );
        marsConfig({project_id: currentItem.id, branch: ""})
          .then((res) => {
            setConfig(res.config);
            setModalBranch(res.branch);
            setLoading(false);
          })
          .catch((e) => {
            message.error(e.response.data.message);
            setLoading(false);
          });
      }).catch((e) => message.error(e.message));
  };
  const onSave = () => {
    currentItem &&
      updateGlobalConfig({project_id: currentItem.id, config: globalConfig || ""})
        .then((res) => {
          message.success("保存成功");
          console.log(res.data.data.global_config);
          setGlobalConfig(res.data.data.global_config);
          setEditMode(false);
        })
        .catch((e) => {
          message.error(e.response.data.message);
          globalConfigApi({project_id: currentItem.id}).then((res) => {
            setGlobalEnabled(res.enabled);
            console.log(res.config);
            // setGlobalConfig(res.config);
          });
        });
  };
  const cmref = useRef<any>();

  return (
    <Modal
      title={<div><MarsExample/>&nbsp;&nbsp;{title}</div>}
      visible={configVisible}
      footer={null}
      width={800}
      onCancel={resetModal}
    >
      {modalBranch ? (
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
                  onClick={() => setEditMode(!editMode)}
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
                defaultChecked={globalEnabled}
                onChange={toggleGlobalEnabled}
              />
            </div>
          </div>
          {editMode ? (
            <CodeMirror
              ref={cmref}
              value={globalConfig || ""}
              options={{
                mode: "yaml",
                theme: "mdn-like",
                lineNumbers: true,
              }}
              onBeforeChange={(editor, d, value) => {
                console.log("valuevalue", globalConfig, value);
                setGlobalConfig(value);
              }}
            />
          ) : (
            <Spin spinning={loading}>
              {!globalEnabled && !config ? (
                <Empty
                  description="未发现该项目的配置文件"
                  style={{ height: 220 }}
                />
              ) : (
                <SyntaxHighlighter
                  language="yaml"
                  style={materialDark}
                  customStyle={{
                    minHeight: 200,
                    lineHeight: 1.2,
                    padding: "10px",
                    fontFamily: '"Fira code", "Fira Mono", monospace',
                    fontSize: 13,
                  }}
                >
                  {globalEnabled ? globalConfig : config}
                </SyntaxHighlighter>
              )}
            </Spin>
          )}
        </>
      ) : (
        <Skeleton active />
      )}
    </Modal>
  );
};

export default ConfigModal;
