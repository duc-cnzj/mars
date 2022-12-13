import React, { useEffect, useState, useCallback, memo } from "react";
import { DraggableModalProvider } from "../pkg/DraggableModal/DraggableModalProvider";
import ItemCard from "./ItemCard";
import { Empty, Row, Col } from "antd";
import "../pkg/DraggableModal/index.css";
import { allNamespaces } from "../api/namespace";
import { useSelector, useDispatch } from "react-redux";
import { setNamespaceReload } from "../store/actions";
import { selectReload } from "../store/reducers/namespace";
import pb from "../api/compiled";
import AddNamespace from "./AddNamespace";
import { useAsyncState } from "../utils/async";

const AppContent: React.FC = () => {
  const reloadNamespace = useSelector(selectReload);
  const dispatch = useDispatch();
  const [loading, setLoading] = useState(false);
  const [namespaceItems, setNamespaceItems] = useAsyncState<
    pb.types.NamespaceModel[]
  >([]);
  const fetchNamespaces = useCallback(() => {
    setLoading(true);
    allNamespaces()
      .then((res) => {
        setNamespaceItems(res.data.items);
        setLoading(false);
      })
      .catch((e) => {
        setLoading(false);
      });
  }, [setNamespaceItems]);

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

  const onNamespaceCreated = useCallback(
    ({ id, name }: { id: number; name: string }) => {
      fetchNamespaces();
    },
    [fetchNamespaces]
  );

  useEffect(() => {
    fetchNamespaces();
  }, [fetchNamespaces]);

  useEffect(() => {
    if (reloadNamespace) {
      fetchNamespaces();
      dispatch(setNamespaceReload(false));
    }
  }, [reloadNamespace, dispatch, fetchNamespaces]);

  return (
    <DraggableModalProvider>
      <div className="content" style={{ marginBottom: 30 }}>
        <AddNamespace onCreated={onNamespaceCreated} />

        {namespaceItems.length < 1 ? (
          <Empty description={false} imageStyle={{ height: 300 }} />
        ) : (
          <Row gutter={[16, 16]}>
            {namespaceItems.map((item: pb.types.NamespaceModel) => (
              <Col md={12} lg={8} sm={12} xs={24} key={item.id}>
                <ItemCard
                  loading={loading}
                  item={item}
                  onNamespaceDeleted={() => fetchNamespaces()}
                />
              </Col>
            ))}
          </Row>
        )}
      </div>
    </DraggableModalProvider>
  );
};

export default memo(AppContent);
