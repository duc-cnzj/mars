declare namespace Mars {
  interface CreateItemInterface {
    gitlabProjectId: number;
    gitlabBranch: string;
    gitlabCommit: string;

    name: string;
    config: string;
    config_type: string;
    debug: boolean;
  }
}
