import React, { memo, useCallback, useEffect, useState } from "react";
import { Button, Card, Divider, message, Spin } from "antd";
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

  const [visible, setVisible] = useState(false);
  const [editItem, setEditItem] =
    useState<components["schemas"]["types.RepoModel"]>();

  const fetch = useCallback(
    (page: number, pageSize: number, refresh?: boolean) => {
      setLoading((v) => ({ ...v, list: true }));
      ajax
        .GET("/api/repos", { params: { query: { page, pageSize } } })
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

  useEffect(() => {
    fetch(1, defaultPageSize);
  }, [fetch]);

  return (
    <>
      <AddRepoModal
        editItem={editItem}
        visible={visible}
        onSuccess={() => {
          setVisible(false);
          fetch(1, defaultPageSize, true);
          setEditItem(undefined);
        }}
        onCancel={() => {
          console.log("cancel");
          setEditItem(undefined);
          setVisible(false);
        }}
      />
      <Card
        className="git"
        title={
          <div style={{ display: "flex", justifyContent: "space-between" }}>
            <div>Repo 管理</div>
            <Button
              style={{ fontSize: 12 }}
              onClick={() => setVisible(true)}
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
        <InfiniteScroll
          dataLength={data.length}
          next={() => {
            fetch(paginate.page + 1, paginate.pageSize);
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
                      setVisible(true);
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
                      if (!error) {
                        setTimeout(() => {
                          setData((items) =>
                            items.map((v) =>
                              v.id === item.id
                                ? { ...v, enabled: !item.enabled }
                                : v
                            )
                          );
                          message.success("操作成功");
                          setLoading((lo) => ({
                            ...lo,
                            items: { ...lo.items, [item.id]: false },
                          }));
                        }, 1000);
                      }
                    }}
                  >
                    {item.enabled ? "禁用" : "启用"}
                  </Button>,
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
