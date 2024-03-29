import React, { memo, useState } from "react";
import { deleteProject } from "../api/project";
import { Skeleton, Button, Modal, message } from "antd";
import {
  BranchesOutlined,
  PushpinOutlined,
  ExclamationCircleOutlined,
  SettingOutlined,
  FireOutlined,
  FieldNumberOutlined,
  LinkOutlined,
} from "@ant-design/icons";
import pb from "../api/compiled";
import { PrismLight as SyntaxHighlighter } from "react-syntax-highlighter";
import { materialDark } from "react-syntax-highlighter/dist/esm/styles/prism";
import yaml from "react-syntax-highlighter/dist/esm/languages/prism/yaml";

SyntaxHighlighter.registerLanguage("yaml", yaml);

const { confirm } = Modal;

const DetailTab: React.FC<{
  detail: pb.types.ProjectModel;
  cpu: string;
  git_commit_web_url: string;
  git_commit_title: string;
  git_commit_author: string;
  git_commit_date: string;
  memory: string;
  urls: pb.types.ServiceEndpoint[];
  onDeleted: () => void;
}> = ({
  detail,
  onDeleted,
  cpu,
  memory,
  urls,
  git_commit_web_url,
  git_commit_title,
  git_commit_author,
  git_commit_date,
}) => {
  const [loading, setLoading] = useState<boolean>(false);

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
        <p>
          cpu: <span className="detail-data">{cpu}</span>
        </p>
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
        <p>
          memory: <span className="detail-data">{memory}</span>
        </p>
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
        <p>
          分支: <span className="detail-data">{detail.git_branch}</span>
        </p>
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
          <p>地址:</p>
        </div>
        <ul style={{ listStyle: "none", padding: "0 0 0 1.5em", margin: 0 }}>
          {urls.map((item, index) => (
            <li key={index}>
              {index + 1}.
              {item.url.startsWith("http") ? (
                <a href={item.url} target="_blank" className="detail-data">
                  {item.url}
                  {item.port_name ? `(${item.port_name})` : ""}
                </a>
              ) : (
                <span>
                  {item.url}
                  {item.port_name ? `(${item.port_name})` : ""}
                </span>
              )}
            </li>
          ))}
        </ul>
      </div>

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
        <p>
          容器镜像:
          <span className="detail-data">{detail.docker_image}</span>
        </p>
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
        <p>
          提交:
          <span className="detail-data">
            <a href={git_commit_web_url} target="_blank">
              {git_commit_title}
            </a>
            by {git_commit_author} 于 {git_commit_date}
          </span>
        </p>
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
        <p>
          部署日期:{" "}
          <span className="detail-data">{detail.humanize_created_at}</span>
        </p>
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
        <p>
          更新日期:{" "}
          <span className="detail-data">{detail.humanize_updated_at}</span>
        </p>
      </div>

      <div>
        <p>
          <FireOutlined
            style={{
              width: 20,
              height: 20,
              marginRight: 4,
              fontSize: 16,
            }}
          />
          相关配置
        </p>
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
            {detail.override_values}
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
              deleteProject(detail.id)
                .then((res) => {
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
        style={{ marginTop: 20 }}
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
