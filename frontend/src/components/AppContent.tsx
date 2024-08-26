import React, { useEffect, useState, useCallback, memo } from "react";
import { DraggableModalProvider } from "../pkg/DraggableModal/DraggableModalProvider";
import ItemCard from "./ItemCard";
import { Empty, Row, Col, Tabs, message, Pagination, Input } from "antd";
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

const defaultPageSize = 12;

const AppContent: React.FC = () => {
  const reloadNamespace = useSelector(selectReload);
  const reloadNsID = useSelector(selectReloadNsID);
  const dispatch = useDispatch();
  const [favorite, setFavorite] = useState(true);
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
  const [params, setParams] = useSearchParams();
  if (!!params.get("pid")) {
    let obj: { [key: number]: boolean } = {};
    sortedUniq((params.get("pid") || "").split(","))
      .filter((v) => isNumber(Number(v)))
      .map((v) => (obj[Number(v)] = true));
    dispatch(setOpenedModals(obj));
  }

  usePreventModalBack();

  useEffect(() => {
    fetchNamespaces(true, 1, defaultPageSize, "");
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
  const onTabsClick = useCallback(
    (v: any) => {
      let fav = v === "1";
      setFavorite(fav);
      fetchNamespaces(fav, 1, defaultPageSize, searchInput.name);
      setParams({});
      dispatch(setOpenedModals({}));
    },
    [dispatch, setParams, fetchNamespaces, searchInput.name],
  );
  const [isFocused, setIsFocused] = useState(false);

  return (
    <DraggableModalProvider>
      <Content>
        <Row justify={"space-between"}>
          <Col span={8}>
            <MyTabs onClick={onTabsClick} />
          </Col>
          <Col span={10} style={{ textAlign: "right" }}>
            <Input
              allowClear
              className={css`
                margin-left: 20px;
                width: ${isFocused ? "60%" : "30%"};
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
        {pageInfo.count > defaultPageSize && (
          <Row style={{ marginTop: 10 }}>
            <Pagination
              showSizeChanger={false}
              defaultCurrent={pageInfo.page}
              total={pageInfo.count}
              defaultPageSize={defaultPageSize}
              pageSize={pageInfo.pageSize}
              onChange={(page, size) => {
                fetchNamespaces(favorite, page, size, searchInput.name);
              }}
            />
          </Row>
        )}
      </Content>
    </DraggableModalProvider>
  );
};

const MyTabs: React.FC<{ onClick: (v: any) => void }> = memo(({ onClick }) => {
  const items: TabsProps["items"] = [
    {
      key: "1",
      label: "我的关注",
      icon: <IconFont name="#icon-wodeguanzhu" color="#a78bfa" />,
    },
    {
      key: "2",
      label: "全部项目",
      icon: <IconFont name="#icon-kongjian" />,
    },
  ];
  return <Tabs onTabClick={onClick} defaultActiveKey="1" items={items} />;
});

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
