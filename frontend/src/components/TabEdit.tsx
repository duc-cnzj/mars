import React, { useState, useEffect, useCallback, memo } from "react";
import { Controlled as CodeMirror } from "react-codemirror2";
import PipelineInfo from "./PipelineInfo";

import {
  branches,
  commit,
  commits,
  configFile,
  Options,
  projects,
} from "../api/gitlab";
import {
  DeployStatus as DeployStatusEnum,
  selectList,
} from "../store/reducers/createProject";
import _ from "lodash";
import { CascaderOptionType } from "antd/lib/cascader";
import { Button, Cascader, Timeline, Skeleton, Progress } from "antd";
import "codemirror/lib/codemirror.css";
import "codemirror/theme/material.css";
import "codemirror/theme/dracula.css";
import { useSelector, useDispatch } from "react-redux";
import {
  clearCreateProjectLog,
  setCreateProjectLoading,
  setDeployStatus,
} from "../store/actions";
import { toSlug } from "../utils/slug";
import { useWs } from "../contexts/useWebsocket";
import { ArrowLeftOutlined, ArrowRightOutlined } from "@ant-design/icons";
import classNames from "classnames";
import { ProjectDetail } from "../api/project";

require("codemirror/mode/go/go");
require("codemirror/mode/css/css");
require("codemirror/mode/javascript/javascript");
require("codemirror/mode/yaml/yaml");
require("codemirror/mode/php/php");
require("codemirror/mode/textile/textile");

interface CreateItemInterface {
  gitlabProjectId: number;
  gitlabBranch: string;
  gitlabCommit: string;

  name: string;
  config: string;
}

