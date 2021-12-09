/*eslint-disable block-scoped-var, id-length, no-control-regex, no-magic-numbers, no-prototype-builtins, no-redeclare, no-shadow, no-var, sort-vars*/
import * as $protobuf from "protobufjs/minimal";

// Common aliases
const $Reader = $protobuf.Reader, $Writer = $protobuf.Writer, $util = $protobuf.util;

// Exported root namespace
const $root = $protobuf.roots["default"] || ($protobuf.roots["default"] = {});

export const LoginRequest = $root.LoginRequest = (() => {

    /**
     * Properties of a LoginRequest.
     * @exports ILoginRequest
     * @interface ILoginRequest
     * @property {string|null} [username] LoginRequest username
     * @property {string|null} [password] LoginRequest password
     */

    /**
     * Constructs a new LoginRequest.
     * @exports LoginRequest
     * @classdesc Represents a LoginRequest.
     * @implements ILoginRequest
     * @constructor
     * @param {ILoginRequest=} [properties] Properties to set
     */
    function LoginRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * LoginRequest username.
     * @member {string} username
     * @memberof LoginRequest
     * @instance
     */
    LoginRequest.prototype.username = "";

    /**
     * LoginRequest password.
     * @member {string} password
     * @memberof LoginRequest
     * @instance
     */
    LoginRequest.prototype.password = "";

    /**
     * Encodes the specified LoginRequest message. Does not implicitly {@link LoginRequest.verify|verify} messages.
     * @function encode
     * @memberof LoginRequest
     * @static
     * @param {LoginRequest} message LoginRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    LoginRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.username != null && Object.hasOwnProperty.call(message, "username"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.username);
        if (message.password != null && Object.hasOwnProperty.call(message, "password"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.password);
        return writer;
    };

    /**
     * Decodes a LoginRequest message from the specified reader or buffer.
     * @function decode
     * @memberof LoginRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {LoginRequest} LoginRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    LoginRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.LoginRequest();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.username = reader.string();
                break;
            case 2:
                message.password = reader.string();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return LoginRequest;
})();

export const LoginResponse = $root.LoginResponse = (() => {

    /**
     * Properties of a LoginResponse.
     * @exports ILoginResponse
     * @interface ILoginResponse
     * @property {string|null} [token] LoginResponse token
     * @property {number|null} [expires_in] LoginResponse expires_in
     */

    /**
     * Constructs a new LoginResponse.
     * @exports LoginResponse
     * @classdesc Represents a LoginResponse.
     * @implements ILoginResponse
     * @constructor
     * @param {ILoginResponse=} [properties] Properties to set
     */
    function LoginResponse(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * LoginResponse token.
     * @member {string} token
     * @memberof LoginResponse
     * @instance
     */
    LoginResponse.prototype.token = "";

    /**
     * LoginResponse expires_in.
     * @member {number} expires_in
     * @memberof LoginResponse
     * @instance
     */
    LoginResponse.prototype.expires_in = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * Encodes the specified LoginResponse message. Does not implicitly {@link LoginResponse.verify|verify} messages.
     * @function encode
     * @memberof LoginResponse
     * @static
     * @param {LoginResponse} message LoginResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    LoginResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.token != null && Object.hasOwnProperty.call(message, "token"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.token);
        if (message.expires_in != null && Object.hasOwnProperty.call(message, "expires_in"))
            writer.uint32(/* id 2, wireType 0 =*/16).int64(message.expires_in);
        return writer;
    };

    /**
     * Decodes a LoginResponse message from the specified reader or buffer.
     * @function decode
     * @memberof LoginResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {LoginResponse} LoginResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    LoginResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.LoginResponse();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.token = reader.string();
                break;
            case 2:
                message.expires_in = reader.int64();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return LoginResponse;
})();

export const InfoResponse = $root.InfoResponse = (() => {

    /**
     * Properties of an InfoResponse.
     * @exports IInfoResponse
     * @interface IInfoResponse
     * @property {string|null} [id] InfoResponse id
     * @property {string|null} [avatar] InfoResponse avatar
     * @property {string|null} [name] InfoResponse name
     * @property {string|null} [email] InfoResponse email
     * @property {string|null} [logout_url] InfoResponse logout_url
     * @property {Array.<string>|null} [roles] InfoResponse roles
     */

    /**
     * Constructs a new InfoResponse.
     * @exports InfoResponse
     * @classdesc Represents an InfoResponse.
     * @implements IInfoResponse
     * @constructor
     * @param {IInfoResponse=} [properties] Properties to set
     */
    function InfoResponse(properties) {
        this.roles = [];
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * InfoResponse id.
     * @member {string} id
     * @memberof InfoResponse
     * @instance
     */
    InfoResponse.prototype.id = "";

    /**
     * InfoResponse avatar.
     * @member {string} avatar
     * @memberof InfoResponse
     * @instance
     */
    InfoResponse.prototype.avatar = "";

    /**
     * InfoResponse name.
     * @member {string} name
     * @memberof InfoResponse
     * @instance
     */
    InfoResponse.prototype.name = "";

    /**
     * InfoResponse email.
     * @member {string} email
     * @memberof InfoResponse
     * @instance
     */
    InfoResponse.prototype.email = "";

    /**
     * InfoResponse logout_url.
     * @member {string} logout_url
     * @memberof InfoResponse
     * @instance
     */
    InfoResponse.prototype.logout_url = "";

    /**
     * InfoResponse roles.
     * @member {Array.<string>} roles
     * @memberof InfoResponse
     * @instance
     */
    InfoResponse.prototype.roles = $util.emptyArray;

    /**
     * Encodes the specified InfoResponse message. Does not implicitly {@link InfoResponse.verify|verify} messages.
     * @function encode
     * @memberof InfoResponse
     * @static
     * @param {InfoResponse} message InfoResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    InfoResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.id != null && Object.hasOwnProperty.call(message, "id"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.id);
        if (message.avatar != null && Object.hasOwnProperty.call(message, "avatar"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.avatar);
        if (message.name != null && Object.hasOwnProperty.call(message, "name"))
            writer.uint32(/* id 3, wireType 2 =*/26).string(message.name);
        if (message.email != null && Object.hasOwnProperty.call(message, "email"))
            writer.uint32(/* id 4, wireType 2 =*/34).string(message.email);
        if (message.logout_url != null && Object.hasOwnProperty.call(message, "logout_url"))
            writer.uint32(/* id 5, wireType 2 =*/42).string(message.logout_url);
        if (message.roles != null && message.roles.length)
            for (let i = 0; i < message.roles.length; ++i)
                writer.uint32(/* id 6, wireType 2 =*/50).string(message.roles[i]);
        return writer;
    };

    /**
     * Decodes an InfoResponse message from the specified reader or buffer.
     * @function decode
     * @memberof InfoResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {InfoResponse} InfoResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    InfoResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.InfoResponse();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.id = reader.string();
                break;
            case 2:
                message.avatar = reader.string();
                break;
            case 3:
                message.name = reader.string();
                break;
            case 4:
                message.email = reader.string();
                break;
            case 5:
                message.logout_url = reader.string();
                break;
            case 6:
                if (!(message.roles && message.roles.length))
                    message.roles = [];
                message.roles.push(reader.string());
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return InfoResponse;
})();

export const OidcSetting = $root.OidcSetting = (() => {

    /**
     * Properties of an OidcSetting.
     * @exports IOidcSetting
     * @interface IOidcSetting
     * @property {boolean|null} [enabled] OidcSetting enabled
     * @property {string|null} [name] OidcSetting name
     * @property {string|null} [url] OidcSetting url
     * @property {string|null} [end_session_endpoint] OidcSetting end_session_endpoint
     * @property {string|null} [state] OidcSetting state
     */

    /**
     * Constructs a new OidcSetting.
     * @exports OidcSetting
     * @classdesc Represents an OidcSetting.
     * @implements IOidcSetting
     * @constructor
     * @param {IOidcSetting=} [properties] Properties to set
     */
    function OidcSetting(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * OidcSetting enabled.
     * @member {boolean} enabled
     * @memberof OidcSetting
     * @instance
     */
    OidcSetting.prototype.enabled = false;

    /**
     * OidcSetting name.
     * @member {string} name
     * @memberof OidcSetting
     * @instance
     */
    OidcSetting.prototype.name = "";

    /**
     * OidcSetting url.
     * @member {string} url
     * @memberof OidcSetting
     * @instance
     */
    OidcSetting.prototype.url = "";

    /**
     * OidcSetting end_session_endpoint.
     * @member {string} end_session_endpoint
     * @memberof OidcSetting
     * @instance
     */
    OidcSetting.prototype.end_session_endpoint = "";

    /**
     * OidcSetting state.
     * @member {string} state
     * @memberof OidcSetting
     * @instance
     */
    OidcSetting.prototype.state = "";

    /**
     * Encodes the specified OidcSetting message. Does not implicitly {@link OidcSetting.verify|verify} messages.
     * @function encode
     * @memberof OidcSetting
     * @static
     * @param {OidcSetting} message OidcSetting message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    OidcSetting.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.enabled != null && Object.hasOwnProperty.call(message, "enabled"))
            writer.uint32(/* id 1, wireType 0 =*/8).bool(message.enabled);
        if (message.name != null && Object.hasOwnProperty.call(message, "name"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.name);
        if (message.url != null && Object.hasOwnProperty.call(message, "url"))
            writer.uint32(/* id 3, wireType 2 =*/26).string(message.url);
        if (message.end_session_endpoint != null && Object.hasOwnProperty.call(message, "end_session_endpoint"))
            writer.uint32(/* id 4, wireType 2 =*/34).string(message.end_session_endpoint);
        if (message.state != null && Object.hasOwnProperty.call(message, "state"))
            writer.uint32(/* id 5, wireType 2 =*/42).string(message.state);
        return writer;
    };

    /**
     * Decodes an OidcSetting message from the specified reader or buffer.
     * @function decode
     * @memberof OidcSetting
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {OidcSetting} OidcSetting
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    OidcSetting.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.OidcSetting();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.enabled = reader.bool();
                break;
            case 2:
                message.name = reader.string();
                break;
            case 3:
                message.url = reader.string();
                break;
            case 4:
                message.end_session_endpoint = reader.string();
                break;
            case 5:
                message.state = reader.string();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return OidcSetting;
})();

export const SettingsResponse = $root.SettingsResponse = (() => {

    /**
     * Properties of a SettingsResponse.
     * @exports ISettingsResponse
     * @interface ISettingsResponse
     * @property {Array.<OidcSetting>|null} [items] SettingsResponse items
     */

    /**
     * Constructs a new SettingsResponse.
     * @exports SettingsResponse
     * @classdesc Represents a SettingsResponse.
     * @implements ISettingsResponse
     * @constructor
     * @param {ISettingsResponse=} [properties] Properties to set
     */
    function SettingsResponse(properties) {
        this.items = [];
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * SettingsResponse items.
     * @member {Array.<OidcSetting>} items
     * @memberof SettingsResponse
     * @instance
     */
    SettingsResponse.prototype.items = $util.emptyArray;

    /**
     * Encodes the specified SettingsResponse message. Does not implicitly {@link SettingsResponse.verify|verify} messages.
     * @function encode
     * @memberof SettingsResponse
     * @static
     * @param {SettingsResponse} message SettingsResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    SettingsResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.items != null && message.items.length)
            for (let i = 0; i < message.items.length; ++i)
                $root.OidcSetting.encode(message.items[i], writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
        return writer;
    };

    /**
     * Decodes a SettingsResponse message from the specified reader or buffer.
     * @function decode
     * @memberof SettingsResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {SettingsResponse} SettingsResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    SettingsResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.SettingsResponse();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                if (!(message.items && message.items.length))
                    message.items = [];
                message.items.push($root.OidcSetting.decode(reader, reader.uint32()));
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return SettingsResponse;
})();

export const ExchangeRequest = $root.ExchangeRequest = (() => {

    /**
     * Properties of an ExchangeRequest.
     * @exports IExchangeRequest
     * @interface IExchangeRequest
     * @property {string|null} [code] ExchangeRequest code
     */

    /**
     * Constructs a new ExchangeRequest.
     * @exports ExchangeRequest
     * @classdesc Represents an ExchangeRequest.
     * @implements IExchangeRequest
     * @constructor
     * @param {IExchangeRequest=} [properties] Properties to set
     */
    function ExchangeRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * ExchangeRequest code.
     * @member {string} code
     * @memberof ExchangeRequest
     * @instance
     */
    ExchangeRequest.prototype.code = "";

    /**
     * Encodes the specified ExchangeRequest message. Does not implicitly {@link ExchangeRequest.verify|verify} messages.
     * @function encode
     * @memberof ExchangeRequest
     * @static
     * @param {ExchangeRequest} message ExchangeRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    ExchangeRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.code != null && Object.hasOwnProperty.call(message, "code"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.code);
        return writer;
    };

    /**
     * Decodes an ExchangeRequest message from the specified reader or buffer.
     * @function decode
     * @memberof ExchangeRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {ExchangeRequest} ExchangeRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    ExchangeRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.ExchangeRequest();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.code = reader.string();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return ExchangeRequest;
})();

export const Auth = $root.Auth = (() => {

    /**
     * Constructs a new Auth service.
     * @exports Auth
     * @classdesc Represents an Auth
     * @extends $protobuf.rpc.Service
     * @constructor
     * @param {$protobuf.RPCImpl} rpcImpl RPC implementation
     * @param {boolean} [requestDelimited=false] Whether requests are length-delimited
     * @param {boolean} [responseDelimited=false] Whether responses are length-delimited
     */
    function Auth(rpcImpl, requestDelimited, responseDelimited) {
        $protobuf.rpc.Service.call(this, rpcImpl, requestDelimited, responseDelimited);
    }

    (Auth.prototype = Object.create($protobuf.rpc.Service.prototype)).constructor = Auth;

    /**
     * Callback as used by {@link Auth#login}.
     * @memberof Auth
     * @typedef LoginCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {LoginResponse} [response] LoginResponse
     */

    /**
     * Calls Login.
     * @function login
     * @memberof Auth
     * @instance
     * @param {LoginRequest} request LoginRequest message or plain object
     * @param {Auth.LoginCallback} callback Node-style callback called with the error, if any, and LoginResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Auth.prototype.login = function login(request, callback) {
        return this.rpcCall(login, $root.LoginRequest, $root.LoginResponse, request, callback);
    }, "name", { value: "Login" });

    /**
     * Calls Login.
     * @function login
     * @memberof Auth
     * @instance
     * @param {LoginRequest} request LoginRequest message or plain object
     * @returns {Promise<LoginResponse>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link Auth#info}.
     * @memberof Auth
     * @typedef InfoCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {InfoResponse} [response] InfoResponse
     */

    /**
     * Calls Info.
     * @function info
     * @memberof Auth
     * @instance
     * @param {google.protobuf.Empty} request Empty message or plain object
     * @param {Auth.InfoCallback} callback Node-style callback called with the error, if any, and InfoResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Auth.prototype.info = function info(request, callback) {
        return this.rpcCall(info, $root.google.protobuf.Empty, $root.InfoResponse, request, callback);
    }, "name", { value: "Info" });

    /**
     * Calls Info.
     * @function info
     * @memberof Auth
     * @instance
     * @param {google.protobuf.Empty} request Empty message or plain object
     * @returns {Promise<InfoResponse>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link Auth#settings}.
     * @memberof Auth
     * @typedef SettingsCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {SettingsResponse} [response] SettingsResponse
     */

    /**
     * Calls Settings.
     * @function settings
     * @memberof Auth
     * @instance
     * @param {google.protobuf.Empty} request Empty message or plain object
     * @param {Auth.SettingsCallback} callback Node-style callback called with the error, if any, and SettingsResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Auth.prototype.settings = function settings(request, callback) {
        return this.rpcCall(settings, $root.google.protobuf.Empty, $root.SettingsResponse, request, callback);
    }, "name", { value: "Settings" });

    /**
     * Calls Settings.
     * @function settings
     * @memberof Auth
     * @instance
     * @param {google.protobuf.Empty} request Empty message or plain object
     * @returns {Promise<SettingsResponse>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link Auth#exchange}.
     * @memberof Auth
     * @typedef ExchangeCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {LoginResponse} [response] LoginResponse
     */

    /**
     * Calls Exchange.
     * @function exchange
     * @memberof Auth
     * @instance
     * @param {ExchangeRequest} request ExchangeRequest message or plain object
     * @param {Auth.ExchangeCallback} callback Node-style callback called with the error, if any, and LoginResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Auth.prototype.exchange = function exchange(request, callback) {
        return this.rpcCall(exchange, $root.ExchangeRequest, $root.LoginResponse, request, callback);
    }, "name", { value: "Exchange" });

    /**
     * Calls Exchange.
     * @function exchange
     * @memberof Auth
     * @instance
     * @param {ExchangeRequest} request ExchangeRequest message or plain object
     * @returns {Promise<LoginResponse>} Promise
     * @variation 2
     */

    return Auth;
})();

/**
 * ClusterStatus enum.
 * @exports ClusterStatus
 * @enum {number}
 * @property {number} StatusUnknown=0 StatusUnknown value
 * @property {number} StatusBad=1 StatusBad value
 * @property {number} StatusNotGood=2 StatusNotGood value
 * @property {number} StatusHealth=3 StatusHealth value
 */
export const ClusterStatus = $root.ClusterStatus = (() => {
    const valuesById = {}, values = Object.create(valuesById);
    values[valuesById[0] = "StatusUnknown"] = 0;
    values[valuesById[1] = "StatusBad"] = 1;
    values[valuesById[2] = "StatusNotGood"] = 2;
    values[valuesById[3] = "StatusHealth"] = 3;
    return values;
})();

export const ClusterInfoResponse = $root.ClusterInfoResponse = (() => {

    /**
     * Properties of a ClusterInfoResponse.
     * @exports IClusterInfoResponse
     * @interface IClusterInfoResponse
     * @property {string|null} [status] ClusterInfoResponse status
     * @property {string|null} [free_memory] ClusterInfoResponse free_memory
     * @property {string|null} [free_cpu] ClusterInfoResponse free_cpu
     * @property {string|null} [free_request_memory] ClusterInfoResponse free_request_memory
     * @property {string|null} [free_request_cpu] ClusterInfoResponse free_request_cpu
     * @property {string|null} [total_memory] ClusterInfoResponse total_memory
     * @property {string|null} [total_cpu] ClusterInfoResponse total_cpu
     * @property {string|null} [usage_memory_rate] ClusterInfoResponse usage_memory_rate
     * @property {string|null} [usage_cpu_rate] ClusterInfoResponse usage_cpu_rate
     * @property {string|null} [request_memory_rate] ClusterInfoResponse request_memory_rate
     * @property {string|null} [request_cpu_rate] ClusterInfoResponse request_cpu_rate
     */

    /**
     * Constructs a new ClusterInfoResponse.
     * @exports ClusterInfoResponse
     * @classdesc Represents a ClusterInfoResponse.
     * @implements IClusterInfoResponse
     * @constructor
     * @param {IClusterInfoResponse=} [properties] Properties to set
     */
    function ClusterInfoResponse(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * ClusterInfoResponse status.
     * @member {string} status
     * @memberof ClusterInfoResponse
     * @instance
     */
    ClusterInfoResponse.prototype.status = "";

    /**
     * ClusterInfoResponse free_memory.
     * @member {string} free_memory
     * @memberof ClusterInfoResponse
     * @instance
     */
    ClusterInfoResponse.prototype.free_memory = "";

    /**
     * ClusterInfoResponse free_cpu.
     * @member {string} free_cpu
     * @memberof ClusterInfoResponse
     * @instance
     */
    ClusterInfoResponse.prototype.free_cpu = "";

    /**
     * ClusterInfoResponse free_request_memory.
     * @member {string} free_request_memory
     * @memberof ClusterInfoResponse
     * @instance
     */
    ClusterInfoResponse.prototype.free_request_memory = "";

    /**
     * ClusterInfoResponse free_request_cpu.
     * @member {string} free_request_cpu
     * @memberof ClusterInfoResponse
     * @instance
     */
    ClusterInfoResponse.prototype.free_request_cpu = "";

    /**
     * ClusterInfoResponse total_memory.
     * @member {string} total_memory
     * @memberof ClusterInfoResponse
     * @instance
     */
    ClusterInfoResponse.prototype.total_memory = "";

    /**
     * ClusterInfoResponse total_cpu.
     * @member {string} total_cpu
     * @memberof ClusterInfoResponse
     * @instance
     */
    ClusterInfoResponse.prototype.total_cpu = "";

    /**
     * ClusterInfoResponse usage_memory_rate.
     * @member {string} usage_memory_rate
     * @memberof ClusterInfoResponse
     * @instance
     */
    ClusterInfoResponse.prototype.usage_memory_rate = "";

    /**
     * ClusterInfoResponse usage_cpu_rate.
     * @member {string} usage_cpu_rate
     * @memberof ClusterInfoResponse
     * @instance
     */
    ClusterInfoResponse.prototype.usage_cpu_rate = "";

    /**
     * ClusterInfoResponse request_memory_rate.
     * @member {string} request_memory_rate
     * @memberof ClusterInfoResponse
     * @instance
     */
    ClusterInfoResponse.prototype.request_memory_rate = "";

    /**
     * ClusterInfoResponse request_cpu_rate.
     * @member {string} request_cpu_rate
     * @memberof ClusterInfoResponse
     * @instance
     */
    ClusterInfoResponse.prototype.request_cpu_rate = "";

    /**
     * Encodes the specified ClusterInfoResponse message. Does not implicitly {@link ClusterInfoResponse.verify|verify} messages.
     * @function encode
     * @memberof ClusterInfoResponse
     * @static
     * @param {ClusterInfoResponse} message ClusterInfoResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    ClusterInfoResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.status != null && Object.hasOwnProperty.call(message, "status"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.status);
        if (message.free_memory != null && Object.hasOwnProperty.call(message, "free_memory"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.free_memory);
        if (message.free_cpu != null && Object.hasOwnProperty.call(message, "free_cpu"))
            writer.uint32(/* id 3, wireType 2 =*/26).string(message.free_cpu);
        if (message.free_request_memory != null && Object.hasOwnProperty.call(message, "free_request_memory"))
            writer.uint32(/* id 4, wireType 2 =*/34).string(message.free_request_memory);
        if (message.free_request_cpu != null && Object.hasOwnProperty.call(message, "free_request_cpu"))
            writer.uint32(/* id 5, wireType 2 =*/42).string(message.free_request_cpu);
        if (message.total_memory != null && Object.hasOwnProperty.call(message, "total_memory"))
            writer.uint32(/* id 6, wireType 2 =*/50).string(message.total_memory);
        if (message.total_cpu != null && Object.hasOwnProperty.call(message, "total_cpu"))
            writer.uint32(/* id 7, wireType 2 =*/58).string(message.total_cpu);
        if (message.usage_memory_rate != null && Object.hasOwnProperty.call(message, "usage_memory_rate"))
            writer.uint32(/* id 8, wireType 2 =*/66).string(message.usage_memory_rate);
        if (message.usage_cpu_rate != null && Object.hasOwnProperty.call(message, "usage_cpu_rate"))
            writer.uint32(/* id 9, wireType 2 =*/74).string(message.usage_cpu_rate);
        if (message.request_memory_rate != null && Object.hasOwnProperty.call(message, "request_memory_rate"))
            writer.uint32(/* id 10, wireType 2 =*/82).string(message.request_memory_rate);
        if (message.request_cpu_rate != null && Object.hasOwnProperty.call(message, "request_cpu_rate"))
            writer.uint32(/* id 11, wireType 2 =*/90).string(message.request_cpu_rate);
        return writer;
    };

    /**
     * Decodes a ClusterInfoResponse message from the specified reader or buffer.
     * @function decode
     * @memberof ClusterInfoResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {ClusterInfoResponse} ClusterInfoResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    ClusterInfoResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.ClusterInfoResponse();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.status = reader.string();
                break;
            case 2:
                message.free_memory = reader.string();
                break;
            case 3:
                message.free_cpu = reader.string();
                break;
            case 4:
                message.free_request_memory = reader.string();
                break;
            case 5:
                message.free_request_cpu = reader.string();
                break;
            case 6:
                message.total_memory = reader.string();
                break;
            case 7:
                message.total_cpu = reader.string();
                break;
            case 8:
                message.usage_memory_rate = reader.string();
                break;
            case 9:
                message.usage_cpu_rate = reader.string();
                break;
            case 10:
                message.request_memory_rate = reader.string();
                break;
            case 11:
                message.request_cpu_rate = reader.string();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return ClusterInfoResponse;
})();

export const Cluster = $root.Cluster = (() => {

    /**
     * Constructs a new Cluster service.
     * @exports Cluster
     * @classdesc Represents a Cluster
     * @extends $protobuf.rpc.Service
     * @constructor
     * @param {$protobuf.RPCImpl} rpcImpl RPC implementation
     * @param {boolean} [requestDelimited=false] Whether requests are length-delimited
     * @param {boolean} [responseDelimited=false] Whether responses are length-delimited
     */
    function Cluster(rpcImpl, requestDelimited, responseDelimited) {
        $protobuf.rpc.Service.call(this, rpcImpl, requestDelimited, responseDelimited);
    }

    (Cluster.prototype = Object.create($protobuf.rpc.Service.prototype)).constructor = Cluster;

    /**
     * Callback as used by {@link Cluster#info}.
     * @memberof Cluster
     * @typedef InfoCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {ClusterInfoResponse} [response] ClusterInfoResponse
     */

    /**
     * Calls Info.
     * @function info
     * @memberof Cluster
     * @instance
     * @param {google.protobuf.Empty} request Empty message or plain object
     * @param {Cluster.InfoCallback} callback Node-style callback called with the error, if any, and ClusterInfoResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Cluster.prototype.info = function info(request, callback) {
        return this.rpcCall(info, $root.google.protobuf.Empty, $root.ClusterInfoResponse, request, callback);
    }, "name", { value: "Info" });

    /**
     * Calls Info.
     * @function info
     * @memberof Cluster
     * @instance
     * @param {google.protobuf.Empty} request Empty message or plain object
     * @returns {Promise<ClusterInfoResponse>} Promise
     * @variation 2
     */

    return Cluster;
})();

export const CopyToPodRequest = $root.CopyToPodRequest = (() => {

    /**
     * Properties of a CopyToPodRequest.
     * @exports ICopyToPodRequest
     * @interface ICopyToPodRequest
     * @property {number|null} [file_id] CopyToPodRequest file_id
     * @property {string|null} [namespace] CopyToPodRequest namespace
     * @property {string|null} [pod] CopyToPodRequest pod
     * @property {string|null} [container] CopyToPodRequest container
     */

    /**
     * Constructs a new CopyToPodRequest.
     * @exports CopyToPodRequest
     * @classdesc Represents a CopyToPodRequest.
     * @implements ICopyToPodRequest
     * @constructor
     * @param {ICopyToPodRequest=} [properties] Properties to set
     */
    function CopyToPodRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * CopyToPodRequest file_id.
     * @member {number} file_id
     * @memberof CopyToPodRequest
     * @instance
     */
    CopyToPodRequest.prototype.file_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * CopyToPodRequest namespace.
     * @member {string} namespace
     * @memberof CopyToPodRequest
     * @instance
     */
    CopyToPodRequest.prototype.namespace = "";

    /**
     * CopyToPodRequest pod.
     * @member {string} pod
     * @memberof CopyToPodRequest
     * @instance
     */
    CopyToPodRequest.prototype.pod = "";

    /**
     * CopyToPodRequest container.
     * @member {string} container
     * @memberof CopyToPodRequest
     * @instance
     */
    CopyToPodRequest.prototype.container = "";

    /**
     * Encodes the specified CopyToPodRequest message. Does not implicitly {@link CopyToPodRequest.verify|verify} messages.
     * @function encode
     * @memberof CopyToPodRequest
     * @static
     * @param {CopyToPodRequest} message CopyToPodRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    CopyToPodRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.file_id != null && Object.hasOwnProperty.call(message, "file_id"))
            writer.uint32(/* id 1, wireType 0 =*/8).int64(message.file_id);
        if (message.namespace != null && Object.hasOwnProperty.call(message, "namespace"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.namespace);
        if (message.pod != null && Object.hasOwnProperty.call(message, "pod"))
            writer.uint32(/* id 3, wireType 2 =*/26).string(message.pod);
        if (message.container != null && Object.hasOwnProperty.call(message, "container"))
            writer.uint32(/* id 4, wireType 2 =*/34).string(message.container);
        return writer;
    };

    /**
     * Decodes a CopyToPodRequest message from the specified reader or buffer.
     * @function decode
     * @memberof CopyToPodRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {CopyToPodRequest} CopyToPodRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    CopyToPodRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.CopyToPodRequest();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.file_id = reader.int64();
                break;
            case 2:
                message.namespace = reader.string();
                break;
            case 3:
                message.pod = reader.string();
                break;
            case 4:
                message.container = reader.string();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return CopyToPodRequest;
})();

export const CopyToPodResponse = $root.CopyToPodResponse = (() => {

    /**
     * Properties of a CopyToPodResponse.
     * @exports ICopyToPodResponse
     * @interface ICopyToPodResponse
     * @property {string|null} [podFilePath] CopyToPodResponse podFilePath
     * @property {string|null} [output] CopyToPodResponse output
     */

    /**
     * Constructs a new CopyToPodResponse.
     * @exports CopyToPodResponse
     * @classdesc Represents a CopyToPodResponse.
     * @implements ICopyToPodResponse
     * @constructor
     * @param {ICopyToPodResponse=} [properties] Properties to set
     */
    function CopyToPodResponse(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * CopyToPodResponse podFilePath.
     * @member {string} podFilePath
     * @memberof CopyToPodResponse
     * @instance
     */
    CopyToPodResponse.prototype.podFilePath = "";

    /**
     * CopyToPodResponse output.
     * @member {string} output
     * @memberof CopyToPodResponse
     * @instance
     */
    CopyToPodResponse.prototype.output = "";

    /**
     * Encodes the specified CopyToPodResponse message. Does not implicitly {@link CopyToPodResponse.verify|verify} messages.
     * @function encode
     * @memberof CopyToPodResponse
     * @static
     * @param {CopyToPodResponse} message CopyToPodResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    CopyToPodResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.podFilePath != null && Object.hasOwnProperty.call(message, "podFilePath"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.podFilePath);
        if (message.output != null && Object.hasOwnProperty.call(message, "output"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.output);
        return writer;
    };

    /**
     * Decodes a CopyToPodResponse message from the specified reader or buffer.
     * @function decode
     * @memberof CopyToPodResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {CopyToPodResponse} CopyToPodResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    CopyToPodResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.CopyToPodResponse();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.podFilePath = reader.string();
                break;
            case 2:
                message.output = reader.string();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return CopyToPodResponse;
})();

export const Cp = $root.Cp = (() => {

    /**
     * Constructs a new Cp service.
     * @exports Cp
     * @classdesc Represents a Cp
     * @extends $protobuf.rpc.Service
     * @constructor
     * @param {$protobuf.RPCImpl} rpcImpl RPC implementation
     * @param {boolean} [requestDelimited=false] Whether requests are length-delimited
     * @param {boolean} [responseDelimited=false] Whether responses are length-delimited
     */
    function Cp(rpcImpl, requestDelimited, responseDelimited) {
        $protobuf.rpc.Service.call(this, rpcImpl, requestDelimited, responseDelimited);
    }

    (Cp.prototype = Object.create($protobuf.rpc.Service.prototype)).constructor = Cp;

    /**
     * Callback as used by {@link Cp#copyToPod}.
     * @memberof Cp
     * @typedef CopyToPodCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {CopyToPodResponse} [response] CopyToPodResponse
     */

    /**
     * Calls CopyToPod.
     * @function copyToPod
     * @memberof Cp
     * @instance
     * @param {CopyToPodRequest} request CopyToPodRequest message or plain object
     * @param {Cp.CopyToPodCallback} callback Node-style callback called with the error, if any, and CopyToPodResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Cp.prototype.copyToPod = function copyToPod(request, callback) {
        return this.rpcCall(copyToPod, $root.CopyToPodRequest, $root.CopyToPodResponse, request, callback);
    }, "name", { value: "CopyToPod" });

    /**
     * Calls CopyToPod.
     * @function copyToPod
     * @memberof Cp
     * @instance
     * @param {CopyToPodRequest} request CopyToPodRequest message or plain object
     * @returns {Promise<CopyToPodResponse>} Promise
     * @variation 2
     */

    return Cp;
})();

export const GitlabDestroyRequest = $root.GitlabDestroyRequest = (() => {

    /**
     * Properties of a GitlabDestroyRequest.
     * @exports IGitlabDestroyRequest
     * @interface IGitlabDestroyRequest
     * @property {string|null} [namespace_id] GitlabDestroyRequest namespace_id
     * @property {string|null} [project_id] GitlabDestroyRequest project_id
     */

    /**
     * Constructs a new GitlabDestroyRequest.
     * @exports GitlabDestroyRequest
     * @classdesc Represents a GitlabDestroyRequest.
     * @implements IGitlabDestroyRequest
     * @constructor
     * @param {IGitlabDestroyRequest=} [properties] Properties to set
     */
    function GitlabDestroyRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * GitlabDestroyRequest namespace_id.
     * @member {string} namespace_id
     * @memberof GitlabDestroyRequest
     * @instance
     */
    GitlabDestroyRequest.prototype.namespace_id = "";

    /**
     * GitlabDestroyRequest project_id.
     * @member {string} project_id
     * @memberof GitlabDestroyRequest
     * @instance
     */
    GitlabDestroyRequest.prototype.project_id = "";

    /**
     * Encodes the specified GitlabDestroyRequest message. Does not implicitly {@link GitlabDestroyRequest.verify|verify} messages.
     * @function encode
     * @memberof GitlabDestroyRequest
     * @static
     * @param {GitlabDestroyRequest} message GitlabDestroyRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    GitlabDestroyRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.namespace_id != null && Object.hasOwnProperty.call(message, "namespace_id"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.namespace_id);
        if (message.project_id != null && Object.hasOwnProperty.call(message, "project_id"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.project_id);
        return writer;
    };

    /**
     * Decodes a GitlabDestroyRequest message from the specified reader or buffer.
     * @function decode
     * @memberof GitlabDestroyRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {GitlabDestroyRequest} GitlabDestroyRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    GitlabDestroyRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.GitlabDestroyRequest();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.namespace_id = reader.string();
                break;
            case 2:
                message.project_id = reader.string();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return GitlabDestroyRequest;
})();

export const EnableProjectRequest = $root.EnableProjectRequest = (() => {

    /**
     * Properties of an EnableProjectRequest.
     * @exports IEnableProjectRequest
     * @interface IEnableProjectRequest
     * @property {string|null} [gitlab_project_id] EnableProjectRequest gitlab_project_id
     */

    /**
     * Constructs a new EnableProjectRequest.
     * @exports EnableProjectRequest
     * @classdesc Represents an EnableProjectRequest.
     * @implements IEnableProjectRequest
     * @constructor
     * @param {IEnableProjectRequest=} [properties] Properties to set
     */
    function EnableProjectRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * EnableProjectRequest gitlab_project_id.
     * @member {string} gitlab_project_id
     * @memberof EnableProjectRequest
     * @instance
     */
    EnableProjectRequest.prototype.gitlab_project_id = "";

    /**
     * Encodes the specified EnableProjectRequest message. Does not implicitly {@link EnableProjectRequest.verify|verify} messages.
     * @function encode
     * @memberof EnableProjectRequest
     * @static
     * @param {EnableProjectRequest} message EnableProjectRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    EnableProjectRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.gitlab_project_id != null && Object.hasOwnProperty.call(message, "gitlab_project_id"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.gitlab_project_id);
        return writer;
    };

    /**
     * Decodes an EnableProjectRequest message from the specified reader or buffer.
     * @function decode
     * @memberof EnableProjectRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {EnableProjectRequest} EnableProjectRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    EnableProjectRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.EnableProjectRequest();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.gitlab_project_id = reader.string();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return EnableProjectRequest;
})();

export const DisableProjectRequest = $root.DisableProjectRequest = (() => {

    /**
     * Properties of a DisableProjectRequest.
     * @exports IDisableProjectRequest
     * @interface IDisableProjectRequest
     * @property {string|null} [gitlab_project_id] DisableProjectRequest gitlab_project_id
     */

    /**
     * Constructs a new DisableProjectRequest.
     * @exports DisableProjectRequest
     * @classdesc Represents a DisableProjectRequest.
     * @implements IDisableProjectRequest
     * @constructor
     * @param {IDisableProjectRequest=} [properties] Properties to set
     */
    function DisableProjectRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * DisableProjectRequest gitlab_project_id.
     * @member {string} gitlab_project_id
     * @memberof DisableProjectRequest
     * @instance
     */
    DisableProjectRequest.prototype.gitlab_project_id = "";

    /**
     * Encodes the specified DisableProjectRequest message. Does not implicitly {@link DisableProjectRequest.verify|verify} messages.
     * @function encode
     * @memberof DisableProjectRequest
     * @static
     * @param {DisableProjectRequest} message DisableProjectRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    DisableProjectRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.gitlab_project_id != null && Object.hasOwnProperty.call(message, "gitlab_project_id"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.gitlab_project_id);
        return writer;
    };

    /**
     * Decodes a DisableProjectRequest message from the specified reader or buffer.
     * @function decode
     * @memberof DisableProjectRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {DisableProjectRequest} DisableProjectRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    DisableProjectRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.DisableProjectRequest();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.gitlab_project_id = reader.string();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return DisableProjectRequest;
})();

export const GitlabProjectInfo = $root.GitlabProjectInfo = (() => {

    /**
     * Properties of a GitlabProjectInfo.
     * @exports IGitlabProjectInfo
     * @interface IGitlabProjectInfo
     * @property {number|null} [id] GitlabProjectInfo id
     * @property {string|null} [name] GitlabProjectInfo name
     * @property {string|null} [path] GitlabProjectInfo path
     * @property {string|null} [web_url] GitlabProjectInfo web_url
     * @property {string|null} [avatar_url] GitlabProjectInfo avatar_url
     * @property {string|null} [description] GitlabProjectInfo description
     * @property {boolean|null} [enabled] GitlabProjectInfo enabled
     * @property {boolean|null} [global_enabled] GitlabProjectInfo global_enabled
     */

    /**
     * Constructs a new GitlabProjectInfo.
     * @exports GitlabProjectInfo
     * @classdesc Represents a GitlabProjectInfo.
     * @implements IGitlabProjectInfo
     * @constructor
     * @param {IGitlabProjectInfo=} [properties] Properties to set
     */
    function GitlabProjectInfo(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * GitlabProjectInfo id.
     * @member {number} id
     * @memberof GitlabProjectInfo
     * @instance
     */
    GitlabProjectInfo.prototype.id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * GitlabProjectInfo name.
     * @member {string} name
     * @memberof GitlabProjectInfo
     * @instance
     */
    GitlabProjectInfo.prototype.name = "";

    /**
     * GitlabProjectInfo path.
     * @member {string} path
     * @memberof GitlabProjectInfo
     * @instance
     */
    GitlabProjectInfo.prototype.path = "";

    /**
     * GitlabProjectInfo web_url.
     * @member {string} web_url
     * @memberof GitlabProjectInfo
     * @instance
     */
    GitlabProjectInfo.prototype.web_url = "";

    /**
     * GitlabProjectInfo avatar_url.
     * @member {string} avatar_url
     * @memberof GitlabProjectInfo
     * @instance
     */
    GitlabProjectInfo.prototype.avatar_url = "";

    /**
     * GitlabProjectInfo description.
     * @member {string} description
     * @memberof GitlabProjectInfo
     * @instance
     */
    GitlabProjectInfo.prototype.description = "";

    /**
     * GitlabProjectInfo enabled.
     * @member {boolean} enabled
     * @memberof GitlabProjectInfo
     * @instance
     */
    GitlabProjectInfo.prototype.enabled = false;

    /**
     * GitlabProjectInfo global_enabled.
     * @member {boolean} global_enabled
     * @memberof GitlabProjectInfo
     * @instance
     */
    GitlabProjectInfo.prototype.global_enabled = false;

    /**
     * Encodes the specified GitlabProjectInfo message. Does not implicitly {@link GitlabProjectInfo.verify|verify} messages.
     * @function encode
     * @memberof GitlabProjectInfo
     * @static
     * @param {GitlabProjectInfo} message GitlabProjectInfo message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    GitlabProjectInfo.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.id != null && Object.hasOwnProperty.call(message, "id"))
            writer.uint32(/* id 1, wireType 0 =*/8).int64(message.id);
        if (message.name != null && Object.hasOwnProperty.call(message, "name"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.name);
        if (message.path != null && Object.hasOwnProperty.call(message, "path"))
            writer.uint32(/* id 3, wireType 2 =*/26).string(message.path);
        if (message.web_url != null && Object.hasOwnProperty.call(message, "web_url"))
            writer.uint32(/* id 4, wireType 2 =*/34).string(message.web_url);
        if (message.avatar_url != null && Object.hasOwnProperty.call(message, "avatar_url"))
            writer.uint32(/* id 5, wireType 2 =*/42).string(message.avatar_url);
        if (message.description != null && Object.hasOwnProperty.call(message, "description"))
            writer.uint32(/* id 6, wireType 2 =*/50).string(message.description);
        if (message.enabled != null && Object.hasOwnProperty.call(message, "enabled"))
            writer.uint32(/* id 7, wireType 0 =*/56).bool(message.enabled);
        if (message.global_enabled != null && Object.hasOwnProperty.call(message, "global_enabled"))
            writer.uint32(/* id 8, wireType 0 =*/64).bool(message.global_enabled);
        return writer;
    };

    /**
     * Decodes a GitlabProjectInfo message from the specified reader or buffer.
     * @function decode
     * @memberof GitlabProjectInfo
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {GitlabProjectInfo} GitlabProjectInfo
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    GitlabProjectInfo.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.GitlabProjectInfo();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.id = reader.int64();
                break;
            case 2:
                message.name = reader.string();
                break;
            case 3:
                message.path = reader.string();
                break;
            case 4:
                message.web_url = reader.string();
                break;
            case 5:
                message.avatar_url = reader.string();
                break;
            case 6:
                message.description = reader.string();
                break;
            case 7:
                message.enabled = reader.bool();
                break;
            case 8:
                message.global_enabled = reader.bool();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return GitlabProjectInfo;
})();

export const ProjectListResponse = $root.ProjectListResponse = (() => {

    /**
     * Properties of a ProjectListResponse.
     * @exports IProjectListResponse
     * @interface IProjectListResponse
     * @property {Array.<GitlabProjectInfo>|null} [data] ProjectListResponse data
     */

    /**
     * Constructs a new ProjectListResponse.
     * @exports ProjectListResponse
     * @classdesc Represents a ProjectListResponse.
     * @implements IProjectListResponse
     * @constructor
     * @param {IProjectListResponse=} [properties] Properties to set
     */
    function ProjectListResponse(properties) {
        this.data = [];
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * ProjectListResponse data.
     * @member {Array.<GitlabProjectInfo>} data
     * @memberof ProjectListResponse
     * @instance
     */
    ProjectListResponse.prototype.data = $util.emptyArray;

    /**
     * Encodes the specified ProjectListResponse message. Does not implicitly {@link ProjectListResponse.verify|verify} messages.
     * @function encode
     * @memberof ProjectListResponse
     * @static
     * @param {ProjectListResponse} message ProjectListResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    ProjectListResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.data != null && message.data.length)
            for (let i = 0; i < message.data.length; ++i)
                $root.GitlabProjectInfo.encode(message.data[i], writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
        return writer;
    };

    /**
     * Decodes a ProjectListResponse message from the specified reader or buffer.
     * @function decode
     * @memberof ProjectListResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {ProjectListResponse} ProjectListResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    ProjectListResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.ProjectListResponse();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                if (!(message.data && message.data.length))
                    message.data = [];
                message.data.push($root.GitlabProjectInfo.decode(reader, reader.uint32()));
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return ProjectListResponse;
})();

export const Option = $root.Option = (() => {

    /**
     * Properties of an Option.
     * @exports IOption
     * @interface IOption
     * @property {string|null} [value] Option value
     * @property {string|null} [label] Option label
     * @property {string|null} [type] Option type
     * @property {boolean|null} [isLeaf] Option isLeaf
     * @property {string|null} [projectId] Option projectId
     * @property {string|null} [branch] Option branch
     */

    /**
     * Constructs a new Option.
     * @exports Option
     * @classdesc Represents an Option.
     * @implements IOption
     * @constructor
     * @param {IOption=} [properties] Properties to set
     */
    function Option(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * Option value.
     * @member {string} value
     * @memberof Option
     * @instance
     */
    Option.prototype.value = "";

    /**
     * Option label.
     * @member {string} label
     * @memberof Option
     * @instance
     */
    Option.prototype.label = "";

    /**
     * Option type.
     * @member {string} type
     * @memberof Option
     * @instance
     */
    Option.prototype.type = "";

    /**
     * Option isLeaf.
     * @member {boolean} isLeaf
     * @memberof Option
     * @instance
     */
    Option.prototype.isLeaf = false;

    /**
     * Option projectId.
     * @member {string} projectId
     * @memberof Option
     * @instance
     */
    Option.prototype.projectId = "";

    /**
     * Option branch.
     * @member {string} branch
     * @memberof Option
     * @instance
     */
    Option.prototype.branch = "";

    /**
     * Encodes the specified Option message. Does not implicitly {@link Option.verify|verify} messages.
     * @function encode
     * @memberof Option
     * @static
     * @param {Option} message Option message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    Option.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.value != null && Object.hasOwnProperty.call(message, "value"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.value);
        if (message.label != null && Object.hasOwnProperty.call(message, "label"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.label);
        if (message.type != null && Object.hasOwnProperty.call(message, "type"))
            writer.uint32(/* id 3, wireType 2 =*/26).string(message.type);
        if (message.isLeaf != null && Object.hasOwnProperty.call(message, "isLeaf"))
            writer.uint32(/* id 4, wireType 0 =*/32).bool(message.isLeaf);
        if (message.projectId != null && Object.hasOwnProperty.call(message, "projectId"))
            writer.uint32(/* id 5, wireType 2 =*/42).string(message.projectId);
        if (message.branch != null && Object.hasOwnProperty.call(message, "branch"))
            writer.uint32(/* id 6, wireType 2 =*/50).string(message.branch);
        return writer;
    };

    /**
     * Decodes an Option message from the specified reader or buffer.
     * @function decode
     * @memberof Option
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {Option} Option
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    Option.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.Option();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.value = reader.string();
                break;
            case 2:
                message.label = reader.string();
                break;
            case 3:
                message.type = reader.string();
                break;
            case 4:
                message.isLeaf = reader.bool();
                break;
            case 5:
                message.projectId = reader.string();
                break;
            case 6:
                message.branch = reader.string();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return Option;
})();

export const ProjectsResponse = $root.ProjectsResponse = (() => {

    /**
     * Properties of a ProjectsResponse.
     * @exports IProjectsResponse
     * @interface IProjectsResponse
     * @property {Array.<Option>|null} [data] ProjectsResponse data
     */

    /**
     * Constructs a new ProjectsResponse.
     * @exports ProjectsResponse
     * @classdesc Represents a ProjectsResponse.
     * @implements IProjectsResponse
     * @constructor
     * @param {IProjectsResponse=} [properties] Properties to set
     */
    function ProjectsResponse(properties) {
        this.data = [];
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * ProjectsResponse data.
     * @member {Array.<Option>} data
     * @memberof ProjectsResponse
     * @instance
     */
    ProjectsResponse.prototype.data = $util.emptyArray;

    /**
     * Encodes the specified ProjectsResponse message. Does not implicitly {@link ProjectsResponse.verify|verify} messages.
     * @function encode
     * @memberof ProjectsResponse
     * @static
     * @param {ProjectsResponse} message ProjectsResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    ProjectsResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.data != null && message.data.length)
            for (let i = 0; i < message.data.length; ++i)
                $root.Option.encode(message.data[i], writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
        return writer;
    };

    /**
     * Decodes a ProjectsResponse message from the specified reader or buffer.
     * @function decode
     * @memberof ProjectsResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {ProjectsResponse} ProjectsResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    ProjectsResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.ProjectsResponse();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                if (!(message.data && message.data.length))
                    message.data = [];
                message.data.push($root.Option.decode(reader, reader.uint32()));
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return ProjectsResponse;
})();

export const BranchesRequest = $root.BranchesRequest = (() => {

    /**
     * Properties of a BranchesRequest.
     * @exports IBranchesRequest
     * @interface IBranchesRequest
     * @property {string|null} [project_id] BranchesRequest project_id
     * @property {boolean|null} [all] BranchesRequest all
     */

    /**
     * Constructs a new BranchesRequest.
     * @exports BranchesRequest
     * @classdesc Represents a BranchesRequest.
     * @implements IBranchesRequest
     * @constructor
     * @param {IBranchesRequest=} [properties] Properties to set
     */
    function BranchesRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * BranchesRequest project_id.
     * @member {string} project_id
     * @memberof BranchesRequest
     * @instance
     */
    BranchesRequest.prototype.project_id = "";

    /**
     * BranchesRequest all.
     * @member {boolean} all
     * @memberof BranchesRequest
     * @instance
     */
    BranchesRequest.prototype.all = false;

    /**
     * Encodes the specified BranchesRequest message. Does not implicitly {@link BranchesRequest.verify|verify} messages.
     * @function encode
     * @memberof BranchesRequest
     * @static
     * @param {BranchesRequest} message BranchesRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    BranchesRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.project_id != null && Object.hasOwnProperty.call(message, "project_id"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.project_id);
        if (message.all != null && Object.hasOwnProperty.call(message, "all"))
            writer.uint32(/* id 2, wireType 0 =*/16).bool(message.all);
        return writer;
    };

    /**
     * Decodes a BranchesRequest message from the specified reader or buffer.
     * @function decode
     * @memberof BranchesRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {BranchesRequest} BranchesRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    BranchesRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.BranchesRequest();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.project_id = reader.string();
                break;
            case 2:
                message.all = reader.bool();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return BranchesRequest;
})();

export const BranchesResponse = $root.BranchesResponse = (() => {

    /**
     * Properties of a BranchesResponse.
     * @exports IBranchesResponse
     * @interface IBranchesResponse
     * @property {Array.<Option>|null} [data] BranchesResponse data
     */

    /**
     * Constructs a new BranchesResponse.
     * @exports BranchesResponse
     * @classdesc Represents a BranchesResponse.
     * @implements IBranchesResponse
     * @constructor
     * @param {IBranchesResponse=} [properties] Properties to set
     */
    function BranchesResponse(properties) {
        this.data = [];
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * BranchesResponse data.
     * @member {Array.<Option>} data
     * @memberof BranchesResponse
     * @instance
     */
    BranchesResponse.prototype.data = $util.emptyArray;

    /**
     * Encodes the specified BranchesResponse message. Does not implicitly {@link BranchesResponse.verify|verify} messages.
     * @function encode
     * @memberof BranchesResponse
     * @static
     * @param {BranchesResponse} message BranchesResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    BranchesResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.data != null && message.data.length)
            for (let i = 0; i < message.data.length; ++i)
                $root.Option.encode(message.data[i], writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
        return writer;
    };

    /**
     * Decodes a BranchesResponse message from the specified reader or buffer.
     * @function decode
     * @memberof BranchesResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {BranchesResponse} BranchesResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    BranchesResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.BranchesResponse();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                if (!(message.data && message.data.length))
                    message.data = [];
                message.data.push($root.Option.decode(reader, reader.uint32()));
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return BranchesResponse;
})();

export const CommitsRequest = $root.CommitsRequest = (() => {

    /**
     * Properties of a CommitsRequest.
     * @exports ICommitsRequest
     * @interface ICommitsRequest
     * @property {string|null} [project_id] CommitsRequest project_id
     * @property {string|null} [branch] CommitsRequest branch
     */

    /**
     * Constructs a new CommitsRequest.
     * @exports CommitsRequest
     * @classdesc Represents a CommitsRequest.
     * @implements ICommitsRequest
     * @constructor
     * @param {ICommitsRequest=} [properties] Properties to set
     */
    function CommitsRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * CommitsRequest project_id.
     * @member {string} project_id
     * @memberof CommitsRequest
     * @instance
     */
    CommitsRequest.prototype.project_id = "";

    /**
     * CommitsRequest branch.
     * @member {string} branch
     * @memberof CommitsRequest
     * @instance
     */
    CommitsRequest.prototype.branch = "";

    /**
     * Encodes the specified CommitsRequest message. Does not implicitly {@link CommitsRequest.verify|verify} messages.
     * @function encode
     * @memberof CommitsRequest
     * @static
     * @param {CommitsRequest} message CommitsRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    CommitsRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.project_id != null && Object.hasOwnProperty.call(message, "project_id"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.project_id);
        if (message.branch != null && Object.hasOwnProperty.call(message, "branch"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.branch);
        return writer;
    };

    /**
     * Decodes a CommitsRequest message from the specified reader or buffer.
     * @function decode
     * @memberof CommitsRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {CommitsRequest} CommitsRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    CommitsRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.CommitsRequest();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.project_id = reader.string();
                break;
            case 2:
                message.branch = reader.string();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return CommitsRequest;
})();

export const CommitsResponse = $root.CommitsResponse = (() => {

    /**
     * Properties of a CommitsResponse.
     * @exports ICommitsResponse
     * @interface ICommitsResponse
     * @property {Array.<Option>|null} [data] CommitsResponse data
     */

    /**
     * Constructs a new CommitsResponse.
     * @exports CommitsResponse
     * @classdesc Represents a CommitsResponse.
     * @implements ICommitsResponse
     * @constructor
     * @param {ICommitsResponse=} [properties] Properties to set
     */
    function CommitsResponse(properties) {
        this.data = [];
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * CommitsResponse data.
     * @member {Array.<Option>} data
     * @memberof CommitsResponse
     * @instance
     */
    CommitsResponse.prototype.data = $util.emptyArray;

    /**
     * Encodes the specified CommitsResponse message. Does not implicitly {@link CommitsResponse.verify|verify} messages.
     * @function encode
     * @memberof CommitsResponse
     * @static
     * @param {CommitsResponse} message CommitsResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    CommitsResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.data != null && message.data.length)
            for (let i = 0; i < message.data.length; ++i)
                $root.Option.encode(message.data[i], writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
        return writer;
    };

    /**
     * Decodes a CommitsResponse message from the specified reader or buffer.
     * @function decode
     * @memberof CommitsResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {CommitsResponse} CommitsResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    CommitsResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.CommitsResponse();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                if (!(message.data && message.data.length))
                    message.data = [];
                message.data.push($root.Option.decode(reader, reader.uint32()));
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return CommitsResponse;
})();

export const CommitRequest = $root.CommitRequest = (() => {

    /**
     * Properties of a CommitRequest.
     * @exports ICommitRequest
     * @interface ICommitRequest
     * @property {string|null} [project_id] CommitRequest project_id
     * @property {string|null} [branch] CommitRequest branch
     * @property {string|null} [commit] CommitRequest commit
     */

    /**
     * Constructs a new CommitRequest.
     * @exports CommitRequest
     * @classdesc Represents a CommitRequest.
     * @implements ICommitRequest
     * @constructor
     * @param {ICommitRequest=} [properties] Properties to set
     */
    function CommitRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * CommitRequest project_id.
     * @member {string} project_id
     * @memberof CommitRequest
     * @instance
     */
    CommitRequest.prototype.project_id = "";

    /**
     * CommitRequest branch.
     * @member {string} branch
     * @memberof CommitRequest
     * @instance
     */
    CommitRequest.prototype.branch = "";

    /**
     * CommitRequest commit.
     * @member {string} commit
     * @memberof CommitRequest
     * @instance
     */
    CommitRequest.prototype.commit = "";

    /**
     * Encodes the specified CommitRequest message. Does not implicitly {@link CommitRequest.verify|verify} messages.
     * @function encode
     * @memberof CommitRequest
     * @static
     * @param {CommitRequest} message CommitRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    CommitRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.project_id != null && Object.hasOwnProperty.call(message, "project_id"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.project_id);
        if (message.branch != null && Object.hasOwnProperty.call(message, "branch"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.branch);
        if (message.commit != null && Object.hasOwnProperty.call(message, "commit"))
            writer.uint32(/* id 3, wireType 2 =*/26).string(message.commit);
        return writer;
    };

    /**
     * Decodes a CommitRequest message from the specified reader or buffer.
     * @function decode
     * @memberof CommitRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {CommitRequest} CommitRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    CommitRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.CommitRequest();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.project_id = reader.string();
                break;
            case 2:
                message.branch = reader.string();
                break;
            case 3:
                message.commit = reader.string();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return CommitRequest;
})();

export const CommitResponse = $root.CommitResponse = (() => {

    /**
     * Properties of a CommitResponse.
     * @exports ICommitResponse
     * @interface ICommitResponse
     * @property {Option|null} [data] CommitResponse data
     */

    /**
     * Constructs a new CommitResponse.
     * @exports CommitResponse
     * @classdesc Represents a CommitResponse.
     * @implements ICommitResponse
     * @constructor
     * @param {ICommitResponse=} [properties] Properties to set
     */
    function CommitResponse(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * CommitResponse data.
     * @member {Option|null|undefined} data
     * @memberof CommitResponse
     * @instance
     */
    CommitResponse.prototype.data = null;

    /**
     * Encodes the specified CommitResponse message. Does not implicitly {@link CommitResponse.verify|verify} messages.
     * @function encode
     * @memberof CommitResponse
     * @static
     * @param {CommitResponse} message CommitResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    CommitResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.data != null && Object.hasOwnProperty.call(message, "data"))
            $root.Option.encode(message.data, writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
        return writer;
    };

    /**
     * Decodes a CommitResponse message from the specified reader or buffer.
     * @function decode
     * @memberof CommitResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {CommitResponse} CommitResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    CommitResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.CommitResponse();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.data = $root.Option.decode(reader, reader.uint32());
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return CommitResponse;
})();

export const PipelineInfoRequest = $root.PipelineInfoRequest = (() => {

    /**
     * Properties of a PipelineInfoRequest.
     * @exports IPipelineInfoRequest
     * @interface IPipelineInfoRequest
     * @property {string|null} [project_id] PipelineInfoRequest project_id
     * @property {string|null} [branch] PipelineInfoRequest branch
     * @property {string|null} [commit] PipelineInfoRequest commit
     */

    /**
     * Constructs a new PipelineInfoRequest.
     * @exports PipelineInfoRequest
     * @classdesc Represents a PipelineInfoRequest.
     * @implements IPipelineInfoRequest
     * @constructor
     * @param {IPipelineInfoRequest=} [properties] Properties to set
     */
    function PipelineInfoRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * PipelineInfoRequest project_id.
     * @member {string} project_id
     * @memberof PipelineInfoRequest
     * @instance
     */
    PipelineInfoRequest.prototype.project_id = "";

    /**
     * PipelineInfoRequest branch.
     * @member {string} branch
     * @memberof PipelineInfoRequest
     * @instance
     */
    PipelineInfoRequest.prototype.branch = "";

    /**
     * PipelineInfoRequest commit.
     * @member {string} commit
     * @memberof PipelineInfoRequest
     * @instance
     */
    PipelineInfoRequest.prototype.commit = "";

    /**
     * Encodes the specified PipelineInfoRequest message. Does not implicitly {@link PipelineInfoRequest.verify|verify} messages.
     * @function encode
     * @memberof PipelineInfoRequest
     * @static
     * @param {PipelineInfoRequest} message PipelineInfoRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    PipelineInfoRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.project_id != null && Object.hasOwnProperty.call(message, "project_id"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.project_id);
        if (message.branch != null && Object.hasOwnProperty.call(message, "branch"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.branch);
        if (message.commit != null && Object.hasOwnProperty.call(message, "commit"))
            writer.uint32(/* id 3, wireType 2 =*/26).string(message.commit);
        return writer;
    };

    /**
     * Decodes a PipelineInfoRequest message from the specified reader or buffer.
     * @function decode
     * @memberof PipelineInfoRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {PipelineInfoRequest} PipelineInfoRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    PipelineInfoRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.PipelineInfoRequest();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.project_id = reader.string();
                break;
            case 2:
                message.branch = reader.string();
                break;
            case 3:
                message.commit = reader.string();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return PipelineInfoRequest;
})();

export const PipelineInfoResponse = $root.PipelineInfoResponse = (() => {

    /**
     * Properties of a PipelineInfoResponse.
     * @exports IPipelineInfoResponse
     * @interface IPipelineInfoResponse
     * @property {string|null} [status] PipelineInfoResponse status
     * @property {string|null} [web_url] PipelineInfoResponse web_url
     */

    /**
     * Constructs a new PipelineInfoResponse.
     * @exports PipelineInfoResponse
     * @classdesc Represents a PipelineInfoResponse.
     * @implements IPipelineInfoResponse
     * @constructor
     * @param {IPipelineInfoResponse=} [properties] Properties to set
     */
    function PipelineInfoResponse(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * PipelineInfoResponse status.
     * @member {string} status
     * @memberof PipelineInfoResponse
     * @instance
     */
    PipelineInfoResponse.prototype.status = "";

    /**
     * PipelineInfoResponse web_url.
     * @member {string} web_url
     * @memberof PipelineInfoResponse
     * @instance
     */
    PipelineInfoResponse.prototype.web_url = "";

    /**
     * Encodes the specified PipelineInfoResponse message. Does not implicitly {@link PipelineInfoResponse.verify|verify} messages.
     * @function encode
     * @memberof PipelineInfoResponse
     * @static
     * @param {PipelineInfoResponse} message PipelineInfoResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    PipelineInfoResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.status != null && Object.hasOwnProperty.call(message, "status"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.status);
        if (message.web_url != null && Object.hasOwnProperty.call(message, "web_url"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.web_url);
        return writer;
    };

    /**
     * Decodes a PipelineInfoResponse message from the specified reader or buffer.
     * @function decode
     * @memberof PipelineInfoResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {PipelineInfoResponse} PipelineInfoResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    PipelineInfoResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.PipelineInfoResponse();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.status = reader.string();
                break;
            case 2:
                message.web_url = reader.string();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return PipelineInfoResponse;
})();

export const ConfigFileRequest = $root.ConfigFileRequest = (() => {

    /**
     * Properties of a ConfigFileRequest.
     * @exports IConfigFileRequest
     * @interface IConfigFileRequest
     * @property {string|null} [project_id] ConfigFileRequest project_id
     * @property {string|null} [branch] ConfigFileRequest branch
     */

    /**
     * Constructs a new ConfigFileRequest.
     * @exports ConfigFileRequest
     * @classdesc Represents a ConfigFileRequest.
     * @implements IConfigFileRequest
     * @constructor
     * @param {IConfigFileRequest=} [properties] Properties to set
     */
    function ConfigFileRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * ConfigFileRequest project_id.
     * @member {string} project_id
     * @memberof ConfigFileRequest
     * @instance
     */
    ConfigFileRequest.prototype.project_id = "";

    /**
     * ConfigFileRequest branch.
     * @member {string} branch
     * @memberof ConfigFileRequest
     * @instance
     */
    ConfigFileRequest.prototype.branch = "";

    /**
     * Encodes the specified ConfigFileRequest message. Does not implicitly {@link ConfigFileRequest.verify|verify} messages.
     * @function encode
     * @memberof ConfigFileRequest
     * @static
     * @param {ConfigFileRequest} message ConfigFileRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    ConfigFileRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.project_id != null && Object.hasOwnProperty.call(message, "project_id"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.project_id);
        if (message.branch != null && Object.hasOwnProperty.call(message, "branch"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.branch);
        return writer;
    };

    /**
     * Decodes a ConfigFileRequest message from the specified reader or buffer.
     * @function decode
     * @memberof ConfigFileRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {ConfigFileRequest} ConfigFileRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    ConfigFileRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.ConfigFileRequest();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.project_id = reader.string();
                break;
            case 2:
                message.branch = reader.string();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return ConfigFileRequest;
})();

export const ConfigFileResponse = $root.ConfigFileResponse = (() => {

    /**
     * Properties of a ConfigFileResponse.
     * @exports IConfigFileResponse
     * @interface IConfigFileResponse
     * @property {string|null} [data] ConfigFileResponse data
     * @property {string|null} [type] ConfigFileResponse type
     */

    /**
     * Constructs a new ConfigFileResponse.
     * @exports ConfigFileResponse
     * @classdesc Represents a ConfigFileResponse.
     * @implements IConfigFileResponse
     * @constructor
     * @param {IConfigFileResponse=} [properties] Properties to set
     */
    function ConfigFileResponse(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * ConfigFileResponse data.
     * @member {string} data
     * @memberof ConfigFileResponse
     * @instance
     */
    ConfigFileResponse.prototype.data = "";

    /**
     * ConfigFileResponse type.
     * @member {string} type
     * @memberof ConfigFileResponse
     * @instance
     */
    ConfigFileResponse.prototype.type = "";

    /**
     * Encodes the specified ConfigFileResponse message. Does not implicitly {@link ConfigFileResponse.verify|verify} messages.
     * @function encode
     * @memberof ConfigFileResponse
     * @static
     * @param {ConfigFileResponse} message ConfigFileResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    ConfigFileResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.data != null && Object.hasOwnProperty.call(message, "data"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.data);
        if (message.type != null && Object.hasOwnProperty.call(message, "type"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.type);
        return writer;
    };

    /**
     * Decodes a ConfigFileResponse message from the specified reader or buffer.
     * @function decode
     * @memberof ConfigFileResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {ConfigFileResponse} ConfigFileResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    ConfigFileResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.ConfigFileResponse();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.data = reader.string();
                break;
            case 2:
                message.type = reader.string();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return ConfigFileResponse;
})();

export const Gitlab = $root.Gitlab = (() => {

    /**
     * Constructs a new Gitlab service.
     * @exports Gitlab
     * @classdesc Represents a Gitlab
     * @extends $protobuf.rpc.Service
     * @constructor
     * @param {$protobuf.RPCImpl} rpcImpl RPC implementation
     * @param {boolean} [requestDelimited=false] Whether requests are length-delimited
     * @param {boolean} [responseDelimited=false] Whether responses are length-delimited
     */
    function Gitlab(rpcImpl, requestDelimited, responseDelimited) {
        $protobuf.rpc.Service.call(this, rpcImpl, requestDelimited, responseDelimited);
    }

    (Gitlab.prototype = Object.create($protobuf.rpc.Service.prototype)).constructor = Gitlab;

    /**
     * Callback as used by {@link Gitlab#enableProject}.
     * @memberof Gitlab
     * @typedef EnableProjectCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {google.protobuf.Empty} [response] Empty
     */

    /**
     * Calls EnableProject.
     * @function enableProject
     * @memberof Gitlab
     * @instance
     * @param {EnableProjectRequest} request EnableProjectRequest message or plain object
     * @param {Gitlab.EnableProjectCallback} callback Node-style callback called with the error, if any, and Empty
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Gitlab.prototype.enableProject = function enableProject(request, callback) {
        return this.rpcCall(enableProject, $root.EnableProjectRequest, $root.google.protobuf.Empty, request, callback);
    }, "name", { value: "EnableProject" });

    /**
     * Calls EnableProject.
     * @function enableProject
     * @memberof Gitlab
     * @instance
     * @param {EnableProjectRequest} request EnableProjectRequest message or plain object
     * @returns {Promise<google.protobuf.Empty>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link Gitlab#disableProject}.
     * @memberof Gitlab
     * @typedef DisableProjectCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {google.protobuf.Empty} [response] Empty
     */

    /**
     * Calls DisableProject.
     * @function disableProject
     * @memberof Gitlab
     * @instance
     * @param {DisableProjectRequest} request DisableProjectRequest message or plain object
     * @param {Gitlab.DisableProjectCallback} callback Node-style callback called with the error, if any, and Empty
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Gitlab.prototype.disableProject = function disableProject(request, callback) {
        return this.rpcCall(disableProject, $root.DisableProjectRequest, $root.google.protobuf.Empty, request, callback);
    }, "name", { value: "DisableProject" });

    /**
     * Calls DisableProject.
     * @function disableProject
     * @memberof Gitlab
     * @instance
     * @param {DisableProjectRequest} request DisableProjectRequest message or plain object
     * @returns {Promise<google.protobuf.Empty>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link Gitlab#projectList}.
     * @memberof Gitlab
     * @typedef ProjectListCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {ProjectListResponse} [response] ProjectListResponse
     */

    /**
     * Calls ProjectList.
     * @function projectList
     * @memberof Gitlab
     * @instance
     * @param {google.protobuf.Empty} request Empty message or plain object
     * @param {Gitlab.ProjectListCallback} callback Node-style callback called with the error, if any, and ProjectListResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Gitlab.prototype.projectList = function projectList(request, callback) {
        return this.rpcCall(projectList, $root.google.protobuf.Empty, $root.ProjectListResponse, request, callback);
    }, "name", { value: "ProjectList" });

    /**
     * Calls ProjectList.
     * @function projectList
     * @memberof Gitlab
     * @instance
     * @param {google.protobuf.Empty} request Empty message or plain object
     * @returns {Promise<ProjectListResponse>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link Gitlab#projects}.
     * @memberof Gitlab
     * @typedef ProjectsCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {ProjectsResponse} [response] ProjectsResponse
     */

    /**
     * Calls Projects.
     * @function projects
     * @memberof Gitlab
     * @instance
     * @param {google.protobuf.Empty} request Empty message or plain object
     * @param {Gitlab.ProjectsCallback} callback Node-style callback called with the error, if any, and ProjectsResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Gitlab.prototype.projects = function projects(request, callback) {
        return this.rpcCall(projects, $root.google.protobuf.Empty, $root.ProjectsResponse, request, callback);
    }, "name", { value: "Projects" });

    /**
     * Calls Projects.
     * @function projects
     * @memberof Gitlab
     * @instance
     * @param {google.protobuf.Empty} request Empty message or plain object
     * @returns {Promise<ProjectsResponse>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link Gitlab#branches}.
     * @memberof Gitlab
     * @typedef BranchesCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {BranchesResponse} [response] BranchesResponse
     */

    /**
     * Calls Branches.
     * @function branches
     * @memberof Gitlab
     * @instance
     * @param {BranchesRequest} request BranchesRequest message or plain object
     * @param {Gitlab.BranchesCallback} callback Node-style callback called with the error, if any, and BranchesResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Gitlab.prototype.branches = function branches(request, callback) {
        return this.rpcCall(branches, $root.BranchesRequest, $root.BranchesResponse, request, callback);
    }, "name", { value: "Branches" });

    /**
     * Calls Branches.
     * @function branches
     * @memberof Gitlab
     * @instance
     * @param {BranchesRequest} request BranchesRequest message or plain object
     * @returns {Promise<BranchesResponse>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link Gitlab#commits}.
     * @memberof Gitlab
     * @typedef CommitsCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {CommitsResponse} [response] CommitsResponse
     */

    /**
     * Calls Commits.
     * @function commits
     * @memberof Gitlab
     * @instance
     * @param {CommitsRequest} request CommitsRequest message or plain object
     * @param {Gitlab.CommitsCallback} callback Node-style callback called with the error, if any, and CommitsResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Gitlab.prototype.commits = function commits(request, callback) {
        return this.rpcCall(commits, $root.CommitsRequest, $root.CommitsResponse, request, callback);
    }, "name", { value: "Commits" });

    /**
     * Calls Commits.
     * @function commits
     * @memberof Gitlab
     * @instance
     * @param {CommitsRequest} request CommitsRequest message or plain object
     * @returns {Promise<CommitsResponse>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link Gitlab#commit}.
     * @memberof Gitlab
     * @typedef CommitCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {CommitResponse} [response] CommitResponse
     */

    /**
     * Calls Commit.
     * @function commit
     * @memberof Gitlab
     * @instance
     * @param {CommitRequest} request CommitRequest message or plain object
     * @param {Gitlab.CommitCallback} callback Node-style callback called with the error, if any, and CommitResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Gitlab.prototype.commit = function commit(request, callback) {
        return this.rpcCall(commit, $root.CommitRequest, $root.CommitResponse, request, callback);
    }, "name", { value: "Commit" });

    /**
     * Calls Commit.
     * @function commit
     * @memberof Gitlab
     * @instance
     * @param {CommitRequest} request CommitRequest message or plain object
     * @returns {Promise<CommitResponse>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link Gitlab#pipelineInfo}.
     * @memberof Gitlab
     * @typedef PipelineInfoCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {PipelineInfoResponse} [response] PipelineInfoResponse
     */

    /**
     * Calls PipelineInfo.
     * @function pipelineInfo
     * @memberof Gitlab
     * @instance
     * @param {PipelineInfoRequest} request PipelineInfoRequest message or plain object
     * @param {Gitlab.PipelineInfoCallback} callback Node-style callback called with the error, if any, and PipelineInfoResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Gitlab.prototype.pipelineInfo = function pipelineInfo(request, callback) {
        return this.rpcCall(pipelineInfo, $root.PipelineInfoRequest, $root.PipelineInfoResponse, request, callback);
    }, "name", { value: "PipelineInfo" });

    /**
     * Calls PipelineInfo.
     * @function pipelineInfo
     * @memberof Gitlab
     * @instance
     * @param {PipelineInfoRequest} request PipelineInfoRequest message or plain object
     * @returns {Promise<PipelineInfoResponse>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link Gitlab#configFile}.
     * @memberof Gitlab
     * @typedef ConfigFileCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {ConfigFileResponse} [response] ConfigFileResponse
     */

    /**
     * Calls ConfigFile.
     * @function configFile
     * @memberof Gitlab
     * @instance
     * @param {ConfigFileRequest} request ConfigFileRequest message or plain object
     * @param {Gitlab.ConfigFileCallback} callback Node-style callback called with the error, if any, and ConfigFileResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Gitlab.prototype.configFile = function configFile(request, callback) {
        return this.rpcCall(configFile, $root.ConfigFileRequest, $root.ConfigFileResponse, request, callback);
    }, "name", { value: "ConfigFile" });

    /**
     * Calls ConfigFile.
     * @function configFile
     * @memberof Gitlab
     * @instance
     * @param {ConfigFileRequest} request ConfigFileRequest message or plain object
     * @returns {Promise<ConfigFileResponse>} Promise
     * @variation 2
     */

    return Gitlab;
})();

export const Config = $root.Config = (() => {

    /**
     * Properties of a Config.
     * @exports IConfig
     * @interface IConfig
     * @property {string|null} [config_file] Config config_file
     * @property {string|null} [config_file_values] Config config_file_values
     * @property {string|null} [config_field] Config config_field
     * @property {boolean|null} [is_simple_env] Config is_simple_env
     * @property {string|null} [config_file_type] Config config_file_type
     * @property {string|null} [local_chart_path] Config local_chart_path
     * @property {Array.<string>|null} [branches] Config branches
     * @property {string|null} [values_yaml] Config values_yaml
     */

    /**
     * Constructs a new Config.
     * @exports Config
     * @classdesc Represents a Config.
     * @implements IConfig
     * @constructor
     * @param {IConfig=} [properties] Properties to set
     */
    function Config(properties) {
        this.branches = [];
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * Config config_file.
     * @member {string} config_file
     * @memberof Config
     * @instance
     */
    Config.prototype.config_file = "";

    /**
     * Config config_file_values.
     * @member {string} config_file_values
     * @memberof Config
     * @instance
     */
    Config.prototype.config_file_values = "";

    /**
     * Config config_field.
     * @member {string} config_field
     * @memberof Config
     * @instance
     */
    Config.prototype.config_field = "";

    /**
     * Config is_simple_env.
     * @member {boolean} is_simple_env
     * @memberof Config
     * @instance
     */
    Config.prototype.is_simple_env = false;

    /**
     * Config config_file_type.
     * @member {string} config_file_type
     * @memberof Config
     * @instance
     */
    Config.prototype.config_file_type = "";

    /**
     * Config local_chart_path.
     * @member {string} local_chart_path
     * @memberof Config
     * @instance
     */
    Config.prototype.local_chart_path = "";

    /**
     * Config branches.
     * @member {Array.<string>} branches
     * @memberof Config
     * @instance
     */
    Config.prototype.branches = $util.emptyArray;

    /**
     * Config values_yaml.
     * @member {string} values_yaml
     * @memberof Config
     * @instance
     */
    Config.prototype.values_yaml = "";

    /**
     * Encodes the specified Config message. Does not implicitly {@link Config.verify|verify} messages.
     * @function encode
     * @memberof Config
     * @static
     * @param {Config} message Config message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    Config.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.config_file != null && Object.hasOwnProperty.call(message, "config_file"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.config_file);
        if (message.config_file_values != null && Object.hasOwnProperty.call(message, "config_file_values"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.config_file_values);
        if (message.config_field != null && Object.hasOwnProperty.call(message, "config_field"))
            writer.uint32(/* id 3, wireType 2 =*/26).string(message.config_field);
        if (message.is_simple_env != null && Object.hasOwnProperty.call(message, "is_simple_env"))
            writer.uint32(/* id 4, wireType 0 =*/32).bool(message.is_simple_env);
        if (message.config_file_type != null && Object.hasOwnProperty.call(message, "config_file_type"))
            writer.uint32(/* id 5, wireType 2 =*/42).string(message.config_file_type);
        if (message.local_chart_path != null && Object.hasOwnProperty.call(message, "local_chart_path"))
            writer.uint32(/* id 6, wireType 2 =*/50).string(message.local_chart_path);
        if (message.branches != null && message.branches.length)
            for (let i = 0; i < message.branches.length; ++i)
                writer.uint32(/* id 7, wireType 2 =*/58).string(message.branches[i]);
        if (message.values_yaml != null && Object.hasOwnProperty.call(message, "values_yaml"))
            writer.uint32(/* id 8, wireType 2 =*/66).string(message.values_yaml);
        return writer;
    };

    /**
     * Decodes a Config message from the specified reader or buffer.
     * @function decode
     * @memberof Config
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {Config} Config
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    Config.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.Config();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.config_file = reader.string();
                break;
            case 2:
                message.config_file_values = reader.string();
                break;
            case 3:
                message.config_field = reader.string();
                break;
            case 4:
                message.is_simple_env = reader.bool();
                break;
            case 5:
                message.config_file_type = reader.string();
                break;
            case 6:
                message.local_chart_path = reader.string();
                break;
            case 7:
                if (!(message.branches && message.branches.length))
                    message.branches = [];
                message.branches.push(reader.string());
                break;
            case 8:
                message.values_yaml = reader.string();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return Config;
})();

export const MarsShowRequest = $root.MarsShowRequest = (() => {

    /**
     * Properties of a MarsShowRequest.
     * @exports IMarsShowRequest
     * @interface IMarsShowRequest
     * @property {number|null} [project_id] MarsShowRequest project_id
     * @property {string|null} [branch] MarsShowRequest branch
     */

    /**
     * Constructs a new MarsShowRequest.
     * @exports MarsShowRequest
     * @classdesc Represents a MarsShowRequest.
     * @implements IMarsShowRequest
     * @constructor
     * @param {IMarsShowRequest=} [properties] Properties to set
     */
    function MarsShowRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * MarsShowRequest project_id.
     * @member {number} project_id
     * @memberof MarsShowRequest
     * @instance
     */
    MarsShowRequest.prototype.project_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * MarsShowRequest branch.
     * @member {string} branch
     * @memberof MarsShowRequest
     * @instance
     */
    MarsShowRequest.prototype.branch = "";

    /**
     * Encodes the specified MarsShowRequest message. Does not implicitly {@link MarsShowRequest.verify|verify} messages.
     * @function encode
     * @memberof MarsShowRequest
     * @static
     * @param {MarsShowRequest} message MarsShowRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    MarsShowRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.project_id != null && Object.hasOwnProperty.call(message, "project_id"))
            writer.uint32(/* id 1, wireType 0 =*/8).int64(message.project_id);
        if (message.branch != null && Object.hasOwnProperty.call(message, "branch"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.branch);
        return writer;
    };

    /**
     * Decodes a MarsShowRequest message from the specified reader or buffer.
     * @function decode
     * @memberof MarsShowRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {MarsShowRequest} MarsShowRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    MarsShowRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.MarsShowRequest();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.project_id = reader.int64();
                break;
            case 2:
                message.branch = reader.string();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return MarsShowRequest;
})();

export const MarsShowResponse = $root.MarsShowResponse = (() => {

    /**
     * Properties of a MarsShowResponse.
     * @exports IMarsShowResponse
     * @interface IMarsShowResponse
     * @property {string|null} [branch] MarsShowResponse branch
     * @property {Config|null} [config] MarsShowResponse config
     */

    /**
     * Constructs a new MarsShowResponse.
     * @exports MarsShowResponse
     * @classdesc Represents a MarsShowResponse.
     * @implements IMarsShowResponse
     * @constructor
     * @param {IMarsShowResponse=} [properties] Properties to set
     */
    function MarsShowResponse(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * MarsShowResponse branch.
     * @member {string} branch
     * @memberof MarsShowResponse
     * @instance
     */
    MarsShowResponse.prototype.branch = "";

    /**
     * MarsShowResponse config.
     * @member {Config|null|undefined} config
     * @memberof MarsShowResponse
     * @instance
     */
    MarsShowResponse.prototype.config = null;

    /**
     * Encodes the specified MarsShowResponse message. Does not implicitly {@link MarsShowResponse.verify|verify} messages.
     * @function encode
     * @memberof MarsShowResponse
     * @static
     * @param {MarsShowResponse} message MarsShowResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    MarsShowResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.branch != null && Object.hasOwnProperty.call(message, "branch"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.branch);
        if (message.config != null && Object.hasOwnProperty.call(message, "config"))
            $root.Config.encode(message.config, writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim();
        return writer;
    };

    /**
     * Decodes a MarsShowResponse message from the specified reader or buffer.
     * @function decode
     * @memberof MarsShowResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {MarsShowResponse} MarsShowResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    MarsShowResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.MarsShowResponse();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.branch = reader.string();
                break;
            case 2:
                message.config = $root.Config.decode(reader, reader.uint32());
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return MarsShowResponse;
})();

export const GlobalConfigRequest = $root.GlobalConfigRequest = (() => {

    /**
     * Properties of a GlobalConfigRequest.
     * @exports IGlobalConfigRequest
     * @interface IGlobalConfigRequest
     * @property {number|null} [project_id] GlobalConfigRequest project_id
     */

    /**
     * Constructs a new GlobalConfigRequest.
     * @exports GlobalConfigRequest
     * @classdesc Represents a GlobalConfigRequest.
     * @implements IGlobalConfigRequest
     * @constructor
     * @param {IGlobalConfigRequest=} [properties] Properties to set
     */
    function GlobalConfigRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * GlobalConfigRequest project_id.
     * @member {number} project_id
     * @memberof GlobalConfigRequest
     * @instance
     */
    GlobalConfigRequest.prototype.project_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * Encodes the specified GlobalConfigRequest message. Does not implicitly {@link GlobalConfigRequest.verify|verify} messages.
     * @function encode
     * @memberof GlobalConfigRequest
     * @static
     * @param {GlobalConfigRequest} message GlobalConfigRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    GlobalConfigRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.project_id != null && Object.hasOwnProperty.call(message, "project_id"))
            writer.uint32(/* id 1, wireType 0 =*/8).int64(message.project_id);
        return writer;
    };

    /**
     * Decodes a GlobalConfigRequest message from the specified reader or buffer.
     * @function decode
     * @memberof GlobalConfigRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {GlobalConfigRequest} GlobalConfigRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    GlobalConfigRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.GlobalConfigRequest();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.project_id = reader.int64();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return GlobalConfigRequest;
})();

export const GlobalConfigResponse = $root.GlobalConfigResponse = (() => {

    /**
     * Properties of a GlobalConfigResponse.
     * @exports IGlobalConfigResponse
     * @interface IGlobalConfigResponse
     * @property {boolean|null} [enabled] GlobalConfigResponse enabled
     * @property {Config|null} [config] GlobalConfigResponse config
     */

    /**
     * Constructs a new GlobalConfigResponse.
     * @exports GlobalConfigResponse
     * @classdesc Represents a GlobalConfigResponse.
     * @implements IGlobalConfigResponse
     * @constructor
     * @param {IGlobalConfigResponse=} [properties] Properties to set
     */
    function GlobalConfigResponse(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * GlobalConfigResponse enabled.
     * @member {boolean} enabled
     * @memberof GlobalConfigResponse
     * @instance
     */
    GlobalConfigResponse.prototype.enabled = false;

    /**
     * GlobalConfigResponse config.
     * @member {Config|null|undefined} config
     * @memberof GlobalConfigResponse
     * @instance
     */
    GlobalConfigResponse.prototype.config = null;

    /**
     * Encodes the specified GlobalConfigResponse message. Does not implicitly {@link GlobalConfigResponse.verify|verify} messages.
     * @function encode
     * @memberof GlobalConfigResponse
     * @static
     * @param {GlobalConfigResponse} message GlobalConfigResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    GlobalConfigResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.enabled != null && Object.hasOwnProperty.call(message, "enabled"))
            writer.uint32(/* id 1, wireType 0 =*/8).bool(message.enabled);
        if (message.config != null && Object.hasOwnProperty.call(message, "config"))
            $root.Config.encode(message.config, writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim();
        return writer;
    };

    /**
     * Decodes a GlobalConfigResponse message from the specified reader or buffer.
     * @function decode
     * @memberof GlobalConfigResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {GlobalConfigResponse} GlobalConfigResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    GlobalConfigResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.GlobalConfigResponse();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.enabled = reader.bool();
                break;
            case 2:
                message.config = $root.Config.decode(reader, reader.uint32());
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return GlobalConfigResponse;
})();

export const MarsUpdateRequest = $root.MarsUpdateRequest = (() => {

    /**
     * Properties of a MarsUpdateRequest.
     * @exports IMarsUpdateRequest
     * @interface IMarsUpdateRequest
     * @property {number|null} [project_id] MarsUpdateRequest project_id
     * @property {Config|null} [config] MarsUpdateRequest config
     */

    /**
     * Constructs a new MarsUpdateRequest.
     * @exports MarsUpdateRequest
     * @classdesc Represents a MarsUpdateRequest.
     * @implements IMarsUpdateRequest
     * @constructor
     * @param {IMarsUpdateRequest=} [properties] Properties to set
     */
    function MarsUpdateRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * MarsUpdateRequest project_id.
     * @member {number} project_id
     * @memberof MarsUpdateRequest
     * @instance
     */
    MarsUpdateRequest.prototype.project_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * MarsUpdateRequest config.
     * @member {Config|null|undefined} config
     * @memberof MarsUpdateRequest
     * @instance
     */
    MarsUpdateRequest.prototype.config = null;

    /**
     * Encodes the specified MarsUpdateRequest message. Does not implicitly {@link MarsUpdateRequest.verify|verify} messages.
     * @function encode
     * @memberof MarsUpdateRequest
     * @static
     * @param {MarsUpdateRequest} message MarsUpdateRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    MarsUpdateRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.project_id != null && Object.hasOwnProperty.call(message, "project_id"))
            writer.uint32(/* id 1, wireType 0 =*/8).int64(message.project_id);
        if (message.config != null && Object.hasOwnProperty.call(message, "config"))
            $root.Config.encode(message.config, writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim();
        return writer;
    };

    /**
     * Decodes a MarsUpdateRequest message from the specified reader or buffer.
     * @function decode
     * @memberof MarsUpdateRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {MarsUpdateRequest} MarsUpdateRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    MarsUpdateRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.MarsUpdateRequest();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.project_id = reader.int64();
                break;
            case 2:
                message.config = $root.Config.decode(reader, reader.uint32());
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return MarsUpdateRequest;
})();

export const MarsUpdateResponse = $root.MarsUpdateResponse = (() => {

    /**
     * Properties of a MarsUpdateResponse.
     * @exports IMarsUpdateResponse
     * @interface IMarsUpdateResponse
     * @property {Config|null} [config] MarsUpdateResponse config
     */

    /**
     * Constructs a new MarsUpdateResponse.
     * @exports MarsUpdateResponse
     * @classdesc Represents a MarsUpdateResponse.
     * @implements IMarsUpdateResponse
     * @constructor
     * @param {IMarsUpdateResponse=} [properties] Properties to set
     */
    function MarsUpdateResponse(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * MarsUpdateResponse config.
     * @member {Config|null|undefined} config
     * @memberof MarsUpdateResponse
     * @instance
     */
    MarsUpdateResponse.prototype.config = null;

    /**
     * Encodes the specified MarsUpdateResponse message. Does not implicitly {@link MarsUpdateResponse.verify|verify} messages.
     * @function encode
     * @memberof MarsUpdateResponse
     * @static
     * @param {MarsUpdateResponse} message MarsUpdateResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    MarsUpdateResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.config != null && Object.hasOwnProperty.call(message, "config"))
            $root.Config.encode(message.config, writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
        return writer;
    };

    /**
     * Decodes a MarsUpdateResponse message from the specified reader or buffer.
     * @function decode
     * @memberof MarsUpdateResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {MarsUpdateResponse} MarsUpdateResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    MarsUpdateResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.MarsUpdateResponse();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.config = $root.Config.decode(reader, reader.uint32());
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return MarsUpdateResponse;
})();

export const ToggleEnabledRequest = $root.ToggleEnabledRequest = (() => {

    /**
     * Properties of a ToggleEnabledRequest.
     * @exports IToggleEnabledRequest
     * @interface IToggleEnabledRequest
     * @property {number|null} [project_id] ToggleEnabledRequest project_id
     * @property {boolean|null} [enabled] ToggleEnabledRequest enabled
     */

    /**
     * Constructs a new ToggleEnabledRequest.
     * @exports ToggleEnabledRequest
     * @classdesc Represents a ToggleEnabledRequest.
     * @implements IToggleEnabledRequest
     * @constructor
     * @param {IToggleEnabledRequest=} [properties] Properties to set
     */
    function ToggleEnabledRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * ToggleEnabledRequest project_id.
     * @member {number} project_id
     * @memberof ToggleEnabledRequest
     * @instance
     */
    ToggleEnabledRequest.prototype.project_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * ToggleEnabledRequest enabled.
     * @member {boolean} enabled
     * @memberof ToggleEnabledRequest
     * @instance
     */
    ToggleEnabledRequest.prototype.enabled = false;

    /**
     * Encodes the specified ToggleEnabledRequest message. Does not implicitly {@link ToggleEnabledRequest.verify|verify} messages.
     * @function encode
     * @memberof ToggleEnabledRequest
     * @static
     * @param {ToggleEnabledRequest} message ToggleEnabledRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    ToggleEnabledRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.project_id != null && Object.hasOwnProperty.call(message, "project_id"))
            writer.uint32(/* id 1, wireType 0 =*/8).int64(message.project_id);
        if (message.enabled != null && Object.hasOwnProperty.call(message, "enabled"))
            writer.uint32(/* id 2, wireType 0 =*/16).bool(message.enabled);
        return writer;
    };

    /**
     * Decodes a ToggleEnabledRequest message from the specified reader or buffer.
     * @function decode
     * @memberof ToggleEnabledRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {ToggleEnabledRequest} ToggleEnabledRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    ToggleEnabledRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.ToggleEnabledRequest();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.project_id = reader.int64();
                break;
            case 2:
                message.enabled = reader.bool();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return ToggleEnabledRequest;
})();

export const DefaultChartValuesRequest = $root.DefaultChartValuesRequest = (() => {

    /**
     * Properties of a DefaultChartValuesRequest.
     * @exports IDefaultChartValuesRequest
     * @interface IDefaultChartValuesRequest
     * @property {number|null} [project_id] DefaultChartValuesRequest project_id
     * @property {string|null} [branch] DefaultChartValuesRequest branch
     */

    /**
     * Constructs a new DefaultChartValuesRequest.
     * @exports DefaultChartValuesRequest
     * @classdesc Represents a DefaultChartValuesRequest.
     * @implements IDefaultChartValuesRequest
     * @constructor
     * @param {IDefaultChartValuesRequest=} [properties] Properties to set
     */
    function DefaultChartValuesRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * DefaultChartValuesRequest project_id.
     * @member {number} project_id
     * @memberof DefaultChartValuesRequest
     * @instance
     */
    DefaultChartValuesRequest.prototype.project_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * DefaultChartValuesRequest branch.
     * @member {string} branch
     * @memberof DefaultChartValuesRequest
     * @instance
     */
    DefaultChartValuesRequest.prototype.branch = "";

    /**
     * Encodes the specified DefaultChartValuesRequest message. Does not implicitly {@link DefaultChartValuesRequest.verify|verify} messages.
     * @function encode
     * @memberof DefaultChartValuesRequest
     * @static
     * @param {DefaultChartValuesRequest} message DefaultChartValuesRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    DefaultChartValuesRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.project_id != null && Object.hasOwnProperty.call(message, "project_id"))
            writer.uint32(/* id 1, wireType 0 =*/8).int64(message.project_id);
        if (message.branch != null && Object.hasOwnProperty.call(message, "branch"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.branch);
        return writer;
    };

    /**
     * Decodes a DefaultChartValuesRequest message from the specified reader or buffer.
     * @function decode
     * @memberof DefaultChartValuesRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {DefaultChartValuesRequest} DefaultChartValuesRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    DefaultChartValuesRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.DefaultChartValuesRequest();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.project_id = reader.int64();
                break;
            case 2:
                message.branch = reader.string();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return DefaultChartValuesRequest;
})();

export const DefaultChartValues = $root.DefaultChartValues = (() => {

    /**
     * Properties of a DefaultChartValues.
     * @exports IDefaultChartValues
     * @interface IDefaultChartValues
     * @property {string|null} [value] DefaultChartValues value
     */

    /**
     * Constructs a new DefaultChartValues.
     * @exports DefaultChartValues
     * @classdesc Represents a DefaultChartValues.
     * @implements IDefaultChartValues
     * @constructor
     * @param {IDefaultChartValues=} [properties] Properties to set
     */
    function DefaultChartValues(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * DefaultChartValues value.
     * @member {string} value
     * @memberof DefaultChartValues
     * @instance
     */
    DefaultChartValues.prototype.value = "";

    /**
     * Encodes the specified DefaultChartValues message. Does not implicitly {@link DefaultChartValues.verify|verify} messages.
     * @function encode
     * @memberof DefaultChartValues
     * @static
     * @param {DefaultChartValues} message DefaultChartValues message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    DefaultChartValues.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.value != null && Object.hasOwnProperty.call(message, "value"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.value);
        return writer;
    };

    /**
     * Decodes a DefaultChartValues message from the specified reader or buffer.
     * @function decode
     * @memberof DefaultChartValues
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {DefaultChartValues} DefaultChartValues
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    DefaultChartValues.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.DefaultChartValues();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.value = reader.string();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return DefaultChartValues;
})();

export const Mars = $root.Mars = (() => {

    /**
     * Constructs a new Mars service.
     * @exports Mars
     * @classdesc Represents a Mars
     * @extends $protobuf.rpc.Service
     * @constructor
     * @param {$protobuf.RPCImpl} rpcImpl RPC implementation
     * @param {boolean} [requestDelimited=false] Whether requests are length-delimited
     * @param {boolean} [responseDelimited=false] Whether responses are length-delimited
     */
    function Mars(rpcImpl, requestDelimited, responseDelimited) {
        $protobuf.rpc.Service.call(this, rpcImpl, requestDelimited, responseDelimited);
    }

    (Mars.prototype = Object.create($protobuf.rpc.Service.prototype)).constructor = Mars;

    /**
     * Callback as used by {@link Mars#show}.
     * @memberof Mars
     * @typedef ShowCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {MarsShowResponse} [response] MarsShowResponse
     */

    /**
     * Calls Show.
     * @function show
     * @memberof Mars
     * @instance
     * @param {MarsShowRequest} request MarsShowRequest message or plain object
     * @param {Mars.ShowCallback} callback Node-style callback called with the error, if any, and MarsShowResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Mars.prototype.show = function show(request, callback) {
        return this.rpcCall(show, $root.MarsShowRequest, $root.MarsShowResponse, request, callback);
    }, "name", { value: "Show" });

    /**
     * Calls Show.
     * @function show
     * @memberof Mars
     * @instance
     * @param {MarsShowRequest} request MarsShowRequest message or plain object
     * @returns {Promise<MarsShowResponse>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link Mars#globalConfig}.
     * @memberof Mars
     * @typedef GlobalConfigCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {GlobalConfigResponse} [response] GlobalConfigResponse
     */

    /**
     * Calls GlobalConfig.
     * @function globalConfig
     * @memberof Mars
     * @instance
     * @param {GlobalConfigRequest} request GlobalConfigRequest message or plain object
     * @param {Mars.GlobalConfigCallback} callback Node-style callback called with the error, if any, and GlobalConfigResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Mars.prototype.globalConfig = function globalConfig(request, callback) {
        return this.rpcCall(globalConfig, $root.GlobalConfigRequest, $root.GlobalConfigResponse, request, callback);
    }, "name", { value: "GlobalConfig" });

    /**
     * Calls GlobalConfig.
     * @function globalConfig
     * @memberof Mars
     * @instance
     * @param {GlobalConfigRequest} request GlobalConfigRequest message or plain object
     * @returns {Promise<GlobalConfigResponse>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link Mars#toggleEnabled}.
     * @memberof Mars
     * @typedef ToggleEnabledCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {google.protobuf.Empty} [response] Empty
     */

    /**
     * Calls ToggleEnabled.
     * @function toggleEnabled
     * @memberof Mars
     * @instance
     * @param {ToggleEnabledRequest} request ToggleEnabledRequest message or plain object
     * @param {Mars.ToggleEnabledCallback} callback Node-style callback called with the error, if any, and Empty
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Mars.prototype.toggleEnabled = function toggleEnabled(request, callback) {
        return this.rpcCall(toggleEnabled, $root.ToggleEnabledRequest, $root.google.protobuf.Empty, request, callback);
    }, "name", { value: "ToggleEnabled" });

    /**
     * Calls ToggleEnabled.
     * @function toggleEnabled
     * @memberof Mars
     * @instance
     * @param {ToggleEnabledRequest} request ToggleEnabledRequest message or plain object
     * @returns {Promise<google.protobuf.Empty>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link Mars#update}.
     * @memberof Mars
     * @typedef UpdateCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {MarsUpdateResponse} [response] MarsUpdateResponse
     */

    /**
     * Calls Update.
     * @function update
     * @memberof Mars
     * @instance
     * @param {MarsUpdateRequest} request MarsUpdateRequest message or plain object
     * @param {Mars.UpdateCallback} callback Node-style callback called with the error, if any, and MarsUpdateResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Mars.prototype.update = function update(request, callback) {
        return this.rpcCall(update, $root.MarsUpdateRequest, $root.MarsUpdateResponse, request, callback);
    }, "name", { value: "Update" });

    /**
     * Calls Update.
     * @function update
     * @memberof Mars
     * @instance
     * @param {MarsUpdateRequest} request MarsUpdateRequest message or plain object
     * @returns {Promise<MarsUpdateResponse>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link Mars#getDefaultChartValues}.
     * @memberof Mars
     * @typedef GetDefaultChartValuesCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {DefaultChartValues} [response] DefaultChartValues
     */

    /**
     * Calls GetDefaultChartValues.
     * @function getDefaultChartValues
     * @memberof Mars
     * @instance
     * @param {DefaultChartValuesRequest} request DefaultChartValuesRequest message or plain object
     * @param {Mars.GetDefaultChartValuesCallback} callback Node-style callback called with the error, if any, and DefaultChartValues
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Mars.prototype.getDefaultChartValues = function getDefaultChartValues(request, callback) {
        return this.rpcCall(getDefaultChartValues, $root.DefaultChartValuesRequest, $root.DefaultChartValues, request, callback);
    }, "name", { value: "GetDefaultChartValues" });

    /**
     * Calls GetDefaultChartValues.
     * @function getDefaultChartValues
     * @memberof Mars
     * @instance
     * @param {DefaultChartValuesRequest} request DefaultChartValuesRequest message or plain object
     * @returns {Promise<DefaultChartValues>} Promise
     * @variation 2
     */

    return Mars;
})();

export const ProjectByIDRequest = $root.ProjectByIDRequest = (() => {

    /**
     * Properties of a ProjectByIDRequest.
     * @exports IProjectByIDRequest
     * @interface IProjectByIDRequest
     * @property {string|null} [namespace] ProjectByIDRequest namespace
     * @property {string|null} [pod] ProjectByIDRequest pod
     */

    /**
     * Constructs a new ProjectByIDRequest.
     * @exports ProjectByIDRequest
     * @classdesc Represents a ProjectByIDRequest.
     * @implements IProjectByIDRequest
     * @constructor
     * @param {IProjectByIDRequest=} [properties] Properties to set
     */
    function ProjectByIDRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * ProjectByIDRequest namespace.
     * @member {string} namespace
     * @memberof ProjectByIDRequest
     * @instance
     */
    ProjectByIDRequest.prototype.namespace = "";

    /**
     * ProjectByIDRequest pod.
     * @member {string} pod
     * @memberof ProjectByIDRequest
     * @instance
     */
    ProjectByIDRequest.prototype.pod = "";

    /**
     * Encodes the specified ProjectByIDRequest message. Does not implicitly {@link ProjectByIDRequest.verify|verify} messages.
     * @function encode
     * @memberof ProjectByIDRequest
     * @static
     * @param {ProjectByIDRequest} message ProjectByIDRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    ProjectByIDRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.namespace != null && Object.hasOwnProperty.call(message, "namespace"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.namespace);
        if (message.pod != null && Object.hasOwnProperty.call(message, "pod"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.pod);
        return writer;
    };

    /**
     * Decodes a ProjectByIDRequest message from the specified reader or buffer.
     * @function decode
     * @memberof ProjectByIDRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {ProjectByIDRequest} ProjectByIDRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    ProjectByIDRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.ProjectByIDRequest();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.namespace = reader.string();
                break;
            case 2:
                message.pod = reader.string();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return ProjectByIDRequest;
})();

export const ProjectByIDResponse = $root.ProjectByIDResponse = (() => {

    /**
     * Properties of a ProjectByIDResponse.
     * @exports IProjectByIDResponse
     * @interface IProjectByIDResponse
     * @property {number|null} [cpu] ProjectByIDResponse cpu
     * @property {number|null} [memory] ProjectByIDResponse memory
     * @property {string|null} [humanize_cpu] ProjectByIDResponse humanize_cpu
     * @property {string|null} [humanize_memory] ProjectByIDResponse humanize_memory
     * @property {string|null} [time] ProjectByIDResponse time
     * @property {number|null} [length] ProjectByIDResponse length
     */

    /**
     * Constructs a new ProjectByIDResponse.
     * @exports ProjectByIDResponse
     * @classdesc Represents a ProjectByIDResponse.
     * @implements IProjectByIDResponse
     * @constructor
     * @param {IProjectByIDResponse=} [properties] Properties to set
     */
    function ProjectByIDResponse(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * ProjectByIDResponse cpu.
     * @member {number} cpu
     * @memberof ProjectByIDResponse
     * @instance
     */
    ProjectByIDResponse.prototype.cpu = 0;

    /**
     * ProjectByIDResponse memory.
     * @member {number} memory
     * @memberof ProjectByIDResponse
     * @instance
     */
    ProjectByIDResponse.prototype.memory = 0;

    /**
     * ProjectByIDResponse humanize_cpu.
     * @member {string} humanize_cpu
     * @memberof ProjectByIDResponse
     * @instance
     */
    ProjectByIDResponse.prototype.humanize_cpu = "";

    /**
     * ProjectByIDResponse humanize_memory.
     * @member {string} humanize_memory
     * @memberof ProjectByIDResponse
     * @instance
     */
    ProjectByIDResponse.prototype.humanize_memory = "";

    /**
     * ProjectByIDResponse time.
     * @member {string} time
     * @memberof ProjectByIDResponse
     * @instance
     */
    ProjectByIDResponse.prototype.time = "";

    /**
     * ProjectByIDResponse length.
     * @member {number} length
     * @memberof ProjectByIDResponse
     * @instance
     */
    ProjectByIDResponse.prototype.length = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * Encodes the specified ProjectByIDResponse message. Does not implicitly {@link ProjectByIDResponse.verify|verify} messages.
     * @function encode
     * @memberof ProjectByIDResponse
     * @static
     * @param {ProjectByIDResponse} message ProjectByIDResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    ProjectByIDResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.cpu != null && Object.hasOwnProperty.call(message, "cpu"))
            writer.uint32(/* id 1, wireType 1 =*/9).double(message.cpu);
        if (message.memory != null && Object.hasOwnProperty.call(message, "memory"))
            writer.uint32(/* id 2, wireType 1 =*/17).double(message.memory);
        if (message.humanize_cpu != null && Object.hasOwnProperty.call(message, "humanize_cpu"))
            writer.uint32(/* id 3, wireType 2 =*/26).string(message.humanize_cpu);
        if (message.humanize_memory != null && Object.hasOwnProperty.call(message, "humanize_memory"))
            writer.uint32(/* id 4, wireType 2 =*/34).string(message.humanize_memory);
        if (message.time != null && Object.hasOwnProperty.call(message, "time"))
            writer.uint32(/* id 5, wireType 2 =*/42).string(message.time);
        if (message.length != null && Object.hasOwnProperty.call(message, "length"))
            writer.uint32(/* id 6, wireType 0 =*/48).int64(message.length);
        return writer;
    };

    /**
     * Decodes a ProjectByIDResponse message from the specified reader or buffer.
     * @function decode
     * @memberof ProjectByIDResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {ProjectByIDResponse} ProjectByIDResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    ProjectByIDResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.ProjectByIDResponse();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.cpu = reader.double();
                break;
            case 2:
                message.memory = reader.double();
                break;
            case 3:
                message.humanize_cpu = reader.string();
                break;
            case 4:
                message.humanize_memory = reader.string();
                break;
            case 5:
                message.time = reader.string();
                break;
            case 6:
                message.length = reader.int64();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return ProjectByIDResponse;
})();

export const Metrics = $root.Metrics = (() => {

    /**
     * Constructs a new Metrics service.
     * @exports Metrics
     * @classdesc Represents a Metrics
     * @extends $protobuf.rpc.Service
     * @constructor
     * @param {$protobuf.RPCImpl} rpcImpl RPC implementation
     * @param {boolean} [requestDelimited=false] Whether requests are length-delimited
     * @param {boolean} [responseDelimited=false] Whether responses are length-delimited
     */
    function Metrics(rpcImpl, requestDelimited, responseDelimited) {
        $protobuf.rpc.Service.call(this, rpcImpl, requestDelimited, responseDelimited);
    }

    (Metrics.prototype = Object.create($protobuf.rpc.Service.prototype)).constructor = Metrics;

    /**
     * Callback as used by {@link Metrics#projectByID}.
     * @memberof Metrics
     * @typedef ProjectByIDCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {ProjectByIDResponse} [response] ProjectByIDResponse
     */

    /**
     * Calls ProjectByID.
     * @function projectByID
     * @memberof Metrics
     * @instance
     * @param {ProjectByIDRequest} request ProjectByIDRequest message or plain object
     * @param {Metrics.ProjectByIDCallback} callback Node-style callback called with the error, if any, and ProjectByIDResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Metrics.prototype.projectByID = function projectByID(request, callback) {
        return this.rpcCall(projectByID, $root.ProjectByIDRequest, $root.ProjectByIDResponse, request, callback);
    }, "name", { value: "ProjectByID" });

    /**
     * Calls ProjectByID.
     * @function projectByID
     * @memberof Metrics
     * @instance
     * @param {ProjectByIDRequest} request ProjectByIDRequest message or plain object
     * @returns {Promise<ProjectByIDResponse>} Promise
     * @variation 2
     */

    return Metrics;
})();

export const GitlabProjectModal = $root.GitlabProjectModal = (() => {

    /**
     * Properties of a GitlabProjectModal.
     * @exports IGitlabProjectModal
     * @interface IGitlabProjectModal
     * @property {number|null} [id] GitlabProjectModal id
     * @property {string|null} [default_branch] GitlabProjectModal default_branch
     * @property {string|null} [name] GitlabProjectModal name
     * @property {number|null} [gitlab_project_id] GitlabProjectModal gitlab_project_id
     * @property {boolean|null} [enabled] GitlabProjectModal enabled
     * @property {boolean|null} [global_enabled] GitlabProjectModal global_enabled
     * @property {string|null} [global_config] GitlabProjectModal global_config
     * @property {google.protobuf.Timestamp|null} [created_at] GitlabProjectModal created_at
     * @property {google.protobuf.Timestamp|null} [updated_at] GitlabProjectModal updated_at
     * @property {google.protobuf.Timestamp|null} [deleted_at] GitlabProjectModal deleted_at
     */

    /**
     * Constructs a new GitlabProjectModal.
     * @exports GitlabProjectModal
     * @classdesc Represents a GitlabProjectModal.
     * @implements IGitlabProjectModal
     * @constructor
     * @param {IGitlabProjectModal=} [properties] Properties to set
     */
    function GitlabProjectModal(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * GitlabProjectModal id.
     * @member {number} id
     * @memberof GitlabProjectModal
     * @instance
     */
    GitlabProjectModal.prototype.id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * GitlabProjectModal default_branch.
     * @member {string} default_branch
     * @memberof GitlabProjectModal
     * @instance
     */
    GitlabProjectModal.prototype.default_branch = "";

    /**
     * GitlabProjectModal name.
     * @member {string} name
     * @memberof GitlabProjectModal
     * @instance
     */
    GitlabProjectModal.prototype.name = "";

    /**
     * GitlabProjectModal gitlab_project_id.
     * @member {number} gitlab_project_id
     * @memberof GitlabProjectModal
     * @instance
     */
    GitlabProjectModal.prototype.gitlab_project_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * GitlabProjectModal enabled.
     * @member {boolean} enabled
     * @memberof GitlabProjectModal
     * @instance
     */
    GitlabProjectModal.prototype.enabled = false;

    /**
     * GitlabProjectModal global_enabled.
     * @member {boolean} global_enabled
     * @memberof GitlabProjectModal
     * @instance
     */
    GitlabProjectModal.prototype.global_enabled = false;

    /**
     * GitlabProjectModal global_config.
     * @member {string} global_config
     * @memberof GitlabProjectModal
     * @instance
     */
    GitlabProjectModal.prototype.global_config = "";

    /**
     * GitlabProjectModal created_at.
     * @member {google.protobuf.Timestamp|null|undefined} created_at
     * @memberof GitlabProjectModal
     * @instance
     */
    GitlabProjectModal.prototype.created_at = null;

    /**
     * GitlabProjectModal updated_at.
     * @member {google.protobuf.Timestamp|null|undefined} updated_at
     * @memberof GitlabProjectModal
     * @instance
     */
    GitlabProjectModal.prototype.updated_at = null;

    /**
     * GitlabProjectModal deleted_at.
     * @member {google.protobuf.Timestamp|null|undefined} deleted_at
     * @memberof GitlabProjectModal
     * @instance
     */
    GitlabProjectModal.prototype.deleted_at = null;

    /**
     * Encodes the specified GitlabProjectModal message. Does not implicitly {@link GitlabProjectModal.verify|verify} messages.
     * @function encode
     * @memberof GitlabProjectModal
     * @static
     * @param {GitlabProjectModal} message GitlabProjectModal message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    GitlabProjectModal.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.id != null && Object.hasOwnProperty.call(message, "id"))
            writer.uint32(/* id 1, wireType 0 =*/8).int64(message.id);
        if (message.default_branch != null && Object.hasOwnProperty.call(message, "default_branch"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.default_branch);
        if (message.name != null && Object.hasOwnProperty.call(message, "name"))
            writer.uint32(/* id 3, wireType 2 =*/26).string(message.name);
        if (message.gitlab_project_id != null && Object.hasOwnProperty.call(message, "gitlab_project_id"))
            writer.uint32(/* id 4, wireType 0 =*/32).int64(message.gitlab_project_id);
        if (message.enabled != null && Object.hasOwnProperty.call(message, "enabled"))
            writer.uint32(/* id 5, wireType 0 =*/40).bool(message.enabled);
        if (message.global_enabled != null && Object.hasOwnProperty.call(message, "global_enabled"))
            writer.uint32(/* id 6, wireType 0 =*/48).bool(message.global_enabled);
        if (message.global_config != null && Object.hasOwnProperty.call(message, "global_config"))
            writer.uint32(/* id 7, wireType 2 =*/58).string(message.global_config);
        if (message.created_at != null && Object.hasOwnProperty.call(message, "created_at"))
            $root.google.protobuf.Timestamp.encode(message.created_at, writer.uint32(/* id 8, wireType 2 =*/66).fork()).ldelim();
        if (message.updated_at != null && Object.hasOwnProperty.call(message, "updated_at"))
            $root.google.protobuf.Timestamp.encode(message.updated_at, writer.uint32(/* id 9, wireType 2 =*/74).fork()).ldelim();
        if (message.deleted_at != null && Object.hasOwnProperty.call(message, "deleted_at"))
            $root.google.protobuf.Timestamp.encode(message.deleted_at, writer.uint32(/* id 10, wireType 2 =*/82).fork()).ldelim();
        return writer;
    };

    /**
     * Decodes a GitlabProjectModal message from the specified reader or buffer.
     * @function decode
     * @memberof GitlabProjectModal
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {GitlabProjectModal} GitlabProjectModal
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    GitlabProjectModal.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.GitlabProjectModal();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.id = reader.int64();
                break;
            case 2:
                message.default_branch = reader.string();
                break;
            case 3:
                message.name = reader.string();
                break;
            case 4:
                message.gitlab_project_id = reader.int64();
                break;
            case 5:
                message.enabled = reader.bool();
                break;
            case 6:
                message.global_enabled = reader.bool();
                break;
            case 7:
                message.global_config = reader.string();
                break;
            case 8:
                message.created_at = $root.google.protobuf.Timestamp.decode(reader, reader.uint32());
                break;
            case 9:
                message.updated_at = $root.google.protobuf.Timestamp.decode(reader, reader.uint32());
                break;
            case 10:
                message.deleted_at = $root.google.protobuf.Timestamp.decode(reader, reader.uint32());
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return GitlabProjectModal;
})();

export const NamespaceModal = $root.NamespaceModal = (() => {

    /**
     * Properties of a NamespaceModal.
     * @exports INamespaceModal
     * @interface INamespaceModal
     * @property {number|null} [id] NamespaceModal id
     * @property {string|null} [name] NamespaceModal name
     * @property {Array.<string>|null} [image_pull_secrets] NamespaceModal image_pull_secrets
     * @property {google.protobuf.Timestamp|null} [created_at] NamespaceModal created_at
     * @property {google.protobuf.Timestamp|null} [updated_at] NamespaceModal updated_at
     * @property {google.protobuf.Timestamp|null} [deleted_at] NamespaceModal deleted_at
     * @property {Array.<ProjectModal>|null} [projects] NamespaceModal projects
     */

    /**
     * Constructs a new NamespaceModal.
     * @exports NamespaceModal
     * @classdesc Represents a NamespaceModal.
     * @implements INamespaceModal
     * @constructor
     * @param {INamespaceModal=} [properties] Properties to set
     */
    function NamespaceModal(properties) {
        this.image_pull_secrets = [];
        this.projects = [];
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * NamespaceModal id.
     * @member {number} id
     * @memberof NamespaceModal
     * @instance
     */
    NamespaceModal.prototype.id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * NamespaceModal name.
     * @member {string} name
     * @memberof NamespaceModal
     * @instance
     */
    NamespaceModal.prototype.name = "";

    /**
     * NamespaceModal image_pull_secrets.
     * @member {Array.<string>} image_pull_secrets
     * @memberof NamespaceModal
     * @instance
     */
    NamespaceModal.prototype.image_pull_secrets = $util.emptyArray;

    /**
     * NamespaceModal created_at.
     * @member {google.protobuf.Timestamp|null|undefined} created_at
     * @memberof NamespaceModal
     * @instance
     */
    NamespaceModal.prototype.created_at = null;

    /**
     * NamespaceModal updated_at.
     * @member {google.protobuf.Timestamp|null|undefined} updated_at
     * @memberof NamespaceModal
     * @instance
     */
    NamespaceModal.prototype.updated_at = null;

    /**
     * NamespaceModal deleted_at.
     * @member {google.protobuf.Timestamp|null|undefined} deleted_at
     * @memberof NamespaceModal
     * @instance
     */
    NamespaceModal.prototype.deleted_at = null;

    /**
     * NamespaceModal projects.
     * @member {Array.<ProjectModal>} projects
     * @memberof NamespaceModal
     * @instance
     */
    NamespaceModal.prototype.projects = $util.emptyArray;

    /**
     * Encodes the specified NamespaceModal message. Does not implicitly {@link NamespaceModal.verify|verify} messages.
     * @function encode
     * @memberof NamespaceModal
     * @static
     * @param {NamespaceModal} message NamespaceModal message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    NamespaceModal.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.id != null && Object.hasOwnProperty.call(message, "id"))
            writer.uint32(/* id 1, wireType 0 =*/8).int64(message.id);
        if (message.name != null && Object.hasOwnProperty.call(message, "name"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.name);
        if (message.image_pull_secrets != null && message.image_pull_secrets.length)
            for (let i = 0; i < message.image_pull_secrets.length; ++i)
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.image_pull_secrets[i]);
        if (message.created_at != null && Object.hasOwnProperty.call(message, "created_at"))
            $root.google.protobuf.Timestamp.encode(message.created_at, writer.uint32(/* id 4, wireType 2 =*/34).fork()).ldelim();
        if (message.updated_at != null && Object.hasOwnProperty.call(message, "updated_at"))
            $root.google.protobuf.Timestamp.encode(message.updated_at, writer.uint32(/* id 5, wireType 2 =*/42).fork()).ldelim();
        if (message.deleted_at != null && Object.hasOwnProperty.call(message, "deleted_at"))
            $root.google.protobuf.Timestamp.encode(message.deleted_at, writer.uint32(/* id 6, wireType 2 =*/50).fork()).ldelim();
        if (message.projects != null && message.projects.length)
            for (let i = 0; i < message.projects.length; ++i)
                $root.ProjectModal.encode(message.projects[i], writer.uint32(/* id 7, wireType 2 =*/58).fork()).ldelim();
        return writer;
    };

    /**
     * Decodes a NamespaceModal message from the specified reader or buffer.
     * @function decode
     * @memberof NamespaceModal
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {NamespaceModal} NamespaceModal
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    NamespaceModal.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.NamespaceModal();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.id = reader.int64();
                break;
            case 2:
                message.name = reader.string();
                break;
            case 3:
                if (!(message.image_pull_secrets && message.image_pull_secrets.length))
                    message.image_pull_secrets = [];
                message.image_pull_secrets.push(reader.string());
                break;
            case 4:
                message.created_at = $root.google.protobuf.Timestamp.decode(reader, reader.uint32());
                break;
            case 5:
                message.updated_at = $root.google.protobuf.Timestamp.decode(reader, reader.uint32());
                break;
            case 6:
                message.deleted_at = $root.google.protobuf.Timestamp.decode(reader, reader.uint32());
                break;
            case 7:
                if (!(message.projects && message.projects.length))
                    message.projects = [];
                message.projects.push($root.ProjectModal.decode(reader, reader.uint32()));
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return NamespaceModal;
})();

export const ProjectModal = $root.ProjectModal = (() => {

    /**
     * Properties of a ProjectModal.
     * @exports IProjectModal
     * @interface IProjectModal
     * @property {number|null} [id] ProjectModal id
     * @property {string|null} [name] ProjectModal name
     * @property {number|null} [gitlab_project_id] ProjectModal gitlab_project_id
     * @property {string|null} [gitlab_branch] ProjectModal gitlab_branch
     * @property {string|null} [gitlab_commit] ProjectModal gitlab_commit
     * @property {string|null} [config] ProjectModal config
     * @property {string|null} [override_values] ProjectModal override_values
     * @property {string|null} [docker_image] ProjectModal docker_image
     * @property {string|null} [pod_selectors] ProjectModal pod_selectors
     * @property {number|null} [namespace_id] ProjectModal namespace_id
     * @property {boolean|null} [atomic] ProjectModal atomic
     * @property {google.protobuf.Timestamp|null} [created_at] ProjectModal created_at
     * @property {google.protobuf.Timestamp|null} [updated_at] ProjectModal updated_at
     * @property {google.protobuf.Timestamp|null} [deleted_at] ProjectModal deleted_at
     * @property {NamespaceModal|null} [namespace] ProjectModal namespace
     */

    /**
     * Constructs a new ProjectModal.
     * @exports ProjectModal
     * @classdesc Represents a ProjectModal.
     * @implements IProjectModal
     * @constructor
     * @param {IProjectModal=} [properties] Properties to set
     */
    function ProjectModal(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * ProjectModal id.
     * @member {number} id
     * @memberof ProjectModal
     * @instance
     */
    ProjectModal.prototype.id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * ProjectModal name.
     * @member {string} name
     * @memberof ProjectModal
     * @instance
     */
    ProjectModal.prototype.name = "";

    /**
     * ProjectModal gitlab_project_id.
     * @member {number} gitlab_project_id
     * @memberof ProjectModal
     * @instance
     */
    ProjectModal.prototype.gitlab_project_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * ProjectModal gitlab_branch.
     * @member {string} gitlab_branch
     * @memberof ProjectModal
     * @instance
     */
    ProjectModal.prototype.gitlab_branch = "";

    /**
     * ProjectModal gitlab_commit.
     * @member {string} gitlab_commit
     * @memberof ProjectModal
     * @instance
     */
    ProjectModal.prototype.gitlab_commit = "";

    /**
     * ProjectModal config.
     * @member {string} config
     * @memberof ProjectModal
     * @instance
     */
    ProjectModal.prototype.config = "";

    /**
     * ProjectModal override_values.
     * @member {string} override_values
     * @memberof ProjectModal
     * @instance
     */
    ProjectModal.prototype.override_values = "";

    /**
     * ProjectModal docker_image.
     * @member {string} docker_image
     * @memberof ProjectModal
     * @instance
     */
    ProjectModal.prototype.docker_image = "";

    /**
     * ProjectModal pod_selectors.
     * @member {string} pod_selectors
     * @memberof ProjectModal
     * @instance
     */
    ProjectModal.prototype.pod_selectors = "";

    /**
     * ProjectModal namespace_id.
     * @member {number} namespace_id
     * @memberof ProjectModal
     * @instance
     */
    ProjectModal.prototype.namespace_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * ProjectModal atomic.
     * @member {boolean} atomic
     * @memberof ProjectModal
     * @instance
     */
    ProjectModal.prototype.atomic = false;

    /**
     * ProjectModal created_at.
     * @member {google.protobuf.Timestamp|null|undefined} created_at
     * @memberof ProjectModal
     * @instance
     */
    ProjectModal.prototype.created_at = null;

    /**
     * ProjectModal updated_at.
     * @member {google.protobuf.Timestamp|null|undefined} updated_at
     * @memberof ProjectModal
     * @instance
     */
    ProjectModal.prototype.updated_at = null;

    /**
     * ProjectModal deleted_at.
     * @member {google.protobuf.Timestamp|null|undefined} deleted_at
     * @memberof ProjectModal
     * @instance
     */
    ProjectModal.prototype.deleted_at = null;

    /**
     * ProjectModal namespace.
     * @member {NamespaceModal|null|undefined} namespace
     * @memberof ProjectModal
     * @instance
     */
    ProjectModal.prototype.namespace = null;

    /**
     * Encodes the specified ProjectModal message. Does not implicitly {@link ProjectModal.verify|verify} messages.
     * @function encode
     * @memberof ProjectModal
     * @static
     * @param {ProjectModal} message ProjectModal message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    ProjectModal.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.id != null && Object.hasOwnProperty.call(message, "id"))
            writer.uint32(/* id 1, wireType 0 =*/8).int64(message.id);
        if (message.name != null && Object.hasOwnProperty.call(message, "name"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.name);
        if (message.gitlab_project_id != null && Object.hasOwnProperty.call(message, "gitlab_project_id"))
            writer.uint32(/* id 3, wireType 0 =*/24).int64(message.gitlab_project_id);
        if (message.gitlab_branch != null && Object.hasOwnProperty.call(message, "gitlab_branch"))
            writer.uint32(/* id 4, wireType 2 =*/34).string(message.gitlab_branch);
        if (message.gitlab_commit != null && Object.hasOwnProperty.call(message, "gitlab_commit"))
            writer.uint32(/* id 5, wireType 2 =*/42).string(message.gitlab_commit);
        if (message.config != null && Object.hasOwnProperty.call(message, "config"))
            writer.uint32(/* id 6, wireType 2 =*/50).string(message.config);
        if (message.override_values != null && Object.hasOwnProperty.call(message, "override_values"))
            writer.uint32(/* id 7, wireType 2 =*/58).string(message.override_values);
        if (message.docker_image != null && Object.hasOwnProperty.call(message, "docker_image"))
            writer.uint32(/* id 8, wireType 2 =*/66).string(message.docker_image);
        if (message.pod_selectors != null && Object.hasOwnProperty.call(message, "pod_selectors"))
            writer.uint32(/* id 9, wireType 2 =*/74).string(message.pod_selectors);
        if (message.namespace_id != null && Object.hasOwnProperty.call(message, "namespace_id"))
            writer.uint32(/* id 10, wireType 0 =*/80).int64(message.namespace_id);
        if (message.atomic != null && Object.hasOwnProperty.call(message, "atomic"))
            writer.uint32(/* id 11, wireType 0 =*/88).bool(message.atomic);
        if (message.created_at != null && Object.hasOwnProperty.call(message, "created_at"))
            $root.google.protobuf.Timestamp.encode(message.created_at, writer.uint32(/* id 12, wireType 2 =*/98).fork()).ldelim();
        if (message.updated_at != null && Object.hasOwnProperty.call(message, "updated_at"))
            $root.google.protobuf.Timestamp.encode(message.updated_at, writer.uint32(/* id 13, wireType 2 =*/106).fork()).ldelim();
        if (message.deleted_at != null && Object.hasOwnProperty.call(message, "deleted_at"))
            $root.google.protobuf.Timestamp.encode(message.deleted_at, writer.uint32(/* id 14, wireType 2 =*/114).fork()).ldelim();
        if (message.namespace != null && Object.hasOwnProperty.call(message, "namespace"))
            $root.NamespaceModal.encode(message.namespace, writer.uint32(/* id 15, wireType 2 =*/122).fork()).ldelim();
        return writer;
    };

    /**
     * Decodes a ProjectModal message from the specified reader or buffer.
     * @function decode
     * @memberof ProjectModal
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {ProjectModal} ProjectModal
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    ProjectModal.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.ProjectModal();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.id = reader.int64();
                break;
            case 2:
                message.name = reader.string();
                break;
            case 3:
                message.gitlab_project_id = reader.int64();
                break;
            case 4:
                message.gitlab_branch = reader.string();
                break;
            case 5:
                message.gitlab_commit = reader.string();
                break;
            case 6:
                message.config = reader.string();
                break;
            case 7:
                message.override_values = reader.string();
                break;
            case 8:
                message.docker_image = reader.string();
                break;
            case 9:
                message.pod_selectors = reader.string();
                break;
            case 10:
                message.namespace_id = reader.int64();
                break;
            case 11:
                message.atomic = reader.bool();
                break;
            case 12:
                message.created_at = $root.google.protobuf.Timestamp.decode(reader, reader.uint32());
                break;
            case 13:
                message.updated_at = $root.google.protobuf.Timestamp.decode(reader, reader.uint32());
                break;
            case 14:
                message.deleted_at = $root.google.protobuf.Timestamp.decode(reader, reader.uint32());
                break;
            case 15:
                message.namespace = $root.NamespaceModal.decode(reader, reader.uint32());
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return ProjectModal;
})();

export const google = $root.google = (() => {

    /**
     * Namespace google.
     * @exports google
     * @namespace
     */
    const google = {};

    google.protobuf = (function() {

        /**
         * Namespace protobuf.
         * @memberof google
         * @namespace
         */
        const protobuf = {};

        protobuf.Empty = (function() {

            /**
             * Properties of an Empty.
             * @memberof google.protobuf
             * @interface IEmpty
             */

            /**
             * Constructs a new Empty.
             * @memberof google.protobuf
             * @classdesc Represents an Empty.
             * @implements IEmpty
             * @constructor
             * @param {google.protobuf.IEmpty=} [properties] Properties to set
             */
            function Empty(properties) {
                if (properties)
                    for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                        if (properties[keys[i]] != null)
                            this[keys[i]] = properties[keys[i]];
            }

            /**
             * Encodes the specified Empty message. Does not implicitly {@link google.protobuf.Empty.verify|verify} messages.
             * @function encode
             * @memberof google.protobuf.Empty
             * @static
             * @param {google.protobuf.Empty} message Empty message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            Empty.encode = function encode(message, writer) {
                if (!writer)
                    writer = $Writer.create();
                return writer;
            };

            /**
             * Decodes an Empty message from the specified reader or buffer.
             * @function decode
             * @memberof google.protobuf.Empty
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @param {number} [length] Message length if known beforehand
             * @returns {google.protobuf.Empty} Empty
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            Empty.decode = function decode(reader, length) {
                if (!(reader instanceof $Reader))
                    reader = $Reader.create(reader);
                let end = length === undefined ? reader.len : reader.pos + length, message = new $root.google.protobuf.Empty();
                while (reader.pos < end) {
                    let tag = reader.uint32();
                    switch (tag >>> 3) {
                    default:
                        reader.skipType(tag & 7);
                        break;
                    }
                }
                return message;
            };

            return Empty;
        })();

        protobuf.FileDescriptorSet = (function() {

            /**
             * Properties of a FileDescriptorSet.
             * @memberof google.protobuf
             * @interface IFileDescriptorSet
             * @property {Array.<google.protobuf.FileDescriptorProto>|null} [file] FileDescriptorSet file
             */

            /**
             * Constructs a new FileDescriptorSet.
             * @memberof google.protobuf
             * @classdesc Represents a FileDescriptorSet.
             * @implements IFileDescriptorSet
             * @constructor
             * @param {google.protobuf.IFileDescriptorSet=} [properties] Properties to set
             */
            function FileDescriptorSet(properties) {
                this.file = [];
                if (properties)
                    for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                        if (properties[keys[i]] != null)
                            this[keys[i]] = properties[keys[i]];
            }

            /**
             * FileDescriptorSet file.
             * @member {Array.<google.protobuf.FileDescriptorProto>} file
             * @memberof google.protobuf.FileDescriptorSet
             * @instance
             */
            FileDescriptorSet.prototype.file = $util.emptyArray;

            /**
             * Encodes the specified FileDescriptorSet message. Does not implicitly {@link google.protobuf.FileDescriptorSet.verify|verify} messages.
             * @function encode
             * @memberof google.protobuf.FileDescriptorSet
             * @static
             * @param {google.protobuf.FileDescriptorSet} message FileDescriptorSet message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            FileDescriptorSet.encode = function encode(message, writer) {
                if (!writer)
                    writer = $Writer.create();
                if (message.file != null && message.file.length)
                    for (let i = 0; i < message.file.length; ++i)
                        $root.google.protobuf.FileDescriptorProto.encode(message.file[i], writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
                return writer;
            };

            /**
             * Decodes a FileDescriptorSet message from the specified reader or buffer.
             * @function decode
             * @memberof google.protobuf.FileDescriptorSet
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @param {number} [length] Message length if known beforehand
             * @returns {google.protobuf.FileDescriptorSet} FileDescriptorSet
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            FileDescriptorSet.decode = function decode(reader, length) {
                if (!(reader instanceof $Reader))
                    reader = $Reader.create(reader);
                let end = length === undefined ? reader.len : reader.pos + length, message = new $root.google.protobuf.FileDescriptorSet();
                while (reader.pos < end) {
                    let tag = reader.uint32();
                    switch (tag >>> 3) {
                    case 1:
                        if (!(message.file && message.file.length))
                            message.file = [];
                        message.file.push($root.google.protobuf.FileDescriptorProto.decode(reader, reader.uint32()));
                        break;
                    default:
                        reader.skipType(tag & 7);
                        break;
                    }
                }
                return message;
            };

            return FileDescriptorSet;
        })();

        protobuf.FileDescriptorProto = (function() {

            /**
             * Properties of a FileDescriptorProto.
             * @memberof google.protobuf
             * @interface IFileDescriptorProto
             * @property {string|null} [name] FileDescriptorProto name
             * @property {string|null} ["package"] FileDescriptorProto package
             * @property {Array.<string>|null} [dependency] FileDescriptorProto dependency
             * @property {Array.<number>|null} [public_dependency] FileDescriptorProto public_dependency
             * @property {Array.<number>|null} [weak_dependency] FileDescriptorProto weak_dependency
             * @property {Array.<google.protobuf.DescriptorProto>|null} [message_type] FileDescriptorProto message_type
             * @property {Array.<google.protobuf.EnumDescriptorProto>|null} [enum_type] FileDescriptorProto enum_type
             * @property {Array.<google.protobuf.ServiceDescriptorProto>|null} [service] FileDescriptorProto service
             * @property {Array.<google.protobuf.FieldDescriptorProto>|null} [extension] FileDescriptorProto extension
             * @property {google.protobuf.FileOptions|null} [options] FileDescriptorProto options
             * @property {google.protobuf.SourceCodeInfo|null} [source_code_info] FileDescriptorProto source_code_info
             * @property {string|null} [syntax] FileDescriptorProto syntax
             */

            /**
             * Constructs a new FileDescriptorProto.
             * @memberof google.protobuf
             * @classdesc Represents a FileDescriptorProto.
             * @implements IFileDescriptorProto
             * @constructor
             * @param {google.protobuf.IFileDescriptorProto=} [properties] Properties to set
             */
            function FileDescriptorProto(properties) {
                this.dependency = [];
                this.public_dependency = [];
                this.weak_dependency = [];
                this.message_type = [];
                this.enum_type = [];
                this.service = [];
                this.extension = [];
                if (properties)
                    for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                        if (properties[keys[i]] != null)
                            this[keys[i]] = properties[keys[i]];
            }

            /**
             * FileDescriptorProto name.
             * @member {string} name
             * @memberof google.protobuf.FileDescriptorProto
             * @instance
             */
            FileDescriptorProto.prototype.name = "";

            /**
             * FileDescriptorProto package.
             * @member {string} package
             * @memberof google.protobuf.FileDescriptorProto
             * @instance
             */
            FileDescriptorProto.prototype["package"] = "";

            /**
             * FileDescriptorProto dependency.
             * @member {Array.<string>} dependency
             * @memberof google.protobuf.FileDescriptorProto
             * @instance
             */
            FileDescriptorProto.prototype.dependency = $util.emptyArray;

            /**
             * FileDescriptorProto public_dependency.
             * @member {Array.<number>} public_dependency
             * @memberof google.protobuf.FileDescriptorProto
             * @instance
             */
            FileDescriptorProto.prototype.public_dependency = $util.emptyArray;

            /**
             * FileDescriptorProto weak_dependency.
             * @member {Array.<number>} weak_dependency
             * @memberof google.protobuf.FileDescriptorProto
             * @instance
             */
            FileDescriptorProto.prototype.weak_dependency = $util.emptyArray;

            /**
             * FileDescriptorProto message_type.
             * @member {Array.<google.protobuf.DescriptorProto>} message_type
             * @memberof google.protobuf.FileDescriptorProto
             * @instance
             */
            FileDescriptorProto.prototype.message_type = $util.emptyArray;

            /**
             * FileDescriptorProto enum_type.
             * @member {Array.<google.protobuf.EnumDescriptorProto>} enum_type
             * @memberof google.protobuf.FileDescriptorProto
             * @instance
             */
            FileDescriptorProto.prototype.enum_type = $util.emptyArray;

            /**
             * FileDescriptorProto service.
             * @member {Array.<google.protobuf.ServiceDescriptorProto>} service
             * @memberof google.protobuf.FileDescriptorProto
             * @instance
             */
            FileDescriptorProto.prototype.service = $util.emptyArray;

            /**
             * FileDescriptorProto extension.
             * @member {Array.<google.protobuf.FieldDescriptorProto>} extension
             * @memberof google.protobuf.FileDescriptorProto
             * @instance
             */
            FileDescriptorProto.prototype.extension = $util.emptyArray;

            /**
             * FileDescriptorProto options.
             * @member {google.protobuf.FileOptions|null|undefined} options
             * @memberof google.protobuf.FileDescriptorProto
             * @instance
             */
            FileDescriptorProto.prototype.options = null;

            /**
             * FileDescriptorProto source_code_info.
             * @member {google.protobuf.SourceCodeInfo|null|undefined} source_code_info
             * @memberof google.protobuf.FileDescriptorProto
             * @instance
             */
            FileDescriptorProto.prototype.source_code_info = null;

            /**
             * FileDescriptorProto syntax.
             * @member {string} syntax
             * @memberof google.protobuf.FileDescriptorProto
             * @instance
             */
            FileDescriptorProto.prototype.syntax = "";

            /**
             * Encodes the specified FileDescriptorProto message. Does not implicitly {@link google.protobuf.FileDescriptorProto.verify|verify} messages.
             * @function encode
             * @memberof google.protobuf.FileDescriptorProto
             * @static
             * @param {google.protobuf.FileDescriptorProto} message FileDescriptorProto message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            FileDescriptorProto.encode = function encode(message, writer) {
                if (!writer)
                    writer = $Writer.create();
                if (message.name != null && Object.hasOwnProperty.call(message, "name"))
                    writer.uint32(/* id 1, wireType 2 =*/10).string(message.name);
                if (message["package"] != null && Object.hasOwnProperty.call(message, "package"))
                    writer.uint32(/* id 2, wireType 2 =*/18).string(message["package"]);
                if (message.dependency != null && message.dependency.length)
                    for (let i = 0; i < message.dependency.length; ++i)
                        writer.uint32(/* id 3, wireType 2 =*/26).string(message.dependency[i]);
                if (message.message_type != null && message.message_type.length)
                    for (let i = 0; i < message.message_type.length; ++i)
                        $root.google.protobuf.DescriptorProto.encode(message.message_type[i], writer.uint32(/* id 4, wireType 2 =*/34).fork()).ldelim();
                if (message.enum_type != null && message.enum_type.length)
                    for (let i = 0; i < message.enum_type.length; ++i)
                        $root.google.protobuf.EnumDescriptorProto.encode(message.enum_type[i], writer.uint32(/* id 5, wireType 2 =*/42).fork()).ldelim();
                if (message.service != null && message.service.length)
                    for (let i = 0; i < message.service.length; ++i)
                        $root.google.protobuf.ServiceDescriptorProto.encode(message.service[i], writer.uint32(/* id 6, wireType 2 =*/50).fork()).ldelim();
                if (message.extension != null && message.extension.length)
                    for (let i = 0; i < message.extension.length; ++i)
                        $root.google.protobuf.FieldDescriptorProto.encode(message.extension[i], writer.uint32(/* id 7, wireType 2 =*/58).fork()).ldelim();
                if (message.options != null && Object.hasOwnProperty.call(message, "options"))
                    $root.google.protobuf.FileOptions.encode(message.options, writer.uint32(/* id 8, wireType 2 =*/66).fork()).ldelim();
                if (message.source_code_info != null && Object.hasOwnProperty.call(message, "source_code_info"))
                    $root.google.protobuf.SourceCodeInfo.encode(message.source_code_info, writer.uint32(/* id 9, wireType 2 =*/74).fork()).ldelim();
                if (message.public_dependency != null && message.public_dependency.length)
                    for (let i = 0; i < message.public_dependency.length; ++i)
                        writer.uint32(/* id 10, wireType 0 =*/80).int32(message.public_dependency[i]);
                if (message.weak_dependency != null && message.weak_dependency.length)
                    for (let i = 0; i < message.weak_dependency.length; ++i)
                        writer.uint32(/* id 11, wireType 0 =*/88).int32(message.weak_dependency[i]);
                if (message.syntax != null && Object.hasOwnProperty.call(message, "syntax"))
                    writer.uint32(/* id 12, wireType 2 =*/98).string(message.syntax);
                return writer;
            };

            /**
             * Decodes a FileDescriptorProto message from the specified reader or buffer.
             * @function decode
             * @memberof google.protobuf.FileDescriptorProto
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @param {number} [length] Message length if known beforehand
             * @returns {google.protobuf.FileDescriptorProto} FileDescriptorProto
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            FileDescriptorProto.decode = function decode(reader, length) {
                if (!(reader instanceof $Reader))
                    reader = $Reader.create(reader);
                let end = length === undefined ? reader.len : reader.pos + length, message = new $root.google.protobuf.FileDescriptorProto();
                while (reader.pos < end) {
                    let tag = reader.uint32();
                    switch (tag >>> 3) {
                    case 1:
                        message.name = reader.string();
                        break;
                    case 2:
                        message["package"] = reader.string();
                        break;
                    case 3:
                        if (!(message.dependency && message.dependency.length))
                            message.dependency = [];
                        message.dependency.push(reader.string());
                        break;
                    case 10:
                        if (!(message.public_dependency && message.public_dependency.length))
                            message.public_dependency = [];
                        if ((tag & 7) === 2) {
                            let end2 = reader.uint32() + reader.pos;
                            while (reader.pos < end2)
                                message.public_dependency.push(reader.int32());
                        } else
                            message.public_dependency.push(reader.int32());
                        break;
                    case 11:
                        if (!(message.weak_dependency && message.weak_dependency.length))
                            message.weak_dependency = [];
                        if ((tag & 7) === 2) {
                            let end2 = reader.uint32() + reader.pos;
                            while (reader.pos < end2)
                                message.weak_dependency.push(reader.int32());
                        } else
                            message.weak_dependency.push(reader.int32());
                        break;
                    case 4:
                        if (!(message.message_type && message.message_type.length))
                            message.message_type = [];
                        message.message_type.push($root.google.protobuf.DescriptorProto.decode(reader, reader.uint32()));
                        break;
                    case 5:
                        if (!(message.enum_type && message.enum_type.length))
                            message.enum_type = [];
                        message.enum_type.push($root.google.protobuf.EnumDescriptorProto.decode(reader, reader.uint32()));
                        break;
                    case 6:
                        if (!(message.service && message.service.length))
                            message.service = [];
                        message.service.push($root.google.protobuf.ServiceDescriptorProto.decode(reader, reader.uint32()));
                        break;
                    case 7:
                        if (!(message.extension && message.extension.length))
                            message.extension = [];
                        message.extension.push($root.google.protobuf.FieldDescriptorProto.decode(reader, reader.uint32()));
                        break;
                    case 8:
                        message.options = $root.google.protobuf.FileOptions.decode(reader, reader.uint32());
                        break;
                    case 9:
                        message.source_code_info = $root.google.protobuf.SourceCodeInfo.decode(reader, reader.uint32());
                        break;
                    case 12:
                        message.syntax = reader.string();
                        break;
                    default:
                        reader.skipType(tag & 7);
                        break;
                    }
                }
                return message;
            };

            return FileDescriptorProto;
        })();

        protobuf.DescriptorProto = (function() {

            /**
             * Properties of a DescriptorProto.
             * @memberof google.protobuf
             * @interface IDescriptorProto
             * @property {string|null} [name] DescriptorProto name
             * @property {Array.<google.protobuf.FieldDescriptorProto>|null} [field] DescriptorProto field
             * @property {Array.<google.protobuf.FieldDescriptorProto>|null} [extension] DescriptorProto extension
             * @property {Array.<google.protobuf.DescriptorProto>|null} [nested_type] DescriptorProto nested_type
             * @property {Array.<google.protobuf.EnumDescriptorProto>|null} [enum_type] DescriptorProto enum_type
             * @property {Array.<google.protobuf.DescriptorProto.ExtensionRange>|null} [extension_range] DescriptorProto extension_range
             * @property {Array.<google.protobuf.OneofDescriptorProto>|null} [oneof_decl] DescriptorProto oneof_decl
             * @property {google.protobuf.MessageOptions|null} [options] DescriptorProto options
             * @property {Array.<google.protobuf.DescriptorProto.ReservedRange>|null} [reserved_range] DescriptorProto reserved_range
             * @property {Array.<string>|null} [reserved_name] DescriptorProto reserved_name
             */

            /**
             * Constructs a new DescriptorProto.
             * @memberof google.protobuf
             * @classdesc Represents a DescriptorProto.
             * @implements IDescriptorProto
             * @constructor
             * @param {google.protobuf.IDescriptorProto=} [properties] Properties to set
             */
            function DescriptorProto(properties) {
                this.field = [];
                this.extension = [];
                this.nested_type = [];
                this.enum_type = [];
                this.extension_range = [];
                this.oneof_decl = [];
                this.reserved_range = [];
                this.reserved_name = [];
                if (properties)
                    for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                        if (properties[keys[i]] != null)
                            this[keys[i]] = properties[keys[i]];
            }

            /**
             * DescriptorProto name.
             * @member {string} name
             * @memberof google.protobuf.DescriptorProto
             * @instance
             */
            DescriptorProto.prototype.name = "";

            /**
             * DescriptorProto field.
             * @member {Array.<google.protobuf.FieldDescriptorProto>} field
             * @memberof google.protobuf.DescriptorProto
             * @instance
             */
            DescriptorProto.prototype.field = $util.emptyArray;

            /**
             * DescriptorProto extension.
             * @member {Array.<google.protobuf.FieldDescriptorProto>} extension
             * @memberof google.protobuf.DescriptorProto
             * @instance
             */
            DescriptorProto.prototype.extension = $util.emptyArray;

            /**
             * DescriptorProto nested_type.
             * @member {Array.<google.protobuf.DescriptorProto>} nested_type
             * @memberof google.protobuf.DescriptorProto
             * @instance
             */
            DescriptorProto.prototype.nested_type = $util.emptyArray;

            /**
             * DescriptorProto enum_type.
             * @member {Array.<google.protobuf.EnumDescriptorProto>} enum_type
             * @memberof google.protobuf.DescriptorProto
             * @instance
             */
            DescriptorProto.prototype.enum_type = $util.emptyArray;

            /**
             * DescriptorProto extension_range.
             * @member {Array.<google.protobuf.DescriptorProto.ExtensionRange>} extension_range
             * @memberof google.protobuf.DescriptorProto
             * @instance
             */
            DescriptorProto.prototype.extension_range = $util.emptyArray;

            /**
             * DescriptorProto oneof_decl.
             * @member {Array.<google.protobuf.OneofDescriptorProto>} oneof_decl
             * @memberof google.protobuf.DescriptorProto
             * @instance
             */
            DescriptorProto.prototype.oneof_decl = $util.emptyArray;

            /**
             * DescriptorProto options.
             * @member {google.protobuf.MessageOptions|null|undefined} options
             * @memberof google.protobuf.DescriptorProto
             * @instance
             */
            DescriptorProto.prototype.options = null;

            /**
             * DescriptorProto reserved_range.
             * @member {Array.<google.protobuf.DescriptorProto.ReservedRange>} reserved_range
             * @memberof google.protobuf.DescriptorProto
             * @instance
             */
            DescriptorProto.prototype.reserved_range = $util.emptyArray;

            /**
             * DescriptorProto reserved_name.
             * @member {Array.<string>} reserved_name
             * @memberof google.protobuf.DescriptorProto
             * @instance
             */
            DescriptorProto.prototype.reserved_name = $util.emptyArray;

            /**
             * Encodes the specified DescriptorProto message. Does not implicitly {@link google.protobuf.DescriptorProto.verify|verify} messages.
             * @function encode
             * @memberof google.protobuf.DescriptorProto
             * @static
             * @param {google.protobuf.DescriptorProto} message DescriptorProto message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            DescriptorProto.encode = function encode(message, writer) {
                if (!writer)
                    writer = $Writer.create();
                if (message.name != null && Object.hasOwnProperty.call(message, "name"))
                    writer.uint32(/* id 1, wireType 2 =*/10).string(message.name);
                if (message.field != null && message.field.length)
                    for (let i = 0; i < message.field.length; ++i)
                        $root.google.protobuf.FieldDescriptorProto.encode(message.field[i], writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim();
                if (message.nested_type != null && message.nested_type.length)
                    for (let i = 0; i < message.nested_type.length; ++i)
                        $root.google.protobuf.DescriptorProto.encode(message.nested_type[i], writer.uint32(/* id 3, wireType 2 =*/26).fork()).ldelim();
                if (message.enum_type != null && message.enum_type.length)
                    for (let i = 0; i < message.enum_type.length; ++i)
                        $root.google.protobuf.EnumDescriptorProto.encode(message.enum_type[i], writer.uint32(/* id 4, wireType 2 =*/34).fork()).ldelim();
                if (message.extension_range != null && message.extension_range.length)
                    for (let i = 0; i < message.extension_range.length; ++i)
                        $root.google.protobuf.DescriptorProto.ExtensionRange.encode(message.extension_range[i], writer.uint32(/* id 5, wireType 2 =*/42).fork()).ldelim();
                if (message.extension != null && message.extension.length)
                    for (let i = 0; i < message.extension.length; ++i)
                        $root.google.protobuf.FieldDescriptorProto.encode(message.extension[i], writer.uint32(/* id 6, wireType 2 =*/50).fork()).ldelim();
                if (message.options != null && Object.hasOwnProperty.call(message, "options"))
                    $root.google.protobuf.MessageOptions.encode(message.options, writer.uint32(/* id 7, wireType 2 =*/58).fork()).ldelim();
                if (message.oneof_decl != null && message.oneof_decl.length)
                    for (let i = 0; i < message.oneof_decl.length; ++i)
                        $root.google.protobuf.OneofDescriptorProto.encode(message.oneof_decl[i], writer.uint32(/* id 8, wireType 2 =*/66).fork()).ldelim();
                if (message.reserved_range != null && message.reserved_range.length)
                    for (let i = 0; i < message.reserved_range.length; ++i)
                        $root.google.protobuf.DescriptorProto.ReservedRange.encode(message.reserved_range[i], writer.uint32(/* id 9, wireType 2 =*/74).fork()).ldelim();
                if (message.reserved_name != null && message.reserved_name.length)
                    for (let i = 0; i < message.reserved_name.length; ++i)
                        writer.uint32(/* id 10, wireType 2 =*/82).string(message.reserved_name[i]);
                return writer;
            };

            /**
             * Decodes a DescriptorProto message from the specified reader or buffer.
             * @function decode
             * @memberof google.protobuf.DescriptorProto
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @param {number} [length] Message length if known beforehand
             * @returns {google.protobuf.DescriptorProto} DescriptorProto
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            DescriptorProto.decode = function decode(reader, length) {
                if (!(reader instanceof $Reader))
                    reader = $Reader.create(reader);
                let end = length === undefined ? reader.len : reader.pos + length, message = new $root.google.protobuf.DescriptorProto();
                while (reader.pos < end) {
                    let tag = reader.uint32();
                    switch (tag >>> 3) {
                    case 1:
                        message.name = reader.string();
                        break;
                    case 2:
                        if (!(message.field && message.field.length))
                            message.field = [];
                        message.field.push($root.google.protobuf.FieldDescriptorProto.decode(reader, reader.uint32()));
                        break;
                    case 6:
                        if (!(message.extension && message.extension.length))
                            message.extension = [];
                        message.extension.push($root.google.protobuf.FieldDescriptorProto.decode(reader, reader.uint32()));
                        break;
                    case 3:
                        if (!(message.nested_type && message.nested_type.length))
                            message.nested_type = [];
                        message.nested_type.push($root.google.protobuf.DescriptorProto.decode(reader, reader.uint32()));
                        break;
                    case 4:
                        if (!(message.enum_type && message.enum_type.length))
                            message.enum_type = [];
                        message.enum_type.push($root.google.protobuf.EnumDescriptorProto.decode(reader, reader.uint32()));
                        break;
                    case 5:
                        if (!(message.extension_range && message.extension_range.length))
                            message.extension_range = [];
                        message.extension_range.push($root.google.protobuf.DescriptorProto.ExtensionRange.decode(reader, reader.uint32()));
                        break;
                    case 8:
                        if (!(message.oneof_decl && message.oneof_decl.length))
                            message.oneof_decl = [];
                        message.oneof_decl.push($root.google.protobuf.OneofDescriptorProto.decode(reader, reader.uint32()));
                        break;
                    case 7:
                        message.options = $root.google.protobuf.MessageOptions.decode(reader, reader.uint32());
                        break;
                    case 9:
                        if (!(message.reserved_range && message.reserved_range.length))
                            message.reserved_range = [];
                        message.reserved_range.push($root.google.protobuf.DescriptorProto.ReservedRange.decode(reader, reader.uint32()));
                        break;
                    case 10:
                        if (!(message.reserved_name && message.reserved_name.length))
                            message.reserved_name = [];
                        message.reserved_name.push(reader.string());
                        break;
                    default:
                        reader.skipType(tag & 7);
                        break;
                    }
                }
                return message;
            };

            DescriptorProto.ExtensionRange = (function() {

                /**
                 * Properties of an ExtensionRange.
                 * @memberof google.protobuf.DescriptorProto
                 * @interface IExtensionRange
                 * @property {number|null} [start] ExtensionRange start
                 * @property {number|null} [end] ExtensionRange end
                 */

                /**
                 * Constructs a new ExtensionRange.
                 * @memberof google.protobuf.DescriptorProto
                 * @classdesc Represents an ExtensionRange.
                 * @implements IExtensionRange
                 * @constructor
                 * @param {google.protobuf.DescriptorProto.IExtensionRange=} [properties] Properties to set
                 */
                function ExtensionRange(properties) {
                    if (properties)
                        for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                            if (properties[keys[i]] != null)
                                this[keys[i]] = properties[keys[i]];
                }

                /**
                 * ExtensionRange start.
                 * @member {number} start
                 * @memberof google.protobuf.DescriptorProto.ExtensionRange
                 * @instance
                 */
                ExtensionRange.prototype.start = 0;

                /**
                 * ExtensionRange end.
                 * @member {number} end
                 * @memberof google.protobuf.DescriptorProto.ExtensionRange
                 * @instance
                 */
                ExtensionRange.prototype.end = 0;

                /**
                 * Encodes the specified ExtensionRange message. Does not implicitly {@link google.protobuf.DescriptorProto.ExtensionRange.verify|verify} messages.
                 * @function encode
                 * @memberof google.protobuf.DescriptorProto.ExtensionRange
                 * @static
                 * @param {google.protobuf.DescriptorProto.ExtensionRange} message ExtensionRange message or plain object to encode
                 * @param {$protobuf.Writer} [writer] Writer to encode to
                 * @returns {$protobuf.Writer} Writer
                 */
                ExtensionRange.encode = function encode(message, writer) {
                    if (!writer)
                        writer = $Writer.create();
                    if (message.start != null && Object.hasOwnProperty.call(message, "start"))
                        writer.uint32(/* id 1, wireType 0 =*/8).int32(message.start);
                    if (message.end != null && Object.hasOwnProperty.call(message, "end"))
                        writer.uint32(/* id 2, wireType 0 =*/16).int32(message.end);
                    return writer;
                };

                /**
                 * Decodes an ExtensionRange message from the specified reader or buffer.
                 * @function decode
                 * @memberof google.protobuf.DescriptorProto.ExtensionRange
                 * @static
                 * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
                 * @param {number} [length] Message length if known beforehand
                 * @returns {google.protobuf.DescriptorProto.ExtensionRange} ExtensionRange
                 * @throws {Error} If the payload is not a reader or valid buffer
                 * @throws {$protobuf.util.ProtocolError} If required fields are missing
                 */
                ExtensionRange.decode = function decode(reader, length) {
                    if (!(reader instanceof $Reader))
                        reader = $Reader.create(reader);
                    let end = length === undefined ? reader.len : reader.pos + length, message = new $root.google.protobuf.DescriptorProto.ExtensionRange();
                    while (reader.pos < end) {
                        let tag = reader.uint32();
                        switch (tag >>> 3) {
                        case 1:
                            message.start = reader.int32();
                            break;
                        case 2:
                            message.end = reader.int32();
                            break;
                        default:
                            reader.skipType(tag & 7);
                            break;
                        }
                    }
                    return message;
                };

                return ExtensionRange;
            })();

            DescriptorProto.ReservedRange = (function() {

                /**
                 * Properties of a ReservedRange.
                 * @memberof google.protobuf.DescriptorProto
                 * @interface IReservedRange
                 * @property {number|null} [start] ReservedRange start
                 * @property {number|null} [end] ReservedRange end
                 */

                /**
                 * Constructs a new ReservedRange.
                 * @memberof google.protobuf.DescriptorProto
                 * @classdesc Represents a ReservedRange.
                 * @implements IReservedRange
                 * @constructor
                 * @param {google.protobuf.DescriptorProto.IReservedRange=} [properties] Properties to set
                 */
                function ReservedRange(properties) {
                    if (properties)
                        for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                            if (properties[keys[i]] != null)
                                this[keys[i]] = properties[keys[i]];
                }

                /**
                 * ReservedRange start.
                 * @member {number} start
                 * @memberof google.protobuf.DescriptorProto.ReservedRange
                 * @instance
                 */
                ReservedRange.prototype.start = 0;

                /**
                 * ReservedRange end.
                 * @member {number} end
                 * @memberof google.protobuf.DescriptorProto.ReservedRange
                 * @instance
                 */
                ReservedRange.prototype.end = 0;

                /**
                 * Encodes the specified ReservedRange message. Does not implicitly {@link google.protobuf.DescriptorProto.ReservedRange.verify|verify} messages.
                 * @function encode
                 * @memberof google.protobuf.DescriptorProto.ReservedRange
                 * @static
                 * @param {google.protobuf.DescriptorProto.ReservedRange} message ReservedRange message or plain object to encode
                 * @param {$protobuf.Writer} [writer] Writer to encode to
                 * @returns {$protobuf.Writer} Writer
                 */
                ReservedRange.encode = function encode(message, writer) {
                    if (!writer)
                        writer = $Writer.create();
                    if (message.start != null && Object.hasOwnProperty.call(message, "start"))
                        writer.uint32(/* id 1, wireType 0 =*/8).int32(message.start);
                    if (message.end != null && Object.hasOwnProperty.call(message, "end"))
                        writer.uint32(/* id 2, wireType 0 =*/16).int32(message.end);
                    return writer;
                };

                /**
                 * Decodes a ReservedRange message from the specified reader or buffer.
                 * @function decode
                 * @memberof google.protobuf.DescriptorProto.ReservedRange
                 * @static
                 * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
                 * @param {number} [length] Message length if known beforehand
                 * @returns {google.protobuf.DescriptorProto.ReservedRange} ReservedRange
                 * @throws {Error} If the payload is not a reader or valid buffer
                 * @throws {$protobuf.util.ProtocolError} If required fields are missing
                 */
                ReservedRange.decode = function decode(reader, length) {
                    if (!(reader instanceof $Reader))
                        reader = $Reader.create(reader);
                    let end = length === undefined ? reader.len : reader.pos + length, message = new $root.google.protobuf.DescriptorProto.ReservedRange();
                    while (reader.pos < end) {
                        let tag = reader.uint32();
                        switch (tag >>> 3) {
                        case 1:
                            message.start = reader.int32();
                            break;
                        case 2:
                            message.end = reader.int32();
                            break;
                        default:
                            reader.skipType(tag & 7);
                            break;
                        }
                    }
                    return message;
                };

                return ReservedRange;
            })();

            return DescriptorProto;
        })();

        protobuf.FieldDescriptorProto = (function() {

            /**
             * Properties of a FieldDescriptorProto.
             * @memberof google.protobuf
             * @interface IFieldDescriptorProto
             * @property {string|null} [name] FieldDescriptorProto name
             * @property {number|null} [number] FieldDescriptorProto number
             * @property {google.protobuf.FieldDescriptorProto.Label|null} [label] FieldDescriptorProto label
             * @property {google.protobuf.FieldDescriptorProto.Type|null} [type] FieldDescriptorProto type
             * @property {string|null} [type_name] FieldDescriptorProto type_name
             * @property {string|null} [extendee] FieldDescriptorProto extendee
             * @property {string|null} [default_value] FieldDescriptorProto default_value
             * @property {number|null} [oneof_index] FieldDescriptorProto oneof_index
             * @property {string|null} [json_name] FieldDescriptorProto json_name
             * @property {google.protobuf.FieldOptions|null} [options] FieldDescriptorProto options
             */

            /**
             * Constructs a new FieldDescriptorProto.
             * @memberof google.protobuf
             * @classdesc Represents a FieldDescriptorProto.
             * @implements IFieldDescriptorProto
             * @constructor
             * @param {google.protobuf.IFieldDescriptorProto=} [properties] Properties to set
             */
            function FieldDescriptorProto(properties) {
                if (properties)
                    for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                        if (properties[keys[i]] != null)
                            this[keys[i]] = properties[keys[i]];
            }

            /**
             * FieldDescriptorProto name.
             * @member {string} name
             * @memberof google.protobuf.FieldDescriptorProto
             * @instance
             */
            FieldDescriptorProto.prototype.name = "";

            /**
             * FieldDescriptorProto number.
             * @member {number} number
             * @memberof google.protobuf.FieldDescriptorProto
             * @instance
             */
            FieldDescriptorProto.prototype.number = 0;

            /**
             * FieldDescriptorProto label.
             * @member {google.protobuf.FieldDescriptorProto.Label} label
             * @memberof google.protobuf.FieldDescriptorProto
             * @instance
             */
            FieldDescriptorProto.prototype.label = 1;

            /**
             * FieldDescriptorProto type.
             * @member {google.protobuf.FieldDescriptorProto.Type} type
             * @memberof google.protobuf.FieldDescriptorProto
             * @instance
             */
            FieldDescriptorProto.prototype.type = 1;

            /**
             * FieldDescriptorProto type_name.
             * @member {string} type_name
             * @memberof google.protobuf.FieldDescriptorProto
             * @instance
             */
            FieldDescriptorProto.prototype.type_name = "";

            /**
             * FieldDescriptorProto extendee.
             * @member {string} extendee
             * @memberof google.protobuf.FieldDescriptorProto
             * @instance
             */
            FieldDescriptorProto.prototype.extendee = "";

            /**
             * FieldDescriptorProto default_value.
             * @member {string} default_value
             * @memberof google.protobuf.FieldDescriptorProto
             * @instance
             */
            FieldDescriptorProto.prototype.default_value = "";

            /**
             * FieldDescriptorProto oneof_index.
             * @member {number} oneof_index
             * @memberof google.protobuf.FieldDescriptorProto
             * @instance
             */
            FieldDescriptorProto.prototype.oneof_index = 0;

            /**
             * FieldDescriptorProto json_name.
             * @member {string} json_name
             * @memberof google.protobuf.FieldDescriptorProto
             * @instance
             */
            FieldDescriptorProto.prototype.json_name = "";

            /**
             * FieldDescriptorProto options.
             * @member {google.protobuf.FieldOptions|null|undefined} options
             * @memberof google.protobuf.FieldDescriptorProto
             * @instance
             */
            FieldDescriptorProto.prototype.options = null;

            /**
             * Encodes the specified FieldDescriptorProto message. Does not implicitly {@link google.protobuf.FieldDescriptorProto.verify|verify} messages.
             * @function encode
             * @memberof google.protobuf.FieldDescriptorProto
             * @static
             * @param {google.protobuf.FieldDescriptorProto} message FieldDescriptorProto message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            FieldDescriptorProto.encode = function encode(message, writer) {
                if (!writer)
                    writer = $Writer.create();
                if (message.name != null && Object.hasOwnProperty.call(message, "name"))
                    writer.uint32(/* id 1, wireType 2 =*/10).string(message.name);
                if (message.extendee != null && Object.hasOwnProperty.call(message, "extendee"))
                    writer.uint32(/* id 2, wireType 2 =*/18).string(message.extendee);
                if (message.number != null && Object.hasOwnProperty.call(message, "number"))
                    writer.uint32(/* id 3, wireType 0 =*/24).int32(message.number);
                if (message.label != null && Object.hasOwnProperty.call(message, "label"))
                    writer.uint32(/* id 4, wireType 0 =*/32).int32(message.label);
                if (message.type != null && Object.hasOwnProperty.call(message, "type"))
                    writer.uint32(/* id 5, wireType 0 =*/40).int32(message.type);
                if (message.type_name != null && Object.hasOwnProperty.call(message, "type_name"))
                    writer.uint32(/* id 6, wireType 2 =*/50).string(message.type_name);
                if (message.default_value != null && Object.hasOwnProperty.call(message, "default_value"))
                    writer.uint32(/* id 7, wireType 2 =*/58).string(message.default_value);
                if (message.options != null && Object.hasOwnProperty.call(message, "options"))
                    $root.google.protobuf.FieldOptions.encode(message.options, writer.uint32(/* id 8, wireType 2 =*/66).fork()).ldelim();
                if (message.oneof_index != null && Object.hasOwnProperty.call(message, "oneof_index"))
                    writer.uint32(/* id 9, wireType 0 =*/72).int32(message.oneof_index);
                if (message.json_name != null && Object.hasOwnProperty.call(message, "json_name"))
                    writer.uint32(/* id 10, wireType 2 =*/82).string(message.json_name);
                return writer;
            };

            /**
             * Decodes a FieldDescriptorProto message from the specified reader or buffer.
             * @function decode
             * @memberof google.protobuf.FieldDescriptorProto
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @param {number} [length] Message length if known beforehand
             * @returns {google.protobuf.FieldDescriptorProto} FieldDescriptorProto
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            FieldDescriptorProto.decode = function decode(reader, length) {
                if (!(reader instanceof $Reader))
                    reader = $Reader.create(reader);
                let end = length === undefined ? reader.len : reader.pos + length, message = new $root.google.protobuf.FieldDescriptorProto();
                while (reader.pos < end) {
                    let tag = reader.uint32();
                    switch (tag >>> 3) {
                    case 1:
                        message.name = reader.string();
                        break;
                    case 3:
                        message.number = reader.int32();
                        break;
                    case 4:
                        message.label = reader.int32();
                        break;
                    case 5:
                        message.type = reader.int32();
                        break;
                    case 6:
                        message.type_name = reader.string();
                        break;
                    case 2:
                        message.extendee = reader.string();
                        break;
                    case 7:
                        message.default_value = reader.string();
                        break;
                    case 9:
                        message.oneof_index = reader.int32();
                        break;
                    case 10:
                        message.json_name = reader.string();
                        break;
                    case 8:
                        message.options = $root.google.protobuf.FieldOptions.decode(reader, reader.uint32());
                        break;
                    default:
                        reader.skipType(tag & 7);
                        break;
                    }
                }
                return message;
            };

            /**
             * Type enum.
             * @name google.protobuf.FieldDescriptorProto.Type
             * @enum {number}
             * @property {number} TYPE_DOUBLE=1 TYPE_DOUBLE value
             * @property {number} TYPE_FLOAT=2 TYPE_FLOAT value
             * @property {number} TYPE_INT64=3 TYPE_INT64 value
             * @property {number} TYPE_UINT64=4 TYPE_UINT64 value
             * @property {number} TYPE_INT32=5 TYPE_INT32 value
             * @property {number} TYPE_FIXED64=6 TYPE_FIXED64 value
             * @property {number} TYPE_FIXED32=7 TYPE_FIXED32 value
             * @property {number} TYPE_BOOL=8 TYPE_BOOL value
             * @property {number} TYPE_STRING=9 TYPE_STRING value
             * @property {number} TYPE_GROUP=10 TYPE_GROUP value
             * @property {number} TYPE_MESSAGE=11 TYPE_MESSAGE value
             * @property {number} TYPE_BYTES=12 TYPE_BYTES value
             * @property {number} TYPE_UINT32=13 TYPE_UINT32 value
             * @property {number} TYPE_ENUM=14 TYPE_ENUM value
             * @property {number} TYPE_SFIXED32=15 TYPE_SFIXED32 value
             * @property {number} TYPE_SFIXED64=16 TYPE_SFIXED64 value
             * @property {number} TYPE_SINT32=17 TYPE_SINT32 value
             * @property {number} TYPE_SINT64=18 TYPE_SINT64 value
             */
            FieldDescriptorProto.Type = (function() {
                const valuesById = {}, values = Object.create(valuesById);
                values[valuesById[1] = "TYPE_DOUBLE"] = 1;
                values[valuesById[2] = "TYPE_FLOAT"] = 2;
                values[valuesById[3] = "TYPE_INT64"] = 3;
                values[valuesById[4] = "TYPE_UINT64"] = 4;
                values[valuesById[5] = "TYPE_INT32"] = 5;
                values[valuesById[6] = "TYPE_FIXED64"] = 6;
                values[valuesById[7] = "TYPE_FIXED32"] = 7;
                values[valuesById[8] = "TYPE_BOOL"] = 8;
                values[valuesById[9] = "TYPE_STRING"] = 9;
                values[valuesById[10] = "TYPE_GROUP"] = 10;
                values[valuesById[11] = "TYPE_MESSAGE"] = 11;
                values[valuesById[12] = "TYPE_BYTES"] = 12;
                values[valuesById[13] = "TYPE_UINT32"] = 13;
                values[valuesById[14] = "TYPE_ENUM"] = 14;
                values[valuesById[15] = "TYPE_SFIXED32"] = 15;
                values[valuesById[16] = "TYPE_SFIXED64"] = 16;
                values[valuesById[17] = "TYPE_SINT32"] = 17;
                values[valuesById[18] = "TYPE_SINT64"] = 18;
                return values;
            })();

            /**
             * Label enum.
             * @name google.protobuf.FieldDescriptorProto.Label
             * @enum {number}
             * @property {number} LABEL_OPTIONAL=1 LABEL_OPTIONAL value
             * @property {number} LABEL_REQUIRED=2 LABEL_REQUIRED value
             * @property {number} LABEL_REPEATED=3 LABEL_REPEATED value
             */
            FieldDescriptorProto.Label = (function() {
                const valuesById = {}, values = Object.create(valuesById);
                values[valuesById[1] = "LABEL_OPTIONAL"] = 1;
                values[valuesById[2] = "LABEL_REQUIRED"] = 2;
                values[valuesById[3] = "LABEL_REPEATED"] = 3;
                return values;
            })();

            return FieldDescriptorProto;
        })();

        protobuf.OneofDescriptorProto = (function() {

            /**
             * Properties of an OneofDescriptorProto.
             * @memberof google.protobuf
             * @interface IOneofDescriptorProto
             * @property {string|null} [name] OneofDescriptorProto name
             * @property {google.protobuf.OneofOptions|null} [options] OneofDescriptorProto options
             */

            /**
             * Constructs a new OneofDescriptorProto.
             * @memberof google.protobuf
             * @classdesc Represents an OneofDescriptorProto.
             * @implements IOneofDescriptorProto
             * @constructor
             * @param {google.protobuf.IOneofDescriptorProto=} [properties] Properties to set
             */
            function OneofDescriptorProto(properties) {
                if (properties)
                    for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                        if (properties[keys[i]] != null)
                            this[keys[i]] = properties[keys[i]];
            }

            /**
             * OneofDescriptorProto name.
             * @member {string} name
             * @memberof google.protobuf.OneofDescriptorProto
             * @instance
             */
            OneofDescriptorProto.prototype.name = "";

            /**
             * OneofDescriptorProto options.
             * @member {google.protobuf.OneofOptions|null|undefined} options
             * @memberof google.protobuf.OneofDescriptorProto
             * @instance
             */
            OneofDescriptorProto.prototype.options = null;

            /**
             * Encodes the specified OneofDescriptorProto message. Does not implicitly {@link google.protobuf.OneofDescriptorProto.verify|verify} messages.
             * @function encode
             * @memberof google.protobuf.OneofDescriptorProto
             * @static
             * @param {google.protobuf.OneofDescriptorProto} message OneofDescriptorProto message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            OneofDescriptorProto.encode = function encode(message, writer) {
                if (!writer)
                    writer = $Writer.create();
                if (message.name != null && Object.hasOwnProperty.call(message, "name"))
                    writer.uint32(/* id 1, wireType 2 =*/10).string(message.name);
                if (message.options != null && Object.hasOwnProperty.call(message, "options"))
                    $root.google.protobuf.OneofOptions.encode(message.options, writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim();
                return writer;
            };

            /**
             * Decodes an OneofDescriptorProto message from the specified reader or buffer.
             * @function decode
             * @memberof google.protobuf.OneofDescriptorProto
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @param {number} [length] Message length if known beforehand
             * @returns {google.protobuf.OneofDescriptorProto} OneofDescriptorProto
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            OneofDescriptorProto.decode = function decode(reader, length) {
                if (!(reader instanceof $Reader))
                    reader = $Reader.create(reader);
                let end = length === undefined ? reader.len : reader.pos + length, message = new $root.google.protobuf.OneofDescriptorProto();
                while (reader.pos < end) {
                    let tag = reader.uint32();
                    switch (tag >>> 3) {
                    case 1:
                        message.name = reader.string();
                        break;
                    case 2:
                        message.options = $root.google.protobuf.OneofOptions.decode(reader, reader.uint32());
                        break;
                    default:
                        reader.skipType(tag & 7);
                        break;
                    }
                }
                return message;
            };

            return OneofDescriptorProto;
        })();

        protobuf.EnumDescriptorProto = (function() {

            /**
             * Properties of an EnumDescriptorProto.
             * @memberof google.protobuf
             * @interface IEnumDescriptorProto
             * @property {string|null} [name] EnumDescriptorProto name
             * @property {Array.<google.protobuf.EnumValueDescriptorProto>|null} [value] EnumDescriptorProto value
             * @property {google.protobuf.EnumOptions|null} [options] EnumDescriptorProto options
             */

            /**
             * Constructs a new EnumDescriptorProto.
             * @memberof google.protobuf
             * @classdesc Represents an EnumDescriptorProto.
             * @implements IEnumDescriptorProto
             * @constructor
             * @param {google.protobuf.IEnumDescriptorProto=} [properties] Properties to set
             */
            function EnumDescriptorProto(properties) {
                this.value = [];
                if (properties)
                    for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                        if (properties[keys[i]] != null)
                            this[keys[i]] = properties[keys[i]];
            }

            /**
             * EnumDescriptorProto name.
             * @member {string} name
             * @memberof google.protobuf.EnumDescriptorProto
             * @instance
             */
            EnumDescriptorProto.prototype.name = "";

            /**
             * EnumDescriptorProto value.
             * @member {Array.<google.protobuf.EnumValueDescriptorProto>} value
             * @memberof google.protobuf.EnumDescriptorProto
             * @instance
             */
            EnumDescriptorProto.prototype.value = $util.emptyArray;

            /**
             * EnumDescriptorProto options.
             * @member {google.protobuf.EnumOptions|null|undefined} options
             * @memberof google.protobuf.EnumDescriptorProto
             * @instance
             */
            EnumDescriptorProto.prototype.options = null;

            /**
             * Encodes the specified EnumDescriptorProto message. Does not implicitly {@link google.protobuf.EnumDescriptorProto.verify|verify} messages.
             * @function encode
             * @memberof google.protobuf.EnumDescriptorProto
             * @static
             * @param {google.protobuf.EnumDescriptorProto} message EnumDescriptorProto message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            EnumDescriptorProto.encode = function encode(message, writer) {
                if (!writer)
                    writer = $Writer.create();
                if (message.name != null && Object.hasOwnProperty.call(message, "name"))
                    writer.uint32(/* id 1, wireType 2 =*/10).string(message.name);
                if (message.value != null && message.value.length)
                    for (let i = 0; i < message.value.length; ++i)
                        $root.google.protobuf.EnumValueDescriptorProto.encode(message.value[i], writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim();
                if (message.options != null && Object.hasOwnProperty.call(message, "options"))
                    $root.google.protobuf.EnumOptions.encode(message.options, writer.uint32(/* id 3, wireType 2 =*/26).fork()).ldelim();
                return writer;
            };

            /**
             * Decodes an EnumDescriptorProto message from the specified reader or buffer.
             * @function decode
             * @memberof google.protobuf.EnumDescriptorProto
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @param {number} [length] Message length if known beforehand
             * @returns {google.protobuf.EnumDescriptorProto} EnumDescriptorProto
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            EnumDescriptorProto.decode = function decode(reader, length) {
                if (!(reader instanceof $Reader))
                    reader = $Reader.create(reader);
                let end = length === undefined ? reader.len : reader.pos + length, message = new $root.google.protobuf.EnumDescriptorProto();
                while (reader.pos < end) {
                    let tag = reader.uint32();
                    switch (tag >>> 3) {
                    case 1:
                        message.name = reader.string();
                        break;
                    case 2:
                        if (!(message.value && message.value.length))
                            message.value = [];
                        message.value.push($root.google.protobuf.EnumValueDescriptorProto.decode(reader, reader.uint32()));
                        break;
                    case 3:
                        message.options = $root.google.protobuf.EnumOptions.decode(reader, reader.uint32());
                        break;
                    default:
                        reader.skipType(tag & 7);
                        break;
                    }
                }
                return message;
            };

            return EnumDescriptorProto;
        })();

        protobuf.EnumValueDescriptorProto = (function() {

            /**
             * Properties of an EnumValueDescriptorProto.
             * @memberof google.protobuf
             * @interface IEnumValueDescriptorProto
             * @property {string|null} [name] EnumValueDescriptorProto name
             * @property {number|null} [number] EnumValueDescriptorProto number
             * @property {google.protobuf.EnumValueOptions|null} [options] EnumValueDescriptorProto options
             */

            /**
             * Constructs a new EnumValueDescriptorProto.
             * @memberof google.protobuf
             * @classdesc Represents an EnumValueDescriptorProto.
             * @implements IEnumValueDescriptorProto
             * @constructor
             * @param {google.protobuf.IEnumValueDescriptorProto=} [properties] Properties to set
             */
            function EnumValueDescriptorProto(properties) {
                if (properties)
                    for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                        if (properties[keys[i]] != null)
                            this[keys[i]] = properties[keys[i]];
            }

            /**
             * EnumValueDescriptorProto name.
             * @member {string} name
             * @memberof google.protobuf.EnumValueDescriptorProto
             * @instance
             */
            EnumValueDescriptorProto.prototype.name = "";

            /**
             * EnumValueDescriptorProto number.
             * @member {number} number
             * @memberof google.protobuf.EnumValueDescriptorProto
             * @instance
             */
            EnumValueDescriptorProto.prototype.number = 0;

            /**
             * EnumValueDescriptorProto options.
             * @member {google.protobuf.EnumValueOptions|null|undefined} options
             * @memberof google.protobuf.EnumValueDescriptorProto
             * @instance
             */
            EnumValueDescriptorProto.prototype.options = null;

            /**
             * Encodes the specified EnumValueDescriptorProto message. Does not implicitly {@link google.protobuf.EnumValueDescriptorProto.verify|verify} messages.
             * @function encode
             * @memberof google.protobuf.EnumValueDescriptorProto
             * @static
             * @param {google.protobuf.EnumValueDescriptorProto} message EnumValueDescriptorProto message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            EnumValueDescriptorProto.encode = function encode(message, writer) {
                if (!writer)
                    writer = $Writer.create();
                if (message.name != null && Object.hasOwnProperty.call(message, "name"))
                    writer.uint32(/* id 1, wireType 2 =*/10).string(message.name);
                if (message.number != null && Object.hasOwnProperty.call(message, "number"))
                    writer.uint32(/* id 2, wireType 0 =*/16).int32(message.number);
                if (message.options != null && Object.hasOwnProperty.call(message, "options"))
                    $root.google.protobuf.EnumValueOptions.encode(message.options, writer.uint32(/* id 3, wireType 2 =*/26).fork()).ldelim();
                return writer;
            };

            /**
             * Decodes an EnumValueDescriptorProto message from the specified reader or buffer.
             * @function decode
             * @memberof google.protobuf.EnumValueDescriptorProto
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @param {number} [length] Message length if known beforehand
             * @returns {google.protobuf.EnumValueDescriptorProto} EnumValueDescriptorProto
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            EnumValueDescriptorProto.decode = function decode(reader, length) {
                if (!(reader instanceof $Reader))
                    reader = $Reader.create(reader);
                let end = length === undefined ? reader.len : reader.pos + length, message = new $root.google.protobuf.EnumValueDescriptorProto();
                while (reader.pos < end) {
                    let tag = reader.uint32();
                    switch (tag >>> 3) {
                    case 1:
                        message.name = reader.string();
                        break;
                    case 2:
                        message.number = reader.int32();
                        break;
                    case 3:
                        message.options = $root.google.protobuf.EnumValueOptions.decode(reader, reader.uint32());
                        break;
                    default:
                        reader.skipType(tag & 7);
                        break;
                    }
                }
                return message;
            };

            return EnumValueDescriptorProto;
        })();

        protobuf.ServiceDescriptorProto = (function() {

            /**
             * Properties of a ServiceDescriptorProto.
             * @memberof google.protobuf
             * @interface IServiceDescriptorProto
             * @property {string|null} [name] ServiceDescriptorProto name
             * @property {Array.<google.protobuf.MethodDescriptorProto>|null} [method] ServiceDescriptorProto method
             * @property {google.protobuf.ServiceOptions|null} [options] ServiceDescriptorProto options
             */

            /**
             * Constructs a new ServiceDescriptorProto.
             * @memberof google.protobuf
             * @classdesc Represents a ServiceDescriptorProto.
             * @implements IServiceDescriptorProto
             * @constructor
             * @param {google.protobuf.IServiceDescriptorProto=} [properties] Properties to set
             */
            function ServiceDescriptorProto(properties) {
                this.method = [];
                if (properties)
                    for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                        if (properties[keys[i]] != null)
                            this[keys[i]] = properties[keys[i]];
            }

            /**
             * ServiceDescriptorProto name.
             * @member {string} name
             * @memberof google.protobuf.ServiceDescriptorProto
             * @instance
             */
            ServiceDescriptorProto.prototype.name = "";

            /**
             * ServiceDescriptorProto method.
             * @member {Array.<google.protobuf.MethodDescriptorProto>} method
             * @memberof google.protobuf.ServiceDescriptorProto
             * @instance
             */
            ServiceDescriptorProto.prototype.method = $util.emptyArray;

            /**
             * ServiceDescriptorProto options.
             * @member {google.protobuf.ServiceOptions|null|undefined} options
             * @memberof google.protobuf.ServiceDescriptorProto
             * @instance
             */
            ServiceDescriptorProto.prototype.options = null;

            /**
             * Encodes the specified ServiceDescriptorProto message. Does not implicitly {@link google.protobuf.ServiceDescriptorProto.verify|verify} messages.
             * @function encode
             * @memberof google.protobuf.ServiceDescriptorProto
             * @static
             * @param {google.protobuf.ServiceDescriptorProto} message ServiceDescriptorProto message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            ServiceDescriptorProto.encode = function encode(message, writer) {
                if (!writer)
                    writer = $Writer.create();
                if (message.name != null && Object.hasOwnProperty.call(message, "name"))
                    writer.uint32(/* id 1, wireType 2 =*/10).string(message.name);
                if (message.method != null && message.method.length)
                    for (let i = 0; i < message.method.length; ++i)
                        $root.google.protobuf.MethodDescriptorProto.encode(message.method[i], writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim();
                if (message.options != null && Object.hasOwnProperty.call(message, "options"))
                    $root.google.protobuf.ServiceOptions.encode(message.options, writer.uint32(/* id 3, wireType 2 =*/26).fork()).ldelim();
                return writer;
            };

            /**
             * Decodes a ServiceDescriptorProto message from the specified reader or buffer.
             * @function decode
             * @memberof google.protobuf.ServiceDescriptorProto
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @param {number} [length] Message length if known beforehand
             * @returns {google.protobuf.ServiceDescriptorProto} ServiceDescriptorProto
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            ServiceDescriptorProto.decode = function decode(reader, length) {
                if (!(reader instanceof $Reader))
                    reader = $Reader.create(reader);
                let end = length === undefined ? reader.len : reader.pos + length, message = new $root.google.protobuf.ServiceDescriptorProto();
                while (reader.pos < end) {
                    let tag = reader.uint32();
                    switch (tag >>> 3) {
                    case 1:
                        message.name = reader.string();
                        break;
                    case 2:
                        if (!(message.method && message.method.length))
                            message.method = [];
                        message.method.push($root.google.protobuf.MethodDescriptorProto.decode(reader, reader.uint32()));
                        break;
                    case 3:
                        message.options = $root.google.protobuf.ServiceOptions.decode(reader, reader.uint32());
                        break;
                    default:
                        reader.skipType(tag & 7);
                        break;
                    }
                }
                return message;
            };

            return ServiceDescriptorProto;
        })();

        protobuf.MethodDescriptorProto = (function() {

            /**
             * Properties of a MethodDescriptorProto.
             * @memberof google.protobuf
             * @interface IMethodDescriptorProto
             * @property {string|null} [name] MethodDescriptorProto name
             * @property {string|null} [input_type] MethodDescriptorProto input_type
             * @property {string|null} [output_type] MethodDescriptorProto output_type
             * @property {google.protobuf.MethodOptions|null} [options] MethodDescriptorProto options
             * @property {boolean|null} [client_streaming] MethodDescriptorProto client_streaming
             * @property {boolean|null} [server_streaming] MethodDescriptorProto server_streaming
             */

            /**
             * Constructs a new MethodDescriptorProto.
             * @memberof google.protobuf
             * @classdesc Represents a MethodDescriptorProto.
             * @implements IMethodDescriptorProto
             * @constructor
             * @param {google.protobuf.IMethodDescriptorProto=} [properties] Properties to set
             */
            function MethodDescriptorProto(properties) {
                if (properties)
                    for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                        if (properties[keys[i]] != null)
                            this[keys[i]] = properties[keys[i]];
            }

            /**
             * MethodDescriptorProto name.
             * @member {string} name
             * @memberof google.protobuf.MethodDescriptorProto
             * @instance
             */
            MethodDescriptorProto.prototype.name = "";

            /**
             * MethodDescriptorProto input_type.
             * @member {string} input_type
             * @memberof google.protobuf.MethodDescriptorProto
             * @instance
             */
            MethodDescriptorProto.prototype.input_type = "";

            /**
             * MethodDescriptorProto output_type.
             * @member {string} output_type
             * @memberof google.protobuf.MethodDescriptorProto
             * @instance
             */
            MethodDescriptorProto.prototype.output_type = "";

            /**
             * MethodDescriptorProto options.
             * @member {google.protobuf.MethodOptions|null|undefined} options
             * @memberof google.protobuf.MethodDescriptorProto
             * @instance
             */
            MethodDescriptorProto.prototype.options = null;

            /**
             * MethodDescriptorProto client_streaming.
             * @member {boolean} client_streaming
             * @memberof google.protobuf.MethodDescriptorProto
             * @instance
             */
            MethodDescriptorProto.prototype.client_streaming = false;

            /**
             * MethodDescriptorProto server_streaming.
             * @member {boolean} server_streaming
             * @memberof google.protobuf.MethodDescriptorProto
             * @instance
             */
            MethodDescriptorProto.prototype.server_streaming = false;

            /**
             * Encodes the specified MethodDescriptorProto message. Does not implicitly {@link google.protobuf.MethodDescriptorProto.verify|verify} messages.
             * @function encode
             * @memberof google.protobuf.MethodDescriptorProto
             * @static
             * @param {google.protobuf.MethodDescriptorProto} message MethodDescriptorProto message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            MethodDescriptorProto.encode = function encode(message, writer) {
                if (!writer)
                    writer = $Writer.create();
                if (message.name != null && Object.hasOwnProperty.call(message, "name"))
                    writer.uint32(/* id 1, wireType 2 =*/10).string(message.name);
                if (message.input_type != null && Object.hasOwnProperty.call(message, "input_type"))
                    writer.uint32(/* id 2, wireType 2 =*/18).string(message.input_type);
                if (message.output_type != null && Object.hasOwnProperty.call(message, "output_type"))
                    writer.uint32(/* id 3, wireType 2 =*/26).string(message.output_type);
                if (message.options != null && Object.hasOwnProperty.call(message, "options"))
                    $root.google.protobuf.MethodOptions.encode(message.options, writer.uint32(/* id 4, wireType 2 =*/34).fork()).ldelim();
                if (message.client_streaming != null && Object.hasOwnProperty.call(message, "client_streaming"))
                    writer.uint32(/* id 5, wireType 0 =*/40).bool(message.client_streaming);
                if (message.server_streaming != null && Object.hasOwnProperty.call(message, "server_streaming"))
                    writer.uint32(/* id 6, wireType 0 =*/48).bool(message.server_streaming);
                return writer;
            };

            /**
             * Decodes a MethodDescriptorProto message from the specified reader or buffer.
             * @function decode
             * @memberof google.protobuf.MethodDescriptorProto
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @param {number} [length] Message length if known beforehand
             * @returns {google.protobuf.MethodDescriptorProto} MethodDescriptorProto
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            MethodDescriptorProto.decode = function decode(reader, length) {
                if (!(reader instanceof $Reader))
                    reader = $Reader.create(reader);
                let end = length === undefined ? reader.len : reader.pos + length, message = new $root.google.protobuf.MethodDescriptorProto();
                while (reader.pos < end) {
                    let tag = reader.uint32();
                    switch (tag >>> 3) {
                    case 1:
                        message.name = reader.string();
                        break;
                    case 2:
                        message.input_type = reader.string();
                        break;
                    case 3:
                        message.output_type = reader.string();
                        break;
                    case 4:
                        message.options = $root.google.protobuf.MethodOptions.decode(reader, reader.uint32());
                        break;
                    case 5:
                        message.client_streaming = reader.bool();
                        break;
                    case 6:
                        message.server_streaming = reader.bool();
                        break;
                    default:
                        reader.skipType(tag & 7);
                        break;
                    }
                }
                return message;
            };

            return MethodDescriptorProto;
        })();

        protobuf.FileOptions = (function() {

            /**
             * Properties of a FileOptions.
             * @memberof google.protobuf
             * @interface IFileOptions
             * @property {string|null} [java_package] FileOptions java_package
             * @property {string|null} [java_outer_classname] FileOptions java_outer_classname
             * @property {boolean|null} [java_multiple_files] FileOptions java_multiple_files
             * @property {boolean|null} [java_generate_equals_and_hash] FileOptions java_generate_equals_and_hash
             * @property {boolean|null} [java_string_check_utf8] FileOptions java_string_check_utf8
             * @property {google.protobuf.FileOptions.OptimizeMode|null} [optimize_for] FileOptions optimize_for
             * @property {string|null} [go_package] FileOptions go_package
             * @property {boolean|null} [cc_generic_services] FileOptions cc_generic_services
             * @property {boolean|null} [java_generic_services] FileOptions java_generic_services
             * @property {boolean|null} [py_generic_services] FileOptions py_generic_services
             * @property {boolean|null} [deprecated] FileOptions deprecated
             * @property {boolean|null} [cc_enable_arenas] FileOptions cc_enable_arenas
             * @property {string|null} [objc_class_prefix] FileOptions objc_class_prefix
             * @property {string|null} [csharp_namespace] FileOptions csharp_namespace
             * @property {Array.<google.protobuf.UninterpretedOption>|null} [uninterpreted_option] FileOptions uninterpreted_option
             */

            /**
             * Constructs a new FileOptions.
             * @memberof google.protobuf
             * @classdesc Represents a FileOptions.
             * @implements IFileOptions
             * @constructor
             * @param {google.protobuf.IFileOptions=} [properties] Properties to set
             */
            function FileOptions(properties) {
                this.uninterpreted_option = [];
                if (properties)
                    for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                        if (properties[keys[i]] != null)
                            this[keys[i]] = properties[keys[i]];
            }

            /**
             * FileOptions java_package.
             * @member {string} java_package
             * @memberof google.protobuf.FileOptions
             * @instance
             */
            FileOptions.prototype.java_package = "";

            /**
             * FileOptions java_outer_classname.
             * @member {string} java_outer_classname
             * @memberof google.protobuf.FileOptions
             * @instance
             */
            FileOptions.prototype.java_outer_classname = "";

            /**
             * FileOptions java_multiple_files.
             * @member {boolean} java_multiple_files
             * @memberof google.protobuf.FileOptions
             * @instance
             */
            FileOptions.prototype.java_multiple_files = false;

            /**
             * FileOptions java_generate_equals_and_hash.
             * @member {boolean} java_generate_equals_and_hash
             * @memberof google.protobuf.FileOptions
             * @instance
             */
            FileOptions.prototype.java_generate_equals_and_hash = false;

            /**
             * FileOptions java_string_check_utf8.
             * @member {boolean} java_string_check_utf8
             * @memberof google.protobuf.FileOptions
             * @instance
             */
            FileOptions.prototype.java_string_check_utf8 = false;

            /**
             * FileOptions optimize_for.
             * @member {google.protobuf.FileOptions.OptimizeMode} optimize_for
             * @memberof google.protobuf.FileOptions
             * @instance
             */
            FileOptions.prototype.optimize_for = 1;

            /**
             * FileOptions go_package.
             * @member {string} go_package
             * @memberof google.protobuf.FileOptions
             * @instance
             */
            FileOptions.prototype.go_package = "";

            /**
             * FileOptions cc_generic_services.
             * @member {boolean} cc_generic_services
             * @memberof google.protobuf.FileOptions
             * @instance
             */
            FileOptions.prototype.cc_generic_services = false;

            /**
             * FileOptions java_generic_services.
             * @member {boolean} java_generic_services
             * @memberof google.protobuf.FileOptions
             * @instance
             */
            FileOptions.prototype.java_generic_services = false;

            /**
             * FileOptions py_generic_services.
             * @member {boolean} py_generic_services
             * @memberof google.protobuf.FileOptions
             * @instance
             */
            FileOptions.prototype.py_generic_services = false;

            /**
             * FileOptions deprecated.
             * @member {boolean} deprecated
             * @memberof google.protobuf.FileOptions
             * @instance
             */
            FileOptions.prototype.deprecated = false;

            /**
             * FileOptions cc_enable_arenas.
             * @member {boolean} cc_enable_arenas
             * @memberof google.protobuf.FileOptions
             * @instance
             */
            FileOptions.prototype.cc_enable_arenas = false;

            /**
             * FileOptions objc_class_prefix.
             * @member {string} objc_class_prefix
             * @memberof google.protobuf.FileOptions
             * @instance
             */
            FileOptions.prototype.objc_class_prefix = "";

            /**
             * FileOptions csharp_namespace.
             * @member {string} csharp_namespace
             * @memberof google.protobuf.FileOptions
             * @instance
             */
            FileOptions.prototype.csharp_namespace = "";

            /**
             * FileOptions uninterpreted_option.
             * @member {Array.<google.protobuf.UninterpretedOption>} uninterpreted_option
             * @memberof google.protobuf.FileOptions
             * @instance
             */
            FileOptions.prototype.uninterpreted_option = $util.emptyArray;

            /**
             * Encodes the specified FileOptions message. Does not implicitly {@link google.protobuf.FileOptions.verify|verify} messages.
             * @function encode
             * @memberof google.protobuf.FileOptions
             * @static
             * @param {google.protobuf.FileOptions} message FileOptions message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            FileOptions.encode = function encode(message, writer) {
                if (!writer)
                    writer = $Writer.create();
                if (message.java_package != null && Object.hasOwnProperty.call(message, "java_package"))
                    writer.uint32(/* id 1, wireType 2 =*/10).string(message.java_package);
                if (message.java_outer_classname != null && Object.hasOwnProperty.call(message, "java_outer_classname"))
                    writer.uint32(/* id 8, wireType 2 =*/66).string(message.java_outer_classname);
                if (message.optimize_for != null && Object.hasOwnProperty.call(message, "optimize_for"))
                    writer.uint32(/* id 9, wireType 0 =*/72).int32(message.optimize_for);
                if (message.java_multiple_files != null && Object.hasOwnProperty.call(message, "java_multiple_files"))
                    writer.uint32(/* id 10, wireType 0 =*/80).bool(message.java_multiple_files);
                if (message.go_package != null && Object.hasOwnProperty.call(message, "go_package"))
                    writer.uint32(/* id 11, wireType 2 =*/90).string(message.go_package);
                if (message.cc_generic_services != null && Object.hasOwnProperty.call(message, "cc_generic_services"))
                    writer.uint32(/* id 16, wireType 0 =*/128).bool(message.cc_generic_services);
                if (message.java_generic_services != null && Object.hasOwnProperty.call(message, "java_generic_services"))
                    writer.uint32(/* id 17, wireType 0 =*/136).bool(message.java_generic_services);
                if (message.py_generic_services != null && Object.hasOwnProperty.call(message, "py_generic_services"))
                    writer.uint32(/* id 18, wireType 0 =*/144).bool(message.py_generic_services);
                if (message.java_generate_equals_and_hash != null && Object.hasOwnProperty.call(message, "java_generate_equals_and_hash"))
                    writer.uint32(/* id 20, wireType 0 =*/160).bool(message.java_generate_equals_and_hash);
                if (message.deprecated != null && Object.hasOwnProperty.call(message, "deprecated"))
                    writer.uint32(/* id 23, wireType 0 =*/184).bool(message.deprecated);
                if (message.java_string_check_utf8 != null && Object.hasOwnProperty.call(message, "java_string_check_utf8"))
                    writer.uint32(/* id 27, wireType 0 =*/216).bool(message.java_string_check_utf8);
                if (message.cc_enable_arenas != null && Object.hasOwnProperty.call(message, "cc_enable_arenas"))
                    writer.uint32(/* id 31, wireType 0 =*/248).bool(message.cc_enable_arenas);
                if (message.objc_class_prefix != null && Object.hasOwnProperty.call(message, "objc_class_prefix"))
                    writer.uint32(/* id 36, wireType 2 =*/290).string(message.objc_class_prefix);
                if (message.csharp_namespace != null && Object.hasOwnProperty.call(message, "csharp_namespace"))
                    writer.uint32(/* id 37, wireType 2 =*/298).string(message.csharp_namespace);
                if (message.uninterpreted_option != null && message.uninterpreted_option.length)
                    for (let i = 0; i < message.uninterpreted_option.length; ++i)
                        $root.google.protobuf.UninterpretedOption.encode(message.uninterpreted_option[i], writer.uint32(/* id 999, wireType 2 =*/7994).fork()).ldelim();
                return writer;
            };

            /**
             * Decodes a FileOptions message from the specified reader or buffer.
             * @function decode
             * @memberof google.protobuf.FileOptions
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @param {number} [length] Message length if known beforehand
             * @returns {google.protobuf.FileOptions} FileOptions
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            FileOptions.decode = function decode(reader, length) {
                if (!(reader instanceof $Reader))
                    reader = $Reader.create(reader);
                let end = length === undefined ? reader.len : reader.pos + length, message = new $root.google.protobuf.FileOptions();
                while (reader.pos < end) {
                    let tag = reader.uint32();
                    switch (tag >>> 3) {
                    case 1:
                        message.java_package = reader.string();
                        break;
                    case 8:
                        message.java_outer_classname = reader.string();
                        break;
                    case 10:
                        message.java_multiple_files = reader.bool();
                        break;
                    case 20:
                        message.java_generate_equals_and_hash = reader.bool();
                        break;
                    case 27:
                        message.java_string_check_utf8 = reader.bool();
                        break;
                    case 9:
                        message.optimize_for = reader.int32();
                        break;
                    case 11:
                        message.go_package = reader.string();
                        break;
                    case 16:
                        message.cc_generic_services = reader.bool();
                        break;
                    case 17:
                        message.java_generic_services = reader.bool();
                        break;
                    case 18:
                        message.py_generic_services = reader.bool();
                        break;
                    case 23:
                        message.deprecated = reader.bool();
                        break;
                    case 31:
                        message.cc_enable_arenas = reader.bool();
                        break;
                    case 36:
                        message.objc_class_prefix = reader.string();
                        break;
                    case 37:
                        message.csharp_namespace = reader.string();
                        break;
                    case 999:
                        if (!(message.uninterpreted_option && message.uninterpreted_option.length))
                            message.uninterpreted_option = [];
                        message.uninterpreted_option.push($root.google.protobuf.UninterpretedOption.decode(reader, reader.uint32()));
                        break;
                    default:
                        reader.skipType(tag & 7);
                        break;
                    }
                }
                return message;
            };

            /**
             * OptimizeMode enum.
             * @name google.protobuf.FileOptions.OptimizeMode
             * @enum {number}
             * @property {number} SPEED=1 SPEED value
             * @property {number} CODE_SIZE=2 CODE_SIZE value
             * @property {number} LITE_RUNTIME=3 LITE_RUNTIME value
             */
            FileOptions.OptimizeMode = (function() {
                const valuesById = {}, values = Object.create(valuesById);
                values[valuesById[1] = "SPEED"] = 1;
                values[valuesById[2] = "CODE_SIZE"] = 2;
                values[valuesById[3] = "LITE_RUNTIME"] = 3;
                return values;
            })();

            return FileOptions;
        })();

        protobuf.MessageOptions = (function() {

            /**
             * Properties of a MessageOptions.
             * @memberof google.protobuf
             * @interface IMessageOptions
             * @property {boolean|null} [message_set_wire_format] MessageOptions message_set_wire_format
             * @property {boolean|null} [no_standard_descriptor_accessor] MessageOptions no_standard_descriptor_accessor
             * @property {boolean|null} [deprecated] MessageOptions deprecated
             * @property {boolean|null} [map_entry] MessageOptions map_entry
             * @property {Array.<google.protobuf.UninterpretedOption>|null} [uninterpreted_option] MessageOptions uninterpreted_option
             */

            /**
             * Constructs a new MessageOptions.
             * @memberof google.protobuf
             * @classdesc Represents a MessageOptions.
             * @implements IMessageOptions
             * @constructor
             * @param {google.protobuf.IMessageOptions=} [properties] Properties to set
             */
            function MessageOptions(properties) {
                this.uninterpreted_option = [];
                if (properties)
                    for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                        if (properties[keys[i]] != null)
                            this[keys[i]] = properties[keys[i]];
            }

            /**
             * MessageOptions message_set_wire_format.
             * @member {boolean} message_set_wire_format
             * @memberof google.protobuf.MessageOptions
             * @instance
             */
            MessageOptions.prototype.message_set_wire_format = false;

            /**
             * MessageOptions no_standard_descriptor_accessor.
             * @member {boolean} no_standard_descriptor_accessor
             * @memberof google.protobuf.MessageOptions
             * @instance
             */
            MessageOptions.prototype.no_standard_descriptor_accessor = false;

            /**
             * MessageOptions deprecated.
             * @member {boolean} deprecated
             * @memberof google.protobuf.MessageOptions
             * @instance
             */
            MessageOptions.prototype.deprecated = false;

            /**
             * MessageOptions map_entry.
             * @member {boolean} map_entry
             * @memberof google.protobuf.MessageOptions
             * @instance
             */
            MessageOptions.prototype.map_entry = false;

            /**
             * MessageOptions uninterpreted_option.
             * @member {Array.<google.protobuf.UninterpretedOption>} uninterpreted_option
             * @memberof google.protobuf.MessageOptions
             * @instance
             */
            MessageOptions.prototype.uninterpreted_option = $util.emptyArray;

            /**
             * Encodes the specified MessageOptions message. Does not implicitly {@link google.protobuf.MessageOptions.verify|verify} messages.
             * @function encode
             * @memberof google.protobuf.MessageOptions
             * @static
             * @param {google.protobuf.MessageOptions} message MessageOptions message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            MessageOptions.encode = function encode(message, writer) {
                if (!writer)
                    writer = $Writer.create();
                if (message.message_set_wire_format != null && Object.hasOwnProperty.call(message, "message_set_wire_format"))
                    writer.uint32(/* id 1, wireType 0 =*/8).bool(message.message_set_wire_format);
                if (message.no_standard_descriptor_accessor != null && Object.hasOwnProperty.call(message, "no_standard_descriptor_accessor"))
                    writer.uint32(/* id 2, wireType 0 =*/16).bool(message.no_standard_descriptor_accessor);
                if (message.deprecated != null && Object.hasOwnProperty.call(message, "deprecated"))
                    writer.uint32(/* id 3, wireType 0 =*/24).bool(message.deprecated);
                if (message.map_entry != null && Object.hasOwnProperty.call(message, "map_entry"))
                    writer.uint32(/* id 7, wireType 0 =*/56).bool(message.map_entry);
                if (message.uninterpreted_option != null && message.uninterpreted_option.length)
                    for (let i = 0; i < message.uninterpreted_option.length; ++i)
                        $root.google.protobuf.UninterpretedOption.encode(message.uninterpreted_option[i], writer.uint32(/* id 999, wireType 2 =*/7994).fork()).ldelim();
                return writer;
            };

            /**
             * Decodes a MessageOptions message from the specified reader or buffer.
             * @function decode
             * @memberof google.protobuf.MessageOptions
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @param {number} [length] Message length if known beforehand
             * @returns {google.protobuf.MessageOptions} MessageOptions
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            MessageOptions.decode = function decode(reader, length) {
                if (!(reader instanceof $Reader))
                    reader = $Reader.create(reader);
                let end = length === undefined ? reader.len : reader.pos + length, message = new $root.google.protobuf.MessageOptions();
                while (reader.pos < end) {
                    let tag = reader.uint32();
                    switch (tag >>> 3) {
                    case 1:
                        message.message_set_wire_format = reader.bool();
                        break;
                    case 2:
                        message.no_standard_descriptor_accessor = reader.bool();
                        break;
                    case 3:
                        message.deprecated = reader.bool();
                        break;
                    case 7:
                        message.map_entry = reader.bool();
                        break;
                    case 999:
                        if (!(message.uninterpreted_option && message.uninterpreted_option.length))
                            message.uninterpreted_option = [];
                        message.uninterpreted_option.push($root.google.protobuf.UninterpretedOption.decode(reader, reader.uint32()));
                        break;
                    default:
                        reader.skipType(tag & 7);
                        break;
                    }
                }
                return message;
            };

            return MessageOptions;
        })();

        protobuf.FieldOptions = (function() {

            /**
             * Properties of a FieldOptions.
             * @memberof google.protobuf
             * @interface IFieldOptions
             * @property {google.protobuf.FieldOptions.CType|null} [ctype] FieldOptions ctype
             * @property {boolean|null} [packed] FieldOptions packed
             * @property {google.protobuf.FieldOptions.JSType|null} [jstype] FieldOptions jstype
             * @property {boolean|null} [lazy] FieldOptions lazy
             * @property {boolean|null} [deprecated] FieldOptions deprecated
             * @property {boolean|null} [weak] FieldOptions weak
             * @property {Array.<google.protobuf.UninterpretedOption>|null} [uninterpreted_option] FieldOptions uninterpreted_option
             */

            /**
             * Constructs a new FieldOptions.
             * @memberof google.protobuf
             * @classdesc Represents a FieldOptions.
             * @implements IFieldOptions
             * @constructor
             * @param {google.protobuf.IFieldOptions=} [properties] Properties to set
             */
            function FieldOptions(properties) {
                this.uninterpreted_option = [];
                if (properties)
                    for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                        if (properties[keys[i]] != null)
                            this[keys[i]] = properties[keys[i]];
            }

            /**
             * FieldOptions ctype.
             * @member {google.protobuf.FieldOptions.CType} ctype
             * @memberof google.protobuf.FieldOptions
             * @instance
             */
            FieldOptions.prototype.ctype = 0;

            /**
             * FieldOptions packed.
             * @member {boolean} packed
             * @memberof google.protobuf.FieldOptions
             * @instance
             */
            FieldOptions.prototype.packed = false;

            /**
             * FieldOptions jstype.
             * @member {google.protobuf.FieldOptions.JSType} jstype
             * @memberof google.protobuf.FieldOptions
             * @instance
             */
            FieldOptions.prototype.jstype = 0;

            /**
             * FieldOptions lazy.
             * @member {boolean} lazy
             * @memberof google.protobuf.FieldOptions
             * @instance
             */
            FieldOptions.prototype.lazy = false;

            /**
             * FieldOptions deprecated.
             * @member {boolean} deprecated
             * @memberof google.protobuf.FieldOptions
             * @instance
             */
            FieldOptions.prototype.deprecated = false;

            /**
             * FieldOptions weak.
             * @member {boolean} weak
             * @memberof google.protobuf.FieldOptions
             * @instance
             */
            FieldOptions.prototype.weak = false;

            /**
             * FieldOptions uninterpreted_option.
             * @member {Array.<google.protobuf.UninterpretedOption>} uninterpreted_option
             * @memberof google.protobuf.FieldOptions
             * @instance
             */
            FieldOptions.prototype.uninterpreted_option = $util.emptyArray;

            /**
             * Encodes the specified FieldOptions message. Does not implicitly {@link google.protobuf.FieldOptions.verify|verify} messages.
             * @function encode
             * @memberof google.protobuf.FieldOptions
             * @static
             * @param {google.protobuf.FieldOptions} message FieldOptions message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            FieldOptions.encode = function encode(message, writer) {
                if (!writer)
                    writer = $Writer.create();
                if (message.ctype != null && Object.hasOwnProperty.call(message, "ctype"))
                    writer.uint32(/* id 1, wireType 0 =*/8).int32(message.ctype);
                if (message.packed != null && Object.hasOwnProperty.call(message, "packed"))
                    writer.uint32(/* id 2, wireType 0 =*/16).bool(message.packed);
                if (message.deprecated != null && Object.hasOwnProperty.call(message, "deprecated"))
                    writer.uint32(/* id 3, wireType 0 =*/24).bool(message.deprecated);
                if (message.lazy != null && Object.hasOwnProperty.call(message, "lazy"))
                    writer.uint32(/* id 5, wireType 0 =*/40).bool(message.lazy);
                if (message.jstype != null && Object.hasOwnProperty.call(message, "jstype"))
                    writer.uint32(/* id 6, wireType 0 =*/48).int32(message.jstype);
                if (message.weak != null && Object.hasOwnProperty.call(message, "weak"))
                    writer.uint32(/* id 10, wireType 0 =*/80).bool(message.weak);
                if (message.uninterpreted_option != null && message.uninterpreted_option.length)
                    for (let i = 0; i < message.uninterpreted_option.length; ++i)
                        $root.google.protobuf.UninterpretedOption.encode(message.uninterpreted_option[i], writer.uint32(/* id 999, wireType 2 =*/7994).fork()).ldelim();
                return writer;
            };

            /**
             * Decodes a FieldOptions message from the specified reader or buffer.
             * @function decode
             * @memberof google.protobuf.FieldOptions
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @param {number} [length] Message length if known beforehand
             * @returns {google.protobuf.FieldOptions} FieldOptions
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            FieldOptions.decode = function decode(reader, length) {
                if (!(reader instanceof $Reader))
                    reader = $Reader.create(reader);
                let end = length === undefined ? reader.len : reader.pos + length, message = new $root.google.protobuf.FieldOptions();
                while (reader.pos < end) {
                    let tag = reader.uint32();
                    switch (tag >>> 3) {
                    case 1:
                        message.ctype = reader.int32();
                        break;
                    case 2:
                        message.packed = reader.bool();
                        break;
                    case 6:
                        message.jstype = reader.int32();
                        break;
                    case 5:
                        message.lazy = reader.bool();
                        break;
                    case 3:
                        message.deprecated = reader.bool();
                        break;
                    case 10:
                        message.weak = reader.bool();
                        break;
                    case 999:
                        if (!(message.uninterpreted_option && message.uninterpreted_option.length))
                            message.uninterpreted_option = [];
                        message.uninterpreted_option.push($root.google.protobuf.UninterpretedOption.decode(reader, reader.uint32()));
                        break;
                    default:
                        reader.skipType(tag & 7);
                        break;
                    }
                }
                return message;
            };

            /**
             * CType enum.
             * @name google.protobuf.FieldOptions.CType
             * @enum {number}
             * @property {number} STRING=0 STRING value
             * @property {number} CORD=1 CORD value
             * @property {number} STRING_PIECE=2 STRING_PIECE value
             */
            FieldOptions.CType = (function() {
                const valuesById = {}, values = Object.create(valuesById);
                values[valuesById[0] = "STRING"] = 0;
                values[valuesById[1] = "CORD"] = 1;
                values[valuesById[2] = "STRING_PIECE"] = 2;
                return values;
            })();

            /**
             * JSType enum.
             * @name google.protobuf.FieldOptions.JSType
             * @enum {number}
             * @property {number} JS_NORMAL=0 JS_NORMAL value
             * @property {number} JS_STRING=1 JS_STRING value
             * @property {number} JS_NUMBER=2 JS_NUMBER value
             */
            FieldOptions.JSType = (function() {
                const valuesById = {}, values = Object.create(valuesById);
                values[valuesById[0] = "JS_NORMAL"] = 0;
                values[valuesById[1] = "JS_STRING"] = 1;
                values[valuesById[2] = "JS_NUMBER"] = 2;
                return values;
            })();

            return FieldOptions;
        })();

        protobuf.OneofOptions = (function() {

            /**
             * Properties of an OneofOptions.
             * @memberof google.protobuf
             * @interface IOneofOptions
             * @property {Array.<google.protobuf.UninterpretedOption>|null} [uninterpreted_option] OneofOptions uninterpreted_option
             */

            /**
             * Constructs a new OneofOptions.
             * @memberof google.protobuf
             * @classdesc Represents an OneofOptions.
             * @implements IOneofOptions
             * @constructor
             * @param {google.protobuf.IOneofOptions=} [properties] Properties to set
             */
            function OneofOptions(properties) {
                this.uninterpreted_option = [];
                if (properties)
                    for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                        if (properties[keys[i]] != null)
                            this[keys[i]] = properties[keys[i]];
            }

            /**
             * OneofOptions uninterpreted_option.
             * @member {Array.<google.protobuf.UninterpretedOption>} uninterpreted_option
             * @memberof google.protobuf.OneofOptions
             * @instance
             */
            OneofOptions.prototype.uninterpreted_option = $util.emptyArray;

            /**
             * Encodes the specified OneofOptions message. Does not implicitly {@link google.protobuf.OneofOptions.verify|verify} messages.
             * @function encode
             * @memberof google.protobuf.OneofOptions
             * @static
             * @param {google.protobuf.OneofOptions} message OneofOptions message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            OneofOptions.encode = function encode(message, writer) {
                if (!writer)
                    writer = $Writer.create();
                if (message.uninterpreted_option != null && message.uninterpreted_option.length)
                    for (let i = 0; i < message.uninterpreted_option.length; ++i)
                        $root.google.protobuf.UninterpretedOption.encode(message.uninterpreted_option[i], writer.uint32(/* id 999, wireType 2 =*/7994).fork()).ldelim();
                return writer;
            };

            /**
             * Decodes an OneofOptions message from the specified reader or buffer.
             * @function decode
             * @memberof google.protobuf.OneofOptions
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @param {number} [length] Message length if known beforehand
             * @returns {google.protobuf.OneofOptions} OneofOptions
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            OneofOptions.decode = function decode(reader, length) {
                if (!(reader instanceof $Reader))
                    reader = $Reader.create(reader);
                let end = length === undefined ? reader.len : reader.pos + length, message = new $root.google.protobuf.OneofOptions();
                while (reader.pos < end) {
                    let tag = reader.uint32();
                    switch (tag >>> 3) {
                    case 999:
                        if (!(message.uninterpreted_option && message.uninterpreted_option.length))
                            message.uninterpreted_option = [];
                        message.uninterpreted_option.push($root.google.protobuf.UninterpretedOption.decode(reader, reader.uint32()));
                        break;
                    default:
                        reader.skipType(tag & 7);
                        break;
                    }
                }
                return message;
            };

            return OneofOptions;
        })();

        protobuf.EnumOptions = (function() {

            /**
             * Properties of an EnumOptions.
             * @memberof google.protobuf
             * @interface IEnumOptions
             * @property {boolean|null} [allow_alias] EnumOptions allow_alias
             * @property {boolean|null} [deprecated] EnumOptions deprecated
             * @property {Array.<google.protobuf.UninterpretedOption>|null} [uninterpreted_option] EnumOptions uninterpreted_option
             */

            /**
             * Constructs a new EnumOptions.
             * @memberof google.protobuf
             * @classdesc Represents an EnumOptions.
             * @implements IEnumOptions
             * @constructor
             * @param {google.protobuf.IEnumOptions=} [properties] Properties to set
             */
            function EnumOptions(properties) {
                this.uninterpreted_option = [];
                if (properties)
                    for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                        if (properties[keys[i]] != null)
                            this[keys[i]] = properties[keys[i]];
            }

            /**
             * EnumOptions allow_alias.
             * @member {boolean} allow_alias
             * @memberof google.protobuf.EnumOptions
             * @instance
             */
            EnumOptions.prototype.allow_alias = false;

            /**
             * EnumOptions deprecated.
             * @member {boolean} deprecated
             * @memberof google.protobuf.EnumOptions
             * @instance
             */
            EnumOptions.prototype.deprecated = false;

            /**
             * EnumOptions uninterpreted_option.
             * @member {Array.<google.protobuf.UninterpretedOption>} uninterpreted_option
             * @memberof google.protobuf.EnumOptions
             * @instance
             */
            EnumOptions.prototype.uninterpreted_option = $util.emptyArray;

            /**
             * Encodes the specified EnumOptions message. Does not implicitly {@link google.protobuf.EnumOptions.verify|verify} messages.
             * @function encode
             * @memberof google.protobuf.EnumOptions
             * @static
             * @param {google.protobuf.EnumOptions} message EnumOptions message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            EnumOptions.encode = function encode(message, writer) {
                if (!writer)
                    writer = $Writer.create();
                if (message.allow_alias != null && Object.hasOwnProperty.call(message, "allow_alias"))
                    writer.uint32(/* id 2, wireType 0 =*/16).bool(message.allow_alias);
                if (message.deprecated != null && Object.hasOwnProperty.call(message, "deprecated"))
                    writer.uint32(/* id 3, wireType 0 =*/24).bool(message.deprecated);
                if (message.uninterpreted_option != null && message.uninterpreted_option.length)
                    for (let i = 0; i < message.uninterpreted_option.length; ++i)
                        $root.google.protobuf.UninterpretedOption.encode(message.uninterpreted_option[i], writer.uint32(/* id 999, wireType 2 =*/7994).fork()).ldelim();
                return writer;
            };

            /**
             * Decodes an EnumOptions message from the specified reader or buffer.
             * @function decode
             * @memberof google.protobuf.EnumOptions
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @param {number} [length] Message length if known beforehand
             * @returns {google.protobuf.EnumOptions} EnumOptions
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            EnumOptions.decode = function decode(reader, length) {
                if (!(reader instanceof $Reader))
                    reader = $Reader.create(reader);
                let end = length === undefined ? reader.len : reader.pos + length, message = new $root.google.protobuf.EnumOptions();
                while (reader.pos < end) {
                    let tag = reader.uint32();
                    switch (tag >>> 3) {
                    case 2:
                        message.allow_alias = reader.bool();
                        break;
                    case 3:
                        message.deprecated = reader.bool();
                        break;
                    case 999:
                        if (!(message.uninterpreted_option && message.uninterpreted_option.length))
                            message.uninterpreted_option = [];
                        message.uninterpreted_option.push($root.google.protobuf.UninterpretedOption.decode(reader, reader.uint32()));
                        break;
                    default:
                        reader.skipType(tag & 7);
                        break;
                    }
                }
                return message;
            };

            return EnumOptions;
        })();

        protobuf.EnumValueOptions = (function() {

            /**
             * Properties of an EnumValueOptions.
             * @memberof google.protobuf
             * @interface IEnumValueOptions
             * @property {boolean|null} [deprecated] EnumValueOptions deprecated
             * @property {Array.<google.protobuf.UninterpretedOption>|null} [uninterpreted_option] EnumValueOptions uninterpreted_option
             */

            /**
             * Constructs a new EnumValueOptions.
             * @memberof google.protobuf
             * @classdesc Represents an EnumValueOptions.
             * @implements IEnumValueOptions
             * @constructor
             * @param {google.protobuf.IEnumValueOptions=} [properties] Properties to set
             */
            function EnumValueOptions(properties) {
                this.uninterpreted_option = [];
                if (properties)
                    for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                        if (properties[keys[i]] != null)
                            this[keys[i]] = properties[keys[i]];
            }

            /**
             * EnumValueOptions deprecated.
             * @member {boolean} deprecated
             * @memberof google.protobuf.EnumValueOptions
             * @instance
             */
            EnumValueOptions.prototype.deprecated = false;

            /**
             * EnumValueOptions uninterpreted_option.
             * @member {Array.<google.protobuf.UninterpretedOption>} uninterpreted_option
             * @memberof google.protobuf.EnumValueOptions
             * @instance
             */
            EnumValueOptions.prototype.uninterpreted_option = $util.emptyArray;

            /**
             * Encodes the specified EnumValueOptions message. Does not implicitly {@link google.protobuf.EnumValueOptions.verify|verify} messages.
             * @function encode
             * @memberof google.protobuf.EnumValueOptions
             * @static
             * @param {google.protobuf.EnumValueOptions} message EnumValueOptions message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            EnumValueOptions.encode = function encode(message, writer) {
                if (!writer)
                    writer = $Writer.create();
                if (message.deprecated != null && Object.hasOwnProperty.call(message, "deprecated"))
                    writer.uint32(/* id 1, wireType 0 =*/8).bool(message.deprecated);
                if (message.uninterpreted_option != null && message.uninterpreted_option.length)
                    for (let i = 0; i < message.uninterpreted_option.length; ++i)
                        $root.google.protobuf.UninterpretedOption.encode(message.uninterpreted_option[i], writer.uint32(/* id 999, wireType 2 =*/7994).fork()).ldelim();
                return writer;
            };

            /**
             * Decodes an EnumValueOptions message from the specified reader or buffer.
             * @function decode
             * @memberof google.protobuf.EnumValueOptions
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @param {number} [length] Message length if known beforehand
             * @returns {google.protobuf.EnumValueOptions} EnumValueOptions
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            EnumValueOptions.decode = function decode(reader, length) {
                if (!(reader instanceof $Reader))
                    reader = $Reader.create(reader);
                let end = length === undefined ? reader.len : reader.pos + length, message = new $root.google.protobuf.EnumValueOptions();
                while (reader.pos < end) {
                    let tag = reader.uint32();
                    switch (tag >>> 3) {
                    case 1:
                        message.deprecated = reader.bool();
                        break;
                    case 999:
                        if (!(message.uninterpreted_option && message.uninterpreted_option.length))
                            message.uninterpreted_option = [];
                        message.uninterpreted_option.push($root.google.protobuf.UninterpretedOption.decode(reader, reader.uint32()));
                        break;
                    default:
                        reader.skipType(tag & 7);
                        break;
                    }
                }
                return message;
            };

            return EnumValueOptions;
        })();

        protobuf.ServiceOptions = (function() {

            /**
             * Properties of a ServiceOptions.
             * @memberof google.protobuf
             * @interface IServiceOptions
             * @property {boolean|null} [deprecated] ServiceOptions deprecated
             * @property {Array.<google.protobuf.UninterpretedOption>|null} [uninterpreted_option] ServiceOptions uninterpreted_option
             */

            /**
             * Constructs a new ServiceOptions.
             * @memberof google.protobuf
             * @classdesc Represents a ServiceOptions.
             * @implements IServiceOptions
             * @constructor
             * @param {google.protobuf.IServiceOptions=} [properties] Properties to set
             */
            function ServiceOptions(properties) {
                this.uninterpreted_option = [];
                if (properties)
                    for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                        if (properties[keys[i]] != null)
                            this[keys[i]] = properties[keys[i]];
            }

            /**
             * ServiceOptions deprecated.
             * @member {boolean} deprecated
             * @memberof google.protobuf.ServiceOptions
             * @instance
             */
            ServiceOptions.prototype.deprecated = false;

            /**
             * ServiceOptions uninterpreted_option.
             * @member {Array.<google.protobuf.UninterpretedOption>} uninterpreted_option
             * @memberof google.protobuf.ServiceOptions
             * @instance
             */
            ServiceOptions.prototype.uninterpreted_option = $util.emptyArray;

            /**
             * Encodes the specified ServiceOptions message. Does not implicitly {@link google.protobuf.ServiceOptions.verify|verify} messages.
             * @function encode
             * @memberof google.protobuf.ServiceOptions
             * @static
             * @param {google.protobuf.ServiceOptions} message ServiceOptions message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            ServiceOptions.encode = function encode(message, writer) {
                if (!writer)
                    writer = $Writer.create();
                if (message.deprecated != null && Object.hasOwnProperty.call(message, "deprecated"))
                    writer.uint32(/* id 33, wireType 0 =*/264).bool(message.deprecated);
                if (message.uninterpreted_option != null && message.uninterpreted_option.length)
                    for (let i = 0; i < message.uninterpreted_option.length; ++i)
                        $root.google.protobuf.UninterpretedOption.encode(message.uninterpreted_option[i], writer.uint32(/* id 999, wireType 2 =*/7994).fork()).ldelim();
                return writer;
            };

            /**
             * Decodes a ServiceOptions message from the specified reader or buffer.
             * @function decode
             * @memberof google.protobuf.ServiceOptions
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @param {number} [length] Message length if known beforehand
             * @returns {google.protobuf.ServiceOptions} ServiceOptions
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            ServiceOptions.decode = function decode(reader, length) {
                if (!(reader instanceof $Reader))
                    reader = $Reader.create(reader);
                let end = length === undefined ? reader.len : reader.pos + length, message = new $root.google.protobuf.ServiceOptions();
                while (reader.pos < end) {
                    let tag = reader.uint32();
                    switch (tag >>> 3) {
                    case 33:
                        message.deprecated = reader.bool();
                        break;
                    case 999:
                        if (!(message.uninterpreted_option && message.uninterpreted_option.length))
                            message.uninterpreted_option = [];
                        message.uninterpreted_option.push($root.google.protobuf.UninterpretedOption.decode(reader, reader.uint32()));
                        break;
                    default:
                        reader.skipType(tag & 7);
                        break;
                    }
                }
                return message;
            };

            return ServiceOptions;
        })();

        protobuf.MethodOptions = (function() {

            /**
             * Properties of a MethodOptions.
             * @memberof google.protobuf
             * @interface IMethodOptions
             * @property {boolean|null} [deprecated] MethodOptions deprecated
             * @property {Array.<google.protobuf.UninterpretedOption>|null} [uninterpreted_option] MethodOptions uninterpreted_option
             * @property {google.api.HttpRule|null} [".google.api.http"] MethodOptions .google.api.http
             */

            /**
             * Constructs a new MethodOptions.
             * @memberof google.protobuf
             * @classdesc Represents a MethodOptions.
             * @implements IMethodOptions
             * @constructor
             * @param {google.protobuf.IMethodOptions=} [properties] Properties to set
             */
            function MethodOptions(properties) {
                this.uninterpreted_option = [];
                if (properties)
                    for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                        if (properties[keys[i]] != null)
                            this[keys[i]] = properties[keys[i]];
            }

            /**
             * MethodOptions deprecated.
             * @member {boolean} deprecated
             * @memberof google.protobuf.MethodOptions
             * @instance
             */
            MethodOptions.prototype.deprecated = false;

            /**
             * MethodOptions uninterpreted_option.
             * @member {Array.<google.protobuf.UninterpretedOption>} uninterpreted_option
             * @memberof google.protobuf.MethodOptions
             * @instance
             */
            MethodOptions.prototype.uninterpreted_option = $util.emptyArray;

            /**
             * MethodOptions .google.api.http.
             * @member {google.api.HttpRule|null|undefined} .google.api.http
             * @memberof google.protobuf.MethodOptions
             * @instance
             */
            MethodOptions.prototype[".google.api.http"] = null;

            /**
             * Encodes the specified MethodOptions message. Does not implicitly {@link google.protobuf.MethodOptions.verify|verify} messages.
             * @function encode
             * @memberof google.protobuf.MethodOptions
             * @static
             * @param {google.protobuf.MethodOptions} message MethodOptions message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            MethodOptions.encode = function encode(message, writer) {
                if (!writer)
                    writer = $Writer.create();
                if (message.deprecated != null && Object.hasOwnProperty.call(message, "deprecated"))
                    writer.uint32(/* id 33, wireType 0 =*/264).bool(message.deprecated);
                if (message.uninterpreted_option != null && message.uninterpreted_option.length)
                    for (let i = 0; i < message.uninterpreted_option.length; ++i)
                        $root.google.protobuf.UninterpretedOption.encode(message.uninterpreted_option[i], writer.uint32(/* id 999, wireType 2 =*/7994).fork()).ldelim();
                if (message[".google.api.http"] != null && Object.hasOwnProperty.call(message, ".google.api.http"))
                    $root.google.api.HttpRule.encode(message[".google.api.http"], writer.uint32(/* id 72295728, wireType 2 =*/578365826).fork()).ldelim();
                return writer;
            };

            /**
             * Decodes a MethodOptions message from the specified reader or buffer.
             * @function decode
             * @memberof google.protobuf.MethodOptions
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @param {number} [length] Message length if known beforehand
             * @returns {google.protobuf.MethodOptions} MethodOptions
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            MethodOptions.decode = function decode(reader, length) {
                if (!(reader instanceof $Reader))
                    reader = $Reader.create(reader);
                let end = length === undefined ? reader.len : reader.pos + length, message = new $root.google.protobuf.MethodOptions();
                while (reader.pos < end) {
                    let tag = reader.uint32();
                    switch (tag >>> 3) {
                    case 33:
                        message.deprecated = reader.bool();
                        break;
                    case 999:
                        if (!(message.uninterpreted_option && message.uninterpreted_option.length))
                            message.uninterpreted_option = [];
                        message.uninterpreted_option.push($root.google.protobuf.UninterpretedOption.decode(reader, reader.uint32()));
                        break;
                    case 72295728:
                        message[".google.api.http"] = $root.google.api.HttpRule.decode(reader, reader.uint32());
                        break;
                    default:
                        reader.skipType(tag & 7);
                        break;
                    }
                }
                return message;
            };

            return MethodOptions;
        })();

        protobuf.UninterpretedOption = (function() {

            /**
             * Properties of an UninterpretedOption.
             * @memberof google.protobuf
             * @interface IUninterpretedOption
             * @property {Array.<google.protobuf.UninterpretedOption.NamePart>|null} [name] UninterpretedOption name
             * @property {string|null} [identifier_value] UninterpretedOption identifier_value
             * @property {number|null} [positive_int_value] UninterpretedOption positive_int_value
             * @property {number|null} [negative_int_value] UninterpretedOption negative_int_value
             * @property {number|null} [double_value] UninterpretedOption double_value
             * @property {Uint8Array|null} [string_value] UninterpretedOption string_value
             * @property {string|null} [aggregate_value] UninterpretedOption aggregate_value
             */

            /**
             * Constructs a new UninterpretedOption.
             * @memberof google.protobuf
             * @classdesc Represents an UninterpretedOption.
             * @implements IUninterpretedOption
             * @constructor
             * @param {google.protobuf.IUninterpretedOption=} [properties] Properties to set
             */
            function UninterpretedOption(properties) {
                this.name = [];
                if (properties)
                    for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                        if (properties[keys[i]] != null)
                            this[keys[i]] = properties[keys[i]];
            }

            /**
             * UninterpretedOption name.
             * @member {Array.<google.protobuf.UninterpretedOption.NamePart>} name
             * @memberof google.protobuf.UninterpretedOption
             * @instance
             */
            UninterpretedOption.prototype.name = $util.emptyArray;

            /**
             * UninterpretedOption identifier_value.
             * @member {string} identifier_value
             * @memberof google.protobuf.UninterpretedOption
             * @instance
             */
            UninterpretedOption.prototype.identifier_value = "";

            /**
             * UninterpretedOption positive_int_value.
             * @member {number} positive_int_value
             * @memberof google.protobuf.UninterpretedOption
             * @instance
             */
            UninterpretedOption.prototype.positive_int_value = $util.Long ? $util.Long.fromBits(0,0,true) : 0;

            /**
             * UninterpretedOption negative_int_value.
             * @member {number} negative_int_value
             * @memberof google.protobuf.UninterpretedOption
             * @instance
             */
            UninterpretedOption.prototype.negative_int_value = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

            /**
             * UninterpretedOption double_value.
             * @member {number} double_value
             * @memberof google.protobuf.UninterpretedOption
             * @instance
             */
            UninterpretedOption.prototype.double_value = 0;

            /**
             * UninterpretedOption string_value.
             * @member {Uint8Array} string_value
             * @memberof google.protobuf.UninterpretedOption
             * @instance
             */
            UninterpretedOption.prototype.string_value = $util.newBuffer([]);

            /**
             * UninterpretedOption aggregate_value.
             * @member {string} aggregate_value
             * @memberof google.protobuf.UninterpretedOption
             * @instance
             */
            UninterpretedOption.prototype.aggregate_value = "";

            /**
             * Encodes the specified UninterpretedOption message. Does not implicitly {@link google.protobuf.UninterpretedOption.verify|verify} messages.
             * @function encode
             * @memberof google.protobuf.UninterpretedOption
             * @static
             * @param {google.protobuf.UninterpretedOption} message UninterpretedOption message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            UninterpretedOption.encode = function encode(message, writer) {
                if (!writer)
                    writer = $Writer.create();
                if (message.name != null && message.name.length)
                    for (let i = 0; i < message.name.length; ++i)
                        $root.google.protobuf.UninterpretedOption.NamePart.encode(message.name[i], writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim();
                if (message.identifier_value != null && Object.hasOwnProperty.call(message, "identifier_value"))
                    writer.uint32(/* id 3, wireType 2 =*/26).string(message.identifier_value);
                if (message.positive_int_value != null && Object.hasOwnProperty.call(message, "positive_int_value"))
                    writer.uint32(/* id 4, wireType 0 =*/32).uint64(message.positive_int_value);
                if (message.negative_int_value != null && Object.hasOwnProperty.call(message, "negative_int_value"))
                    writer.uint32(/* id 5, wireType 0 =*/40).int64(message.negative_int_value);
                if (message.double_value != null && Object.hasOwnProperty.call(message, "double_value"))
                    writer.uint32(/* id 6, wireType 1 =*/49).double(message.double_value);
                if (message.string_value != null && Object.hasOwnProperty.call(message, "string_value"))
                    writer.uint32(/* id 7, wireType 2 =*/58).bytes(message.string_value);
                if (message.aggregate_value != null && Object.hasOwnProperty.call(message, "aggregate_value"))
                    writer.uint32(/* id 8, wireType 2 =*/66).string(message.aggregate_value);
                return writer;
            };

            /**
             * Decodes an UninterpretedOption message from the specified reader or buffer.
             * @function decode
             * @memberof google.protobuf.UninterpretedOption
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @param {number} [length] Message length if known beforehand
             * @returns {google.protobuf.UninterpretedOption} UninterpretedOption
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            UninterpretedOption.decode = function decode(reader, length) {
                if (!(reader instanceof $Reader))
                    reader = $Reader.create(reader);
                let end = length === undefined ? reader.len : reader.pos + length, message = new $root.google.protobuf.UninterpretedOption();
                while (reader.pos < end) {
                    let tag = reader.uint32();
                    switch (tag >>> 3) {
                    case 2:
                        if (!(message.name && message.name.length))
                            message.name = [];
                        message.name.push($root.google.protobuf.UninterpretedOption.NamePart.decode(reader, reader.uint32()));
                        break;
                    case 3:
                        message.identifier_value = reader.string();
                        break;
                    case 4:
                        message.positive_int_value = reader.uint64();
                        break;
                    case 5:
                        message.negative_int_value = reader.int64();
                        break;
                    case 6:
                        message.double_value = reader.double();
                        break;
                    case 7:
                        message.string_value = reader.bytes();
                        break;
                    case 8:
                        message.aggregate_value = reader.string();
                        break;
                    default:
                        reader.skipType(tag & 7);
                        break;
                    }
                }
                return message;
            };

            UninterpretedOption.NamePart = (function() {

                /**
                 * Properties of a NamePart.
                 * @memberof google.protobuf.UninterpretedOption
                 * @interface INamePart
                 * @property {string} name_part NamePart name_part
                 * @property {boolean} is_extension NamePart is_extension
                 */

                /**
                 * Constructs a new NamePart.
                 * @memberof google.protobuf.UninterpretedOption
                 * @classdesc Represents a NamePart.
                 * @implements INamePart
                 * @constructor
                 * @param {google.protobuf.UninterpretedOption.INamePart=} [properties] Properties to set
                 */
                function NamePart(properties) {
                    if (properties)
                        for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                            if (properties[keys[i]] != null)
                                this[keys[i]] = properties[keys[i]];
                }

                /**
                 * NamePart name_part.
                 * @member {string} name_part
                 * @memberof google.protobuf.UninterpretedOption.NamePart
                 * @instance
                 */
                NamePart.prototype.name_part = "";

                /**
                 * NamePart is_extension.
                 * @member {boolean} is_extension
                 * @memberof google.protobuf.UninterpretedOption.NamePart
                 * @instance
                 */
                NamePart.prototype.is_extension = false;

                /**
                 * Encodes the specified NamePart message. Does not implicitly {@link google.protobuf.UninterpretedOption.NamePart.verify|verify} messages.
                 * @function encode
                 * @memberof google.protobuf.UninterpretedOption.NamePart
                 * @static
                 * @param {google.protobuf.UninterpretedOption.NamePart} message NamePart message or plain object to encode
                 * @param {$protobuf.Writer} [writer] Writer to encode to
                 * @returns {$protobuf.Writer} Writer
                 */
                NamePart.encode = function encode(message, writer) {
                    if (!writer)
                        writer = $Writer.create();
                    writer.uint32(/* id 1, wireType 2 =*/10).string(message.name_part);
                    writer.uint32(/* id 2, wireType 0 =*/16).bool(message.is_extension);
                    return writer;
                };

                /**
                 * Decodes a NamePart message from the specified reader or buffer.
                 * @function decode
                 * @memberof google.protobuf.UninterpretedOption.NamePart
                 * @static
                 * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
                 * @param {number} [length] Message length if known beforehand
                 * @returns {google.protobuf.UninterpretedOption.NamePart} NamePart
                 * @throws {Error} If the payload is not a reader or valid buffer
                 * @throws {$protobuf.util.ProtocolError} If required fields are missing
                 */
                NamePart.decode = function decode(reader, length) {
                    if (!(reader instanceof $Reader))
                        reader = $Reader.create(reader);
                    let end = length === undefined ? reader.len : reader.pos + length, message = new $root.google.protobuf.UninterpretedOption.NamePart();
                    while (reader.pos < end) {
                        let tag = reader.uint32();
                        switch (tag >>> 3) {
                        case 1:
                            message.name_part = reader.string();
                            break;
                        case 2:
                            message.is_extension = reader.bool();
                            break;
                        default:
                            reader.skipType(tag & 7);
                            break;
                        }
                    }
                    if (!message.hasOwnProperty("name_part"))
                        throw $util.ProtocolError("missing required 'name_part'", { instance: message });
                    if (!message.hasOwnProperty("is_extension"))
                        throw $util.ProtocolError("missing required 'is_extension'", { instance: message });
                    return message;
                };

                return NamePart;
            })();

            return UninterpretedOption;
        })();

        protobuf.SourceCodeInfo = (function() {

            /**
             * Properties of a SourceCodeInfo.
             * @memberof google.protobuf
             * @interface ISourceCodeInfo
             * @property {Array.<google.protobuf.SourceCodeInfo.Location>|null} [location] SourceCodeInfo location
             */

            /**
             * Constructs a new SourceCodeInfo.
             * @memberof google.protobuf
             * @classdesc Represents a SourceCodeInfo.
             * @implements ISourceCodeInfo
             * @constructor
             * @param {google.protobuf.ISourceCodeInfo=} [properties] Properties to set
             */
            function SourceCodeInfo(properties) {
                this.location = [];
                if (properties)
                    for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                        if (properties[keys[i]] != null)
                            this[keys[i]] = properties[keys[i]];
            }

            /**
             * SourceCodeInfo location.
             * @member {Array.<google.protobuf.SourceCodeInfo.Location>} location
             * @memberof google.protobuf.SourceCodeInfo
             * @instance
             */
            SourceCodeInfo.prototype.location = $util.emptyArray;

            /**
             * Encodes the specified SourceCodeInfo message. Does not implicitly {@link google.protobuf.SourceCodeInfo.verify|verify} messages.
             * @function encode
             * @memberof google.protobuf.SourceCodeInfo
             * @static
             * @param {google.protobuf.SourceCodeInfo} message SourceCodeInfo message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            SourceCodeInfo.encode = function encode(message, writer) {
                if (!writer)
                    writer = $Writer.create();
                if (message.location != null && message.location.length)
                    for (let i = 0; i < message.location.length; ++i)
                        $root.google.protobuf.SourceCodeInfo.Location.encode(message.location[i], writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
                return writer;
            };

            /**
             * Decodes a SourceCodeInfo message from the specified reader or buffer.
             * @function decode
             * @memberof google.protobuf.SourceCodeInfo
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @param {number} [length] Message length if known beforehand
             * @returns {google.protobuf.SourceCodeInfo} SourceCodeInfo
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            SourceCodeInfo.decode = function decode(reader, length) {
                if (!(reader instanceof $Reader))
                    reader = $Reader.create(reader);
                let end = length === undefined ? reader.len : reader.pos + length, message = new $root.google.protobuf.SourceCodeInfo();
                while (reader.pos < end) {
                    let tag = reader.uint32();
                    switch (tag >>> 3) {
                    case 1:
                        if (!(message.location && message.location.length))
                            message.location = [];
                        message.location.push($root.google.protobuf.SourceCodeInfo.Location.decode(reader, reader.uint32()));
                        break;
                    default:
                        reader.skipType(tag & 7);
                        break;
                    }
                }
                return message;
            };

            SourceCodeInfo.Location = (function() {

                /**
                 * Properties of a Location.
                 * @memberof google.protobuf.SourceCodeInfo
                 * @interface ILocation
                 * @property {Array.<number>|null} [path] Location path
                 * @property {Array.<number>|null} [span] Location span
                 * @property {string|null} [leading_comments] Location leading_comments
                 * @property {string|null} [trailing_comments] Location trailing_comments
                 * @property {Array.<string>|null} [leading_detached_comments] Location leading_detached_comments
                 */

                /**
                 * Constructs a new Location.
                 * @memberof google.protobuf.SourceCodeInfo
                 * @classdesc Represents a Location.
                 * @implements ILocation
                 * @constructor
                 * @param {google.protobuf.SourceCodeInfo.ILocation=} [properties] Properties to set
                 */
                function Location(properties) {
                    this.path = [];
                    this.span = [];
                    this.leading_detached_comments = [];
                    if (properties)
                        for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                            if (properties[keys[i]] != null)
                                this[keys[i]] = properties[keys[i]];
                }

                /**
                 * Location path.
                 * @member {Array.<number>} path
                 * @memberof google.protobuf.SourceCodeInfo.Location
                 * @instance
                 */
                Location.prototype.path = $util.emptyArray;

                /**
                 * Location span.
                 * @member {Array.<number>} span
                 * @memberof google.protobuf.SourceCodeInfo.Location
                 * @instance
                 */
                Location.prototype.span = $util.emptyArray;

                /**
                 * Location leading_comments.
                 * @member {string} leading_comments
                 * @memberof google.protobuf.SourceCodeInfo.Location
                 * @instance
                 */
                Location.prototype.leading_comments = "";

                /**
                 * Location trailing_comments.
                 * @member {string} trailing_comments
                 * @memberof google.protobuf.SourceCodeInfo.Location
                 * @instance
                 */
                Location.prototype.trailing_comments = "";

                /**
                 * Location leading_detached_comments.
                 * @member {Array.<string>} leading_detached_comments
                 * @memberof google.protobuf.SourceCodeInfo.Location
                 * @instance
                 */
                Location.prototype.leading_detached_comments = $util.emptyArray;

                /**
                 * Encodes the specified Location message. Does not implicitly {@link google.protobuf.SourceCodeInfo.Location.verify|verify} messages.
                 * @function encode
                 * @memberof google.protobuf.SourceCodeInfo.Location
                 * @static
                 * @param {google.protobuf.SourceCodeInfo.Location} message Location message or plain object to encode
                 * @param {$protobuf.Writer} [writer] Writer to encode to
                 * @returns {$protobuf.Writer} Writer
                 */
                Location.encode = function encode(message, writer) {
                    if (!writer)
                        writer = $Writer.create();
                    if (message.path != null && message.path.length) {
                        writer.uint32(/* id 1, wireType 2 =*/10).fork();
                        for (let i = 0; i < message.path.length; ++i)
                            writer.int32(message.path[i]);
                        writer.ldelim();
                    }
                    if (message.span != null && message.span.length) {
                        writer.uint32(/* id 2, wireType 2 =*/18).fork();
                        for (let i = 0; i < message.span.length; ++i)
                            writer.int32(message.span[i]);
                        writer.ldelim();
                    }
                    if (message.leading_comments != null && Object.hasOwnProperty.call(message, "leading_comments"))
                        writer.uint32(/* id 3, wireType 2 =*/26).string(message.leading_comments);
                    if (message.trailing_comments != null && Object.hasOwnProperty.call(message, "trailing_comments"))
                        writer.uint32(/* id 4, wireType 2 =*/34).string(message.trailing_comments);
                    if (message.leading_detached_comments != null && message.leading_detached_comments.length)
                        for (let i = 0; i < message.leading_detached_comments.length; ++i)
                            writer.uint32(/* id 6, wireType 2 =*/50).string(message.leading_detached_comments[i]);
                    return writer;
                };

                /**
                 * Decodes a Location message from the specified reader or buffer.
                 * @function decode
                 * @memberof google.protobuf.SourceCodeInfo.Location
                 * @static
                 * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
                 * @param {number} [length] Message length if known beforehand
                 * @returns {google.protobuf.SourceCodeInfo.Location} Location
                 * @throws {Error} If the payload is not a reader or valid buffer
                 * @throws {$protobuf.util.ProtocolError} If required fields are missing
                 */
                Location.decode = function decode(reader, length) {
                    if (!(reader instanceof $Reader))
                        reader = $Reader.create(reader);
                    let end = length === undefined ? reader.len : reader.pos + length, message = new $root.google.protobuf.SourceCodeInfo.Location();
                    while (reader.pos < end) {
                        let tag = reader.uint32();
                        switch (tag >>> 3) {
                        case 1:
                            if (!(message.path && message.path.length))
                                message.path = [];
                            if ((tag & 7) === 2) {
                                let end2 = reader.uint32() + reader.pos;
                                while (reader.pos < end2)
                                    message.path.push(reader.int32());
                            } else
                                message.path.push(reader.int32());
                            break;
                        case 2:
                            if (!(message.span && message.span.length))
                                message.span = [];
                            if ((tag & 7) === 2) {
                                let end2 = reader.uint32() + reader.pos;
                                while (reader.pos < end2)
                                    message.span.push(reader.int32());
                            } else
                                message.span.push(reader.int32());
                            break;
                        case 3:
                            message.leading_comments = reader.string();
                            break;
                        case 4:
                            message.trailing_comments = reader.string();
                            break;
                        case 6:
                            if (!(message.leading_detached_comments && message.leading_detached_comments.length))
                                message.leading_detached_comments = [];
                            message.leading_detached_comments.push(reader.string());
                            break;
                        default:
                            reader.skipType(tag & 7);
                            break;
                        }
                    }
                    return message;
                };

                return Location;
            })();

            return SourceCodeInfo;
        })();

        protobuf.GeneratedCodeInfo = (function() {

            /**
             * Properties of a GeneratedCodeInfo.
             * @memberof google.protobuf
             * @interface IGeneratedCodeInfo
             * @property {Array.<google.protobuf.GeneratedCodeInfo.Annotation>|null} [annotation] GeneratedCodeInfo annotation
             */

            /**
             * Constructs a new GeneratedCodeInfo.
             * @memberof google.protobuf
             * @classdesc Represents a GeneratedCodeInfo.
             * @implements IGeneratedCodeInfo
             * @constructor
             * @param {google.protobuf.IGeneratedCodeInfo=} [properties] Properties to set
             */
            function GeneratedCodeInfo(properties) {
                this.annotation = [];
                if (properties)
                    for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                        if (properties[keys[i]] != null)
                            this[keys[i]] = properties[keys[i]];
            }

            /**
             * GeneratedCodeInfo annotation.
             * @member {Array.<google.protobuf.GeneratedCodeInfo.Annotation>} annotation
             * @memberof google.protobuf.GeneratedCodeInfo
             * @instance
             */
            GeneratedCodeInfo.prototype.annotation = $util.emptyArray;

            /**
             * Encodes the specified GeneratedCodeInfo message. Does not implicitly {@link google.protobuf.GeneratedCodeInfo.verify|verify} messages.
             * @function encode
             * @memberof google.protobuf.GeneratedCodeInfo
             * @static
             * @param {google.protobuf.GeneratedCodeInfo} message GeneratedCodeInfo message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            GeneratedCodeInfo.encode = function encode(message, writer) {
                if (!writer)
                    writer = $Writer.create();
                if (message.annotation != null && message.annotation.length)
                    for (let i = 0; i < message.annotation.length; ++i)
                        $root.google.protobuf.GeneratedCodeInfo.Annotation.encode(message.annotation[i], writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
                return writer;
            };

            /**
             * Decodes a GeneratedCodeInfo message from the specified reader or buffer.
             * @function decode
             * @memberof google.protobuf.GeneratedCodeInfo
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @param {number} [length] Message length if known beforehand
             * @returns {google.protobuf.GeneratedCodeInfo} GeneratedCodeInfo
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            GeneratedCodeInfo.decode = function decode(reader, length) {
                if (!(reader instanceof $Reader))
                    reader = $Reader.create(reader);
                let end = length === undefined ? reader.len : reader.pos + length, message = new $root.google.protobuf.GeneratedCodeInfo();
                while (reader.pos < end) {
                    let tag = reader.uint32();
                    switch (tag >>> 3) {
                    case 1:
                        if (!(message.annotation && message.annotation.length))
                            message.annotation = [];
                        message.annotation.push($root.google.protobuf.GeneratedCodeInfo.Annotation.decode(reader, reader.uint32()));
                        break;
                    default:
                        reader.skipType(tag & 7);
                        break;
                    }
                }
                return message;
            };

            GeneratedCodeInfo.Annotation = (function() {

                /**
                 * Properties of an Annotation.
                 * @memberof google.protobuf.GeneratedCodeInfo
                 * @interface IAnnotation
                 * @property {Array.<number>|null} [path] Annotation path
                 * @property {string|null} [source_file] Annotation source_file
                 * @property {number|null} [begin] Annotation begin
                 * @property {number|null} [end] Annotation end
                 */

                /**
                 * Constructs a new Annotation.
                 * @memberof google.protobuf.GeneratedCodeInfo
                 * @classdesc Represents an Annotation.
                 * @implements IAnnotation
                 * @constructor
                 * @param {google.protobuf.GeneratedCodeInfo.IAnnotation=} [properties] Properties to set
                 */
                function Annotation(properties) {
                    this.path = [];
                    if (properties)
                        for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                            if (properties[keys[i]] != null)
                                this[keys[i]] = properties[keys[i]];
                }

                /**
                 * Annotation path.
                 * @member {Array.<number>} path
                 * @memberof google.protobuf.GeneratedCodeInfo.Annotation
                 * @instance
                 */
                Annotation.prototype.path = $util.emptyArray;

                /**
                 * Annotation source_file.
                 * @member {string} source_file
                 * @memberof google.protobuf.GeneratedCodeInfo.Annotation
                 * @instance
                 */
                Annotation.prototype.source_file = "";

                /**
                 * Annotation begin.
                 * @member {number} begin
                 * @memberof google.protobuf.GeneratedCodeInfo.Annotation
                 * @instance
                 */
                Annotation.prototype.begin = 0;

                /**
                 * Annotation end.
                 * @member {number} end
                 * @memberof google.protobuf.GeneratedCodeInfo.Annotation
                 * @instance
                 */
                Annotation.prototype.end = 0;

                /**
                 * Encodes the specified Annotation message. Does not implicitly {@link google.protobuf.GeneratedCodeInfo.Annotation.verify|verify} messages.
                 * @function encode
                 * @memberof google.protobuf.GeneratedCodeInfo.Annotation
                 * @static
                 * @param {google.protobuf.GeneratedCodeInfo.Annotation} message Annotation message or plain object to encode
                 * @param {$protobuf.Writer} [writer] Writer to encode to
                 * @returns {$protobuf.Writer} Writer
                 */
                Annotation.encode = function encode(message, writer) {
                    if (!writer)
                        writer = $Writer.create();
                    if (message.path != null && message.path.length) {
                        writer.uint32(/* id 1, wireType 2 =*/10).fork();
                        for (let i = 0; i < message.path.length; ++i)
                            writer.int32(message.path[i]);
                        writer.ldelim();
                    }
                    if (message.source_file != null && Object.hasOwnProperty.call(message, "source_file"))
                        writer.uint32(/* id 2, wireType 2 =*/18).string(message.source_file);
                    if (message.begin != null && Object.hasOwnProperty.call(message, "begin"))
                        writer.uint32(/* id 3, wireType 0 =*/24).int32(message.begin);
                    if (message.end != null && Object.hasOwnProperty.call(message, "end"))
                        writer.uint32(/* id 4, wireType 0 =*/32).int32(message.end);
                    return writer;
                };

                /**
                 * Decodes an Annotation message from the specified reader or buffer.
                 * @function decode
                 * @memberof google.protobuf.GeneratedCodeInfo.Annotation
                 * @static
                 * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
                 * @param {number} [length] Message length if known beforehand
                 * @returns {google.protobuf.GeneratedCodeInfo.Annotation} Annotation
                 * @throws {Error} If the payload is not a reader or valid buffer
                 * @throws {$protobuf.util.ProtocolError} If required fields are missing
                 */
                Annotation.decode = function decode(reader, length) {
                    if (!(reader instanceof $Reader))
                        reader = $Reader.create(reader);
                    let end = length === undefined ? reader.len : reader.pos + length, message = new $root.google.protobuf.GeneratedCodeInfo.Annotation();
                    while (reader.pos < end) {
                        let tag = reader.uint32();
                        switch (tag >>> 3) {
                        case 1:
                            if (!(message.path && message.path.length))
                                message.path = [];
                            if ((tag & 7) === 2) {
                                let end2 = reader.uint32() + reader.pos;
                                while (reader.pos < end2)
                                    message.path.push(reader.int32());
                            } else
                                message.path.push(reader.int32());
                            break;
                        case 2:
                            message.source_file = reader.string();
                            break;
                        case 3:
                            message.begin = reader.int32();
                            break;
                        case 4:
                            message.end = reader.int32();
                            break;
                        default:
                            reader.skipType(tag & 7);
                            break;
                        }
                    }
                    return message;
                };

                return Annotation;
            })();

            return GeneratedCodeInfo;
        })();

        protobuf.Timestamp = (function() {

            /**
             * Properties of a Timestamp.
             * @memberof google.protobuf
             * @interface ITimestamp
             * @property {number|null} [seconds] Timestamp seconds
             * @property {number|null} [nanos] Timestamp nanos
             */

            /**
             * Constructs a new Timestamp.
             * @memberof google.protobuf
             * @classdesc Represents a Timestamp.
             * @implements ITimestamp
             * @constructor
             * @param {google.protobuf.ITimestamp=} [properties] Properties to set
             */
            function Timestamp(properties) {
                if (properties)
                    for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                        if (properties[keys[i]] != null)
                            this[keys[i]] = properties[keys[i]];
            }

            /**
             * Timestamp seconds.
             * @member {number} seconds
             * @memberof google.protobuf.Timestamp
             * @instance
             */
            Timestamp.prototype.seconds = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

            /**
             * Timestamp nanos.
             * @member {number} nanos
             * @memberof google.protobuf.Timestamp
             * @instance
             */
            Timestamp.prototype.nanos = 0;

            /**
             * Encodes the specified Timestamp message. Does not implicitly {@link google.protobuf.Timestamp.verify|verify} messages.
             * @function encode
             * @memberof google.protobuf.Timestamp
             * @static
             * @param {google.protobuf.Timestamp} message Timestamp message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            Timestamp.encode = function encode(message, writer) {
                if (!writer)
                    writer = $Writer.create();
                if (message.seconds != null && Object.hasOwnProperty.call(message, "seconds"))
                    writer.uint32(/* id 1, wireType 0 =*/8).int64(message.seconds);
                if (message.nanos != null && Object.hasOwnProperty.call(message, "nanos"))
                    writer.uint32(/* id 2, wireType 0 =*/16).int32(message.nanos);
                return writer;
            };

            /**
             * Decodes a Timestamp message from the specified reader or buffer.
             * @function decode
             * @memberof google.protobuf.Timestamp
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @param {number} [length] Message length if known beforehand
             * @returns {google.protobuf.Timestamp} Timestamp
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            Timestamp.decode = function decode(reader, length) {
                if (!(reader instanceof $Reader))
                    reader = $Reader.create(reader);
                let end = length === undefined ? reader.len : reader.pos + length, message = new $root.google.protobuf.Timestamp();
                while (reader.pos < end) {
                    let tag = reader.uint32();
                    switch (tag >>> 3) {
                    case 1:
                        message.seconds = reader.int64();
                        break;
                    case 2:
                        message.nanos = reader.int32();
                        break;
                    default:
                        reader.skipType(tag & 7);
                        break;
                    }
                }
                return message;
            };

            return Timestamp;
        })();

        return protobuf;
    })();

    google.api = (function() {

        /**
         * Namespace api.
         * @memberof google
         * @namespace
         */
        const api = {};

        api.Http = (function() {

            /**
             * Properties of a Http.
             * @memberof google.api
             * @interface IHttp
             * @property {Array.<google.api.HttpRule>|null} [rules] Http rules
             */

            /**
             * Constructs a new Http.
             * @memberof google.api
             * @classdesc Represents a Http.
             * @implements IHttp
             * @constructor
             * @param {google.api.IHttp=} [properties] Properties to set
             */
            function Http(properties) {
                this.rules = [];
                if (properties)
                    for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                        if (properties[keys[i]] != null)
                            this[keys[i]] = properties[keys[i]];
            }

            /**
             * Http rules.
             * @member {Array.<google.api.HttpRule>} rules
             * @memberof google.api.Http
             * @instance
             */
            Http.prototype.rules = $util.emptyArray;

            /**
             * Encodes the specified Http message. Does not implicitly {@link google.api.Http.verify|verify} messages.
             * @function encode
             * @memberof google.api.Http
             * @static
             * @param {google.api.Http} message Http message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            Http.encode = function encode(message, writer) {
                if (!writer)
                    writer = $Writer.create();
                if (message.rules != null && message.rules.length)
                    for (let i = 0; i < message.rules.length; ++i)
                        $root.google.api.HttpRule.encode(message.rules[i], writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
                return writer;
            };

            /**
             * Decodes a Http message from the specified reader or buffer.
             * @function decode
             * @memberof google.api.Http
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @param {number} [length] Message length if known beforehand
             * @returns {google.api.Http} Http
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            Http.decode = function decode(reader, length) {
                if (!(reader instanceof $Reader))
                    reader = $Reader.create(reader);
                let end = length === undefined ? reader.len : reader.pos + length, message = new $root.google.api.Http();
                while (reader.pos < end) {
                    let tag = reader.uint32();
                    switch (tag >>> 3) {
                    case 1:
                        if (!(message.rules && message.rules.length))
                            message.rules = [];
                        message.rules.push($root.google.api.HttpRule.decode(reader, reader.uint32()));
                        break;
                    default:
                        reader.skipType(tag & 7);
                        break;
                    }
                }
                return message;
            };

            return Http;
        })();

        api.HttpRule = (function() {

            /**
             * Properties of a HttpRule.
             * @memberof google.api
             * @interface IHttpRule
             * @property {string|null} [get] HttpRule get
             * @property {string|null} [put] HttpRule put
             * @property {string|null} [post] HttpRule post
             * @property {string|null} ["delete"] HttpRule delete
             * @property {string|null} [patch] HttpRule patch
             * @property {google.api.CustomHttpPattern|null} [custom] HttpRule custom
             * @property {string|null} [selector] HttpRule selector
             * @property {string|null} [body] HttpRule body
             * @property {Array.<google.api.HttpRule>|null} [additional_bindings] HttpRule additional_bindings
             */

            /**
             * Constructs a new HttpRule.
             * @memberof google.api
             * @classdesc Represents a HttpRule.
             * @implements IHttpRule
             * @constructor
             * @param {google.api.IHttpRule=} [properties] Properties to set
             */
            function HttpRule(properties) {
                this.additional_bindings = [];
                if (properties)
                    for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                        if (properties[keys[i]] != null)
                            this[keys[i]] = properties[keys[i]];
            }

            /**
             * HttpRule get.
             * @member {string|null|undefined} get
             * @memberof google.api.HttpRule
             * @instance
             */
            HttpRule.prototype.get = null;

            /**
             * HttpRule put.
             * @member {string|null|undefined} put
             * @memberof google.api.HttpRule
             * @instance
             */
            HttpRule.prototype.put = null;

            /**
             * HttpRule post.
             * @member {string|null|undefined} post
             * @memberof google.api.HttpRule
             * @instance
             */
            HttpRule.prototype.post = null;

            /**
             * HttpRule delete.
             * @member {string|null|undefined} delete
             * @memberof google.api.HttpRule
             * @instance
             */
            HttpRule.prototype["delete"] = null;

            /**
             * HttpRule patch.
             * @member {string|null|undefined} patch
             * @memberof google.api.HttpRule
             * @instance
             */
            HttpRule.prototype.patch = null;

            /**
             * HttpRule custom.
             * @member {google.api.CustomHttpPattern|null|undefined} custom
             * @memberof google.api.HttpRule
             * @instance
             */
            HttpRule.prototype.custom = null;

            /**
             * HttpRule selector.
             * @member {string} selector
             * @memberof google.api.HttpRule
             * @instance
             */
            HttpRule.prototype.selector = "";

            /**
             * HttpRule body.
             * @member {string} body
             * @memberof google.api.HttpRule
             * @instance
             */
            HttpRule.prototype.body = "";

            /**
             * HttpRule additional_bindings.
             * @member {Array.<google.api.HttpRule>} additional_bindings
             * @memberof google.api.HttpRule
             * @instance
             */
            HttpRule.prototype.additional_bindings = $util.emptyArray;

            // OneOf field names bound to virtual getters and setters
            let $oneOfFields;

            /**
             * HttpRule pattern.
             * @member {"get"|"put"|"post"|"delete"|"patch"|"custom"|undefined} pattern
             * @memberof google.api.HttpRule
             * @instance
             */
            Object.defineProperty(HttpRule.prototype, "pattern", {
                get: $util.oneOfGetter($oneOfFields = ["get", "put", "post", "delete", "patch", "custom"]),
                set: $util.oneOfSetter($oneOfFields)
            });

            /**
             * Encodes the specified HttpRule message. Does not implicitly {@link google.api.HttpRule.verify|verify} messages.
             * @function encode
             * @memberof google.api.HttpRule
             * @static
             * @param {google.api.HttpRule} message HttpRule message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            HttpRule.encode = function encode(message, writer) {
                if (!writer)
                    writer = $Writer.create();
                if (message.selector != null && Object.hasOwnProperty.call(message, "selector"))
                    writer.uint32(/* id 1, wireType 2 =*/10).string(message.selector);
                if (message.get != null && Object.hasOwnProperty.call(message, "get"))
                    writer.uint32(/* id 2, wireType 2 =*/18).string(message.get);
                if (message.put != null && Object.hasOwnProperty.call(message, "put"))
                    writer.uint32(/* id 3, wireType 2 =*/26).string(message.put);
                if (message.post != null && Object.hasOwnProperty.call(message, "post"))
                    writer.uint32(/* id 4, wireType 2 =*/34).string(message.post);
                if (message["delete"] != null && Object.hasOwnProperty.call(message, "delete"))
                    writer.uint32(/* id 5, wireType 2 =*/42).string(message["delete"]);
                if (message.patch != null && Object.hasOwnProperty.call(message, "patch"))
                    writer.uint32(/* id 6, wireType 2 =*/50).string(message.patch);
                if (message.body != null && Object.hasOwnProperty.call(message, "body"))
                    writer.uint32(/* id 7, wireType 2 =*/58).string(message.body);
                if (message.custom != null && Object.hasOwnProperty.call(message, "custom"))
                    $root.google.api.CustomHttpPattern.encode(message.custom, writer.uint32(/* id 8, wireType 2 =*/66).fork()).ldelim();
                if (message.additional_bindings != null && message.additional_bindings.length)
                    for (let i = 0; i < message.additional_bindings.length; ++i)
                        $root.google.api.HttpRule.encode(message.additional_bindings[i], writer.uint32(/* id 11, wireType 2 =*/90).fork()).ldelim();
                return writer;
            };

            /**
             * Decodes a HttpRule message from the specified reader or buffer.
             * @function decode
             * @memberof google.api.HttpRule
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @param {number} [length] Message length if known beforehand
             * @returns {google.api.HttpRule} HttpRule
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            HttpRule.decode = function decode(reader, length) {
                if (!(reader instanceof $Reader))
                    reader = $Reader.create(reader);
                let end = length === undefined ? reader.len : reader.pos + length, message = new $root.google.api.HttpRule();
                while (reader.pos < end) {
                    let tag = reader.uint32();
                    switch (tag >>> 3) {
                    case 2:
                        message.get = reader.string();
                        break;
                    case 3:
                        message.put = reader.string();
                        break;
                    case 4:
                        message.post = reader.string();
                        break;
                    case 5:
                        message["delete"] = reader.string();
                        break;
                    case 6:
                        message.patch = reader.string();
                        break;
                    case 8:
                        message.custom = $root.google.api.CustomHttpPattern.decode(reader, reader.uint32());
                        break;
                    case 1:
                        message.selector = reader.string();
                        break;
                    case 7:
                        message.body = reader.string();
                        break;
                    case 11:
                        if (!(message.additional_bindings && message.additional_bindings.length))
                            message.additional_bindings = [];
                        message.additional_bindings.push($root.google.api.HttpRule.decode(reader, reader.uint32()));
                        break;
                    default:
                        reader.skipType(tag & 7);
                        break;
                    }
                }
                return message;
            };

            return HttpRule;
        })();

        api.CustomHttpPattern = (function() {

            /**
             * Properties of a CustomHttpPattern.
             * @memberof google.api
             * @interface ICustomHttpPattern
             * @property {string|null} [kind] CustomHttpPattern kind
             * @property {string|null} [path] CustomHttpPattern path
             */

            /**
             * Constructs a new CustomHttpPattern.
             * @memberof google.api
             * @classdesc Represents a CustomHttpPattern.
             * @implements ICustomHttpPattern
             * @constructor
             * @param {google.api.ICustomHttpPattern=} [properties] Properties to set
             */
            function CustomHttpPattern(properties) {
                if (properties)
                    for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                        if (properties[keys[i]] != null)
                            this[keys[i]] = properties[keys[i]];
            }

            /**
             * CustomHttpPattern kind.
             * @member {string} kind
             * @memberof google.api.CustomHttpPattern
             * @instance
             */
            CustomHttpPattern.prototype.kind = "";

            /**
             * CustomHttpPattern path.
             * @member {string} path
             * @memberof google.api.CustomHttpPattern
             * @instance
             */
            CustomHttpPattern.prototype.path = "";

            /**
             * Encodes the specified CustomHttpPattern message. Does not implicitly {@link google.api.CustomHttpPattern.verify|verify} messages.
             * @function encode
             * @memberof google.api.CustomHttpPattern
             * @static
             * @param {google.api.CustomHttpPattern} message CustomHttpPattern message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            CustomHttpPattern.encode = function encode(message, writer) {
                if (!writer)
                    writer = $Writer.create();
                if (message.kind != null && Object.hasOwnProperty.call(message, "kind"))
                    writer.uint32(/* id 1, wireType 2 =*/10).string(message.kind);
                if (message.path != null && Object.hasOwnProperty.call(message, "path"))
                    writer.uint32(/* id 2, wireType 2 =*/18).string(message.path);
                return writer;
            };

            /**
             * Decodes a CustomHttpPattern message from the specified reader or buffer.
             * @function decode
             * @memberof google.api.CustomHttpPattern
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @param {number} [length] Message length if known beforehand
             * @returns {google.api.CustomHttpPattern} CustomHttpPattern
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            CustomHttpPattern.decode = function decode(reader, length) {
                if (!(reader instanceof $Reader))
                    reader = $Reader.create(reader);
                let end = length === undefined ? reader.len : reader.pos + length, message = new $root.google.api.CustomHttpPattern();
                while (reader.pos < end) {
                    let tag = reader.uint32();
                    switch (tag >>> 3) {
                    case 1:
                        message.kind = reader.string();
                        break;
                    case 2:
                        message.path = reader.string();
                        break;
                    default:
                        reader.skipType(tag & 7);
                        break;
                    }
                }
                return message;
            };

            return CustomHttpPattern;
        })();

        return api;
    })();

    return google;
})();

export const NamespaceID = $root.NamespaceID = (() => {

    /**
     * Properties of a NamespaceID.
     * @exports INamespaceID
     * @interface INamespaceID
     * @property {number|null} [namespace_id] NamespaceID namespace_id
     */

    /**
     * Constructs a new NamespaceID.
     * @exports NamespaceID
     * @classdesc Represents a NamespaceID.
     * @implements INamespaceID
     * @constructor
     * @param {INamespaceID=} [properties] Properties to set
     */
    function NamespaceID(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * NamespaceID namespace_id.
     * @member {number} namespace_id
     * @memberof NamespaceID
     * @instance
     */
    NamespaceID.prototype.namespace_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * Encodes the specified NamespaceID message. Does not implicitly {@link NamespaceID.verify|verify} messages.
     * @function encode
     * @memberof NamespaceID
     * @static
     * @param {NamespaceID} message NamespaceID message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    NamespaceID.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.namespace_id != null && Object.hasOwnProperty.call(message, "namespace_id"))
            writer.uint32(/* id 1, wireType 0 =*/8).int64(message.namespace_id);
        return writer;
    };

    /**
     * Decodes a NamespaceID message from the specified reader or buffer.
     * @function decode
     * @memberof NamespaceID
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {NamespaceID} NamespaceID
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    NamespaceID.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.NamespaceID();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.namespace_id = reader.int64();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return NamespaceID;
})();

export const NamespaceResponse = $root.NamespaceResponse = (() => {

    /**
     * Properties of a NamespaceResponse.
     * @exports INamespaceResponse
     * @interface INamespaceResponse
     * @property {number|null} [id] NamespaceResponse id
     * @property {string|null} [name] NamespaceResponse name
     * @property {Array.<string>|null} [image_pull_secrets] NamespaceResponse image_pull_secrets
     * @property {google.protobuf.Timestamp|null} [created_at] NamespaceResponse created_at
     * @property {google.protobuf.Timestamp|null} [updated_at] NamespaceResponse updated_at
     * @property {google.protobuf.Timestamp|null} [deleted_at] NamespaceResponse deleted_at
     */

    /**
     * Constructs a new NamespaceResponse.
     * @exports NamespaceResponse
     * @classdesc Represents a NamespaceResponse.
     * @implements INamespaceResponse
     * @constructor
     * @param {INamespaceResponse=} [properties] Properties to set
     */
    function NamespaceResponse(properties) {
        this.image_pull_secrets = [];
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * NamespaceResponse id.
     * @member {number} id
     * @memberof NamespaceResponse
     * @instance
     */
    NamespaceResponse.prototype.id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * NamespaceResponse name.
     * @member {string} name
     * @memberof NamespaceResponse
     * @instance
     */
    NamespaceResponse.prototype.name = "";

    /**
     * NamespaceResponse image_pull_secrets.
     * @member {Array.<string>} image_pull_secrets
     * @memberof NamespaceResponse
     * @instance
     */
    NamespaceResponse.prototype.image_pull_secrets = $util.emptyArray;

    /**
     * NamespaceResponse created_at.
     * @member {google.protobuf.Timestamp|null|undefined} created_at
     * @memberof NamespaceResponse
     * @instance
     */
    NamespaceResponse.prototype.created_at = null;

    /**
     * NamespaceResponse updated_at.
     * @member {google.protobuf.Timestamp|null|undefined} updated_at
     * @memberof NamespaceResponse
     * @instance
     */
    NamespaceResponse.prototype.updated_at = null;

    /**
     * NamespaceResponse deleted_at.
     * @member {google.protobuf.Timestamp|null|undefined} deleted_at
     * @memberof NamespaceResponse
     * @instance
     */
    NamespaceResponse.prototype.deleted_at = null;

    /**
     * Encodes the specified NamespaceResponse message. Does not implicitly {@link NamespaceResponse.verify|verify} messages.
     * @function encode
     * @memberof NamespaceResponse
     * @static
     * @param {NamespaceResponse} message NamespaceResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    NamespaceResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.id != null && Object.hasOwnProperty.call(message, "id"))
            writer.uint32(/* id 1, wireType 0 =*/8).int64(message.id);
        if (message.name != null && Object.hasOwnProperty.call(message, "name"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.name);
        if (message.image_pull_secrets != null && message.image_pull_secrets.length)
            for (let i = 0; i < message.image_pull_secrets.length; ++i)
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.image_pull_secrets[i]);
        if (message.created_at != null && Object.hasOwnProperty.call(message, "created_at"))
            $root.google.protobuf.Timestamp.encode(message.created_at, writer.uint32(/* id 4, wireType 2 =*/34).fork()).ldelim();
        if (message.updated_at != null && Object.hasOwnProperty.call(message, "updated_at"))
            $root.google.protobuf.Timestamp.encode(message.updated_at, writer.uint32(/* id 5, wireType 2 =*/42).fork()).ldelim();
        if (message.deleted_at != null && Object.hasOwnProperty.call(message, "deleted_at"))
            $root.google.protobuf.Timestamp.encode(message.deleted_at, writer.uint32(/* id 6, wireType 2 =*/50).fork()).ldelim();
        return writer;
    };

    /**
     * Decodes a NamespaceResponse message from the specified reader or buffer.
     * @function decode
     * @memberof NamespaceResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {NamespaceResponse} NamespaceResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    NamespaceResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.NamespaceResponse();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.id = reader.int64();
                break;
            case 2:
                message.name = reader.string();
                break;
            case 3:
                if (!(message.image_pull_secrets && message.image_pull_secrets.length))
                    message.image_pull_secrets = [];
                message.image_pull_secrets.push(reader.string());
                break;
            case 4:
                message.created_at = $root.google.protobuf.Timestamp.decode(reader, reader.uint32());
                break;
            case 5:
                message.updated_at = $root.google.protobuf.Timestamp.decode(reader, reader.uint32());
                break;
            case 6:
                message.deleted_at = $root.google.protobuf.Timestamp.decode(reader, reader.uint32());
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return NamespaceResponse;
})();

export const NamespaceItem = $root.NamespaceItem = (() => {

    /**
     * Properties of a NamespaceItem.
     * @exports INamespaceItem
     * @interface INamespaceItem
     * @property {number|null} [id] NamespaceItem id
     * @property {string|null} [name] NamespaceItem name
     * @property {google.protobuf.Timestamp|null} [created_at] NamespaceItem created_at
     * @property {google.protobuf.Timestamp|null} [updated_at] NamespaceItem updated_at
     * @property {Array.<NamespaceItem.SimpleProjectItem>|null} [projects] NamespaceItem projects
     */

    /**
     * Constructs a new NamespaceItem.
     * @exports NamespaceItem
     * @classdesc Represents a NamespaceItem.
     * @implements INamespaceItem
     * @constructor
     * @param {INamespaceItem=} [properties] Properties to set
     */
    function NamespaceItem(properties) {
        this.projects = [];
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * NamespaceItem id.
     * @member {number} id
     * @memberof NamespaceItem
     * @instance
     */
    NamespaceItem.prototype.id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * NamespaceItem name.
     * @member {string} name
     * @memberof NamespaceItem
     * @instance
     */
    NamespaceItem.prototype.name = "";

    /**
     * NamespaceItem created_at.
     * @member {google.protobuf.Timestamp|null|undefined} created_at
     * @memberof NamespaceItem
     * @instance
     */
    NamespaceItem.prototype.created_at = null;

    /**
     * NamespaceItem updated_at.
     * @member {google.protobuf.Timestamp|null|undefined} updated_at
     * @memberof NamespaceItem
     * @instance
     */
    NamespaceItem.prototype.updated_at = null;

    /**
     * NamespaceItem projects.
     * @member {Array.<NamespaceItem.SimpleProjectItem>} projects
     * @memberof NamespaceItem
     * @instance
     */
    NamespaceItem.prototype.projects = $util.emptyArray;

    /**
     * Encodes the specified NamespaceItem message. Does not implicitly {@link NamespaceItem.verify|verify} messages.
     * @function encode
     * @memberof NamespaceItem
     * @static
     * @param {NamespaceItem} message NamespaceItem message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    NamespaceItem.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.id != null && Object.hasOwnProperty.call(message, "id"))
            writer.uint32(/* id 1, wireType 0 =*/8).int64(message.id);
        if (message.name != null && Object.hasOwnProperty.call(message, "name"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.name);
        if (message.created_at != null && Object.hasOwnProperty.call(message, "created_at"))
            $root.google.protobuf.Timestamp.encode(message.created_at, writer.uint32(/* id 3, wireType 2 =*/26).fork()).ldelim();
        if (message.updated_at != null && Object.hasOwnProperty.call(message, "updated_at"))
            $root.google.protobuf.Timestamp.encode(message.updated_at, writer.uint32(/* id 4, wireType 2 =*/34).fork()).ldelim();
        if (message.projects != null && message.projects.length)
            for (let i = 0; i < message.projects.length; ++i)
                $root.NamespaceItem.SimpleProjectItem.encode(message.projects[i], writer.uint32(/* id 5, wireType 2 =*/42).fork()).ldelim();
        return writer;
    };

    /**
     * Decodes a NamespaceItem message from the specified reader or buffer.
     * @function decode
     * @memberof NamespaceItem
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {NamespaceItem} NamespaceItem
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    NamespaceItem.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.NamespaceItem();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.id = reader.int64();
                break;
            case 2:
                message.name = reader.string();
                break;
            case 3:
                message.created_at = $root.google.protobuf.Timestamp.decode(reader, reader.uint32());
                break;
            case 4:
                message.updated_at = $root.google.protobuf.Timestamp.decode(reader, reader.uint32());
                break;
            case 5:
                if (!(message.projects && message.projects.length))
                    message.projects = [];
                message.projects.push($root.NamespaceItem.SimpleProjectItem.decode(reader, reader.uint32()));
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    NamespaceItem.SimpleProjectItem = (function() {

        /**
         * Properties of a SimpleProjectItem.
         * @memberof NamespaceItem
         * @interface ISimpleProjectItem
         * @property {number|null} [id] SimpleProjectItem id
         * @property {string|null} [name] SimpleProjectItem name
         * @property {string|null} [status] SimpleProjectItem status
         */

        /**
         * Constructs a new SimpleProjectItem.
         * @memberof NamespaceItem
         * @classdesc Represents a SimpleProjectItem.
         * @implements ISimpleProjectItem
         * @constructor
         * @param {NamespaceItem.ISimpleProjectItem=} [properties] Properties to set
         */
        function SimpleProjectItem(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * SimpleProjectItem id.
         * @member {number} id
         * @memberof NamespaceItem.SimpleProjectItem
         * @instance
         */
        SimpleProjectItem.prototype.id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * SimpleProjectItem name.
         * @member {string} name
         * @memberof NamespaceItem.SimpleProjectItem
         * @instance
         */
        SimpleProjectItem.prototype.name = "";

        /**
         * SimpleProjectItem status.
         * @member {string} status
         * @memberof NamespaceItem.SimpleProjectItem
         * @instance
         */
        SimpleProjectItem.prototype.status = "";

        /**
         * Encodes the specified SimpleProjectItem message. Does not implicitly {@link NamespaceItem.SimpleProjectItem.verify|verify} messages.
         * @function encode
         * @memberof NamespaceItem.SimpleProjectItem
         * @static
         * @param {NamespaceItem.SimpleProjectItem} message SimpleProjectItem message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        SimpleProjectItem.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.id != null && Object.hasOwnProperty.call(message, "id"))
                writer.uint32(/* id 1, wireType 0 =*/8).int64(message.id);
            if (message.name != null && Object.hasOwnProperty.call(message, "name"))
                writer.uint32(/* id 2, wireType 2 =*/18).string(message.name);
            if (message.status != null && Object.hasOwnProperty.call(message, "status"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.status);
            return writer;
        };

        /**
         * Decodes a SimpleProjectItem message from the specified reader or buffer.
         * @function decode
         * @memberof NamespaceItem.SimpleProjectItem
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {NamespaceItem.SimpleProjectItem} SimpleProjectItem
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        SimpleProjectItem.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.NamespaceItem.SimpleProjectItem();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1:
                    message.id = reader.int64();
                    break;
                case 2:
                    message.name = reader.string();
                    break;
                case 3:
                    message.status = reader.string();
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        return SimpleProjectItem;
    })();

    return NamespaceItem;
})();

export const NamespaceList = $root.NamespaceList = (() => {

    /**
     * Properties of a NamespaceList.
     * @exports INamespaceList
     * @interface INamespaceList
     * @property {Array.<NamespaceItem>|null} [data] NamespaceList data
     */

    /**
     * Constructs a new NamespaceList.
     * @exports NamespaceList
     * @classdesc Represents a NamespaceList.
     * @implements INamespaceList
     * @constructor
     * @param {INamespaceList=} [properties] Properties to set
     */
    function NamespaceList(properties) {
        this.data = [];
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * NamespaceList data.
     * @member {Array.<NamespaceItem>} data
     * @memberof NamespaceList
     * @instance
     */
    NamespaceList.prototype.data = $util.emptyArray;

    /**
     * Encodes the specified NamespaceList message. Does not implicitly {@link NamespaceList.verify|verify} messages.
     * @function encode
     * @memberof NamespaceList
     * @static
     * @param {NamespaceList} message NamespaceList message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    NamespaceList.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.data != null && message.data.length)
            for (let i = 0; i < message.data.length; ++i)
                $root.NamespaceItem.encode(message.data[i], writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
        return writer;
    };

    /**
     * Decodes a NamespaceList message from the specified reader or buffer.
     * @function decode
     * @memberof NamespaceList
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {NamespaceList} NamespaceList
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    NamespaceList.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.NamespaceList();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                if (!(message.data && message.data.length))
                    message.data = [];
                message.data.push($root.NamespaceItem.decode(reader, reader.uint32()));
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return NamespaceList;
})();

export const NsStoreRequest = $root.NsStoreRequest = (() => {

    /**
     * Properties of a NsStoreRequest.
     * @exports INsStoreRequest
     * @interface INsStoreRequest
     * @property {string|null} [namespace] NsStoreRequest namespace
     */

    /**
     * Constructs a new NsStoreRequest.
     * @exports NsStoreRequest
     * @classdesc Represents a NsStoreRequest.
     * @implements INsStoreRequest
     * @constructor
     * @param {INsStoreRequest=} [properties] Properties to set
     */
    function NsStoreRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * NsStoreRequest namespace.
     * @member {string} namespace
     * @memberof NsStoreRequest
     * @instance
     */
    NsStoreRequest.prototype.namespace = "";

    /**
     * Encodes the specified NsStoreRequest message. Does not implicitly {@link NsStoreRequest.verify|verify} messages.
     * @function encode
     * @memberof NsStoreRequest
     * @static
     * @param {NsStoreRequest} message NsStoreRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    NsStoreRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.namespace != null && Object.hasOwnProperty.call(message, "namespace"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.namespace);
        return writer;
    };

    /**
     * Decodes a NsStoreRequest message from the specified reader or buffer.
     * @function decode
     * @memberof NsStoreRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {NsStoreRequest} NsStoreRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    NsStoreRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.NsStoreRequest();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.namespace = reader.string();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return NsStoreRequest;
})();

export const NsStoreResponse = $root.NsStoreResponse = (() => {

    /**
     * Properties of a NsStoreResponse.
     * @exports INsStoreResponse
     * @interface INsStoreResponse
     * @property {NamespaceResponse|null} [data] NsStoreResponse data
     */

    /**
     * Constructs a new NsStoreResponse.
     * @exports NsStoreResponse
     * @classdesc Represents a NsStoreResponse.
     * @implements INsStoreResponse
     * @constructor
     * @param {INsStoreResponse=} [properties] Properties to set
     */
    function NsStoreResponse(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * NsStoreResponse data.
     * @member {NamespaceResponse|null|undefined} data
     * @memberof NsStoreResponse
     * @instance
     */
    NsStoreResponse.prototype.data = null;

    /**
     * Encodes the specified NsStoreResponse message. Does not implicitly {@link NsStoreResponse.verify|verify} messages.
     * @function encode
     * @memberof NsStoreResponse
     * @static
     * @param {NsStoreResponse} message NsStoreResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    NsStoreResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.data != null && Object.hasOwnProperty.call(message, "data"))
            $root.NamespaceResponse.encode(message.data, writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
        return writer;
    };

    /**
     * Decodes a NsStoreResponse message from the specified reader or buffer.
     * @function decode
     * @memberof NsStoreResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {NsStoreResponse} NsStoreResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    NsStoreResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.NsStoreResponse();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.data = $root.NamespaceResponse.decode(reader, reader.uint32());
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return NsStoreResponse;
})();

export const CpuAndMemoryResponse = $root.CpuAndMemoryResponse = (() => {

    /**
     * Properties of a CpuAndMemoryResponse.
     * @exports ICpuAndMemoryResponse
     * @interface ICpuAndMemoryResponse
     * @property {string|null} [cpu] CpuAndMemoryResponse cpu
     * @property {string|null} [memory] CpuAndMemoryResponse memory
     */

    /**
     * Constructs a new CpuAndMemoryResponse.
     * @exports CpuAndMemoryResponse
     * @classdesc Represents a CpuAndMemoryResponse.
     * @implements ICpuAndMemoryResponse
     * @constructor
     * @param {ICpuAndMemoryResponse=} [properties] Properties to set
     */
    function CpuAndMemoryResponse(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * CpuAndMemoryResponse cpu.
     * @member {string} cpu
     * @memberof CpuAndMemoryResponse
     * @instance
     */
    CpuAndMemoryResponse.prototype.cpu = "";

    /**
     * CpuAndMemoryResponse memory.
     * @member {string} memory
     * @memberof CpuAndMemoryResponse
     * @instance
     */
    CpuAndMemoryResponse.prototype.memory = "";

    /**
     * Encodes the specified CpuAndMemoryResponse message. Does not implicitly {@link CpuAndMemoryResponse.verify|verify} messages.
     * @function encode
     * @memberof CpuAndMemoryResponse
     * @static
     * @param {CpuAndMemoryResponse} message CpuAndMemoryResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    CpuAndMemoryResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.cpu != null && Object.hasOwnProperty.call(message, "cpu"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.cpu);
        if (message.memory != null && Object.hasOwnProperty.call(message, "memory"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.memory);
        return writer;
    };

    /**
     * Decodes a CpuAndMemoryResponse message from the specified reader or buffer.
     * @function decode
     * @memberof CpuAndMemoryResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {CpuAndMemoryResponse} CpuAndMemoryResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    CpuAndMemoryResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.CpuAndMemoryResponse();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.cpu = reader.string();
                break;
            case 2:
                message.memory = reader.string();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return CpuAndMemoryResponse;
})();

export const ServiceEndpointsResponse = $root.ServiceEndpointsResponse = (() => {

    /**
     * Properties of a ServiceEndpointsResponse.
     * @exports IServiceEndpointsResponse
     * @interface IServiceEndpointsResponse
     * @property {Array.<ServiceEndpointsResponse.item>|null} [data] ServiceEndpointsResponse data
     */

    /**
     * Constructs a new ServiceEndpointsResponse.
     * @exports ServiceEndpointsResponse
     * @classdesc Represents a ServiceEndpointsResponse.
     * @implements IServiceEndpointsResponse
     * @constructor
     * @param {IServiceEndpointsResponse=} [properties] Properties to set
     */
    function ServiceEndpointsResponse(properties) {
        this.data = [];
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * ServiceEndpointsResponse data.
     * @member {Array.<ServiceEndpointsResponse.item>} data
     * @memberof ServiceEndpointsResponse
     * @instance
     */
    ServiceEndpointsResponse.prototype.data = $util.emptyArray;

    /**
     * Encodes the specified ServiceEndpointsResponse message. Does not implicitly {@link ServiceEndpointsResponse.verify|verify} messages.
     * @function encode
     * @memberof ServiceEndpointsResponse
     * @static
     * @param {ServiceEndpointsResponse} message ServiceEndpointsResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    ServiceEndpointsResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.data != null && message.data.length)
            for (let i = 0; i < message.data.length; ++i)
                $root.ServiceEndpointsResponse.item.encode(message.data[i], writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
        return writer;
    };

    /**
     * Decodes a ServiceEndpointsResponse message from the specified reader or buffer.
     * @function decode
     * @memberof ServiceEndpointsResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {ServiceEndpointsResponse} ServiceEndpointsResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    ServiceEndpointsResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.ServiceEndpointsResponse();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                if (!(message.data && message.data.length))
                    message.data = [];
                message.data.push($root.ServiceEndpointsResponse.item.decode(reader, reader.uint32()));
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    ServiceEndpointsResponse.item = (function() {

        /**
         * Properties of an item.
         * @memberof ServiceEndpointsResponse
         * @interface Iitem
         * @property {string|null} [name] item name
         * @property {Array.<string>|null} [url] item url
         */

        /**
         * Constructs a new item.
         * @memberof ServiceEndpointsResponse
         * @classdesc Represents an item.
         * @implements Iitem
         * @constructor
         * @param {ServiceEndpointsResponse.Iitem=} [properties] Properties to set
         */
        function item(properties) {
            this.url = [];
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * item name.
         * @member {string} name
         * @memberof ServiceEndpointsResponse.item
         * @instance
         */
        item.prototype.name = "";

        /**
         * item url.
         * @member {Array.<string>} url
         * @memberof ServiceEndpointsResponse.item
         * @instance
         */
        item.prototype.url = $util.emptyArray;

        /**
         * Encodes the specified item message. Does not implicitly {@link ServiceEndpointsResponse.item.verify|verify} messages.
         * @function encode
         * @memberof ServiceEndpointsResponse.item
         * @static
         * @param {ServiceEndpointsResponse.item} message item message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        item.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.name != null && Object.hasOwnProperty.call(message, "name"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.name);
            if (message.url != null && message.url.length)
                for (let i = 0; i < message.url.length; ++i)
                    writer.uint32(/* id 2, wireType 2 =*/18).string(message.url[i]);
            return writer;
        };

        /**
         * Decodes an item message from the specified reader or buffer.
         * @function decode
         * @memberof ServiceEndpointsResponse.item
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {ServiceEndpointsResponse.item} item
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        item.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.ServiceEndpointsResponse.item();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1:
                    message.name = reader.string();
                    break;
                case 2:
                    if (!(message.url && message.url.length))
                        message.url = [];
                    message.url.push(reader.string());
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        return item;
    })();

    return ServiceEndpointsResponse;
})();

export const ServiceEndpointsRequest = $root.ServiceEndpointsRequest = (() => {

    /**
     * Properties of a ServiceEndpointsRequest.
     * @exports IServiceEndpointsRequest
     * @interface IServiceEndpointsRequest
     * @property {number|null} [namespace_id] ServiceEndpointsRequest namespace_id
     * @property {string|null} [project_name] ServiceEndpointsRequest project_name
     */

    /**
     * Constructs a new ServiceEndpointsRequest.
     * @exports ServiceEndpointsRequest
     * @classdesc Represents a ServiceEndpointsRequest.
     * @implements IServiceEndpointsRequest
     * @constructor
     * @param {IServiceEndpointsRequest=} [properties] Properties to set
     */
    function ServiceEndpointsRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * ServiceEndpointsRequest namespace_id.
     * @member {number} namespace_id
     * @memberof ServiceEndpointsRequest
     * @instance
     */
    ServiceEndpointsRequest.prototype.namespace_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * ServiceEndpointsRequest project_name.
     * @member {string} project_name
     * @memberof ServiceEndpointsRequest
     * @instance
     */
    ServiceEndpointsRequest.prototype.project_name = "";

    /**
     * Encodes the specified ServiceEndpointsRequest message. Does not implicitly {@link ServiceEndpointsRequest.verify|verify} messages.
     * @function encode
     * @memberof ServiceEndpointsRequest
     * @static
     * @param {ServiceEndpointsRequest} message ServiceEndpointsRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    ServiceEndpointsRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.namespace_id != null && Object.hasOwnProperty.call(message, "namespace_id"))
            writer.uint32(/* id 1, wireType 0 =*/8).int64(message.namespace_id);
        if (message.project_name != null && Object.hasOwnProperty.call(message, "project_name"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.project_name);
        return writer;
    };

    /**
     * Decodes a ServiceEndpointsRequest message from the specified reader or buffer.
     * @function decode
     * @memberof ServiceEndpointsRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {ServiceEndpointsRequest} ServiceEndpointsRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    ServiceEndpointsRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.ServiceEndpointsRequest();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.namespace_id = reader.int64();
                break;
            case 2:
                message.project_name = reader.string();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return ServiceEndpointsRequest;
})();

export const Namespace = $root.Namespace = (() => {

    /**
     * Constructs a new Namespace service.
     * @exports Namespace
     * @classdesc Represents a Namespace
     * @extends $protobuf.rpc.Service
     * @constructor
     * @param {$protobuf.RPCImpl} rpcImpl RPC implementation
     * @param {boolean} [requestDelimited=false] Whether requests are length-delimited
     * @param {boolean} [responseDelimited=false] Whether responses are length-delimited
     */
    function Namespace(rpcImpl, requestDelimited, responseDelimited) {
        $protobuf.rpc.Service.call(this, rpcImpl, requestDelimited, responseDelimited);
    }

    (Namespace.prototype = Object.create($protobuf.rpc.Service.prototype)).constructor = Namespace;

    /**
     * Callback as used by {@link Namespace#index}.
     * @memberof Namespace
     * @typedef IndexCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {NamespaceList} [response] NamespaceList
     */

    /**
     * Calls Index.
     * @function index
     * @memberof Namespace
     * @instance
     * @param {google.protobuf.Empty} request Empty message or plain object
     * @param {Namespace.IndexCallback} callback Node-style callback called with the error, if any, and NamespaceList
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Namespace.prototype.index = function index(request, callback) {
        return this.rpcCall(index, $root.google.protobuf.Empty, $root.NamespaceList, request, callback);
    }, "name", { value: "Index" });

    /**
     * Calls Index.
     * @function index
     * @memberof Namespace
     * @instance
     * @param {google.protobuf.Empty} request Empty message or plain object
     * @returns {Promise<NamespaceList>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link Namespace#store}.
     * @memberof Namespace
     * @typedef StoreCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {NsStoreResponse} [response] NsStoreResponse
     */

    /**
     * Calls Store.
     * @function store
     * @memberof Namespace
     * @instance
     * @param {NsStoreRequest} request NsStoreRequest message or plain object
     * @param {Namespace.StoreCallback} callback Node-style callback called with the error, if any, and NsStoreResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Namespace.prototype.store = function store(request, callback) {
        return this.rpcCall(store, $root.NsStoreRequest, $root.NsStoreResponse, request, callback);
    }, "name", { value: "Store" });

    /**
     * Calls Store.
     * @function store
     * @memberof Namespace
     * @instance
     * @param {NsStoreRequest} request NsStoreRequest message or plain object
     * @returns {Promise<NsStoreResponse>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link Namespace#cpuAndMemory}.
     * @memberof Namespace
     * @typedef CpuAndMemoryCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {CpuAndMemoryResponse} [response] CpuAndMemoryResponse
     */

    /**
     * Calls CpuAndMemory.
     * @function cpuAndMemory
     * @memberof Namespace
     * @instance
     * @param {NamespaceID} request NamespaceID message or plain object
     * @param {Namespace.CpuAndMemoryCallback} callback Node-style callback called with the error, if any, and CpuAndMemoryResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Namespace.prototype.cpuAndMemory = function cpuAndMemory(request, callback) {
        return this.rpcCall(cpuAndMemory, $root.NamespaceID, $root.CpuAndMemoryResponse, request, callback);
    }, "name", { value: "CpuAndMemory" });

    /**
     * Calls CpuAndMemory.
     * @function cpuAndMemory
     * @memberof Namespace
     * @instance
     * @param {NamespaceID} request NamespaceID message or plain object
     * @returns {Promise<CpuAndMemoryResponse>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link Namespace#serviceEndpoints}.
     * @memberof Namespace
     * @typedef ServiceEndpointsCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {ServiceEndpointsResponse} [response] ServiceEndpointsResponse
     */

    /**
     * Calls ServiceEndpoints.
     * @function serviceEndpoints
     * @memberof Namespace
     * @instance
     * @param {ServiceEndpointsRequest} request ServiceEndpointsRequest message or plain object
     * @param {Namespace.ServiceEndpointsCallback} callback Node-style callback called with the error, if any, and ServiceEndpointsResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Namespace.prototype.serviceEndpoints = function serviceEndpoints(request, callback) {
        return this.rpcCall(serviceEndpoints, $root.ServiceEndpointsRequest, $root.ServiceEndpointsResponse, request, callback);
    }, "name", { value: "ServiceEndpoints" });

    /**
     * Calls ServiceEndpoints.
     * @function serviceEndpoints
     * @memberof Namespace
     * @instance
     * @param {ServiceEndpointsRequest} request ServiceEndpointsRequest message or plain object
     * @returns {Promise<ServiceEndpointsResponse>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link Namespace#destroy}.
     * @memberof Namespace
     * @typedef DestroyCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {google.protobuf.Empty} [response] Empty
     */

    /**
     * Calls Destroy.
     * @function destroy
     * @memberof Namespace
     * @instance
     * @param {NamespaceID} request NamespaceID message or plain object
     * @param {Namespace.DestroyCallback} callback Node-style callback called with the error, if any, and Empty
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Namespace.prototype.destroy = function destroy(request, callback) {
        return this.rpcCall(destroy, $root.NamespaceID, $root.google.protobuf.Empty, request, callback);
    }, "name", { value: "Destroy" });

    /**
     * Calls Destroy.
     * @function destroy
     * @memberof Namespace
     * @instance
     * @param {NamespaceID} request NamespaceID message or plain object
     * @returns {Promise<google.protobuf.Empty>} Promise
     * @variation 2
     */

    return Namespace;
})();

export const BackgroundRequest = $root.BackgroundRequest = (() => {

    /**
     * Properties of a BackgroundRequest.
     * @exports IBackgroundRequest
     * @interface IBackgroundRequest
     * @property {boolean|null} [random] BackgroundRequest random
     */

    /**
     * Constructs a new BackgroundRequest.
     * @exports BackgroundRequest
     * @classdesc Represents a BackgroundRequest.
     * @implements IBackgroundRequest
     * @constructor
     * @param {IBackgroundRequest=} [properties] Properties to set
     */
    function BackgroundRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * BackgroundRequest random.
     * @member {boolean} random
     * @memberof BackgroundRequest
     * @instance
     */
    BackgroundRequest.prototype.random = false;

    /**
     * Encodes the specified BackgroundRequest message. Does not implicitly {@link BackgroundRequest.verify|verify} messages.
     * @function encode
     * @memberof BackgroundRequest
     * @static
     * @param {BackgroundRequest} message BackgroundRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    BackgroundRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.random != null && Object.hasOwnProperty.call(message, "random"))
            writer.uint32(/* id 1, wireType 0 =*/8).bool(message.random);
        return writer;
    };

    /**
     * Decodes a BackgroundRequest message from the specified reader or buffer.
     * @function decode
     * @memberof BackgroundRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {BackgroundRequest} BackgroundRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    BackgroundRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.BackgroundRequest();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.random = reader.bool();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return BackgroundRequest;
})();

export const BackgroundResponse = $root.BackgroundResponse = (() => {

    /**
     * Properties of a BackgroundResponse.
     * @exports IBackgroundResponse
     * @interface IBackgroundResponse
     * @property {string|null} [url] BackgroundResponse url
     * @property {string|null} [copyright] BackgroundResponse copyright
     */

    /**
     * Constructs a new BackgroundResponse.
     * @exports BackgroundResponse
     * @classdesc Represents a BackgroundResponse.
     * @implements IBackgroundResponse
     * @constructor
     * @param {IBackgroundResponse=} [properties] Properties to set
     */
    function BackgroundResponse(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * BackgroundResponse url.
     * @member {string} url
     * @memberof BackgroundResponse
     * @instance
     */
    BackgroundResponse.prototype.url = "";

    /**
     * BackgroundResponse copyright.
     * @member {string} copyright
     * @memberof BackgroundResponse
     * @instance
     */
    BackgroundResponse.prototype.copyright = "";

    /**
     * Encodes the specified BackgroundResponse message. Does not implicitly {@link BackgroundResponse.verify|verify} messages.
     * @function encode
     * @memberof BackgroundResponse
     * @static
     * @param {BackgroundResponse} message BackgroundResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    BackgroundResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.url != null && Object.hasOwnProperty.call(message, "url"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.url);
        if (message.copyright != null && Object.hasOwnProperty.call(message, "copyright"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.copyright);
        return writer;
    };

    /**
     * Decodes a BackgroundResponse message from the specified reader or buffer.
     * @function decode
     * @memberof BackgroundResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {BackgroundResponse} BackgroundResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    BackgroundResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.BackgroundResponse();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.url = reader.string();
                break;
            case 2:
                message.copyright = reader.string();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return BackgroundResponse;
})();

export const Picture = $root.Picture = (() => {

    /**
     * Constructs a new Picture service.
     * @exports Picture
     * @classdesc Represents a Picture
     * @extends $protobuf.rpc.Service
     * @constructor
     * @param {$protobuf.RPCImpl} rpcImpl RPC implementation
     * @param {boolean} [requestDelimited=false] Whether requests are length-delimited
     * @param {boolean} [responseDelimited=false] Whether responses are length-delimited
     */
    function Picture(rpcImpl, requestDelimited, responseDelimited) {
        $protobuf.rpc.Service.call(this, rpcImpl, requestDelimited, responseDelimited);
    }

    (Picture.prototype = Object.create($protobuf.rpc.Service.prototype)).constructor = Picture;

    /**
     * Callback as used by {@link Picture#background}.
     * @memberof Picture
     * @typedef BackgroundCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {BackgroundResponse} [response] BackgroundResponse
     */

    /**
     * Calls Background.
     * @function background
     * @memberof Picture
     * @instance
     * @param {BackgroundRequest} request BackgroundRequest message or plain object
     * @param {Picture.BackgroundCallback} callback Node-style callback called with the error, if any, and BackgroundResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Picture.prototype.background = function background(request, callback) {
        return this.rpcCall(background, $root.BackgroundRequest, $root.BackgroundResponse, request, callback);
    }, "name", { value: "Background" });

    /**
     * Calls Background.
     * @function background
     * @memberof Picture
     * @instance
     * @param {BackgroundRequest} request BackgroundRequest message or plain object
     * @returns {Promise<BackgroundResponse>} Promise
     * @variation 2
     */

    return Picture;
})();

export const ProjectDestroyRequest = $root.ProjectDestroyRequest = (() => {

    /**
     * Properties of a ProjectDestroyRequest.
     * @exports IProjectDestroyRequest
     * @interface IProjectDestroyRequest
     * @property {number|null} [namespace_id] ProjectDestroyRequest namespace_id
     * @property {number|null} [project_id] ProjectDestroyRequest project_id
     */

    /**
     * Constructs a new ProjectDestroyRequest.
     * @exports ProjectDestroyRequest
     * @classdesc Represents a ProjectDestroyRequest.
     * @implements IProjectDestroyRequest
     * @constructor
     * @param {IProjectDestroyRequest=} [properties] Properties to set
     */
    function ProjectDestroyRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * ProjectDestroyRequest namespace_id.
     * @member {number} namespace_id
     * @memberof ProjectDestroyRequest
     * @instance
     */
    ProjectDestroyRequest.prototype.namespace_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * ProjectDestroyRequest project_id.
     * @member {number} project_id
     * @memberof ProjectDestroyRequest
     * @instance
     */
    ProjectDestroyRequest.prototype.project_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * Encodes the specified ProjectDestroyRequest message. Does not implicitly {@link ProjectDestroyRequest.verify|verify} messages.
     * @function encode
     * @memberof ProjectDestroyRequest
     * @static
     * @param {ProjectDestroyRequest} message ProjectDestroyRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    ProjectDestroyRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.namespace_id != null && Object.hasOwnProperty.call(message, "namespace_id"))
            writer.uint32(/* id 1, wireType 0 =*/8).int64(message.namespace_id);
        if (message.project_id != null && Object.hasOwnProperty.call(message, "project_id"))
            writer.uint32(/* id 2, wireType 0 =*/16).int64(message.project_id);
        return writer;
    };

    /**
     * Decodes a ProjectDestroyRequest message from the specified reader or buffer.
     * @function decode
     * @memberof ProjectDestroyRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {ProjectDestroyRequest} ProjectDestroyRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    ProjectDestroyRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.ProjectDestroyRequest();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.namespace_id = reader.int64();
                break;
            case 2:
                message.project_id = reader.int64();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return ProjectDestroyRequest;
})();

export const ProjectShowRequest = $root.ProjectShowRequest = (() => {

    /**
     * Properties of a ProjectShowRequest.
     * @exports IProjectShowRequest
     * @interface IProjectShowRequest
     * @property {number|null} [namespace_id] ProjectShowRequest namespace_id
     * @property {number|null} [project_id] ProjectShowRequest project_id
     */

    /**
     * Constructs a new ProjectShowRequest.
     * @exports ProjectShowRequest
     * @classdesc Represents a ProjectShowRequest.
     * @implements IProjectShowRequest
     * @constructor
     * @param {IProjectShowRequest=} [properties] Properties to set
     */
    function ProjectShowRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * ProjectShowRequest namespace_id.
     * @member {number} namespace_id
     * @memberof ProjectShowRequest
     * @instance
     */
    ProjectShowRequest.prototype.namespace_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * ProjectShowRequest project_id.
     * @member {number} project_id
     * @memberof ProjectShowRequest
     * @instance
     */
    ProjectShowRequest.prototype.project_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * Encodes the specified ProjectShowRequest message. Does not implicitly {@link ProjectShowRequest.verify|verify} messages.
     * @function encode
     * @memberof ProjectShowRequest
     * @static
     * @param {ProjectShowRequest} message ProjectShowRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    ProjectShowRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.namespace_id != null && Object.hasOwnProperty.call(message, "namespace_id"))
            writer.uint32(/* id 1, wireType 0 =*/8).int64(message.namespace_id);
        if (message.project_id != null && Object.hasOwnProperty.call(message, "project_id"))
            writer.uint32(/* id 2, wireType 0 =*/16).int64(message.project_id);
        return writer;
    };

    /**
     * Decodes a ProjectShowRequest message from the specified reader or buffer.
     * @function decode
     * @memberof ProjectShowRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {ProjectShowRequest} ProjectShowRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    ProjectShowRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.ProjectShowRequest();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.namespace_id = reader.int64();
                break;
            case 2:
                message.project_id = reader.int64();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return ProjectShowRequest;
})();

export const ProjectShowResponse = $root.ProjectShowResponse = (() => {

    /**
     * Properties of a ProjectShowResponse.
     * @exports IProjectShowResponse
     * @interface IProjectShowResponse
     * @property {number|null} [id] ProjectShowResponse id
     * @property {string|null} [name] ProjectShowResponse name
     * @property {number|null} [gitlab_project_id] ProjectShowResponse gitlab_project_id
     * @property {string|null} [gitlab_branch] ProjectShowResponse gitlab_branch
     * @property {string|null} [gitlab_commit] ProjectShowResponse gitlab_commit
     * @property {string|null} [config] ProjectShowResponse config
     * @property {string|null} [docker_image] ProjectShowResponse docker_image
     * @property {boolean|null} [atomic] ProjectShowResponse atomic
     * @property {string|null} [gitlab_commit_web_url] ProjectShowResponse gitlab_commit_web_url
     * @property {string|null} [gitlab_commit_title] ProjectShowResponse gitlab_commit_title
     * @property {string|null} [gitlab_commit_author] ProjectShowResponse gitlab_commit_author
     * @property {string|null} [gitlab_commit_date] ProjectShowResponse gitlab_commit_date
     * @property {Array.<string>|null} [urls] ProjectShowResponse urls
     * @property {ProjectShowResponse.Namespace|null} [namespace] ProjectShowResponse namespace
     * @property {string|null} [cpu] ProjectShowResponse cpu
     * @property {string|null} [memory] ProjectShowResponse memory
     * @property {string|null} [override_values] ProjectShowResponse override_values
     * @property {string|null} [created_at] ProjectShowResponse created_at
     * @property {string|null} [updated_at] ProjectShowResponse updated_at
     */

    /**
     * Constructs a new ProjectShowResponse.
     * @exports ProjectShowResponse
     * @classdesc Represents a ProjectShowResponse.
     * @implements IProjectShowResponse
     * @constructor
     * @param {IProjectShowResponse=} [properties] Properties to set
     */
    function ProjectShowResponse(properties) {
        this.urls = [];
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * ProjectShowResponse id.
     * @member {number} id
     * @memberof ProjectShowResponse
     * @instance
     */
    ProjectShowResponse.prototype.id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * ProjectShowResponse name.
     * @member {string} name
     * @memberof ProjectShowResponse
     * @instance
     */
    ProjectShowResponse.prototype.name = "";

    /**
     * ProjectShowResponse gitlab_project_id.
     * @member {number} gitlab_project_id
     * @memberof ProjectShowResponse
     * @instance
     */
    ProjectShowResponse.prototype.gitlab_project_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * ProjectShowResponse gitlab_branch.
     * @member {string} gitlab_branch
     * @memberof ProjectShowResponse
     * @instance
     */
    ProjectShowResponse.prototype.gitlab_branch = "";

    /**
     * ProjectShowResponse gitlab_commit.
     * @member {string} gitlab_commit
     * @memberof ProjectShowResponse
     * @instance
     */
    ProjectShowResponse.prototype.gitlab_commit = "";

    /**
     * ProjectShowResponse config.
     * @member {string} config
     * @memberof ProjectShowResponse
     * @instance
     */
    ProjectShowResponse.prototype.config = "";

    /**
     * ProjectShowResponse docker_image.
     * @member {string} docker_image
     * @memberof ProjectShowResponse
     * @instance
     */
    ProjectShowResponse.prototype.docker_image = "";

    /**
     * ProjectShowResponse atomic.
     * @member {boolean} atomic
     * @memberof ProjectShowResponse
     * @instance
     */
    ProjectShowResponse.prototype.atomic = false;

    /**
     * ProjectShowResponse gitlab_commit_web_url.
     * @member {string} gitlab_commit_web_url
     * @memberof ProjectShowResponse
     * @instance
     */
    ProjectShowResponse.prototype.gitlab_commit_web_url = "";

    /**
     * ProjectShowResponse gitlab_commit_title.
     * @member {string} gitlab_commit_title
     * @memberof ProjectShowResponse
     * @instance
     */
    ProjectShowResponse.prototype.gitlab_commit_title = "";

    /**
     * ProjectShowResponse gitlab_commit_author.
     * @member {string} gitlab_commit_author
     * @memberof ProjectShowResponse
     * @instance
     */
    ProjectShowResponse.prototype.gitlab_commit_author = "";

    /**
     * ProjectShowResponse gitlab_commit_date.
     * @member {string} gitlab_commit_date
     * @memberof ProjectShowResponse
     * @instance
     */
    ProjectShowResponse.prototype.gitlab_commit_date = "";

    /**
     * ProjectShowResponse urls.
     * @member {Array.<string>} urls
     * @memberof ProjectShowResponse
     * @instance
     */
    ProjectShowResponse.prototype.urls = $util.emptyArray;

    /**
     * ProjectShowResponse namespace.
     * @member {ProjectShowResponse.Namespace|null|undefined} namespace
     * @memberof ProjectShowResponse
     * @instance
     */
    ProjectShowResponse.prototype.namespace = null;

    /**
     * ProjectShowResponse cpu.
     * @member {string} cpu
     * @memberof ProjectShowResponse
     * @instance
     */
    ProjectShowResponse.prototype.cpu = "";

    /**
     * ProjectShowResponse memory.
     * @member {string} memory
     * @memberof ProjectShowResponse
     * @instance
     */
    ProjectShowResponse.prototype.memory = "";

    /**
     * ProjectShowResponse override_values.
     * @member {string} override_values
     * @memberof ProjectShowResponse
     * @instance
     */
    ProjectShowResponse.prototype.override_values = "";

    /**
     * ProjectShowResponse created_at.
     * @member {string} created_at
     * @memberof ProjectShowResponse
     * @instance
     */
    ProjectShowResponse.prototype.created_at = "";

    /**
     * ProjectShowResponse updated_at.
     * @member {string} updated_at
     * @memberof ProjectShowResponse
     * @instance
     */
    ProjectShowResponse.prototype.updated_at = "";

    /**
     * Encodes the specified ProjectShowResponse message. Does not implicitly {@link ProjectShowResponse.verify|verify} messages.
     * @function encode
     * @memberof ProjectShowResponse
     * @static
     * @param {ProjectShowResponse} message ProjectShowResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    ProjectShowResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.id != null && Object.hasOwnProperty.call(message, "id"))
            writer.uint32(/* id 1, wireType 0 =*/8).int64(message.id);
        if (message.name != null && Object.hasOwnProperty.call(message, "name"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.name);
        if (message.gitlab_project_id != null && Object.hasOwnProperty.call(message, "gitlab_project_id"))
            writer.uint32(/* id 3, wireType 0 =*/24).int64(message.gitlab_project_id);
        if (message.gitlab_branch != null && Object.hasOwnProperty.call(message, "gitlab_branch"))
            writer.uint32(/* id 4, wireType 2 =*/34).string(message.gitlab_branch);
        if (message.gitlab_commit != null && Object.hasOwnProperty.call(message, "gitlab_commit"))
            writer.uint32(/* id 5, wireType 2 =*/42).string(message.gitlab_commit);
        if (message.config != null && Object.hasOwnProperty.call(message, "config"))
            writer.uint32(/* id 6, wireType 2 =*/50).string(message.config);
        if (message.docker_image != null && Object.hasOwnProperty.call(message, "docker_image"))
            writer.uint32(/* id 7, wireType 2 =*/58).string(message.docker_image);
        if (message.atomic != null && Object.hasOwnProperty.call(message, "atomic"))
            writer.uint32(/* id 8, wireType 0 =*/64).bool(message.atomic);
        if (message.gitlab_commit_web_url != null && Object.hasOwnProperty.call(message, "gitlab_commit_web_url"))
            writer.uint32(/* id 9, wireType 2 =*/74).string(message.gitlab_commit_web_url);
        if (message.gitlab_commit_title != null && Object.hasOwnProperty.call(message, "gitlab_commit_title"))
            writer.uint32(/* id 10, wireType 2 =*/82).string(message.gitlab_commit_title);
        if (message.gitlab_commit_author != null && Object.hasOwnProperty.call(message, "gitlab_commit_author"))
            writer.uint32(/* id 11, wireType 2 =*/90).string(message.gitlab_commit_author);
        if (message.gitlab_commit_date != null && Object.hasOwnProperty.call(message, "gitlab_commit_date"))
            writer.uint32(/* id 12, wireType 2 =*/98).string(message.gitlab_commit_date);
        if (message.urls != null && message.urls.length)
            for (let i = 0; i < message.urls.length; ++i)
                writer.uint32(/* id 13, wireType 2 =*/106).string(message.urls[i]);
        if (message.namespace != null && Object.hasOwnProperty.call(message, "namespace"))
            $root.ProjectShowResponse.Namespace.encode(message.namespace, writer.uint32(/* id 14, wireType 2 =*/114).fork()).ldelim();
        if (message.cpu != null && Object.hasOwnProperty.call(message, "cpu"))
            writer.uint32(/* id 15, wireType 2 =*/122).string(message.cpu);
        if (message.memory != null && Object.hasOwnProperty.call(message, "memory"))
            writer.uint32(/* id 16, wireType 2 =*/130).string(message.memory);
        if (message.override_values != null && Object.hasOwnProperty.call(message, "override_values"))
            writer.uint32(/* id 17, wireType 2 =*/138).string(message.override_values);
        if (message.created_at != null && Object.hasOwnProperty.call(message, "created_at"))
            writer.uint32(/* id 18, wireType 2 =*/146).string(message.created_at);
        if (message.updated_at != null && Object.hasOwnProperty.call(message, "updated_at"))
            writer.uint32(/* id 19, wireType 2 =*/154).string(message.updated_at);
        return writer;
    };

    /**
     * Decodes a ProjectShowResponse message from the specified reader or buffer.
     * @function decode
     * @memberof ProjectShowResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {ProjectShowResponse} ProjectShowResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    ProjectShowResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.ProjectShowResponse();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.id = reader.int64();
                break;
            case 2:
                message.name = reader.string();
                break;
            case 3:
                message.gitlab_project_id = reader.int64();
                break;
            case 4:
                message.gitlab_branch = reader.string();
                break;
            case 5:
                message.gitlab_commit = reader.string();
                break;
            case 6:
                message.config = reader.string();
                break;
            case 7:
                message.docker_image = reader.string();
                break;
            case 8:
                message.atomic = reader.bool();
                break;
            case 9:
                message.gitlab_commit_web_url = reader.string();
                break;
            case 10:
                message.gitlab_commit_title = reader.string();
                break;
            case 11:
                message.gitlab_commit_author = reader.string();
                break;
            case 12:
                message.gitlab_commit_date = reader.string();
                break;
            case 13:
                if (!(message.urls && message.urls.length))
                    message.urls = [];
                message.urls.push(reader.string());
                break;
            case 14:
                message.namespace = $root.ProjectShowResponse.Namespace.decode(reader, reader.uint32());
                break;
            case 15:
                message.cpu = reader.string();
                break;
            case 16:
                message.memory = reader.string();
                break;
            case 17:
                message.override_values = reader.string();
                break;
            case 18:
                message.created_at = reader.string();
                break;
            case 19:
                message.updated_at = reader.string();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    ProjectShowResponse.Namespace = (function() {

        /**
         * Properties of a Namespace.
         * @memberof ProjectShowResponse
         * @interface INamespace
         * @property {number|null} [id] Namespace id
         * @property {string|null} [name] Namespace name
         */

        /**
         * Constructs a new Namespace.
         * @memberof ProjectShowResponse
         * @classdesc Represents a Namespace.
         * @implements INamespace
         * @constructor
         * @param {ProjectShowResponse.INamespace=} [properties] Properties to set
         */
        function Namespace(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * Namespace id.
         * @member {number} id
         * @memberof ProjectShowResponse.Namespace
         * @instance
         */
        Namespace.prototype.id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Namespace name.
         * @member {string} name
         * @memberof ProjectShowResponse.Namespace
         * @instance
         */
        Namespace.prototype.name = "";

        /**
         * Encodes the specified Namespace message. Does not implicitly {@link ProjectShowResponse.Namespace.verify|verify} messages.
         * @function encode
         * @memberof ProjectShowResponse.Namespace
         * @static
         * @param {ProjectShowResponse.Namespace} message Namespace message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        Namespace.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.id != null && Object.hasOwnProperty.call(message, "id"))
                writer.uint32(/* id 1, wireType 0 =*/8).int64(message.id);
            if (message.name != null && Object.hasOwnProperty.call(message, "name"))
                writer.uint32(/* id 2, wireType 2 =*/18).string(message.name);
            return writer;
        };

        /**
         * Decodes a Namespace message from the specified reader or buffer.
         * @function decode
         * @memberof ProjectShowResponse.Namespace
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {ProjectShowResponse.Namespace} Namespace
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Namespace.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.ProjectShowResponse.Namespace();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1:
                    message.id = reader.int64();
                    break;
                case 2:
                    message.name = reader.string();
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        return Namespace;
    })();

    return ProjectShowResponse;
})();

export const AllPodContainersRequest = $root.AllPodContainersRequest = (() => {

    /**
     * Properties of an AllPodContainersRequest.
     * @exports IAllPodContainersRequest
     * @interface IAllPodContainersRequest
     * @property {number|null} [namespace_id] AllPodContainersRequest namespace_id
     * @property {number|null} [project_id] AllPodContainersRequest project_id
     */

    /**
     * Constructs a new AllPodContainersRequest.
     * @exports AllPodContainersRequest
     * @classdesc Represents an AllPodContainersRequest.
     * @implements IAllPodContainersRequest
     * @constructor
     * @param {IAllPodContainersRequest=} [properties] Properties to set
     */
    function AllPodContainersRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * AllPodContainersRequest namespace_id.
     * @member {number} namespace_id
     * @memberof AllPodContainersRequest
     * @instance
     */
    AllPodContainersRequest.prototype.namespace_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * AllPodContainersRequest project_id.
     * @member {number} project_id
     * @memberof AllPodContainersRequest
     * @instance
     */
    AllPodContainersRequest.prototype.project_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * Encodes the specified AllPodContainersRequest message. Does not implicitly {@link AllPodContainersRequest.verify|verify} messages.
     * @function encode
     * @memberof AllPodContainersRequest
     * @static
     * @param {AllPodContainersRequest} message AllPodContainersRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    AllPodContainersRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.namespace_id != null && Object.hasOwnProperty.call(message, "namespace_id"))
            writer.uint32(/* id 1, wireType 0 =*/8).int64(message.namespace_id);
        if (message.project_id != null && Object.hasOwnProperty.call(message, "project_id"))
            writer.uint32(/* id 2, wireType 0 =*/16).int64(message.project_id);
        return writer;
    };

    /**
     * Decodes an AllPodContainersRequest message from the specified reader or buffer.
     * @function decode
     * @memberof AllPodContainersRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {AllPodContainersRequest} AllPodContainersRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    AllPodContainersRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.AllPodContainersRequest();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.namespace_id = reader.int64();
                break;
            case 2:
                message.project_id = reader.int64();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return AllPodContainersRequest;
})();

export const PodLog = $root.PodLog = (() => {

    /**
     * Properties of a PodLog.
     * @exports IPodLog
     * @interface IPodLog
     * @property {string|null} [pod_name] PodLog pod_name
     * @property {string|null} [container_name] PodLog container_name
     * @property {string|null} [log] PodLog log
     */

    /**
     * Constructs a new PodLog.
     * @exports PodLog
     * @classdesc Represents a PodLog.
     * @implements IPodLog
     * @constructor
     * @param {IPodLog=} [properties] Properties to set
     */
    function PodLog(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * PodLog pod_name.
     * @member {string} pod_name
     * @memberof PodLog
     * @instance
     */
    PodLog.prototype.pod_name = "";

    /**
     * PodLog container_name.
     * @member {string} container_name
     * @memberof PodLog
     * @instance
     */
    PodLog.prototype.container_name = "";

    /**
     * PodLog log.
     * @member {string} log
     * @memberof PodLog
     * @instance
     */
    PodLog.prototype.log = "";

    /**
     * Encodes the specified PodLog message. Does not implicitly {@link PodLog.verify|verify} messages.
     * @function encode
     * @memberof PodLog
     * @static
     * @param {PodLog} message PodLog message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    PodLog.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.pod_name != null && Object.hasOwnProperty.call(message, "pod_name"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.pod_name);
        if (message.container_name != null && Object.hasOwnProperty.call(message, "container_name"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.container_name);
        if (message.log != null && Object.hasOwnProperty.call(message, "log"))
            writer.uint32(/* id 3, wireType 2 =*/26).string(message.log);
        return writer;
    };

    /**
     * Decodes a PodLog message from the specified reader or buffer.
     * @function decode
     * @memberof PodLog
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {PodLog} PodLog
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    PodLog.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.PodLog();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.pod_name = reader.string();
                break;
            case 2:
                message.container_name = reader.string();
                break;
            case 3:
                message.log = reader.string();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return PodLog;
})();

export const AllPodContainersResponse = $root.AllPodContainersResponse = (() => {

    /**
     * Properties of an AllPodContainersResponse.
     * @exports IAllPodContainersResponse
     * @interface IAllPodContainersResponse
     * @property {Array.<PodLog>|null} [data] AllPodContainersResponse data
     */

    /**
     * Constructs a new AllPodContainersResponse.
     * @exports AllPodContainersResponse
     * @classdesc Represents an AllPodContainersResponse.
     * @implements IAllPodContainersResponse
     * @constructor
     * @param {IAllPodContainersResponse=} [properties] Properties to set
     */
    function AllPodContainersResponse(properties) {
        this.data = [];
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * AllPodContainersResponse data.
     * @member {Array.<PodLog>} data
     * @memberof AllPodContainersResponse
     * @instance
     */
    AllPodContainersResponse.prototype.data = $util.emptyArray;

    /**
     * Encodes the specified AllPodContainersResponse message. Does not implicitly {@link AllPodContainersResponse.verify|verify} messages.
     * @function encode
     * @memberof AllPodContainersResponse
     * @static
     * @param {AllPodContainersResponse} message AllPodContainersResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    AllPodContainersResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.data != null && message.data.length)
            for (let i = 0; i < message.data.length; ++i)
                $root.PodLog.encode(message.data[i], writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
        return writer;
    };

    /**
     * Decodes an AllPodContainersResponse message from the specified reader or buffer.
     * @function decode
     * @memberof AllPodContainersResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {AllPodContainersResponse} AllPodContainersResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    AllPodContainersResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.AllPodContainersResponse();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                if (!(message.data && message.data.length))
                    message.data = [];
                message.data.push($root.PodLog.decode(reader, reader.uint32()));
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return AllPodContainersResponse;
})();

export const PodContainerLogRequest = $root.PodContainerLogRequest = (() => {

    /**
     * Properties of a PodContainerLogRequest.
     * @exports IPodContainerLogRequest
     * @interface IPodContainerLogRequest
     * @property {number|null} [namespace_id] PodContainerLogRequest namespace_id
     * @property {number|null} [project_id] PodContainerLogRequest project_id
     * @property {string|null} [pod] PodContainerLogRequest pod
     * @property {string|null} [container] PodContainerLogRequest container
     */

    /**
     * Constructs a new PodContainerLogRequest.
     * @exports PodContainerLogRequest
     * @classdesc Represents a PodContainerLogRequest.
     * @implements IPodContainerLogRequest
     * @constructor
     * @param {IPodContainerLogRequest=} [properties] Properties to set
     */
    function PodContainerLogRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * PodContainerLogRequest namespace_id.
     * @member {number} namespace_id
     * @memberof PodContainerLogRequest
     * @instance
     */
    PodContainerLogRequest.prototype.namespace_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * PodContainerLogRequest project_id.
     * @member {number} project_id
     * @memberof PodContainerLogRequest
     * @instance
     */
    PodContainerLogRequest.prototype.project_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * PodContainerLogRequest pod.
     * @member {string} pod
     * @memberof PodContainerLogRequest
     * @instance
     */
    PodContainerLogRequest.prototype.pod = "";

    /**
     * PodContainerLogRequest container.
     * @member {string} container
     * @memberof PodContainerLogRequest
     * @instance
     */
    PodContainerLogRequest.prototype.container = "";

    /**
     * Encodes the specified PodContainerLogRequest message. Does not implicitly {@link PodContainerLogRequest.verify|verify} messages.
     * @function encode
     * @memberof PodContainerLogRequest
     * @static
     * @param {PodContainerLogRequest} message PodContainerLogRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    PodContainerLogRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.namespace_id != null && Object.hasOwnProperty.call(message, "namespace_id"))
            writer.uint32(/* id 1, wireType 0 =*/8).int64(message.namespace_id);
        if (message.project_id != null && Object.hasOwnProperty.call(message, "project_id"))
            writer.uint32(/* id 2, wireType 0 =*/16).int64(message.project_id);
        if (message.pod != null && Object.hasOwnProperty.call(message, "pod"))
            writer.uint32(/* id 3, wireType 2 =*/26).string(message.pod);
        if (message.container != null && Object.hasOwnProperty.call(message, "container"))
            writer.uint32(/* id 4, wireType 2 =*/34).string(message.container);
        return writer;
    };

    /**
     * Decodes a PodContainerLogRequest message from the specified reader or buffer.
     * @function decode
     * @memberof PodContainerLogRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {PodContainerLogRequest} PodContainerLogRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    PodContainerLogRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.PodContainerLogRequest();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.namespace_id = reader.int64();
                break;
            case 2:
                message.project_id = reader.int64();
                break;
            case 3:
                message.pod = reader.string();
                break;
            case 4:
                message.container = reader.string();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return PodContainerLogRequest;
})();

export const PodContainerLogResponse = $root.PodContainerLogResponse = (() => {

    /**
     * Properties of a PodContainerLogResponse.
     * @exports IPodContainerLogResponse
     * @interface IPodContainerLogResponse
     * @property {PodLog|null} [data] PodContainerLogResponse data
     */

    /**
     * Constructs a new PodContainerLogResponse.
     * @exports PodContainerLogResponse
     * @classdesc Represents a PodContainerLogResponse.
     * @implements IPodContainerLogResponse
     * @constructor
     * @param {IPodContainerLogResponse=} [properties] Properties to set
     */
    function PodContainerLogResponse(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * PodContainerLogResponse data.
     * @member {PodLog|null|undefined} data
     * @memberof PodContainerLogResponse
     * @instance
     */
    PodContainerLogResponse.prototype.data = null;

    /**
     * Encodes the specified PodContainerLogResponse message. Does not implicitly {@link PodContainerLogResponse.verify|verify} messages.
     * @function encode
     * @memberof PodContainerLogResponse
     * @static
     * @param {PodContainerLogResponse} message PodContainerLogResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    PodContainerLogResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.data != null && Object.hasOwnProperty.call(message, "data"))
            $root.PodLog.encode(message.data, writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
        return writer;
    };

    /**
     * Decodes a PodContainerLogResponse message from the specified reader or buffer.
     * @function decode
     * @memberof PodContainerLogResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {PodContainerLogResponse} PodContainerLogResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    PodContainerLogResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.PodContainerLogResponse();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.data = $root.PodLog.decode(reader, reader.uint32());
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return PodContainerLogResponse;
})();

export const IsPodRunningRequest = $root.IsPodRunningRequest = (() => {

    /**
     * Properties of an IsPodRunningRequest.
     * @exports IIsPodRunningRequest
     * @interface IIsPodRunningRequest
     * @property {string|null} [namespace] IsPodRunningRequest namespace
     * @property {string|null} [pod] IsPodRunningRequest pod
     */

    /**
     * Constructs a new IsPodRunningRequest.
     * @exports IsPodRunningRequest
     * @classdesc Represents an IsPodRunningRequest.
     * @implements IIsPodRunningRequest
     * @constructor
     * @param {IIsPodRunningRequest=} [properties] Properties to set
     */
    function IsPodRunningRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * IsPodRunningRequest namespace.
     * @member {string} namespace
     * @memberof IsPodRunningRequest
     * @instance
     */
    IsPodRunningRequest.prototype.namespace = "";

    /**
     * IsPodRunningRequest pod.
     * @member {string} pod
     * @memberof IsPodRunningRequest
     * @instance
     */
    IsPodRunningRequest.prototype.pod = "";

    /**
     * Encodes the specified IsPodRunningRequest message. Does not implicitly {@link IsPodRunningRequest.verify|verify} messages.
     * @function encode
     * @memberof IsPodRunningRequest
     * @static
     * @param {IsPodRunningRequest} message IsPodRunningRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    IsPodRunningRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.namespace != null && Object.hasOwnProperty.call(message, "namespace"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.namespace);
        if (message.pod != null && Object.hasOwnProperty.call(message, "pod"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.pod);
        return writer;
    };

    /**
     * Decodes an IsPodRunningRequest message from the specified reader or buffer.
     * @function decode
     * @memberof IsPodRunningRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {IsPodRunningRequest} IsPodRunningRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    IsPodRunningRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.IsPodRunningRequest();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.namespace = reader.string();
                break;
            case 2:
                message.pod = reader.string();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return IsPodRunningRequest;
})();

export const IsPodRunningResponse = $root.IsPodRunningResponse = (() => {

    /**
     * Properties of an IsPodRunningResponse.
     * @exports IIsPodRunningResponse
     * @interface IIsPodRunningResponse
     * @property {boolean|null} [running] IsPodRunningResponse running
     * @property {string|null} [reason] IsPodRunningResponse reason
     */

    /**
     * Constructs a new IsPodRunningResponse.
     * @exports IsPodRunningResponse
     * @classdesc Represents an IsPodRunningResponse.
     * @implements IIsPodRunningResponse
     * @constructor
     * @param {IIsPodRunningResponse=} [properties] Properties to set
     */
    function IsPodRunningResponse(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * IsPodRunningResponse running.
     * @member {boolean} running
     * @memberof IsPodRunningResponse
     * @instance
     */
    IsPodRunningResponse.prototype.running = false;

    /**
     * IsPodRunningResponse reason.
     * @member {string} reason
     * @memberof IsPodRunningResponse
     * @instance
     */
    IsPodRunningResponse.prototype.reason = "";

    /**
     * Encodes the specified IsPodRunningResponse message. Does not implicitly {@link IsPodRunningResponse.verify|verify} messages.
     * @function encode
     * @memberof IsPodRunningResponse
     * @static
     * @param {IsPodRunningResponse} message IsPodRunningResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    IsPodRunningResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.running != null && Object.hasOwnProperty.call(message, "running"))
            writer.uint32(/* id 1, wireType 0 =*/8).bool(message.running);
        if (message.reason != null && Object.hasOwnProperty.call(message, "reason"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.reason);
        return writer;
    };

    /**
     * Decodes an IsPodRunningResponse message from the specified reader or buffer.
     * @function decode
     * @memberof IsPodRunningResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {IsPodRunningResponse} IsPodRunningResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    IsPodRunningResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.IsPodRunningResponse();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.running = reader.bool();
                break;
            case 2:
                message.reason = reader.string();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return IsPodRunningResponse;
})();

export const Project = $root.Project = (() => {

    /**
     * Constructs a new Project service.
     * @exports Project
     * @classdesc Represents a Project
     * @extends $protobuf.rpc.Service
     * @constructor
     * @param {$protobuf.RPCImpl} rpcImpl RPC implementation
     * @param {boolean} [requestDelimited=false] Whether requests are length-delimited
     * @param {boolean} [responseDelimited=false] Whether responses are length-delimited
     */
    function Project(rpcImpl, requestDelimited, responseDelimited) {
        $protobuf.rpc.Service.call(this, rpcImpl, requestDelimited, responseDelimited);
    }

    (Project.prototype = Object.create($protobuf.rpc.Service.prototype)).constructor = Project;

    /**
     * Callback as used by {@link Project#destroy}.
     * @memberof Project
     * @typedef DestroyCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {google.protobuf.Empty} [response] Empty
     */

    /**
     * Calls Destroy.
     * @function destroy
     * @memberof Project
     * @instance
     * @param {ProjectDestroyRequest} request ProjectDestroyRequest message or plain object
     * @param {Project.DestroyCallback} callback Node-style callback called with the error, if any, and Empty
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Project.prototype.destroy = function destroy(request, callback) {
        return this.rpcCall(destroy, $root.ProjectDestroyRequest, $root.google.protobuf.Empty, request, callback);
    }, "name", { value: "Destroy" });

    /**
     * Calls Destroy.
     * @function destroy
     * @memberof Project
     * @instance
     * @param {ProjectDestroyRequest} request ProjectDestroyRequest message or plain object
     * @returns {Promise<google.protobuf.Empty>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link Project#show}.
     * @memberof Project
     * @typedef ShowCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {ProjectShowResponse} [response] ProjectShowResponse
     */

    /**
     * Calls Show.
     * @function show
     * @memberof Project
     * @instance
     * @param {ProjectShowRequest} request ProjectShowRequest message or plain object
     * @param {Project.ShowCallback} callback Node-style callback called with the error, if any, and ProjectShowResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Project.prototype.show = function show(request, callback) {
        return this.rpcCall(show, $root.ProjectShowRequest, $root.ProjectShowResponse, request, callback);
    }, "name", { value: "Show" });

    /**
     * Calls Show.
     * @function show
     * @memberof Project
     * @instance
     * @param {ProjectShowRequest} request ProjectShowRequest message or plain object
     * @returns {Promise<ProjectShowResponse>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link Project#isPodRunning}.
     * @memberof Project
     * @typedef IsPodRunningCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {IsPodRunningResponse} [response] IsPodRunningResponse
     */

    /**
     * Calls IsPodRunning.
     * @function isPodRunning
     * @memberof Project
     * @instance
     * @param {IsPodRunningRequest} request IsPodRunningRequest message or plain object
     * @param {Project.IsPodRunningCallback} callback Node-style callback called with the error, if any, and IsPodRunningResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Project.prototype.isPodRunning = function isPodRunning(request, callback) {
        return this.rpcCall(isPodRunning, $root.IsPodRunningRequest, $root.IsPodRunningResponse, request, callback);
    }, "name", { value: "IsPodRunning" });

    /**
     * Calls IsPodRunning.
     * @function isPodRunning
     * @memberof Project
     * @instance
     * @param {IsPodRunningRequest} request IsPodRunningRequest message or plain object
     * @returns {Promise<IsPodRunningResponse>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link Project#allPodContainers}.
     * @memberof Project
     * @typedef AllPodContainersCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {AllPodContainersResponse} [response] AllPodContainersResponse
     */

    /**
     * Calls AllPodContainers.
     * @function allPodContainers
     * @memberof Project
     * @instance
     * @param {AllPodContainersRequest} request AllPodContainersRequest message or plain object
     * @param {Project.AllPodContainersCallback} callback Node-style callback called with the error, if any, and AllPodContainersResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Project.prototype.allPodContainers = function allPodContainers(request, callback) {
        return this.rpcCall(allPodContainers, $root.AllPodContainersRequest, $root.AllPodContainersResponse, request, callback);
    }, "name", { value: "AllPodContainers" });

    /**
     * Calls AllPodContainers.
     * @function allPodContainers
     * @memberof Project
     * @instance
     * @param {AllPodContainersRequest} request AllPodContainersRequest message or plain object
     * @returns {Promise<AllPodContainersResponse>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link Project#podContainerLog}.
     * @memberof Project
     * @typedef PodContainerLogCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {PodContainerLogResponse} [response] PodContainerLogResponse
     */

    /**
     * Calls PodContainerLog.
     * @function podContainerLog
     * @memberof Project
     * @instance
     * @param {PodContainerLogRequest} request PodContainerLogRequest message or plain object
     * @param {Project.PodContainerLogCallback} callback Node-style callback called with the error, if any, and PodContainerLogResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Project.prototype.podContainerLog = function podContainerLog(request, callback) {
        return this.rpcCall(podContainerLog, $root.PodContainerLogRequest, $root.PodContainerLogResponse, request, callback);
    }, "name", { value: "PodContainerLog" });

    /**
     * Calls PodContainerLog.
     * @function podContainerLog
     * @memberof Project
     * @instance
     * @param {PodContainerLogRequest} request PodContainerLogRequest message or plain object
     * @returns {Promise<PodContainerLogResponse>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link Project#streamPodContainerLog}.
     * @memberof Project
     * @typedef StreamPodContainerLogCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {PodContainerLogResponse} [response] PodContainerLogResponse
     */

    /**
     * Calls StreamPodContainerLog.
     * @function streamPodContainerLog
     * @memberof Project
     * @instance
     * @param {PodContainerLogRequest} request PodContainerLogRequest message or plain object
     * @param {Project.StreamPodContainerLogCallback} callback Node-style callback called with the error, if any, and PodContainerLogResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Project.prototype.streamPodContainerLog = function streamPodContainerLog(request, callback) {
        return this.rpcCall(streamPodContainerLog, $root.PodContainerLogRequest, $root.PodContainerLogResponse, request, callback);
    }, "name", { value: "StreamPodContainerLog" });

    /**
     * Calls StreamPodContainerLog.
     * @function streamPodContainerLog
     * @memberof Project
     * @instance
     * @param {PodContainerLogRequest} request PodContainerLogRequest message or plain object
     * @returns {Promise<PodContainerLogResponse>} Promise
     * @variation 2
     */

    return Project;
})();

export const VersionResponse = $root.VersionResponse = (() => {

    /**
     * Properties of a VersionResponse.
     * @exports IVersionResponse
     * @interface IVersionResponse
     * @property {string|null} [Version] VersionResponse Version
     * @property {string|null} [BuildDate] VersionResponse BuildDate
     * @property {string|null} [gitBranch] VersionResponse gitBranch
     * @property {string|null} [GitCommit] VersionResponse GitCommit
     * @property {string|null} [GitTag] VersionResponse GitTag
     * @property {string|null} [GoVersion] VersionResponse GoVersion
     * @property {string|null} [Compiler] VersionResponse Compiler
     * @property {string|null} [Platform] VersionResponse Platform
     * @property {string|null} [KubectlVersion] VersionResponse KubectlVersion
     * @property {string|null} [HelmVersion] VersionResponse HelmVersion
     * @property {string|null} [GitRepo] VersionResponse GitRepo
     */

    /**
     * Constructs a new VersionResponse.
     * @exports VersionResponse
     * @classdesc Represents a VersionResponse.
     * @implements IVersionResponse
     * @constructor
     * @param {IVersionResponse=} [properties] Properties to set
     */
    function VersionResponse(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * VersionResponse Version.
     * @member {string} Version
     * @memberof VersionResponse
     * @instance
     */
    VersionResponse.prototype.Version = "";

    /**
     * VersionResponse BuildDate.
     * @member {string} BuildDate
     * @memberof VersionResponse
     * @instance
     */
    VersionResponse.prototype.BuildDate = "";

    /**
     * VersionResponse gitBranch.
     * @member {string} gitBranch
     * @memberof VersionResponse
     * @instance
     */
    VersionResponse.prototype.gitBranch = "";

    /**
     * VersionResponse GitCommit.
     * @member {string} GitCommit
     * @memberof VersionResponse
     * @instance
     */
    VersionResponse.prototype.GitCommit = "";

    /**
     * VersionResponse GitTag.
     * @member {string} GitTag
     * @memberof VersionResponse
     * @instance
     */
    VersionResponse.prototype.GitTag = "";

    /**
     * VersionResponse GoVersion.
     * @member {string} GoVersion
     * @memberof VersionResponse
     * @instance
     */
    VersionResponse.prototype.GoVersion = "";

    /**
     * VersionResponse Compiler.
     * @member {string} Compiler
     * @memberof VersionResponse
     * @instance
     */
    VersionResponse.prototype.Compiler = "";

    /**
     * VersionResponse Platform.
     * @member {string} Platform
     * @memberof VersionResponse
     * @instance
     */
    VersionResponse.prototype.Platform = "";

    /**
     * VersionResponse KubectlVersion.
     * @member {string} KubectlVersion
     * @memberof VersionResponse
     * @instance
     */
    VersionResponse.prototype.KubectlVersion = "";

    /**
     * VersionResponse HelmVersion.
     * @member {string} HelmVersion
     * @memberof VersionResponse
     * @instance
     */
    VersionResponse.prototype.HelmVersion = "";

    /**
     * VersionResponse GitRepo.
     * @member {string} GitRepo
     * @memberof VersionResponse
     * @instance
     */
    VersionResponse.prototype.GitRepo = "";

    /**
     * Encodes the specified VersionResponse message. Does not implicitly {@link VersionResponse.verify|verify} messages.
     * @function encode
     * @memberof VersionResponse
     * @static
     * @param {VersionResponse} message VersionResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    VersionResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.Version != null && Object.hasOwnProperty.call(message, "Version"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.Version);
        if (message.BuildDate != null && Object.hasOwnProperty.call(message, "BuildDate"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.BuildDate);
        if (message.gitBranch != null && Object.hasOwnProperty.call(message, "gitBranch"))
            writer.uint32(/* id 3, wireType 2 =*/26).string(message.gitBranch);
        if (message.GitCommit != null && Object.hasOwnProperty.call(message, "GitCommit"))
            writer.uint32(/* id 4, wireType 2 =*/34).string(message.GitCommit);
        if (message.GitTag != null && Object.hasOwnProperty.call(message, "GitTag"))
            writer.uint32(/* id 5, wireType 2 =*/42).string(message.GitTag);
        if (message.GoVersion != null && Object.hasOwnProperty.call(message, "GoVersion"))
            writer.uint32(/* id 6, wireType 2 =*/50).string(message.GoVersion);
        if (message.Compiler != null && Object.hasOwnProperty.call(message, "Compiler"))
            writer.uint32(/* id 7, wireType 2 =*/58).string(message.Compiler);
        if (message.Platform != null && Object.hasOwnProperty.call(message, "Platform"))
            writer.uint32(/* id 8, wireType 2 =*/66).string(message.Platform);
        if (message.KubectlVersion != null && Object.hasOwnProperty.call(message, "KubectlVersion"))
            writer.uint32(/* id 9, wireType 2 =*/74).string(message.KubectlVersion);
        if (message.HelmVersion != null && Object.hasOwnProperty.call(message, "HelmVersion"))
            writer.uint32(/* id 10, wireType 2 =*/82).string(message.HelmVersion);
        if (message.GitRepo != null && Object.hasOwnProperty.call(message, "GitRepo"))
            writer.uint32(/* id 11, wireType 2 =*/90).string(message.GitRepo);
        return writer;
    };

    /**
     * Decodes a VersionResponse message from the specified reader or buffer.
     * @function decode
     * @memberof VersionResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {VersionResponse} VersionResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    VersionResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.VersionResponse();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.Version = reader.string();
                break;
            case 2:
                message.BuildDate = reader.string();
                break;
            case 3:
                message.gitBranch = reader.string();
                break;
            case 4:
                message.GitCommit = reader.string();
                break;
            case 5:
                message.GitTag = reader.string();
                break;
            case 6:
                message.GoVersion = reader.string();
                break;
            case 7:
                message.Compiler = reader.string();
                break;
            case 8:
                message.Platform = reader.string();
                break;
            case 9:
                message.KubectlVersion = reader.string();
                break;
            case 10:
                message.HelmVersion = reader.string();
                break;
            case 11:
                message.GitRepo = reader.string();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return VersionResponse;
})();

export const Version = $root.Version = (() => {

    /**
     * Constructs a new Version service.
     * @exports Version
     * @classdesc Represents a Version
     * @extends $protobuf.rpc.Service
     * @constructor
     * @param {$protobuf.RPCImpl} rpcImpl RPC implementation
     * @param {boolean} [requestDelimited=false] Whether requests are length-delimited
     * @param {boolean} [responseDelimited=false] Whether responses are length-delimited
     */
    function Version(rpcImpl, requestDelimited, responseDelimited) {
        $protobuf.rpc.Service.call(this, rpcImpl, requestDelimited, responseDelimited);
    }

    (Version.prototype = Object.create($protobuf.rpc.Service.prototype)).constructor = Version;

    /**
     * Callback as used by {@link Version#get}.
     * @memberof Version
     * @typedef GetCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {VersionResponse} [response] VersionResponse
     */

    /**
     * Calls Get.
     * @function get
     * @memberof Version
     * @instance
     * @param {google.protobuf.Empty} request Empty message or plain object
     * @param {Version.GetCallback} callback Node-style callback called with the error, if any, and VersionResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Version.prototype.get = function get(request, callback) {
        return this.rpcCall(get, $root.google.protobuf.Empty, $root.VersionResponse, request, callback);
    }, "name", { value: "Get" });

    /**
     * Calls Get.
     * @function get
     * @memberof Version
     * @instance
     * @param {google.protobuf.Empty} request Empty message or plain object
     * @returns {Promise<VersionResponse>} Promise
     * @variation 2
     */

    return Version;
})();

/**
 * Type enum.
 * @exports Type
 * @enum {number}
 * @property {number} TypeUnknown=0 TypeUnknown value
 * @property {number} SetUid=1 SetUid value
 * @property {number} ReloadProjects=2 ReloadProjects value
 * @property {number} CancelProject=3 CancelProject value
 * @property {number} CreateProject=4 CreateProject value
 * @property {number} UpdateProject=5 UpdateProject value
 * @property {number} ProcessPercent=6 ProcessPercent value
 * @property {number} ClusterInfoSync=7 ClusterInfoSync value
 * @property {number} InternalError=8 InternalError value
 * @property {number} HandleExecShell=9 HandleExecShell value
 * @property {number} HandleExecShellMsg=10 HandleExecShellMsg value
 * @property {number} HandleCloseShell=11 HandleCloseShell value
 * @property {number} HandleAuthorize=12 HandleAuthorize value
 */
export const Type = $root.Type = (() => {
    const valuesById = {}, values = Object.create(valuesById);
    values[valuesById[0] = "TypeUnknown"] = 0;
    values[valuesById[1] = "SetUid"] = 1;
    values[valuesById[2] = "ReloadProjects"] = 2;
    values[valuesById[3] = "CancelProject"] = 3;
    values[valuesById[4] = "CreateProject"] = 4;
    values[valuesById[5] = "UpdateProject"] = 5;
    values[valuesById[6] = "ProcessPercent"] = 6;
    values[valuesById[7] = "ClusterInfoSync"] = 7;
    values[valuesById[8] = "InternalError"] = 8;
    values[valuesById[9] = "HandleExecShell"] = 9;
    values[valuesById[10] = "HandleExecShellMsg"] = 10;
    values[valuesById[11] = "HandleCloseShell"] = 11;
    values[valuesById[12] = "HandleAuthorize"] = 12;
    return values;
})();

/**
 * ResultType enum.
 * @exports ResultType
 * @enum {number}
 * @property {number} ResultUnknown=0 ResultUnknown value
 * @property {number} Error=1 Error value
 * @property {number} Success=2 Success value
 * @property {number} Deployed=3 Deployed value
 * @property {number} DeployedFailed=4 DeployedFailed value
 * @property {number} DeployedCanceled=5 DeployedCanceled value
 */
export const ResultType = $root.ResultType = (() => {
    const valuesById = {}, values = Object.create(valuesById);
    values[valuesById[0] = "ResultUnknown"] = 0;
    values[valuesById[1] = "Error"] = 1;
    values[valuesById[2] = "Success"] = 2;
    values[valuesById[3] = "Deployed"] = 3;
    values[valuesById[4] = "DeployedFailed"] = 4;
    values[valuesById[5] = "DeployedCanceled"] = 5;
    return values;
})();

/**
 * To enum.
 * @exports To
 * @enum {number}
 * @property {number} ToSelf=0 ToSelf value
 * @property {number} ToAll=1 ToAll value
 * @property {number} ToOthers=2 ToOthers value
 */
export const To = $root.To = (() => {
    const valuesById = {}, values = Object.create(valuesById);
    values[valuesById[0] = "ToSelf"] = 0;
    values[valuesById[1] = "ToAll"] = 1;
    values[valuesById[2] = "ToOthers"] = 2;
    return values;
})();

export const WsRequestMetadata = $root.WsRequestMetadata = (() => {

    /**
     * Properties of a WsRequestMetadata.
     * @exports IWsRequestMetadata
     * @interface IWsRequestMetadata
     * @property {Type|null} [type] WsRequestMetadata type
     */

    /**
     * Constructs a new WsRequestMetadata.
     * @exports WsRequestMetadata
     * @classdesc Represents a WsRequestMetadata.
     * @implements IWsRequestMetadata
     * @constructor
     * @param {IWsRequestMetadata=} [properties] Properties to set
     */
    function WsRequestMetadata(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * WsRequestMetadata type.
     * @member {Type} type
     * @memberof WsRequestMetadata
     * @instance
     */
    WsRequestMetadata.prototype.type = 0;

    /**
     * Encodes the specified WsRequestMetadata message. Does not implicitly {@link WsRequestMetadata.verify|verify} messages.
     * @function encode
     * @memberof WsRequestMetadata
     * @static
     * @param {WsRequestMetadata} message WsRequestMetadata message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    WsRequestMetadata.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.type != null && Object.hasOwnProperty.call(message, "type"))
            writer.uint32(/* id 1, wireType 0 =*/8).int32(message.type);
        return writer;
    };

    /**
     * Decodes a WsRequestMetadata message from the specified reader or buffer.
     * @function decode
     * @memberof WsRequestMetadata
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {WsRequestMetadata} WsRequestMetadata
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    WsRequestMetadata.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.WsRequestMetadata();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.type = reader.int32();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return WsRequestMetadata;
})();

export const AuthorizeTokenInput = $root.AuthorizeTokenInput = (() => {

    /**
     * Properties of an AuthorizeTokenInput.
     * @exports IAuthorizeTokenInput
     * @interface IAuthorizeTokenInput
     * @property {Type|null} [type] AuthorizeTokenInput type
     * @property {string|null} [token] AuthorizeTokenInput token
     */

    /**
     * Constructs a new AuthorizeTokenInput.
     * @exports AuthorizeTokenInput
     * @classdesc Represents an AuthorizeTokenInput.
     * @implements IAuthorizeTokenInput
     * @constructor
     * @param {IAuthorizeTokenInput=} [properties] Properties to set
     */
    function AuthorizeTokenInput(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * AuthorizeTokenInput type.
     * @member {Type} type
     * @memberof AuthorizeTokenInput
     * @instance
     */
    AuthorizeTokenInput.prototype.type = 0;

    /**
     * AuthorizeTokenInput token.
     * @member {string} token
     * @memberof AuthorizeTokenInput
     * @instance
     */
    AuthorizeTokenInput.prototype.token = "";

    /**
     * Encodes the specified AuthorizeTokenInput message. Does not implicitly {@link AuthorizeTokenInput.verify|verify} messages.
     * @function encode
     * @memberof AuthorizeTokenInput
     * @static
     * @param {AuthorizeTokenInput} message AuthorizeTokenInput message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    AuthorizeTokenInput.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.type != null && Object.hasOwnProperty.call(message, "type"))
            writer.uint32(/* id 1, wireType 0 =*/8).int32(message.type);
        if (message.token != null && Object.hasOwnProperty.call(message, "token"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.token);
        return writer;
    };

    /**
     * Decodes an AuthorizeTokenInput message from the specified reader or buffer.
     * @function decode
     * @memberof AuthorizeTokenInput
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {AuthorizeTokenInput} AuthorizeTokenInput
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    AuthorizeTokenInput.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.AuthorizeTokenInput();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.type = reader.int32();
                break;
            case 2:
                message.token = reader.string();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return AuthorizeTokenInput;
})();

export const TerminalMessage = $root.TerminalMessage = (() => {

    /**
     * Properties of a TerminalMessage.
     * @exports ITerminalMessage
     * @interface ITerminalMessage
     * @property {string|null} [op] TerminalMessage op
     * @property {string|null} [data] TerminalMessage data
     * @property {string|null} [session_id] TerminalMessage session_id
     * @property {number|null} [rows] TerminalMessage rows
     * @property {number|null} [cols] TerminalMessage cols
     */

    /**
     * Constructs a new TerminalMessage.
     * @exports TerminalMessage
     * @classdesc Represents a TerminalMessage.
     * @implements ITerminalMessage
     * @constructor
     * @param {ITerminalMessage=} [properties] Properties to set
     */
    function TerminalMessage(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * TerminalMessage op.
     * @member {string} op
     * @memberof TerminalMessage
     * @instance
     */
    TerminalMessage.prototype.op = "";

    /**
     * TerminalMessage data.
     * @member {string} data
     * @memberof TerminalMessage
     * @instance
     */
    TerminalMessage.prototype.data = "";

    /**
     * TerminalMessage session_id.
     * @member {string} session_id
     * @memberof TerminalMessage
     * @instance
     */
    TerminalMessage.prototype.session_id = "";

    /**
     * TerminalMessage rows.
     * @member {number} rows
     * @memberof TerminalMessage
     * @instance
     */
    TerminalMessage.prototype.rows = 0;

    /**
     * TerminalMessage cols.
     * @member {number} cols
     * @memberof TerminalMessage
     * @instance
     */
    TerminalMessage.prototype.cols = 0;

    /**
     * Encodes the specified TerminalMessage message. Does not implicitly {@link TerminalMessage.verify|verify} messages.
     * @function encode
     * @memberof TerminalMessage
     * @static
     * @param {TerminalMessage} message TerminalMessage message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    TerminalMessage.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.op != null && Object.hasOwnProperty.call(message, "op"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.op);
        if (message.data != null && Object.hasOwnProperty.call(message, "data"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.data);
        if (message.session_id != null && Object.hasOwnProperty.call(message, "session_id"))
            writer.uint32(/* id 3, wireType 2 =*/26).string(message.session_id);
        if (message.rows != null && Object.hasOwnProperty.call(message, "rows"))
            writer.uint32(/* id 4, wireType 0 =*/32).uint32(message.rows);
        if (message.cols != null && Object.hasOwnProperty.call(message, "cols"))
            writer.uint32(/* id 5, wireType 0 =*/40).uint32(message.cols);
        return writer;
    };

    /**
     * Decodes a TerminalMessage message from the specified reader or buffer.
     * @function decode
     * @memberof TerminalMessage
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {TerminalMessage} TerminalMessage
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    TerminalMessage.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.TerminalMessage();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.op = reader.string();
                break;
            case 2:
                message.data = reader.string();
                break;
            case 3:
                message.session_id = reader.string();
                break;
            case 4:
                message.rows = reader.uint32();
                break;
            case 5:
                message.cols = reader.uint32();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return TerminalMessage;
})();

export const TerminalMessageInput = $root.TerminalMessageInput = (() => {

    /**
     * Properties of a TerminalMessageInput.
     * @exports ITerminalMessageInput
     * @interface ITerminalMessageInput
     * @property {Type|null} [type] TerminalMessageInput type
     * @property {TerminalMessage|null} [message] TerminalMessageInput message
     */

    /**
     * Constructs a new TerminalMessageInput.
     * @exports TerminalMessageInput
     * @classdesc Represents a TerminalMessageInput.
     * @implements ITerminalMessageInput
     * @constructor
     * @param {ITerminalMessageInput=} [properties] Properties to set
     */
    function TerminalMessageInput(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * TerminalMessageInput type.
     * @member {Type} type
     * @memberof TerminalMessageInput
     * @instance
     */
    TerminalMessageInput.prototype.type = 0;

    /**
     * TerminalMessageInput message.
     * @member {TerminalMessage|null|undefined} message
     * @memberof TerminalMessageInput
     * @instance
     */
    TerminalMessageInput.prototype.message = null;

    /**
     * Encodes the specified TerminalMessageInput message. Does not implicitly {@link TerminalMessageInput.verify|verify} messages.
     * @function encode
     * @memberof TerminalMessageInput
     * @static
     * @param {TerminalMessageInput} message TerminalMessageInput message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    TerminalMessageInput.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.type != null && Object.hasOwnProperty.call(message, "type"))
            writer.uint32(/* id 1, wireType 0 =*/8).int32(message.type);
        if (message.message != null && Object.hasOwnProperty.call(message, "message"))
            $root.TerminalMessage.encode(message.message, writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim();
        return writer;
    };

    /**
     * Decodes a TerminalMessageInput message from the specified reader or buffer.
     * @function decode
     * @memberof TerminalMessageInput
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {TerminalMessageInput} TerminalMessageInput
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    TerminalMessageInput.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.TerminalMessageInput();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.type = reader.int32();
                break;
            case 2:
                message.message = $root.TerminalMessage.decode(reader, reader.uint32());
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return TerminalMessageInput;
})();

export const WsHandleExecShellInput = $root.WsHandleExecShellInput = (() => {

    /**
     * Properties of a WsHandleExecShellInput.
     * @exports IWsHandleExecShellInput
     * @interface IWsHandleExecShellInput
     * @property {Type|null} [type] WsHandleExecShellInput type
     * @property {string|null} [namespace] WsHandleExecShellInput namespace
     * @property {string|null} [pod] WsHandleExecShellInput pod
     * @property {string|null} [container] WsHandleExecShellInput container
     */

    /**
     * Constructs a new WsHandleExecShellInput.
     * @exports WsHandleExecShellInput
     * @classdesc Represents a WsHandleExecShellInput.
     * @implements IWsHandleExecShellInput
     * @constructor
     * @param {IWsHandleExecShellInput=} [properties] Properties to set
     */
    function WsHandleExecShellInput(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * WsHandleExecShellInput type.
     * @member {Type} type
     * @memberof WsHandleExecShellInput
     * @instance
     */
    WsHandleExecShellInput.prototype.type = 0;

    /**
     * WsHandleExecShellInput namespace.
     * @member {string} namespace
     * @memberof WsHandleExecShellInput
     * @instance
     */
    WsHandleExecShellInput.prototype.namespace = "";

    /**
     * WsHandleExecShellInput pod.
     * @member {string} pod
     * @memberof WsHandleExecShellInput
     * @instance
     */
    WsHandleExecShellInput.prototype.pod = "";

    /**
     * WsHandleExecShellInput container.
     * @member {string} container
     * @memberof WsHandleExecShellInput
     * @instance
     */
    WsHandleExecShellInput.prototype.container = "";

    /**
     * Encodes the specified WsHandleExecShellInput message. Does not implicitly {@link WsHandleExecShellInput.verify|verify} messages.
     * @function encode
     * @memberof WsHandleExecShellInput
     * @static
     * @param {WsHandleExecShellInput} message WsHandleExecShellInput message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    WsHandleExecShellInput.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.type != null && Object.hasOwnProperty.call(message, "type"))
            writer.uint32(/* id 1, wireType 0 =*/8).int32(message.type);
        if (message.namespace != null && Object.hasOwnProperty.call(message, "namespace"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.namespace);
        if (message.pod != null && Object.hasOwnProperty.call(message, "pod"))
            writer.uint32(/* id 3, wireType 2 =*/26).string(message.pod);
        if (message.container != null && Object.hasOwnProperty.call(message, "container"))
            writer.uint32(/* id 4, wireType 2 =*/34).string(message.container);
        return writer;
    };

    /**
     * Decodes a WsHandleExecShellInput message from the specified reader or buffer.
     * @function decode
     * @memberof WsHandleExecShellInput
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {WsHandleExecShellInput} WsHandleExecShellInput
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    WsHandleExecShellInput.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.WsHandleExecShellInput();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.type = reader.int32();
                break;
            case 2:
                message.namespace = reader.string();
                break;
            case 3:
                message.pod = reader.string();
                break;
            case 4:
                message.container = reader.string();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return WsHandleExecShellInput;
})();

export const CancelInput = $root.CancelInput = (() => {

    /**
     * Properties of a CancelInput.
     * @exports ICancelInput
     * @interface ICancelInput
     * @property {Type|null} [type] CancelInput type
     * @property {number|null} [namespace_id] CancelInput namespace_id
     * @property {string|null} [name] CancelInput name
     */

    /**
     * Constructs a new CancelInput.
     * @exports CancelInput
     * @classdesc Represents a CancelInput.
     * @implements ICancelInput
     * @constructor
     * @param {ICancelInput=} [properties] Properties to set
     */
    function CancelInput(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * CancelInput type.
     * @member {Type} type
     * @memberof CancelInput
     * @instance
     */
    CancelInput.prototype.type = 0;

    /**
     * CancelInput namespace_id.
     * @member {number} namespace_id
     * @memberof CancelInput
     * @instance
     */
    CancelInput.prototype.namespace_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * CancelInput name.
     * @member {string} name
     * @memberof CancelInput
     * @instance
     */
    CancelInput.prototype.name = "";

    /**
     * Encodes the specified CancelInput message. Does not implicitly {@link CancelInput.verify|verify} messages.
     * @function encode
     * @memberof CancelInput
     * @static
     * @param {CancelInput} message CancelInput message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    CancelInput.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.type != null && Object.hasOwnProperty.call(message, "type"))
            writer.uint32(/* id 1, wireType 0 =*/8).int32(message.type);
        if (message.namespace_id != null && Object.hasOwnProperty.call(message, "namespace_id"))
            writer.uint32(/* id 2, wireType 0 =*/16).int64(message.namespace_id);
        if (message.name != null && Object.hasOwnProperty.call(message, "name"))
            writer.uint32(/* id 3, wireType 2 =*/26).string(message.name);
        return writer;
    };

    /**
     * Decodes a CancelInput message from the specified reader or buffer.
     * @function decode
     * @memberof CancelInput
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {CancelInput} CancelInput
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    CancelInput.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.CancelInput();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.type = reader.int32();
                break;
            case 2:
                message.namespace_id = reader.int64();
                break;
            case 3:
                message.name = reader.string();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return CancelInput;
})();

export const ProjectInput = $root.ProjectInput = (() => {

    /**
     * Properties of a ProjectInput.
     * @exports IProjectInput
     * @interface IProjectInput
     * @property {Type|null} [type] ProjectInput type
     * @property {number|null} [namespace_id] ProjectInput namespace_id
     * @property {string|null} [name] ProjectInput name
     * @property {number|null} [gitlab_project_id] ProjectInput gitlab_project_id
     * @property {string|null} [gitlab_branch] ProjectInput gitlab_branch
     * @property {string|null} [gitlab_commit] ProjectInput gitlab_commit
     * @property {string|null} [config] ProjectInput config
     * @property {boolean|null} [atomic] ProjectInput atomic
     */

    /**
     * Constructs a new ProjectInput.
     * @exports ProjectInput
     * @classdesc Represents a ProjectInput.
     * @implements IProjectInput
     * @constructor
     * @param {IProjectInput=} [properties] Properties to set
     */
    function ProjectInput(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * ProjectInput type.
     * @member {Type} type
     * @memberof ProjectInput
     * @instance
     */
    ProjectInput.prototype.type = 0;

    /**
     * ProjectInput namespace_id.
     * @member {number} namespace_id
     * @memberof ProjectInput
     * @instance
     */
    ProjectInput.prototype.namespace_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * ProjectInput name.
     * @member {string} name
     * @memberof ProjectInput
     * @instance
     */
    ProjectInput.prototype.name = "";

    /**
     * ProjectInput gitlab_project_id.
     * @member {number} gitlab_project_id
     * @memberof ProjectInput
     * @instance
     */
    ProjectInput.prototype.gitlab_project_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * ProjectInput gitlab_branch.
     * @member {string} gitlab_branch
     * @memberof ProjectInput
     * @instance
     */
    ProjectInput.prototype.gitlab_branch = "";

    /**
     * ProjectInput gitlab_commit.
     * @member {string} gitlab_commit
     * @memberof ProjectInput
     * @instance
     */
    ProjectInput.prototype.gitlab_commit = "";

    /**
     * ProjectInput config.
     * @member {string} config
     * @memberof ProjectInput
     * @instance
     */
    ProjectInput.prototype.config = "";

    /**
     * ProjectInput atomic.
     * @member {boolean} atomic
     * @memberof ProjectInput
     * @instance
     */
    ProjectInput.prototype.atomic = false;

    /**
     * Encodes the specified ProjectInput message. Does not implicitly {@link ProjectInput.verify|verify} messages.
     * @function encode
     * @memberof ProjectInput
     * @static
     * @param {ProjectInput} message ProjectInput message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    ProjectInput.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.type != null && Object.hasOwnProperty.call(message, "type"))
            writer.uint32(/* id 1, wireType 0 =*/8).int32(message.type);
        if (message.namespace_id != null && Object.hasOwnProperty.call(message, "namespace_id"))
            writer.uint32(/* id 2, wireType 0 =*/16).int64(message.namespace_id);
        if (message.name != null && Object.hasOwnProperty.call(message, "name"))
            writer.uint32(/* id 3, wireType 2 =*/26).string(message.name);
        if (message.gitlab_project_id != null && Object.hasOwnProperty.call(message, "gitlab_project_id"))
            writer.uint32(/* id 4, wireType 0 =*/32).int64(message.gitlab_project_id);
        if (message.gitlab_branch != null && Object.hasOwnProperty.call(message, "gitlab_branch"))
            writer.uint32(/* id 5, wireType 2 =*/42).string(message.gitlab_branch);
        if (message.gitlab_commit != null && Object.hasOwnProperty.call(message, "gitlab_commit"))
            writer.uint32(/* id 6, wireType 2 =*/50).string(message.gitlab_commit);
        if (message.config != null && Object.hasOwnProperty.call(message, "config"))
            writer.uint32(/* id 7, wireType 2 =*/58).string(message.config);
        if (message.atomic != null && Object.hasOwnProperty.call(message, "atomic"))
            writer.uint32(/* id 8, wireType 0 =*/64).bool(message.atomic);
        return writer;
    };

    /**
     * Decodes a ProjectInput message from the specified reader or buffer.
     * @function decode
     * @memberof ProjectInput
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {ProjectInput} ProjectInput
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    ProjectInput.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.ProjectInput();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.type = reader.int32();
                break;
            case 2:
                message.namespace_id = reader.int64();
                break;
            case 3:
                message.name = reader.string();
                break;
            case 4:
                message.gitlab_project_id = reader.int64();
                break;
            case 5:
                message.gitlab_branch = reader.string();
                break;
            case 6:
                message.gitlab_commit = reader.string();
                break;
            case 7:
                message.config = reader.string();
                break;
            case 8:
                message.atomic = reader.bool();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return ProjectInput;
})();

export const UpdateProjectInput = $root.UpdateProjectInput = (() => {

    /**
     * Properties of an UpdateProjectInput.
     * @exports IUpdateProjectInput
     * @interface IUpdateProjectInput
     * @property {Type|null} [type] UpdateProjectInput type
     * @property {number|null} [project_id] UpdateProjectInput project_id
     * @property {string|null} [gitlab_branch] UpdateProjectInput gitlab_branch
     * @property {string|null} [gitlab_commit] UpdateProjectInput gitlab_commit
     * @property {string|null} [config] UpdateProjectInput config
     * @property {boolean|null} [atomic] UpdateProjectInput atomic
     */

    /**
     * Constructs a new UpdateProjectInput.
     * @exports UpdateProjectInput
     * @classdesc Represents an UpdateProjectInput.
     * @implements IUpdateProjectInput
     * @constructor
     * @param {IUpdateProjectInput=} [properties] Properties to set
     */
    function UpdateProjectInput(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * UpdateProjectInput type.
     * @member {Type} type
     * @memberof UpdateProjectInput
     * @instance
     */
    UpdateProjectInput.prototype.type = 0;

    /**
     * UpdateProjectInput project_id.
     * @member {number} project_id
     * @memberof UpdateProjectInput
     * @instance
     */
    UpdateProjectInput.prototype.project_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * UpdateProjectInput gitlab_branch.
     * @member {string} gitlab_branch
     * @memberof UpdateProjectInput
     * @instance
     */
    UpdateProjectInput.prototype.gitlab_branch = "";

    /**
     * UpdateProjectInput gitlab_commit.
     * @member {string} gitlab_commit
     * @memberof UpdateProjectInput
     * @instance
     */
    UpdateProjectInput.prototype.gitlab_commit = "";

    /**
     * UpdateProjectInput config.
     * @member {string} config
     * @memberof UpdateProjectInput
     * @instance
     */
    UpdateProjectInput.prototype.config = "";

    /**
     * UpdateProjectInput atomic.
     * @member {boolean} atomic
     * @memberof UpdateProjectInput
     * @instance
     */
    UpdateProjectInput.prototype.atomic = false;

    /**
     * Encodes the specified UpdateProjectInput message. Does not implicitly {@link UpdateProjectInput.verify|verify} messages.
     * @function encode
     * @memberof UpdateProjectInput
     * @static
     * @param {UpdateProjectInput} message UpdateProjectInput message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    UpdateProjectInput.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.type != null && Object.hasOwnProperty.call(message, "type"))
            writer.uint32(/* id 1, wireType 0 =*/8).int32(message.type);
        if (message.project_id != null && Object.hasOwnProperty.call(message, "project_id"))
            writer.uint32(/* id 2, wireType 0 =*/16).int64(message.project_id);
        if (message.gitlab_branch != null && Object.hasOwnProperty.call(message, "gitlab_branch"))
            writer.uint32(/* id 3, wireType 2 =*/26).string(message.gitlab_branch);
        if (message.gitlab_commit != null && Object.hasOwnProperty.call(message, "gitlab_commit"))
            writer.uint32(/* id 4, wireType 2 =*/34).string(message.gitlab_commit);
        if (message.config != null && Object.hasOwnProperty.call(message, "config"))
            writer.uint32(/* id 5, wireType 2 =*/42).string(message.config);
        if (message.atomic != null && Object.hasOwnProperty.call(message, "atomic"))
            writer.uint32(/* id 6, wireType 0 =*/48).bool(message.atomic);
        return writer;
    };

    /**
     * Decodes an UpdateProjectInput message from the specified reader or buffer.
     * @function decode
     * @memberof UpdateProjectInput
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {UpdateProjectInput} UpdateProjectInput
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    UpdateProjectInput.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.UpdateProjectInput();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.type = reader.int32();
                break;
            case 2:
                message.project_id = reader.int64();
                break;
            case 3:
                message.gitlab_branch = reader.string();
                break;
            case 4:
                message.gitlab_commit = reader.string();
                break;
            case 5:
                message.config = reader.string();
                break;
            case 6:
                message.atomic = reader.bool();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return UpdateProjectInput;
})();

export const ResponseMetadata = $root.ResponseMetadata = (() => {

    /**
     * Properties of a ResponseMetadata.
     * @exports IResponseMetadata
     * @interface IResponseMetadata
     * @property {string|null} [id] ResponseMetadata id
     * @property {string|null} [uid] ResponseMetadata uid
     * @property {string|null} [slug] ResponseMetadata slug
     * @property {Type|null} [type] ResponseMetadata type
     * @property {boolean|null} [end] ResponseMetadata end
     * @property {ResultType|null} [result] ResponseMetadata result
     * @property {To|null} [to] ResponseMetadata to
     * @property {string|null} [data] ResponseMetadata data
     */

    /**
     * Constructs a new ResponseMetadata.
     * @exports ResponseMetadata
     * @classdesc Represents a ResponseMetadata.
     * @implements IResponseMetadata
     * @constructor
     * @param {IResponseMetadata=} [properties] Properties to set
     */
    function ResponseMetadata(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * ResponseMetadata id.
     * @member {string} id
     * @memberof ResponseMetadata
     * @instance
     */
    ResponseMetadata.prototype.id = "";

    /**
     * ResponseMetadata uid.
     * @member {string} uid
     * @memberof ResponseMetadata
     * @instance
     */
    ResponseMetadata.prototype.uid = "";

    /**
     * ResponseMetadata slug.
     * @member {string} slug
     * @memberof ResponseMetadata
     * @instance
     */
    ResponseMetadata.prototype.slug = "";

    /**
     * ResponseMetadata type.
     * @member {Type} type
     * @memberof ResponseMetadata
     * @instance
     */
    ResponseMetadata.prototype.type = 0;

    /**
     * ResponseMetadata end.
     * @member {boolean} end
     * @memberof ResponseMetadata
     * @instance
     */
    ResponseMetadata.prototype.end = false;

    /**
     * ResponseMetadata result.
     * @member {ResultType} result
     * @memberof ResponseMetadata
     * @instance
     */
    ResponseMetadata.prototype.result = 0;

    /**
     * ResponseMetadata to.
     * @member {To} to
     * @memberof ResponseMetadata
     * @instance
     */
    ResponseMetadata.prototype.to = 0;

    /**
     * ResponseMetadata data.
     * @member {string} data
     * @memberof ResponseMetadata
     * @instance
     */
    ResponseMetadata.prototype.data = "";

    /**
     * Encodes the specified ResponseMetadata message. Does not implicitly {@link ResponseMetadata.verify|verify} messages.
     * @function encode
     * @memberof ResponseMetadata
     * @static
     * @param {ResponseMetadata} message ResponseMetadata message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    ResponseMetadata.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.id != null && Object.hasOwnProperty.call(message, "id"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.id);
        if (message.uid != null && Object.hasOwnProperty.call(message, "uid"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.uid);
        if (message.slug != null && Object.hasOwnProperty.call(message, "slug"))
            writer.uint32(/* id 3, wireType 2 =*/26).string(message.slug);
        if (message.type != null && Object.hasOwnProperty.call(message, "type"))
            writer.uint32(/* id 4, wireType 0 =*/32).int32(message.type);
        if (message.end != null && Object.hasOwnProperty.call(message, "end"))
            writer.uint32(/* id 5, wireType 0 =*/40).bool(message.end);
        if (message.result != null && Object.hasOwnProperty.call(message, "result"))
            writer.uint32(/* id 6, wireType 0 =*/48).int32(message.result);
        if (message.to != null && Object.hasOwnProperty.call(message, "to"))
            writer.uint32(/* id 7, wireType 0 =*/56).int32(message.to);
        if (message.data != null && Object.hasOwnProperty.call(message, "data"))
            writer.uint32(/* id 8, wireType 2 =*/66).string(message.data);
        return writer;
    };

    /**
     * Decodes a ResponseMetadata message from the specified reader or buffer.
     * @function decode
     * @memberof ResponseMetadata
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {ResponseMetadata} ResponseMetadata
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    ResponseMetadata.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.ResponseMetadata();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.id = reader.string();
                break;
            case 2:
                message.uid = reader.string();
                break;
            case 3:
                message.slug = reader.string();
                break;
            case 4:
                message.type = reader.int32();
                break;
            case 5:
                message.end = reader.bool();
                break;
            case 6:
                message.result = reader.int32();
                break;
            case 7:
                message.to = reader.int32();
                break;
            case 8:
                message.data = reader.string();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return ResponseMetadata;
})();

export const WsResponseMetadata = $root.WsResponseMetadata = (() => {

    /**
     * Properties of a WsResponseMetadata.
     * @exports IWsResponseMetadata
     * @interface IWsResponseMetadata
     * @property {ResponseMetadata|null} [metadata] WsResponseMetadata metadata
     */

    /**
     * Constructs a new WsResponseMetadata.
     * @exports WsResponseMetadata
     * @classdesc Represents a WsResponseMetadata.
     * @implements IWsResponseMetadata
     * @constructor
     * @param {IWsResponseMetadata=} [properties] Properties to set
     */
    function WsResponseMetadata(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * WsResponseMetadata metadata.
     * @member {ResponseMetadata|null|undefined} metadata
     * @memberof WsResponseMetadata
     * @instance
     */
    WsResponseMetadata.prototype.metadata = null;

    /**
     * Encodes the specified WsResponseMetadata message. Does not implicitly {@link WsResponseMetadata.verify|verify} messages.
     * @function encode
     * @memberof WsResponseMetadata
     * @static
     * @param {WsResponseMetadata} message WsResponseMetadata message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    WsResponseMetadata.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.metadata != null && Object.hasOwnProperty.call(message, "metadata"))
            $root.ResponseMetadata.encode(message.metadata, writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
        return writer;
    };

    /**
     * Decodes a WsResponseMetadata message from the specified reader or buffer.
     * @function decode
     * @memberof WsResponseMetadata
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {WsResponseMetadata} WsResponseMetadata
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    WsResponseMetadata.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.WsResponseMetadata();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.metadata = $root.ResponseMetadata.decode(reader, reader.uint32());
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return WsResponseMetadata;
})();

export const Container = $root.Container = (() => {

    /**
     * Properties of a Container.
     * @exports IContainer
     * @interface IContainer
     * @property {string|null} [namespace] Container namespace
     * @property {string|null} [pod] Container pod
     * @property {string|null} [container] Container container
     */

    /**
     * Constructs a new Container.
     * @exports Container
     * @classdesc Represents a Container.
     * @implements IContainer
     * @constructor
     * @param {IContainer=} [properties] Properties to set
     */
    function Container(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * Container namespace.
     * @member {string} namespace
     * @memberof Container
     * @instance
     */
    Container.prototype.namespace = "";

    /**
     * Container pod.
     * @member {string} pod
     * @memberof Container
     * @instance
     */
    Container.prototype.pod = "";

    /**
     * Container container.
     * @member {string} container
     * @memberof Container
     * @instance
     */
    Container.prototype.container = "";

    /**
     * Encodes the specified Container message. Does not implicitly {@link Container.verify|verify} messages.
     * @function encode
     * @memberof Container
     * @static
     * @param {Container} message Container message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    Container.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.namespace != null && Object.hasOwnProperty.call(message, "namespace"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.namespace);
        if (message.pod != null && Object.hasOwnProperty.call(message, "pod"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.pod);
        if (message.container != null && Object.hasOwnProperty.call(message, "container"))
            writer.uint32(/* id 3, wireType 2 =*/26).string(message.container);
        return writer;
    };

    /**
     * Decodes a Container message from the specified reader or buffer.
     * @function decode
     * @memberof Container
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {Container} Container
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    Container.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.Container();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.namespace = reader.string();
                break;
            case 2:
                message.pod = reader.string();
                break;
            case 3:
                message.container = reader.string();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return Container;
})();

export const WsHandleShellResponse = $root.WsHandleShellResponse = (() => {

    /**
     * Properties of a WsHandleShellResponse.
     * @exports IWsHandleShellResponse
     * @interface IWsHandleShellResponse
     * @property {ResponseMetadata|null} [metadata] WsHandleShellResponse metadata
     * @property {TerminalMessage|null} [terminal_message] WsHandleShellResponse terminal_message
     * @property {Container|null} [container] WsHandleShellResponse container
     */

    /**
     * Constructs a new WsHandleShellResponse.
     * @exports WsHandleShellResponse
     * @classdesc Represents a WsHandleShellResponse.
     * @implements IWsHandleShellResponse
     * @constructor
     * @param {IWsHandleShellResponse=} [properties] Properties to set
     */
    function WsHandleShellResponse(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * WsHandleShellResponse metadata.
     * @member {ResponseMetadata|null|undefined} metadata
     * @memberof WsHandleShellResponse
     * @instance
     */
    WsHandleShellResponse.prototype.metadata = null;

    /**
     * WsHandleShellResponse terminal_message.
     * @member {TerminalMessage|null|undefined} terminal_message
     * @memberof WsHandleShellResponse
     * @instance
     */
    WsHandleShellResponse.prototype.terminal_message = null;

    /**
     * WsHandleShellResponse container.
     * @member {Container|null|undefined} container
     * @memberof WsHandleShellResponse
     * @instance
     */
    WsHandleShellResponse.prototype.container = null;

    /**
     * Encodes the specified WsHandleShellResponse message. Does not implicitly {@link WsHandleShellResponse.verify|verify} messages.
     * @function encode
     * @memberof WsHandleShellResponse
     * @static
     * @param {WsHandleShellResponse} message WsHandleShellResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    WsHandleShellResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.metadata != null && Object.hasOwnProperty.call(message, "metadata"))
            $root.ResponseMetadata.encode(message.metadata, writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
        if (message.terminal_message != null && Object.hasOwnProperty.call(message, "terminal_message"))
            $root.TerminalMessage.encode(message.terminal_message, writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim();
        if (message.container != null && Object.hasOwnProperty.call(message, "container"))
            $root.Container.encode(message.container, writer.uint32(/* id 3, wireType 2 =*/26).fork()).ldelim();
        return writer;
    };

    /**
     * Decodes a WsHandleShellResponse message from the specified reader or buffer.
     * @function decode
     * @memberof WsHandleShellResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {WsHandleShellResponse} WsHandleShellResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    WsHandleShellResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.WsHandleShellResponse();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.metadata = $root.ResponseMetadata.decode(reader, reader.uint32());
                break;
            case 2:
                message.terminal_message = $root.TerminalMessage.decode(reader, reader.uint32());
                break;
            case 3:
                message.container = $root.Container.decode(reader, reader.uint32());
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return WsHandleShellResponse;
})();

export const WsHandleClusterResponse = $root.WsHandleClusterResponse = (() => {

    /**
     * Properties of a WsHandleClusterResponse.
     * @exports IWsHandleClusterResponse
     * @interface IWsHandleClusterResponse
     * @property {ResponseMetadata|null} [metadata] WsHandleClusterResponse metadata
     * @property {ClusterInfoResponse|null} [info] WsHandleClusterResponse info
     */

    /**
     * Constructs a new WsHandleClusterResponse.
     * @exports WsHandleClusterResponse
     * @classdesc Represents a WsHandleClusterResponse.
     * @implements IWsHandleClusterResponse
     * @constructor
     * @param {IWsHandleClusterResponse=} [properties] Properties to set
     */
    function WsHandleClusterResponse(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * WsHandleClusterResponse metadata.
     * @member {ResponseMetadata|null|undefined} metadata
     * @memberof WsHandleClusterResponse
     * @instance
     */
    WsHandleClusterResponse.prototype.metadata = null;

    /**
     * WsHandleClusterResponse info.
     * @member {ClusterInfoResponse|null|undefined} info
     * @memberof WsHandleClusterResponse
     * @instance
     */
    WsHandleClusterResponse.prototype.info = null;

    /**
     * Encodes the specified WsHandleClusterResponse message. Does not implicitly {@link WsHandleClusterResponse.verify|verify} messages.
     * @function encode
     * @memberof WsHandleClusterResponse
     * @static
     * @param {WsHandleClusterResponse} message WsHandleClusterResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    WsHandleClusterResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.metadata != null && Object.hasOwnProperty.call(message, "metadata"))
            $root.ResponseMetadata.encode(message.metadata, writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
        if (message.info != null && Object.hasOwnProperty.call(message, "info"))
            $root.ClusterInfoResponse.encode(message.info, writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim();
        return writer;
    };

    /**
     * Decodes a WsHandleClusterResponse message from the specified reader or buffer.
     * @function decode
     * @memberof WsHandleClusterResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {WsHandleClusterResponse} WsHandleClusterResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    WsHandleClusterResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.WsHandleClusterResponse();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.metadata = $root.ResponseMetadata.decode(reader, reader.uint32());
                break;
            case 2:
                message.info = $root.ClusterInfoResponse.decode(reader, reader.uint32());
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return WsHandleClusterResponse;
})();

export { $root as default };
