import * as $protobuf from "protobufjs";
/** Properties of an AuthLoginRequest. */
export interface IAuthLoginRequest {

    /** AuthLoginRequest username */
    username?: (string|null);

    /** AuthLoginRequest password */
    password?: (string|null);
}

/** Represents an AuthLoginRequest. */
export class AuthLoginRequest implements IAuthLoginRequest {

    /**
     * Constructs a new AuthLoginRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IAuthLoginRequest);

    /** AuthLoginRequest username. */
    public username: string;

    /** AuthLoginRequest password. */
    public password: string;

    /**
     * Encodes the specified AuthLoginRequest message. Does not implicitly {@link AuthLoginRequest.verify|verify} messages.
     * @param message AuthLoginRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: AuthLoginRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes an AuthLoginRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns AuthLoginRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): AuthLoginRequest;
}

/** Properties of an AuthLoginResponse. */
export interface IAuthLoginResponse {

    /** AuthLoginResponse token */
    token?: (string|null);

    /** AuthLoginResponse expires_in */
    expires_in?: (number|null);
}

/** Represents an AuthLoginResponse. */
export class AuthLoginResponse implements IAuthLoginResponse {

    /**
     * Constructs a new AuthLoginResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IAuthLoginResponse);

    /** AuthLoginResponse token. */
    public token: string;

    /** AuthLoginResponse expires_in. */
    public expires_in: number;

    /**
     * Encodes the specified AuthLoginResponse message. Does not implicitly {@link AuthLoginResponse.verify|verify} messages.
     * @param message AuthLoginResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: AuthLoginResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes an AuthLoginResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns AuthLoginResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): AuthLoginResponse;
}

/** Properties of an AuthExchangeRequest. */
export interface IAuthExchangeRequest {

    /** AuthExchangeRequest code */
    code?: (string|null);
}

/** Represents an AuthExchangeRequest. */
export class AuthExchangeRequest implements IAuthExchangeRequest {

    /**
     * Constructs a new AuthExchangeRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IAuthExchangeRequest);

    /** AuthExchangeRequest code. */
    public code: string;

    /**
     * Encodes the specified AuthExchangeRequest message. Does not implicitly {@link AuthExchangeRequest.verify|verify} messages.
     * @param message AuthExchangeRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: AuthExchangeRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes an AuthExchangeRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns AuthExchangeRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): AuthExchangeRequest;
}

/** Properties of an AuthExchangeResponse. */
export interface IAuthExchangeResponse {

    /** AuthExchangeResponse token */
    token?: (string|null);

    /** AuthExchangeResponse expires_in */
    expires_in?: (number|null);
}

/** Represents an AuthExchangeResponse. */
export class AuthExchangeResponse implements IAuthExchangeResponse {

    /**
     * Constructs a new AuthExchangeResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IAuthExchangeResponse);

    /** AuthExchangeResponse token. */
    public token: string;

    /** AuthExchangeResponse expires_in. */
    public expires_in: number;

    /**
     * Encodes the specified AuthExchangeResponse message. Does not implicitly {@link AuthExchangeResponse.verify|verify} messages.
     * @param message AuthExchangeResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: AuthExchangeResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes an AuthExchangeResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns AuthExchangeResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): AuthExchangeResponse;
}

/** Properties of an AuthInfoRequest. */
export interface IAuthInfoRequest {
}

/** Represents an AuthInfoRequest. */
export class AuthInfoRequest implements IAuthInfoRequest {

    /**
     * Constructs a new AuthInfoRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IAuthInfoRequest);

    /**
     * Encodes the specified AuthInfoRequest message. Does not implicitly {@link AuthInfoRequest.verify|verify} messages.
     * @param message AuthInfoRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: AuthInfoRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes an AuthInfoRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns AuthInfoRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): AuthInfoRequest;
}

/** Properties of an AuthInfoResponse. */
export interface IAuthInfoResponse {

    /** AuthInfoResponse id */
    id?: (string|null);

    /** AuthInfoResponse avatar */
    avatar?: (string|null);

    /** AuthInfoResponse name */
    name?: (string|null);

    /** AuthInfoResponse email */
    email?: (string|null);

    /** AuthInfoResponse logout_url */
    logout_url?: (string|null);

    /** AuthInfoResponse roles */
    roles?: (string[]|null);
}

/** Represents an AuthInfoResponse. */
export class AuthInfoResponse implements IAuthInfoResponse {

    /**
     * Constructs a new AuthInfoResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IAuthInfoResponse);

    /** AuthInfoResponse id. */
    public id: string;

    /** AuthInfoResponse avatar. */
    public avatar: string;

    /** AuthInfoResponse name. */
    public name: string;

    /** AuthInfoResponse email. */
    public email: string;

    /** AuthInfoResponse logout_url. */
    public logout_url: string;

    /** AuthInfoResponse roles. */
    public roles: string[];

    /**
     * Encodes the specified AuthInfoResponse message. Does not implicitly {@link AuthInfoResponse.verify|verify} messages.
     * @param message AuthInfoResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: AuthInfoResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes an AuthInfoResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns AuthInfoResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): AuthInfoResponse;
}

/** Properties of an AuthSettingsRequest. */
export interface IAuthSettingsRequest {
}

/** Represents an AuthSettingsRequest. */
export class AuthSettingsRequest implements IAuthSettingsRequest {

    /**
     * Constructs a new AuthSettingsRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IAuthSettingsRequest);

    /**
     * Encodes the specified AuthSettingsRequest message. Does not implicitly {@link AuthSettingsRequest.verify|verify} messages.
     * @param message AuthSettingsRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: AuthSettingsRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes an AuthSettingsRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns AuthSettingsRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): AuthSettingsRequest;
}

/** Properties of an AuthSettingsResponse. */
export interface IAuthSettingsResponse {

    /** AuthSettingsResponse items */
    items?: (AuthSettingsResponse.OidcSetting[]|null);
}

/** Represents an AuthSettingsResponse. */
export class AuthSettingsResponse implements IAuthSettingsResponse {

    /**
     * Constructs a new AuthSettingsResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IAuthSettingsResponse);

    /** AuthSettingsResponse items. */
    public items: AuthSettingsResponse.OidcSetting[];

    /**
     * Encodes the specified AuthSettingsResponse message. Does not implicitly {@link AuthSettingsResponse.verify|verify} messages.
     * @param message AuthSettingsResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: AuthSettingsResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes an AuthSettingsResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns AuthSettingsResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): AuthSettingsResponse;
}

export namespace AuthSettingsResponse {

    /** Properties of an OidcSetting. */
    interface IOidcSetting {

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
    class OidcSetting implements IOidcSetting {

        /**
         * Constructs a new OidcSetting.
         * @param [properties] Properties to set
         */
        constructor(properties?: AuthSettingsResponse.IOidcSetting);

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
         * Encodes the specified OidcSetting message. Does not implicitly {@link AuthSettingsResponse.OidcSetting.verify|verify} messages.
         * @param message OidcSetting message or plain object to encode
         * @param [writer] Writer to encode to
         * @returns Writer
         */
        public static encode(message: AuthSettingsResponse.OidcSetting, writer?: $protobuf.Writer): $protobuf.Writer;

        /**
         * Decodes an OidcSetting message from the specified reader or buffer.
         * @param reader Reader or buffer to decode from
         * @param [length] Message length if known beforehand
         * @returns OidcSetting
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): AuthSettingsResponse.OidcSetting;
    }
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
     * @param request AuthLoginRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and AuthLoginResponse
     */
    public login(request: AuthLoginRequest, callback: Auth.LoginCallback): void;

    /**
     * Calls Login.
     * @param request AuthLoginRequest message or plain object
     * @returns Promise
     */
    public login(request: AuthLoginRequest): Promise<AuthLoginResponse>;

    /**
     * Calls Info.
     * @param request AuthInfoRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and AuthInfoResponse
     */
    public info(request: AuthInfoRequest, callback: Auth.InfoCallback): void;

    /**
     * Calls Info.
     * @param request AuthInfoRequest message or plain object
     * @returns Promise
     */
    public info(request: AuthInfoRequest): Promise<AuthInfoResponse>;

    /**
     * Calls Settings.
     * @param request AuthSettingsRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and AuthSettingsResponse
     */
    public settings(request: AuthSettingsRequest, callback: Auth.SettingsCallback): void;

    /**
     * Calls Settings.
     * @param request AuthSettingsRequest message or plain object
     * @returns Promise
     */
    public settings(request: AuthSettingsRequest): Promise<AuthSettingsResponse>;

    /**
     * Calls Exchange.
     * @param request AuthExchangeRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and AuthExchangeResponse
     */
    public exchange(request: AuthExchangeRequest, callback: Auth.ExchangeCallback): void;

    /**
     * Calls Exchange.
     * @param request AuthExchangeRequest message or plain object
     * @returns Promise
     */
    public exchange(request: AuthExchangeRequest): Promise<AuthExchangeResponse>;
}

export namespace Auth {

    /**
     * Callback as used by {@link Auth#login}.
     * @param error Error, if any
     * @param [response] AuthLoginResponse
     */
    type LoginCallback = (error: (Error|null), response?: AuthLoginResponse) => void;

    /**
     * Callback as used by {@link Auth#info}.
     * @param error Error, if any
     * @param [response] AuthInfoResponse
     */
    type InfoCallback = (error: (Error|null), response?: AuthInfoResponse) => void;

    /**
     * Callback as used by {@link Auth#settings}.
     * @param error Error, if any
     * @param [response] AuthSettingsResponse
     */
    type SettingsCallback = (error: (Error|null), response?: AuthSettingsResponse) => void;

    /**
     * Callback as used by {@link Auth#exchange}.
     * @param error Error, if any
     * @param [response] AuthExchangeResponse
     */
    type ExchangeCallback = (error: (Error|null), response?: AuthExchangeResponse) => void;
}

/** Namespace google. */
export namespace google {

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

    /** Namespace protobuf. */
    namespace protobuf {

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
    }
}

/** Properties of a ChangelogShowRequest. */
export interface IChangelogShowRequest {

    /** ChangelogShowRequest project_id */
    project_id?: (number|null);

    /** ChangelogShowRequest only_changed */
    only_changed?: (boolean|null);
}

/** Represents a ChangelogShowRequest. */
export class ChangelogShowRequest implements IChangelogShowRequest {

    /**
     * Constructs a new ChangelogShowRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IChangelogShowRequest);

    /** ChangelogShowRequest project_id. */
    public project_id: number;

    /** ChangelogShowRequest only_changed. */
    public only_changed: boolean;

    /**
     * Encodes the specified ChangelogShowRequest message. Does not implicitly {@link ChangelogShowRequest.verify|verify} messages.
     * @param message ChangelogShowRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: ChangelogShowRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a ChangelogShowRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns ChangelogShowRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): ChangelogShowRequest;
}

/** Properties of a ChangelogShowItem. */
export interface IChangelogShowItem {

    /** ChangelogShowItem version */
    version?: (number|null);

    /** ChangelogShowItem config */
    config?: (string|null);

    /** ChangelogShowItem date */
    date?: (string|null);

    /** ChangelogShowItem username */
    username?: (string|null);
}

/** Represents a ChangelogShowItem. */
export class ChangelogShowItem implements IChangelogShowItem {

    /**
     * Constructs a new ChangelogShowItem.
     * @param [properties] Properties to set
     */
    constructor(properties?: IChangelogShowItem);

    /** ChangelogShowItem version. */
    public version: number;

    /** ChangelogShowItem config. */
    public config: string;

    /** ChangelogShowItem date. */
    public date: string;

    /** ChangelogShowItem username. */
    public username: string;

    /**
     * Encodes the specified ChangelogShowItem message. Does not implicitly {@link ChangelogShowItem.verify|verify} messages.
     * @param message ChangelogShowItem message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: ChangelogShowItem, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a ChangelogShowItem message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns ChangelogShowItem
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): ChangelogShowItem;
}

/** Properties of a ChangelogShowResponse. */
export interface IChangelogShowResponse {

    /** ChangelogShowResponse items */
    items?: (ChangelogShowItem[]|null);
}

/** Represents a ChangelogShowResponse. */
export class ChangelogShowResponse implements IChangelogShowResponse {

    /**
     * Constructs a new ChangelogShowResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IChangelogShowResponse);

    /** ChangelogShowResponse items. */
    public items: ChangelogShowItem[];

    /**
     * Encodes the specified ChangelogShowResponse message. Does not implicitly {@link ChangelogShowResponse.verify|verify} messages.
     * @param message ChangelogShowResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: ChangelogShowResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a ChangelogShowResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns ChangelogShowResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): ChangelogShowResponse;
}

/** Represents a Changelog */
export class Changelog extends $protobuf.rpc.Service {

    /**
     * Constructs a new Changelog service.
     * @param rpcImpl RPC implementation
     * @param [requestDelimited=false] Whether requests are length-delimited
     * @param [responseDelimited=false] Whether responses are length-delimited
     */
    constructor(rpcImpl: $protobuf.RPCImpl, requestDelimited?: boolean, responseDelimited?: boolean);

    /**
     * Calls Show.
     * @param request ChangelogShowRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and ChangelogShowResponse
     */
    public show(request: ChangelogShowRequest, callback: Changelog.ShowCallback): void;

    /**
     * Calls Show.
     * @param request ChangelogShowRequest message or plain object
     * @returns Promise
     */
    public show(request: ChangelogShowRequest): Promise<ChangelogShowResponse>;
}

export namespace Changelog {

    /**
     * Callback as used by {@link Changelog#show}.
     * @param error Error, if any
     * @param [response] ChangelogShowResponse
     */
    type ShowCallback = (error: (Error|null), response?: ChangelogShowResponse) => void;
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

/** Represents a ClusterInfoRequest. */
export class ClusterInfoRequest implements IClusterInfoRequest {

    /**
     * Constructs a new ClusterInfoRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IClusterInfoRequest);

    /**
     * Encodes the specified ClusterInfoRequest message. Does not implicitly {@link ClusterInfoRequest.verify|verify} messages.
     * @param message ClusterInfoRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: ClusterInfoRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a ClusterInfoRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns ClusterInfoRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): ClusterInfoRequest;
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
     * Calls ClusterInfo.
     * @param request ClusterInfoRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and ClusterInfoResponse
     */
    public clusterInfo(request: ClusterInfoRequest, callback: Cluster.ClusterInfoCallback): void;

    /**
     * Calls ClusterInfo.
     * @param request ClusterInfoRequest message or plain object
     * @returns Promise
     */
    public clusterInfo(request: ClusterInfoRequest): Promise<ClusterInfoResponse>;
}

export namespace Cluster {

    /**
     * Callback as used by {@link Cluster#clusterInfo}.
     * @param error Error, if any
     * @param [response] ClusterInfoResponse
     */
    type ClusterInfoCallback = (error: (Error|null), response?: ClusterInfoResponse) => void;
}

/** Represents a ContainerCopyToPodRequest. */
export class ContainerCopyToPodRequest implements IContainerCopyToPodRequest {

