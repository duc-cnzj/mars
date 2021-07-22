import React, { useState, useEffect, useCallback } from "react";
import {
  disabledProject,
  enabledProject,
  Info,
  projectList,
} from "../api/gitlab";
import { List, Avatar, Card, Button, message, Tooltip } from "antd";
import ConfigModal from "./ConfigModal";
import { GlobalOutlined } from "@ant-design/icons";

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
      return;
    }

    fetchList().then((res) => {
      message.success("操作成功");
      setLoadingList((l) => ({ ...l, [item.id]: false }));
    });
  };

  const [currentItem, setCurrentItem] = useState<Info>();
  const [configVisible, setConfigVisible] = useState(false);

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
                title={
                  <div>
                    {item.name}{" "}
                    {item.global_enabled ? (
                        <Tooltip placement="top" title="已使用全局配置" overlayStyle={{fontSize: "10px"}}>
                          <GlobalOutlined
                            style={{ color: item.enabled ? "green" : "red" }}
                          />
                      </Tooltip>
                    ) : (
                      <></>
                    )}
                  </div>
                }
                description={
                  item.description ? item.description : "该项目还没有描述信息哦"
                }
              />
            </List.Item>
          )}
        />
        <ConfigModal
          visible={configVisible}
          item={currentItem}
          onCancel={() => setConfigVisible(false)}
          onChange={()=>fetchList()}
        />
      </Card>
    </>
  );
};

export default GitlabProjectManager;
