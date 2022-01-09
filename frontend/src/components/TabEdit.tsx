import React, { useMemo, useState, useEffect, useCallback, memo } from "react";
import { MyCodeMirror as CodeMirror, getMode } from "./MyCodeMirror";
import ReactDiffViewer from "react-diff-viewer";
import PipelineInfo from "./PipelineInfo";
import ConfigHistory from "./ConfigHistory";
import { commit, configFile, allProjects } from "../api/gitlab";
import { getHighlightSyntax } from "../utils/highlight";
import {
  DeployStatus as DeployStatusEnum,
  selectList,
} from "../store/reducers/createProject";
import { Button, Skeleton, Progress, message, Row, Col } from "antd";
import { useSelector, useDispatch } from "react-redux";
import {
  clearCreateProjectLog,
  setCreateProjectLoading,
  setDeployStatus,
} from "../store/actions";
import { toSlug } from "../utils/slug";
import { useWs, useWsReady } from "../contexts/useWebsocket";
import {
  ArrowLeftOutlined,
  StopOutlined,
  ArrowRightOutlined,
} from "@ant-design/icons";
import classNames from "classnames";
import LogOutput from "./LogOutput";
import ProjectSelector from "./ProjectSelector";
import TimeCost from "./TimeCost";
import DebugModeSwitch from "./DebugModeSwitch";
import pb from "../api/compiled";

