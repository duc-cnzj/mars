import React, { memo, useEffect, useState } from "react";
import { Alert } from "antd";
import { ClockCircleOutlined } from "@ant-design/icons";
import ajax from "../api/ajax";

// gray 标记 manual（等待手动触发）走灰色卡片，不用 warning 的黄色。
const pipelines: {
  [status: string]: {
    type: "error" | "success" | "warning" | "info";
    message: string;
    gray?: boolean;
  };
} = {
  failed: {
    type: "error",
    message: "pipeline 执行失败",
  },
  running: {
    type: "warning",
    message: "pipeline 还在执行中",
  },
  manual: {
    type: "warning",
    message: "pipeline 等待手动触发",
    gray: true,
  },
  unknown: {
    type: "info",
    message: "pipeline 状态未知",
    gray: true,
  },
  success: {
    type: "success",
    message: "pipeline 执行成功",
  },
};

const PipelineInfo: React.FC<{
  repoId: number;
  branch: string;
  commit: string;
}> = ({ repoId, branch, commit }) => {
  const [info, setInfo] = useState<{
    message: string;
    web_url: string;
    type: "success" | "warning" | "error" | "info";
    gray?: boolean;
  } | null>();

  useEffect(() => {
    if (repoId && branch && commit) {
      ajax
        .GET(
          "/api/git/repos/{repoId}/branches/{branch}/commits/{commit}/pipeline_info",
          {
            params: {
              path: {
                repoId,
                branch,
                commit,
              },
            },
          },
        )
        .then(({ data, error }) => {
          if (error) {
            setInfo(null);
            return;
          }
          let p = pipelines[data.status];
          if (p) {
            setInfo({
              type: p.type,
              message: p.message,
              web_url: data.webUrl,
              gray: p.gray,
            });
          }
        });
      return;
    }
    setInfo(null);
  }, [repoId, branch, commit]);

  return (
    <>
      {info ? (
        <Alert
          style={{
            marginBottom: 10,
            ...(info.gray
              ? { background: "#f5f5f5", borderColor: "#d9d9d9" }
              : {}),
          }}
          icon={
            info.gray ? (
              <ClockCircleOutlined style={{ color: "#8c8c8c" }} />
            ) : undefined
          }
          message={
            <div style={{ display: "flex", alignItems: "center" }}>
              <span style={{ marginRight: 10 }}>{info.message}</span>
              <a
                href={info.web_url}
                target="_blank"
                style={{ display: "flex", alignItems: "center" }}
              >
                点击查看 pipeline 详细信息
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  className="h-6 w-6"
                  fill="none"
                  viewBox="0 0 24 24"
                  style={{ width: 18, height: 18, flexShrink: 0 }}
                  stroke="currentColor"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"
                  />
                </svg>
              </a>
            </div>
          }
          type={info.type}
          showIcon
        />
      ) : (
        ""
      )}
    </>
  );
};

export default memo(PipelineInfo);
