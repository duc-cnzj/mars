import React, { useState, useEffect, memo } from "react";
import { Cascader } from "antd";
import { branches, commits, Options, projects } from "../api/gitlab";
import { CascaderOptionType } from "antd/lib/cascader";
import _ from "lodash";

const ProjectSelector: React.FC<{
  value?: {
    projectName: string;
    gitlabProjectId: number;
    gitlabBranch: string;
    gitlabCommit: string;
    time?: number;
  };
  onChange: (data: {
    projectName: string;
    gitlabProjectId: number;
    gitlabBranch: string;
    gitlabCommit: string;
  }) => void;
}> = ({ value: v, onChange: onCh }) => {
  const [options, setOptions] = useState<Options[]>([]);
  let initValue: (string | number)[] = [];
  if (v) {
    initValue = [v.projectName, v.gitlabBranch, v.gitlabCommit];
  }
  const [value, setValue] = useState<(string | number)[]>(initValue);

  useEffect(() => {
    projects().then((res) => {
      if (v) {
        if (v.gitlabProjectId) {
          let r = res.data.data.find(
            (item) => item.projectId === v.gitlabProjectId
          );
          setOptions(r ? [r] : []);
        }
      } else {
        setOptions(res.data.data);
      }
    });
  }, []);

  useEffect(() => {
    if (v) {
      setValue([v.projectName, v.gitlabBranch, v.gitlabCommit]);
    }
  }, [v]);

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
    values: (string | number)[],
    selectedOptions: CascaderOptionType[] | undefined
  ) => {
    let gitlabId = _.get(values, 0, 0);
    let gbranch = _.get(values, 1, "");
    let gcommit = _.get(values, 2, "");

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
            }
          }
        }
      }
    }
    onCh({
      projectName: _.get(
        options.find((item) => item.value === values[0]),
        "label",
        ""
      ),
      gitlabProjectId: Number(gitlabId),
      gitlabBranch: String(gbranch),
      gitlabCommit: String(gcommit),
    });
  };

  return (
    <Cascader
      options={options}
      style={{ width: "100%", marginBottom: "10px" }}
      autoFocus
      value={value}
      allowClear={false}
      loadData={loadData}
      onChange={onChange}
      changeOnSelect
      placeholder="选择项目/分支/提交"
    />
  );
};

export default memo(ProjectSelector);
