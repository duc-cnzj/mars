import React from "react";
import pb from "../api/compiled";
import { Tag } from "antd";
import { SyncOutlined } from "@ant-design/icons";

const PodStateTag: React.FC<{ pod: pb.types.StateContainer }> = ({ pod }) => {
  if (pod.terminating) {
    return (
      <Tag
        icon={<SyncOutlined spin />}
        color="#fca5a5"
        style={{ marginLeft: 5 }}
      >
        {pod.pod} 停止中
      </Tag>
    );
  }

  if (pod.is_old) {
    return (
      <Tag
        icon={<SyncOutlined spin />}
        color="#fde047"
        style={{ marginLeft: 5 }}
      >
        {pod.pod} 即将停止
      </Tag>
    );
  }

  return (
    <Tag className="pod-running-tag" style={{ marginLeft: 5 }}>
      {pod.pod}
    </Tag>
  );
};

export default PodStateTag;
