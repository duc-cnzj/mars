import * as $protobuf from "protobufjs";
/** ClusterStatus enum. */
export enum ClusterStatus {
    StatusUnknown = 0,
    StatusBad = 1,
    StatusNotGood = 2,
    StatusHealth = 3
}

/** Represents a ClusterInfoResponse. */
export class ClusterInfoResponse implements IClusterInfoResponse {

    /**
     * Constructs a new ClusterInfoResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IClusterInfoResponse);

    /** ClusterInfoResponse status. */
    public status: string;

    /** ClusterInfoResponse free_memory. */
    public free_memory: string;

    /** ClusterInfoResponse free_cpu. */
    public free_cpu: string;

    /** ClusterInfoResponse free_request_memory. */
    public free_request_memory: string;

    /** ClusterInfoResponse free_request_cpu. */
    public free_request_cpu: string;

    /** ClusterInfoResponse total_memory. */
    public total_memory: string;

    /** ClusterInfoResponse total_cpu. */
    public total_cpu: string;

    /** ClusterInfoResponse usage_memory_rate. */
    public usage_memory_rate: string;

    /** ClusterInfoResponse usage_cpu_rate. */
    public usage_cpu_rate: string;

    /** ClusterInfoResponse request_memory_rate. */
    public request_memory_rate: string;

    /** ClusterInfoResponse request_cpu_rate. */
    public request_cpu_rate: string;
}

/** Represents a Cluster */
export class Cluster extends $protobuf.rpc.Service {

    /**
     * Constructs a new Cluster service.
     * @param rpcImpl RPC implementation
     * @param [requestDelimited=false] Whether requests are length-delimited
     * @param [responseDelimited=false] Whether responses are length-delimited
     */
    constructor(rpcImpl: $protobuf.RPCImpl, requestDelimited?: boolean, responseDelimited?: boolean);

    /**
     * Calls Info.
     * @param request Empty message or plain object
     * @param callback Node-style callback called with the error, if any, and ClusterInfoResponse
     */
    public info(request: google.protobuf.IEmpty, callback: Cluster.InfoCallback): void;

    /**
     * Calls Info.
     * @param request Empty message or plain object
     * @returns Promise
     */
    public info(request: google.protobuf.IEmpty): Promise<ClusterInfoResponse>;
}

export namespace Cluster {

    /**
     * Callback as used by {@link Cluster#info}.
     * @param error Error, if any
     * @param [response] ClusterInfoResponse
     */
    type InfoCallback = (error: (Error|null), response?: ClusterInfoResponse) => void;
}

/** Represents a GitlabDestroyRequest. */
export class GitlabDestroyRequest implements IGitlabDestroyRequest {

    /**
     * Constructs a new GitlabDestroyRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IGitlabDestroyRequest);

    /** GitlabDestroyRequest namespace_id. */
    public namespace_id: string;

    /** GitlabDestroyRequest project_id. */
    public project_id: string;
}

/** Represents an EnableProjectRequest. */
export class EnableProjectRequest implements IEnableProjectRequest {

    /**
     * Constructs a new EnableProjectRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IEnableProjectRequest);

    /** EnableProjectRequest gitlab_project_id. */
    public gitlab_project_id: string;
}

/** Represents a DisableProjectRequest. */
export class DisableProjectRequest implements IDisableProjectRequest {

    /**
     * Constructs a new DisableProjectRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IDisableProjectRequest);

    /** DisableProjectRequest gitlab_project_id. */
    public gitlab_project_id: string;
}

/** Represents a GitlabProjectInfo. */
export class GitlabProjectInfo implements IGitlabProjectInfo {

    /**
     * Constructs a new GitlabProjectInfo.
     * @param [properties] Properties to set
     */
    constructor(properties?: IGitlabProjectInfo);

    /** GitlabProjectInfo id. */
    public id: (number|Long);

    /** GitlabProjectInfo name. */
    public name: string;

    /** GitlabProjectInfo path. */
    public path: string;

    /** GitlabProjectInfo web_url. */
    public web_url: string;

    /** GitlabProjectInfo avatar_url. */
    public avatar_url: string;

    /** GitlabProjectInfo description. */
    public description: string;

    /** GitlabProjectInfo enabled. */
    public enabled: boolean;

    /** GitlabProjectInfo global_enabled. */
    public global_enabled: boolean;
}

/** Represents a ProjectListResponse. */
export class ProjectListResponse implements IProjectListResponse {

    /**
     * Constructs a new ProjectListResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IProjectListResponse);

    /** ProjectListResponse data. */
    public data: IGitlabProjectInfo[];
}

/** Represents an Option. */
export class Option implements IOption {

    /**
     * Constructs a new Option.
     * @param [properties] Properties to set
     */
    constructor(properties?: IOption);

    /** Option value. */
    public value: string;

    /** Option label. */
    public label: string;

    /** Option type. */
    public type: string;

    /** Option isLeaf. */
    public isLeaf: boolean;

    /** Option projectId. */
    public projectId: string;

    /** Option branch. */
    public branch: string;
}

/** Represents a ProjectsResponse. */
export class ProjectsResponse implements IProjectsResponse {

    /**
     * Constructs a new ProjectsResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IProjectsResponse);

    /** ProjectsResponse data. */
    public data: IOption[];
}

/** Represents a BranchesRequest. */
export class BranchesRequest implements IBranchesRequest {

    /**
     * Constructs a new BranchesRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IBranchesRequest);

    /** BranchesRequest project_id. */
    public project_id: string;
}

/** Represents a BranchesResponse. */
export class BranchesResponse implements IBranchesResponse {

    /**
     * Constructs a new BranchesResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IBranchesResponse);

    /** BranchesResponse data. */
    public data: IOption[];
}

/** Represents a CommitsRequest. */
export class CommitsRequest implements ICommitsRequest {

    /**
     * Constructs a new CommitsRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: ICommitsRequest);

    /** CommitsRequest project_id. */
    public project_id: string;

    /** CommitsRequest branch. */
    public branch: string;
}

/** Represents a CommitsResponse. */
export class CommitsResponse implements ICommitsResponse {

    /**
     * Constructs a new CommitsResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: ICommitsResponse);

    /** CommitsResponse data. */
    public data: IOption[];
}

/** Represents a CommitRequest. */
export class CommitRequest implements ICommitRequest {

    /**
     * Constructs a new CommitRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: ICommitRequest);

    /** CommitRequest project_id. */
    public project_id: string;

    /** CommitRequest branch. */
    public branch: string;

    /** CommitRequest commit. */
    public commit: string;
}

/** Represents a CommitResponse. */
export class CommitResponse implements ICommitResponse {

    /**
     * Constructs a new CommitResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: ICommitResponse);

    /** CommitResponse data. */
    public data?: (IOption|null);
}

/** Represents a PipelineInfoRequest. */
export class PipelineInfoRequest implements IPipelineInfoRequest {

    /**
     * Constructs a new PipelineInfoRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IPipelineInfoRequest);

    /** PipelineInfoRequest project_id. */
    public project_id: string;

    /** PipelineInfoRequest branch. */
    public branch: string;

    /** PipelineInfoRequest commit. */
    public commit: string;
}

/** Represents a PipelineInfoResponse. */
export class PipelineInfoResponse implements IPipelineInfoResponse {

    /**
     * Constructs a new PipelineInfoResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IPipelineInfoResponse);

    /** PipelineInfoResponse status. */
    public status: string;

    /** PipelineInfoResponse web_url. */
    public web_url: string;
}

/** Represents a ConfigFileRequest. */
export class ConfigFileRequest implements IConfigFileRequest {

    /**
     * Constructs a new ConfigFileRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IConfigFileRequest);

    /** ConfigFileRequest project_id. */
    public project_id: string;

    /** ConfigFileRequest branch. */
    public branch: string;
}

/** Represents a ConfigFileResponse. */
export class ConfigFileResponse implements IConfigFileResponse {

    /**
     * Constructs a new ConfigFileResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IConfigFileResponse);

    /** ConfigFileResponse data. */
    public data: string;

    /** ConfigFileResponse type. */
    public type: string;
}

/** Represents a Gitlab */
export class Gitlab extends $protobuf.rpc.Service {

    /**
     * Constructs a new Gitlab service.
     * @param rpcImpl RPC implementation
     * @param [requestDelimited=false] Whether requests are length-delimited
     * @param [responseDelimited=false] Whether responses are length-delimited
     */
    constructor(rpcImpl: $protobuf.RPCImpl, requestDelimited?: boolean, responseDelimited?: boolean);

    /**
     * Calls EnableProject.
     * @param request EnableProjectRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and Empty
     */
    public enableProject(request: IEnableProjectRequest, callback: Gitlab.EnableProjectCallback): void;

    /**
     * Calls EnableProject.
     * @param request EnableProjectRequest message or plain object
     * @returns Promise
     */
    public enableProject(request: IEnableProjectRequest): Promise<google.protobuf.Empty>;

    /**
     * Calls DisableProject.
     * @param request DisableProjectRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and Empty
     */
    public disableProject(request: IDisableProjectRequest, callback: Gitlab.DisableProjectCallback): void;

