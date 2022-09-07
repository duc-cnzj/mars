import React, { memo, useCallback } from "react";
import ReactDiffViewer, {
  ReactDiffViewerStylesOverride,
} from "react-diff-viewer";
import { getHighlightSyntax } from "../utils/highlight";

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
}> = ({ mode, oldValue, newValue, splitView, showDiffOnly, styles }) => {
  const highlightSyntax = useCallback(
    (str: string) => (
      <pre
        style={{ display: 'inline' }}
        dangerouslySetInnerHTML={{
          __html: getHighlightSyntax(str, mode),
        }}
      />
    ),
    [mode]
  );
  return (
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
  );
};

export default memo(DiffViewer);
