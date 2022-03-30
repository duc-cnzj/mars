/*eslint-disable block-scoped-var, id-length, no-control-regex, no-magic-numbers, no-prototype-builtins, no-redeclare, no-shadow, no-var, sort-vars*/
import * as $protobuf from "protobufjs/minimal";

// Common aliases
const $Reader = $protobuf.Reader, $Writer = $protobuf.Writer, $util = $protobuf.util;

// Exported root namespace
const $root = $protobuf.roots["default"] || ($protobuf.roots["default"] = {});

export const AuthLoginRequest = $root.AuthLoginRequest = (() => {

    /**
     * Properties of an AuthLoginRequest.
     * @exports IAuthLoginRequest
     * @interface IAuthLoginRequest
     * @property {string|null} [username] AuthLoginRequest username
     * @property {string|null} [password] AuthLoginRequest password
     */

    /**
     * Constructs a new AuthLoginRequest.
     * @exports AuthLoginRequest
     * @classdesc Represents an AuthLoginRequest.
     * @implements IAuthLoginRequest
     * @constructor
     * @param {IAuthLoginRequest=} [properties] Properties to set
     */
    function AuthLoginRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * AuthLoginRequest username.
     * @member {string} username
     * @memberof AuthLoginRequest
     * @instance
     */
    AuthLoginRequest.prototype.username = "";

    /**
     * AuthLoginRequest password.
     * @member {string} password
     * @memberof AuthLoginRequest
     * @instance
     */
    AuthLoginRequest.prototype.password = "";

    /**
     * Encodes the specified AuthLoginRequest message. Does not implicitly {@link AuthLoginRequest.verify|verify} messages.
     * @function encode
     * @memberof AuthLoginRequest
     * @static
     * @param {AuthLoginRequest} message AuthLoginRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    AuthLoginRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.username != null && Object.hasOwnProperty.call(message, "username"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.username);
        if (message.password != null && Object.hasOwnProperty.call(message, "password"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.password);
        return writer;
    };

    /**
     * Decodes an AuthLoginRequest message from the specified reader or buffer.
     * @function decode
     * @memberof AuthLoginRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {AuthLoginRequest} AuthLoginRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    AuthLoginRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.AuthLoginRequest();
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

    return AuthLoginRequest;
})();

export const AuthLoginResponse = $root.AuthLoginResponse = (() => {

    /**
     * Properties of an AuthLoginResponse.
     * @exports IAuthLoginResponse
     * @interface IAuthLoginResponse
     * @property {string|null} [token] AuthLoginResponse token
     * @property {number|null} [expires_in] AuthLoginResponse expires_in
     */

    /**
     * Constructs a new AuthLoginResponse.
     * @exports AuthLoginResponse
     * @classdesc Represents an AuthLoginResponse.
     * @implements IAuthLoginResponse
     * @constructor
     * @param {IAuthLoginResponse=} [properties] Properties to set
     */
    function AuthLoginResponse(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * AuthLoginResponse token.
     * @member {string} token
     * @memberof AuthLoginResponse
     * @instance
     */
    AuthLoginResponse.prototype.token = "";

    /**
     * AuthLoginResponse expires_in.
     * @member {number} expires_in
     * @memberof AuthLoginResponse
     * @instance
     */
    AuthLoginResponse.prototype.expires_in = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * Encodes the specified AuthLoginResponse message. Does not implicitly {@link AuthLoginResponse.verify|verify} messages.
     * @function encode
     * @memberof AuthLoginResponse
     * @static
     * @param {AuthLoginResponse} message AuthLoginResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    AuthLoginResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.token != null && Object.hasOwnProperty.call(message, "token"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.token);
        if (message.expires_in != null && Object.hasOwnProperty.call(message, "expires_in"))
            writer.uint32(/* id 2, wireType 0 =*/16).int64(message.expires_in);
        return writer;
    };

    /**
     * Decodes an AuthLoginResponse message from the specified reader or buffer.
     * @function decode
     * @memberof AuthLoginResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {AuthLoginResponse} AuthLoginResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    AuthLoginResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.AuthLoginResponse();
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

    return AuthLoginResponse;
})();

export const AuthExchangeRequest = $root.AuthExchangeRequest = (() => {

    /**
     * Properties of an AuthExchangeRequest.
     * @exports IAuthExchangeRequest
     * @interface IAuthExchangeRequest
     * @property {string|null} [code] AuthExchangeRequest code
     */

    /**
     * Constructs a new AuthExchangeRequest.
     * @exports AuthExchangeRequest
     * @classdesc Represents an AuthExchangeRequest.
     * @implements IAuthExchangeRequest
     * @constructor
     * @param {IAuthExchangeRequest=} [properties] Properties to set
     */
    function AuthExchangeRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * AuthExchangeRequest code.
     * @member {string} code
     * @memberof AuthExchangeRequest
     * @instance
     */
    AuthExchangeRequest.prototype.code = "";

    /**
     * Encodes the specified AuthExchangeRequest message. Does not implicitly {@link AuthExchangeRequest.verify|verify} messages.
     * @function encode
     * @memberof AuthExchangeRequest
     * @static
     * @param {AuthExchangeRequest} message AuthExchangeRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    AuthExchangeRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.code != null && Object.hasOwnProperty.call(message, "code"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.code);
        return writer;
    };

    /**
     * Decodes an AuthExchangeRequest message from the specified reader or buffer.
     * @function decode
     * @memberof AuthExchangeRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {AuthExchangeRequest} AuthExchangeRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    AuthExchangeRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.AuthExchangeRequest();
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

    return AuthExchangeRequest;
})();

export const AuthExchangeResponse = $root.AuthExchangeResponse = (() => {

    /**
     * Properties of an AuthExchangeResponse.
     * @exports IAuthExchangeResponse
     * @interface IAuthExchangeResponse
     * @property {string|null} [token] AuthExchangeResponse token
     * @property {number|null} [expires_in] AuthExchangeResponse expires_in
     */

    /**
     * Constructs a new AuthExchangeResponse.
     * @exports AuthExchangeResponse
     * @classdesc Represents an AuthExchangeResponse.
     * @implements IAuthExchangeResponse
     * @constructor
     * @param {IAuthExchangeResponse=} [properties] Properties to set
     */
    function AuthExchangeResponse(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * AuthExchangeResponse token.
     * @member {string} token
     * @memberof AuthExchangeResponse
     * @instance
     */
    AuthExchangeResponse.prototype.token = "";

    /**
     * AuthExchangeResponse expires_in.
     * @member {number} expires_in
     * @memberof AuthExchangeResponse
     * @instance
     */
    AuthExchangeResponse.prototype.expires_in = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * Encodes the specified AuthExchangeResponse message. Does not implicitly {@link AuthExchangeResponse.verify|verify} messages.
     * @function encode
     * @memberof AuthExchangeResponse
     * @static
     * @param {AuthExchangeResponse} message AuthExchangeResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    AuthExchangeResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.token != null && Object.hasOwnProperty.call(message, "token"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.token);
        if (message.expires_in != null && Object.hasOwnProperty.call(message, "expires_in"))
            writer.uint32(/* id 2, wireType 0 =*/16).int64(message.expires_in);
        return writer;
    };

    /**
     * Decodes an AuthExchangeResponse message from the specified reader or buffer.
     * @function decode
     * @memberof AuthExchangeResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {AuthExchangeResponse} AuthExchangeResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    AuthExchangeResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.AuthExchangeResponse();
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

    return AuthExchangeResponse;
})();

export const AuthInfoRequest = $root.AuthInfoRequest = (() => {

    /**
     * Properties of an AuthInfoRequest.
     * @exports IAuthInfoRequest
     * @interface IAuthInfoRequest
     */

    /**
     * Constructs a new AuthInfoRequest.
     * @exports AuthInfoRequest
     * @classdesc Represents an AuthInfoRequest.
     * @implements IAuthInfoRequest
     * @constructor
     * @param {IAuthInfoRequest=} [properties] Properties to set
     */
    function AuthInfoRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * Encodes the specified AuthInfoRequest message. Does not implicitly {@link AuthInfoRequest.verify|verify} messages.
     * @function encode
     * @memberof AuthInfoRequest
     * @static
     * @param {AuthInfoRequest} message AuthInfoRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    AuthInfoRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        return writer;
    };

    /**
     * Decodes an AuthInfoRequest message from the specified reader or buffer.
     * @function decode
     * @memberof AuthInfoRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {AuthInfoRequest} AuthInfoRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    AuthInfoRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.AuthInfoRequest();
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

    return AuthInfoRequest;
})();

export const AuthInfoResponse = $root.AuthInfoResponse = (() => {

    /**
     * Properties of an AuthInfoResponse.
     * @exports IAuthInfoResponse
     * @interface IAuthInfoResponse
     * @property {string|null} [id] AuthInfoResponse id
     * @property {string|null} [avatar] AuthInfoResponse avatar
     * @property {string|null} [name] AuthInfoResponse name
     * @property {string|null} [email] AuthInfoResponse email
     * @property {string|null} [logout_url] AuthInfoResponse logout_url
     * @property {Array.<string>|null} [roles] AuthInfoResponse roles
     */

    /**
     * Constructs a new AuthInfoResponse.
     * @exports AuthInfoResponse
     * @classdesc Represents an AuthInfoResponse.
     * @implements IAuthInfoResponse
     * @constructor
     * @param {IAuthInfoResponse=} [properties] Properties to set
     */
    function AuthInfoResponse(properties) {
        this.roles = [];
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * AuthInfoResponse id.
     * @member {string} id
     * @memberof AuthInfoResponse
     * @instance
     */
    AuthInfoResponse.prototype.id = "";

    /**
     * AuthInfoResponse avatar.
     * @member {string} avatar
     * @memberof AuthInfoResponse
     * @instance
     */
    AuthInfoResponse.prototype.avatar = "";

    /**
     * AuthInfoResponse name.
     * @member {string} name
     * @memberof AuthInfoResponse
     * @instance
     */
    AuthInfoResponse.prototype.name = "";

    /**
     * AuthInfoResponse email.
     * @member {string} email
     * @memberof AuthInfoResponse
     * @instance
     */
    AuthInfoResponse.prototype.email = "";

    /**
     * AuthInfoResponse logout_url.
     * @member {string} logout_url
     * @memberof AuthInfoResponse
     * @instance
     */
    AuthInfoResponse.prototype.logout_url = "";

    /**
     * AuthInfoResponse roles.
     * @member {Array.<string>} roles
     * @memberof AuthInfoResponse
     * @instance
     */
    AuthInfoResponse.prototype.roles = $util.emptyArray;

    /**
     * Encodes the specified AuthInfoResponse message. Does not implicitly {@link AuthInfoResponse.verify|verify} messages.
     * @function encode
     * @memberof AuthInfoResponse
     * @static
     * @param {AuthInfoResponse} message AuthInfoResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    AuthInfoResponse.encode = function encode(message, writer) {
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
     * Decodes an AuthInfoResponse message from the specified reader or buffer.
     * @function decode
     * @memberof AuthInfoResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {AuthInfoResponse} AuthInfoResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    AuthInfoResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.AuthInfoResponse();
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

    return AuthInfoResponse;
})();

export const AuthSettingsRequest = $root.AuthSettingsRequest = (() => {

    /**
     * Properties of an AuthSettingsRequest.
     * @exports IAuthSettingsRequest
     * @interface IAuthSettingsRequest
     */

    /**
     * Constructs a new AuthSettingsRequest.
     * @exports AuthSettingsRequest
     * @classdesc Represents an AuthSettingsRequest.
     * @implements IAuthSettingsRequest
     * @constructor
     * @param {IAuthSettingsRequest=} [properties] Properties to set
     */
    function AuthSettingsRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * Encodes the specified AuthSettingsRequest message. Does not implicitly {@link AuthSettingsRequest.verify|verify} messages.
     * @function encode
     * @memberof AuthSettingsRequest
     * @static
     * @param {AuthSettingsRequest} message AuthSettingsRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    AuthSettingsRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        return writer;
    };

    /**
     * Decodes an AuthSettingsRequest message from the specified reader or buffer.
     * @function decode
     * @memberof AuthSettingsRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {AuthSettingsRequest} AuthSettingsRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    AuthSettingsRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.AuthSettingsRequest();
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

    return AuthSettingsRequest;
})();

export const AuthSettingsResponse = $root.AuthSettingsResponse = (() => {

    /**
     * Properties of an AuthSettingsResponse.
     * @exports IAuthSettingsResponse
     * @interface IAuthSettingsResponse
     * @property {Array.<AuthSettingsResponse.OidcSetting>|null} [items] AuthSettingsResponse items
     */

    /**
     * Constructs a new AuthSettingsResponse.
     * @exports AuthSettingsResponse
     * @classdesc Represents an AuthSettingsResponse.
     * @implements IAuthSettingsResponse
     * @constructor
     * @param {IAuthSettingsResponse=} [properties] Properties to set
     */
    function AuthSettingsResponse(properties) {
        this.items = [];
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * AuthSettingsResponse items.
     * @member {Array.<AuthSettingsResponse.OidcSetting>} items
     * @memberof AuthSettingsResponse
     * @instance
     */
    AuthSettingsResponse.prototype.items = $util.emptyArray;

    /**
     * Encodes the specified AuthSettingsResponse message. Does not implicitly {@link AuthSettingsResponse.verify|verify} messages.
     * @function encode
     * @memberof AuthSettingsResponse
     * @static
     * @param {AuthSettingsResponse} message AuthSettingsResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    AuthSettingsResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.items != null && message.items.length)
            for (let i = 0; i < message.items.length; ++i)
                $root.AuthSettingsResponse.OidcSetting.encode(message.items[i], writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
        return writer;
    };

    /**
     * Decodes an AuthSettingsResponse message from the specified reader or buffer.
     * @function decode
     * @memberof AuthSettingsResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {AuthSettingsResponse} AuthSettingsResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    AuthSettingsResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.AuthSettingsResponse();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                if (!(message.items && message.items.length))
                    message.items = [];
                message.items.push($root.AuthSettingsResponse.OidcSetting.decode(reader, reader.uint32()));
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    AuthSettingsResponse.OidcSetting = (function() {

        /**
         * Properties of an OidcSetting.
         * @memberof AuthSettingsResponse
         * @interface IOidcSetting
         * @property {boolean|null} [enabled] OidcSetting enabled
         * @property {string|null} [name] OidcSetting name
         * @property {string|null} [url] OidcSetting url
         * @property {string|null} [end_session_endpoint] OidcSetting end_session_endpoint
         * @property {string|null} [state] OidcSetting state
         */

        /**
         * Constructs a new OidcSetting.
         * @memberof AuthSettingsResponse
         * @classdesc Represents an OidcSetting.
         * @implements IOidcSetting
         * @constructor
         * @param {AuthSettingsResponse.IOidcSetting=} [properties] Properties to set
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
         * @memberof AuthSettingsResponse.OidcSetting
         * @instance
         */
        OidcSetting.prototype.enabled = false;

        /**
         * OidcSetting name.
         * @member {string} name
         * @memberof AuthSettingsResponse.OidcSetting
         * @instance
         */
        OidcSetting.prototype.name = "";

        /**
         * OidcSetting url.
         * @member {string} url
         * @memberof AuthSettingsResponse.OidcSetting
         * @instance
         */
        OidcSetting.prototype.url = "";

        /**
         * OidcSetting end_session_endpoint.
         * @member {string} end_session_endpoint
         * @memberof AuthSettingsResponse.OidcSetting
         * @instance
         */
        OidcSetting.prototype.end_session_endpoint = "";

        /**
         * OidcSetting state.
         * @member {string} state
         * @memberof AuthSettingsResponse.OidcSetting
         * @instance
         */
        OidcSetting.prototype.state = "";

        /**
         * Encodes the specified OidcSetting message. Does not implicitly {@link AuthSettingsResponse.OidcSetting.verify|verify} messages.
         * @function encode
         * @memberof AuthSettingsResponse.OidcSetting
         * @static
         * @param {AuthSettingsResponse.OidcSetting} message OidcSetting message or plain object to encode
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
         * @memberof AuthSettingsResponse.OidcSetting
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {AuthSettingsResponse.OidcSetting} OidcSetting
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        OidcSetting.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.AuthSettingsResponse.OidcSetting();
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

    return AuthSettingsResponse;
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
     * @param {AuthLoginResponse} [response] AuthLoginResponse
     */

    /**
     * Calls Login.
     * @function login
     * @memberof Auth
     * @instance
     * @param {AuthLoginRequest} request AuthLoginRequest message or plain object
     * @param {Auth.LoginCallback} callback Node-style callback called with the error, if any, and AuthLoginResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Auth.prototype.login = function login(request, callback) {
        return this.rpcCall(login, $root.AuthLoginRequest, $root.AuthLoginResponse, request, callback);
    }, "name", { value: "Login" });

    /**
     * Calls Login.
     * @function login
     * @memberof Auth
     * @instance
     * @param {AuthLoginRequest} request AuthLoginRequest message or plain object
     * @returns {Promise<AuthLoginResponse>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link Auth#info}.
     * @memberof Auth
     * @typedef InfoCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {AuthInfoResponse} [response] AuthInfoResponse
     */

    /**
     * Calls Info.
     * @function info
     * @memberof Auth
     * @instance
     * @param {AuthInfoRequest} request AuthInfoRequest message or plain object
     * @param {Auth.InfoCallback} callback Node-style callback called with the error, if any, and AuthInfoResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Auth.prototype.info = function info(request, callback) {
        return this.rpcCall(info, $root.AuthInfoRequest, $root.AuthInfoResponse, request, callback);
    }, "name", { value: "Info" });

    /**
     * Calls Info.
     * @function info
     * @memberof Auth
     * @instance
     * @param {AuthInfoRequest} request AuthInfoRequest message or plain object
     * @returns {Promise<AuthInfoResponse>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link Auth#settings}.
     * @memberof Auth
     * @typedef SettingsCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {AuthSettingsResponse} [response] AuthSettingsResponse
     */

    /**
     * Calls Settings.
     * @function settings
     * @memberof Auth
     * @instance
     * @param {AuthSettingsRequest} request AuthSettingsRequest message or plain object
     * @param {Auth.SettingsCallback} callback Node-style callback called with the error, if any, and AuthSettingsResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Auth.prototype.settings = function settings(request, callback) {
        return this.rpcCall(settings, $root.AuthSettingsRequest, $root.AuthSettingsResponse, request, callback);
    }, "name", { value: "Settings" });

    /**
     * Calls Settings.
     * @function settings
     * @memberof Auth
     * @instance
     * @param {AuthSettingsRequest} request AuthSettingsRequest message or plain object
     * @returns {Promise<AuthSettingsResponse>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link Auth#exchange}.
     * @memberof Auth
     * @typedef ExchangeCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {AuthExchangeResponse} [response] AuthExchangeResponse
     */

    /**
     * Calls Exchange.
     * @function exchange
     * @memberof Auth
     * @instance
     * @param {AuthExchangeRequest} request AuthExchangeRequest message or plain object
     * @param {Auth.ExchangeCallback} callback Node-style callback called with the error, if any, and AuthExchangeResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Auth.prototype.exchange = function exchange(request, callback) {
        return this.rpcCall(exchange, $root.AuthExchangeRequest, $root.AuthExchangeResponse, request, callback);
    }, "name", { value: "Exchange" });

    /**
     * Calls Exchange.
     * @function exchange
     * @memberof Auth
     * @instance
     * @param {AuthExchangeRequest} request AuthExchangeRequest message or plain object
     * @returns {Promise<AuthExchangeResponse>} Promise
     * @variation 2
     */

    return Auth;
})();

export const google = $root.google = (() => {

    /**
     * Namespace google.
     * @exports google
     * @namespace
     */
    const google = {};

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

    google.protobuf = (function() {

        /**
         * Namespace protobuf.
         * @memberof google
         * @namespace
         */
        const protobuf = {};

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

        return protobuf;
    })();

    return google;
})();

export const ChangelogShowRequest = $root.ChangelogShowRequest = (() => {

    /**
     * Properties of a ChangelogShowRequest.
     * @exports IChangelogShowRequest
     * @interface IChangelogShowRequest
     * @property {number|null} [project_id] ChangelogShowRequest project_id
     * @property {boolean|null} [only_changed] ChangelogShowRequest only_changed
     */

    /**
     * Constructs a new ChangelogShowRequest.
     * @exports ChangelogShowRequest
     * @classdesc Represents a ChangelogShowRequest.
     * @implements IChangelogShowRequest
     * @constructor
     * @param {IChangelogShowRequest=} [properties] Properties to set
     */
    function ChangelogShowRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * ChangelogShowRequest project_id.
     * @member {number} project_id
     * @memberof ChangelogShowRequest
     * @instance
     */
    ChangelogShowRequest.prototype.project_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * ChangelogShowRequest only_changed.
     * @member {boolean} only_changed
     * @memberof ChangelogShowRequest
     * @instance
     */
    ChangelogShowRequest.prototype.only_changed = false;

    /**
     * Encodes the specified ChangelogShowRequest message. Does not implicitly {@link ChangelogShowRequest.verify|verify} messages.
     * @function encode
     * @memberof ChangelogShowRequest
     * @static
     * @param {ChangelogShowRequest} message ChangelogShowRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    ChangelogShowRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.project_id != null && Object.hasOwnProperty.call(message, "project_id"))
            writer.uint32(/* id 1, wireType 0 =*/8).int64(message.project_id);
        if (message.only_changed != null && Object.hasOwnProperty.call(message, "only_changed"))
            writer.uint32(/* id 2, wireType 0 =*/16).bool(message.only_changed);
        return writer;
    };

    /**
     * Decodes a ChangelogShowRequest message from the specified reader or buffer.
     * @function decode
     * @memberof ChangelogShowRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {ChangelogShowRequest} ChangelogShowRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    ChangelogShowRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.ChangelogShowRequest();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.project_id = reader.int64();
                break;
            case 2:
                message.only_changed = reader.bool();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return ChangelogShowRequest;
})();

export const ChangelogShowItem = $root.ChangelogShowItem = (() => {

    /**
     * Properties of a ChangelogShowItem.
     * @exports IChangelogShowItem
     * @interface IChangelogShowItem
     * @property {number|null} [version] ChangelogShowItem version
     * @property {string|null} [config] ChangelogShowItem config
     * @property {string|null} [date] ChangelogShowItem date
     * @property {string|null} [username] ChangelogShowItem username
     */

    /**
     * Constructs a new ChangelogShowItem.
     * @exports ChangelogShowItem
     * @classdesc Represents a ChangelogShowItem.
     * @implements IChangelogShowItem
     * @constructor
     * @param {IChangelogShowItem=} [properties] Properties to set
     */
    function ChangelogShowItem(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * ChangelogShowItem version.
     * @member {number} version
     * @memberof ChangelogShowItem
     * @instance
     */
    ChangelogShowItem.prototype.version = 0;

    /**
     * ChangelogShowItem config.
     * @member {string} config
     * @memberof ChangelogShowItem
     * @instance
     */
    ChangelogShowItem.prototype.config = "";

    /**
     * ChangelogShowItem date.
     * @member {string} date
     * @memberof ChangelogShowItem
     * @instance
     */
    ChangelogShowItem.prototype.date = "";

    /**
     * ChangelogShowItem username.
     * @member {string} username
     * @memberof ChangelogShowItem
     * @instance
     */
    ChangelogShowItem.prototype.username = "";

    /**
     * Encodes the specified ChangelogShowItem message. Does not implicitly {@link ChangelogShowItem.verify|verify} messages.
     * @function encode
     * @memberof ChangelogShowItem
     * @static
     * @param {ChangelogShowItem} message ChangelogShowItem message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    ChangelogShowItem.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.version != null && Object.hasOwnProperty.call(message, "version"))
            writer.uint32(/* id 1, wireType 0 =*/8).int32(message.version);
        if (message.config != null && Object.hasOwnProperty.call(message, "config"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.config);
        if (message.date != null && Object.hasOwnProperty.call(message, "date"))
            writer.uint32(/* id 3, wireType 2 =*/26).string(message.date);
        if (message.username != null && Object.hasOwnProperty.call(message, "username"))
            writer.uint32(/* id 4, wireType 2 =*/34).string(message.username);
        return writer;
    };

    /**
     * Decodes a ChangelogShowItem message from the specified reader or buffer.
     * @function decode
     * @memberof ChangelogShowItem
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {ChangelogShowItem} ChangelogShowItem
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    ChangelogShowItem.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.ChangelogShowItem();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.version = reader.int32();
                break;
            case 2:
                message.config = reader.string();
                break;
            case 3:
                message.date = reader.string();
                break;
            case 4:
                message.username = reader.string();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return ChangelogShowItem;
})();

export const ChangelogShowResponse = $root.ChangelogShowResponse = (() => {

    /**
     * Properties of a ChangelogShowResponse.
     * @exports IChangelogShowResponse
     * @interface IChangelogShowResponse
     * @property {Array.<ChangelogShowItem>|null} [items] ChangelogShowResponse items
     */

    /**
     * Constructs a new ChangelogShowResponse.
     * @exports ChangelogShowResponse
     * @classdesc Represents a ChangelogShowResponse.
     * @implements IChangelogShowResponse
     * @constructor
     * @param {IChangelogShowResponse=} [properties] Properties to set
     */
    function ChangelogShowResponse(properties) {
        this.items = [];
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * ChangelogShowResponse items.
     * @member {Array.<ChangelogShowItem>} items
     * @memberof ChangelogShowResponse
     * @instance
     */
    ChangelogShowResponse.prototype.items = $util.emptyArray;

    /**
     * Encodes the specified ChangelogShowResponse message. Does not implicitly {@link ChangelogShowResponse.verify|verify} messages.
     * @function encode
     * @memberof ChangelogShowResponse
     * @static
     * @param {ChangelogShowResponse} message ChangelogShowResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    ChangelogShowResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.items != null && message.items.length)
            for (let i = 0; i < message.items.length; ++i)
                $root.ChangelogShowItem.encode(message.items[i], writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
        return writer;
    };

    /**
     * Decodes a ChangelogShowResponse message from the specified reader or buffer.
     * @function decode
     * @memberof ChangelogShowResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {ChangelogShowResponse} ChangelogShowResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    ChangelogShowResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.ChangelogShowResponse();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                if (!(message.items && message.items.length))
                    message.items = [];
                message.items.push($root.ChangelogShowItem.decode(reader, reader.uint32()));
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return ChangelogShowResponse;
})();

export const Changelog = $root.Changelog = (() => {

    /**
     * Constructs a new Changelog service.
     * @exports Changelog
     * @classdesc Represents a Changelog
     * @extends $protobuf.rpc.Service
     * @constructor
     * @param {$protobuf.RPCImpl} rpcImpl RPC implementation
     * @param {boolean} [requestDelimited=false] Whether requests are length-delimited
     * @param {boolean} [responseDelimited=false] Whether responses are length-delimited
     */
    function Changelog(rpcImpl, requestDelimited, responseDelimited) {
        $protobuf.rpc.Service.call(this, rpcImpl, requestDelimited, responseDelimited);
    }

    (Changelog.prototype = Object.create($protobuf.rpc.Service.prototype)).constructor = Changelog;

    /**
     * Callback as used by {@link Changelog#show}.
     * @memberof Changelog
     * @typedef ShowCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {ChangelogShowResponse} [response] ChangelogShowResponse
     */

    /**
     * Calls Show.
     * @function show
     * @memberof Changelog
     * @instance
     * @param {ChangelogShowRequest} request ChangelogShowRequest message or plain object
     * @param {Changelog.ShowCallback} callback Node-style callback called with the error, if any, and ChangelogShowResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Changelog.prototype.show = function show(request, callback) {
        return this.rpcCall(show, $root.ChangelogShowRequest, $root.ChangelogShowResponse, request, callback);
    }, "name", { value: "Show" });

    /**
     * Calls Show.
     * @function show
     * @memberof Changelog
     * @instance
     * @param {ChangelogShowRequest} request ChangelogShowRequest message or plain object
     * @returns {Promise<ChangelogShowResponse>} Promise
     * @variation 2
     */

    return Changelog;
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

export const ClusterInfoRequest = $root.ClusterInfoRequest = (() => {

    /**
     * Properties of a ClusterInfoRequest.
     * @exports IClusterInfoRequest
     * @interface IClusterInfoRequest
     */

    /**
     * Constructs a new ClusterInfoRequest.
     * @exports ClusterInfoRequest
     * @classdesc Represents a ClusterInfoRequest.
     * @implements IClusterInfoRequest
     * @constructor
     * @param {IClusterInfoRequest=} [properties] Properties to set
     */
    function ClusterInfoRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * Encodes the specified ClusterInfoRequest message. Does not implicitly {@link ClusterInfoRequest.verify|verify} messages.
     * @function encode
     * @memberof ClusterInfoRequest
     * @static
     * @param {ClusterInfoRequest} message ClusterInfoRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    ClusterInfoRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        return writer;
    };

    /**
     * Decodes a ClusterInfoRequest message from the specified reader or buffer.
     * @function decode
     * @memberof ClusterInfoRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {ClusterInfoRequest} ClusterInfoRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    ClusterInfoRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.ClusterInfoRequest();
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

    return ClusterInfoRequest;
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
     * Callback as used by {@link Cluster#clusterInfo}.
     * @memberof Cluster
     * @typedef ClusterInfoCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {ClusterInfoResponse} [response] ClusterInfoResponse
     */

    /**
     * Calls ClusterInfo.
     * @function clusterInfo
     * @memberof Cluster
     * @instance
     * @param {ClusterInfoRequest} request ClusterInfoRequest message or plain object
     * @param {Cluster.ClusterInfoCallback} callback Node-style callback called with the error, if any, and ClusterInfoResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Cluster.prototype.clusterInfo = function clusterInfo(request, callback) {
        return this.rpcCall(clusterInfo, $root.ClusterInfoRequest, $root.ClusterInfoResponse, request, callback);
    }, "name", { value: "ClusterInfo" });

    /**
     * Calls ClusterInfo.
     * @function clusterInfo
     * @memberof Cluster
     * @instance
     * @param {ClusterInfoRequest} request ClusterInfoRequest message or plain object
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
     * @property {string|null} [file_name] CopyToPodResponse file_name
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
     * CopyToPodResponse file_name.
     * @member {string} file_name
     * @memberof CopyToPodResponse
     * @instance
     */
    CopyToPodResponse.prototype.file_name = "";

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
        if (message.file_name != null && Object.hasOwnProperty.call(message, "file_name"))
            writer.uint32(/* id 3, wireType 2 =*/26).string(message.file_name);
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
            case 3:
                message.file_name = reader.string();
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

export const ExecRequest = $root.ExecRequest = (() => {

    /**
     * Properties of an ExecRequest.
     * @exports IExecRequest
     * @interface IExecRequest
     * @property {string|null} [namespace] ExecRequest namespace
     * @property {string|null} [pod] ExecRequest pod
     * @property {string|null} [container] ExecRequest container
     * @property {Array.<string>|null} [command] ExecRequest command
     */

    /**
     * Constructs a new ExecRequest.
     * @exports ExecRequest
     * @classdesc Represents an ExecRequest.
     * @implements IExecRequest
     * @constructor
     * @param {IExecRequest=} [properties] Properties to set
     */
    function ExecRequest(properties) {
        this.command = [];
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * ExecRequest namespace.
     * @member {string} namespace
     * @memberof ExecRequest
     * @instance
     */
    ExecRequest.prototype.namespace = "";

    /**
     * ExecRequest pod.
     * @member {string} pod
     * @memberof ExecRequest
     * @instance
     */
    ExecRequest.prototype.pod = "";

    /**
     * ExecRequest container.
     * @member {string} container
     * @memberof ExecRequest
     * @instance
     */
    ExecRequest.prototype.container = "";

    /**
     * ExecRequest command.
     * @member {Array.<string>} command
     * @memberof ExecRequest
     * @instance
     */
    ExecRequest.prototype.command = $util.emptyArray;

    /**
     * Encodes the specified ExecRequest message. Does not implicitly {@link ExecRequest.verify|verify} messages.
     * @function encode
     * @memberof ExecRequest
     * @static
     * @param {ExecRequest} message ExecRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    ExecRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.namespace != null && Object.hasOwnProperty.call(message, "namespace"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.namespace);
        if (message.pod != null && Object.hasOwnProperty.call(message, "pod"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.pod);
        if (message.container != null && Object.hasOwnProperty.call(message, "container"))
            writer.uint32(/* id 3, wireType 2 =*/26).string(message.container);
        if (message.command != null && message.command.length)
            for (let i = 0; i < message.command.length; ++i)
                writer.uint32(/* id 4, wireType 2 =*/34).string(message.command[i]);
        return writer;
    };

    /**
     * Decodes an ExecRequest message from the specified reader or buffer.
     * @function decode
     * @memberof ExecRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {ExecRequest} ExecRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    ExecRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.ExecRequest();
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
            case 4:
                if (!(message.command && message.command.length))
                    message.command = [];
                message.command.push(reader.string());
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return ExecRequest;
})();

export const ExecResponse = $root.ExecResponse = (() => {

    /**
     * Properties of an ExecResponse.
     * @exports IExecResponse
     * @interface IExecResponse
     * @property {string|null} [data] ExecResponse data
     */

    /**
     * Constructs a new ExecResponse.
     * @exports ExecResponse
     * @classdesc Represents an ExecResponse.
     * @implements IExecResponse
     * @constructor
     * @param {IExecResponse=} [properties] Properties to set
     */
    function ExecResponse(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * ExecResponse data.
     * @member {string} data
     * @memberof ExecResponse
     * @instance
     */
    ExecResponse.prototype.data = "";

    /**
     * Encodes the specified ExecResponse message. Does not implicitly {@link ExecResponse.verify|verify} messages.
     * @function encode
     * @memberof ExecResponse
     * @static
     * @param {ExecResponse} message ExecResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    ExecResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.data != null && Object.hasOwnProperty.call(message, "data"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.data);
        return writer;
    };

    /**
     * Decodes an ExecResponse message from the specified reader or buffer.
     * @function decode
     * @memberof ExecResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {ExecResponse} ExecResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    ExecResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.ExecResponse();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.data = reader.string();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return ExecResponse;
})();

export const StreamCopyToPodRequest = $root.StreamCopyToPodRequest = (() => {

    /**
     * Properties of a StreamCopyToPodRequest.
     * @exports IStreamCopyToPodRequest
     * @interface IStreamCopyToPodRequest
     * @property {string|null} [file_name] StreamCopyToPodRequest file_name
     * @property {Uint8Array|null} [data] StreamCopyToPodRequest data
     * @property {string|null} [namespace] StreamCopyToPodRequest namespace
     * @property {string|null} [pod] StreamCopyToPodRequest pod
     * @property {string|null} [container] StreamCopyToPodRequest container
     */

    /**
     * Constructs a new StreamCopyToPodRequest.
     * @exports StreamCopyToPodRequest
     * @classdesc Represents a StreamCopyToPodRequest.
     * @implements IStreamCopyToPodRequest
     * @constructor
     * @param {IStreamCopyToPodRequest=} [properties] Properties to set
     */
    function StreamCopyToPodRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * StreamCopyToPodRequest file_name.
     * @member {string} file_name
     * @memberof StreamCopyToPodRequest
     * @instance
     */
    StreamCopyToPodRequest.prototype.file_name = "";

    /**
     * StreamCopyToPodRequest data.
     * @member {Uint8Array} data
     * @memberof StreamCopyToPodRequest
     * @instance
     */
    StreamCopyToPodRequest.prototype.data = $util.newBuffer([]);

    /**
     * StreamCopyToPodRequest namespace.
     * @member {string} namespace
     * @memberof StreamCopyToPodRequest
     * @instance
     */
    StreamCopyToPodRequest.prototype.namespace = "";

    /**
     * StreamCopyToPodRequest pod.
     * @member {string} pod
     * @memberof StreamCopyToPodRequest
     * @instance
     */
    StreamCopyToPodRequest.prototype.pod = "";

    /**
     * StreamCopyToPodRequest container.
     * @member {string} container
     * @memberof StreamCopyToPodRequest
     * @instance
     */
    StreamCopyToPodRequest.prototype.container = "";

    /**
     * Encodes the specified StreamCopyToPodRequest message. Does not implicitly {@link StreamCopyToPodRequest.verify|verify} messages.
     * @function encode
     * @memberof StreamCopyToPodRequest
     * @static
     * @param {StreamCopyToPodRequest} message StreamCopyToPodRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    StreamCopyToPodRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.file_name != null && Object.hasOwnProperty.call(message, "file_name"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.file_name);
        if (message.data != null && Object.hasOwnProperty.call(message, "data"))
            writer.uint32(/* id 2, wireType 2 =*/18).bytes(message.data);
        if (message.namespace != null && Object.hasOwnProperty.call(message, "namespace"))
            writer.uint32(/* id 3, wireType 2 =*/26).string(message.namespace);
        if (message.pod != null && Object.hasOwnProperty.call(message, "pod"))
            writer.uint32(/* id 4, wireType 2 =*/34).string(message.pod);
        if (message.container != null && Object.hasOwnProperty.call(message, "container"))
            writer.uint32(/* id 5, wireType 2 =*/42).string(message.container);
        return writer;
    };

    /**
     * Decodes a StreamCopyToPodRequest message from the specified reader or buffer.
     * @function decode
     * @memberof StreamCopyToPodRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {StreamCopyToPodRequest} StreamCopyToPodRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    StreamCopyToPodRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.StreamCopyToPodRequest();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.file_name = reader.string();
                break;
            case 2:
                message.data = reader.bytes();
                break;
            case 3:
                message.namespace = reader.string();
                break;
            case 4:
                message.pod = reader.string();
                break;
            case 5:
                message.container = reader.string();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return StreamCopyToPodRequest;
})();

export const StreamCopyToPodResponse = $root.StreamCopyToPodResponse = (() => {

    /**
     * Properties of a StreamCopyToPodResponse.
     * @exports IStreamCopyToPodResponse
     * @interface IStreamCopyToPodResponse
     * @property {number|null} [size] StreamCopyToPodResponse size
     * @property {string|null} [podFilePath] StreamCopyToPodResponse podFilePath
     * @property {string|null} [output] StreamCopyToPodResponse output
     * @property {string|null} [pod] StreamCopyToPodResponse pod
     * @property {string|null} [namespace] StreamCopyToPodResponse namespace
     * @property {string|null} [container] StreamCopyToPodResponse container
     * @property {string|null} [filename] StreamCopyToPodResponse filename
     */

    /**
     * Constructs a new StreamCopyToPodResponse.
     * @exports StreamCopyToPodResponse
     * @classdesc Represents a StreamCopyToPodResponse.
     * @implements IStreamCopyToPodResponse
     * @constructor
     * @param {IStreamCopyToPodResponse=} [properties] Properties to set
     */
    function StreamCopyToPodResponse(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * StreamCopyToPodResponse size.
     * @member {number} size
     * @memberof StreamCopyToPodResponse
     * @instance
     */
    StreamCopyToPodResponse.prototype.size = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * StreamCopyToPodResponse podFilePath.
     * @member {string} podFilePath
     * @memberof StreamCopyToPodResponse
     * @instance
     */
    StreamCopyToPodResponse.prototype.podFilePath = "";

    /**
     * StreamCopyToPodResponse output.
     * @member {string} output
     * @memberof StreamCopyToPodResponse
     * @instance
     */
    StreamCopyToPodResponse.prototype.output = "";

    /**
     * StreamCopyToPodResponse pod.
     * @member {string} pod
     * @memberof StreamCopyToPodResponse
     * @instance
     */
    StreamCopyToPodResponse.prototype.pod = "";

    /**
     * StreamCopyToPodResponse namespace.
     * @member {string} namespace
     * @memberof StreamCopyToPodResponse
     * @instance
     */
    StreamCopyToPodResponse.prototype.namespace = "";

    /**
     * StreamCopyToPodResponse container.
     * @member {string} container
     * @memberof StreamCopyToPodResponse
     * @instance
     */
    StreamCopyToPodResponse.prototype.container = "";

    /**
     * StreamCopyToPodResponse filename.
     * @member {string} filename
     * @memberof StreamCopyToPodResponse
     * @instance
     */
    StreamCopyToPodResponse.prototype.filename = "";

    /**
     * Encodes the specified StreamCopyToPodResponse message. Does not implicitly {@link StreamCopyToPodResponse.verify|verify} messages.
     * @function encode
     * @memberof StreamCopyToPodResponse
     * @static
     * @param {StreamCopyToPodResponse} message StreamCopyToPodResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    StreamCopyToPodResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.size != null && Object.hasOwnProperty.call(message, "size"))
            writer.uint32(/* id 1, wireType 0 =*/8).int64(message.size);
        if (message.podFilePath != null && Object.hasOwnProperty.call(message, "podFilePath"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.podFilePath);
        if (message.output != null && Object.hasOwnProperty.call(message, "output"))
            writer.uint32(/* id 3, wireType 2 =*/26).string(message.output);
        if (message.pod != null && Object.hasOwnProperty.call(message, "pod"))
            writer.uint32(/* id 4, wireType 2 =*/34).string(message.pod);
        if (message.namespace != null && Object.hasOwnProperty.call(message, "namespace"))
            writer.uint32(/* id 5, wireType 2 =*/42).string(message.namespace);
        if (message.container != null && Object.hasOwnProperty.call(message, "container"))
            writer.uint32(/* id 6, wireType 2 =*/50).string(message.container);
        if (message.filename != null && Object.hasOwnProperty.call(message, "filename"))
            writer.uint32(/* id 7, wireType 2 =*/58).string(message.filename);
        return writer;
    };

    /**
     * Decodes a StreamCopyToPodResponse message from the specified reader or buffer.
     * @function decode
     * @memberof StreamCopyToPodResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {StreamCopyToPodResponse} StreamCopyToPodResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    StreamCopyToPodResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.StreamCopyToPodResponse();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.size = reader.int64();
                break;
            case 2:
                message.podFilePath = reader.string();
                break;
            case 3:
                message.output = reader.string();
                break;
            case 4:
                message.pod = reader.string();
                break;
            case 5:
                message.namespace = reader.string();
                break;
            case 6:
                message.container = reader.string();
                break;
            case 7:
                message.filename = reader.string();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return StreamCopyToPodResponse;
})();

export const ContainerSvc = $root.ContainerSvc = (() => {

    /**
     * Constructs a new ContainerSvc service.
     * @exports ContainerSvc
     * @classdesc Represents a ContainerSvc
     * @extends $protobuf.rpc.Service
     * @constructor
     * @param {$protobuf.RPCImpl} rpcImpl RPC implementation
     * @param {boolean} [requestDelimited=false] Whether requests are length-delimited
     * @param {boolean} [responseDelimited=false] Whether responses are length-delimited
     */
    function ContainerSvc(rpcImpl, requestDelimited, responseDelimited) {
        $protobuf.rpc.Service.call(this, rpcImpl, requestDelimited, responseDelimited);
    }

    (ContainerSvc.prototype = Object.create($protobuf.rpc.Service.prototype)).constructor = ContainerSvc;

    /**
     * Callback as used by {@link ContainerSvc#copyToPod}.
     * @memberof ContainerSvc
     * @typedef CopyToPodCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {CopyToPodResponse} [response] CopyToPodResponse
     */

    /**
     * Calls CopyToPod.
     * @function copyToPod
     * @memberof ContainerSvc
     * @instance
     * @param {CopyToPodRequest} request CopyToPodRequest message or plain object
     * @param {ContainerSvc.CopyToPodCallback} callback Node-style callback called with the error, if any, and CopyToPodResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(ContainerSvc.prototype.copyToPod = function copyToPod(request, callback) {
        return this.rpcCall(copyToPod, $root.CopyToPodRequest, $root.CopyToPodResponse, request, callback);
    }, "name", { value: "CopyToPod" });

    /**
     * Calls CopyToPod.
     * @function copyToPod
     * @memberof ContainerSvc
     * @instance
     * @param {CopyToPodRequest} request CopyToPodRequest message or plain object
     * @returns {Promise<CopyToPodResponse>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link ContainerSvc#exec}.
     * @memberof ContainerSvc
     * @typedef ExecCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {ExecResponse} [response] ExecResponse
     */

    /**
     * Calls Exec.
     * @function exec
     * @memberof ContainerSvc
     * @instance
     * @param {ExecRequest} request ExecRequest message or plain object
     * @param {ContainerSvc.ExecCallback} callback Node-style callback called with the error, if any, and ExecResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(ContainerSvc.prototype.exec = function exec(request, callback) {
        return this.rpcCall(exec, $root.ExecRequest, $root.ExecResponse, request, callback);
    }, "name", { value: "Exec" });

    /**
     * Calls Exec.
     * @function exec
     * @memberof ContainerSvc
     * @instance
     * @param {ExecRequest} request ExecRequest message or plain object
     * @returns {Promise<ExecResponse>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link ContainerSvc#streamCopyToPod}.
     * @memberof ContainerSvc
     * @typedef StreamCopyToPodCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {StreamCopyToPodResponse} [response] StreamCopyToPodResponse
     */

    /**
     * Calls StreamCopyToPod.
     * @function streamCopyToPod
     * @memberof ContainerSvc
     * @instance
     * @param {StreamCopyToPodRequest} request StreamCopyToPodRequest message or plain object
     * @param {ContainerSvc.StreamCopyToPodCallback} callback Node-style callback called with the error, if any, and StreamCopyToPodResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(ContainerSvc.prototype.streamCopyToPod = function streamCopyToPod(request, callback) {
        return this.rpcCall(streamCopyToPod, $root.StreamCopyToPodRequest, $root.StreamCopyToPodResponse, request, callback);
    }, "name", { value: "StreamCopyToPod" });

    /**
     * Calls StreamCopyToPod.
     * @function streamCopyToPod
     * @memberof ContainerSvc
     * @instance
     * @param {StreamCopyToPodRequest} request StreamCopyToPodRequest message or plain object
     * @returns {Promise<StreamCopyToPodResponse>} Promise
     * @variation 2
     */

    return ContainerSvc;
})();