    /**
     * Constructs a new ContainerCopyToPodRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IContainerCopyToPodRequest);

    /** ContainerCopyToPodRequest file_id. */
    public file_id: number;

    /** ContainerCopyToPodRequest namespace. */
    public namespace: string;

    /** ContainerCopyToPodRequest pod. */
    public pod: string;

    /** ContainerCopyToPodRequest container. */
    public container: string;

    /**
     * Encodes the specified ContainerCopyToPodRequest message. Does not implicitly {@link ContainerCopyToPodRequest.verify|verify} messages.
     * @param message ContainerCopyToPodRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: ContainerCopyToPodRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a ContainerCopyToPodRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns ContainerCopyToPodRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): ContainerCopyToPodRequest;
}

/** Represents a ContainerCopyToPodResponse. */
export class ContainerCopyToPodResponse implements IContainerCopyToPodResponse {

    /**
     * Constructs a new ContainerCopyToPodResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IContainerCopyToPodResponse);

    /** ContainerCopyToPodResponse pod_file_path. */
    public pod_file_path: string;

    /** ContainerCopyToPodResponse output. */
    public output: string;

    /** ContainerCopyToPodResponse file_name. */
    public file_name: string;

    /**
     * Encodes the specified ContainerCopyToPodResponse message. Does not implicitly {@link ContainerCopyToPodResponse.verify|verify} messages.
     * @param message ContainerCopyToPodResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: ContainerCopyToPodResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a ContainerCopyToPodResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns ContainerCopyToPodResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): ContainerCopyToPodResponse;
}

/** Represents a ContainerExecRequest. */
export class ContainerExecRequest implements IContainerExecRequest {

    /**
     * Constructs a new ContainerExecRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IContainerExecRequest);

    /** ContainerExecRequest namespace. */
    public namespace: string;

    /** ContainerExecRequest pod. */
    public pod: string;

    /** ContainerExecRequest container. */
    public container: string;

    /** ContainerExecRequest command. */
    public command: string[];

    /**
     * Encodes the specified ContainerExecRequest message. Does not implicitly {@link ContainerExecRequest.verify|verify} messages.
     * @param message ContainerExecRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: ContainerExecRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a ContainerExecRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns ContainerExecRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): ContainerExecRequest;
}

/** Represents a ContainerExecResponse. */
export class ContainerExecResponse implements IContainerExecResponse {

    /**
     * Constructs a new ContainerExecResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IContainerExecResponse);

    /** ContainerExecResponse data. */
    public data: string;

    /**
     * Encodes the specified ContainerExecResponse message. Does not implicitly {@link ContainerExecResponse.verify|verify} messages.
     * @param message ContainerExecResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: ContainerExecResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a ContainerExecResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns ContainerExecResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): ContainerExecResponse;
}

/** Represents a ContainerStreamCopyToPodRequest. */
export class ContainerStreamCopyToPodRequest implements IContainerStreamCopyToPodRequest {

    /**
     * Constructs a new ContainerStreamCopyToPodRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IContainerStreamCopyToPodRequest);

    /** ContainerStreamCopyToPodRequest file_name. */
    public file_name: string;

    /** ContainerStreamCopyToPodRequest data. */
    public data: Uint8Array;

    /** ContainerStreamCopyToPodRequest namespace. */
    public namespace: string;

    /** ContainerStreamCopyToPodRequest pod. */
    public pod: string;

    /** ContainerStreamCopyToPodRequest container. */
    public container: string;

    /**
     * Encodes the specified ContainerStreamCopyToPodRequest message. Does not implicitly {@link ContainerStreamCopyToPodRequest.verify|verify} messages.
     * @param message ContainerStreamCopyToPodRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: ContainerStreamCopyToPodRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a ContainerStreamCopyToPodRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns ContainerStreamCopyToPodRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): ContainerStreamCopyToPodRequest;
}

/** Represents a ContainerStreamCopyToPodResponse. */
export class ContainerStreamCopyToPodResponse implements IContainerStreamCopyToPodResponse {

    /**
     * Constructs a new ContainerStreamCopyToPodResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IContainerStreamCopyToPodResponse);

    /** ContainerStreamCopyToPodResponse size. */
    public size: number;

    /** ContainerStreamCopyToPodResponse pod_file_path. */
    public pod_file_path: string;

    /** ContainerStreamCopyToPodResponse output. */
    public output: string;

    /** ContainerStreamCopyToPodResponse pod. */
    public pod: string;

    /** ContainerStreamCopyToPodResponse namespace. */
    public namespace: string;

    /** ContainerStreamCopyToPodResponse container. */
    public container: string;

    /** ContainerStreamCopyToPodResponse filename. */
    public filename: string;

    /**
     * Encodes the specified ContainerStreamCopyToPodResponse message. Does not implicitly {@link ContainerStreamCopyToPodResponse.verify|verify} messages.
     * @param message ContainerStreamCopyToPodResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: ContainerStreamCopyToPodResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a ContainerStreamCopyToPodResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns ContainerStreamCopyToPodResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): ContainerStreamCopyToPodResponse;
}

/** Represents a ContainerIsPodRunningRequest. */
export class ContainerIsPodRunningRequest implements IContainerIsPodRunningRequest {

    /**
     * Constructs a new ContainerIsPodRunningRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IContainerIsPodRunningRequest);

    /** ContainerIsPodRunningRequest namespace. */
    public namespace: string;

    /** ContainerIsPodRunningRequest pod. */
    public pod: string;

    /**
     * Encodes the specified ContainerIsPodRunningRequest message. Does not implicitly {@link ContainerIsPodRunningRequest.verify|verify} messages.
     * @param message ContainerIsPodRunningRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: ContainerIsPodRunningRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a ContainerIsPodRunningRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns ContainerIsPodRunningRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): ContainerIsPodRunningRequest;
}

/** Represents a ContainerIsPodRunningResponse. */
export class ContainerIsPodRunningResponse implements IContainerIsPodRunningResponse {

    /**
     * Constructs a new ContainerIsPodRunningResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IContainerIsPodRunningResponse);

    /** ContainerIsPodRunningResponse running. */
    public running: boolean;

    /** ContainerIsPodRunningResponse reason. */
    public reason: string;

    /**
     * Encodes the specified ContainerIsPodRunningResponse message. Does not implicitly {@link ContainerIsPodRunningResponse.verify|verify} messages.
     * @param message ContainerIsPodRunningResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: ContainerIsPodRunningResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a ContainerIsPodRunningResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns ContainerIsPodRunningResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): ContainerIsPodRunningResponse;
}

/** Represents a ContainerIsPodExistsRequest. */
export class ContainerIsPodExistsRequest implements IContainerIsPodExistsRequest {

    /**
     * Constructs a new ContainerIsPodExistsRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IContainerIsPodExistsRequest);

    /** ContainerIsPodExistsRequest namespace. */
    public namespace: string;

    /** ContainerIsPodExistsRequest pod. */
    public pod: string;

    /**
     * Encodes the specified ContainerIsPodExistsRequest message. Does not implicitly {@link ContainerIsPodExistsRequest.verify|verify} messages.
     * @param message ContainerIsPodExistsRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: ContainerIsPodExistsRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a ContainerIsPodExistsRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns ContainerIsPodExistsRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): ContainerIsPodExistsRequest;
}

/** Represents a ContainerIsPodExistsResponse. */
export class ContainerIsPodExistsResponse implements IContainerIsPodExistsResponse {

    /**
     * Constructs a new ContainerIsPodExistsResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IContainerIsPodExistsResponse);

    /** ContainerIsPodExistsResponse exists. */
    public exists: boolean;

    /**
     * Encodes the specified ContainerIsPodExistsResponse message. Does not implicitly {@link ContainerIsPodExistsResponse.verify|verify} messages.
     * @param message ContainerIsPodExistsResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: ContainerIsPodExistsResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a ContainerIsPodExistsResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns ContainerIsPodExistsResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): ContainerIsPodExistsResponse;
}

/** Represents a ContainerLogRequest. */
export class ContainerLogRequest implements IContainerLogRequest {

    /**
     * Constructs a new ContainerLogRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IContainerLogRequest);

    /** ContainerLogRequest namespace. */
    public namespace: string;

    /** ContainerLogRequest pod. */
    public pod: string;

    /** ContainerLogRequest container. */
    public container: string;

    /**
     * Encodes the specified ContainerLogRequest message. Does not implicitly {@link ContainerLogRequest.verify|verify} messages.
     * @param message ContainerLogRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: ContainerLogRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a ContainerLogRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns ContainerLogRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): ContainerLogRequest;
}

/** Represents a ContainerLogResponse. */
export class ContainerLogResponse implements IContainerLogResponse {

    /**
     * Constructs a new ContainerLogResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IContainerLogResponse);

    /** ContainerLogResponse namespace. */
    public namespace: string;

    /** ContainerLogResponse pod_name. */
    public pod_name: string;

    /** ContainerLogResponse container_name. */
    public container_name: string;

    /** ContainerLogResponse log. */
    public log: string;

    /**
     * Encodes the specified ContainerLogResponse message. Does not implicitly {@link ContainerLogResponse.verify|verify} messages.
     * @param message ContainerLogResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: ContainerLogResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a ContainerLogResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns ContainerLogResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): ContainerLogResponse;
}

/** Represents a ContainerSvc */
export class ContainerSvc extends $protobuf.rpc.Service {

    /**
     * Constructs a new ContainerSvc service.
     * @param rpcImpl RPC implementation
     * @param [requestDelimited=false] Whether requests are length-delimited
     * @param [responseDelimited=false] Whether responses are length-delimited
     */
    constructor(rpcImpl: $protobuf.RPCImpl, requestDelimited?: boolean, responseDelimited?: boolean);

    /**
     * Calls CopyToPod.
     * @param request ContainerCopyToPodRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and ContainerCopyToPodResponse
     */
    public copyToPod(request: ContainerCopyToPodRequest, callback: ContainerSvc.CopyToPodCallback): void;

    /**
     * Calls CopyToPod.
     * @param request ContainerCopyToPodRequest message or plain object
     * @returns Promise
     */
    public copyToPod(request: ContainerCopyToPodRequest): Promise<ContainerCopyToPodResponse>;

    /**
     * Calls Exec.
     * @param request ContainerExecRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and ContainerExecResponse
     */
    public exec(request: ContainerExecRequest, callback: ContainerSvc.ExecCallback): void;

    /**
     * Calls Exec.
     * @param request ContainerExecRequest message or plain object
     * @returns Promise
     */
    public exec(request: ContainerExecRequest): Promise<ContainerExecResponse>;

    /**
     * Calls StreamCopyToPod.
     * @param request ContainerStreamCopyToPodRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and ContainerStreamCopyToPodResponse
     */
    public streamCopyToPod(request: ContainerStreamCopyToPodRequest, callback: ContainerSvc.StreamCopyToPodCallback): void;

    /**
     * Calls StreamCopyToPod.
     * @param request ContainerStreamCopyToPodRequest message or plain object
     * @returns Promise
     */
    public streamCopyToPod(request: ContainerStreamCopyToPodRequest): Promise<ContainerStreamCopyToPodResponse>;

    /**
     * Calls IsPodRunning.
     * @param request ContainerIsPodRunningRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and ContainerIsPodRunningResponse
     */
    public isPodRunning(request: ContainerIsPodRunningRequest, callback: ContainerSvc.IsPodRunningCallback): void;

    /**
     * Calls IsPodRunning.
     * @param request ContainerIsPodRunningRequest message or plain object
     * @returns Promise
     */
    public isPodRunning(request: ContainerIsPodRunningRequest): Promise<ContainerIsPodRunningResponse>;

    /**
     * Calls IsPodExists.
     * @param request ContainerIsPodExistsRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and ContainerIsPodExistsResponse
     */
    public isPodExists(request: ContainerIsPodExistsRequest, callback: ContainerSvc.IsPodExistsCallback): void;

    /**
     * Calls IsPodExists.
     * @param request ContainerIsPodExistsRequest message or plain object
     * @returns Promise
     */
    public isPodExists(request: ContainerIsPodExistsRequest): Promise<ContainerIsPodExistsResponse>;

    /**
     * Calls ContainerLog.
     * @param request ContainerLogRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and ContainerLogResponse
     */
    public containerLog(request: ContainerLogRequest, callback: ContainerSvc.ContainerLogCallback): void;

    /**
     * Calls ContainerLog.
     * @param request ContainerLogRequest message or plain object
     * @returns Promise
     */
    public containerLog(request: ContainerLogRequest): Promise<ContainerLogResponse>;

    /**
     * Calls StreamContainerLog.
     * @param request ContainerLogRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and ContainerLogResponse
     */
    public streamContainerLog(request: ContainerLogRequest, callback: ContainerSvc.StreamContainerLogCallback): void;

    /**
     * Calls StreamContainerLog.
     * @param request ContainerLogRequest message or plain object
     * @returns Promise
     */
    public streamContainerLog(request: ContainerLogRequest): Promise<ContainerLogResponse>;
}

export namespace ContainerSvc {

    /**
     * Callback as used by {@link ContainerSvc#copyToPod}.
     * @param error Error, if any
     * @param [response] ContainerCopyToPodResponse
     */
    type CopyToPodCallback = (error: (Error|null), response?: ContainerCopyToPodResponse) => void;

    /**
     * Callback as used by {@link ContainerSvc#exec}.
     * @param error Error, if any
     * @param [response] ContainerExecResponse
     */
    type ExecCallback = (error: (Error|null), response?: ContainerExecResponse) => void;

    /**
     * Callback as used by {@link ContainerSvc#streamCopyToPod}.
     * @param error Error, if any
     * @param [response] ContainerStreamCopyToPodResponse
     */
    type StreamCopyToPodCallback = (error: (Error|null), response?: ContainerStreamCopyToPodResponse) => void;

    /**
     * Callback as used by {@link ContainerSvc#isPodRunning}.
     * @param error Error, if any
     * @param [response] ContainerIsPodRunningResponse
     */
    type IsPodRunningCallback = (error: (Error|null), response?: ContainerIsPodRunningResponse) => void;

    /**
     * Callback as used by {@link ContainerSvc#isPodExists}.
     * @param error Error, if any
     * @param [response] ContainerIsPodExistsResponse
     */
    type IsPodExistsCallback = (error: (Error|null), response?: ContainerIsPodExistsResponse) => void;

    /**
     * Callback as used by {@link ContainerSvc#containerLog}.
     * @param error Error, if any
     * @param [response] ContainerLogResponse
     */
    type ContainerLogCallback = (error: (Error|null), response?: ContainerLogResponse) => void;

    /**
     * Callback as used by {@link ContainerSvc#streamContainerLog}.
     * @param error Error, if any
     * @param [response] ContainerLogResponse
     */
    type StreamContainerLogCallback = (error: (Error|null), response?: ContainerLogResponse) => void;
}

/** Represents a ServiceEndpoint. */
export class ServiceEndpoint implements IServiceEndpoint {

    /**
     * Constructs a new ServiceEndpoint.
     * @param [properties] Properties to set
     */
    constructor(properties?: IServiceEndpoint);

