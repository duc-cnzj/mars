import React, { useState, useEffect, useCallback } from "react";
import {
  disabledProject,
  enabledProject,
  projectList,
} from "../api/gitlab";
import { List, Avatar, Card, Button, Select, message, Tooltip } from "antd";
import ConfigModal from "./ConfigModal";
import { GlobalOutlined } from "@ant-design/icons";
import pb from '../api/compiled'

const { Option } = Select;
const GitlabProjectManager: React.FC = () => {
  const [list, setList] = useState<pb.GitlabProjectInfo[]>([]);
  const [initLoading, setInitLoading] = useState(true);
  const [loadingList, setLoadingList] = useState<{ [name: number]: boolean }>();

  const fetchList = useCallback(() => {
    return projectList().then((res) => {
      setList(res.data.data);
    }).catch((e) => message.error(e.response.data.message));
  }, [setList]);

  useEffect(() => {
    fetchList().then(() => {
      setInitLoading(false);
    });
  }, [fetchList, setInitLoading]);

  const toggleStatus = async (item: pb.GitlabProjectInfo) => {
    setLoadingList((l) => ({ ...l, [item.id]: true }));
    console.log("loadingList", loadingList);
    try {
      if (item.enabled) {
        await disabledProject({gitlab_project_id: String(item.id)});
      } else {
        await enabledProject({gitlab_project_id: String(item.id)});
      }
    } catch (e: any) {
      message.error(e.response.data.message);
      setLoadingList((l) => ({ ...l, [item.id]: false }));
      return;
    }

    fetchList().then((res) => {
      setLoadingList((l) => ({ ...l, [item.id]: false }));
      message.success("操作成功");
    });
  };

  const [currentItem, setCurrentItem] = useState<pb.GitlabProjectInfo>();
  const [configVisible, setConfigVisible] = useState(false);
  const [selected, setSelected] = useState<pb.GitlabProjectInfo>();

  const onChange = (v: any) => {
    console.log(v);
    if (!v) {
      setSelected(undefined);
      return;
    }
    let item = list.find((item) => item.id === v);
    if (item) {
      setSelected(item);
      console.log(item)
    }
  };

  return (
    <>
      <Card
        title={"gitlab项目管理"}
        style={{ marginTop: 20, marginBottom: 30 }}
      >
        <div>
          <Select
            showSearch
            allowClear
            style={{ width: 500, marginBottom: 10 }}
            placeholder="搜索项目"
            optionFilterProp="children"
            onChange={onChange}
            filterOption={(input, option: any) =>
              option.children.toLowerCase().indexOf(input.toLowerCase()) >= 0
            }
          >
            {list ? list.map((item, key) => (
              <Option value={item.id} key={key}>{item.name}</Option>
            )):<></>}
          </Select>
        </div>
        <List
          itemLayout="horizontal"
          loading={initLoading}
          dataSource={list.filter(item=> selected ? item.id === selected.id : true)}
          renderItem={(item: pb.GitlabProjectInfo) => (
            <List.Item
              key={item.id}
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
                key={item.id}
                avatar={<Avatar src={item.avatar_url} />}
                title={
                  <div>
                    {item.name}
                    {item.global_enabled ? (
                      <Tooltip
                        placement="top"
                        title="已使用全局配置"
                        overlayStyle={{ fontSize: "10px" }}
                      >
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
        />
      </Card>
    </>
  );
};

export default GitlabProjectManager;