/**
 * ActionType enum.
 * @exports ActionType
 * @enum {number}
 * @property {number} Unknown=0 Unknown value
 * @property {number} Create=1 Create value
 * @property {number} Update=2 Update value
 * @property {number} Delete=3 Delete value
 * @property {number} Upload=4 Upload value
 * @property {number} Download=5 Download value
 * @property {number} DryRun=6 DryRun value
 */
export const ActionType = $root.ActionType = (() => {
    const valuesById = {}, values = Object.create(valuesById);
    values[valuesById[0] = "Unknown"] = 0;
    values[valuesById[1] = "Create"] = 1;
    values[valuesById[2] = "Update"] = 2;
    values[valuesById[3] = "Delete"] = 3;
    values[valuesById[4] = "Upload"] = 4;
    values[valuesById[5] = "Download"] = 5;
    values[valuesById[6] = "DryRun"] = 6;
    return values;
})();

export const EventListRequest = $root.EventListRequest = (() => {

    /**
     * Properties of an EventListRequest.
     * @exports IEventListRequest
     * @interface IEventListRequest
     * @property {number|null} [page] EventListRequest page
     * @property {number|null} [page_size] EventListRequest page_size
     */

    /**
     * Constructs a new EventListRequest.
     * @exports EventListRequest
     * @classdesc Represents an EventListRequest.
     * @implements IEventListRequest
     * @constructor
     * @param {IEventListRequest=} [properties] Properties to set
     */
    function EventListRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * EventListRequest page.
     * @member {number} page
     * @memberof EventListRequest
     * @instance
     */
    EventListRequest.prototype.page = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * EventListRequest page_size.
     * @member {number} page_size
     * @memberof EventListRequest
     * @instance
     */
    EventListRequest.prototype.page_size = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * Encodes the specified EventListRequest message. Does not implicitly {@link EventListRequest.verify|verify} messages.
     * @function encode
     * @memberof EventListRequest
     * @static
     * @param {EventListRequest} message EventListRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    EventListRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.page != null && Object.hasOwnProperty.call(message, "page"))
            writer.uint32(/* id 1, wireType 0 =*/8).int64(message.page);
        if (message.page_size != null && Object.hasOwnProperty.call(message, "page_size"))
            writer.uint32(/* id 2, wireType 0 =*/16).int64(message.page_size);
        return writer;
    };

    /**
     * Decodes an EventListRequest message from the specified reader or buffer.
     * @function decode
     * @memberof EventListRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {EventListRequest} EventListRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    EventListRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.EventListRequest();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.page = reader.int64();
                break;
            case 2:
                message.page_size = reader.int64();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return EventListRequest;
})();

export const EventListItem = $root.EventListItem = (() => {

    /**
     * Properties of an EventListItem.
     * @exports IEventListItem
     * @interface IEventListItem
     * @property {number|null} [id] EventListItem id
     * @property {ActionType|null} [action] EventListItem action
     * @property {string|null} [username] EventListItem username
     * @property {string|null} [message] EventListItem message
     * @property {string|null} [old] EventListItem old
     * @property {string|null} ["new"] EventListItem new
     * @property {string|null} [event_at] EventListItem event_at
     * @property {number|null} [file_id] EventListItem file_id
     */

    /**
     * Constructs a new EventListItem.
     * @exports EventListItem
     * @classdesc Represents an EventListItem.
     * @implements IEventListItem
     * @constructor
     * @param {IEventListItem=} [properties] Properties to set
     */
    function EventListItem(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * EventListItem id.
     * @member {number} id
     * @memberof EventListItem
     * @instance
     */
    EventListItem.prototype.id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * EventListItem action.
     * @member {ActionType} action
     * @memberof EventListItem
     * @instance
     */
    EventListItem.prototype.action = 0;

    /**
     * EventListItem username.
     * @member {string} username
     * @memberof EventListItem
     * @instance
     */
    EventListItem.prototype.username = "";

    /**
     * EventListItem message.
     * @member {string} message
     * @memberof EventListItem
     * @instance
     */
    EventListItem.prototype.message = "";

    /**
     * EventListItem old.
     * @member {string} old
     * @memberof EventListItem
     * @instance
     */
    EventListItem.prototype.old = "";

    /**
     * EventListItem new.
     * @member {string} new
     * @memberof EventListItem
     * @instance
     */
    EventListItem.prototype["new"] = "";

    /**
     * EventListItem event_at.
     * @member {string} event_at
     * @memberof EventListItem
     * @instance
     */
    EventListItem.prototype.event_at = "";

    /**
     * EventListItem file_id.
     * @member {number} file_id
     * @memberof EventListItem
     * @instance
     */
    EventListItem.prototype.file_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * Encodes the specified EventListItem message. Does not implicitly {@link EventListItem.verify|verify} messages.
     * @function encode
     * @memberof EventListItem
     * @static
     * @param {EventListItem} message EventListItem message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    EventListItem.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.id != null && Object.hasOwnProperty.call(message, "id"))
            writer.uint32(/* id 1, wireType 0 =*/8).int64(message.id);
        if (message.action != null && Object.hasOwnProperty.call(message, "action"))
            writer.uint32(/* id 2, wireType 0 =*/16).int32(message.action);
        if (message.username != null && Object.hasOwnProperty.call(message, "username"))
            writer.uint32(/* id 3, wireType 2 =*/26).string(message.username);
        if (message.message != null && Object.hasOwnProperty.call(message, "message"))
            writer.uint32(/* id 4, wireType 2 =*/34).string(message.message);
        if (message.old != null && Object.hasOwnProperty.call(message, "old"))
            writer.uint32(/* id 5, wireType 2 =*/42).string(message.old);
        if (message["new"] != null && Object.hasOwnProperty.call(message, "new"))
            writer.uint32(/* id 6, wireType 2 =*/50).string(message["new"]);
        if (message.event_at != null && Object.hasOwnProperty.call(message, "event_at"))
            writer.uint32(/* id 7, wireType 2 =*/58).string(message.event_at);
        if (message.file_id != null && Object.hasOwnProperty.call(message, "file_id"))
            writer.uint32(/* id 8, wireType 0 =*/64).int64(message.file_id);
        return writer;
    };

    /**
     * Decodes an EventListItem message from the specified reader or buffer.
     * @function decode
     * @memberof EventListItem
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {EventListItem} EventListItem
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    EventListItem.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.EventListItem();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.id = reader.int64();
                break;
            case 2:
                message.action = reader.int32();
                break;
            case 3:
                message.username = reader.string();
                break;
            case 4:
                message.message = reader.string();
                break;
            case 5:
                message.old = reader.string();
                break;
            case 6:
                message["new"] = reader.string();
                break;
            case 7:
                message.event_at = reader.string();
                break;
            case 8:
                message.file_id = reader.int64();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return EventListItem;
})();

export const EventListResponse = $root.EventListResponse = (() => {

    /**
     * Properties of an EventListResponse.
     * @exports IEventListResponse
     * @interface IEventListResponse
     * @property {number|null} [page] EventListResponse page
     * @property {number|null} [page_size] EventListResponse page_size
     * @property {Array.<EventListItem>|null} [items] EventListResponse items
     * @property {number|null} [count] EventListResponse count
     */

    /**
     * Constructs a new EventListResponse.
     * @exports EventListResponse
     * @classdesc Represents an EventListResponse.
     * @implements IEventListResponse
     * @constructor
     * @param {IEventListResponse=} [properties] Properties to set
     */
    function EventListResponse(properties) {
        this.items = [];
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * EventListResponse page.
     * @member {number} page
     * @memberof EventListResponse
     * @instance
     */
    EventListResponse.prototype.page = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * EventListResponse page_size.
     * @member {number} page_size
     * @memberof EventListResponse
     * @instance
     */
    EventListResponse.prototype.page_size = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * EventListResponse items.
     * @member {Array.<EventListItem>} items
     * @memberof EventListResponse
     * @instance
     */
    EventListResponse.prototype.items = $util.emptyArray;

    /**
     * EventListResponse count.
     * @member {number} count
     * @memberof EventListResponse
     * @instance
     */
    EventListResponse.prototype.count = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * Encodes the specified EventListResponse message. Does not implicitly {@link EventListResponse.verify|verify} messages.
     * @function encode
     * @memberof EventListResponse
     * @static
     * @param {EventListResponse} message EventListResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    EventListResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.page != null && Object.hasOwnProperty.call(message, "page"))
            writer.uint32(/* id 1, wireType 0 =*/8).int64(message.page);
        if (message.page_size != null && Object.hasOwnProperty.call(message, "page_size"))
            writer.uint32(/* id 2, wireType 0 =*/16).int64(message.page_size);
        if (message.items != null && message.items.length)
            for (let i = 0; i < message.items.length; ++i)
                $root.EventListItem.encode(message.items[i], writer.uint32(/* id 3, wireType 2 =*/26).fork()).ldelim();
        if (message.count != null && Object.hasOwnProperty.call(message, "count"))
            writer.uint32(/* id 4, wireType 0 =*/32).int64(message.count);
        return writer;
    };

    /**
     * Decodes an EventListResponse message from the specified reader or buffer.
     * @function decode
     * @memberof EventListResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {EventListResponse} EventListResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    EventListResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.EventListResponse();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.page = reader.int64();
                break;
            case 2:
                message.page_size = reader.int64();
                break;
            case 3:
                if (!(message.items && message.items.length))
                    message.items = [];
                message.items.push($root.EventListItem.decode(reader, reader.uint32()));
                break;
            case 4:
                message.count = reader.int64();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return EventListResponse;
})();

export const Event = $root.Event = (() => {

    /**
     * Constructs a new Event service.
     * @exports Event
     * @classdesc Represents an Event
     * @extends $protobuf.rpc.Service
     * @constructor
     * @param {$protobuf.RPCImpl} rpcImpl RPC implementation
     * @param {boolean} [requestDelimited=false] Whether requests are length-delimited
     * @param {boolean} [responseDelimited=false] Whether responses are length-delimited
     */
    function Event(rpcImpl, requestDelimited, responseDelimited) {
        $protobuf.rpc.Service.call(this, rpcImpl, requestDelimited, responseDelimited);
    }

    (Event.prototype = Object.create($protobuf.rpc.Service.prototype)).constructor = Event;

    /**
     * Callback as used by {@link Event#list}.
     * @memberof Event
     * @typedef ListCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {EventListResponse} [response] EventListResponse
     */

    /**
     * Calls List.
     * @function list
     * @memberof Event
     * @instance
     * @param {EventListRequest} request EventListRequest message or plain object
     * @param {Event.ListCallback} callback Node-style callback called with the error, if any, and EventListResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Event.prototype.list = function list(request, callback) {
        return this.rpcCall(list, $root.EventListRequest, $root.EventListResponse, request, callback);
    }, "name", { value: "List" });

    /**
     * Calls List.
     * @function list
     * @memberof Event
     * @instance
     * @param {EventListRequest} request EventListRequest message or plain object
     * @returns {Promise<EventListResponse>} Promise
     * @variation 2
     */

    return Event;
})();

export const FileDeleteRequest = $root.FileDeleteRequest = (() => {

    /**
     * Properties of a FileDeleteRequest.
     * @exports IFileDeleteRequest
     * @interface IFileDeleteRequest
     * @property {number|null} [id] FileDeleteRequest id
     */

    /**
     * Constructs a new FileDeleteRequest.
     * @exports FileDeleteRequest
     * @classdesc Represents a FileDeleteRequest.
     * @implements IFileDeleteRequest
     * @constructor
     * @param {IFileDeleteRequest=} [properties] Properties to set
     */
    function FileDeleteRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * FileDeleteRequest id.
     * @member {number} id
     * @memberof FileDeleteRequest
     * @instance
     */
    FileDeleteRequest.prototype.id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * Encodes the specified FileDeleteRequest message. Does not implicitly {@link FileDeleteRequest.verify|verify} messages.
     * @function encode
     * @memberof FileDeleteRequest
     * @static
     * @param {FileDeleteRequest} message FileDeleteRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    FileDeleteRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.id != null && Object.hasOwnProperty.call(message, "id"))
            writer.uint32(/* id 1, wireType 0 =*/8).int64(message.id);
        return writer;
    };

    /**
     * Decodes a FileDeleteRequest message from the specified reader or buffer.
     * @function decode
     * @memberof FileDeleteRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {FileDeleteRequest} FileDeleteRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    FileDeleteRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.FileDeleteRequest();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.id = reader.int64();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return FileDeleteRequest;
})();

export const FileDeleteResponse = $root.FileDeleteResponse = (() => {

    /**
     * Properties of a FileDeleteResponse.
     * @exports IFileDeleteResponse
     * @interface IFileDeleteResponse
     * @property {File|null} [file] FileDeleteResponse file
     */

    /**
     * Constructs a new FileDeleteResponse.
     * @exports FileDeleteResponse
     * @classdesc Represents a FileDeleteResponse.
     * @implements IFileDeleteResponse
     * @constructor
     * @param {IFileDeleteResponse=} [properties] Properties to set
     */
    function FileDeleteResponse(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * FileDeleteResponse file.
     * @member {File|null|undefined} file
     * @memberof FileDeleteResponse
     * @instance
     */
    FileDeleteResponse.prototype.file = null;

    /**
     * Encodes the specified FileDeleteResponse message. Does not implicitly {@link FileDeleteResponse.verify|verify} messages.
     * @function encode
     * @memberof FileDeleteResponse
     * @static
     * @param {FileDeleteResponse} message FileDeleteResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    FileDeleteResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.file != null && Object.hasOwnProperty.call(message, "file"))
            $root.File.encode(message.file, writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
        return writer;
    };

    /**
     * Decodes a FileDeleteResponse message from the specified reader or buffer.
     * @function decode
     * @memberof FileDeleteResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {FileDeleteResponse} FileDeleteResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    FileDeleteResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.FileDeleteResponse();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.file = $root.File.decode(reader, reader.uint32());
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return FileDeleteResponse;
})();

export const DeleteUndocumentedFilesRequest = $root.DeleteUndocumentedFilesRequest = (() => {

    /**
     * Properties of a DeleteUndocumentedFilesRequest.
     * @exports IDeleteUndocumentedFilesRequest
     * @interface IDeleteUndocumentedFilesRequest
     */

    /**
     * Constructs a new DeleteUndocumentedFilesRequest.
     * @exports DeleteUndocumentedFilesRequest
     * @classdesc Represents a DeleteUndocumentedFilesRequest.
     * @implements IDeleteUndocumentedFilesRequest
     * @constructor
     * @param {IDeleteUndocumentedFilesRequest=} [properties] Properties to set
     */
    function DeleteUndocumentedFilesRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * Encodes the specified DeleteUndocumentedFilesRequest message. Does not implicitly {@link DeleteUndocumentedFilesRequest.verify|verify} messages.
     * @function encode
     * @memberof DeleteUndocumentedFilesRequest
     * @static
     * @param {DeleteUndocumentedFilesRequest} message DeleteUndocumentedFilesRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    DeleteUndocumentedFilesRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        return writer;
    };

    /**
     * Decodes a DeleteUndocumentedFilesRequest message from the specified reader or buffer.
     * @function decode
     * @memberof DeleteUndocumentedFilesRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {DeleteUndocumentedFilesRequest} DeleteUndocumentedFilesRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    DeleteUndocumentedFilesRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.DeleteUndocumentedFilesRequest();
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

    return DeleteUndocumentedFilesRequest;
})();

export const File = $root.File = (() => {

    /**
     * Properties of a File.
     * @exports IFile
     * @interface IFile
     * @property {string|null} [path] File path
     * @property {string|null} [humanize_size] File humanize_size
     * @property {number|null} [size] File size
     * @property {string|null} [upload_by] File upload_by
     */

    /**
     * Constructs a new File.
     * @exports File
     * @classdesc Represents a File.
     * @implements IFile
     * @constructor
     * @param {IFile=} [properties] Properties to set
     */
    function File(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * File path.
     * @member {string} path
     * @memberof File
     * @instance
     */
    File.prototype.path = "";

    /**
     * File humanize_size.
     * @member {string} humanize_size
     * @memberof File
     * @instance
     */
    File.prototype.humanize_size = "";

    /**
     * File size.
     * @member {number} size
     * @memberof File
     * @instance
     */
    File.prototype.size = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * File upload_by.
     * @member {string} upload_by
     * @memberof File
     * @instance
     */
    File.prototype.upload_by = "";

    /**
     * Encodes the specified File message. Does not implicitly {@link File.verify|verify} messages.
     * @function encode
     * @memberof File
     * @static
     * @param {File} message File message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    File.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.path != null && Object.hasOwnProperty.call(message, "path"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.path);
        if (message.humanize_size != null && Object.hasOwnProperty.call(message, "humanize_size"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.humanize_size);
        if (message.size != null && Object.hasOwnProperty.call(message, "size"))
            writer.uint32(/* id 3, wireType 0 =*/24).int64(message.size);
        if (message.upload_by != null && Object.hasOwnProperty.call(message, "upload_by"))
            writer.uint32(/* id 4, wireType 2 =*/34).string(message.upload_by);
        return writer;
    };

    /**
     * Decodes a File message from the specified reader or buffer.
     * @function decode
     * @memberof File
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {File} File
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    File.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.File();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.path = reader.string();
                break;
            case 2:
                message.humanize_size = reader.string();
                break;
            case 3:
                message.size = reader.int64();
                break;
            case 4:
                message.upload_by = reader.string();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return File;
})();

export const DeleteUndocumentedFilesResponse = $root.DeleteUndocumentedFilesResponse = (() => {

    /**
     * Properties of a DeleteUndocumentedFilesResponse.
     * @exports IDeleteUndocumentedFilesResponse
     * @interface IDeleteUndocumentedFilesResponse
     * @property {Array.<File>|null} [files] DeleteUndocumentedFilesResponse files
     */

    /**
     * Constructs a new DeleteUndocumentedFilesResponse.
     * @exports DeleteUndocumentedFilesResponse
     * @classdesc Represents a DeleteUndocumentedFilesResponse.
     * @implements IDeleteUndocumentedFilesResponse
     * @constructor
     * @param {IDeleteUndocumentedFilesResponse=} [properties] Properties to set
     */
    function DeleteUndocumentedFilesResponse(properties) {
        this.files = [];
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * DeleteUndocumentedFilesResponse files.
     * @member {Array.<File>} files
     * @memberof DeleteUndocumentedFilesResponse
     * @instance
     */
    DeleteUndocumentedFilesResponse.prototype.files = $util.emptyArray;

    /**
     * Encodes the specified DeleteUndocumentedFilesResponse message. Does not implicitly {@link DeleteUndocumentedFilesResponse.verify|verify} messages.
     * @function encode
     * @memberof DeleteUndocumentedFilesResponse
     * @static
     * @param {DeleteUndocumentedFilesResponse} message DeleteUndocumentedFilesResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    DeleteUndocumentedFilesResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.files != null && message.files.length)
            for (let i = 0; i < message.files.length; ++i)
                $root.File.encode(message.files[i], writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
        return writer;
    };

    /**
     * Decodes a DeleteUndocumentedFilesResponse message from the specified reader or buffer.
     * @function decode
     * @memberof DeleteUndocumentedFilesResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {DeleteUndocumentedFilesResponse} DeleteUndocumentedFilesResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    DeleteUndocumentedFilesResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.DeleteUndocumentedFilesResponse();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                if (!(message.files && message.files.length))
                    message.files = [];
                message.files.push($root.File.decode(reader, reader.uint32()));
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return DeleteUndocumentedFilesResponse;
})();

export const DiskInfoRequest = $root.DiskInfoRequest = (() => {

    /**
     * Properties of a DiskInfoRequest.
     * @exports IDiskInfoRequest
     * @interface IDiskInfoRequest
     */

    /**
     * Constructs a new DiskInfoRequest.
     * @exports DiskInfoRequest
     * @classdesc Represents a DiskInfoRequest.
     * @implements IDiskInfoRequest
     * @constructor
     * @param {IDiskInfoRequest=} [properties] Properties to set
     */
    function DiskInfoRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * Encodes the specified DiskInfoRequest message. Does not implicitly {@link DiskInfoRequest.verify|verify} messages.
     * @function encode
     * @memberof DiskInfoRequest
     * @static
     * @param {DiskInfoRequest} message DiskInfoRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    DiskInfoRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        return writer;
    };

    /**
     * Decodes a DiskInfoRequest message from the specified reader or buffer.
     * @function decode
     * @memberof DiskInfoRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {DiskInfoRequest} DiskInfoRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    DiskInfoRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.DiskInfoRequest();
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

    return DiskInfoRequest;
})();

export const DiskInfoResponse = $root.DiskInfoResponse = (() => {

    /**
     * Properties of a DiskInfoResponse.
     * @exports IDiskInfoResponse
     * @interface IDiskInfoResponse
     * @property {number|null} [usage] DiskInfoResponse usage
     * @property {string|null} [humanize_usage] DiskInfoResponse humanize_usage
     */

    /**
     * Constructs a new DiskInfoResponse.
     * @exports DiskInfoResponse
     * @classdesc Represents a DiskInfoResponse.
     * @implements IDiskInfoResponse
     * @constructor
     * @param {IDiskInfoResponse=} [properties] Properties to set
     */
    function DiskInfoResponse(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * DiskInfoResponse usage.
     * @member {number} usage
     * @memberof DiskInfoResponse
     * @instance
     */
    DiskInfoResponse.prototype.usage = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * DiskInfoResponse humanize_usage.
     * @member {string} humanize_usage
     * @memberof DiskInfoResponse
     * @instance
     */
    DiskInfoResponse.prototype.humanize_usage = "";

    /**
     * Encodes the specified DiskInfoResponse message. Does not implicitly {@link DiskInfoResponse.verify|verify} messages.
     * @function encode
     * @memberof DiskInfoResponse
     * @static
     * @param {DiskInfoResponse} message DiskInfoResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    DiskInfoResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.usage != null && Object.hasOwnProperty.call(message, "usage"))
            writer.uint32(/* id 1, wireType 0 =*/8).int64(message.usage);
        if (message.humanize_usage != null && Object.hasOwnProperty.call(message, "humanize_usage"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.humanize_usage);
        return writer;
    };

    /**
     * Decodes a DiskInfoResponse message from the specified reader or buffer.
     * @function decode
     * @memberof DiskInfoResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {DiskInfoResponse} DiskInfoResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    DiskInfoResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.DiskInfoResponse();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.usage = reader.int64();
                break;
            case 2:
                message.humanize_usage = reader.string();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return DiskInfoResponse;
})();

export const FileListRequest = $root.FileListRequest = (() => {

    /**
     * Properties of a FileListRequest.
     * @exports IFileListRequest
     * @interface IFileListRequest
     * @property {number|null} [page] FileListRequest page
     * @property {number|null} [page_size] FileListRequest page_size
     * @property {boolean|null} [without_deleted] FileListRequest without_deleted
     */

    /**
     * Constructs a new FileListRequest.
     * @exports FileListRequest
     * @classdesc Represents a FileListRequest.
     * @implements IFileListRequest
     * @constructor
     * @param {IFileListRequest=} [properties] Properties to set
     */
    function FileListRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * FileListRequest page.
     * @member {number} page
     * @memberof FileListRequest
     * @instance
     */
    FileListRequest.prototype.page = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * FileListRequest page_size.
     * @member {number} page_size
     * @memberof FileListRequest
     * @instance
     */
    FileListRequest.prototype.page_size = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * FileListRequest without_deleted.
     * @member {boolean} without_deleted
     * @memberof FileListRequest
     * @instance
     */
    FileListRequest.prototype.without_deleted = false;

    /**
     * Encodes the specified FileListRequest message. Does not implicitly {@link FileListRequest.verify|verify} messages.
     * @function encode
     * @memberof FileListRequest
     * @static
     * @param {FileListRequest} message FileListRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    FileListRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.page != null && Object.hasOwnProperty.call(message, "page"))
            writer.uint32(/* id 1, wireType 0 =*/8).int64(message.page);
        if (message.page_size != null && Object.hasOwnProperty.call(message, "page_size"))
            writer.uint32(/* id 2, wireType 0 =*/16).int64(message.page_size);
        if (message.without_deleted != null && Object.hasOwnProperty.call(message, "without_deleted"))
            writer.uint32(/* id 3, wireType 0 =*/24).bool(message.without_deleted);
        return writer;
    };

    /**
     * Decodes a FileListRequest message from the specified reader or buffer.
     * @function decode
     * @memberof FileListRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {FileListRequest} FileListRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    FileListRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.FileListRequest();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.page = reader.int64();
                break;
            case 2:
                message.page_size = reader.int64();
                break;
            case 3:
                message.without_deleted = reader.bool();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return FileListRequest;
})();

export const FileListResponse = $root.FileListResponse = (() => {

    /**
     * Properties of a FileListResponse.
     * @exports IFileListResponse
     * @interface IFileListResponse
     * @property {number|null} [page] FileListResponse page
     * @property {number|null} [page_size] FileListResponse page_size
     * @property {Array.<FileModel>|null} [items] FileListResponse items
     * @property {number|null} [count] FileListResponse count
     */

    /**
     * Constructs a new FileListResponse.
     * @exports FileListResponse
     * @classdesc Represents a FileListResponse.
     * @implements IFileListResponse
     * @constructor
     * @param {IFileListResponse=} [properties] Properties to set
     */
    function FileListResponse(properties) {
        this.items = [];
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * FileListResponse page.
     * @member {number} page
     * @memberof FileListResponse
     * @instance
     */
    FileListResponse.prototype.page = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * FileListResponse page_size.
     * @member {number} page_size
     * @memberof FileListResponse
     * @instance
     */
    FileListResponse.prototype.page_size = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * FileListResponse items.
     * @member {Array.<FileModel>} items
     * @memberof FileListResponse
     * @instance
     */
    FileListResponse.prototype.items = $util.emptyArray;

    /**
     * FileListResponse count.
     * @member {number} count
     * @memberof FileListResponse
     * @instance
     */
    FileListResponse.prototype.count = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * Encodes the specified FileListResponse message. Does not implicitly {@link FileListResponse.verify|verify} messages.
     * @function encode
     * @memberof FileListResponse
     * @static
     * @param {FileListResponse} message FileListResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    FileListResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.page != null && Object.hasOwnProperty.call(message, "page"))
            writer.uint32(/* id 1, wireType 0 =*/8).int64(message.page);
        if (message.page_size != null && Object.hasOwnProperty.call(message, "page_size"))
            writer.uint32(/* id 2, wireType 0 =*/16).int64(message.page_size);
        if (message.items != null && message.items.length)
            for (let i = 0; i < message.items.length; ++i)
                $root.FileModel.encode(message.items[i], writer.uint32(/* id 3, wireType 2 =*/26).fork()).ldelim();
        if (message.count != null && Object.hasOwnProperty.call(message, "count"))
            writer.uint32(/* id 4, wireType 0 =*/32).int64(message.count);
        return writer;
    };

    /**
     * Decodes a FileListResponse message from the specified reader or buffer.
     * @function decode
     * @memberof FileListResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {FileListResponse} FileListResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    FileListResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.FileListResponse();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.page = reader.int64();
                break;
            case 2:
                message.page_size = reader.int64();
                break;
            case 3:
                if (!(message.items && message.items.length))
                    message.items = [];
                message.items.push($root.FileModel.decode(reader, reader.uint32()));
                break;
            case 4:
                message.count = reader.int64();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return FileListResponse;
})();

export const FileSvc = $root.FileSvc = (() => {

    /**
     * Constructs a new FileSvc service.
     * @exports FileSvc
     * @classdesc Represents a FileSvc
     * @extends $protobuf.rpc.Service
     * @constructor
     * @param {$protobuf.RPCImpl} rpcImpl RPC implementation
     * @param {boolean} [requestDelimited=false] Whether requests are length-delimited
     * @param {boolean} [responseDelimited=false] Whether responses are length-delimited
     */
    function FileSvc(rpcImpl, requestDelimited, responseDelimited) {
        $protobuf.rpc.Service.call(this, rpcImpl, requestDelimited, responseDelimited);
    }

    (FileSvc.prototype = Object.create($protobuf.rpc.Service.prototype)).constructor = FileSvc;

    /**
     * Callback as used by {@link FileSvc#list}.
     * @memberof FileSvc
     * @typedef ListCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {FileListResponse} [response] FileListResponse
     */

    /**
     * Calls List.
     * @function list
     * @memberof FileSvc
     * @instance
     * @param {FileListRequest} request FileListRequest message or plain object
     * @param {FileSvc.ListCallback} callback Node-style callback called with the error, if any, and FileListResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(FileSvc.prototype.list = function list(request, callback) {
        return this.rpcCall(list, $root.FileListRequest, $root.FileListResponse, request, callback);
    }, "name", { value: "List" });

    /**
     * Calls List.
     * @function list
     * @memberof FileSvc
     * @instance
     * @param {FileListRequest} request FileListRequest message or plain object
     * @returns {Promise<FileListResponse>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link FileSvc#delete_}.
     * @memberof FileSvc
     * @typedef DeleteCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {FileDeleteResponse} [response] FileDeleteResponse
     */

    /**
     * Calls Delete.
     * @function delete
     * @memberof FileSvc
     * @instance
     * @param {FileDeleteRequest} request FileDeleteRequest message or plain object
     * @param {FileSvc.DeleteCallback} callback Node-style callback called with the error, if any, and FileDeleteResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(FileSvc.prototype["delete"] = function delete_(request, callback) {
        return this.rpcCall(delete_, $root.FileDeleteRequest, $root.FileDeleteResponse, request, callback);
    }, "name", { value: "Delete" });

    /**
     * Calls Delete.
     * @function delete
     * @memberof FileSvc
     * @instance
     * @param {FileDeleteRequest} request FileDeleteRequest message or plain object
     * @returns {Promise<FileDeleteResponse>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link FileSvc#deleteUndocumentedFiles}.
     * @memberof FileSvc
     * @typedef DeleteUndocumentedFilesCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {DeleteUndocumentedFilesResponse} [response] DeleteUndocumentedFilesResponse
     */

    /**
     * Calls DeleteUndocumentedFiles.
     * @function deleteUndocumentedFiles
     * @memberof FileSvc
     * @instance
     * @param {DeleteUndocumentedFilesRequest} request DeleteUndocumentedFilesRequest message or plain object
     * @param {FileSvc.DeleteUndocumentedFilesCallback} callback Node-style callback called with the error, if any, and DeleteUndocumentedFilesResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(FileSvc.prototype.deleteUndocumentedFiles = function deleteUndocumentedFiles(request, callback) {
        return this.rpcCall(deleteUndocumentedFiles, $root.DeleteUndocumentedFilesRequest, $root.DeleteUndocumentedFilesResponse, request, callback);
    }, "name", { value: "DeleteUndocumentedFiles" });

    /**
     * Calls DeleteUndocumentedFiles.
     * @function deleteUndocumentedFiles
     * @memberof FileSvc
     * @instance
     * @param {DeleteUndocumentedFilesRequest} request DeleteUndocumentedFilesRequest message or plain object
     * @returns {Promise<DeleteUndocumentedFilesResponse>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link FileSvc#diskInfo}.
     * @memberof FileSvc
     * @typedef DiskInfoCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {DiskInfoResponse} [response] DiskInfoResponse
     */

    /**
     * Calls DiskInfo.
     * @function diskInfo
     * @memberof FileSvc
     * @instance
     * @param {DiskInfoRequest} request DiskInfoRequest message or plain object
     * @param {FileSvc.DiskInfoCallback} callback Node-style callback called with the error, if any, and DiskInfoResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(FileSvc.prototype.diskInfo = function diskInfo(request, callback) {
        return this.rpcCall(diskInfo, $root.DiskInfoRequest, $root.DiskInfoResponse, request, callback);
    }, "name", { value: "DiskInfo" });

    /**
     * Calls DiskInfo.
     * @function diskInfo
     * @memberof FileSvc
     * @instance
     * @param {DiskInfoRequest} request DiskInfoRequest message or plain object
     * @returns {Promise<DiskInfoResponse>} Promise
     * @variation 2
     */

    return FileSvc;
})();

