import React, { memo, useState } from "react";
import { Popover, Button, Collapse, Tooltip, Spin } from "antd";
import { CloseOutlined, HistoryOutlined } from "@ant-design/icons";
import DiffViewer from "./DiffViewer";
import ajax from "../api/ajax";
import { components } from "../api/schema";
const { Panel } = Collapse;

const ConfigHistory: React.FC<{
  projectID: number;
  configType: string;
}> = ({ projectID, configType }) => {
  const [list, setList] = useState<
    components["schemas"]["types.ChangelogModel"][]
  >([]);

  const [loading, setLoading] = useState(false);
  const [visible, setVisible] = useState(false);
  return (
    <Popover
      placement="right"
      open={visible}
      destroyTooltipOnHide
      onOpenChange={(v) => {
        if (v) {
          setLoading(true);
          ajax
            .POST("/api/changelogs/find_last_changelogs_by_project_id", {
              body: {
                projectId: projectID,
                onlyChanged: true,
              },
            })
            .then(({ data }) => {
              data && setList(data.items);
              setLoading(false);
            });
        }
      }}
      content={
        <Content loading={loading} list={list} configType={configType} />
      }
      trigger="click"
      title={
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
          }}
        >
          <span>历史修改记录</span>{" "}
          <Button
            size="small"
            type="link"
            onClick={() => setVisible(false)}
            icon={<CloseOutlined />}
          ></Button>
        </div>
      }
    >
      <Button
        size="small"
        onClick={() => {
          setVisible((v) => !v);
        }}
        style={{ fontSize: 12 }}
        icon={<HistoryOutlined />}
      />
    </Popover>
  );
};

const Content: React.FC<{
  list: components["schemas"]["types.ChangelogModel"][];
  configType: string;
  loading: boolean;
}> = ({ configType, list, loading }) => {
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
        {!loading ? (
          list.map((item, idx) => (
            <Panel
              key={item.version}
              header={
                <>
                  <div style={{ display: "flex", alignItems: "center" }}>
                    <div style={{ flexShrink: 0 }}>
                      <span style={{ color: "red", margin: "0 5px" }}>
                        {item.username}
                      </span>{" "}
                      于{" "}
                      <strong style={{ margin: "0 5px" }}>{item.date}</strong>{" "}
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
                            {item.gitCommitAuthor} 提交于
                            {item.gitCommitDate}
                          </div>
                        }
                      >
                        <a href={item.gitCommitWebUrl} target="_blank">
                          {item.gitCommitTitle}
                        </a>
                      </Tooltip>
                    </div>
                  </div>
                </>
              }
            >
              <div style={{ marginTop: 5 }}>
                <div>
                  <DiffViewer
                    mode={configType}
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
                    showDiffOnly
                    oldValue={
                      list && idx + 1 < list.length ? list[idx + 1].config : ""
                    }
                    newValue={item.config}
                    splitView={false}
                  />
                </div>
              </div>
            </Panel>
          ))
        ) : (
          <div
            style={{
              display: "flex",
              justifyContent: "center",
              alignItems: "center",
              padding: "5px 0",
            }}
          >
            <Spin />
          </div>
        )}
      </Collapse>
    </div>
  );
};

export default memo(ConfigHistory);
