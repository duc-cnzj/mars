import React, { useState, useEffect, memo, useCallback } from "react";
import { Form, Input, Button, Select, SelectProps, InputNumber } from "antd";
import { MinusCircleOutlined, PlusOutlined } from "@ant-design/icons";
import TextArea from "antd/lib/input/TextArea";
import { DragDropContext, Droppable, Draggable } from "react-beautiful-dnd";
import { slice } from "lodash";
import { css } from "@emotion/css";
import { FormInstance } from "antd/lib";
import { MarsElementType } from "../../api/schema.d";

function isDefaultRequired(t: MarsElementType): boolean {
  switch (t) {
    case MarsElementType.ElementTypeInputNumber:
    case MarsElementType.ElementTypeRadio:
    case MarsElementType.ElementTypeNumberRadio:
    case MarsElementType.ElementTypeSelect:
    case MarsElementType.ElementTypeNumberSelect:
    case MarsElementType.ElementTypeSwitch:
      return true;
  }
  return false;
}

const DynamicElement: React.FC<{ form: FormInstance }> = ({ form }) => {
  const [isDragging, setIsDragging] = useState(false);
  const onDragEnd = useCallback(
    (result: any) => {
      if (result.destination.index === result.source.index) {
        return;
      }

      let deleteIdx = result.source.index;
      let eles = form.getFieldValue(["marsConfig", "elements"]) as any[];
      let n = [
        ...slice(
          eles,
          0,
          result.source.index > result.destination.index
            ? result.destination.index
            : result.destination.index + 1,
        ),
        eles[result.source.index],
        ...slice(
          eles,
          result.source.index > result.destination.index
            ? result.destination.index
            : result.destination.index + 1,
        ),
      ];
      n.splice(
        result.source.index > result.destination.index
          ? deleteIdx + 1
          : deleteIdx,
        1,
      );

      form.setFieldValue(["marsConfig", "elements"], [...n]);
      setIsDragging(false);
    },
    [form],
  );

  return (
    <DragDropContext
      onDragEnd={onDragEnd}
      onDragStart={() => setIsDragging(true)}
    >
      <Droppable droppableId="dynamic-elements">
        {(provided) => (
          <div ref={provided.innerRef} {...provided.droppableProps}>
            <Form.List name={["marsConfig", "elements"]}>
              {(fields, { add, remove }) => (
                <>
                  {fields.map((field, index) => {
                    const type = form.getFieldValue([
                      "marsConfig",
                      "elements",
                      field.name,
                      "type",
                    ]);
                    return (
                      <Draggable
                        draggableId={String(field.name)}
                        index={index}
                        key={index}
                      >
                        {(provided) => (
                          <div
                            ref={provided.innerRef}
                            {...provided.draggableProps}
                            {...provided.dragHandleProps}
                            key={field.name}
                            className={css`
                              background-image: linear-gradient(
                                to right,
                                #a855f7,
                                #ec4899
                              );
                              padding: 2px;
                              margin-bottom: 5px;
                              border-radius: 7px;
                            `}
                          >
                            <div
                              className={css`
                                background-color: white;
                                overflow: hidden;
                                border-radius: 5px;
                                width: 100%;
                                height: 100%;
                                padding: 5px;
                              `}
                            >
                              <div style={{ display: "flex", width: "100%" }}>
                                <Form.Item
                                  hidden
                                  label={"字段顺序"}
                                  name={[field.name, "order"]}
                                >
                                  <InputNumber placeholder="字段顺序" />
                                </Form.Item>
                                <Form.Item
                                  style={{ width: "100%" }}
                                  label={"字段路径"}
                                  name={[field.name, "path"]}
                                  rules={[
                                    { required: true, message: "字段路径必填" },
                                  ]}
                                >
                                  <Input placeholder="字段路径" />
                                </Form.Item>
                                <Form.Item
                                  style={{ width: "100%" }}
                                  label="表单类型"
                                  name={[field.name, "type"]}
                                  rules={[
                                    { required: true, message: "表单类型必填" },
                                  ]}
                                >
                                  <Select
                                    placeholder="选择类型"
                                    onChange={(v) => {
                                      let eles = form.getFieldValue([
                                        "marsConfig",
                                        "elements",
                                      ]);
                                      Object.assign(eles[field.name], {
                                        type: v,
                                      });
                                      form.setFieldValue(
                                        ["marsConfig", "elements"],
                                        eles,
                                      );
                                    }}
                                  >
                                    <Select.Option
                                      value={MarsElementType.ElementTypeUnknown}
                                    >
                                      未设置
                                    </Select.Option>
                                    <Select.Option
                                      value={MarsElementType.ElementTypeInput}
                                    >
                                      Input
                                    </Select.Option>
                                    <Select.Option
                                      value={
                                        MarsElementType.ElementTypeInputNumber
                                      }
                                    >
                                      InputNumber
                                    </Select.Option>
                                    <Select.Option
                                      value={
                                        MarsElementType.ElementTypeTextArea
                                      }
                                    >
                                      TextArea
                                    </Select.Option>
                                    <Select.Option
                                      value={MarsElementType.ElementTypeRadio}
                                    >
                                      Radio
                                    </Select.Option>
                                    <Select.Option
                                      value={
                                        MarsElementType.ElementTypeNumberRadio
                                      }
                                    >
                                      Number Radio
                                    </Select.Option>
                                    <Select.Option
                                      value={MarsElementType.ElementTypeSelect}
                                    >
                                      Select
                                    </Select.Option>
                                    <Select.Option
                                      value={
                                        MarsElementType.ElementTypeNumberSelect
                                      }
                                    >
                                      Number Select
                                    </Select.Option>
                                    <Select.Option
                                      value={MarsElementType.ElementTypeSwitch}
                                    >
                                      Switch
                                    </Select.Option>
                                  </Select>
                                </Form.Item>

                                {type !==
                                  MarsElementType.ElementTypeTextArea && (
                                  <Form.Item
                                    style={{ width: "100%" }}
                                    label="默认值"
                                    dependencies={[
                                      [
                                        ["marsConfig", "elements"],
                                        field.name,
                                        "type",
                                      ],
                                      [
                                        ["marsConfig", "elements"],
                                        field.name,
                                        "selectValues",
                                      ],
                                    ]}
                                    name={[field.name, "default"]}
                                    rules={[
                                      {
                                        required: isDefaultRequired(type),
                                        message: "默认值必填",
                                      },
                                      ({ getFieldValue }) => ({
                                        validator(_, value) {
                                          const fieldType = getFieldValue([
                                            ["marsConfig", "elements"],
                                            field.name,
                                            "type",
                                          ]);
                                          const selectValues = getFieldValue([
                                            ["marsConfig", "elements"],
                                            field.name,
                                            "selectValues",
                                          ]);
                                          let flag = false;

                                          switch (fieldType) {
                                            case MarsElementType.ElementTypeSelect:
                                            case MarsElementType.ElementTypeNumberSelect:
                                            case MarsElementType.ElementTypeRadio:
                                            case MarsElementType.ElementTypeNumberRadio:
                                              if (Array.isArray(selectValues)) {
                                                for (const key in selectValues) {
                                                  if (
                                                    selectValues[key] === value
                                                  ) {
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
                                            new Error(
                                              "default 默认值必须在选择器中",
                                            ),
                                          );
                                        },
                                      }),
                                    ]}
                                  >
                                    <DefaultValueElement type={type} />
                                  </Form.Item>
                                )}
                              </div>
                              <div style={{ display: "flex" }}>
                                <Form.Item
                                  style={{ width: "100%" }}
                                  label="字段描述"
                                  name={[field.name, "description"]}
                                  rules={[
                                    { required: true, message: "字段描述必填" },
                                  ]}
                                >
                                  <Input placeholder="字段描述" />
                                </Form.Item>

                                <Form.Item
                                  hidden={
                                    !(
                                      type &&
                                      (type ===
                                        MarsElementType.ElementTypeNumberSelect ||
                                        type ===
                                          MarsElementType.ElementTypeSelect ||
                                        type ===
                                          MarsElementType.ElementTypeNumberRadio ||
                                        type ===
                                          MarsElementType.ElementTypeRadio)
                                    )
                                  }
                                  style={{ width: "100%" }}
                                  label="选择器"
                                  name={[field.name, "selectValues"]}
                                >
                                  <MySelect />
                                </Form.Item>
                              </div>
                              {type === MarsElementType.ElementTypeTextArea && (
                                <div style={{ display: "flex" }}>
                                  <Form.Item
                                    style={{ width: "100%" }}
                                    label="默认值"
                                    dependencies={[
                                      [
                                        ["marsConfig", "elements"],
                                        field.name,
                                        "type",
                                      ],
                                      [
                                        ["marsConfig", "elements"],
                                        field.name,
                                        "selectValues",
                                      ],
                                    ]}
                                    name={[field.name, "default"]}
                                  >
                                    <DefaultValueElement type={type} />
                                  </Form.Item>
                                </div>
                              )}
                              <MinusCircleOutlined
                                onClick={() => remove(field.name)}
                              />
                            </div>
                          </div>
                        )}
                      </Draggable>
                    );
                  })}
                  <Form.Item>
                    <Button
                      hidden={isDragging}
                      disabled={isDragging}
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
            {provided.placeholder}
          </div>
        )}
      </Droppable>
    </DragDropContext>
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
  type: MarsElementType;
}> = ({ type, value, onChange }) => {
  const [t, setT] = useState(type);
  useEffect(() => {
    setT(type);
    if (t !== type) {
      switch (type) {
        case MarsElementType.ElementTypeSwitch:
          if (value !== "false" || value !== "true") {
            onChange?.("false");
          }
          break;
        case MarsElementType.ElementTypeInputNumber:
          if (isNaN(Number(value))) {
            onChange?.("0");
          }
          break;
        default:
          break;
      }
    }
  }, [type, t, value, onChange]);

  switch (t) {
    case MarsElementType.ElementTypeInputNumber:
      return (
        <InputNumber
          value={value}
          onChange={(v) => onChange?.(String(v))}
          placeholder="默认值"
        />
      );
    case MarsElementType.ElementTypeSwitch:
      return (
        <Select value={value} onChange={(v) => onChange?.(String(v))}>
          <Select.Option value={"false"}>false</Select.Option>
          <Select.Option value={"true"}>true</Select.Option>
        </Select>
      );
    case MarsElementType.ElementTypeTextArea:
      return (
        <TextArea value={value} onChange={onChange} placeholder="默认值" />
      );
    default:
      return <Input value={value} onChange={onChange} placeholder="默认值" />;
  }
};

export default memo(DynamicElement);
