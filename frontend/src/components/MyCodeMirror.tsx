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
require("./autorefresh.ext");
(window as any).JSHINT = JSHINT;
(window as any).jsyaml = jsyaml;

var cm = require("codemirror");

const list = [
  { text: "<.ImagePullSecrets>", displayText: "imagePullSecrets" },
  { text: "<.Branch>", displayText: "branch" },
  { text: "<.Commit>", displayText: "commit" },
  { text: "<.Pipeline>", displayText: "pipeline" },
  { text: "<.ClusterIssuer>", displayText: "clusterIssuer" },
  { text: "<.Host1>", displayText: "host1" },
  { text: "<.Host2>", displayText: "host2" },
  { text: "<.Host3>", displayText: "host3" },
  { text: "<.Host4>", displayText: "host4" },
  { text: "<.Host5>", displayText: "host5" },
  { text: "<.Host6>", displayText: "host6" },
  { text: "<.Host7>", displayText: "host7" },
  { text: "<.Host8>", displayText: "host8" },
  { text: "<.Host9>", displayText: "host9" },
  { text: "<.Host10>", displayText: "host10" },
  { text: "<.TlsSecret1>", displayText: "tlsSecret1" },
  { text: "<.TlsSecret2>", displayText: "tlsSecret2" },
  { text: "<.TlsSecret3>", displayText: "tlsSecret3" },
  { text: "<.TlsSecret4>", displayText: "tlsSecret4" },
  { text: "<.TlsSecret5>", displayText: "tlsSecret5" },
  { text: "<.TlsSecret6>", displayText: "tlsSecret6" },
  { text: "<.TlsSecret7>", displayText: "tlsSecret7" },
  { text: "<.TlsSecret8>", displayText: "tlsSecret8" },
  { text: "<.TlsSecret9>", displayText: "tlsSecret9" },
  { text: "<.TlsSecret10>", displayText: "tlsSecret10" },
  {
    text: 'cert-manager.io/cluster-issuer: "<.ClusterIssuer>"',
    displayText: "certManager",
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
  autoRefresh: { force: true },
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