    /**
     * Calls DisableProject.
     * @param request DisableProjectRequest message or plain object
     * @returns Promise
     */
    public disableProject(request: IDisableProjectRequest): Promise<google.protobuf.Empty>;

    /**
     * Calls ProjectList.
     * @param request Empty message or plain object
     * @param callback Node-style callback called with the error, if any, and ProjectListResponse
     */
    public projectList(request: google.protobuf.IEmpty, callback: Gitlab.ProjectListCallback): void;

    /**
     * Calls ProjectList.
     * @param request Empty message or plain object
     * @returns Promise
     */
    public projectList(request: google.protobuf.IEmpty): Promise<ProjectListResponse>;

    /**
     * Calls Projects.
     * @param request Empty message or plain object
     * @param callback Node-style callback called with the error, if any, and ProjectsResponse
     */
    public projects(request: google.protobuf.IEmpty, callback: Gitlab.ProjectsCallback): void;

    /**
     * Calls Projects.
     * @param request Empty message or plain object
     * @returns Promise
     */
    public projects(request: google.protobuf.IEmpty): Promise<ProjectsResponse>;

    /**
     * Calls Branches.
     * @param request BranchesRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and BranchesResponse
     */
    public branches(request: IBranchesRequest, callback: Gitlab.BranchesCallback): void;

    /**
     * Calls Branches.
     * @param request BranchesRequest message or plain object
     * @returns Promise
     */
    public branches(request: IBranchesRequest): Promise<BranchesResponse>;

    /**
     * Calls Commits.
     * @param request CommitsRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and CommitsResponse
     */
    public commits(request: ICommitsRequest, callback: Gitlab.CommitsCallback): void;

    /**
     * Calls Commits.
     * @param request CommitsRequest message or plain object
     * @returns Promise
     */
    public commits(request: ICommitsRequest): Promise<CommitsResponse>;

    /**
     * Calls Commit.
     * @param request CommitRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and CommitResponse
     */
    public commit(request: ICommitRequest, callback: Gitlab.CommitCallback): void;

    /**
     * Calls Commit.
     * @param request CommitRequest message or plain object
     * @returns Promise
     */
    public commit(request: ICommitRequest): Promise<CommitResponse>;

    /**
     * Calls PipelineInfo.
     * @param request PipelineInfoRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and PipelineInfoResponse
     */
    public pipelineInfo(request: IPipelineInfoRequest, callback: Gitlab.PipelineInfoCallback): void;

    /**
     * Calls PipelineInfo.
     * @param request PipelineInfoRequest message or plain object
     * @returns Promise
     */
    public pipelineInfo(request: IPipelineInfoRequest): Promise<PipelineInfoResponse>;

    /**
     * Calls ConfigFile.
     * @param request ConfigFileRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and ConfigFileResponse
     */
    public configFile(request: IConfigFileRequest, callback: Gitlab.ConfigFileCallback): void;

    /**
     * Calls ConfigFile.
     * @param request ConfigFileRequest message or plain object
     * @returns Promise
     */
    public configFile(request: IConfigFileRequest): Promise<ConfigFileResponse>;
}

export namespace Gitlab {

    /**
     * Callback as used by {@link Gitlab#enableProject}.
     * @param error Error, if any
     * @param [response] Empty
     */
    type EnableProjectCallback = (error: (Error|null), response?: google.protobuf.Empty) => void;

    /**
     * Callback as used by {@link Gitlab#disableProject}.
     * @param error Error, if any
     * @param [response] Empty
     */
    type DisableProjectCallback = (error: (Error|null), response?: google.protobuf.Empty) => void;

    /**
     * Callback as used by {@link Gitlab#projectList}.
     * @param error Error, if any
     * @param [response] ProjectListResponse
     */
    type ProjectListCallback = (error: (Error|null), response?: ProjectListResponse) => void;

    /**
     * Callback as used by {@link Gitlab#projects}.
     * @param error Error, if any
     * @param [response] ProjectsResponse
     */
    type ProjectsCallback = (error: (Error|null), response?: ProjectsResponse) => void;

    /**
     * Callback as used by {@link Gitlab#branches}.
     * @param error Error, if any
     * @param [response] BranchesResponse
     */
    type BranchesCallback = (error: (Error|null), response?: BranchesResponse) => void;

    /**
     * Callback as used by {@link Gitlab#commits}.
     * @param error Error, if any
     * @param [response] CommitsResponse
     */
    type CommitsCallback = (error: (Error|null), response?: CommitsResponse) => void;

    /**
     * Callback as used by {@link Gitlab#commit}.
     * @param error Error, if any
     * @param [response] CommitResponse
     */
    type CommitCallback = (error: (Error|null), response?: CommitResponse) => void;

    /**
     * Callback as used by {@link Gitlab#pipelineInfo}.
     * @param error Error, if any
     * @param [response] PipelineInfoResponse
     */
    type PipelineInfoCallback = (error: (Error|null), response?: PipelineInfoResponse) => void;

    /**
     * Callback as used by {@link Gitlab#configFile}.
     * @param error Error, if any
     * @param [response] ConfigFileResponse
     */
    type ConfigFileCallback = (error: (Error|null), response?: ConfigFileResponse) => void;
}

/** Represents a MarsShowRequest. */
export class MarsShowRequest implements IMarsShowRequest {

    /**
     * Constructs a new MarsShowRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IMarsShowRequest);

    /** MarsShowRequest project_id. */
    public project_id: (number|Long);

    /** MarsShowRequest branch. */
    public branch: string;
}

/** Represents a MarsShowResponse. */
export class MarsShowResponse implements IMarsShowResponse {

    /**
     * Constructs a new MarsShowResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IMarsShowResponse);

    /** MarsShowResponse branch. */
    public branch: string;

    /** MarsShowResponse config. */
    public config: string;
}

/** Represents a GlobalConfigRequest. */
export class GlobalConfigRequest implements IGlobalConfigRequest {

    /**
     * Constructs a new GlobalConfigRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IGlobalConfigRequest);

    /** GlobalConfigRequest project_id. */
    public project_id: (number|Long);
}

/** Represents a GlobalConfigResponse. */
export class GlobalConfigResponse implements IGlobalConfigResponse {

    /**
     * Constructs a new GlobalConfigResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IGlobalConfigResponse);

    /** GlobalConfigResponse enabled. */
    public enabled: boolean;

    /** GlobalConfigResponse config. */
    public config: string;
}

/** Represents a MarsUpdateRequest. */
export class MarsUpdateRequest implements IMarsUpdateRequest {

    /**
     * Constructs a new MarsUpdateRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IMarsUpdateRequest);

    /** MarsUpdateRequest project_id. */
    public project_id: (number|Long);

    /** MarsUpdateRequest config. */
    public config: string;
}

/** Represents a MarsUpdateResponse. */
export class MarsUpdateResponse implements IMarsUpdateResponse {

    /**
     * Constructs a new MarsUpdateResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IMarsUpdateResponse);

    /** MarsUpdateResponse data. */
    public data?: (IGitlabProjectModal|null);
}

/** Represents a ToggleEnabledRequest. */
export class ToggleEnabledRequest implements IToggleEnabledRequest {

    /**
     * Constructs a new ToggleEnabledRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IToggleEnabledRequest);

    /** ToggleEnabledRequest project_id. */
    public project_id: (number|Long);

    /** ToggleEnabledRequest enabled. */
    public enabled: boolean;
}

/** Represents a Mars */
export class Mars extends $protobuf.rpc.Service {

    /**
     * Constructs a new Mars service.
     * @param rpcImpl RPC implementation
     * @param [requestDelimited=false] Whether requests are length-delimited
     * @param [responseDelimited=false] Whether responses are length-delimited
     */
    constructor(rpcImpl: $protobuf.RPCImpl, requestDelimited?: boolean, responseDelimited?: boolean);

    /**
     * Calls Show.
     * @param request MarsShowRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and MarsShowResponse
     */
    public show(request: IMarsShowRequest, callback: Mars.ShowCallback): void;

    /**
     * Calls Show.
     * @param request MarsShowRequest message or plain object
     * @returns Promise
     */
    public show(request: IMarsShowRequest): Promise<MarsShowResponse>;

    /**
     * Calls GlobalConfig.
     * @param request GlobalConfigRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and GlobalConfigResponse
     */
    public globalConfig(request: IGlobalConfigRequest, callback: Mars.GlobalConfigCallback): void;

    /**
     * Calls GlobalConfig.
     * @param request GlobalConfigRequest message or plain object
     * @returns Promise
     */
    public globalConfig(request: IGlobalConfigRequest): Promise<GlobalConfigResponse>;

    /**
     * Calls ToggleEnabled.
     * @param request ToggleEnabledRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and Empty
     */
    public toggleEnabled(request: IToggleEnabledRequest, callback: Mars.ToggleEnabledCallback): void;

    /**
     * Calls ToggleEnabled.
     * @param request ToggleEnabledRequest message or plain object
     * @returns Promise
     */
    public toggleEnabled(request: IToggleEnabledRequest): Promise<google.protobuf.Empty>;

