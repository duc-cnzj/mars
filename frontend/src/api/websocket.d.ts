import * as $protobuf from "protobufjs";
import Long = require("long");
/** Namespace websocket. */
export namespace websocket {

    /** Type enum. */
    enum Type {
        TypeUnknown = 0,
        SetUid = 1,
        ReloadProjects = 2,
        CancelProject = 3,
        CreateProject = 4,
        UpdateProject = 5,
        ProcessPercent = 6,
        ClusterInfoSync = 7,
        InternalError = 8,
        ApplyProject = 9,
        ProjectPodEvent = 10,
        HandleExecShell = 50,
        HandleExecShellMsg = 51,
        HandleCloseShell = 52,
        HandleAuthorize = 53
    }

    /** ResultType enum. */
    enum ResultType {
        ResultUnknown = 0,
        Error = 1,
        Success = 2,
        Deployed = 3,
        DeployedFailed = 4,
        DeployedCanceled = 5,
        LogWithContainers = 6
    }

    /** To enum. */
    enum To {
        ToSelf = 0,
        ToAll = 1,
        ToOthers = 2
    }

    /** Properties of a ClusterInfo. */
    interface IClusterInfo {

        /** ClusterInfo status */
        status?: (string|null);

        /** ClusterInfo freeMemory */
        freeMemory?: (string|null);

        /** ClusterInfo freeCpu */
        freeCpu?: (string|null);

        /** ClusterInfo freeRequestMemory */
        freeRequestMemory?: (string|null);

        /** ClusterInfo freeRequestCpu */
        freeRequestCpu?: (string|null);

        /** ClusterInfo totalMemory */
        totalMemory?: (string|null);

        /** ClusterInfo totalCpu */
        totalCpu?: (string|null);

        /** ClusterInfo usageMemoryRate */
        usageMemoryRate?: (string|null);

        /** ClusterInfo usageCpuRate */
        usageCpuRate?: (string|null);

        /** ClusterInfo requestMemoryRate */
        requestMemoryRate?: (string|null);

        /** ClusterInfo requestCpuRate */
        requestCpuRate?: (string|null);
    }

    /** Represents a ClusterInfo. */
    class ClusterInfo implements IClusterInfo {

        /**
         * Constructs a new ClusterInfo.
         * @param [properties] Properties to set
         */
        constructor(properties?: websocket.IClusterInfo);

        /** ClusterInfo status. */
        public status: string;

        /** ClusterInfo freeMemory. */
        public freeMemory: string;

        /** ClusterInfo freeCpu. */
        public freeCpu: string;

        /** ClusterInfo freeRequestMemory. */
        public freeRequestMemory: string;

        /** ClusterInfo freeRequestCpu. */
        public freeRequestCpu: string;

        /** ClusterInfo totalMemory. */
        public totalMemory: string;

        /** ClusterInfo totalCpu. */
        public totalCpu: string;

        /** ClusterInfo usageMemoryRate. */
        public usageMemoryRate: string;

        /** ClusterInfo usageCpuRate. */
        public usageCpuRate: string;

        /** ClusterInfo requestMemoryRate. */
        public requestMemoryRate: string;

        /** ClusterInfo requestCpuRate. */
        public requestCpuRate: string;

        /**
         * Encodes the specified ClusterInfo message. Does not implicitly {@link websocket.ClusterInfo.verify|verify} messages.
         * @param message ClusterInfo message or plain object to encode
         * @param [writer] Writer to encode to
         * @returns Writer
         */
        public static encode(message: websocket.ClusterInfo, writer?: $protobuf.Writer): $protobuf.Writer;

        /**
         * Decodes a ClusterInfo message from the specified reader or buffer.
         * @param reader Reader or buffer to decode from
         * @param [length] Message length if known beforehand
         * @returns ClusterInfo
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): websocket.ClusterInfo;

        /**
         * Gets the default type url for ClusterInfo
         * @param [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns The default type url
         */
        public static getTypeUrl(typeUrlPrefix?: string): string;
    }

    /** Properties of an ExtraValue. */
    interface IExtraValue {

        /** ExtraValue path */
        path?: (string|null);

        /** ExtraValue value */
        value?: (string|null);
    }

    /** Represents an ExtraValue. */
    class ExtraValue implements IExtraValue {

