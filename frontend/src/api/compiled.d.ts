import * as $protobuf from "protobufjs";
/** Properties of a LoginRequest. */
export interface ILoginRequest {

    /** LoginRequest username */
    username?: (string|null);

    /** LoginRequest password */
    password?: (string|null);
}

/** Represents a LoginRequest. */
export class LoginRequest implements ILoginRequest {

    /**
     * Constructs a new LoginRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: ILoginRequest);

    /** LoginRequest username. */
    public username: string;

    /** LoginRequest password. */
    public password: string;

    /**
     * Encodes the specified LoginRequest message. Does not implicitly {@link LoginRequest.verify|verify} messages.
     * @param message LoginRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: LoginRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a LoginRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns LoginRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): LoginRequest;
}

/** Properties of a LoginResponse. */
export interface ILoginResponse {

    /** LoginResponse token */
    token?: (string|null);

    /** LoginResponse expires_in */
    expires_in?: (number|null);
}

/** Represents a LoginResponse. */
export class LoginResponse implements ILoginResponse {

    /**
     * Constructs a new LoginResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: ILoginResponse);

    /** LoginResponse token. */
    public token: string;

    /** LoginResponse expires_in. */
    public expires_in: number;

    /**
     * Encodes the specified LoginResponse message. Does not implicitly {@link LoginResponse.verify|verify} messages.
     * @param message LoginResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: LoginResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a LoginResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns LoginResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): LoginResponse;
}

/** Properties of an InfoResponse. */
export interface IInfoResponse {

    /** InfoResponse id */
    id?: (string|null);

    /** InfoResponse avatar */
    avatar?: (string|null);

    /** InfoResponse name */
    name?: (string|null);

    /** InfoResponse email */
    email?: (string|null);

    /** InfoResponse logout_url */
    logout_url?: (string|null);

    /** InfoResponse roles */
    roles?: (string[]|null);
}

/** Represents an InfoResponse. */
export class InfoResponse implements IInfoResponse {

    /**
     * Constructs a new InfoResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IInfoResponse);

    /** InfoResponse id. */
    public id: string;

    /** InfoResponse avatar. */
    public avatar: string;

    /** InfoResponse name. */
    public name: string;

    /** InfoResponse email. */
    public email: string;

    /** InfoResponse logout_url. */
    public logout_url: string;

    /** InfoResponse roles. */
    public roles: string[];

    /**
     * Encodes the specified InfoResponse message. Does not implicitly {@link InfoResponse.verify|verify} messages.
     * @param message InfoResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: InfoResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes an InfoResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns InfoResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): InfoResponse;
}

/** Properties of an OidcSetting. */
export interface IOidcSetting {

    /** OidcSetting enabled */
    enabled?: (boolean|null);

    /** OidcSetting name */
    name?: (string|null);

    /** OidcSetting url */
    url?: (string|null);

    /** OidcSetting end_session_endpoint */
    end_session_endpoint?: (string|null);

    /** OidcSetting state */
    state?: (string|null);
}

/** Represents an OidcSetting. */
export class OidcSetting implements IOidcSetting {

    /**
     * Constructs a new OidcSetting.
     * @param [properties] Properties to set
     */
    constructor(properties?: IOidcSetting);

    /** OidcSetting enabled. */
    public enabled: boolean;

    /** OidcSetting name. */
    public name: string;

    /** OidcSetting url. */
    public url: string;

    /** OidcSetting end_session_endpoint. */
    public end_session_endpoint: string;

    /** OidcSetting state. */
    public state: string;

    /**
     * Encodes the specified OidcSetting message. Does not implicitly {@link OidcSetting.verify|verify} messages.
     * @param message OidcSetting message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: OidcSetting, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes an OidcSetting message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns OidcSetting
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): OidcSetting;
}

/** Properties of a SettingsResponse. */
export interface ISettingsResponse {

    /** SettingsResponse items */
    items?: (OidcSetting[]|null);
}

/** Represents a SettingsResponse. */
export class SettingsResponse implements ISettingsResponse {

    /**
     * Constructs a new SettingsResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: ISettingsResponse);

    /** SettingsResponse items. */
    public items: OidcSetting[];

    /**
     * Encodes the specified SettingsResponse message. Does not implicitly {@link SettingsResponse.verify|verify} messages.
     * @param message SettingsResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: SettingsResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a SettingsResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns SettingsResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): SettingsResponse;
}

/** Properties of an ExchangeRequest. */
export interface IExchangeRequest {

    /** ExchangeRequest code */
    code?: (string|null);
}

/** Represents an ExchangeRequest. */
export class ExchangeRequest implements IExchangeRequest {

    /**
     * Constructs a new ExchangeRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IExchangeRequest);

    /** ExchangeRequest code. */
    public code: string;

    /**
     * Encodes the specified ExchangeRequest message. Does not implicitly {@link ExchangeRequest.verify|verify} messages.
     * @param message ExchangeRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: ExchangeRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes an ExchangeRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns ExchangeRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): ExchangeRequest;
}

/** Represents an Auth */
export class Auth extends $protobuf.rpc.Service {

    /**
     * Constructs a new Auth service.
     * @param rpcImpl RPC implementation
     * @param [requestDelimited=false] Whether requests are length-delimited
     * @param [responseDelimited=false] Whether responses are length-delimited
     */
    constructor(rpcImpl: $protobuf.RPCImpl, requestDelimited?: boolean, responseDelimited?: boolean);

    /**
     * Calls Login.
     * @param request LoginRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and LoginResponse
     */
    public login(request: LoginRequest, callback: Auth.LoginCallback): void;

    /**
     * Calls Login.
     * @param request LoginRequest message or plain object
     * @returns Promise
     */
    public login(request: LoginRequest): Promise<LoginResponse>;

    /**
     * Calls Info.
     * @param request Empty message or plain object
     * @param callback Node-style callback called with the error, if any, and InfoResponse
     */
    public info(request: google.protobuf.Empty, callback: Auth.InfoCallback): void;

    /**
     * Calls Info.
     * @param request Empty message or plain object
     * @returns Promise
     */
    public info(request: google.protobuf.Empty): Promise<InfoResponse>;

    /**
     * Calls Settings.
     * @param request Empty message or plain object
     * @param callback Node-style callback called with the error, if any, and SettingsResponse
     */
    public settings(request: google.protobuf.Empty, callback: Auth.SettingsCallback): void;

    /**
     * Calls Settings.
     * @param request Empty message or plain object
     * @returns Promise
     */
    public settings(request: google.protobuf.Empty): Promise<SettingsResponse>;

    /**
     * Calls Exchange.
     * @param request ExchangeRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and LoginResponse
     */
    public exchange(request: ExchangeRequest, callback: Auth.ExchangeCallback): void;

    /**
     * Calls Exchange.
     * @param request ExchangeRequest message or plain object
     * @returns Promise
     */
    public exchange(request: ExchangeRequest): Promise<LoginResponse>;
}

export namespace Auth {

    /**
     * Callback as used by {@link Auth#login}.
     * @param error Error, if any
     * @param [response] LoginResponse
     */
    type LoginCallback = (error: (Error|null), response?: LoginResponse) => void;

    /**
     * Callback as used by {@link Auth#info}.
     * @param error Error, if any
     * @param [response] InfoResponse
     */
    type InfoCallback = (error: (Error|null), response?: InfoResponse) => void;

    /**
     * Callback as used by {@link Auth#settings}.
     * @param error Error, if any
     * @param [response] SettingsResponse
     */
    type SettingsCallback = (error: (Error|null), response?: SettingsResponse) => void;

    /**
     * Callback as used by {@link Auth#exchange}.
     * @param error Error, if any
     * @param [response] LoginResponse
     */
    type ExchangeCallback = (error: (Error|null), response?: LoginResponse) => void;
}

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

    /**
     * Encodes the specified ClusterInfoResponse message. Does not implicitly {@link ClusterInfoResponse.verify|verify} messages.
     * @param message ClusterInfoResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: ClusterInfoResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a ClusterInfoResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns ClusterInfoResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): ClusterInfoResponse;
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
    public info(request: google.protobuf.Empty, callback: Cluster.InfoCallback): void;

    /**
     * Calls Info.
     * @param request Empty message or plain object
     * @returns Promise
     */
    public info(request: google.protobuf.Empty): Promise<ClusterInfoResponse>;
}

export namespace Cluster {

    /**
     * Callback as used by {@link Cluster#info}.
     * @param error Error, if any
     * @param [response] ClusterInfoResponse
     */
    type InfoCallback = (error: (Error|null), response?: ClusterInfoResponse) => void;
}

/** Represents a CopyToPodRequest. */
export class CopyToPodRequest implements ICopyToPodRequest {

    /**
     * Constructs a new CopyToPodRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: ICopyToPodRequest);

    /** CopyToPodRequest file_id. */
    public file_id: number;

    /** CopyToPodRequest namespace. */
    public namespace: string;

    /** CopyToPodRequest pod. */
    public pod: string;

    /** CopyToPodRequest container. */
    public container: string;

    /**
     * Encodes the specified CopyToPodRequest message. Does not implicitly {@link CopyToPodRequest.verify|verify} messages.
     * @param message CopyToPodRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: CopyToPodRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a CopyToPodRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns CopyToPodRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): CopyToPodRequest;
}

/** Represents a CopyToPodResponse. */
export class CopyToPodResponse implements ICopyToPodResponse {

    /**
     * Constructs a new CopyToPodResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: ICopyToPodResponse);

    /** CopyToPodResponse podFilePath. */
    public podFilePath: string;

    /** CopyToPodResponse output. */
    public output: string;

    /**
     * Encodes the specified CopyToPodResponse message. Does not implicitly {@link CopyToPodResponse.verify|verify} messages.
     * @param message CopyToPodResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: CopyToPodResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a CopyToPodResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns CopyToPodResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): CopyToPodResponse;
}

/** Represents a Cp */
export class Cp extends $protobuf.rpc.Service {

    /**
     * Constructs a new Cp service.
     * @param rpcImpl RPC implementation
     * @param [requestDelimited=false] Whether requests are length-delimited
     * @param [responseDelimited=false] Whether responses are length-delimited
     */
    constructor(rpcImpl: $protobuf.RPCImpl, requestDelimited?: boolean, responseDelimited?: boolean);

    /**
     * Calls CopyToPod.
     * @param request CopyToPodRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and CopyToPodResponse
     */
    public copyToPod(request: CopyToPodRequest, callback: Cp.CopyToPodCallback): void;

    /**
     * Calls CopyToPod.
     * @param request CopyToPodRequest message or plain object
     * @returns Promise
     */
    public copyToPod(request: CopyToPodRequest): Promise<CopyToPodResponse>;
}

export namespace Cp {

    /**
     * Callback as used by {@link Cp#copyToPod}.
     * @param error Error, if any
     * @param [response] CopyToPodResponse
     */
    type CopyToPodCallback = (error: (Error|null), response?: CopyToPodResponse) => void;
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

    /**
     * Encodes the specified GitlabDestroyRequest message. Does not implicitly {@link GitlabDestroyRequest.verify|verify} messages.
     * @param message GitlabDestroyRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: GitlabDestroyRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a GitlabDestroyRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns GitlabDestroyRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): GitlabDestroyRequest;
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

    /**
     * Encodes the specified EnableProjectRequest message. Does not implicitly {@link EnableProjectRequest.verify|verify} messages.
     * @param message EnableProjectRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: EnableProjectRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes an EnableProjectRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns EnableProjectRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): EnableProjectRequest;
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

    /**
     * Encodes the specified DisableProjectRequest message. Does not implicitly {@link DisableProjectRequest.verify|verify} messages.
     * @param message DisableProjectRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: DisableProjectRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a DisableProjectRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns DisableProjectRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): DisableProjectRequest;
}

/** Represents a GitlabProjectInfo. */
export class GitlabProjectInfo implements IGitlabProjectInfo {

