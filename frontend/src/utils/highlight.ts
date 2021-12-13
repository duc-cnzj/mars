import Prism from "prismjs";

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