        /**
         * Constructs a new ExtraValue.
         * @param [properties] Properties to set
         */
        constructor(properties?: websocket.IExtraValue);

        /** ExtraValue path. */
        public path: string;

        /** ExtraValue value. */
        public value: string;

        /**
         * Encodes the specified ExtraValue message. Does not implicitly {@link websocket.ExtraValue.verify|verify} messages.
         * @param message ExtraValue message or plain object to encode
         * @param [writer] Writer to encode to
         * @returns Writer
         */
        public static encode(message: websocket.ExtraValue, writer?: $protobuf.Writer): $protobuf.Writer;

        /**
         * Decodes an ExtraValue message from the specified reader or buffer.
         * @param reader Reader or buffer to decode from
         * @param [length] Message length if known beforehand
         * @returns ExtraValue
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): websocket.ExtraValue;

        /**
         * Gets the default type url for ExtraValue
         * @param [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns The default type url
         */
        public static getTypeUrl(typeUrlPrefix?: string): string;
    }

    /** Properties of a Container. */
    interface IContainer {

        /** Container namespace */
        namespace?: (string|null);

        /** Container pod */
        pod?: (string|null);

        /** Container container */
        container?: (string|null);
    }

    /** Represents a Container. */
    class Container implements IContainer {

        /**
         * Constructs a new Container.
         * @param [properties] Properties to set
         */
        constructor(properties?: websocket.IContainer);

        /** Container namespace. */
        public namespace: string;

        /** Container pod. */
        public pod: string;

        /** Container container. */
        public container: string;

        /**
         * Encodes the specified Container message. Does not implicitly {@link websocket.Container.verify|verify} messages.
         * @param message Container message or plain object to encode
         * @param [writer] Writer to encode to
         * @returns Writer
         */
        public static encode(message: websocket.Container, writer?: $protobuf.Writer): $protobuf.Writer;

        /**
         * Decodes a Container message from the specified reader or buffer.
         * @param reader Reader or buffer to decode from
         * @param [length] Message length if known beforehand
         * @returns Container
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): websocket.Container;

        /**
         * Gets the default type url for Container
         * @param [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns The default type url
         */
        public static getTypeUrl(typeUrlPrefix?: string): string;
    }

    /** Properties of a WsRequestMetadata. */
    interface IWsRequestMetadata {

        /** WsRequestMetadata type */
        type?: (websocket.Type|null);
    }

    /** Represents a WsRequestMetadata. */
    class WsRequestMetadata implements IWsRequestMetadata {

        /**
         * Constructs a new WsRequestMetadata.
         * @param [properties] Properties to set
         */
        constructor(properties?: websocket.IWsRequestMetadata);

        /** WsRequestMetadata type. */
        public type: websocket.Type;

        /**
         * Encodes the specified WsRequestMetadata message. Does not implicitly {@link websocket.WsRequestMetadata.verify|verify} messages.
         * @param message WsRequestMetadata message or plain object to encode
         * @param [writer] Writer to encode to
         * @returns Writer
         */
        public static encode(message: websocket.WsRequestMetadata, writer?: $protobuf.Writer): $protobuf.Writer;

        /**
         * Decodes a WsRequestMetadata message from the specified reader or buffer.
         * @param reader Reader or buffer to decode from
         * @param [length] Message length if known beforehand
         * @returns WsRequestMetadata
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): websocket.WsRequestMetadata;

        /**
         * Gets the default type url for WsRequestMetadata
         * @param [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns The default type url
         */
        public static getTypeUrl(typeUrlPrefix?: string): string;
    }

    /** Properties of an AuthorizeTokenInput. */
    interface IAuthorizeTokenInput {

        /** AuthorizeTokenInput type */
        type?: (websocket.Type|null);

        /** AuthorizeTokenInput token */
        token?: (string|null);
    }

    /** Represents an AuthorizeTokenInput. */
    class AuthorizeTokenInput implements IAuthorizeTokenInput {

        /**
         * Constructs a new AuthorizeTokenInput.
         * @param [properties] Properties to set
         */
        constructor(properties?: websocket.IAuthorizeTokenInput);

        /** AuthorizeTokenInput type. */
        public type: websocket.Type;

        /** AuthorizeTokenInput token. */
        public token: string;