export const GitDestroyRequest = $root.GitDestroyRequest = (() => {

    /**
     * Properties of a GitDestroyRequest.
     * @exports IGitDestroyRequest
     * @interface IGitDestroyRequest
     * @property {string|null} [namespace_id] GitDestroyRequest namespace_id
     * @property {string|null} [project_id] GitDestroyRequest project_id
     */

    /**
     * Constructs a new GitDestroyRequest.
     * @exports GitDestroyRequest
     * @classdesc Represents a GitDestroyRequest.
     * @implements IGitDestroyRequest
     * @constructor
     * @param {IGitDestroyRequest=} [properties] Properties to set
     */
    function GitDestroyRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * GitDestroyRequest namespace_id.
     * @member {string} namespace_id
     * @memberof GitDestroyRequest
     * @instance
     */
    GitDestroyRequest.prototype.namespace_id = "";

    /**
     * GitDestroyRequest project_id.
     * @member {string} project_id
     * @memberof GitDestroyRequest
     * @instance
     */
    GitDestroyRequest.prototype.project_id = "";

    /**
     * Encodes the specified GitDestroyRequest message. Does not implicitly {@link GitDestroyRequest.verify|verify} messages.
     * @function encode
     * @memberof GitDestroyRequest
     * @static
     * @param {GitDestroyRequest} message GitDestroyRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    GitDestroyRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.namespace_id != null && Object.hasOwnProperty.call(message, "namespace_id"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.namespace_id);
        if (message.project_id != null && Object.hasOwnProperty.call(message, "project_id"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.project_id);
        return writer;
    };

    /**
     * Decodes a GitDestroyRequest message from the specified reader or buffer.
     * @function decode
     * @memberof GitDestroyRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {GitDestroyRequest} GitDestroyRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    GitDestroyRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.GitDestroyRequest();
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

    return GitDestroyRequest;
})();

export const GitEnableProjectRequest = $root.GitEnableProjectRequest = (() => {

    /**
     * Properties of a GitEnableProjectRequest.
     * @exports IGitEnableProjectRequest
     * @interface IGitEnableProjectRequest
     * @property {string|null} [git_project_id] GitEnableProjectRequest git_project_id
     */

    /**
     * Constructs a new GitEnableProjectRequest.
     * @exports GitEnableProjectRequest
     * @classdesc Represents a GitEnableProjectRequest.
     * @implements IGitEnableProjectRequest
     * @constructor
     * @param {IGitEnableProjectRequest=} [properties] Properties to set
     */
    function GitEnableProjectRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * GitEnableProjectRequest git_project_id.
     * @member {string} git_project_id
     * @memberof GitEnableProjectRequest
     * @instance
     */
    GitEnableProjectRequest.prototype.git_project_id = "";

    /**
     * Encodes the specified GitEnableProjectRequest message. Does not implicitly {@link GitEnableProjectRequest.verify|verify} messages.
     * @function encode
     * @memberof GitEnableProjectRequest
     * @static
     * @param {GitEnableProjectRequest} message GitEnableProjectRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    GitEnableProjectRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.git_project_id != null && Object.hasOwnProperty.call(message, "git_project_id"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.git_project_id);
        return writer;
    };

    /**
     * Decodes a GitEnableProjectRequest message from the specified reader or buffer.
     * @function decode
     * @memberof GitEnableProjectRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {GitEnableProjectRequest} GitEnableProjectRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    GitEnableProjectRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.GitEnableProjectRequest();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.git_project_id = reader.string();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return GitEnableProjectRequest;
})();

export const GitDisableProjectRequest = $root.GitDisableProjectRequest = (() => {

    /**
     * Properties of a GitDisableProjectRequest.
     * @exports IGitDisableProjectRequest
     * @interface IGitDisableProjectRequest
     * @property {string|null} [git_project_id] GitDisableProjectRequest git_project_id
     */

    /**
     * Constructs a new GitDisableProjectRequest.
     * @exports GitDisableProjectRequest
     * @classdesc Represents a GitDisableProjectRequest.
     * @implements IGitDisableProjectRequest
     * @constructor
     * @param {IGitDisableProjectRequest=} [properties] Properties to set
     */
    function GitDisableProjectRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * GitDisableProjectRequest git_project_id.
     * @member {string} git_project_id
     * @memberof GitDisableProjectRequest
     * @instance
     */
    GitDisableProjectRequest.prototype.git_project_id = "";

    /**
     * Encodes the specified GitDisableProjectRequest message. Does not implicitly {@link GitDisableProjectRequest.verify|verify} messages.
     * @function encode
     * @memberof GitDisableProjectRequest
     * @static
     * @param {GitDisableProjectRequest} message GitDisableProjectRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    GitDisableProjectRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.git_project_id != null && Object.hasOwnProperty.call(message, "git_project_id"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.git_project_id);
        return writer;
    };

    /**
     * Decodes a GitDisableProjectRequest message from the specified reader or buffer.
     * @function decode
     * @memberof GitDisableProjectRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {GitDisableProjectRequest} GitDisableProjectRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    GitDisableProjectRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.GitDisableProjectRequest();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.git_project_id = reader.string();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return GitDisableProjectRequest;
})();

export const GitProjectItem = $root.GitProjectItem = (() => {

    /**
     * Properties of a GitProjectItem.
     * @exports IGitProjectItem
     * @interface IGitProjectItem
     * @property {number|null} [id] GitProjectItem id
     * @property {string|null} [name] GitProjectItem name
     * @property {string|null} [path] GitProjectItem path
     * @property {string|null} [web_url] GitProjectItem web_url
     * @property {string|null} [avatar_url] GitProjectItem avatar_url
     * @property {string|null} [description] GitProjectItem description
     * @property {boolean|null} [enabled] GitProjectItem enabled
     * @property {boolean|null} [global_enabled] GitProjectItem global_enabled
     */

    /**
     * Constructs a new GitProjectItem.
     * @exports GitProjectItem
     * @classdesc Represents a GitProjectItem.
     * @implements IGitProjectItem
     * @constructor
     * @param {IGitProjectItem=} [properties] Properties to set
     */
    function GitProjectItem(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * GitProjectItem id.
     * @member {number} id
     * @memberof GitProjectItem
     * @instance
     */
    GitProjectItem.prototype.id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * GitProjectItem name.
     * @member {string} name
     * @memberof GitProjectItem
     * @instance
     */
    GitProjectItem.prototype.name = "";

    /**
     * GitProjectItem path.
     * @member {string} path
     * @memberof GitProjectItem
     * @instance
     */
    GitProjectItem.prototype.path = "";

    /**
     * GitProjectItem web_url.
     * @member {string} web_url
     * @memberof GitProjectItem
     * @instance
     */
    GitProjectItem.prototype.web_url = "";

    /**
     * GitProjectItem avatar_url.
     * @member {string} avatar_url
     * @memberof GitProjectItem
     * @instance
     */
    GitProjectItem.prototype.avatar_url = "";

    /**
     * GitProjectItem description.
     * @member {string} description
     * @memberof GitProjectItem
     * @instance
     */
    GitProjectItem.prototype.description = "";

    /**
     * GitProjectItem enabled.
     * @member {boolean} enabled
     * @memberof GitProjectItem
     * @instance
     */
    GitProjectItem.prototype.enabled = false;

    /**
     * GitProjectItem global_enabled.
     * @member {boolean} global_enabled
     * @memberof GitProjectItem
     * @instance
     */
    GitProjectItem.prototype.global_enabled = false;

    /**
     * Encodes the specified GitProjectItem message. Does not implicitly {@link GitProjectItem.verify|verify} messages.
     * @function encode
     * @memberof GitProjectItem
     * @static
     * @param {GitProjectItem} message GitProjectItem message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    GitProjectItem.encode = function encode(message, writer) {
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
     * Decodes a GitProjectItem message from the specified reader or buffer.
     * @function decode
     * @memberof GitProjectItem
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {GitProjectItem} GitProjectItem
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    GitProjectItem.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.GitProjectItem();
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

    return GitProjectItem;
})();

export const GitAllProjectsResponse = $root.GitAllProjectsResponse = (() => {

    /**
     * Properties of a GitAllProjectsResponse.
     * @exports IGitAllProjectsResponse
     * @interface IGitAllProjectsResponse
     * @property {Array.<GitProjectItem>|null} [data] GitAllProjectsResponse data
     */

    /**
     * Constructs a new GitAllProjectsResponse.
     * @exports GitAllProjectsResponse
     * @classdesc Represents a GitAllProjectsResponse.
     * @implements IGitAllProjectsResponse
     * @constructor
     * @param {IGitAllProjectsResponse=} [properties] Properties to set
     */
    function GitAllProjectsResponse(properties) {
        this.data = [];
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * GitAllProjectsResponse data.
     * @member {Array.<GitProjectItem>} data
     * @memberof GitAllProjectsResponse
     * @instance
     */
    GitAllProjectsResponse.prototype.data = $util.emptyArray;

    /**
     * Encodes the specified GitAllProjectsResponse message. Does not implicitly {@link GitAllProjectsResponse.verify|verify} messages.
     * @function encode
     * @memberof GitAllProjectsResponse
     * @static
     * @param {GitAllProjectsResponse} message GitAllProjectsResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    GitAllProjectsResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.data != null && message.data.length)
            for (let i = 0; i < message.data.length; ++i)
                $root.GitProjectItem.encode(message.data[i], writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
        return writer;
    };

    /**
     * Decodes a GitAllProjectsResponse message from the specified reader or buffer.
     * @function decode
     * @memberof GitAllProjectsResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {GitAllProjectsResponse} GitAllProjectsResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    GitAllProjectsResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.GitAllProjectsResponse();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                if (!(message.data && message.data.length))
                    message.data = [];
                message.data.push($root.GitProjectItem.decode(reader, reader.uint32()));
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return GitAllProjectsResponse;
})();

export const GitOption = $root.GitOption = (() => {

    /**
     * Properties of a GitOption.
     * @exports IGitOption
     * @interface IGitOption
     * @property {string|null} [value] GitOption value
     * @property {string|null} [label] GitOption label
     * @property {string|null} [type] GitOption type
     * @property {boolean|null} [isLeaf] GitOption isLeaf
     * @property {string|null} [projectId] GitOption projectId
     * @property {string|null} [branch] GitOption branch
     */

    /**
     * Constructs a new GitOption.
     * @exports GitOption
     * @classdesc Represents a GitOption.
     * @implements IGitOption
     * @constructor
     * @param {IGitOption=} [properties] Properties to set
     */
    function GitOption(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * GitOption value.
     * @member {string} value
     * @memberof GitOption
     * @instance
     */
    GitOption.prototype.value = "";

    /**
     * GitOption label.
     * @member {string} label
     * @memberof GitOption
     * @instance
     */
    GitOption.prototype.label = "";

    /**
     * GitOption type.
     * @member {string} type
     * @memberof GitOption
     * @instance
     */
    GitOption.prototype.type = "";

    /**
     * GitOption isLeaf.
     * @member {boolean} isLeaf
     * @memberof GitOption
     * @instance
     */
    GitOption.prototype.isLeaf = false;

    /**
     * GitOption projectId.
     * @member {string} projectId
     * @memberof GitOption
     * @instance
     */
    GitOption.prototype.projectId = "";

    /**
     * GitOption branch.
     * @member {string} branch
     * @memberof GitOption
     * @instance
     */
    GitOption.prototype.branch = "";

    /**
     * Encodes the specified GitOption message. Does not implicitly {@link GitOption.verify|verify} messages.
     * @function encode
     * @memberof GitOption
     * @static
     * @param {GitOption} message GitOption message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    GitOption.encode = function encode(message, writer) {
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
     * Decodes a GitOption message from the specified reader or buffer.
     * @function decode
     * @memberof GitOption
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {GitOption} GitOption
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    GitOption.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.GitOption();
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

    return GitOption;
})();

export const GitProjectOptionsResponse = $root.GitProjectOptionsResponse = (() => {

    /**
     * Properties of a GitProjectOptionsResponse.
     * @exports IGitProjectOptionsResponse
     * @interface IGitProjectOptionsResponse
     * @property {Array.<GitOption>|null} [data] GitProjectOptionsResponse data
     */

    /**
     * Constructs a new GitProjectOptionsResponse.
     * @exports GitProjectOptionsResponse
     * @classdesc Represents a GitProjectOptionsResponse.
     * @implements IGitProjectOptionsResponse
     * @constructor
     * @param {IGitProjectOptionsResponse=} [properties] Properties to set
     */
    function GitProjectOptionsResponse(properties) {
        this.data = [];
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * GitProjectOptionsResponse data.
     * @member {Array.<GitOption>} data
     * @memberof GitProjectOptionsResponse
     * @instance
     */
    GitProjectOptionsResponse.prototype.data = $util.emptyArray;

    /**
     * Encodes the specified GitProjectOptionsResponse message. Does not implicitly {@link GitProjectOptionsResponse.verify|verify} messages.
     * @function encode
     * @memberof GitProjectOptionsResponse
     * @static
     * @param {GitProjectOptionsResponse} message GitProjectOptionsResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    GitProjectOptionsResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.data != null && message.data.length)
            for (let i = 0; i < message.data.length; ++i)
                $root.GitOption.encode(message.data[i], writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
        return writer;
    };

    /**
     * Decodes a GitProjectOptionsResponse message from the specified reader or buffer.
     * @function decode
     * @memberof GitProjectOptionsResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {GitProjectOptionsResponse} GitProjectOptionsResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    GitProjectOptionsResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.GitProjectOptionsResponse();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                if (!(message.data && message.data.length))
                    message.data = [];
                message.data.push($root.GitOption.decode(reader, reader.uint32()));
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return GitProjectOptionsResponse;
})();

export const GitBranchOptionsRequest = $root.GitBranchOptionsRequest = (() => {

    /**
     * Properties of a GitBranchOptionsRequest.
     * @exports IGitBranchOptionsRequest
     * @interface IGitBranchOptionsRequest
     * @property {string|null} [project_id] GitBranchOptionsRequest project_id
     * @property {boolean|null} [all] GitBranchOptionsRequest all
     */

    /**
     * Constructs a new GitBranchOptionsRequest.
     * @exports GitBranchOptionsRequest
     * @classdesc Represents a GitBranchOptionsRequest.
     * @implements IGitBranchOptionsRequest
     * @constructor
     * @param {IGitBranchOptionsRequest=} [properties] Properties to set
     */
    function GitBranchOptionsRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * GitBranchOptionsRequest project_id.
     * @member {string} project_id
     * @memberof GitBranchOptionsRequest
     * @instance
     */
    GitBranchOptionsRequest.prototype.project_id = "";

    /**
     * GitBranchOptionsRequest all.
     * @member {boolean} all
     * @memberof GitBranchOptionsRequest
     * @instance
     */
    GitBranchOptionsRequest.prototype.all = false;

    /**
     * Encodes the specified GitBranchOptionsRequest message. Does not implicitly {@link GitBranchOptionsRequest.verify|verify} messages.
     * @function encode
     * @memberof GitBranchOptionsRequest
     * @static
     * @param {GitBranchOptionsRequest} message GitBranchOptionsRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    GitBranchOptionsRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.project_id != null && Object.hasOwnProperty.call(message, "project_id"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.project_id);
        if (message.all != null && Object.hasOwnProperty.call(message, "all"))
            writer.uint32(/* id 2, wireType 0 =*/16).bool(message.all);
        return writer;
    };

    /**
     * Decodes a GitBranchOptionsRequest message from the specified reader or buffer.
     * @function decode
     * @memberof GitBranchOptionsRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {GitBranchOptionsRequest} GitBranchOptionsRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    GitBranchOptionsRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.GitBranchOptionsRequest();
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

    return GitBranchOptionsRequest;
})();

export const GitBranchOptionsResponse = $root.GitBranchOptionsResponse = (() => {

    /**
     * Properties of a GitBranchOptionsResponse.
     * @exports IGitBranchOptionsResponse
     * @interface IGitBranchOptionsResponse
     * @property {Array.<GitOption>|null} [data] GitBranchOptionsResponse data
     */

    /**
     * Constructs a new GitBranchOptionsResponse.
     * @exports GitBranchOptionsResponse
     * @classdesc Represents a GitBranchOptionsResponse.
     * @implements IGitBranchOptionsResponse
     * @constructor
     * @param {IGitBranchOptionsResponse=} [properties] Properties to set
     */
    function GitBranchOptionsResponse(properties) {
        this.data = [];
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * GitBranchOptionsResponse data.
     * @member {Array.<GitOption>} data
     * @memberof GitBranchOptionsResponse
     * @instance
     */
    GitBranchOptionsResponse.prototype.data = $util.emptyArray;

    /**
     * Encodes the specified GitBranchOptionsResponse message. Does not implicitly {@link GitBranchOptionsResponse.verify|verify} messages.
     * @function encode
     * @memberof GitBranchOptionsResponse
     * @static
     * @param {GitBranchOptionsResponse} message GitBranchOptionsResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    GitBranchOptionsResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.data != null && message.data.length)
            for (let i = 0; i < message.data.length; ++i)
                $root.GitOption.encode(message.data[i], writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
        return writer;
    };

    /**
     * Decodes a GitBranchOptionsResponse message from the specified reader or buffer.
     * @function decode
     * @memberof GitBranchOptionsResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {GitBranchOptionsResponse} GitBranchOptionsResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    GitBranchOptionsResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.GitBranchOptionsResponse();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                if (!(message.data && message.data.length))
                    message.data = [];
                message.data.push($root.GitOption.decode(reader, reader.uint32()));
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return GitBranchOptionsResponse;
})();

export const GitCommitOptionsRequest = $root.GitCommitOptionsRequest = (() => {

    /**
     * Properties of a GitCommitOptionsRequest.
     * @exports IGitCommitOptionsRequest
     * @interface IGitCommitOptionsRequest
     * @property {string|null} [project_id] GitCommitOptionsRequest project_id
     * @property {string|null} [branch] GitCommitOptionsRequest branch
     */

    /**
     * Constructs a new GitCommitOptionsRequest.
     * @exports GitCommitOptionsRequest
     * @classdesc Represents a GitCommitOptionsRequest.
     * @implements IGitCommitOptionsRequest
     * @constructor
     * @param {IGitCommitOptionsRequest=} [properties] Properties to set
     */
    function GitCommitOptionsRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * GitCommitOptionsRequest project_id.
     * @member {string} project_id
     * @memberof GitCommitOptionsRequest
     * @instance
     */
    GitCommitOptionsRequest.prototype.project_id = "";

    /**
     * GitCommitOptionsRequest branch.
     * @member {string} branch
     * @memberof GitCommitOptionsRequest
     * @instance
     */
    GitCommitOptionsRequest.prototype.branch = "";

    /**
     * Encodes the specified GitCommitOptionsRequest message. Does not implicitly {@link GitCommitOptionsRequest.verify|verify} messages.
     * @function encode
     * @memberof GitCommitOptionsRequest
     * @static
     * @param {GitCommitOptionsRequest} message GitCommitOptionsRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    GitCommitOptionsRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.project_id != null && Object.hasOwnProperty.call(message, "project_id"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.project_id);
        if (message.branch != null && Object.hasOwnProperty.call(message, "branch"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.branch);
        return writer;
    };

    /**
     * Decodes a GitCommitOptionsRequest message from the specified reader or buffer.
     * @function decode
     * @memberof GitCommitOptionsRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {GitCommitOptionsRequest} GitCommitOptionsRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    GitCommitOptionsRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.GitCommitOptionsRequest();
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

    return GitCommitOptionsRequest;
})();

export const GitCommitOptionsResponse = $root.GitCommitOptionsResponse = (() => {

    /**
     * Properties of a GitCommitOptionsResponse.
     * @exports IGitCommitOptionsResponse
     * @interface IGitCommitOptionsResponse
     * @property {Array.<GitOption>|null} [data] GitCommitOptionsResponse data
     */

    /**
     * Constructs a new GitCommitOptionsResponse.
     * @exports GitCommitOptionsResponse
     * @classdesc Represents a GitCommitOptionsResponse.
     * @implements IGitCommitOptionsResponse
     * @constructor
     * @param {IGitCommitOptionsResponse=} [properties] Properties to set
     */
    function GitCommitOptionsResponse(properties) {
        this.data = [];
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * GitCommitOptionsResponse data.
     * @member {Array.<GitOption>} data
     * @memberof GitCommitOptionsResponse
     * @instance
     */
    GitCommitOptionsResponse.prototype.data = $util.emptyArray;

    /**
     * Encodes the specified GitCommitOptionsResponse message. Does not implicitly {@link GitCommitOptionsResponse.verify|verify} messages.
     * @function encode
     * @memberof GitCommitOptionsResponse
     * @static
     * @param {GitCommitOptionsResponse} message GitCommitOptionsResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    GitCommitOptionsResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.data != null && message.data.length)
            for (let i = 0; i < message.data.length; ++i)
                $root.GitOption.encode(message.data[i], writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
        return writer;
    };

    /**
     * Decodes a GitCommitOptionsResponse message from the specified reader or buffer.
     * @function decode
     * @memberof GitCommitOptionsResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {GitCommitOptionsResponse} GitCommitOptionsResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    GitCommitOptionsResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.GitCommitOptionsResponse();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                if (!(message.data && message.data.length))
                    message.data = [];
                message.data.push($root.GitOption.decode(reader, reader.uint32()));
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return GitCommitOptionsResponse;
})();

export const GitCommitRequest = $root.GitCommitRequest = (() => {

    /**
     * Properties of a GitCommitRequest.
     * @exports IGitCommitRequest
     * @interface IGitCommitRequest
     * @property {string|null} [project_id] GitCommitRequest project_id
     * @property {string|null} [branch] GitCommitRequest branch
     * @property {string|null} [commit] GitCommitRequest commit
     */

    /**
     * Constructs a new GitCommitRequest.
     * @exports GitCommitRequest
     * @classdesc Represents a GitCommitRequest.
     * @implements IGitCommitRequest
     * @constructor
     * @param {IGitCommitRequest=} [properties] Properties to set
     */
    function GitCommitRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * GitCommitRequest project_id.
     * @member {string} project_id
     * @memberof GitCommitRequest
     * @instance
     */
    GitCommitRequest.prototype.project_id = "";

    /**
     * GitCommitRequest branch.
     * @member {string} branch
     * @memberof GitCommitRequest
     * @instance
     */
    GitCommitRequest.prototype.branch = "";

    /**
     * GitCommitRequest commit.
     * @member {string} commit
     * @memberof GitCommitRequest
     * @instance
     */
    GitCommitRequest.prototype.commit = "";

    /**
     * Encodes the specified GitCommitRequest message. Does not implicitly {@link GitCommitRequest.verify|verify} messages.
     * @function encode
     * @memberof GitCommitRequest
     * @static
     * @param {GitCommitRequest} message GitCommitRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    GitCommitRequest.encode = function encode(message, writer) {
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
     * Decodes a GitCommitRequest message from the specified reader or buffer.
     * @function decode
     * @memberof GitCommitRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {GitCommitRequest} GitCommitRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    GitCommitRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.GitCommitRequest();
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

    return GitCommitRequest;
})();

export const GitCommitResponse = $root.GitCommitResponse = (() => {

    /**
     * Properties of a GitCommitResponse.
     * @exports IGitCommitResponse
     * @interface IGitCommitResponse
     * @property {GitOption|null} [data] GitCommitResponse data
     */

    /**
     * Constructs a new GitCommitResponse.
     * @exports GitCommitResponse
     * @classdesc Represents a GitCommitResponse.
     * @implements IGitCommitResponse
     * @constructor
     * @param {IGitCommitResponse=} [properties] Properties to set
     */
    function GitCommitResponse(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * GitCommitResponse data.
     * @member {GitOption|null|undefined} data
     * @memberof GitCommitResponse
     * @instance
     */
    GitCommitResponse.prototype.data = null;

    /**
     * Encodes the specified GitCommitResponse message. Does not implicitly {@link GitCommitResponse.verify|verify} messages.
     * @function encode
     * @memberof GitCommitResponse
     * @static
     * @param {GitCommitResponse} message GitCommitResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    GitCommitResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.data != null && Object.hasOwnProperty.call(message, "data"))
            $root.GitOption.encode(message.data, writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
        return writer;
    };

    /**
     * Decodes a GitCommitResponse message from the specified reader or buffer.
     * @function decode
     * @memberof GitCommitResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {GitCommitResponse} GitCommitResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    GitCommitResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.GitCommitResponse();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.data = $root.GitOption.decode(reader, reader.uint32());
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return GitCommitResponse;
})();

export const GitPipelineInfoRequest = $root.GitPipelineInfoRequest = (() => {

    /**
     * Properties of a GitPipelineInfoRequest.
     * @exports IGitPipelineInfoRequest
     * @interface IGitPipelineInfoRequest
     * @property {string|null} [project_id] GitPipelineInfoRequest project_id
     * @property {string|null} [branch] GitPipelineInfoRequest branch
     * @property {string|null} [commit] GitPipelineInfoRequest commit
     */

    /**
     * Constructs a new GitPipelineInfoRequest.
     * @exports GitPipelineInfoRequest
     * @classdesc Represents a GitPipelineInfoRequest.
     * @implements IGitPipelineInfoRequest
     * @constructor
     * @param {IGitPipelineInfoRequest=} [properties] Properties to set
     */
    function GitPipelineInfoRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * GitPipelineInfoRequest project_id.
     * @member {string} project_id
     * @memberof GitPipelineInfoRequest
     * @instance
     */
    GitPipelineInfoRequest.prototype.project_id = "";

    /**
     * GitPipelineInfoRequest branch.
     * @member {string} branch
     * @memberof GitPipelineInfoRequest
     * @instance
     */
    GitPipelineInfoRequest.prototype.branch = "";

    /**
     * GitPipelineInfoRequest commit.
     * @member {string} commit
     * @memberof GitPipelineInfoRequest
     * @instance
     */
    GitPipelineInfoRequest.prototype.commit = "";

    /**
     * Encodes the specified GitPipelineInfoRequest message. Does not implicitly {@link GitPipelineInfoRequest.verify|verify} messages.
     * @function encode
     * @memberof GitPipelineInfoRequest
     * @static
     * @param {GitPipelineInfoRequest} message GitPipelineInfoRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    GitPipelineInfoRequest.encode = function encode(message, writer) {
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
     * Decodes a GitPipelineInfoRequest message from the specified reader or buffer.
     * @function decode
     * @memberof GitPipelineInfoRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {GitPipelineInfoRequest} GitPipelineInfoRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    GitPipelineInfoRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.GitPipelineInfoRequest();
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

    return GitPipelineInfoRequest;
})();

export const GitPipelineInfoResponse = $root.GitPipelineInfoResponse = (() => {

    /**
     * Properties of a GitPipelineInfoResponse.
     * @exports IGitPipelineInfoResponse
     * @interface IGitPipelineInfoResponse
     * @property {string|null} [status] GitPipelineInfoResponse status
     * @property {string|null} [web_url] GitPipelineInfoResponse web_url
     */

    /**
     * Constructs a new GitPipelineInfoResponse.
     * @exports GitPipelineInfoResponse
     * @classdesc Represents a GitPipelineInfoResponse.
     * @implements IGitPipelineInfoResponse
     * @constructor
     * @param {IGitPipelineInfoResponse=} [properties] Properties to set
     */
    function GitPipelineInfoResponse(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * GitPipelineInfoResponse status.
     * @member {string} status
     * @memberof GitPipelineInfoResponse
     * @instance
     */
    GitPipelineInfoResponse.prototype.status = "";

    /**
     * GitPipelineInfoResponse web_url.
     * @member {string} web_url
     * @memberof GitPipelineInfoResponse
     * @instance
     */
    GitPipelineInfoResponse.prototype.web_url = "";

    /**
     * Encodes the specified GitPipelineInfoResponse message. Does not implicitly {@link GitPipelineInfoResponse.verify|verify} messages.
     * @function encode
     * @memberof GitPipelineInfoResponse
     * @static
     * @param {GitPipelineInfoResponse} message GitPipelineInfoResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    GitPipelineInfoResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.status != null && Object.hasOwnProperty.call(message, "status"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.status);
        if (message.web_url != null && Object.hasOwnProperty.call(message, "web_url"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.web_url);
        return writer;
    };

    /**
     * Decodes a GitPipelineInfoResponse message from the specified reader or buffer.
     * @function decode
     * @memberof GitPipelineInfoResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {GitPipelineInfoResponse} GitPipelineInfoResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    GitPipelineInfoResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.GitPipelineInfoResponse();
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

    return GitPipelineInfoResponse;
})();

export const GitConfigFileRequest = $root.GitConfigFileRequest = (() => {

    /**
     * Properties of a GitConfigFileRequest.
     * @exports IGitConfigFileRequest
     * @interface IGitConfigFileRequest
     * @property {string|null} [project_id] GitConfigFileRequest project_id
     * @property {string|null} [branch] GitConfigFileRequest branch
     */

    /**
     * Constructs a new GitConfigFileRequest.
     * @exports GitConfigFileRequest
     * @classdesc Represents a GitConfigFileRequest.
     * @implements IGitConfigFileRequest
     * @constructor
     * @param {IGitConfigFileRequest=} [properties] Properties to set
     */
    function GitConfigFileRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * GitConfigFileRequest project_id.
     * @member {string} project_id
     * @memberof GitConfigFileRequest
     * @instance
     */
    GitConfigFileRequest.prototype.project_id = "";

    /**
     * GitConfigFileRequest branch.
     * @member {string} branch
     * @memberof GitConfigFileRequest
     * @instance
     */
    GitConfigFileRequest.prototype.branch = "";

    /**
     * Encodes the specified GitConfigFileRequest message. Does not implicitly {@link GitConfigFileRequest.verify|verify} messages.
     * @function encode
     * @memberof GitConfigFileRequest
     * @static
     * @param {GitConfigFileRequest} message GitConfigFileRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    GitConfigFileRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.project_id != null && Object.hasOwnProperty.call(message, "project_id"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.project_id);
        if (message.branch != null && Object.hasOwnProperty.call(message, "branch"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.branch);
        return writer;
    };

    /**
     * Decodes a GitConfigFileRequest message from the specified reader or buffer.
     * @function decode
     * @memberof GitConfigFileRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {GitConfigFileRequest} GitConfigFileRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    GitConfigFileRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.GitConfigFileRequest();
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

    return GitConfigFileRequest;
})();

export const GitConfigFileResponse = $root.GitConfigFileResponse = (() => {

    /**
     * Properties of a GitConfigFileResponse.
     * @exports IGitConfigFileResponse
     * @interface IGitConfigFileResponse
     * @property {string|null} [data] GitConfigFileResponse data
     * @property {string|null} [type] GitConfigFileResponse type
     * @property {Array.<Element>|null} [elements] GitConfigFileResponse elements
     */

    /**
     * Constructs a new GitConfigFileResponse.
     * @exports GitConfigFileResponse
     * @classdesc Represents a GitConfigFileResponse.
     * @implements IGitConfigFileResponse
     * @constructor
     * @param {IGitConfigFileResponse=} [properties] Properties to set
     */
    function GitConfigFileResponse(properties) {
        this.elements = [];
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * GitConfigFileResponse data.
     * @member {string} data
     * @memberof GitConfigFileResponse
     * @instance
     */
    GitConfigFileResponse.prototype.data = "";

    /**
     * GitConfigFileResponse type.
     * @member {string} type
     * @memberof GitConfigFileResponse
     * @instance
     */
    GitConfigFileResponse.prototype.type = "";

    /**
     * GitConfigFileResponse elements.
     * @member {Array.<Element>} elements
     * @memberof GitConfigFileResponse
     * @instance
     */
    GitConfigFileResponse.prototype.elements = $util.emptyArray;

    /**
     * Encodes the specified GitConfigFileResponse message. Does not implicitly {@link GitConfigFileResponse.verify|verify} messages.
     * @function encode
     * @memberof GitConfigFileResponse
     * @static
     * @param {GitConfigFileResponse} message GitConfigFileResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    GitConfigFileResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.data != null && Object.hasOwnProperty.call(message, "data"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.data);
        if (message.type != null && Object.hasOwnProperty.call(message, "type"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.type);
        if (message.elements != null && message.elements.length)
            for (let i = 0; i < message.elements.length; ++i)
                $root.Element.encode(message.elements[i], writer.uint32(/* id 3, wireType 2 =*/26).fork()).ldelim();
        return writer;
    };

    /**
     * Decodes a GitConfigFileResponse message from the specified reader or buffer.
     * @function decode
     * @memberof GitConfigFileResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {GitConfigFileResponse} GitConfigFileResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    GitConfigFileResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.GitConfigFileResponse();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.data = reader.string();
                break;
            case 2:
                message.type = reader.string();
                break;
            case 3:
                if (!(message.elements && message.elements.length))
                    message.elements = [];
                message.elements.push($root.Element.decode(reader, reader.uint32()));
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return GitConfigFileResponse;
})();

export const GitEnableProjectResponse = $root.GitEnableProjectResponse = (() => {

    /**
     * Properties of a GitEnableProjectResponse.
     * @exports IGitEnableProjectResponse
     * @interface IGitEnableProjectResponse
     */

    /**
     * Constructs a new GitEnableProjectResponse.
     * @exports GitEnableProjectResponse
     * @classdesc Represents a GitEnableProjectResponse.
     * @implements IGitEnableProjectResponse
     * @constructor
     * @param {IGitEnableProjectResponse=} [properties] Properties to set
     */
    function GitEnableProjectResponse(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * Encodes the specified GitEnableProjectResponse message. Does not implicitly {@link GitEnableProjectResponse.verify|verify} messages.
     * @function encode
     * @memberof GitEnableProjectResponse
     * @static
     * @param {GitEnableProjectResponse} message GitEnableProjectResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    GitEnableProjectResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        return writer;
    };

    /**
     * Decodes a GitEnableProjectResponse message from the specified reader or buffer.
     * @function decode
     * @memberof GitEnableProjectResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {GitEnableProjectResponse} GitEnableProjectResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    GitEnableProjectResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.GitEnableProjectResponse();
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

    return GitEnableProjectResponse;
})();

export const GitDisableProjectResponse = $root.GitDisableProjectResponse = (() => {

    /**
     * Properties of a GitDisableProjectResponse.
     * @exports IGitDisableProjectResponse
     * @interface IGitDisableProjectResponse
     */

    /**
     * Constructs a new GitDisableProjectResponse.
     * @exports GitDisableProjectResponse
     * @classdesc Represents a GitDisableProjectResponse.
     * @implements IGitDisableProjectResponse
     * @constructor
     * @param {IGitDisableProjectResponse=} [properties] Properties to set
     */
    function GitDisableProjectResponse(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * Encodes the specified GitDisableProjectResponse message. Does not implicitly {@link GitDisableProjectResponse.verify|verify} messages.
     * @function encode
     * @memberof GitDisableProjectResponse
     * @static
     * @param {GitDisableProjectResponse} message GitDisableProjectResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    GitDisableProjectResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        return writer;
    };

    /**
     * Decodes a GitDisableProjectResponse message from the specified reader or buffer.
     * @function decode
     * @memberof GitDisableProjectResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {GitDisableProjectResponse} GitDisableProjectResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    GitDisableProjectResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.GitDisableProjectResponse();
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

    return GitDisableProjectResponse;
})();

export const GitAllProjectsRequest = $root.GitAllProjectsRequest = (() => {

    /**
     * Properties of a GitAllProjectsRequest.
     * @exports IGitAllProjectsRequest
     * @interface IGitAllProjectsRequest
     */

    /**
     * Constructs a new GitAllProjectsRequest.
     * @exports GitAllProjectsRequest
     * @classdesc Represents a GitAllProjectsRequest.
     * @implements IGitAllProjectsRequest
     * @constructor
     * @param {IGitAllProjectsRequest=} [properties] Properties to set
     */
    function GitAllProjectsRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * Encodes the specified GitAllProjectsRequest message. Does not implicitly {@link GitAllProjectsRequest.verify|verify} messages.
     * @function encode
     * @memberof GitAllProjectsRequest
     * @static
     * @param {GitAllProjectsRequest} message GitAllProjectsRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    GitAllProjectsRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        return writer;
    };

    /**
     * Decodes a GitAllProjectsRequest message from the specified reader or buffer.
     * @function decode
     * @memberof GitAllProjectsRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {GitAllProjectsRequest} GitAllProjectsRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    GitAllProjectsRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.GitAllProjectsRequest();
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

    return GitAllProjectsRequest;
})();

export const GitProjectOptionsRequest = $root.GitProjectOptionsRequest = (() => {

    /**
     * Properties of a GitProjectOptionsRequest.
     * @exports IGitProjectOptionsRequest
     * @interface IGitProjectOptionsRequest
     */

    /**
     * Constructs a new GitProjectOptionsRequest.
     * @exports GitProjectOptionsRequest
     * @classdesc Represents a GitProjectOptionsRequest.
     * @implements IGitProjectOptionsRequest
     * @constructor
     * @param {IGitProjectOptionsRequest=} [properties] Properties to set
     */
    function GitProjectOptionsRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * Encodes the specified GitProjectOptionsRequest message. Does not implicitly {@link GitProjectOptionsRequest.verify|verify} messages.
     * @function encode
     * @memberof GitProjectOptionsRequest
     * @static
     * @param {GitProjectOptionsRequest} message GitProjectOptionsRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    GitProjectOptionsRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        return writer;
    };

    /**
     * Decodes a GitProjectOptionsRequest message from the specified reader or buffer.
     * @function decode
     * @memberof GitProjectOptionsRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {GitProjectOptionsRequest} GitProjectOptionsRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    GitProjectOptionsRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.GitProjectOptionsRequest();
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

    return GitProjectOptionsRequest;
})();

export const GitServer = $root.GitServer = (() => {

    /**
     * Constructs a new GitServer service.
     * @exports GitServer
     * @classdesc Represents a GitServer
     * @extends $protobuf.rpc.Service
     * @constructor
     * @param {$protobuf.RPCImpl} rpcImpl RPC implementation
     * @param {boolean} [requestDelimited=false] Whether requests are length-delimited
     * @param {boolean} [responseDelimited=false] Whether responses are length-delimited
     */
    function GitServer(rpcImpl, requestDelimited, responseDelimited) {
        $protobuf.rpc.Service.call(this, rpcImpl, requestDelimited, responseDelimited);
    }

    (GitServer.prototype = Object.create($protobuf.rpc.Service.prototype)).constructor = GitServer;

    /**
     * Callback as used by {@link GitServer#enableProject}.
     * @memberof GitServer
     * @typedef EnableProjectCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {GitEnableProjectResponse} [response] GitEnableProjectResponse
     */

    /**
     * Calls EnableProject.
     * @function enableProject
     * @memberof GitServer
     * @instance
     * @param {GitEnableProjectRequest} request GitEnableProjectRequest message or plain object
     * @param {GitServer.EnableProjectCallback} callback Node-style callback called with the error, if any, and GitEnableProjectResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(GitServer.prototype.enableProject = function enableProject(request, callback) {
        return this.rpcCall(enableProject, $root.GitEnableProjectRequest, $root.GitEnableProjectResponse, request, callback);
    }, "name", { value: "EnableProject" });

    /**
     * Calls EnableProject.
     * @function enableProject
     * @memberof GitServer
     * @instance
     * @param {GitEnableProjectRequest} request GitEnableProjectRequest message or plain object
     * @returns {Promise<GitEnableProjectResponse>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link GitServer#disableProject}.
     * @memberof GitServer
     * @typedef DisableProjectCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {GitDisableProjectResponse} [response] GitDisableProjectResponse
     */

    /**
     * Calls DisableProject.
     * @function disableProject
     * @memberof GitServer
     * @instance
     * @param {GitDisableProjectRequest} request GitDisableProjectRequest message or plain object
     * @param {GitServer.DisableProjectCallback} callback Node-style callback called with the error, if any, and GitDisableProjectResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(GitServer.prototype.disableProject = function disableProject(request, callback) {
        return this.rpcCall(disableProject, $root.GitDisableProjectRequest, $root.GitDisableProjectResponse, request, callback);
    }, "name", { value: "DisableProject" });

    /**
     * Calls DisableProject.
     * @function disableProject
     * @memberof GitServer
     * @instance
     * @param {GitDisableProjectRequest} request GitDisableProjectRequest message or plain object
     * @returns {Promise<GitDisableProjectResponse>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link GitServer#all}.
     * @memberof GitServer
     * @typedef AllCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {GitAllProjectsResponse} [response] GitAllProjectsResponse
     */

    /**
     * Calls All.
     * @function all
     * @memberof GitServer
     * @instance
     * @param {GitAllProjectsRequest} request GitAllProjectsRequest message or plain object
     * @param {GitServer.AllCallback} callback Node-style callback called with the error, if any, and GitAllProjectsResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(GitServer.prototype.all = function all(request, callback) {
        return this.rpcCall(all, $root.GitAllProjectsRequest, $root.GitAllProjectsResponse, request, callback);
    }, "name", { value: "All" });

    /**
     * Calls All.
     * @function all
     * @memberof GitServer
     * @instance
     * @param {GitAllProjectsRequest} request GitAllProjectsRequest message or plain object
     * @returns {Promise<GitAllProjectsResponse>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link GitServer#projectOptions}.
     * @memberof GitServer
     * @typedef ProjectOptionsCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {GitProjectOptionsResponse} [response] GitProjectOptionsResponse
     */

    /**
     * Calls ProjectOptions.
     * @function projectOptions
     * @memberof GitServer
     * @instance
     * @param {GitProjectOptionsRequest} request GitProjectOptionsRequest message or plain object
     * @param {GitServer.ProjectOptionsCallback} callback Node-style callback called with the error, if any, and GitProjectOptionsResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(GitServer.prototype.projectOptions = function projectOptions(request, callback) {
        return this.rpcCall(projectOptions, $root.GitProjectOptionsRequest, $root.GitProjectOptionsResponse, request, callback);
    }, "name", { value: "ProjectOptions" });

    /**
     * Calls ProjectOptions.
     * @function projectOptions
     * @memberof GitServer
     * @instance
     * @param {GitProjectOptionsRequest} request GitProjectOptionsRequest message or plain object
     * @returns {Promise<GitProjectOptionsResponse>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link GitServer#branchOptions}.
     * @memberof GitServer
     * @typedef BranchOptionsCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {GitBranchOptionsResponse} [response] GitBranchOptionsResponse
     */

    /**
     * Calls BranchOptions.
     * @function branchOptions
     * @memberof GitServer
     * @instance
     * @param {GitBranchOptionsRequest} request GitBranchOptionsRequest message or plain object
     * @param {GitServer.BranchOptionsCallback} callback Node-style callback called with the error, if any, and GitBranchOptionsResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(GitServer.prototype.branchOptions = function branchOptions(request, callback) {
        return this.rpcCall(branchOptions, $root.GitBranchOptionsRequest, $root.GitBranchOptionsResponse, request, callback);
    }, "name", { value: "BranchOptions" });

    /**
     * Calls BranchOptions.
     * @function branchOptions
     * @memberof GitServer
     * @instance
     * @param {GitBranchOptionsRequest} request GitBranchOptionsRequest message or plain object
     * @returns {Promise<GitBranchOptionsResponse>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link GitServer#commitOptions}.
     * @memberof GitServer
     * @typedef CommitOptionsCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {GitCommitOptionsResponse} [response] GitCommitOptionsResponse
     */

    /**
     * Calls CommitOptions.
     * @function commitOptions
     * @memberof GitServer
     * @instance
     * @param {GitCommitOptionsRequest} request GitCommitOptionsRequest message or plain object
     * @param {GitServer.CommitOptionsCallback} callback Node-style callback called with the error, if any, and GitCommitOptionsResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(GitServer.prototype.commitOptions = function commitOptions(request, callback) {
        return this.rpcCall(commitOptions, $root.GitCommitOptionsRequest, $root.GitCommitOptionsResponse, request, callback);
    }, "name", { value: "CommitOptions" });

    /**
     * Calls CommitOptions.
     * @function commitOptions
     * @memberof GitServer
     * @instance
     * @param {GitCommitOptionsRequest} request GitCommitOptionsRequest message or plain object
     * @returns {Promise<GitCommitOptionsResponse>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link GitServer#commit}.
     * @memberof GitServer
     * @typedef CommitCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {GitCommitResponse} [response] GitCommitResponse
     */

    /**
     * Calls Commit.
     * @function commit
     * @memberof GitServer
     * @instance
     * @param {GitCommitRequest} request GitCommitRequest message or plain object
     * @param {GitServer.CommitCallback} callback Node-style callback called with the error, if any, and GitCommitResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(GitServer.prototype.commit = function commit(request, callback) {
        return this.rpcCall(commit, $root.GitCommitRequest, $root.GitCommitResponse, request, callback);
    }, "name", { value: "Commit" });

    /**
     * Calls Commit.
     * @function commit
     * @memberof GitServer
     * @instance
     * @param {GitCommitRequest} request GitCommitRequest message or plain object
     * @returns {Promise<GitCommitResponse>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link GitServer#pipelineInfo}.
     * @memberof GitServer
     * @typedef PipelineInfoCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {GitPipelineInfoResponse} [response] GitPipelineInfoResponse
     */

    /**
     * Calls PipelineInfo.
     * @function pipelineInfo
     * @memberof GitServer
     * @instance
     * @param {GitPipelineInfoRequest} request GitPipelineInfoRequest message or plain object
     * @param {GitServer.PipelineInfoCallback} callback Node-style callback called with the error, if any, and GitPipelineInfoResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(GitServer.prototype.pipelineInfo = function pipelineInfo(request, callback) {
        return this.rpcCall(pipelineInfo, $root.GitPipelineInfoRequest, $root.GitPipelineInfoResponse, request, callback);
    }, "name", { value: "PipelineInfo" });

    /**
     * Calls PipelineInfo.
     * @function pipelineInfo
     * @memberof GitServer
     * @instance
     * @param {GitPipelineInfoRequest} request GitPipelineInfoRequest message or plain object
     * @returns {Promise<GitPipelineInfoResponse>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link GitServer#marsConfigFile}.
     * @memberof GitServer
     * @typedef MarsConfigFileCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {GitConfigFileResponse} [response] GitConfigFileResponse
     */

    /**
     * Calls MarsConfigFile.
     * @function marsConfigFile
     * @memberof GitServer
     * @instance
     * @param {GitConfigFileRequest} request GitConfigFileRequest message or plain object
     * @param {GitServer.MarsConfigFileCallback} callback Node-style callback called with the error, if any, and GitConfigFileResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(GitServer.prototype.marsConfigFile = function marsConfigFile(request, callback) {
        return this.rpcCall(marsConfigFile, $root.GitConfigFileRequest, $root.GitConfigFileResponse, request, callback);
    }, "name", { value: "MarsConfigFile" });

    /**
     * Calls MarsConfigFile.
     * @function marsConfigFile
     * @memberof GitServer
     * @instance
     * @param {GitConfigFileRequest} request GitConfigFileRequest message or plain object
     * @returns {Promise<GitConfigFileResponse>} Promise
     * @variation 2
     */

    return GitServer;
})();

export const MarsConfig = $root.MarsConfig = (() => {

    /**
     * Properties of a MarsConfig.
     * @exports IMarsConfig
     * @interface IMarsConfig
     * @property {string|null} [config_file] MarsConfig config_file
     * @property {string|null} [config_file_values] MarsConfig config_file_values
     * @property {string|null} [config_field] MarsConfig config_field
     * @property {boolean|null} [is_simple_env] MarsConfig is_simple_env
     * @property {string|null} [config_file_type] MarsConfig config_file_type
     * @property {string|null} [local_chart_path] MarsConfig local_chart_path
     * @property {Array.<string>|null} [branches] MarsConfig branches
     * @property {string|null} [values_yaml] MarsConfig values_yaml
     * @property {Array.<Element>|null} [elements] MarsConfig elements
     */

    /**
     * Constructs a new MarsConfig.
     * @exports MarsConfig
     * @classdesc Represents a MarsConfig.
     * @implements IMarsConfig
     * @constructor
     * @param {IMarsConfig=} [properties] Properties to set
     */
    function MarsConfig(properties) {
        this.branches = [];
        this.elements = [];
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * MarsConfig config_file.
     * @member {string} config_file
     * @memberof MarsConfig
     * @instance
     */
    MarsConfig.prototype.config_file = "";

    /**
     * MarsConfig config_file_values.
     * @member {string} config_file_values
     * @memberof MarsConfig
     * @instance
     */
    MarsConfig.prototype.config_file_values = "";

    /**
     * MarsConfig config_field.
     * @member {string} config_field
     * @memberof MarsConfig
     * @instance
     */
    MarsConfig.prototype.config_field = "";

    /**
     * MarsConfig is_simple_env.
     * @member {boolean} is_simple_env
     * @memberof MarsConfig
     * @instance
     */
    MarsConfig.prototype.is_simple_env = false;

    /**
     * MarsConfig config_file_type.
     * @member {string} config_file_type
     * @memberof MarsConfig
     * @instance
     */
    MarsConfig.prototype.config_file_type = "";

    /**
     * MarsConfig local_chart_path.
     * @member {string} local_chart_path
     * @memberof MarsConfig
     * @instance
     */
    MarsConfig.prototype.local_chart_path = "";

    /**
     * MarsConfig branches.
     * @member {Array.<string>} branches
     * @memberof MarsConfig
     * @instance
     */
    MarsConfig.prototype.branches = $util.emptyArray;

    /**
     * MarsConfig values_yaml.
     * @member {string} values_yaml
     * @memberof MarsConfig
     * @instance
     */
    MarsConfig.prototype.values_yaml = "";

    /**
     * MarsConfig elements.
     * @member {Array.<Element>} elements
     * @memberof MarsConfig
     * @instance
     */
    MarsConfig.prototype.elements = $util.emptyArray;

    /**
     * Encodes the specified MarsConfig message. Does not implicitly {@link MarsConfig.verify|verify} messages.
     * @function encode
     * @memberof MarsConfig
     * @static
     * @param {MarsConfig} message MarsConfig message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    MarsConfig.encode = function encode(message, writer) {
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
        if (message.elements != null && message.elements.length)
            for (let i = 0; i < message.elements.length; ++i)
                $root.Element.encode(message.elements[i], writer.uint32(/* id 9, wireType 2 =*/74).fork()).ldelim();
        return writer;
    };

    /**
     * Decodes a MarsConfig message from the specified reader or buffer.
     * @function decode
     * @memberof MarsConfig
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {MarsConfig} MarsConfig
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    MarsConfig.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.MarsConfig();
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
            case 9:
                if (!(message.elements && message.elements.length))
                    message.elements = [];
                message.elements.push($root.Element.decode(reader, reader.uint32()));
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return MarsConfig;
})();

/**
 * ElementType enum.
 * @exports ElementType
 * @enum {number}
 * @property {number} ElementTypeUnknown=0 ElementTypeUnknown value
 * @property {number} ElementTypeInput=1 ElementTypeInput value
 * @property {number} ElementTypeInputNumber=2 ElementTypeInputNumber value
 * @property {number} ElementTypeSelect=3 ElementTypeSelect value
 * @property {number} ElementTypeRadio=4 ElementTypeRadio value
 * @property {number} ElementTypeSwitch=5 ElementTypeSwitch value
 */
export const ElementType = $root.ElementType = (() => {
    const valuesById = {}, values = Object.create(valuesById);
    values[valuesById[0] = "ElementTypeUnknown"] = 0;
    values[valuesById[1] = "ElementTypeInput"] = 1;
    values[valuesById[2] = "ElementTypeInputNumber"] = 2;
    values[valuesById[3] = "ElementTypeSelect"] = 3;
    values[valuesById[4] = "ElementTypeRadio"] = 4;
    values[valuesById[5] = "ElementTypeSwitch"] = 5;
    return values;
})();

export const Element = $root.Element = (() => {

    /**
     * Properties of an Element.
     * @exports IElement
     * @interface IElement
     * @property {string|null} [path] Element path
     * @property {ElementType|null} [type] Element type
     * @property {string|null} ["default"] Element default
     * @property {string|null} [description] Element description
     * @property {Array.<string>|null} [select_values] Element select_values
     */

    /**
     * Constructs a new Element.
     * @exports Element
     * @classdesc Represents an Element.
     * @implements IElement
     * @constructor
     * @param {IElement=} [properties] Properties to set
     */
    function Element(properties) {
        this.select_values = [];
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * Element path.
     * @member {string} path
     * @memberof Element
     * @instance
     */
    Element.prototype.path = "";

    /**
     * Element type.
     * @member {ElementType} type
     * @memberof Element
     * @instance
     */
    Element.prototype.type = 0;

    /**
     * Element default.
     * @member {string} default
     * @memberof Element
     * @instance
     */
    Element.prototype["default"] = "";

    /**
     * Element description.
     * @member {string} description
     * @memberof Element
     * @instance
     */
    Element.prototype.description = "";

    /**
     * Element select_values.
     * @member {Array.<string>} select_values
     * @memberof Element
     * @instance
     */
    Element.prototype.select_values = $util.emptyArray;

    /**
     * Encodes the specified Element message. Does not implicitly {@link Element.verify|verify} messages.
     * @function encode
     * @memberof Element
     * @static
     * @param {Element} message Element message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    Element.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.path != null && Object.hasOwnProperty.call(message, "path"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.path);
        if (message.type != null && Object.hasOwnProperty.call(message, "type"))
            writer.uint32(/* id 2, wireType 0 =*/16).int32(message.type);
        if (message["default"] != null && Object.hasOwnProperty.call(message, "default"))
            writer.uint32(/* id 3, wireType 2 =*/26).string(message["default"]);
        if (message.description != null && Object.hasOwnProperty.call(message, "description"))
            writer.uint32(/* id 4, wireType 2 =*/34).string(message.description);
        if (message.select_values != null && message.select_values.length)
            for (let i = 0; i < message.select_values.length; ++i)
                writer.uint32(/* id 6, wireType 2 =*/50).string(message.select_values[i]);
        return writer;
    };

    /**
     * Decodes an Element message from the specified reader or buffer.
     * @function decode
     * @memberof Element
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {Element} Element
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    Element.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.Element();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.path = reader.string();
                break;
            case 2:
                message.type = reader.int32();
                break;
            case 3:
                message["default"] = reader.string();
                break;
            case 4:
                message.description = reader.string();
                break;
            case 6:
                if (!(message.select_values && message.select_values.length))
                    message.select_values = [];
                message.select_values.push(reader.string());
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return Element;
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
     * @property {MarsConfig|null} [config] MarsShowResponse config
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
     * @member {MarsConfig|null|undefined} config
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
            $root.MarsConfig.encode(message.config, writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim();
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
                message.config = $root.MarsConfig.decode(reader, reader.uint32());
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

export const MarsGlobalConfigRequest = $root.MarsGlobalConfigRequest = (() => {

    /**
     * Properties of a MarsGlobalConfigRequest.
     * @exports IMarsGlobalConfigRequest
     * @interface IMarsGlobalConfigRequest
     * @property {number|null} [project_id] MarsGlobalConfigRequest project_id
     */

    /**
     * Constructs a new MarsGlobalConfigRequest.
     * @exports MarsGlobalConfigRequest
     * @classdesc Represents a MarsGlobalConfigRequest.
     * @implements IMarsGlobalConfigRequest
     * @constructor
     * @param {IMarsGlobalConfigRequest=} [properties] Properties to set
     */
    function MarsGlobalConfigRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * MarsGlobalConfigRequest project_id.
     * @member {number} project_id
     * @memberof MarsGlobalConfigRequest
     * @instance
     */
    MarsGlobalConfigRequest.prototype.project_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * Encodes the specified MarsGlobalConfigRequest message. Does not implicitly {@link MarsGlobalConfigRequest.verify|verify} messages.
     * @function encode
     * @memberof MarsGlobalConfigRequest
     * @static
     * @param {MarsGlobalConfigRequest} message MarsGlobalConfigRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    MarsGlobalConfigRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.project_id != null && Object.hasOwnProperty.call(message, "project_id"))
            writer.uint32(/* id 1, wireType 0 =*/8).int64(message.project_id);
        return writer;
    };

    /**
     * Decodes a MarsGlobalConfigRequest message from the specified reader or buffer.
     * @function decode
     * @memberof MarsGlobalConfigRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {MarsGlobalConfigRequest} MarsGlobalConfigRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    MarsGlobalConfigRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.MarsGlobalConfigRequest();
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

    return MarsGlobalConfigRequest;
})();

export const MarsGlobalConfigResponse = $root.MarsGlobalConfigResponse = (() => {

    /**
     * Properties of a MarsGlobalConfigResponse.
     * @exports IMarsGlobalConfigResponse
     * @interface IMarsGlobalConfigResponse
     * @property {boolean|null} [enabled] MarsGlobalConfigResponse enabled
     * @property {MarsConfig|null} [config] MarsGlobalConfigResponse config
     */

    /**
     * Constructs a new MarsGlobalConfigResponse.
     * @exports MarsGlobalConfigResponse
     * @classdesc Represents a MarsGlobalConfigResponse.
     * @implements IMarsGlobalConfigResponse
     * @constructor
     * @param {IMarsGlobalConfigResponse=} [properties] Properties to set
     */
    function MarsGlobalConfigResponse(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * MarsGlobalConfigResponse enabled.
     * @member {boolean} enabled
     * @memberof MarsGlobalConfigResponse
     * @instance
     */
    MarsGlobalConfigResponse.prototype.enabled = false;

    /**
     * MarsGlobalConfigResponse config.
     * @member {MarsConfig|null|undefined} config
     * @memberof MarsGlobalConfigResponse
     * @instance
     */
    MarsGlobalConfigResponse.prototype.config = null;

    /**
     * Encodes the specified MarsGlobalConfigResponse message. Does not implicitly {@link MarsGlobalConfigResponse.verify|verify} messages.
     * @function encode
     * @memberof MarsGlobalConfigResponse
     * @static
     * @param {MarsGlobalConfigResponse} message MarsGlobalConfigResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    MarsGlobalConfigResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.enabled != null && Object.hasOwnProperty.call(message, "enabled"))
            writer.uint32(/* id 1, wireType 0 =*/8).bool(message.enabled);
        if (message.config != null && Object.hasOwnProperty.call(message, "config"))
            $root.MarsConfig.encode(message.config, writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim();
        return writer;
    };

    /**
     * Decodes a MarsGlobalConfigResponse message from the specified reader or buffer.
     * @function decode
     * @memberof MarsGlobalConfigResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {MarsGlobalConfigResponse} MarsGlobalConfigResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    MarsGlobalConfigResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.MarsGlobalConfigResponse();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.enabled = reader.bool();
                break;
            case 2:
                message.config = $root.MarsConfig.decode(reader, reader.uint32());
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return MarsGlobalConfigResponse;
})();

export const MarsUpdateRequest = $root.MarsUpdateRequest = (() => {

    /**
     * Properties of a MarsUpdateRequest.
     * @exports IMarsUpdateRequest
     * @interface IMarsUpdateRequest
     * @property {number|null} [project_id] MarsUpdateRequest project_id
     * @property {MarsConfig|null} [config] MarsUpdateRequest config
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
     * @member {MarsConfig|null|undefined} config
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
            $root.MarsConfig.encode(message.config, writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim();
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
                message.config = $root.MarsConfig.decode(reader, reader.uint32());
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
     * @property {MarsConfig|null} [config] MarsUpdateResponse config
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
     * @member {MarsConfig|null|undefined} config
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
            $root.MarsConfig.encode(message.config, writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
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
                message.config = $root.MarsConfig.decode(reader, reader.uint32());
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

export const MarsToggleEnabledRequest = $root.MarsToggleEnabledRequest = (() => {

    /**
     * Properties of a MarsToggleEnabledRequest.
     * @exports IMarsToggleEnabledRequest
     * @interface IMarsToggleEnabledRequest
     * @property {number|null} [project_id] MarsToggleEnabledRequest project_id
     * @property {boolean|null} [enabled] MarsToggleEnabledRequest enabled
     */

    /**
     * Constructs a new MarsToggleEnabledRequest.
     * @exports MarsToggleEnabledRequest
     * @classdesc Represents a MarsToggleEnabledRequest.
     * @implements IMarsToggleEnabledRequest
     * @constructor
     * @param {IMarsToggleEnabledRequest=} [properties] Properties to set
     */
    function MarsToggleEnabledRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * MarsToggleEnabledRequest project_id.
     * @member {number} project_id
     * @memberof MarsToggleEnabledRequest
     * @instance
     */
    MarsToggleEnabledRequest.prototype.project_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * MarsToggleEnabledRequest enabled.
     * @member {boolean} enabled
     * @memberof MarsToggleEnabledRequest
     * @instance
     */
    MarsToggleEnabledRequest.prototype.enabled = false;

    /**
     * Encodes the specified MarsToggleEnabledRequest message. Does not implicitly {@link MarsToggleEnabledRequest.verify|verify} messages.
     * @function encode
     * @memberof MarsToggleEnabledRequest
     * @static
     * @param {MarsToggleEnabledRequest} message MarsToggleEnabledRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    MarsToggleEnabledRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.project_id != null && Object.hasOwnProperty.call(message, "project_id"))
            writer.uint32(/* id 1, wireType 0 =*/8).int64(message.project_id);
        if (message.enabled != null && Object.hasOwnProperty.call(message, "enabled"))
            writer.uint32(/* id 2, wireType 0 =*/16).bool(message.enabled);
        return writer;
    };

    /**
     * Decodes a MarsToggleEnabledRequest message from the specified reader or buffer.
     * @function decode
     * @memberof MarsToggleEnabledRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {MarsToggleEnabledRequest} MarsToggleEnabledRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    MarsToggleEnabledRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.MarsToggleEnabledRequest();
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

    return MarsToggleEnabledRequest;
})();

export const MarsDefaultChartValuesRequest = $root.MarsDefaultChartValuesRequest = (() => {

    /**
     * Properties of a MarsDefaultChartValuesRequest.
     * @exports IMarsDefaultChartValuesRequest
     * @interface IMarsDefaultChartValuesRequest
     * @property {number|null} [project_id] MarsDefaultChartValuesRequest project_id
     * @property {string|null} [branch] MarsDefaultChartValuesRequest branch
     */

    /**
     * Constructs a new MarsDefaultChartValuesRequest.
     * @exports MarsDefaultChartValuesRequest
     * @classdesc Represents a MarsDefaultChartValuesRequest.
     * @implements IMarsDefaultChartValuesRequest
     * @constructor
     * @param {IMarsDefaultChartValuesRequest=} [properties] Properties to set
     */
    function MarsDefaultChartValuesRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * MarsDefaultChartValuesRequest project_id.
     * @member {number} project_id
     * @memberof MarsDefaultChartValuesRequest
     * @instance
     */
    MarsDefaultChartValuesRequest.prototype.project_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * MarsDefaultChartValuesRequest branch.
     * @member {string} branch
     * @memberof MarsDefaultChartValuesRequest
     * @instance
     */
    MarsDefaultChartValuesRequest.prototype.branch = "";

    /**
     * Encodes the specified MarsDefaultChartValuesRequest message. Does not implicitly {@link MarsDefaultChartValuesRequest.verify|verify} messages.
     * @function encode
     * @memberof MarsDefaultChartValuesRequest
     * @static
     * @param {MarsDefaultChartValuesRequest} message MarsDefaultChartValuesRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    MarsDefaultChartValuesRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.project_id != null && Object.hasOwnProperty.call(message, "project_id"))
            writer.uint32(/* id 1, wireType 0 =*/8).int64(message.project_id);
        if (message.branch != null && Object.hasOwnProperty.call(message, "branch"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.branch);
        return writer;
    };

    /**
     * Decodes a MarsDefaultChartValuesRequest message from the specified reader or buffer.
     * @function decode
     * @memberof MarsDefaultChartValuesRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {MarsDefaultChartValuesRequest} MarsDefaultChartValuesRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    MarsDefaultChartValuesRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.MarsDefaultChartValuesRequest();
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

    return MarsDefaultChartValuesRequest;
})();

export const MarsDefaultChartValuesResponse = $root.MarsDefaultChartValuesResponse = (() => {

    /**
     * Properties of a MarsDefaultChartValuesResponse.
     * @exports IMarsDefaultChartValuesResponse
     * @interface IMarsDefaultChartValuesResponse
     * @property {string|null} [value] MarsDefaultChartValuesResponse value
     */

    /**
     * Constructs a new MarsDefaultChartValuesResponse.
     * @exports MarsDefaultChartValuesResponse
     * @classdesc Represents a MarsDefaultChartValuesResponse.
     * @implements IMarsDefaultChartValuesResponse
     * @constructor
     * @param {IMarsDefaultChartValuesResponse=} [properties] Properties to set
     */
    function MarsDefaultChartValuesResponse(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * MarsDefaultChartValuesResponse value.
     * @member {string} value
     * @memberof MarsDefaultChartValuesResponse
     * @instance
     */
    MarsDefaultChartValuesResponse.prototype.value = "";

    /**
     * Encodes the specified MarsDefaultChartValuesResponse message. Does not implicitly {@link MarsDefaultChartValuesResponse.verify|verify} messages.
     * @function encode
     * @memberof MarsDefaultChartValuesResponse
     * @static
     * @param {MarsDefaultChartValuesResponse} message MarsDefaultChartValuesResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    MarsDefaultChartValuesResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.value != null && Object.hasOwnProperty.call(message, "value"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.value);
        return writer;
    };

    /**
     * Decodes a MarsDefaultChartValuesResponse message from the specified reader or buffer.
     * @function decode
     * @memberof MarsDefaultChartValuesResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {MarsDefaultChartValuesResponse} MarsDefaultChartValuesResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    MarsDefaultChartValuesResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.MarsDefaultChartValuesResponse();
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

    return MarsDefaultChartValuesResponse;
})();

export const MarsToggleEnabledResponse = $root.MarsToggleEnabledResponse = (() => {

    /**
     * Properties of a MarsToggleEnabledResponse.
     * @exports IMarsToggleEnabledResponse
     * @interface IMarsToggleEnabledResponse
     */

    /**
     * Constructs a new MarsToggleEnabledResponse.
     * @exports MarsToggleEnabledResponse
     * @classdesc Represents a MarsToggleEnabledResponse.
     * @implements IMarsToggleEnabledResponse
     * @constructor
     * @param {IMarsToggleEnabledResponse=} [properties] Properties to set
     */
    function MarsToggleEnabledResponse(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * Encodes the specified MarsToggleEnabledResponse message. Does not implicitly {@link MarsToggleEnabledResponse.verify|verify} messages.
     * @function encode
     * @memberof MarsToggleEnabledResponse
     * @static
     * @param {MarsToggleEnabledResponse} message MarsToggleEnabledResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    MarsToggleEnabledResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        return writer;
    };

    /**
     * Decodes a MarsToggleEnabledResponse message from the specified reader or buffer.
     * @function decode
     * @memberof MarsToggleEnabledResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {MarsToggleEnabledResponse} MarsToggleEnabledResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    MarsToggleEnabledResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.MarsToggleEnabledResponse();
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

    return MarsToggleEnabledResponse;
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
     * @param {MarsGlobalConfigResponse} [response] MarsGlobalConfigResponse
     */

    /**
     * Calls GlobalConfig.
     * @function globalConfig
     * @memberof Mars
     * @instance
     * @param {MarsGlobalConfigRequest} request MarsGlobalConfigRequest message or plain object
     * @param {Mars.GlobalConfigCallback} callback Node-style callback called with the error, if any, and MarsGlobalConfigResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Mars.prototype.globalConfig = function globalConfig(request, callback) {
        return this.rpcCall(globalConfig, $root.MarsGlobalConfigRequest, $root.MarsGlobalConfigResponse, request, callback);
    }, "name", { value: "GlobalConfig" });

    /**
     * Calls GlobalConfig.
     * @function globalConfig
     * @memberof Mars
     * @instance
     * @param {MarsGlobalConfigRequest} request MarsGlobalConfigRequest message or plain object
     * @returns {Promise<MarsGlobalConfigResponse>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link Mars#toggleEnabled}.
     * @memberof Mars
     * @typedef ToggleEnabledCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {MarsToggleEnabledResponse} [response] MarsToggleEnabledResponse
     */

    /**
     * Calls ToggleEnabled.
     * @function toggleEnabled
     * @memberof Mars
     * @instance
     * @param {MarsToggleEnabledRequest} request MarsToggleEnabledRequest message or plain object
     * @param {Mars.ToggleEnabledCallback} callback Node-style callback called with the error, if any, and MarsToggleEnabledResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Mars.prototype.toggleEnabled = function toggleEnabled(request, callback) {
        return this.rpcCall(toggleEnabled, $root.MarsToggleEnabledRequest, $root.MarsToggleEnabledResponse, request, callback);
    }, "name", { value: "ToggleEnabled" });

    /**
     * Calls ToggleEnabled.
     * @function toggleEnabled
     * @memberof Mars
     * @instance
     * @param {MarsToggleEnabledRequest} request MarsToggleEnabledRequest message or plain object
     * @returns {Promise<MarsToggleEnabledResponse>} Promise
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
     * @param {MarsDefaultChartValuesResponse} [response] MarsDefaultChartValuesResponse
     */

    /**
     * Calls GetDefaultChartValues.
     * @function getDefaultChartValues
     * @memberof Mars
     * @instance
     * @param {MarsDefaultChartValuesRequest} request MarsDefaultChartValuesRequest message or plain object
     * @param {Mars.GetDefaultChartValuesCallback} callback Node-style callback called with the error, if any, and MarsDefaultChartValuesResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Mars.prototype.getDefaultChartValues = function getDefaultChartValues(request, callback) {
        return this.rpcCall(getDefaultChartValues, $root.MarsDefaultChartValuesRequest, $root.MarsDefaultChartValuesResponse, request, callback);
    }, "name", { value: "GetDefaultChartValues" });

    /**
     * Calls GetDefaultChartValues.
     * @function getDefaultChartValues
     * @memberof Mars
     * @instance
     * @param {MarsDefaultChartValuesRequest} request MarsDefaultChartValuesRequest message or plain object
     * @returns {Promise<MarsDefaultChartValuesResponse>} Promise
     * @variation 2
     */

    return Mars;
})();

