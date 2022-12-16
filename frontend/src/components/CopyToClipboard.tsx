import React, { memo } from "react";
import { copy } from "../utils/copy";

const CopyToClipboard: React.FC<{
  text: string;
  successText?: string;
  children: React.ReactNode;
}> = ({ text, successText, children }) => {
  return (
    <span style={{ cursor: "pointer" }} onClick={() => copy(text, successText)}>
      {children}
    </span>
  );
};

export default memo(CopyToClipboard);
