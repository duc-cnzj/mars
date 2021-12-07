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
declare namespace API {
  interface ClusterInfo {
    status: string;

    free_memory: string;
    free_cpu: string;

    total_memory: string;
    total_cpu: string;

    usage_memory_rate: string;
    usage_cpu_rate: string;

    request_memory_rate: string;
    request_cpu_rate: string;

    free_request_memory: string;
    free_request_cpu: string;
  }
}