export const MetricsShowRequest = $root.MetricsShowRequest = (() => {

    /**
     * Properties of a MetricsShowRequest.
     * @exports IMetricsShowRequest
     * @interface IMetricsShowRequest
     * @property {string|null} [namespace] MetricsShowRequest namespace
     * @property {string|null} [pod] MetricsShowRequest pod
     */

    /**
     * Constructs a new MetricsShowRequest.
     * @exports MetricsShowRequest
     * @classdesc Represents a MetricsShowRequest.
     * @implements IMetricsShowRequest
     * @constructor
     * @param {IMetricsShowRequest=} [properties] Properties to set
     */
    function MetricsShowRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * MetricsShowRequest namespace.
     * @member {string} namespace
     * @memberof MetricsShowRequest
     * @instance
     */
    MetricsShowRequest.prototype.namespace = "";

    /**
     * MetricsShowRequest pod.
     * @member {string} pod
     * @memberof MetricsShowRequest
     * @instance
     */
    MetricsShowRequest.prototype.pod = "";

    /**
     * Encodes the specified MetricsShowRequest message. Does not implicitly {@link MetricsShowRequest.verify|verify} messages.
     * @function encode
     * @memberof MetricsShowRequest
     * @static
     * @param {MetricsShowRequest} message MetricsShowRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    MetricsShowRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.namespace != null && Object.hasOwnProperty.call(message, "namespace"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.namespace);
        if (message.pod != null && Object.hasOwnProperty.call(message, "pod"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.pod);
        return writer;
    };

    /**
     * Decodes a MetricsShowRequest message from the specified reader or buffer.
     * @function decode
     * @memberof MetricsShowRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {MetricsShowRequest} MetricsShowRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    MetricsShowRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.MetricsShowRequest();
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

    return MetricsShowRequest;
})();

export const MetricsShowResponse = $root.MetricsShowResponse = (() => {

    /**
     * Properties of a MetricsShowResponse.
     * @exports IMetricsShowResponse
     * @interface IMetricsShowResponse
     * @property {number|null} [cpu] MetricsShowResponse cpu
     * @property {number|null} [memory] MetricsShowResponse memory
     * @property {string|null} [humanize_cpu] MetricsShowResponse humanize_cpu
     * @property {string|null} [humanize_memory] MetricsShowResponse humanize_memory
     * @property {string|null} [time] MetricsShowResponse time
     * @property {number|null} [length] MetricsShowResponse length
     */

    /**
     * Constructs a new MetricsShowResponse.
     * @exports MetricsShowResponse
     * @classdesc Represents a MetricsShowResponse.
     * @implements IMetricsShowResponse
     * @constructor
     * @param {IMetricsShowResponse=} [properties] Properties to set
     */
    function MetricsShowResponse(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * MetricsShowResponse cpu.
     * @member {number} cpu
     * @memberof MetricsShowResponse
     * @instance
     */
    MetricsShowResponse.prototype.cpu = 0;

    /**
     * MetricsShowResponse memory.
     * @member {number} memory
     * @memberof MetricsShowResponse
     * @instance
     */
    MetricsShowResponse.prototype.memory = 0;

    /**
     * MetricsShowResponse humanize_cpu.
     * @member {string} humanize_cpu
     * @memberof MetricsShowResponse
     * @instance
     */
    MetricsShowResponse.prototype.humanize_cpu = "";

    /**
     * MetricsShowResponse humanize_memory.
     * @member {string} humanize_memory
     * @memberof MetricsShowResponse
     * @instance
     */
    MetricsShowResponse.prototype.humanize_memory = "";

    /**
     * MetricsShowResponse time.
     * @member {string} time
     * @memberof MetricsShowResponse
     * @instance
     */
    MetricsShowResponse.prototype.time = "";

    /**
     * MetricsShowResponse length.
     * @member {number} length
     * @memberof MetricsShowResponse
     * @instance
     */
    MetricsShowResponse.prototype.length = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * Encodes the specified MetricsShowResponse message. Does not implicitly {@link MetricsShowResponse.verify|verify} messages.
     * @function encode
     * @memberof MetricsShowResponse
     * @static
     * @param {MetricsShowResponse} message MetricsShowResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    MetricsShowResponse.encode = function encode(message, writer) {
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
     * Decodes a MetricsShowResponse message from the specified reader or buffer.
     * @function decode
     * @memberof MetricsShowResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {MetricsShowResponse} MetricsShowResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    MetricsShowResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.MetricsShowResponse();
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

    return MetricsShowResponse;
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
     * Callback as used by {@link Metrics#show}.
     * @memberof Metrics
     * @typedef ShowCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {MetricsShowResponse} [response] MetricsShowResponse
     */

    /**
     * Calls Show.
     * @function show
     * @memberof Metrics
     * @instance
     * @param {MetricsShowRequest} request MetricsShowRequest message or plain object
     * @param {Metrics.ShowCallback} callback Node-style callback called with the error, if any, and MetricsShowResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Metrics.prototype.show = function show(request, callback) {
        return this.rpcCall(show, $root.MetricsShowRequest, $root.MetricsShowResponse, request, callback);
    }, "name", { value: "Show" });

    /**
     * Calls Show.
     * @function show
     * @memberof Metrics
     * @instance
     * @param {MetricsShowRequest} request MetricsShowRequest message or plain object
     * @returns {Promise<MetricsShowResponse>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link Metrics#streamShow}.
     * @memberof Metrics
     * @typedef StreamShowCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {MetricsShowResponse} [response] MetricsShowResponse
     */

    /**
     * Calls StreamShow.
     * @function streamShow
     * @memberof Metrics
     * @instance
     * @param {MetricsShowRequest} request MetricsShowRequest message or plain object
     * @param {Metrics.StreamShowCallback} callback Node-style callback called with the error, if any, and MetricsShowResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Metrics.prototype.streamShow = function streamShow(request, callback) {
        return this.rpcCall(streamShow, $root.MetricsShowRequest, $root.MetricsShowResponse, request, callback);
    }, "name", { value: "StreamShow" });

    /**
     * Calls StreamShow.
     * @function streamShow
     * @memberof Metrics
     * @instance
     * @param {MetricsShowRequest} request MetricsShowRequest message or plain object
     * @returns {Promise<MetricsShowResponse>} Promise
     * @variation 2
     */

    return Metrics;
})();

export const GitlabProjectModel = $root.GitlabProjectModel = (() => {

    /**
     * Properties of a GitlabProjectModel.
     * @exports IGitlabProjectModel
     * @interface IGitlabProjectModel
     * @property {number|null} [id] GitlabProjectModel id
     * @property {string|null} [default_branch] GitlabProjectModel default_branch
     * @property {string|null} [name] GitlabProjectModel name
     * @property {number|null} [gitlab_project_id] GitlabProjectModel gitlab_project_id
     * @property {boolean|null} [enabled] GitlabProjectModel enabled
     * @property {boolean|null} [global_enabled] GitlabProjectModel global_enabled
     * @property {string|null} [global_config] GitlabProjectModel global_config
     * @property {string|null} [created_at] GitlabProjectModel created_at
     * @property {string|null} [updated_at] GitlabProjectModel updated_at
     */

    /**
     * Constructs a new GitlabProjectModel.
     * @exports GitlabProjectModel
     * @classdesc Represents a GitlabProjectModel.
     * @implements IGitlabProjectModel
     * @constructor
     * @param {IGitlabProjectModel=} [properties] Properties to set
     */
    function GitlabProjectModel(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * GitlabProjectModel id.
     * @member {number} id
     * @memberof GitlabProjectModel
     * @instance
     */
    GitlabProjectModel.prototype.id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * GitlabProjectModel default_branch.
     * @member {string} default_branch
     * @memberof GitlabProjectModel
     * @instance
     */
    GitlabProjectModel.prototype.default_branch = "";

    /**
     * GitlabProjectModel name.
     * @member {string} name
     * @memberof GitlabProjectModel
     * @instance
     */
    GitlabProjectModel.prototype.name = "";

    /**
     * GitlabProjectModel gitlab_project_id.
     * @member {number} gitlab_project_id
     * @memberof GitlabProjectModel
     * @instance
     */
    GitlabProjectModel.prototype.gitlab_project_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * GitlabProjectModel enabled.
     * @member {boolean} enabled
     * @memberof GitlabProjectModel
     * @instance
     */
    GitlabProjectModel.prototype.enabled = false;

    /**
     * GitlabProjectModel global_enabled.
     * @member {boolean} global_enabled
     * @memberof GitlabProjectModel
     * @instance
     */
    GitlabProjectModel.prototype.global_enabled = false;

    /**
     * GitlabProjectModel global_config.
     * @member {string} global_config
     * @memberof GitlabProjectModel
     * @instance
     */
    GitlabProjectModel.prototype.global_config = "";

    /**
     * GitlabProjectModel created_at.
     * @member {string} created_at
     * @memberof GitlabProjectModel
     * @instance
     */
    GitlabProjectModel.prototype.created_at = "";

    /**
     * GitlabProjectModel updated_at.
     * @member {string} updated_at
     * @memberof GitlabProjectModel
     * @instance
     */
    GitlabProjectModel.prototype.updated_at = "";

    /**
     * Encodes the specified GitlabProjectModel message. Does not implicitly {@link GitlabProjectModel.verify|verify} messages.
     * @function encode
     * @memberof GitlabProjectModel
     * @static
     * @param {GitlabProjectModel} message GitlabProjectModel message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    GitlabProjectModel.encode = function encode(message, writer) {
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
            writer.uint32(/* id 8, wireType 2 =*/66).string(message.created_at);
        if (message.updated_at != null && Object.hasOwnProperty.call(message, "updated_at"))
            writer.uint32(/* id 9, wireType 2 =*/74).string(message.updated_at);
        return writer;
    };

    /**
     * Decodes a GitlabProjectModel message from the specified reader or buffer.
     * @function decode
     * @memberof GitlabProjectModel
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {GitlabProjectModel} GitlabProjectModel
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    GitlabProjectModel.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.GitlabProjectModel();
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
                message.created_at = reader.string();
                break;
            case 9:
                message.updated_at = reader.string();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return GitlabProjectModel;
})();

export const NamespaceModel = $root.NamespaceModel = (() => {

    /**
     * Properties of a NamespaceModel.
     * @exports INamespaceModel
     * @interface INamespaceModel
     * @property {number|null} [id] NamespaceModel id
     * @property {string|null} [name] NamespaceModel name
     * @property {Array.<string>|null} [image_pull_secrets] NamespaceModel image_pull_secrets
     * @property {string|null} [created_at] NamespaceModel created_at
     * @property {string|null} [updated_at] NamespaceModel updated_at
     * @property {Array.<ProjectModel>|null} [projects] NamespaceModel projects
     */

    /**
     * Constructs a new NamespaceModel.
     * @exports NamespaceModel
     * @classdesc Represents a NamespaceModel.
     * @implements INamespaceModel
     * @constructor
     * @param {INamespaceModel=} [properties] Properties to set
     */
    function NamespaceModel(properties) {
        this.image_pull_secrets = [];
        this.projects = [];
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * NamespaceModel id.
     * @member {number} id
     * @memberof NamespaceModel
     * @instance
     */
    NamespaceModel.prototype.id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * NamespaceModel name.
     * @member {string} name
     * @memberof NamespaceModel
     * @instance
     */
    NamespaceModel.prototype.name = "";

    /**
     * NamespaceModel image_pull_secrets.
     * @member {Array.<string>} image_pull_secrets
     * @memberof NamespaceModel
     * @instance
     */
    NamespaceModel.prototype.image_pull_secrets = $util.emptyArray;

    /**
     * NamespaceModel created_at.
     * @member {string} created_at
     * @memberof NamespaceModel
     * @instance
     */
    NamespaceModel.prototype.created_at = "";

    /**
     * NamespaceModel updated_at.
     * @member {string} updated_at
     * @memberof NamespaceModel
     * @instance
     */
    NamespaceModel.prototype.updated_at = "";

    /**
     * NamespaceModel projects.
     * @member {Array.<ProjectModel>} projects
     * @memberof NamespaceModel
     * @instance
     */
    NamespaceModel.prototype.projects = $util.emptyArray;

    /**
     * Encodes the specified NamespaceModel message. Does not implicitly {@link NamespaceModel.verify|verify} messages.
     * @function encode
     * @memberof NamespaceModel
     * @static
     * @param {NamespaceModel} message NamespaceModel message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    NamespaceModel.encode = function encode(message, writer) {
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
            writer.uint32(/* id 4, wireType 2 =*/34).string(message.created_at);
        if (message.updated_at != null && Object.hasOwnProperty.call(message, "updated_at"))
            writer.uint32(/* id 5, wireType 2 =*/42).string(message.updated_at);
        if (message.projects != null && message.projects.length)
            for (let i = 0; i < message.projects.length; ++i)
                $root.ProjectModel.encode(message.projects[i], writer.uint32(/* id 7, wireType 2 =*/58).fork()).ldelim();
        return writer;
    };

    /**
     * Decodes a NamespaceModel message from the specified reader or buffer.
     * @function decode
     * @memberof NamespaceModel
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {NamespaceModel} NamespaceModel
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    NamespaceModel.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.NamespaceModel();
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
                message.created_at = reader.string();
                break;
            case 5:
                message.updated_at = reader.string();
                break;
            case 7:
                if (!(message.projects && message.projects.length))
                    message.projects = [];
                message.projects.push($root.ProjectModel.decode(reader, reader.uint32()));
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return NamespaceModel;
})();

export const ProjectModel = $root.ProjectModel = (() => {

    /**
     * Properties of a ProjectModel.
     * @exports IProjectModel
     * @interface IProjectModel
     * @property {number|null} [id] ProjectModel id
     * @property {string|null} [name] ProjectModel name
     * @property {number|null} [gitlab_project_id] ProjectModel gitlab_project_id
     * @property {string|null} [gitlab_branch] ProjectModel gitlab_branch
     * @property {string|null} [gitlab_commit] ProjectModel gitlab_commit
     * @property {string|null} [config] ProjectModel config
     * @property {string|null} [override_values] ProjectModel override_values
     * @property {string|null} [docker_image] ProjectModel docker_image
     * @property {string|null} [pod_selectors] ProjectModel pod_selectors
     * @property {number|null} [namespace_id] ProjectModel namespace_id
     * @property {boolean|null} [atomic] ProjectModel atomic
     * @property {string|null} [created_at] ProjectModel created_at
     * @property {string|null} [updated_at] ProjectModel updated_at
     * @property {string|null} [extra_values] ProjectModel extra_values
     * @property {NamespaceModel|null} [namespace] ProjectModel namespace
     */

    /**
     * Constructs a new ProjectModel.
     * @exports ProjectModel
     * @classdesc Represents a ProjectModel.
     * @implements IProjectModel
     * @constructor
     * @param {IProjectModel=} [properties] Properties to set
     */
    function ProjectModel(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * ProjectModel id.
     * @member {number} id
     * @memberof ProjectModel
     * @instance
     */
    ProjectModel.prototype.id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * ProjectModel name.
     * @member {string} name
     * @memberof ProjectModel
     * @instance
     */
    ProjectModel.prototype.name = "";

    /**
     * ProjectModel gitlab_project_id.
     * @member {number} gitlab_project_id
     * @memberof ProjectModel
     * @instance
     */
    ProjectModel.prototype.gitlab_project_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * ProjectModel gitlab_branch.
     * @member {string} gitlab_branch
     * @memberof ProjectModel
     * @instance
     */
    ProjectModel.prototype.gitlab_branch = "";

    /**
     * ProjectModel gitlab_commit.
     * @member {string} gitlab_commit
     * @memberof ProjectModel
     * @instance
     */
    ProjectModel.prototype.gitlab_commit = "";

    /**
     * ProjectModel config.
     * @member {string} config
     * @memberof ProjectModel
     * @instance
     */
    ProjectModel.prototype.config = "";

    /**
     * ProjectModel override_values.
     * @member {string} override_values
     * @memberof ProjectModel
     * @instance
     */
    ProjectModel.prototype.override_values = "";

    /**
     * ProjectModel docker_image.
     * @member {string} docker_image
     * @memberof ProjectModel
     * @instance
     */
    ProjectModel.prototype.docker_image = "";

    /**
     * ProjectModel pod_selectors.
     * @member {string} pod_selectors
     * @memberof ProjectModel
     * @instance
     */
    ProjectModel.prototype.pod_selectors = "";

    /**
     * ProjectModel namespace_id.
     * @member {number} namespace_id
     * @memberof ProjectModel
     * @instance
     */
    ProjectModel.prototype.namespace_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * ProjectModel atomic.
     * @member {boolean} atomic
     * @memberof ProjectModel
     * @instance
     */
    ProjectModel.prototype.atomic = false;

    /**
     * ProjectModel created_at.
     * @member {string} created_at
     * @memberof ProjectModel
     * @instance
     */
    ProjectModel.prototype.created_at = "";

    /**
     * ProjectModel updated_at.
     * @member {string} updated_at
     * @memberof ProjectModel
     * @instance
     */
    ProjectModel.prototype.updated_at = "";

    /**
     * ProjectModel extra_values.
     * @member {string} extra_values
     * @memberof ProjectModel
     * @instance
     */
    ProjectModel.prototype.extra_values = "";

    /**
     * ProjectModel namespace.
     * @member {NamespaceModel|null|undefined} namespace
     * @memberof ProjectModel
     * @instance
     */
    ProjectModel.prototype.namespace = null;

    /**
     * Encodes the specified ProjectModel message. Does not implicitly {@link ProjectModel.verify|verify} messages.
     * @function encode
     * @memberof ProjectModel
     * @static
     * @param {ProjectModel} message ProjectModel message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    ProjectModel.encode = function encode(message, writer) {
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
            writer.uint32(/* id 12, wireType 2 =*/98).string(message.created_at);
        if (message.updated_at != null && Object.hasOwnProperty.call(message, "updated_at"))
            writer.uint32(/* id 13, wireType 2 =*/106).string(message.updated_at);
        if (message.extra_values != null && Object.hasOwnProperty.call(message, "extra_values"))
            writer.uint32(/* id 14, wireType 2 =*/114).string(message.extra_values);
        if (message.namespace != null && Object.hasOwnProperty.call(message, "namespace"))
            $root.NamespaceModel.encode(message.namespace, writer.uint32(/* id 15, wireType 2 =*/122).fork()).ldelim();
        return writer;
    };

    /**
     * Decodes a ProjectModel message from the specified reader or buffer.
     * @function decode
     * @memberof ProjectModel
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {ProjectModel} ProjectModel
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    ProjectModel.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.ProjectModel();
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
                message.created_at = reader.string();
                break;
            case 13:
                message.updated_at = reader.string();
                break;
            case 14:
                message.extra_values = reader.string();
                break;
            case 15:
                message.namespace = $root.NamespaceModel.decode(reader, reader.uint32());
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return ProjectModel;
})();

export const FileModel = $root.FileModel = (() => {

    /**
     * Properties of a FileModel.
     * @exports IFileModel
     * @interface IFileModel
     * @property {number|null} [id] FileModel id
     * @property {string|null} [path] FileModel path
     * @property {number|null} [size] FileModel size
     * @property {string|null} [username] FileModel username
     * @property {string|null} [namespace] FileModel namespace
     * @property {string|null} [pod] FileModel pod
     * @property {string|null} [container] FileModel container
     * @property {string|null} [container_path] FileModel container_path
     * @property {string|null} [created_at] FileModel created_at
     * @property {string|null} [updated_at] FileModel updated_at
     * @property {string|null} [deleted_at] FileModel deleted_at
     * @property {boolean|null} [is_deleted] FileModel is_deleted
     */

    /**
     * Constructs a new FileModel.
     * @exports FileModel
     * @classdesc Represents a FileModel.
     * @implements IFileModel
     * @constructor
     * @param {IFileModel=} [properties] Properties to set
     */
    function FileModel(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * FileModel id.
     * @member {number} id
     * @memberof FileModel
     * @instance
     */
    FileModel.prototype.id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * FileModel path.
     * @member {string} path
     * @memberof FileModel
     * @instance
     */
    FileModel.prototype.path = "";

    /**
     * FileModel size.
     * @member {number} size
     * @memberof FileModel
     * @instance
     */
    FileModel.prototype.size = $util.Long ? $util.Long.fromBits(0,0,true) : 0;

    /**
     * FileModel username.
     * @member {string} username
     * @memberof FileModel
     * @instance
     */
    FileModel.prototype.username = "";

    /**
     * FileModel namespace.
     * @member {string} namespace
     * @memberof FileModel
     * @instance
     */
    FileModel.prototype.namespace = "";

    /**
     * FileModel pod.
     * @member {string} pod
     * @memberof FileModel
     * @instance
     */
    FileModel.prototype.pod = "";

    /**
     * FileModel container.
     * @member {string} container
     * @memberof FileModel
     * @instance
     */
    FileModel.prototype.container = "";

    /**
     * FileModel container_path.
     * @member {string} container_path
     * @memberof FileModel
     * @instance
     */
    FileModel.prototype.container_path = "";

    /**
     * FileModel created_at.
     * @member {string} created_at
     * @memberof FileModel
     * @instance
     */
    FileModel.prototype.created_at = "";

    /**
     * FileModel updated_at.
     * @member {string} updated_at
     * @memberof FileModel
     * @instance
     */
    FileModel.prototype.updated_at = "";

    /**
     * FileModel deleted_at.
     * @member {string} deleted_at
     * @memberof FileModel
     * @instance
     */
    FileModel.prototype.deleted_at = "";

    /**
     * FileModel is_deleted.
     * @member {boolean} is_deleted
     * @memberof FileModel
     * @instance
     */
    FileModel.prototype.is_deleted = false;

    /**
     * Encodes the specified FileModel message. Does not implicitly {@link FileModel.verify|verify} messages.
     * @function encode
     * @memberof FileModel
     * @static
     * @param {FileModel} message FileModel message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    FileModel.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.id != null && Object.hasOwnProperty.call(message, "id"))
            writer.uint32(/* id 1, wireType 0 =*/8).int64(message.id);
        if (message.path != null && Object.hasOwnProperty.call(message, "path"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.path);
        if (message.size != null && Object.hasOwnProperty.call(message, "size"))
            writer.uint32(/* id 3, wireType 0 =*/24).uint64(message.size);
        if (message.username != null && Object.hasOwnProperty.call(message, "username"))
            writer.uint32(/* id 4, wireType 2 =*/34).string(message.username);
        if (message.namespace != null && Object.hasOwnProperty.call(message, "namespace"))
            writer.uint32(/* id 5, wireType 2 =*/42).string(message.namespace);
        if (message.pod != null && Object.hasOwnProperty.call(message, "pod"))
            writer.uint32(/* id 6, wireType 2 =*/50).string(message.pod);
        if (message.container != null && Object.hasOwnProperty.call(message, "container"))
            writer.uint32(/* id 7, wireType 2 =*/58).string(message.container);
        if (message.container_path != null && Object.hasOwnProperty.call(message, "container_path"))
            writer.uint32(/* id 8, wireType 2 =*/66).string(message.container_path);
        if (message.created_at != null && Object.hasOwnProperty.call(message, "created_at"))
            writer.uint32(/* id 9, wireType 2 =*/74).string(message.created_at);
        if (message.updated_at != null && Object.hasOwnProperty.call(message, "updated_at"))
            writer.uint32(/* id 10, wireType 2 =*/82).string(message.updated_at);
        if (message.deleted_at != null && Object.hasOwnProperty.call(message, "deleted_at"))
            writer.uint32(/* id 11, wireType 2 =*/90).string(message.deleted_at);
        if (message.is_deleted != null && Object.hasOwnProperty.call(message, "is_deleted"))
            writer.uint32(/* id 12, wireType 0 =*/96).bool(message.is_deleted);
        return writer;
    };

    /**
     * Decodes a FileModel message from the specified reader or buffer.
     * @function decode
     * @memberof FileModel
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {FileModel} FileModel
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    FileModel.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.FileModel();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.id = reader.int64();
                break;
            case 2:
                message.path = reader.string();
                break;
            case 3:
                message.size = reader.uint64();
                break;
            case 4:
                message.username = reader.string();
                break;
            case 5:
                message.namespace = reader.string();
                break;
            case 6:
                message.pod = reader.string();
                break;
            case 7:
                message.container = reader.string();
                break;
            case 8:
                message.container_path = reader.string();
                break;
            case 9:
                message.created_at = reader.string();
                break;
            case 10:
                message.updated_at = reader.string();
                break;
            case 11:
                message.deleted_at = reader.string();
                break;
            case 12:
                message.is_deleted = reader.bool();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return FileModel;
})();

export const NamespaceCreateRequest = $root.NamespaceCreateRequest = (() => {

    /**
     * Properties of a NamespaceCreateRequest.
     * @exports INamespaceCreateRequest
     * @interface INamespaceCreateRequest
     * @property {string|null} [namespace] NamespaceCreateRequest namespace
     */

    /**
     * Constructs a new NamespaceCreateRequest.
     * @exports NamespaceCreateRequest
     * @classdesc Represents a NamespaceCreateRequest.
     * @implements INamespaceCreateRequest
     * @constructor
     * @param {INamespaceCreateRequest=} [properties] Properties to set
     */
    function NamespaceCreateRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * NamespaceCreateRequest namespace.
     * @member {string} namespace
     * @memberof NamespaceCreateRequest
     * @instance
     */
    NamespaceCreateRequest.prototype.namespace = "";

    /**
     * Encodes the specified NamespaceCreateRequest message. Does not implicitly {@link NamespaceCreateRequest.verify|verify} messages.
     * @function encode
     * @memberof NamespaceCreateRequest
     * @static
     * @param {NamespaceCreateRequest} message NamespaceCreateRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    NamespaceCreateRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.namespace != null && Object.hasOwnProperty.call(message, "namespace"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.namespace);
        return writer;
    };

    /**
     * Decodes a NamespaceCreateRequest message from the specified reader or buffer.
     * @function decode
     * @memberof NamespaceCreateRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {NamespaceCreateRequest} NamespaceCreateRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    NamespaceCreateRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.NamespaceCreateRequest();
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

    return NamespaceCreateRequest;
})();

export const NamespaceShowRequest = $root.NamespaceShowRequest = (() => {

    /**
     * Properties of a NamespaceShowRequest.
     * @exports INamespaceShowRequest
     * @interface INamespaceShowRequest
     * @property {number|null} [namespace_id] NamespaceShowRequest namespace_id
     */

    /**
     * Constructs a new NamespaceShowRequest.
     * @exports NamespaceShowRequest
     * @classdesc Represents a NamespaceShowRequest.
     * @implements INamespaceShowRequest
     * @constructor
     * @param {INamespaceShowRequest=} [properties] Properties to set
     */
    function NamespaceShowRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * NamespaceShowRequest namespace_id.
     * @member {number} namespace_id
     * @memberof NamespaceShowRequest
     * @instance
     */
    NamespaceShowRequest.prototype.namespace_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * Encodes the specified NamespaceShowRequest message. Does not implicitly {@link NamespaceShowRequest.verify|verify} messages.
     * @function encode
     * @memberof NamespaceShowRequest
     * @static
     * @param {NamespaceShowRequest} message NamespaceShowRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    NamespaceShowRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.namespace_id != null && Object.hasOwnProperty.call(message, "namespace_id"))
            writer.uint32(/* id 1, wireType 0 =*/8).int64(message.namespace_id);
        return writer;
    };

    /**
     * Decodes a NamespaceShowRequest message from the specified reader or buffer.
     * @function decode
     * @memberof NamespaceShowRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {NamespaceShowRequest} NamespaceShowRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    NamespaceShowRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.NamespaceShowRequest();
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

    return NamespaceShowRequest;
})();

export const NamespaceDeleteRequest = $root.NamespaceDeleteRequest = (() => {

    /**
     * Properties of a NamespaceDeleteRequest.
     * @exports INamespaceDeleteRequest
     * @interface INamespaceDeleteRequest
     * @property {number|null} [namespace_id] NamespaceDeleteRequest namespace_id
     */

    /**
     * Constructs a new NamespaceDeleteRequest.
     * @exports NamespaceDeleteRequest
     * @classdesc Represents a NamespaceDeleteRequest.
     * @implements INamespaceDeleteRequest
     * @constructor
     * @param {INamespaceDeleteRequest=} [properties] Properties to set
     */
    function NamespaceDeleteRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * NamespaceDeleteRequest namespace_id.
     * @member {number} namespace_id
     * @memberof NamespaceDeleteRequest
     * @instance
     */
    NamespaceDeleteRequest.prototype.namespace_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * Encodes the specified NamespaceDeleteRequest message. Does not implicitly {@link NamespaceDeleteRequest.verify|verify} messages.
     * @function encode
     * @memberof NamespaceDeleteRequest
     * @static
     * @param {NamespaceDeleteRequest} message NamespaceDeleteRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    NamespaceDeleteRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.namespace_id != null && Object.hasOwnProperty.call(message, "namespace_id"))
            writer.uint32(/* id 1, wireType 0 =*/8).int64(message.namespace_id);
        return writer;
    };

    /**
     * Decodes a NamespaceDeleteRequest message from the specified reader or buffer.
     * @function decode
     * @memberof NamespaceDeleteRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {NamespaceDeleteRequest} NamespaceDeleteRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    NamespaceDeleteRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.NamespaceDeleteRequest();
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

    return NamespaceDeleteRequest;
})();

export const NamespaceIsExistsRequest = $root.NamespaceIsExistsRequest = (() => {

    /**
     * Properties of a NamespaceIsExistsRequest.
     * @exports INamespaceIsExistsRequest
     * @interface INamespaceIsExistsRequest
     * @property {string|null} [name] NamespaceIsExistsRequest name
     */

    /**
     * Constructs a new NamespaceIsExistsRequest.
     * @exports NamespaceIsExistsRequest
     * @classdesc Represents a NamespaceIsExistsRequest.
     * @implements INamespaceIsExistsRequest
     * @constructor
     * @param {INamespaceIsExistsRequest=} [properties] Properties to set
     */
    function NamespaceIsExistsRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * NamespaceIsExistsRequest name.
     * @member {string} name
     * @memberof NamespaceIsExistsRequest
     * @instance
     */
    NamespaceIsExistsRequest.prototype.name = "";

    /**
     * Encodes the specified NamespaceIsExistsRequest message. Does not implicitly {@link NamespaceIsExistsRequest.verify|verify} messages.
     * @function encode
     * @memberof NamespaceIsExistsRequest
     * @static
     * @param {NamespaceIsExistsRequest} message NamespaceIsExistsRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    NamespaceIsExistsRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.name != null && Object.hasOwnProperty.call(message, "name"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.name);
        return writer;
    };

    /**
     * Decodes a NamespaceIsExistsRequest message from the specified reader or buffer.
     * @function decode
     * @memberof NamespaceIsExistsRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {NamespaceIsExistsRequest} NamespaceIsExistsRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    NamespaceIsExistsRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.NamespaceIsExistsRequest();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.name = reader.string();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return NamespaceIsExistsRequest;
})();

export const NamespaceCpuMemoryRequest = $root.NamespaceCpuMemoryRequest = (() => {

    /**
     * Properties of a NamespaceCpuMemoryRequest.
     * @exports INamespaceCpuMemoryRequest
     * @interface INamespaceCpuMemoryRequest
     * @property {number|null} [namespace_id] NamespaceCpuMemoryRequest namespace_id
     */

    /**
     * Constructs a new NamespaceCpuMemoryRequest.
     * @exports NamespaceCpuMemoryRequest
     * @classdesc Represents a NamespaceCpuMemoryRequest.
     * @implements INamespaceCpuMemoryRequest
     * @constructor
     * @param {INamespaceCpuMemoryRequest=} [properties] Properties to set
     */
    function NamespaceCpuMemoryRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * NamespaceCpuMemoryRequest namespace_id.
     * @member {number} namespace_id
     * @memberof NamespaceCpuMemoryRequest
     * @instance
     */
    NamespaceCpuMemoryRequest.prototype.namespace_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * Encodes the specified NamespaceCpuMemoryRequest message. Does not implicitly {@link NamespaceCpuMemoryRequest.verify|verify} messages.
     * @function encode
     * @memberof NamespaceCpuMemoryRequest
     * @static
     * @param {NamespaceCpuMemoryRequest} message NamespaceCpuMemoryRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    NamespaceCpuMemoryRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.namespace_id != null && Object.hasOwnProperty.call(message, "namespace_id"))
            writer.uint32(/* id 1, wireType 0 =*/8).int64(message.namespace_id);
        return writer;
    };

    /**
     * Decodes a NamespaceCpuMemoryRequest message from the specified reader or buffer.
     * @function decode
     * @memberof NamespaceCpuMemoryRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {NamespaceCpuMemoryRequest} NamespaceCpuMemoryRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    NamespaceCpuMemoryRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.NamespaceCpuMemoryRequest();
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

    return NamespaceCpuMemoryRequest;
})();

export const NamespaceServiceEndpointsRequest = $root.NamespaceServiceEndpointsRequest = (() => {

    /**
     * Properties of a NamespaceServiceEndpointsRequest.
     * @exports INamespaceServiceEndpointsRequest
     * @interface INamespaceServiceEndpointsRequest
     * @property {number|null} [namespace_id] NamespaceServiceEndpointsRequest namespace_id
     * @property {string|null} [project_name] NamespaceServiceEndpointsRequest project_name
     */

    /**
     * Constructs a new NamespaceServiceEndpointsRequest.
     * @exports NamespaceServiceEndpointsRequest
     * @classdesc Represents a NamespaceServiceEndpointsRequest.
     * @implements INamespaceServiceEndpointsRequest
     * @constructor
     * @param {INamespaceServiceEndpointsRequest=} [properties] Properties to set
     */
    function NamespaceServiceEndpointsRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * NamespaceServiceEndpointsRequest namespace_id.
     * @member {number} namespace_id
     * @memberof NamespaceServiceEndpointsRequest
     * @instance
     */
    NamespaceServiceEndpointsRequest.prototype.namespace_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * NamespaceServiceEndpointsRequest project_name.
     * @member {string} project_name
     * @memberof NamespaceServiceEndpointsRequest
     * @instance
     */
    NamespaceServiceEndpointsRequest.prototype.project_name = "";

    /**
     * Encodes the specified NamespaceServiceEndpointsRequest message. Does not implicitly {@link NamespaceServiceEndpointsRequest.verify|verify} messages.
     * @function encode
     * @memberof NamespaceServiceEndpointsRequest
     * @static
     * @param {NamespaceServiceEndpointsRequest} message NamespaceServiceEndpointsRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    NamespaceServiceEndpointsRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.namespace_id != null && Object.hasOwnProperty.call(message, "namespace_id"))
            writer.uint32(/* id 1, wireType 0 =*/8).int64(message.namespace_id);
        if (message.project_name != null && Object.hasOwnProperty.call(message, "project_name"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.project_name);
        return writer;
    };

    /**
     * Decodes a NamespaceServiceEndpointsRequest message from the specified reader or buffer.
     * @function decode
     * @memberof NamespaceServiceEndpointsRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {NamespaceServiceEndpointsRequest} NamespaceServiceEndpointsRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    NamespaceServiceEndpointsRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.NamespaceServiceEndpointsRequest();
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

    return NamespaceServiceEndpointsRequest;
})();

export const NamespaceSimpleProject = $root.NamespaceSimpleProject = (() => {

    /**
     * Properties of a NamespaceSimpleProject.
     * @exports INamespaceSimpleProject
     * @interface INamespaceSimpleProject
     * @property {number|null} [id] NamespaceSimpleProject id
     * @property {string|null} [name] NamespaceSimpleProject name
     * @property {string|null} [status] NamespaceSimpleProject status
     */

    /**
     * Constructs a new NamespaceSimpleProject.
     * @exports NamespaceSimpleProject
     * @classdesc Represents a NamespaceSimpleProject.
     * @implements INamespaceSimpleProject
     * @constructor
     * @param {INamespaceSimpleProject=} [properties] Properties to set
     */
    function NamespaceSimpleProject(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * NamespaceSimpleProject id.
     * @member {number} id
     * @memberof NamespaceSimpleProject
     * @instance
     */
    NamespaceSimpleProject.prototype.id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * NamespaceSimpleProject name.
     * @member {string} name
     * @memberof NamespaceSimpleProject
     * @instance
     */
    NamespaceSimpleProject.prototype.name = "";

    /**
     * NamespaceSimpleProject status.
     * @member {string} status
     * @memberof NamespaceSimpleProject
     * @instance
     */
    NamespaceSimpleProject.prototype.status = "";

    /**
     * Encodes the specified NamespaceSimpleProject message. Does not implicitly {@link NamespaceSimpleProject.verify|verify} messages.
     * @function encode
     * @memberof NamespaceSimpleProject
     * @static
     * @param {NamespaceSimpleProject} message NamespaceSimpleProject message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    NamespaceSimpleProject.encode = function encode(message, writer) {
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
     * Decodes a NamespaceSimpleProject message from the specified reader or buffer.
     * @function decode
     * @memberof NamespaceSimpleProject
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {NamespaceSimpleProject} NamespaceSimpleProject
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    NamespaceSimpleProject.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.NamespaceSimpleProject();
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

    return NamespaceSimpleProject;
})();

export const NamespaceItem = $root.NamespaceItem = (() => {

    /**
     * Properties of a NamespaceItem.
     * @exports INamespaceItem
     * @interface INamespaceItem
     * @property {number|null} [id] NamespaceItem id
     * @property {string|null} [name] NamespaceItem name
     * @property {string|null} [created_at] NamespaceItem created_at
     * @property {string|null} [updated_at] NamespaceItem updated_at
     * @property {Array.<NamespaceSimpleProject>|null} [projects] NamespaceItem projects
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
     * @member {string} created_at
     * @memberof NamespaceItem
     * @instance
     */
    NamespaceItem.prototype.created_at = "";

    /**
     * NamespaceItem updated_at.
     * @member {string} updated_at
     * @memberof NamespaceItem
     * @instance
     */
    NamespaceItem.prototype.updated_at = "";

    /**
     * NamespaceItem projects.
     * @member {Array.<NamespaceSimpleProject>} projects
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
            writer.uint32(/* id 3, wireType 2 =*/26).string(message.created_at);
        if (message.updated_at != null && Object.hasOwnProperty.call(message, "updated_at"))
            writer.uint32(/* id 4, wireType 2 =*/34).string(message.updated_at);
        if (message.projects != null && message.projects.length)
            for (let i = 0; i < message.projects.length; ++i)
                $root.NamespaceSimpleProject.encode(message.projects[i], writer.uint32(/* id 5, wireType 2 =*/42).fork()).ldelim();
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
                message.created_at = reader.string();
                break;
            case 4:
                message.updated_at = reader.string();
                break;
            case 5:
                if (!(message.projects && message.projects.length))
                    message.projects = [];
                message.projects.push($root.NamespaceSimpleProject.decode(reader, reader.uint32()));
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return NamespaceItem;
})();

