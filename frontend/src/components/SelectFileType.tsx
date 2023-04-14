import React, { memo } from "react";
import { Select, SelectProps } from "antd";

const SelectFileType: React.FC<
  {
    value?: string;
    onChange?: (value: string) => void;
  } & SelectProps
> = ({ value, onChange, ...rest }) => {
  return (
    <Select showSearch {...rest} value={value} onChange={onChange}>
      <Select.Option value="env">.env</Select.Option>
      <Select.Option value="php">php</Select.Option>
      <Select.Option value="json">json</Select.Option>
      <Select.Option value="yaml">yaml</Select.Option>
      <Select.Option value="go">go</Select.Option>
      <Select.Option value="c">c</Select.Option>
      <Select.Option value="csharp">csharp</Select.Option>
      <Select.Option value="scala">scala</Select.Option>
      <Select.Option value="kotlin">kotlin</Select.Option>
      <Select.Option value="objectiveC">objectiveC</Select.Option>
      <Select.Option value="objectiveCpp">objectiveCpp</Select.Option>
      <Select.Option value="dart">dart</Select.Option>
      <Select.Option value="cmake">cmake</Select.Option>
      <Select.Option value="groovy">groovy</Select.Option>
      <Select.Option value="haskell">haskell</Select.Option>
      <Select.Option value="dockerfile">dockerfile</Select.Option>
      <Select.Option value="http">http</Select.Option>
      <Select.Option value="jinja2">jinja2</Select.Option>
      <Select.Option value="properties">properties</Select.Option>
      <Select.Option value="protobuf">protobuf</Select.Option>
      <Select.Option value="puppet">puppet</Select.Option>
      <Select.Option value="sass">sass</Select.Option>
      <Select.Option value="textile">textile</Select.Option>
      <Select.Option value="javascript">javascript</Select.Option>
      <Select.Option value="jsx">jsx</Select.Option>
      <Select.Option value="typescript">typescript</Select.Option>
      <Select.Option value="tsx">tsx</Select.Option>
      <Select.Option value="html">html</Select.Option>
      <Select.Option value="css">css</Select.Option>
      <Select.Option value="python">python</Select.Option>
      <Select.Option value="markdown">markdown</Select.Option>
      <Select.Option value="xml">xml</Select.Option>
      <Select.Option value="sql">sql</Select.Option>
      <Select.Option value="mysql">mysql</Select.Option>
      <Select.Option value="pgsql">pgsql</Select.Option>
      <Select.Option value="java">java</Select.Option>
      <Select.Option value="rust">rust</Select.Option>
      <Select.Option value="cpp">cpp</Select.Option>
      <Select.Option value="shell">shell</Select.Option>
      <Select.Option value="lua">lua</Select.Option>
      <Select.Option value="swift">swift</Select.Option>
      <Select.Option value="vb">vb</Select.Option>
      <Select.Option value="powershell">powershell</Select.Option>
      <Select.Option value="stylus">stylus</Select.Option>
      <Select.Option value="ruby">ruby</Select.Option>
      <Select.Option value="erlang">erlang</Select.Option>
      <Select.Option value="nginx">nginx</Select.Option>
      <Select.Option value="perl">perl</Select.Option>
      <Select.Option value="less">less</Select.Option>
      <Select.Option value="toml">toml</Select.Option>
      <Select.Option value="vbscript">vbscript</Select.Option>
      <Select.Option value="coffeescript">coffeescript</Select.Option>
      <Select.Option value="julia">julia</Select.Option>
    </Select>
  );
};

export default memo(SelectFileType);