    /**
     * Calls Update.
     * @param request MarsUpdateRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and MarsUpdateResponse
     */
    public update(request: IMarsUpdateRequest, callback: Mars.UpdateCallback): void;

    /**
     * Calls Update.
     * @param request MarsUpdateRequest message or plain object
     * @returns Promise
     */
    public update(request: IMarsUpdateRequest): Promise<MarsUpdateResponse>;
}

export namespace Mars {

    /**
     * Callback as used by {@link Mars#show}.
     * @param error Error, if any
     * @param [response] MarsShowResponse
     */
    type ShowCallback = (error: (Error|null), response?: MarsShowResponse) => void;

    /**
     * Callback as used by {@link Mars#globalConfig}.
     * @param error Error, if any
     * @param [response] GlobalConfigResponse
     */
    type GlobalConfigCallback = (error: (Error|null), response?: GlobalConfigResponse) => void;

    /**
     * Callback as used by {@link Mars#toggleEnabled}.
     * @param error Error, if any
     * @param [response] Empty
     */
    type ToggleEnabledCallback = (error: (Error|null), response?: google.protobuf.Empty) => void;

    /**
     * Callback as used by {@link Mars#update}.
     * @param error Error, if any
     * @param [response] MarsUpdateResponse
     */
    type UpdateCallback = (error: (Error|null), response?: MarsUpdateResponse) => void;
}

/** Represents a GitlabProjectModal. */
export class GitlabProjectModal implements IGitlabProjectModal {

    /**
     * Constructs a new GitlabProjectModal.
     * @param [properties] Properties to set
     */
    constructor(properties?: IGitlabProjectModal);

    /** GitlabProjectModal id. */
    public id: (number|Long);

    /** GitlabProjectModal default_branch. */
    public default_branch: string;

    /** GitlabProjectModal name. */
    public name: string;

    /** GitlabProjectModal gitlab_project_id. */
    public gitlab_project_id: (number|Long);

    /** GitlabProjectModal enabled. */
    public enabled: boolean;

    /** GitlabProjectModal global_enabled. */
    public global_enabled: boolean;

    /** GitlabProjectModal global_config. */
    public global_config: string;

    /** GitlabProjectModal created_at. */
    public created_at?: (google.protobuf.ITimestamp|null);

    /** GitlabProjectModal updated_at. */
    public updated_at?: (google.protobuf.ITimestamp|null);

    /** GitlabProjectModal deleted_at. */
    public deleted_at?: (google.protobuf.ITimestamp|null);
}

/** Represents a NamespaceModal. */
export class NamespaceModal implements INamespaceModal {

    /**
     * Constructs a new NamespaceModal.
     * @param [properties] Properties to set
     */
    constructor(properties?: INamespaceModal);

    /** NamespaceModal id. */
    public id: (number|Long);

    /** NamespaceModal name. */
    public name: string;

    /** NamespaceModal image_pull_secrets. */
    public image_pull_secrets: string[];

    /** NamespaceModal created_at. */
    public created_at?: (google.protobuf.ITimestamp|null);

    /** NamespaceModal updated_at. */
    public updated_at?: (google.protobuf.ITimestamp|null);

    /** NamespaceModal deleted_at. */
    public deleted_at?: (google.protobuf.ITimestamp|null);

    /** NamespaceModal projects. */
    public projects: IProjectModal[];
}

/** Represents a ProjectModal. */
export class ProjectModal implements IProjectModal {

    /**
     * Constructs a new ProjectModal.
     * @param [properties] Properties to set
     */
    constructor(properties?: IProjectModal);

    /** ProjectModal id. */
    public id: (number|Long);

    /** ProjectModal name. */
    public name: string;

    /** ProjectModal gitlab_project_id. */
    public gitlab_project_id: (number|Long);

    /** ProjectModal gitlab_branch. */
    public gitlab_branch: string;

    /** ProjectModal gitlab_commit. */
    public gitlab_commit: string;

    /** ProjectModal config. */
    public config: string;

    /** ProjectModal override_values. */
    public override_values: string;

    /** ProjectModal docker_image. */
    public docker_image: string;

    /** ProjectModal pod_selectors. */
    public pod_selectors: string;

    /** ProjectModal namespace_id. */
    public namespace_id: (number|Long);

    /** ProjectModal atomic. */
    public atomic: boolean;

    /** ProjectModal created_at. */
    public created_at?: (google.protobuf.ITimestamp|null);

    /** ProjectModal updated_at. */
    public updated_at?: (google.protobuf.ITimestamp|null);

    /** ProjectModal deleted_at. */
    public deleted_at?: (google.protobuf.ITimestamp|null);

    /** ProjectModal namespace. */
    public namespace?: (INamespaceModal|null);
}

/** Namespace google. */
export namespace google {

    /** Namespace protobuf. */
    namespace protobuf {

        /** Properties of an Empty. */
        interface IEmpty {
        }

        /** Represents an Empty. */
        class Empty implements IEmpty {

            /**
             * Constructs a new Empty.
             * @param [properties] Properties to set
             */
            constructor(properties?: google.protobuf.IEmpty);
        }

        /** Properties of a FileDescriptorSet. */
        interface IFileDescriptorSet {

            /** FileDescriptorSet file */
            file?: (google.protobuf.IFileDescriptorProto[]|null);
        }

        /** Represents a FileDescriptorSet. */
        class FileDescriptorSet implements IFileDescriptorSet {

            /**
             * Constructs a new FileDescriptorSet.
             * @param [properties] Properties to set
             */
            constructor(properties?: google.protobuf.IFileDescriptorSet);

            /** FileDescriptorSet file. */
            public file: google.protobuf.IFileDescriptorProto[];
        }

        /** Properties of a FileDescriptorProto. */
        interface IFileDescriptorProto {

            /** FileDescriptorProto name */
            name?: (string|null);

            /** FileDescriptorProto package */
            "package"?: (string|null);

            /** FileDescriptorProto dependency */
            dependency?: (string[]|null);

            /** FileDescriptorProto public_dependency */
            public_dependency?: (number[]|null);

            /** FileDescriptorProto weak_dependency */
            weak_dependency?: (number[]|null);

            /** FileDescriptorProto message_type */
            message_type?: (google.protobuf.IDescriptorProto[]|null);

            /** FileDescriptorProto enum_type */
            enum_type?: (google.protobuf.IEnumDescriptorProto[]|null);

            /** FileDescriptorProto service */
            service?: (google.protobuf.IServiceDescriptorProto[]|null);

            /** FileDescriptorProto extension */
            extension?: (google.protobuf.IFieldDescriptorProto[]|null);

            /** FileDescriptorProto options */
            options?: (google.protobuf.IFileOptions|null);

            /** FileDescriptorProto source_code_info */
            source_code_info?: (google.protobuf.ISourceCodeInfo|null);

            /** FileDescriptorProto syntax */
            syntax?: (string|null);
        }

        /** Represents a FileDescriptorProto. */
        class FileDescriptorProto implements IFileDescriptorProto {

            /**
             * Constructs a new FileDescriptorProto.
             * @param [properties] Properties to set
             */
            constructor(properties?: google.protobuf.IFileDescriptorProto);

            /** FileDescriptorProto name. */
            public name: string;

            /** FileDescriptorProto package. */
            public package: string;

            /** FileDescriptorProto dependency. */
            public dependency: string[];

            /** FileDescriptorProto public_dependency. */
            public public_dependency: number[];

            /** FileDescriptorProto weak_dependency. */
            public weak_dependency: number[];

            /** FileDescriptorProto message_type. */
            public message_type: google.protobuf.IDescriptorProto[];

            /** FileDescriptorProto enum_type. */
            public enum_type: google.protobuf.IEnumDescriptorProto[];

            /** FileDescriptorProto service. */
            public service: google.protobuf.IServiceDescriptorProto[];

            /** FileDescriptorProto extension. */
            public extension: google.protobuf.IFieldDescriptorProto[];

            /** FileDescriptorProto options. */
            public options?: (google.protobuf.IFileOptions|null);

            /** FileDescriptorProto source_code_info. */
            public source_code_info?: (google.protobuf.ISourceCodeInfo|null);

            /** FileDescriptorProto syntax. */
            public syntax: string;
        }

        /** Properties of a DescriptorProto. */
        interface IDescriptorProto {

            /** DescriptorProto name */
            name?: (string|null);

            /** DescriptorProto field */
            field?: (google.protobuf.IFieldDescriptorProto[]|null);

            /** DescriptorProto extension */
            extension?: (google.protobuf.IFieldDescriptorProto[]|null);

            /** DescriptorProto nested_type */
            nested_type?: (google.protobuf.IDescriptorProto[]|null);

            /** DescriptorProto enum_type */
            enum_type?: (google.protobuf.IEnumDescriptorProto[]|null);

            /** DescriptorProto extension_range */
            extension_range?: (google.protobuf.DescriptorProto.IExtensionRange[]|null);

            /** DescriptorProto oneof_decl */
            oneof_decl?: (google.protobuf.IOneofDescriptorProto[]|null);

            /** DescriptorProto options */
            options?: (google.protobuf.IMessageOptions|null);

