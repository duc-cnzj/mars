import React, { useState, useCallback, memo } from "react";
import { Affix, Button, Modal, Input, message, Form } from "antd";
import { PlusOutlined } from "@ant-design/icons";
import ajax from "../api/ajax";
import TextArea from "antd/es/input/TextArea";
import { components } from "../api/schema";

interface IProps {
  onCreated: ({ id, name }: { id: number; name: string }) => void;
}

const AddNamespace: React.FC<IProps> = ({ onCreated }) => {
  const [isVisible, setIsVisible] = useState<boolean>(false);
  const [form] = Form.useForm();
  const submit = useCallback(
    (values: any) => {
      ajax.POST("/api/namespaces", { body: values }).then(({ data, error }) => {
        if (error) {
          message.error(error.message);
          return;
        }
        data && onCreated({ id: data.item.id, name: data.item.name });
        message.success("名称空间创建成功");
        setIsVisible(false);
        form.resetFields();
      });
    },
    [onCreated, form],
  );

  return (
    <>
      <Affix offsetTop={80} style={{ position: "absolute", right: "10px" }}>
        <Button
          size="large"
          type="primary"
          shape="circle"
          className="add-namespace__button"
          icon={<PlusOutlined />}
          onClick={() => setIsVisible(true)}
        />
      </Affix>
      <Modal
        title="创建项目空间"
        open={isVisible}
        onOk={() => form.submit()}
        okText={"创建"}
        cancelText={"取消"}
        onCancel={() => {
          setIsVisible(false);
          form.resetFields();
        }}
      >
        <Form
          name="basic"
          form={form}
          initialValues={{ remember: true }}
          onFinish={submit}
          autoComplete="off"
        >
          <Form.Item<components["schemas"]["namespace.CreateRequest"]>
            name="namespace"
            rules={[
              { required: true, message: "空间名称必填" },
              () => ({
                validator(_, value) {
                  if (
                    !!value &&
                    new RegExp(/^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/).test(value)
                  ) {
                    return Promise.resolve();
                  }
                  return Promise.reject(
                    new Error(
                      "名称空间格式有问题, (e.g. 'my-name',  or '123-abc')",
                    ),
                  );
                },
              }),
            ]}
          >
            <Input placeholder="空间名称" />
          </Form.Item>

          <Form.Item<
            components["schemas"]["namespace.CreateRequest"]
          > name="description">
            <TextArea role="2" placeholder="输入描述..." />
          </Form.Item>
        </Form>
      </Modal>
    </>
  );
};

export default memo(AddNamespace);
