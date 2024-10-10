import { Tooltip } from "antd";
import React, { memo, useState } from "react";
import ajax from "../api/ajax";
import IconFont from "./Icon";

const CpuMemory: React.FC<{
  namespaceID: number;
  title: string;
  projectID?: number;
}> = ({ namespaceID, title, projectID }) => {
  const [cpuAndMemory, setCpuAndMemory] = useState({ cpu: "", memory: "" });
  return (
    <Tooltip
      onOpenChange={(visible: boolean) => {
        if (visible) {
          if (projectID) {
            ajax
              .GET("/api/metrics/projects/{projectId}/cpu_memory", {
                params: { path: { projectId: projectID } },
              })
              .then(({ data }) => {
                data &&
                  setCpuAndMemory({
                    cpu: data.cpu,
                    memory: data.memory,
                  });
              });
            return;
          }
          ajax
            .GET("/api/metrics/namespace/{namespaceId}/cpu_memory", {
              params: { path: { namespaceId: namespaceID } },
            })
            .then(({ data }) => {
              data &&
                setCpuAndMemory({
                  cpu: data.cpu,
                  memory: data.memory,
                });
            });
        }
      }}
      title={
        <div style={{ fontSize: "10px" }}>
          <div>{title}</div>
          <div>
            <span>cpu: </span>
            <span>{cpuAndMemory.cpu}</span>
          </div>
          <div>
            <span>memory: </span>
            <span>{cpuAndMemory.memory}</span>
          </div>
        </div>
      }
      trigger="hover"
    >
      <IconFont style={{ cursor: "pointer" }} name="#icon-dianboxindiantu" />
    </Tooltip>
  );
};

export default memo(CpuMemory);
