import React, { useEffect, useState, useCallback } from "react";
import { DraggableModalProvider } from "ant-design-draggable-modal";
import ItemCard from "./ItemCard";
import { Row, Col } from "antd";
import AddNamespace from "./AddNamespace";
import "ant-design-draggable-modal/dist/index.css";
import { listNamespaces, NamespaceItem } from "../api/namespace";

import { useWs } from "../contexts/useWebsocket";
import { useSelector, useDispatch } from "react-redux";
import { setNamespaceReload } from "../store/actions";
import { selectReload } from "../store/reducers/namespace";

const AppContent: React.FC = () => {
  const reloadNamespace = useSelector(selectReload)
  const dispatch = useDispatch()
  const ws = useWs();
  console.log(ws, "console.log(ws)");
  const [loading, setLoading] = useState(false);
  const [namespaceItems, setNamespaceItems] = useState<NamespaceItem[]>([]);
  const onNamespaceCreated = ({ id, name }: { id: number; name: string }) => {
    console.log(id, name);
    fetchNamespaces();
  };
  const fetchNamespaces = useCallback(() => {
    setLoading(true);
    listNamespaces()
      .then((res) => {
        setNamespaceItems(res.data.data);
        setLoading(false);
      })
      .catch((e) => {
        setLoading(false);
      });
  }, []);

  useEffect(() => {
    fetchNamespaces();
  }, [fetchNamespaces]);

  useEffect(() => {
    console.log("reloadNamespace", reloadNamespace);
    if (reloadNamespace) {
      fetchNamespaces();
      dispatch(setNamespaceReload(false));
    }
  }, [reloadNamespace,dispatch, fetchNamespaces]);

  return (
    <DraggableModalProvider>
      <div className="content">
        <AddNamespace onCreated={onNamespaceCreated} />

        <Row gutter={[16, 16]}>
          {namespaceItems.map((item: NamespaceItem) => (
            <Col md={12} lg={8} sm={12} xs={24} key={item.id}>
              <ItemCard
                loading={loading}
                item={item}
                onNamespaceDeleted={() => fetchNamespaces()}
              />
            </Col>
          ))}
        </Row>
      </div>
    </DraggableModalProvider>
  );
};

export default AppContent;