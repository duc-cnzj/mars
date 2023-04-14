import { languages } from "@codemirror/language-data";
import { StreamLanguage } from "@codemirror/language";
import { markdown, markdownLanguage } from "@codemirror/lang-markdown";
import { javascript } from "@codemirror/lang-javascript";
import { html } from "@codemirror/lang-html";
import { css } from "@codemirror/lang-css";
import { json } from "@codemirror/lang-json";
import { python } from "@codemirror/lang-python";
import { xml } from "@codemirror/lang-xml";
import { sql, MySQL, PostgreSQL } from "@codemirror/lang-sql";
import { java } from "@codemirror/lang-java";
import { rust } from "@codemirror/lang-rust";
import { cpp } from "@codemirror/lang-cpp";
import { php } from "@codemirror/lang-php";
import {
  c,
  csharp,
  scala,
  kotlin,
  objectiveC,
  objectiveCpp,
  dart,
} from "@codemirror/legacy-modes/mode/clike";
import { less } from "@codemirror/legacy-modes/mode/css";
import { cmake } from "@codemirror/legacy-modes/mode/cmake";
import { coffeeScript } from "@codemirror/legacy-modes/mode/coffeescript";
import { dockerFile } from "@codemirror/legacy-modes/mode/dockerfile";
import { erlang } from "@codemirror/legacy-modes/mode/erlang";
import { go } from "@codemirror/legacy-modes/mode/go";
import { groovy } from "@codemirror/legacy-modes/mode/groovy";
import { haskell } from "@codemirror/legacy-modes/mode/haskell";
import { http } from "@codemirror/legacy-modes/mode/http";
import { jinja2 } from "@codemirror/legacy-modes/mode/jinja2";
import { julia } from "@codemirror/legacy-modes/mode/julia";
import { lua } from "@codemirror/legacy-modes/mode/lua";
import { nginx } from "@codemirror/legacy-modes/mode/nginx";
import { perl } from "@codemirror/legacy-modes/mode/perl";
import { powerShell } from "@codemirror/legacy-modes/mode/powershell";
import { properties } from "@codemirror/legacy-modes/mode/properties";
import { protobuf } from "@codemirror/legacy-modes/mode/protobuf";
import { puppet } from "@codemirror/legacy-modes/mode/puppet";
import { ruby } from "@codemirror/legacy-modes/mode/ruby";
import { sass } from "@codemirror/legacy-modes/mode/sass";
import { shell } from "@codemirror/legacy-modes/mode/shell";
import { stylus } from "@codemirror/legacy-modes/mode/stylus";
import { swift } from "@codemirror/legacy-modes/mode/swift";
import { textile } from "@codemirror/legacy-modes/mode/textile";
import { toml } from "@codemirror/legacy-modes/mode/toml";
import { vb } from "@codemirror/legacy-modes/mode/vb";
import { vbScript } from "@codemirror/legacy-modes/mode/vbscript";
import { yaml } from "@codemirror/legacy-modes/mode/yaml";

export const langs = {
  c: () => StreamLanguage.define(c),
  csharp: () => StreamLanguage.define(csharp),
  scala: () => StreamLanguage.define(scala),
  kotlin: () => StreamLanguage.define(kotlin),
  objectiveC: () => StreamLanguage.define(objectiveC),
  objectiveCpp: () => StreamLanguage.define(objectiveCpp),
  dart: () => StreamLanguage.define(dart),
  cmake: () => StreamLanguage.define(cmake),
  groovy: () => StreamLanguage.define(groovy),
  haskell: () => StreamLanguage.define(haskell),
  http: () => StreamLanguage.define(http),
  jinja2: () => StreamLanguage.define(jinja2),
  properties: () => StreamLanguage.define(properties),
  protobuf: () => StreamLanguage.define(protobuf),
  puppet: () => StreamLanguage.define(puppet),
  sass: () => StreamLanguage.define(sass),
  textile: () => StreamLanguage.define(textile),
  javascript,
  jsx: () => javascript({ jsx: true }),
  typescript: () => javascript({ typescript: true }),
  tsx: () => javascript({ jsx: true, typescript: true }),
  json,
  html,
  css,
  python,
  markdown: () =>
    markdown({ base: markdownLanguage, codeLanguages: languages }),
  xml,
  sql,
  mysql: () => sql({ dialect: MySQL }),
  pgsql: () => sql({ dialect: PostgreSQL }),
  java,
  rust,
  cpp,
  php,
  go: () => StreamLanguage.define(go),
  shell: () => StreamLanguage.define(shell),
  lua: () => StreamLanguage.define(lua),
  swift: () => StreamLanguage.define(swift),
  yaml: () => StreamLanguage.define(yaml),
  vb: () => StreamLanguage.define(vb),
  powershell: () => StreamLanguage.define(powerShell),
  stylus: () => StreamLanguage.define(stylus),
  erlang: () => StreamLanguage.define(erlang),
  nginx: () => StreamLanguage.define(nginx),
  perl: () => StreamLanguage.define(perl),
  ruby: () => StreamLanguage.define(ruby),
  less: () => StreamLanguage.define(less),
  toml: () => StreamLanguage.define(toml),
  vbscript: () => StreamLanguage.define(vbScript),
  coffeescript: () => StreamLanguage.define(coffeeScript),
  julia: () => StreamLanguage.define(julia),
  dockerfile: () => StreamLanguage.define(dockerFile),
};

/** Language list */
export const langNames = Object.keys(langs) as LanguageName[];
export type LanguageName = keyof typeof langs;

export function loadLanguage(name: LanguageName) {
  return langs[name] ? langs[name]() : null;
}