    /** ServiceEndpoint name. */
    public name: string;

    /** ServiceEndpoint url. */
    public url: string;

    /** ServiceEndpoint port_name. */
    public port_name: string;

    /**
     * Encodes the specified ServiceEndpoint message. Does not implicitly {@link ServiceEndpoint.verify|verify} messages.
     * @param message ServiceEndpoint message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: ServiceEndpoint, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a ServiceEndpoint message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns ServiceEndpoint
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): ServiceEndpoint;
}

/** Represents an EndpointInNamespaceRequest. */
export class EndpointInNamespaceRequest implements IEndpointInNamespaceRequest {

    /**
     * Constructs a new EndpointInNamespaceRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IEndpointInNamespaceRequest);

    /** EndpointInNamespaceRequest namespace_id. */
    public namespace_id: number;

    /**
     * Encodes the specified EndpointInNamespaceRequest message. Does not implicitly {@link EndpointInNamespaceRequest.verify|verify} messages.
     * @param message EndpointInNamespaceRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: EndpointInNamespaceRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes an EndpointInNamespaceRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns EndpointInNamespaceRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): EndpointInNamespaceRequest;
}

/** Represents an EndpointInNamespaceResponse. */
export class EndpointInNamespaceResponse implements IEndpointInNamespaceResponse {

    /**
     * Constructs a new EndpointInNamespaceResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IEndpointInNamespaceResponse);

    /** EndpointInNamespaceResponse items. */
    public items: ServiceEndpoint[];

    /**
     * Encodes the specified EndpointInNamespaceResponse message. Does not implicitly {@link EndpointInNamespaceResponse.verify|verify} messages.
     * @param message EndpointInNamespaceResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: EndpointInNamespaceResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes an EndpointInNamespaceResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns EndpointInNamespaceResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): EndpointInNamespaceResponse;
}

/** Represents an EndpointInProjectRequest. */
export class EndpointInProjectRequest implements IEndpointInProjectRequest {

    /**
     * Constructs a new EndpointInProjectRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IEndpointInProjectRequest);

    /** EndpointInProjectRequest project_id. */
    public project_id: number;

    /**
     * Encodes the specified EndpointInProjectRequest message. Does not implicitly {@link EndpointInProjectRequest.verify|verify} messages.
     * @param message EndpointInProjectRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: EndpointInProjectRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes an EndpointInProjectRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns EndpointInProjectRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): EndpointInProjectRequest;
}

/** Represents an EndpointInProjectResponse. */
export class EndpointInProjectResponse implements IEndpointInProjectResponse {

    /**
     * Constructs a new EndpointInProjectResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IEndpointInProjectResponse);

    /** EndpointInProjectResponse items. */
    public items: ServiceEndpoint[];

    /**
     * Encodes the specified EndpointInProjectResponse message. Does not implicitly {@link EndpointInProjectResponse.verify|verify} messages.
     * @param message EndpointInProjectResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: EndpointInProjectResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes an EndpointInProjectResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns EndpointInProjectResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): EndpointInProjectResponse;
}

/** Represents an Endpoint */
export class Endpoint extends $protobuf.rpc.Service {

    /**
     * Constructs a new Endpoint service.
     * @param rpcImpl RPC implementation
     * @param [requestDelimited=false] Whether requests are length-delimited
     * @param [responseDelimited=false] Whether responses are length-delimited
     */
    constructor(rpcImpl: $protobuf.RPCImpl, requestDelimited?: boolean, responseDelimited?: boolean);

    /**
     * Calls InNamespace.
     * @param request EndpointInNamespaceRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and EndpointInNamespaceResponse
     */
    public inNamespace(request: EndpointInNamespaceRequest, callback: Endpoint.InNamespaceCallback): void;

    /**
     * Calls InNamespace.
     * @param request EndpointInNamespaceRequest message or plain object
     * @returns Promise
     */
    public inNamespace(request: EndpointInNamespaceRequest): Promise<EndpointInNamespaceResponse>;

    /**
     * Calls InProject.
     * @param request EndpointInProjectRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and EndpointInProjectResponse
     */
    public inProject(request: EndpointInProjectRequest, callback: Endpoint.InProjectCallback): void;

    /**
     * Calls InProject.
     * @param request EndpointInProjectRequest message or plain object
     * @returns Promise
     */
    public inProject(request: EndpointInProjectRequest): Promise<EndpointInProjectResponse>;
}

export namespace Endpoint {

    /**
     * Callback as used by {@link Endpoint#inNamespace}.
     * @param error Error, if any
     * @param [response] EndpointInNamespaceResponse
     */
    type InNamespaceCallback = (error: (Error|null), response?: EndpointInNamespaceResponse) => void;

    /**
     * Callback as used by {@link Endpoint#inProject}.
     * @param error Error, if any
     * @param [response] EndpointInProjectResponse
     */
    type InProjectCallback = (error: (Error|null), response?: EndpointInProjectResponse) => void;
}

/** ActionType enum. */
export enum ActionType {
    Unknown = 0,
    Create = 1,
    Update = 2,
    Delete = 3,
    Upload = 4,
    Download = 5,
    DryRun = 6,
    Shell = 7
}

/** Represents an EventListRequest. */
export class EventListRequest implements IEventListRequest {

    /**
     * Constructs a new EventListRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IEventListRequest);

    /** EventListRequest page. */
    public page: number;

    /** EventListRequest page_size. */
    public page_size: number;

    /**
     * Encodes the specified EventListRequest message. Does not implicitly {@link EventListRequest.verify|verify} messages.
     * @param message EventListRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: EventListRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes an EventListRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns EventListRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): EventListRequest;
}

/** Represents an EventListItem. */
export class EventListItem implements IEventListItem {

    /**
     * Constructs a new EventListItem.
     * @param [properties] Properties to set
     */
    constructor(properties?: IEventListItem);

    /** EventListItem id. */
    public id: number;

    /** EventListItem action. */
    public action: ActionType;

    /** EventListItem username. */
    public username: string;

    /** EventListItem message. */
    public message: string;

    /** EventListItem old. */
    public old: string;

    /** EventListItem new. */
    public new: string;

    /** EventListItem event_at. */
    public event_at: string;

    /** EventListItem file_id. */
    public file_id: number;

    /**
     * Encodes the specified EventListItem message. Does not implicitly {@link EventListItem.verify|verify} messages.
     * @param message EventListItem message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: EventListItem, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes an EventListItem message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns EventListItem
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): EventListItem;
}

/** Represents an EventListResponse. */
export class EventListResponse implements IEventListResponse {

    /**
     * Constructs a new EventListResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IEventListResponse);

    /** EventListResponse page. */
    public page: number;

    /** EventListResponse page_size. */
    public page_size: number;

    /** EventListResponse items. */
    public items: EventListItem[];

    /** EventListResponse count. */
    public count: number;

    /**
     * Encodes the specified EventListResponse message. Does not implicitly {@link EventListResponse.verify|verify} messages.
     * @param message EventListResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: EventListResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes an EventListResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns EventListResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): EventListResponse;
}

/** Represents an Event */
export class Event extends $protobuf.rpc.Service {

    /**
     * Constructs a new Event service.
     * @param rpcImpl RPC implementation
     * @param [requestDelimited=false] Whether requests are length-delimited
     * @param [responseDelimited=false] Whether responses are length-delimited
     */
    constructor(rpcImpl: $protobuf.RPCImpl, requestDelimited?: boolean, responseDelimited?: boolean);

    /**
     * Calls List.
     * @param request EventListRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and EventListResponse
     */
    public list(request: EventListRequest, callback: Event.ListCallback): void;

    /**
     * Calls List.
     * @param request EventListRequest message or plain object
     * @returns Promise
     */
    public list(request: EventListRequest): Promise<EventListResponse>;
}

export namespace Event {

    /**
     * Callback as used by {@link Event#list}.
     * @param error Error, if any
     * @param [response] EventListResponse
     */
    type ListCallback = (error: (Error|null), response?: EventListResponse) => void;
}

/** Represents a FileDeleteRequest. */
export class FileDeleteRequest implements IFileDeleteRequest {

    /**
     * Constructs a new FileDeleteRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IFileDeleteRequest);

    /** FileDeleteRequest id. */
    public id: number;

    /**
     * Encodes the specified FileDeleteRequest message. Does not implicitly {@link FileDeleteRequest.verify|verify} messages.
     * @param message FileDeleteRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: FileDeleteRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a FileDeleteRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns FileDeleteRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): FileDeleteRequest;
}

/** Represents a FileDeleteResponse. */
export class FileDeleteResponse implements IFileDeleteResponse {

    /**
     * Constructs a new FileDeleteResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IFileDeleteResponse);

    /** FileDeleteResponse file. */
    public file?: (File|null);

    /**
     * Encodes the specified FileDeleteResponse message. Does not implicitly {@link FileDeleteResponse.verify|verify} messages.
     * @param message FileDeleteResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: FileDeleteResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a FileDeleteResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns FileDeleteResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): FileDeleteResponse;
}

/** Represents a DeleteUndocumentedFilesRequest. */
export class DeleteUndocumentedFilesRequest implements IDeleteUndocumentedFilesRequest {

    /**
     * Constructs a new DeleteUndocumentedFilesRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IDeleteUndocumentedFilesRequest);

    /**
     * Encodes the specified DeleteUndocumentedFilesRequest message. Does not implicitly {@link DeleteUndocumentedFilesRequest.verify|verify} messages.
     * @param message DeleteUndocumentedFilesRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: DeleteUndocumentedFilesRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a DeleteUndocumentedFilesRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns DeleteUndocumentedFilesRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): DeleteUndocumentedFilesRequest;
}

/** Represents a File. */
export class File implements IFile {

    /**
     * Constructs a new File.
     * @param [properties] Properties to set
     */
    constructor(properties?: IFile);

    /** File path. */
    public path: string;

    /** File humanize_size. */
    public humanize_size: string;

    /** File size. */
    public size: number;

    /** File upload_by. */
    public upload_by: string;

    /**
     * Encodes the specified File message. Does not implicitly {@link File.verify|verify} messages.
     * @param message File message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: File, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a File message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns File
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): File;
}

/** Represents a DeleteUndocumentedFilesResponse. */
export class DeleteUndocumentedFilesResponse implements IDeleteUndocumentedFilesResponse {

    /**
     * Constructs a new DeleteUndocumentedFilesResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IDeleteUndocumentedFilesResponse);

    /** DeleteUndocumentedFilesResponse items. */
    public items: File[];

    /**
     * Encodes the specified DeleteUndocumentedFilesResponse message. Does not implicitly {@link DeleteUndocumentedFilesResponse.verify|verify} messages.
     * @param message DeleteUndocumentedFilesResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: DeleteUndocumentedFilesResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a DeleteUndocumentedFilesResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns DeleteUndocumentedFilesResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): DeleteUndocumentedFilesResponse;
}

/** Represents a DiskInfoRequest. */
export class DiskInfoRequest implements IDiskInfoRequest {

    /**
     * Constructs a new DiskInfoRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IDiskInfoRequest);

    /**
     * Encodes the specified DiskInfoRequest message. Does not implicitly {@link DiskInfoRequest.verify|verify} messages.
     * @param message DiskInfoRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: DiskInfoRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a DiskInfoRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns DiskInfoRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): DiskInfoRequest;
}

/** Represents a DiskInfoResponse. */
export class DiskInfoResponse implements IDiskInfoResponse {

    /**
     * Constructs a new DiskInfoResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IDiskInfoResponse);

    /** DiskInfoResponse usage. */
    public usage: number;

    /** DiskInfoResponse humanize_usage. */
    public humanize_usage: string;

    /**
     * Encodes the specified DiskInfoResponse message. Does not implicitly {@link DiskInfoResponse.verify|verify} messages.
     * @param message DiskInfoResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: DiskInfoResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a DiskInfoResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns DiskInfoResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): DiskInfoResponse;
}

/** Represents a FileListRequest. */
export class FileListRequest implements IFileListRequest {

    /**
     * Constructs a new FileListRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IFileListRequest);

    /** FileListRequest page. */
    public page: number;

    /** FileListRequest page_size. */
    public page_size: number;

    /** FileListRequest without_deleted. */
    public without_deleted: boolean;

    /**
     * Encodes the specified FileListRequest message. Does not implicitly {@link FileListRequest.verify|verify} messages.
     * @param message FileListRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: FileListRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a FileListRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns FileListRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): FileListRequest;
}

/** Represents a FileListResponse. */
export class FileListResponse implements IFileListResponse {

    /**
     * Constructs a new FileListResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IFileListResponse);

    /** FileListResponse page. */
    public page: number;

    /** FileListResponse page_size. */
    public page_size: number;

    /** FileListResponse items. */
    public items: FileModel[];

    /** FileListResponse count. */
    public count: number;

    /**
     * Encodes the specified FileListResponse message. Does not implicitly {@link FileListResponse.verify|verify} messages.
     * @param message FileListResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: FileListResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a FileListResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns FileListResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): FileListResponse;
}

/** Represents a FileSvc */
export class FileSvc extends $protobuf.rpc.Service {

    /**
     * Constructs a new FileSvc service.
     * @param rpcImpl RPC implementation
     * @param [requestDelimited=false] Whether requests are length-delimited
     * @param [responseDelimited=false] Whether responses are length-delimited
     */
    constructor(rpcImpl: $protobuf.RPCImpl, requestDelimited?: boolean, responseDelimited?: boolean);

    /**
     * Calls List.
     * @param request FileListRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and FileListResponse
     */
    public list(request: FileListRequest, callback: FileSvc.ListCallback): void;

    /**
     * Calls List.
     * @param request FileListRequest message or plain object
     * @returns Promise
     */
    public list(request: FileListRequest): Promise<FileListResponse>;

    /**
     * Calls Delete.
     * @param request FileDeleteRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and FileDeleteResponse
     */
    public delete(request: FileDeleteRequest, callback: FileSvc.DeleteCallback): void;

    /**
     * Calls Delete.
     * @param request FileDeleteRequest message or plain object
     * @returns Promise
     */
    public delete(request: FileDeleteRequest): Promise<FileDeleteResponse>;

    /**
     * Calls DeleteUndocumentedFiles.
     * @param request DeleteUndocumentedFilesRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and DeleteUndocumentedFilesResponse
     */
    public deleteUndocumentedFiles(request: DeleteUndocumentedFilesRequest, callback: FileSvc.DeleteUndocumentedFilesCallback): void;

    /**
     * Calls DeleteUndocumentedFiles.
     * @param request DeleteUndocumentedFilesRequest message or plain object
     * @returns Promise
     */
    public deleteUndocumentedFiles(request: DeleteUndocumentedFilesRequest): Promise<DeleteUndocumentedFilesResponse>;

    /**
     * Calls DiskInfo.
     * @param request DiskInfoRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and DiskInfoResponse
     */
    public diskInfo(request: DiskInfoRequest, callback: FileSvc.DiskInfoCallback): void;