    /**
     * Constructs a new GitlabProjectInfo.
     * @param [properties] Properties to set
     */
    constructor(properties?: IGitlabProjectInfo);

    /** GitlabProjectInfo id. */
    public id: number;

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

    /**
     * Encodes the specified GitlabProjectInfo message. Does not implicitly {@link GitlabProjectInfo.verify|verify} messages.
     * @param message GitlabProjectInfo message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: GitlabProjectInfo, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a GitlabProjectInfo message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns GitlabProjectInfo
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): GitlabProjectInfo;
}

/** Represents a ProjectListResponse. */
export class ProjectListResponse implements IProjectListResponse {

    /**
     * Constructs a new ProjectListResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IProjectListResponse);

    /** ProjectListResponse data. */
    public data: GitlabProjectInfo[];

    /**
     * Encodes the specified ProjectListResponse message. Does not implicitly {@link ProjectListResponse.verify|verify} messages.
     * @param message ProjectListResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: ProjectListResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a ProjectListResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns ProjectListResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): ProjectListResponse;
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

    /**
     * Encodes the specified Option message. Does not implicitly {@link Option.verify|verify} messages.
     * @param message Option message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: Option, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes an Option message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns Option
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): Option;
}

/** Represents a ProjectsResponse. */
export class ProjectsResponse implements IProjectsResponse {

    /**
     * Constructs a new ProjectsResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IProjectsResponse);

    /** ProjectsResponse data. */
    public data: Option[];

    /**
     * Encodes the specified ProjectsResponse message. Does not implicitly {@link ProjectsResponse.verify|verify} messages.
     * @param message ProjectsResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: ProjectsResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a ProjectsResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns ProjectsResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): ProjectsResponse;
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

    /** BranchesRequest all. */
    public all: boolean;

    /**
     * Encodes the specified BranchesRequest message. Does not implicitly {@link BranchesRequest.verify|verify} messages.
     * @param message BranchesRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: BranchesRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a BranchesRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns BranchesRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): BranchesRequest;
}

/** Represents a BranchesResponse. */
export class BranchesResponse implements IBranchesResponse {

    /**
     * Constructs a new BranchesResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IBranchesResponse);

    /** BranchesResponse data. */
    public data: Option[];

    /**
     * Encodes the specified BranchesResponse message. Does not implicitly {@link BranchesResponse.verify|verify} messages.
     * @param message BranchesResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: BranchesResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a BranchesResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns BranchesResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): BranchesResponse;
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

    /**
     * Encodes the specified CommitsRequest message. Does not implicitly {@link CommitsRequest.verify|verify} messages.
     * @param message CommitsRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: CommitsRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a CommitsRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns CommitsRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): CommitsRequest;
}

/** Represents a CommitsResponse. */
export class CommitsResponse implements ICommitsResponse {

    /**
     * Constructs a new CommitsResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: ICommitsResponse);

    /** CommitsResponse data. */
    public data: Option[];

    /**
     * Encodes the specified CommitsResponse message. Does not implicitly {@link CommitsResponse.verify|verify} messages.
     * @param message CommitsResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: CommitsResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a CommitsResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns CommitsResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): CommitsResponse;
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

    /**
     * Encodes the specified CommitRequest message. Does not implicitly {@link CommitRequest.verify|verify} messages.
     * @param message CommitRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: CommitRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a CommitRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns CommitRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): CommitRequest;
}

/** Represents a CommitResponse. */
export class CommitResponse implements ICommitResponse {

    /**
     * Constructs a new CommitResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: ICommitResponse);

    /** CommitResponse data. */
    public data?: (Option|null);

    /**
     * Encodes the specified CommitResponse message. Does not implicitly {@link CommitResponse.verify|verify} messages.
     * @param message CommitResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: CommitResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a CommitResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns CommitResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): CommitResponse;
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

    /**
     * Encodes the specified PipelineInfoRequest message. Does not implicitly {@link PipelineInfoRequest.verify|verify} messages.
     * @param message PipelineInfoRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: PipelineInfoRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a PipelineInfoRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns PipelineInfoRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): PipelineInfoRequest;
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

    /**
     * Encodes the specified PipelineInfoResponse message. Does not implicitly {@link PipelineInfoResponse.verify|verify} messages.
     * @param message PipelineInfoResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: PipelineInfoResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a PipelineInfoResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns PipelineInfoResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): PipelineInfoResponse;
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

    /**
     * Encodes the specified ConfigFileRequest message. Does not implicitly {@link ConfigFileRequest.verify|verify} messages.
     * @param message ConfigFileRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: ConfigFileRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a ConfigFileRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns ConfigFileRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): ConfigFileRequest;
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

    /**
     * Encodes the specified ConfigFileResponse message. Does not implicitly {@link ConfigFileResponse.verify|verify} messages.
     * @param message ConfigFileResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: ConfigFileResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a ConfigFileResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns ConfigFileResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): ConfigFileResponse;
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
    public enableProject(request: EnableProjectRequest, callback: Gitlab.EnableProjectCallback): void;

    /**
     * Calls EnableProject.
     * @param request EnableProjectRequest message or plain object
     * @returns Promise
     */
    public enableProject(request: EnableProjectRequest): Promise<google.protobuf.Empty>;

    /**
     * Calls DisableProject.
     * @param request DisableProjectRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and Empty
     */
    public disableProject(request: DisableProjectRequest, callback: Gitlab.DisableProjectCallback): void;

    /**
     * Calls DisableProject.
     * @param request DisableProjectRequest message or plain object
     * @returns Promise
     */
    public disableProject(request: DisableProjectRequest): Promise<google.protobuf.Empty>;

    /**
     * Calls ProjectList.
     * @param request Empty message or plain object
     * @param callback Node-style callback called with the error, if any, and ProjectListResponse
     */
    public projectList(request: google.protobuf.Empty, callback: Gitlab.ProjectListCallback): void;

    /**
     * Calls ProjectList.
     * @param request Empty message or plain object
     * @returns Promise
     */
    public projectList(request: google.protobuf.Empty): Promise<ProjectListResponse>;

    /**
     * Calls Projects.
     * @param request Empty message or plain object
     * @param callback Node-style callback called with the error, if any, and ProjectsResponse
     */
    public projects(request: google.protobuf.Empty, callback: Gitlab.ProjectsCallback): void;

    /**
     * Calls Projects.
     * @param request Empty message or plain object
     * @returns Promise
     */
    public projects(request: google.protobuf.Empty): Promise<ProjectsResponse>;

    /**
     * Calls Branches.
     * @param request BranchesRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and BranchesResponse
     */
    public branches(request: BranchesRequest, callback: Gitlab.BranchesCallback): void;

    /**
     * Calls Branches.
     * @param request BranchesRequest message or plain object
     * @returns Promise
     */
    public branches(request: BranchesRequest): Promise<BranchesResponse>;

    /**
     * Calls Commits.
     * @param request CommitsRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and CommitsResponse
     */
    public commits(request: CommitsRequest, callback: Gitlab.CommitsCallback): void;

    /**
     * Calls Commits.
     * @param request CommitsRequest message or plain object
     * @returns Promise
     */
    public commits(request: CommitsRequest): Promise<CommitsResponse>;

    /**
     * Calls Commit.
     * @param request CommitRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and CommitResponse
     */
    public commit(request: CommitRequest, callback: Gitlab.CommitCallback): void;

    /**
     * Calls Commit.
     * @param request CommitRequest message or plain object
     * @returns Promise
     */
    public commit(request: CommitRequest): Promise<CommitResponse>;

    /**
     * Calls PipelineInfo.
     * @param request PipelineInfoRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and PipelineInfoResponse
     */
    public pipelineInfo(request: PipelineInfoRequest, callback: Gitlab.PipelineInfoCallback): void;

    /**
     * Calls PipelineInfo.
     * @param request PipelineInfoRequest message or plain object
     * @returns Promise
     */
    public pipelineInfo(request: PipelineInfoRequest): Promise<PipelineInfoResponse>;

    /**
     * Calls ConfigFile.
     * @param request ConfigFileRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and ConfigFileResponse
     */
    public configFile(request: ConfigFileRequest, callback: Gitlab.ConfigFileCallback): void;

    /**
     * Calls ConfigFile.
     * @param request ConfigFileRequest message or plain object
     * @returns Promise
     */
    public configFile(request: ConfigFileRequest): Promise<ConfigFileResponse>;
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

/** Represents a Config. */
export class Config implements IConfig {

    /**
     * Constructs a new Config.
     * @param [properties] Properties to set
     */
    constructor(properties?: IConfig);

    /** Config config_file. */
    public config_file: string;

    /** Config config_file_values. */
    public config_file_values: string;

    /** Config config_field. */
    public config_field: string;

    /** Config is_simple_env. */
    public is_simple_env: boolean;

    /** Config config_file_type. */
    public config_file_type: string;

    /** Config local_chart_path. */
    public local_chart_path: string;

    /** Config branches. */
    public branches: string[];

    /** Config values_yaml. */
    public values_yaml: string;

    /**
     * Encodes the specified Config message. Does not implicitly {@link Config.verify|verify} messages.
     * @param message Config message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: Config, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a Config message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns Config
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): Config;
}

/** Represents a MarsShowRequest. */
export class MarsShowRequest implements IMarsShowRequest {

    /**
     * Constructs a new MarsShowRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IMarsShowRequest);

    /** MarsShowRequest project_id. */
    public project_id: number;

    /** MarsShowRequest branch. */
    public branch: string;

    /**
     * Encodes the specified MarsShowRequest message. Does not implicitly {@link MarsShowRequest.verify|verify} messages.
     * @param message MarsShowRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: MarsShowRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a MarsShowRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns MarsShowRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): MarsShowRequest;
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
    public config?: (Config|null);

    /**
     * Encodes the specified MarsShowResponse message. Does not implicitly {@link MarsShowResponse.verify|verify} messages.
     * @param message MarsShowResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: MarsShowResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a MarsShowResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns MarsShowResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): MarsShowResponse;
}

/** Represents a GlobalConfigRequest. */
export class GlobalConfigRequest implements IGlobalConfigRequest {

    /**
     * Constructs a new GlobalConfigRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IGlobalConfigRequest);

    /** GlobalConfigRequest project_id. */
    public project_id: number;

    /**
     * Encodes the specified GlobalConfigRequest message. Does not implicitly {@link GlobalConfigRequest.verify|verify} messages.
     * @param message GlobalConfigRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: GlobalConfigRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a GlobalConfigRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns GlobalConfigRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): GlobalConfigRequest;
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
    public config?: (Config|null);

    /**
     * Encodes the specified GlobalConfigResponse message. Does not implicitly {@link GlobalConfigResponse.verify|verify} messages.
     * @param message GlobalConfigResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: GlobalConfigResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a GlobalConfigResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns GlobalConfigResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): GlobalConfigResponse;
}

/** Represents a MarsUpdateRequest. */
export class MarsUpdateRequest implements IMarsUpdateRequest {