export const NamespaceAllResponse = $root.NamespaceAllResponse = (() => {

    /**
     * Properties of a NamespaceAllResponse.
     * @exports INamespaceAllResponse
     * @interface INamespaceAllResponse
     * @property {Array.<NamespaceItem>|null} [data] NamespaceAllResponse data
     */

    /**
     * Constructs a new NamespaceAllResponse.
     * @exports NamespaceAllResponse
     * @classdesc Represents a NamespaceAllResponse.
     * @implements INamespaceAllResponse
     * @constructor
     * @param {INamespaceAllResponse=} [properties] Properties to set
     */
    function NamespaceAllResponse(properties) {
        this.data = [];
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * NamespaceAllResponse data.
     * @member {Array.<NamespaceItem>} data
     * @memberof NamespaceAllResponse
     * @instance
     */
    NamespaceAllResponse.prototype.data = $util.emptyArray;

    /**
     * Encodes the specified NamespaceAllResponse message. Does not implicitly {@link NamespaceAllResponse.verify|verify} messages.
     * @function encode
     * @memberof NamespaceAllResponse
     * @static
     * @param {NamespaceAllResponse} message NamespaceAllResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    NamespaceAllResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.data != null && message.data.length)
            for (let i = 0; i < message.data.length; ++i)
                $root.NamespaceItem.encode(message.data[i], writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
        return writer;
    };

    /**
     * Decodes a NamespaceAllResponse message from the specified reader or buffer.
     * @function decode
     * @memberof NamespaceAllResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {NamespaceAllResponse} NamespaceAllResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    NamespaceAllResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.NamespaceAllResponse();
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

    return NamespaceAllResponse;
})();

export const NamespaceCreateResponse = $root.NamespaceCreateResponse = (() => {

    /**
     * Properties of a NamespaceCreateResponse.
     * @exports INamespaceCreateResponse
     * @interface INamespaceCreateResponse
     * @property {number|null} [id] NamespaceCreateResponse id
     * @property {string|null} [name] NamespaceCreateResponse name
     * @property {Array.<string>|null} [image_pull_secrets] NamespaceCreateResponse image_pull_secrets
     * @property {string|null} [created_at] NamespaceCreateResponse created_at
     * @property {string|null} [updated_at] NamespaceCreateResponse updated_at
     */

    /**
     * Constructs a new NamespaceCreateResponse.
     * @exports NamespaceCreateResponse
     * @classdesc Represents a NamespaceCreateResponse.
     * @implements INamespaceCreateResponse
     * @constructor
     * @param {INamespaceCreateResponse=} [properties] Properties to set
     */
    function NamespaceCreateResponse(properties) {
        this.image_pull_secrets = [];
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * NamespaceCreateResponse id.
     * @member {number} id
     * @memberof NamespaceCreateResponse
     * @instance
     */
    NamespaceCreateResponse.prototype.id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * NamespaceCreateResponse name.
     * @member {string} name
     * @memberof NamespaceCreateResponse
     * @instance
     */
    NamespaceCreateResponse.prototype.name = "";

    /**
     * NamespaceCreateResponse image_pull_secrets.
     * @member {Array.<string>} image_pull_secrets
     * @memberof NamespaceCreateResponse
     * @instance
     */
    NamespaceCreateResponse.prototype.image_pull_secrets = $util.emptyArray;

    /**
     * NamespaceCreateResponse created_at.
     * @member {string} created_at
     * @memberof NamespaceCreateResponse
     * @instance
     */
    NamespaceCreateResponse.prototype.created_at = "";

    /**
     * NamespaceCreateResponse updated_at.
     * @member {string} updated_at
     * @memberof NamespaceCreateResponse
     * @instance
     */
    NamespaceCreateResponse.prototype.updated_at = "";

    /**
     * Encodes the specified NamespaceCreateResponse message. Does not implicitly {@link NamespaceCreateResponse.verify|verify} messages.
     * @function encode
     * @memberof NamespaceCreateResponse
     * @static
     * @param {NamespaceCreateResponse} message NamespaceCreateResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    NamespaceCreateResponse.encode = function encode(message, writer) {
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
            writer.uint32(/* id 4, wireType 2 =*/34).string(message.created_at);
        if (message.updated_at != null && Object.hasOwnProperty.call(message, "updated_at"))
            writer.uint32(/* id 5, wireType 2 =*/42).string(message.updated_at);
        return writer;
    };

    /**
     * Decodes a NamespaceCreateResponse message from the specified reader or buffer.
     * @function decode
     * @memberof NamespaceCreateResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {NamespaceCreateResponse} NamespaceCreateResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    NamespaceCreateResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.NamespaceCreateResponse();
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
                message.created_at = reader.string();
                break;
            case 5:
                message.updated_at = reader.string();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return NamespaceCreateResponse;
})();

export const NamespaceShowResponse = $root.NamespaceShowResponse = (() => {

    /**
     * Properties of a NamespaceShowResponse.
     * @exports INamespaceShowResponse
     * @interface INamespaceShowResponse
     * @property {number|null} [id] NamespaceShowResponse id
     * @property {string|null} [name] NamespaceShowResponse name
     * @property {Array.<string>|null} [image_pull_secrets] NamespaceShowResponse image_pull_secrets
     * @property {string|null} [created_at] NamespaceShowResponse created_at
     * @property {string|null} [updated_at] NamespaceShowResponse updated_at
     * @property {Array.<ProjectModel>|null} [projects] NamespaceShowResponse projects
     */

    /**
     * Constructs a new NamespaceShowResponse.
     * @exports NamespaceShowResponse
     * @classdesc Represents a NamespaceShowResponse.
     * @implements INamespaceShowResponse
     * @constructor
     * @param {INamespaceShowResponse=} [properties] Properties to set
     */
    function NamespaceShowResponse(properties) {
        this.image_pull_secrets = [];
        this.projects = [];
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * NamespaceShowResponse id.
     * @member {number} id
     * @memberof NamespaceShowResponse
     * @instance
     */
    NamespaceShowResponse.prototype.id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * NamespaceShowResponse name.
     * @member {string} name
     * @memberof NamespaceShowResponse
     * @instance
     */
    NamespaceShowResponse.prototype.name = "";

    /**
     * NamespaceShowResponse image_pull_secrets.
     * @member {Array.<string>} image_pull_secrets
     * @memberof NamespaceShowResponse
     * @instance
     */
    NamespaceShowResponse.prototype.image_pull_secrets = $util.emptyArray;

    /**
     * NamespaceShowResponse created_at.
     * @member {string} created_at
     * @memberof NamespaceShowResponse
     * @instance
     */
    NamespaceShowResponse.prototype.created_at = "";

    /**
     * NamespaceShowResponse updated_at.
     * @member {string} updated_at
     * @memberof NamespaceShowResponse
     * @instance
     */
    NamespaceShowResponse.prototype.updated_at = "";

    /**
     * NamespaceShowResponse projects.
     * @member {Array.<ProjectModel>} projects
     * @memberof NamespaceShowResponse
     * @instance
     */
    NamespaceShowResponse.prototype.projects = $util.emptyArray;

    /**
     * Encodes the specified NamespaceShowResponse message. Does not implicitly {@link NamespaceShowResponse.verify|verify} messages.
     * @function encode
     * @memberof NamespaceShowResponse
     * @static
     * @param {NamespaceShowResponse} message NamespaceShowResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    NamespaceShowResponse.encode = function encode(message, writer) {
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
            writer.uint32(/* id 4, wireType 2 =*/34).string(message.created_at);
        if (message.updated_at != null && Object.hasOwnProperty.call(message, "updated_at"))
            writer.uint32(/* id 5, wireType 2 =*/42).string(message.updated_at);
        if (message.projects != null && message.projects.length)
            for (let i = 0; i < message.projects.length; ++i)
                $root.ProjectModel.encode(message.projects[i], writer.uint32(/* id 6, wireType 2 =*/50).fork()).ldelim();
        return writer;
    };

    /**
     * Decodes a NamespaceShowResponse message from the specified reader or buffer.
     * @function decode
     * @memberof NamespaceShowResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {NamespaceShowResponse} NamespaceShowResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    NamespaceShowResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.NamespaceShowResponse();
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
                message.created_at = reader.string();
                break;
            case 5:
                message.updated_at = reader.string();
                break;
            case 6:
                if (!(message.projects && message.projects.length))
                    message.projects = [];
                message.projects.push($root.ProjectModel.decode(reader, reader.uint32()));
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return NamespaceShowResponse;
})();

export const NamespaceCpuMemoryResponse = $root.NamespaceCpuMemoryResponse = (() => {

    /**
     * Properties of a NamespaceCpuMemoryResponse.
     * @exports INamespaceCpuMemoryResponse
     * @interface INamespaceCpuMemoryResponse
     * @property {string|null} [cpu] NamespaceCpuMemoryResponse cpu
     * @property {string|null} [memory] NamespaceCpuMemoryResponse memory
     */

    /**
     * Constructs a new NamespaceCpuMemoryResponse.
     * @exports NamespaceCpuMemoryResponse
     * @classdesc Represents a NamespaceCpuMemoryResponse.
     * @implements INamespaceCpuMemoryResponse
     * @constructor
     * @param {INamespaceCpuMemoryResponse=} [properties] Properties to set
     */
    function NamespaceCpuMemoryResponse(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * NamespaceCpuMemoryResponse cpu.
     * @member {string} cpu
     * @memberof NamespaceCpuMemoryResponse
     * @instance
     */
    NamespaceCpuMemoryResponse.prototype.cpu = "";

    /**
     * NamespaceCpuMemoryResponse memory.
     * @member {string} memory
     * @memberof NamespaceCpuMemoryResponse
     * @instance
     */
    NamespaceCpuMemoryResponse.prototype.memory = "";

    /**
     * Encodes the specified NamespaceCpuMemoryResponse message. Does not implicitly {@link NamespaceCpuMemoryResponse.verify|verify} messages.
     * @function encode
     * @memberof NamespaceCpuMemoryResponse
     * @static
     * @param {NamespaceCpuMemoryResponse} message NamespaceCpuMemoryResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    NamespaceCpuMemoryResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.cpu != null && Object.hasOwnProperty.call(message, "cpu"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.cpu);
        if (message.memory != null && Object.hasOwnProperty.call(message, "memory"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.memory);
        return writer;
    };

    /**
     * Decodes a NamespaceCpuMemoryResponse message from the specified reader or buffer.
     * @function decode
     * @memberof NamespaceCpuMemoryResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {NamespaceCpuMemoryResponse} NamespaceCpuMemoryResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    NamespaceCpuMemoryResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.NamespaceCpuMemoryResponse();
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

    return NamespaceCpuMemoryResponse;
})();

export const NamespaceServiceEndpoint = $root.NamespaceServiceEndpoint = (() => {

    /**
     * Properties of a NamespaceServiceEndpoint.
     * @exports INamespaceServiceEndpoint
     * @interface INamespaceServiceEndpoint
     * @property {string|null} [name] NamespaceServiceEndpoint name
     * @property {string|null} [url] NamespaceServiceEndpoint url
     * @property {string|null} [port_name] NamespaceServiceEndpoint port_name
     */

    /**
     * Constructs a new NamespaceServiceEndpoint.
     * @exports NamespaceServiceEndpoint
     * @classdesc Represents a NamespaceServiceEndpoint.
     * @implements INamespaceServiceEndpoint
     * @constructor
     * @param {INamespaceServiceEndpoint=} [properties] Properties to set
     */
    function NamespaceServiceEndpoint(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * NamespaceServiceEndpoint name.
     * @member {string} name
     * @memberof NamespaceServiceEndpoint
     * @instance
     */
    NamespaceServiceEndpoint.prototype.name = "";

    /**
     * NamespaceServiceEndpoint url.
     * @member {string} url
     * @memberof NamespaceServiceEndpoint
     * @instance
     */
    NamespaceServiceEndpoint.prototype.url = "";

    /**
     * NamespaceServiceEndpoint port_name.
     * @member {string} port_name
     * @memberof NamespaceServiceEndpoint
     * @instance
     */
    NamespaceServiceEndpoint.prototype.port_name = "";

    /**
     * Encodes the specified NamespaceServiceEndpoint message. Does not implicitly {@link NamespaceServiceEndpoint.verify|verify} messages.
     * @function encode
     * @memberof NamespaceServiceEndpoint
     * @static
     * @param {NamespaceServiceEndpoint} message NamespaceServiceEndpoint message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    NamespaceServiceEndpoint.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.name != null && Object.hasOwnProperty.call(message, "name"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.name);
        if (message.url != null && Object.hasOwnProperty.call(message, "url"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.url);
        if (message.port_name != null && Object.hasOwnProperty.call(message, "port_name"))
            writer.uint32(/* id 3, wireType 2 =*/26).string(message.port_name);
        return writer;
    };

    /**
     * Decodes a NamespaceServiceEndpoint message from the specified reader or buffer.
     * @function decode
     * @memberof NamespaceServiceEndpoint
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {NamespaceServiceEndpoint} NamespaceServiceEndpoint
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    NamespaceServiceEndpoint.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.NamespaceServiceEndpoint();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.name = reader.string();
                break;
            case 2:
                message.url = reader.string();
                break;
            case 3:
                message.port_name = reader.string();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return NamespaceServiceEndpoint;
})();

export const NamespaceServiceEndpointsResponse = $root.NamespaceServiceEndpointsResponse = (() => {

    /**
     * Properties of a NamespaceServiceEndpointsResponse.
     * @exports INamespaceServiceEndpointsResponse
     * @interface INamespaceServiceEndpointsResponse
     * @property {Array.<NamespaceServiceEndpoint>|null} [data] NamespaceServiceEndpointsResponse data
     */

    /**
     * Constructs a new NamespaceServiceEndpointsResponse.
     * @exports NamespaceServiceEndpointsResponse
     * @classdesc Represents a NamespaceServiceEndpointsResponse.
     * @implements INamespaceServiceEndpointsResponse
     * @constructor
     * @param {INamespaceServiceEndpointsResponse=} [properties] Properties to set
     */
    function NamespaceServiceEndpointsResponse(properties) {
        this.data = [];
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * NamespaceServiceEndpointsResponse data.
     * @member {Array.<NamespaceServiceEndpoint>} data
     * @memberof NamespaceServiceEndpointsResponse
     * @instance
     */
    NamespaceServiceEndpointsResponse.prototype.data = $util.emptyArray;

    /**
     * Encodes the specified NamespaceServiceEndpointsResponse message. Does not implicitly {@link NamespaceServiceEndpointsResponse.verify|verify} messages.
     * @function encode
     * @memberof NamespaceServiceEndpointsResponse
     * @static
     * @param {NamespaceServiceEndpointsResponse} message NamespaceServiceEndpointsResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    NamespaceServiceEndpointsResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.data != null && message.data.length)
            for (let i = 0; i < message.data.length; ++i)
                $root.NamespaceServiceEndpoint.encode(message.data[i], writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
        return writer;
    };

    /**
     * Decodes a NamespaceServiceEndpointsResponse message from the specified reader or buffer.
     * @function decode
     * @memberof NamespaceServiceEndpointsResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {NamespaceServiceEndpointsResponse} NamespaceServiceEndpointsResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    NamespaceServiceEndpointsResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.NamespaceServiceEndpointsResponse();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                if (!(message.data && message.data.length))
                    message.data = [];
                message.data.push($root.NamespaceServiceEndpoint.decode(reader, reader.uint32()));
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return NamespaceServiceEndpointsResponse;
})();

export const NamespaceIsExistsResponse = $root.NamespaceIsExistsResponse = (() => {

    /**
     * Properties of a NamespaceIsExistsResponse.
     * @exports INamespaceIsExistsResponse
     * @interface INamespaceIsExistsResponse
     * @property {boolean|null} [exists] NamespaceIsExistsResponse exists
     * @property {number|null} [id] NamespaceIsExistsResponse id
     */

    /**
     * Constructs a new NamespaceIsExistsResponse.
     * @exports NamespaceIsExistsResponse
     * @classdesc Represents a NamespaceIsExistsResponse.
     * @implements INamespaceIsExistsResponse
     * @constructor
     * @param {INamespaceIsExistsResponse=} [properties] Properties to set
     */
    function NamespaceIsExistsResponse(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * NamespaceIsExistsResponse exists.
     * @member {boolean} exists
     * @memberof NamespaceIsExistsResponse
     * @instance
     */
    NamespaceIsExistsResponse.prototype.exists = false;

    /**
     * NamespaceIsExistsResponse id.
     * @member {number} id
     * @memberof NamespaceIsExistsResponse
     * @instance
     */
    NamespaceIsExistsResponse.prototype.id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * Encodes the specified NamespaceIsExistsResponse message. Does not implicitly {@link NamespaceIsExistsResponse.verify|verify} messages.
     * @function encode
     * @memberof NamespaceIsExistsResponse
     * @static
     * @param {NamespaceIsExistsResponse} message NamespaceIsExistsResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    NamespaceIsExistsResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.exists != null && Object.hasOwnProperty.call(message, "exists"))
            writer.uint32(/* id 1, wireType 0 =*/8).bool(message.exists);
        if (message.id != null && Object.hasOwnProperty.call(message, "id"))
            writer.uint32(/* id 2, wireType 0 =*/16).int64(message.id);
        return writer;
    };

    /**
     * Decodes a NamespaceIsExistsResponse message from the specified reader or buffer.
     * @function decode
     * @memberof NamespaceIsExistsResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {NamespaceIsExistsResponse} NamespaceIsExistsResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    NamespaceIsExistsResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.NamespaceIsExistsResponse();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.exists = reader.bool();
                break;
            case 2:
                message.id = reader.int64();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return NamespaceIsExistsResponse;
})();

export const NamespaceAllRequest = $root.NamespaceAllRequest = (() => {

    /**
     * Properties of a NamespaceAllRequest.
     * @exports INamespaceAllRequest
     * @interface INamespaceAllRequest
     */

    /**
     * Constructs a new NamespaceAllRequest.
     * @exports NamespaceAllRequest
     * @classdesc Represents a NamespaceAllRequest.
     * @implements INamespaceAllRequest
     * @constructor
     * @param {INamespaceAllRequest=} [properties] Properties to set
     */
    function NamespaceAllRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * Encodes the specified NamespaceAllRequest message. Does not implicitly {@link NamespaceAllRequest.verify|verify} messages.
     * @function encode
     * @memberof NamespaceAllRequest
     * @static
     * @param {NamespaceAllRequest} message NamespaceAllRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    NamespaceAllRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        return writer;
    };

    /**
     * Decodes a NamespaceAllRequest message from the specified reader or buffer.
     * @function decode
     * @memberof NamespaceAllRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {NamespaceAllRequest} NamespaceAllRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    NamespaceAllRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.NamespaceAllRequest();
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

    return NamespaceAllRequest;
})();

export const NamespaceDeleteResponse = $root.NamespaceDeleteResponse = (() => {

    /**
     * Properties of a NamespaceDeleteResponse.
     * @exports INamespaceDeleteResponse
     * @interface INamespaceDeleteResponse
     */

    /**
     * Constructs a new NamespaceDeleteResponse.
     * @exports NamespaceDeleteResponse
     * @classdesc Represents a NamespaceDeleteResponse.
     * @implements INamespaceDeleteResponse
     * @constructor
     * @param {INamespaceDeleteResponse=} [properties] Properties to set
     */
    function NamespaceDeleteResponse(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * Encodes the specified NamespaceDeleteResponse message. Does not implicitly {@link NamespaceDeleteResponse.verify|verify} messages.
     * @function encode
     * @memberof NamespaceDeleteResponse
     * @static
     * @param {NamespaceDeleteResponse} message NamespaceDeleteResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    NamespaceDeleteResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        return writer;
    };

    /**
     * Decodes a NamespaceDeleteResponse message from the specified reader or buffer.
     * @function decode
     * @memberof NamespaceDeleteResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {NamespaceDeleteResponse} NamespaceDeleteResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    NamespaceDeleteResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.NamespaceDeleteResponse();
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

    return NamespaceDeleteResponse;
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
     * Callback as used by {@link Namespace#all}.
     * @memberof Namespace
     * @typedef AllCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {NamespaceAllResponse} [response] NamespaceAllResponse
     */

    /**
     * Calls All.
     * @function all
     * @memberof Namespace
     * @instance
     * @param {NamespaceAllRequest} request NamespaceAllRequest message or plain object
     * @param {Namespace.AllCallback} callback Node-style callback called with the error, if any, and NamespaceAllResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Namespace.prototype.all = function all(request, callback) {
        return this.rpcCall(all, $root.NamespaceAllRequest, $root.NamespaceAllResponse, request, callback);
    }, "name", { value: "All" });

    /**
     * Calls All.
     * @function all
     * @memberof Namespace
     * @instance
     * @param {NamespaceAllRequest} request NamespaceAllRequest message or plain object
     * @returns {Promise<NamespaceAllResponse>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link Namespace#create}.
     * @memberof Namespace
     * @typedef CreateCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {NamespaceCreateResponse} [response] NamespaceCreateResponse
     */

    /**
     * Calls Create.
     * @function create
     * @memberof Namespace
     * @instance
     * @param {NamespaceCreateRequest} request NamespaceCreateRequest message or plain object
     * @param {Namespace.CreateCallback} callback Node-style callback called with the error, if any, and NamespaceCreateResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Namespace.prototype.create = function create(request, callback) {
        return this.rpcCall(create, $root.NamespaceCreateRequest, $root.NamespaceCreateResponse, request, callback);
    }, "name", { value: "Create" });

    /**
     * Calls Create.
     * @function create
     * @memberof Namespace
     * @instance
     * @param {NamespaceCreateRequest} request NamespaceCreateRequest message or plain object
     * @returns {Promise<NamespaceCreateResponse>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link Namespace#show}.
     * @memberof Namespace
     * @typedef ShowCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {NamespaceShowResponse} [response] NamespaceShowResponse
     */

    /**
     * Calls Show.
     * @function show
     * @memberof Namespace
     * @instance
     * @param {NamespaceShowRequest} request NamespaceShowRequest message or plain object
     * @param {Namespace.ShowCallback} callback Node-style callback called with the error, if any, and NamespaceShowResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Namespace.prototype.show = function show(request, callback) {
        return this.rpcCall(show, $root.NamespaceShowRequest, $root.NamespaceShowResponse, request, callback);
    }, "name", { value: "Show" });

    /**
     * Calls Show.
     * @function show
     * @memberof Namespace
     * @instance
     * @param {NamespaceShowRequest} request NamespaceShowRequest message or plain object
     * @returns {Promise<NamespaceShowResponse>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link Namespace#delete_}.
     * @memberof Namespace
     * @typedef DeleteCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {NamespaceDeleteResponse} [response] NamespaceDeleteResponse
     */

    /**
     * Calls Delete.
     * @function delete
     * @memberof Namespace
     * @instance
     * @param {NamespaceDeleteRequest} request NamespaceDeleteRequest message or plain object
     * @param {Namespace.DeleteCallback} callback Node-style callback called with the error, if any, and NamespaceDeleteResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Namespace.prototype["delete"] = function delete_(request, callback) {
        return this.rpcCall(delete_, $root.NamespaceDeleteRequest, $root.NamespaceDeleteResponse, request, callback);
    }, "name", { value: "Delete" });

    /**
     * Calls Delete.
     * @function delete
     * @memberof Namespace
     * @instance
     * @param {NamespaceDeleteRequest} request NamespaceDeleteRequest message or plain object
     * @returns {Promise<NamespaceDeleteResponse>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link Namespace#isExists}.
     * @memberof Namespace
     * @typedef IsExistsCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {NamespaceIsExistsResponse} [response] NamespaceIsExistsResponse
     */

    /**
     * Calls IsExists.
     * @function isExists
     * @memberof Namespace
     * @instance
     * @param {NamespaceIsExistsRequest} request NamespaceIsExistsRequest message or plain object
     * @param {Namespace.IsExistsCallback} callback Node-style callback called with the error, if any, and NamespaceIsExistsResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Namespace.prototype.isExists = function isExists(request, callback) {
        return this.rpcCall(isExists, $root.NamespaceIsExistsRequest, $root.NamespaceIsExistsResponse, request, callback);
    }, "name", { value: "IsExists" });

    /**
     * Calls IsExists.
     * @function isExists
     * @memberof Namespace
     * @instance
     * @param {NamespaceIsExistsRequest} request NamespaceIsExistsRequest message or plain object
     * @returns {Promise<NamespaceIsExistsResponse>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link Namespace#cpuMemory}.
     * @memberof Namespace
     * @typedef CpuMemoryCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {NamespaceCpuMemoryResponse} [response] NamespaceCpuMemoryResponse
     */

    /**
     * Calls CpuMemory.
     * @function cpuMemory
     * @memberof Namespace
     * @instance
     * @param {NamespaceCpuMemoryRequest} request NamespaceCpuMemoryRequest message or plain object
     * @param {Namespace.CpuMemoryCallback} callback Node-style callback called with the error, if any, and NamespaceCpuMemoryResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Namespace.prototype.cpuMemory = function cpuMemory(request, callback) {
        return this.rpcCall(cpuMemory, $root.NamespaceCpuMemoryRequest, $root.NamespaceCpuMemoryResponse, request, callback);
    }, "name", { value: "CpuMemory" });

    /**
     * Calls CpuMemory.
     * @function cpuMemory
     * @memberof Namespace
     * @instance
     * @param {NamespaceCpuMemoryRequest} request NamespaceCpuMemoryRequest message or plain object
     * @returns {Promise<NamespaceCpuMemoryResponse>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link Namespace#serviceEndpoints}.
     * @memberof Namespace
     * @typedef ServiceEndpointsCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {NamespaceServiceEndpointsResponse} [response] NamespaceServiceEndpointsResponse
     */

    /**
     * Calls ServiceEndpoints.
     * @function serviceEndpoints
     * @memberof Namespace
     * @instance
     * @param {NamespaceServiceEndpointsRequest} request NamespaceServiceEndpointsRequest message or plain object
     * @param {Namespace.ServiceEndpointsCallback} callback Node-style callback called with the error, if any, and NamespaceServiceEndpointsResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Namespace.prototype.serviceEndpoints = function serviceEndpoints(request, callback) {
        return this.rpcCall(serviceEndpoints, $root.NamespaceServiceEndpointsRequest, $root.NamespaceServiceEndpointsResponse, request, callback);
    }, "name", { value: "ServiceEndpoints" });

    /**
     * Calls ServiceEndpoints.
     * @function serviceEndpoints
     * @memberof Namespace
     * @instance
     * @param {NamespaceServiceEndpointsRequest} request NamespaceServiceEndpointsRequest message or plain object
     * @returns {Promise<NamespaceServiceEndpointsResponse>} Promise
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

export const ProjectDeleteRequest = $root.ProjectDeleteRequest = (() => {

    /**
     * Properties of a ProjectDeleteRequest.
     * @exports IProjectDeleteRequest
     * @interface IProjectDeleteRequest
     * @property {number|null} [project_id] ProjectDeleteRequest project_id
     */

    /**
     * Constructs a new ProjectDeleteRequest.
     * @exports ProjectDeleteRequest
     * @classdesc Represents a ProjectDeleteRequest.
     * @implements IProjectDeleteRequest
     * @constructor
     * @param {IProjectDeleteRequest=} [properties] Properties to set
     */
    function ProjectDeleteRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * ProjectDeleteRequest project_id.
     * @member {number} project_id
     * @memberof ProjectDeleteRequest
     * @instance
     */
    ProjectDeleteRequest.prototype.project_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * Encodes the specified ProjectDeleteRequest message. Does not implicitly {@link ProjectDeleteRequest.verify|verify} messages.
     * @function encode
     * @memberof ProjectDeleteRequest
     * @static
     * @param {ProjectDeleteRequest} message ProjectDeleteRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    ProjectDeleteRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.project_id != null && Object.hasOwnProperty.call(message, "project_id"))
            writer.uint32(/* id 1, wireType 0 =*/8).int64(message.project_id);
        return writer;
    };

    /**
     * Decodes a ProjectDeleteRequest message from the specified reader or buffer.
     * @function decode
     * @memberof ProjectDeleteRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {ProjectDeleteRequest} ProjectDeleteRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    ProjectDeleteRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.ProjectDeleteRequest();
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

    return ProjectDeleteRequest;
})();

export const ProjectShowRequest = $root.ProjectShowRequest = (() => {

    /**
     * Properties of a ProjectShowRequest.
     * @exports IProjectShowRequest
     * @interface IProjectShowRequest
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
        if (message.project_id != null && Object.hasOwnProperty.call(message, "project_id"))
            writer.uint32(/* id 1, wireType 0 =*/8).int64(message.project_id);
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
     * @property {Array.<NamespaceServiceEndpoint>|null} [urls] ProjectShowResponse urls
     * @property {ProjectShowResponse.Namespace|null} [namespace] ProjectShowResponse namespace
     * @property {string|null} [cpu] ProjectShowResponse cpu
     * @property {string|null} [memory] ProjectShowResponse memory
     * @property {string|null} [override_values] ProjectShowResponse override_values
     * @property {string|null} [created_at] ProjectShowResponse created_at
     * @property {string|null} [updated_at] ProjectShowResponse updated_at
     * @property {string|null} [humanize_created_at] ProjectShowResponse humanize_created_at
     * @property {string|null} [humanize_updated_at] ProjectShowResponse humanize_updated_at
     * @property {Array.<ProjectExtraItem>|null} [extra_values] ProjectShowResponse extra_values
     * @property {Array.<Element>|null} [elements] ProjectShowResponse elements
     * @property {string|null} [config_type] ProjectShowResponse config_type
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
        this.extra_values = [];
        this.elements = [];
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
     * @member {Array.<NamespaceServiceEndpoint>} urls
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
     * ProjectShowResponse humanize_created_at.
     * @member {string} humanize_created_at
     * @memberof ProjectShowResponse
     * @instance
     */
    ProjectShowResponse.prototype.humanize_created_at = "";

    /**
     * ProjectShowResponse humanize_updated_at.
     * @member {string} humanize_updated_at
     * @memberof ProjectShowResponse
     * @instance
     */
    ProjectShowResponse.prototype.humanize_updated_at = "";

    /**
     * ProjectShowResponse extra_values.
     * @member {Array.<ProjectExtraItem>} extra_values
     * @memberof ProjectShowResponse
     * @instance
     */
    ProjectShowResponse.prototype.extra_values = $util.emptyArray;

    /**
     * ProjectShowResponse elements.
     * @member {Array.<Element>} elements
     * @memberof ProjectShowResponse
     * @instance
     */
    ProjectShowResponse.prototype.elements = $util.emptyArray;

    /**
     * ProjectShowResponse config_type.
     * @member {string} config_type
     * @memberof ProjectShowResponse
     * @instance
     */
    ProjectShowResponse.prototype.config_type = "";

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
                $root.NamespaceServiceEndpoint.encode(message.urls[i], writer.uint32(/* id 13, wireType 2 =*/106).fork()).ldelim();
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
        if (message.humanize_created_at != null && Object.hasOwnProperty.call(message, "humanize_created_at"))
            writer.uint32(/* id 20, wireType 2 =*/162).string(message.humanize_created_at);
        if (message.humanize_updated_at != null && Object.hasOwnProperty.call(message, "humanize_updated_at"))
            writer.uint32(/* id 21, wireType 2 =*/170).string(message.humanize_updated_at);
        if (message.extra_values != null && message.extra_values.length)
            for (let i = 0; i < message.extra_values.length; ++i)
                $root.ProjectExtraItem.encode(message.extra_values[i], writer.uint32(/* id 22, wireType 2 =*/178).fork()).ldelim();
        if (message.elements != null && message.elements.length)
            for (let i = 0; i < message.elements.length; ++i)
                $root.Element.encode(message.elements[i], writer.uint32(/* id 23, wireType 2 =*/186).fork()).ldelim();
        if (message.config_type != null && Object.hasOwnProperty.call(message, "config_type"))
            writer.uint32(/* id 24, wireType 2 =*/194).string(message.config_type);
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
                message.urls.push($root.NamespaceServiceEndpoint.decode(reader, reader.uint32()));
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
            case 20:
                message.humanize_created_at = reader.string();
                break;
            case 21:
                message.humanize_updated_at = reader.string();
                break;
            case 22:
                if (!(message.extra_values && message.extra_values.length))
                    message.extra_values = [];
                message.extra_values.push($root.ProjectExtraItem.decode(reader, reader.uint32()));
                break;
            case 23:
                if (!(message.elements && message.elements.length))
                    message.elements = [];
                message.elements.push($root.Element.decode(reader, reader.uint32()));
                break;
            case 24:
                message.config_type = reader.string();
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

