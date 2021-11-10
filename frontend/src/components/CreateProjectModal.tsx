import React, { useState, useRef, useCallback, useEffect } from "react";
import { selectClusterInfo } from "../store/reducers/cluster";
import PipelineInfo from "./PipelineInfo";
import { DraggableModal } from "../pkg/DraggableModal/DraggableModal";
import { MyCodeMirror as CodeMirror, getMode } from "./MyCodeMirror";

import { configFile } from "../api/gitlab";
import {
  DeployStatus as DeployStatusEnum,
  selectList,
} from "../store/reducers/createProject";
import { useWs, useWsReady } from "../contexts/useWebsocket";
import { message, Progress, Button } from "antd";
import {
  PlusOutlined,
  StopOutlined,
  ArrowLeftOutlined,
  ArrowRightOutlined,
} from "@ant-design/icons";
import { useDispatch, useSelector } from "react-redux";
import {
  clearCreateProjectLog,
  setCreateProjectLoading,
  setDeployStatus,
} from "../store/actions";
import classNames from "classnames";
import { toSlug } from "../utils/slug";
import LogOutput from "./LogOutput";
import ProjectSelector from "./ProjectSelector";
import DebugModeSwitch from "./DebugModeSwitch";
import TimeCost from "./TimeCost";

const initItemData: Mars.CreateItemInterface = {
  name: "",
  gitlabProjectId: 0,
  gitlabBranch: "",
  gitlabCommit: "",
  config: "",
  debug: false,
  config_type: "yaml",
};

const CreateProjectModal: React.FC<{
  namespaceId: number;
}> = ({ namespaceId }) => {
  const list = useSelector(selectList);
  const dispatch = useDispatch();
  const [data, setData] = useState<Mars.CreateItemInterface>(initItemData);
  const [mode, setMode] = useState<string>("text/x-yaml");
  const [visible, setVisible] = useState<boolean>(false);
  const [editVisible, setEditVisible] = useState<boolean>(true);
  const [timelineVisible, setTimelineVisible] = useState<boolean>(false);

  let slug = toSlug(namespaceId, data.name);

  const onCancel = useCallback(() => {
    setVisible(false);
    setEditVisible(true);
    setTimelineVisible(false);
    setData(initItemData);
    dispatch(clearCreateProjectLog(slug));
  }, [dispatch, slug]);

  useEffect(() => {
    if (list[slug]?.deployStatus === DeployStatusEnum.DeploySuccess) {
      setTimelineVisible(false);
      setEditVisible(true);
      dispatch(setDeployStatus(slug, DeployStatusEnum.DeployUnknown));
      setTimeout(() => {
        setVisible(false);
        setData(initItemData);
      }, 500);
    }
  }, [list, dispatch, slug]);
  useEffect(() => {
    if (list[slug]?.deployStatus !== DeployStatusEnum.DeployUnknown) {
      setStart(false);
    }
  }, [list, slug]);

  const onChange = ({
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
  };
  const cmref = useRef<any>()

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
  }, [data.gitlabBranch, data.gitlabProjectId]);

  useEffect(() => {
    setMode(getMode(data.config_type))
  }, [data.config_type]);

  useEffect(() => {
    if (cmref && cmref.current && data.config) {
      cmref.current.editor.setSize("100%", "100%");
    }
  }, [data.config]);

  const ws = useWs();
  const wsReady = useWsReady();

  useEffect(() => {
    if (!wsReady) {
      setStart(false);
      dispatch(setCreateProjectLoading(slug, false));
    }
  }, [wsReady]);

  const onOk = useCallback(() => {
    console.log(data);
    if (!wsReady) {
      message.error("连接断开了");
      return;
    }
    if (data.gitlabProjectId && data.gitlabBranch && data.gitlabCommit) {
      // todo ws connected!
      setEditVisible(false);
      setTimelineVisible(true);

      let re = {
        type: "create_project",
        data: JSON.stringify({
          namespace_id: Number(namespaceId),
          name: data.name,
          gitlab_project_id: Number(data.gitlabProjectId),
          gitlab_branch: data.gitlabBranch,
          gitlab_commit: data.gitlabCommit,
          config: data.config,
          atomic: !data.debug,
        }),
      };

      dispatch(setDeployStatus(slug, DeployStatusEnum.DeployUnknown));

      let s = JSON.stringify(re);
      dispatch(clearCreateProjectLog(slug));
      dispatch(setCreateProjectLoading(slug, true));
      setStart(true);
      ws?.send(s);
      return;
    }

    message.error("项目id, 分支，提交必填");
  }, [data, dispatch, slug, ws, namespaceId, wsReady]);
  const onRemove = useCallback(() => {
    if (data.gitlabProjectId && data.gitlabBranch && data.gitlabCommit) {
      let re = {
        type: "cancel_project",
        data: JSON.stringify({
          namespace_id: Number(namespaceId),
          name: data.name,
        }),
      };

      let s = JSON.stringify(re);
      ws?.send(s);
      return;
    }
  }, [data, ws, namespaceId]);

  const [start, setStart] = useState(false);
  const info = useSelector(selectClusterInfo);

  return (
    <div>
      <Button
        onClick={() => setVisible(true)}
        style={{ width: "100%" }}
        type="dashed"
        icon={<PlusOutlined />}
      ></Button>
      <DraggableModal
        visible={visible}
        okButtonProps={{
          loading: list[slug]?.isLoading,
          danger: info.status === "bad",
        }}
        cancelButtonProps={{ disabled: list[slug]?.isLoading }}
        closable={!list[slug]?.isLoading}
        okText={info.status === "bad" ? "集群资源不足" : "部署"}
        cancelText="取消"
        onOk={onOk}
        initialWidth={800}
        initialHeight={500}
        title={<div style={{textAlign: "center"}}>创建项目</div>}
        className="drag-item-modal"
        onCancel={onCancel}
      >
        <PipelineInfo
          projectId={data.gitlabProjectId}
          branch={data.gitlabBranch}
          commit={data.gitlabCommit}
        />
        <div className={classNames({ "display-none": !editVisible })}>
          <div
            style={{ display: "flex", alignItems: "center", marginBottom: 10 }}
          >
            {list[slug]?.output?.length > 0 ? (
              <Button
                style={{ marginRight: 5 }}
                type="dashed"
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
            <ProjectSelector onChange={onChange} />
          </div>
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              alignItems: "center",
              paddingBottom: 10,
            }}
          >
            <span>配置文件:</span>
            <DebugModeSwitch
              value={data.debug}
              onchange={(checked: boolean, event: MouseEvent) => {
                setData((data) => ({ ...data, debug: checked }));
              }}
            />
          </div>
          <div
            style={{
              minWidth: 200,
              maxWidth: 1280,
              marginBottom: 20,
              height: "100%",
            }}
          >
            <CodeMirror
              ref={cmref}
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
            <Button
              hidden={
                list[slug]?.deployStatus === DeployStatusEnum.DeployCanceled
              }
              style={{ marginRight: 10 }}
              danger
              icon={<StopOutlined />}
              type="dashed"
              onClick={onRemove}
            >
              取消
            </Button>
            <TimeCost start={start} />
          </div>

          <LogOutput slug={slug} />
        </div>
      </DraggableModal>
    </div>
  );
};

export default CreateProjectModal;
