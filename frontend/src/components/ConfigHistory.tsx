import React, { useEffect, useCallback, useState } from "react";
import { Popover, Button, Collapse } from "antd";
import { HistoryOutlined } from "@ant-design/icons";
import ReactDiffViewer from "react-diff-viewer";
import { getHighlightSyntax } from "../utils/highlight";
import { changelogs } from "../api/changelog";
import pb from "../api/compiled";
const { Panel } = Collapse;

const ConfigHistory: React.FC<{
  show: boolean;
  projectID: number;
  configType: string;
  currentConfig: string;
  onDataChange: (s: string) => void;
  updatedAt: any;
}> = ({
  currentConfig,
  projectID,
  configType,
  updatedAt,
  onDataChange,
  show,
}) => {
  const [visible, setVisible] = useState(false);
  useEffect(() => {
    if (!show) {
      setVisible(false);
    }
  }, [show]);
  return (
    <Popover
      placement="right"
      visible={visible}
      onVisibleChange={(v) => setVisible(v)}
      content={
        <Content
          onDataChange={(s) => {
            onDataChange(s);
            setVisible(false);
          }}
          updatedAt={updatedAt}
          projectID={projectID}
          configType={configType}
          currentConfig={currentConfig}
        />
      }
      trigger="click"
      title="历史修改记录"
    >
      <Button
        size="small"
        style={{ fontSize: 12 }}
        icon={<HistoryOutlined />}
      />
    </Popover>
  );
};

const Content: React.FC<{
  projectID: number;
  configType: string;
  currentConfig: string;
  updatedAt: any;
  onDataChange: (s: string) => void;
}> = ({ currentConfig, projectID, configType, updatedAt, onDataChange }) => {
  const [list, setList] = useState<pb.ChangelogGetResponse.Item[]>();
  const [data, setData] = useState("");
  useEffect(() => {
    changelogs({ project_id: projectID, only_changed: true }).then((res) =>
      setList(res.data.items.filter((i) => i.config !== currentConfig))
    );
  }, [projectID, updatedAt, currentConfig]);

  const highlightSyntax = useCallback(
    (str: string) => (
      <code
        dangerouslySetInnerHTML={{
          __html: getHighlightSyntax(str, configType),
        }}
      />
    ),
    [configType]
  );
  return (
    <div style={{ width: 600, maxHeight: 600, overflowY: "auto" }}>
      <Collapse
        accordion
        onChange={(k) => {
          let l = list?.find((item) => item.version === Number(k));
          l && setData(l.config);
        }}
      >
        {list?.map((item) => (
          <Panel
            key={item.version}
            header={
              <>
                [{item.date} {item.username}]: version {item.version}
              </>
            }
          >
            <div>
              <ReactDiffViewer
                disableWordDiff
                styles={{
                  line: { fontSize: 12 },
                  gutter: { padding: "0 5px", minWidth: 20 },
                  marker: { padding: "0 6px" },
                  diffContainer: {
                    display: "block",
                    width: "100%",
                    overflowX: "auto",
                  },
                }}
                useDarkTheme
                renderContent={highlightSyntax}
                showDiffOnly
                oldValue={currentConfig}
                newValue={data}
                splitView={false}
              />
              <div style={{ display: "flex", flexDirection: "row-reverse" }}>
                <Button
                  onClick={() => {
                    onDataChange(data);
                    setData("");
                  }}
                  size="small"
                  type="dashed"
                  style={{ marginTop: 3 }}
                >
                  使用这个配置
                </Button>
              </div>
            </div>
          </Panel>
        ))}
      </Collapse>
    </div>
  );
};

export default ConfigHistory;
