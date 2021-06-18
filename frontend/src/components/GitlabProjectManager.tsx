import React, { useState, useEffect, useCallback } from "react";
import {
  disabledProject,
  enabledProject,
  Info,
  projectList,
} from "../api/gitlab";
import { List, Avatar, Card, Button, message } from "antd";

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

  const toggleStatus = (item: Info) => {
    setLoadingList({ ...loadingList, [item.id]: true });
    console.log(loadingList);
    if (item.enabled) {
      disabledProject(item.id)
        .then((res) => {
          message.success("成功");
          setTimeout(() => {
            setLoadingList({ ...loadingList, [item.id]: false });
          }, 1500);
        })
        .catch((e) => {
          message.error(e.message);
          setTimeout(() => {
            setLoadingList({ ...loadingList, [item.id]: false });
          }, 1500);
        });
    } else {
      enabledProject(item.id)
        .then((res) => {
          message.success("成功");
          setTimeout(() => {
            setLoadingList({ ...loadingList, [item.id]: false });
          }, 500);
        })
        .catch((e) => {
          message.error(e.response.data.message);
          setTimeout(() => {
            setLoadingList({ ...loadingList, [item.id]: false });
          }, 500);
        });
    }
    fetchList();
  };

  return (
    <>
      <Card title={"manager"} style={{ marginTop: 20, marginBottom: 30 }}>
        <List
          itemLayout="horizontal"
          loading={initLoading}
          dataSource={list}
          renderItem={(item: Info) => (
            <List.Item
              actions={[
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
      </Card>
    </>
  );
};

export default GitlabProjectManager;