            /** DescriptorProto reserved_range */
            reserved_range?: (google.protobuf.DescriptorProto.IReservedRange[]|null);

            /** DescriptorProto reserved_name */
            reserved_name?: (string[]|null);
        }

        /** Represents a DescriptorProto. */
        class DescriptorProto implements IDescriptorProto {

            /**
             * Constructs a new DescriptorProto.
             * @param [properties] Properties to set
             */
            constructor(properties?: google.protobuf.IDescriptorProto);

            /** DescriptorProto name. */
            public name: string;

            /** DescriptorProto field. */
            public field: google.protobuf.IFieldDescriptorProto[];

            /** DescriptorProto extension. */
            public extension: google.protobuf.IFieldDescriptorProto[];

            /** DescriptorProto nested_type. */
            public nested_type: google.protobuf.IDescriptorProto[];

            /** DescriptorProto enum_type. */
            public enum_type: google.protobuf.IEnumDescriptorProto[];

            /** DescriptorProto extension_range. */
            public extension_range: google.protobuf.DescriptorProto.IExtensionRange[];

            /** DescriptorProto oneof_decl. */
            public oneof_decl: google.protobuf.IOneofDescriptorProto[];

            /** DescriptorProto options. */
            public options?: (google.protobuf.IMessageOptions|null);

            /** DescriptorProto reserved_range. */
            public reserved_range: google.protobuf.DescriptorProto.IReservedRange[];

            /** DescriptorProto reserved_name. */
            public reserved_name: string[];
        }

        namespace DescriptorProto {

            /** Properties of an ExtensionRange. */
            interface IExtensionRange {

                /** ExtensionRange start */
                start?: (number|null);

                /** ExtensionRange end */
                end?: (number|null);
            }

            /** Represents an ExtensionRange. */
            class ExtensionRange implements IExtensionRange {

                /**
                 * Constructs a new ExtensionRange.
                 * @param [properties] Properties to set
                 */
                constructor(properties?: google.protobuf.DescriptorProto.IExtensionRange);

                /** ExtensionRange start. */
                public start: number;

                /** ExtensionRange end. */
                public end: number;
            }

            /** Properties of a ReservedRange. */
            interface IReservedRange {

                /** ReservedRange start */
                start?: (number|null);

                /** ReservedRange end */
                end?: (number|null);
            }

            /** Represents a ReservedRange. */
            class ReservedRange implements IReservedRange {

                /**
                 * Constructs a new ReservedRange.
                 * @param [properties] Properties to set
                 */
                constructor(properties?: google.protobuf.DescriptorProto.IReservedRange);

                /** ReservedRange start. */
                public start: number;

                /** ReservedRange end. */
                public end: number;
            }
        }

        /** Properties of a FieldDescriptorProto. */
        interface IFieldDescriptorProto {

            /** FieldDescriptorProto name */
            name?: (string|null);

            /** FieldDescriptorProto number */
            number?: (number|null);

            /** FieldDescriptorProto label */
            label?: (google.protobuf.FieldDescriptorProto.Label|null);

            /** FieldDescriptorProto type */
            type?: (google.protobuf.FieldDescriptorProto.Type|null);

            /** FieldDescriptorProto type_name */
            type_name?: (string|null);

            /** FieldDescriptorProto extendee */
            extendee?: (string|null);

            /** FieldDescriptorProto default_value */
            default_value?: (string|null);

            /** FieldDescriptorProto oneof_index */
            oneof_index?: (number|null);

            /** FieldDescriptorProto json_name */
            json_name?: (string|null);

            /** FieldDescriptorProto options */
            options?: (google.protobuf.IFieldOptions|null);
        }

        /** Represents a FieldDescriptorProto. */
        class FieldDescriptorProto implements IFieldDescriptorProto {

            /**
             * Constructs a new FieldDescriptorProto.
             * @param [properties] Properties to set
             */
            constructor(properties?: google.protobuf.IFieldDescriptorProto);

            /** FieldDescriptorProto name. */
            public name: string;

            /** FieldDescriptorProto number. */
            public number: number;

            /** FieldDescriptorProto label. */
            public label: google.protobuf.FieldDescriptorProto.Label;

            /** FieldDescriptorProto type. */
            public type: google.protobuf.FieldDescriptorProto.Type;

            /** FieldDescriptorProto type_name. */
            public type_name: string;

            /** FieldDescriptorProto extendee. */
            public extendee: string;

            /** FieldDescriptorProto default_value. */
            public default_value: string;

            /** FieldDescriptorProto oneof_index. */
            public oneof_index: number;

            /** FieldDescriptorProto json_name. */
            public json_name: string;

            /** FieldDescriptorProto options. */
            public options?: (google.protobuf.IFieldOptions|null);
        }

        namespace FieldDescriptorProto {

            /** Type enum. */
            enum Type {
                TYPE_DOUBLE = 1,
                TYPE_FLOAT = 2,
                TYPE_INT64 = 3,
                TYPE_UINT64 = 4,
                TYPE_INT32 = 5,
                TYPE_FIXED64 = 6,
                TYPE_FIXED32 = 7,
                TYPE_BOOL = 8,
                TYPE_STRING = 9,
                TYPE_GROUP = 10,
                TYPE_MESSAGE = 11,
                TYPE_BYTES = 12,
                TYPE_UINT32 = 13,
                TYPE_ENUM = 14,
                TYPE_SFIXED32 = 15,
                TYPE_SFIXED64 = 16,
                TYPE_SINT32 = 17,
                TYPE_SINT64 = 18
            }

            /** Label enum. */
            enum Label {
                LABEL_OPTIONAL = 1,
                LABEL_REQUIRED = 2,
                LABEL_REPEATED = 3
            }
        }

        /** Properties of an OneofDescriptorProto. */
        interface IOneofDescriptorProto {

            /** OneofDescriptorProto name */
            name?: (string|null);

            /** OneofDescriptorProto options */
            options?: (google.protobuf.IOneofOptions|null);
        }

        /** Represents an OneofDescriptorProto. */
        class OneofDescriptorProto implements IOneofDescriptorProto {

            /**
             * Constructs a new OneofDescriptorProto.
             * @param [properties] Properties to set
             */
            constructor(properties?: google.protobuf.IOneofDescriptorProto);

            /** OneofDescriptorProto name. */
            public name: string;

            /** OneofDescriptorProto options. */
            public options?: (google.protobuf.IOneofOptions|null);
        }

        /** Properties of an EnumDescriptorProto. */
        interface IEnumDescriptorProto {

            /** EnumDescriptorProto name */
            name?: (string|null);

            /** EnumDescriptorProto value */
            value?: (google.protobuf.IEnumValueDescriptorProto[]|null);

            /** EnumDescriptorProto options */
            options?: (google.protobuf.IEnumOptions|null);
        }

        /** Represents an EnumDescriptorProto. */
        class EnumDescriptorProto implements IEnumDescriptorProto {

            /**
             * Constructs a new EnumDescriptorProto.
             * @param [properties] Properties to set
             */
            constructor(properties?: google.protobuf.IEnumDescriptorProto);

            /** EnumDescriptorProto name. */
            public name: string;

            /** EnumDescriptorProto value. */
            public value: google.protobuf.IEnumValueDescriptorProto[];

            /** EnumDescriptorProto options. */
            public options?: (google.protobuf.IEnumOptions|null);
        }

        /** Properties of an EnumValueDescriptorProto. */
        interface IEnumValueDescriptorProto {

            /** EnumValueDescriptorProto name */
            name?: (string|null);

            /** EnumValueDescriptorProto number */
            number?: (number|null);

            /** EnumValueDescriptorProto options */
            options?: (google.protobuf.IEnumValueOptions|null);
        }

        /** Represents an EnumValueDescriptorProto. */
        class EnumValueDescriptorProto implements IEnumValueDescriptorProto {

            /**
             * Constructs a new EnumValueDescriptorProto.
             * @param [properties] Properties to set
             */
            constructor(properties?: google.protobuf.IEnumValueDescriptorProto);

            /** EnumValueDescriptorProto name. */
            public name: string;

            /** EnumValueDescriptorProto number. */
            public number: number;

            /** EnumValueDescriptorProto options. */
            public options?: (google.protobuf.IEnumValueOptions|null);
        }

        /** Properties of a ServiceDescriptorProto. */
        interface IServiceDescriptorProto {

            /** ServiceDescriptorProto name */
            name?: (string|null);

            /** ServiceDescriptorProto method */
            method?: (google.protobuf.IMethodDescriptorProto[]|null);

            /** ServiceDescriptorProto options */
            options?: (google.protobuf.IServiceOptions|null);
        }

        /** Represents a ServiceDescriptorProto. */
        class ServiceDescriptorProto implements IServiceDescriptorProto {

            /**
             * Constructs a new ServiceDescriptorProto.
             * @param [properties] Properties to set
             */
            constructor(properties?: google.protobuf.IServiceDescriptorProto);

            /** ServiceDescriptorProto name. */
            public name: string;

            /** ServiceDescriptorProto method. */
            public method: google.protobuf.IMethodDescriptorProto[];

