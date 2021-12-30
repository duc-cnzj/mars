import React, { useState, memo } from "react";
import {
  Card,
  Popconfirm,
  Row,
  Spin,
  Col,
  message,
  Tooltip,
  Button,
} from "antd";
import "../pkg/DraggableModal/index.css";
import { CloseOutlined } from "@ant-design/icons";
import { deleteNamespace, getNamespaceCpuAndMemory } from "../api/namespace";
import ServiceEndpoint from "./ServiceEndpoint";
import ProjectDetail from "./ProjectDetail";
import CreateProjectModal from "./CreateProjectModal";
import pb from "../api/compiled";

const Item: React.FC<{
  item: pb.NamespaceItem;
  onNamespaceDeleted: () => void;
  loading: boolean;
}> = ({ item, onNamespaceDeleted, loading }) => {
  const [cpuAndMemory, setCpuAndMemory] = useState({ cpu: "", memory: "" });

  const [deleting, setDeleting] = useState<boolean>(false);

  return (
    <Card
      title={
        <div className="title">
          <div className="title-left">
            <div className="title-namespace">
              项目空间: <span>{item.name}</span>
            </div>
            <div className="title-cpu-memory">
              <Tooltip
                onVisibleChange={(visible) => {
                  if (visible) {
                    getNamespaceCpuAndMemory({ namespace_id: item.id }).then(
                      (res) => {
                        setCpuAndMemory({
                          cpu: res.data.cpu,
                          memory: res.data.memory,
                        });
                      }
                    );
                  }
                }}
                title={
                  <div style={{ fontSize: "10px" }}>
                    <div>
                      <span>cpu: </span>
                      <span>{cpuAndMemory.cpu}</span>
                    </div>
                    <div>
                      <span>memory: </span>
                      <span>{cpuAndMemory.memory}</span>
                    </div>
                  </div>
                }
                trigger="hover"
              >
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  className="h-5 w-5"
                  viewBox="0 0 20 20"
                  style={{ width: "20px", height: "20px" }}
                  fill="currentColor"
                >
                  <path d="M13 7H7v6h6V7z" />
                  <path
                    fillRule="evenodd"
                    d="M7 2a1 1 0 012 0v1h2V2a1 1 0 112 0v1h2a2 2 0 012 2v2h1a1 1 0 110 2h-1v2h1a1 1 0 110 2h-1v2a2 2 0 01-2 2h-2v1a1 1 0 11-2 0v-1H9v1a1 1 0 11-2 0v-1H5a2 2 0 01-2-2v-2H2a1 1 0 110-2h1V9H2a1 1 0 010-2h1V5a2 2 0 012-2h2V2zM5 5h10v10H5V5z"
                    clipRule="evenodd"
                  />
                </svg>
              </Tooltip>
            </div>
            <div className="title-service-endpoint">
              <ServiceEndpoint namespaceId={item.id} />
            </div>
          </div>
        </div>
      }
      extra={
        <Popconfirm
          title={`确定要删除 '${item.name}' 这个名称空间吗？`}
          okText="Yes"
          cancelText="No"
          onConfirm={() => {
            setDeleting(true);
            deleteNamespace({ namespace_id: item.id })
              .then((res) => {
                message.success("删除成功");
                onNamespaceDeleted();
              })
              .catch((e) => message.error(e.response.data.message));
          }}
        >
          <Button type="link" size="middle" icon={<CloseOutlined />} />
        </Popconfirm>
      }
      bordered={false}
    >
      <Spin spinning={deleting || loading}>
        <Row gutter={[8, 8]}>
          {item.projects?.map((data) => (
            <Col key={data.id} md={12} xs={24} sm={24}>
              <ProjectDetail
                namespace={item.name}
                namespaceId={item.id}
                item={data}
              />
            </Col>
          ))}

          <Col md={12} xs={24} sm={24}>
            <CreateProjectModal namespaceId={item.id} />
          </Col>
        </Row>
      </Spin>
    </Card>
  );
};

export default memo(Item);