    /**
     * Calls DiskInfo.
     * @param request DiskInfoRequest message or plain object
     * @returns Promise
     */
    public diskInfo(request: DiskInfoRequest): Promise<DiskInfoResponse>;
}

export namespace FileSvc {

    /**
     * Callback as used by {@link FileSvc#list}.
     * @param error Error, if any
     * @param [response] FileListResponse
     */
    type ListCallback = (error: (Error|null), response?: FileListResponse) => void;

    /**
     * Callback as used by {@link FileSvc#delete_}.
     * @param error Error, if any
     * @param [response] FileDeleteResponse
     */
    type DeleteCallback = (error: (Error|null), response?: FileDeleteResponse) => void;

    /**
     * Callback as used by {@link FileSvc#deleteUndocumentedFiles}.
     * @param error Error, if any
     * @param [response] DeleteUndocumentedFilesResponse
     */
    type DeleteUndocumentedFilesCallback = (error: (Error|null), response?: DeleteUndocumentedFilesResponse) => void;

    /**
     * Callback as used by {@link FileSvc#diskInfo}.
     * @param error Error, if any
     * @param [response] DiskInfoResponse
     */
    type DiskInfoCallback = (error: (Error|null), response?: DiskInfoResponse) => void;
}

/** Represents a GitEnableProjectRequest. */
export class GitEnableProjectRequest implements IGitEnableProjectRequest {

    /**
     * Constructs a new GitEnableProjectRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IGitEnableProjectRequest);

    /** GitEnableProjectRequest git_project_id. */
    public git_project_id: string;

    /**
     * Encodes the specified GitEnableProjectRequest message. Does not implicitly {@link GitEnableProjectRequest.verify|verify} messages.
     * @param message GitEnableProjectRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: GitEnableProjectRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a GitEnableProjectRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns GitEnableProjectRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): GitEnableProjectRequest;
}

/** Represents a GitDisableProjectRequest. */
export class GitDisableProjectRequest implements IGitDisableProjectRequest {

    /**
     * Constructs a new GitDisableProjectRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IGitDisableProjectRequest);

    /** GitDisableProjectRequest git_project_id. */
    public git_project_id: string;

    /**
     * Encodes the specified GitDisableProjectRequest message. Does not implicitly {@link GitDisableProjectRequest.verify|verify} messages.
     * @param message GitDisableProjectRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: GitDisableProjectRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a GitDisableProjectRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns GitDisableProjectRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): GitDisableProjectRequest;
}

/** Represents a GitProjectItem. */
export class GitProjectItem implements IGitProjectItem {

    /**
     * Constructs a new GitProjectItem.
     * @param [properties] Properties to set
     */
    constructor(properties?: IGitProjectItem);

    /** GitProjectItem id. */
    public id: number;

    /** GitProjectItem name. */
    public name: string;

    /** GitProjectItem path. */
    public path: string;

    /** GitProjectItem web_url. */
    public web_url: string;

    /** GitProjectItem avatar_url. */
    public avatar_url: string;

    /** GitProjectItem description. */
    public description: string;

    /** GitProjectItem enabled. */
    public enabled: boolean;

    /** GitProjectItem global_enabled. */
    public global_enabled: boolean;

    /**
     * Encodes the specified GitProjectItem message. Does not implicitly {@link GitProjectItem.verify|verify} messages.
     * @param message GitProjectItem message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: GitProjectItem, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a GitProjectItem message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns GitProjectItem
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): GitProjectItem;
}

/** Represents a GitAllProjectsResponse. */
export class GitAllProjectsResponse implements IGitAllProjectsResponse {

    /**
     * Constructs a new GitAllProjectsResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IGitAllProjectsResponse);

    /** GitAllProjectsResponse items. */
    public items: GitProjectItem[];

    /**
     * Encodes the specified GitAllProjectsResponse message. Does not implicitly {@link GitAllProjectsResponse.verify|verify} messages.
     * @param message GitAllProjectsResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: GitAllProjectsResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a GitAllProjectsResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns GitAllProjectsResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): GitAllProjectsResponse;
}

/** Represents a GitOption. */
export class GitOption implements IGitOption {

    /**
     * Constructs a new GitOption.
     * @param [properties] Properties to set
     */
    constructor(properties?: IGitOption);

    /** GitOption value. */
    public value: string;

    /** GitOption label. */
    public label: string;

    /** GitOption type. */
    public type: string;

    /** GitOption isLeaf. */
    public isLeaf: boolean;

    /** GitOption gitProjectId. */
    public gitProjectId: string;

    /** GitOption branch. */
    public branch: string;

    /**
     * Encodes the specified GitOption message. Does not implicitly {@link GitOption.verify|verify} messages.
     * @param message GitOption message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: GitOption, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a GitOption message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns GitOption
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): GitOption;
}

/** Represents a GitProjectOptionsResponse. */
export class GitProjectOptionsResponse implements IGitProjectOptionsResponse {

    /**
     * Constructs a new GitProjectOptionsResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IGitProjectOptionsResponse);

    /** GitProjectOptionsResponse items. */
    public items: GitOption[];

    /**
     * Encodes the specified GitProjectOptionsResponse message. Does not implicitly {@link GitProjectOptionsResponse.verify|verify} messages.
     * @param message GitProjectOptionsResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: GitProjectOptionsResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a GitProjectOptionsResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns GitProjectOptionsResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): GitProjectOptionsResponse;
}

/** Represents a GitBranchOptionsRequest. */
export class GitBranchOptionsRequest implements IGitBranchOptionsRequest {

    /**
     * Constructs a new GitBranchOptionsRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IGitBranchOptionsRequest);

    /** GitBranchOptionsRequest git_project_id. */
    public git_project_id: string;

    /** GitBranchOptionsRequest all. */
    public all: boolean;

    /**
     * Encodes the specified GitBranchOptionsRequest message. Does not implicitly {@link GitBranchOptionsRequest.verify|verify} messages.
     * @param message GitBranchOptionsRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: GitBranchOptionsRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a GitBranchOptionsRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns GitBranchOptionsRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): GitBranchOptionsRequest;
}

/** Represents a GitBranchOptionsResponse. */
export class GitBranchOptionsResponse implements IGitBranchOptionsResponse {

    /**
     * Constructs a new GitBranchOptionsResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IGitBranchOptionsResponse);

    /** GitBranchOptionsResponse items. */
    public items: GitOption[];

    /**
     * Encodes the specified GitBranchOptionsResponse message. Does not implicitly {@link GitBranchOptionsResponse.verify|verify} messages.
     * @param message GitBranchOptionsResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: GitBranchOptionsResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a GitBranchOptionsResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns GitBranchOptionsResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): GitBranchOptionsResponse;
}

/** Represents a GitCommitOptionsRequest. */
export class GitCommitOptionsRequest implements IGitCommitOptionsRequest {

    /**
     * Constructs a new GitCommitOptionsRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IGitCommitOptionsRequest);

    /** GitCommitOptionsRequest git_project_id. */
    public git_project_id: string;

    /** GitCommitOptionsRequest branch. */
    public branch: string;

    /**
     * Encodes the specified GitCommitOptionsRequest message. Does not implicitly {@link GitCommitOptionsRequest.verify|verify} messages.
     * @param message GitCommitOptionsRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: GitCommitOptionsRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a GitCommitOptionsRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns GitCommitOptionsRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): GitCommitOptionsRequest;
}

/** Represents a GitCommitOptionsResponse. */
export class GitCommitOptionsResponse implements IGitCommitOptionsResponse {

    /**
     * Constructs a new GitCommitOptionsResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IGitCommitOptionsResponse);

    /** GitCommitOptionsResponse items. */
    public items: GitOption[];

    /**
     * Encodes the specified GitCommitOptionsResponse message. Does not implicitly {@link GitCommitOptionsResponse.verify|verify} messages.
     * @param message GitCommitOptionsResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: GitCommitOptionsResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a GitCommitOptionsResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns GitCommitOptionsResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): GitCommitOptionsResponse;
}

/** Represents a GitCommitRequest. */
export class GitCommitRequest implements IGitCommitRequest {

    /**
     * Constructs a new GitCommitRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IGitCommitRequest);

    /** GitCommitRequest git_project_id. */
    public git_project_id: string;

    /** GitCommitRequest branch. */
    public branch: string;

    /** GitCommitRequest commit. */
    public commit: string;

    /**
     * Encodes the specified GitCommitRequest message. Does not implicitly {@link GitCommitRequest.verify|verify} messages.
     * @param message GitCommitRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: GitCommitRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a GitCommitRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns GitCommitRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): GitCommitRequest;
}

/** Represents a GitCommitResponse. */
export class GitCommitResponse implements IGitCommitResponse {

    /**
     * Constructs a new GitCommitResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IGitCommitResponse);

    /** GitCommitResponse id. */
    public id: string;

    /** GitCommitResponse short_id. */
    public short_id: string;

    /** GitCommitResponse git_project_id. */
    public git_project_id: string;

    /** GitCommitResponse label. */
    public label: string;

    /** GitCommitResponse title. */
    public title: string;

    /** GitCommitResponse branch. */
    public branch: string;

    /** GitCommitResponse author_name. */
    public author_name: string;

    /** GitCommitResponse author_email. */
    public author_email: string;

    /** GitCommitResponse committer_name. */
    public committer_name: string;

    /** GitCommitResponse committer_email. */
    public committer_email: string;

    /** GitCommitResponse web_url. */
    public web_url: string;

    /** GitCommitResponse message. */
    public message: string;

    /** GitCommitResponse committed_date. */
    public committed_date: string;

    /** GitCommitResponse created_at. */
    public created_at: string;

    /**
     * Encodes the specified GitCommitResponse message. Does not implicitly {@link GitCommitResponse.verify|verify} messages.
     * @param message GitCommitResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: GitCommitResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a GitCommitResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns GitCommitResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): GitCommitResponse;
}

/** Represents a GitPipelineInfoRequest. */
export class GitPipelineInfoRequest implements IGitPipelineInfoRequest {

    /**
     * Constructs a new GitPipelineInfoRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IGitPipelineInfoRequest);

    /** GitPipelineInfoRequest git_project_id. */
    public git_project_id: string;

    /** GitPipelineInfoRequest branch. */
    public branch: string;

    /** GitPipelineInfoRequest commit. */
    public commit: string;

    /**
     * Encodes the specified GitPipelineInfoRequest message. Does not implicitly {@link GitPipelineInfoRequest.verify|verify} messages.
     * @param message GitPipelineInfoRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: GitPipelineInfoRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a GitPipelineInfoRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns GitPipelineInfoRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): GitPipelineInfoRequest;
}

/** Represents a GitPipelineInfoResponse. */
export class GitPipelineInfoResponse implements IGitPipelineInfoResponse {

    /**
     * Constructs a new GitPipelineInfoResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IGitPipelineInfoResponse);

    /** GitPipelineInfoResponse status. */
    public status: string;

    /** GitPipelineInfoResponse web_url. */
    public web_url: string;

    /**
     * Encodes the specified GitPipelineInfoResponse message. Does not implicitly {@link GitPipelineInfoResponse.verify|verify} messages.
     * @param message GitPipelineInfoResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: GitPipelineInfoResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a GitPipelineInfoResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns GitPipelineInfoResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): GitPipelineInfoResponse;
}

/** Represents a GitConfigFileRequest. */
export class GitConfigFileRequest implements IGitConfigFileRequest {

    /**
     * Constructs a new GitConfigFileRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IGitConfigFileRequest);

    /** GitConfigFileRequest git_project_id. */
    public git_project_id: string;

    /** GitConfigFileRequest branch. */
    public branch: string;

    /**
     * Encodes the specified GitConfigFileRequest message. Does not implicitly {@link GitConfigFileRequest.verify|verify} messages.
     * @param message GitConfigFileRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: GitConfigFileRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a GitConfigFileRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns GitConfigFileRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): GitConfigFileRequest;
}

/** Represents a GitConfigFileResponse. */
export class GitConfigFileResponse implements IGitConfigFileResponse {

    /**
     * Constructs a new GitConfigFileResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IGitConfigFileResponse);

    /** GitConfigFileResponse data. */
    public data: string;

    /** GitConfigFileResponse type. */
    public type: string;

    /** GitConfigFileResponse elements. */
    public elements: Element[];

    /**
     * Encodes the specified GitConfigFileResponse message. Does not implicitly {@link GitConfigFileResponse.verify|verify} messages.
     * @param message GitConfigFileResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: GitConfigFileResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a GitConfigFileResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns GitConfigFileResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): GitConfigFileResponse;
}

/** Represents a GitEnableProjectResponse. */
export class GitEnableProjectResponse implements IGitEnableProjectResponse {

    /**
     * Constructs a new GitEnableProjectResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IGitEnableProjectResponse);

    /**
     * Encodes the specified GitEnableProjectResponse message. Does not implicitly {@link GitEnableProjectResponse.verify|verify} messages.
     * @param message GitEnableProjectResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: GitEnableProjectResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a GitEnableProjectResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns GitEnableProjectResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): GitEnableProjectResponse;
}

/** Represents a GitDisableProjectResponse. */
export class GitDisableProjectResponse implements IGitDisableProjectResponse {

    /**
     * Constructs a new GitDisableProjectResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IGitDisableProjectResponse);

    /**
     * Encodes the specified GitDisableProjectResponse message. Does not implicitly {@link GitDisableProjectResponse.verify|verify} messages.
     * @param message GitDisableProjectResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: GitDisableProjectResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a GitDisableProjectResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns GitDisableProjectResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): GitDisableProjectResponse;
}

/** Represents a GitAllProjectsRequest. */
export class GitAllProjectsRequest implements IGitAllProjectsRequest {

    /**
     * Constructs a new GitAllProjectsRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IGitAllProjectsRequest);

    /**
     * Encodes the specified GitAllProjectsRequest message. Does not implicitly {@link GitAllProjectsRequest.verify|verify} messages.
     * @param message GitAllProjectsRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: GitAllProjectsRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a GitAllProjectsRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns GitAllProjectsRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): GitAllProjectsRequest;
}

/** Represents a GitProjectOptionsRequest. */
export class GitProjectOptionsRequest implements IGitProjectOptionsRequest {

    /**
     * Constructs a new GitProjectOptionsRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IGitProjectOptionsRequest);

    /**
     * Encodes the specified GitProjectOptionsRequest message. Does not implicitly {@link GitProjectOptionsRequest.verify|verify} messages.
     * @param message GitProjectOptionsRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: GitProjectOptionsRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a GitProjectOptionsRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns GitProjectOptionsRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): GitProjectOptionsRequest;
}

/** Represents a Git */
export class Git extends $protobuf.rpc.Service {

    /**
     * Constructs a new Git service.
     * @param rpcImpl RPC implementation
     * @param [requestDelimited=false] Whether requests are length-delimited
     * @param [responseDelimited=false] Whether responses are length-delimited
     */
    constructor(rpcImpl: $protobuf.RPCImpl, requestDelimited?: boolean, responseDelimited?: boolean);

