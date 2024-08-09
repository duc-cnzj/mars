import React, { memo } from "react";
import { css } from "@emotion/css";
import { components } from "../api/schema";
import ProjectSelectorV2 from "./ProjectSelectorV2";

const ModalSub: React.FC<{
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
      <ProjectSelectorV2
        onSuccess={onSuccess}
        isEdit
        namespaceId={namespaceId}
        project={detail}
      />
    </div>
  );
};
export default memo(ModalSub);
