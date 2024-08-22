import React, { memo, useCallback, useMemo, useState } from "react";
import { Timeline, Button, TimelineItemProps } from "antd";
import { useSelector } from "react-redux";
import { selectList } from "../store/reducers/createProject";
import pb from "../api/websocket";
import { LogUtil } from "./TabLog";
import { DraggableModal } from "../pkg/DraggableModal";

const LogOutput: React.FC<{ slug: string; pending?: React.ReactNode }> = ({
  slug,
  pending,
}) => {
  const list = useSelector(selectList);
  const getResultColor = useCallback((data: pb.websocket.ResultType) => {
    switch (data) {
      case pb.websocket.ResultType.DeployedCanceled:
        return "#F59E0B";
      case pb.websocket.ResultType.Error:
      case pb.websocket.ResultType.DeployedFailed:
        return "red";
      case pb.websocket.ResultType.Deployed:
        return "blue";
      default:
        return "blue";
    }
  }, []);

  const items = useMemo(
    () =>
      list[slug]?.output?.map(
        (data): TimelineItemProps => ({
          label: "",
          children: (
            <div style={{ display: "flex", alignItems: "center" }}>
              <span style={{ marginRight: 5 }}>{data.log}</span>
              {data.containers &&
                data.containers.map((item, k) => (
                  <LogButton c={item} key={k} />
                ))}
            </div>
          ),
          color: getResultColor(data.type),
        }),
      ),
    [getResultColor, list, slug],
  );

  return (
    <Timeline
      pending={
        list[slug]?.isLoading ? (pending ? pending : "loading...") : false
      }
      reverse={true}
      style={{ paddingLeft: 2 }}
      items={items}
    />
  );
};

const LogButton: React.FC<{ c: pb.websocket.Container }> = memo(({ c }) => {
  const [visible, setVisible] = useState(false);
  const [timestamp, setTimestamp] = useState(new Date().getTime());

  const handleVisibleChange = useCallback((newVisible: boolean) => {
    setVisible(newVisible);
    newVisible && setTimestamp(new Date().getTime());
  }, []);

  return (
    <div>
      <DraggableModal
        zIndex={99999}
        className="draggable-modal"
        destroyOnClose
        open={visible}
        initialWidth={900}
        initialHeight={600}
        footer={null}
        keyboard={false}
        onCancel={() => handleVisibleChange(false)}
        title={`容器日志: ${c.pod}`}
      >
        {visible && (
          <LogUtil
            freshTime={timestamp}
            pod={c.pod}
            container={c.container}
            namespace={c.namespace}
          />
        )}
      </DraggableModal>
      <Button
        danger
        size="small"
        style={{ marginRight: 5 }}
        onClick={() => handleVisibleChange(true)}
      >
        查看日志 {c.pod}
      </Button>
    </div>
  );
});

export default memo(LogOutput);