    /**
     * Calls EnableProject.
     * @param request GitEnableProjectRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and GitEnableProjectResponse
     */
    public enableProject(request: GitEnableProjectRequest, callback: Git.EnableProjectCallback): void;

    /**
     * Calls EnableProject.
     * @param request GitEnableProjectRequest message or plain object
     * @returns Promise
     */
    public enableProject(request: GitEnableProjectRequest): Promise<GitEnableProjectResponse>;

    /**
     * Calls DisableProject.
     * @param request GitDisableProjectRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and GitDisableProjectResponse
     */
    public disableProject(request: GitDisableProjectRequest, callback: Git.DisableProjectCallback): void;

    /**
     * Calls DisableProject.
     * @param request GitDisableProjectRequest message or plain object
     * @returns Promise
     */
    public disableProject(request: GitDisableProjectRequest): Promise<GitDisableProjectResponse>;

    /**
     * Calls All.
     * @param request GitAllProjectsRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and GitAllProjectsResponse
     */
    public all(request: GitAllProjectsRequest, callback: Git.AllCallback): void;

    /**
     * Calls All.
     * @param request GitAllProjectsRequest message or plain object
     * @returns Promise
     */
    public all(request: GitAllProjectsRequest): Promise<GitAllProjectsResponse>;

    /**
     * Calls ProjectOptions.
     * @param request GitProjectOptionsRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and GitProjectOptionsResponse
     */
    public projectOptions(request: GitProjectOptionsRequest, callback: Git.ProjectOptionsCallback): void;

    /**
     * Calls ProjectOptions.
     * @param request GitProjectOptionsRequest message or plain object
     * @returns Promise
     */
    public projectOptions(request: GitProjectOptionsRequest): Promise<GitProjectOptionsResponse>;

    /**
     * Calls BranchOptions.
     * @param request GitBranchOptionsRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and GitBranchOptionsResponse
     */
    public branchOptions(request: GitBranchOptionsRequest, callback: Git.BranchOptionsCallback): void;

    /**
     * Calls BranchOptions.
     * @param request GitBranchOptionsRequest message or plain object
     * @returns Promise
     */
    public branchOptions(request: GitBranchOptionsRequest): Promise<GitBranchOptionsResponse>;

    /**
     * Calls CommitOptions.
     * @param request GitCommitOptionsRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and GitCommitOptionsResponse
     */
    public commitOptions(request: GitCommitOptionsRequest, callback: Git.CommitOptionsCallback): void;

    /**
     * Calls CommitOptions.
     * @param request GitCommitOptionsRequest message or plain object
     * @returns Promise
     */
    public commitOptions(request: GitCommitOptionsRequest): Promise<GitCommitOptionsResponse>;

    /**
     * Calls Commit.
     * @param request GitCommitRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and GitCommitResponse
     */
    public commit(request: GitCommitRequest, callback: Git.CommitCallback): void;

    /**
     * Calls Commit.
     * @param request GitCommitRequest message or plain object
     * @returns Promise
     */
    public commit(request: GitCommitRequest): Promise<GitCommitResponse>;

    /**
     * Calls PipelineInfo.
     * @param request GitPipelineInfoRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and GitPipelineInfoResponse
     */
    public pipelineInfo(request: GitPipelineInfoRequest, callback: Git.PipelineInfoCallback): void;

    /**
     * Calls PipelineInfo.
     * @param request GitPipelineInfoRequest message or plain object
     * @returns Promise
     */
    public pipelineInfo(request: GitPipelineInfoRequest): Promise<GitPipelineInfoResponse>;

    /**
     * Calls MarsConfigFile.
     * @param request GitConfigFileRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and GitConfigFileResponse
     */
    public marsConfigFile(request: GitConfigFileRequest, callback: Git.MarsConfigFileCallback): void;

    /**
     * Calls MarsConfigFile.
     * @param request GitConfigFileRequest message or plain object
     * @returns Promise
     */
    public marsConfigFile(request: GitConfigFileRequest): Promise<GitConfigFileResponse>;
}

export namespace Git {

    /**
     * Callback as used by {@link Git#enableProject}.
     * @param error Error, if any
     * @param [response] GitEnableProjectResponse
     */
    type EnableProjectCallback = (error: (Error|null), response?: GitEnableProjectResponse) => void;

    /**
     * Callback as used by {@link Git#disableProject}.
     * @param error Error, if any
     * @param [response] GitDisableProjectResponse
     */
    type DisableProjectCallback = (error: (Error|null), response?: GitDisableProjectResponse) => void;

    /**
     * Callback as used by {@link Git#all}.
     * @param error Error, if any
     * @param [response] GitAllProjectsResponse
     */
    type AllCallback = (error: (Error|null), response?: GitAllProjectsResponse) => void;

    /**
     * Callback as used by {@link Git#projectOptions}.
     * @param error Error, if any
     * @param [response] GitProjectOptionsResponse
     */
    type ProjectOptionsCallback = (error: (Error|null), response?: GitProjectOptionsResponse) => void;

    /**
     * Callback as used by {@link Git#branchOptions}.
     * @param error Error, if any
     * @param [response] GitBranchOptionsResponse
     */
    type BranchOptionsCallback = (error: (Error|null), response?: GitBranchOptionsResponse) => void;

    /**
     * Callback as used by {@link Git#commitOptions}.
     * @param error Error, if any
     * @param [response] GitCommitOptionsResponse
     */
    type CommitOptionsCallback = (error: (Error|null), response?: GitCommitOptionsResponse) => void;

    /**
     * Callback as used by {@link Git#commit}.
     * @param error Error, if any
     * @param [response] GitCommitResponse
     */
    type CommitCallback = (error: (Error|null), response?: GitCommitResponse) => void;

    /**
     * Callback as used by {@link Git#pipelineInfo}.
     * @param error Error, if any
     * @param [response] GitPipelineInfoResponse
     */
    type PipelineInfoCallback = (error: (Error|null), response?: GitPipelineInfoResponse) => void;

    /**
     * Callback as used by {@link Git#marsConfigFile}.
     * @param error Error, if any
     * @param [response] GitConfigFileResponse
     */
    type MarsConfigFileCallback = (error: (Error|null), response?: GitConfigFileResponse) => void;
}

/** Represents a GitConfigShowRequest. */
export class GitConfigShowRequest implements IGitConfigShowRequest {

    /**
     * Constructs a new GitConfigShowRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IGitConfigShowRequest);

    /** GitConfigShowRequest git_project_id. */
    public git_project_id: number;

    /** GitConfigShowRequest branch. */
    public branch: string;

    /**
     * Encodes the specified GitConfigShowRequest message. Does not implicitly {@link GitConfigShowRequest.verify|verify} messages.
     * @param message GitConfigShowRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: GitConfigShowRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a GitConfigShowRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns GitConfigShowRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): GitConfigShowRequest;
}

/** Represents a GitConfigShowResponse. */
export class GitConfigShowResponse implements IGitConfigShowResponse {

    /**
     * Constructs a new GitConfigShowResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IGitConfigShowResponse);

    /** GitConfigShowResponse branch. */
    public branch: string;

    /** GitConfigShowResponse config. */
    public config?: (MarsConfig|null);

    /**
     * Encodes the specified GitConfigShowResponse message. Does not implicitly {@link GitConfigShowResponse.verify|verify} messages.
     * @param message GitConfigShowResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: GitConfigShowResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a GitConfigShowResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns GitConfigShowResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): GitConfigShowResponse;
}

/** Represents a GitConfigGlobalConfigRequest. */
export class GitConfigGlobalConfigRequest implements IGitConfigGlobalConfigRequest {

    /**
     * Constructs a new GitConfigGlobalConfigRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IGitConfigGlobalConfigRequest);

    /** GitConfigGlobalConfigRequest git_project_id. */
    public git_project_id: number;

    /**
     * Encodes the specified GitConfigGlobalConfigRequest message. Does not implicitly {@link GitConfigGlobalConfigRequest.verify|verify} messages.
     * @param message GitConfigGlobalConfigRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: GitConfigGlobalConfigRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a GitConfigGlobalConfigRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns GitConfigGlobalConfigRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): GitConfigGlobalConfigRequest;
}

/** Represents a GitConfigGlobalConfigResponse. */
export class GitConfigGlobalConfigResponse implements IGitConfigGlobalConfigResponse {

    /**
     * Constructs a new GitConfigGlobalConfigResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IGitConfigGlobalConfigResponse);

    /** GitConfigGlobalConfigResponse enabled. */
    public enabled: boolean;

    /** GitConfigGlobalConfigResponse config. */
    public config?: (MarsConfig|null);

    /**
     * Encodes the specified GitConfigGlobalConfigResponse message. Does not implicitly {@link GitConfigGlobalConfigResponse.verify|verify} messages.
     * @param message GitConfigGlobalConfigResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: GitConfigGlobalConfigResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a GitConfigGlobalConfigResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns GitConfigGlobalConfigResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): GitConfigGlobalConfigResponse;
}

/** Represents a GitConfigUpdateRequest. */
export class GitConfigUpdateRequest implements IGitConfigUpdateRequest {

    /**
     * Constructs a new GitConfigUpdateRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IGitConfigUpdateRequest);

    /** GitConfigUpdateRequest git_project_id. */
    public git_project_id: number;

    /** GitConfigUpdateRequest config. */
    public config?: (MarsConfig|null);

    /**
     * Encodes the specified GitConfigUpdateRequest message. Does not implicitly {@link GitConfigUpdateRequest.verify|verify} messages.
     * @param message GitConfigUpdateRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: GitConfigUpdateRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a GitConfigUpdateRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns GitConfigUpdateRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): GitConfigUpdateRequest;
}

/** Represents a GitConfigUpdateResponse. */
export class GitConfigUpdateResponse implements IGitConfigUpdateResponse {

    /**
     * Constructs a new GitConfigUpdateResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IGitConfigUpdateResponse);

    /** GitConfigUpdateResponse config. */
    public config?: (MarsConfig|null);

    /**
     * Encodes the specified GitConfigUpdateResponse message. Does not implicitly {@link GitConfigUpdateResponse.verify|verify} messages.
     * @param message GitConfigUpdateResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: GitConfigUpdateResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a GitConfigUpdateResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns GitConfigUpdateResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): GitConfigUpdateResponse;
}

/** Represents a GitConfigToggleGlobalStatusRequest. */
export class GitConfigToggleGlobalStatusRequest implements IGitConfigToggleGlobalStatusRequest {

    /**
     * Constructs a new GitConfigToggleGlobalStatusRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IGitConfigToggleGlobalStatusRequest);

    /** GitConfigToggleGlobalStatusRequest git_project_id. */
    public git_project_id: number;

    /** GitConfigToggleGlobalStatusRequest enabled. */
    public enabled: boolean;

    /**
     * Encodes the specified GitConfigToggleGlobalStatusRequest message. Does not implicitly {@link GitConfigToggleGlobalStatusRequest.verify|verify} messages.
     * @param message GitConfigToggleGlobalStatusRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: GitConfigToggleGlobalStatusRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a GitConfigToggleGlobalStatusRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns GitConfigToggleGlobalStatusRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): GitConfigToggleGlobalStatusRequest;
}

/** Represents a GitConfigDefaultChartValuesRequest. */
export class GitConfigDefaultChartValuesRequest implements IGitConfigDefaultChartValuesRequest {

    /**
     * Constructs a new GitConfigDefaultChartValuesRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IGitConfigDefaultChartValuesRequest);

    /** GitConfigDefaultChartValuesRequest git_project_id. */
    public git_project_id: number;

    /** GitConfigDefaultChartValuesRequest branch. */
    public branch: string;

    /**
     * Encodes the specified GitConfigDefaultChartValuesRequest message. Does not implicitly {@link GitConfigDefaultChartValuesRequest.verify|verify} messages.
     * @param message GitConfigDefaultChartValuesRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: GitConfigDefaultChartValuesRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a GitConfigDefaultChartValuesRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns GitConfigDefaultChartValuesRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): GitConfigDefaultChartValuesRequest;
}

/** Represents a GitConfigDefaultChartValuesResponse. */
export class GitConfigDefaultChartValuesResponse implements IGitConfigDefaultChartValuesResponse {

    /**
     * Constructs a new GitConfigDefaultChartValuesResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IGitConfigDefaultChartValuesResponse);

    /** GitConfigDefaultChartValuesResponse value. */
    public value: string;

    /**
     * Encodes the specified GitConfigDefaultChartValuesResponse message. Does not implicitly {@link GitConfigDefaultChartValuesResponse.verify|verify} messages.
     * @param message GitConfigDefaultChartValuesResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: GitConfigDefaultChartValuesResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a GitConfigDefaultChartValuesResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns GitConfigDefaultChartValuesResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): GitConfigDefaultChartValuesResponse;
}

/** Represents a GitConfigToggleGlobalStatusResponse. */
export class GitConfigToggleGlobalStatusResponse implements IGitConfigToggleGlobalStatusResponse {

    /**
     * Constructs a new GitConfigToggleGlobalStatusResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IGitConfigToggleGlobalStatusResponse);

    /**
     * Encodes the specified GitConfigToggleGlobalStatusResponse message. Does not implicitly {@link GitConfigToggleGlobalStatusResponse.verify|verify} messages.
     * @param message GitConfigToggleGlobalStatusResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: GitConfigToggleGlobalStatusResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a GitConfigToggleGlobalStatusResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns GitConfigToggleGlobalStatusResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): GitConfigToggleGlobalStatusResponse;
}

/** Represents a GitConfig */
export class GitConfig extends $protobuf.rpc.Service {

    /**
     * Constructs a new GitConfig service.
     * @param rpcImpl RPC implementation
     * @param [requestDelimited=false] Whether requests are length-delimited
     * @param [responseDelimited=false] Whether responses are length-delimited
     */
    constructor(rpcImpl: $protobuf.RPCImpl, requestDelimited?: boolean, responseDelimited?: boolean);

    /**
     * Calls Show.
     * @param request GitConfigShowRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and GitConfigShowResponse
     */
    public show(request: GitConfigShowRequest, callback: GitConfig.ShowCallback): void;

    /**
     * Calls Show.
     * @param request GitConfigShowRequest message or plain object
     * @returns Promise
     */
    public show(request: GitConfigShowRequest): Promise<GitConfigShowResponse>;

    /**
     * Calls GlobalConfig.
     * @param request GitConfigGlobalConfigRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and GitConfigGlobalConfigResponse
     */
    public globalConfig(request: GitConfigGlobalConfigRequest, callback: GitConfig.GlobalConfigCallback): void;

    /**
     * Calls GlobalConfig.
     * @param request GitConfigGlobalConfigRequest message or plain object
     * @returns Promise
     */
    public globalConfig(request: GitConfigGlobalConfigRequest): Promise<GitConfigGlobalConfigResponse>;

    /**
     * Calls ToggleGlobalStatus.
     * @param request GitConfigToggleGlobalStatusRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and GitConfigToggleGlobalStatusResponse
     */
    public toggleGlobalStatus(request: GitConfigToggleGlobalStatusRequest, callback: GitConfig.ToggleGlobalStatusCallback): void;

