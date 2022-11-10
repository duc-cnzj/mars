import React, { memo, useCallback, useState } from "react";
import { Timeline, Button} from "antd";
import { useSelector } from "react-redux";
import { selectList } from "../store/reducers/createProject";
import pb from "../api/compiled";
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
      case pb.websocket.ResultType.DeployedFailed:
        return "red";
      case pb.websocket.ResultType.Deployed:
        return "blue";
      default:
        return "blue";
    }
  }, []);

  return (
    <>
      <Timeline
        pending={
          list[slug]?.isLoading ? (pending ? pending : "loading...") : false
        }
        reverse={true}
        style={{ paddingLeft: 2 }}
      >
        {list[slug]?.output?.map((data, index) => (
          <Timeline.Item key={index} color={getResultColor(data.type)}>
            <div style={{display: "flex", alignItems: "center"}}>
              <span style={{marginRight: 5}}>{data.log}</span>
              {data.containers &&
                data.containers.map((item, k) => (
                  <LogButton c={item} key={k} />
                ))}
            </div>
          </Timeline.Item>
        ))}
      </Timeline>
    </>
  );
};

const LogButton: React.FC<{ c: pb.types.Container }> = ({ c }) => {
  const [visible, setVisible] = useState(false);
  const [timestamp, setTimestamp] = useState(new Date().getTime());

  const handleVisibleChange = (newVisible: boolean) => {
    setVisible(newVisible);
    newVisible && setTimestamp(new Date().getTime());
  };
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
      <Button danger size="small" style={{marginRight: 5}} onClick={() => handleVisibleChange(true)}>
        查看日志 {c.pod}
      </Button>
    </div>
  );
};

export default memo(LogOutput);
