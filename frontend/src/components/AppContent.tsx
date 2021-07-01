import React, { useEffect, useState, useCallback } from "react";
import { DraggableModalProvider } from "ant-design-draggable-modal";
import ItemCard from "./ItemCard";
import { Row, Col, Affix, Button } from "antd";
import AddNamespace from "./AddNamespace";
import "ant-design-draggable-modal/dist/index.css";
import { listNamespaces, NamespaceItem } from "../api/namespace";
import { SettingOutlined } from "@ant-design/icons";

import { useWs } from "../contexts/useWebsocket";
import { useSelector, useDispatch } from "react-redux";
import { setNamespaceReload } from "../store/actions";
import { selectReload } from "../store/reducers/namespace";
import { useHistory } from "react-router-dom";
import {
  CreateProjectItem,
  List,
  selectList,
} from "../store/reducers/createProject";

const AppContent: React.FC = () => {
  const reloadNamespace = useSelector(selectReload);
  const dispatch = useDispatch();
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
  }, [reloadNamespace, dispatch, fetchNamespaces]);
  const list = useSelector<any, List>(selectList);
  const [can, setCan] = useState(true);
  useEffect(() => {
    let flag: boolean = false;
    for (const key in list) {
      if ((list[key] as CreateProjectItem).isLoading) {
        flag = true;
        break;
      }
    }

    setCan(!flag);
  }, [list]);

  let h = useHistory();

  return (
    <DraggableModalProvider>
      <div className="content">
        <AddNamespace onCreated={onNamespaceCreated} />
        <Affix offsetTop={130} style={{ position: "absolute", right: "10px" }}>
          <Button
            disabled={!can}
            size="large"
            type="ghost"
            shape="circle"
            icon={<SettingOutlined />}
            onClick={() => h.push("/web/gitlab_project_manager")}
          />
        </Affix>

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