        /**
         * Encodes the specified AuthorizeTokenInput message. Does not implicitly {@link websocket.AuthorizeTokenInput.verify|verify} messages.
         * @param message AuthorizeTokenInput message or plain object to encode
         * @param [writer] Writer to encode to
         * @returns Writer
         */
        public static encode(message: websocket.AuthorizeTokenInput, writer?: $protobuf.Writer): $protobuf.Writer;

        /**
         * Decodes an AuthorizeTokenInput message from the specified reader or buffer.
         * @param reader Reader or buffer to decode from
         * @param [length] Message length if known beforehand
         * @returns AuthorizeTokenInput
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): websocket.AuthorizeTokenInput;

        /**
         * Gets the default type url for AuthorizeTokenInput
         * @param [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns The default type url
         */
        public static getTypeUrl(typeUrlPrefix?: string): string;
    }

    /** Properties of a TerminalMessage. */
    interface ITerminalMessage {

        /** TerminalMessage op */
        op?: (string|null);

        /** TerminalMessage data */
        data?: (Uint8Array|null);

        /** TerminalMessage sessionId */
        sessionId?: (string|null);

        /** TerminalMessage rows */
        rows?: (number|null);

        /** TerminalMessage cols */
        cols?: (number|null);
    }

    /** Represents a TerminalMessage. */
    class TerminalMessage implements ITerminalMessage {

        /**
         * Constructs a new TerminalMessage.
         * @param [properties] Properties to set
         */
        constructor(properties?: websocket.ITerminalMessage);

        /** TerminalMessage op. */
        public op: string;

        /** TerminalMessage data. */
        public data: Uint8Array;

        /** TerminalMessage sessionId. */
        public sessionId: string;

        /** TerminalMessage rows. */
        public rows: number;

        /** TerminalMessage cols. */
        public cols: number;

        /**
         * Encodes the specified TerminalMessage message. Does not implicitly {@link websocket.TerminalMessage.verify|verify} messages.
         * @param message TerminalMessage message or plain object to encode
         * @param [writer] Writer to encode to
         * @returns Writer
         */
        public static encode(message: websocket.TerminalMessage, writer?: $protobuf.Writer): $protobuf.Writer;

        /**
         * Decodes a TerminalMessage message from the specified reader or buffer.
         * @param reader Reader or buffer to decode from
         * @param [length] Message length if known beforehand
         * @returns TerminalMessage
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): websocket.TerminalMessage;

        /**
         * Gets the default type url for TerminalMessage
         * @param [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns The default type url
         */
        public static getTypeUrl(typeUrlPrefix?: string): string;
    }

    /** Properties of a ProjectPodEventJoinInput. */
    interface IProjectPodEventJoinInput {

        /** ProjectPodEventJoinInput type */
        type?: (websocket.Type|null);

        /** ProjectPodEventJoinInput join */
        join?: (boolean|null);

        /** ProjectPodEventJoinInput projectId */
        projectId?: (number|null);

        /** ProjectPodEventJoinInput namespaceId */
        namespaceId?: (number|null);
    }

    /** Represents a ProjectPodEventJoinInput. */
    class ProjectPodEventJoinInput implements IProjectPodEventJoinInput {

        /**
         * Constructs a new ProjectPodEventJoinInput.
         * @param [properties] Properties to set
         */
        constructor(properties?: websocket.IProjectPodEventJoinInput);

        /** ProjectPodEventJoinInput type. */
        public type: websocket.Type;

        /** ProjectPodEventJoinInput join. */
        public join: boolean;

        /** ProjectPodEventJoinInput projectId. */
        public projectId: number;

        /** ProjectPodEventJoinInput namespaceId. */
        public namespaceId: number;

        /**
         * Encodes the specified ProjectPodEventJoinInput message. Does not implicitly {@link websocket.ProjectPodEventJoinInput.verify|verify} messages.
         * @param message ProjectPodEventJoinInput message or plain object to encode
         * @param [writer] Writer to encode to
         * @returns Writer
         */
        public static encode(message: websocket.ProjectPodEventJoinInput, writer?: $protobuf.Writer): $protobuf.Writer;

        /**
         * Decodes a ProjectPodEventJoinInput message from the specified reader or buffer.
         * @param reader Reader or buffer to decode from
         * @param [length] Message length if known beforehand
         * @returns ProjectPodEventJoinInput
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): websocket.ProjectPodEventJoinInput;

        /**
         * Gets the default type url for ProjectPodEventJoinInput
         * @param [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns The default type url
         */
        public static getTypeUrl(typeUrlPrefix?: string): string;
    }