    /**
     * Constructs a new MarsUpdateRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IMarsUpdateRequest);

    /** MarsUpdateRequest project_id. */
    public project_id: number;

    /** MarsUpdateRequest config. */
    public config?: (Config|null);

    /**
     * Encodes the specified MarsUpdateRequest message. Does not implicitly {@link MarsUpdateRequest.verify|verify} messages.
     * @param message MarsUpdateRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: MarsUpdateRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a MarsUpdateRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns MarsUpdateRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): MarsUpdateRequest;
}

/** Represents a MarsUpdateResponse. */
export class MarsUpdateResponse implements IMarsUpdateResponse {

    /**
     * Constructs a new MarsUpdateResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IMarsUpdateResponse);

    /** MarsUpdateResponse config. */
    public config?: (Config|null);

    /**
     * Encodes the specified MarsUpdateResponse message. Does not implicitly {@link MarsUpdateResponse.verify|verify} messages.
     * @param message MarsUpdateResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: MarsUpdateResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a MarsUpdateResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns MarsUpdateResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): MarsUpdateResponse;
}

/** Represents a ToggleEnabledRequest. */
export class ToggleEnabledRequest implements IToggleEnabledRequest {

    /**
     * Constructs a new ToggleEnabledRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IToggleEnabledRequest);

    /** ToggleEnabledRequest project_id. */
    public project_id: number;

    /** ToggleEnabledRequest enabled. */
    public enabled: boolean;

    /**
     * Encodes the specified ToggleEnabledRequest message. Does not implicitly {@link ToggleEnabledRequest.verify|verify} messages.
     * @param message ToggleEnabledRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: ToggleEnabledRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a ToggleEnabledRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns ToggleEnabledRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): ToggleEnabledRequest;
}

/** Represents a DefaultChartValuesRequest. */
export class DefaultChartValuesRequest implements IDefaultChartValuesRequest {

    /**
     * Constructs a new DefaultChartValuesRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IDefaultChartValuesRequest);

    /** DefaultChartValuesRequest project_id. */
    public project_id: number;

    /** DefaultChartValuesRequest branch. */
    public branch: string;

    /**
     * Encodes the specified DefaultChartValuesRequest message. Does not implicitly {@link DefaultChartValuesRequest.verify|verify} messages.
     * @param message DefaultChartValuesRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: DefaultChartValuesRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a DefaultChartValuesRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns DefaultChartValuesRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): DefaultChartValuesRequest;
}

/** Represents a DefaultChartValues. */
export class DefaultChartValues implements IDefaultChartValues {

    /**
     * Constructs a new DefaultChartValues.
     * @param [properties] Properties to set
     */
    constructor(properties?: IDefaultChartValues);

    /** DefaultChartValues value. */
    public value: string;

    /**
     * Encodes the specified DefaultChartValues message. Does not implicitly {@link DefaultChartValues.verify|verify} messages.
     * @param message DefaultChartValues message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: DefaultChartValues, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a DefaultChartValues message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns DefaultChartValues
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): DefaultChartValues;
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
    public show(request: MarsShowRequest, callback: Mars.ShowCallback): void;

    /**
     * Calls Show.
     * @param request MarsShowRequest message or plain object
     * @returns Promise
     */
    public show(request: MarsShowRequest): Promise<MarsShowResponse>;

    /**
     * Calls GlobalConfig.
     * @param request GlobalConfigRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and GlobalConfigResponse
     */
    public globalConfig(request: GlobalConfigRequest, callback: Mars.GlobalConfigCallback): void;

    /**
     * Calls GlobalConfig.
     * @param request GlobalConfigRequest message or plain object
     * @returns Promise
     */
    public globalConfig(request: GlobalConfigRequest): Promise<GlobalConfigResponse>;

    /**
     * Calls ToggleEnabled.
     * @param request ToggleEnabledRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and Empty
     */
    public toggleEnabled(request: ToggleEnabledRequest, callback: Mars.ToggleEnabledCallback): void;

    /**
     * Calls ToggleEnabled.
     * @param request ToggleEnabledRequest message or plain object
     * @returns Promise
     */
    public toggleEnabled(request: ToggleEnabledRequest): Promise<google.protobuf.Empty>;

    /**
     * Calls Update.
     * @param request MarsUpdateRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and MarsUpdateResponse
     */
    public update(request: MarsUpdateRequest, callback: Mars.UpdateCallback): void;

    /**
     * Calls Update.
     * @param request MarsUpdateRequest message or plain object
     * @returns Promise
     */
    public update(request: MarsUpdateRequest): Promise<MarsUpdateResponse>;

    /**
     * Calls GetDefaultChartValues.
     * @param request DefaultChartValuesRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and DefaultChartValues
     */
    public getDefaultChartValues(request: DefaultChartValuesRequest, callback: Mars.GetDefaultChartValuesCallback): void;

    /**
     * Calls GetDefaultChartValues.
     * @param request DefaultChartValuesRequest message or plain object
     * @returns Promise
     */
    public getDefaultChartValues(request: DefaultChartValuesRequest): Promise<DefaultChartValues>;
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

    /**
     * Callback as used by {@link Mars#getDefaultChartValues}.
     * @param error Error, if any
     * @param [response] DefaultChartValues
     */
    type GetDefaultChartValuesCallback = (error: (Error|null), response?: DefaultChartValues) => void;
}

/** Represents a ProjectByIDRequest. */
export class ProjectByIDRequest implements IProjectByIDRequest {

    /**
     * Constructs a new ProjectByIDRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IProjectByIDRequest);

    /** ProjectByIDRequest namespace. */
    public namespace: string;

    /** ProjectByIDRequest pod. */
    public pod: string;

    /**
     * Encodes the specified ProjectByIDRequest message. Does not implicitly {@link ProjectByIDRequest.verify|verify} messages.
     * @param message ProjectByIDRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: ProjectByIDRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a ProjectByIDRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns ProjectByIDRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): ProjectByIDRequest;
}

/** Represents a ProjectByIDResponse. */
export class ProjectByIDResponse implements IProjectByIDResponse {

    /**
     * Constructs a new ProjectByIDResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IProjectByIDResponse);

    /** ProjectByIDResponse cpu. */
    public cpu: number;

    /** ProjectByIDResponse memory. */
    public memory: number;

    /** ProjectByIDResponse humanize_cpu. */
    public humanize_cpu: string;

    /** ProjectByIDResponse humanize_memory. */
    public humanize_memory: string;

    /** ProjectByIDResponse time. */
    public time: string;

    /** ProjectByIDResponse length. */
    public length: number;

    /**
     * Encodes the specified ProjectByIDResponse message. Does not implicitly {@link ProjectByIDResponse.verify|verify} messages.
     * @param message ProjectByIDResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: ProjectByIDResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a ProjectByIDResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns ProjectByIDResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): ProjectByIDResponse;
}

/** Represents a Metrics */
export class Metrics extends $protobuf.rpc.Service {

    /**
     * Constructs a new Metrics service.
     * @param rpcImpl RPC implementation
     * @param [requestDelimited=false] Whether requests are length-delimited
     * @param [responseDelimited=false] Whether responses are length-delimited
     */
    constructor(rpcImpl: $protobuf.RPCImpl, requestDelimited?: boolean, responseDelimited?: boolean);

    /**
     * Calls ProjectByID.
     * @param request ProjectByIDRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and ProjectByIDResponse
     */
    public projectByID(request: ProjectByIDRequest, callback: Metrics.ProjectByIDCallback): void;

    /**
     * Calls ProjectByID.
     * @param request ProjectByIDRequest message or plain object
     * @returns Promise
     */
    public projectByID(request: ProjectByIDRequest): Promise<ProjectByIDResponse>;
}

export namespace Metrics {

    /**
     * Callback as used by {@link Metrics#projectByID}.
     * @param error Error, if any
     * @param [response] ProjectByIDResponse
     */
    type ProjectByIDCallback = (error: (Error|null), response?: ProjectByIDResponse) => void;
}

/** Represents a GitlabProjectModal. */
export class GitlabProjectModal implements IGitlabProjectModal {

    /**
     * Constructs a new GitlabProjectModal.
     * @param [properties] Properties to set
     */
    constructor(properties?: IGitlabProjectModal);

    /** GitlabProjectModal id. */
    public id: number;

    /** GitlabProjectModal default_branch. */
    public default_branch: string;

    /** GitlabProjectModal name. */
    public name: string;

    /** GitlabProjectModal gitlab_project_id. */
    public gitlab_project_id: number;

    /** GitlabProjectModal enabled. */
    public enabled: boolean;

    /** GitlabProjectModal global_enabled. */
    public global_enabled: boolean;

    /** GitlabProjectModal global_config. */
    public global_config: string;

    /** GitlabProjectModal created_at. */
    public created_at?: (google.protobuf.Timestamp|null);

    /** GitlabProjectModal updated_at. */
    public updated_at?: (google.protobuf.Timestamp|null);

    /** GitlabProjectModal deleted_at. */
    public deleted_at?: (google.protobuf.Timestamp|null);

    /**
     * Encodes the specified GitlabProjectModal message. Does not implicitly {@link GitlabProjectModal.verify|verify} messages.
     * @param message GitlabProjectModal message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: GitlabProjectModal, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a GitlabProjectModal message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns GitlabProjectModal
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): GitlabProjectModal;
}

/** Represents a NamespaceModal. */
export class NamespaceModal implements INamespaceModal {

    /**
     * Constructs a new NamespaceModal.
     * @param [properties] Properties to set
     */
    constructor(properties?: INamespaceModal);

    /** NamespaceModal id. */
    public id: number;

    /** NamespaceModal name. */
    public name: string;

    /** NamespaceModal image_pull_secrets. */
    public image_pull_secrets: string[];

    /** NamespaceModal created_at. */
    public created_at?: (google.protobuf.Timestamp|null);

    /** NamespaceModal updated_at. */
    public updated_at?: (google.protobuf.Timestamp|null);

    /** NamespaceModal deleted_at. */
    public deleted_at?: (google.protobuf.Timestamp|null);

    /** NamespaceModal projects. */
    public projects: ProjectModal[];

    /**
     * Encodes the specified NamespaceModal message. Does not implicitly {@link NamespaceModal.verify|verify} messages.
     * @param message NamespaceModal message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: NamespaceModal, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a NamespaceModal message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns NamespaceModal
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): NamespaceModal;
}

/** Represents a ProjectModal. */
export class ProjectModal implements IProjectModal {

    /**
     * Constructs a new ProjectModal.
     * @param [properties] Properties to set
     */
    constructor(properties?: IProjectModal);

    /** ProjectModal id. */
    public id: number;

    /** ProjectModal name. */
    public name: string;

    /** ProjectModal gitlab_project_id. */
    public gitlab_project_id: number;

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
    public namespace_id: number;

    /** ProjectModal atomic. */
    public atomic: boolean;

    /** ProjectModal created_at. */
    public created_at?: (google.protobuf.Timestamp|null);

    /** ProjectModal updated_at. */
    public updated_at?: (google.protobuf.Timestamp|null);

    /** ProjectModal deleted_at. */
    public deleted_at?: (google.protobuf.Timestamp|null);

    /** ProjectModal namespace. */
    public namespace?: (NamespaceModal|null);

