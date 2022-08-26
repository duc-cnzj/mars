import React, { useState, useEffect, memo, useCallback } from "react";
import { Cascader, Skeleton } from "antd";
import { commit } from "../api/git";
import { branchOptions, commitOptions, projectOptions } from "../api/git";
import { get } from "lodash";
import pb from "../api/compiled";
import { useAsyncState } from "../utils/async";

const ProjectSelector: React.FC<{
  isCreate: boolean;
  disabled?: boolean;
  value?: {
    projectName: string;
    gitProjectId: string;
    gitBranch: string;
    gitCommit: string;
    time?: number;
  };
  onChange?: (data: {
    projectName: string;
    gitProjectId: number;
    gitBranch: string;
    gitCommit: string;
  }) => void;
}> = ({ value: v, onChange: onCh, isCreate, disabled }) => {
  const [options, setOptions] = useAsyncState<pb.git.Option[]>([]);
  const [value, setValue] = useState<(string | number)[]>([]);
  const [loading, setLoading] = useState(v ? !!v.gitCommit : false);

  const [selectedValues, setSelectedValues] = useState<(string | number)[]>([]);

  // 初始化，设置 initvalue
  useEffect(() => {
    if (value.length < 1 && v && v.gitCommit && v.gitBranch && v.gitProjectId) {
      projectOptions().then((res) => {
        let r = res.data.items.find(
          (item) => item.gitProjectId === String(v.gitProjectId)
        );

        commit({
          git_project_id: String(v.gitProjectId),
          branch: v.gitBranch,
          commit: v.gitCommit,
        }).then((res) => {
          r && res.data && setValue([r.label, v.gitBranch, res.data.label]);
          setLoading(false);
        });
      });
    }
  }, [v, value]);

  useEffect(() => {
    projectOptions().then((res) => {
      if (!isCreate && v?.gitProjectId) {
        let r = res.data.items.find(
          (item) => item.gitProjectId === String(v.gitProjectId)
        );
        if (r) {
          (r as any).children = [];
        }
        setOptions(r ? [r] : []);
      } else {
        setOptions(
          res.data.items.map((i: any) => {
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
            git_project_id: String(targetOption.value),
            all: false,
          }).then((res) => {
            targetOption.loading = false;
            targetOption.children = res.data.items;
            setOptions((opts) => [...opts]);
          });
          return;
        case "branch":
          commitOptions({
            git_project_id: String(targetOption.gitProjectId),
            branch: String(targetOption.value),
          }).then((res) => {
            targetOption.loading = false;
            targetOption.children = res.data.items;
            setOptions((opts) => [...opts]);
          });
          return;
      }
    },
    [setOptions]
  );

  const onChange = (values: (string | number)[]) => {
    setSelectedValues(values);
    let gitId = get(values, 0, 0);
    let gbranch = get(values, 1, "");
    let gcommit = get(values, 2, "");

    if (gitId) {
      let o = options.find((item) => item.value === values[0]);
      setValue([o ? o.label : ""]);
      if (gbranch) {
        // @ts-ignore
        if (o && o.children) {
          // @ts-ignore
          let b = o.children.find(
            (item: pb.git.Option) => item.value === gbranch
          );
          setValue([o.label, b ? b.label : ""]);
          if (gcommit) {
            if (b && b.children) {
              let c = b.children.find(
                (item: pb.git.Option) => item.value === gcommit
              );
              setValue([o.label, b.label, c ? c.label : ""]);
              onCh?.({
                projectName: get(
                  options.find((item) => item.value === values[0]),
                  "display_name",
                  ""
                ),
                gitProjectId: Number(gitId),
                gitBranch: String(gbranch),
                gitCommit: String(gcommit),
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
        onDropdownVisibleChange={() => {
          if (selectedValues && selectedValues.length !== 3) {
            setSelectedValues([])
          }
        }}
        disabled={disabled}
        showSearch={{
          filter: (inputValue: string, options: any, fieldNames: any) => {
            return options.some((option: any) => {
              switch (selectedValues.length) {
                case 0:
                  if (option.type !== "project") {
                    return false;
                  }
                  option.children = []
                  break;
                case 1:
                  if (option.type !== "branch") {
                    return false;
                  }
                  if (String(option.gitProjectId) !== String(selectedValues[0])) {
                    return false;
                  }
                  break;
                case 2:
                  if (option.type !== "commit") {
                    return false;
                  }
                  if (String(option.gitProjectId) !== String(selectedValues[0])) {
                    return false;
                  }
                  break;
              }

              return (
                (option.label as string)
                  .toLowerCase()
                  .indexOf(inputValue.toLowerCase()) > -1
              );
            });
          },
        }}
        options={options}
        style={{ width: "100%" }}
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