    /** Properties of a TerminalMessageInput. */
    interface ITerminalMessageInput {

        /** TerminalMessageInput type */
        type?: (websocket.Type|null);

        /** TerminalMessageInput message */
        message?: (websocket.TerminalMessage|null);
    }

    /** Represents a TerminalMessageInput. */
    class TerminalMessageInput implements ITerminalMessageInput {

        /**
         * Constructs a new TerminalMessageInput.
         * @param [properties] Properties to set
         */
        constructor(properties?: websocket.ITerminalMessageInput);

        /** TerminalMessageInput type. */
        public type: websocket.Type;

        /** TerminalMessageInput message. */
        public message?: (websocket.TerminalMessage|null);

        /**
         * Encodes the specified TerminalMessageInput message. Does not implicitly {@link websocket.TerminalMessageInput.verify|verify} messages.
         * @param message TerminalMessageInput message or plain object to encode
         * @param [writer] Writer to encode to
         * @returns Writer
         */
        public static encode(message: websocket.TerminalMessageInput, writer?: $protobuf.Writer): $protobuf.Writer;

        /**
         * Decodes a TerminalMessageInput message from the specified reader or buffer.
         * @param reader Reader or buffer to decode from
         * @param [length] Message length if known beforehand
         * @returns TerminalMessageInput
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): websocket.TerminalMessageInput;

        /**
         * Gets the default type url for TerminalMessageInput
         * @param [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns The default type url
         */
        public static getTypeUrl(typeUrlPrefix?: string): string;
    }

    /** Properties of a WsHandleExecShellInput. */
    interface IWsHandleExecShellInput {

        /** WsHandleExecShellInput type */
        type?: (websocket.Type|null);

        /** WsHandleExecShellInput container */
        container?: (websocket.Container|null);

        /** WsHandleExecShellInput sessionId */
        sessionId?: (string|null);
    }

    /** Represents a WsHandleExecShellInput. */
    class WsHandleExecShellInput implements IWsHandleExecShellInput {

        /**
         * Constructs a new WsHandleExecShellInput.
         * @param [properties] Properties to set
         */
        constructor(properties?: websocket.IWsHandleExecShellInput);

        /** WsHandleExecShellInput type. */
        public type: websocket.Type;

        /** WsHandleExecShellInput container. */
        public container?: (websocket.Container|null);

        /** WsHandleExecShellInput sessionId. */
        public sessionId: string;

        /**
         * Encodes the specified WsHandleExecShellInput message. Does not implicitly {@link websocket.WsHandleExecShellInput.verify|verify} messages.
         * @param message WsHandleExecShellInput message or plain object to encode
         * @param [writer] Writer to encode to
         * @returns Writer
         */
        public static encode(message: websocket.WsHandleExecShellInput, writer?: $protobuf.Writer): $protobuf.Writer;

        /**
         * Decodes a WsHandleExecShellInput message from the specified reader or buffer.
         * @param reader Reader or buffer to decode from
         * @param [length] Message length if known beforehand
         * @returns WsHandleExecShellInput
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): websocket.WsHandleExecShellInput;

        /**
         * Gets the default type url for WsHandleExecShellInput
         * @param [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns The default type url
         */
        public static getTypeUrl(typeUrlPrefix?: string): string;
    }

    /** Properties of a CancelInput. */
    interface ICancelInput {

        /** CancelInput type */
        type?: (websocket.Type|null);

        /** CancelInput namespaceId */
        namespaceId?: (number|null);

        /** CancelInput name */
        name?: (string|null);
    }

    /** Represents a CancelInput. */
    class CancelInput implements ICancelInput {

        /**
         * Constructs a new CancelInput.
         * @param [properties] Properties to set
         */
        constructor(properties?: websocket.ICancelInput);

        /** CancelInput type. */
        public type: websocket.Type;

        /** CancelInput namespaceId. */
        public namespaceId: number;

        /** CancelInput name. */
        public name: string;

        /**
         * Encodes the specified CancelInput message. Does not implicitly {@link websocket.CancelInput.verify|verify} messages.
         * @param message CancelInput message or plain object to encode
         * @param [writer] Writer to encode to
         * @returns Writer
         */
        public static encode(message: websocket.CancelInput, writer?: $protobuf.Writer): $protobuf.Writer;

