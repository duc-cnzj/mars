import React, { useEffect, useState, useCallback, memo } from "react";
import { DraggableModalProvider } from "../pkg/DraggableModal/DraggableModalProvider";
import ItemCard from "./ItemCard";
import { Empty, Row, Col, Tabs, message } from "antd";
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

const AppContent: React.FC = () => {
  const reloadNamespace = useSelector(selectReload);
  const reloadNsID = useSelector(selectReloadNsID);
  const dispatch = useDispatch();
  const [favorite, setFavorite] = useState(true);
  const [loading, setLoading] = useState(false);
  const [namespaceItems, setNamespaceItems] = useAsyncState<
    components["schemas"]["types.NamespaceModel"][]
  >([]);
  const fetchNamespaces = useCallback(
    (favorite: boolean) => {
      setLoading(true);
      return ajax
        .GET("/api/namespaces", { params: { query: { favorite } } })
        .then(({ data, error }) => {
          if (error) {
            return;
          }
          setNamespaceItems(data.items);
        })
        .finally(() => setLoading(false));
    },
    [setNamespaceItems],
  );

  const [params] = useSearchParams();
  if (!!params.get("pid")) {
    let obj: { [key: number]: boolean } = {};
    sortedUniq((params.get("pid") || "").split(","))
      .filter((v) => isNumber(Number(v)))
      .map((v) => (obj[Number(v)] = true));
    dispatch(setOpenedModals(obj));
  }

  usePreventModalBack();

  useEffect(() => {
    fetchNamespaces(favorite);
  }, [fetchNamespaces, favorite]);

  useEffect(() => {
    if (reloadNamespace) {
      fetchNamespaces(favorite).finally(() =>
        dispatch(setNamespaceReload(false, 0)),
      );
    }
  }, [reloadNamespace, dispatch, fetchNamespaces, favorite]);

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

  return (
    <DraggableModalProvider>
      <Content>
        <AddNamespace
          onCreated={() => {
            fetchNamespaces(favorite);
          }}
        />
        <Tabs
          onTabClick={(v) => setFavorite(v === "1")}
          defaultActiveKey="1"
          items={items}
        />
        <NamespaceList
          loading={loading}
          reloadNsID={reloadNsID}
          list={namespaceItems}
          fetchNamespaces={fetchNamespaces}
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