            /** ServiceDescriptorProto options. */
            public options?: (google.protobuf.IServiceOptions|null);
        }

        /** Properties of a MethodDescriptorProto. */
        interface IMethodDescriptorProto {

            /** MethodDescriptorProto name */
            name?: (string|null);

            /** MethodDescriptorProto input_type */
            input_type?: (string|null);

            /** MethodDescriptorProto output_type */
            output_type?: (string|null);

            /** MethodDescriptorProto options */
            options?: (google.protobuf.IMethodOptions|null);

            /** MethodDescriptorProto client_streaming */
            client_streaming?: (boolean|null);

            /** MethodDescriptorProto server_streaming */
            server_streaming?: (boolean|null);
        }

        /** Represents a MethodDescriptorProto. */
        class MethodDescriptorProto implements IMethodDescriptorProto {

            /**
             * Constructs a new MethodDescriptorProto.
             * @param [properties] Properties to set
             */
            constructor(properties?: google.protobuf.IMethodDescriptorProto);

            /** MethodDescriptorProto name. */
            public name: string;

            /** MethodDescriptorProto input_type. */
            public input_type: string;

            /** MethodDescriptorProto output_type. */
            public output_type: string;

            /** MethodDescriptorProto options. */
            public options?: (google.protobuf.IMethodOptions|null);

            /** MethodDescriptorProto client_streaming. */
            public client_streaming: boolean;

            /** MethodDescriptorProto server_streaming. */
            public server_streaming: boolean;
        }

        /** Properties of a FileOptions. */
        interface IFileOptions {

            /** FileOptions java_package */
            java_package?: (string|null);

            /** FileOptions java_outer_classname */
            java_outer_classname?: (string|null);

            /** FileOptions java_multiple_files */
            java_multiple_files?: (boolean|null);

            /** FileOptions java_generate_equals_and_hash */
            java_generate_equals_and_hash?: (boolean|null);

            /** FileOptions java_string_check_utf8 */
            java_string_check_utf8?: (boolean|null);

            /** FileOptions optimize_for */
            optimize_for?: (google.protobuf.FileOptions.OptimizeMode|null);

            /** FileOptions go_package */
            go_package?: (string|null);

            /** FileOptions cc_generic_services */
            cc_generic_services?: (boolean|null);

            /** FileOptions java_generic_services */
            java_generic_services?: (boolean|null);

            /** FileOptions py_generic_services */
            py_generic_services?: (boolean|null);

            /** FileOptions deprecated */
            deprecated?: (boolean|null);

            /** FileOptions cc_enable_arenas */
            cc_enable_arenas?: (boolean|null);

            /** FileOptions objc_class_prefix */
            objc_class_prefix?: (string|null);

            /** FileOptions csharp_namespace */
            csharp_namespace?: (string|null);

            /** FileOptions uninterpreted_option */
            uninterpreted_option?: (google.protobuf.IUninterpretedOption[]|null);
        }

        /** Represents a FileOptions. */
        class FileOptions implements IFileOptions {

            /**
             * Constructs a new FileOptions.
             * @param [properties] Properties to set
             */
            constructor(properties?: google.protobuf.IFileOptions);

            /** FileOptions java_package. */
            public java_package: string;

            /** FileOptions java_outer_classname. */
            public java_outer_classname: string;

            /** FileOptions java_multiple_files. */
            public java_multiple_files: boolean;

            /** FileOptions java_generate_equals_and_hash. */
            public java_generate_equals_and_hash: boolean;

            /** FileOptions java_string_check_utf8. */
            public java_string_check_utf8: boolean;

            /** FileOptions optimize_for. */
            public optimize_for: google.protobuf.FileOptions.OptimizeMode;

            /** FileOptions go_package. */
            public go_package: string;

            /** FileOptions cc_generic_services. */
            public cc_generic_services: boolean;

            /** FileOptions java_generic_services. */
            public java_generic_services: boolean;

            /** FileOptions py_generic_services. */
            public py_generic_services: boolean;

            /** FileOptions deprecated. */
            public deprecated: boolean;

            /** FileOptions cc_enable_arenas. */
            public cc_enable_arenas: boolean;

            /** FileOptions objc_class_prefix. */
            public objc_class_prefix: string;

            /** FileOptions csharp_namespace. */
            public csharp_namespace: string;

            /** FileOptions uninterpreted_option. */
            public uninterpreted_option: google.protobuf.IUninterpretedOption[];
        }

        namespace FileOptions {

            /** OptimizeMode enum. */
            enum OptimizeMode {
                SPEED = 1,
                CODE_SIZE = 2,
                LITE_RUNTIME = 3
            }
        }

        /** Properties of a MessageOptions. */
        interface IMessageOptions {

            /** MessageOptions message_set_wire_format */
            message_set_wire_format?: (boolean|null);

            /** MessageOptions no_standard_descriptor_accessor */
            no_standard_descriptor_accessor?: (boolean|null);

            /** MessageOptions deprecated */
            deprecated?: (boolean|null);

            /** MessageOptions map_entry */
            map_entry?: (boolean|null);

            /** MessageOptions uninterpreted_option */
            uninterpreted_option?: (google.protobuf.IUninterpretedOption[]|null);
        }

        /** Represents a MessageOptions. */
        class MessageOptions implements IMessageOptions {

            /**
             * Constructs a new MessageOptions.
             * @param [properties] Properties to set
             */
            constructor(properties?: google.protobuf.IMessageOptions);

            /** MessageOptions message_set_wire_format. */
            public message_set_wire_format: boolean;

            /** MessageOptions no_standard_descriptor_accessor. */
            public no_standard_descriptor_accessor: boolean;

            /** MessageOptions deprecated. */
            public deprecated: boolean;

            /** MessageOptions map_entry. */
            public map_entry: boolean;

            /** MessageOptions uninterpreted_option. */
            public uninterpreted_option: google.protobuf.IUninterpretedOption[];
        }

        /** Properties of a FieldOptions. */
        interface IFieldOptions {

            /** FieldOptions ctype */
            ctype?: (google.protobuf.FieldOptions.CType|null);

            /** FieldOptions packed */
            packed?: (boolean|null);

            /** FieldOptions jstype */
            jstype?: (google.protobuf.FieldOptions.JSType|null);

            /** FieldOptions lazy */
            lazy?: (boolean|null);

            /** FieldOptions deprecated */
            deprecated?: (boolean|null);

            /** FieldOptions weak */
            weak?: (boolean|null);

            /** FieldOptions uninterpreted_option */
            uninterpreted_option?: (google.protobuf.IUninterpretedOption[]|null);
        }

        /** Represents a FieldOptions. */
        class FieldOptions implements IFieldOptions {

            /**
             * Constructs a new FieldOptions.
             * @param [properties] Properties to set
             */
            constructor(properties?: google.protobuf.IFieldOptions);

            /** FieldOptions ctype. */
            public ctype: google.protobuf.FieldOptions.CType;

            /** FieldOptions packed. */
            public packed: boolean;

            /** FieldOptions jstype. */
            public jstype: google.protobuf.FieldOptions.JSType;

            /** FieldOptions lazy. */
            public lazy: boolean;

            /** FieldOptions deprecated. */
            public deprecated: boolean;

            /** FieldOptions weak. */
            public weak: boolean;

            /** FieldOptions uninterpreted_option. */
            public uninterpreted_option: google.protobuf.IUninterpretedOption[];
        }

        namespace FieldOptions {

            /** CType enum. */
            enum CType {
                STRING = 0,
                CORD = 1,
                STRING_PIECE = 2
            }

            /** JSType enum. */
            enum JSType {
                JS_NORMAL = 0,
                JS_STRING = 1,
                JS_NUMBER = 2
            }
        }

        /** Properties of an OneofOptions. */
        interface IOneofOptions {

            /** OneofOptions uninterpreted_option */
            uninterpreted_option?: (google.protobuf.IUninterpretedOption[]|null);
        }

        /** Represents an OneofOptions. */
        class OneofOptions implements IOneofOptions {

            /**
             * Constructs a new OneofOptions.
             * @param [properties] Properties to set
             */
            constructor(properties?: google.protobuf.IOneofOptions);

            /** OneofOptions uninterpreted_option. */
            public uninterpreted_option: google.protobuf.IUninterpretedOption[];
        }

        /** Properties of an EnumOptions. */
        interface IEnumOptions {

            /** EnumOptions allow_alias */
            allow_alias?: (boolean|null);

            /** EnumOptions deprecated */
            deprecated?: (boolean|null);

            /** EnumOptions uninterpreted_option */
            uninterpreted_option?: (google.protobuf.IUninterpretedOption[]|null);
        }

        /** Represents an EnumOptions. */
        class EnumOptions implements IEnumOptions {

            /**
             * Constructs a new EnumOptions.
             * @param [properties] Properties to set
             */
            constructor(properties?: google.protobuf.IEnumOptions);

            /** EnumOptions allow_alias. */
            public allow_alias: boolean;

            /** EnumOptions deprecated. */
            public deprecated: boolean;

            /** EnumOptions uninterpreted_option. */
            public uninterpreted_option: google.protobuf.IUninterpretedOption[];
        }

