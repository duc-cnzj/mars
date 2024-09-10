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
  Alert,
} from "antd";
import yaml from "js-yaml";
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
import { MyCodeMirror } from "./MyCodeMirror";
import DiffViewer from "./DiffViewer";

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
        <div style={{ marginTop: 8, marginBottom: 5 }}>
          <Row>
            <Col
              span={18}
              offset={3}
              style={{
                display: "flex",
                justifyContent: "center",
              }}
            >
              <IconFont
                onClick={() => onFavorite(item.id, !item.favorite)}
                name="#icon-wodeguanzhu"
                className={css`
                  margin-right: 3px;
                  margin-top: 3px;
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
            </Col>
            <Col span={3} style={{ textAlign: "right" }}>
              {isAdmin() && (
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
                  <Button
                    type="link"
                    danger
                    size="small"
                    style={{ border: 0 }}
                    icon={<CloseOutlined />}
                  />
                </Popconfirm>
              )}
            </Col>
          </Row>

          <Row>
            <Col
              offset={8}
              span={8}
              style={{
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
              }}
            >
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
                      <div style={{ textAlign: "left" }}>
                        {item.description}
                      </div>
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
                    color: "gray",
                    fontWeight: "normal",
                    whiteSpace: "nowrap",
                    overflow: "hidden",
                    textOverflow: "ellipsis",
                    cursor: "pointer",
                  }}
                >
                  {item.description ? (
                    item.description
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
            </Col>
            <Col span={8} style={{ textAlign: "right" }}>
              <Space style={{ marginRight: 4 }}>
                <NamespacePrivate reload={reload} item={item} />
                <Tooltip
                  onOpenChange={(visible: boolean) => {
                    if (visible) {
                      ajax
                        .GET(
                          "/api/metrics/namespace/{namespaceId}/cpu_memory",
                          {
                            params: { path: { namespaceId: item.id } },
                          },
                        )
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
                  <IconFont
                    style={{ cursor: "pointer" }}
                    name="#icon-dianboxindiantu"
                  />
                </Tooltip>
                <ServiceEndpoint namespaceId={item.id} />
              </Space>
            </Col>
          </Row>
        </div>
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

const TitleNamespace = styled.div`
  font-size: 12px;
  font-weight: normal;
`;

const TitleNamespaceName = styled.span`
  font-family: "Gill Sans", "Gill Sans MT", Calibri, "Trebuchet MS", sans-serif;
  font-weight: 500;
  font-size: 18px;
  margin-left: 3px;
`;

const NamespacePrivate: React.FC<{
  item: components["schemas"]["types.NamespaceModel"];
  reload: () => void;
}> = ({ item, reload }) => {
  const { user, isAdmin } = useAuth();
  const [isModalOpen, setIsModalOpen] = useState(false);

  const initValues = yaml.dump(item.members.map((v) => v.email));
  const [value, setValue] = useState(initValues);
  const isOwned = useCallback(() => {
    return isAdmin() || item.creatorEmail === user.email;
  }, [isAdmin, item.creatorEmail, user.email]);

  if (!isOwned()) {
    return (
      <Tooltip
        placement="top"
        overlayStyle={{ fontSize: 12 }}
        title={`此项目管理员是: ${item.creatorEmail}`}
      >
        {!item.private ? <UnlockOutlined /> : <LockOutlined />}
      </Tooltip>
    );
  }

  return (
    <Space>
      <Modal
        width={"60%"}
        destroyOnClose
        title="修改空间成员"
        open={isModalOpen}
        onOk={() => {
          try {
            let v = yaml.load(value);
            ajax
              .POST("/api/namespaces/sync_members", {
                body: {
                  id: item.id,
                  emails: (v as string[]) || [],
                },
              })
              .then(({ error }) => {
                if (error) {
                  message.error(error.message);
                  return;
                }
                message.success("修改成功");
                reload();
                setIsModalOpen(false);
              });
          } catch (e) {
            message.error("yaml 格式不正确");
          }
        }}
        onCancel={() => {
          setIsModalOpen(false);
        }}
      >
        <Form>
          <Row style={{ marginBottom: 5 }}>
            <Col span={24}>
              <Alert
                message={
                  <div>
                    <div>参考下面格式👇，自行修改，输入用户邮箱</div>
                    <div>- a@qq.com</div>
                    <div>- v@qq.com</div>
                  </div>
                }
                type="success"
              />
            </Col>
          </Row>
          <Row>
            <Col span={12}>
              <MyCodeMirror
                value={value}
                onChange={(v) => setValue(v)}
                mode="yaml"
              />
            </Col>
            <Col span={12}>
              <DiffViewer
                mode={"yaml"}
                styles={{
                  line: { fontSize: 12, lineHeight: 10 },
                  gutter: { padding: "0 5px", minWidth: 20 },
                  marker: { padding: "0 6px" },
                  diffContainer: {
                    height: "100%",
                    display: "block",
                    width: "100%",
                    overflowX: "auto",
                  },
                }}
                showDiffOnly={false}
                oldValue={initValues}
                newValue={value}
                splitView={false}
              />
            </Col>
          </Row>
        </Form>
      </Modal>
      <Tooltip
        overlayStyle={{ fontSize: 12 }}
        title={
          user.email !== item.creatorEmail
            ? `空间管理员邮箱: ${item.creatorEmail}`
            : "你是此空间管理员"
        }
      >
        <IconFont style={{ cursor: "pointer" }} name="#icon-crown" />
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
        <Tooltip overlayStyle={{ fontSize: 12 }} title="修改空间访问权限">
          {!item.private ? <UnlockOutlined /> : <LockOutlined />}
        </Tooltip>
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
