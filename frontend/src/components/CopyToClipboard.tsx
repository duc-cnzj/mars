import React from "react";
import { copy } from "../utils/copy";

const CopyToClipboard: React.FC<{ text: string; successText?: string }> = ({
  text,
  successText,
  children,
}) => {
  return (
    <span style={{ cursor: "pointer" }} onClick={() => copy(text, successText)}>
      {children}
    </span>
  );
};

export default CopyToClipboard;
