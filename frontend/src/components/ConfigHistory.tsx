import React, { memo, useEffect, useCallback, useState } from "react";
import { Popover, Button, Collapse, Tooltip } from "antd";
import { HistoryOutlined, CarryOutOutlined } from "@ant-design/icons";
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
  const [list, setList] = useState<pb.types.ChangelogModel[]>();
  useEffect(() => {
    changelogs({ project_id: projectID, only_changed: true }).then((res) => {
      setList(res.data.items);
    });
  }, [projectID, updatedAt]);

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
    <div
      style={{
        width: 600,
        maxHeight: 600,
        overflowY: "auto",
        pointerEvents: "auto",
      }}
    >
      <Collapse accordion bordered={false}>
        {list?.map((item) => (
          <Panel
            key={item.version}
            header={
              <>
                <div style={{ display: "flex", alignItems: "center" }}>
                  <div style={{ flexShrink: 0 }}>
                    <span style={{ color: "red", margin: "0 5px" }}>
                      {item.username}
                    </span>{" "}
                    于 <strong style={{ margin: "0 5px" }}>{item.date}</strong>{" "}
                    更新了项目
                  </div>

                  <div
                    style={{
                      textOverflow: "ellipsis",
                      whiteSpace: "nowrap",
                      overflow: "hidden",
                      marginLeft: 5,
                      flexShrink: 1,
                    }}
                  >
                    <Tooltip
                      placement="top"
                      title={
                        <div style={{ fontSize: 12 }}>
                          {item.git_commit_author} 提交于
                          {item.git_commit_date}
                        </div>
                      }
                    >
                      <a href={item.git_commit_web_url} target="_blank">
                        {item.git_commit_title}
                      </a>
                    </Tooltip>
                  </div>
                </div>
              </>
            }
          >
            <div style={{ marginTop: 5 }}>
              {currentConfig === item.config ? (
                <div>和当前配置一致</div>
              ) : (
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
                    newValue={item.config}
                    splitView={false}
                  />

                  <div
                    style={{ display: "flex", flexDirection: "row-reverse" }}
                  >
                    <Button
                      onClick={() => {
                        onDataChange(item.config);
                      }}
                      size="small"
                      type="dashed"
                      style={{ marginTop: 3 }}
                    >
                      使用这个配置
                    </Button>
                  </div>
                </div>
              )}
            </div>
          </Panel>
        ))}
      </Collapse>
    </div>
  );
};

export default memo(ConfigHistory);
