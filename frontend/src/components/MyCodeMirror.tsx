import React from "react";
import { Controlled as CodeMirror, IControlledCodeMirror, Controlled } from "react-codemirror2";

require("codemirror/mode/go/go");
require("codemirror/mode/css/css");
require("codemirror/mode/javascript/javascript");
require("codemirror/mode/yaml/yaml");
require("codemirror/mode/php/php");
require("codemirror/mode/python/python");
require("codemirror/mode/properties/properties");
require("codemirror/mode/textile/textile");

export const getMode = (mode: string):string => {
  switch (mode) {
    case "dotenv":
    case "env":
    case ".env":
      return "text/x-textile";
    case "yaml":
      return "text/x-yaml";
    case "js":
    case "javascript":
      return "text/javascript";
    case "ini":
      return "text/x-properties";
    case "php":
      return "php";
    case "go":
      return "text/x-go";
    case "py":
    case "python":
      return "text/x-python";
    default:
      return mode;
  }
}

const myCodeMirror: React.ForwardRefRenderFunction<Controlled, IControlledCodeMirror & React.RefAttributes<Controlled>> = (props, ref) => {
  return (
    <CodeMirror
      ref={ref}
      {...props}
    />
  );
};
export const MyCodeMirror = React.forwardRef(myCodeMirror);

export default MyCodeMirror;
