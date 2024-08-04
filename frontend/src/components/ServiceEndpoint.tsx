import React, { memo, MouseEvent, useCallback, useState } from "react";

import { Popover } from "antd";
import { CopyOutlined } from "@ant-design/icons";
import CopyToClipboard from "./CopyToClipboard";
import ajax from "../api/ajax";
import { components } from "../api/schema";

const ServiceEndpoint: React.FC<{
  namespaceId?: number;
  projectId?: number;
}> = ({ namespaceId, projectId }) => {
  const [endpoints, setEndpoints] =
    useState<components["schemas"]["types.ServiceEndpoint"][]>();
  const enterFn = useCallback(
    (e: MouseEvent<SVGElement>) => {
      if (namespaceId) {
        ajax
          .GET("/api/endpoints/namespaces/{namespaceId}", {
            params: { path: { namespaceId } },
          })
          .then(({ data }) => {
            data && setEndpoints(data.items);
          });
      }
      if (projectId) {
        ajax
          .GET("/api/endpoints/projects/{projectId}", {
            params: { path: { projectId: projectId } },
          })
          .then(({ data }) => {
            data && setEndpoints(data.items);
          });
      }
    },
    [namespaceId, projectId]
  );

  return (
    <Popover
      placement="right"
      title={"链接"}
      content={endpoints?.map((v, k) => (
        <div key={v.url} onClick={(e) => e.stopPropagation()}>
          <span style={{ marginRight: 5 }}>
            {v.name}
            {v.portName ? `(${v.portName})` : ""}:
          </span>
          {v.url.startsWith("http") ? (
            <a href={v.url} target="_blank" style={{ marginRight: 10 }}>
              {v.url}
            </a>
          ) : (
            <span style={{ marginRight: 10 }}>{v.url}</span>
          )}

          <CopyToClipboard text={v.url} successText="已复制！">
            <CopyOutlined />
          </CopyToClipboard>
        </div>
      ))}
      trigger="hover"
    >
      <svg
        xmlns="http://www.w3.org/2000/svg"
        className="h-6 w-6"
        fill="none"
        viewBox="0 0 24 24"
        style={{ width: 18, height: 18, flexShrink: 0 }}
        stroke="currentColor"
        onMouseEnter={enterFn}
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth={2}
          d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"
        />
      </svg>
    </Popover>
  );
};

export default memo(ServiceEndpoint);
