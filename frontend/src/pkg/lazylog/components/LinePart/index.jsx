import { Component } from "react";
import { func, object, shape, string } from "prop-types";
import styles from "./index.module.css";
import Ansi from "ansi-to-react";

const getClassName = (part) => {
  const className = [];

  if (part.foreground && part.bold) {
    className.push(styles[`${part.foreground}Bold`], styles.bold);
  } else if (part.foreground) {
    className.push(styles[part.foreground]);
  } else if (part.bold) {
    className.push(styles.bold);
  }

  if (part.background) {
    className.push(styles[`${part.background}Bg`]);
  }

  if (part.italic) {
    className.push(styles.italic);
  }

  if (part.underline) {
    className.push(styles.underline);
  }

  return className.join(" ");
};

/**
 * An individual segment of text within a line. When the text content
 * is ANSI-parsed, each boundary is placed within its own `LinePart`
 * and styled separately (colors, text formatting, etc.) from the
 * rest of the line's content.
 */
export default class LinePart extends Component {
  static propTypes = {
    /**
     * The pieces of data to render in a line. Will typically
     * be multiple items in the array if ANSI parsed prior.
     */
    part: shape({
      text: string,
    }).isRequired,
    /**
     * Execute a function against each line part's
     * `text` property in `data` to process and
     * return a new value to render for the part.
     */
    format: func,
    style: object,
  };

  static defaultProps = {
    format: null,
    style: null,
  };

  render() {
    const { format, part, style } = this.props;
    let text = format ? format(part.text) : part.text;
    let itemText = [];
    if (typeof text === "string") {
      itemText.push(<Ansi>{text}</Ansi>);
    } else if (typeof text === "object") {
      for (let index = 0; index < text.length; index++) {
        const element = text[index];
        if (typeof element === "string") {
          itemText.push(<Ansi>{element}</Ansi>);
        } else {
          itemText.push(element);
        }
      }
    }

    return (
      <span className={getClassName(part)} style={style}>
        {itemText.map((v, k) => (
          <span key={k}>{v}</span>
        ))}
      </span>
    );
  }
}
