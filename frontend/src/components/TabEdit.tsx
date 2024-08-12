import React, { memo } from "react";
import { css } from "@emotion/css";
import { components } from "../api/schema";
import DeployProjectForm from "./DeployProjectForm";

const EditProject: React.FC<{
  namespaceId: number;
  detail: components["schemas"]["types.ProjectModel"];
  onSuccess: () => void;
}> = ({ detail, onSuccess, namespaceId }) => {
  return (
    <div
      className={css`
        height: 100%;
        overflow-y: auto;
        .diff-viewer {
          pre {
            white-space: pre !important;
          }
        }
      `}
    >
      <DeployProjectForm
        onSuccess={onSuccess}
        isEdit
        namespaceId={namespaceId}
        project={detail}
      />
    </div>
  );
};
export default memo(EditProject);
