import React, { useEffect, useState, useCallback, memo, useRef } from "react";
import { DraggableModalProvider } from "../pkg/DraggableModal/DraggableModalProvider";
import ItemCard from "./ItemCard";
import { Empty, Row, Col, Tabs, message, Pagination, Input, Space } from "antd";
import "../pkg/DraggableModal/index.css";
import { useSelector, useDispatch } from "react-redux";
import { setNamespaceReload, setOpenedModals } from "../store/actions";
import { selectReload, selectReloadNsID } from "../store/reducers/namespace";
import AddNamespace from "./AddNamespace";
import { useAsyncState } from "../utils/async";
import styled from "@emotion/styled";
import { useSearchParams } from "react-router-dom";
import { isNumber, sortedUniq } from "lodash";
import ajax from "../api/ajax";
import { components } from "../api/schema";
import IconFont from "./Icon";
import { TabsProps } from "antd/lib";
import { SearchOutlined } from "@ant-design/icons";
import { css } from "@emotion/css";
import useLocalStorage from "../contexts/useLocalstorage";

const defaultPageSize = 12;

const isFavorite = (v: string) => {
  return v === "2";
};

const AppContent: React.FC = () => {
  const { store, setStore } = useLocalStorage("active-tabs", "1");
  const init = useRef(store);
  const reloadNamespace = useSelector(selectReload);
  const reloadNsID = useSelector(selectReloadNsID);
  const dispatch = useDispatch();
  const [params, setParams] = useSearchParams();
  const [favorite, setFavorite] = useState(isFavorite(init.current));
  const [loading, setLoading] = useState(false);
  const [namespaceItems, setNamespaceItems] = useAsyncState<
    components["schemas"]["types.NamespaceModel"][]
  >([]);
  const [pageInfo, setPageInfo] = useState({
    page: 1,
    pageSize: defaultPageSize,
    count: 0,
  });
  const [searchInput, setSearchInput] = useState({ name: "" });

  const fetchNamespaces = useCallback(
    (favorite: boolean, page: number, pageSize: number, name?: string) => {
      setLoading(true);
      return ajax
        .GET("/api/namespaces", {
          params: {
            query: {
              favorite,
              page: page,
              pageSize: pageSize,
              name: name,
            },
          },
        })
        .then(({ data, error }) => {
          if (error) {
            return;
          }
          setNamespaceItems(data.items);
          setPageInfo({
            page: data.page,
            pageSize: data.pageSize,
            count: data.count,
          });
        })
        .finally(() => setLoading(false));
    },
    [setNamespaceItems],
  );
  if (!!params.get("pid") && !favorite) {
    let obj: { [key: number]: boolean } = {};
    sortedUniq((params.get("pid") || "").split(","))
      .filter((v) => isNumber(Number(v)))
      .map((v) => (obj[Number(v)] = true));
    dispatch(setOpenedModals(obj));
  }

  usePreventModalBack();

  useEffect(() => {
    fetchNamespaces(isFavorite(init.current), 1, defaultPageSize, "");
  }, [fetchNamespaces]);

  useEffect(() => {
    if (reloadNamespace) {
      fetchNamespaces(
        favorite,
        pageInfo.page,
        pageInfo.pageSize,
        searchInput.name,
      ).finally(() => dispatch(setNamespaceReload(false, 0)));
    }
  }, [
    reloadNamespace,
    dispatch,
    fetchNamespaces,
    favorite,
    pageInfo,
    searchInput.name,
  ]);
  const [isFocused, setIsFocused] = useState(false);
  const items: TabsProps["items"] = [
    {
      key: "1",
      label: "全部项目",
      icon: <IconFont name="#icon-kongjian" />,
    },
    {
      key: "2",
      label: "我的关注",
      icon: <IconFont name="#icon-wodeguanzhu" color="#a78bfa" />,
    },
  ];

  return (
    <DraggableModalProvider>
      <Content>
        <Row justify={"space-between"}>
          <Col span={8}>
            <Tabs
              onTabClick={(v) => {
                let fav = isFavorite(v);
                setStore(v);
                setFavorite(fav);
                fetchNamespaces(fav, 1, defaultPageSize, searchInput.name);
                if (fav) {
                  setParams({});
                  dispatch(setOpenedModals({}));
                }
              }}
              defaultActiveKey={store}
              items={items}
            />
          </Col>
          <Col span={16} style={{ textAlign: "right" }}>
            <Space>
              <Input
                allowClear
                placeholder="搜索空间名称"
                className={css`
                  font-size: 12px;
                  width: ${isFocused ? "80%" : "50%"};
                  margin-right: 10px;
                  transition:
                    width 0.3s ease-in-out,
                    background-color 0.3s ease-in-out;
                  &:focus {
                    background-color: black;
                    color: white; /* 使文本在黑色背景上可见 */
                  }
                `}
                onFocus={() => setIsFocused(true)}
                onBlur={() => {
                  setIsFocused(false);
                  fetchNamespaces(
                    favorite,
                    pageInfo.page,
                    pageInfo.pageSize,
                    searchInput.name,
                  );
                }}
                suffix={<SearchOutlined />}
                value={searchInput.name}
                onChange={(v) => setSearchInput({ name: v.target.value })}
                onKeyDown={(k) => {
                  if (k.code === "Enter") {
                    fetchNamespaces(
                      favorite,
                      pageInfo.page,
                      pageInfo.pageSize,
                      searchInput.name,
                    );
                  }
                }}
              />
              <Pagination
                simple
                showSizeChanger={false}
                style={{ fontSize: 12 }}
                defaultCurrent={pageInfo.page}
                total={pageInfo.count}
                defaultPageSize={defaultPageSize}
                pageSize={pageInfo.pageSize}
                onChange={(page, size) => {
                  fetchNamespaces(favorite, page, size, searchInput.name);
                }}
              />
              <AddNamespace
                onCreated={() => {
                  fetchNamespaces(
                    favorite,
                    pageInfo.page,
                    pageInfo.pageSize,
                    searchInput.name,
                  );
                }}
              />
            </Space>
          </Col>
        </Row>

        <NamespaceList
          loading={loading}
          reloadNsID={reloadNsID}
          list={namespaceItems}
          fetchNamespaces={() =>
            fetchNamespaces(
              favorite,
              pageInfo.page,
              pageInfo.pageSize,
              searchInput.name,
            )
          }
          favorite={favorite}
        />
      </Content>
    </DraggableModalProvider>
  );
};

