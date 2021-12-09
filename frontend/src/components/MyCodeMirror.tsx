import React from "react";
import {
  Controlled as CodeMirror,
  IControlledCodeMirror,
  Controlled,
} from "react-codemirror2";

import "codemirror/mode/go/go";
import "codemirror/mode/css/css";
import "codemirror/mode/javascript/javascript";
import "codemirror/mode/yaml/yaml";
import "codemirror/mode/php/php";
import "codemirror/mode/python/python";
import "codemirror/mode/properties/properties";
import "codemirror/mode/textile/textile";

import "codemirror/addon/lint/javascript-lint";
import "codemirror/addon/lint/yaml-lint";
import "codemirror/addon/lint/lint.js";
import "codemirror/addon/hint/javascript-hint";

import { JSHINT } from "jshint";
import jsyaml from "js-yaml";
require("./autorefresh.ext");
(window as any).JSHINT = JSHINT;
(window as any).jsyaml = jsyaml;

export const getMode = (mode: string): string => {
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
    case "json":
      return "application/json";
    case "go":
      return "text/x-go";
    case "py":
    case "python":
      return "text/x-python";
    default:
      return mode;
  }
};

const defaultOpt = {
  autoRefresh: { force: true },
  lineNumbers: true,
  lint: true,
  gutters: ["CodeMirror-lint-markers"],
};

const myCodeMirror: React.ForwardRefRenderFunction<
  Controlled,
  IControlledCodeMirror & React.RefAttributes<Controlled>
> = (props, ref) => {
  return (
    <CodeMirror
      ref={ref}
      {...props}
      options={{ ...props.options, ...defaultOpt }}
    />
  );
};
export const MyCodeMirror = React.forwardRef(myCodeMirror);

export default MyCodeMirror;
