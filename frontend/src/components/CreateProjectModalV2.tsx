import React, { memo } from "react";
import { DraggableModal } from "../pkg/DraggableModal/DraggableModal";
import { useAsyncState } from "../utils/async";
import { Button } from "antd";
import { PlusOutlined } from "@ant-design/icons";
import ProjectSelector from "./ProjectSelectorV2";

const CreateProjectModal: React.FC<{
  namespaceId: number;
}> = ({ namespaceId }) => {
  const [visible, setVisible] = useAsyncState<boolean>(false);
  return (
    <div>
      <Button
        onClick={() => setVisible(true)}
        style={{ width: "100%" }}
        type="dashed"
        icon={<PlusOutlined />}
      ></Button>
      <DraggableModal
        destroyOnClose
        open={visible}
        onCancel={() => setVisible(false)}
        footer={null}
        initialWidth={900}
        initialHeight={600}
        title={<div style={{ textAlign: "center" }}>创建项目</div>}
        className="draggable-modal"
      >
        <ProjectSelector
          onSuccess={() => setVisible(false)}
          namespaceId={namespaceId}
        />
      </DraggableModal>
    </div>
  );
};

export default memo(CreateProjectModal);