        /**
         * Decodes a CancelInput message from the specified reader or buffer.
         * @param reader Reader or buffer to decode from
         * @param [length] Message length if known beforehand
         * @returns CancelInput
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): websocket.CancelInput;

        /**
         * Gets the default type url for CancelInput
         * @param [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns The default type url
         */
        public static getTypeUrl(typeUrlPrefix?: string): string;
    }

    /** Properties of a CreateProjectInput. */
    interface ICreateProjectInput {

        /** CreateProjectInput type */
        type?: (websocket.Type|null);

        /** CreateProjectInput namespaceId */
        namespaceId?: (number|null);

        /** CreateProjectInput name */
        name?: (string|null);

        /** CreateProjectInput repoId */
        repoId?: (number|null);

        /** CreateProjectInput gitBranch */
        gitBranch?: (string|null);

        /** CreateProjectInput gitCommit */
        gitCommit?: (string|null);

        /** CreateProjectInput config */
        config?: (string|null);

        /** CreateProjectInput extraValues */
        extraValues?: (websocket.ExtraValue[]|null);

        /** CreateProjectInput atomic */
        atomic?: (boolean|null);
    }

    /** Represents a CreateProjectInput. */
    class CreateProjectInput implements ICreateProjectInput {

        /**
         * Constructs a new CreateProjectInput.
         * @param [properties] Properties to set
         */
        constructor(properties?: websocket.ICreateProjectInput);

        /** CreateProjectInput type. */
        public type: websocket.Type;

        /** CreateProjectInput namespaceId. */
        public namespaceId: number;

        /** CreateProjectInput name. */
        public name?: (string|null);

        /** CreateProjectInput repoId. */
        public repoId: number;

        /** CreateProjectInput gitBranch. */
        public gitBranch: string;

        /** CreateProjectInput gitCommit. */
        public gitCommit: string;

        /** CreateProjectInput config. */
        public config: string;

        /** CreateProjectInput extraValues. */
        public extraValues: websocket.ExtraValue[];

        /** CreateProjectInput atomic. */
        public atomic?: (boolean|null);

        /**
         * Encodes the specified CreateProjectInput message. Does not implicitly {@link websocket.CreateProjectInput.verify|verify} messages.
         * @param message CreateProjectInput message or plain object to encode
         * @param [writer] Writer to encode to
         * @returns Writer
         */
        public static encode(message: websocket.CreateProjectInput, writer?: $protobuf.Writer): $protobuf.Writer;

        /**
         * Decodes a CreateProjectInput message from the specified reader or buffer.
         * @param reader Reader or buffer to decode from
         * @param [length] Message length if known beforehand
         * @returns CreateProjectInput
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): websocket.CreateProjectInput;

        /**
         * Gets the default type url for CreateProjectInput
         * @param [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns The default type url
         */
        public static getTypeUrl(typeUrlPrefix?: string): string;
    }

    /** Properties of an UpdateProjectInput. */
    interface IUpdateProjectInput {

        /** UpdateProjectInput type */
        type?: (websocket.Type|null);

        /** UpdateProjectInput projectId */
        projectId?: (number|null);

        /** UpdateProjectInput gitBranch */
        gitBranch?: (string|null);

        /** UpdateProjectInput gitCommit */
        gitCommit?: (string|null);

        /** UpdateProjectInput config */
        config?: (string|null);

        /** UpdateProjectInput extraValues */
        extraValues?: (websocket.ExtraValue[]|null);

        /** UpdateProjectInput version */
        version?: (number|null);

        /** UpdateProjectInput atomic */
        atomic?: (boolean|null);
    }

    /** Represents an UpdateProjectInput. */
    class UpdateProjectInput implements IUpdateProjectInput {

        /**
         * Constructs a new UpdateProjectInput.
         * @param [properties] Properties to set
         */
        constructor(properties?: websocket.IUpdateProjectInput);

        /** UpdateProjectInput type. */
        public type: websocket.Type;

        /** UpdateProjectInput projectId. */
        public projectId: number;

        /** UpdateProjectInput gitBranch. */
        public gitBranch: string;

        /** UpdateProjectInput gitCommit. */
        public gitCommit: string;

