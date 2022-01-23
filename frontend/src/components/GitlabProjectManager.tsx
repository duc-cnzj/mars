import React, { useState, useEffect, useCallback, memo } from "react";
import { disabledProject, enabledProject, allProjects } from "../api/gitlab";
import { CopyOutlined } from "@ant-design/icons";
import { CopyToClipboard } from "react-copy-to-clipboard";
import {
  List,
  Avatar,
  Card,
  Button,
  Select,
  message,
  Tooltip,
  Divider,
} from "antd";
import ConfigModal from "./ConfigModal";
import { GlobalOutlined } from "@ant-design/icons";
import pb from "../api/compiled";

const { Option } = Select;
const GitlabProjectManager: React.FC = () => {
  const [list, setList] = useState<pb.GitProjectItem[]>([]);
  const [initLoading, setInitLoading] = useState(true);
  const [loadingList, setLoadingList] = useState<{ [name: number]: boolean }>();

  const fetchList = useCallback(() => {
    return allProjects()
      .then((res) => {
        setList(res.data.data);
      })
      .catch((e) => message.error(e.response.data.message));
  }, [setList]);

  useEffect(() => {
    fetchList().then(() => {
      setInitLoading(false);
    });
  }, [fetchList, setInitLoading]);

  const toggleStatus = async (item: pb.GitProjectItem) => {
    setLoadingList((l) => ({ ...l, [item.id]: true }));
    try {
      if (item.enabled) {
        await disabledProject({ git_project_id: String(item.id) });
      } else {
        await enabledProject({ git_project_id: String(item.id) });
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

  const [currentItem, setCurrentItem] = useState<pb.GitProjectItem>();
  const [configVisible, setConfigVisible] = useState(false);
  const [selected, setSelected] = useState<pb.GitProjectItem>();

  const onChange = useCallback(
    (v: any) => {
      if (!v) {
        setSelected(undefined);
        return;
      }
      let item = list.find((item) => item.id === v);
      if (item) {
        setSelected(item);
      }
    },
    [list]
  );

  return (
    <>
      <Card
        className="gitlab"
        title={"gitlab项目管理"}
        style={{ marginTop: 20, marginBottom: 30 }}
        bodyStyle={{ padding: 0 }}
      >
        <div style={{ padding: "24px 24px 0 24px" }}>
          <Select
            showSearch
            allowClear
            style={{ width: 500 }}
            placeholder="搜索项目"
            optionFilterProp="children"
            onChange={onChange}
            filterOption={(input, option: any) =>
              option.children.toLowerCase().indexOf(input.toLowerCase()) >= 0
            }
          >
            {list ? (
              list.map((item, key) => (
                <Option value={item.id} key={key}>
                  {item.name}
                </Option>
              ))
            ) : (
              <></>
            )}
          </Select>
        </div>
        <Divider />
        <List
          itemLayout="horizontal"
          loading={initLoading}
          dataSource={list.filter((item) =>
            selected ? item.id === selected.id : true
          )}
          renderItem={(item: pb.GitProjectItem) => (
            <List.Item
              className="gitlab__list-item"
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
                  <div style={{ fontSize: 16 }}>
                    {item.name}
                    <div
                      style={{
                        display: "inline-block",
                        fontSize: 10,
                        marginLeft: 3,
                        marginRight: 1,
                        fontWeight: "normal",
                      }}
                    >
                      (id: <span style={{ marginRight: 1 }}>{item.id}</span>
                      <CopyToClipboard
                        text={String(item.id)}
                        onCopy={() => message.success("已复制项目id！")}
                      >
                        <CopyOutlined />
                      </CopyToClipboard>
                      )
                    </div>
                    {item.global_enabled ? (
                      <Tooltip
                        placement="top"
                        title="已使用全局配置"
                        overlayStyle={{ fontSize: "10px" }}
                      >
                        <GlobalOutlined
                          style={{
                            color: item.enabled ? "green" : "red",
                            marginLeft: 3,
                          }}
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
        {configVisible ? (
          <ConfigModal
            visible={configVisible}
            item={currentItem}
            onCancel={() => setConfigVisible(false)}
          />
        ) : (
          <></>
        )}
      </Card>
    </>
  );
};

export default memo(GitlabProjectManager);
