import React, { useEffect, useState, useCallback, memo } from "react";
import { DraggableModalProvider } from "../pkg/DraggableModal/DraggableModalProvider";
import ItemCard from "./ItemCard";
import { Empty, Row, Col, message } from "antd";
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

const AppContent: React.FC = () => {
  const reloadNamespace = useSelector(selectReload);
  const reloadNsID = useSelector(selectReloadNsID);
  const dispatch = useDispatch();
  const [loading, setLoading] = useState(false);
  const [namespaceItems, setNamespaceItems] = useAsyncState<
    components["schemas"]["types.NamespaceModel"][]
  >([]);
  const fetchNamespaces = useCallback(() => {
    setLoading(true);
    return ajax
      .GET("/api/namespaces")
      .then(({ data, error }) => {
        if (error) {
          return;
        }
        setNamespaceItems(data.items);
      })
      .finally(() => setLoading(false));
  }, [setNamespaceItems]);

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
    fetchNamespaces();
  }, [fetchNamespaces]);

  useEffect(() => {
    if (reloadNamespace) {
      fetchNamespaces().finally(() => dispatch(setNamespaceReload(false, 0)));
    }
  }, [reloadNamespace, dispatch, fetchNamespaces]);

  return (
    <DraggableModalProvider>
      <Content>
        <AddNamespace onCreated={fetchNamespaces} />

        {namespaceItems.length < 1 ? (
          <Empty description={false} imageStyle={{ height: 300 }} />
        ) : (
          <Row gutter={[16, 16]}>
            {namespaceItems.map(
              (item: components["schemas"]["types.NamespaceModel"]) => (
                <Col md={12} lg={8} sm={12} xs={24} key={item.id}>
                  <ItemCard
                    loading={
                      loading && (item.id === reloadNsID || reloadNsID === 0)
                    }
                    item={item}
                    onNamespaceDeleted={fetchNamespaces}
                  />
                </Col>
              )
            )}
          </Row>
        )}
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
