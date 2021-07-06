import React, { memo, useState } from "react";
import { deleteProject, ProjectDetail } from "../api/project";
import { Skeleton, Button, Modal, message } from "antd";
import SyntaxHighlighter from "react-syntax-highlighter";
import { monokaiSublime } from "react-syntax-highlighter/dist/esm/styles/hljs";

import {
  BranchesOutlined,
  PushpinOutlined,
  ExclamationCircleOutlined,
  SettingOutlined,
  FireOutlined,
  LinkOutlined,
} from "@ant-design/icons";
const { confirm } = Modal;

const DetailTab: React.FC<{ detail?: ProjectDetail; onDeleted: () => void }> =
  ({ detail, onDeleted }) => {
    const [loading, setLoading] = useState<boolean>(false);

    return detail ? (
      <>
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
            cpu: <span className="detail-data">{detail.cpu}</span>
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
            memory: <span className="detail-data">{detail.memory}</span>
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
            分支: <span className="detail-data">{detail.gitlab_branch}</span>
          </p>
        </div>
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
          <p>
            地址:
            {detail.urls.map((item) => (
              <a href={item} target="_blank" className="detail-data">
                {item}
              </a>
            ))}
          </p>
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
              <a href={detail.gitlab_commit_web_url} target="_blank">
                {detail.gitlab_commit_title}
              </a>
              by {detail.gitlab_commit_author}
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
            提交日期: <span className="detail-data">{detail.created_at}</span>
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
              style={monokaiSublime}
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
              title: `确定要删除'${detail.name}'项目吗?`,
              icon: <ExclamationCircleOutlined />,
              onOk() {
                setLoading(true);
                deleteProject(detail.namespace.id, detail.id)
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
                console.log("OK");
              },
              onCancel() {
                console.log("Cancel");
              },
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
      </>
    ) : (
      <Skeleton />
    );
  };

export default memo(DetailTab);
