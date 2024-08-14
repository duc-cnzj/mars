import React, { memo, useCallback, useEffect, useState } from "react";
import {
  Button,
  Card,
  Divider,
  Form,
  Input,
  message,
  Modal,
  Popconfirm,
  Popover,
  Spin,
} from "antd";
import theme from "../styles/theme";
import { List } from "antd";
import { css } from "@emotion/css";
import InfiniteScroll from "react-infinite-scroll-component";
import { components } from "../api/schema";
import ajax from "../api/ajax";
import AddRepoModal from "../components/AddRepoModal";
import { copy } from "../utils/copy";
import { CopyOutlined } from "@ant-design/icons";

const defaultPageSize = 15;

const RepoPage: React.FC = () => {
  const [data, setData] = useState<components["schemas"]["types.RepoModel"][]>(
    []
  );
  const [paginate, setPaginate] = useState({
    page: 1,
    pageSize: defaultPageSize,
  });
  const [loading, setLoading] = useState<{
    list: boolean;
    items: { [name: number]: boolean };
  }>({
    list: false,
    items: {},
  });

  const [visible, setVisible] = useState({
    add: false,
    clone: false,
  });
  const [editItem, setEditItem] =
    useState<components["schemas"]["types.RepoModel"]>();

  const fetch = useCallback(
    (page: number, pageSize: number, name?: string, refresh?: boolean) => {
      setLoading((v) => ({ ...v, list: true }));
      ajax
        .GET("/api/repos", {
          params: { query: { page, pageSize, name: name } },
        })
        .then(({ data }) => {
          data &&
            setData((v) => (refresh ? data.items : [...v, ...data.items]));
          data && setPaginate({ page: data.page, pageSize: data.pageSize });
        })
        .finally(() => {
          setLoading((v) => ({ ...v, list: false }));
        });
    },
    []
  );

  const [searchInput, setSearchInput] = useState({ name: "" });

  useEffect(() => {
    fetch(1, defaultPageSize, "");
  }, [fetch]);

  const [cloneInput, setCloneInput] = useState<
    components["schemas"]["repo.CloneRequest"]
  >({ id: 0, name: "" });

  return (
    <>
      <AddRepoModal
        editItem={editItem}
        visible={visible.add}
        onSuccess={() => {
          setVisible((v) => ({ ...v, add: false }));
          fetch(1, defaultPageSize, searchInput.name, true);
          setEditItem(undefined);
        }}
        onCancel={() => {
          console.log("cancel");
          setEditItem(undefined);
          setVisible((v) => ({ ...v, add: false }));
        }}
      />
      <Modal
        title="克隆项目"
        open={visible.clone}
        okText={"确定"}
        cancelText={"取消"}
        onOk={() => {
          if (cloneInput.id <= 0 || !cloneInput.name) {
            message.error("克隆数据不能为空");
            return;
          }
          ajax
            .POST("/api/repos/clone", { body: cloneInput })
            .then(({ error }) => {
              if (error) {
                message.error(error.message);
                return;
              }
              message.success("克隆成功");
              setVisible((v) => ({ ...v, clone: false }));
              fetch(1, defaultPageSize, searchInput.name, true);
            });
        }}
        onCancel={() => {
          setVisible((v) => ({ ...v, clone: false }));
          setCloneInput({ id: 0, name: "" });
        }}
      >
        <Input
          placeholder="请输入 clone 后的项目名称"
          value={cloneInput.name}
          onChange={(v) => {
            setCloneInput((input) => ({ ...input, name: v.target.value }));
          }}
        />
      </Modal>
      <Card
        className="git"
        title={
          <div style={{ display: "flex", justifyContent: "space-between" }}>
            <div>Repo 管理</div>
            <Button
              style={{ fontSize: 12 }}
              onClick={() => setVisible((v) => ({ ...v, add: true }))}
              type="default"
              size="small"
            >
              添加 repo
            </Button>
          </div>
        }
        styles={{ body: { padding: 0 } }}
        style={{ marginTop: 20, marginBottom: 30 }}
      >
        <div>
          <Input
            value={searchInput.name}
            allowClear
            onKeyDown={(k) => {
              if (k.code === "Enter") {
                fetch(1, defaultPageSize, searchInput.name, true);
              }
            }}
            onChange={(v) => {
              setSearchInput((inp) => ({ ...inp, name: v.target.value }));
              if (v.target.value === "") {
                fetch(1, defaultPageSize, "", true);
              }
            }}
            style={{ width: "40%", marginTop: 20, marginLeft: 20 }}
            placeholder="搜索 repo 名称..."
          />
        </div>
        <Divider />
        <InfiniteScroll
          dataLength={data.length}
          next={() => {
            fetch(paginate.page + 1, paginate.pageSize, searchInput.name);
          }}
          hasMore={data.length !== 0 && data.length % paginate.pageSize === 0}
          loader={
            <Spin
              style={{
                margin: "20px 0",
                display: "flex",
                justifyContent: "center",
              }}
              size="large"
            />
          }
          endMessage={<Divider plain>老铁，别翻了，到底了！</Divider>}
          scrollableTarget="scrollableDiv"
        >
          <List
            itemLayout="horizontal"
            loading={loading.list}
            dataSource={data}
            renderItem={(item: components["schemas"]["types.RepoModel"]) => (
              <List.Item
                className={css`
                  padding: 14px 24px !important;
                  &:hover {
                    background-image: ${theme.lightLinear};
                  }
                `}
                key={item.id}
                actions={[
                  <Button
                    type="dashed"
                    onClick={() => {
                      setCloneInput((v) => ({ ...v, id: item.id }));
                      setVisible((v) => ({ ...v, clone: true }));
                    }}
                  >
                    克隆
                  </Button>,
                  <Button
                    onClick={async () => {
                      const { data, error } = await ajax.GET(
                        "/api/repos/{id}",
                        { params: { path: { id: item.id } } }
                      );
                      if (error) {
                        message.error(error.message);
                        return;
                      }
                      setEditItem(data.item);
                      setVisible((v) => ({ ...v, add: true }));
                    }}
                  >
                    编辑配置
                  </Button>,
                  <Button
                    danger={item.enabled}
                    loading={!!loading.items[item.id]}
                    type={!item.enabled ? "primary" : "dashed"}
                    onClick={async () => {
                      setLoading((lo) => ({
                        ...lo,
                        items: { ...lo.items, [item.id]: true },
                      }));
                      const { error } = await ajax.POST(
                        "/api/repos/toggle_enabled",
                        {
                          body: { id: item.id, enabled: !item.enabled },
                        }
                      );
                      if (error) {
                        setTimeout(() => {
                          setLoading((lo) => ({
                            ...lo,
                            items: { ...lo.items, [item.id]: false },
                          }));
                        }, 1000);
                        message.error(error.message);
                        return;
                      }
                      setTimeout(() => {
                        setData((items) =>
                          items.map((v) =>
                            v.id === item.id
                              ? { ...v, enabled: !item.enabled }
                              : v
                          )
                        );
                        setLoading((lo) => ({
                          ...lo,
                          items: { ...lo.items, [item.id]: false },
                        }));
                        message.success("操作成功");
                      }, 1000);
                    }}
                  >
                    {item.enabled ? "禁用" : "启用"}
                  </Button>,
                  <Popconfirm
                    title={`确定删除 ${item.name} ?`}
                    description="下面有关联 project 时无法删除"
                    onConfirm={() => {
                      ajax
                        .DELETE("/api/repos/{id}", {
                          params: { path: { id: item.id } },
                        })
                        .then(({ error }) => {
                          if (error) {
                            message.error(error.message);
                            return;
                          }
                          message.success("删除成功");
                          fetch(1, defaultPageSize, searchInput.name, true);
                        });
                    }}
                    okText="确定"
                    cancelText="取消"
                  >
                    <Button danger>删除 repo</Button>
                  </Popconfirm>,
                ]}
              >
                <List.Item.Meta
                  key={item.id}
                  title={
                    <div>
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
                        <span
                          style={{ cursor: "pointer" }}
                          onClick={() =>
                            copy(String(item.id), "已复制项目id！")
                          }
                        >
                          <CopyOutlined />
                        </span>
                        )
                      </div>
                    </div>
                  }
                  description={
                    <div>
                      {!!item.gitProjectId ? (
                        <div>
                          已关联 git 项目: {item.gitProjectName}, git 项目 id:{" "}
                          {item.gitProjectId}
                        </div>
                      ) : (
                        "无 git 项目关联"
                      )}
                    </div>
                  }
                />
              </List.Item>
            )}
          />
        </InfiniteScroll>
      </Card>
    </>
  );
};

export default memo(RepoPage);
