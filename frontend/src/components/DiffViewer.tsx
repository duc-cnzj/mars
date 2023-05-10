import React, { memo, useCallback } from "react";
import ReactDiffViewer, {
  ReactDiffViewerStylesOverride,
} from "react-diff-viewer";
import { getHighlightSyntax } from "../utils/highlight";
import { Button } from "antd";
import { css } from "@emotion/css";
import { copy } from "../utils/copy";

const defaultStyle: ReactDiffViewerStylesOverride = {
  gutter: { padding: "0 5px", minWidth: 25 },
  marker: { padding: "0 6px" },
  diffContainer: {
    display: "block",
    width: "100%",
    overflowX: "auto",
  },
};

const DiffViewer: React.FC<{
  mode: string;
  oldValue: string;
  newValue: string;
  showDiffOnly: boolean;
  splitView: boolean;
  styles?: ReactDiffViewerStylesOverride;
  showCopyButton?: boolean;
}> = ({
  mode,
  oldValue,
  newValue,
  splitView,
  showDiffOnly,
  styles,
  showCopyButton,
}) => {
  const highlightSyntax = useCallback(
    (str?: string) => (
      <pre
        style={{ display: "inline" }}
        dangerouslySetInnerHTML={{
          __html: getHighlightSyntax(str || "", mode),
        }}
      />
    ),
    [mode]
  );
  return (
    <div style={{ height: "100%" }}>
      {showCopyButton && (
        <div
          className={css`
            display: flex;
            font-size: 12px;
            margin-bottom: 5px;
            justify-content: ${splitView ? "space-between" : "flex-start"};
          `}
        >
          <Button
            size="small"
            type="dashed"
            className={css`
              margin-right: ${!splitView ? "5px" : "0"};
            `}
            onClick={() => copy(oldValue)}
            danger
          >
            copy old
          </Button>
          <Button size="small" type="dashed" onClick={() => copy(newValue)}>
            copy new
          </Button>
        </div>
      )}
      <ReactDiffViewer
        styles={styles ? styles : defaultStyle}
        useDarkTheme
        disableWordDiff
        renderContent={highlightSyntax}
        showDiffOnly={showDiffOnly}
        oldValue={oldValue}
        newValue={newValue}
        splitView={splitView}
      />
    </div>
  );
};

export default memo(DiffViewer);
