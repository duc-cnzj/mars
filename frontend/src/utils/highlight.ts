import Prism from "prismjs";

const getLoader = require("prismjs/dependencies");
const components = require("prismjs/components");

const componentsToLoad = [
  "markup",
  "css",
  "php",
  "yaml",
  "go",
  "ini",
  "python",
  "javascript",
];
const loadedComponents = [""];

const loader = getLoader(components, componentsToLoad, loadedComponents);
loader.load((id: string) => {
  require(`prismjs/components/prism-${id}.min.js`);
});

export const getHighlightSyntax = (str: string, lang: string): string => {
  switch (lang) {
    case "yaml":
      return Prism.highlight(str, Prism.languages.yaml, "yaml");
    case "php":
      return Prism.highlight(str, Prism.languages.php, "php");
    case "py":
    case "python":
      return Prism.highlight(str, Prism.languages.python, "python");
    case "go":
    case "golang":
      return Prism.highlight(str, Prism.languages.go, "go");
    case "js":
    case "javascript":
      return Prism.highlight(str, Prism.languages.javascript, "javascript");
    case "ini":
      return Prism.highlight(str, Prism.languages.ini, "ini");
    default:
      return str;
  }
};