    /**
     * Calls ToggleGlobalStatus.
     * @param request GitConfigToggleGlobalStatusRequest message or plain object
     * @returns Promise
     */
    public toggleGlobalStatus(request: GitConfigToggleGlobalStatusRequest): Promise<GitConfigToggleGlobalStatusResponse>;

    /**
     * Calls Update.
     * @param request GitConfigUpdateRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and GitConfigUpdateResponse
     */
    public update(request: GitConfigUpdateRequest, callback: GitConfig.UpdateCallback): void;

    /**
     * Calls Update.
     * @param request GitConfigUpdateRequest message or plain object
     * @returns Promise
     */
    public update(request: GitConfigUpdateRequest): Promise<GitConfigUpdateResponse>;

    /**
     * Calls GetDefaultChartValues.
     * @param request GitConfigDefaultChartValuesRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and GitConfigDefaultChartValuesResponse
     */
    public getDefaultChartValues(request: GitConfigDefaultChartValuesRequest, callback: GitConfig.GetDefaultChartValuesCallback): void;

    /**
     * Calls GetDefaultChartValues.
     * @param request GitConfigDefaultChartValuesRequest message or plain object
     * @returns Promise
     */
    public getDefaultChartValues(request: GitConfigDefaultChartValuesRequest): Promise<GitConfigDefaultChartValuesResponse>;
}

export namespace GitConfig {

    /**
     * Callback as used by {@link GitConfig#show}.
     * @param error Error, if any
     * @param [response] GitConfigShowResponse
     */
    type ShowCallback = (error: (Error|null), response?: GitConfigShowResponse) => void;

    /**
     * Callback as used by {@link GitConfig#globalConfig}.
     * @param error Error, if any
     * @param [response] GitConfigGlobalConfigResponse
     */
    type GlobalConfigCallback = (error: (Error|null), response?: GitConfigGlobalConfigResponse) => void;

    /**
     * Callback as used by {@link GitConfig#toggleGlobalStatus}.
     * @param error Error, if any
     * @param [response] GitConfigToggleGlobalStatusResponse
     */
    type ToggleGlobalStatusCallback = (error: (Error|null), response?: GitConfigToggleGlobalStatusResponse) => void;

    /**
     * Callback as used by {@link GitConfig#update}.
     * @param error Error, if any
     * @param [response] GitConfigUpdateResponse
     */
    type UpdateCallback = (error: (Error|null), response?: GitConfigUpdateResponse) => void;

    /**
     * Callback as used by {@link GitConfig#getDefaultChartValues}.
     * @param error Error, if any
     * @param [response] GitConfigDefaultChartValuesResponse
     */
    type GetDefaultChartValuesCallback = (error: (Error|null), response?: GitConfigDefaultChartValuesResponse) => void;
}

/** Represents a MarsConfig. */
export class MarsConfig implements IMarsConfig {

    /**
     * Constructs a new MarsConfig.
     * @param [properties] Properties to set
     */
    constructor(properties?: IMarsConfig);

    /** MarsConfig config_file. */
    public config_file: string;

    /** MarsConfig config_file_values. */
    public config_file_values: string;

    /** MarsConfig config_field. */
    public config_field: string;

    /** MarsConfig is_simple_env. */
    public is_simple_env: boolean;

    /** MarsConfig config_file_type. */
    public config_file_type: string;

    /** MarsConfig local_chart_path. */
    public local_chart_path: string;

    /** MarsConfig branches. */
    public branches: string[];

    /** MarsConfig values_yaml. */
    public values_yaml: string;

    /** MarsConfig elements. */
    public elements: Element[];

    /**
     * Encodes the specified MarsConfig message. Does not implicitly {@link MarsConfig.verify|verify} messages.
     * @param message MarsConfig message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: MarsConfig, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a MarsConfig message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns MarsConfig
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): MarsConfig;
}

/** ElementType enum. */
export enum ElementType {
    ElementTypeUnknown = 0,
    ElementTypeInput = 1,
    ElementTypeInputNumber = 2,
    ElementTypeSelect = 3,
    ElementTypeRadio = 4,
    ElementTypeSwitch = 5
}

/** Represents an Element. */
export class Element implements IElement {

    /**
     * Constructs a new Element.
     * @param [properties] Properties to set
     */
    constructor(properties?: IElement);

    /** Element path. */
    public path: string;

    /** Element type. */
    public type: ElementType;

    /** Element default. */
    public default: string;

    /** Element description. */
    public description: string;

    /** Element select_values. */
    public select_values: string[];

    /**
     * Encodes the specified Element message. Does not implicitly {@link Element.verify|verify} messages.
     * @param message Element message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: Element, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes an Element message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns Element
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): Element;
}

/** Represents a MetricsTopPodRequest. */
export class MetricsTopPodRequest implements IMetricsTopPodRequest {

    /**
     * Constructs a new MetricsTopPodRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IMetricsTopPodRequest);

    /** MetricsTopPodRequest namespace. */
    public namespace: string;

    /** MetricsTopPodRequest pod. */
    public pod: string;

    /**
     * Encodes the specified MetricsTopPodRequest message. Does not implicitly {@link MetricsTopPodRequest.verify|verify} messages.
     * @param message MetricsTopPodRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: MetricsTopPodRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a MetricsTopPodRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns MetricsTopPodRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): MetricsTopPodRequest;
}

/** Represents a MetricsTopPodResponse. */
export class MetricsTopPodResponse implements IMetricsTopPodResponse {

    /**
     * Constructs a new MetricsTopPodResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IMetricsTopPodResponse);

    /** MetricsTopPodResponse cpu. */
    public cpu: number;

    /** MetricsTopPodResponse memory. */
    public memory: number;

    /** MetricsTopPodResponse humanize_cpu. */
    public humanize_cpu: string;

    /** MetricsTopPodResponse humanize_memory. */
    public humanize_memory: string;

    /** MetricsTopPodResponse time. */
    public time: string;

    /** MetricsTopPodResponse length. */
    public length: number;

    /**
     * Encodes the specified MetricsTopPodResponse message. Does not implicitly {@link MetricsTopPodResponse.verify|verify} messages.
     * @param message MetricsTopPodResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: MetricsTopPodResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a MetricsTopPodResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns MetricsTopPodResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): MetricsTopPodResponse;
}

/** Represents a MetricsCpuMemoryInNamespaceRequest. */
export class MetricsCpuMemoryInNamespaceRequest implements IMetricsCpuMemoryInNamespaceRequest {

    /**
     * Constructs a new MetricsCpuMemoryInNamespaceRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IMetricsCpuMemoryInNamespaceRequest);

    /** MetricsCpuMemoryInNamespaceRequest namespace_id. */
    public namespace_id: number;

    /**
     * Encodes the specified MetricsCpuMemoryInNamespaceRequest message. Does not implicitly {@link MetricsCpuMemoryInNamespaceRequest.verify|verify} messages.
     * @param message MetricsCpuMemoryInNamespaceRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: MetricsCpuMemoryInNamespaceRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a MetricsCpuMemoryInNamespaceRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns MetricsCpuMemoryInNamespaceRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): MetricsCpuMemoryInNamespaceRequest;
}

/** Represents a MetricsCpuMemoryInNamespaceResponse. */
export class MetricsCpuMemoryInNamespaceResponse implements IMetricsCpuMemoryInNamespaceResponse {

    /**
     * Constructs a new MetricsCpuMemoryInNamespaceResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IMetricsCpuMemoryInNamespaceResponse);

    /** MetricsCpuMemoryInNamespaceResponse cpu. */
    public cpu: string;

    /** MetricsCpuMemoryInNamespaceResponse memory. */
    public memory: string;

    /**
     * Encodes the specified MetricsCpuMemoryInNamespaceResponse message. Does not implicitly {@link MetricsCpuMemoryInNamespaceResponse.verify|verify} messages.
     * @param message MetricsCpuMemoryInNamespaceResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: MetricsCpuMemoryInNamespaceResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a MetricsCpuMemoryInNamespaceResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns MetricsCpuMemoryInNamespaceResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): MetricsCpuMemoryInNamespaceResponse;
}

/** Represents a MetricsCpuMemoryInProjectRequest. */
export class MetricsCpuMemoryInProjectRequest implements IMetricsCpuMemoryInProjectRequest {

    /**
     * Constructs a new MetricsCpuMemoryInProjectRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IMetricsCpuMemoryInProjectRequest);

    /** MetricsCpuMemoryInProjectRequest project_id. */
    public project_id: number;

    /**
     * Encodes the specified MetricsCpuMemoryInProjectRequest message. Does not implicitly {@link MetricsCpuMemoryInProjectRequest.verify|verify} messages.
     * @param message MetricsCpuMemoryInProjectRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: MetricsCpuMemoryInProjectRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a MetricsCpuMemoryInProjectRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns MetricsCpuMemoryInProjectRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): MetricsCpuMemoryInProjectRequest;
}

/** Represents a MetricsCpuMemoryInProjectResponse. */
export class MetricsCpuMemoryInProjectResponse implements IMetricsCpuMemoryInProjectResponse {

    /**
     * Constructs a new MetricsCpuMemoryInProjectResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IMetricsCpuMemoryInProjectResponse);

    /** MetricsCpuMemoryInProjectResponse cpu. */
    public cpu: string;

    /** MetricsCpuMemoryInProjectResponse memory. */
    public memory: string;

    /**
     * Encodes the specified MetricsCpuMemoryInProjectResponse message. Does not implicitly {@link MetricsCpuMemoryInProjectResponse.verify|verify} messages.
     * @param message MetricsCpuMemoryInProjectResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: MetricsCpuMemoryInProjectResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a MetricsCpuMemoryInProjectResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns MetricsCpuMemoryInProjectResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): MetricsCpuMemoryInProjectResponse;
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
     * Calls CpuMemoryInNamespace.
     * @param request MetricsCpuMemoryInNamespaceRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and MetricsCpuMemoryInNamespaceResponse
     */
    public cpuMemoryInNamespace(request: MetricsCpuMemoryInNamespaceRequest, callback: Metrics.CpuMemoryInNamespaceCallback): void;

    /**
     * Calls CpuMemoryInNamespace.
     * @param request MetricsCpuMemoryInNamespaceRequest message or plain object
     * @returns Promise
     */
    public cpuMemoryInNamespace(request: MetricsCpuMemoryInNamespaceRequest): Promise<MetricsCpuMemoryInNamespaceResponse>;

    /**
     * Calls CpuMemoryInProject.
     * @param request MetricsCpuMemoryInProjectRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and MetricsCpuMemoryInProjectResponse
     */
    public cpuMemoryInProject(request: MetricsCpuMemoryInProjectRequest, callback: Metrics.CpuMemoryInProjectCallback): void;

    /**
     * Calls CpuMemoryInProject.
     * @param request MetricsCpuMemoryInProjectRequest message or plain object
     * @returns Promise
     */
    public cpuMemoryInProject(request: MetricsCpuMemoryInProjectRequest): Promise<MetricsCpuMemoryInProjectResponse>;

    /**
     * Calls TopPod.
     * @param request MetricsTopPodRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and MetricsTopPodResponse
     */
    public topPod(request: MetricsTopPodRequest, callback: Metrics.TopPodCallback): void;

    /**
     * Calls TopPod.
     * @param request MetricsTopPodRequest message or plain object
     * @returns Promise
     */
    public topPod(request: MetricsTopPodRequest): Promise<MetricsTopPodResponse>;

    /**
     * Calls StreamTopPod.
     * @param request MetricsTopPodRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and MetricsTopPodResponse
     */
    public streamTopPod(request: MetricsTopPodRequest, callback: Metrics.StreamTopPodCallback): void;

    /**
     * Calls StreamTopPod.
     * @param request MetricsTopPodRequest message or plain object
     * @returns Promise
     */
    public streamTopPod(request: MetricsTopPodRequest): Promise<MetricsTopPodResponse>;
}

export namespace Metrics {

    /**
     * Callback as used by {@link Metrics#cpuMemoryInNamespace}.
     * @param error Error, if any
     * @param [response] MetricsCpuMemoryInNamespaceResponse
     */
    type CpuMemoryInNamespaceCallback = (error: (Error|null), response?: MetricsCpuMemoryInNamespaceResponse) => void;

    /**
     * Callback as used by {@link Metrics#cpuMemoryInProject}.
     * @param error Error, if any
     * @param [response] MetricsCpuMemoryInProjectResponse
     */
    type CpuMemoryInProjectCallback = (error: (Error|null), response?: MetricsCpuMemoryInProjectResponse) => void;

    /**
     * Callback as used by {@link Metrics#topPod}.
     * @param error Error, if any
     * @param [response] MetricsTopPodResponse
     */
    type TopPodCallback = (error: (Error|null), response?: MetricsTopPodResponse) => void;

    /**
     * Callback as used by {@link Metrics#streamTopPod}.
     * @param error Error, if any
     * @param [response] MetricsTopPodResponse
     */
    type StreamTopPodCallback = (error: (Error|null), response?: MetricsTopPodResponse) => void;
}

/** Represents a GitProjectModel. */
export class GitProjectModel implements IGitProjectModel {

    /**
     * Constructs a new GitProjectModel.
     * @param [properties] Properties to set
     */
    constructor(properties?: IGitProjectModel);

    /** GitProjectModel id. */
    public id: number;

    /** GitProjectModel default_branch. */
    public default_branch: string;

    /** GitProjectModel name. */
    public name: string;

    /** GitProjectModel git_project_id. */
    public git_project_id: number;

    /** GitProjectModel enabled. */
    public enabled: boolean;

    /** GitProjectModel global_enabled. */
    public global_enabled: boolean;

    /** GitProjectModel global_config. */
    public global_config: string;

    /** GitProjectModel created_at. */
    public created_at: string;

    /** GitProjectModel updated_at. */
    public updated_at: string;

    /**
     * Encodes the specified GitProjectModel message. Does not implicitly {@link GitProjectModel.verify|verify} messages.
     * @param message GitProjectModel message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: GitProjectModel, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a GitProjectModel message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns GitProjectModel
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): GitProjectModel;
}

/** Represents a NamespaceModel. */
export class NamespaceModel implements INamespaceModel {

    /**
     * Constructs a new NamespaceModel.
     * @param [properties] Properties to set
     */
    constructor(properties?: INamespaceModel);

    /** NamespaceModel id. */
    public id: number;

    /** NamespaceModel name. */
    public name: string;

    /** NamespaceModel image_pull_secrets. */
    public image_pull_secrets: string[];

    /** NamespaceModel created_at. */
    public created_at: string;

    /** NamespaceModel updated_at. */
    public updated_at: string;

    /** NamespaceModel projects. */
    public projects: ProjectModel[];

    /**
     * Encodes the specified NamespaceModel message. Does not implicitly {@link NamespaceModel.verify|verify} messages.
     * @param message NamespaceModel message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: NamespaceModel, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a NamespaceModel message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns NamespaceModel
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): NamespaceModel;
}

/** Represents a ProjectModel. */
export class ProjectModel implements IProjectModel {

