import React, { memo } from "react";
import {
    CheckCircleTwoTone,
    QuestionCircleTwoTone,
    CloseCircleTwoTone,
  } from "@ant-design/icons";

const DeployStatus: React.FC<{status: string}> = ({status}) => {
  return (
      <>
      {status === "unknown" && (
          <QuestionCircleTwoTone
            twoToneColor="#f9ca24"
            style={{
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
            }}
          />
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