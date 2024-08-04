import React, { useState, useCallback, useEffect, memo, Suspense } from "react";
import { DraggableModal } from "../pkg/DraggableModal";
import { detailProject } from "../api/project";
import { css } from "@emotion/css";
import theme from "../styles/theme";
import { Button, Tabs, Skeleton, Badge, Spin } from "antd";
import DeployStatus from "./DeployStatus";
import { setNamespaceReload } from "../store/actions";
import ErrorBoundary from "./ErrorBoundary";
import ServiceEndpoint from "./ServiceEndpoint";
import { useDispatch } from "react-redux";
import pb from "../api/compiled";
import TabInfo from "./TabInfo";
import TabEdit from "./TabEdit";
import Shell from "./TabShell";
import TabLog from "./TabLog";
import useProjectRoom from "../contexts/useProjectRoom";
import { useWs } from "../contexts/useWebsocket";
import { Tab } from "rc-tabs/lib/interface";
import { useSelector } from "react-redux";
import { modals } from "../store/reducers/openedModal";
import { useSearchParams } from "react-router-dom";
import { sortedUniq } from "lodash";
import { components } from "../api/schema";

const ItemDetailModal: React.FC<{
  item: components["schemas"]["types.ProjectModel"];
  namespace: string;
  namespaceId: number;
}> = ({ item, namespace, namespaceId }) => {
  const dispatch = useDispatch();
  const openModals = useSelector(modals);
  const [visible, setVisible] = useState(openModals[item.id] || false);
  const [params, setParams] = useSearchParams();
  const onOpenModal = useCallback(() => {
    let pIDs = (params.get("pid") || "").split(",");
    pIDs.push(String(item.id));
    setParams({ pid: sortedUniq(pIDs).join(",") });
    setVisible(true);
  }, [item.id, setParams, params]);
  const onCloseModal = useCallback(() => {
    setVisible(false);
    let pidStr = sortedUniq(
      (params.get("pid") || "").split(",").filter((v) => v !== String(item.id))
    ).join(",");
    setParams(!!pidStr ? { pid: pidStr } : {});
  }, [item.id, setParams, params]);

  const [detail, setDetail] = useState<
    components["schemas"]["project.ShowResponse"] | undefined
  >();
  const [resizeAt, setResizeAt] = useState<number>(0);

  useEffect(() => {
    if (visible && namespaceId && item.id) {
      detailProject(item.id).then((res) => {
        setDetail(res.data);
      });
    }
  }, [item.id, visible, namespaceId]);

  const onDelete = useCallback(() => {
    dispatch(setNamespaceReload(true, namespaceId));
    onCloseModal();
  }, [dispatch, namespaceId, onCloseModal]);

  const onSuccess = useCallback(() => {
    item.id &&
      detailProject(item.id).then((res) => {
        setDetail(res.data);
      });
  }, [item.id]);

  return (
    <div className="project-detail">
      <Button
        onClick={() => {
          onOpenModal();
        }}
        className={css`
          width: 100%;
          display: flex;
          align-items: center;
          justify-content: center;
        `}
        type="dashed"
      >
        <DeployStatus status={item.deploy_status} />
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
        {item.deploy_status === pb.types.Deploy.StatusDeployed && (
          <ServiceEndpoint projectId={item.id} />
        )}
      </Button>
      <DraggableModal
        onResize={() => {
          setResizeAt(new Date().getTime());
        }}
        className="draggable-modal"
        destroyOnClose
        open={visible}
        initialWidth={900}
        initialHeight={600}
        footer={null}
        keyboard={false}
        onCancel={onCloseModal}
        title={
          <Badge.Ribbon
            className={css`
              top: 0;
              color: ${theme.mainColor};
              cursor: auto;
              font-family: '"Fira code", "Fira Mono", monospace';
              margin-left: -16px;
            `}
            placement="start"
            text={namespace}
          >
            <div style={{ textAlign: "center", fontSize: 18 }}>{item.name}</div>
          </Badge.Ribbon>
        }
      >
        {detail && detail.project ? (
          <MyTabs
            namespaceId={namespaceId}
            projectID={detail.project.id}
            detail={detail}
            onDelete={onDelete}
            onSuccess={onSuccess}
            item={item}
            resizeAt={resizeAt}
          />
        ) : (
          <div
            style={{
              display: "flex",
              height: "100%",
              justifyContent: "center",
              alignItems: "center",
            }}
          >
            <Spin spinning={true} size="large" />
          </div>
        )}
      </DraggableModal>
    </div>
  );
};

const MyTabs: React.FC<{
  detail: components["schemas"]["project.ShowResponse"];
  item: components["schemas"]["types.NamespaceModel"];
  resizeAt: any;
  onSuccess: () => void;
  onDelete: () => void;
  projectID: number;
  namespaceId: number;
}> = memo(
  ({ detail, item, namespaceId, projectID, resizeAt, onSuccess, onDelete }) => {
    console.log("render MyTabs");
    let items: Tab[] = [
      {
        key: "container-logs",
        label: "容器日志",
        children: (
          <>
            {detail?.project && detail.project.namespace ? (
              <TabLog
                updatedAt={detail.project.updated_at}
                id={detail.project.id}
                namespace={detail.project.namespace.name}
                namespaceID={detail.project.namespace.id}
              />
            ) : (
              <Skeleton active />
            )}
          </>
        ),
      },
      {
        key: "shell",
        label: "命令行",
        children: (
          <div style={{ height: "100%" }}>
            <Suspense fallback={<Skeleton active />}>
              <ErrorBoundary>
                {detail?.project && detail.project.namespace && (
                  <Shell
                    namespaceID={detail.project.namespace.id}
                    namespace={detail.project.namespace.name}
                    id={detail.project.id}
                    updatedAt={detail.project.updated_at}
                    resizeAt={resizeAt}
                  />
                )}
              </ErrorBoundary>
            </Suspense>
          </div>
        ),
      },
      {
        key: "update-config",
        label: "配置更新",
        children: (
          <>
            <Suspense fallback={<Skeleton active />}>
              {detail?.project && detail.project.namespace && (
                <TabEdit
                  elements={detail.elements}
                  namespaceId={detail.project.namespace.id}
                  detail={detail.project}
                  updatedAt={detail.project.updated_at}
                  onSuccess={onSuccess}
                />
              )}
            </Suspense>
          </>
        ),
      },
    ];
    items = [
      ...(item.deploy_status === pb.types.Deploy.StatusDeployed ||
      item.deploy_status === pb.types.Deploy.StatusDeploying
        ? items
        : []),
      {
        key: "detail",
        label: "详细信息",
        children: (
          <div className="detail-tab">
            <Suspense fallback={<Skeleton active />}>
              {detail?.project && (
                <TabInfo
                  detail={detail.project}
                  cpu={detail.cpu}
                  memory={detail.memory}
                  git_commit_web_url={detail.project.git_commit_web_url}
                  git_commit_title={detail.project.git_commit_title}
                  git_commit_author={detail.project.git_commit_author}
                  git_commit_date={detail.project.git_commit_date}
                  urls={detail.urls}
                  onDeleted={onDelete}
                />
              )}
            </Suspense>
          </div>
        ),
      },
    ];
    useProjectRoom(namespaceId, projectID, useWs());
    return (
      <Tabs
        destroyInactiveTabPane
        defaultActiveKey="1"
        centered
        items={items}
        style={{ height: "100%" }}
      />
    );
  }
);

export default memo(ItemDetailModal);
