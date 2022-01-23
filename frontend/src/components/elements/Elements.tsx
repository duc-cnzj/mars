import React, { useState, Fragment, useMemo } from "react";
import pb from "../../api/compiled";
import { Form, Input, InputNumber, Radio, Select, Switch } from "antd";
const Option = Select.Option;

interface st {
  input?: React.CSSProperties;
  inputNumber?: React.CSSProperties;
  label?: React.CSSProperties;
  formItem?: React.CSSProperties;
  radioGroup?: React.CSSProperties;
  radio?: React.CSSProperties;
  select?: React.CSSProperties;
  selectOption?: React.CSSProperties;
  switch?: React.CSSProperties;
}
const isTrue = (v: any): boolean => {
  switch (v) {
    case 1:
    case true:
    case "1":
    case "True":
    case "true":
      return true;
    default:
      return false;
  }
};
const initStyle = {
  input: {},
  label: {},
  formItem: {},
  radioGroup: {},
  radio: {},
  select: {},
  selectOption: {},
  switch: {},
};

const Elements: React.FC<{
  value?: pb.ProjectExtraItem[];
  onChange?: (value: pb.ProjectExtraItem[]) => void;
  elements: pb.Element[];
  style?: st;
}> = ({ elements, style, value, onChange }) => {
  let vvv = useMemo(() => {
    return elements
      ? elements.map((item): pb.ProjectExtraItem => {
          let de: any = item.default;
          if (!!value) {
            for (let i = 0; i < value.length; i++) {
              if (value[i].path === item.path) {
                de = value[i].value;
                if (item.type === pb.ElementType.ElementTypeSwitch) {
                  de = isTrue(de);
                }
                if (item.type === pb.ElementType.ElementTypeInputNumber) {
                  de = Number(de);
                }
                break;
              }
            }
          }
          return { path: item.path, value: de };
        })
      : [];
  }, [elements, value]);

  const getElement = (
    v: pb.ProjectExtraItem,
    e: pb.Element[],
    index: number
  ): React.ReactNode => {
    for (let i = 0; i < e.length; i++) {
      let ev = e[i];
      if (ev.path === v.path) {
        return (
          <Element
            value={v.value}
            onChange={(vv) => {
              let va: any = vvv;
              va[index].value = String(vv)
              onChange?.(va);
              return va;
            }}
            element={ev}
            style={style ? style : initStyle}
          />
        );
      }
    }
    return <></>;
  };
  return (
    <div style={{ width: "100%" }}>
      {vvv.map((item, index) => (
        <Fragment key={item.path}>{getElement(item, elements, index)}</Fragment>
      ))}
    </div>
  );
};

const Element: React.FC<{
  value: any;
  onChange: (v: any) => void;
  element: pb.Element;
  style: st;
}> = ({ element, style, value: v, onChange }) => {
  const [value, setValue] = useState(v);
  switch (element.type) {
    case pb.ElementType.ElementTypeInput:
      return (
        <Form.Item
          label={<div style={style.label}>{element.description}</div>}
          style={style.formItem}
        >
          <Input
            defaultValue={element.default}
            value={value}
            onChange={(e) => {
              setValue(e.target.value);
              onChange(e.target.value);
            }}
            style={style.input}
          />
        </Form.Item>
      );
    case pb.ElementType.ElementTypeInputNumber:
      return (
        <Form.Item
          label={<div style={style.label}>{element.description}</div>}
          style={style.formItem}
        >
          <InputNumber
            defaultValue={Number(element.default)}
            style={style.inputNumber}
            value={value}
            onChange={(e) => {
              setValue(e);
              onChange(e);
            }}
          />
        </Form.Item>
      );
    case pb.ElementType.ElementTypeRadio:
      return (
        <Form.Item
          label={<div style={style.label}>{element.description}</div>}
          style={style.formItem}
        >
          <Radio.Group
            defaultValue={element.default}
            style={style.radioGroup}
            value={value}
            onChange={(e) => {
              setValue(e.target.value);
              onChange(e.target.value);
            }}
          >
            {element.select_values.map((i, k) => (
              <Radio key={k} value={i} style={style.radio}>
                {i}
              </Radio>
            ))}
          </Radio.Group>
        </Form.Item>
      );
    case pb.ElementType.ElementTypeSelect:
      return (
        <Form.Item
          label={<div style={style.label}>{element.description}</div>}
          style={style.formItem}
        >
          <Select
            defaultValue={element.default}
            style={style.select}
            value={value}
            onChange={(e) => {
              setValue(e);
              onChange(e);
            }}
          >
            {element.select_values.map((i, k) => (
              <Option value={i} key={k} style={style}>
                {i}
              </Option>
            ))}
          </Select>
        </Form.Item>
      );
    case pb.ElementType.ElementTypeSwitch:
      return (
        <Form.Item
          label={<div style={style.label}>{element.description}</div>}
          style={style.formItem}
        >
          <Switch
            defaultChecked={isTrue(element.default)}
            style={style.switch}
            checked={value}
            onChange={(e) => {
              setValue(e);
              onChange(e);
            }}
          />
        </Form.Item>
      );
    default:
      return <></>;
  }
};

export default Elements;
