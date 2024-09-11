import React, { memo, useEffect, useState } from "react";
import { Skeleton, Button, Modal, message, Spin } from "antd";
import {
  BranchesOutlined,
  PushpinOutlined,
  ExclamationCircleOutlined,
  SettingOutlined,
  FireOutlined,
  FieldNumberOutlined,
  LinkOutlined,
} from "@ant-design/icons";
import { PrismLight as SyntaxHighlighter } from "react-syntax-highlighter";
import { materialDark } from "react-syntax-highlighter/dist/esm/styles/prism";
import yaml from "react-syntax-highlighter/dist/esm/languages/prism/yaml";
import { components } from "../api/schema";
import ajax from "../api/ajax";

SyntaxHighlighter.registerLanguage("yaml", yaml);

const { confirm } = Modal;

const DetailTab: React.FC<{
  detail: components["schemas"]["types.ProjectModel"];
  onDeleted: () => void;
}> = ({ detail, onDeleted }) => {
  const [loading, setLoading] = useState<boolean>(false);
  const [cpuMemEndpointsLoading, setCpuMemEndpointsLoading] =
    useState<boolean>(true);
  const [cpuMemEndpoints, setCpuMemEndpoints] = useState<{
    cpu: string;
    mem: string;
    urls: components["schemas"]["types.ServiceEndpoint"][];
  }>({ cpu: "", mem: "", urls: [] });
  useEffect(() => {
    setCpuMemEndpointsLoading(true);
    ajax
      .GET("/api/projects/{id}/memory_cpu_and_endpoints", {
        params: { path: { id: detail.id } },
      })
      .then(({ data, error }) => {
        setCpuMemEndpointsLoading(false);
        if (error) {
          return;
        }
        setCpuMemEndpoints({
          cpu: data.cpu,
          mem: data.memory,
          urls: data.urls,
        });
      });
  }, [detail.id]);

  return detail ? (
    <div style={{ height: "100%", overflowY: "auto" }}>
      <div
        style={{
          display: "flex",
          alignItems: "center",
        }}
      >
        <FieldNumberOutlined
          style={{
            display: "flex",
            alignItems: "center",
            height: 30,
            marginRight: 4,
            fontSize: 20,
          }}
        />
        {detail.id}
      </div>
      <div
        style={{
          display: "flex",
          alignItems: "center",
        }}
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className="h-5 w-5"
          viewBox="0 0 20 20"
          fill="currentColor"
          style={{
            width: 20,
            height: 20,
            marginRight: 4,
          }}
        >
          <path d="M13 7H7v6h6V7z" />
          <path
            fillRule="evenodd"
            d="M7 2a1 1 0 012 0v1h2V2a1 1 0 112 0v1h2a2 2 0 012 2v2h1a1 1 0 110 2h-1v2h1a1 1 0 110 2h-1v2a2 2 0 01-2 2h-2v1a1 1 0 11-2 0v-1H9v1a1 1 0 11-2 0v-1H5a2 2 0 01-2-2v-2H2a1 1 0 110-2h1V9H2a1 1 0 010-2h1V5a2 2 0 012-2h2V2zM5 5h10v10H5V5z"
            clipRule="evenodd"
          />
        </svg>
        <div style={{ fontWeight: 700 }}>
          cpu:{" "}
          {!cpuMemEndpointsLoading ? (
            <span className="detail-data">{cpuMemEndpoints.cpu}</span>
          ) : (
            <Spin />
          )}
        </div>
      </div>

      <div
        style={{
          display: "flex",
          alignItems: "center",
        }}
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className="h-5 w-5"
          viewBox="0 0 20 20"
          fill="currentColor"
          style={{
            width: 20,
            height: 20,
            marginRight: 4,
          }}
        >
          <path d="M13 7H7v6h6V7z" />
          <path
            fillRule="evenodd"
            d="M7 2a1 1 0 012 0v1h2V2a1 1 0 112 0v1h2a2 2 0 012 2v2h1a1 1 0 110 2h-1v2h1a1 1 0 110 2h-1v2a2 2 0 01-2 2h-2v1a1 1 0 11-2 0v-1H9v1a1 1 0 11-2 0v-1H5a2 2 0 01-2-2v-2H2a1 1 0 110-2h1V9H2a1 1 0 010-2h1V5a2 2 0 012-2h2V2zM5 5h10v10H5V5z"
            clipRule="evenodd"
          />
        </svg>
        <div style={{ fontWeight: 700 }}>
          memory:{" "}
          {!cpuMemEndpointsLoading ? (
            <span className="detail-data">{cpuMemEndpoints.mem}</span>
          ) : (
            <Spin />
          )}
        </div>
      </div>

      <div
        style={{
          display: "flex",
          alignItems: "center",
        }}
      >
        <BranchesOutlined
          style={{
            width: 20,
            height: 20,
            marginRight: 4,
            fontSize: 16,
          }}
        />
        <div style={{ fontWeight: 700 }}>
          分支: <span className="detail-data">{detail.gitBranch}</span>
        </div>
      </div>

      <div>
        <div
          style={{
            display: "flex",
            alignItems: "center",
          }}
        >
          <LinkOutlined
            style={{
              width: 20,
              height: 20,
              marginRight: 4,
              fontSize: 16,
            }}
          />
          <div style={{ fontWeight: 700 }}>地址:</div>
        </div>
        <ul style={{ listStyle: "none", padding: "0 0 0 1.5em", margin: 0 }}>
          {!cpuMemEndpointsLoading ? (
            cpuMemEndpoints.urls.map((item, index) => (
              <li key={index}>
                {index + 1}.
                {item.url.startsWith("http") ? (
                  <a href={item.url} target="_blank" className="detail-data">
                    {item.url}
                    {item.portName ? `(${item.portName})` : ""}
                  </a>
                ) : (
                  <span>
                    {item.url}
                    {item.portName ? `(${item.portName})` : ""}
                  </span>
                )}
              </li>
            ))
          ) : (
            <Spin size={"small"} />
          )}
        </ul>
      </div>

      <div>
        <div
          style={{
            display: "flex",
            alignItems: "center",
          }}
        >
          <SettingOutlined
            style={{
              width: 20,
              height: 20,
              marginRight: 4,
              fontSize: 16,
            }}
          />
          <div style={{ fontWeight: 700 }}>容器镜像:</div>
        </div>
        <div style={{ marginLeft: 20 }}>
          {detail.dockerImage?.map((v, idx) => <div key={idx}>{v}</div>)}
        </div>
      </div>
      <div
        style={{
          display: "flex",
          alignItems: "center",
        }}
      >
        <PushpinOutlined
          style={{
            width: 20,
            height: 20,
            marginRight: 4,
            fontSize: 16,
          }}
        />
        <div style={{ fontWeight: 700 }}>
          提交:
          <span className="detail-data">
            <a href={detail.gitCommitWebUrl} target="_blank">
              {detail.gitCommitTitle}
            </a>
            by {detail.gitCommitAuthor} 于 {detail.gitCommitDate}
          </span>
        </div>
      </div>

      <div
        style={{
          display: "flex",
          alignItems: "center",
        }}
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className="h-6 w-6"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          style={{
            width: 20,
            height: 20,
            marginRight: 4,
          }}
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
          />
        </svg>
        <div style={{ fontWeight: 700 }}>
          部署日期:{" "}
          <span className="detail-data">{detail.humanizeCreatedAt}</span>
        </div>
      </div>
      <div
        style={{
          display: "flex",
          alignItems: "center",
        }}
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          className="h-6 w-6"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
          style={{
            width: 20,
            height: 20,
            marginRight: 4,
          }}
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
          />
        </svg>
        <div style={{ fontWeight: 700 }}>
          更新日期:{" "}
          <span className="detail-data">{detail.humanizeUpdatedAt}</span>
        </div>
      </div>

      <div>
        <div style={{ fontWeight: 700 }}>
          <FireOutlined
            style={{
              width: 20,
              height: 20,
              marginRight: 4,
              fontSize: 16,
            }}
          />
          相关配置
        </div>
        <details>
          <summary style={{ cursor: "pointer" }}>展开查看</summary>
          <SyntaxHighlighter
            language="yaml"
            style={materialDark}
            customStyle={{
              fontFamily: '"Fira code", "Fira Mono", monospace',
              fontSize: 12,
            }}
          >
            {detail.overrideValues}
          </SyntaxHighlighter>
        </details>
      </div>

      <Button
        onClick={() =>
          confirm({
            okText: "确定",
            cancelText: "取消",
            title: (
              <div style={{ fontSize: 14 }}>
                确定要删除空间{" "}
                <span style={{ color: "red" }}>{detail.namespace?.name}</span>{" "}
                下的 <span style={{ color: "red" }}>{detail.name}</span> 项目吗?
              </div>
            ),
            icon: <ExclamationCircleOutlined />,
            onOk() {
              setLoading(true);
              ajax
                .DELETE("/api/projects/{id}", {
                  params: { path: { id: detail.id } },
                })
                .then(() => {
                  message.success("删除成功");
                  setLoading(false);
                  onDeleted();
                })
                .catch((e) => {
                  console.log(e);
                  message.error("删除失败");
                  setLoading(false);
                });
            },
            onCancel() {},
          })
        }
        disabled={loading}
        loading={loading}
        type="primary"
        style={{ marginTop: 10, marginBottom: 40 }}
        size="middle"
        danger
      >
        删除项目
      </Button>
    </div>
  ) : (
    <Skeleton />
  );
};

export default memo(DetailTab);