        /** Properties of an EnumValueOptions. */
        interface IEnumValueOptions {

            /** EnumValueOptions deprecated */
            deprecated?: (boolean|null);

            /** EnumValueOptions uninterpreted_option */
            uninterpreted_option?: (google.protobuf.IUninterpretedOption[]|null);
        }

        /** Represents an EnumValueOptions. */
        class EnumValueOptions implements IEnumValueOptions {

            /**
             * Constructs a new EnumValueOptions.
             * @param [properties] Properties to set
             */
            constructor(properties?: google.protobuf.IEnumValueOptions);

            /** EnumValueOptions deprecated. */
            public deprecated: boolean;

            /** EnumValueOptions uninterpreted_option. */
            public uninterpreted_option: google.protobuf.IUninterpretedOption[];
        }

        /** Properties of a ServiceOptions. */
        interface IServiceOptions {

            /** ServiceOptions deprecated */
            deprecated?: (boolean|null);

            /** ServiceOptions uninterpreted_option */
            uninterpreted_option?: (google.protobuf.IUninterpretedOption[]|null);
        }

        /** Represents a ServiceOptions. */
        class ServiceOptions implements IServiceOptions {

            /**
             * Constructs a new ServiceOptions.
             * @param [properties] Properties to set
             */
            constructor(properties?: google.protobuf.IServiceOptions);

            /** ServiceOptions deprecated. */
            public deprecated: boolean;

            /** ServiceOptions uninterpreted_option. */
            public uninterpreted_option: google.protobuf.IUninterpretedOption[];
        }

        /** Properties of a MethodOptions. */
        interface IMethodOptions {

            /** MethodOptions deprecated */
            deprecated?: (boolean|null);

            /** MethodOptions uninterpreted_option */
            uninterpreted_option?: (google.protobuf.IUninterpretedOption[]|null);

            /** MethodOptions .google.api.http */
            ".google.api.http"?: (google.api.IHttpRule|null);
        }

        /** Represents a MethodOptions. */
        class MethodOptions implements IMethodOptions {

            /**
             * Constructs a new MethodOptions.
             * @param [properties] Properties to set
             */
            constructor(properties?: google.protobuf.IMethodOptions);

            /** MethodOptions deprecated. */
            public deprecated: boolean;

            /** MethodOptions uninterpreted_option. */
            public uninterpreted_option: google.protobuf.IUninterpretedOption[];
        }

        /** Properties of an UninterpretedOption. */
        interface IUninterpretedOption {

            /** UninterpretedOption name */
            name?: (google.protobuf.UninterpretedOption.INamePart[]|null);

            /** UninterpretedOption identifier_value */
            identifier_value?: (string|null);

            /** UninterpretedOption positive_int_value */
            positive_int_value?: (number|Long|null);

            /** UninterpretedOption negative_int_value */
            negative_int_value?: (number|Long|null);

            /** UninterpretedOption double_value */
            double_value?: (number|null);

            /** UninterpretedOption string_value */
            string_value?: (Uint8Array|null);

            /** UninterpretedOption aggregate_value */
            aggregate_value?: (string|null);
        }

        /** Represents an UninterpretedOption. */
        class UninterpretedOption implements IUninterpretedOption {

            /**
             * Constructs a new UninterpretedOption.
             * @param [properties] Properties to set
             */
            constructor(properties?: google.protobuf.IUninterpretedOption);

            /** UninterpretedOption name. */
            public name: google.protobuf.UninterpretedOption.INamePart[];

            /** UninterpretedOption identifier_value. */
            public identifier_value: string;

            /** UninterpretedOption positive_int_value. */
            public positive_int_value: (number|Long);

            /** UninterpretedOption negative_int_value. */
            public negative_int_value: (number|Long);

            /** UninterpretedOption double_value. */
            public double_value: number;

            /** UninterpretedOption string_value. */
            public string_value: Uint8Array;

            /** UninterpretedOption aggregate_value. */
            public aggregate_value: string;
        }

        namespace UninterpretedOption {

            /** Properties of a NamePart. */
            interface INamePart {

                /** NamePart name_part */
                name_part: string;

                /** NamePart is_extension */
                is_extension: boolean;
            }

            /** Represents a NamePart. */
            class NamePart implements INamePart {

                /**
                 * Constructs a new NamePart.
                 * @param [properties] Properties to set
                 */
                constructor(properties?: google.protobuf.UninterpretedOption.INamePart);

                /** NamePart name_part. */
                public name_part: string;

                /** NamePart is_extension. */
                public is_extension: boolean;
            }
        }

        /** Properties of a SourceCodeInfo. */
        interface ISourceCodeInfo {

            /** SourceCodeInfo location */
            location?: (google.protobuf.SourceCodeInfo.ILocation[]|null);
        }

        /** Represents a SourceCodeInfo. */
        class SourceCodeInfo implements ISourceCodeInfo {

            /**
             * Constructs a new SourceCodeInfo.
             * @param [properties] Properties to set
             */
            constructor(properties?: google.protobuf.ISourceCodeInfo);

            /** SourceCodeInfo location. */
            public location: google.protobuf.SourceCodeInfo.ILocation[];
        }

        namespace SourceCodeInfo {

            /** Properties of a Location. */
            interface ILocation {

                /** Location path */
                path?: (number[]|null);

                /** Location span */
                span?: (number[]|null);

                /** Location leading_comments */
                leading_comments?: (string|null);

                /** Location trailing_comments */
                trailing_comments?: (string|null);

                /** Location leading_detached_comments */
                leading_detached_comments?: (string[]|null);
            }

            /** Represents a Location. */
            class Location implements ILocation {

                /**
                 * Constructs a new Location.
                 * @param [properties] Properties to set
                 */
                constructor(properties?: google.protobuf.SourceCodeInfo.ILocation);

                /** Location path. */
                public path: number[];

                /** Location span. */
                public span: number[];

                /** Location leading_comments. */
                public leading_comments: string;

                /** Location trailing_comments. */
                public trailing_comments: string;

                /** Location leading_detached_comments. */
                public leading_detached_comments: string[];
            }
        }

        /** Properties of a GeneratedCodeInfo. */
        interface IGeneratedCodeInfo {

            /** GeneratedCodeInfo annotation */
            annotation?: (google.protobuf.GeneratedCodeInfo.IAnnotation[]|null);
        }

        /** Represents a GeneratedCodeInfo. */
        class GeneratedCodeInfo implements IGeneratedCodeInfo {

            /**
             * Constructs a new GeneratedCodeInfo.
             * @param [properties] Properties to set
             */
            constructor(properties?: google.protobuf.IGeneratedCodeInfo);

            /** GeneratedCodeInfo annotation. */
            public annotation: google.protobuf.GeneratedCodeInfo.IAnnotation[];
        }

        namespace GeneratedCodeInfo {

            /** Properties of an Annotation. */
            interface IAnnotation {

                /** Annotation path */
                path?: (number[]|null);

                /** Annotation source_file */
                source_file?: (string|null);

                /** Annotation begin */
                begin?: (number|null);

                /** Annotation end */
                end?: (number|null);
            }

            /** Represents an Annotation. */
            class Annotation implements IAnnotation {

                /**
                 * Constructs a new Annotation.
                 * @param [properties] Properties to set
                 */
                constructor(properties?: google.protobuf.GeneratedCodeInfo.IAnnotation);

                /** Annotation path. */
                public path: number[];

                /** Annotation source_file. */
                public source_file: string;

                /** Annotation begin. */
                public begin: number;

                /** Annotation end. */
                public end: number;
            }
        }

        /** Properties of a Timestamp. */
        interface ITimestamp {

            /** Timestamp seconds */
            seconds?: (number|Long|null);

            /** Timestamp nanos */
            nanos?: (number|null);
        }

        /** Represents a Timestamp. */
        class Timestamp implements ITimestamp {

            /**
             * Constructs a new Timestamp.
             * @param [properties] Properties to set
             */
            constructor(properties?: google.protobuf.ITimestamp);

            /** Timestamp seconds. */
            public seconds: (number|Long);

            /** Timestamp nanos. */
            public nanos: number;
        }
    }

    /** Namespace api. */
    namespace api {

        /** Properties of a Http. */
        interface IHttp {

            /** Http rules */
            rules?: (google.api.IHttpRule[]|null);
        }

        /** Represents a Http. */
        class Http implements IHttp {

            /**
             * Constructs a new Http.
             * @param [properties] Properties to set
             */
            constructor(properties?: google.api.IHttp);

            /** Http rules. */
            public rules: google.api.IHttpRule[];
        }

        /** Properties of a HttpRule. */
        interface IHttpRule {

            /** HttpRule get */
            get?: (string|null);

            /** HttpRule put */
            put?: (string|null);

            /** HttpRule post */
            post?: (string|null);

            /** HttpRule delete */
            "delete"?: (string|null);

            /** HttpRule patch */
            patch?: (string|null);

            /** HttpRule custom */
            custom?: (google.api.ICustomHttpPattern|null);

            /** HttpRule selector */
            selector?: (string|null);

