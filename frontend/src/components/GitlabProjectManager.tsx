import React, { useState, useEffect, useCallback } from "react";
import yaml from "js-yaml";
import SyntaxHighlighter from "react-syntax-highlighter";
import { monokaiSublime } from "react-syntax-highlighter/dist/esm/styles/hljs";
import {
  branches,
  disabledProject,
  enabledProject,
  Info,
  projectList,
} from "../api/gitlab";
import {
  List,
  Select,
  Avatar,
  Card,
  Button,
  message,
  Modal,
  Skeleton,
  Spin,
} from "antd";
import { marsConfig } from "../api/mars";

const { Option } = Select;
const GitlabProjectManager: React.FC = () => {
  const [list, setList] = useState<Info[]>([]);

  const [initLoading, setInitLoading] = useState(true);

  const [loadingList, setLoadingList] = useState<{ [name: number]: boolean }>();

  const fetchList = useCallback(() => {
    return projectList().then((res) => {
      setList(res.data.data);
    });
  }, [setList]);

  useEffect(() => {
    fetchList().then(() => {
      setInitLoading(false);
    });
  }, [fetchList, setInitLoading]);

  const toggleStatus = async (item: Info) => {
    setLoadingList((l) => ({ ...l, [item.id]: true }));
    console.log("loadingList", loadingList);
    try {
      if (item.enabled) {
        await disabledProject(item.id);
      } else {
        await enabledProject(item.id);
      }
    } catch (e) {
      message.error(e.response.data.message);
      setLoadingList((l) => ({ ...l, [item.id]: false }));
      return
    }

    fetchList().then((res) => {
      message.success("操作成功");
      setLoadingList((l) => ({ ...l, [item.id]: false }));
    });
  };

  const [modalBranch, setModalBranch] = useState("");
  const [currentItem, setCurrentItem] = useState<Info>();
  const [title, setTitle] = useState("");
  const [config, setConfig] = useState<API.Mars>();
  const [configVisible, setConfigVisible] = useState(false);
  const [mbranches, setMbranches] = useState<string[]>([]);
  const [loading, setLoading] = useState(false)

  const loadConfig = (id: number, branch = "") => {
    setLoading(true)
    marsConfig(id, { branch }).then((res) => {
      setConfig(res.config);
      setModalBranch(res.branch);
      setLoading(false)
    }).catch(e=>{
      message.error(e.response.data.message)
      setLoading(false)
    });
  };

  const resetModal = () => {
    setTitle("");
    setModalBranch("")
    setCurrentItem(undefined)
    setMbranches([])
    setLoading(false)
    setConfig(undefined);
    setConfigVisible(false);
  };

  const selectBranch = (value: string) => {
    if (currentItem) {
      loadConfig(currentItem.id, value);
    }
  };

  return (
    <>
      <Card
        title={"gitlab项目管理"}
        style={{ marginTop: 20, marginBottom: 30 }}
      >
        <List
          itemLayout="horizontal"
          loading={initLoading}
          dataSource={list}
          renderItem={(item: Info) => (
            <List.Item
              actions={[
                item.enabled ? (
                  <Button
                    onClick={() => {
                      setCurrentItem(item);
                      setConfigVisible(true);
                      branches(item.id).then((res) =>
                        setMbranches(res.data.data.map((item) => item.value))
                      );
                      loadConfig(item.id);
                    }}
                  >
                    查看配置
                  </Button>
                ) : (
                  <></>
                ),
                <Button
                  danger={item.enabled}
                  loading={loadingList && loadingList[item.id]}
                  type={!item.enabled ? "primary" : "ghost"}
                  onClick={() => toggleStatus(item)}
                >
                  {item.enabled ? "关闭" : "开启"}
                </Button>,
              ]}
            >
              <List.Item.Meta
                avatar={<Avatar src={item.avatar_url} />}
                title={item.name}
                description={
                  item.description ? item.description : "该项目还没有描述信息哦"
                }
              />
            </List.Item>
          )}
        />
        <Modal
          title={title}
          visible={configVisible}
          footer={null}
          width={800}
          onCancel={resetModal}
        >
          {modalBranch ? (
            <>
              <Select
                placeholder="请选择"
                value={modalBranch}
                style={{ width: 250, marginBottom: 10 }}
                onChange={selectBranch}
              >
                {mbranches.map((item) => (
                  <Option value={item} key={item}>
                    {item}
                  </Option>
                ))}
              </Select>
              <Spin spinning={loading}>
                <SyntaxHighlighter
                  language="yaml"
                  style={monokaiSublime}
                  customStyle={{
                    minHeight: 200,
                    lineHeight: 1.2,
                    padding: "10px",
                    fontFamily: '"Fira code", "Fira Mono", monospace',
                    fontSize: 15,
                  }}
                >
                  {yaml.dump(config)}
                </SyntaxHighlighter>
              </Spin>
            </>
          ) : (
            <Skeleton active />
          )}
        </Modal>
      </Card>
    </>
  );
};

export default GitlabProjectManager;
