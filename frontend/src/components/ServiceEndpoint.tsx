import React, { memo, MouseEvent, useCallback, useState } from "react";

import { Popover } from "antd";
import { CopyOutlined } from "@ant-design/icons";
import CopyToClipboard from "./CopyToClipboard";
import ajax from "../api/ajax";
import { components } from "../api/schema";
import IconFont from "./Icon";

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
    [namespaceId, projectId],
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
      <IconFont
        onMouseEnter={enterFn}
        style={{ cursor: "pointer" }}
        name="#icon-netlink"
      />
    </Popover>
  );
};

export default memo(ServiceEndpoint);