            /** HttpRule body */
            body?: (string|null);

            /** HttpRule additional_bindings */
            additional_bindings?: (google.api.IHttpRule[]|null);
        }

        /** Represents a HttpRule. */
        class HttpRule implements IHttpRule {

            /**
             * Constructs a new HttpRule.
             * @param [properties] Properties to set
             */
            constructor(properties?: google.api.IHttpRule);

            /** HttpRule get. */
            public get?: (string|null);

            /** HttpRule put. */
            public put?: (string|null);

            /** HttpRule post. */
            public post?: (string|null);

            /** HttpRule delete. */
            public delete?: (string|null);

            /** HttpRule patch. */
            public patch?: (string|null);

            /** HttpRule custom. */
            public custom?: (google.api.ICustomHttpPattern|null);

            /** HttpRule selector. */
            public selector: string;

            /** HttpRule body. */
            public body: string;

            /** HttpRule additional_bindings. */
            public additional_bindings: google.api.IHttpRule[];

            /** HttpRule pattern. */
            public pattern?: ("get"|"put"|"post"|"delete"|"patch"|"custom");
        }

        /** Properties of a CustomHttpPattern. */
        interface ICustomHttpPattern {

            /** CustomHttpPattern kind */
            kind?: (string|null);

            /** CustomHttpPattern path */
            path?: (string|null);
        }

        /** Represents a CustomHttpPattern. */
        class CustomHttpPattern implements ICustomHttpPattern {

            /**
             * Constructs a new CustomHttpPattern.
             * @param [properties] Properties to set
             */
            constructor(properties?: google.api.ICustomHttpPattern);

            /** CustomHttpPattern kind. */
            public kind: string;

            /** CustomHttpPattern path. */
            public path: string;
        }
    }
}

/** Represents a NamespaceID. */
export class NamespaceID implements INamespaceID {

    /**
     * Constructs a new NamespaceID.
     * @param [properties] Properties to set
     */
    constructor(properties?: INamespaceID);

    /** NamespaceID namespace_id. */
    public namespace_id: (number|Long);
}

/** Represents a NamespaceResponse. */
export class NamespaceResponse implements INamespaceResponse {

    /**
     * Constructs a new NamespaceResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: INamespaceResponse);

    /** NamespaceResponse id. */
    public id: (number|Long);

    /** NamespaceResponse name. */
    public name: string;

    /** NamespaceResponse image_pull_secrets. */
    public image_pull_secrets: string[];

    /** NamespaceResponse created_at. */
    public created_at?: (google.protobuf.ITimestamp|null);

    /** NamespaceResponse updated_at. */
    public updated_at?: (google.protobuf.ITimestamp|null);

    /** NamespaceResponse deleted_at. */
    public deleted_at?: (google.protobuf.ITimestamp|null);
}

/** Represents a NamespaceItem. */
export class NamespaceItem implements INamespaceItem {

    /**
     * Constructs a new NamespaceItem.
     * @param [properties] Properties to set
     */
    constructor(properties?: INamespaceItem);

    /** NamespaceItem id. */
    public id: (number|Long);

    /** NamespaceItem name. */
    public name: string;

    /** NamespaceItem created_at. */
    public created_at?: (google.protobuf.ITimestamp|null);

    /** NamespaceItem updated_at. */
    public updated_at?: (google.protobuf.ITimestamp|null);

    /** NamespaceItem projects. */
    public projects: NamespaceItem.ISimpleProjectItem[];
}

export namespace NamespaceItem {

    /** Properties of a SimpleProjectItem. */
    interface ISimpleProjectItem {

        /** SimpleProjectItem id */
        id?: (number|Long|null);

        /** SimpleProjectItem name */
        name?: (string|null);

        /** SimpleProjectItem status */
        status?: (string|null);
    }

    /** Represents a SimpleProjectItem. */
    class SimpleProjectItem implements ISimpleProjectItem {

        /**
         * Constructs a new SimpleProjectItem.
         * @param [properties] Properties to set
         */
        constructor(properties?: NamespaceItem.ISimpleProjectItem);

        /** SimpleProjectItem id. */
        public id: (number|Long);

        /** SimpleProjectItem name. */
        public name: string;

        /** SimpleProjectItem status. */
        public status: string;
    }
}

/** Represents a NamespaceList. */
export class NamespaceList implements INamespaceList {

    /**
     * Constructs a new NamespaceList.
     * @param [properties] Properties to set
     */
    constructor(properties?: INamespaceList);

    /** NamespaceList data. */
    public data: INamespaceItem[];
}

/** Represents a NsStoreRequest. */
export class NsStoreRequest implements INsStoreRequest {

    /**
     * Constructs a new NsStoreRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: INsStoreRequest);

    /** NsStoreRequest namespace. */
    public namespace: string;
}

/** Represents a NsStoreResponse. */
export class NsStoreResponse implements INsStoreResponse {

    /**
     * Constructs a new NsStoreResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: INsStoreResponse);

    /** NsStoreResponse data. */
    public data?: (INamespaceResponse|null);
}

/** Represents a CpuAndMemoryResponse. */
export class CpuAndMemoryResponse implements ICpuAndMemoryResponse {

    /**
     * Constructs a new CpuAndMemoryResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: ICpuAndMemoryResponse);

    /** CpuAndMemoryResponse cpu. */
    public cpu: string;

    /** CpuAndMemoryResponse memory. */
    public memory: string;
}

/** Represents a ServiceEndpointsResponse. */
export class ServiceEndpointsResponse implements IServiceEndpointsResponse {

    /**
     * Constructs a new ServiceEndpointsResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IServiceEndpointsResponse);

    /** ServiceEndpointsResponse data. */
    public data: { [k: string]: ServiceEndpointsResponse.Iitem };
}

export namespace ServiceEndpointsResponse {

    /** Properties of an item. */
    interface Iitem {

        /** item name */
        name?: (string[]|null);
    }

    /** Represents an item. */
    class item implements Iitem {

        /**
         * Constructs a new item.
         * @param [properties] Properties to set
         */
        constructor(properties?: ServiceEndpointsResponse.Iitem);

        /** item name. */
        public name: string[];
    }
}

/** Represents a ServiceEndpointsRequest. */
export class ServiceEndpointsRequest implements IServiceEndpointsRequest {

    /**
     * Constructs a new ServiceEndpointsRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IServiceEndpointsRequest);

    /** ServiceEndpointsRequest namespace_id. */
    public namespace_id: (number|Long);

    /** ServiceEndpointsRequest project_name. */
    public project_name: string;
}

/** Represents a Namespace */
export class Namespace extends $protobuf.rpc.Service {

    /**
     * Constructs a new Namespace service.
     * @param rpcImpl RPC implementation
     * @param [requestDelimited=false] Whether requests are length-delimited
     * @param [responseDelimited=false] Whether responses are length-delimited
     */
    constructor(rpcImpl: $protobuf.RPCImpl, requestDelimited?: boolean, responseDelimited?: boolean);

    /**
     * Calls Index.
     * @param request Empty message or plain object
     * @param callback Node-style callback called with the error, if any, and NamespaceList
     */
    public index(request: google.protobuf.IEmpty, callback: Namespace.IndexCallback): void;

    /**
     * Calls Index.
     * @param request Empty message or plain object
     * @returns Promise
     */
    public index(request: google.protobuf.IEmpty): Promise<NamespaceList>;

    /**
     * Calls Store.
     * @param request NsStoreRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and NsStoreResponse
     */
    public store(request: INsStoreRequest, callback: Namespace.StoreCallback): void;

    /**
     * Calls Store.
     * @param request NsStoreRequest message or plain object
     * @returns Promise
     */
    public store(request: INsStoreRequest): Promise<NsStoreResponse>;

    /**
     * Calls CpuAndMemory.
     * @param request NamespaceID message or plain object
     * @param callback Node-style callback called with the error, if any, and CpuAndMemoryResponse
     */
    public cpuAndMemory(request: INamespaceID, callback: Namespace.CpuAndMemoryCallback): void;

    /**
     * Calls CpuAndMemory.
     * @param request NamespaceID message or plain object
     * @returns Promise
     */
    public cpuAndMemory(request: INamespaceID): Promise<CpuAndMemoryResponse>;

    /**
     * Calls ServiceEndpoints.
     * @param request ServiceEndpointsRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and ServiceEndpointsResponse
     */
    public serviceEndpoints(request: IServiceEndpointsRequest, callback: Namespace.ServiceEndpointsCallback): void;

    /**
     * Calls ServiceEndpoints.
     * @param request ServiceEndpointsRequest message or plain object
     * @returns Promise
     */
    public serviceEndpoints(request: IServiceEndpointsRequest): Promise<ServiceEndpointsResponse>;

    /**
     * Calls Destroy.
     * @param request NamespaceID message or plain object
     * @param callback Node-style callback called with the error, if any, and Empty
     */
    public destroy(request: INamespaceID, callback: Namespace.DestroyCallback): void;

    /**
     * Calls Destroy.
     * @param request NamespaceID message or plain object
     * @returns Promise
     */
    public destroy(request: INamespaceID): Promise<google.protobuf.Empty>;
}