    /**
     * Encodes the specified ProjectModal message. Does not implicitly {@link ProjectModal.verify|verify} messages.
     * @param message ProjectModal message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: ProjectModal, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a ProjectModal message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns ProjectModal
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): ProjectModal;
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

            /**
             * Encodes the specified Empty message. Does not implicitly {@link google.protobuf.Empty.verify|verify} messages.
             * @param message Empty message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: google.protobuf.Empty, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an Empty message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns Empty
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): google.protobuf.Empty;
        }

        /** Properties of a FileDescriptorSet. */
        interface IFileDescriptorSet {

            /** FileDescriptorSet file */
            file?: (google.protobuf.FileDescriptorProto[]|null);
        }

        /** Represents a FileDescriptorSet. */
        class FileDescriptorSet implements IFileDescriptorSet {

            /**
             * Constructs a new FileDescriptorSet.
             * @param [properties] Properties to set
             */
            constructor(properties?: google.protobuf.IFileDescriptorSet);

            /** FileDescriptorSet file. */
            public file: google.protobuf.FileDescriptorProto[];

            /**
             * Encodes the specified FileDescriptorSet message. Does not implicitly {@link google.protobuf.FileDescriptorSet.verify|verify} messages.
             * @param message FileDescriptorSet message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: google.protobuf.FileDescriptorSet, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a FileDescriptorSet message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns FileDescriptorSet
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): google.protobuf.FileDescriptorSet;
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
            message_type?: (google.protobuf.DescriptorProto[]|null);

            /** FileDescriptorProto enum_type */
            enum_type?: (google.protobuf.EnumDescriptorProto[]|null);

            /** FileDescriptorProto service */
            service?: (google.protobuf.ServiceDescriptorProto[]|null);

            /** FileDescriptorProto extension */
            extension?: (google.protobuf.FieldDescriptorProto[]|null);

            /** FileDescriptorProto options */
            options?: (google.protobuf.FileOptions|null);

            /** FileDescriptorProto source_code_info */
            source_code_info?: (google.protobuf.SourceCodeInfo|null);

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
            public message_type: google.protobuf.DescriptorProto[];

            /** FileDescriptorProto enum_type. */
            public enum_type: google.protobuf.EnumDescriptorProto[];

            /** FileDescriptorProto service. */
            public service: google.protobuf.ServiceDescriptorProto[];

            /** FileDescriptorProto extension. */
            public extension: google.protobuf.FieldDescriptorProto[];

            /** FileDescriptorProto options. */
            public options?: (google.protobuf.FileOptions|null);

            /** FileDescriptorProto source_code_info. */
            public source_code_info?: (google.protobuf.SourceCodeInfo|null);

            /** FileDescriptorProto syntax. */
            public syntax: string;

            /**
             * Encodes the specified FileDescriptorProto message. Does not implicitly {@link google.protobuf.FileDescriptorProto.verify|verify} messages.
             * @param message FileDescriptorProto message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: google.protobuf.FileDescriptorProto, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a FileDescriptorProto message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns FileDescriptorProto
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): google.protobuf.FileDescriptorProto;
        }

        /** Properties of a DescriptorProto. */
        interface IDescriptorProto {

            /** DescriptorProto name */
            name?: (string|null);

            /** DescriptorProto field */
            field?: (google.protobuf.FieldDescriptorProto[]|null);

            /** DescriptorProto extension */
            extension?: (google.protobuf.FieldDescriptorProto[]|null);

            /** DescriptorProto nested_type */
            nested_type?: (google.protobuf.DescriptorProto[]|null);

            /** DescriptorProto enum_type */
            enum_type?: (google.protobuf.EnumDescriptorProto[]|null);

            /** DescriptorProto extension_range */
            extension_range?: (google.protobuf.DescriptorProto.ExtensionRange[]|null);

            /** DescriptorProto oneof_decl */
            oneof_decl?: (google.protobuf.OneofDescriptorProto[]|null);

            /** DescriptorProto options */
            options?: (google.protobuf.MessageOptions|null);

            /** DescriptorProto reserved_range */
            reserved_range?: (google.protobuf.DescriptorProto.ReservedRange[]|null);

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
            public field: google.protobuf.FieldDescriptorProto[];

            /** DescriptorProto extension. */
            public extension: google.protobuf.FieldDescriptorProto[];

            /** DescriptorProto nested_type. */
            public nested_type: google.protobuf.DescriptorProto[];

            /** DescriptorProto enum_type. */
            public enum_type: google.protobuf.EnumDescriptorProto[];

            /** DescriptorProto extension_range. */
            public extension_range: google.protobuf.DescriptorProto.ExtensionRange[];

            /** DescriptorProto oneof_decl. */
            public oneof_decl: google.protobuf.OneofDescriptorProto[];

            /** DescriptorProto options. */
            public options?: (google.protobuf.MessageOptions|null);

            /** DescriptorProto reserved_range. */
            public reserved_range: google.protobuf.DescriptorProto.ReservedRange[];

            /** DescriptorProto reserved_name. */
            public reserved_name: string[];

            /**
             * Encodes the specified DescriptorProto message. Does not implicitly {@link google.protobuf.DescriptorProto.verify|verify} messages.
             * @param message DescriptorProto message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: google.protobuf.DescriptorProto, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a DescriptorProto message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns DescriptorProto
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): google.protobuf.DescriptorProto;
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

                /**
                 * Encodes the specified ExtensionRange message. Does not implicitly {@link google.protobuf.DescriptorProto.ExtensionRange.verify|verify} messages.
                 * @param message ExtensionRange message or plain object to encode
                 * @param [writer] Writer to encode to
                 * @returns Writer
                 */
                public static encode(message: google.protobuf.DescriptorProto.ExtensionRange, writer?: $protobuf.Writer): $protobuf.Writer;

                /**
                 * Decodes an ExtensionRange message from the specified reader or buffer.
                 * @param reader Reader or buffer to decode from
                 * @param [length] Message length if known beforehand
                 * @returns ExtensionRange
                 * @throws {Error} If the payload is not a reader or valid buffer
                 * @throws {$protobuf.util.ProtocolError} If required fields are missing
                 */
                public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): google.protobuf.DescriptorProto.ExtensionRange;
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

                /**
                 * Encodes the specified ReservedRange message. Does not implicitly {@link google.protobuf.DescriptorProto.ReservedRange.verify|verify} messages.
                 * @param message ReservedRange message or plain object to encode
                 * @param [writer] Writer to encode to
                 * @returns Writer
                 */
                public static encode(message: google.protobuf.DescriptorProto.ReservedRange, writer?: $protobuf.Writer): $protobuf.Writer;

                /**
                 * Decodes a ReservedRange message from the specified reader or buffer.
                 * @param reader Reader or buffer to decode from
                 * @param [length] Message length if known beforehand
                 * @returns ReservedRange
                 * @throws {Error} If the payload is not a reader or valid buffer
                 * @throws {$protobuf.util.ProtocolError} If required fields are missing
                 */
                public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): google.protobuf.DescriptorProto.ReservedRange;
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
            options?: (google.protobuf.FieldOptions|null);
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
            public options?: (google.protobuf.FieldOptions|null);

            /**
             * Encodes the specified FieldDescriptorProto message. Does not implicitly {@link google.protobuf.FieldDescriptorProto.verify|verify} messages.
             * @param message FieldDescriptorProto message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: google.protobuf.FieldDescriptorProto, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a FieldDescriptorProto message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns FieldDescriptorProto
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): google.protobuf.FieldDescriptorProto;
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
            options?: (google.protobuf.OneofOptions|null);
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
            public options?: (google.protobuf.OneofOptions|null);

            /**
             * Encodes the specified OneofDescriptorProto message. Does not implicitly {@link google.protobuf.OneofDescriptorProto.verify|verify} messages.
             * @param message OneofDescriptorProto message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: google.protobuf.OneofDescriptorProto, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an OneofDescriptorProto message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns OneofDescriptorProto
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): google.protobuf.OneofDescriptorProto;
        }

        /** Properties of an EnumDescriptorProto. */
        interface IEnumDescriptorProto {

            /** EnumDescriptorProto name */
            name?: (string|null);

            /** EnumDescriptorProto value */
            value?: (google.protobuf.EnumValueDescriptorProto[]|null);

            /** EnumDescriptorProto options */
            options?: (google.protobuf.EnumOptions|null);
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
            public value: google.protobuf.EnumValueDescriptorProto[];

            /** EnumDescriptorProto options. */
            public options?: (google.protobuf.EnumOptions|null);

            /**
             * Encodes the specified EnumDescriptorProto message. Does not implicitly {@link google.protobuf.EnumDescriptorProto.verify|verify} messages.
             * @param message EnumDescriptorProto message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: google.protobuf.EnumDescriptorProto, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an EnumDescriptorProto message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns EnumDescriptorProto
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): google.protobuf.EnumDescriptorProto;
        }

        /** Properties of an EnumValueDescriptorProto. */
        interface IEnumValueDescriptorProto {

            /** EnumValueDescriptorProto name */
            name?: (string|null);

            /** EnumValueDescriptorProto number */
            number?: (number|null);

            /** EnumValueDescriptorProto options */
            options?: (google.protobuf.EnumValueOptions|null);
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
            public options?: (google.protobuf.EnumValueOptions|null);

            /**
             * Encodes the specified EnumValueDescriptorProto message. Does not implicitly {@link google.protobuf.EnumValueDescriptorProto.verify|verify} messages.
             * @param message EnumValueDescriptorProto message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: google.protobuf.EnumValueDescriptorProto, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an EnumValueDescriptorProto message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns EnumValueDescriptorProto
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): google.protobuf.EnumValueDescriptorProto;
        }

        /** Properties of a ServiceDescriptorProto. */
        interface IServiceDescriptorProto {

            /** ServiceDescriptorProto name */
            name?: (string|null);

            /** ServiceDescriptorProto method */
            method?: (google.protobuf.MethodDescriptorProto[]|null);

            /** ServiceDescriptorProto options */
            options?: (google.protobuf.ServiceOptions|null);
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
            public method: google.protobuf.MethodDescriptorProto[];

            /** ServiceDescriptorProto options. */
            public options?: (google.protobuf.ServiceOptions|null);

            /**
             * Encodes the specified ServiceDescriptorProto message. Does not implicitly {@link google.protobuf.ServiceDescriptorProto.verify|verify} messages.
             * @param message ServiceDescriptorProto message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: google.protobuf.ServiceDescriptorProto, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a ServiceDescriptorProto message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns ServiceDescriptorProto
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): google.protobuf.ServiceDescriptorProto;
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
            options?: (google.protobuf.MethodOptions|null);

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
            public options?: (google.protobuf.MethodOptions|null);

            /** MethodDescriptorProto client_streaming. */
            public client_streaming: boolean;

            /** MethodDescriptorProto server_streaming. */
            public server_streaming: boolean;