    /**
     * Constructs a new ProjectModel.
     * @param [properties] Properties to set
     */
    constructor(properties?: IProjectModel);

    /** ProjectModel id. */
    public id: number;

    /** ProjectModel name. */
    public name: string;

    /** ProjectModel git_project_id. */
    public git_project_id: number;

    /** ProjectModel git_branch. */
    public git_branch: string;

    /** ProjectModel git_commit. */
    public git_commit: string;

    /** ProjectModel config. */
    public config: string;

    /** ProjectModel override_values. */
    public override_values: string;

    /** ProjectModel docker_image. */
    public docker_image: string;

    /** ProjectModel pod_selectors. */
    public pod_selectors: string;

    /** ProjectModel namespace_id. */
    public namespace_id: number;

    /** ProjectModel atomic. */
    public atomic: boolean;

    /** ProjectModel created_at. */
    public created_at: string;

    /** ProjectModel updated_at. */
    public updated_at: string;

    /** ProjectModel extra_values. */
    public extra_values: string;

    /** ProjectModel namespace. */
    public namespace?: (NamespaceModel|null);

    /**
     * Encodes the specified ProjectModel message. Does not implicitly {@link ProjectModel.verify|verify} messages.
     * @param message ProjectModel message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: ProjectModel, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a ProjectModel message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns ProjectModel
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): ProjectModel;
}

/** Represents a FileModel. */
export class FileModel implements IFileModel {

    /**
     * Constructs a new FileModel.
     * @param [properties] Properties to set
     */
    constructor(properties?: IFileModel);

    /** FileModel id. */
    public id: number;

    /** FileModel path. */
    public path: string;

    /** FileModel size. */
    public size: number;

    /** FileModel username. */
    public username: string;

    /** FileModel namespace. */
    public namespace: string;

    /** FileModel pod. */
    public pod: string;

    /** FileModel container. */
    public container: string;

    /** FileModel container_path. */
    public container_path: string;

    /** FileModel created_at. */
    public created_at: string;

    /** FileModel updated_at. */
    public updated_at: string;

    /** FileModel deleted_at. */
    public deleted_at: string;

    /** FileModel is_deleted. */
    public is_deleted: boolean;

    /**
     * Encodes the specified FileModel message. Does not implicitly {@link FileModel.verify|verify} messages.
     * @param message FileModel message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: FileModel, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a FileModel message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns FileModel
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): FileModel;
}

/** Represents a NamespaceCreateRequest. */
export class NamespaceCreateRequest implements INamespaceCreateRequest {

    /**
     * Constructs a new NamespaceCreateRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: INamespaceCreateRequest);

    /** NamespaceCreateRequest namespace. */
    public namespace: string;

    /**
     * Encodes the specified NamespaceCreateRequest message. Does not implicitly {@link NamespaceCreateRequest.verify|verify} messages.
     * @param message NamespaceCreateRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: NamespaceCreateRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a NamespaceCreateRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns NamespaceCreateRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): NamespaceCreateRequest;
}

/** Represents a NamespaceShowRequest. */
export class NamespaceShowRequest implements INamespaceShowRequest {

    /**
     * Constructs a new NamespaceShowRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: INamespaceShowRequest);

    /** NamespaceShowRequest namespace_id. */
    public namespace_id: number;

    /**
     * Encodes the specified NamespaceShowRequest message. Does not implicitly {@link NamespaceShowRequest.verify|verify} messages.
     * @param message NamespaceShowRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: NamespaceShowRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a NamespaceShowRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns NamespaceShowRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): NamespaceShowRequest;
}

/** Represents a NamespaceDeleteRequest. */
export class NamespaceDeleteRequest implements INamespaceDeleteRequest {

    /**
     * Constructs a new NamespaceDeleteRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: INamespaceDeleteRequest);

    /** NamespaceDeleteRequest namespace_id. */
    public namespace_id: number;

    /**
     * Encodes the specified NamespaceDeleteRequest message. Does not implicitly {@link NamespaceDeleteRequest.verify|verify} messages.
     * @param message NamespaceDeleteRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: NamespaceDeleteRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a NamespaceDeleteRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns NamespaceDeleteRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): NamespaceDeleteRequest;
}

/** Represents a NamespaceIsExistsRequest. */
export class NamespaceIsExistsRequest implements INamespaceIsExistsRequest {

    /**
     * Constructs a new NamespaceIsExistsRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: INamespaceIsExistsRequest);

    /** NamespaceIsExistsRequest name. */
    public name: string;

    /**
     * Encodes the specified NamespaceIsExistsRequest message. Does not implicitly {@link NamespaceIsExistsRequest.verify|verify} messages.
     * @param message NamespaceIsExistsRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: NamespaceIsExistsRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a NamespaceIsExistsRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns NamespaceIsExistsRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): NamespaceIsExistsRequest;
}

/** Represents a NamespaceSimpleProject. */
export class NamespaceSimpleProject implements INamespaceSimpleProject {

    /**
     * Constructs a new NamespaceSimpleProject.
     * @param [properties] Properties to set
     */
    constructor(properties?: INamespaceSimpleProject);

    /** NamespaceSimpleProject id. */
    public id: number;

    /** NamespaceSimpleProject name. */
    public name: string;

    /** NamespaceSimpleProject status. */
    public status: string;

    /**
     * Encodes the specified NamespaceSimpleProject message. Does not implicitly {@link NamespaceSimpleProject.verify|verify} messages.
     * @param message NamespaceSimpleProject message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: NamespaceSimpleProject, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a NamespaceSimpleProject message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns NamespaceSimpleProject
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): NamespaceSimpleProject;
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
    public created_at: string;

    /** NamespaceItem updated_at. */
    public updated_at: string;

    /** NamespaceItem projects. */
    public projects: NamespaceSimpleProject[];

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

/** Represents a NamespaceAllResponse. */
export class NamespaceAllResponse implements INamespaceAllResponse {

    /**
     * Constructs a new NamespaceAllResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: INamespaceAllResponse);

    /** NamespaceAllResponse items. */
    public items: NamespaceItem[];

    /**
     * Encodes the specified NamespaceAllResponse message. Does not implicitly {@link NamespaceAllResponse.verify|verify} messages.
     * @param message NamespaceAllResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: NamespaceAllResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a NamespaceAllResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns NamespaceAllResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): NamespaceAllResponse;
}

/** Represents a NamespaceCreateResponse. */
export class NamespaceCreateResponse implements INamespaceCreateResponse {

    /**
     * Constructs a new NamespaceCreateResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: INamespaceCreateResponse);

    /** NamespaceCreateResponse id. */
    public id: number;

    /** NamespaceCreateResponse name. */
    public name: string;

    /** NamespaceCreateResponse image_pull_secrets. */
    public image_pull_secrets: string[];

    /** NamespaceCreateResponse created_at. */
    public created_at: string;

    /** NamespaceCreateResponse updated_at. */
    public updated_at: string;

    /**
     * Encodes the specified NamespaceCreateResponse message. Does not implicitly {@link NamespaceCreateResponse.verify|verify} messages.
     * @param message NamespaceCreateResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: NamespaceCreateResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a NamespaceCreateResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns NamespaceCreateResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): NamespaceCreateResponse;
}

/** Represents a NamespaceShowResponse. */
export class NamespaceShowResponse implements INamespaceShowResponse {

    /**
     * Constructs a new NamespaceShowResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: INamespaceShowResponse);

    /** NamespaceShowResponse id. */
    public id: number;

    /** NamespaceShowResponse name. */
    public name: string;

    /** NamespaceShowResponse image_pull_secrets. */
    public image_pull_secrets: string[];

    /** NamespaceShowResponse created_at. */
    public created_at: string;

    /** NamespaceShowResponse updated_at. */
    public updated_at: string;

    /** NamespaceShowResponse projects. */
    public projects: ProjectModel[];

    /**
     * Encodes the specified NamespaceShowResponse message. Does not implicitly {@link NamespaceShowResponse.verify|verify} messages.
     * @param message NamespaceShowResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: NamespaceShowResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a NamespaceShowResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns NamespaceShowResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): NamespaceShowResponse;
}

/** Represents a NamespaceIsExistsResponse. */
export class NamespaceIsExistsResponse implements INamespaceIsExistsResponse {

    /**
     * Constructs a new NamespaceIsExistsResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: INamespaceIsExistsResponse);

    /** NamespaceIsExistsResponse exists. */
    public exists: boolean;

    /** NamespaceIsExistsResponse id. */
    public id: number;

    /**
     * Encodes the specified NamespaceIsExistsResponse message. Does not implicitly {@link NamespaceIsExistsResponse.verify|verify} messages.
     * @param message NamespaceIsExistsResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: NamespaceIsExistsResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a NamespaceIsExistsResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns NamespaceIsExistsResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): NamespaceIsExistsResponse;
}

/** Represents a NamespaceAllRequest. */
export class NamespaceAllRequest implements INamespaceAllRequest {

    /**
     * Constructs a new NamespaceAllRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: INamespaceAllRequest);

    /**
     * Encodes the specified NamespaceAllRequest message. Does not implicitly {@link NamespaceAllRequest.verify|verify} messages.
     * @param message NamespaceAllRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: NamespaceAllRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a NamespaceAllRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns NamespaceAllRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): NamespaceAllRequest;
}

/** Represents a NamespaceDeleteResponse. */
export class NamespaceDeleteResponse implements INamespaceDeleteResponse {

    /**
     * Constructs a new NamespaceDeleteResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: INamespaceDeleteResponse);

    /**
     * Encodes the specified NamespaceDeleteResponse message. Does not implicitly {@link NamespaceDeleteResponse.verify|verify} messages.
     * @param message NamespaceDeleteResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: NamespaceDeleteResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a NamespaceDeleteResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns NamespaceDeleteResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): NamespaceDeleteResponse;
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
     * Calls All.
     * @param request NamespaceAllRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and NamespaceAllResponse
     */
    public all(request: NamespaceAllRequest, callback: Namespace.AllCallback): void;

    /**
     * Calls All.
     * @param request NamespaceAllRequest message or plain object
     * @returns Promise
     */
    public all(request: NamespaceAllRequest): Promise<NamespaceAllResponse>;

    /**
     * Calls Create.
     * @param request NamespaceCreateRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and NamespaceCreateResponse
     */
    public create(request: NamespaceCreateRequest, callback: Namespace.CreateCallback): void;

    /**
     * Calls Create.
     * @param request NamespaceCreateRequest message or plain object
     * @returns Promise
     */
    public create(request: NamespaceCreateRequest): Promise<NamespaceCreateResponse>;

    /**
     * Calls Show.
     * @param request NamespaceShowRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and NamespaceShowResponse
     */
    public show(request: NamespaceShowRequest, callback: Namespace.ShowCallback): void;

    /**
     * Calls Show.
     * @param request NamespaceShowRequest message or plain object
     * @returns Promise
     */
    public show(request: NamespaceShowRequest): Promise<NamespaceShowResponse>;

    /**
     * Calls Delete.
     * @param request NamespaceDeleteRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and NamespaceDeleteResponse
     */
    public delete(request: NamespaceDeleteRequest, callback: Namespace.DeleteCallback): void;

    /**
     * Calls Delete.
     * @param request NamespaceDeleteRequest message or plain object
     * @returns Promise
     */
    public delete(request: NamespaceDeleteRequest): Promise<NamespaceDeleteResponse>;

    /**
     * Calls IsExists.
     * @param request NamespaceIsExistsRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and NamespaceIsExistsResponse
     */
    public isExists(request: NamespaceIsExistsRequest, callback: Namespace.IsExistsCallback): void;

    /**
     * Calls IsExists.
     * @param request NamespaceIsExistsRequest message or plain object
     * @returns Promise
     */
    public isExists(request: NamespaceIsExistsRequest): Promise<NamespaceIsExistsResponse>;
}

export namespace Namespace {

    /**
     * Callback as used by {@link Namespace#all}.
     * @param error Error, if any
     * @param [response] NamespaceAllResponse
     */
    type AllCallback = (error: (Error|null), response?: NamespaceAllResponse) => void;

    /**
     * Callback as used by {@link Namespace#create}.
     * @param error Error, if any
     * @param [response] NamespaceCreateResponse
     */
    type CreateCallback = (error: (Error|null), response?: NamespaceCreateResponse) => void;

    /**
     * Callback as used by {@link Namespace#show}.
     * @param error Error, if any
     * @param [response] NamespaceShowResponse
     */
    type ShowCallback = (error: (Error|null), response?: NamespaceShowResponse) => void;

    /**
     * Callback as used by {@link Namespace#delete_}.
     * @param error Error, if any
     * @param [response] NamespaceDeleteResponse
     */
    type DeleteCallback = (error: (Error|null), response?: NamespaceDeleteResponse) => void;

    /**
     * Callback as used by {@link Namespace#isExists}.
     * @param error Error, if any
     * @param [response] NamespaceIsExistsResponse
     */
    type IsExistsCallback = (error: (Error|null), response?: NamespaceIsExistsResponse) => void;
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

/** Represents a ProjectDeleteRequest. */
export class ProjectDeleteRequest implements IProjectDeleteRequest {

    /**
     * Constructs a new ProjectDeleteRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IProjectDeleteRequest);

    /** ProjectDeleteRequest project_id. */
    public project_id: number;

    /**
     * Encodes the specified ProjectDeleteRequest message. Does not implicitly {@link ProjectDeleteRequest.verify|verify} messages.
     * @param message ProjectDeleteRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: ProjectDeleteRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a ProjectDeleteRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns ProjectDeleteRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): ProjectDeleteRequest;
}

/** Represents a ProjectShowRequest. */
export class ProjectShowRequest implements IProjectShowRequest {

    /**
     * Constructs a new ProjectShowRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IProjectShowRequest);

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

    /** ProjectShowResponse git_project_id. */
    public git_project_id: number;

    /** ProjectShowResponse git_branch. */
    public git_branch: string;

    /** ProjectShowResponse git_commit. */
    public git_commit: string;

    /** ProjectShowResponse config. */
    public config: string;

    /** ProjectShowResponse docker_image. */
    public docker_image: string;

    /** ProjectShowResponse atomic. */
    public atomic: boolean;

    /** ProjectShowResponse git_commit_web_url. */
    public git_commit_web_url: string;

    /** ProjectShowResponse git_commit_title. */
    public git_commit_title: string;

    /** ProjectShowResponse git_commit_author. */
    public git_commit_author: string;

    /** ProjectShowResponse git_commit_date. */
    public git_commit_date: string;

    /** ProjectShowResponse urls. */
    public urls: ServiceEndpoint[];

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

    /** ProjectShowResponse humanize_created_at. */
    public humanize_created_at: string;

    /** ProjectShowResponse humanize_updated_at. */
    public humanize_updated_at: string;

    /** ProjectShowResponse extra_values. */
    public extra_values: ProjectExtraItem[];

    /** ProjectShowResponse elements. */
    public elements: Element[];