export namespace Namespace {

    /**
     * Callback as used by {@link Namespace#index}.
     * @param error Error, if any
     * @param [response] NamespaceList
     */
    type IndexCallback = (error: (Error|null), response?: NamespaceList) => void;

    /**
     * Callback as used by {@link Namespace#store}.
     * @param error Error, if any
     * @param [response] NsStoreResponse
     */
    type StoreCallback = (error: (Error|null), response?: NsStoreResponse) => void;

    /**
     * Callback as used by {@link Namespace#cpuAndMemory}.
     * @param error Error, if any
     * @param [response] CpuAndMemoryResponse
     */
    type CpuAndMemoryCallback = (error: (Error|null), response?: CpuAndMemoryResponse) => void;

    /**
     * Callback as used by {@link Namespace#serviceEndpoints}.
     * @param error Error, if any
     * @param [response] ServiceEndpointsResponse
     */
    type ServiceEndpointsCallback = (error: (Error|null), response?: ServiceEndpointsResponse) => void;

    /**
     * Callback as used by {@link Namespace#destroy}.
     * @param error Error, if any
     * @param [response] Empty
     */
    type DestroyCallback = (error: (Error|null), response?: google.protobuf.Empty) => void;
}

/** Represents a ProjectDestroyRequest. */
export class ProjectDestroyRequest implements IProjectDestroyRequest {

    /**
     * Constructs a new ProjectDestroyRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IProjectDestroyRequest);

    /** ProjectDestroyRequest namespace_id. */
    public namespace_id: (number|Long);

    /** ProjectDestroyRequest project_id. */
    public project_id: (number|Long);
}

/** Represents a ProjectShowRequest. */
export class ProjectShowRequest implements IProjectShowRequest {

    /**
     * Constructs a new ProjectShowRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IProjectShowRequest);

    /** ProjectShowRequest namespace_id. */
    public namespace_id: (number|Long);

    /** ProjectShowRequest project_id. */
    public project_id: (number|Long);
}

/** Represents a ProjectShowResponse. */
export class ProjectShowResponse implements IProjectShowResponse {

    /**
     * Constructs a new ProjectShowResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IProjectShowResponse);

    /** ProjectShowResponse id. */
    public id: (number|Long);

    /** ProjectShowResponse name. */
    public name: string;

    /** ProjectShowResponse gitlab_project_id. */
    public gitlab_project_id: (number|Long);

    /** ProjectShowResponse gitlab_branch. */
    public gitlab_branch: string;

    /** ProjectShowResponse gitlab_commit. */
    public gitlab_commit: string;

    /** ProjectShowResponse config. */
    public config: string;

    /** ProjectShowResponse docker_image. */
    public docker_image: string;

    /** ProjectShowResponse atomic. */
    public atomic: boolean;

    /** ProjectShowResponse gitlab_commit_web_url. */
    public gitlab_commit_web_url: string;

    /** ProjectShowResponse gitlab_commit_title. */
    public gitlab_commit_title: string;

    /** ProjectShowResponse gitlab_commit_author. */
    public gitlab_commit_author: string;

    /** ProjectShowResponse gitlab_commit_date. */
    public gitlab_commit_date: string;

    /** ProjectShowResponse urls. */
    public urls: string[];

    /** ProjectShowResponse namespace. */
    public namespace?: (ProjectShowResponse.INamespace|null);

    /** ProjectShowResponse cpu. */
    public cpu: string;

    /** ProjectShowResponse memory. */
    public memory: string;

    /** ProjectShowResponse override_values. */
    public override_values: string;

    /** ProjectShowResponse created_at. */
    public created_at?: (google.protobuf.ITimestamp|null);

    /** ProjectShowResponse updated_at. */
    public updated_at?: (google.protobuf.ITimestamp|null);
}

export namespace ProjectShowResponse {

    /** Properties of a Namespace. */
    interface INamespace {

        /** Namespace id */
        id?: (number|Long|null);

        /** Namespace name */
        name?: (string|null);
    }

    /** Represents a Namespace. */
    class Namespace implements INamespace {

        /**
         * Constructs a new Namespace.
         * @param [properties] Properties to set
         */
        constructor(properties?: ProjectShowResponse.INamespace);

        /** Namespace id. */
        public id: (number|Long);

        /** Namespace name. */
        public name: string;
    }
}

/** Represents an AllPodContainersRequest. */
export class AllPodContainersRequest implements IAllPodContainersRequest {

    /**
     * Constructs a new AllPodContainersRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IAllPodContainersRequest);

    /** AllPodContainersRequest namespace_id. */
    public namespace_id: (number|Long);

    /** AllPodContainersRequest project_id. */
    public project_id: (number|Long);
}

/** Represents a PodLog. */
export class PodLog implements IPodLog {

    /**
     * Constructs a new PodLog.
     * @param [properties] Properties to set
     */
    constructor(properties?: IPodLog);

    /** PodLog pod_name. */
    public pod_name: string;

    /** PodLog container_name. */
    public container_name: string;

    /** PodLog log. */
    public log: string;
}

/** Represents an AllPodContainersResponse. */
export class AllPodContainersResponse implements IAllPodContainersResponse {

    /**
     * Constructs a new AllPodContainersResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IAllPodContainersResponse);

    /** AllPodContainersResponse data. */
    public data: IPodLog[];
}

/** Represents a PodContainerLogRequest. */
export class PodContainerLogRequest implements IPodContainerLogRequest {

    /**
     * Constructs a new PodContainerLogRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IPodContainerLogRequest);

    /** PodContainerLogRequest namespace_id. */
    public namespace_id: (number|Long);

    /** PodContainerLogRequest project_id. */
    public project_id: (number|Long);

    /** PodContainerLogRequest pod. */
    public pod: string;

    /** PodContainerLogRequest container. */
    public container: string;
}

/** Represents a PodContainerLogResponse. */
export class PodContainerLogResponse implements IPodContainerLogResponse {

    /**
     * Constructs a new PodContainerLogResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IPodContainerLogResponse);

    /** PodContainerLogResponse data. */
    public data?: (IPodLog|null);
}

/** Represents a Project */
export class Project extends $protobuf.rpc.Service {

    /**
     * Constructs a new Project service.
     * @param rpcImpl RPC implementation
     * @param [requestDelimited=false] Whether requests are length-delimited
     * @param [responseDelimited=false] Whether responses are length-delimited
     */
    constructor(rpcImpl: $protobuf.RPCImpl, requestDelimited?: boolean, responseDelimited?: boolean);

    /**
     * Calls Destroy.
     * @param request ProjectDestroyRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and Empty
     */
    public destroy(request: IProjectDestroyRequest, callback: Project.DestroyCallback): void;

    /**
     * Calls Destroy.
     * @param request ProjectDestroyRequest message or plain object
     * @returns Promise
     */
    public destroy(request: IProjectDestroyRequest): Promise<google.protobuf.Empty>;

    /**
     * Calls Show.
     * @param request ProjectShowRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and ProjectShowResponse
     */
    public show(request: IProjectShowRequest, callback: Project.ShowCallback): void;

    /**
     * Calls Show.
     * @param request ProjectShowRequest message or plain object
     * @returns Promise
     */
    public show(request: IProjectShowRequest): Promise<ProjectShowResponse>;

    /**
     * Calls AllPodContainers.
     * @param request AllPodContainersRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and AllPodContainersResponse
     */
    public allPodContainers(request: IAllPodContainersRequest, callback: Project.AllPodContainersCallback): void;

    /**
     * Calls AllPodContainers.
     * @param request AllPodContainersRequest message or plain object
     * @returns Promise
     */
    public allPodContainers(request: IAllPodContainersRequest): Promise<AllPodContainersResponse>;

    /**
     * Calls PodContainerLog.
     * @param request PodContainerLogRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and PodContainerLogResponse
     */
    public podContainerLog(request: IPodContainerLogRequest, callback: Project.PodContainerLogCallback): void;

    /**
     * Calls PodContainerLog.
     * @param request PodContainerLogRequest message or plain object
     * @returns Promise
     */
    public podContainerLog(request: IPodContainerLogRequest): Promise<PodContainerLogResponse>;
}

export namespace Project {

    /**
     * Callback as used by {@link Project#destroy}.
     * @param error Error, if any
     * @param [response] Empty
     */
    type DestroyCallback = (error: (Error|null), response?: google.protobuf.Empty) => void;

    /**
     * Callback as used by {@link Project#show}.
     * @param error Error, if any
     * @param [response] ProjectShowResponse
     */
    type ShowCallback = (error: (Error|null), response?: ProjectShowResponse) => void;

    /**
     * Callback as used by {@link Project#allPodContainers}.
     * @param error Error, if any
     * @param [response] AllPodContainersResponse
     */
    type AllPodContainersCallback = (error: (Error|null), response?: AllPodContainersResponse) => void;

    /**
     * Callback as used by {@link Project#podContainerLog}.
     * @param error Error, if any
     * @param [response] PodContainerLogResponse
     */
    type PodContainerLogCallback = (error: (Error|null), response?: PodContainerLogResponse) => void;
}