const ModalSub: React.FC<{
  detail: pb.ProjectShowResponse;
  onSuccess: () => void;
}> = ({ detail, onSuccess }) => {
  let id = detail.id;
  let namespaceId = detail.namespace?.id;
  const ws = useWs();
  const wsReady = useWsReady();

  const [editVisible, setEditVisible] = useState<boolean>(true);
  const [timelineVisible, setTimelineVisible] = useState<boolean>(false);
  const list = useSelector(selectList);
  const dispatch = useDispatch();
  const [data, setData] = useState<Mars.CreateItemInterface>({
    name: detail.name,
    gitlabProjectId: Number(detail.gitlab_project_id),
    gitlabBranch: detail.gitlab_branch,
    gitlabCommit: detail.gitlab_commit,
    config: detail.config,
    config_type: "yaml",
    debug: !detail.atomic,
  });
  const [mode, setMode] = useState<string>("text/x-yaml");
  const [initValue, setInitValue] = useState<{
    projectName: string;
    gitlabProjectId: string;
    gitlabBranch: string;
    gitlabCommit: string;
    time?: number;
  }>();
  let slug = useMemo(
    () => toSlug(namespaceId || 0, data.name),
    [namespaceId, data.name]
  );

  // 初始化，设置 initvalue
  useEffect(() => {
    allProjects().then((res) => {
      if (
        detail &&
        detail.gitlab_project_id &&
        detail.gitlab_branch &&
        detail.gitlab_commit
      ) {
        configFile({
          project_id: String(detail.gitlab_project_id),
          branch: detail.gitlab_branch,
        }).then((res) => {
          setData((d) => ({ ...d, config_type: res.data.type }));
        });
        commit({
          project_id: String(detail.gitlab_project_id),
          branch: detail.gitlab_branch,
          commit: detail.gitlab_commit,
        }).then((res) => {
          res.data.data &&
            setInitValue({
              projectName: detail.name,
              gitlabProjectId: String(detail.gitlab_project_id),
              gitlabBranch: detail.gitlab_branch,
              gitlabCommit: res.data.data.label,
            });
        });
      }
    });
  }, [setInitValue, detail]);

  // 更新成功，触发 onSuccess
  useEffect(() => {
    if (list[slug]?.deployStatus === DeployStatusEnum.DeployUpdateSuccess) {
      setStart(false);
      setTimelineVisible(false);
      setEditVisible(true);
      dispatch(setDeployStatus(slug, DeployStatusEnum.DeployUnknown));
      onSuccess();
    }
  }, [list, dispatch, slug, onSuccess]);

  // 更新 config 文件的类型， TODO 支持动态加载 mode css 文件
  const loadConfigFile = useCallback(() => {
    configFile({
      project_id: String(data.gitlabProjectId),
      branch: data.gitlabBranch,
    }).then((res) => {
      setData((d) => ({
        ...d,
        config: res.data.data,
        config_type: res.data.type,
      }));
    });
  }, [data.gitlabProjectId, data.gitlabBranch]);

  const onChange = useCallback(
    ({
      projectName,
      gitlabProjectId,
      gitlabBranch,
      gitlabCommit,
    }: {
      projectName: string;
      gitlabProjectId: number;
      gitlabBranch: string;
      gitlabCommit: string;
    }) => {
      setData((d) => ({
        ...d,
        name: projectName,
        gitlabProjectId: gitlabProjectId,
        gitlabBranch: gitlabBranch,
        gitlabCommit: gitlabCommit,
      }));

      if (gitlabCommit !== "" && data.config === "") {
        loadConfigFile();
      }
    },
    [loadConfigFile, data.config]
  );
  useEffect(() => {
    if (!wsReady) {
      setStart(false);
      dispatch(setCreateProjectLoading(slug, false));
    }
  }, [wsReady, dispatch, slug]);
  const updateDeploy = () => {
    if (!wsReady) {
      message.error("连接断开了");
      return;
    }
    if (data.gitlabCommit && data.gitlabBranch) {
      setStart(true);
      setEditVisible(false);
      setTimelineVisible(true);

      let s = pb.UpdateProjectInput.encode({
        type: pb.Type.UpdateProject,

        project_id: Number(id),
        gitlab_branch: data.gitlabBranch,
        gitlab_commit: data.gitlabCommit,
        config: data.config,
        atomic: !data.debug,
      }).finish();
      dispatch(setDeployStatus(slug, DeployStatusEnum.DeployUnknown));

      dispatch(clearCreateProjectLog(slug));
      dispatch(setCreateProjectLoading(slug, true));
      ws?.send(s);
    }
  };
  const [start, setStart] = useState(false);

  useEffect(() => {
    if (list[slug]?.deployStatus !== DeployStatusEnum.DeployUnknown) {
      setStart(false);
    }
  }, [list, slug]);

  const onReset = () => {
    setData({
      name: detail.name,
      gitlabProjectId: Number(detail.gitlab_project_id),
      gitlabBranch: detail.gitlab_branch,
      gitlabCommit: detail.gitlab_commit,
      config: detail.config,
      debug: !detail.atomic,
      config_type: data.config_type,
    });
    if (initValue) {
      setInitValue({ ...initValue, time: new Date().getUTCSeconds() });
    }
  };

  const onRemove = useCallback(() => {
    if (!wsReady) {
      message.error("连接断开了");
      return;
    }
    if (data.gitlabProjectId && data.gitlabBranch && data.gitlabCommit) {
      let s = pb.CancelInput.encode({
        type: pb.Type.CancelProject,
        namespace_id: Number(namespaceId),
        name: data.name,
      }).finish();
      ws?.send(s);
      return;
    }
  }, [data, ws, namespaceId, wsReady]);

  useEffect(() => {
    setMode(getMode(data.config_type));
  }, [data.config_type]);

  const highlightSyntax = useCallback(
    (str: string) => (
      <code
        dangerouslySetInnerHTML={{
          __html: getHighlightSyntax(str, data.config_type),
        }}
      />
    ),
    [data.config_type]
  );

  return (
    <div className="edit-project">
      <div
        className={classNames({ "display-none": !editVisible })}
        style={{ height: "100%", display: "flex", flexDirection: "column" }}
      >
        <PipelineInfo
          projectId={data.gitlabProjectId}
          branch={data.gitlabBranch}
          commit={data.gitlabCommit}
        />
        <div
          style={{
            width: "100%",
            display: "flex",
            alignItems: "center",
            marginBottom: 10,
          }}
        >
          {list[slug]?.output?.length > 0 ? (
            <Button
              type="dashed"
              style={{ marginRight: 5 }}
              disabled={list[slug]?.isLoading}
              onClick={() => {
                setEditVisible(false);
                setTimelineVisible(true);
              }}
              icon={<ArrowRightOutlined />}
            />
          ) : (
            ""
          )}

          <Skeleton
            active
            paragraph={false}
            avatar={false}
            loading={!initValue}
            title={{ style: { marginTop: 0, height: 24 } }}
          >
            <ProjectSelector value={initValue} onChange={onChange} />
          </Skeleton>
        </div>
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
            paddingBottom: 10,
          }}
        >
          <div
            className={classNames("edit-project__footer", {
              "edit-project--hidden": list[slug]?.isLoading,
            })}
          >
            <span style={{ marginRight: 5 }}>配置文件:</span>

            <Button
              size="small"
              style={{ marginRight: 5, fontSize: 12 }}
              disabled={list[slug]?.isLoading}
              onClick={onReset}
            >
              重置
            </Button>
            <Button
              style={{ fontSize: 12, marginRight: 5 }}
              size="small"
              type="primary"
              loading={list[slug]?.isLoading}
              onClick={updateDeploy}
            >
              部署
            </Button>
            <ConfigHistory
              show={editVisible}
              onDataChange={(s: string) =>
                setData((data) => ({ ...data, config: s }))
              }
              projectID={detail.id}
              configType={data.config_type}
              currentConfig={data.config}
              updatedAt={detail.updated_at}
            />
          </div>

          <DebugModeSwitch
            value={data.debug}
            onchange={(checked: boolean, event: MouseEvent) => {
              setData((data) => ({ ...data, debug: checked }));
            }}
          />
        </div>
        <div style={{ minWidth: 200, marginBottom: 20, height: "100%" }}>
          <Row style={{ height: "100%" }}>
            <Col span={detail.config === data.config ? 24 : 12}>
              <CodeMirror
                value={data.config}
                options={{
                  mode: mode,
                  theme: "dracula",
                  lineNumbers: true,
                }}
                onBeforeChange={(editor, d, value) => {
                  console.log(editor, d, value);
                  setData({ ...data, config: value });
                }}
              />
            </Col>
            <Col
              className="diff-viewer"
              span={detail.config === data.config ? 0 : 12}
              style={{ fontSize: 13 }}
            >
              <ReactDiffViewer
                styles={{
                  gutter: { padding: "0 5px", minWidth: 25 },
                  marker: { padding: "0 6px" },
                  diffContainer: {
                    display: "block",
                    width: "100%",
                    overflowX: "auto",
                  },
                }}
                useDarkTheme
                disableWordDiff
                renderContent={highlightSyntax}
                showDiffOnly={false}
                oldValue={detail.config}
                newValue={data.config}
                splitView={false}
              />
            </Col>
          </Row>
        </div>
      </div>
      <div
        id="preview"
        style={{ height: "100%", overflow: "auto" }}
        className={classNames("preview", {
          "display-none": !timelineVisible,
        })}
      >
        <div
          style={{ display: "flex", alignItems: "center", marginBottom: 20 }}
        >
          <Button
            type="dashed"
            disabled={list[slug]?.isLoading}
            onClick={() => {
              setEditVisible(true);
              setTimelineVisible(false);
            }}
            icon={<ArrowLeftOutlined />}
          />
          <Progress
            strokeColor={{
              from: "#108ee9",
              to: "#87d068",
            }}
            style={{ padding: "0 10px" }}
            percent={list[slug]?.processPercent}
            status="active"
          />
        </div>
        <div
          style={{ display: "flex", alignItems: "center", marginBottom: 10 }}
        >
          <TimeCost start={start} />

          <Button
            size="small"
            type="primary"
            loading={list[slug]?.isLoading}
            onClick={updateDeploy}
            style={{ marginRight: 10, marginLeft: 10, fontSize: 12 }}
          >
            部署
          </Button>
          <Button
            style={{ fontSize: 12 }}
            size="small"
            hidden={
              list[slug]?.deployStatus === DeployStatusEnum.DeployCanceled
            }
            danger
            icon={<StopOutlined />}
            type="dashed"
            onClick={onRemove}
          >
            取消
          </Button>
        </div>
        <LogOutput slug={slug} />
      </div>
    </div>
  );
};

export default memo(ModalSub);