const ModalSub: React.FC<{
  detail: ProjectDetail;
  onSuccess: () => void;
}> = ({ detail, onSuccess }) => {
  let id = detail.id;
  let namespaceId = detail.namespace.id;
  const ws = useWs();

  const [editVisible, setEditVisible] = useState<boolean>(true);
  const [timelineVisible, setTimelineVisible] = useState<boolean>(false);
  const list = useSelector(selectList);
  const dispatch = useDispatch();
  const [data, setData] = useState<CreateItemInterface>({
    name: detail.name,
    gitlabProjectId: Number(detail.gitlab_project_id),
    gitlabBranch: detail.gitlab_branch,
    gitlabCommit: detail.gitlab_commit,
    config: detail.config,
  });
  const [mode, setMode] = useState<string>("text/x-yaml");
  const [options, setOptions] = useState<Options[]>([]);
  const [value, setValue] = useState<string[]>([]);
  const [initValue, setInitValue] = useState<string[]>([]);
  let slug = toSlug(namespaceId, data.name);

  // 初始化，设置 initvalue
  useEffect(() => {
    projects().then((res) => {
      setOptions(() => {
        let data = res.data.data;
        // 如果是编辑，则需要获取对应的默认值
        if (
          detail &&
          detail.gitlab_project_id &&
          detail.gitlab_branch &&
          detail.gitlab_commit
        ) {
          commit(
            Number(detail.gitlab_project_id),
            detail.gitlab_branch,
            detail.gitlab_commit
          ).then((res) => {
            let match = data.find(
              (item) => Number(item.value) === Number(detail.gitlab_project_id)
            );
            if (match) {
              setValue([
                match.label,
                detail.gitlab_branch,
                res.data.data.label,
              ]);
              setInitValue([
                match.label,
                detail.gitlab_branch,
                res.data.data.label,
              ]);
            }
          });
        }
        if (detail && Number(detail.gitlab_project_id)) {
          let r = data.find(
            (item) => item.projectId === Number(detail.gitlab_project_id)
          );
          return r ? [r] : [];
        }

        return data;
      });
    });
  }, [setValue, detail]);

  // 更新成功，触发 onSuccess
  useEffect(() => {
    if (list[slug]?.deployStatus === DeployStatusEnum.DeployUpdateSuccess) {
      setTimelineVisible(false);
      setEditVisible(true);
      dispatch(setDeployStatus(slug, DeployStatusEnum.DeployUnknown));
      onSuccess();
    }
  }, [list, dispatch, slug, onSuccess]);

  // 更新 config 文件的类型， TODO 支持动态加载 mode css 文件
  const loadConfigFile = useCallback(() => {
    configFile(data.gitlabProjectId, data.gitlabBranch).then((res) => {
      setData((d) => ({ ...d, config: res.data.data.data }));
      switch (res.data.data.type) {
        case "dotenv":
        case "env":
        case ".env":
          setMode("text/x-textile");
          break;
        case "yaml":
          setMode("text/x-yaml");
          break;
        case "php":
          setMode("php");
          break;
        default:
          setMode(res.data.data.type);
          break;
      }
    });
  }, [data.gitlabProjectId, data.gitlabBranch]);

  const loadData = (selectedOptions: CascaderOptionType[] | undefined) => {
    if (!selectedOptions) {
      return;
    }
    const targetOption = selectedOptions[selectedOptions.length - 1];
    targetOption.loading = true;

    console.log(targetOption);
    switch (targetOption.type) {
      case "project":
        branches(Number(targetOption.value)).then((res) => {
          targetOption.loading = false;
          targetOption.children = res.data.data;
          setOptions([...options]);
        });
        return;
      case "branch":
        commits(
          Number(targetOption.projectId),
          String(targetOption.value)
        ).then((res) => {
          targetOption.loading = false;
          targetOption.children = res.data.data;
          setOptions([...options]);
        });
        return;
    }
  };

  const onChange = (
    values: any[],
    selectedOptions: CascaderOptionType[] | undefined
  ) => {
    let gitlabId = _.get(values, 0, 0);
    let gbranch = _.get(values, 1, "");
    let gcommit = _.get(values, 2, "");
    setData((d) => ({
      ...d,
      name: _.get(
        options.find((item) => item.value === values[0]),
        "label",
        ""
      ),
      gitlabProjectId: gitlabId,
      gitlabBranch: gbranch,
      gitlabCommit: gcommit,
    }));

    if (selectedOptions) {
      const targetOption = selectedOptions[selectedOptions.length - 1];
      if (targetOption.children) {
        targetOption.loading = true;
        targetOption.children = undefined;
        switch (targetOption.type) {
          case "project":
            branches(Number(targetOption.value)).then((res) => {
              targetOption.loading = false;
              targetOption.children = res.data.data;
              setOptions([...options]);
            });
            return;
          case "branch":
            commits(
              Number(targetOption.projectId),
              String(targetOption.value)
            ).then((res) => {
              targetOption.loading = false;
              targetOption.children = res.data.data;
              setOptions([...options]);
            });
            return;
        }
      }
    }

    if (gitlabId) {
      let o = options.find((item) => item.value === values[0]);
      setValue([o ? o.label : ""]);
      if (gbranch) {
        if (o && o.children) {
          let b = o.children.find((item) => item.value === gbranch);
          setValue([o.label, b ? b.label : ""]);

          if (gcommit) {
            if (b && b.children) {
              let c = b.children.find((item) => item.value === gcommit);
              setValue([o.label, b.label, c ? c.label : ""]);
              if (data.config === "") {
                loadConfigFile();
              }
            }
          }
        }
      }
    }
  };

  const updateDeploy = () => {
    if (data.gitlabCommit && data.gitlabBranch) {
      setEditVisible(false);
      setTimelineVisible(true);

      let re = {
        type: "update_project",
        data: JSON.stringify({
          project_id: Number(id),
          gitlab_branch: data.gitlabBranch,
          gitlab_commit: data.gitlabCommit,
          config: data.config,
        }),
      };
      let s = JSON.stringify(re);
      dispatch(clearCreateProjectLog(slug));
      dispatch(setCreateProjectLoading(slug, true));
      ws?.send(s);
    }
  };

  const onReset = () => {
    setData({
      name: detail.name,
      gitlabProjectId: Number(detail.gitlab_project_id),
      gitlabBranch: detail.gitlab_branch,
      gitlabCommit: detail.gitlab_commit,
      config: detail.config,
    });
    setValue(initValue);
  };

  return (
    <div className="edit-project">
      <PipelineInfo
        projectId={data.gitlabProjectId}
        branch={data.gitlabBranch}
        commit={data.gitlabCommit}
      />
      <div className={classNames({ "display-none": !editVisible })}>
        <div style={{ width: "100%" }}>
          {list[slug]?.output.length > 0 ? (
            <Button
              style={{ marginBottom: 20 }}
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
          {value.length > 0 ? (
            <Cascader
              options={options}
              style={{ width: "100%", marginBottom: "10px" }}
              autoFocus
              allowClear={false}
              value={value}
              loadData={loadData}
              onChange={onChange}
              changeOnSelect
              placeholder="选择项目/分支/提交"
            />
          ) : (
            <Skeleton.Input active style={{ width: 800 }} size="small" />
          )}
        </div>
        配置文件:
        <div style={{ minWidth: 200, maxWidth: 1280, marginBottom: 20 }}>
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
            style={{ padding: "0 10px" }}
            percent={list[slug]?.processPercent}
            status="active"
          />
        </div>
        <Timeline
          pending={list[slug]?.isLoading ? "loading..." : false}
          reverse={true}
          style={{ paddingLeft: 2 }}
        >
          {list[slug]?.output.map((data, index) => (
            <Timeline.Item
              key={index}
              color={data === "部署失败" ? "red" : "blue"}
            >
              {data}
            </Timeline.Item>
          ))}
        </Timeline>
      </div>

      <div className="edit-project__footer">
        <Button
          style={{ marginRight: 5 }}
          disabled={list[slug]?.isLoading}
          onClick={onReset}
        >
          重置
        </Button>
        <Button
          type="primary"
          loading={list[slug]?.isLoading}
          onClick={updateDeploy}
        >
          部署
        </Button>
      </div>
    </div>
  );
};

export default memo(ModalSub);
