import React, { useState, useCallback, memo } from "react";
import { Affix, Button, Modal, Input, message } from "antd";
import { PlusOutlined } from "@ant-design/icons";
import { createNamespace } from "../api/namespace";

interface IProps {
  onCreated: ({ id, name }: { id: number; name: string }) => void;
}

const AddNamespace: React.FC<IProps> = ({ onCreated }) => {
  const [isVisible, setIsVisible] = useState<boolean>(false);
  const [namespace, setNamespace] = useState<string>("");
  const submit = useCallback(() => {
    if (!namespace) {
      message.error("名称空间必填");
      return;
    }

    if (!new RegExp(/^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/).test(namespace)) {
      message.error("名称空间格式有问题, (e.g. 'my-name',  or '123-abc')");
      return;
    }

    createNamespace(namespace)
      .then(({ data }) => {
        data && onCreated({ id: data.id, name: data.name });
        message.success("名称空间创建成功");
        setIsVisible(false);
        setNamespace("");
      })
      .catch((e) => message.error(e.response.data.message));
  }, [namespace, onCreated]);

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
        visible={isVisible}
        onOk={() => submit()}
        okText={"创建"}
        cancelText={"取消"}
        onCancel={() => {
          setIsVisible(false);
          setNamespace("");
        }}
      >
        <Input
          placeholder="空间名称"
          value={namespace}
          onChange={(e) => {
            setNamespace(e.target.value);
          }}
        />
      </Modal>
    </>
  );
};

export default memo(AddNamespace);
