import React, { useState, memo, useCallback } from "react";
import { Col, Row, Select, SelectProps } from "antd";
import { branchOptions, commitOptions, projectOptions } from "../api/git";
import pb from "../api/compiled";
import { omitEqual } from "../utils/obj";
import { css } from "@emotion/css";
import styled from "@emotion/styled";

const ProjectSelector: React.FC<{
  isCreate: boolean;
  disabled?: boolean;
  value?: {
    projectName: string;
    gitProjectId: number;
    gitBranch: string;
    gitCommit: string;
    gitCommitTitle: string;
    time?: number;
  };
  onChange?: (data: {
    projectName: string;
    gitProjectId: number;
    gitBranch: string;
    gitCommit: string;
    gitCommitTitle: string;
  }) => void;
}> = ({ value: v, onChange: onCh, isCreate, disabled }) => {
  console.log("ProjectSelector render");

  const [options, setOptions] = useState<{
    projects: pb.git.Option[];
    branches: pb.git.Option[];
    commits: pb.git.Option[];
  }>({ projects: [], branches: [], commits: [] });
  const [loading, setLoading] = useState({
    project: false,
    branch: false,
    commit: false,
  });
  const [focusIdx, setFocusIdx] = useState<number | null>(null);

  const onProjectVisibleChange = useCallback((open: boolean) => {
    if (!open) {
      return;
    }
    setLoading((l) => ({ ...l, project: true }));
    projectOptions()
      .then(({ data }) =>
        setOptions({ projects: data.items, branches: [], commits: [] })
      )
      .finally(() => setLoading((l) => ({ ...l, project: false })));
  }, []);

  const onProjectChange = useCallback(
    (value: any) => {
      let pid = Number(value);
      let currentBranch =
        v && Number(v.gitProjectId) === Number(value) ? v?.gitBranch : "";
      let currentCommit = currentBranch === v?.gitBranch ? v.gitCommit : "";
      let commitTitle = currentCommit === v?.gitCommit ? v?.gitCommitTitle : "";
      onCh?.({
        projectName:
          options.projects.find((pro) => Number(pro.gitProjectId) === pid)
            ?.display_name || "",
        gitProjectId: pid,
        gitBranch: currentBranch,
        gitCommit: currentCommit,
        gitCommitTitle: commitTitle,
      });
    },
    [onCh, options.projects, v]
  );

  const onBranchVisibleChange = useCallback(
    (open: boolean) => {
      if (!v?.gitProjectId || !open) {
        return;
      }
      setLoading((l) => ({ ...l, branch: true }));
      branchOptions({
        git_project_id: String(v?.gitProjectId),
        all: false,
      })
        .then(({ data }) =>
          setOptions((opts) => ({ ...opts, branches: data.items, commits: [] }))
        )
        .finally(() => setLoading((l) => ({ ...l, branch: false })));
    },
    [v?.gitProjectId]
  );

  const onBranchChange = useCallback(
    (vv: any) => {
      onCh?.({
        projectName: v?.projectName || "",
        gitProjectId: v?.gitProjectId || 0,
        gitBranch: String(vv),
        gitCommit: "",
        gitCommitTitle: "",
      });
    },
    [onCh, v?.gitProjectId, v?.projectName]
  );

  const onCommitClickVisibleChange = useCallback(
    (open: boolean) => {
      if (!v?.gitProjectId || !v?.gitBranch || !open) {
        return;
      }
      setLoading((l) => ({ ...l, commit: true }));
      commitOptions({
        git_project_id: String(v?.gitProjectId),
        branch: String(v?.gitBranch),
      })
        .then(({ data }) => {
          setOptions((opts) => ({ ...opts, commits: data.items }));
        })
        .finally(() => {
          setLoading((l) => ({ ...l, commit: false }));
        });
    },
    [v?.gitProjectId, v?.gitBranch]
  );

  const onCommitChange = useCallback(
    (vv: any) => {
      onCh?.({
        projectName: v?.projectName || "",
        gitProjectId: v?.gitProjectId || 0,
        gitBranch: v?.gitBranch || "",
        gitCommit: String(vv),
        gitCommitTitle:
          options.commits.find((it) => it.value === String(vv))?.label || "",
      });
    },
    [onCh, options.commits, v?.gitBranch, v?.gitProjectId, v?.projectName]
  );

  return (
    <Row>
      <MyCol
        span={6}
        onFocus={() => setFocusIdx(1)}
        onBlur={() => setFocusIdx(null)}
        focus={focusIdx === 1 ? 1 : 0}
      >
        <SelectorItem
          loading={loading.project}
          className={css`
            .ant-select-selector {
              border-top-right-radius: 0 !important;
              border-bottom-right-radius: 0 !important;
            }
          `}
          placeholder="选择项目"
          disabled={disabled || loading.branch || loading.commit}
          value={v?.projectName}
          onDropdownVisibleChange={onProjectVisibleChange}
          onChange={onProjectChange}
          options={
            isCreate
              ? options.projects
              : options.projects.filter(
                  (p) => String(p.gitProjectId) === String(v?.gitProjectId)
                )
          }
        />
      </MyCol>
      <MyCol
        span={6}
        onFocus={() => setFocusIdx(2)}
        onBlur={() => setFocusIdx(null)}
        focus={focusIdx === 2 ? 1 : 0}
      >
        <SelectorItem
          className={css`
            .ant-select-selector {
              border-radius: 0 !important;
            }
          `}
          loading={loading.branch}
          onDropdownVisibleChange={onBranchVisibleChange}
          placeholder="选择分支"
          disabled={disabled || loading.commit}
          value={v?.gitBranch}
          onChange={onBranchChange}
          options={options.branches}
        />
      </MyCol>
      <MyCol
        span={12}
        onFocus={() => setFocusIdx(3)}
        onBlur={() => setFocusIdx(null)}
        focus={focusIdx === 3 ? 1 : 0}
      >
        <SelectorItem
          className={css`
            .ant-select-selector {
              border-top-left-radius: 0 !important;
              border-bottom-left-radius: 0 !important;
            }
          `}
          loading={loading.commit}
          onDropdownVisibleChange={onCommitClickVisibleChange}
          placeholder="选择 Commit"
          disabled={disabled}
          value={v?.gitCommitTitle}
          onChange={onCommitChange}
          options={options.commits}
        />
      </MyCol>
    </Row>
  );
};

const SelectorItem: React.FC<
  {
    className?: string;
    value: any;
    onChange: (v: any) => void;
    options: pb.git.Option[];
    disabled?: boolean;
    placeholder: string;
  } & SelectProps
> = memo(
  ({ className, value, onChange, options, disabled, placeholder, ...rest }) => {
    return (
      <Select
        className={className}
        showSearch
        disabled={disabled}
        placeholder={placeholder}
        value={value === "" ? null : value}
        defaultActiveFirstOption={false}
        optionFilterProp="label"
        onChange={onChange}
        options={options}
        {...rest}
      />
    );
  }
);

export default memo(ProjectSelector, (prev, next) =>
  omitEqual(prev, next, "onChange")
);

const MyCol = styled(Col)<{ focus: number }>`
  margin-right: -1px;
  &:hover {
    z-index: 100;
  }
  ${(p) =>
    p.focus &&
    `
    z-index: 100;
  `}
`;