        /** UpdateProjectInput config. */
        public config: string;

        /** UpdateProjectInput extraValues. */
        public extraValues: websocket.ExtraValue[];

        /** UpdateProjectInput version. */
        public version: number;

        /** UpdateProjectInput atomic. */
        public atomic?: (boolean|null);

        /**
         * Encodes the specified UpdateProjectInput message. Does not implicitly {@link websocket.UpdateProjectInput.verify|verify} messages.
         * @param message UpdateProjectInput message or plain object to encode
         * @param [writer] Writer to encode to
         * @returns Writer
         */
        public static encode(message: websocket.UpdateProjectInput, writer?: $protobuf.Writer): $protobuf.Writer;

        /**
         * Decodes an UpdateProjectInput message from the specified reader or buffer.
         * @param reader Reader or buffer to decode from
         * @param [length] Message length if known beforehand
         * @returns UpdateProjectInput
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): websocket.UpdateProjectInput;

        /**
         * Gets the default type url for UpdateProjectInput
         * @param [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns The default type url
         */
        public static getTypeUrl(typeUrlPrefix?: string): string;
    }

    /** Properties of a Metadata. */
    interface IMetadata {

        /** Metadata id */
        id?: (string|null);

        /** Metadata uid */
        uid?: (string|null);

        /** Metadata slug */
        slug?: (string|null);

        /** Metadata type */
        type?: (websocket.Type|null);

        /** Metadata end */
        end?: (boolean|null);

        /** Metadata result */
        result?: (websocket.ResultType|null);

        /** Metadata to */
        to?: (websocket.To|null);

        /** Metadata message */
        message?: (string|null);

        /** Metadata percent */
        percent?: (number|null);
    }

    /** Represents a Metadata. */
    class Metadata implements IMetadata {

        /**
         * Constructs a new Metadata.
         * @param [properties] Properties to set
         */
        constructor(properties?: websocket.IMetadata);

        /** Metadata id. */
        public id: string;

        /** Metadata uid. */
        public uid: string;

        /** Metadata slug. */
        public slug: string;

        /** Metadata type. */
        public type: websocket.Type;

        /** Metadata end. */
        public end: boolean;

        /** Metadata result. */
        public result: websocket.ResultType;

        /** Metadata to. */
        public to: websocket.To;

        /** Metadata message. */
        public message: string;

        /** Metadata percent. */
        public percent: number;

        /**
         * Encodes the specified Metadata message. Does not implicitly {@link websocket.Metadata.verify|verify} messages.
         * @param message Metadata message or plain object to encode
         * @param [writer] Writer to encode to
         * @returns Writer
         */
        public static encode(message: websocket.Metadata, writer?: $protobuf.Writer): $protobuf.Writer;

        /**
         * Decodes a Metadata message from the specified reader or buffer.
         * @param reader Reader or buffer to decode from
         * @param [length] Message length if known beforehand
         * @returns Metadata
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): websocket.Metadata;

        /**
         * Gets the default type url for Metadata
         * @param [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns The default type url
         */
        public static getTypeUrl(typeUrlPrefix?: string): string;
    }

    /** Properties of a WsMetadataResponse. */
    interface IWsMetadataResponse {

        /** WsMetadataResponse metadata */
        metadata?: (websocket.Metadata|null);
    }

    /** Represents a WsMetadataResponse. */
    class WsMetadataResponse implements IWsMetadataResponse {

        /**
         * Constructs a new WsMetadataResponse.
         * @param [properties] Properties to set
         */
        constructor(properties?: websocket.IWsMetadataResponse);

        /** WsMetadataResponse metadata. */
        public metadata?: (websocket.Metadata|null);

        /**
         * Encodes the specified WsMetadataResponse message. Does not implicitly {@link websocket.WsMetadataResponse.verify|verify} messages.
         * @param message WsMetadataResponse message or plain object to encode
         * @param [writer] Writer to encode to
         * @returns Writer
         */
        public static encode(message: websocket.WsMetadataResponse, writer?: $protobuf.Writer): $protobuf.Writer;

        /**
         * Decodes a WsMetadataResponse message from the specified reader or buffer.
         * @param reader Reader or buffer to decode from
         * @param [length] Message length if known beforehand
         * @returns WsMetadataResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): websocket.WsMetadataResponse;

        /**
         * Gets the default type url for WsMetadataResponse
         * @param [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns The default type url
         */
        public static getTypeUrl(typeUrlPrefix?: string): string;
    }

