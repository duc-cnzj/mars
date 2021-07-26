import React, { useState, useCallback, useEffect, memo } from "react";
import { DraggableModal } from "../pkg/DraggableModal/DraggableModal";
import { detailProject, ProjectDetail } from "../api/project";
import { Button, Tabs, Skeleton, Switch } from "antd";
import DeployStatus from "./DeployStatus";
import TabInfo from "./TabInfo";
import TabLog from "./TabLog";
import { setNamespaceReload } from "../store/actions";
import Shell from "./TabShell";
import EditProject from "./TabEdit";
import ErrorBoundary from "./ErrorBoundary";
import ServiceEndpoint from "./ServiceEndpoint";
import { useDispatch } from "react-redux";

const { TabPane } = Tabs;

const ItemDetailModal: React.FC<{
  item: { id: number; name: string; status: string };
  namespace: string;
  namespaceId: number;
}> = ({ item, namespace, namespaceId }) => {
  const dispatch = useDispatch();
  const [visible, setVisible] = useState(false);
  const onOk = useCallback(() => setVisible(true), []);
  const [detail, setDetail] = useState<ProjectDetail | undefined>();
  const [resizeAt, setResizeAt] = useState<number>(0);

  console.log("render ItemDetailModal");
  useEffect(() => {
    if (visible && namespaceId && item.id) {
      detailProject(namespaceId, item.id).then((res) => {
        console.log(res.data.data);
        setDetail(res.data.data);
      });
    }
  }, [namespaceId, item.id, visible]);

  const onSuccess = () => {
    detailProject(namespaceId, item.id).then((res) => {
      console.log(res.data.data);
      setDetail(res.data.data);
    });
  };

  const [autoRefresh, setAutoRefresh] = useState(false);
  const handleAutoRefresh = (f: boolean) => {
    setAutoRefresh(f);
  };

  const onCancel = useCallback(() => {
    setVisible(false);
    setAutoRefresh(false);
  }, []);

  return (
    <>
      <Button
        onClick={() => {
          onOk();
        }}
        style={{
          width: "100%",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
        }}
        type="dashed"
      >
        <DeployStatus status={item.status} />
        <span
          title={item.name}
          style={{
            textOverflow: "ellipsis",
            whiteSpace: "nowrap",
            overflow: "hidden",
            marginRight: 5,
          }}
        >
          {item.name}
        </span>
        {item.status === "deployed" ? (
          <ServiceEndpoint namespaceId={namespaceId} projectName={item.name} />
        ) : (
          <></>
        )}
      </Button>
      <DraggableModal
        onResize={() => {
          console.log("DraggableModal onResize");
          setResizeAt(new Date().getTime());
        }}
        className="draggable-modal"
        visible={visible}
        initialWidth={800}
        initialHeight={600}
        footer={null}
        onCancel={onCancel}
        title={item.name + "(" + namespace + ")"}
      >
        <Tabs defaultActiveKey="1" centered>
          {item.status === "deployed" ? (
            <>
              <TabPane tab="容器日志" key="container-logs">
                <div style={{ marginBottom: 10 }}>
                  <span style={{ marginRight: 5 }}>自动刷新(5s):</span>
                  <Switch
                    checked={autoRefresh}
                    onChange={handleAutoRefresh}
                    defaultChecked={autoRefresh}
                  />
                </div>
                {detail ? (
                  <TabLog
                    updatedAt={detail.updated_at}
                    autoRefresh={autoRefresh}
                    id={detail.id}
                    namespaceId={detail.namespace.id}
                  />
                ) : (
                  <Skeleton active />
                )}
              </TabPane>
              <TabPane tab="命令行" key="shell">
                <ErrorBoundary>
                  {detail ? (
                    <Shell resizeAt={resizeAt} detail={detail} />
                  ) : (
                    <Skeleton active />
                  )}
                </ErrorBoundary>
              </TabPane>
              <TabPane tab="配置更新" key="update-config">
                {detail ? (
                  <EditProject detail={detail} onSuccess={onSuccess} />
                ) : (
                  <Skeleton active />
                )}
              </TabPane>
            </>
          ) : (
            <></>
          )}
          <TabPane tab="详细信息" key="detail" className="detail-tab">
            <TabInfo
              detail={detail}
              onDeleted={() => {
                setTimeout(() => {
                  dispatch(setNamespaceReload(true));
                  setVisible(false);
                }, 500);
              }}
            />
          </TabPane>
        </Tabs>
      </DraggableModal>
    </>
  );
};

export default memo(ItemDetailModal);
