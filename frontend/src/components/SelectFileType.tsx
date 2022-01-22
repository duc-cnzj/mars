import React from "react";
import { Select, SelectProps } from "antd";

const SelectFileType: React.FC<
  {
    value?: string;
    onChange?: (value: string) => void;
  } & SelectProps
> = ({ value, onChange, ...rest }) => {
  return (
    <Select {...rest} value={value} onChange={onChange}>
      <Select.Option value="env">.env</Select.Option>
      <Select.Option value="yaml">yaml</Select.Option>
      <Select.Option value="js">js</Select.Option>
      <Select.Option value="ini">ini</Select.Option>
      <Select.Option value="php">php</Select.Option>
      <Select.Option value="sql">sql</Select.Option>
      <Select.Option value="go">go</Select.Option>
      <Select.Option value="python">python</Select.Option>
      <Select.Option value="json">json</Select.Option>
      <Select.Option value="others">其他</Select.Option>
    </Select>
  );
};

export default SelectFileType;
