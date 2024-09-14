import React, { memo } from "react";
import CodeMirror from "@uiw/react-codemirror";
import { dracula } from "@uiw/codemirror-theme-dracula";
import { langs } from "../utils/langs";
import {
  Completion,
  CompletionContext,
  autocompletion,
  startCompletion,
} from "@codemirror/autocomplete";
import { color } from "@uiw/codemirror-extensions-color";
import { EditorView, keymap } from "@codemirror/view";
import { jsonParseLinter } from "@codemirror/lang-json";
import { linter } from "@codemirror/lint";
import { omitEqual } from "../utils/obj";
import parser from "js-yaml";
import { Diagnostic } from "@codemirror/lint";

const yamlLinter = linter((view: EditorView): Diagnostic[] => {
  const diagnostics: Diagnostic[] = [];

  try {
    parser.load(String(view.state.doc));
  } catch (e: any) {
    const loc = e.mark;
    let from = loc ? loc.position : 0;
    let to = from;
    const severity = "error";
    if (to > view.state.doc.length) {
      from = to = view.state.doc.length;
    }

    diagnostics.push({
      from,
      to,
      message: e.message,
      severity,
    });
  }

  return diagnostics;
});

// https://codesandbox.io/s/codemirror-6-demo-forked-mce50r?file=/src/index.js:626-692
export const MyCodeMirror: React.FC<{
  mode: string;
  value?: string;
  disabled?: boolean;
  completionValues?: boolean;
  onChange?: (v: string) => void;
}> = memo(
  ({ mode, value, onChange, disabled, completionValues }) => {
    console.log("MyCodeMirror render");
    const langeExt = getLangs(mode);
    const extensions = [
      color,
      langeExt,
      theme,
      keymap.of([{ key: "Alt-Enter", run: startCompletion }]),
    ];

    switch (mode) {
      case "yaml":
        extensions.push(
          yamlLinter,
          autocompletion(
            completionValues ? { override: [yamlCompletions] } : undefined,
          ),
        );
        break;
      case "json":
        extensions.push(linter(jsonParseLinter()));
        break;
    }

    return (
      <CodeMirror
        readOnly={disabled}
        style={{ height: "100%" }}
        value={value}
        onChange={onChange}
        theme={dracula}
        basicSetup={{
          lineNumbers: true,
          highlightActiveLineGutter: false,
          foldGutter: true,
          dropCursor: true,
          allowMultipleSelections: true,
          indentOnInput: true,
          bracketMatching: true,
          closeBrackets: true,
          autocompletion: true,
          rectangularSelection: true,
          crosshairCursor: true,
          highlightActiveLine: false,
          highlightSelectionMatches: true,
          closeBracketsKeymap: true,
          searchKeymap: true,
          foldKeymap: true,
          completionKeymap: true,
          lintKeymap: true,
        }}
        extensions={extensions}
      />
    );
  },
  (prevProps, nextProps) => omitEqual(prevProps, nextProps, "onChange"),
);

const theme = EditorView.theme(
  {
    "&": {
      outline: "none",
      height: "100%",
    },
    ".cm-content": {
      paddingTop: 0,
    },
    "&.cm-editor .cm-scroller .cm-gutters": {
      marginRight: "5px",
    },
    "&.cm-editor.cm-focused": {
      outline: "none",
    },
    ".cm-completionIcon-text": {
      "&:after": { content: "''", fontSize: "50%", verticalAlign: "middle" },
    },
    ".cm-line": {
      padding: "1px 0",
    },
  },
  {},
);

function yamlCompletions(context: CompletionContext) {
  let word: any = context.matchBefore(/\w*/);

  if (word.from === word.to && !context.explicit) return null;
  return {
    from: word.from,
    options: [...list],
  };
}

