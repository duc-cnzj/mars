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