export default memo(AppContent);

const usePreventModalBack = () => {
  useEffect(() => {
    window.history.pushState(null, document.title, window.location.href);
    let fn = function (event: any) {
      console.log("first");
      window.history.pushState(null, document.title, window.location.href);
    };
    console.log("add");
    window.addEventListener("popstate", fn);
    return () => {
      console.log("remove");
      window.removeEventListener("popstate", fn);
    };
  }, []);
};

const Content = styled.div`
  padding-top: 15px;
  margin-bottom: 30px;
`;

const NamespaceList: React.FC<{
  list: components["schemas"]["types.NamespaceModel"][];
  loading: boolean;
  reloadNsID: number;
  favorite: boolean;
  fetchNamespaces: (favorite: boolean) => void;
}> = memo(({ list, loading, reloadNsID, favorite, fetchNamespaces }) => {
  return (
    <>
      {list.length < 1 ? (
        <Empty description={false} imageStyle={{ height: 300 }} />
      ) : (
        <Row gutter={[16, 16]}>
          {list.map((item: components["schemas"]["types.NamespaceModel"]) => (
            <Col md={12} lg={8} sm={12} xs={24} key={item.id}>
              <ItemCard
                reload={() => fetchNamespaces(favorite)}
                onFavorite={(id: number, fav: boolean) => {
                  ajax
                    .POST("/api/namespaces/favorite", {
                      body: { id: id, favorite: fav },
                    })
                    .then(() => {
                      message.success(fav ? "关注成功" : "已取消关注");
                      fetchNamespaces(favorite);
                    });
                }}
                loading={
                  loading && (item.id === reloadNsID || reloadNsID === 0)
                }
                item={item}
                onNamespaceDeleted={() => {
                  fetchNamespaces(favorite);
                }}
              />
            </Col>
          ))}
        </Row>
      )}
    </>
  );
});
