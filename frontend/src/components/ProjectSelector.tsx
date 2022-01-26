import React, { useState, useEffect, memo, useCallback } from "react";
import { Cascader, Skeleton } from "antd";
import { commit } from "../api/gitlab";
import { branchOptions, commitOptions, projectOptions } from "../api/gitlab";
import { get } from "lodash";
import pb from "../api/compiled";
import { useAsyncState } from "../utils/async";

const ProjectSelector: React.FC<{
  isCreate: boolean;
  value?: {
    projectName: string;
    gitlabProjectId: string;
    gitlabBranch: string;
    gitlabCommit: string;
    time?: number;
  };
  onChange?: (data: {
    projectName: string;
    gitlabProjectId: number;
    gitlabBranch: string;
    gitlabCommit: string;
  }) => void;
}> = ({ value: v, onChange: onCh, isCreate }) => {
  const [options, setOptions] = useAsyncState<pb.GitOption[]>([]);
  const [value, setValue] = useState<(string | number)[]>([]);
  const [loading, setLoading] = useState(v ? !!v.gitlabCommit : false);

  // 初始化，设置 initvalue
  useEffect(() => {
    if (
      value.length < 1 &&
      v &&
      v.gitlabCommit &&
      v.gitlabBranch &&
      v.gitlabProjectId
    ) {
      projectOptions().then((res) => {
        let r = res.data.data.find(
          (item) => item.projectId === String(v.gitlabProjectId)
        );

        commit({
          project_id: String(v.gitlabProjectId),
          branch: v.gitlabBranch,
          commit: v.gitlabCommit,
        }).then((res) => {
          r &&
            res.data.data &&
            setValue([r.label, v.gitlabBranch, res.data.data.label]);
          setLoading(false);
        });
      });
    }
  }, [v, value]);

  useEffect(() => {
    projectOptions().then((res) => {
      if (!isCreate && v?.gitlabProjectId) {
        let r = res.data.data.find(
          (item) => item.projectId === String(v.gitlabProjectId)
        );
        if (r) {
          (r as any).children = [];
        }
        setOptions(r ? [r] : []);
      } else {
        setOptions(
          res.data.data.map((i: any) => {
            i.children = [];
            return i;
          })
        );
      }
    });
  }, [v, isCreate, setOptions]);

  const loadData = useCallback(
    (selectedOptions: any | undefined) => {
      if (!selectedOptions) {
        return;
      }
      const targetOption = selectedOptions[selectedOptions.length - 1];
      targetOption.loading = true;
      targetOption.children = undefined;

      switch (targetOption.type) {
        case "project":
          branchOptions({
            project_id: String(targetOption.value),
            all: false,
          }).then((res) => {
            targetOption.loading = false;
            targetOption.children = res.data.data;
            setOptions((opts) => [...opts]);
          });
          return;
        case "branch":
          commitOptions({
            project_id: String(targetOption.projectId),
            branch: String(targetOption.value),
          }).then((res) => {
            targetOption.loading = false;
            targetOption.children = res.data.data;
            setOptions((opts) => [...opts]);
          });
          return;
      }
    },
    [setOptions]
  );

  const onChange = (values: (string | number)[]) => {
    let gitlabId = get(values, 0, 0);
    let gbranch = get(values, 1, "");
    let gcommit = get(values, 2, "");

    if (gitlabId) {
      let o = options.find((item) => item.value === values[0]);
      setValue([o ? o.label : ""]);
      if (gbranch) {
        // @ts-ignore
        if (o && o.children) {
          // @ts-ignore
          let b = o.children.find(
            (item: pb.GitOption) => item.value === gbranch
          );
          setValue([o.label, b ? b.label : ""]);
          if (gcommit) {
            if (b && b.children) {
              let c = b.children.find(
                (item: pb.GitOption) => item.value === gcommit
              );
              setValue([o.label, b.label, c ? c.label : ""]);
              onCh?.({
                projectName: get(
                  options.find((item) => item.value === values[0]),
                  "label",
                  ""
                ),
                gitlabProjectId: Number(gitlabId),
                gitlabBranch: String(gbranch),
                gitlabCommit: String(gcommit),
              });
            }
          }
        }
      }
    }
  };

  return (
    <Skeleton
      active
      paragraph={false}
      avatar={false}
      loading={loading}
      title={{ style: { marginTop: 0, height: 24 } }}
    >
      <Cascader
        options={options}
        style={{ width: "100%" }}
        autoFocus
        value={value}
        allowClear={false}
        loadData={loadData}
        onChange={onChange}
        changeOnSelect
        placeholder="选择项目/分支/提交"
      />
    </Skeleton>
  );
};

export default memo(ProjectSelector);