    /** Properties of a WsHandleShellResponse. */
    interface IWsHandleShellResponse {

        /** WsHandleShellResponse metadata */
        metadata?: (websocket.Metadata|null);

        /** WsHandleShellResponse terminalMessage */
        terminalMessage?: (websocket.TerminalMessage|null);

        /** WsHandleShellResponse container */
        container?: (websocket.Container|null);
    }

    /** Represents a WsHandleShellResponse. */
    class WsHandleShellResponse implements IWsHandleShellResponse {

        /**
         * Constructs a new WsHandleShellResponse.
         * @param [properties] Properties to set
         */
        constructor(properties?: websocket.IWsHandleShellResponse);

        /** WsHandleShellResponse metadata. */
        public metadata?: (websocket.Metadata|null);

        /** WsHandleShellResponse terminalMessage. */
        public terminalMessage?: (websocket.TerminalMessage|null);

        /** WsHandleShellResponse container. */
        public container?: (websocket.Container|null);

        /**
         * Encodes the specified WsHandleShellResponse message. Does not implicitly {@link websocket.WsHandleShellResponse.verify|verify} messages.
         * @param message WsHandleShellResponse message or plain object to encode
         * @param [writer] Writer to encode to
         * @returns Writer
         */
        public static encode(message: websocket.WsHandleShellResponse, writer?: $protobuf.Writer): $protobuf.Writer;

        /**
         * Decodes a WsHandleShellResponse message from the specified reader or buffer.
         * @param reader Reader or buffer to decode from
         * @param [length] Message length if known beforehand
         * @returns WsHandleShellResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): websocket.WsHandleShellResponse;

        /**
         * Gets the default type url for WsHandleShellResponse
         * @param [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns The default type url
         */
        public static getTypeUrl(typeUrlPrefix?: string): string;
    }

    /** Properties of a WsHandleClusterResponse. */
    interface IWsHandleClusterResponse {

        /** WsHandleClusterResponse metadata */
        metadata?: (websocket.Metadata|null);

        /** WsHandleClusterResponse info */
        info?: (websocket.ClusterInfo|null);
    }

    /** Represents a WsHandleClusterResponse. */
    class WsHandleClusterResponse implements IWsHandleClusterResponse {

        /**
         * Constructs a new WsHandleClusterResponse.
         * @param [properties] Properties to set
         */
        constructor(properties?: websocket.IWsHandleClusterResponse);

        /** WsHandleClusterResponse metadata. */
        public metadata?: (websocket.Metadata|null);

        /** WsHandleClusterResponse info. */
        public info?: (websocket.ClusterInfo|null);

        /**
         * Encodes the specified WsHandleClusterResponse message. Does not implicitly {@link websocket.WsHandleClusterResponse.verify|verify} messages.
         * @param message WsHandleClusterResponse message or plain object to encode
         * @param [writer] Writer to encode to
         * @returns Writer
         */
        public static encode(message: websocket.WsHandleClusterResponse, writer?: $protobuf.Writer): $protobuf.Writer;

        /**
         * Decodes a WsHandleClusterResponse message from the specified reader or buffer.
         * @param reader Reader or buffer to decode from
         * @param [length] Message length if known beforehand
         * @returns WsHandleClusterResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): websocket.WsHandleClusterResponse;

        /**
         * Gets the default type url for WsHandleClusterResponse
         * @param [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns The default type url
         */
        public static getTypeUrl(typeUrlPrefix?: string): string;
    }

    /** Properties of a WsWithContainerMessageResponse. */
    interface IWsWithContainerMessageResponse {

        /** WsWithContainerMessageResponse metadata */
        metadata?: (websocket.Metadata|null);

        /** WsWithContainerMessageResponse containers */
        containers?: (websocket.Container[]|null);
    }

    /** Represents a WsWithContainerMessageResponse. */
    class WsWithContainerMessageResponse implements IWsWithContainerMessageResponse {

        /**
         * Constructs a new WsWithContainerMessageResponse.
         * @param [properties] Properties to set
         */
        constructor(properties?: websocket.IWsWithContainerMessageResponse);

        /** WsWithContainerMessageResponse metadata. */
        public metadata?: (websocket.Metadata|null);