            /**
             * Encodes the specified MethodDescriptorProto message. Does not implicitly {@link google.protobuf.MethodDescriptorProto.verify|verify} messages.
             * @param message MethodDescriptorProto message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: google.protobuf.MethodDescriptorProto, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a MethodDescriptorProto message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns MethodDescriptorProto
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): google.protobuf.MethodDescriptorProto;
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
            uninterpreted_option?: (google.protobuf.UninterpretedOption[]|null);
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
            public uninterpreted_option: google.protobuf.UninterpretedOption[];

            /**
             * Encodes the specified FileOptions message. Does not implicitly {@link google.protobuf.FileOptions.verify|verify} messages.
             * @param message FileOptions message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: google.protobuf.FileOptions, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a FileOptions message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns FileOptions
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): google.protobuf.FileOptions;
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
            uninterpreted_option?: (google.protobuf.UninterpretedOption[]|null);
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
            public uninterpreted_option: google.protobuf.UninterpretedOption[];

            /**
             * Encodes the specified MessageOptions message. Does not implicitly {@link google.protobuf.MessageOptions.verify|verify} messages.
             * @param message MessageOptions message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: google.protobuf.MessageOptions, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a MessageOptions message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns MessageOptions
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): google.protobuf.MessageOptions;
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
            uninterpreted_option?: (google.protobuf.UninterpretedOption[]|null);
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
            public uninterpreted_option: google.protobuf.UninterpretedOption[];

            /**
             * Encodes the specified FieldOptions message. Does not implicitly {@link google.protobuf.FieldOptions.verify|verify} messages.
             * @param message FieldOptions message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: google.protobuf.FieldOptions, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a FieldOptions message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns FieldOptions
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): google.protobuf.FieldOptions;
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
            uninterpreted_option?: (google.protobuf.UninterpretedOption[]|null);
        }

        /** Represents an OneofOptions. */
        class OneofOptions implements IOneofOptions {

            /**
             * Constructs a new OneofOptions.
             * @param [properties] Properties to set
             */
            constructor(properties?: google.protobuf.IOneofOptions);

            /** OneofOptions uninterpreted_option. */
            public uninterpreted_option: google.protobuf.UninterpretedOption[];

            /**
             * Encodes the specified OneofOptions message. Does not implicitly {@link google.protobuf.OneofOptions.verify|verify} messages.
             * @param message OneofOptions message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: google.protobuf.OneofOptions, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an OneofOptions message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns OneofOptions
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): google.protobuf.OneofOptions;
        }

        /** Properties of an EnumOptions. */
        interface IEnumOptions {

            /** EnumOptions allow_alias */
            allow_alias?: (boolean|null);

            /** EnumOptions deprecated */
            deprecated?: (boolean|null);

            /** EnumOptions uninterpreted_option */
            uninterpreted_option?: (google.protobuf.UninterpretedOption[]|null);
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
            public uninterpreted_option: google.protobuf.UninterpretedOption[];

            /**
             * Encodes the specified EnumOptions message. Does not implicitly {@link google.protobuf.EnumOptions.verify|verify} messages.
             * @param message EnumOptions message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: google.protobuf.EnumOptions, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an EnumOptions message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns EnumOptions
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): google.protobuf.EnumOptions;
        }

        /** Properties of an EnumValueOptions. */
        interface IEnumValueOptions {

            /** EnumValueOptions deprecated */
            deprecated?: (boolean|null);

            /** EnumValueOptions uninterpreted_option */
            uninterpreted_option?: (google.protobuf.UninterpretedOption[]|null);
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
            public uninterpreted_option: google.protobuf.UninterpretedOption[];

            /**
             * Encodes the specified EnumValueOptions message. Does not implicitly {@link google.protobuf.EnumValueOptions.verify|verify} messages.
             * @param message EnumValueOptions message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: google.protobuf.EnumValueOptions, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an EnumValueOptions message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns EnumValueOptions
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): google.protobuf.EnumValueOptions;
        }

        /** Properties of a ServiceOptions. */
        interface IServiceOptions {

            /** ServiceOptions deprecated */
            deprecated?: (boolean|null);

            /** ServiceOptions uninterpreted_option */
            uninterpreted_option?: (google.protobuf.UninterpretedOption[]|null);
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
            public uninterpreted_option: google.protobuf.UninterpretedOption[];

            /**
             * Encodes the specified ServiceOptions message. Does not implicitly {@link google.protobuf.ServiceOptions.verify|verify} messages.
             * @param message ServiceOptions message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: google.protobuf.ServiceOptions, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a ServiceOptions message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns ServiceOptions
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): google.protobuf.ServiceOptions;
        }

        /** Properties of a MethodOptions. */
        interface IMethodOptions {

            /** MethodOptions deprecated */
            deprecated?: (boolean|null);

            /** MethodOptions uninterpreted_option */
            uninterpreted_option?: (google.protobuf.UninterpretedOption[]|null);

            /** MethodOptions .google.api.http */
            ".google.api.http"?: (google.api.HttpRule|null);
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
            public uninterpreted_option: google.protobuf.UninterpretedOption[];

            /**
             * Encodes the specified MethodOptions message. Does not implicitly {@link google.protobuf.MethodOptions.verify|verify} messages.
             * @param message MethodOptions message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: google.protobuf.MethodOptions, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a MethodOptions message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns MethodOptions
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): google.protobuf.MethodOptions;
        }

        /** Properties of an UninterpretedOption. */
        interface IUninterpretedOption {

            /** UninterpretedOption name */
            name?: (google.protobuf.UninterpretedOption.NamePart[]|null);

            /** UninterpretedOption identifier_value */
            identifier_value?: (string|null);

            /** UninterpretedOption positive_int_value */
            positive_int_value?: (number|null);

            /** UninterpretedOption negative_int_value */
            negative_int_value?: (number|null);

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
            public name: google.protobuf.UninterpretedOption.NamePart[];

            /** UninterpretedOption identifier_value. */
            public identifier_value: string;

            /** UninterpretedOption positive_int_value. */
            public positive_int_value: number;

            /** UninterpretedOption negative_int_value. */
            public negative_int_value: number;

            /** UninterpretedOption double_value. */
            public double_value: number;

            /** UninterpretedOption string_value. */
            public string_value: Uint8Array;

            /** UninterpretedOption aggregate_value. */
            public aggregate_value: string;

            /**
             * Encodes the specified UninterpretedOption message. Does not implicitly {@link google.protobuf.UninterpretedOption.verify|verify} messages.
             * @param message UninterpretedOption message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: google.protobuf.UninterpretedOption, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes an UninterpretedOption message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns UninterpretedOption
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): google.protobuf.UninterpretedOption;
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

                /**
                 * Encodes the specified NamePart message. Does not implicitly {@link google.protobuf.UninterpretedOption.NamePart.verify|verify} messages.
                 * @param message NamePart message or plain object to encode
                 * @param [writer] Writer to encode to
                 * @returns Writer
                 */
                public static encode(message: google.protobuf.UninterpretedOption.NamePart, writer?: $protobuf.Writer): $protobuf.Writer;

                /**
                 * Decodes a NamePart message from the specified reader or buffer.
                 * @param reader Reader or buffer to decode from
                 * @param [length] Message length if known beforehand
                 * @returns NamePart
                 * @throws {Error} If the payload is not a reader or valid buffer
                 * @throws {$protobuf.util.ProtocolError} If required fields are missing
                 */
                public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): google.protobuf.UninterpretedOption.NamePart;
            }
        }

        /** Properties of a SourceCodeInfo. */
        interface ISourceCodeInfo {

            /** SourceCodeInfo location */
            location?: (google.protobuf.SourceCodeInfo.Location[]|null);
        }

        /** Represents a SourceCodeInfo. */
        class SourceCodeInfo implements ISourceCodeInfo {

            /**
             * Constructs a new SourceCodeInfo.
             * @param [properties] Properties to set
             */
            constructor(properties?: google.protobuf.ISourceCodeInfo);

            /** SourceCodeInfo location. */
            public location: google.protobuf.SourceCodeInfo.Location[];

            /**
             * Encodes the specified SourceCodeInfo message. Does not implicitly {@link google.protobuf.SourceCodeInfo.verify|verify} messages.
             * @param message SourceCodeInfo message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: google.protobuf.SourceCodeInfo, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a SourceCodeInfo message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns SourceCodeInfo
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): google.protobuf.SourceCodeInfo;
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

                /**
                 * Encodes the specified Location message. Does not implicitly {@link google.protobuf.SourceCodeInfo.Location.verify|verify} messages.
                 * @param message Location message or plain object to encode
                 * @param [writer] Writer to encode to
                 * @returns Writer
                 */
                public static encode(message: google.protobuf.SourceCodeInfo.Location, writer?: $protobuf.Writer): $protobuf.Writer;

                /**
                 * Decodes a Location message from the specified reader or buffer.
                 * @param reader Reader or buffer to decode from
                 * @param [length] Message length if known beforehand
                 * @returns Location
                 * @throws {Error} If the payload is not a reader or valid buffer
                 * @throws {$protobuf.util.ProtocolError} If required fields are missing
                 */
                public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): google.protobuf.SourceCodeInfo.Location;
            }
        }

        /** Properties of a GeneratedCodeInfo. */
        interface IGeneratedCodeInfo {

            /** GeneratedCodeInfo annotation */
            annotation?: (google.protobuf.GeneratedCodeInfo.Annotation[]|null);
        }

        /** Represents a GeneratedCodeInfo. */
        class GeneratedCodeInfo implements IGeneratedCodeInfo {

            /**
             * Constructs a new GeneratedCodeInfo.
             * @param [properties] Properties to set
             */
            constructor(properties?: google.protobuf.IGeneratedCodeInfo);

            /** GeneratedCodeInfo annotation. */
            public annotation: google.protobuf.GeneratedCodeInfo.Annotation[];

            /**
             * Encodes the specified GeneratedCodeInfo message. Does not implicitly {@link google.protobuf.GeneratedCodeInfo.verify|verify} messages.
             * @param message GeneratedCodeInfo message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: google.protobuf.GeneratedCodeInfo, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a GeneratedCodeInfo message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns GeneratedCodeInfo
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): google.protobuf.GeneratedCodeInfo;
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

                /**
                 * Encodes the specified Annotation message. Does not implicitly {@link google.protobuf.GeneratedCodeInfo.Annotation.verify|verify} messages.
                 * @param message Annotation message or plain object to encode
                 * @param [writer] Writer to encode to
                 * @returns Writer
                 */
                public static encode(message: google.protobuf.GeneratedCodeInfo.Annotation, writer?: $protobuf.Writer): $protobuf.Writer;

                /**
                 * Decodes an Annotation message from the specified reader or buffer.
                 * @param reader Reader or buffer to decode from
                 * @param [length] Message length if known beforehand
                 * @returns Annotation
                 * @throws {Error} If the payload is not a reader or valid buffer
                 * @throws {$protobuf.util.ProtocolError} If required fields are missing
                 */
                public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): google.protobuf.GeneratedCodeInfo.Annotation;
            }
        }

        /** Properties of a Timestamp. */
        interface ITimestamp {

            /** Timestamp seconds */
            seconds?: (number|null);

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
            public seconds: number;

            /** Timestamp nanos. */
            public nanos: number;

            /**
             * Encodes the specified Timestamp message. Does not implicitly {@link google.protobuf.Timestamp.verify|verify} messages.
             * @param message Timestamp message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: google.protobuf.Timestamp, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a Timestamp message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns Timestamp
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): google.protobuf.Timestamp;
        }
    }

    /** Namespace api. */
    namespace api {

        /** Properties of a Http. */
        interface IHttp {

            /** Http rules */
            rules?: (google.api.HttpRule[]|null);
        }

        /** Represents a Http. */
        class Http implements IHttp {

            /**
             * Constructs a new Http.
             * @param [properties] Properties to set
             */
            constructor(properties?: google.api.IHttp);

            /** Http rules. */
            public rules: google.api.HttpRule[];

            /**
             * Encodes the specified Http message. Does not implicitly {@link google.api.Http.verify|verify} messages.
             * @param message Http message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: google.api.Http, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a Http message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns Http
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): google.api.Http;
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
            custom?: (google.api.CustomHttpPattern|null);

            /** HttpRule selector */
            selector?: (string|null);

            /** HttpRule body */
            body?: (string|null);

            /** HttpRule additional_bindings */
            additional_bindings?: (google.api.HttpRule[]|null);
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
            public custom?: (google.api.CustomHttpPattern|null);

