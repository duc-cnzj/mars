import React, { useState, memo, useCallback } from "react";
import {
  Card,
  Popconfirm,
  Row,
  Spin,
  Col,
  message,
  Tooltip,
  Button,
  Space,
  Popover,
  Form,
  Input,
  Modal,
  Select,
} from "antd";
import "../pkg/DraggableModal/index.css";
import {
  CloseOutlined,
  EditFilled,
  LockOutlined,
  UnlockOutlined,
} from "@ant-design/icons";
import ServiceEndpoint from "./ServiceEndpoint";
import ProjectDetail from "./ProjectDetail";
import CreateProjectModal from "./CreateProjectModal";
import { copy } from "../utils/copy";
import styled from "@emotion/styled";
import { useAuth } from "../contexts/auth";
import { components } from "../api/schema";
import ajax from "../api/ajax";
import IconFont from "./Icon";
import TextArea from "antd/es/input/TextArea";
import { css } from "@emotion/css";
import { SelectProps } from "antd/lib";

const Item: React.FC<{
  item: components["schemas"]["types.NamespaceModel"];
  onNamespaceDeleted: () => void;
  onFavorite: (nsID: number, favorite: boolean) => void;
  loading: boolean;
  reload: () => void;
}> = ({ item, onNamespaceDeleted, loading, onFavorite, reload }) => {
  const [cpuAndMemory, setCpuAndMemory] = useState({ cpu: "", memory: "" });
  const { isAdmin } = useAuth();
  const [deleting, setDeleting] = useState<boolean>(false);

  const [editDesc, setEditDesc] = useState(false);
  const [popoverVisible, setPopoverVisible] = useState(false);

  return (
    <Card
      style={{ height: "100%" }}
      title={
        <CardTitle>
          <Space size={"small"}>
            <IconFont
              onClick={() => onFavorite(item.id, !item.favorite)}
              name="#icon-wodeguanzhu"
              className={css`
                transition: all 0.3s ease;
                &:hover {
                  transform: scale(1.2);
                }
              `}
              style={{
                color: !item.favorite ? "gray" : "#a78bfa",
                cursor: "pointer",
              }}
            />
            <Tooltip
              title={<span style={{ fontSize: 10 }}>id: {item.id}</span>}
            >
              <TitleNamespace onClick={() => copy(item.id, "已复制 id")}>
                项目空间: <TitleNamespaceName>{item.name}</TitleNamespaceName>
              </TitleNamespace>
            </Tooltip>
            <NamespacePrivate reload={reload} item={item} />
            <TitleSubItem>
              <Tooltip
                onOpenChange={(visible: boolean) => {
                  if (visible) {
                    ajax
                      .GET("/api/metrics/namespace/{namespaceId}/cpu_memory", {
                        params: { path: { namespaceId: item.id } },
                      })
                      .then(({ data }) => {
                        data &&
                          setCpuAndMemory({
                            cpu: data.cpu,
                            memory: data.memory,
                          });
                      });
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
            </TitleSubItem>
            <TitleSubItem>
              <ServiceEndpoint namespaceId={item.id} />
            </TitleSubItem>
            <Popover
              destroyTooltipOnHide
              trigger="click"
              onOpenChange={(v) => {
                setPopoverVisible(v);
                if (!v) {
                  setEditDesc(false);
                }
              }}
              content={
                editDesc ? (
                  <Form
                    name="basic"
                    style={{ width: 300 }}
                    initialValues={{ id: item.id, desc: item.description }}
                    autoComplete="off"
                    onFinish={(values) => {
                      console.log(values);
                      ajax
                        .POST("/api/namespaces/{id}/update_desc", {
                          params: { path: { id: values.id } },
                          body: values,
                        })
                        .then(({ error }) => {
                          if (error) {
                            return;
                          }
                          reload();
                          setEditDesc(false);
                          message.success("修改成功");
                        });
                    }}
                  >
                    <Form.Item<
                      components["schemas"]["namespace.UpdateDescRequest"]
                    >
                      name="id"
                      hidden
                    >
                      <Input />
                    </Form.Item>
                    <Form.Item<
                      components["schemas"]["namespace.UpdateDescRequest"]
                    >
                      name="desc"
                      style={{ marginBottom: 5 }}
                    >
                      <TextArea style={{ fontSize: 10 }} rows={5} />
                    </Form.Item>
                    <div style={{ textAlign: "right" }}>
                      <Button
                        style={{ fontSize: 10 }}
                        size="small"
                        htmlType="submit"
                      >
                        提交
                      </Button>
                    </div>
                  </Form>
                ) : (
                  <div
                    style={{
                      width: 300,
                      fontSize: 12,
                      textAlign: "right",
                    }}
                  >
                    <div style={{ textAlign: "left" }}>{item.description}</div>
                    <EditFilled
                      style={{ textAlign: "right" }}
                      onClick={() => setEditDesc(true)}
                    />
                  </div>
                )
              }
            >
              <div
                style={{
                  fontSize: 10,
                  width: "100px",
                  color: "gray",
                  fontWeight: "normal",
                  whiteSpace: "nowrap",
                  overflow: "hidden",
                  textOverflow: "ellipsis",
                  cursor: "pointer",
                }}
              >
                {item.description ? (
                  <div>{item.description}</div>
                ) : (
                  <div
                    className={css`
                      opacity: ${popoverVisible ? 1 : 0};
                      transition: all 0.2s ease;
                      &:hover {
                        opacity: 1;
                      }
                    `}
                  >
                    暂无描述，点击添加
                  </div>
                )}
              </div>
            </Popover>
          </Space>
        </CardTitle>
      }
      extra={
        isAdmin() ? (
          <Popconfirm
            title={`确定要删除 '${item.name}' 这个名称空间吗？`}
            okText="Yes"
            cancelText="No"
            onConfirm={() => {
              setDeleting(true);
              ajax
                .DELETE("/api/namespaces/{id}", {
                  params: { path: { id: item.id } },
                })
                .then(({ error }) => {
                  if (error) {
                    message.error(error.message);
                    return;
                  }
                  message.success("删除成功");
                  onNamespaceDeleted();
                });
            }}
          >
            <Button type="link" size="middle" icon={<CloseOutlined />} />
          </Popconfirm>
        ) : (
          <></>
        )
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

const CardTitle = styled.div`
  display: flex;
  align-items: center;
  justify-content: space-between;
`;

const TitleNamespace = styled.div`
  font-size: 12px;
  font-weight: normal;
`;

const TitleNamespaceName = styled.span`
  font-family: "Gill Sans", "Gill Sans MT", Calibri, "Trebuchet MS", sans-serif;
  font-weight: 500;
  font-size: 18px;
`;

const TitleSubItem = styled.div`
  display: flex;
  align-items: center;
  justify-content: center;
`;

const NamespacePrivate: React.FC<{
  item: components["schemas"]["types.NamespaceModel"];
  reload: () => void;
}> = ({ item, reload }) => {
  const { user, isAdmin } = useAuth();
  const [isModalOpen, setIsModalOpen] = useState(false);

  const [options, setOptions] = useState<SelectProps["options"]>(
    item.members.map((v) => ({
      value: v.email,
      label: v.email,
    })),
  );

  const isOwned = useCallback(() => {
    return isAdmin() || item.creatorEmail === user.email;
  }, [isAdmin, item.creatorEmail, user.email]);

  if (!isOwned()) {
    return <div>{item.private ? <UnlockOutlined /> : <LockOutlined />}</div>;
  }

  return (
    <Space>
      <Modal
        destroyOnClose
        title="修改空间成员"
        open={isModalOpen}
        onOk={() => {
          ajax
            .POST("/api/namespaces/sync_members", {
              body: {
                id: item.id,
                emails: options
                  ? options.map((v): string => String(v.value))
                  : [],
              },
            })
            .then(() => {
              message.success("修改成功");
              reload();
              setIsModalOpen(false);
            });
        }}
        onCancel={() => {
          setIsModalOpen(false);
        }}
      >
        <Select
          mode="tags"
          style={{ width: "100%" }}
          placeholder="添加成员"
          onChange={(options) => {
            setOptions(options.map((v: any) => ({ label: v, value: v })));
          }}
          defaultValue={options?.map((v) => v.value)}
          options={options}
        />
      </Modal>
      <Tooltip overlayStyle={{ fontSize: 12 }} title="你是此空间管理员">
        <IconFont name="#icon-crown" />
      </Tooltip>
      <Popconfirm
        title="修改访问权限"
        description={`确定要修改成 ${item.private ? "public" : "private"} 吗`}
        onConfirm={() => {
          ajax
            .POST("/api/namespaces/update_private", {
              body: { id: item.id, private: !item.private },
            })
            .then(() => {
              message.success("修改成功");
              reload();
            });
        }}
        okText="Yes"
        cancelText="No"
      >
        {!item.private ? <UnlockOutlined /> : <LockOutlined />}
      </Popconfirm>
      {item.private && (
        <Tooltip overlayStyle={{ fontSize: 12 }} title="成员管理">
          <IconFont
            name="#icon-chengyuanguanli_huaban1"
            style={{ cursor: "pointer" }}
            onClick={() => setIsModalOpen(true)}
          />
        </Tooltip>
      )}
    </Space>
  );
};
