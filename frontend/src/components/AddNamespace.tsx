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
    }

    if (!new RegExp(/^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/).test(namespace)) {
      message.error("名称空间格式有问题, (e.g. 'my-name',  or '123-abc')");
    }

    createNamespace(namespace)
      .then(({ data }) => {
        const {
          data: { id, name },
        } = data;
        onCreated({ id: id, name: name });
        message.success("名称空间创建成功");
        setIsVisible(false);
      })
      .catch((e) => message.error(e.message));
  }, [namespace, onCreated]);

  return (
    <>
      <Affix offsetTop={80} style={{ position: "absolute", right: "10px" }}>
        <Button
          size="large"
          type="primary"
          shape="circle"
          icon={<PlusOutlined />}
          onClick={() => setIsVisible(true)}
        />
      </Affix>
      <Modal
        title="创建项目空间"
        visible={isVisible}
        onOk={() => submit()}
        onCancel={() => {
          setIsVisible(false);
          setNamespace("");
        }}
      >
        <Input
          placeholder="空间名称"
          onChange={(e) => {
            setNamespace(e.target.value);
          }}
        />
      </Modal>
    </>
  );
};

export default memo(AddNamespace);
