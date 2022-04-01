import React from "react";
import {
  Controlled as CodeMirror,
  ICodeMirror,
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

import "codemirror/addon/comment/comment";

import "codemirror/addon/hint/show-hint";
import "codemirror/addon/hint/anyword-hint";
import "codemirror/addon/hint/javascript-hint";
import "codemirror/addon/hint/sql-hint";

import "codemirror/addon/lint/javascript-lint";
import "codemirror/addon/lint/yaml-lint";
import "codemirror/addon/lint/lint.js";

import { JSHINT } from "jshint";
import jsyaml from "js-yaml";
import { lowerCase } from "lodash";

(window as any).JSHINT = JSHINT;
(window as any).jsyaml = jsyaml;

var cm = require("codemirror");

const list = [
  { text: "<.ImagePullSecrets>", displayText: "<.ImagePullSecrets>" },
  { text: "<.Branch>", displayText: "<.Branch>" },
  { text: "<.Commit>", displayText: "<.Commit>" },
  { text: "<.Pipeline>", displayText: "<.Pipeline>" },
  { text: "<.ClusterIssuer>", displayText: "<.ClusterIssuer>" },
  { text: "<.Host1>", displayText: "<.Host1>" },
  { text: "<.Host2>", displayText: "<.Host2>" },
  { text: "<.Host3>", displayText: "<.Host3>" },
  { text: "<.Host4>", displayText: "<.Host4>" },
  { text: "<.Host5>", displayText: "<.Host5>" },
  { text: "<.Host6>", displayText: "<.Host6>" },
  { text: "<.Host7>", displayText: "<.Host7>" },
  { text: "<.Host8>", displayText: "<.Host8>" },
  { text: "<.Host9>", displayText: "<.Host9>" },
  { text: "<.Host10>", displayText: "<.Host10>" },
  { text: "<.TlsSecret1>", displayText: "<.TlsSecret1>" },
  { text: "<.TlsSecret2>", displayText: "<.TlsSecret2>" },
  { text: "<.TlsSecret3>", displayText: "<.TlsSecret3>" },
  { text: "<.TlsSecret4>", displayText: "<.TlsSecret4>" },
  { text: "<.TlsSecret5>", displayText: "<.TlsSecret5>" },
  { text: "<.TlsSecret6>", displayText: "<.TlsSecret6>" },
  { text: "<.TlsSecret7>", displayText: "<.TlsSecret7>" },
  { text: "<.TlsSecret8>", displayText: "<.TlsSecret8>" },
  { text: "<.TlsSecret9>", displayText: "<.TlsSecret9>" },
  { text: "<.TlsSecret10>", displayText: "<.TlsSecret10>" },
  {
    text: 'cert-manager.io/cluster-issuer: "<.ClusterIssuer>"',
    displayText: "certManager",
  },
  {
    text: "<.Branch>-<.Pipeline>",
    displayText: "imageTag",
  },
];

var wordRegexp = /[^"\s>\-_]+/;

let orig = cm.hint.anyword;
cm.hint.yaml = function (e: any) {
  let cur = e.getCursor();
  let curLine = e.getLine(cur.line);
  var end = cur.ch,
    start = end;
  while (start && wordRegexp.test(curLine.charAt(start - 1))) --start;
  var curWord = start !== end && curLine.slice(start, end);

  let filteredList =
    curWord.length > 0
      ? list.filter((item) => {
          return lowerCase(item.text).includes(lowerCase(curWord));
        })
      : list;
  let innter = orig(e) || {
    from: cm.Pos(cur.line, start),
    to: cm.Pos(cur.line, cur.ch),
    list: [],
  };

  return {
    from: cm.Pos(cur.line, start),
    to: cm.Pos(cur.line, cur.ch),
    list: [...innter.list, ...filteredList],
  };
};

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
    case "sql":
      return "text/x-sql";
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
  lineNumbers: true,
  lint: true,
  gutters: ["CodeMirror-lint-markers"],
  extraKeys: {
    "Alt-Enter": "autocomplete",
    "Ctrl-/": (editor: any) => editor.execCommand("toggleComment"),
    "Cmd-/": (editor: any) => editor.execCommand("toggleComment"),
  },
  hintOptions: {
    completeSingle: false,
  },
};

interface InputMirror extends ICodeMirror {
  onChange?: (v: string) => void;
  value?: string;
}

const myCodeMirror: React.ForwardRefRenderFunction<
  Controlled,
  React.RefAttributes<Controlled> & InputMirror
> = (props, ref) => {
  return (
    <CodeMirror
      ref={ref}
      editorDidMount={(editor) => {
        setTimeout(() => {
          editor.refresh();
        }, 200);
      }}
      value={props.value ? props.value : ""}
      onBeforeChange={(e: any, d: any, v: string) => {
        props.onChange?.(v);
      }}
      options={{ ...props.options, ...defaultOpt }}
    />
  );
};
export const MyCodeMirror = React.forwardRef(myCodeMirror);

export default MyCodeMirror;