export const ProjectAllPodContainersRequest = $root.ProjectAllPodContainersRequest = (() => {

    /**
     * Properties of a ProjectAllPodContainersRequest.
     * @exports IProjectAllPodContainersRequest
     * @interface IProjectAllPodContainersRequest
     * @property {number|null} [project_id] ProjectAllPodContainersRequest project_id
     */

    /**
     * Constructs a new ProjectAllPodContainersRequest.
     * @exports ProjectAllPodContainersRequest
     * @classdesc Represents a ProjectAllPodContainersRequest.
     * @implements IProjectAllPodContainersRequest
     * @constructor
     * @param {IProjectAllPodContainersRequest=} [properties] Properties to set
     */
    function ProjectAllPodContainersRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * ProjectAllPodContainersRequest project_id.
     * @member {number} project_id
     * @memberof ProjectAllPodContainersRequest
     * @instance
     */
    ProjectAllPodContainersRequest.prototype.project_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * Encodes the specified ProjectAllPodContainersRequest message. Does not implicitly {@link ProjectAllPodContainersRequest.verify|verify} messages.
     * @function encode
     * @memberof ProjectAllPodContainersRequest
     * @static
     * @param {ProjectAllPodContainersRequest} message ProjectAllPodContainersRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    ProjectAllPodContainersRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.project_id != null && Object.hasOwnProperty.call(message, "project_id"))
            writer.uint32(/* id 1, wireType 0 =*/8).int64(message.project_id);
        return writer;
    };

    /**
     * Decodes a ProjectAllPodContainersRequest message from the specified reader or buffer.
     * @function decode
     * @memberof ProjectAllPodContainersRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {ProjectAllPodContainersRequest} ProjectAllPodContainersRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    ProjectAllPodContainersRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.ProjectAllPodContainersRequest();
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

    return ProjectAllPodContainersRequest;
})();

export const ProjectPodLog = $root.ProjectPodLog = (() => {

    /**
     * Properties of a ProjectPodLog.
     * @exports IProjectPodLog
     * @interface IProjectPodLog
     * @property {string|null} [namespace] ProjectPodLog namespace
     * @property {string|null} [pod_name] ProjectPodLog pod_name
     * @property {string|null} [container_name] ProjectPodLog container_name
     * @property {string|null} [log] ProjectPodLog log
     */

    /**
     * Constructs a new ProjectPodLog.
     * @exports ProjectPodLog
     * @classdesc Represents a ProjectPodLog.
     * @implements IProjectPodLog
     * @constructor
     * @param {IProjectPodLog=} [properties] Properties to set
     */
    function ProjectPodLog(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * ProjectPodLog namespace.
     * @member {string} namespace
     * @memberof ProjectPodLog
     * @instance
     */
    ProjectPodLog.prototype.namespace = "";

    /**
     * ProjectPodLog pod_name.
     * @member {string} pod_name
     * @memberof ProjectPodLog
     * @instance
     */
    ProjectPodLog.prototype.pod_name = "";

    /**
     * ProjectPodLog container_name.
     * @member {string} container_name
     * @memberof ProjectPodLog
     * @instance
     */
    ProjectPodLog.prototype.container_name = "";

    /**
     * ProjectPodLog log.
     * @member {string} log
     * @memberof ProjectPodLog
     * @instance
     */
    ProjectPodLog.prototype.log = "";

    /**
     * Encodes the specified ProjectPodLog message. Does not implicitly {@link ProjectPodLog.verify|verify} messages.
     * @function encode
     * @memberof ProjectPodLog
     * @static
     * @param {ProjectPodLog} message ProjectPodLog message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    ProjectPodLog.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.namespace != null && Object.hasOwnProperty.call(message, "namespace"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.namespace);
        if (message.pod_name != null && Object.hasOwnProperty.call(message, "pod_name"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.pod_name);
        if (message.container_name != null && Object.hasOwnProperty.call(message, "container_name"))
            writer.uint32(/* id 3, wireType 2 =*/26).string(message.container_name);
        if (message.log != null && Object.hasOwnProperty.call(message, "log"))
            writer.uint32(/* id 4, wireType 2 =*/34).string(message.log);
        return writer;
    };

    /**
     * Decodes a ProjectPodLog message from the specified reader or buffer.
     * @function decode
     * @memberof ProjectPodLog
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {ProjectPodLog} ProjectPodLog
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    ProjectPodLog.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.ProjectPodLog();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.namespace = reader.string();
                break;
            case 2:
                message.pod_name = reader.string();
                break;
            case 3:
                message.container_name = reader.string();
                break;
            case 4:
                message.log = reader.string();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return ProjectPodLog;
})();

export const ProjectAllPodContainersResponse = $root.ProjectAllPodContainersResponse = (() => {

    /**
     * Properties of a ProjectAllPodContainersResponse.
     * @exports IProjectAllPodContainersResponse
     * @interface IProjectAllPodContainersResponse
     * @property {Array.<ProjectPodLog>|null} [data] ProjectAllPodContainersResponse data
     */

    /**
     * Constructs a new ProjectAllPodContainersResponse.
     * @exports ProjectAllPodContainersResponse
     * @classdesc Represents a ProjectAllPodContainersResponse.
     * @implements IProjectAllPodContainersResponse
     * @constructor
     * @param {IProjectAllPodContainersResponse=} [properties] Properties to set
     */
    function ProjectAllPodContainersResponse(properties) {
        this.data = [];
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * ProjectAllPodContainersResponse data.
     * @member {Array.<ProjectPodLog>} data
     * @memberof ProjectAllPodContainersResponse
     * @instance
     */
    ProjectAllPodContainersResponse.prototype.data = $util.emptyArray;

    /**
     * Encodes the specified ProjectAllPodContainersResponse message. Does not implicitly {@link ProjectAllPodContainersResponse.verify|verify} messages.
     * @function encode
     * @memberof ProjectAllPodContainersResponse
     * @static
     * @param {ProjectAllPodContainersResponse} message ProjectAllPodContainersResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    ProjectAllPodContainersResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.data != null && message.data.length)
            for (let i = 0; i < message.data.length; ++i)
                $root.ProjectPodLog.encode(message.data[i], writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
        return writer;
    };

    /**
     * Decodes a ProjectAllPodContainersResponse message from the specified reader or buffer.
     * @function decode
     * @memberof ProjectAllPodContainersResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {ProjectAllPodContainersResponse} ProjectAllPodContainersResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    ProjectAllPodContainersResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.ProjectAllPodContainersResponse();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                if (!(message.data && message.data.length))
                    message.data = [];
                message.data.push($root.ProjectPodLog.decode(reader, reader.uint32()));
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return ProjectAllPodContainersResponse;
})();

export const ProjectPodContainerLogRequest = $root.ProjectPodContainerLogRequest = (() => {

    /**
     * Properties of a ProjectPodContainerLogRequest.
     * @exports IProjectPodContainerLogRequest
     * @interface IProjectPodContainerLogRequest
     * @property {number|null} [project_id] ProjectPodContainerLogRequest project_id
     * @property {string|null} [pod] ProjectPodContainerLogRequest pod
     * @property {string|null} [container] ProjectPodContainerLogRequest container
     */

    /**
     * Constructs a new ProjectPodContainerLogRequest.
     * @exports ProjectPodContainerLogRequest
     * @classdesc Represents a ProjectPodContainerLogRequest.
     * @implements IProjectPodContainerLogRequest
     * @constructor
     * @param {IProjectPodContainerLogRequest=} [properties] Properties to set
     */
    function ProjectPodContainerLogRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * ProjectPodContainerLogRequest project_id.
     * @member {number} project_id
     * @memberof ProjectPodContainerLogRequest
     * @instance
     */
    ProjectPodContainerLogRequest.prototype.project_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * ProjectPodContainerLogRequest pod.
     * @member {string} pod
     * @memberof ProjectPodContainerLogRequest
     * @instance
     */
    ProjectPodContainerLogRequest.prototype.pod = "";

    /**
     * ProjectPodContainerLogRequest container.
     * @member {string} container
     * @memberof ProjectPodContainerLogRequest
     * @instance
     */
    ProjectPodContainerLogRequest.prototype.container = "";

    /**
     * Encodes the specified ProjectPodContainerLogRequest message. Does not implicitly {@link ProjectPodContainerLogRequest.verify|verify} messages.
     * @function encode
     * @memberof ProjectPodContainerLogRequest
     * @static
     * @param {ProjectPodContainerLogRequest} message ProjectPodContainerLogRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    ProjectPodContainerLogRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.project_id != null && Object.hasOwnProperty.call(message, "project_id"))
            writer.uint32(/* id 1, wireType 0 =*/8).int64(message.project_id);
        if (message.pod != null && Object.hasOwnProperty.call(message, "pod"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.pod);
        if (message.container != null && Object.hasOwnProperty.call(message, "container"))
            writer.uint32(/* id 3, wireType 2 =*/26).string(message.container);
        return writer;
    };

    /**
     * Decodes a ProjectPodContainerLogRequest message from the specified reader or buffer.
     * @function decode
     * @memberof ProjectPodContainerLogRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {ProjectPodContainerLogRequest} ProjectPodContainerLogRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    ProjectPodContainerLogRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.ProjectPodContainerLogRequest();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.project_id = reader.int64();
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

    return ProjectPodContainerLogRequest;
})();

export const ProjectPodContainerLogResponse = $root.ProjectPodContainerLogResponse = (() => {

    /**
     * Properties of a ProjectPodContainerLogResponse.
     * @exports IProjectPodContainerLogResponse
     * @interface IProjectPodContainerLogResponse
     * @property {ProjectPodLog|null} [data] ProjectPodContainerLogResponse data
     */

    /**
     * Constructs a new ProjectPodContainerLogResponse.
     * @exports ProjectPodContainerLogResponse
     * @classdesc Represents a ProjectPodContainerLogResponse.
     * @implements IProjectPodContainerLogResponse
     * @constructor
     * @param {IProjectPodContainerLogResponse=} [properties] Properties to set
     */
    function ProjectPodContainerLogResponse(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * ProjectPodContainerLogResponse data.
     * @member {ProjectPodLog|null|undefined} data
     * @memberof ProjectPodContainerLogResponse
     * @instance
     */
    ProjectPodContainerLogResponse.prototype.data = null;

    /**
     * Encodes the specified ProjectPodContainerLogResponse message. Does not implicitly {@link ProjectPodContainerLogResponse.verify|verify} messages.
     * @function encode
     * @memberof ProjectPodContainerLogResponse
     * @static
     * @param {ProjectPodContainerLogResponse} message ProjectPodContainerLogResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    ProjectPodContainerLogResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.data != null && Object.hasOwnProperty.call(message, "data"))
            $root.ProjectPodLog.encode(message.data, writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
        return writer;
    };

    /**
     * Decodes a ProjectPodContainerLogResponse message from the specified reader or buffer.
     * @function decode
     * @memberof ProjectPodContainerLogResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {ProjectPodContainerLogResponse} ProjectPodContainerLogResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    ProjectPodContainerLogResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.ProjectPodContainerLogResponse();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.data = $root.ProjectPodLog.decode(reader, reader.uint32());
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return ProjectPodContainerLogResponse;
})();

export const ProjectIsPodRunningRequest = $root.ProjectIsPodRunningRequest = (() => {

    /**
     * Properties of a ProjectIsPodRunningRequest.
     * @exports IProjectIsPodRunningRequest
     * @interface IProjectIsPodRunningRequest
     * @property {string|null} [namespace] ProjectIsPodRunningRequest namespace
     * @property {string|null} [pod] ProjectIsPodRunningRequest pod
     */

    /**
     * Constructs a new ProjectIsPodRunningRequest.
     * @exports ProjectIsPodRunningRequest
     * @classdesc Represents a ProjectIsPodRunningRequest.
     * @implements IProjectIsPodRunningRequest
     * @constructor
     * @param {IProjectIsPodRunningRequest=} [properties] Properties to set
     */
    function ProjectIsPodRunningRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * ProjectIsPodRunningRequest namespace.
     * @member {string} namespace
     * @memberof ProjectIsPodRunningRequest
     * @instance
     */
    ProjectIsPodRunningRequest.prototype.namespace = "";

    /**
     * ProjectIsPodRunningRequest pod.
     * @member {string} pod
     * @memberof ProjectIsPodRunningRequest
     * @instance
     */
    ProjectIsPodRunningRequest.prototype.pod = "";

    /**
     * Encodes the specified ProjectIsPodRunningRequest message. Does not implicitly {@link ProjectIsPodRunningRequest.verify|verify} messages.
     * @function encode
     * @memberof ProjectIsPodRunningRequest
     * @static
     * @param {ProjectIsPodRunningRequest} message ProjectIsPodRunningRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    ProjectIsPodRunningRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.namespace != null && Object.hasOwnProperty.call(message, "namespace"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.namespace);
        if (message.pod != null && Object.hasOwnProperty.call(message, "pod"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.pod);
        return writer;
    };

    /**
     * Decodes a ProjectIsPodRunningRequest message from the specified reader or buffer.
     * @function decode
     * @memberof ProjectIsPodRunningRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {ProjectIsPodRunningRequest} ProjectIsPodRunningRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    ProjectIsPodRunningRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.ProjectIsPodRunningRequest();
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

    return ProjectIsPodRunningRequest;
})();

export const ProjectIsPodRunningResponse = $root.ProjectIsPodRunningResponse = (() => {

    /**
     * Properties of a ProjectIsPodRunningResponse.
     * @exports IProjectIsPodRunningResponse
     * @interface IProjectIsPodRunningResponse
     * @property {boolean|null} [running] ProjectIsPodRunningResponse running
     * @property {string|null} [reason] ProjectIsPodRunningResponse reason
     */

    /**
     * Constructs a new ProjectIsPodRunningResponse.
     * @exports ProjectIsPodRunningResponse
     * @classdesc Represents a ProjectIsPodRunningResponse.
     * @implements IProjectIsPodRunningResponse
     * @constructor
     * @param {IProjectIsPodRunningResponse=} [properties] Properties to set
     */
    function ProjectIsPodRunningResponse(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * ProjectIsPodRunningResponse running.
     * @member {boolean} running
     * @memberof ProjectIsPodRunningResponse
     * @instance
     */
    ProjectIsPodRunningResponse.prototype.running = false;

    /**
     * ProjectIsPodRunningResponse reason.
     * @member {string} reason
     * @memberof ProjectIsPodRunningResponse
     * @instance
     */
    ProjectIsPodRunningResponse.prototype.reason = "";

    /**
     * Encodes the specified ProjectIsPodRunningResponse message. Does not implicitly {@link ProjectIsPodRunningResponse.verify|verify} messages.
     * @function encode
     * @memberof ProjectIsPodRunningResponse
     * @static
     * @param {ProjectIsPodRunningResponse} message ProjectIsPodRunningResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    ProjectIsPodRunningResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.running != null && Object.hasOwnProperty.call(message, "running"))
            writer.uint32(/* id 1, wireType 0 =*/8).bool(message.running);
        if (message.reason != null && Object.hasOwnProperty.call(message, "reason"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.reason);
        return writer;
    };

    /**
     * Decodes a ProjectIsPodRunningResponse message from the specified reader or buffer.
     * @function decode
     * @memberof ProjectIsPodRunningResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {ProjectIsPodRunningResponse} ProjectIsPodRunningResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    ProjectIsPodRunningResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.ProjectIsPodRunningResponse();
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

    return ProjectIsPodRunningResponse;
})();

export const ProjectApplyResponse = $root.ProjectApplyResponse = (() => {

    /**
     * Properties of a ProjectApplyResponse.
     * @exports IProjectApplyResponse
     * @interface IProjectApplyResponse
     * @property {Metadata|null} [metadata] ProjectApplyResponse metadata
     * @property {ProjectModel|null} [project] ProjectApplyResponse project
     */

    /**
     * Constructs a new ProjectApplyResponse.
     * @exports ProjectApplyResponse
     * @classdesc Represents a ProjectApplyResponse.
     * @implements IProjectApplyResponse
     * @constructor
     * @param {IProjectApplyResponse=} [properties] Properties to set
     */
    function ProjectApplyResponse(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * ProjectApplyResponse metadata.
     * @member {Metadata|null|undefined} metadata
     * @memberof ProjectApplyResponse
     * @instance
     */
    ProjectApplyResponse.prototype.metadata = null;

    /**
     * ProjectApplyResponse project.
     * @member {ProjectModel|null|undefined} project
     * @memberof ProjectApplyResponse
     * @instance
     */
    ProjectApplyResponse.prototype.project = null;

    /**
     * Encodes the specified ProjectApplyResponse message. Does not implicitly {@link ProjectApplyResponse.verify|verify} messages.
     * @function encode
     * @memberof ProjectApplyResponse
     * @static
     * @param {ProjectApplyResponse} message ProjectApplyResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    ProjectApplyResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.metadata != null && Object.hasOwnProperty.call(message, "metadata"))
            $root.Metadata.encode(message.metadata, writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
        if (message.project != null && Object.hasOwnProperty.call(message, "project"))
            $root.ProjectModel.encode(message.project, writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim();
        return writer;
    };

    /**
     * Decodes a ProjectApplyResponse message from the specified reader or buffer.
     * @function decode
     * @memberof ProjectApplyResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {ProjectApplyResponse} ProjectApplyResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    ProjectApplyResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.ProjectApplyResponse();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.metadata = $root.Metadata.decode(reader, reader.uint32());
                break;
            case 2:
                message.project = $root.ProjectModel.decode(reader, reader.uint32());
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return ProjectApplyResponse;
})();

export const ProjectDryRunApplyResponse = $root.ProjectDryRunApplyResponse = (() => {

    /**
     * Properties of a ProjectDryRunApplyResponse.
     * @exports IProjectDryRunApplyResponse
     * @interface IProjectDryRunApplyResponse
     * @property {Array.<string>|null} [results] ProjectDryRunApplyResponse results
     */

    /**
     * Constructs a new ProjectDryRunApplyResponse.
     * @exports ProjectDryRunApplyResponse
     * @classdesc Represents a ProjectDryRunApplyResponse.
     * @implements IProjectDryRunApplyResponse
     * @constructor
     * @param {IProjectDryRunApplyResponse=} [properties] Properties to set
     */
    function ProjectDryRunApplyResponse(properties) {
        this.results = [];
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * ProjectDryRunApplyResponse results.
     * @member {Array.<string>} results
     * @memberof ProjectDryRunApplyResponse
     * @instance
     */
    ProjectDryRunApplyResponse.prototype.results = $util.emptyArray;

    /**
     * Encodes the specified ProjectDryRunApplyResponse message. Does not implicitly {@link ProjectDryRunApplyResponse.verify|verify} messages.
     * @function encode
     * @memberof ProjectDryRunApplyResponse
     * @static
     * @param {ProjectDryRunApplyResponse} message ProjectDryRunApplyResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    ProjectDryRunApplyResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.results != null && message.results.length)
            for (let i = 0; i < message.results.length; ++i)
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.results[i]);
        return writer;
    };

    /**
     * Decodes a ProjectDryRunApplyResponse message from the specified reader or buffer.
     * @function decode
     * @memberof ProjectDryRunApplyResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {ProjectDryRunApplyResponse} ProjectDryRunApplyResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    ProjectDryRunApplyResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.ProjectDryRunApplyResponse();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                if (!(message.results && message.results.length))
                    message.results = [];
                message.results.push(reader.string());
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return ProjectDryRunApplyResponse;
})();

export const ProjectApplyRequest = $root.ProjectApplyRequest = (() => {

    /**
     * Properties of a ProjectApplyRequest.
     * @exports IProjectApplyRequest
     * @interface IProjectApplyRequest
     * @property {number|null} [namespace_id] ProjectApplyRequest namespace_id
     * @property {string|null} [name] ProjectApplyRequest name
     * @property {number|null} [gitlab_project_id] ProjectApplyRequest gitlab_project_id
     * @property {string|null} [gitlab_branch] ProjectApplyRequest gitlab_branch
     * @property {string|null} [gitlab_commit] ProjectApplyRequest gitlab_commit
     * @property {string|null} [config] ProjectApplyRequest config
     * @property {boolean|null} [atomic] ProjectApplyRequest atomic
     * @property {boolean|null} [websocket_sync] ProjectApplyRequest websocket_sync
     * @property {Array.<ProjectExtraItem>|null} [extra_values] ProjectApplyRequest extra_values
     * @property {number|null} [install_timeout_seconds] ProjectApplyRequest install_timeout_seconds
     */

    /**
     * Constructs a new ProjectApplyRequest.
     * @exports ProjectApplyRequest
     * @classdesc Represents a ProjectApplyRequest.
     * @implements IProjectApplyRequest
     * @constructor
     * @param {IProjectApplyRequest=} [properties] Properties to set
     */
    function ProjectApplyRequest(properties) {
        this.extra_values = [];
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * ProjectApplyRequest namespace_id.
     * @member {number} namespace_id
     * @memberof ProjectApplyRequest
     * @instance
     */
    ProjectApplyRequest.prototype.namespace_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * ProjectApplyRequest name.
     * @member {string} name
     * @memberof ProjectApplyRequest
     * @instance
     */
    ProjectApplyRequest.prototype.name = "";

    /**
     * ProjectApplyRequest gitlab_project_id.
     * @member {number} gitlab_project_id
     * @memberof ProjectApplyRequest
     * @instance
     */
    ProjectApplyRequest.prototype.gitlab_project_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * ProjectApplyRequest gitlab_branch.
     * @member {string} gitlab_branch
     * @memberof ProjectApplyRequest
     * @instance
     */
    ProjectApplyRequest.prototype.gitlab_branch = "";

    /**
     * ProjectApplyRequest gitlab_commit.
     * @member {string} gitlab_commit
     * @memberof ProjectApplyRequest
     * @instance
     */
    ProjectApplyRequest.prototype.gitlab_commit = "";

    /**
     * ProjectApplyRequest config.
     * @member {string} config
     * @memberof ProjectApplyRequest
     * @instance
     */
    ProjectApplyRequest.prototype.config = "";

    /**
     * ProjectApplyRequest atomic.
     * @member {boolean} atomic
     * @memberof ProjectApplyRequest
     * @instance
     */
    ProjectApplyRequest.prototype.atomic = false;

    /**
     * ProjectApplyRequest websocket_sync.
     * @member {boolean} websocket_sync
     * @memberof ProjectApplyRequest
     * @instance
     */
    ProjectApplyRequest.prototype.websocket_sync = false;

    /**
     * ProjectApplyRequest extra_values.
     * @member {Array.<ProjectExtraItem>} extra_values
     * @memberof ProjectApplyRequest
     * @instance
     */
    ProjectApplyRequest.prototype.extra_values = $util.emptyArray;

    /**
     * ProjectApplyRequest install_timeout_seconds.
     * @member {number} install_timeout_seconds
     * @memberof ProjectApplyRequest
     * @instance
     */
    ProjectApplyRequest.prototype.install_timeout_seconds = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * Encodes the specified ProjectApplyRequest message. Does not implicitly {@link ProjectApplyRequest.verify|verify} messages.
     * @function encode
     * @memberof ProjectApplyRequest
     * @static
     * @param {ProjectApplyRequest} message ProjectApplyRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    ProjectApplyRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.namespace_id != null && Object.hasOwnProperty.call(message, "namespace_id"))
            writer.uint32(/* id 1, wireType 0 =*/8).int64(message.namespace_id);
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
        if (message.atomic != null && Object.hasOwnProperty.call(message, "atomic"))
            writer.uint32(/* id 7, wireType 0 =*/56).bool(message.atomic);
        if (message.websocket_sync != null && Object.hasOwnProperty.call(message, "websocket_sync"))
            writer.uint32(/* id 8, wireType 0 =*/64).bool(message.websocket_sync);
        if (message.extra_values != null && message.extra_values.length)
            for (let i = 0; i < message.extra_values.length; ++i)
                $root.ProjectExtraItem.encode(message.extra_values[i], writer.uint32(/* id 9, wireType 2 =*/74).fork()).ldelim();
        if (message.install_timeout_seconds != null && Object.hasOwnProperty.call(message, "install_timeout_seconds"))
            writer.uint32(/* id 10, wireType 0 =*/80).int64(message.install_timeout_seconds);
        return writer;
    };

    /**
     * Decodes a ProjectApplyRequest message from the specified reader or buffer.
     * @function decode
     * @memberof ProjectApplyRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {ProjectApplyRequest} ProjectApplyRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    ProjectApplyRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.ProjectApplyRequest();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.namespace_id = reader.int64();
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
                message.atomic = reader.bool();
                break;
            case 8:
                message.websocket_sync = reader.bool();
                break;
            case 9:
                if (!(message.extra_values && message.extra_values.length))
                    message.extra_values = [];
                message.extra_values.push($root.ProjectExtraItem.decode(reader, reader.uint32()));
                break;
            case 10:
                message.install_timeout_seconds = reader.int64();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return ProjectApplyRequest;
})();

export const ProjectDeleteResponse = $root.ProjectDeleteResponse = (() => {

    /**
     * Properties of a ProjectDeleteResponse.
     * @exports IProjectDeleteResponse
     * @interface IProjectDeleteResponse
     */

    /**
     * Constructs a new ProjectDeleteResponse.
     * @exports ProjectDeleteResponse
     * @classdesc Represents a ProjectDeleteResponse.
     * @implements IProjectDeleteResponse
     * @constructor
     * @param {IProjectDeleteResponse=} [properties] Properties to set
     */
    function ProjectDeleteResponse(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * Encodes the specified ProjectDeleteResponse message. Does not implicitly {@link ProjectDeleteResponse.verify|verify} messages.
     * @function encode
     * @memberof ProjectDeleteResponse
     * @static
     * @param {ProjectDeleteResponse} message ProjectDeleteResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    ProjectDeleteResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        return writer;
    };

    /**
     * Decodes a ProjectDeleteResponse message from the specified reader or buffer.
     * @function decode
     * @memberof ProjectDeleteResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {ProjectDeleteResponse} ProjectDeleteResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    ProjectDeleteResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.ProjectDeleteResponse();
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

    return ProjectDeleteResponse;
})();

export const ProjectListRequest = $root.ProjectListRequest = (() => {

    /**
     * Properties of a ProjectListRequest.
     * @exports IProjectListRequest
     * @interface IProjectListRequest
     * @property {number|null} [page] ProjectListRequest page
     * @property {number|null} [page_size] ProjectListRequest page_size
     */

    /**
     * Constructs a new ProjectListRequest.
     * @exports ProjectListRequest
     * @classdesc Represents a ProjectListRequest.
     * @implements IProjectListRequest
     * @constructor
     * @param {IProjectListRequest=} [properties] Properties to set
     */
    function ProjectListRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * ProjectListRequest page.
     * @member {number} page
     * @memberof ProjectListRequest
     * @instance
     */
    ProjectListRequest.prototype.page = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * ProjectListRequest page_size.
     * @member {number} page_size
     * @memberof ProjectListRequest
     * @instance
     */
    ProjectListRequest.prototype.page_size = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * Encodes the specified ProjectListRequest message. Does not implicitly {@link ProjectListRequest.verify|verify} messages.
     * @function encode
     * @memberof ProjectListRequest
     * @static
     * @param {ProjectListRequest} message ProjectListRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    ProjectListRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.page != null && Object.hasOwnProperty.call(message, "page"))
            writer.uint32(/* id 1, wireType 0 =*/8).int64(message.page);
        if (message.page_size != null && Object.hasOwnProperty.call(message, "page_size"))
            writer.uint32(/* id 2, wireType 0 =*/16).int64(message.page_size);
        return writer;
    };

    /**
     * Decodes a ProjectListRequest message from the specified reader or buffer.
     * @function decode
     * @memberof ProjectListRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {ProjectListRequest} ProjectListRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    ProjectListRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.ProjectListRequest();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.page = reader.int64();
                break;
            case 2:
                message.page_size = reader.int64();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return ProjectListRequest;
})();

export const ProjectListResponse = $root.ProjectListResponse = (() => {

    /**
     * Properties of a ProjectListResponse.
     * @exports IProjectListResponse
     * @interface IProjectListResponse
     * @property {number|null} [page] ProjectListResponse page
     * @property {number|null} [page_size] ProjectListResponse page_size
     * @property {number|null} [count] ProjectListResponse count
     * @property {Array.<ProjectModel>|null} [data] ProjectListResponse data
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
     * ProjectListResponse page.
     * @member {number} page
     * @memberof ProjectListResponse
     * @instance
     */
    ProjectListResponse.prototype.page = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * ProjectListResponse page_size.
     * @member {number} page_size
     * @memberof ProjectListResponse
     * @instance
     */
    ProjectListResponse.prototype.page_size = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * ProjectListResponse count.
     * @member {number} count
     * @memberof ProjectListResponse
     * @instance
     */
    ProjectListResponse.prototype.count = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * ProjectListResponse data.
     * @member {Array.<ProjectModel>} data
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
        if (message.page != null && Object.hasOwnProperty.call(message, "page"))
            writer.uint32(/* id 1, wireType 0 =*/8).int64(message.page);
        if (message.page_size != null && Object.hasOwnProperty.call(message, "page_size"))
            writer.uint32(/* id 2, wireType 0 =*/16).int64(message.page_size);
        if (message.count != null && Object.hasOwnProperty.call(message, "count"))
            writer.uint32(/* id 3, wireType 0 =*/24).int64(message.count);
        if (message.data != null && message.data.length)
            for (let i = 0; i < message.data.length; ++i)
                $root.ProjectModel.encode(message.data[i], writer.uint32(/* id 4, wireType 2 =*/34).fork()).ldelim();
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
                message.page = reader.int64();
                break;
            case 2:
                message.page_size = reader.int64();
                break;
            case 3:
                message.count = reader.int64();
                break;
            case 4:
                if (!(message.data && message.data.length))
                    message.data = [];
                message.data.push($root.ProjectModel.decode(reader, reader.uint32()));
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