const list = [
  {
    apply: "<.ImagePullSecrets>",
    label: "<.ImagePullSecrets>",
    type: "text",
    info: (completion: Completion) => {
      const div = document.createElement("div");
      div.textContent = `- name: secret1
- name: secret2
- name: secret3
`;
      return div;
    },
  },
  {
    apply: "<.ImagePullSecretsNoName>",
    label: "<.ImagePullSecretsNoName>",
    type: "text",
    info: (completion: Completion) => {
      const div = document.createElement("div");
      div.textContent = `- secret1
- secret2
- secret3
`;
      return div;
    },
  },
  { apply: "<.Branch>", label: "<.Branch>", type: "text" },
  { apply: "<.Commit>", label: "<.Commit>", type: "text" },
  { apply: "<.Pipeline>", label: "<.Pipeline>", type: "text" },
  { apply: "<.ClusterIssuer>", label: "<.ClusterIssuer>", type: "text" },
  { apply: "<.Host1>", label: "<.Host1>", type: "text" },
  { apply: "<.Host2>", label: "<.Host2>", type: "text" },
  { apply: "<.Host3>", label: "<.Host3>", type: "text" },
  { apply: "<.Host4>", label: "<.Host4>", type: "text" },
  { apply: "<.Host5>", label: "<.Host5>", type: "text" },
  { apply: "<.Host6>", label: "<.Host6>", type: "text" },
  { apply: "<.Host7>", label: "<.Host7>", type: "text" },
  { apply: "<.Host8>", label: "<.Host8>", type: "text" },
  { apply: "<.Host9>", label: "<.Host9>", type: "text" },
  { apply: "<.Host10>", label: "<.Host10>", type: "text" },
  { apply: "<.TlsSecret1>", label: "<.TlsSecret1>", type: "text" },
  { apply: "<.TlsSecret2>", label: "<.TlsSecret2>", type: "text" },
  { apply: "<.TlsSecret3>", label: "<.TlsSecret3>", type: "text" },
  { apply: "<.TlsSecret4>", label: "<.TlsSecret4>", type: "text" },
  { apply: "<.TlsSecret5>", label: "<.TlsSecret5>", type: "text" },
  { apply: "<.TlsSecret6>", label: "<.TlsSecret6>", type: "text" },
  { apply: "<.TlsSecret7>", label: "<.TlsSecret7>", type: "text" },
  { apply: "<.TlsSecret8>", label: "<.TlsSecret8>", type: "text" },
  { apply: "<.TlsSecret9>", label: "<.TlsSecret9>", type: "text" },
  { apply: "<.TlsSecret10>", label: "<.TlsSecret10>", type: "text" },
  {
    apply: 'cert-manager.io/cluster-issuer: "<.ClusterIssuer>"',
    label: "certManager",
  },
  {
    apply: `"<.Branch>-<.Pipeline>"`,
    label: "imageTag",
    detail: "<.Branch>-<.Pipeline>",
  },
  {
    apply: `mars.duc-cnzj.github.io/ignore-containers: "app1,app2" # 自行修改 app1,app2 的值`,
    label: "annotationIgnoreContainerNames",
    detail: `# 过滤容器`,
    info: (completion: Completion) => {
      const div = document.createElement("div");
      div.textContent = `mars.duc-cnzj.github.io/ignore-containers: "app1,app2"`;
      return div;
    },
  },
  {
    apply: `mars.duc-cnzj.github.io/order-index: "10" # 值为字符串类型, 数值越大越靠前`,
    label: "annotationPodOrderIndex",
    detail: `# 排序，数值越大越靠前`,
    info: (completion: Completion) => {
      const div = document.createElement("div");
      div.textContent = `mars.duc-cnzj.github.io/order-index: "10"`;
      return div;
    },
  },
];

const getLangs = (name: string) => {
  let res = (langs as any)[name];
  if (!res) {
    res = langs["textile"];
  }
  return res();
};

export const getMode = (mode: string): string => {
  switch (mode) {
    case "dotenv":
    case "env":
    case ".env":
      return "textile";
    case "js":
      return "javascript";
    case "ini":
      return "properties";
    case "py":
      return "python";
    default:
      return mode;
  }
};
