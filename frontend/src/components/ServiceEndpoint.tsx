import React, { memo, useState } from "react";
import { getServiceEndpoints } from "../api/namespace";
import { Popover, message } from "antd";
import { CopyOutlined } from "@ant-design/icons";
import { CopyToClipboard } from "react-copy-to-clipboard";
import pb from "../api/compiled";

const ServiceEndpoint: React.FC<{ namespaceId: number; projectName?: string }> =
  ({ namespaceId, projectName }) => {
    const [endpoints, setEndpoints] = useState<{ [name: string]: string[] }>(
      {}
    );

    return (
      <Popover
        placement="right"
        title={"链接"}
        content={Object.entries(endpoints)?.map(([k, v]) =>
          v.map((link) => (
            <div key={link} onClick={(e) => e.stopPropagation()}>
              <span style={{ marginRight: 5 }}>{k}:</span>
              <a href={link} target="_blank" style={{ marginRight: 10 }}>
                {link}
              </a>
              <CopyToClipboard
                text={link}
                onCopy={() => message.success("已复制！")}
              >
                <CopyOutlined />
              </CopyToClipboard>
            </div>
          ))
        )}
        trigger="hover"
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className="h-6 w-6"
          fill="none"
          viewBox="0 0 24 24"
          style={{ width: 18, height: 18, flexShrink: 0 }}
          stroke="currentColor"
          onMouseEnter={(e) => {
            getServiceEndpoints({project_name: String(projectName), namespace_id: namespaceId}).then((res) => {
              // @ts-ignore
                setEndpoints(res.data.data.map((name: string, item: pb.ServiceEndpointsResponse.Iitem) => {
                  return {
                      name: item.name
                  }
              }));
            });
          }}
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