    /** ProjectShowResponse config_type. */
    public config_type: string;

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

/** Represents a ProjectAllContainersRequest. */
export class ProjectAllContainersRequest implements IProjectAllContainersRequest {

    /**
     * Constructs a new ProjectAllContainersRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IProjectAllContainersRequest);

    /** ProjectAllContainersRequest project_id. */
    public project_id: number;

    /**
     * Encodes the specified ProjectAllContainersRequest message. Does not implicitly {@link ProjectAllContainersRequest.verify|verify} messages.
     * @param message ProjectAllContainersRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: ProjectAllContainersRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a ProjectAllContainersRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns ProjectAllContainersRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): ProjectAllContainersRequest;
}

/** Represents a ProjectPod. */
export class ProjectPod implements IProjectPod {

    /**
     * Constructs a new ProjectPod.
     * @param [properties] Properties to set
     */
    constructor(properties?: IProjectPod);

    /** ProjectPod namespace. */
    public namespace: string;

    /** ProjectPod pod_name. */
    public pod_name: string;

    /** ProjectPod container_name. */
    public container_name: string;

    /**
     * Encodes the specified ProjectPod message. Does not implicitly {@link ProjectPod.verify|verify} messages.
     * @param message ProjectPod message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: ProjectPod, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a ProjectPod message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns ProjectPod
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): ProjectPod;
}

/** Represents a ProjectAllContainersResponse. */
export class ProjectAllContainersResponse implements IProjectAllContainersResponse {

    /**
     * Constructs a new ProjectAllContainersResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IProjectAllContainersResponse);

    /** ProjectAllContainersResponse items. */
    public items: ProjectPod[];

    /**
     * Encodes the specified ProjectAllContainersResponse message. Does not implicitly {@link ProjectAllContainersResponse.verify|verify} messages.
     * @param message ProjectAllContainersResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: ProjectAllContainersResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a ProjectAllContainersResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns ProjectAllContainersResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): ProjectAllContainersResponse;
}

/** Represents a ProjectApplyResponse. */
export class ProjectApplyResponse implements IProjectApplyResponse {

    /**
     * Constructs a new ProjectApplyResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IProjectApplyResponse);

    /** ProjectApplyResponse metadata. */
    public metadata?: (Metadata|null);

    /** ProjectApplyResponse project. */
    public project?: (ProjectModel|null);

    /**
     * Encodes the specified ProjectApplyResponse message. Does not implicitly {@link ProjectApplyResponse.verify|verify} messages.
     * @param message ProjectApplyResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: ProjectApplyResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a ProjectApplyResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns ProjectApplyResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): ProjectApplyResponse;
}

/** Represents a ProjectDryRunApplyResponse. */
export class ProjectDryRunApplyResponse implements IProjectDryRunApplyResponse {

    /**
     * Constructs a new ProjectDryRunApplyResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IProjectDryRunApplyResponse);

    /** ProjectDryRunApplyResponse results. */
    public results: string[];

    /**
     * Encodes the specified ProjectDryRunApplyResponse message. Does not implicitly {@link ProjectDryRunApplyResponse.verify|verify} messages.
     * @param message ProjectDryRunApplyResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: ProjectDryRunApplyResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a ProjectDryRunApplyResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns ProjectDryRunApplyResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): ProjectDryRunApplyResponse;
}

/** Represents a ProjectApplyRequest. */
export class ProjectApplyRequest implements IProjectApplyRequest {

    /**
     * Constructs a new ProjectApplyRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IProjectApplyRequest);

    /** ProjectApplyRequest namespace_id. */
    public namespace_id: number;

    /** ProjectApplyRequest name. */
    public name: string;

    /** ProjectApplyRequest git_project_id. */
    public git_project_id: number;

    /** ProjectApplyRequest git_branch. */
    public git_branch: string;

    /** ProjectApplyRequest git_commit. */
    public git_commit: string;

    /** ProjectApplyRequest config. */
    public config: string;

    /** ProjectApplyRequest atomic. */
    public atomic: boolean;

    /** ProjectApplyRequest websocket_sync. */
    public websocket_sync: boolean;

    /** ProjectApplyRequest extra_values. */
    public extra_values: ProjectExtraItem[];

    /** ProjectApplyRequest install_timeout_seconds. */
    public install_timeout_seconds: number;

    /**
     * Encodes the specified ProjectApplyRequest message. Does not implicitly {@link ProjectApplyRequest.verify|verify} messages.
     * @param message ProjectApplyRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: ProjectApplyRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a ProjectApplyRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns ProjectApplyRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): ProjectApplyRequest;
}

/** Represents a ProjectDeleteResponse. */
export class ProjectDeleteResponse implements IProjectDeleteResponse {

    /**
     * Constructs a new ProjectDeleteResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IProjectDeleteResponse);

    /**
     * Encodes the specified ProjectDeleteResponse message. Does not implicitly {@link ProjectDeleteResponse.verify|verify} messages.
     * @param message ProjectDeleteResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: ProjectDeleteResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a ProjectDeleteResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns ProjectDeleteResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): ProjectDeleteResponse;
}

/** Represents a ProjectListRequest. */
export class ProjectListRequest implements IProjectListRequest {

    /**
     * Constructs a new ProjectListRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IProjectListRequest);

    /** ProjectListRequest page. */
    public page: number;

    /** ProjectListRequest page_size. */
    public page_size: number;

    /**
     * Encodes the specified ProjectListRequest message. Does not implicitly {@link ProjectListRequest.verify|verify} messages.
     * @param message ProjectListRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: ProjectListRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a ProjectListRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns ProjectListRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): ProjectListRequest;
}

/** Represents a ProjectListResponse. */
export class ProjectListResponse implements IProjectListResponse {

    /**
     * Constructs a new ProjectListResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IProjectListResponse);

    /** ProjectListResponse page. */
    public page: number;

    /** ProjectListResponse page_size. */
    public page_size: number;

    /** ProjectListResponse count. */
    public count: number;

    /** ProjectListResponse items. */
    public items: ProjectModel[];

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
     * Calls List.
     * @param request ProjectListRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and ProjectListResponse
     */
    public list(request: ProjectListRequest, callback: Project.ListCallback): void;

    /**
     * Calls List.
     * @param request ProjectListRequest message or plain object
     * @returns Promise
     */
    public list(request: ProjectListRequest): Promise<ProjectListResponse>;

    /**
     * Calls Apply.
     * @param request ProjectApplyRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and ProjectApplyResponse
     */
    public apply(request: ProjectApplyRequest, callback: Project.ApplyCallback): void;

    /**
     * Calls Apply.
     * @param request ProjectApplyRequest message or plain object
     * @returns Promise
     */
    public apply(request: ProjectApplyRequest): Promise<ProjectApplyResponse>;

    /**
     * Calls ApplyDryRun.
     * @param request ProjectApplyRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and ProjectDryRunApplyResponse
     */
    public applyDryRun(request: ProjectApplyRequest, callback: Project.ApplyDryRunCallback): void;

    /**
     * Calls ApplyDryRun.
     * @param request ProjectApplyRequest message or plain object
     * @returns Promise
     */
    public applyDryRun(request: ProjectApplyRequest): Promise<ProjectDryRunApplyResponse>;

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
     * Calls Delete.
     * @param request ProjectDeleteRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and ProjectDeleteResponse
     */
    public delete(request: ProjectDeleteRequest, callback: Project.DeleteCallback): void;

    /**
     * Calls Delete.
     * @param request ProjectDeleteRequest message or plain object
     * @returns Promise
     */
    public delete(request: ProjectDeleteRequest): Promise<ProjectDeleteResponse>;

    /**
     * Calls AllContainers.
     * @param request ProjectAllContainersRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and ProjectAllContainersResponse
     */
    public allContainers(request: ProjectAllContainersRequest, callback: Project.AllContainersCallback): void;

    /**
     * Calls AllContainers.
     * @param request ProjectAllContainersRequest message or plain object
     * @returns Promise
     */
    public allContainers(request: ProjectAllContainersRequest): Promise<ProjectAllContainersResponse>;
}

export namespace Project {

    /**
     * Callback as used by {@link Project#list}.
     * @param error Error, if any
     * @param [response] ProjectListResponse
     */
    type ListCallback = (error: (Error|null), response?: ProjectListResponse) => void;

    /**
     * Callback as used by {@link Project#apply}.
     * @param error Error, if any
     * @param [response] ProjectApplyResponse
     */
    type ApplyCallback = (error: (Error|null), response?: ProjectApplyResponse) => void;

    /**
     * Callback as used by {@link Project#applyDryRun}.
     * @param error Error, if any
     * @param [response] ProjectDryRunApplyResponse
     */
    type ApplyDryRunCallback = (error: (Error|null), response?: ProjectDryRunApplyResponse) => void;

    /**
     * Callback as used by {@link Project#show}.
     * @param error Error, if any
     * @param [response] ProjectShowResponse
     */
    type ShowCallback = (error: (Error|null), response?: ProjectShowResponse) => void;

    /**
     * Callback as used by {@link Project#delete_}.
     * @param error Error, if any
     * @param [response] ProjectDeleteResponse
     */
    type DeleteCallback = (error: (Error|null), response?: ProjectDeleteResponse) => void;

    /**
     * Callback as used by {@link Project#allContainers}.
     * @param error Error, if any
     * @param [response] ProjectAllContainersResponse
     */
    type AllContainersCallback = (error: (Error|null), response?: ProjectAllContainersResponse) => void;
}

/** Represents a VersionRequest. */
export class VersionRequest implements IVersionRequest {

    /**
     * Constructs a new VersionRequest.
     * @param [properties] Properties to set
     */
    constructor(properties?: IVersionRequest);

    /**
     * Encodes the specified VersionRequest message. Does not implicitly {@link VersionRequest.verify|verify} messages.
     * @param message VersionRequest message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: VersionRequest, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a VersionRequest message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns VersionRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): VersionRequest;
}

/** Represents a VersionResponse. */
export class VersionResponse implements IVersionResponse {

    /**
     * Constructs a new VersionResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IVersionResponse);

    /** VersionResponse version. */
    public version: string;

    /** VersionResponse build_date. */
    public build_date: string;

    /** VersionResponse git_branch. */
    public git_branch: string;

    /** VersionResponse git_commit. */
    public git_commit: string;

    /** VersionResponse git_tag. */
    public git_tag: string;

    /** VersionResponse go_version. */
    public go_version: string;

    /** VersionResponse compiler. */
    public compiler: string;

    /** VersionResponse platform. */
    public platform: string;

    /** VersionResponse kubectl_version. */
    public kubectl_version: string;

    /** VersionResponse helm_version. */
    public helm_version: string;

    /** VersionResponse git_repo. */
    public git_repo: string;

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
     * Calls Version.
     * @param request VersionRequest message or plain object
     * @param callback Node-style callback called with the error, if any, and VersionResponse
     */
    public version(request: VersionRequest, callback: Version.VersionCallback): void;

    /**
     * Calls Version.
     * @param request VersionRequest message or plain object
     * @returns Promise
     */
    public version(request: VersionRequest): Promise<VersionResponse>;
}

export namespace Version {

    /**
     * Callback as used by {@link Version#version}.
     * @param error Error, if any
     * @param [response] VersionResponse
     */
    type VersionCallback = (error: (Error|null), response?: VersionResponse) => void;
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
    ApplyProject = 9,
    HandleExecShell = 50,
    HandleExecShellMsg = 51,
    HandleCloseShell = 52,
    HandleAuthorize = 53
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

/** Represents a ProjectExtraItem. */
export class ProjectExtraItem implements IProjectExtraItem {

    /**
     * Constructs a new ProjectExtraItem.
     * @param [properties] Properties to set
     */
    constructor(properties?: IProjectExtraItem);

    /** ProjectExtraItem path. */
    public path: string;

    /** ProjectExtraItem value. */
    public value: string;

    /**
     * Encodes the specified ProjectExtraItem message. Does not implicitly {@link ProjectExtraItem.verify|verify} messages.
     * @param message ProjectExtraItem message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: ProjectExtraItem, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a ProjectExtraItem message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns ProjectExtraItem
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): ProjectExtraItem;
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

    /** ProjectInput git_project_id. */
    public git_project_id: number;

    /** ProjectInput git_branch. */
    public git_branch: string;

    /** ProjectInput git_commit. */
    public git_commit: string;

    /** ProjectInput config. */
    public config: string;

    /** ProjectInput atomic. */
    public atomic: boolean;

    /** ProjectInput extra_values. */
    public extra_values: ProjectExtraItem[];

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

    /** UpdateProjectInput git_branch. */
    public git_branch: string;

    /** UpdateProjectInput git_commit. */
    public git_commit: string;

    /** UpdateProjectInput config. */
    public config: string;

    /** UpdateProjectInput atomic. */
    public atomic: boolean;

    /** UpdateProjectInput extra_values. */
    public extra_values: ProjectExtraItem[];

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

/** Represents a Metadata. */
export class Metadata implements IMetadata {

    /**
     * Constructs a new Metadata.
     * @param [properties] Properties to set
     */
    constructor(properties?: IMetadata);

    /** Metadata id. */
    public id: string;

    /** Metadata uid. */
    public uid: string;

    /** Metadata slug. */
    public slug: string;

    /** Metadata type. */
    public type: Type;

    /** Metadata end. */
    public end: boolean;

    /** Metadata result. */
    public result: ResultType;

    /** Metadata to. */
    public to: To;

    /** Metadata data. */
    public data: string;

    /**
     * Encodes the specified Metadata message. Does not implicitly {@link Metadata.verify|verify} messages.
     * @param message Metadata message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: Metadata, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a Metadata message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns Metadata
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): Metadata;
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

/** Represents a WsMetadataResponse. */
export class WsMetadataResponse implements IWsMetadataResponse {

    /**
     * Constructs a new WsMetadataResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IWsMetadataResponse);

    /** WsMetadataResponse metadata. */
    public metadata?: (Metadata|null);

    /**
     * Encodes the specified WsMetadataResponse message. Does not implicitly {@link WsMetadataResponse.verify|verify} messages.
     * @param message WsMetadataResponse message or plain object to encode
     * @param [writer] Writer to encode to
     * @returns Writer
     */
    public static encode(message: WsMetadataResponse, writer?: $protobuf.Writer): $protobuf.Writer;

    /**
     * Decodes a WsMetadataResponse message from the specified reader or buffer.
     * @param reader Reader or buffer to decode from
     * @param [length] Message length if known beforehand
     * @returns WsMetadataResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    public static decode(reader: ($protobuf.Reader|Uint8Array), length?: number): WsMetadataResponse;
}

/** Represents a WsHandleShellResponse. */
export class WsHandleShellResponse implements IWsHandleShellResponse {

    /**
     * Constructs a new WsHandleShellResponse.
     * @param [properties] Properties to set
     */
    constructor(properties?: IWsHandleShellResponse);

    /** WsHandleShellResponse metadata. */
    public metadata?: (Metadata|null);

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
    public metadata?: (Metadata|null);

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
