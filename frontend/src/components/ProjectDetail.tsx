import React, {
  useState,
  useCallback,
  useEffect,
  memo,
  lazy,
  Suspense,
} from "react";
import { DraggableModal } from "../pkg/DraggableModal";
import { detailProject } from "../api/project";
import { Button, Tabs, Skeleton, Badge } from "antd";
import DeployStatus from "./DeployStatus";
import { setNamespaceReload } from "../store/actions";
import ErrorBoundary from "./ErrorBoundary";
import ServiceEndpoint from "./ServiceEndpoint";
import { useDispatch } from "react-redux";
import pb from "../api/compiled";

import TabLog from "./TabLog";
const TabInfo = lazy(() => import("./TabInfo"));
const TabEdit = lazy(() => import("./TabEdit"));
const Shell = lazy(() => import("./TabShell"));

const { TabPane } = Tabs;

const ItemDetailModal: React.FC<{
  item: pb.NamespaceItem.ISimpleProjectItem;
  namespace: string;
  namespaceId: number;
}> = ({ item, namespace, namespaceId }) => {
  const dispatch = useDispatch();
  const [visible, setVisible] = useState(false);
  const onOk = useCallback(() => setVisible(true), []);
  const [detail, setDetail] = useState<pb.ProjectShowResponse | undefined>();
  const [resizeAt, setResizeAt] = useState<number>(0);

  useEffect(() => {
    if (visible && namespaceId && item.id) {
      detailProject(namespaceId, item.id).then((res) => {
        console.log(res.data);
        setDetail(res.data);
      });
    }
  }, [namespaceId, item.id, visible]);

  const onSuccess = useCallback(() => {
    detailProject(namespaceId, item.id || 0).then((res) => {
      console.log(res.data);
      setDetail(res.data);
    });
  }, [item.id, namespaceId]);

  const onCancel = useCallback(() => {
    setVisible(false);
  }, []);

  return (
    <>
      <Button
        onClick={() => {
          onOk();
        }}
        className="project-detail__show-button"
        type="dashed"
      >
        <DeployStatus status={item.status || ""} />
        <span
          title={item.name || ""}
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
          <ServiceEndpoint
            namespaceId={namespaceId}
            projectName={item.name || ""}
          />
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
        initialWidth={900}
        initialHeight={600}
        footer={null}
        keyboard={false}
        onCancel={onCancel}
        title={
          <Badge.Ribbon
            className="project-detail__badge"
            placement="start"
            text={namespace}
          >
            <div style={{ textAlign: "center", fontSize: 18 }}>{item.name}</div>
          </Badge.Ribbon>
        }
      >
        <Tabs
          destroyInactiveTabPane
          defaultActiveKey="1"
          centered
          style={{ height: "100%" }}
        >
          {item.status === "deployed" ? (
            <>
              <TabPane tab="容器日志" key="container-logs">
                {detail ? (
                  <TabLog
                    updatedAt={detail.updated_timestamp}
                    id={detail.id}
                    namespaceId={detail.namespace?.id || 0}
                  />
                ) : (
                  <Skeleton active />
                )}
              </TabPane>
              <TabPane tab="命令行" key="shell" style={{ height: "100%" }}>
                <Suspense fallback={<Skeleton active />}>
                  <ErrorBoundary>
                    {detail ? (
                      <Shell
                        updatedAt={detail.updated_timestamp}
                        resizeAt={resizeAt}
                        detail={detail}
                      />
                    ) : (
                      <></>
                    )}
                  </ErrorBoundary>
                </Suspense>
              </TabPane>
              <TabPane tab="配置更新" key="update-config">
                <Suspense fallback={<Skeleton active />}>
                  {detail ? (
                    <TabEdit detail={detail} onSuccess={onSuccess} />
                  ) : (
                    <></>
                  )}
                </Suspense>
              </TabPane>
            </>
          ) : (
            <></>
          )}
          <TabPane tab="详细信息" key="detail" className="detail-tab">
            <Suspense fallback={<Skeleton active />}>
              {detail ? (
                <TabInfo
                  detail={detail}
                  onDeleted={() => {
                    dispatch(setNamespaceReload(true));
                    setVisible(false);
                  }}
                />
              ) : (
                <></>
              )}
            </Suspense>
          </TabPane>
        </Tabs>
      </DraggableModal>
    </>
  );
};

export default memo(ItemDetailModal);
