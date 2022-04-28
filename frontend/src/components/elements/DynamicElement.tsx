import React, { useState, useEffect, memo } from "react";
import {
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
                          value={pb.mars.ElementType.ElementTypeUnknown}
                        >
                          未设置
                        </Select.Option>
                        <Select.Option value={pb.mars.ElementType.ElementTypeInput}>
                          Input
                        </Select.Option>
                        <Select.Option
                          value={pb.mars.ElementType.ElementTypeInputNumber}
                        >
                          InputNumber
                        </Select.Option>
                        <Select.Option value={pb.mars.ElementType.ElementTypeRadio}>
                          Radio
                        </Select.Option>
                        <Select.Option value={pb.mars.ElementType.ElementTypeSelect}>
                          Select
                        </Select.Option>
                        <Select.Option value={pb.mars.ElementType.ElementTypeSwitch}>
                          Switch
                        </Select.Option>
                      </Select>
                    </Form.Item>
                    <Form.Item
                      style={{ width: "100%" }}
                      label="默认值"
                      dependencies={[
                        ["elements", field.name, "type"],
                        ["elements", field.name, "select_values"],
                      ]}
                      name={[field.name, "default"]}
                      rules={[
                        { required: true, message: "默认值必填" },
                        ({ getFieldValue }) => ({
                          validator(_, value) {
                            const fieldType = getFieldValue([
                              "elements",
                              field.name,
                              "type",
                            ]);
                            const selectValues = getFieldValue([
                              "elements",
                              field.name,
                              "select_values",
                            ]);
                            let flag = false;

                            switch (fieldType) {
                              case pb.mars.ElementType.ElementTypeSelect:
                              case pb.mars.ElementType.ElementTypeRadio:
                                if (Array.isArray(selectValues)) {
                                  for (const key in selectValues) {
                                    if (selectValues[key] === value) {
                                      flag = true;
                                      break;
                                    }
                                  }
                                }
                                break;
                              default:
                                flag = true;
                                break;
                            }
                            if (flag) {
                              return Promise.resolve();
                            }
                            return Promise.reject(
                              new Error("default 默认值必须在选择器中")
                            );
                          },
                        }),
                      ]}
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
                            pb.mars.ElementType.ElementTypeSelect ||
                            type[field.key] === pb.mars.ElementType.ElementTypeRadio)
                        )
                      }
                      style={{ width: "100%" }}
                      label="选择器"
                      name={[field.name, "select_values"]}
                    >
                      <MySelect disabled={disabled} />
                    </Form.Item>
                  </div>
                  {!disabled && (
                    <MinusCircleOutlined onClick={() => remove(field.name)} />
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
  type: pb.mars.ElementType;
  disabled: boolean;
}> = ({ type, disabled, value, onChange }) => {
  const [t, setT] = useState(type);
  useEffect(() => {
    setT(type);
    if (t !== type) {
      switch (type) {
        case pb.mars.ElementType.ElementTypeSwitch:
          if (value !== "false" || value !== "true") {
            onChange?.("false");
          }
          break;
        case pb.mars.ElementType.ElementTypeInputNumber:
          if (isNaN(Number(value))) {
            onChange?.("0");
          }
          break;
        default:
          break;
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [type, t, value]);

  switch (t) {
    case pb.mars.ElementType.ElementTypeInputNumber:
      return (
        <InputNumber
          disabled={disabled}
          value={value}
          onChange={(v) => onChange?.(String(v))}
          placeholder="默认值"
        />
      );
    case pb.mars.ElementType.ElementTypeSwitch:
      return (
        <Select
          disabled={disabled}
          value={value}
          onChange={(v) => onChange?.(String(v))}
        >
          <Select.Option value={"false"}>false</Select.Option>
          <Select.Option value={"true"}>true</Select.Option>
        </Select>
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
