declare namespace API {
  interface Mars {
    config_file: string;
    config_file_type: string;
    docker_repository: string;
    docker_tag_format: string;
    local_chart_path: string;
    config_field: string;
    is_simple_env: boolean;
    default_values: null | { [k: string]: any };
    branches: null | string[];
    ingress_overwrite_values: null | string;
  }

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

  interface WsResponse {
    type: string;
    slug: string;
    result: string;
    data: string;
    end: boolean;
  }

  interface WsHandleExecShellResponse {
    data: string;
    namespace: string;
    pod: string;
    container: string;
    session_id: string;
  }
}