export const ProjectIsPodExistsRequest = $root.ProjectIsPodExistsRequest = (() => {

    /**
     * Properties of a ProjectIsPodExistsRequest.
     * @exports IProjectIsPodExistsRequest
     * @interface IProjectIsPodExistsRequest
     * @property {string|null} [namespace] ProjectIsPodExistsRequest namespace
     * @property {string|null} [pod] ProjectIsPodExistsRequest pod
     */

    /**
     * Constructs a new ProjectIsPodExistsRequest.
     * @exports ProjectIsPodExistsRequest
     * @classdesc Represents a ProjectIsPodExistsRequest.
     * @implements IProjectIsPodExistsRequest
     * @constructor
     * @param {IProjectIsPodExistsRequest=} [properties] Properties to set
     */
    function ProjectIsPodExistsRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * ProjectIsPodExistsRequest namespace.
     * @member {string} namespace
     * @memberof ProjectIsPodExistsRequest
     * @instance
     */
    ProjectIsPodExistsRequest.prototype.namespace = "";

    /**
     * ProjectIsPodExistsRequest pod.
     * @member {string} pod
     * @memberof ProjectIsPodExistsRequest
     * @instance
     */
    ProjectIsPodExistsRequest.prototype.pod = "";

    /**
     * Encodes the specified ProjectIsPodExistsRequest message. Does not implicitly {@link ProjectIsPodExistsRequest.verify|verify} messages.
     * @function encode
     * @memberof ProjectIsPodExistsRequest
     * @static
     * @param {ProjectIsPodExistsRequest} message ProjectIsPodExistsRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    ProjectIsPodExistsRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.namespace != null && Object.hasOwnProperty.call(message, "namespace"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.namespace);
        if (message.pod != null && Object.hasOwnProperty.call(message, "pod"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.pod);
        return writer;
    };

    /**
     * Decodes a ProjectIsPodExistsRequest message from the specified reader or buffer.
     * @function decode
     * @memberof ProjectIsPodExistsRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {ProjectIsPodExistsRequest} ProjectIsPodExistsRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    ProjectIsPodExistsRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.ProjectIsPodExistsRequest();
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

    return ProjectIsPodExistsRequest;
})();

export const ProjectIsPodExistsResponse = $root.ProjectIsPodExistsResponse = (() => {

    /**
     * Properties of a ProjectIsPodExistsResponse.
     * @exports IProjectIsPodExistsResponse
     * @interface IProjectIsPodExistsResponse
     * @property {boolean|null} [exists] ProjectIsPodExistsResponse exists
     */

    /**
     * Constructs a new ProjectIsPodExistsResponse.
     * @exports ProjectIsPodExistsResponse
     * @classdesc Represents a ProjectIsPodExistsResponse.
     * @implements IProjectIsPodExistsResponse
     * @constructor
     * @param {IProjectIsPodExistsResponse=} [properties] Properties to set
     */
    function ProjectIsPodExistsResponse(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * ProjectIsPodExistsResponse exists.
     * @member {boolean} exists
     * @memberof ProjectIsPodExistsResponse
     * @instance
     */
    ProjectIsPodExistsResponse.prototype.exists = false;

    /**
     * Encodes the specified ProjectIsPodExistsResponse message. Does not implicitly {@link ProjectIsPodExistsResponse.verify|verify} messages.
     * @function encode
     * @memberof ProjectIsPodExistsResponse
     * @static
     * @param {ProjectIsPodExistsResponse} message ProjectIsPodExistsResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    ProjectIsPodExistsResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.exists != null && Object.hasOwnProperty.call(message, "exists"))
            writer.uint32(/* id 1, wireType 0 =*/8).bool(message.exists);
        return writer;
    };

    /**
     * Decodes a ProjectIsPodExistsResponse message from the specified reader or buffer.
     * @function decode
     * @memberof ProjectIsPodExistsResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {ProjectIsPodExistsResponse} ProjectIsPodExistsResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    ProjectIsPodExistsResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.ProjectIsPodExistsResponse();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.exists = reader.bool();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return ProjectIsPodExistsResponse;
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
     * Callback as used by {@link Project#list}.
     * @memberof Project
     * @typedef ListCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {ProjectListResponse} [response] ProjectListResponse
     */

    /**
     * Calls List.
     * @function list
     * @memberof Project
     * @instance
     * @param {ProjectListRequest} request ProjectListRequest message or plain object
     * @param {Project.ListCallback} callback Node-style callback called with the error, if any, and ProjectListResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Project.prototype.list = function list(request, callback) {
        return this.rpcCall(list, $root.ProjectListRequest, $root.ProjectListResponse, request, callback);
    }, "name", { value: "List" });

    /**
     * Calls List.
     * @function list
     * @memberof Project
     * @instance
     * @param {ProjectListRequest} request ProjectListRequest message or plain object
     * @returns {Promise<ProjectListResponse>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link Project#apply}.
     * @memberof Project
     * @typedef ApplyCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {ProjectApplyResponse} [response] ProjectApplyResponse
     */

    /**
     * Calls Apply.
     * @function apply
     * @memberof Project
     * @instance
     * @param {ProjectApplyRequest} request ProjectApplyRequest message or plain object
     * @param {Project.ApplyCallback} callback Node-style callback called with the error, if any, and ProjectApplyResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Project.prototype.apply = function apply(request, callback) {
        return this.rpcCall(apply, $root.ProjectApplyRequest, $root.ProjectApplyResponse, request, callback);
    }, "name", { value: "Apply" });

    /**
     * Calls Apply.
     * @function apply
     * @memberof Project
     * @instance
     * @param {ProjectApplyRequest} request ProjectApplyRequest message or plain object
     * @returns {Promise<ProjectApplyResponse>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link Project#applyDryRun}.
     * @memberof Project
     * @typedef ApplyDryRunCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {ProjectDryRunApplyResponse} [response] ProjectDryRunApplyResponse
     */

    /**
     * Calls ApplyDryRun.
     * @function applyDryRun
     * @memberof Project
     * @instance
     * @param {ProjectApplyRequest} request ProjectApplyRequest message or plain object
     * @param {Project.ApplyDryRunCallback} callback Node-style callback called with the error, if any, and ProjectDryRunApplyResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Project.prototype.applyDryRun = function applyDryRun(request, callback) {
        return this.rpcCall(applyDryRun, $root.ProjectApplyRequest, $root.ProjectDryRunApplyResponse, request, callback);
    }, "name", { value: "ApplyDryRun" });

    /**
     * Calls ApplyDryRun.
     * @function applyDryRun
     * @memberof Project
     * @instance
     * @param {ProjectApplyRequest} request ProjectApplyRequest message or plain object
     * @returns {Promise<ProjectDryRunApplyResponse>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link Project#delete_}.
     * @memberof Project
     * @typedef DeleteCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {ProjectDeleteResponse} [response] ProjectDeleteResponse
     */

    /**
     * Calls Delete.
     * @function delete
     * @memberof Project
     * @instance
     * @param {ProjectDeleteRequest} request ProjectDeleteRequest message or plain object
     * @param {Project.DeleteCallback} callback Node-style callback called with the error, if any, and ProjectDeleteResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Project.prototype["delete"] = function delete_(request, callback) {
        return this.rpcCall(delete_, $root.ProjectDeleteRequest, $root.ProjectDeleteResponse, request, callback);
    }, "name", { value: "Delete" });

    /**
     * Calls Delete.
     * @function delete
     * @memberof Project
     * @instance
     * @param {ProjectDeleteRequest} request ProjectDeleteRequest message or plain object
     * @returns {Promise<ProjectDeleteResponse>} Promise
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
     * @param {ProjectIsPodRunningResponse} [response] ProjectIsPodRunningResponse
     */

    /**
     * Calls IsPodRunning.
     * @function isPodRunning
     * @memberof Project
     * @instance
     * @param {ProjectIsPodRunningRequest} request ProjectIsPodRunningRequest message or plain object
     * @param {Project.IsPodRunningCallback} callback Node-style callback called with the error, if any, and ProjectIsPodRunningResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Project.prototype.isPodRunning = function isPodRunning(request, callback) {
        return this.rpcCall(isPodRunning, $root.ProjectIsPodRunningRequest, $root.ProjectIsPodRunningResponse, request, callback);
    }, "name", { value: "IsPodRunning" });

    /**
     * Calls IsPodRunning.
     * @function isPodRunning
     * @memberof Project
     * @instance
     * @param {ProjectIsPodRunningRequest} request ProjectIsPodRunningRequest message or plain object
     * @returns {Promise<ProjectIsPodRunningResponse>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link Project#isPodExists}.
     * @memberof Project
     * @typedef IsPodExistsCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {ProjectIsPodExistsResponse} [response] ProjectIsPodExistsResponse
     */

    /**
     * Calls IsPodExists.
     * @function isPodExists
     * @memberof Project
     * @instance
     * @param {ProjectIsPodExistsRequest} request ProjectIsPodExistsRequest message or plain object
     * @param {Project.IsPodExistsCallback} callback Node-style callback called with the error, if any, and ProjectIsPodExistsResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Project.prototype.isPodExists = function isPodExists(request, callback) {
        return this.rpcCall(isPodExists, $root.ProjectIsPodExistsRequest, $root.ProjectIsPodExistsResponse, request, callback);
    }, "name", { value: "IsPodExists" });

    /**
     * Calls IsPodExists.
     * @function isPodExists
     * @memberof Project
     * @instance
     * @param {ProjectIsPodExistsRequest} request ProjectIsPodExistsRequest message or plain object
     * @returns {Promise<ProjectIsPodExistsResponse>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link Project#allPodContainers}.
     * @memberof Project
     * @typedef AllPodContainersCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {ProjectAllPodContainersResponse} [response] ProjectAllPodContainersResponse
     */

    /**
     * Calls AllPodContainers.
     * @function allPodContainers
     * @memberof Project
     * @instance
     * @param {ProjectAllPodContainersRequest} request ProjectAllPodContainersRequest message or plain object
     * @param {Project.AllPodContainersCallback} callback Node-style callback called with the error, if any, and ProjectAllPodContainersResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Project.prototype.allPodContainers = function allPodContainers(request, callback) {
        return this.rpcCall(allPodContainers, $root.ProjectAllPodContainersRequest, $root.ProjectAllPodContainersResponse, request, callback);
    }, "name", { value: "AllPodContainers" });

    /**
     * Calls AllPodContainers.
     * @function allPodContainers
     * @memberof Project
     * @instance
     * @param {ProjectAllPodContainersRequest} request ProjectAllPodContainersRequest message or plain object
     * @returns {Promise<ProjectAllPodContainersResponse>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link Project#podContainerLog}.
     * @memberof Project
     * @typedef PodContainerLogCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {ProjectPodContainerLogResponse} [response] ProjectPodContainerLogResponse
     */

    /**
     * Calls PodContainerLog.
     * @function podContainerLog
     * @memberof Project
     * @instance
     * @param {ProjectPodContainerLogRequest} request ProjectPodContainerLogRequest message or plain object
     * @param {Project.PodContainerLogCallback} callback Node-style callback called with the error, if any, and ProjectPodContainerLogResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Project.prototype.podContainerLog = function podContainerLog(request, callback) {
        return this.rpcCall(podContainerLog, $root.ProjectPodContainerLogRequest, $root.ProjectPodContainerLogResponse, request, callback);
    }, "name", { value: "PodContainerLog" });

    /**
     * Calls PodContainerLog.
     * @function podContainerLog
     * @memberof Project
     * @instance
     * @param {ProjectPodContainerLogRequest} request ProjectPodContainerLogRequest message or plain object
     * @returns {Promise<ProjectPodContainerLogResponse>} Promise
     * @variation 2
     */

    /**
     * Callback as used by {@link Project#streamPodContainerLog}.
     * @memberof Project
     * @typedef StreamPodContainerLogCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {ProjectPodContainerLogResponse} [response] ProjectPodContainerLogResponse
     */

    /**
     * Calls StreamPodContainerLog.
     * @function streamPodContainerLog
     * @memberof Project
     * @instance
     * @param {ProjectPodContainerLogRequest} request ProjectPodContainerLogRequest message or plain object
     * @param {Project.StreamPodContainerLogCallback} callback Node-style callback called with the error, if any, and ProjectPodContainerLogResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Project.prototype.streamPodContainerLog = function streamPodContainerLog(request, callback) {
        return this.rpcCall(streamPodContainerLog, $root.ProjectPodContainerLogRequest, $root.ProjectPodContainerLogResponse, request, callback);
    }, "name", { value: "StreamPodContainerLog" });

    /**
     * Calls StreamPodContainerLog.
     * @function streamPodContainerLog
     * @memberof Project
     * @instance
     * @param {ProjectPodContainerLogRequest} request ProjectPodContainerLogRequest message or plain object
     * @returns {Promise<ProjectPodContainerLogResponse>} Promise
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

export const VersionRequest = $root.VersionRequest = (() => {

    /**
     * Properties of a VersionRequest.
     * @exports IVersionRequest
     * @interface IVersionRequest
     */

    /**
     * Constructs a new VersionRequest.
     * @exports VersionRequest
     * @classdesc Represents a VersionRequest.
     * @implements IVersionRequest
     * @constructor
     * @param {IVersionRequest=} [properties] Properties to set
     */
    function VersionRequest(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * Encodes the specified VersionRequest message. Does not implicitly {@link VersionRequest.verify|verify} messages.
     * @function encode
     * @memberof VersionRequest
     * @static
     * @param {VersionRequest} message VersionRequest message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    VersionRequest.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        return writer;
    };

    /**
     * Decodes a VersionRequest message from the specified reader or buffer.
     * @function decode
     * @memberof VersionRequest
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {VersionRequest} VersionRequest
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    VersionRequest.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.VersionRequest();
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

    return VersionRequest;
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
     * Callback as used by {@link Version#version}.
     * @memberof Version
     * @typedef VersionCallback
     * @type {function}
     * @param {Error|null} error Error, if any
     * @param {VersionResponse} [response] VersionResponse
     */

    /**
     * Calls Version.
     * @function version
     * @memberof Version
     * @instance
     * @param {VersionRequest} request VersionRequest message or plain object
     * @param {Version.VersionCallback} callback Node-style callback called with the error, if any, and VersionResponse
     * @returns {undefined}
     * @variation 1
     */
    Object.defineProperty(Version.prototype.version = function version(request, callback) {
        return this.rpcCall(version, $root.VersionRequest, $root.VersionResponse, request, callback);
    }, "name", { value: "Version" });

    /**
     * Calls Version.
     * @function version
     * @memberof Version
     * @instance
     * @param {VersionRequest} request VersionRequest message or plain object
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
 * @property {number} ApplyProject=9 ApplyProject value
 * @property {number} HandleExecShell=50 HandleExecShell value
 * @property {number} HandleExecShellMsg=51 HandleExecShellMsg value
 * @property {number} HandleCloseShell=52 HandleCloseShell value
 * @property {number} HandleAuthorize=53 HandleAuthorize value
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
    values[valuesById[9] = "ApplyProject"] = 9;
    values[valuesById[50] = "HandleExecShell"] = 50;
    values[valuesById[51] = "HandleExecShellMsg"] = 51;
    values[valuesById[52] = "HandleCloseShell"] = 52;
    values[valuesById[53] = "HandleAuthorize"] = 53;
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

export const ProjectExtraItem = $root.ProjectExtraItem = (() => {

    /**
     * Properties of a ProjectExtraItem.
     * @exports IProjectExtraItem
     * @interface IProjectExtraItem
     * @property {string|null} [path] ProjectExtraItem path
     * @property {string|null} [value] ProjectExtraItem value
     */

    /**
     * Constructs a new ProjectExtraItem.
     * @exports ProjectExtraItem
     * @classdesc Represents a ProjectExtraItem.
     * @implements IProjectExtraItem
     * @constructor
     * @param {IProjectExtraItem=} [properties] Properties to set
     */
    function ProjectExtraItem(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * ProjectExtraItem path.
     * @member {string} path
     * @memberof ProjectExtraItem
     * @instance
     */
    ProjectExtraItem.prototype.path = "";

    /**
     * ProjectExtraItem value.
     * @member {string} value
     * @memberof ProjectExtraItem
     * @instance
     */
    ProjectExtraItem.prototype.value = "";

    /**
     * Encodes the specified ProjectExtraItem message. Does not implicitly {@link ProjectExtraItem.verify|verify} messages.
     * @function encode
     * @memberof ProjectExtraItem
     * @static
     * @param {ProjectExtraItem} message ProjectExtraItem message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    ProjectExtraItem.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.path != null && Object.hasOwnProperty.call(message, "path"))
            writer.uint32(/* id 1, wireType 2 =*/10).string(message.path);
        if (message.value != null && Object.hasOwnProperty.call(message, "value"))
            writer.uint32(/* id 2, wireType 2 =*/18).string(message.value);
        return writer;
    };

    /**
     * Decodes a ProjectExtraItem message from the specified reader or buffer.
     * @function decode
     * @memberof ProjectExtraItem
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {ProjectExtraItem} ProjectExtraItem
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    ProjectExtraItem.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.ProjectExtraItem();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.path = reader.string();
                break;
            case 2:
                message.value = reader.string();
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return ProjectExtraItem;
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
     * @property {Array.<ProjectExtraItem>|null} [extra_values] ProjectInput extra_values
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
        this.extra_values = [];
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
     * ProjectInput extra_values.
     * @member {Array.<ProjectExtraItem>} extra_values
     * @memberof ProjectInput
     * @instance
     */
    ProjectInput.prototype.extra_values = $util.emptyArray;

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
        if (message.extra_values != null && message.extra_values.length)
            for (let i = 0; i < message.extra_values.length; ++i)
                $root.ProjectExtraItem.encode(message.extra_values[i], writer.uint32(/* id 9, wireType 2 =*/74).fork()).ldelim();
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
            case 9:
                if (!(message.extra_values && message.extra_values.length))
                    message.extra_values = [];
                message.extra_values.push($root.ProjectExtraItem.decode(reader, reader.uint32()));
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
     * @property {Array.<ProjectExtraItem>|null} [extra_values] UpdateProjectInput extra_values
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
        this.extra_values = [];
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
     * UpdateProjectInput extra_values.
     * @member {Array.<ProjectExtraItem>} extra_values
     * @memberof UpdateProjectInput
     * @instance
     */
    UpdateProjectInput.prototype.extra_values = $util.emptyArray;

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
        if (message.extra_values != null && message.extra_values.length)
            for (let i = 0; i < message.extra_values.length; ++i)
                $root.ProjectExtraItem.encode(message.extra_values[i], writer.uint32(/* id 7, wireType 2 =*/58).fork()).ldelim();
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
            case 7:
                if (!(message.extra_values && message.extra_values.length))
                    message.extra_values = [];
                message.extra_values.push($root.ProjectExtraItem.decode(reader, reader.uint32()));
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

