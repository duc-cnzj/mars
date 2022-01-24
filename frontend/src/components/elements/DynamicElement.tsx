import React, { useState, useEffect, memo } from "react";
import {
  Switch,
  Form,
  Input,
  Button,
  FormInstance,
  Select,
  SelectProps,
  InputNumber,
} from "antd";
import { MinusCircleOutlined, PlusOutlined } from "@ant-design/icons";
import pb from "../../api/compiled";

const DynamicElement: React.FC<{
  form: FormInstance;
  disabled: boolean;
}> = ({ form, disabled }) => {
  const [type, setType] = useState<{ [key: number]: number }>({});
  useEffect(() => {
    if (form) {
      (form.getFieldValue("elements") as Array<any>).forEach((e, i) => {
        setType((t) => ({
          ...t,
          [i]: e.type,
        }));
      });
    }
  }, [form]);

  return (
    <Form.List name="elements">
      {(fields, { add, remove }) => (
        <>
          {fields.map((field) => {
            return (
              <div key={field.name} className="dynamic-element">
                <div className="dynamic-element__wrapper">
                  <div style={{ display: "flex", width: "100%" }}>
                    <Form.Item
                      style={{ width: "100%" }}
                      label={"字段路径"}
                      name={[field.name, "path"]}
                      rules={[{ required: true, message: "字段路径必填" }]}
                    >
                      <Input disabled={disabled} placeholder="字段路径" />
                    </Form.Item>
                    <Form.Item
                      style={{ width: "100%" }}
                      label="表单类型"
                      name={[field.name, "type"]}
                      rules={[{ required: true, message: "表单类型必填" }]}
                    >
                      <Select
                        disabled={disabled}
                        onChange={(v) => {
                          setType((t) => ({ ...t, [field.key]: v }));
                          form.setFieldsValue(["elements", Number(v)]);
                        }}
                      >
                        <Select.Option
                          value={pb.ElementType.ElementTypeUnknown}
                        >
                          未设置
                        </Select.Option>
                        <Select.Option value={pb.ElementType.ElementTypeInput}>
                          Input
                        </Select.Option>
                        <Select.Option
                          value={pb.ElementType.ElementTypeInputNumber}
                        >
                          InputNumber
                        </Select.Option>
                        <Select.Option value={pb.ElementType.ElementTypeRadio}>
                          Radio
                        </Select.Option>
                        <Select.Option value={pb.ElementType.ElementTypeSelect}>
                          Select
                        </Select.Option>
                        <Select.Option value={pb.ElementType.ElementTypeSwitch}>
                          Switch
                        </Select.Option>
                      </Select>
                    </Form.Item>
                    <Form.Item
                      style={{ width: "100%" }}
                      label="默认值"
                      name={[field.name, "default"]}
                    >
                      <DefaultValueElement
                        disabled={disabled}
                        type={type[field.key] ? type[field.key] : 0}
                      />
                    </Form.Item>
                  </div>
                  <div style={{ display: "flex" }}>
                    <Form.Item
                      style={{ width: "100%" }}
                      label="字段描述"
                      name={[field.name, "description"]}
                      rules={[{ required: true, message: "字段描述必填" }]}
                    >
                      <Input disabled={disabled} placeholder="字段描述" />
                    </Form.Item>

                    <Form.Item
                      hidden={
                        !(
                          type[field.key] &&
                          (type[field.key] ===
                            pb.ElementType.ElementTypeSelect ||
                            type[field.key] === pb.ElementType.ElementTypeRadio)
                        )
                      }
                      style={{ width: "100%" }}
                      label="选择器"
                      name={[field.name, "select_values"]}
                    >
                      <MySelect disabled={disabled} />
                    </Form.Item>
                  </div>
                  {!disabled ? (
                    <MinusCircleOutlined onClick={() => remove(field.name)} />
                  ) : (
                    <></>
                  )}
                </div>
              </div>
            );
          })}
          <Form.Item>
            <Button
              disabled={disabled}
              type="dashed"
              onClick={() => add()}
              block
              icon={<PlusOutlined />}
            >
              添加自定义配置
            </Button>
          </Form.Item>
        </>
      )}
    </Form.List>
  );
};

const MySelect: React.FC<
  {
    value?: string[];
    onChange?: (value: string) => void;
  } & SelectProps
> = ({ value, onChange }) => {
  return (
    <Select
      mode="tags"
      value={value}
      style={{ width: "100%" }}
      tokenSeparators={[","]}
      onChange={onChange}
    />
  );
};

const DefaultValueElement: React.FC<{
  value?: any;
  onChange?: (v: any) => void;
  type: pb.ElementType;
  disabled: boolean;
}> = ({ type, disabled, value, onChange }) => {
  switch (type) {
    case pb.ElementType.ElementTypeInputNumber:
      return (
        <InputNumber
          disabled={disabled}
          value={value}
          onChange={onChange}
          placeholder="默认值"
        />
      );
    case pb.ElementType.ElementTypeSwitch:
      return (
        <Switch
          disabled={disabled}
          checked={value}
          onChange={onChange}
          defaultChecked={false}
        />
      );
    default:
      return (
        <Input
          disabled={disabled}
          value={value}
          onChange={onChange}
          placeholder="默认值"
        />
      );
  }
};

export default memo(DynamicElement);
