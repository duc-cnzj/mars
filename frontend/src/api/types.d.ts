

declare namespace Mars {
  import pb from '../api/compiled';
  interface CreateItemInterface {
    gitlabProjectId: number;
    gitlabBranch: string;
    gitlabCommit: string;

    name: string;
    config: string;
    config_type: string;
    debug: boolean;
    extra_values: pb.ProjectExtraItem[];
  }
}
