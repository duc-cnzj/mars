import React, { memo, useEffect, useState } from "react";
import { Skeleton, Button, Modal, message, Spin, Space } from "antd";
import {
  BranchesOutlined,
  PushpinOutlined,
  ExclamationCircleOutlined,
  SettingOutlined,
  FireOutlined,
  FieldNumberOutlined,
  LinkOutlined,
  ScheduleOutlined,
  DotChartOutlined,
  PieChartOutlined,
} from "@ant-design/icons";
import { PrismLight as SyntaxHighlighter } from "react-syntax-highlighter";
import { materialDark } from "react-syntax-highlighter/dist/esm/styles/prism";
import yaml from "react-syntax-highlighter/dist/esm/languages/prism/yaml";
import { components } from "../api/schema";
import ajax from "../api/ajax";
import { IconBaseProps } from "@ant-design/icons/lib/components/Icon";
import styled from "@emotion/styled";

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
    <Space direction="vertical" style={{ width: "100%" }}>
      <LineItem icon={FieldNumberOutlined} title={detail.id} />
      <LineItem
        icon={DotChartOutlined}
        title={"此项目 cpu"}
        children={
          !cpuMemEndpointsLoading ? (
            <span className="detail-data">{cpuMemEndpoints.cpu}</span>
          ) : (
            <Spin />
          )
        }
      />
      <LineItem
        icon={PieChartOutlined}
        title={"此项目 memory"}
        children={
          !cpuMemEndpointsLoading ? (
            <span className="detail-data">{cpuMemEndpoints.mem}</span>
          ) : (
            <Spin />
          )
        }
      />
      <div>
        <LineItem icon={LinkOutlined} title={"地址"} />
        <ul style={{ listStyle: "none", padding: "0 0 0 1.5rem", margin: 0 }}>
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

      {detail.repo.needGitRepo && (
        <>
          <LineItem
            icon={BranchesOutlined}
            title={"分支"}
            children={<span className="detail-data">{detail.gitBranch}</span>}
          />

          <LineItem
            icon={PushpinOutlined}
            title={"提交"}
            children={
              <span className="detail-data">
                <a href={detail.gitCommitWebUrl} target="_blank">
                  {detail.gitCommitTitle}
                </a>
                by {detail.gitCommitAuthor} 于 {detail.gitCommitDate}
              </span>
            }
          />
        </>
      )}

      <div>
        <LineItem icon={SettingOutlined} title={"容器镜像"} />
        <div style={{ marginLeft: "1.5rem" }}>
          {detail.dockerImage?.map((v, idx) => <div key={idx}>{v}</div>)}
        </div>
      </div>
      <LineItem
        icon={ScheduleOutlined}
        title={"部署日期"}
        children={
          <span className="detail-data">{detail.humanizeCreatedAt}</span>
        }
      />
      <LineItem
        icon={ScheduleOutlined}
        title={"更新日期"}
        children={
          <span className="detail-data">{detail.humanizeUpdatedAt}</span>
        }
      />
      <div>
        <LineItem
          icon={FireOutlined}
          title={"相关配置"}
          children={
            <span className="detail-data">{detail.humanizeUpdatedAt}</span>
          }
        />
        <details style={{ marginTop: 3, marginLeft: "1.5rem" }}>
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
    </Space>
  ) : (
    <Skeleton />
  );
};

export default memo(DetailTab);

const Label = styled.div`
  font-weight: 700;
`;

const LineItem: React.FC<{
  icon: React.ComponentType<IconBaseProps>;
  title: string | number;
  children?: React.ReactNode;
}> = ({ icon, title, children }) => {
  return (
    <div
      style={{
        display: "flex",
        alignItems: "center",
      }}
    >
      <IconWithStyle IconComponent={icon} />
      <Label>
        {title}
        {children && <>: {children}</>}
      </Label>
    </div>
  );
};

const IconWithStyle: React.FC<{
  IconComponent: React.ComponentType<IconBaseProps>;
}> = ({ IconComponent }) => {
  return (
    <IconComponent
      style={{
        width: 20,
        height: 20,
        marginRight: 4,
        fontSize: 16,
      }}
    />
  );
};