        /** WsWithContainerMessageResponse containers. */
        public containers: websocket.Container[];

        /**
         * Encodes the specified WsWithContainerMessageResponse message. Does not implicitly {@link websocket.WsWithContainerMessageResponse.verify|verify} messages.
         * @param message WsWithContainerMessageResponse message or plain object to encode
         * @param [writer] Writer to encode to
         * @returns Writer
         */
        public static encode(message: websocket.WsWithContainerMessageResponse, writer?: $protobuf.Writer): $protobuf.Writer;

        /**
         * Decodes a WsWithContainerMessageResponse message from the specified reader or buffer.
         * @param reader Reader or buffer to decode from
         * @param [length] Message length if known beforehand
         * @returns WsWithContainerMessageResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): websocket.WsWithContainerMessageResponse;

        /**
         * Gets the default type url for WsWithContainerMessageResponse
         * @param [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns The default type url
         */
        public static getTypeUrl(typeUrlPrefix?: string): string;
    }

    /** Properties of a WsProjectPodEventResponse. */
    interface IWsProjectPodEventResponse {

        /** WsProjectPodEventResponse metadata */
        metadata?: (websocket.Metadata|null);

        /** WsProjectPodEventResponse projectId */
        projectId?: (number|null);
    }

    /** Represents a WsProjectPodEventResponse. */
    class WsProjectPodEventResponse implements IWsProjectPodEventResponse {

        /**
         * Constructs a new WsProjectPodEventResponse.
         * @param [properties] Properties to set
         */
        constructor(properties?: websocket.IWsProjectPodEventResponse);

        /** WsProjectPodEventResponse metadata. */
        public metadata?: (websocket.Metadata|null);

        /** WsProjectPodEventResponse projectId. */
        public projectId: number;

        /**
         * Encodes the specified WsProjectPodEventResponse message. Does not implicitly {@link websocket.WsProjectPodEventResponse.verify|verify} messages.
         * @param message WsProjectPodEventResponse message or plain object to encode
         * @param [writer] Writer to encode to
         * @returns Writer
         */
        public static encode(message: websocket.WsProjectPodEventResponse, writer?: $protobuf.Writer): $protobuf.Writer;

        /**
         * Decodes a WsProjectPodEventResponse message from the specified reader or buffer.
         * @param reader Reader or buffer to decode from
         * @param [length] Message length if known beforehand
         * @returns WsProjectPodEventResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): websocket.WsProjectPodEventResponse;

        /**
         * Gets the default type url for WsProjectPodEventResponse
         * @param [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns The default type url
         */
        public static getTypeUrl(typeUrlPrefix?: string): string;
    }

    /** Properties of a WsReloadProjectsResponse. */
    interface IWsReloadProjectsResponse {

        /** WsReloadProjectsResponse metadata */
        metadata?: (websocket.Metadata|null);

        /** WsReloadProjectsResponse namespaceId */
        namespaceId?: (number|null);
    }

    /** Represents a WsReloadProjectsResponse. */
    class WsReloadProjectsResponse implements IWsReloadProjectsResponse {

        /**
         * Constructs a new WsReloadProjectsResponse.
         * @param [properties] Properties to set
         */
        constructor(properties?: websocket.IWsReloadProjectsResponse);

        /** WsReloadProjectsResponse metadata. */
        public metadata?: (websocket.Metadata|null);

        /** WsReloadProjectsResponse namespaceId. */
        public namespaceId: number;

        /**
         * Encodes the specified WsReloadProjectsResponse message. Does not implicitly {@link websocket.WsReloadProjectsResponse.verify|verify} messages.
         * @param message WsReloadProjectsResponse message or plain object to encode
         * @param [writer] Writer to encode to
         * @returns Writer
         */
        public static encode(message: websocket.WsReloadProjectsResponse, writer?: $protobuf.Writer): $protobuf.Writer;

        /**
         * Decodes a WsReloadProjectsResponse message from the specified reader or buffer.
         * @param reader Reader or buffer to decode from
         * @param [length] Message length if known beforehand
         * @returns WsReloadProjectsResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): websocket.WsReloadProjectsResponse;

        /**
         * Gets the default type url for WsReloadProjectsResponse
         * @param [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns The default type url
         */
        public static getTypeUrl(typeUrlPrefix?: string): string;
    }
}