export const Metadata = $root.Metadata = (() => {

    /**
     * Properties of a Metadata.
     * @exports IMetadata
     * @interface IMetadata
     * @property {string|null} [id] Metadata id
     * @property {string|null} [uid] Metadata uid
     * @property {string|null} [slug] Metadata slug
     * @property {Type|null} [type] Metadata type
     * @property {boolean|null} [end] Metadata end
     * @property {ResultType|null} [result] Metadata result
     * @property {To|null} [to] Metadata to
     * @property {string|null} [data] Metadata data
     */

    /**
     * Constructs a new Metadata.
     * @exports Metadata
     * @classdesc Represents a Metadata.
     * @implements IMetadata
     * @constructor
     * @param {IMetadata=} [properties] Properties to set
     */
    function Metadata(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * Metadata id.
     * @member {string} id
     * @memberof Metadata
     * @instance
     */
    Metadata.prototype.id = "";

    /**
     * Metadata uid.
     * @member {string} uid
     * @memberof Metadata
     * @instance
     */
    Metadata.prototype.uid = "";

    /**
     * Metadata slug.
     * @member {string} slug
     * @memberof Metadata
     * @instance
     */
    Metadata.prototype.slug = "";

    /**
     * Metadata type.
     * @member {Type} type
     * @memberof Metadata
     * @instance
     */
    Metadata.prototype.type = 0;

    /**
     * Metadata end.
     * @member {boolean} end
     * @memberof Metadata
     * @instance
     */
    Metadata.prototype.end = false;

    /**
     * Metadata result.
     * @member {ResultType} result
     * @memberof Metadata
     * @instance
     */
    Metadata.prototype.result = 0;

    /**
     * Metadata to.
     * @member {To} to
     * @memberof Metadata
     * @instance
     */
    Metadata.prototype.to = 0;

    /**
     * Metadata data.
     * @member {string} data
     * @memberof Metadata
     * @instance
     */
    Metadata.prototype.data = "";

    /**
     * Encodes the specified Metadata message. Does not implicitly {@link Metadata.verify|verify} messages.
     * @function encode
     * @memberof Metadata
     * @static
     * @param {Metadata} message Metadata message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    Metadata.encode = function encode(message, writer) {
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
     * Decodes a Metadata message from the specified reader or buffer.
     * @function decode
     * @memberof Metadata
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {Metadata} Metadata
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    Metadata.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.Metadata();
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

    return Metadata;
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

export const WsMetadataResponse = $root.WsMetadataResponse = (() => {

    /**
     * Properties of a WsMetadataResponse.
     * @exports IWsMetadataResponse
     * @interface IWsMetadataResponse
     * @property {Metadata|null} [metadata] WsMetadataResponse metadata
     */

    /**
     * Constructs a new WsMetadataResponse.
     * @exports WsMetadataResponse
     * @classdesc Represents a WsMetadataResponse.
     * @implements IWsMetadataResponse
     * @constructor
     * @param {IWsMetadataResponse=} [properties] Properties to set
     */
    function WsMetadataResponse(properties) {
        if (properties)
            for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                if (properties[keys[i]] != null)
                    this[keys[i]] = properties[keys[i]];
    }

    /**
     * WsMetadataResponse metadata.
     * @member {Metadata|null|undefined} metadata
     * @memberof WsMetadataResponse
     * @instance
     */
    WsMetadataResponse.prototype.metadata = null;

    /**
     * Encodes the specified WsMetadataResponse message. Does not implicitly {@link WsMetadataResponse.verify|verify} messages.
     * @function encode
     * @memberof WsMetadataResponse
     * @static
     * @param {WsMetadataResponse} message WsMetadataResponse message or plain object to encode
     * @param {$protobuf.Writer} [writer] Writer to encode to
     * @returns {$protobuf.Writer} Writer
     */
    WsMetadataResponse.encode = function encode(message, writer) {
        if (!writer)
            writer = $Writer.create();
        if (message.metadata != null && Object.hasOwnProperty.call(message, "metadata"))
            $root.Metadata.encode(message.metadata, writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
        return writer;
    };

    /**
     * Decodes a WsMetadataResponse message from the specified reader or buffer.
     * @function decode
     * @memberof WsMetadataResponse
     * @static
     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
     * @param {number} [length] Message length if known beforehand
     * @returns {WsMetadataResponse} WsMetadataResponse
     * @throws {Error} If the payload is not a reader or valid buffer
     * @throws {$protobuf.util.ProtocolError} If required fields are missing
     */
    WsMetadataResponse.decode = function decode(reader, length) {
        if (!(reader instanceof $Reader))
            reader = $Reader.create(reader);
        let end = length === undefined ? reader.len : reader.pos + length, message = new $root.WsMetadataResponse();
        while (reader.pos < end) {
            let tag = reader.uint32();
            switch (tag >>> 3) {
            case 1:
                message.metadata = $root.Metadata.decode(reader, reader.uint32());
                break;
            default:
                reader.skipType(tag & 7);
                break;
            }
        }
        return message;
    };

    return WsMetadataResponse;
})();

export const WsHandleShellResponse = $root.WsHandleShellResponse = (() => {

    /**
     * Properties of a WsHandleShellResponse.
     * @exports IWsHandleShellResponse
     * @interface IWsHandleShellResponse
     * @property {Metadata|null} [metadata] WsHandleShellResponse metadata
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
     * @member {Metadata|null|undefined} metadata
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
            $root.Metadata.encode(message.metadata, writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
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
                message.metadata = $root.Metadata.decode(reader, reader.uint32());
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
     * @property {Metadata|null} [metadata] WsHandleClusterResponse metadata
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
     * @member {Metadata|null|undefined} metadata
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
            $root.Metadata.encode(message.metadata, writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
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
                message.metadata = $root.Metadata.decode(reader, reader.uint32());
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