            /** HttpRule selector. */
            public selector: string;

            /** HttpRule body. */
            public body: string;

            /** HttpRule additional_bindings. */
            public additional_bindings: google.api.HttpRule[];

            /** HttpRule pattern. */
            public pattern?: ("get"|"put"|"post"|"delete"|"patch"|"custom");

            /**
             * Encodes the specified HttpRule message. Does not implicitly {@link google.api.HttpRule.verify|verify} messages.
             * @param message HttpRule message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: google.api.HttpRule, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a HttpRule message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns HttpRule
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): google.api.HttpRule;
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

            /**
             * Encodes the specified CustomHttpPattern message. Does not implicitly {@link google.api.CustomHttpPattern.verify|verify} messages.
             * @param message CustomHttpPattern message or plain object to encode
             * @param [writer] Writer to encode to
             * @returns Writer
             */
            public static encode(message: google.api.CustomHttpPattern, writer?: $protobuf.Writer): $protobuf.Writer;

            /**
             * Decodes a CustomHttpPattern message from the specified reader or buffer.
             * @param reader Reader or buffer to decode from
             * @param [length] Message length if known beforehand
             * @returns CustomHttpPattern
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): google.api.CustomHttpPattern;
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
    public namespace_id: number;

    /**
     * Encodes the specified NamespaceID message. Does not implicitly {@link NamespaceID.verify|verify} messages.
     * @param message NamespaceID message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: NamespaceID, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a NamespaceID message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns NamespaceID
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): NamespaceID;
}

/** Represents a NamespaceResponse. */
export class NamespaceResponse implements INamespaceResponse {

    /**
     * Constructs a new NamespaceResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: INamespaceResponse);

    /** NamespaceResponse id. */
    public id: number;

    /** NamespaceResponse name. */
    public name: string;

    /** NamespaceResponse image_pull_secrets. */
    public image_pull_secrets: string[];

    /** NamespaceResponse created_at. */
    public created_at?: (google.protobuf.Timestamp|null);

    /** NamespaceResponse updated_at. */
    public updated_at?: (google.protobuf.Timestamp|null);

    /** NamespaceResponse deleted_at. */
    public deleted_at?: (google.protobuf.Timestamp|null);

    /**
     * Encodes the specified NamespaceResponse message. Does not implicitly {@link NamespaceResponse.verify|verify} messages.
     * @param message NamespaceResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: NamespaceResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a NamespaceResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns NamespaceResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): NamespaceResponse;
}

/** Represents a NamespaceItem. */
export class NamespaceItem implements INamespaceItem {

    /**
     * Constructs a new NamespaceItem.
     * @param [properties] Properties to set
     */
    constructor(properties?: INamespaceItem);

    /** NamespaceItem id. */
    public id: number;

    /** NamespaceItem name. */
    public name: string;

    /** NamespaceItem created_at. */
    public created_at?: (google.protobuf.Timestamp|null);

    /** NamespaceItem updated_at. */
    public updated_at?: (google.protobuf.Timestamp|null);

    /** NamespaceItem projects. */
    public projects: NamespaceItem.SimpleProjectItem[];

    /**
     * Encodes the specified NamespaceItem message. Does not implicitly {@link NamespaceItem.verify|verify} messages.
     * @param message NamespaceItem message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: NamespaceItem, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a NamespaceItem message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns NamespaceItem
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): NamespaceItem;
}

export namespace NamespaceItem {

    /** Properties of a SimpleProjectItem. */
    interface ISimpleProjectItem {

        /** SimpleProjectItem id */
        id?: (number|null);

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
        public id: number;

        /** SimpleProjectItem name. */
        public name: string;

        /** SimpleProjectItem status. */
        public status: string;

        /**
         * Encodes the specified SimpleProjectItem message. Does not implicitly {@link NamespaceItem.SimpleProjectItem.verify|verify} messages.
         * @param message SimpleProjectItem message or plain object to encode
         * @param [writer] Writer to encode to
         * @returns Writer
         */
        public static encode(message: NamespaceItem.SimpleProjectItem, writer?: $protobuf.Writer): $protobuf.Writer;

        /**
         * Decodes a SimpleProjectItem message from the specified reader or buffer.
         * @param reader Reader or buffer to decode from
         * @param [length] Message length if known beforehand
         * @returns SimpleProjectItem
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): NamespaceItem.SimpleProjectItem;
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
    public data: NamespaceItem[];

    /**
     * Encodes the specified NamespaceList message. Does not implicitly {@link NamespaceList.verify|verify} messages.
     * @param message NamespaceList message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: NamespaceList, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a NamespaceList message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns NamespaceList
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): NamespaceList;
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

    /**
     * Encodes the specified NsStoreRequest message. Does not implicitly {@link NsStoreRequest.verify|verify} messages.
     * @param message NsStoreRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: NsStoreRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a NsStoreRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns NsStoreRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): NsStoreRequest;
}

/** Represents a NsStoreResponse. */
export class NsStoreResponse implements INsStoreResponse {

    /**
     * Constructs a new NsStoreResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: INsStoreResponse);

    /** NsStoreResponse data. */
    public data?: (NamespaceResponse|null);

    /**
     * Encodes the specified NsStoreResponse message. Does not implicitly {@link NsStoreResponse.verify|verify} messages.
     * @param message NsStoreResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: NsStoreResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a NsStoreResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns NsStoreResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): NsStoreResponse;
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

    /**
     * Encodes the specified CpuAndMemoryResponse message. Does not implicitly {@link CpuAndMemoryResponse.verify|verify} messages.
     * @param message CpuAndMemoryResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: CpuAndMemoryResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a CpuAndMemoryResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns CpuAndMemoryResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): CpuAndMemoryResponse;
}

/** Represents a ServiceEndpointsResponse. */
export class ServiceEndpointsResponse implements IServiceEndpointsResponse {

    /**
     * Constructs a new ServiceEndpointsResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IServiceEndpointsResponse);

    /** ServiceEndpointsResponse data. */
    public data: ServiceEndpointsResponse.item[];

    /**
     * Encodes the specified ServiceEndpointsResponse message. Does not implicitly {@link ServiceEndpointsResponse.verify|verify} messages.
     * @param message ServiceEndpointsResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: ServiceEndpointsResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a ServiceEndpointsResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns ServiceEndpointsResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): ServiceEndpointsResponse;
}

export namespace ServiceEndpointsResponse {

    /** Properties of an item. */
    interface Iitem {

        /** item name */
        name?: (string|null);

        /** item url */
        url?: (string[]|null);
    }

    /** Represents an item. */
    class item implements Iitem {

        /**
         * Constructs a new item.
         * @param [properties] Properties to set
         */
        constructor(properties?: ServiceEndpointsResponse.Iitem);

        /** item name. */
        public name: string;

        /** item url. */
        public url: string[];

        /**
         * Encodes the specified item message. Does not implicitly {@link ServiceEndpointsResponse.item.verify|verify} messages.
         * @param message item message or plain object to encode
         * @param [writer] Writer to encode to
         * @returns Writer
         */
        public static encode(message: ServiceEndpointsResponse.item, writer?: $protobuf.Writer): $protobuf.Writer;

        /**
         * Decodes an item message from the specified reader or buffer.
         * @param reader Reader or buffer to decode from
         * @param [length] Message length if known beforehand
         * @returns item
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): ServiceEndpointsResponse.item;
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
    public namespace_id: number;

    /** ServiceEndpointsRequest project_name. */
    public project_name: string;

    /**
     * Encodes the specified ServiceEndpointsRequest message. Does not implicitly {@link ServiceEndpointsRequest.verify|verify} messages.
     * @param message ServiceEndpointsRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: ServiceEndpointsRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a ServiceEndpointsRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns ServiceEndpointsRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): ServiceEndpointsRequest;
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
    public index(request: google.protobuf.Empty, callback: Namespace.IndexCallback): void;

    /**
     * Calls Index.
     * @param request Empty message or plain object
     * @returns Promise
     */
    public index(request: google.protobuf.Empty): Promise<NamespaceList>;

    /**
     * Calls Store.
     * @param request NsStoreRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and NsStoreResponse
     */
    public store(request: NsStoreRequest, callback: Namespace.StoreCallback): void;

    /**
     * Calls Store.
     * @param request NsStoreRequest message or plain object
     * @returns Promise
     */
    public store(request: NsStoreRequest): Promise<NsStoreResponse>;

    /**
     * Calls CpuAndMemory.
     * @param request NamespaceID message or plain object
     * @param callback Node-style callback called with the error, if any, and CpuAndMemoryResponse
     */
    public cpuAndMemory(request: NamespaceID, callback: Namespace.CpuAndMemoryCallback): void;

    /**
     * Calls CpuAndMemory.
     * @param request NamespaceID message or plain object
     * @returns Promise
     */
    public cpuAndMemory(request: NamespaceID): Promise<CpuAndMemoryResponse>;

    /**
     * Calls ServiceEndpoints.
     * @param request ServiceEndpointsRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and ServiceEndpointsResponse
     */
    public serviceEndpoints(request: ServiceEndpointsRequest, callback: Namespace.ServiceEndpointsCallback): void;

    /**
     * Calls ServiceEndpoints.
     * @param request ServiceEndpointsRequest message or plain object
     * @returns Promise
     */
    public serviceEndpoints(request: ServiceEndpointsRequest): Promise<ServiceEndpointsResponse>;

    /**
     * Calls Destroy.
     * @param request NamespaceID message or plain object
     * @param callback Node-style callback called with the error, if any, and Empty
     */
    public destroy(request: NamespaceID, callback: Namespace.DestroyCallback): void;

    /**
     * Calls Destroy.
     * @param request NamespaceID message or plain object
     * @returns Promise
     */
    public destroy(request: NamespaceID): Promise<google.protobuf.Empty>;
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

/** Represents a BackgroundRequest. */
export class BackgroundRequest implements IBackgroundRequest {

    /**
     * Constructs a new BackgroundRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IBackgroundRequest);

    /** BackgroundRequest random. */
    public random: boolean;

    /**
     * Encodes the specified BackgroundRequest message. Does not implicitly {@link BackgroundRequest.verify|verify} messages.
     * @param message BackgroundRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: BackgroundRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a BackgroundRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns BackgroundRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): BackgroundRequest;
}

/** Represents a BackgroundResponse. */
export class BackgroundResponse implements IBackgroundResponse {

    /**
     * Constructs a new BackgroundResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IBackgroundResponse);

    /** BackgroundResponse url. */
    public url: string;

    /** BackgroundResponse copyright. */
    public copyright: string;

    /**
     * Encodes the specified BackgroundResponse message. Does not implicitly {@link BackgroundResponse.verify|verify} messages.
     * @param message BackgroundResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: BackgroundResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a BackgroundResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns BackgroundResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): BackgroundResponse;
}

/** Represents a Picture */
export class Picture extends $protobuf.rpc.Service {

    /**
     * Constructs a new Picture service.
     * @param rpcImpl RPC implementation
     * @param [requestDelimited=false] Whether requests are length-delimited
     * @param [responseDelimited=false] Whether responses are length-delimited
     */
    constructor(rpcImpl: $protobuf.RPCImpl, requestDelimited?: boolean, responseDelimited?: boolean);

    /**
     * Calls Background.
     * @param request BackgroundRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and BackgroundResponse
     */
    public background(request: BackgroundRequest, callback: Picture.BackgroundCallback): void;

