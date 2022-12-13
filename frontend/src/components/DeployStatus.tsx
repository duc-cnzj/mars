import React, { memo } from "react";
import {
  CheckCircleTwoTone,
  QuestionCircleTwoTone,
  CloseCircleTwoTone,
  ClockCircleTwoTone,
} from "@ant-design/icons";
import { Tooltip } from "antd";
import pb from "../api/compiled";

const DeployStatus: React.FC<{ status: pb.types.Deploy }> = ({ status }) => {
  return (
    <>
      {status === pb.types.Deploy.StatusUnknown && (
        <Tooltip
          placement="top"
          overlayStyle={{ fontSize: "10px" }}
          title={"状态未知，刷新试试~"}
        >
          <QuestionCircleTwoTone
            twoToneColor="#f9ca24"
            style={{
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
            }}
          />
        </Tooltip>
      )}
      {status === pb.types.Deploy.StatusDeployed && (
        <CheckCircleTwoTone
          twoToneColor="#52c41a"
          style={{
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
          }}
        />
      )}

      {status === pb.types.Deploy.StatusDeploying && (
        <Tooltip
          placement="top"
          overlayStyle={{ fontSize: "10px" }}
          title={"正在部署中..."}
        >
          <ClockCircleTwoTone
            twoToneColor="#6EE7B7"
            style={{
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
            }}
          />
        </Tooltip>
      )}

      {status === pb.types.Deploy.StatusFailed && (
        <CloseCircleTwoTone
          twoToneColor="#eb4d4b"
          style={{
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
          }}
        />
      )}
    </>
  );
};

export default memo(DeployStatus);
