import React, { memo } from "react";
import {
  CheckCircleTwoTone,
  QuestionCircleTwoTone,
  CloseCircleTwoTone,
  ClockCircleTwoTone,
} from "@ant-design/icons";
import { Tooltip } from "antd";

const DeployStatus: React.FC<{ status: string }> = ({ status }) => {
  return (
    <>
      {status === "unknown" && (
        <Tooltip
          placement="top"
          overlayStyle={{ fontSize: "10px" }}
          title={"状态未知，建议重新部署~"}
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
      {status === "deployed" && (
        <CheckCircleTwoTone
          twoToneColor="#52c41a"
          style={{
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
          }}
        />
      )}

      {status === "pending" && (
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

      {status === "failed" && (
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