    /**
     * Calls Background.
     * @param request BackgroundRequest message or plain object
     * @returns Promise
     */
    public background(request: BackgroundRequest): Promise<BackgroundResponse>;
}

export namespace Picture {

    /**
     * Callback as used by {@link Picture#background}.
     * @param error Error, if any
     * @param [response] BackgroundResponse
     */
    type BackgroundCallback = (error: (Error|null), response?: BackgroundResponse) => void;
}

/** Represents a ProjectDestroyRequest. */
export class ProjectDestroyRequest implements IProjectDestroyRequest {

    /**
     * Constructs a new ProjectDestroyRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IProjectDestroyRequest);

    /** ProjectDestroyRequest namespace_id. */
    public namespace_id: number;

    /** ProjectDestroyRequest project_id. */
    public project_id: number;

    /**
     * Encodes the specified ProjectDestroyRequest message. Does not implicitly {@link ProjectDestroyRequest.verify|verify} messages.
     * @param message ProjectDestroyRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: ProjectDestroyRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a ProjectDestroyRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns ProjectDestroyRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): ProjectDestroyRequest;
}

/** Represents a ProjectShowRequest. */
export class ProjectShowRequest implements IProjectShowRequest {

    /**
     * Constructs a new ProjectShowRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IProjectShowRequest);

    /** ProjectShowRequest namespace_id. */
    public namespace_id: number;

    /** ProjectShowRequest project_id. */
    public project_id: number;

    /**
     * Encodes the specified ProjectShowRequest message. Does not implicitly {@link ProjectShowRequest.verify|verify} messages.
     * @param message ProjectShowRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: ProjectShowRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a ProjectShowRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns ProjectShowRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): ProjectShowRequest;
}

/** Represents a ProjectShowResponse. */
export class ProjectShowResponse implements IProjectShowResponse {

    /**
     * Constructs a new ProjectShowResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IProjectShowResponse);

    /** ProjectShowResponse id. */
    public id: number;

    /** ProjectShowResponse name. */
    public name: string;

    /** ProjectShowResponse gitlab_project_id. */
    public gitlab_project_id: number;

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
    public namespace?: (ProjectShowResponse.Namespace|null);

    /** ProjectShowResponse cpu. */
    public cpu: string;

    /** ProjectShowResponse memory. */
    public memory: string;

    /** ProjectShowResponse override_values. */
    public override_values: string;

    /** ProjectShowResponse created_at. */
    public created_at: string;

    /** ProjectShowResponse updated_at. */
    public updated_at: string;

    /**
     * Encodes the specified ProjectShowResponse message. Does not implicitly {@link ProjectShowResponse.verify|verify} messages.
     * @param message ProjectShowResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: ProjectShowResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a ProjectShowResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns ProjectShowResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): ProjectShowResponse;
}

export namespace ProjectShowResponse {

    /** Properties of a Namespace. */
    interface INamespace {

        /** Namespace id */
        id?: (number|null);

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
        public id: number;

        /** Namespace name. */
        public name: string;

        /**
         * Encodes the specified Namespace message. Does not implicitly {@link ProjectShowResponse.Namespace.verify|verify} messages.
         * @param message Namespace message or plain object to encode
         * @param [writer] Writer to encode to
         * @returns Writer
         */
        public static encode(message: ProjectShowResponse.Namespace, writer?: $protobuf.Writer): $protobuf.Writer;

        /**
         * Decodes a Namespace message from the specified reader or buffer.
         * @param reader Reader or buffer to decode from
         * @param [length] Message length if known beforehand
         * @returns Namespace
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): ProjectShowResponse.Namespace;
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
    public namespace_id: number;

    /** AllPodContainersRequest project_id. */
    public project_id: number;

    /**
     * Encodes the specified AllPodContainersRequest message. Does not implicitly {@link AllPodContainersRequest.verify|verify} messages.
     * @param message AllPodContainersRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: AllPodContainersRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes an AllPodContainersRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns AllPodContainersRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): AllPodContainersRequest;
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

    /**
     * Encodes the specified PodLog message. Does not implicitly {@link PodLog.verify|verify} messages.
     * @param message PodLog message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: PodLog, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a PodLog message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns PodLog
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): PodLog;
}

/** Represents an AllPodContainersResponse. */
export class AllPodContainersResponse implements IAllPodContainersResponse {

    /**
     * Constructs a new AllPodContainersResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IAllPodContainersResponse);

    /** AllPodContainersResponse data. */
    public data: PodLog[];

    /**
     * Encodes the specified AllPodContainersResponse message. Does not implicitly {@link AllPodContainersResponse.verify|verify} messages.
     * @param message AllPodContainersResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: AllPodContainersResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes an AllPodContainersResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns AllPodContainersResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): AllPodContainersResponse;
}

/** Represents a PodContainerLogRequest. */
export class PodContainerLogRequest implements IPodContainerLogRequest {

    /**
     * Constructs a new PodContainerLogRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IPodContainerLogRequest);

    /** PodContainerLogRequest namespace_id. */
    public namespace_id: number;

    /** PodContainerLogRequest project_id. */
    public project_id: number;

    /** PodContainerLogRequest pod. */
    public pod: string;

    /** PodContainerLogRequest container. */
    public container: string;

    /**
     * Encodes the specified PodContainerLogRequest message. Does not implicitly {@link PodContainerLogRequest.verify|verify} messages.
     * @param message PodContainerLogRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: PodContainerLogRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a PodContainerLogRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns PodContainerLogRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): PodContainerLogRequest;
}

/** Represents a PodContainerLogResponse. */
export class PodContainerLogResponse implements IPodContainerLogResponse {

    /**
     * Constructs a new PodContainerLogResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IPodContainerLogResponse);

    /** PodContainerLogResponse data. */
    public data?: (PodLog|null);

    /**
     * Encodes the specified PodContainerLogResponse message. Does not implicitly {@link PodContainerLogResponse.verify|verify} messages.
     * @param message PodContainerLogResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: PodContainerLogResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a PodContainerLogResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns PodContainerLogResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): PodContainerLogResponse;
}

/** Represents an IsPodRunningRequest. */
export class IsPodRunningRequest implements IIsPodRunningRequest {

    /**
     * Constructs a new IsPodRunningRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IIsPodRunningRequest);

    /** IsPodRunningRequest namespace. */
    public namespace: string;

    /** IsPodRunningRequest pod. */
    public pod: string;

    /**
     * Encodes the specified IsPodRunningRequest message. Does not implicitly {@link IsPodRunningRequest.verify|verify} messages.
     * @param message IsPodRunningRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: IsPodRunningRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes an IsPodRunningRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns IsPodRunningRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): IsPodRunningRequest;
}

/** Represents an IsPodRunningResponse. */
export class IsPodRunningResponse implements IIsPodRunningResponse {

    /**
     * Constructs a new IsPodRunningResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IIsPodRunningResponse);

    /** IsPodRunningResponse running. */
    public running: boolean;

    /** IsPodRunningResponse reason. */
    public reason: string;

    /**
     * Encodes the specified IsPodRunningResponse message. Does not implicitly {@link IsPodRunningResponse.verify|verify} messages.
     * @param message IsPodRunningResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: IsPodRunningResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes an IsPodRunningResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns IsPodRunningResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): IsPodRunningResponse;
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
    public destroy(request: ProjectDestroyRequest, callback: Project.DestroyCallback): void;

    /**
     * Calls Destroy.
     * @param request ProjectDestroyRequest message or plain object
     * @returns Promise
     */
    public destroy(request: ProjectDestroyRequest): Promise<google.protobuf.Empty>;

    /**
     * Calls Show.
     * @param request ProjectShowRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and ProjectShowResponse
     */
    public show(request: ProjectShowRequest, callback: Project.ShowCallback): void;

    /**
     * Calls Show.
     * @param request ProjectShowRequest message or plain object
     * @returns Promise
     */
    public show(request: ProjectShowRequest): Promise<ProjectShowResponse>;

    /**
     * Calls IsPodRunning.
     * @param request IsPodRunningRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and IsPodRunningResponse
     */
    public isPodRunning(request: IsPodRunningRequest, callback: Project.IsPodRunningCallback): void;

    /**
     * Calls IsPodRunning.
     * @param request IsPodRunningRequest message or plain object
     * @returns Promise
     */
    public isPodRunning(request: IsPodRunningRequest): Promise<IsPodRunningResponse>;

    /**
     * Calls AllPodContainers.
     * @param request AllPodContainersRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and AllPodContainersResponse
     */
    public allPodContainers(request: AllPodContainersRequest, callback: Project.AllPodContainersCallback): void;

    /**
     * Calls AllPodContainers.
     * @param request AllPodContainersRequest message or plain object
     * @returns Promise
     */
    public allPodContainers(request: AllPodContainersRequest): Promise<AllPodContainersResponse>;

    /**
     * Calls PodContainerLog.
     * @param request PodContainerLogRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and PodContainerLogResponse
     */
    public podContainerLog(request: PodContainerLogRequest, callback: Project.PodContainerLogCallback): void;

    /**
     * Calls PodContainerLog.
     * @param request PodContainerLogRequest message or plain object
     * @returns Promise
     */
    public podContainerLog(request: PodContainerLogRequest): Promise<PodContainerLogResponse>;

    /**
     * Calls StreamPodContainerLog.
     * @param request PodContainerLogRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and PodContainerLogResponse
     */
    public streamPodContainerLog(request: PodContainerLogRequest, callback: Project.StreamPodContainerLogCallback): void;

    /**
     * Calls StreamPodContainerLog.
     * @param request PodContainerLogRequest message or plain object
     * @returns Promise
     */
    public streamPodContainerLog(request: PodContainerLogRequest): Promise<PodContainerLogResponse>;
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
     * Callback as used by {@link Project#isPodRunning}.
     * @param error Error, if any
     * @param [response] IsPodRunningResponse
     */
    type IsPodRunningCallback = (error: (Error|null), response?: IsPodRunningResponse) => void;

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

    /**
     * Callback as used by {@link Project#streamPodContainerLog}.
     * @param error Error, if any
     * @param [response] PodContainerLogResponse
     */
    type StreamPodContainerLogCallback = (error: (Error|null), response?: PodContainerLogResponse) => void;
}

/** Represents a VersionResponse. */
export class VersionResponse implements IVersionResponse {

    /**
     * Constructs a new VersionResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IVersionResponse);

    /** VersionResponse Version. */
    public Version: string;

    /** VersionResponse BuildDate. */
    public BuildDate: string;

    /** VersionResponse gitBranch. */
    public gitBranch: string;

    /** VersionResponse GitCommit. */
    public GitCommit: string;

    /** VersionResponse GitTag. */
    public GitTag: string;

    /** VersionResponse GoVersion. */
    public GoVersion: string;

    /** VersionResponse Compiler. */
    public Compiler: string;

    /** VersionResponse Platform. */
    public Platform: string;

    /** VersionResponse KubectlVersion. */
    public KubectlVersion: string;

    /** VersionResponse HelmVersion. */
    public HelmVersion: string;

    /** VersionResponse GitRepo. */
    public GitRepo: string;

    /**
     * Encodes the specified VersionResponse message. Does not implicitly {@link VersionResponse.verify|verify} messages.
     * @param message VersionResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: VersionResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a VersionResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns VersionResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): VersionResponse;
}

/** Represents a Version */
export class Version extends $protobuf.rpc.Service {

