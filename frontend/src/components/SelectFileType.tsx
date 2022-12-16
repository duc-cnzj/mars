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
      <Select.Option value="yaml">yaml</Select.Option>
      <Select.Option value="javascript">javascript</Select.Option>
      <Select.Option value="ini">ini</Select.Option>
      <Select.Option value="php">php</Select.Option>
      <Select.Option value="sql">sql</Select.Option>
      <Select.Option value="go">go</Select.Option>
      <Select.Option value="python">python</Select.Option>
      <Select.Option value="json">json</Select.Option>
      <Select.Option value="apl">apl</Select.Option>
      <Select.Option value="asciiArmor">asciiArmor</Select.Option>
      <Select.Option value="asterisk">asterisk</Select.Option>
      <Select.Option value="c">c</Select.Option>
      <Select.Option value="csharp">csharp</Select.Option>
      <Select.Option value="scala">scala</Select.Option>
      <Select.Option value="kotlin">kotlin</Select.Option>
      <Select.Option value="shader">shader</Select.Option>
      <Select.Option value="nesC">nesC</Select.Option>
      <Select.Option value="objectiveC">objectiveC</Select.Option>
      <Select.Option value="objectiveCpp">objectiveCpp</Select.Option>
      <Select.Option value="squirrel">squirrel</Select.Option>
      <Select.Option value="ceylon">ceylon</Select.Option>
      <Select.Option value="dart">dart</Select.Option>
      <Select.Option value="cmake">cmake</Select.Option>
      <Select.Option value="cobol">cobol</Select.Option>
      <Select.Option value="commonLisp">commonLisp</Select.Option>
      <Select.Option value="crystal">crystal</Select.Option>
      <Select.Option value="cypher">cypher</Select.Option>
      <Select.Option value="d">d</Select.Option>
      <Select.Option value="diff">diff</Select.Option>
      <Select.Option value="dtd">dtd</Select.Option>
      <Select.Option value="dylan">dylan</Select.Option>
      <Select.Option value="ebnf">ebnf</Select.Option>
      <Select.Option value="ecl">ecl</Select.Option>
      <Select.Option value="eiffel">eiffel</Select.Option>
      <Select.Option value="elm">elm</Select.Option>
      <Select.Option value="factor">factor</Select.Option>
      <Select.Option value="fcl">fcl</Select.Option>
      <Select.Option value="forth">forth</Select.Option>
      <Select.Option value="fortran">fortran</Select.Option>
      <Select.Option value="gas">gas</Select.Option>
      <Select.Option value="gherkin">gherkin</Select.Option>
      <Select.Option value="groovy">groovy</Select.Option>
      <Select.Option value="haskell">haskell</Select.Option>
      <Select.Option value="haxe">haxe</Select.Option>
      <Select.Option value="http">http</Select.Option>
      <Select.Option value="idl">idl</Select.Option>
      <Select.Option value="jinja2">jinja2</Select.Option>
      <Select.Option value="mathematica">mathematica</Select.Option>
      <Select.Option value="mbox">mbox</Select.Option>
      <Select.Option value="mirc">mirc</Select.Option>
      <Select.Option value="modelica">modelica</Select.Option>
      <Select.Option value="mscgen">mscgen</Select.Option>
      <Select.Option value="mumps">mumps</Select.Option>
      <Select.Option value="nsis">nsis</Select.Option>
      <Select.Option value="ntriples">ntriples</Select.Option>
      <Select.Option value="octave">octave</Select.Option>
      <Select.Option value="oz">oz</Select.Option>
      <Select.Option value="pig">pig</Select.Option>
      <Select.Option value="properties">properties</Select.Option>
      <Select.Option value="protobuf">protobuf</Select.Option>
      <Select.Option value="puppet">puppet</Select.Option>
      <Select.Option value="q">q</Select.Option>
      <Select.Option value="sas">sas</Select.Option>
      <Select.Option value="sass">sass</Select.Option>
      <Select.Option value="sieve">sieve</Select.Option>
      <Select.Option value="smalltalk">smalltalk</Select.Option>
      <Select.Option value="solr">solr</Select.Option>
      <Select.Option value="sparql">sparql</Select.Option>
      <Select.Option value="spreadsheet">spreadsheet</Select.Option>
      <Select.Option value="stex">stex</Select.Option>
      <Select.Option value="textile">textile</Select.Option>
      <Select.Option value="tiddlyWiki">tiddlyWiki</Select.Option>
      <Select.Option value="tiki">tiki</Select.Option>
      <Select.Option value="troff">troff</Select.Option>
      <Select.Option value="ttcn">ttcn</Select.Option>
      <Select.Option value="turtle">turtle</Select.Option>
      <Select.Option value="velocity">velocity</Select.Option>
      <Select.Option value="verilog">verilog</Select.Option>
      <Select.Option value="vhdl">vhdl</Select.Option>
      <Select.Option value="wast">wast</Select.Option>
      <Select.Option value="webIDL">webIDL</Select.Option>
      <Select.Option value="xQuery">xQuery</Select.Option>
      <Select.Option value="yacas">yacas</Select.Option>
      <Select.Option value="z80">z80</Select.Option>
      <Select.Option value="jsx">jsx</Select.Option>
      <Select.Option value="typescript">typescript</Select.Option>
      <Select.Option value="tsx">tsx</Select.Option>
      <Select.Option value="html">html</Select.Option>
      <Select.Option value="css">css</Select.Option>
      <Select.Option value="markdown">markdown</Select.Option>
      <Select.Option value="xml">xml</Select.Option>
      <Select.Option value="mysql">mysql</Select.Option>
      <Select.Option value="pgsql">pgsql</Select.Option>
      <Select.Option value="java">java</Select.Option>
      <Select.Option value="rust">rust</Select.Option>
      <Select.Option value="cpp">cpp</Select.Option>
      <Select.Option value="lezer">lezer</Select.Option>
      <Select.Option value="shell">shell</Select.Option>
      <Select.Option value="lua">lua</Select.Option>
      <Select.Option value="swift">swift</Select.Option>
      <Select.Option value="tcl">tcl</Select.Option>
      <Select.Option value="vb">vb</Select.Option>
      <Select.Option value="powershell">powershell</Select.Option>
      <Select.Option value="brainfuck">brainfuck</Select.Option>
      <Select.Option value="stylus">stylus</Select.Option>
      <Select.Option value="erlang">erlang</Select.Option>
      <Select.Option value="nginx">nginx</Select.Option>
      <Select.Option value="perl">perl</Select.Option>
      <Select.Option value="ruby">ruby</Select.Option>
      <Select.Option value="pascal">pascal</Select.Option>
      <Select.Option value="livescript">livescript</Select.Option>
      <Select.Option value="scheme">scheme</Select.Option>
      <Select.Option value="toml">toml</Select.Option>
      <Select.Option value="vbscript">vbscript</Select.Option>
      <Select.Option value="clojure">clojure</Select.Option>
      <Select.Option value="coffeescript">coffeescript</Select.Option>
      <Select.Option value="julia">julia</Select.Option>
      <Select.Option value="dockerfile">dockerfile</Select.Option>
      <Select.Option value="r">r</Select.Option>
      <Select.Option value="其他">其他</Select.Option>
    </Select>
  );
};

export default memo(SelectFileType);
