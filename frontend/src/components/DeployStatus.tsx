import React, { memo } from "react";
import {
  CheckCircleTwoTone,
  QuestionCircleTwoTone,
  CloseCircleTwoTone,
  ClockCircleTwoTone,
} from "@ant-design/icons";
import { Tooltip } from "antd";
import { TypesProjectModelDeployStatus } from "../api/schema.d";

const DeployStatus: React.FC<{ status: TypesProjectModelDeployStatus }> = ({
  status,
}) => {
  return (
    <>
      {status === TypesProjectModelDeployStatus.StatusUnknown && (
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
      {status === TypesProjectModelDeployStatus.StatusDeployed && (
        <CheckCircleTwoTone
          twoToneColor="#52c41a"
          style={{
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
          }}
        />
      )}

      {status === TypesProjectModelDeployStatus.StatusDeploying && (
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

      {status === TypesProjectModelDeployStatus.StatusFailed && (
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