    /**
     * Constructs a new Version service.
     * @param rpcImpl RPC implementation
     * @param [requestDelimited=false] Whether requests are length-delimited
     * @param [responseDelimited=false] Whether responses are length-delimited
     */
    constructor(rpcImpl: $protobuf.RPCImpl, requestDelimited?: boolean, responseDelimited?: boolean);

    /**
     * Calls Get.
     * @param request Empty message or plain object
     * @param callback Node-style callback called with the error, if any, and VersionResponse
     */
    public get(request: google.protobuf.Empty, callback: Version.GetCallback): void;

    /**
     * Calls Get.
     * @param request Empty message or plain object
     * @returns Promise
     */
    public get(request: google.protobuf.Empty): Promise<VersionResponse>;
}

export namespace Version {

    /**
     * Callback as used by {@link Version#get}.
     * @param error Error, if any
     * @param [response] VersionResponse
     */
    type GetCallback = (error: (Error|null), response?: VersionResponse) => void;
}

/** Type enum. */
export enum Type {
    TypeUnknown = 0,
    SetUid = 1,
    ReloadProjects = 2,
    CancelProject = 3,
    CreateProject = 4,
    UpdateProject = 5,
    ProcessPercent = 6,
    ClusterInfoSync = 7,
    InternalError = 8,
    HandleExecShell = 9,
    HandleExecShellMsg = 10,
    HandleCloseShell = 11,
    HandleAuthorize = 12
}

/** ResultType enum. */
export enum ResultType {
    ResultUnknown = 0,
    Error = 1,
    Success = 2,
    Deployed = 3,
    DeployedFailed = 4,
    DeployedCanceled = 5
}

/** To enum. */
export enum To {
    ToSelf = 0,
    ToAll = 1,
    ToOthers = 2
}

/** Represents a WsRequestMetadata. */
export class WsRequestMetadata implements IWsRequestMetadata {

    /**
     * Constructs a new WsRequestMetadata.
     * @param [properties] Properties to set
     */
    constructor(properties?: IWsRequestMetadata);

    /** WsRequestMetadata type. */
    public type: Type;

    /**
     * Encodes the specified WsRequestMetadata message. Does not implicitly {@link WsRequestMetadata.verify|verify} messages.
     * @param message WsRequestMetadata message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: WsRequestMetadata, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a WsRequestMetadata message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns WsRequestMetadata
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): WsRequestMetadata;
}

/** Represents an AuthorizeTokenInput. */
export class AuthorizeTokenInput implements IAuthorizeTokenInput {

    /**
     * Constructs a new AuthorizeTokenInput.
     * @param [properties] Properties to set
     */
    constructor(properties?: IAuthorizeTokenInput);

    /** AuthorizeTokenInput type. */
    public type: Type;

    /** AuthorizeTokenInput token. */
    public token: string;

    /**
     * Encodes the specified AuthorizeTokenInput message. Does not implicitly {@link AuthorizeTokenInput.verify|verify} messages.
     * @param message AuthorizeTokenInput message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: AuthorizeTokenInput, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes an AuthorizeTokenInput message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns AuthorizeTokenInput
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): AuthorizeTokenInput;
}

/** Represents a TerminalMessage. */
export class TerminalMessage implements ITerminalMessage {

    /**
     * Constructs a new TerminalMessage.
     * @param [properties] Properties to set
     */
    constructor(properties?: ITerminalMessage);

    /** TerminalMessage op. */
    public op: string;

    /** TerminalMessage data. */
    public data: string;

    /** TerminalMessage session_id. */
    public session_id: string;

    /** TerminalMessage rows. */
    public rows: number;

    /** TerminalMessage cols. */
    public cols: number;

    /**
     * Encodes the specified TerminalMessage message. Does not implicitly {@link TerminalMessage.verify|verify} messages.
     * @param message TerminalMessage message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: TerminalMessage, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a TerminalMessage message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns TerminalMessage
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): TerminalMessage;
}

/** Represents a TerminalMessageInput. */
export class TerminalMessageInput implements ITerminalMessageInput {

    /**
     * Constructs a new TerminalMessageInput.
     * @param [properties] Properties to set
     */
    constructor(properties?: ITerminalMessageInput);

    /** TerminalMessageInput type. */
    public type: Type;

    /** TerminalMessageInput message. */
    public message?: (TerminalMessage|null);

    /**
     * Encodes the specified TerminalMessageInput message. Does not implicitly {@link TerminalMessageInput.verify|verify} messages.
     * @param message TerminalMessageInput message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: TerminalMessageInput, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a TerminalMessageInput message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns TerminalMessageInput
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): TerminalMessageInput;
}

/** Represents a WsHandleExecShellInput. */
export class WsHandleExecShellInput implements IWsHandleExecShellInput {

    /**
     * Constructs a new WsHandleExecShellInput.
     * @param [properties] Properties to set
     */
    constructor(properties?: IWsHandleExecShellInput);

    /** WsHandleExecShellInput type. */
    public type: Type;

    /** WsHandleExecShellInput namespace. */
    public namespace: string;

    /** WsHandleExecShellInput pod. */
    public pod: string;

    /** WsHandleExecShellInput container. */
    public container: string;

    /**
     * Encodes the specified WsHandleExecShellInput message. Does not implicitly {@link WsHandleExecShellInput.verify|verify} messages.
     * @param message WsHandleExecShellInput message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: WsHandleExecShellInput, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a WsHandleExecShellInput message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns WsHandleExecShellInput
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): WsHandleExecShellInput;
}

/** Represents a CancelInput. */
export class CancelInput implements ICancelInput {

    /**
     * Constructs a new CancelInput.
     * @param [properties] Properties to set
     */
    constructor(properties?: ICancelInput);

    /** CancelInput type. */
    public type: Type;

    /** CancelInput namespace_id. */
    public namespace_id: number;

    /** CancelInput name. */
    public name: string;

    /**
     * Encodes the specified CancelInput message. Does not implicitly {@link CancelInput.verify|verify} messages.
     * @param message CancelInput message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: CancelInput, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a CancelInput message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns CancelInput
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): CancelInput;
}

/** Represents a ProjectInput. */
export class ProjectInput implements IProjectInput {

    /**
     * Constructs a new ProjectInput.
     * @param [properties] Properties to set
     */
    constructor(properties?: IProjectInput);

    /** ProjectInput type. */
    public type: Type;

    /** ProjectInput namespace_id. */
    public namespace_id: number;

    /** ProjectInput name. */
    public name: string;

    /** ProjectInput gitlab_project_id. */
    public gitlab_project_id: number;

    /** ProjectInput gitlab_branch. */
    public gitlab_branch: string;

    /** ProjectInput gitlab_commit. */
    public gitlab_commit: string;

    /** ProjectInput config. */
    public config: string;

    /** ProjectInput atomic. */
    public atomic: boolean;

    /**
     * Encodes the specified ProjectInput message. Does not implicitly {@link ProjectInput.verify|verify} messages.
     * @param message ProjectInput message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: ProjectInput, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a ProjectInput message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns ProjectInput
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): ProjectInput;
}

/** Represents an UpdateProjectInput. */
export class UpdateProjectInput implements IUpdateProjectInput {

    /**
     * Constructs a new UpdateProjectInput.
     * @param [properties] Properties to set
     */
    constructor(properties?: IUpdateProjectInput);

    /** UpdateProjectInput type. */
    public type: Type;

    /** UpdateProjectInput project_id. */
    public project_id: number;

    /** UpdateProjectInput gitlab_branch. */
    public gitlab_branch: string;

    /** UpdateProjectInput gitlab_commit. */
    public gitlab_commit: string;

    /** UpdateProjectInput config. */
    public config: string;

    /** UpdateProjectInput atomic. */
    public atomic: boolean;

    /**
     * Encodes the specified UpdateProjectInput message. Does not implicitly {@link UpdateProjectInput.verify|verify} messages.
     * @param message UpdateProjectInput message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: UpdateProjectInput, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes an UpdateProjectInput message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns UpdateProjectInput
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): UpdateProjectInput;
}

/** Represents a ResponseMetadata. */
export class ResponseMetadata implements IResponseMetadata {

    /**
     * Constructs a new ResponseMetadata.
     * @param [properties] Properties to set
     */
    constructor(properties?: IResponseMetadata);

    /** ResponseMetadata id. */
    public id: string;

    /** ResponseMetadata uid. */
    public uid: string;

    /** ResponseMetadata slug. */
    public slug: string;

    /** ResponseMetadata type. */
    public type: Type;

    /** ResponseMetadata end. */
    public end: boolean;

    /** ResponseMetadata result. */
    public result: ResultType;

    /** ResponseMetadata to. */
    public to: To;

    /** ResponseMetadata data. */
    public data: string;

    /**
     * Encodes the specified ResponseMetadata message. Does not implicitly {@link ResponseMetadata.verify|verify} messages.
     * @param message ResponseMetadata message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: ResponseMetadata, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a ResponseMetadata message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns ResponseMetadata
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): ResponseMetadata;
}

/** Represents a WsResponseMetadata. */
export class WsResponseMetadata implements IWsResponseMetadata {

    /**
     * Constructs a new WsResponseMetadata.
     * @param [properties] Properties to set
     */
    constructor(properties?: IWsResponseMetadata);

    /** WsResponseMetadata metadata. */
    public metadata?: (ResponseMetadata|null);

    /**
     * Encodes the specified WsResponseMetadata message. Does not implicitly {@link WsResponseMetadata.verify|verify} messages.
     * @param message WsResponseMetadata message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: WsResponseMetadata, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a WsResponseMetadata message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns WsResponseMetadata
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): WsResponseMetadata;
}

/** Represents a Container. */
export class Container implements IContainer {

    /**
     * Constructs a new Container.
     * @param [properties] Properties to set
     */
    constructor(properties?: IContainer);

    /** Container namespace. */
    public namespace: string;

    /** Container pod. */
    public pod: string;

    /** Container container. */
    public container: string;

    /**
     * Encodes the specified Container message. Does not implicitly {@link Container.verify|verify} messages.
     * @param message Container message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: Container, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a Container message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns Container
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): Container;
}

/** Represents a WsHandleShellResponse. */
export class WsHandleShellResponse implements IWsHandleShellResponse {

    /**
     * Constructs a new WsHandleShellResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IWsHandleShellResponse);

    /** WsHandleShellResponse metadata. */
    public metadata?: (ResponseMetadata|null);

    /** WsHandleShellResponse terminal_message. */
    public terminal_message?: (TerminalMessage|null);

    /** WsHandleShellResponse container. */
    public container?: (Container|null);

    /**
     * Encodes the specified WsHandleShellResponse message. Does not implicitly {@link WsHandleShellResponse.verify|verify} messages.
     * @param message WsHandleShellResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: WsHandleShellResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a WsHandleShellResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns WsHandleShellResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): WsHandleShellResponse;
}

/** Represents a WsHandleClusterResponse. */
export class WsHandleClusterResponse implements IWsHandleClusterResponse {

    /**
     * Constructs a new WsHandleClusterResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IWsHandleClusterResponse);

    /** WsHandleClusterResponse metadata. */
    public metadata?: (ResponseMetadata|null);

    /** WsHandleClusterResponse info. */
    public info?: (ClusterInfoResponse|null);

    /**
     * Encodes the specified WsHandleClusterResponse message. Does not implicitly {@link WsHandleClusterResponse.verify|verify} messages.
     * @param message WsHandleClusterResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: WsHandleClusterResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a WsHandleClusterResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns WsHandleClusterResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): WsHandleClusterResponse;
}
