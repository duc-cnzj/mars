import React, { useEffect, useState } from "react";
import { v4 as uuidv4 } from "uuid";
import {
  Switch,
  Form,
  Input,
  Button,
  Select,
  SelectProps,
  InputNumber,
} from "antd";
import { MinusCircleOutlined, PlusOutlined } from "@ant-design/icons";
import pb from "../../api/compiled";

interface Ele extends pb.Element {
  id?: string;
}
const initValue: Ele = {
  type: pb.ElementType.ElementTypeUnknown,
  path: "",
  default: "",
  select_values: [],
  description: "",
  id: "",
};

const DynamicElement: React.FC<{
  value: Ele[];
  disabled: boolean;
  onChange?: (eles: Ele[]) => void;
}> = ({ value: iValue, onChange, disabled }) => {
  const [value, setValue] = useState<Ele[]>(iValue);
  useEffect(() => {
    setValue(
      iValue.map((i) => {
        if (!i.id) {
          return { ...i, id: uuidv4() };
        }
        return i;
      })
    );
  }, [iValue]);

  return (
    <div>
      {value &&
        value.map((item, key) => {
          console.log(item, "key", key);
          return (
            <div key={item.id}>
              <div style={{ position: "relative" }}>
                <Element
                  disabled={disabled}
                  value={item}
                  onChange={(v: Ele) => {
                    let vv = value ? value : [];
                    vv = vv.map((a) => {
                      if (a.id === v.id) {
                        return v;
                      }
                      return a;
                    });
                    setValue([...vv]);
                    onChange?.([...vv]);
                  }}
                />{" "}
                {disabled ? null : (
                  <MinusCircleOutlined
                    style={{ position: "absolute", bottom: 10, left: 10 }}
                    onClick={() => {
                      let vv = value ? value : [];
                      vv = vv.filter((v) => v.id !== item.id);
                      setValue([...vv]);
                      onChange?.([...vv]);
                    }}
                  />
                )}
              </div>
            </div>
          );
        })}
      <Form.Item>
        <Button
          disabled={disabled}
          type="dashed"
          onClick={() => {
            setValue((v) => [...v, { ...initValue, id: uuidv4() }]);
          }}
          block
          icon={<PlusOutlined />}
        >
          Add field
        </Button>
      </Form.Item>
    </div>
  );
};

const Element: React.FC<{
  value: Ele;
  disabled: boolean;
  onChange: (v: Ele) => void;
}> = ({ value, onChange, disabled }) => {
  const [config, setConfig] = useState(value);
  return (
    <div className="dynamic-element">
      <div className="dynamic-element__wrapper">
        <div style={{ display: "flex", width: "100%" }}>
          <Form.Item style={{ width: "100%" }} label={"字段路径"}>
            <Input
              disabled={disabled}
              placeholder="字段路径"
              value={config.path}
              onChange={(v) => {
                setConfig((config) => {
                  let data = {
                    ...config,
                    path: String(v.target.value),
                  };
                  onChange({
                    ...config,
                    path: String(v.target.value),
                  });
                  return data;
                });
              }}
            />
          </Form.Item>
          <Form.Item style={{ width: "100%" }} label="表单类型">
            <Select
              disabled={disabled}
              value={config.type}
              onChange={(v) => {
                setConfig((config) => {
                  let data = {
                    ...config,
                    type: Number(v),
                  };
                  onChange({
                    ...config,
                    type: Number(v),
                  });
                  return data;
                });
              }}
            >
              <Select.Option value={pb.ElementType.ElementTypeUnknown}>
                未设置
              </Select.Option>
              <Select.Option value={pb.ElementType.ElementTypeInput}>
                Input
              </Select.Option>
              <Select.Option value={pb.ElementType.ElementTypeInputNumber}>
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
            hidden={config.type === pb.ElementType.ElementTypeUnknown}
          >
            <DefaultValueElement
              disabled={disabled}
              type={config.type}
              value={config.default}
              onChange={(v: any) => {
                setConfig((config) => {
                  let data = {
                    ...config,
                    default: String(v),
                  };
                  onChange({
                    ...config,
                    default: String(v),
                  });
                  return data;
                });
              }}
            />
          </Form.Item>
        </div>
        <div style={{ display: "flex" }}>
          <Form.Item style={{ width: "100%" }} label="字段描述">
            <Input
              disabled={disabled}
              placeholder="字段描述"
              value={config.description}
              onChange={(v) => {
                setConfig((config) => {
                  let data = {
                    ...config,
                    description: String(v.target.value),
                  };
                  onChange({
                    ...config,
                    description: String(v.target.value),
                  });
                  return data;
                });
              }}
            />
          </Form.Item>

          {!(
            config.type === pb.ElementType.ElementTypeRadio ||
            config.type === pb.ElementType.ElementTypeSelect
          ) ? (
            ""
          ) : (
            <Form.Item style={{ width: "100%" }} label="选择器">
              <MySelect
                disabled={disabled}
                value={config.select_values}
                onChange={(v) => {
                  setConfig((config) => {
                    let data = {
                      ...config,
                      select_values: v,
                    };
                    onChange({
                      ...config,
                      select_values: v,
                    });
                    return data;
                  });
                }}
              />
            </Form.Item>
          )}
        </div>
      </div>
    </div>
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
    ></Select>
  );
};

const DefaultValueElement: React.FC<{
  type: pb.ElementType;
  value: any;
  disabled: boolean;
  onChange: (v: any) => void;
}> = ({ type, value, onChange, disabled }) => {
  console.log(type);
  switch (type) {
    case pb.ElementType.ElementTypeInputNumber:
      return (
        <InputNumber
          disabled={disabled}
          placeholder="默认值"
          value={Number(value)}
          onChange={(e) => {
            onChange(e);
          }}
        />
      );
    case pb.ElementType.ElementTypeSwitch:
      return (
        <Switch
          disabled={disabled}
          defaultChecked={value}
          onChange={(e) => onChange(e)}
        />
      );
    default:
      return (
        <Input
          disabled={disabled}
          placeholder="默认值"
          value={value}
          onChange={(e) => {
            onChange(e.target.value);
          }}
        />
      );
  }
};

export default DynamicElement;
