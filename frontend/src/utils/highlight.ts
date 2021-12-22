import Prism from "prismjs";

require("prismjs/components/prism-markup-templating");
require("prismjs/components/prism-markup");
require("prismjs/components/prism-css");
require("prismjs/components/prism-php");
require("prismjs/components/prism-yaml");
require("prismjs/components/prism-go");
require("prismjs/components/prism-ini");
require("prismjs/components/prism-python");
require("prismjs/components/prism-javascript");

export const getHighlightSyntax = (str: string, lang: string): string => {
  try {
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
  } catch (e) {
    console.log(e);
    return str;
  }
};
