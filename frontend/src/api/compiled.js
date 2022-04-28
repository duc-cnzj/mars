/*eslint-disable block-scoped-var, id-length, no-control-regex, no-magic-numbers, no-prototype-builtins, no-redeclare, no-shadow, no-var, sort-vars*/
import * as $protobuf from "protobufjs/minimal";

// Common aliases
const $Reader = $protobuf.Reader, $Writer = $protobuf.Writer, $util = $protobuf.util;

// Exported root namespace
const $root = $protobuf.roots["default"] || ($protobuf.roots["default"] = {});

export const auth = $root.auth = (() => {

    /**
     * Namespace auth.
     * @exports auth
     * @namespace
     */
    const auth = {};

    auth.LoginRequest = (function() {

        /**
         * Properties of a LoginRequest.
         * @memberof auth
         * @interface ILoginRequest
         * @property {string|null} [username] LoginRequest username
         * @property {string|null} [password] LoginRequest password
         */

        /**
         * Constructs a new LoginRequest.
         * @memberof auth
         * @classdesc Represents a LoginRequest.
         * @implements ILoginRequest
         * @constructor
         * @param {auth.ILoginRequest=} [properties] Properties to set
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
         * @memberof auth.LoginRequest
         * @instance
         */
        LoginRequest.prototype.username = "";

        /**
         * LoginRequest password.
         * @member {string} password
         * @memberof auth.LoginRequest
         * @instance
         */
        LoginRequest.prototype.password = "";

        /**
         * Encodes the specified LoginRequest message. Does not implicitly {@link auth.LoginRequest.verify|verify} messages.
         * @function encode
         * @memberof auth.LoginRequest
         * @static
         * @param {auth.LoginRequest} message LoginRequest message or plain object to encode
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
         * @memberof auth.LoginRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {auth.LoginRequest} LoginRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        LoginRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.auth.LoginRequest();
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

    auth.LoginResponse = (function() {

        /**
         * Properties of a LoginResponse.
         * @memberof auth
         * @interface ILoginResponse
         * @property {string|null} [token] LoginResponse token
         * @property {number|null} [expires_in] LoginResponse expires_in
         */

        /**
         * Constructs a new LoginResponse.
         * @memberof auth
         * @classdesc Represents a LoginResponse.
         * @implements ILoginResponse
         * @constructor
         * @param {auth.ILoginResponse=} [properties] Properties to set
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
         * @memberof auth.LoginResponse
         * @instance
         */
        LoginResponse.prototype.token = "";

        /**
         * LoginResponse expires_in.
         * @member {number} expires_in
         * @memberof auth.LoginResponse
         * @instance
         */
        LoginResponse.prototype.expires_in = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Encodes the specified LoginResponse message. Does not implicitly {@link auth.LoginResponse.verify|verify} messages.
         * @function encode
         * @memberof auth.LoginResponse
         * @static
         * @param {auth.LoginResponse} message LoginResponse message or plain object to encode
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
         * @memberof auth.LoginResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {auth.LoginResponse} LoginResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        LoginResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.auth.LoginResponse();
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

    auth.ExchangeRequest = (function() {

        /**
         * Properties of an ExchangeRequest.
         * @memberof auth
         * @interface IExchangeRequest
         * @property {string|null} [code] ExchangeRequest code
         */

        /**
         * Constructs a new ExchangeRequest.
         * @memberof auth
         * @classdesc Represents an ExchangeRequest.
         * @implements IExchangeRequest
         * @constructor
         * @param {auth.IExchangeRequest=} [properties] Properties to set
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
         * @memberof auth.ExchangeRequest
         * @instance
         */
        ExchangeRequest.prototype.code = "";

        /**
         * Encodes the specified ExchangeRequest message. Does not implicitly {@link auth.ExchangeRequest.verify|verify} messages.
         * @function encode
         * @memberof auth.ExchangeRequest
         * @static
         * @param {auth.ExchangeRequest} message ExchangeRequest message or plain object to encode
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
         * @memberof auth.ExchangeRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {auth.ExchangeRequest} ExchangeRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ExchangeRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.auth.ExchangeRequest();
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

    auth.ExchangeResponse = (function() {

        /**
         * Properties of an ExchangeResponse.
         * @memberof auth
         * @interface IExchangeResponse
         * @property {string|null} [token] ExchangeResponse token
         * @property {number|null} [expires_in] ExchangeResponse expires_in
         */

        /**
         * Constructs a new ExchangeResponse.
         * @memberof auth
         * @classdesc Represents an ExchangeResponse.
         * @implements IExchangeResponse
         * @constructor
         * @param {auth.IExchangeResponse=} [properties] Properties to set
         */
        function ExchangeResponse(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * ExchangeResponse token.
         * @member {string} token
         * @memberof auth.ExchangeResponse
         * @instance
         */
        ExchangeResponse.prototype.token = "";

        /**
         * ExchangeResponse expires_in.
         * @member {number} expires_in
         * @memberof auth.ExchangeResponse
         * @instance
         */
        ExchangeResponse.prototype.expires_in = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Encodes the specified ExchangeResponse message. Does not implicitly {@link auth.ExchangeResponse.verify|verify} messages.
         * @function encode
         * @memberof auth.ExchangeResponse
         * @static
         * @param {auth.ExchangeResponse} message ExchangeResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ExchangeResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.token != null && Object.hasOwnProperty.call(message, "token"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.token);
            if (message.expires_in != null && Object.hasOwnProperty.call(message, "expires_in"))
                writer.uint32(/* id 2, wireType 0 =*/16).int64(message.expires_in);
            return writer;
        };

        /**
         * Decodes an ExchangeResponse message from the specified reader or buffer.
         * @function decode
         * @memberof auth.ExchangeResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {auth.ExchangeResponse} ExchangeResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ExchangeResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.auth.ExchangeResponse();
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

        return ExchangeResponse;
    })();

    auth.InfoRequest = (function() {

        /**
         * Properties of an InfoRequest.
         * @memberof auth
         * @interface IInfoRequest
         */

        /**
         * Constructs a new InfoRequest.
         * @memberof auth
         * @classdesc Represents an InfoRequest.
         * @implements IInfoRequest
         * @constructor
         * @param {auth.IInfoRequest=} [properties] Properties to set
         */
        function InfoRequest(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * Encodes the specified InfoRequest message. Does not implicitly {@link auth.InfoRequest.verify|verify} messages.
         * @function encode
         * @memberof auth.InfoRequest
         * @static
         * @param {auth.InfoRequest} message InfoRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        InfoRequest.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            return writer;
        };

        /**
         * Decodes an InfoRequest message from the specified reader or buffer.
         * @function decode
         * @memberof auth.InfoRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {auth.InfoRequest} InfoRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        InfoRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.auth.InfoRequest();
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

        return InfoRequest;
    })();

    auth.InfoResponse = (function() {

        /**
         * Properties of an InfoResponse.
         * @memberof auth
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
         * @memberof auth
         * @classdesc Represents an InfoResponse.
         * @implements IInfoResponse
         * @constructor
         * @param {auth.IInfoResponse=} [properties] Properties to set
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
         * @memberof auth.InfoResponse
         * @instance
         */
        InfoResponse.prototype.id = "";

        /**
         * InfoResponse avatar.
         * @member {string} avatar
         * @memberof auth.InfoResponse
         * @instance
         */
        InfoResponse.prototype.avatar = "";

        /**
         * InfoResponse name.
         * @member {string} name
         * @memberof auth.InfoResponse
         * @instance
         */
        InfoResponse.prototype.name = "";

        /**
         * InfoResponse email.
         * @member {string} email
         * @memberof auth.InfoResponse
         * @instance
         */
        InfoResponse.prototype.email = "";

        /**
         * InfoResponse logout_url.
         * @member {string} logout_url
         * @memberof auth.InfoResponse
         * @instance
         */
        InfoResponse.prototype.logout_url = "";

        /**
         * InfoResponse roles.
         * @member {Array.<string>} roles
         * @memberof auth.InfoResponse
         * @instance
         */
        InfoResponse.prototype.roles = $util.emptyArray;

        /**
         * Encodes the specified InfoResponse message. Does not implicitly {@link auth.InfoResponse.verify|verify} messages.
         * @function encode
         * @memberof auth.InfoResponse
         * @static
         * @param {auth.InfoResponse} message InfoResponse message or plain object to encode
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
         * @memberof auth.InfoResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {auth.InfoResponse} InfoResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        InfoResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.auth.InfoResponse();
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

    auth.SettingsRequest = (function() {

        /**
         * Properties of a SettingsRequest.
         * @memberof auth
         * @interface ISettingsRequest
         */

        /**
         * Constructs a new SettingsRequest.
         * @memberof auth
         * @classdesc Represents a SettingsRequest.
         * @implements ISettingsRequest
         * @constructor
         * @param {auth.ISettingsRequest=} [properties] Properties to set
         */
        function SettingsRequest(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * Encodes the specified SettingsRequest message. Does not implicitly {@link auth.SettingsRequest.verify|verify} messages.
         * @function encode
         * @memberof auth.SettingsRequest
         * @static
         * @param {auth.SettingsRequest} message SettingsRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        SettingsRequest.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            return writer;
        };

        /**
         * Decodes a SettingsRequest message from the specified reader or buffer.
         * @function decode
         * @memberof auth.SettingsRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {auth.SettingsRequest} SettingsRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        SettingsRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.auth.SettingsRequest();
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

        return SettingsRequest;
    })();

    auth.SettingsResponse = (function() {

        /**
         * Properties of a SettingsResponse.
         * @memberof auth
         * @interface ISettingsResponse
         * @property {Array.<auth.SettingsResponse.OidcSetting>|null} [items] SettingsResponse items
         */

        /**
         * Constructs a new SettingsResponse.
         * @memberof auth
         * @classdesc Represents a SettingsResponse.
         * @implements ISettingsResponse
         * @constructor
         * @param {auth.ISettingsResponse=} [properties] Properties to set
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
         * @member {Array.<auth.SettingsResponse.OidcSetting>} items
         * @memberof auth.SettingsResponse
         * @instance
         */
        SettingsResponse.prototype.items = $util.emptyArray;

        /**
         * Encodes the specified SettingsResponse message. Does not implicitly {@link auth.SettingsResponse.verify|verify} messages.
         * @function encode
         * @memberof auth.SettingsResponse
         * @static
         * @param {auth.SettingsResponse} message SettingsResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        SettingsResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.items != null && message.items.length)
                for (let i = 0; i < message.items.length; ++i)
                    $root.auth.SettingsResponse.OidcSetting.encode(message.items[i], writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
            return writer;
        };

        /**
         * Decodes a SettingsResponse message from the specified reader or buffer.
         * @function decode
         * @memberof auth.SettingsResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {auth.SettingsResponse} SettingsResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        SettingsResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.auth.SettingsResponse();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1:
                    if (!(message.items && message.items.length))
                        message.items = [];
                    message.items.push($root.auth.SettingsResponse.OidcSetting.decode(reader, reader.uint32()));
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        SettingsResponse.OidcSetting = (function() {

            /**
             * Properties of an OidcSetting.
             * @memberof auth.SettingsResponse
             * @interface IOidcSetting
             * @property {boolean|null} [enabled] OidcSetting enabled
             * @property {string|null} [name] OidcSetting name
             * @property {string|null} [url] OidcSetting url
             * @property {string|null} [end_session_endpoint] OidcSetting end_session_endpoint
             * @property {string|null} [state] OidcSetting state
             */

            /**
             * Constructs a new OidcSetting.
             * @memberof auth.SettingsResponse
             * @classdesc Represents an OidcSetting.
             * @implements IOidcSetting
             * @constructor
             * @param {auth.SettingsResponse.IOidcSetting=} [properties] Properties to set
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
             * @memberof auth.SettingsResponse.OidcSetting
             * @instance
             */
            OidcSetting.prototype.enabled = false;

            /**
             * OidcSetting name.
             * @member {string} name
             * @memberof auth.SettingsResponse.OidcSetting
             * @instance
             */
            OidcSetting.prototype.name = "";

            /**
             * OidcSetting url.
             * @member {string} url
             * @memberof auth.SettingsResponse.OidcSetting
             * @instance
             */
            OidcSetting.prototype.url = "";

            /**
             * OidcSetting end_session_endpoint.
             * @member {string} end_session_endpoint
             * @memberof auth.SettingsResponse.OidcSetting
             * @instance
             */
            OidcSetting.prototype.end_session_endpoint = "";

            /**
             * OidcSetting state.
             * @member {string} state
             * @memberof auth.SettingsResponse.OidcSetting
             * @instance
             */
            OidcSetting.prototype.state = "";

            /**
             * Encodes the specified OidcSetting message. Does not implicitly {@link auth.SettingsResponse.OidcSetting.verify|verify} messages.
             * @function encode
             * @memberof auth.SettingsResponse.OidcSetting
             * @static
             * @param {auth.SettingsResponse.OidcSetting} message OidcSetting message or plain object to encode
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
             * @memberof auth.SettingsResponse.OidcSetting
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @param {number} [length] Message length if known beforehand
             * @returns {auth.SettingsResponse.OidcSetting} OidcSetting
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            OidcSetting.decode = function decode(reader, length) {
                if (!(reader instanceof $Reader))
                    reader = $Reader.create(reader);
                let end = length === undefined ? reader.len : reader.pos + length, message = new $root.auth.SettingsResponse.OidcSetting();
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

        return SettingsResponse;
    })();

    auth.Auth = (function() {

        /**
         * Constructs a new Auth service.
         * @memberof auth
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
         * Callback as used by {@link auth.Auth#login}.
         * @memberof auth.Auth
         * @typedef LoginCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {auth.LoginResponse} [response] LoginResponse
         */

        /**
         * Calls Login.
         * @function login
         * @memberof auth.Auth
         * @instance
         * @param {auth.LoginRequest} request LoginRequest message or plain object
         * @param {auth.Auth.LoginCallback} callback Node-style callback called with the error, if any, and LoginResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(Auth.prototype.login = function login(request, callback) {
            return this.rpcCall(login, $root.auth.LoginRequest, $root.auth.LoginResponse, request, callback);
        }, "name", { value: "Login" });

        /**
         * Calls Login.
         * @function login
         * @memberof auth.Auth
         * @instance
         * @param {auth.LoginRequest} request LoginRequest message or plain object
         * @returns {Promise<auth.LoginResponse>} Promise
         * @variation 2
         */

        /**
         * Callback as used by {@link auth.Auth#info}.
         * @memberof auth.Auth
         * @typedef InfoCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {auth.InfoResponse} [response] InfoResponse
         */

        /**
         * Calls Info.
         * @function info
         * @memberof auth.Auth
         * @instance
         * @param {auth.InfoRequest} request InfoRequest message or plain object
         * @param {auth.Auth.InfoCallback} callback Node-style callback called with the error, if any, and InfoResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(Auth.prototype.info = function info(request, callback) {
            return this.rpcCall(info, $root.auth.InfoRequest, $root.auth.InfoResponse, request, callback);
        }, "name", { value: "Info" });

        /**
         * Calls Info.
         * @function info
         * @memberof auth.Auth
         * @instance
         * @param {auth.InfoRequest} request InfoRequest message or plain object
         * @returns {Promise<auth.InfoResponse>} Promise
         * @variation 2
         */

        /**
         * Callback as used by {@link auth.Auth#settings}.
         * @memberof auth.Auth
         * @typedef SettingsCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {auth.SettingsResponse} [response] SettingsResponse
         */

        /**
         * Calls Settings.
         * @function settings
         * @memberof auth.Auth
         * @instance
         * @param {auth.SettingsRequest} request SettingsRequest message or plain object
         * @param {auth.Auth.SettingsCallback} callback Node-style callback called with the error, if any, and SettingsResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(Auth.prototype.settings = function settings(request, callback) {
            return this.rpcCall(settings, $root.auth.SettingsRequest, $root.auth.SettingsResponse, request, callback);
        }, "name", { value: "Settings" });

        /**
         * Calls Settings.
         * @function settings
         * @memberof auth.Auth
         * @instance
         * @param {auth.SettingsRequest} request SettingsRequest message or plain object
         * @returns {Promise<auth.SettingsResponse>} Promise
         * @variation 2
         */

        /**
         * Callback as used by {@link auth.Auth#exchange}.
         * @memberof auth.Auth
         * @typedef ExchangeCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {auth.ExchangeResponse} [response] ExchangeResponse
         */

        /**
         * Calls Exchange.
         * @function exchange
         * @memberof auth.Auth
         * @instance
         * @param {auth.ExchangeRequest} request ExchangeRequest message or plain object
         * @param {auth.Auth.ExchangeCallback} callback Node-style callback called with the error, if any, and ExchangeResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(Auth.prototype.exchange = function exchange(request, callback) {
            return this.rpcCall(exchange, $root.auth.ExchangeRequest, $root.auth.ExchangeResponse, request, callback);
        }, "name", { value: "Exchange" });

        /**
         * Calls Exchange.
         * @function exchange
         * @memberof auth.Auth
         * @instance
         * @param {auth.ExchangeRequest} request ExchangeRequest message or plain object
         * @returns {Promise<auth.ExchangeResponse>} Promise
         * @variation 2
         */

        return Auth;
    })();

    return auth;
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

export const changelog = $root.changelog = (() => {

    /**
     * Namespace changelog.
     * @exports changelog
     * @namespace
     */
    const changelog = {};

    changelog.ShowRequest = (function() {

        /**
         * Properties of a ShowRequest.
         * @memberof changelog
         * @interface IShowRequest
         * @property {number|null} [project_id] ShowRequest project_id
         * @property {boolean|null} [only_changed] ShowRequest only_changed
         */

        /**
         * Constructs a new ShowRequest.
         * @memberof changelog
         * @classdesc Represents a ShowRequest.
         * @implements IShowRequest
         * @constructor
         * @param {changelog.IShowRequest=} [properties] Properties to set
         */
        function ShowRequest(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * ShowRequest project_id.
         * @member {number} project_id
         * @memberof changelog.ShowRequest
         * @instance
         */
        ShowRequest.prototype.project_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * ShowRequest only_changed.
         * @member {boolean} only_changed
         * @memberof changelog.ShowRequest
         * @instance
         */
        ShowRequest.prototype.only_changed = false;

        /**
         * Encodes the specified ShowRequest message. Does not implicitly {@link changelog.ShowRequest.verify|verify} messages.
         * @function encode
         * @memberof changelog.ShowRequest
         * @static
         * @param {changelog.ShowRequest} message ShowRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ShowRequest.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.project_id != null && Object.hasOwnProperty.call(message, "project_id"))
                writer.uint32(/* id 1, wireType 0 =*/8).int64(message.project_id);
            if (message.only_changed != null && Object.hasOwnProperty.call(message, "only_changed"))
                writer.uint32(/* id 2, wireType 0 =*/16).bool(message.only_changed);
            return writer;
        };

        /**
         * Decodes a ShowRequest message from the specified reader or buffer.
         * @function decode
         * @memberof changelog.ShowRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {changelog.ShowRequest} ShowRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ShowRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.changelog.ShowRequest();
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

        return ShowRequest;
    })();

    changelog.ShowResponse = (function() {

        /**
         * Properties of a ShowResponse.
         * @memberof changelog
         * @interface IShowResponse
         * @property {Array.<types.ChangelogModel>|null} [items] ShowResponse items
         */

        /**
         * Constructs a new ShowResponse.
         * @memberof changelog
         * @classdesc Represents a ShowResponse.
         * @implements IShowResponse
         * @constructor
         * @param {changelog.IShowResponse=} [properties] Properties to set
         */
        function ShowResponse(properties) {
            this.items = [];
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * ShowResponse items.
         * @member {Array.<types.ChangelogModel>} items
         * @memberof changelog.ShowResponse
         * @instance
         */
        ShowResponse.prototype.items = $util.emptyArray;

        /**
         * Encodes the specified ShowResponse message. Does not implicitly {@link changelog.ShowResponse.verify|verify} messages.
         * @function encode
         * @memberof changelog.ShowResponse
         * @static
         * @param {changelog.ShowResponse} message ShowResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ShowResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.items != null && message.items.length)
                for (let i = 0; i < message.items.length; ++i)
                    $root.types.ChangelogModel.encode(message.items[i], writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
            return writer;
        };

        /**
         * Decodes a ShowResponse message from the specified reader or buffer.
         * @function decode
         * @memberof changelog.ShowResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {changelog.ShowResponse} ShowResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ShowResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.changelog.ShowResponse();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1:
                    if (!(message.items && message.items.length))
                        message.items = [];
                    message.items.push($root.types.ChangelogModel.decode(reader, reader.uint32()));
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        return ShowResponse;
    })();

    changelog.Changelog = (function() {

        /**
         * Constructs a new Changelog service.
         * @memberof changelog
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
         * Callback as used by {@link changelog.Changelog#show}.
         * @memberof changelog.Changelog
         * @typedef ShowCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {changelog.ShowResponse} [response] ShowResponse
         */

        /**
         * Calls Show.
         * @function show
         * @memberof changelog.Changelog
         * @instance
         * @param {changelog.ShowRequest} request ShowRequest message or plain object
         * @param {changelog.Changelog.ShowCallback} callback Node-style callback called with the error, if any, and ShowResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(Changelog.prototype.show = function show(request, callback) {
            return this.rpcCall(show, $root.changelog.ShowRequest, $root.changelog.ShowResponse, request, callback);
        }, "name", { value: "Show" });

        /**
         * Calls Show.
         * @function show
         * @memberof changelog.Changelog
         * @instance
         * @param {changelog.ShowRequest} request ShowRequest message or plain object
         * @returns {Promise<changelog.ShowResponse>} Promise
         * @variation 2
         */

        return Changelog;
    })();

    return changelog;
})();

export const cluster = $root.cluster = (() => {

    /**
     * Namespace cluster.
     * @exports cluster
     * @namespace
     */
    const cluster = {};

    /**
     * Status enum.
     * @name cluster.Status
     * @enum {number}
     * @property {number} StatusUnknown=0 StatusUnknown value
     * @property {number} StatusBad=1 StatusBad value
     * @property {number} StatusNotGood=2 StatusNotGood value
     * @property {number} StatusHealth=3 StatusHealth value
     */
    cluster.Status = (function() {
        const valuesById = {}, values = Object.create(valuesById);
        values[valuesById[0] = "StatusUnknown"] = 0;
        values[valuesById[1] = "StatusBad"] = 1;
        values[valuesById[2] = "StatusNotGood"] = 2;
        values[valuesById[3] = "StatusHealth"] = 3;
        return values;
    })();

    cluster.InfoResponse = (function() {

        /**
         * Properties of an InfoResponse.
         * @memberof cluster
         * @interface IInfoResponse
         * @property {string|null} [status] InfoResponse status
         * @property {string|null} [free_memory] InfoResponse free_memory
         * @property {string|null} [free_cpu] InfoResponse free_cpu
         * @property {string|null} [free_request_memory] InfoResponse free_request_memory
         * @property {string|null} [free_request_cpu] InfoResponse free_request_cpu
         * @property {string|null} [total_memory] InfoResponse total_memory
         * @property {string|null} [total_cpu] InfoResponse total_cpu
         * @property {string|null} [usage_memory_rate] InfoResponse usage_memory_rate
         * @property {string|null} [usage_cpu_rate] InfoResponse usage_cpu_rate
         * @property {string|null} [request_memory_rate] InfoResponse request_memory_rate
         * @property {string|null} [request_cpu_rate] InfoResponse request_cpu_rate
         */

        /**
         * Constructs a new InfoResponse.
         * @memberof cluster
         * @classdesc Represents an InfoResponse.
         * @implements IInfoResponse
         * @constructor
         * @param {cluster.IInfoResponse=} [properties] Properties to set
         */
        function InfoResponse(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * InfoResponse status.
         * @member {string} status
         * @memberof cluster.InfoResponse
         * @instance
         */
        InfoResponse.prototype.status = "";

        /**
         * InfoResponse free_memory.
         * @member {string} free_memory
         * @memberof cluster.InfoResponse
         * @instance
         */
        InfoResponse.prototype.free_memory = "";

        /**
         * InfoResponse free_cpu.
         * @member {string} free_cpu
         * @memberof cluster.InfoResponse
         * @instance
         */
        InfoResponse.prototype.free_cpu = "";

        /**
         * InfoResponse free_request_memory.
         * @member {string} free_request_memory
         * @memberof cluster.InfoResponse
         * @instance
         */
        InfoResponse.prototype.free_request_memory = "";

        /**
         * InfoResponse free_request_cpu.
         * @member {string} free_request_cpu
         * @memberof cluster.InfoResponse
         * @instance
         */
        InfoResponse.prototype.free_request_cpu = "";

        /**
         * InfoResponse total_memory.
         * @member {string} total_memory
         * @memberof cluster.InfoResponse
         * @instance
         */
        InfoResponse.prototype.total_memory = "";

        /**
         * InfoResponse total_cpu.
         * @member {string} total_cpu
         * @memberof cluster.InfoResponse
         * @instance
         */
        InfoResponse.prototype.total_cpu = "";

        /**
         * InfoResponse usage_memory_rate.
         * @member {string} usage_memory_rate
         * @memberof cluster.InfoResponse
         * @instance
         */
        InfoResponse.prototype.usage_memory_rate = "";

        /**
         * InfoResponse usage_cpu_rate.
         * @member {string} usage_cpu_rate
         * @memberof cluster.InfoResponse
         * @instance
         */
        InfoResponse.prototype.usage_cpu_rate = "";

        /**
         * InfoResponse request_memory_rate.
         * @member {string} request_memory_rate
         * @memberof cluster.InfoResponse
         * @instance
         */
        InfoResponse.prototype.request_memory_rate = "";

        /**
         * InfoResponse request_cpu_rate.
         * @member {string} request_cpu_rate
         * @memberof cluster.InfoResponse
         * @instance
         */
        InfoResponse.prototype.request_cpu_rate = "";

        /**
         * Encodes the specified InfoResponse message. Does not implicitly {@link cluster.InfoResponse.verify|verify} messages.
         * @function encode
         * @memberof cluster.InfoResponse
         * @static
         * @param {cluster.InfoResponse} message InfoResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        InfoResponse.encode = function encode(message, writer) {
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
         * Decodes an InfoResponse message from the specified reader or buffer.
         * @function decode
         * @memberof cluster.InfoResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {cluster.InfoResponse} InfoResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        InfoResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.cluster.InfoResponse();
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

        return InfoResponse;
    })();

    cluster.InfoRequest = (function() {

        /**
         * Properties of an InfoRequest.
         * @memberof cluster
         * @interface IInfoRequest
         */

        /**
         * Constructs a new InfoRequest.
         * @memberof cluster
         * @classdesc Represents an InfoRequest.
         * @implements IInfoRequest
         * @constructor
         * @param {cluster.IInfoRequest=} [properties] Properties to set
         */
        function InfoRequest(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * Encodes the specified InfoRequest message. Does not implicitly {@link cluster.InfoRequest.verify|verify} messages.
         * @function encode
         * @memberof cluster.InfoRequest
         * @static
         * @param {cluster.InfoRequest} message InfoRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        InfoRequest.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            return writer;
        };

        /**
         * Decodes an InfoRequest message from the specified reader or buffer.
         * @function decode
         * @memberof cluster.InfoRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {cluster.InfoRequest} InfoRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        InfoRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.cluster.InfoRequest();
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

        return InfoRequest;
    })();

    cluster.Cluster = (function() {

        /**
         * Constructs a new Cluster service.
         * @memberof cluster
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
         * Callback as used by {@link cluster.Cluster#clusterInfo}.
         * @memberof cluster.Cluster
         * @typedef ClusterInfoCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {cluster.InfoResponse} [response] InfoResponse
         */

        /**
         * Calls ClusterInfo.
         * @function clusterInfo
         * @memberof cluster.Cluster
         * @instance
         * @param {cluster.InfoRequest} request InfoRequest message or plain object
         * @param {cluster.Cluster.ClusterInfoCallback} callback Node-style callback called with the error, if any, and InfoResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(Cluster.prototype.clusterInfo = function clusterInfo(request, callback) {
            return this.rpcCall(clusterInfo, $root.cluster.InfoRequest, $root.cluster.InfoResponse, request, callback);
        }, "name", { value: "ClusterInfo" });

        /**
         * Calls ClusterInfo.
         * @function clusterInfo
         * @memberof cluster.Cluster
         * @instance
         * @param {cluster.InfoRequest} request InfoRequest message or plain object
         * @returns {Promise<cluster.InfoResponse>} Promise
         * @variation 2
         */

        return Cluster;
    })();

    return cluster;
})();

export const container = $root.container = (() => {

    /**
     * Namespace container.
     * @exports container
     * @namespace
     */
    const container = {};

    container.CopyToPodRequest = (function() {

        /**
         * Properties of a CopyToPodRequest.
         * @memberof container
         * @interface ICopyToPodRequest
         * @property {number|null} [file_id] CopyToPodRequest file_id
         * @property {string|null} [namespace] CopyToPodRequest namespace
         * @property {string|null} [pod] CopyToPodRequest pod
         * @property {string|null} [container] CopyToPodRequest container
         */

        /**
         * Constructs a new CopyToPodRequest.
         * @memberof container
         * @classdesc Represents a CopyToPodRequest.
         * @implements ICopyToPodRequest
         * @constructor
         * @param {container.ICopyToPodRequest=} [properties] Properties to set
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
         * @memberof container.CopyToPodRequest
         * @instance
         */
        CopyToPodRequest.prototype.file_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * CopyToPodRequest namespace.
         * @member {string} namespace
         * @memberof container.CopyToPodRequest
         * @instance
         */
        CopyToPodRequest.prototype.namespace = "";

        /**
         * CopyToPodRequest pod.
         * @member {string} pod
         * @memberof container.CopyToPodRequest
         * @instance
         */
        CopyToPodRequest.prototype.pod = "";

        /**
         * CopyToPodRequest container.
         * @member {string} container
         * @memberof container.CopyToPodRequest
         * @instance
         */
        CopyToPodRequest.prototype.container = "";

        /**
         * Encodes the specified CopyToPodRequest message. Does not implicitly {@link container.CopyToPodRequest.verify|verify} messages.
         * @function encode
         * @memberof container.CopyToPodRequest
         * @static
         * @param {container.CopyToPodRequest} message CopyToPodRequest message or plain object to encode
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
         * @memberof container.CopyToPodRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {container.CopyToPodRequest} CopyToPodRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        CopyToPodRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.container.CopyToPodRequest();
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

    container.CopyToPodResponse = (function() {

        /**
         * Properties of a CopyToPodResponse.
         * @memberof container
         * @interface ICopyToPodResponse
         * @property {string|null} [pod_file_path] CopyToPodResponse pod_file_path
         * @property {string|null} [output] CopyToPodResponse output
         * @property {string|null} [file_name] CopyToPodResponse file_name
         */

        /**
         * Constructs a new CopyToPodResponse.
         * @memberof container
         * @classdesc Represents a CopyToPodResponse.
         * @implements ICopyToPodResponse
         * @constructor
         * @param {container.ICopyToPodResponse=} [properties] Properties to set
         */
        function CopyToPodResponse(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * CopyToPodResponse pod_file_path.
         * @member {string} pod_file_path
         * @memberof container.CopyToPodResponse
         * @instance
         */
        CopyToPodResponse.prototype.pod_file_path = "";

        /**
         * CopyToPodResponse output.
         * @member {string} output
         * @memberof container.CopyToPodResponse
         * @instance
         */
        CopyToPodResponse.prototype.output = "";

        /**
         * CopyToPodResponse file_name.
         * @member {string} file_name
         * @memberof container.CopyToPodResponse
         * @instance
         */
        CopyToPodResponse.prototype.file_name = "";

        /**
         * Encodes the specified CopyToPodResponse message. Does not implicitly {@link container.CopyToPodResponse.verify|verify} messages.
         * @function encode
         * @memberof container.CopyToPodResponse
         * @static
         * @param {container.CopyToPodResponse} message CopyToPodResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        CopyToPodResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.pod_file_path != null && Object.hasOwnProperty.call(message, "pod_file_path"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.pod_file_path);
            if (message.output != null && Object.hasOwnProperty.call(message, "output"))
                writer.uint32(/* id 2, wireType 2 =*/18).string(message.output);
            if (message.file_name != null && Object.hasOwnProperty.call(message, "file_name"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.file_name);
            return writer;
        };

        /**
         * Decodes a CopyToPodResponse message from the specified reader or buffer.
         * @function decode
         * @memberof container.CopyToPodResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {container.CopyToPodResponse} CopyToPodResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        CopyToPodResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.container.CopyToPodResponse();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1:
                    message.pod_file_path = reader.string();
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

    container.ExecRequest = (function() {

        /**
         * Properties of an ExecRequest.
         * @memberof container
         * @interface IExecRequest
         * @property {string|null} [namespace] ExecRequest namespace
         * @property {string|null} [pod] ExecRequest pod
         * @property {string|null} [container] ExecRequest container
         * @property {Array.<string>|null} [command] ExecRequest command
         */

        /**
         * Constructs a new ExecRequest.
         * @memberof container
         * @classdesc Represents an ExecRequest.
         * @implements IExecRequest
         * @constructor
         * @param {container.IExecRequest=} [properties] Properties to set
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
         * @memberof container.ExecRequest
         * @instance
         */
        ExecRequest.prototype.namespace = "";

        /**
         * ExecRequest pod.
         * @member {string} pod
         * @memberof container.ExecRequest
         * @instance
         */
        ExecRequest.prototype.pod = "";

        /**
         * ExecRequest container.
         * @member {string} container
         * @memberof container.ExecRequest
         * @instance
         */
        ExecRequest.prototype.container = "";

        /**
         * ExecRequest command.
         * @member {Array.<string>} command
         * @memberof container.ExecRequest
         * @instance
         */
        ExecRequest.prototype.command = $util.emptyArray;

        /**
         * Encodes the specified ExecRequest message. Does not implicitly {@link container.ExecRequest.verify|verify} messages.
         * @function encode
         * @memberof container.ExecRequest
         * @static
         * @param {container.ExecRequest} message ExecRequest message or plain object to encode
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
         * @memberof container.ExecRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {container.ExecRequest} ExecRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ExecRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.container.ExecRequest();
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

    container.ExecResponse = (function() {

        /**
         * Properties of an ExecResponse.
         * @memberof container
         * @interface IExecResponse
         * @property {string|null} [message] ExecResponse message
         */

        /**
         * Constructs a new ExecResponse.
         * @memberof container
         * @classdesc Represents an ExecResponse.
         * @implements IExecResponse
         * @constructor
         * @param {container.IExecResponse=} [properties] Properties to set
         */
        function ExecResponse(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * ExecResponse message.
         * @member {string} message
         * @memberof container.ExecResponse
         * @instance
         */
        ExecResponse.prototype.message = "";

        /**
         * Encodes the specified ExecResponse message. Does not implicitly {@link container.ExecResponse.verify|verify} messages.
         * @function encode
         * @memberof container.ExecResponse
         * @static
         * @param {container.ExecResponse} message ExecResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ExecResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.message != null && Object.hasOwnProperty.call(message, "message"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.message);
            return writer;
        };

        /**
         * Decodes an ExecResponse message from the specified reader or buffer.
         * @function decode
         * @memberof container.ExecResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {container.ExecResponse} ExecResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ExecResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.container.ExecResponse();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1:
                    message.message = reader.string();
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

    container.StreamCopyToPodRequest = (function() {

        /**
         * Properties of a StreamCopyToPodRequest.
         * @memberof container
         * @interface IStreamCopyToPodRequest
         * @property {string|null} [file_name] StreamCopyToPodRequest file_name
         * @property {Uint8Array|null} [data] StreamCopyToPodRequest data
         * @property {string|null} [namespace] StreamCopyToPodRequest namespace
         * @property {string|null} [pod] StreamCopyToPodRequest pod
         * @property {string|null} [container] StreamCopyToPodRequest container
         */

        /**
         * Constructs a new StreamCopyToPodRequest.
         * @memberof container
         * @classdesc Represents a StreamCopyToPodRequest.
         * @implements IStreamCopyToPodRequest
         * @constructor
         * @param {container.IStreamCopyToPodRequest=} [properties] Properties to set
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
         * @memberof container.StreamCopyToPodRequest
         * @instance
         */
        StreamCopyToPodRequest.prototype.file_name = "";

        /**
         * StreamCopyToPodRequest data.
         * @member {Uint8Array} data
         * @memberof container.StreamCopyToPodRequest
         * @instance
         */
        StreamCopyToPodRequest.prototype.data = $util.newBuffer([]);

        /**
         * StreamCopyToPodRequest namespace.
         * @member {string} namespace
         * @memberof container.StreamCopyToPodRequest
         * @instance
         */
        StreamCopyToPodRequest.prototype.namespace = "";

        /**
         * StreamCopyToPodRequest pod.
         * @member {string} pod
         * @memberof container.StreamCopyToPodRequest
         * @instance
         */
        StreamCopyToPodRequest.prototype.pod = "";

        /**
         * StreamCopyToPodRequest container.
         * @member {string} container
         * @memberof container.StreamCopyToPodRequest
         * @instance
         */
        StreamCopyToPodRequest.prototype.container = "";

        /**
         * Encodes the specified StreamCopyToPodRequest message. Does not implicitly {@link container.StreamCopyToPodRequest.verify|verify} messages.
         * @function encode
         * @memberof container.StreamCopyToPodRequest
         * @static
         * @param {container.StreamCopyToPodRequest} message StreamCopyToPodRequest message or plain object to encode
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
         * @memberof container.StreamCopyToPodRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {container.StreamCopyToPodRequest} StreamCopyToPodRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        StreamCopyToPodRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.container.StreamCopyToPodRequest();
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

    container.StreamCopyToPodResponse = (function() {

        /**
         * Properties of a StreamCopyToPodResponse.
         * @memberof container
         * @interface IStreamCopyToPodResponse
         * @property {number|null} [size] StreamCopyToPodResponse size
         * @property {string|null} [pod_file_path] StreamCopyToPodResponse pod_file_path
         * @property {string|null} [output] StreamCopyToPodResponse output
         * @property {string|null} [pod] StreamCopyToPodResponse pod
         * @property {string|null} [namespace] StreamCopyToPodResponse namespace
         * @property {string|null} [container] StreamCopyToPodResponse container
         * @property {string|null} [filename] StreamCopyToPodResponse filename
         */

        /**
         * Constructs a new StreamCopyToPodResponse.
         * @memberof container
         * @classdesc Represents a StreamCopyToPodResponse.
         * @implements IStreamCopyToPodResponse
         * @constructor
         * @param {container.IStreamCopyToPodResponse=} [properties] Properties to set
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
         * @memberof container.StreamCopyToPodResponse
         * @instance
         */
        StreamCopyToPodResponse.prototype.size = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * StreamCopyToPodResponse pod_file_path.
         * @member {string} pod_file_path
         * @memberof container.StreamCopyToPodResponse
         * @instance
         */
        StreamCopyToPodResponse.prototype.pod_file_path = "";

        /**
         * StreamCopyToPodResponse output.
         * @member {string} output
         * @memberof container.StreamCopyToPodResponse
         * @instance
         */
        StreamCopyToPodResponse.prototype.output = "";

        /**
         * StreamCopyToPodResponse pod.
         * @member {string} pod
         * @memberof container.StreamCopyToPodResponse
         * @instance
         */
        StreamCopyToPodResponse.prototype.pod = "";

        /**
         * StreamCopyToPodResponse namespace.
         * @member {string} namespace
         * @memberof container.StreamCopyToPodResponse
         * @instance
         */
        StreamCopyToPodResponse.prototype.namespace = "";

        /**
         * StreamCopyToPodResponse container.
         * @member {string} container
         * @memberof container.StreamCopyToPodResponse
         * @instance
         */
        StreamCopyToPodResponse.prototype.container = "";

        /**
         * StreamCopyToPodResponse filename.
         * @member {string} filename
         * @memberof container.StreamCopyToPodResponse
         * @instance
         */
        StreamCopyToPodResponse.prototype.filename = "";

        /**
         * Encodes the specified StreamCopyToPodResponse message. Does not implicitly {@link container.StreamCopyToPodResponse.verify|verify} messages.
         * @function encode
         * @memberof container.StreamCopyToPodResponse
         * @static
         * @param {container.StreamCopyToPodResponse} message StreamCopyToPodResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        StreamCopyToPodResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.size != null && Object.hasOwnProperty.call(message, "size"))
                writer.uint32(/* id 1, wireType 0 =*/8).int64(message.size);
            if (message.pod_file_path != null && Object.hasOwnProperty.call(message, "pod_file_path"))
                writer.uint32(/* id 2, wireType 2 =*/18).string(message.pod_file_path);
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
         * @memberof container.StreamCopyToPodResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {container.StreamCopyToPodResponse} StreamCopyToPodResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        StreamCopyToPodResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.container.StreamCopyToPodResponse();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1:
                    message.size = reader.int64();
                    break;
                case 2:
                    message.pod_file_path = reader.string();
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

    container.IsPodRunningRequest = (function() {

        /**
         * Properties of an IsPodRunningRequest.
         * @memberof container
         * @interface IIsPodRunningRequest
         * @property {string|null} [namespace] IsPodRunningRequest namespace
         * @property {string|null} [pod] IsPodRunningRequest pod
         */

        /**
         * Constructs a new IsPodRunningRequest.
         * @memberof container
         * @classdesc Represents an IsPodRunningRequest.
         * @implements IIsPodRunningRequest
         * @constructor
         * @param {container.IIsPodRunningRequest=} [properties] Properties to set
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
         * @memberof container.IsPodRunningRequest
         * @instance
         */
        IsPodRunningRequest.prototype.namespace = "";

        /**
         * IsPodRunningRequest pod.
         * @member {string} pod
         * @memberof container.IsPodRunningRequest
         * @instance
         */
        IsPodRunningRequest.prototype.pod = "";

        /**
         * Encodes the specified IsPodRunningRequest message. Does not implicitly {@link container.IsPodRunningRequest.verify|verify} messages.
         * @function encode
         * @memberof container.IsPodRunningRequest
         * @static
         * @param {container.IsPodRunningRequest} message IsPodRunningRequest message or plain object to encode
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
         * @memberof container.IsPodRunningRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {container.IsPodRunningRequest} IsPodRunningRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        IsPodRunningRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.container.IsPodRunningRequest();
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

    container.IsPodRunningResponse = (function() {

        /**
         * Properties of an IsPodRunningResponse.
         * @memberof container
         * @interface IIsPodRunningResponse
         * @property {boolean|null} [running] IsPodRunningResponse running
         * @property {string|null} [reason] IsPodRunningResponse reason
         */

        /**
         * Constructs a new IsPodRunningResponse.
         * @memberof container
         * @classdesc Represents an IsPodRunningResponse.
         * @implements IIsPodRunningResponse
         * @constructor
         * @param {container.IIsPodRunningResponse=} [properties] Properties to set
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
         * @memberof container.IsPodRunningResponse
         * @instance
         */
        IsPodRunningResponse.prototype.running = false;

        /**
         * IsPodRunningResponse reason.
         * @member {string} reason
         * @memberof container.IsPodRunningResponse
         * @instance
         */
        IsPodRunningResponse.prototype.reason = "";

        /**
         * Encodes the specified IsPodRunningResponse message. Does not implicitly {@link container.IsPodRunningResponse.verify|verify} messages.
         * @function encode
         * @memberof container.IsPodRunningResponse
         * @static
         * @param {container.IsPodRunningResponse} message IsPodRunningResponse message or plain object to encode
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
         * @memberof container.IsPodRunningResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {container.IsPodRunningResponse} IsPodRunningResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        IsPodRunningResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.container.IsPodRunningResponse();
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

    container.IsPodExistsRequest = (function() {

        /**
         * Properties of an IsPodExistsRequest.
         * @memberof container
         * @interface IIsPodExistsRequest
         * @property {string|null} [namespace] IsPodExistsRequest namespace
         * @property {string|null} [pod] IsPodExistsRequest pod
         */

        /**
         * Constructs a new IsPodExistsRequest.
         * @memberof container
         * @classdesc Represents an IsPodExistsRequest.
         * @implements IIsPodExistsRequest
         * @constructor
         * @param {container.IIsPodExistsRequest=} [properties] Properties to set
         */
        function IsPodExistsRequest(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * IsPodExistsRequest namespace.
         * @member {string} namespace
         * @memberof container.IsPodExistsRequest
         * @instance
         */
        IsPodExistsRequest.prototype.namespace = "";

        /**
         * IsPodExistsRequest pod.
         * @member {string} pod
         * @memberof container.IsPodExistsRequest
         * @instance
         */
        IsPodExistsRequest.prototype.pod = "";

        /**
         * Encodes the specified IsPodExistsRequest message. Does not implicitly {@link container.IsPodExistsRequest.verify|verify} messages.
         * @function encode
         * @memberof container.IsPodExistsRequest
         * @static
         * @param {container.IsPodExistsRequest} message IsPodExistsRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        IsPodExistsRequest.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.namespace != null && Object.hasOwnProperty.call(message, "namespace"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.namespace);
            if (message.pod != null && Object.hasOwnProperty.call(message, "pod"))
                writer.uint32(/* id 2, wireType 2 =*/18).string(message.pod);
            return writer;
        };

        /**
         * Decodes an IsPodExistsRequest message from the specified reader or buffer.
         * @function decode
         * @memberof container.IsPodExistsRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {container.IsPodExistsRequest} IsPodExistsRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        IsPodExistsRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.container.IsPodExistsRequest();
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

        return IsPodExistsRequest;
    })();

    container.IsPodExistsResponse = (function() {

        /**
         * Properties of an IsPodExistsResponse.
         * @memberof container
         * @interface IIsPodExistsResponse
         * @property {boolean|null} [exists] IsPodExistsResponse exists
         */

        /**
         * Constructs a new IsPodExistsResponse.
         * @memberof container
         * @classdesc Represents an IsPodExistsResponse.
         * @implements IIsPodExistsResponse
         * @constructor
         * @param {container.IIsPodExistsResponse=} [properties] Properties to set
         */
        function IsPodExistsResponse(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * IsPodExistsResponse exists.
         * @member {boolean} exists
         * @memberof container.IsPodExistsResponse
         * @instance
         */
        IsPodExistsResponse.prototype.exists = false;

        /**
         * Encodes the specified IsPodExistsResponse message. Does not implicitly {@link container.IsPodExistsResponse.verify|verify} messages.
         * @function encode
         * @memberof container.IsPodExistsResponse
         * @static
         * @param {container.IsPodExistsResponse} message IsPodExistsResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        IsPodExistsResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.exists != null && Object.hasOwnProperty.call(message, "exists"))
                writer.uint32(/* id 1, wireType 0 =*/8).bool(message.exists);
            return writer;
        };

        /**
         * Decodes an IsPodExistsResponse message from the specified reader or buffer.
         * @function decode
         * @memberof container.IsPodExistsResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {container.IsPodExistsResponse} IsPodExistsResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        IsPodExistsResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.container.IsPodExistsResponse();
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

        return IsPodExistsResponse;
    })();

    container.LogRequest = (function() {

        /**
         * Properties of a LogRequest.
         * @memberof container
         * @interface ILogRequest
         * @property {string|null} [namespace] LogRequest namespace
         * @property {string|null} [pod] LogRequest pod
         * @property {string|null} [container] LogRequest container
         */

        /**
         * Constructs a new LogRequest.
         * @memberof container
         * @classdesc Represents a LogRequest.
         * @implements ILogRequest
         * @constructor
         * @param {container.ILogRequest=} [properties] Properties to set
         */
        function LogRequest(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * LogRequest namespace.
         * @member {string} namespace
         * @memberof container.LogRequest
         * @instance
         */
        LogRequest.prototype.namespace = "";

        /**
         * LogRequest pod.
         * @member {string} pod
         * @memberof container.LogRequest
         * @instance
         */
        LogRequest.prototype.pod = "";

        /**
         * LogRequest container.
         * @member {string} container
         * @memberof container.LogRequest
         * @instance
         */
        LogRequest.prototype.container = "";

        /**
         * Encodes the specified LogRequest message. Does not implicitly {@link container.LogRequest.verify|verify} messages.
         * @function encode
         * @memberof container.LogRequest
         * @static
         * @param {container.LogRequest} message LogRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        LogRequest.encode = function encode(message, writer) {
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
         * Decodes a LogRequest message from the specified reader or buffer.
         * @function decode
         * @memberof container.LogRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {container.LogRequest} LogRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        LogRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.container.LogRequest();
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

        return LogRequest;
    })();

    container.LogResponse = (function() {

        /**
         * Properties of a LogResponse.
         * @memberof container
         * @interface ILogResponse
         * @property {string|null} [namespace] LogResponse namespace
         * @property {string|null} [pod_name] LogResponse pod_name
         * @property {string|null} [container_name] LogResponse container_name
         * @property {string|null} [log] LogResponse log
         */

        /**
         * Constructs a new LogResponse.
         * @memberof container
         * @classdesc Represents a LogResponse.
         * @implements ILogResponse
         * @constructor
         * @param {container.ILogResponse=} [properties] Properties to set
         */
        function LogResponse(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * LogResponse namespace.
         * @member {string} namespace
         * @memberof container.LogResponse
         * @instance
         */
        LogResponse.prototype.namespace = "";

        /**
         * LogResponse pod_name.
         * @member {string} pod_name
         * @memberof container.LogResponse
         * @instance
         */
        LogResponse.prototype.pod_name = "";

        /**
         * LogResponse container_name.
         * @member {string} container_name
         * @memberof container.LogResponse
         * @instance
         */
        LogResponse.prototype.container_name = "";

        /**
         * LogResponse log.
         * @member {string} log
         * @memberof container.LogResponse
         * @instance
         */
        LogResponse.prototype.log = "";

        /**
         * Encodes the specified LogResponse message. Does not implicitly {@link container.LogResponse.verify|verify} messages.
         * @function encode
         * @memberof container.LogResponse
         * @static
         * @param {container.LogResponse} message LogResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        LogResponse.encode = function encode(message, writer) {
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
         * Decodes a LogResponse message from the specified reader or buffer.
         * @function decode
         * @memberof container.LogResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {container.LogResponse} LogResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        LogResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.container.LogResponse();
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

        return LogResponse;
    })();

    container.Container = (function() {

        /**
         * Constructs a new Container service.
         * @memberof container
         * @classdesc Represents a Container
         * @extends $protobuf.rpc.Service
         * @constructor
         * @param {$protobuf.RPCImpl} rpcImpl RPC implementation
         * @param {boolean} [requestDelimited=false] Whether requests are length-delimited
         * @param {boolean} [responseDelimited=false] Whether responses are length-delimited
         */
        function Container(rpcImpl, requestDelimited, responseDelimited) {
            $protobuf.rpc.Service.call(this, rpcImpl, requestDelimited, responseDelimited);
        }

        (Container.prototype = Object.create($protobuf.rpc.Service.prototype)).constructor = Container;

        /**
         * Callback as used by {@link container.Container#copyToPod}.
         * @memberof container.Container
         * @typedef CopyToPodCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {container.CopyToPodResponse} [response] CopyToPodResponse
         */

        /**
         * Calls CopyToPod.
         * @function copyToPod
         * @memberof container.Container
         * @instance
         * @param {container.CopyToPodRequest} request CopyToPodRequest message or plain object
         * @param {container.Container.CopyToPodCallback} callback Node-style callback called with the error, if any, and CopyToPodResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(Container.prototype.copyToPod = function copyToPod(request, callback) {
            return this.rpcCall(copyToPod, $root.container.CopyToPodRequest, $root.container.CopyToPodResponse, request, callback);
        }, "name", { value: "CopyToPod" });

        /**
         * Calls CopyToPod.
         * @function copyToPod
         * @memberof container.Container
         * @instance
         * @param {container.CopyToPodRequest} request CopyToPodRequest message or plain object
         * @returns {Promise<container.CopyToPodResponse>} Promise
         * @variation 2
         */

        /**
         * Callback as used by {@link container.Container#exec}.
         * @memberof container.Container
         * @typedef ExecCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {container.ExecResponse} [response] ExecResponse
         */

        /**
         * Calls Exec.
         * @function exec
         * @memberof container.Container
         * @instance
         * @param {container.ExecRequest} request ExecRequest message or plain object
         * @param {container.Container.ExecCallback} callback Node-style callback called with the error, if any, and ExecResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(Container.prototype.exec = function exec(request, callback) {
            return this.rpcCall(exec, $root.container.ExecRequest, $root.container.ExecResponse, request, callback);
        }, "name", { value: "Exec" });

        /**
         * Calls Exec.
         * @function exec
         * @memberof container.Container
         * @instance
         * @param {container.ExecRequest} request ExecRequest message or plain object
         * @returns {Promise<container.ExecResponse>} Promise
         * @variation 2
         */

        /**
         * Callback as used by {@link container.Container#streamCopyToPod}.
         * @memberof container.Container
         * @typedef StreamCopyToPodCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {container.StreamCopyToPodResponse} [response] StreamCopyToPodResponse
         */

        /**
         * Calls StreamCopyToPod.
         * @function streamCopyToPod
         * @memberof container.Container
         * @instance
         * @param {container.StreamCopyToPodRequest} request StreamCopyToPodRequest message or plain object
         * @param {container.Container.StreamCopyToPodCallback} callback Node-style callback called with the error, if any, and StreamCopyToPodResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(Container.prototype.streamCopyToPod = function streamCopyToPod(request, callback) {
            return this.rpcCall(streamCopyToPod, $root.container.StreamCopyToPodRequest, $root.container.StreamCopyToPodResponse, request, callback);
        }, "name", { value: "StreamCopyToPod" });

        /**
         * Calls StreamCopyToPod.
         * @function streamCopyToPod
         * @memberof container.Container
         * @instance
         * @param {container.StreamCopyToPodRequest} request StreamCopyToPodRequest message or plain object
         * @returns {Promise<container.StreamCopyToPodResponse>} Promise
         * @variation 2
         */

        /**
         * Callback as used by {@link container.Container#isPodRunning}.
         * @memberof container.Container
         * @typedef IsPodRunningCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {container.IsPodRunningResponse} [response] IsPodRunningResponse
         */

        /**
         * Calls IsPodRunning.
         * @function isPodRunning
         * @memberof container.Container
         * @instance
         * @param {container.IsPodRunningRequest} request IsPodRunningRequest message or plain object
         * @param {container.Container.IsPodRunningCallback} callback Node-style callback called with the error, if any, and IsPodRunningResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(Container.prototype.isPodRunning = function isPodRunning(request, callback) {
            return this.rpcCall(isPodRunning, $root.container.IsPodRunningRequest, $root.container.IsPodRunningResponse, request, callback);
        }, "name", { value: "IsPodRunning" });

        /**
         * Calls IsPodRunning.
         * @function isPodRunning
         * @memberof container.Container
         * @instance
         * @param {container.IsPodRunningRequest} request IsPodRunningRequest message or plain object
         * @returns {Promise<container.IsPodRunningResponse>} Promise
         * @variation 2
         */

        /**
         * Callback as used by {@link container.Container#isPodExists}.
         * @memberof container.Container
         * @typedef IsPodExistsCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {container.IsPodExistsResponse} [response] IsPodExistsResponse
         */

        /**
         * Calls IsPodExists.
         * @function isPodExists
         * @memberof container.Container
         * @instance
         * @param {container.IsPodExistsRequest} request IsPodExistsRequest message or plain object
         * @param {container.Container.IsPodExistsCallback} callback Node-style callback called with the error, if any, and IsPodExistsResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(Container.prototype.isPodExists = function isPodExists(request, callback) {
            return this.rpcCall(isPodExists, $root.container.IsPodExistsRequest, $root.container.IsPodExistsResponse, request, callback);
        }, "name", { value: "IsPodExists" });

        /**
         * Calls IsPodExists.
         * @function isPodExists
         * @memberof container.Container
         * @instance
         * @param {container.IsPodExistsRequest} request IsPodExistsRequest message or plain object
         * @returns {Promise<container.IsPodExistsResponse>} Promise
         * @variation 2
         */

        /**
         * Callback as used by {@link container.Container#containerLog}.
         * @memberof container.Container
         * @typedef ContainerLogCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {container.LogResponse} [response] LogResponse
         */

        /**
         * Calls ContainerLog.
         * @function containerLog
         * @memberof container.Container
         * @instance
         * @param {container.LogRequest} request LogRequest message or plain object
         * @param {container.Container.ContainerLogCallback} callback Node-style callback called with the error, if any, and LogResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(Container.prototype.containerLog = function containerLog(request, callback) {
            return this.rpcCall(containerLog, $root.container.LogRequest, $root.container.LogResponse, request, callback);
        }, "name", { value: "ContainerLog" });

        /**
         * Calls ContainerLog.
         * @function containerLog
         * @memberof container.Container
         * @instance
         * @param {container.LogRequest} request LogRequest message or plain object
         * @returns {Promise<container.LogResponse>} Promise
         * @variation 2
         */

        /**
         * Callback as used by {@link container.Container#streamContainerLog}.
         * @memberof container.Container
         * @typedef StreamContainerLogCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {container.LogResponse} [response] LogResponse
         */

        /**
         * Calls StreamContainerLog.
         * @function streamContainerLog
         * @memberof container.Container
         * @instance
         * @param {container.LogRequest} request LogRequest message or plain object
         * @param {container.Container.StreamContainerLogCallback} callback Node-style callback called with the error, if any, and LogResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(Container.prototype.streamContainerLog = function streamContainerLog(request, callback) {
            return this.rpcCall(streamContainerLog, $root.container.LogRequest, $root.container.LogResponse, request, callback);
        }, "name", { value: "StreamContainerLog" });

        /**
         * Calls StreamContainerLog.
         * @function streamContainerLog
         * @memberof container.Container
         * @instance
         * @param {container.LogRequest} request LogRequest message or plain object
         * @returns {Promise<container.LogResponse>} Promise
         * @variation 2
         */

        return Container;
    })();

    return container;
})();

export const endpoint = $root.endpoint = (() => {

    /**
     * Namespace endpoint.
     * @exports endpoint
     * @namespace
     */
    const endpoint = {};

    endpoint.InNamespaceRequest = (function() {

        /**
         * Properties of an InNamespaceRequest.
         * @memberof endpoint
         * @interface IInNamespaceRequest
         * @property {number|null} [namespace_id] InNamespaceRequest namespace_id
         */

        /**
         * Constructs a new InNamespaceRequest.
         * @memberof endpoint
         * @classdesc Represents an InNamespaceRequest.
         * @implements IInNamespaceRequest
         * @constructor
         * @param {endpoint.IInNamespaceRequest=} [properties] Properties to set
         */
        function InNamespaceRequest(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * InNamespaceRequest namespace_id.
         * @member {number} namespace_id
         * @memberof endpoint.InNamespaceRequest
         * @instance
         */
        InNamespaceRequest.prototype.namespace_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Encodes the specified InNamespaceRequest message. Does not implicitly {@link endpoint.InNamespaceRequest.verify|verify} messages.
         * @function encode
         * @memberof endpoint.InNamespaceRequest
         * @static
         * @param {endpoint.InNamespaceRequest} message InNamespaceRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        InNamespaceRequest.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.namespace_id != null && Object.hasOwnProperty.call(message, "namespace_id"))
                writer.uint32(/* id 1, wireType 0 =*/8).int64(message.namespace_id);
            return writer;
        };

        /**
         * Decodes an InNamespaceRequest message from the specified reader or buffer.
         * @function decode
         * @memberof endpoint.InNamespaceRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {endpoint.InNamespaceRequest} InNamespaceRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        InNamespaceRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.endpoint.InNamespaceRequest();
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

        return InNamespaceRequest;
    })();

    endpoint.InNamespaceResponse = (function() {

        /**
         * Properties of an InNamespaceResponse.
         * @memberof endpoint
         * @interface IInNamespaceResponse
         * @property {Array.<types.ServiceEndpoint>|null} [items] InNamespaceResponse items
         */

        /**
         * Constructs a new InNamespaceResponse.
         * @memberof endpoint
         * @classdesc Represents an InNamespaceResponse.
         * @implements IInNamespaceResponse
         * @constructor
         * @param {endpoint.IInNamespaceResponse=} [properties] Properties to set
         */
        function InNamespaceResponse(properties) {
            this.items = [];
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * InNamespaceResponse items.
         * @member {Array.<types.ServiceEndpoint>} items
         * @memberof endpoint.InNamespaceResponse
         * @instance
         */
        InNamespaceResponse.prototype.items = $util.emptyArray;

        /**
         * Encodes the specified InNamespaceResponse message. Does not implicitly {@link endpoint.InNamespaceResponse.verify|verify} messages.
         * @function encode
         * @memberof endpoint.InNamespaceResponse
         * @static
         * @param {endpoint.InNamespaceResponse} message InNamespaceResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        InNamespaceResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.items != null && message.items.length)
                for (let i = 0; i < message.items.length; ++i)
                    $root.types.ServiceEndpoint.encode(message.items[i], writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
            return writer;
        };

        /**
         * Decodes an InNamespaceResponse message from the specified reader or buffer.
         * @function decode
         * @memberof endpoint.InNamespaceResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {endpoint.InNamespaceResponse} InNamespaceResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        InNamespaceResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.endpoint.InNamespaceResponse();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1:
                    if (!(message.items && message.items.length))
                        message.items = [];
                    message.items.push($root.types.ServiceEndpoint.decode(reader, reader.uint32()));
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        return InNamespaceResponse;
    })();

    endpoint.InProjectRequest = (function() {

        /**
         * Properties of an InProjectRequest.
         * @memberof endpoint
         * @interface IInProjectRequest
         * @property {number|null} [project_id] InProjectRequest project_id
         */

        /**
         * Constructs a new InProjectRequest.
         * @memberof endpoint
         * @classdesc Represents an InProjectRequest.
         * @implements IInProjectRequest
         * @constructor
         * @param {endpoint.IInProjectRequest=} [properties] Properties to set
         */
        function InProjectRequest(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * InProjectRequest project_id.
         * @member {number} project_id
         * @memberof endpoint.InProjectRequest
         * @instance
         */
        InProjectRequest.prototype.project_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Encodes the specified InProjectRequest message. Does not implicitly {@link endpoint.InProjectRequest.verify|verify} messages.
         * @function encode
         * @memberof endpoint.InProjectRequest
         * @static
         * @param {endpoint.InProjectRequest} message InProjectRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        InProjectRequest.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.project_id != null && Object.hasOwnProperty.call(message, "project_id"))
                writer.uint32(/* id 1, wireType 0 =*/8).int64(message.project_id);
            return writer;
        };

        /**
         * Decodes an InProjectRequest message from the specified reader or buffer.
         * @function decode
         * @memberof endpoint.InProjectRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {endpoint.InProjectRequest} InProjectRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        InProjectRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.endpoint.InProjectRequest();
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

        return InProjectRequest;
    })();

    endpoint.InProjectResponse = (function() {

        /**
         * Properties of an InProjectResponse.
         * @memberof endpoint
         * @interface IInProjectResponse
         * @property {Array.<types.ServiceEndpoint>|null} [items] InProjectResponse items
         */

        /**
         * Constructs a new InProjectResponse.
         * @memberof endpoint
         * @classdesc Represents an InProjectResponse.
         * @implements IInProjectResponse
         * @constructor
         * @param {endpoint.IInProjectResponse=} [properties] Properties to set
         */
        function InProjectResponse(properties) {
            this.items = [];
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * InProjectResponse items.
         * @member {Array.<types.ServiceEndpoint>} items
         * @memberof endpoint.InProjectResponse
         * @instance
         */
        InProjectResponse.prototype.items = $util.emptyArray;

        /**
         * Encodes the specified InProjectResponse message. Does not implicitly {@link endpoint.InProjectResponse.verify|verify} messages.
         * @function encode
         * @memberof endpoint.InProjectResponse
         * @static
         * @param {endpoint.InProjectResponse} message InProjectResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        InProjectResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.items != null && message.items.length)
                for (let i = 0; i < message.items.length; ++i)
                    $root.types.ServiceEndpoint.encode(message.items[i], writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
            return writer;
        };

        /**
         * Decodes an InProjectResponse message from the specified reader or buffer.
         * @function decode
         * @memberof endpoint.InProjectResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {endpoint.InProjectResponse} InProjectResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        InProjectResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.endpoint.InProjectResponse();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1:
                    if (!(message.items && message.items.length))
                        message.items = [];
                    message.items.push($root.types.ServiceEndpoint.decode(reader, reader.uint32()));
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        return InProjectResponse;
    })();

    endpoint.Endpoint = (function() {

        /**
         * Constructs a new Endpoint service.
         * @memberof endpoint
         * @classdesc Represents an Endpoint
         * @extends $protobuf.rpc.Service
         * @constructor
         * @param {$protobuf.RPCImpl} rpcImpl RPC implementation
         * @param {boolean} [requestDelimited=false] Whether requests are length-delimited
         * @param {boolean} [responseDelimited=false] Whether responses are length-delimited
         */
        function Endpoint(rpcImpl, requestDelimited, responseDelimited) {
            $protobuf.rpc.Service.call(this, rpcImpl, requestDelimited, responseDelimited);
        }

        (Endpoint.prototype = Object.create($protobuf.rpc.Service.prototype)).constructor = Endpoint;

        /**
         * Callback as used by {@link endpoint.Endpoint#inNamespace}.
         * @memberof endpoint.Endpoint
         * @typedef InNamespaceCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {endpoint.InNamespaceResponse} [response] InNamespaceResponse
         */

        /**
         * Calls InNamespace.
         * @function inNamespace
         * @memberof endpoint.Endpoint
         * @instance
         * @param {endpoint.InNamespaceRequest} request InNamespaceRequest message or plain object
         * @param {endpoint.Endpoint.InNamespaceCallback} callback Node-style callback called with the error, if any, and InNamespaceResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(Endpoint.prototype.inNamespace = function inNamespace(request, callback) {
            return this.rpcCall(inNamespace, $root.endpoint.InNamespaceRequest, $root.endpoint.InNamespaceResponse, request, callback);
        }, "name", { value: "InNamespace" });

        /**
         * Calls InNamespace.
         * @function inNamespace
         * @memberof endpoint.Endpoint
         * @instance
         * @param {endpoint.InNamespaceRequest} request InNamespaceRequest message or plain object
         * @returns {Promise<endpoint.InNamespaceResponse>} Promise
         * @variation 2
         */

        /**
         * Callback as used by {@link endpoint.Endpoint#inProject}.
         * @memberof endpoint.Endpoint
         * @typedef InProjectCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {endpoint.InProjectResponse} [response] InProjectResponse
         */

        /**
         * Calls InProject.
         * @function inProject
         * @memberof endpoint.Endpoint
         * @instance
         * @param {endpoint.InProjectRequest} request InProjectRequest message or plain object
         * @param {endpoint.Endpoint.InProjectCallback} callback Node-style callback called with the error, if any, and InProjectResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(Endpoint.prototype.inProject = function inProject(request, callback) {
            return this.rpcCall(inProject, $root.endpoint.InProjectRequest, $root.endpoint.InProjectResponse, request, callback);
        }, "name", { value: "InProject" });

        /**
         * Calls InProject.
         * @function inProject
         * @memberof endpoint.Endpoint
         * @instance
         * @param {endpoint.InProjectRequest} request InProjectRequest message or plain object
         * @returns {Promise<endpoint.InProjectResponse>} Promise
         * @variation 2
         */

        return Endpoint;
    })();

    return endpoint;
})();

export const event = $root.event = (() => {

    /**
     * Namespace event.
     * @exports event
     * @namespace
     */
    const event = {};

    event.ListRequest = (function() {

        /**
         * Properties of a ListRequest.
         * @memberof event
         * @interface IListRequest
         * @property {number|null} [page] ListRequest page
         * @property {number|null} [page_size] ListRequest page_size
         */

        /**
         * Constructs a new ListRequest.
         * @memberof event
         * @classdesc Represents a ListRequest.
         * @implements IListRequest
         * @constructor
         * @param {event.IListRequest=} [properties] Properties to set
         */
        function ListRequest(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * ListRequest page.
         * @member {number} page
         * @memberof event.ListRequest
         * @instance
         */
        ListRequest.prototype.page = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * ListRequest page_size.
         * @member {number} page_size
         * @memberof event.ListRequest
         * @instance
         */
        ListRequest.prototype.page_size = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Encodes the specified ListRequest message. Does not implicitly {@link event.ListRequest.verify|verify} messages.
         * @function encode
         * @memberof event.ListRequest
         * @static
         * @param {event.ListRequest} message ListRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ListRequest.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.page != null && Object.hasOwnProperty.call(message, "page"))
                writer.uint32(/* id 1, wireType 0 =*/8).int64(message.page);
            if (message.page_size != null && Object.hasOwnProperty.call(message, "page_size"))
                writer.uint32(/* id 2, wireType 0 =*/16).int64(message.page_size);
            return writer;
        };

        /**
         * Decodes a ListRequest message from the specified reader or buffer.
         * @function decode
         * @memberof event.ListRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {event.ListRequest} ListRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ListRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.event.ListRequest();
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

        return ListRequest;
    })();

    event.ListResponse = (function() {

        /**
         * Properties of a ListResponse.
         * @memberof event
         * @interface IListResponse
         * @property {number|null} [page] ListResponse page
         * @property {number|null} [page_size] ListResponse page_size
         * @property {Array.<types.EventModel>|null} [items] ListResponse items
         * @property {number|null} [count] ListResponse count
         */

        /**
         * Constructs a new ListResponse.
         * @memberof event
         * @classdesc Represents a ListResponse.
         * @implements IListResponse
         * @constructor
         * @param {event.IListResponse=} [properties] Properties to set
         */
        function ListResponse(properties) {
            this.items = [];
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * ListResponse page.
         * @member {number} page
         * @memberof event.ListResponse
         * @instance
         */
        ListResponse.prototype.page = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * ListResponse page_size.
         * @member {number} page_size
         * @memberof event.ListResponse
         * @instance
         */
        ListResponse.prototype.page_size = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * ListResponse items.
         * @member {Array.<types.EventModel>} items
         * @memberof event.ListResponse
         * @instance
         */
        ListResponse.prototype.items = $util.emptyArray;

        /**
         * ListResponse count.
         * @member {number} count
         * @memberof event.ListResponse
         * @instance
         */
        ListResponse.prototype.count = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Encodes the specified ListResponse message. Does not implicitly {@link event.ListResponse.verify|verify} messages.
         * @function encode
         * @memberof event.ListResponse
         * @static
         * @param {event.ListResponse} message ListResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ListResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.page != null && Object.hasOwnProperty.call(message, "page"))
                writer.uint32(/* id 1, wireType 0 =*/8).int64(message.page);
            if (message.page_size != null && Object.hasOwnProperty.call(message, "page_size"))
                writer.uint32(/* id 2, wireType 0 =*/16).int64(message.page_size);
            if (message.items != null && message.items.length)
                for (let i = 0; i < message.items.length; ++i)
                    $root.types.EventModel.encode(message.items[i], writer.uint32(/* id 3, wireType 2 =*/26).fork()).ldelim();
            if (message.count != null && Object.hasOwnProperty.call(message, "count"))
                writer.uint32(/* id 4, wireType 0 =*/32).int64(message.count);
            return writer;
        };

        /**
         * Decodes a ListResponse message from the specified reader or buffer.
         * @function decode
         * @memberof event.ListResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {event.ListResponse} ListResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ListResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.event.ListResponse();
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
                    message.items.push($root.types.EventModel.decode(reader, reader.uint32()));
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

        return ListResponse;
    })();

    event.Event = (function() {

        /**
         * Constructs a new Event service.
         * @memberof event
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
         * Callback as used by {@link event.Event#list}.
         * @memberof event.Event
         * @typedef ListCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {event.ListResponse} [response] ListResponse
         */

        /**
         * Calls List.
         * @function list
         * @memberof event.Event
         * @instance
         * @param {event.ListRequest} request ListRequest message or plain object
         * @param {event.Event.ListCallback} callback Node-style callback called with the error, if any, and ListResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(Event.prototype.list = function list(request, callback) {
            return this.rpcCall(list, $root.event.ListRequest, $root.event.ListResponse, request, callback);
        }, "name", { value: "List" });

        /**
         * Calls List.
         * @function list
         * @memberof event.Event
         * @instance
         * @param {event.ListRequest} request ListRequest message or plain object
         * @returns {Promise<event.ListResponse>} Promise
         * @variation 2
         */

        return Event;
    })();

    return event;
})();

export const file = $root.file = (() => {

    /**
     * Namespace file.
     * @exports file
     * @namespace
     */
    const file = {};

    file.DeleteRequest = (function() {

        /**
         * Properties of a DeleteRequest.
         * @memberof file
         * @interface IDeleteRequest
         * @property {number|null} [id] DeleteRequest id
         */

        /**
         * Constructs a new DeleteRequest.
         * @memberof file
         * @classdesc Represents a DeleteRequest.
         * @implements IDeleteRequest
         * @constructor
         * @param {file.IDeleteRequest=} [properties] Properties to set
         */
        function DeleteRequest(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * DeleteRequest id.
         * @member {number} id
         * @memberof file.DeleteRequest
         * @instance
         */
        DeleteRequest.prototype.id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Encodes the specified DeleteRequest message. Does not implicitly {@link file.DeleteRequest.verify|verify} messages.
         * @function encode
         * @memberof file.DeleteRequest
         * @static
         * @param {file.DeleteRequest} message DeleteRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        DeleteRequest.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.id != null && Object.hasOwnProperty.call(message, "id"))
                writer.uint32(/* id 1, wireType 0 =*/8).int64(message.id);
            return writer;
        };

        /**
         * Decodes a DeleteRequest message from the specified reader or buffer.
         * @function decode
         * @memberof file.DeleteRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {file.DeleteRequest} DeleteRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        DeleteRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.file.DeleteRequest();
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

        return DeleteRequest;
    })();

    file.DeleteResponse = (function() {

        /**
         * Properties of a DeleteResponse.
         * @memberof file
         * @interface IDeleteResponse
         * @property {types.FileModel|null} [file] DeleteResponse file
         */

        /**
         * Constructs a new DeleteResponse.
         * @memberof file
         * @classdesc Represents a DeleteResponse.
         * @implements IDeleteResponse
         * @constructor
         * @param {file.IDeleteResponse=} [properties] Properties to set
         */
        function DeleteResponse(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * DeleteResponse file.
         * @member {types.FileModel|null|undefined} file
         * @memberof file.DeleteResponse
         * @instance
         */
        DeleteResponse.prototype.file = null;

        /**
         * Encodes the specified DeleteResponse message. Does not implicitly {@link file.DeleteResponse.verify|verify} messages.
         * @function encode
         * @memberof file.DeleteResponse
         * @static
         * @param {file.DeleteResponse} message DeleteResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        DeleteResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.file != null && Object.hasOwnProperty.call(message, "file"))
                $root.types.FileModel.encode(message.file, writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
            return writer;
        };

        /**
         * Decodes a DeleteResponse message from the specified reader or buffer.
         * @function decode
         * @memberof file.DeleteResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {file.DeleteResponse} DeleteResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        DeleteResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.file.DeleteResponse();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1:
                    message.file = $root.types.FileModel.decode(reader, reader.uint32());
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        return DeleteResponse;
    })();

    file.DeleteUndocumentedFilesRequest = (function() {

        /**
         * Properties of a DeleteUndocumentedFilesRequest.
         * @memberof file
         * @interface IDeleteUndocumentedFilesRequest
         */

        /**
         * Constructs a new DeleteUndocumentedFilesRequest.
         * @memberof file
         * @classdesc Represents a DeleteUndocumentedFilesRequest.
         * @implements IDeleteUndocumentedFilesRequest
         * @constructor
         * @param {file.IDeleteUndocumentedFilesRequest=} [properties] Properties to set
         */
        function DeleteUndocumentedFilesRequest(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * Encodes the specified DeleteUndocumentedFilesRequest message. Does not implicitly {@link file.DeleteUndocumentedFilesRequest.verify|verify} messages.
         * @function encode
         * @memberof file.DeleteUndocumentedFilesRequest
         * @static
         * @param {file.DeleteUndocumentedFilesRequest} message DeleteUndocumentedFilesRequest message or plain object to encode
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
         * @memberof file.DeleteUndocumentedFilesRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {file.DeleteUndocumentedFilesRequest} DeleteUndocumentedFilesRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        DeleteUndocumentedFilesRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.file.DeleteUndocumentedFilesRequest();
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

    file.DeleteUndocumentedFilesResponse = (function() {

        /**
         * Properties of a DeleteUndocumentedFilesResponse.
         * @memberof file
         * @interface IDeleteUndocumentedFilesResponse
         * @property {Array.<types.FileModel>|null} [items] DeleteUndocumentedFilesResponse items
         */

        /**
         * Constructs a new DeleteUndocumentedFilesResponse.
         * @memberof file
         * @classdesc Represents a DeleteUndocumentedFilesResponse.
         * @implements IDeleteUndocumentedFilesResponse
         * @constructor
         * @param {file.IDeleteUndocumentedFilesResponse=} [properties] Properties to set
         */
        function DeleteUndocumentedFilesResponse(properties) {
            this.items = [];
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * DeleteUndocumentedFilesResponse items.
         * @member {Array.<types.FileModel>} items
         * @memberof file.DeleteUndocumentedFilesResponse
         * @instance
         */
        DeleteUndocumentedFilesResponse.prototype.items = $util.emptyArray;

        /**
         * Encodes the specified DeleteUndocumentedFilesResponse message. Does not implicitly {@link file.DeleteUndocumentedFilesResponse.verify|verify} messages.
         * @function encode
         * @memberof file.DeleteUndocumentedFilesResponse
         * @static
         * @param {file.DeleteUndocumentedFilesResponse} message DeleteUndocumentedFilesResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        DeleteUndocumentedFilesResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.items != null && message.items.length)
                for (let i = 0; i < message.items.length; ++i)
                    $root.types.FileModel.encode(message.items[i], writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
            return writer;
        };

        /**
         * Decodes a DeleteUndocumentedFilesResponse message from the specified reader or buffer.
         * @function decode
         * @memberof file.DeleteUndocumentedFilesResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {file.DeleteUndocumentedFilesResponse} DeleteUndocumentedFilesResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        DeleteUndocumentedFilesResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.file.DeleteUndocumentedFilesResponse();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1:
                    if (!(message.items && message.items.length))
                        message.items = [];
                    message.items.push($root.types.FileModel.decode(reader, reader.uint32()));
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

    file.DiskInfoRequest = (function() {

        /**
         * Properties of a DiskInfoRequest.
         * @memberof file
         * @interface IDiskInfoRequest
         */

        /**
         * Constructs a new DiskInfoRequest.
         * @memberof file
         * @classdesc Represents a DiskInfoRequest.
         * @implements IDiskInfoRequest
         * @constructor
         * @param {file.IDiskInfoRequest=} [properties] Properties to set
         */
        function DiskInfoRequest(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * Encodes the specified DiskInfoRequest message. Does not implicitly {@link file.DiskInfoRequest.verify|verify} messages.
         * @function encode
         * @memberof file.DiskInfoRequest
         * @static
         * @param {file.DiskInfoRequest} message DiskInfoRequest message or plain object to encode
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
         * @memberof file.DiskInfoRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {file.DiskInfoRequest} DiskInfoRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        DiskInfoRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.file.DiskInfoRequest();
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

    file.DiskInfoResponse = (function() {

        /**
         * Properties of a DiskInfoResponse.
         * @memberof file
         * @interface IDiskInfoResponse
         * @property {number|null} [usage] DiskInfoResponse usage
         * @property {string|null} [humanize_usage] DiskInfoResponse humanize_usage
         */

        /**
         * Constructs a new DiskInfoResponse.
         * @memberof file
         * @classdesc Represents a DiskInfoResponse.
         * @implements IDiskInfoResponse
         * @constructor
         * @param {file.IDiskInfoResponse=} [properties] Properties to set
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
         * @memberof file.DiskInfoResponse
         * @instance
         */
        DiskInfoResponse.prototype.usage = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * DiskInfoResponse humanize_usage.
         * @member {string} humanize_usage
         * @memberof file.DiskInfoResponse
         * @instance
         */
        DiskInfoResponse.prototype.humanize_usage = "";

        /**
         * Encodes the specified DiskInfoResponse message. Does not implicitly {@link file.DiskInfoResponse.verify|verify} messages.
         * @function encode
         * @memberof file.DiskInfoResponse
         * @static
         * @param {file.DiskInfoResponse} message DiskInfoResponse message or plain object to encode
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
         * @memberof file.DiskInfoResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {file.DiskInfoResponse} DiskInfoResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        DiskInfoResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.file.DiskInfoResponse();
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

    file.ListRequest = (function() {

        /**
         * Properties of a ListRequest.
         * @memberof file
         * @interface IListRequest
         * @property {number|null} [page] ListRequest page
         * @property {number|null} [page_size] ListRequest page_size
         * @property {boolean|null} [without_deleted] ListRequest without_deleted
         */

        /**
         * Constructs a new ListRequest.
         * @memberof file
         * @classdesc Represents a ListRequest.
         * @implements IListRequest
         * @constructor
         * @param {file.IListRequest=} [properties] Properties to set
         */
        function ListRequest(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * ListRequest page.
         * @member {number} page
         * @memberof file.ListRequest
         * @instance
         */
        ListRequest.prototype.page = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * ListRequest page_size.
         * @member {number} page_size
         * @memberof file.ListRequest
         * @instance
         */
        ListRequest.prototype.page_size = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * ListRequest without_deleted.
         * @member {boolean} without_deleted
         * @memberof file.ListRequest
         * @instance
         */
        ListRequest.prototype.without_deleted = false;

        /**
         * Encodes the specified ListRequest message. Does not implicitly {@link file.ListRequest.verify|verify} messages.
         * @function encode
         * @memberof file.ListRequest
         * @static
         * @param {file.ListRequest} message ListRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ListRequest.encode = function encode(message, writer) {
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
         * Decodes a ListRequest message from the specified reader or buffer.
         * @function decode
         * @memberof file.ListRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {file.ListRequest} ListRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ListRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.file.ListRequest();
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

        return ListRequest;
    })();

    file.ListResponse = (function() {

        /**
         * Properties of a ListResponse.
         * @memberof file
         * @interface IListResponse
         * @property {number|null} [page] ListResponse page
         * @property {number|null} [page_size] ListResponse page_size
         * @property {Array.<types.FileModel>|null} [items] ListResponse items
         * @property {number|null} [count] ListResponse count
         */

        /**
         * Constructs a new ListResponse.
         * @memberof file
         * @classdesc Represents a ListResponse.
         * @implements IListResponse
         * @constructor
         * @param {file.IListResponse=} [properties] Properties to set
         */
        function ListResponse(properties) {
            this.items = [];
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * ListResponse page.
         * @member {number} page
         * @memberof file.ListResponse
         * @instance
         */
        ListResponse.prototype.page = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * ListResponse page_size.
         * @member {number} page_size
         * @memberof file.ListResponse
         * @instance
         */
        ListResponse.prototype.page_size = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * ListResponse items.
         * @member {Array.<types.FileModel>} items
         * @memberof file.ListResponse
         * @instance
         */
        ListResponse.prototype.items = $util.emptyArray;

        /**
         * ListResponse count.
         * @member {number} count
         * @memberof file.ListResponse
         * @instance
         */
        ListResponse.prototype.count = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Encodes the specified ListResponse message. Does not implicitly {@link file.ListResponse.verify|verify} messages.
         * @function encode
         * @memberof file.ListResponse
         * @static
         * @param {file.ListResponse} message ListResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ListResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.page != null && Object.hasOwnProperty.call(message, "page"))
                writer.uint32(/* id 1, wireType 0 =*/8).int64(message.page);
            if (message.page_size != null && Object.hasOwnProperty.call(message, "page_size"))
                writer.uint32(/* id 2, wireType 0 =*/16).int64(message.page_size);
            if (message.items != null && message.items.length)
                for (let i = 0; i < message.items.length; ++i)
                    $root.types.FileModel.encode(message.items[i], writer.uint32(/* id 3, wireType 2 =*/26).fork()).ldelim();
            if (message.count != null && Object.hasOwnProperty.call(message, "count"))
                writer.uint32(/* id 4, wireType 0 =*/32).int64(message.count);
            return writer;
        };

        /**
         * Decodes a ListResponse message from the specified reader or buffer.
         * @function decode
         * @memberof file.ListResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {file.ListResponse} ListResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ListResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.file.ListResponse();
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
                    message.items.push($root.types.FileModel.decode(reader, reader.uint32()));
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

        return ListResponse;
    })();

    file.File = (function() {

        /**
         * Constructs a new File service.
         * @memberof file
         * @classdesc Represents a File
         * @extends $protobuf.rpc.Service
         * @constructor
         * @param {$protobuf.RPCImpl} rpcImpl RPC implementation
         * @param {boolean} [requestDelimited=false] Whether requests are length-delimited
         * @param {boolean} [responseDelimited=false] Whether responses are length-delimited
         */
        function File(rpcImpl, requestDelimited, responseDelimited) {
            $protobuf.rpc.Service.call(this, rpcImpl, requestDelimited, responseDelimited);
        }

        (File.prototype = Object.create($protobuf.rpc.Service.prototype)).constructor = File;

        /**
         * Callback as used by {@link file.File#list}.
         * @memberof file.File
         * @typedef ListCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {file.ListResponse} [response] ListResponse
         */

        /**
         * Calls List.
         * @function list
         * @memberof file.File
         * @instance
         * @param {file.ListRequest} request ListRequest message or plain object
         * @param {file.File.ListCallback} callback Node-style callback called with the error, if any, and ListResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(File.prototype.list = function list(request, callback) {
            return this.rpcCall(list, $root.file.ListRequest, $root.file.ListResponse, request, callback);
        }, "name", { value: "List" });

        /**
         * Calls List.
         * @function list
         * @memberof file.File
         * @instance
         * @param {file.ListRequest} request ListRequest message or plain object
         * @returns {Promise<file.ListResponse>} Promise
         * @variation 2
         */

        /**
         * Callback as used by {@link file.File#delete_}.
         * @memberof file.File
         * @typedef DeleteCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {file.DeleteResponse} [response] DeleteResponse
         */

        /**
         * Calls Delete.
         * @function delete
         * @memberof file.File
         * @instance
         * @param {file.DeleteRequest} request DeleteRequest message or plain object
         * @param {file.File.DeleteCallback} callback Node-style callback called with the error, if any, and DeleteResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(File.prototype["delete"] = function delete_(request, callback) {
            return this.rpcCall(delete_, $root.file.DeleteRequest, $root.file.DeleteResponse, request, callback);
        }, "name", { value: "Delete" });

        /**
         * Calls Delete.
         * @function delete
         * @memberof file.File
         * @instance
         * @param {file.DeleteRequest} request DeleteRequest message or plain object
         * @returns {Promise<file.DeleteResponse>} Promise
         * @variation 2
         */

        /**
         * Callback as used by {@link file.File#deleteUndocumentedFiles}.
         * @memberof file.File
         * @typedef DeleteUndocumentedFilesCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {file.DeleteUndocumentedFilesResponse} [response] DeleteUndocumentedFilesResponse
         */

        /**
         * Calls DeleteUndocumentedFiles.
         * @function deleteUndocumentedFiles
         * @memberof file.File
         * @instance
         * @param {file.DeleteUndocumentedFilesRequest} request DeleteUndocumentedFilesRequest message or plain object
         * @param {file.File.DeleteUndocumentedFilesCallback} callback Node-style callback called with the error, if any, and DeleteUndocumentedFilesResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(File.prototype.deleteUndocumentedFiles = function deleteUndocumentedFiles(request, callback) {
            return this.rpcCall(deleteUndocumentedFiles, $root.file.DeleteUndocumentedFilesRequest, $root.file.DeleteUndocumentedFilesResponse, request, callback);
        }, "name", { value: "DeleteUndocumentedFiles" });

        /**
         * Calls DeleteUndocumentedFiles.
         * @function deleteUndocumentedFiles
         * @memberof file.File
         * @instance
         * @param {file.DeleteUndocumentedFilesRequest} request DeleteUndocumentedFilesRequest message or plain object
         * @returns {Promise<file.DeleteUndocumentedFilesResponse>} Promise
         * @variation 2
         */

        /**
         * Callback as used by {@link file.File#diskInfo}.
         * @memberof file.File
         * @typedef DiskInfoCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {file.DiskInfoResponse} [response] DiskInfoResponse
         */

        /**
         * Calls DiskInfo.
         * @function diskInfo
         * @memberof file.File
         * @instance
         * @param {file.DiskInfoRequest} request DiskInfoRequest message or plain object
         * @param {file.File.DiskInfoCallback} callback Node-style callback called with the error, if any, and DiskInfoResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(File.prototype.diskInfo = function diskInfo(request, callback) {
            return this.rpcCall(diskInfo, $root.file.DiskInfoRequest, $root.file.DiskInfoResponse, request, callback);
        }, "name", { value: "DiskInfo" });

        /**
         * Calls DiskInfo.
         * @function diskInfo
         * @memberof file.File
         * @instance
         * @param {file.DiskInfoRequest} request DiskInfoRequest message or plain object
         * @returns {Promise<file.DiskInfoResponse>} Promise
         * @variation 2
         */

        return File;
    })();

    return file;
})();

export const git = $root.git = (() => {

    /**
     * Namespace git.
     * @exports git
     * @namespace
     */
    const git = {};

    git.EnableProjectRequest = (function() {

        /**
         * Properties of an EnableProjectRequest.
         * @memberof git
         * @interface IEnableProjectRequest
         * @property {string|null} [git_project_id] EnableProjectRequest git_project_id
         */

        /**
         * Constructs a new EnableProjectRequest.
         * @memberof git
         * @classdesc Represents an EnableProjectRequest.
         * @implements IEnableProjectRequest
         * @constructor
         * @param {git.IEnableProjectRequest=} [properties] Properties to set
         */
        function EnableProjectRequest(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * EnableProjectRequest git_project_id.
         * @member {string} git_project_id
         * @memberof git.EnableProjectRequest
         * @instance
         */
        EnableProjectRequest.prototype.git_project_id = "";

        /**
         * Encodes the specified EnableProjectRequest message. Does not implicitly {@link git.EnableProjectRequest.verify|verify} messages.
         * @function encode
         * @memberof git.EnableProjectRequest
         * @static
         * @param {git.EnableProjectRequest} message EnableProjectRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        EnableProjectRequest.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.git_project_id != null && Object.hasOwnProperty.call(message, "git_project_id"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.git_project_id);
            return writer;
        };

        /**
         * Decodes an EnableProjectRequest message from the specified reader or buffer.
         * @function decode
         * @memberof git.EnableProjectRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {git.EnableProjectRequest} EnableProjectRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        EnableProjectRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.git.EnableProjectRequest();
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

        return EnableProjectRequest;
    })();

    git.DisableProjectRequest = (function() {

        /**
         * Properties of a DisableProjectRequest.
         * @memberof git
         * @interface IDisableProjectRequest
         * @property {string|null} [git_project_id] DisableProjectRequest git_project_id
         */

        /**
         * Constructs a new DisableProjectRequest.
         * @memberof git
         * @classdesc Represents a DisableProjectRequest.
         * @implements IDisableProjectRequest
         * @constructor
         * @param {git.IDisableProjectRequest=} [properties] Properties to set
         */
        function DisableProjectRequest(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * DisableProjectRequest git_project_id.
         * @member {string} git_project_id
         * @memberof git.DisableProjectRequest
         * @instance
         */
        DisableProjectRequest.prototype.git_project_id = "";

        /**
         * Encodes the specified DisableProjectRequest message. Does not implicitly {@link git.DisableProjectRequest.verify|verify} messages.
         * @function encode
         * @memberof git.DisableProjectRequest
         * @static
         * @param {git.DisableProjectRequest} message DisableProjectRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        DisableProjectRequest.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.git_project_id != null && Object.hasOwnProperty.call(message, "git_project_id"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.git_project_id);
            return writer;
        };

        /**
         * Decodes a DisableProjectRequest message from the specified reader or buffer.
         * @function decode
         * @memberof git.DisableProjectRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {git.DisableProjectRequest} DisableProjectRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        DisableProjectRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.git.DisableProjectRequest();
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

        return DisableProjectRequest;
    })();

    git.ProjectItem = (function() {

        /**
         * Properties of a ProjectItem.
         * @memberof git
         * @interface IProjectItem
         * @property {number|null} [id] ProjectItem id
         * @property {string|null} [name] ProjectItem name
         * @property {string|null} [path] ProjectItem path
         * @property {string|null} [web_url] ProjectItem web_url
         * @property {string|null} [avatar_url] ProjectItem avatar_url
         * @property {string|null} [description] ProjectItem description
         * @property {boolean|null} [enabled] ProjectItem enabled
         * @property {boolean|null} [global_enabled] ProjectItem global_enabled
         */

        /**
         * Constructs a new ProjectItem.
         * @memberof git
         * @classdesc Represents a ProjectItem.
         * @implements IProjectItem
         * @constructor
         * @param {git.IProjectItem=} [properties] Properties to set
         */
        function ProjectItem(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * ProjectItem id.
         * @member {number} id
         * @memberof git.ProjectItem
         * @instance
         */
        ProjectItem.prototype.id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * ProjectItem name.
         * @member {string} name
         * @memberof git.ProjectItem
         * @instance
         */
        ProjectItem.prototype.name = "";

        /**
         * ProjectItem path.
         * @member {string} path
         * @memberof git.ProjectItem
         * @instance
         */
        ProjectItem.prototype.path = "";

        /**
         * ProjectItem web_url.
         * @member {string} web_url
         * @memberof git.ProjectItem
         * @instance
         */
        ProjectItem.prototype.web_url = "";

        /**
         * ProjectItem avatar_url.
         * @member {string} avatar_url
         * @memberof git.ProjectItem
         * @instance
         */
        ProjectItem.prototype.avatar_url = "";

        /**
         * ProjectItem description.
         * @member {string} description
         * @memberof git.ProjectItem
         * @instance
         */
        ProjectItem.prototype.description = "";

        /**
         * ProjectItem enabled.
         * @member {boolean} enabled
         * @memberof git.ProjectItem
         * @instance
         */
        ProjectItem.prototype.enabled = false;

        /**
         * ProjectItem global_enabled.
         * @member {boolean} global_enabled
         * @memberof git.ProjectItem
         * @instance
         */
        ProjectItem.prototype.global_enabled = false;

        /**
         * Encodes the specified ProjectItem message. Does not implicitly {@link git.ProjectItem.verify|verify} messages.
         * @function encode
         * @memberof git.ProjectItem
         * @static
         * @param {git.ProjectItem} message ProjectItem message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ProjectItem.encode = function encode(message, writer) {
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
         * Decodes a ProjectItem message from the specified reader or buffer.
         * @function decode
         * @memberof git.ProjectItem
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {git.ProjectItem} ProjectItem
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ProjectItem.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.git.ProjectItem();
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

        return ProjectItem;
    })();

    git.AllProjectsResponse = (function() {

        /**
         * Properties of an AllProjectsResponse.
         * @memberof git
         * @interface IAllProjectsResponse
         * @property {Array.<git.ProjectItem>|null} [items] AllProjectsResponse items
         */

        /**
         * Constructs a new AllProjectsResponse.
         * @memberof git
         * @classdesc Represents an AllProjectsResponse.
         * @implements IAllProjectsResponse
         * @constructor
         * @param {git.IAllProjectsResponse=} [properties] Properties to set
         */
        function AllProjectsResponse(properties) {
            this.items = [];
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * AllProjectsResponse items.
         * @member {Array.<git.ProjectItem>} items
         * @memberof git.AllProjectsResponse
         * @instance
         */
        AllProjectsResponse.prototype.items = $util.emptyArray;

        /**
         * Encodes the specified AllProjectsResponse message. Does not implicitly {@link git.AllProjectsResponse.verify|verify} messages.
         * @function encode
         * @memberof git.AllProjectsResponse
         * @static
         * @param {git.AllProjectsResponse} message AllProjectsResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        AllProjectsResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.items != null && message.items.length)
                for (let i = 0; i < message.items.length; ++i)
                    $root.git.ProjectItem.encode(message.items[i], writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
            return writer;
        };

        /**
         * Decodes an AllProjectsResponse message from the specified reader or buffer.
         * @function decode
         * @memberof git.AllProjectsResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {git.AllProjectsResponse} AllProjectsResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        AllProjectsResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.git.AllProjectsResponse();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1:
                    if (!(message.items && message.items.length))
                        message.items = [];
                    message.items.push($root.git.ProjectItem.decode(reader, reader.uint32()));
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        return AllProjectsResponse;
    })();

    git.Option = (function() {

        /**
         * Properties of an Option.
         * @memberof git
         * @interface IOption
         * @property {string|null} [value] Option value
         * @property {string|null} [label] Option label
         * @property {string|null} [type] Option type
         * @property {boolean|null} [isLeaf] Option isLeaf
         * @property {string|null} [gitProjectId] Option gitProjectId
         * @property {string|null} [branch] Option branch
         */

        /**
         * Constructs a new Option.
         * @memberof git
         * @classdesc Represents an Option.
         * @implements IOption
         * @constructor
         * @param {git.IOption=} [properties] Properties to set
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
         * @memberof git.Option
         * @instance
         */
        Option.prototype.value = "";

        /**
         * Option label.
         * @member {string} label
         * @memberof git.Option
         * @instance
         */
        Option.prototype.label = "";

        /**
         * Option type.
         * @member {string} type
         * @memberof git.Option
         * @instance
         */
        Option.prototype.type = "";

        /**
         * Option isLeaf.
         * @member {boolean} isLeaf
         * @memberof git.Option
         * @instance
         */
        Option.prototype.isLeaf = false;

        /**
         * Option gitProjectId.
         * @member {string} gitProjectId
         * @memberof git.Option
         * @instance
         */
        Option.prototype.gitProjectId = "";

        /**
         * Option branch.
         * @member {string} branch
         * @memberof git.Option
         * @instance
         */
        Option.prototype.branch = "";

        /**
         * Encodes the specified Option message. Does not implicitly {@link git.Option.verify|verify} messages.
         * @function encode
         * @memberof git.Option
         * @static
         * @param {git.Option} message Option message or plain object to encode
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
            if (message.gitProjectId != null && Object.hasOwnProperty.call(message, "gitProjectId"))
                writer.uint32(/* id 5, wireType 2 =*/42).string(message.gitProjectId);
            if (message.branch != null && Object.hasOwnProperty.call(message, "branch"))
                writer.uint32(/* id 6, wireType 2 =*/50).string(message.branch);
            return writer;
        };

        /**
         * Decodes an Option message from the specified reader or buffer.
         * @function decode
         * @memberof git.Option
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {git.Option} Option
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Option.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.git.Option();
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
                    message.gitProjectId = reader.string();
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

    git.ProjectOptionsResponse = (function() {

        /**
         * Properties of a ProjectOptionsResponse.
         * @memberof git
         * @interface IProjectOptionsResponse
         * @property {Array.<git.Option>|null} [items] ProjectOptionsResponse items
         */

        /**
         * Constructs a new ProjectOptionsResponse.
         * @memberof git
         * @classdesc Represents a ProjectOptionsResponse.
         * @implements IProjectOptionsResponse
         * @constructor
         * @param {git.IProjectOptionsResponse=} [properties] Properties to set
         */
        function ProjectOptionsResponse(properties) {
            this.items = [];
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * ProjectOptionsResponse items.
         * @member {Array.<git.Option>} items
         * @memberof git.ProjectOptionsResponse
         * @instance
         */
        ProjectOptionsResponse.prototype.items = $util.emptyArray;

        /**
         * Encodes the specified ProjectOptionsResponse message. Does not implicitly {@link git.ProjectOptionsResponse.verify|verify} messages.
         * @function encode
         * @memberof git.ProjectOptionsResponse
         * @static
         * @param {git.ProjectOptionsResponse} message ProjectOptionsResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ProjectOptionsResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.items != null && message.items.length)
                for (let i = 0; i < message.items.length; ++i)
                    $root.git.Option.encode(message.items[i], writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
            return writer;
        };

        /**
         * Decodes a ProjectOptionsResponse message from the specified reader or buffer.
         * @function decode
         * @memberof git.ProjectOptionsResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {git.ProjectOptionsResponse} ProjectOptionsResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ProjectOptionsResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.git.ProjectOptionsResponse();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1:
                    if (!(message.items && message.items.length))
                        message.items = [];
                    message.items.push($root.git.Option.decode(reader, reader.uint32()));
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        return ProjectOptionsResponse;
    })();

    git.BranchOptionsRequest = (function() {

        /**
         * Properties of a BranchOptionsRequest.
         * @memberof git
         * @interface IBranchOptionsRequest
         * @property {string|null} [git_project_id] BranchOptionsRequest git_project_id
         * @property {boolean|null} [all] BranchOptionsRequest all
         */

        /**
         * Constructs a new BranchOptionsRequest.
         * @memberof git
         * @classdesc Represents a BranchOptionsRequest.
         * @implements IBranchOptionsRequest
         * @constructor
         * @param {git.IBranchOptionsRequest=} [properties] Properties to set
         */
        function BranchOptionsRequest(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * BranchOptionsRequest git_project_id.
         * @member {string} git_project_id
         * @memberof git.BranchOptionsRequest
         * @instance
         */
        BranchOptionsRequest.prototype.git_project_id = "";

        /**
         * BranchOptionsRequest all.
         * @member {boolean} all
         * @memberof git.BranchOptionsRequest
         * @instance
         */
        BranchOptionsRequest.prototype.all = false;

        /**
         * Encodes the specified BranchOptionsRequest message. Does not implicitly {@link git.BranchOptionsRequest.verify|verify} messages.
         * @function encode
         * @memberof git.BranchOptionsRequest
         * @static
         * @param {git.BranchOptionsRequest} message BranchOptionsRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        BranchOptionsRequest.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.git_project_id != null && Object.hasOwnProperty.call(message, "git_project_id"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.git_project_id);
            if (message.all != null && Object.hasOwnProperty.call(message, "all"))
                writer.uint32(/* id 2, wireType 0 =*/16).bool(message.all);
            return writer;
        };

        /**
         * Decodes a BranchOptionsRequest message from the specified reader or buffer.
         * @function decode
         * @memberof git.BranchOptionsRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {git.BranchOptionsRequest} BranchOptionsRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        BranchOptionsRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.git.BranchOptionsRequest();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1:
                    message.git_project_id = reader.string();
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

        return BranchOptionsRequest;
    })();

    git.BranchOptionsResponse = (function() {

        /**
         * Properties of a BranchOptionsResponse.
         * @memberof git
         * @interface IBranchOptionsResponse
         * @property {Array.<git.Option>|null} [items] BranchOptionsResponse items
         */

        /**
         * Constructs a new BranchOptionsResponse.
         * @memberof git
         * @classdesc Represents a BranchOptionsResponse.
         * @implements IBranchOptionsResponse
         * @constructor
         * @param {git.IBranchOptionsResponse=} [properties] Properties to set
         */
        function BranchOptionsResponse(properties) {
            this.items = [];
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * BranchOptionsResponse items.
         * @member {Array.<git.Option>} items
         * @memberof git.BranchOptionsResponse
         * @instance
         */
        BranchOptionsResponse.prototype.items = $util.emptyArray;

        /**
         * Encodes the specified BranchOptionsResponse message. Does not implicitly {@link git.BranchOptionsResponse.verify|verify} messages.
         * @function encode
         * @memberof git.BranchOptionsResponse
         * @static
         * @param {git.BranchOptionsResponse} message BranchOptionsResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        BranchOptionsResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.items != null && message.items.length)
                for (let i = 0; i < message.items.length; ++i)
                    $root.git.Option.encode(message.items[i], writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
            return writer;
        };

        /**
         * Decodes a BranchOptionsResponse message from the specified reader or buffer.
         * @function decode
         * @memberof git.BranchOptionsResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {git.BranchOptionsResponse} BranchOptionsResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        BranchOptionsResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.git.BranchOptionsResponse();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1:
                    if (!(message.items && message.items.length))
                        message.items = [];
                    message.items.push($root.git.Option.decode(reader, reader.uint32()));
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        return BranchOptionsResponse;
    })();

    git.CommitOptionsRequest = (function() {

        /**
         * Properties of a CommitOptionsRequest.
         * @memberof git
         * @interface ICommitOptionsRequest
         * @property {string|null} [git_project_id] CommitOptionsRequest git_project_id
         * @property {string|null} [branch] CommitOptionsRequest branch
         */

        /**
         * Constructs a new CommitOptionsRequest.
         * @memberof git
         * @classdesc Represents a CommitOptionsRequest.
         * @implements ICommitOptionsRequest
         * @constructor
         * @param {git.ICommitOptionsRequest=} [properties] Properties to set
         */
        function CommitOptionsRequest(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * CommitOptionsRequest git_project_id.
         * @member {string} git_project_id
         * @memberof git.CommitOptionsRequest
         * @instance
         */
        CommitOptionsRequest.prototype.git_project_id = "";

        /**
         * CommitOptionsRequest branch.
         * @member {string} branch
         * @memberof git.CommitOptionsRequest
         * @instance
         */
        CommitOptionsRequest.prototype.branch = "";

        /**
         * Encodes the specified CommitOptionsRequest message. Does not implicitly {@link git.CommitOptionsRequest.verify|verify} messages.
         * @function encode
         * @memberof git.CommitOptionsRequest
         * @static
         * @param {git.CommitOptionsRequest} message CommitOptionsRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        CommitOptionsRequest.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.git_project_id != null && Object.hasOwnProperty.call(message, "git_project_id"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.git_project_id);
            if (message.branch != null && Object.hasOwnProperty.call(message, "branch"))
                writer.uint32(/* id 2, wireType 2 =*/18).string(message.branch);
            return writer;
        };

        /**
         * Decodes a CommitOptionsRequest message from the specified reader or buffer.
         * @function decode
         * @memberof git.CommitOptionsRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {git.CommitOptionsRequest} CommitOptionsRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        CommitOptionsRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.git.CommitOptionsRequest();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1:
                    message.git_project_id = reader.string();
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

        return CommitOptionsRequest;
    })();

    git.CommitOptionsResponse = (function() {

        /**
         * Properties of a CommitOptionsResponse.
         * @memberof git
         * @interface ICommitOptionsResponse
         * @property {Array.<git.Option>|null} [items] CommitOptionsResponse items
         */

        /**
         * Constructs a new CommitOptionsResponse.
         * @memberof git
         * @classdesc Represents a CommitOptionsResponse.
         * @implements ICommitOptionsResponse
         * @constructor
         * @param {git.ICommitOptionsResponse=} [properties] Properties to set
         */
        function CommitOptionsResponse(properties) {
            this.items = [];
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * CommitOptionsResponse items.
         * @member {Array.<git.Option>} items
         * @memberof git.CommitOptionsResponse
         * @instance
         */
        CommitOptionsResponse.prototype.items = $util.emptyArray;

        /**
         * Encodes the specified CommitOptionsResponse message. Does not implicitly {@link git.CommitOptionsResponse.verify|verify} messages.
         * @function encode
         * @memberof git.CommitOptionsResponse
         * @static
         * @param {git.CommitOptionsResponse} message CommitOptionsResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        CommitOptionsResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.items != null && message.items.length)
                for (let i = 0; i < message.items.length; ++i)
                    $root.git.Option.encode(message.items[i], writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
            return writer;
        };

        /**
         * Decodes a CommitOptionsResponse message from the specified reader or buffer.
         * @function decode
         * @memberof git.CommitOptionsResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {git.CommitOptionsResponse} CommitOptionsResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        CommitOptionsResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.git.CommitOptionsResponse();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1:
                    if (!(message.items && message.items.length))
                        message.items = [];
                    message.items.push($root.git.Option.decode(reader, reader.uint32()));
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        return CommitOptionsResponse;
    })();

    git.CommitRequest = (function() {

        /**
         * Properties of a CommitRequest.
         * @memberof git
         * @interface ICommitRequest
         * @property {string|null} [git_project_id] CommitRequest git_project_id
         * @property {string|null} [branch] CommitRequest branch
         * @property {string|null} [commit] CommitRequest commit
         */

        /**
         * Constructs a new CommitRequest.
         * @memberof git
         * @classdesc Represents a CommitRequest.
         * @implements ICommitRequest
         * @constructor
         * @param {git.ICommitRequest=} [properties] Properties to set
         */
        function CommitRequest(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * CommitRequest git_project_id.
         * @member {string} git_project_id
         * @memberof git.CommitRequest
         * @instance
         */
        CommitRequest.prototype.git_project_id = "";

        /**
         * CommitRequest branch.
         * @member {string} branch
         * @memberof git.CommitRequest
         * @instance
         */
        CommitRequest.prototype.branch = "";

        /**
         * CommitRequest commit.
         * @member {string} commit
         * @memberof git.CommitRequest
         * @instance
         */
        CommitRequest.prototype.commit = "";

        /**
         * Encodes the specified CommitRequest message. Does not implicitly {@link git.CommitRequest.verify|verify} messages.
         * @function encode
         * @memberof git.CommitRequest
         * @static
         * @param {git.CommitRequest} message CommitRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        CommitRequest.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.git_project_id != null && Object.hasOwnProperty.call(message, "git_project_id"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.git_project_id);
            if (message.branch != null && Object.hasOwnProperty.call(message, "branch"))
                writer.uint32(/* id 2, wireType 2 =*/18).string(message.branch);
            if (message.commit != null && Object.hasOwnProperty.call(message, "commit"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.commit);
            return writer;
        };

        /**
         * Decodes a CommitRequest message from the specified reader or buffer.
         * @function decode
         * @memberof git.CommitRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {git.CommitRequest} CommitRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        CommitRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.git.CommitRequest();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1:
                    message.git_project_id = reader.string();
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

    git.CommitResponse = (function() {

        /**
         * Properties of a CommitResponse.
         * @memberof git
         * @interface ICommitResponse
         * @property {string|null} [id] CommitResponse id
         * @property {string|null} [short_id] CommitResponse short_id
         * @property {string|null} [git_project_id] CommitResponse git_project_id
         * @property {string|null} [label] CommitResponse label
         * @property {string|null} [title] CommitResponse title
         * @property {string|null} [branch] CommitResponse branch
         * @property {string|null} [author_name] CommitResponse author_name
         * @property {string|null} [author_email] CommitResponse author_email
         * @property {string|null} [committer_name] CommitResponse committer_name
         * @property {string|null} [committer_email] CommitResponse committer_email
         * @property {string|null} [web_url] CommitResponse web_url
         * @property {string|null} [message] CommitResponse message
         * @property {string|null} [committed_date] CommitResponse committed_date
         * @property {string|null} [created_at] CommitResponse created_at
         */

        /**
         * Constructs a new CommitResponse.
         * @memberof git
         * @classdesc Represents a CommitResponse.
         * @implements ICommitResponse
         * @constructor
         * @param {git.ICommitResponse=} [properties] Properties to set
         */
        function CommitResponse(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * CommitResponse id.
         * @member {string} id
         * @memberof git.CommitResponse
         * @instance
         */
        CommitResponse.prototype.id = "";

        /**
         * CommitResponse short_id.
         * @member {string} short_id
         * @memberof git.CommitResponse
         * @instance
         */
        CommitResponse.prototype.short_id = "";

        /**
         * CommitResponse git_project_id.
         * @member {string} git_project_id
         * @memberof git.CommitResponse
         * @instance
         */
        CommitResponse.prototype.git_project_id = "";

        /**
         * CommitResponse label.
         * @member {string} label
         * @memberof git.CommitResponse
         * @instance
         */
        CommitResponse.prototype.label = "";

        /**
         * CommitResponse title.
         * @member {string} title
         * @memberof git.CommitResponse
         * @instance
         */
        CommitResponse.prototype.title = "";

        /**
         * CommitResponse branch.
         * @member {string} branch
         * @memberof git.CommitResponse
         * @instance
         */
        CommitResponse.prototype.branch = "";

        /**
         * CommitResponse author_name.
         * @member {string} author_name
         * @memberof git.CommitResponse
         * @instance
         */
        CommitResponse.prototype.author_name = "";

        /**
         * CommitResponse author_email.
         * @member {string} author_email
         * @memberof git.CommitResponse
         * @instance
         */
        CommitResponse.prototype.author_email = "";

        /**
         * CommitResponse committer_name.
         * @member {string} committer_name
         * @memberof git.CommitResponse
         * @instance
         */
        CommitResponse.prototype.committer_name = "";

        /**
         * CommitResponse committer_email.
         * @member {string} committer_email
         * @memberof git.CommitResponse
         * @instance
         */
        CommitResponse.prototype.committer_email = "";

        /**
         * CommitResponse web_url.
         * @member {string} web_url
         * @memberof git.CommitResponse
         * @instance
         */
        CommitResponse.prototype.web_url = "";

        /**
         * CommitResponse message.
         * @member {string} message
         * @memberof git.CommitResponse
         * @instance
         */
        CommitResponse.prototype.message = "";

        /**
         * CommitResponse committed_date.
         * @member {string} committed_date
         * @memberof git.CommitResponse
         * @instance
         */
        CommitResponse.prototype.committed_date = "";

        /**
         * CommitResponse created_at.
         * @member {string} created_at
         * @memberof git.CommitResponse
         * @instance
         */
        CommitResponse.prototype.created_at = "";

        /**
         * Encodes the specified CommitResponse message. Does not implicitly {@link git.CommitResponse.verify|verify} messages.
         * @function encode
         * @memberof git.CommitResponse
         * @static
         * @param {git.CommitResponse} message CommitResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        CommitResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.id != null && Object.hasOwnProperty.call(message, "id"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.id);
            if (message.short_id != null && Object.hasOwnProperty.call(message, "short_id"))
                writer.uint32(/* id 2, wireType 2 =*/18).string(message.short_id);
            if (message.git_project_id != null && Object.hasOwnProperty.call(message, "git_project_id"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.git_project_id);
            if (message.label != null && Object.hasOwnProperty.call(message, "label"))
                writer.uint32(/* id 4, wireType 2 =*/34).string(message.label);
            if (message.title != null && Object.hasOwnProperty.call(message, "title"))
                writer.uint32(/* id 5, wireType 2 =*/42).string(message.title);
            if (message.branch != null && Object.hasOwnProperty.call(message, "branch"))
                writer.uint32(/* id 6, wireType 2 =*/50).string(message.branch);
            if (message.author_name != null && Object.hasOwnProperty.call(message, "author_name"))
                writer.uint32(/* id 7, wireType 2 =*/58).string(message.author_name);
            if (message.author_email != null && Object.hasOwnProperty.call(message, "author_email"))
                writer.uint32(/* id 8, wireType 2 =*/66).string(message.author_email);
            if (message.committer_name != null && Object.hasOwnProperty.call(message, "committer_name"))
                writer.uint32(/* id 9, wireType 2 =*/74).string(message.committer_name);
            if (message.committer_email != null && Object.hasOwnProperty.call(message, "committer_email"))
                writer.uint32(/* id 10, wireType 2 =*/82).string(message.committer_email);
            if (message.web_url != null && Object.hasOwnProperty.call(message, "web_url"))
                writer.uint32(/* id 11, wireType 2 =*/90).string(message.web_url);
            if (message.message != null && Object.hasOwnProperty.call(message, "message"))
                writer.uint32(/* id 12, wireType 2 =*/98).string(message.message);
            if (message.committed_date != null && Object.hasOwnProperty.call(message, "committed_date"))
                writer.uint32(/* id 13, wireType 2 =*/106).string(message.committed_date);
            if (message.created_at != null && Object.hasOwnProperty.call(message, "created_at"))
                writer.uint32(/* id 14, wireType 2 =*/114).string(message.created_at);
            return writer;
        };

        /**
         * Decodes a CommitResponse message from the specified reader or buffer.
         * @function decode
         * @memberof git.CommitResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {git.CommitResponse} CommitResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        CommitResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.git.CommitResponse();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1:
                    message.id = reader.string();
                    break;
                case 2:
                    message.short_id = reader.string();
                    break;
                case 3:
                    message.git_project_id = reader.string();
                    break;
                case 4:
                    message.label = reader.string();
                    break;
                case 5:
                    message.title = reader.string();
                    break;
                case 6:
                    message.branch = reader.string();
                    break;
                case 7:
                    message.author_name = reader.string();
                    break;
                case 8:
                    message.author_email = reader.string();
                    break;
                case 9:
                    message.committer_name = reader.string();
                    break;
                case 10:
                    message.committer_email = reader.string();
                    break;
                case 11:
                    message.web_url = reader.string();
                    break;
                case 12:
                    message.message = reader.string();
                    break;
                case 13:
                    message.committed_date = reader.string();
                    break;
                case 14:
                    message.created_at = reader.string();
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

    git.PipelineInfoRequest = (function() {

        /**
         * Properties of a PipelineInfoRequest.
         * @memberof git
         * @interface IPipelineInfoRequest
         * @property {string|null} [git_project_id] PipelineInfoRequest git_project_id
         * @property {string|null} [branch] PipelineInfoRequest branch
         * @property {string|null} [commit] PipelineInfoRequest commit
         */

        /**
         * Constructs a new PipelineInfoRequest.
         * @memberof git
         * @classdesc Represents a PipelineInfoRequest.
         * @implements IPipelineInfoRequest
         * @constructor
         * @param {git.IPipelineInfoRequest=} [properties] Properties to set
         */
        function PipelineInfoRequest(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * PipelineInfoRequest git_project_id.
         * @member {string} git_project_id
         * @memberof git.PipelineInfoRequest
         * @instance
         */
        PipelineInfoRequest.prototype.git_project_id = "";

        /**
         * PipelineInfoRequest branch.
         * @member {string} branch
         * @memberof git.PipelineInfoRequest
         * @instance
         */
        PipelineInfoRequest.prototype.branch = "";

        /**
         * PipelineInfoRequest commit.
         * @member {string} commit
         * @memberof git.PipelineInfoRequest
         * @instance
         */
        PipelineInfoRequest.prototype.commit = "";

        /**
         * Encodes the specified PipelineInfoRequest message. Does not implicitly {@link git.PipelineInfoRequest.verify|verify} messages.
         * @function encode
         * @memberof git.PipelineInfoRequest
         * @static
         * @param {git.PipelineInfoRequest} message PipelineInfoRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        PipelineInfoRequest.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.git_project_id != null && Object.hasOwnProperty.call(message, "git_project_id"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.git_project_id);
            if (message.branch != null && Object.hasOwnProperty.call(message, "branch"))
                writer.uint32(/* id 2, wireType 2 =*/18).string(message.branch);
            if (message.commit != null && Object.hasOwnProperty.call(message, "commit"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.commit);
            return writer;
        };

        /**
         * Decodes a PipelineInfoRequest message from the specified reader or buffer.
         * @function decode
         * @memberof git.PipelineInfoRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {git.PipelineInfoRequest} PipelineInfoRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        PipelineInfoRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.git.PipelineInfoRequest();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1:
                    message.git_project_id = reader.string();
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

    git.PipelineInfoResponse = (function() {

        /**
         * Properties of a PipelineInfoResponse.
         * @memberof git
         * @interface IPipelineInfoResponse
         * @property {string|null} [status] PipelineInfoResponse status
         * @property {string|null} [web_url] PipelineInfoResponse web_url
         */

        /**
         * Constructs a new PipelineInfoResponse.
         * @memberof git
         * @classdesc Represents a PipelineInfoResponse.
         * @implements IPipelineInfoResponse
         * @constructor
         * @param {git.IPipelineInfoResponse=} [properties] Properties to set
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
         * @memberof git.PipelineInfoResponse
         * @instance
         */
        PipelineInfoResponse.prototype.status = "";

        /**
         * PipelineInfoResponse web_url.
         * @member {string} web_url
         * @memberof git.PipelineInfoResponse
         * @instance
         */
        PipelineInfoResponse.prototype.web_url = "";

        /**
         * Encodes the specified PipelineInfoResponse message. Does not implicitly {@link git.PipelineInfoResponse.verify|verify} messages.
         * @function encode
         * @memberof git.PipelineInfoResponse
         * @static
         * @param {git.PipelineInfoResponse} message PipelineInfoResponse message or plain object to encode
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
         * @memberof git.PipelineInfoResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {git.PipelineInfoResponse} PipelineInfoResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        PipelineInfoResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.git.PipelineInfoResponse();
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

    git.ConfigFileRequest = (function() {

        /**
         * Properties of a ConfigFileRequest.
         * @memberof git
         * @interface IConfigFileRequest
         * @property {string|null} [git_project_id] ConfigFileRequest git_project_id
         * @property {string|null} [branch] ConfigFileRequest branch
         */

        /**
         * Constructs a new ConfigFileRequest.
         * @memberof git
         * @classdesc Represents a ConfigFileRequest.
         * @implements IConfigFileRequest
         * @constructor
         * @param {git.IConfigFileRequest=} [properties] Properties to set
         */
        function ConfigFileRequest(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * ConfigFileRequest git_project_id.
         * @member {string} git_project_id
         * @memberof git.ConfigFileRequest
         * @instance
         */
        ConfigFileRequest.prototype.git_project_id = "";

        /**
         * ConfigFileRequest branch.
         * @member {string} branch
         * @memberof git.ConfigFileRequest
         * @instance
         */
        ConfigFileRequest.prototype.branch = "";

        /**
         * Encodes the specified ConfigFileRequest message. Does not implicitly {@link git.ConfigFileRequest.verify|verify} messages.
         * @function encode
         * @memberof git.ConfigFileRequest
         * @static
         * @param {git.ConfigFileRequest} message ConfigFileRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ConfigFileRequest.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.git_project_id != null && Object.hasOwnProperty.call(message, "git_project_id"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.git_project_id);
            if (message.branch != null && Object.hasOwnProperty.call(message, "branch"))
                writer.uint32(/* id 2, wireType 2 =*/18).string(message.branch);
            return writer;
        };

        /**
         * Decodes a ConfigFileRequest message from the specified reader or buffer.
         * @function decode
         * @memberof git.ConfigFileRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {git.ConfigFileRequest} ConfigFileRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ConfigFileRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.git.ConfigFileRequest();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1:
                    message.git_project_id = reader.string();
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

    git.ConfigFileResponse = (function() {

        /**
         * Properties of a ConfigFileResponse.
         * @memberof git
         * @interface IConfigFileResponse
         * @property {string|null} [data] ConfigFileResponse data
         * @property {string|null} [type] ConfigFileResponse type
         * @property {Array.<mars.Element>|null} [elements] ConfigFileResponse elements
         */

        /**
         * Constructs a new ConfigFileResponse.
         * @memberof git
         * @classdesc Represents a ConfigFileResponse.
         * @implements IConfigFileResponse
         * @constructor
         * @param {git.IConfigFileResponse=} [properties] Properties to set
         */
        function ConfigFileResponse(properties) {
            this.elements = [];
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * ConfigFileResponse data.
         * @member {string} data
         * @memberof git.ConfigFileResponse
         * @instance
         */
        ConfigFileResponse.prototype.data = "";

        /**
         * ConfigFileResponse type.
         * @member {string} type
         * @memberof git.ConfigFileResponse
         * @instance
         */
        ConfigFileResponse.prototype.type = "";

        /**
         * ConfigFileResponse elements.
         * @member {Array.<mars.Element>} elements
         * @memberof git.ConfigFileResponse
         * @instance
         */
        ConfigFileResponse.prototype.elements = $util.emptyArray;

        /**
         * Encodes the specified ConfigFileResponse message. Does not implicitly {@link git.ConfigFileResponse.verify|verify} messages.
         * @function encode
         * @memberof git.ConfigFileResponse
         * @static
         * @param {git.ConfigFileResponse} message ConfigFileResponse message or plain object to encode
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
            if (message.elements != null && message.elements.length)
                for (let i = 0; i < message.elements.length; ++i)
                    $root.mars.Element.encode(message.elements[i], writer.uint32(/* id 3, wireType 2 =*/26).fork()).ldelim();
            return writer;
        };

        /**
         * Decodes a ConfigFileResponse message from the specified reader or buffer.
         * @function decode
         * @memberof git.ConfigFileResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {git.ConfigFileResponse} ConfigFileResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ConfigFileResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.git.ConfigFileResponse();
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
                    message.elements.push($root.mars.Element.decode(reader, reader.uint32()));
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

    git.EnableProjectResponse = (function() {

        /**
         * Properties of an EnableProjectResponse.
         * @memberof git
         * @interface IEnableProjectResponse
         */

        /**
         * Constructs a new EnableProjectResponse.
         * @memberof git
         * @classdesc Represents an EnableProjectResponse.
         * @implements IEnableProjectResponse
         * @constructor
         * @param {git.IEnableProjectResponse=} [properties] Properties to set
         */
        function EnableProjectResponse(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * Encodes the specified EnableProjectResponse message. Does not implicitly {@link git.EnableProjectResponse.verify|verify} messages.
         * @function encode
         * @memberof git.EnableProjectResponse
         * @static
         * @param {git.EnableProjectResponse} message EnableProjectResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        EnableProjectResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            return writer;
        };

        /**
         * Decodes an EnableProjectResponse message from the specified reader or buffer.
         * @function decode
         * @memberof git.EnableProjectResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {git.EnableProjectResponse} EnableProjectResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        EnableProjectResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.git.EnableProjectResponse();
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

        return EnableProjectResponse;
    })();

    git.DisableProjectResponse = (function() {

        /**
         * Properties of a DisableProjectResponse.
         * @memberof git
         * @interface IDisableProjectResponse
         */

        /**
         * Constructs a new DisableProjectResponse.
         * @memberof git
         * @classdesc Represents a DisableProjectResponse.
         * @implements IDisableProjectResponse
         * @constructor
         * @param {git.IDisableProjectResponse=} [properties] Properties to set
         */
        function DisableProjectResponse(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * Encodes the specified DisableProjectResponse message. Does not implicitly {@link git.DisableProjectResponse.verify|verify} messages.
         * @function encode
         * @memberof git.DisableProjectResponse
         * @static
         * @param {git.DisableProjectResponse} message DisableProjectResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        DisableProjectResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            return writer;
        };

        /**
         * Decodes a DisableProjectResponse message from the specified reader or buffer.
         * @function decode
         * @memberof git.DisableProjectResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {git.DisableProjectResponse} DisableProjectResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        DisableProjectResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.git.DisableProjectResponse();
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

        return DisableProjectResponse;
    })();

    git.AllProjectsRequest = (function() {

        /**
         * Properties of an AllProjectsRequest.
         * @memberof git
         * @interface IAllProjectsRequest
         */

        /**
         * Constructs a new AllProjectsRequest.
         * @memberof git
         * @classdesc Represents an AllProjectsRequest.
         * @implements IAllProjectsRequest
         * @constructor
         * @param {git.IAllProjectsRequest=} [properties] Properties to set
         */
        function AllProjectsRequest(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * Encodes the specified AllProjectsRequest message. Does not implicitly {@link git.AllProjectsRequest.verify|verify} messages.
         * @function encode
         * @memberof git.AllProjectsRequest
         * @static
         * @param {git.AllProjectsRequest} message AllProjectsRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        AllProjectsRequest.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            return writer;
        };

        /**
         * Decodes an AllProjectsRequest message from the specified reader or buffer.
         * @function decode
         * @memberof git.AllProjectsRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {git.AllProjectsRequest} AllProjectsRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        AllProjectsRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.git.AllProjectsRequest();
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

        return AllProjectsRequest;
    })();

    git.ProjectOptionsRequest = (function() {

        /**
         * Properties of a ProjectOptionsRequest.
         * @memberof git
         * @interface IProjectOptionsRequest
         */

        /**
         * Constructs a new ProjectOptionsRequest.
         * @memberof git
         * @classdesc Represents a ProjectOptionsRequest.
         * @implements IProjectOptionsRequest
         * @constructor
         * @param {git.IProjectOptionsRequest=} [properties] Properties to set
         */
        function ProjectOptionsRequest(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * Encodes the specified ProjectOptionsRequest message. Does not implicitly {@link git.ProjectOptionsRequest.verify|verify} messages.
         * @function encode
         * @memberof git.ProjectOptionsRequest
         * @static
         * @param {git.ProjectOptionsRequest} message ProjectOptionsRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ProjectOptionsRequest.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            return writer;
        };

        /**
         * Decodes a ProjectOptionsRequest message from the specified reader or buffer.
         * @function decode
         * @memberof git.ProjectOptionsRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {git.ProjectOptionsRequest} ProjectOptionsRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ProjectOptionsRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.git.ProjectOptionsRequest();
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

        return ProjectOptionsRequest;
    })();

    git.Git = (function() {

        /**
         * Constructs a new Git service.
         * @memberof git
         * @classdesc Represents a Git
         * @extends $protobuf.rpc.Service
         * @constructor
         * @param {$protobuf.RPCImpl} rpcImpl RPC implementation
         * @param {boolean} [requestDelimited=false] Whether requests are length-delimited
         * @param {boolean} [responseDelimited=false] Whether responses are length-delimited
         */
        function Git(rpcImpl, requestDelimited, responseDelimited) {
            $protobuf.rpc.Service.call(this, rpcImpl, requestDelimited, responseDelimited);
        }

        (Git.prototype = Object.create($protobuf.rpc.Service.prototype)).constructor = Git;

        /**
         * Callback as used by {@link git.Git#enableProject}.
         * @memberof git.Git
         * @typedef EnableProjectCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {git.EnableProjectResponse} [response] EnableProjectResponse
         */

        /**
         * Calls EnableProject.
         * @function enableProject
         * @memberof git.Git
         * @instance
         * @param {git.EnableProjectRequest} request EnableProjectRequest message or plain object
         * @param {git.Git.EnableProjectCallback} callback Node-style callback called with the error, if any, and EnableProjectResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(Git.prototype.enableProject = function enableProject(request, callback) {
            return this.rpcCall(enableProject, $root.git.EnableProjectRequest, $root.git.EnableProjectResponse, request, callback);
        }, "name", { value: "EnableProject" });

        /**
         * Calls EnableProject.
         * @function enableProject
         * @memberof git.Git
         * @instance
         * @param {git.EnableProjectRequest} request EnableProjectRequest message or plain object
         * @returns {Promise<git.EnableProjectResponse>} Promise
         * @variation 2
         */

        /**
         * Callback as used by {@link git.Git#disableProject}.
         * @memberof git.Git
         * @typedef DisableProjectCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {git.DisableProjectResponse} [response] DisableProjectResponse
         */

        /**
         * Calls DisableProject.
         * @function disableProject
         * @memberof git.Git
         * @instance
         * @param {git.DisableProjectRequest} request DisableProjectRequest message or plain object
         * @param {git.Git.DisableProjectCallback} callback Node-style callback called with the error, if any, and DisableProjectResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(Git.prototype.disableProject = function disableProject(request, callback) {
            return this.rpcCall(disableProject, $root.git.DisableProjectRequest, $root.git.DisableProjectResponse, request, callback);
        }, "name", { value: "DisableProject" });

        /**
         * Calls DisableProject.
         * @function disableProject
         * @memberof git.Git
         * @instance
         * @param {git.DisableProjectRequest} request DisableProjectRequest message or plain object
         * @returns {Promise<git.DisableProjectResponse>} Promise
         * @variation 2
         */

        /**
         * Callback as used by {@link git.Git#all}.
         * @memberof git.Git
         * @typedef AllCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {git.AllProjectsResponse} [response] AllProjectsResponse
         */

        /**
         * Calls All.
         * @function all
         * @memberof git.Git
         * @instance
         * @param {git.AllProjectsRequest} request AllProjectsRequest message or plain object
         * @param {git.Git.AllCallback} callback Node-style callback called with the error, if any, and AllProjectsResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(Git.prototype.all = function all(request, callback) {
            return this.rpcCall(all, $root.git.AllProjectsRequest, $root.git.AllProjectsResponse, request, callback);
        }, "name", { value: "All" });

        /**
         * Calls All.
         * @function all
         * @memberof git.Git
         * @instance
         * @param {git.AllProjectsRequest} request AllProjectsRequest message or plain object
         * @returns {Promise<git.AllProjectsResponse>} Promise
         * @variation 2
         */

        /**
         * Callback as used by {@link git.Git#projectOptions}.
         * @memberof git.Git
         * @typedef ProjectOptionsCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {git.ProjectOptionsResponse} [response] ProjectOptionsResponse
         */

        /**
         * Calls ProjectOptions.
         * @function projectOptions
         * @memberof git.Git
         * @instance
         * @param {git.ProjectOptionsRequest} request ProjectOptionsRequest message or plain object
         * @param {git.Git.ProjectOptionsCallback} callback Node-style callback called with the error, if any, and ProjectOptionsResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(Git.prototype.projectOptions = function projectOptions(request, callback) {
            return this.rpcCall(projectOptions, $root.git.ProjectOptionsRequest, $root.git.ProjectOptionsResponse, request, callback);
        }, "name", { value: "ProjectOptions" });

        /**
         * Calls ProjectOptions.
         * @function projectOptions
         * @memberof git.Git
         * @instance
         * @param {git.ProjectOptionsRequest} request ProjectOptionsRequest message or plain object
         * @returns {Promise<git.ProjectOptionsResponse>} Promise
         * @variation 2
         */

        /**
         * Callback as used by {@link git.Git#branchOptions}.
         * @memberof git.Git
         * @typedef BranchOptionsCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {git.BranchOptionsResponse} [response] BranchOptionsResponse
         */

        /**
         * Calls BranchOptions.
         * @function branchOptions
         * @memberof git.Git
         * @instance
         * @param {git.BranchOptionsRequest} request BranchOptionsRequest message or plain object
         * @param {git.Git.BranchOptionsCallback} callback Node-style callback called with the error, if any, and BranchOptionsResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(Git.prototype.branchOptions = function branchOptions(request, callback) {
            return this.rpcCall(branchOptions, $root.git.BranchOptionsRequest, $root.git.BranchOptionsResponse, request, callback);
        }, "name", { value: "BranchOptions" });

        /**
         * Calls BranchOptions.
         * @function branchOptions
         * @memberof git.Git
         * @instance
         * @param {git.BranchOptionsRequest} request BranchOptionsRequest message or plain object
         * @returns {Promise<git.BranchOptionsResponse>} Promise
         * @variation 2
         */

        /**
         * Callback as used by {@link git.Git#commitOptions}.
         * @memberof git.Git
         * @typedef CommitOptionsCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {git.CommitOptionsResponse} [response] CommitOptionsResponse
         */

        /**
         * Calls CommitOptions.
         * @function commitOptions
         * @memberof git.Git
         * @instance
         * @param {git.CommitOptionsRequest} request CommitOptionsRequest message or plain object
         * @param {git.Git.CommitOptionsCallback} callback Node-style callback called with the error, if any, and CommitOptionsResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(Git.prototype.commitOptions = function commitOptions(request, callback) {
            return this.rpcCall(commitOptions, $root.git.CommitOptionsRequest, $root.git.CommitOptionsResponse, request, callback);
        }, "name", { value: "CommitOptions" });

        /**
         * Calls CommitOptions.
         * @function commitOptions
         * @memberof git.Git
         * @instance
         * @param {git.CommitOptionsRequest} request CommitOptionsRequest message or plain object
         * @returns {Promise<git.CommitOptionsResponse>} Promise
         * @variation 2
         */

        /**
         * Callback as used by {@link git.Git#commit}.
         * @memberof git.Git
         * @typedef CommitCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {git.CommitResponse} [response] CommitResponse
         */

        /**
         * Calls Commit.
         * @function commit
         * @memberof git.Git
         * @instance
         * @param {git.CommitRequest} request CommitRequest message or plain object
         * @param {git.Git.CommitCallback} callback Node-style callback called with the error, if any, and CommitResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(Git.prototype.commit = function commit(request, callback) {
            return this.rpcCall(commit, $root.git.CommitRequest, $root.git.CommitResponse, request, callback);
        }, "name", { value: "Commit" });

        /**
         * Calls Commit.
         * @function commit
         * @memberof git.Git
         * @instance
         * @param {git.CommitRequest} request CommitRequest message or plain object
         * @returns {Promise<git.CommitResponse>} Promise
         * @variation 2
         */

        /**
         * Callback as used by {@link git.Git#pipelineInfo}.
         * @memberof git.Git
         * @typedef PipelineInfoCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {git.PipelineInfoResponse} [response] PipelineInfoResponse
         */

        /**
         * Calls PipelineInfo.
         * @function pipelineInfo
         * @memberof git.Git
         * @instance
         * @param {git.PipelineInfoRequest} request PipelineInfoRequest message or plain object
         * @param {git.Git.PipelineInfoCallback} callback Node-style callback called with the error, if any, and PipelineInfoResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(Git.prototype.pipelineInfo = function pipelineInfo(request, callback) {
            return this.rpcCall(pipelineInfo, $root.git.PipelineInfoRequest, $root.git.PipelineInfoResponse, request, callback);
        }, "name", { value: "PipelineInfo" });

        /**
         * Calls PipelineInfo.
         * @function pipelineInfo
         * @memberof git.Git
         * @instance
         * @param {git.PipelineInfoRequest} request PipelineInfoRequest message or plain object
         * @returns {Promise<git.PipelineInfoResponse>} Promise
         * @variation 2
         */

        /**
         * Callback as used by {@link git.Git#marsConfigFile}.
         * @memberof git.Git
         * @typedef MarsConfigFileCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {git.ConfigFileResponse} [response] ConfigFileResponse
         */

        /**
         * Calls MarsConfigFile.
         * @function marsConfigFile
         * @memberof git.Git
         * @instance
         * @param {git.ConfigFileRequest} request ConfigFileRequest message or plain object
         * @param {git.Git.MarsConfigFileCallback} callback Node-style callback called with the error, if any, and ConfigFileResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(Git.prototype.marsConfigFile = function marsConfigFile(request, callback) {
            return this.rpcCall(marsConfigFile, $root.git.ConfigFileRequest, $root.git.ConfigFileResponse, request, callback);
        }, "name", { value: "MarsConfigFile" });

        /**
         * Calls MarsConfigFile.
         * @function marsConfigFile
         * @memberof git.Git
         * @instance
         * @param {git.ConfigFileRequest} request ConfigFileRequest message or plain object
         * @returns {Promise<git.ConfigFileResponse>} Promise
         * @variation 2
         */

        return Git;
    })();

    return git;
})();

export const gitconfig = $root.gitconfig = (() => {

    /**
     * Namespace gitconfig.
     * @exports gitconfig
     * @namespace
     */
    const gitconfig = {};

    gitconfig.FileRequest = (function() {

        /**
         * Properties of a FileRequest.
         * @memberof gitconfig
         * @interface IFileRequest
         * @property {string|null} [git_project_id] FileRequest git_project_id
         * @property {string|null} [branch] FileRequest branch
         */

        /**
         * Constructs a new FileRequest.
         * @memberof gitconfig
         * @classdesc Represents a FileRequest.
         * @implements IFileRequest
         * @constructor
         * @param {gitconfig.IFileRequest=} [properties] Properties to set
         */
        function FileRequest(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * FileRequest git_project_id.
         * @member {string} git_project_id
         * @memberof gitconfig.FileRequest
         * @instance
         */
        FileRequest.prototype.git_project_id = "";

        /**
         * FileRequest branch.
         * @member {string} branch
         * @memberof gitconfig.FileRequest
         * @instance
         */
        FileRequest.prototype.branch = "";

        /**
         * Encodes the specified FileRequest message. Does not implicitly {@link gitconfig.FileRequest.verify|verify} messages.
         * @function encode
         * @memberof gitconfig.FileRequest
         * @static
         * @param {gitconfig.FileRequest} message FileRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        FileRequest.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.git_project_id != null && Object.hasOwnProperty.call(message, "git_project_id"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.git_project_id);
            if (message.branch != null && Object.hasOwnProperty.call(message, "branch"))
                writer.uint32(/* id 2, wireType 2 =*/18).string(message.branch);
            return writer;
        };

        /**
         * Decodes a FileRequest message from the specified reader or buffer.
         * @function decode
         * @memberof gitconfig.FileRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {gitconfig.FileRequest} FileRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        FileRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.gitconfig.FileRequest();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1:
                    message.git_project_id = reader.string();
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

        return FileRequest;
    })();

    gitconfig.FileResponse = (function() {

        /**
         * Properties of a FileResponse.
         * @memberof gitconfig
         * @interface IFileResponse
         * @property {string|null} [data] FileResponse data
         * @property {string|null} [type] FileResponse type
         * @property {Array.<mars.Element>|null} [elements] FileResponse elements
         */

        /**
         * Constructs a new FileResponse.
         * @memberof gitconfig
         * @classdesc Represents a FileResponse.
         * @implements IFileResponse
         * @constructor
         * @param {gitconfig.IFileResponse=} [properties] Properties to set
         */
        function FileResponse(properties) {
            this.elements = [];
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * FileResponse data.
         * @member {string} data
         * @memberof gitconfig.FileResponse
         * @instance
         */
        FileResponse.prototype.data = "";

        /**
         * FileResponse type.
         * @member {string} type
         * @memberof gitconfig.FileResponse
         * @instance
         */
        FileResponse.prototype.type = "";

        /**
         * FileResponse elements.
         * @member {Array.<mars.Element>} elements
         * @memberof gitconfig.FileResponse
         * @instance
         */
        FileResponse.prototype.elements = $util.emptyArray;

        /**
         * Encodes the specified FileResponse message. Does not implicitly {@link gitconfig.FileResponse.verify|verify} messages.
         * @function encode
         * @memberof gitconfig.FileResponse
         * @static
         * @param {gitconfig.FileResponse} message FileResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        FileResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.data != null && Object.hasOwnProperty.call(message, "data"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.data);
            if (message.type != null && Object.hasOwnProperty.call(message, "type"))
                writer.uint32(/* id 2, wireType 2 =*/18).string(message.type);
            if (message.elements != null && message.elements.length)
                for (let i = 0; i < message.elements.length; ++i)
                    $root.mars.Element.encode(message.elements[i], writer.uint32(/* id 3, wireType 2 =*/26).fork()).ldelim();
            return writer;
        };

        /**
         * Decodes a FileResponse message from the specified reader or buffer.
         * @function decode
         * @memberof gitconfig.FileResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {gitconfig.FileResponse} FileResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        FileResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.gitconfig.FileResponse();
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
                    message.elements.push($root.mars.Element.decode(reader, reader.uint32()));
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        return FileResponse;
    })();

    gitconfig.ShowRequest = (function() {

        /**
         * Properties of a ShowRequest.
         * @memberof gitconfig
         * @interface IShowRequest
         * @property {number|null} [git_project_id] ShowRequest git_project_id
         * @property {string|null} [branch] ShowRequest branch
         */

        /**
         * Constructs a new ShowRequest.
         * @memberof gitconfig
         * @classdesc Represents a ShowRequest.
         * @implements IShowRequest
         * @constructor
         * @param {gitconfig.IShowRequest=} [properties] Properties to set
         */
        function ShowRequest(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * ShowRequest git_project_id.
         * @member {number} git_project_id
         * @memberof gitconfig.ShowRequest
         * @instance
         */
        ShowRequest.prototype.git_project_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * ShowRequest branch.
         * @member {string} branch
         * @memberof gitconfig.ShowRequest
         * @instance
         */
        ShowRequest.prototype.branch = "";

        /**
         * Encodes the specified ShowRequest message. Does not implicitly {@link gitconfig.ShowRequest.verify|verify} messages.
         * @function encode
         * @memberof gitconfig.ShowRequest
         * @static
         * @param {gitconfig.ShowRequest} message ShowRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ShowRequest.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.git_project_id != null && Object.hasOwnProperty.call(message, "git_project_id"))
                writer.uint32(/* id 1, wireType 0 =*/8).int64(message.git_project_id);
            if (message.branch != null && Object.hasOwnProperty.call(message, "branch"))
                writer.uint32(/* id 2, wireType 2 =*/18).string(message.branch);
            return writer;
        };

        /**
         * Decodes a ShowRequest message from the specified reader or buffer.
         * @function decode
         * @memberof gitconfig.ShowRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {gitconfig.ShowRequest} ShowRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ShowRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.gitconfig.ShowRequest();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1:
                    message.git_project_id = reader.int64();
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

        return ShowRequest;
    })();

    gitconfig.ShowResponse = (function() {

        /**
         * Properties of a ShowResponse.
         * @memberof gitconfig
         * @interface IShowResponse
         * @property {string|null} [branch] ShowResponse branch
         * @property {mars.Config|null} [config] ShowResponse config
         */

        /**
         * Constructs a new ShowResponse.
         * @memberof gitconfig
         * @classdesc Represents a ShowResponse.
         * @implements IShowResponse
         * @constructor
         * @param {gitconfig.IShowResponse=} [properties] Properties to set
         */
        function ShowResponse(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * ShowResponse branch.
         * @member {string} branch
         * @memberof gitconfig.ShowResponse
         * @instance
         */
        ShowResponse.prototype.branch = "";

        /**
         * ShowResponse config.
         * @member {mars.Config|null|undefined} config
         * @memberof gitconfig.ShowResponse
         * @instance
         */
        ShowResponse.prototype.config = null;

        /**
         * Encodes the specified ShowResponse message. Does not implicitly {@link gitconfig.ShowResponse.verify|verify} messages.
         * @function encode
         * @memberof gitconfig.ShowResponse
         * @static
         * @param {gitconfig.ShowResponse} message ShowResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ShowResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.branch != null && Object.hasOwnProperty.call(message, "branch"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.branch);
            if (message.config != null && Object.hasOwnProperty.call(message, "config"))
                $root.mars.Config.encode(message.config, writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim();
            return writer;
        };

        /**
         * Decodes a ShowResponse message from the specified reader or buffer.
         * @function decode
         * @memberof gitconfig.ShowResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {gitconfig.ShowResponse} ShowResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ShowResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.gitconfig.ShowResponse();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1:
                    message.branch = reader.string();
                    break;
                case 2:
                    message.config = $root.mars.Config.decode(reader, reader.uint32());
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        return ShowResponse;
    })();

    gitconfig.GlobalConfigRequest = (function() {

        /**
         * Properties of a GlobalConfigRequest.
         * @memberof gitconfig
         * @interface IGlobalConfigRequest
         * @property {number|null} [git_project_id] GlobalConfigRequest git_project_id
         */

        /**
         * Constructs a new GlobalConfigRequest.
         * @memberof gitconfig
         * @classdesc Represents a GlobalConfigRequest.
         * @implements IGlobalConfigRequest
         * @constructor
         * @param {gitconfig.IGlobalConfigRequest=} [properties] Properties to set
         */
        function GlobalConfigRequest(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * GlobalConfigRequest git_project_id.
         * @member {number} git_project_id
         * @memberof gitconfig.GlobalConfigRequest
         * @instance
         */
        GlobalConfigRequest.prototype.git_project_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Encodes the specified GlobalConfigRequest message. Does not implicitly {@link gitconfig.GlobalConfigRequest.verify|verify} messages.
         * @function encode
         * @memberof gitconfig.GlobalConfigRequest
         * @static
         * @param {gitconfig.GlobalConfigRequest} message GlobalConfigRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        GlobalConfigRequest.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.git_project_id != null && Object.hasOwnProperty.call(message, "git_project_id"))
                writer.uint32(/* id 1, wireType 0 =*/8).int64(message.git_project_id);
            return writer;
        };

        /**
         * Decodes a GlobalConfigRequest message from the specified reader or buffer.
         * @function decode
         * @memberof gitconfig.GlobalConfigRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {gitconfig.GlobalConfigRequest} GlobalConfigRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        GlobalConfigRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.gitconfig.GlobalConfigRequest();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1:
                    message.git_project_id = reader.int64();
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

    gitconfig.GlobalConfigResponse = (function() {

        /**
         * Properties of a GlobalConfigResponse.
         * @memberof gitconfig
         * @interface IGlobalConfigResponse
         * @property {boolean|null} [enabled] GlobalConfigResponse enabled
         * @property {mars.Config|null} [config] GlobalConfigResponse config
         */

        /**
         * Constructs a new GlobalConfigResponse.
         * @memberof gitconfig
         * @classdesc Represents a GlobalConfigResponse.
         * @implements IGlobalConfigResponse
         * @constructor
         * @param {gitconfig.IGlobalConfigResponse=} [properties] Properties to set
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
         * @memberof gitconfig.GlobalConfigResponse
         * @instance
         */
        GlobalConfigResponse.prototype.enabled = false;

        /**
         * GlobalConfigResponse config.
         * @member {mars.Config|null|undefined} config
         * @memberof gitconfig.GlobalConfigResponse
         * @instance
         */
        GlobalConfigResponse.prototype.config = null;

        /**
         * Encodes the specified GlobalConfigResponse message. Does not implicitly {@link gitconfig.GlobalConfigResponse.verify|verify} messages.
         * @function encode
         * @memberof gitconfig.GlobalConfigResponse
         * @static
         * @param {gitconfig.GlobalConfigResponse} message GlobalConfigResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        GlobalConfigResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.enabled != null && Object.hasOwnProperty.call(message, "enabled"))
                writer.uint32(/* id 1, wireType 0 =*/8).bool(message.enabled);
            if (message.config != null && Object.hasOwnProperty.call(message, "config"))
                $root.mars.Config.encode(message.config, writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim();
            return writer;
        };

        /**
         * Decodes a GlobalConfigResponse message from the specified reader or buffer.
         * @function decode
         * @memberof gitconfig.GlobalConfigResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {gitconfig.GlobalConfigResponse} GlobalConfigResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        GlobalConfigResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.gitconfig.GlobalConfigResponse();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1:
                    message.enabled = reader.bool();
                    break;
                case 2:
                    message.config = $root.mars.Config.decode(reader, reader.uint32());
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

    gitconfig.UpdateRequest = (function() {

        /**
         * Properties of an UpdateRequest.
         * @memberof gitconfig
         * @interface IUpdateRequest
         * @property {number|null} [git_project_id] UpdateRequest git_project_id
         * @property {mars.Config|null} [config] UpdateRequest config
         */

        /**
         * Constructs a new UpdateRequest.
         * @memberof gitconfig
         * @classdesc Represents an UpdateRequest.
         * @implements IUpdateRequest
         * @constructor
         * @param {gitconfig.IUpdateRequest=} [properties] Properties to set
         */
        function UpdateRequest(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * UpdateRequest git_project_id.
         * @member {number} git_project_id
         * @memberof gitconfig.UpdateRequest
         * @instance
         */
        UpdateRequest.prototype.git_project_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * UpdateRequest config.
         * @member {mars.Config|null|undefined} config
         * @memberof gitconfig.UpdateRequest
         * @instance
         */
        UpdateRequest.prototype.config = null;

        /**
         * Encodes the specified UpdateRequest message. Does not implicitly {@link gitconfig.UpdateRequest.verify|verify} messages.
         * @function encode
         * @memberof gitconfig.UpdateRequest
         * @static
         * @param {gitconfig.UpdateRequest} message UpdateRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        UpdateRequest.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.git_project_id != null && Object.hasOwnProperty.call(message, "git_project_id"))
                writer.uint32(/* id 1, wireType 0 =*/8).int64(message.git_project_id);
            if (message.config != null && Object.hasOwnProperty.call(message, "config"))
                $root.mars.Config.encode(message.config, writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim();
            return writer;
        };

        /**
         * Decodes an UpdateRequest message from the specified reader or buffer.
         * @function decode
         * @memberof gitconfig.UpdateRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {gitconfig.UpdateRequest} UpdateRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        UpdateRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.gitconfig.UpdateRequest();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1:
                    message.git_project_id = reader.int64();
                    break;
                case 2:
                    message.config = $root.mars.Config.decode(reader, reader.uint32());
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        return UpdateRequest;
    })();

    gitconfig.UpdateResponse = (function() {

        /**
         * Properties of an UpdateResponse.
         * @memberof gitconfig
         * @interface IUpdateResponse
         * @property {mars.Config|null} [config] UpdateResponse config
         */

        /**
         * Constructs a new UpdateResponse.
         * @memberof gitconfig
         * @classdesc Represents an UpdateResponse.
         * @implements IUpdateResponse
         * @constructor
         * @param {gitconfig.IUpdateResponse=} [properties] Properties to set
         */
        function UpdateResponse(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * UpdateResponse config.
         * @member {mars.Config|null|undefined} config
         * @memberof gitconfig.UpdateResponse
         * @instance
         */
        UpdateResponse.prototype.config = null;

        /**
         * Encodes the specified UpdateResponse message. Does not implicitly {@link gitconfig.UpdateResponse.verify|verify} messages.
         * @function encode
         * @memberof gitconfig.UpdateResponse
         * @static
         * @param {gitconfig.UpdateResponse} message UpdateResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        UpdateResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.config != null && Object.hasOwnProperty.call(message, "config"))
                $root.mars.Config.encode(message.config, writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
            return writer;
        };

        /**
         * Decodes an UpdateResponse message from the specified reader or buffer.
         * @function decode
         * @memberof gitconfig.UpdateResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {gitconfig.UpdateResponse} UpdateResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        UpdateResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.gitconfig.UpdateResponse();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1:
                    message.config = $root.mars.Config.decode(reader, reader.uint32());
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        return UpdateResponse;
    })();

    gitconfig.ToggleGlobalStatusRequest = (function() {

        /**
         * Properties of a ToggleGlobalStatusRequest.
         * @memberof gitconfig
         * @interface IToggleGlobalStatusRequest
         * @property {number|null} [git_project_id] ToggleGlobalStatusRequest git_project_id
         * @property {boolean|null} [enabled] ToggleGlobalStatusRequest enabled
         */

        /**
         * Constructs a new ToggleGlobalStatusRequest.
         * @memberof gitconfig
         * @classdesc Represents a ToggleGlobalStatusRequest.
         * @implements IToggleGlobalStatusRequest
         * @constructor
         * @param {gitconfig.IToggleGlobalStatusRequest=} [properties] Properties to set
         */
        function ToggleGlobalStatusRequest(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * ToggleGlobalStatusRequest git_project_id.
         * @member {number} git_project_id
         * @memberof gitconfig.ToggleGlobalStatusRequest
         * @instance
         */
        ToggleGlobalStatusRequest.prototype.git_project_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * ToggleGlobalStatusRequest enabled.
         * @member {boolean} enabled
         * @memberof gitconfig.ToggleGlobalStatusRequest
         * @instance
         */
        ToggleGlobalStatusRequest.prototype.enabled = false;

        /**
         * Encodes the specified ToggleGlobalStatusRequest message. Does not implicitly {@link gitconfig.ToggleGlobalStatusRequest.verify|verify} messages.
         * @function encode
         * @memberof gitconfig.ToggleGlobalStatusRequest
         * @static
         * @param {gitconfig.ToggleGlobalStatusRequest} message ToggleGlobalStatusRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ToggleGlobalStatusRequest.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.git_project_id != null && Object.hasOwnProperty.call(message, "git_project_id"))
                writer.uint32(/* id 1, wireType 0 =*/8).int64(message.git_project_id);
            if (message.enabled != null && Object.hasOwnProperty.call(message, "enabled"))
                writer.uint32(/* id 2, wireType 0 =*/16).bool(message.enabled);
            return writer;
        };

        /**
         * Decodes a ToggleGlobalStatusRequest message from the specified reader or buffer.
         * @function decode
         * @memberof gitconfig.ToggleGlobalStatusRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {gitconfig.ToggleGlobalStatusRequest} ToggleGlobalStatusRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ToggleGlobalStatusRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.gitconfig.ToggleGlobalStatusRequest();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1:
                    message.git_project_id = reader.int64();
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

        return ToggleGlobalStatusRequest;
    })();

    gitconfig.DefaultChartValuesRequest = (function() {

        /**
         * Properties of a DefaultChartValuesRequest.
         * @memberof gitconfig
         * @interface IDefaultChartValuesRequest
         * @property {number|null} [git_project_id] DefaultChartValuesRequest git_project_id
         * @property {string|null} [branch] DefaultChartValuesRequest branch
         */

        /**
         * Constructs a new DefaultChartValuesRequest.
         * @memberof gitconfig
         * @classdesc Represents a DefaultChartValuesRequest.
         * @implements IDefaultChartValuesRequest
         * @constructor
         * @param {gitconfig.IDefaultChartValuesRequest=} [properties] Properties to set
         */
        function DefaultChartValuesRequest(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * DefaultChartValuesRequest git_project_id.
         * @member {number} git_project_id
         * @memberof gitconfig.DefaultChartValuesRequest
         * @instance
         */
        DefaultChartValuesRequest.prototype.git_project_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * DefaultChartValuesRequest branch.
         * @member {string} branch
         * @memberof gitconfig.DefaultChartValuesRequest
         * @instance
         */
        DefaultChartValuesRequest.prototype.branch = "";

        /**
         * Encodes the specified DefaultChartValuesRequest message. Does not implicitly {@link gitconfig.DefaultChartValuesRequest.verify|verify} messages.
         * @function encode
         * @memberof gitconfig.DefaultChartValuesRequest
         * @static
         * @param {gitconfig.DefaultChartValuesRequest} message DefaultChartValuesRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        DefaultChartValuesRequest.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.git_project_id != null && Object.hasOwnProperty.call(message, "git_project_id"))
                writer.uint32(/* id 1, wireType 0 =*/8).int64(message.git_project_id);
            if (message.branch != null && Object.hasOwnProperty.call(message, "branch"))
                writer.uint32(/* id 2, wireType 2 =*/18).string(message.branch);
            return writer;
        };

        /**
         * Decodes a DefaultChartValuesRequest message from the specified reader or buffer.
         * @function decode
         * @memberof gitconfig.DefaultChartValuesRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {gitconfig.DefaultChartValuesRequest} DefaultChartValuesRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        DefaultChartValuesRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.gitconfig.DefaultChartValuesRequest();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1:
                    message.git_project_id = reader.int64();
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

    gitconfig.DefaultChartValuesResponse = (function() {

        /**
         * Properties of a DefaultChartValuesResponse.
         * @memberof gitconfig
         * @interface IDefaultChartValuesResponse
         * @property {string|null} [value] DefaultChartValuesResponse value
         */

        /**
         * Constructs a new DefaultChartValuesResponse.
         * @memberof gitconfig
         * @classdesc Represents a DefaultChartValuesResponse.
         * @implements IDefaultChartValuesResponse
         * @constructor
         * @param {gitconfig.IDefaultChartValuesResponse=} [properties] Properties to set
         */
        function DefaultChartValuesResponse(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * DefaultChartValuesResponse value.
         * @member {string} value
         * @memberof gitconfig.DefaultChartValuesResponse
         * @instance
         */
        DefaultChartValuesResponse.prototype.value = "";

        /**
         * Encodes the specified DefaultChartValuesResponse message. Does not implicitly {@link gitconfig.DefaultChartValuesResponse.verify|verify} messages.
         * @function encode
         * @memberof gitconfig.DefaultChartValuesResponse
         * @static
         * @param {gitconfig.DefaultChartValuesResponse} message DefaultChartValuesResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        DefaultChartValuesResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.value != null && Object.hasOwnProperty.call(message, "value"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.value);
            return writer;
        };

        /**
         * Decodes a DefaultChartValuesResponse message from the specified reader or buffer.
         * @function decode
         * @memberof gitconfig.DefaultChartValuesResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {gitconfig.DefaultChartValuesResponse} DefaultChartValuesResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        DefaultChartValuesResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.gitconfig.DefaultChartValuesResponse();
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

        return DefaultChartValuesResponse;
    })();

    gitconfig.ToggleGlobalStatusResponse = (function() {

        /**
         * Properties of a ToggleGlobalStatusResponse.
         * @memberof gitconfig
         * @interface IToggleGlobalStatusResponse
         */

        /**
         * Constructs a new ToggleGlobalStatusResponse.
         * @memberof gitconfig
         * @classdesc Represents a ToggleGlobalStatusResponse.
         * @implements IToggleGlobalStatusResponse
         * @constructor
         * @param {gitconfig.IToggleGlobalStatusResponse=} [properties] Properties to set
         */
        function ToggleGlobalStatusResponse(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * Encodes the specified ToggleGlobalStatusResponse message. Does not implicitly {@link gitconfig.ToggleGlobalStatusResponse.verify|verify} messages.
         * @function encode
         * @memberof gitconfig.ToggleGlobalStatusResponse
         * @static
         * @param {gitconfig.ToggleGlobalStatusResponse} message ToggleGlobalStatusResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ToggleGlobalStatusResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            return writer;
        };

        /**
         * Decodes a ToggleGlobalStatusResponse message from the specified reader or buffer.
         * @function decode
         * @memberof gitconfig.ToggleGlobalStatusResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {gitconfig.ToggleGlobalStatusResponse} ToggleGlobalStatusResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ToggleGlobalStatusResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.gitconfig.ToggleGlobalStatusResponse();
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

        return ToggleGlobalStatusResponse;
    })();

    gitconfig.GitConfig = (function() {

        /**
         * Constructs a new GitConfig service.
         * @memberof gitconfig
         * @classdesc Represents a GitConfig
         * @extends $protobuf.rpc.Service
         * @constructor
         * @param {$protobuf.RPCImpl} rpcImpl RPC implementation
         * @param {boolean} [requestDelimited=false] Whether requests are length-delimited
         * @param {boolean} [responseDelimited=false] Whether responses are length-delimited
         */
        function GitConfig(rpcImpl, requestDelimited, responseDelimited) {
            $protobuf.rpc.Service.call(this, rpcImpl, requestDelimited, responseDelimited);
        }

        (GitConfig.prototype = Object.create($protobuf.rpc.Service.prototype)).constructor = GitConfig;

        /**
         * Callback as used by {@link gitconfig.GitConfig#show}.
         * @memberof gitconfig.GitConfig
         * @typedef ShowCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {gitconfig.ShowResponse} [response] ShowResponse
         */

        /**
         * Calls Show.
         * @function show
         * @memberof gitconfig.GitConfig
         * @instance
         * @param {gitconfig.ShowRequest} request ShowRequest message or plain object
         * @param {gitconfig.GitConfig.ShowCallback} callback Node-style callback called with the error, if any, and ShowResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(GitConfig.prototype.show = function show(request, callback) {
            return this.rpcCall(show, $root.gitconfig.ShowRequest, $root.gitconfig.ShowResponse, request, callback);
        }, "name", { value: "Show" });

        /**
         * Calls Show.
         * @function show
         * @memberof gitconfig.GitConfig
         * @instance
         * @param {gitconfig.ShowRequest} request ShowRequest message or plain object
         * @returns {Promise<gitconfig.ShowResponse>} Promise
         * @variation 2
         */

        /**
         * Callback as used by {@link gitconfig.GitConfig#globalConfig}.
         * @memberof gitconfig.GitConfig
         * @typedef GlobalConfigCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {gitconfig.GlobalConfigResponse} [response] GlobalConfigResponse
         */

        /**
         * Calls GlobalConfig.
         * @function globalConfig
         * @memberof gitconfig.GitConfig
         * @instance
         * @param {gitconfig.GlobalConfigRequest} request GlobalConfigRequest message or plain object
         * @param {gitconfig.GitConfig.GlobalConfigCallback} callback Node-style callback called with the error, if any, and GlobalConfigResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(GitConfig.prototype.globalConfig = function globalConfig(request, callback) {
            return this.rpcCall(globalConfig, $root.gitconfig.GlobalConfigRequest, $root.gitconfig.GlobalConfigResponse, request, callback);
        }, "name", { value: "GlobalConfig" });

        /**
         * Calls GlobalConfig.
         * @function globalConfig
         * @memberof gitconfig.GitConfig
         * @instance
         * @param {gitconfig.GlobalConfigRequest} request GlobalConfigRequest message or plain object
         * @returns {Promise<gitconfig.GlobalConfigResponse>} Promise
         * @variation 2
         */

        /**
         * Callback as used by {@link gitconfig.GitConfig#toggleGlobalStatus}.
         * @memberof gitconfig.GitConfig
         * @typedef ToggleGlobalStatusCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {gitconfig.ToggleGlobalStatusResponse} [response] ToggleGlobalStatusResponse
         */

        /**
         * Calls ToggleGlobalStatus.
         * @function toggleGlobalStatus
         * @memberof gitconfig.GitConfig
         * @instance
         * @param {gitconfig.ToggleGlobalStatusRequest} request ToggleGlobalStatusRequest message or plain object
         * @param {gitconfig.GitConfig.ToggleGlobalStatusCallback} callback Node-style callback called with the error, if any, and ToggleGlobalStatusResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(GitConfig.prototype.toggleGlobalStatus = function toggleGlobalStatus(request, callback) {
            return this.rpcCall(toggleGlobalStatus, $root.gitconfig.ToggleGlobalStatusRequest, $root.gitconfig.ToggleGlobalStatusResponse, request, callback);
        }, "name", { value: "ToggleGlobalStatus" });

        /**
         * Calls ToggleGlobalStatus.
         * @function toggleGlobalStatus
         * @memberof gitconfig.GitConfig
         * @instance
         * @param {gitconfig.ToggleGlobalStatusRequest} request ToggleGlobalStatusRequest message or plain object
         * @returns {Promise<gitconfig.ToggleGlobalStatusResponse>} Promise
         * @variation 2
         */

        /**
         * Callback as used by {@link gitconfig.GitConfig#update}.
         * @memberof gitconfig.GitConfig
         * @typedef UpdateCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {gitconfig.UpdateResponse} [response] UpdateResponse
         */

        /**
         * Calls Update.
         * @function update
         * @memberof gitconfig.GitConfig
         * @instance
         * @param {gitconfig.UpdateRequest} request UpdateRequest message or plain object
         * @param {gitconfig.GitConfig.UpdateCallback} callback Node-style callback called with the error, if any, and UpdateResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(GitConfig.prototype.update = function update(request, callback) {
            return this.rpcCall(update, $root.gitconfig.UpdateRequest, $root.gitconfig.UpdateResponse, request, callback);
        }, "name", { value: "Update" });

        /**
         * Calls Update.
         * @function update
         * @memberof gitconfig.GitConfig
         * @instance
         * @param {gitconfig.UpdateRequest} request UpdateRequest message or plain object
         * @returns {Promise<gitconfig.UpdateResponse>} Promise
         * @variation 2
         */

        /**
         * Callback as used by {@link gitconfig.GitConfig#getDefaultChartValues}.
         * @memberof gitconfig.GitConfig
         * @typedef GetDefaultChartValuesCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {gitconfig.DefaultChartValuesResponse} [response] DefaultChartValuesResponse
         */

        /**
         * Calls GetDefaultChartValues.
         * @function getDefaultChartValues
         * @memberof gitconfig.GitConfig
         * @instance
         * @param {gitconfig.DefaultChartValuesRequest} request DefaultChartValuesRequest message or plain object
         * @param {gitconfig.GitConfig.GetDefaultChartValuesCallback} callback Node-style callback called with the error, if any, and DefaultChartValuesResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(GitConfig.prototype.getDefaultChartValues = function getDefaultChartValues(request, callback) {
            return this.rpcCall(getDefaultChartValues, $root.gitconfig.DefaultChartValuesRequest, $root.gitconfig.DefaultChartValuesResponse, request, callback);
        }, "name", { value: "GetDefaultChartValues" });

        /**
         * Calls GetDefaultChartValues.
         * @function getDefaultChartValues
         * @memberof gitconfig.GitConfig
         * @instance
         * @param {gitconfig.DefaultChartValuesRequest} request DefaultChartValuesRequest message or plain object
         * @returns {Promise<gitconfig.DefaultChartValuesResponse>} Promise
         * @variation 2
         */

        return GitConfig;
    })();

    return gitconfig;
})();

export const mars = $root.mars = (() => {

    /**
     * Namespace mars.
     * @exports mars
     * @namespace
     */
    const mars = {};

    mars.Config = (function() {

        /**
         * Properties of a Config.
         * @memberof mars
         * @interface IConfig
         * @property {string|null} [config_file] Config config_file
         * @property {string|null} [config_file_values] Config config_file_values
         * @property {string|null} [config_field] Config config_field
         * @property {boolean|null} [is_simple_env] Config is_simple_env
         * @property {string|null} [config_file_type] Config config_file_type
         * @property {string|null} [local_chart_path] Config local_chart_path
         * @property {Array.<string>|null} [branches] Config branches
         * @property {string|null} [values_yaml] Config values_yaml
         * @property {Array.<mars.Element>|null} [elements] Config elements
         */

        /**
         * Constructs a new Config.
         * @memberof mars
         * @classdesc Represents a Config.
         * @implements IConfig
         * @constructor
         * @param {mars.IConfig=} [properties] Properties to set
         */
        function Config(properties) {
            this.branches = [];
            this.elements = [];
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * Config config_file.
         * @member {string} config_file
         * @memberof mars.Config
         * @instance
         */
        Config.prototype.config_file = "";

        /**
         * Config config_file_values.
         * @member {string} config_file_values
         * @memberof mars.Config
         * @instance
         */
        Config.prototype.config_file_values = "";

        /**
         * Config config_field.
         * @member {string} config_field
         * @memberof mars.Config
         * @instance
         */
        Config.prototype.config_field = "";

        /**
         * Config is_simple_env.
         * @member {boolean} is_simple_env
         * @memberof mars.Config
         * @instance
         */
        Config.prototype.is_simple_env = false;

        /**
         * Config config_file_type.
         * @member {string} config_file_type
         * @memberof mars.Config
         * @instance
         */
        Config.prototype.config_file_type = "";

        /**
         * Config local_chart_path.
         * @member {string} local_chart_path
         * @memberof mars.Config
         * @instance
         */
        Config.prototype.local_chart_path = "";

        /**
         * Config branches.
         * @member {Array.<string>} branches
         * @memberof mars.Config
         * @instance
         */
        Config.prototype.branches = $util.emptyArray;

        /**
         * Config values_yaml.
         * @member {string} values_yaml
         * @memberof mars.Config
         * @instance
         */
        Config.prototype.values_yaml = "";

        /**
         * Config elements.
         * @member {Array.<mars.Element>} elements
         * @memberof mars.Config
         * @instance
         */
        Config.prototype.elements = $util.emptyArray;

        /**
         * Encodes the specified Config message. Does not implicitly {@link mars.Config.verify|verify} messages.
         * @function encode
         * @memberof mars.Config
         * @static
         * @param {mars.Config} message Config message or plain object to encode
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
            if (message.elements != null && message.elements.length)
                for (let i = 0; i < message.elements.length; ++i)
                    $root.mars.Element.encode(message.elements[i], writer.uint32(/* id 9, wireType 2 =*/74).fork()).ldelim();
            return writer;
        };

        /**
         * Decodes a Config message from the specified reader or buffer.
         * @function decode
         * @memberof mars.Config
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {mars.Config} Config
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Config.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.mars.Config();
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
                    message.elements.push($root.mars.Element.decode(reader, reader.uint32()));
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

    /**
     * ElementType enum.
     * @name mars.ElementType
     * @enum {number}
     * @property {number} ElementTypeUnknown=0 ElementTypeUnknown value
     * @property {number} ElementTypeInput=1 ElementTypeInput value
     * @property {number} ElementTypeInputNumber=2 ElementTypeInputNumber value
     * @property {number} ElementTypeSelect=3 ElementTypeSelect value
     * @property {number} ElementTypeRadio=4 ElementTypeRadio value
     * @property {number} ElementTypeSwitch=5 ElementTypeSwitch value
     */
    mars.ElementType = (function() {
        const valuesById = {}, values = Object.create(valuesById);
        values[valuesById[0] = "ElementTypeUnknown"] = 0;
        values[valuesById[1] = "ElementTypeInput"] = 1;
        values[valuesById[2] = "ElementTypeInputNumber"] = 2;
        values[valuesById[3] = "ElementTypeSelect"] = 3;
        values[valuesById[4] = "ElementTypeRadio"] = 4;
        values[valuesById[5] = "ElementTypeSwitch"] = 5;
        return values;
    })();

    mars.Element = (function() {

        /**
         * Properties of an Element.
         * @memberof mars
         * @interface IElement
         * @property {string|null} [path] Element path
         * @property {mars.ElementType|null} [type] Element type
         * @property {string|null} ["default"] Element default
         * @property {string|null} [description] Element description
         * @property {Array.<string>|null} [select_values] Element select_values
         */

        /**
         * Constructs a new Element.
         * @memberof mars
         * @classdesc Represents an Element.
         * @implements IElement
         * @constructor
         * @param {mars.IElement=} [properties] Properties to set
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
         * @memberof mars.Element
         * @instance
         */
        Element.prototype.path = "";

        /**
         * Element type.
         * @member {mars.ElementType} type
         * @memberof mars.Element
         * @instance
         */
        Element.prototype.type = 0;

        /**
         * Element default.
         * @member {string} default
         * @memberof mars.Element
         * @instance
         */
        Element.prototype["default"] = "";

        /**
         * Element description.
         * @member {string} description
         * @memberof mars.Element
         * @instance
         */
        Element.prototype.description = "";

        /**
         * Element select_values.
         * @member {Array.<string>} select_values
         * @memberof mars.Element
         * @instance
         */
        Element.prototype.select_values = $util.emptyArray;

        /**
         * Encodes the specified Element message. Does not implicitly {@link mars.Element.verify|verify} messages.
         * @function encode
         * @memberof mars.Element
         * @static
         * @param {mars.Element} message Element message or plain object to encode
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
         * @memberof mars.Element
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {mars.Element} Element
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Element.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.mars.Element();
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

    return mars;
})();

export const metrics = $root.metrics = (() => {

    /**
     * Namespace metrics.
     * @exports metrics
     * @namespace
     */
    const metrics = {};

    metrics.TopPodRequest = (function() {

        /**
         * Properties of a TopPodRequest.
         * @memberof metrics
         * @interface ITopPodRequest
         * @property {string|null} [namespace] TopPodRequest namespace
         * @property {string|null} [pod] TopPodRequest pod
         */

        /**
         * Constructs a new TopPodRequest.
         * @memberof metrics
         * @classdesc Represents a TopPodRequest.
         * @implements ITopPodRequest
         * @constructor
         * @param {metrics.ITopPodRequest=} [properties] Properties to set
         */
        function TopPodRequest(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * TopPodRequest namespace.
         * @member {string} namespace
         * @memberof metrics.TopPodRequest
         * @instance
         */
        TopPodRequest.prototype.namespace = "";

        /**
         * TopPodRequest pod.
         * @member {string} pod
         * @memberof metrics.TopPodRequest
         * @instance
         */
        TopPodRequest.prototype.pod = "";

        /**
         * Encodes the specified TopPodRequest message. Does not implicitly {@link metrics.TopPodRequest.verify|verify} messages.
         * @function encode
         * @memberof metrics.TopPodRequest
         * @static
         * @param {metrics.TopPodRequest} message TopPodRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        TopPodRequest.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.namespace != null && Object.hasOwnProperty.call(message, "namespace"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.namespace);
            if (message.pod != null && Object.hasOwnProperty.call(message, "pod"))
                writer.uint32(/* id 2, wireType 2 =*/18).string(message.pod);
            return writer;
        };

        /**
         * Decodes a TopPodRequest message from the specified reader or buffer.
         * @function decode
         * @memberof metrics.TopPodRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {metrics.TopPodRequest} TopPodRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        TopPodRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.metrics.TopPodRequest();
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

        return TopPodRequest;
    })();

    metrics.TopPodResponse = (function() {

        /**
         * Properties of a TopPodResponse.
         * @memberof metrics
         * @interface ITopPodResponse
         * @property {number|null} [cpu] TopPodResponse cpu
         * @property {number|null} [memory] TopPodResponse memory
         * @property {string|null} [humanize_cpu] TopPodResponse humanize_cpu
         * @property {string|null} [humanize_memory] TopPodResponse humanize_memory
         * @property {string|null} [time] TopPodResponse time
         * @property {number|null} [length] TopPodResponse length
         */

        /**
         * Constructs a new TopPodResponse.
         * @memberof metrics
         * @classdesc Represents a TopPodResponse.
         * @implements ITopPodResponse
         * @constructor
         * @param {metrics.ITopPodResponse=} [properties] Properties to set
         */
        function TopPodResponse(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * TopPodResponse cpu.
         * @member {number} cpu
         * @memberof metrics.TopPodResponse
         * @instance
         */
        TopPodResponse.prototype.cpu = 0;

        /**
         * TopPodResponse memory.
         * @member {number} memory
         * @memberof metrics.TopPodResponse
         * @instance
         */
        TopPodResponse.prototype.memory = 0;

        /**
         * TopPodResponse humanize_cpu.
         * @member {string} humanize_cpu
         * @memberof metrics.TopPodResponse
         * @instance
         */
        TopPodResponse.prototype.humanize_cpu = "";

        /**
         * TopPodResponse humanize_memory.
         * @member {string} humanize_memory
         * @memberof metrics.TopPodResponse
         * @instance
         */
        TopPodResponse.prototype.humanize_memory = "";

        /**
         * TopPodResponse time.
         * @member {string} time
         * @memberof metrics.TopPodResponse
         * @instance
         */
        TopPodResponse.prototype.time = "";

        /**
         * TopPodResponse length.
         * @member {number} length
         * @memberof metrics.TopPodResponse
         * @instance
         */
        TopPodResponse.prototype.length = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Encodes the specified TopPodResponse message. Does not implicitly {@link metrics.TopPodResponse.verify|verify} messages.
         * @function encode
         * @memberof metrics.TopPodResponse
         * @static
         * @param {metrics.TopPodResponse} message TopPodResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        TopPodResponse.encode = function encode(message, writer) {
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
         * Decodes a TopPodResponse message from the specified reader or buffer.
         * @function decode
         * @memberof metrics.TopPodResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {metrics.TopPodResponse} TopPodResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        TopPodResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.metrics.TopPodResponse();
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

        return TopPodResponse;
    })();

    metrics.CpuMemoryInNamespaceRequest = (function() {

        /**
         * Properties of a CpuMemoryInNamespaceRequest.
         * @memberof metrics
         * @interface ICpuMemoryInNamespaceRequest
         * @property {number|null} [namespace_id] CpuMemoryInNamespaceRequest namespace_id
         */

        /**
         * Constructs a new CpuMemoryInNamespaceRequest.
         * @memberof metrics
         * @classdesc Represents a CpuMemoryInNamespaceRequest.
         * @implements ICpuMemoryInNamespaceRequest
         * @constructor
         * @param {metrics.ICpuMemoryInNamespaceRequest=} [properties] Properties to set
         */
        function CpuMemoryInNamespaceRequest(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * CpuMemoryInNamespaceRequest namespace_id.
         * @member {number} namespace_id
         * @memberof metrics.CpuMemoryInNamespaceRequest
         * @instance
         */
        CpuMemoryInNamespaceRequest.prototype.namespace_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Encodes the specified CpuMemoryInNamespaceRequest message. Does not implicitly {@link metrics.CpuMemoryInNamespaceRequest.verify|verify} messages.
         * @function encode
         * @memberof metrics.CpuMemoryInNamespaceRequest
         * @static
         * @param {metrics.CpuMemoryInNamespaceRequest} message CpuMemoryInNamespaceRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        CpuMemoryInNamespaceRequest.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.namespace_id != null && Object.hasOwnProperty.call(message, "namespace_id"))
                writer.uint32(/* id 1, wireType 0 =*/8).int64(message.namespace_id);
            return writer;
        };

        /**
         * Decodes a CpuMemoryInNamespaceRequest message from the specified reader or buffer.
         * @function decode
         * @memberof metrics.CpuMemoryInNamespaceRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {metrics.CpuMemoryInNamespaceRequest} CpuMemoryInNamespaceRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        CpuMemoryInNamespaceRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.metrics.CpuMemoryInNamespaceRequest();
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

        return CpuMemoryInNamespaceRequest;
    })();

    metrics.CpuMemoryInNamespaceResponse = (function() {

        /**
         * Properties of a CpuMemoryInNamespaceResponse.
         * @memberof metrics
         * @interface ICpuMemoryInNamespaceResponse
         * @property {string|null} [cpu] CpuMemoryInNamespaceResponse cpu
         * @property {string|null} [memory] CpuMemoryInNamespaceResponse memory
         */

        /**
         * Constructs a new CpuMemoryInNamespaceResponse.
         * @memberof metrics
         * @classdesc Represents a CpuMemoryInNamespaceResponse.
         * @implements ICpuMemoryInNamespaceResponse
         * @constructor
         * @param {metrics.ICpuMemoryInNamespaceResponse=} [properties] Properties to set
         */
        function CpuMemoryInNamespaceResponse(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * CpuMemoryInNamespaceResponse cpu.
         * @member {string} cpu
         * @memberof metrics.CpuMemoryInNamespaceResponse
         * @instance
         */
        CpuMemoryInNamespaceResponse.prototype.cpu = "";

        /**
         * CpuMemoryInNamespaceResponse memory.
         * @member {string} memory
         * @memberof metrics.CpuMemoryInNamespaceResponse
         * @instance
         */
        CpuMemoryInNamespaceResponse.prototype.memory = "";

        /**
         * Encodes the specified CpuMemoryInNamespaceResponse message. Does not implicitly {@link metrics.CpuMemoryInNamespaceResponse.verify|verify} messages.
         * @function encode
         * @memberof metrics.CpuMemoryInNamespaceResponse
         * @static
         * @param {metrics.CpuMemoryInNamespaceResponse} message CpuMemoryInNamespaceResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        CpuMemoryInNamespaceResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.cpu != null && Object.hasOwnProperty.call(message, "cpu"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.cpu);
            if (message.memory != null && Object.hasOwnProperty.call(message, "memory"))
                writer.uint32(/* id 2, wireType 2 =*/18).string(message.memory);
            return writer;
        };

        /**
         * Decodes a CpuMemoryInNamespaceResponse message from the specified reader or buffer.
         * @function decode
         * @memberof metrics.CpuMemoryInNamespaceResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {metrics.CpuMemoryInNamespaceResponse} CpuMemoryInNamespaceResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        CpuMemoryInNamespaceResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.metrics.CpuMemoryInNamespaceResponse();
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

        return CpuMemoryInNamespaceResponse;
    })();

    metrics.CpuMemoryInProjectRequest = (function() {

        /**
         * Properties of a CpuMemoryInProjectRequest.
         * @memberof metrics
         * @interface ICpuMemoryInProjectRequest
         * @property {number|null} [project_id] CpuMemoryInProjectRequest project_id
         */

        /**
         * Constructs a new CpuMemoryInProjectRequest.
         * @memberof metrics
         * @classdesc Represents a CpuMemoryInProjectRequest.
         * @implements ICpuMemoryInProjectRequest
         * @constructor
         * @param {metrics.ICpuMemoryInProjectRequest=} [properties] Properties to set
         */
        function CpuMemoryInProjectRequest(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * CpuMemoryInProjectRequest project_id.
         * @member {number} project_id
         * @memberof metrics.CpuMemoryInProjectRequest
         * @instance
         */
        CpuMemoryInProjectRequest.prototype.project_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Encodes the specified CpuMemoryInProjectRequest message. Does not implicitly {@link metrics.CpuMemoryInProjectRequest.verify|verify} messages.
         * @function encode
         * @memberof metrics.CpuMemoryInProjectRequest
         * @static
         * @param {metrics.CpuMemoryInProjectRequest} message CpuMemoryInProjectRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        CpuMemoryInProjectRequest.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.project_id != null && Object.hasOwnProperty.call(message, "project_id"))
                writer.uint32(/* id 1, wireType 0 =*/8).int64(message.project_id);
            return writer;
        };

        /**
         * Decodes a CpuMemoryInProjectRequest message from the specified reader or buffer.
         * @function decode
         * @memberof metrics.CpuMemoryInProjectRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {metrics.CpuMemoryInProjectRequest} CpuMemoryInProjectRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        CpuMemoryInProjectRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.metrics.CpuMemoryInProjectRequest();
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

        return CpuMemoryInProjectRequest;
    })();

    metrics.CpuMemoryInProjectResponse = (function() {

        /**
         * Properties of a CpuMemoryInProjectResponse.
         * @memberof metrics
         * @interface ICpuMemoryInProjectResponse
         * @property {string|null} [cpu] CpuMemoryInProjectResponse cpu
         * @property {string|null} [memory] CpuMemoryInProjectResponse memory
         */

        /**
         * Constructs a new CpuMemoryInProjectResponse.
         * @memberof metrics
         * @classdesc Represents a CpuMemoryInProjectResponse.
         * @implements ICpuMemoryInProjectResponse
         * @constructor
         * @param {metrics.ICpuMemoryInProjectResponse=} [properties] Properties to set
         */
        function CpuMemoryInProjectResponse(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * CpuMemoryInProjectResponse cpu.
         * @member {string} cpu
         * @memberof metrics.CpuMemoryInProjectResponse
         * @instance
         */
        CpuMemoryInProjectResponse.prototype.cpu = "";

        /**
         * CpuMemoryInProjectResponse memory.
         * @member {string} memory
         * @memberof metrics.CpuMemoryInProjectResponse
         * @instance
         */
        CpuMemoryInProjectResponse.prototype.memory = "";

        /**
         * Encodes the specified CpuMemoryInProjectResponse message. Does not implicitly {@link metrics.CpuMemoryInProjectResponse.verify|verify} messages.
         * @function encode
         * @memberof metrics.CpuMemoryInProjectResponse
         * @static
         * @param {metrics.CpuMemoryInProjectResponse} message CpuMemoryInProjectResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        CpuMemoryInProjectResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.cpu != null && Object.hasOwnProperty.call(message, "cpu"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.cpu);
            if (message.memory != null && Object.hasOwnProperty.call(message, "memory"))
                writer.uint32(/* id 2, wireType 2 =*/18).string(message.memory);
            return writer;
        };

        /**
         * Decodes a CpuMemoryInProjectResponse message from the specified reader or buffer.
         * @function decode
         * @memberof metrics.CpuMemoryInProjectResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {metrics.CpuMemoryInProjectResponse} CpuMemoryInProjectResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        CpuMemoryInProjectResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.metrics.CpuMemoryInProjectResponse();
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

        return CpuMemoryInProjectResponse;
    })();

    metrics.Metrics = (function() {

        /**
         * Constructs a new Metrics service.
         * @memberof metrics
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
         * Callback as used by {@link metrics.Metrics#cpuMemoryInNamespace}.
         * @memberof metrics.Metrics
         * @typedef CpuMemoryInNamespaceCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {metrics.CpuMemoryInNamespaceResponse} [response] CpuMemoryInNamespaceResponse
         */

        /**
         * Calls CpuMemoryInNamespace.
         * @function cpuMemoryInNamespace
         * @memberof metrics.Metrics
         * @instance
         * @param {metrics.CpuMemoryInNamespaceRequest} request CpuMemoryInNamespaceRequest message or plain object
         * @param {metrics.Metrics.CpuMemoryInNamespaceCallback} callback Node-style callback called with the error, if any, and CpuMemoryInNamespaceResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(Metrics.prototype.cpuMemoryInNamespace = function cpuMemoryInNamespace(request, callback) {
            return this.rpcCall(cpuMemoryInNamespace, $root.metrics.CpuMemoryInNamespaceRequest, $root.metrics.CpuMemoryInNamespaceResponse, request, callback);
        }, "name", { value: "CpuMemoryInNamespace" });

        /**
         * Calls CpuMemoryInNamespace.
         * @function cpuMemoryInNamespace
         * @memberof metrics.Metrics
         * @instance
         * @param {metrics.CpuMemoryInNamespaceRequest} request CpuMemoryInNamespaceRequest message or plain object
         * @returns {Promise<metrics.CpuMemoryInNamespaceResponse>} Promise
         * @variation 2
         */

        /**
         * Callback as used by {@link metrics.Metrics#cpuMemoryInProject}.
         * @memberof metrics.Metrics
         * @typedef CpuMemoryInProjectCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {metrics.CpuMemoryInProjectResponse} [response] CpuMemoryInProjectResponse
         */

        /**
         * Calls CpuMemoryInProject.
         * @function cpuMemoryInProject
         * @memberof metrics.Metrics
         * @instance
         * @param {metrics.CpuMemoryInProjectRequest} request CpuMemoryInProjectRequest message or plain object
         * @param {metrics.Metrics.CpuMemoryInProjectCallback} callback Node-style callback called with the error, if any, and CpuMemoryInProjectResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(Metrics.prototype.cpuMemoryInProject = function cpuMemoryInProject(request, callback) {
            return this.rpcCall(cpuMemoryInProject, $root.metrics.CpuMemoryInProjectRequest, $root.metrics.CpuMemoryInProjectResponse, request, callback);
        }, "name", { value: "CpuMemoryInProject" });

        /**
         * Calls CpuMemoryInProject.
         * @function cpuMemoryInProject
         * @memberof metrics.Metrics
         * @instance
         * @param {metrics.CpuMemoryInProjectRequest} request CpuMemoryInProjectRequest message or plain object
         * @returns {Promise<metrics.CpuMemoryInProjectResponse>} Promise
         * @variation 2
         */

        /**
         * Callback as used by {@link metrics.Metrics#topPod}.
         * @memberof metrics.Metrics
         * @typedef TopPodCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {metrics.TopPodResponse} [response] TopPodResponse
         */

        /**
         * Calls TopPod.
         * @function topPod
         * @memberof metrics.Metrics
         * @instance
         * @param {metrics.TopPodRequest} request TopPodRequest message or plain object
         * @param {metrics.Metrics.TopPodCallback} callback Node-style callback called with the error, if any, and TopPodResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(Metrics.prototype.topPod = function topPod(request, callback) {
            return this.rpcCall(topPod, $root.metrics.TopPodRequest, $root.metrics.TopPodResponse, request, callback);
        }, "name", { value: "TopPod" });

        /**
         * Calls TopPod.
         * @function topPod
         * @memberof metrics.Metrics
         * @instance
         * @param {metrics.TopPodRequest} request TopPodRequest message or plain object
         * @returns {Promise<metrics.TopPodResponse>} Promise
         * @variation 2
         */

        /**
         * Callback as used by {@link metrics.Metrics#streamTopPod}.
         * @memberof metrics.Metrics
         * @typedef StreamTopPodCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {metrics.TopPodResponse} [response] TopPodResponse
         */

        /**
         * Calls StreamTopPod.
         * @function streamTopPod
         * @memberof metrics.Metrics
         * @instance
         * @param {metrics.TopPodRequest} request TopPodRequest message or plain object
         * @param {metrics.Metrics.StreamTopPodCallback} callback Node-style callback called with the error, if any, and TopPodResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(Metrics.prototype.streamTopPod = function streamTopPod(request, callback) {
            return this.rpcCall(streamTopPod, $root.metrics.TopPodRequest, $root.metrics.TopPodResponse, request, callback);
        }, "name", { value: "StreamTopPod" });

        /**
         * Calls StreamTopPod.
         * @function streamTopPod
         * @memberof metrics.Metrics
         * @instance
         * @param {metrics.TopPodRequest} request TopPodRequest message or plain object
         * @returns {Promise<metrics.TopPodResponse>} Promise
         * @variation 2
         */

        return Metrics;
    })();

    return metrics;
})();

export const namespace = $root.namespace = (() => {

    /**
     * Namespace namespace.
     * @exports namespace
     * @namespace
     */
    const namespace = {};

    namespace.CreateRequest = (function() {

        /**
         * Properties of a CreateRequest.
         * @memberof namespace
         * @interface ICreateRequest
         * @property {string|null} [namespace] CreateRequest namespace
         * @property {boolean|null} [ignore_if_exists] CreateRequest ignore_if_exists
         */

        /**
         * Constructs a new CreateRequest.
         * @memberof namespace
         * @classdesc Represents a CreateRequest.
         * @implements ICreateRequest
         * @constructor
         * @param {namespace.ICreateRequest=} [properties] Properties to set
         */
        function CreateRequest(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * CreateRequest namespace.
         * @member {string} namespace
         * @memberof namespace.CreateRequest
         * @instance
         */
        CreateRequest.prototype.namespace = "";

        /**
         * CreateRequest ignore_if_exists.
         * @member {boolean} ignore_if_exists
         * @memberof namespace.CreateRequest
         * @instance
         */
        CreateRequest.prototype.ignore_if_exists = false;

        /**
         * Encodes the specified CreateRequest message. Does not implicitly {@link namespace.CreateRequest.verify|verify} messages.
         * @function encode
         * @memberof namespace.CreateRequest
         * @static
         * @param {namespace.CreateRequest} message CreateRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        CreateRequest.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.namespace != null && Object.hasOwnProperty.call(message, "namespace"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.namespace);
            if (message.ignore_if_exists != null && Object.hasOwnProperty.call(message, "ignore_if_exists"))
                writer.uint32(/* id 2, wireType 0 =*/16).bool(message.ignore_if_exists);
            return writer;
        };

        /**
         * Decodes a CreateRequest message from the specified reader or buffer.
         * @function decode
         * @memberof namespace.CreateRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {namespace.CreateRequest} CreateRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        CreateRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.namespace.CreateRequest();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1:
                    message.namespace = reader.string();
                    break;
                case 2:
                    message.ignore_if_exists = reader.bool();
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        return CreateRequest;
    })();

    namespace.ShowRequest = (function() {

        /**
         * Properties of a ShowRequest.
         * @memberof namespace
         * @interface IShowRequest
         * @property {number|null} [namespace_id] ShowRequest namespace_id
         */

        /**
         * Constructs a new ShowRequest.
         * @memberof namespace
         * @classdesc Represents a ShowRequest.
         * @implements IShowRequest
         * @constructor
         * @param {namespace.IShowRequest=} [properties] Properties to set
         */
        function ShowRequest(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * ShowRequest namespace_id.
         * @member {number} namespace_id
         * @memberof namespace.ShowRequest
         * @instance
         */
        ShowRequest.prototype.namespace_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Encodes the specified ShowRequest message. Does not implicitly {@link namespace.ShowRequest.verify|verify} messages.
         * @function encode
         * @memberof namespace.ShowRequest
         * @static
         * @param {namespace.ShowRequest} message ShowRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ShowRequest.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.namespace_id != null && Object.hasOwnProperty.call(message, "namespace_id"))
                writer.uint32(/* id 1, wireType 0 =*/8).int64(message.namespace_id);
            return writer;
        };

        /**
         * Decodes a ShowRequest message from the specified reader or buffer.
         * @function decode
         * @memberof namespace.ShowRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {namespace.ShowRequest} ShowRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ShowRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.namespace.ShowRequest();
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

        return ShowRequest;
    })();

    namespace.DeleteRequest = (function() {

        /**
         * Properties of a DeleteRequest.
         * @memberof namespace
         * @interface IDeleteRequest
         * @property {number|null} [namespace_id] DeleteRequest namespace_id
         */

        /**
         * Constructs a new DeleteRequest.
         * @memberof namespace
         * @classdesc Represents a DeleteRequest.
         * @implements IDeleteRequest
         * @constructor
         * @param {namespace.IDeleteRequest=} [properties] Properties to set
         */
        function DeleteRequest(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * DeleteRequest namespace_id.
         * @member {number} namespace_id
         * @memberof namespace.DeleteRequest
         * @instance
         */
        DeleteRequest.prototype.namespace_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Encodes the specified DeleteRequest message. Does not implicitly {@link namespace.DeleteRequest.verify|verify} messages.
         * @function encode
         * @memberof namespace.DeleteRequest
         * @static
         * @param {namespace.DeleteRequest} message DeleteRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        DeleteRequest.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.namespace_id != null && Object.hasOwnProperty.call(message, "namespace_id"))
                writer.uint32(/* id 1, wireType 0 =*/8).int64(message.namespace_id);
            return writer;
        };

        /**
         * Decodes a DeleteRequest message from the specified reader or buffer.
         * @function decode
         * @memberof namespace.DeleteRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {namespace.DeleteRequest} DeleteRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        DeleteRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.namespace.DeleteRequest();
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

        return DeleteRequest;
    })();

    namespace.IsExistsRequest = (function() {

        /**
         * Properties of an IsExistsRequest.
         * @memberof namespace
         * @interface IIsExistsRequest
         * @property {string|null} [name] IsExistsRequest name
         */

        /**
         * Constructs a new IsExistsRequest.
         * @memberof namespace
         * @classdesc Represents an IsExistsRequest.
         * @implements IIsExistsRequest
         * @constructor
         * @param {namespace.IIsExistsRequest=} [properties] Properties to set
         */
        function IsExistsRequest(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * IsExistsRequest name.
         * @member {string} name
         * @memberof namespace.IsExistsRequest
         * @instance
         */
        IsExistsRequest.prototype.name = "";

        /**
         * Encodes the specified IsExistsRequest message. Does not implicitly {@link namespace.IsExistsRequest.verify|verify} messages.
         * @function encode
         * @memberof namespace.IsExistsRequest
         * @static
         * @param {namespace.IsExistsRequest} message IsExistsRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        IsExistsRequest.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.name != null && Object.hasOwnProperty.call(message, "name"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.name);
            return writer;
        };

        /**
         * Decodes an IsExistsRequest message from the specified reader or buffer.
         * @function decode
         * @memberof namespace.IsExistsRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {namespace.IsExistsRequest} IsExistsRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        IsExistsRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.namespace.IsExistsRequest();
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

        return IsExistsRequest;
    })();

    namespace.AllResponse = (function() {

        /**
         * Properties of an AllResponse.
         * @memberof namespace
         * @interface IAllResponse
         * @property {Array.<types.NamespaceModel>|null} [items] AllResponse items
         */

        /**
         * Constructs a new AllResponse.
         * @memberof namespace
         * @classdesc Represents an AllResponse.
         * @implements IAllResponse
         * @constructor
         * @param {namespace.IAllResponse=} [properties] Properties to set
         */
        function AllResponse(properties) {
            this.items = [];
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * AllResponse items.
         * @member {Array.<types.NamespaceModel>} items
         * @memberof namespace.AllResponse
         * @instance
         */
        AllResponse.prototype.items = $util.emptyArray;

        /**
         * Encodes the specified AllResponse message. Does not implicitly {@link namespace.AllResponse.verify|verify} messages.
         * @function encode
         * @memberof namespace.AllResponse
         * @static
         * @param {namespace.AllResponse} message AllResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        AllResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.items != null && message.items.length)
                for (let i = 0; i < message.items.length; ++i)
                    $root.types.NamespaceModel.encode(message.items[i], writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
            return writer;
        };

        /**
         * Decodes an AllResponse message from the specified reader or buffer.
         * @function decode
         * @memberof namespace.AllResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {namespace.AllResponse} AllResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        AllResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.namespace.AllResponse();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1:
                    if (!(message.items && message.items.length))
                        message.items = [];
                    message.items.push($root.types.NamespaceModel.decode(reader, reader.uint32()));
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        return AllResponse;
    })();

    namespace.CreateResponse = (function() {

        /**
         * Properties of a CreateResponse.
         * @memberof namespace
         * @interface ICreateResponse
         * @property {types.NamespaceModel|null} [namespace] CreateResponse namespace
         * @property {boolean|null} [exists] CreateResponse exists
         */

        /**
         * Constructs a new CreateResponse.
         * @memberof namespace
         * @classdesc Represents a CreateResponse.
         * @implements ICreateResponse
         * @constructor
         * @param {namespace.ICreateResponse=} [properties] Properties to set
         */
        function CreateResponse(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * CreateResponse namespace.
         * @member {types.NamespaceModel|null|undefined} namespace
         * @memberof namespace.CreateResponse
         * @instance
         */
        CreateResponse.prototype.namespace = null;

        /**
         * CreateResponse exists.
         * @member {boolean} exists
         * @memberof namespace.CreateResponse
         * @instance
         */
        CreateResponse.prototype.exists = false;

        /**
         * Encodes the specified CreateResponse message. Does not implicitly {@link namespace.CreateResponse.verify|verify} messages.
         * @function encode
         * @memberof namespace.CreateResponse
         * @static
         * @param {namespace.CreateResponse} message CreateResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        CreateResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.namespace != null && Object.hasOwnProperty.call(message, "namespace"))
                $root.types.NamespaceModel.encode(message.namespace, writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
            if (message.exists != null && Object.hasOwnProperty.call(message, "exists"))
                writer.uint32(/* id 2, wireType 0 =*/16).bool(message.exists);
            return writer;
        };

        /**
         * Decodes a CreateResponse message from the specified reader or buffer.
         * @function decode
         * @memberof namespace.CreateResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {namespace.CreateResponse} CreateResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        CreateResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.namespace.CreateResponse();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1:
                    message.namespace = $root.types.NamespaceModel.decode(reader, reader.uint32());
                    break;
                case 2:
                    message.exists = reader.bool();
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        return CreateResponse;
    })();

    namespace.ShowResponse = (function() {

        /**
         * Properties of a ShowResponse.
         * @memberof namespace
         * @interface IShowResponse
         * @property {types.NamespaceModel|null} [namespace] ShowResponse namespace
         */

        /**
         * Constructs a new ShowResponse.
         * @memberof namespace
         * @classdesc Represents a ShowResponse.
         * @implements IShowResponse
         * @constructor
         * @param {namespace.IShowResponse=} [properties] Properties to set
         */
        function ShowResponse(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * ShowResponse namespace.
         * @member {types.NamespaceModel|null|undefined} namespace
         * @memberof namespace.ShowResponse
         * @instance
         */
        ShowResponse.prototype.namespace = null;

        /**
         * Encodes the specified ShowResponse message. Does not implicitly {@link namespace.ShowResponse.verify|verify} messages.
         * @function encode
         * @memberof namespace.ShowResponse
         * @static
         * @param {namespace.ShowResponse} message ShowResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ShowResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.namespace != null && Object.hasOwnProperty.call(message, "namespace"))
                $root.types.NamespaceModel.encode(message.namespace, writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
            return writer;
        };

        /**
         * Decodes a ShowResponse message from the specified reader or buffer.
         * @function decode
         * @memberof namespace.ShowResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {namespace.ShowResponse} ShowResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ShowResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.namespace.ShowResponse();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1:
                    message.namespace = $root.types.NamespaceModel.decode(reader, reader.uint32());
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        return ShowResponse;
    })();

    namespace.IsExistsResponse = (function() {

        /**
         * Properties of an IsExistsResponse.
         * @memberof namespace
         * @interface IIsExistsResponse
         * @property {boolean|null} [exists] IsExistsResponse exists
         * @property {number|null} [id] IsExistsResponse id
         */

        /**
         * Constructs a new IsExistsResponse.
         * @memberof namespace
         * @classdesc Represents an IsExistsResponse.
         * @implements IIsExistsResponse
         * @constructor
         * @param {namespace.IIsExistsResponse=} [properties] Properties to set
         */
        function IsExistsResponse(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * IsExistsResponse exists.
         * @member {boolean} exists
         * @memberof namespace.IsExistsResponse
         * @instance
         */
        IsExistsResponse.prototype.exists = false;

        /**
         * IsExistsResponse id.
         * @member {number} id
         * @memberof namespace.IsExistsResponse
         * @instance
         */
        IsExistsResponse.prototype.id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Encodes the specified IsExistsResponse message. Does not implicitly {@link namespace.IsExistsResponse.verify|verify} messages.
         * @function encode
         * @memberof namespace.IsExistsResponse
         * @static
         * @param {namespace.IsExistsResponse} message IsExistsResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        IsExistsResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.exists != null && Object.hasOwnProperty.call(message, "exists"))
                writer.uint32(/* id 1, wireType 0 =*/8).bool(message.exists);
            if (message.id != null && Object.hasOwnProperty.call(message, "id"))
                writer.uint32(/* id 2, wireType 0 =*/16).int64(message.id);
            return writer;
        };

        /**
         * Decodes an IsExistsResponse message from the specified reader or buffer.
         * @function decode
         * @memberof namespace.IsExistsResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {namespace.IsExistsResponse} IsExistsResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        IsExistsResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.namespace.IsExistsResponse();
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

        return IsExistsResponse;
    })();

    namespace.AllRequest = (function() {

        /**
         * Properties of an AllRequest.
         * @memberof namespace
         * @interface IAllRequest
         */

        /**
         * Constructs a new AllRequest.
         * @memberof namespace
         * @classdesc Represents an AllRequest.
         * @implements IAllRequest
         * @constructor
         * @param {namespace.IAllRequest=} [properties] Properties to set
         */
        function AllRequest(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * Encodes the specified AllRequest message. Does not implicitly {@link namespace.AllRequest.verify|verify} messages.
         * @function encode
         * @memberof namespace.AllRequest
         * @static
         * @param {namespace.AllRequest} message AllRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        AllRequest.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            return writer;
        };

        /**
         * Decodes an AllRequest message from the specified reader or buffer.
         * @function decode
         * @memberof namespace.AllRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {namespace.AllRequest} AllRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        AllRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.namespace.AllRequest();
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

        return AllRequest;
    })();

    namespace.DeleteResponse = (function() {

        /**
         * Properties of a DeleteResponse.
         * @memberof namespace
         * @interface IDeleteResponse
         */

        /**
         * Constructs a new DeleteResponse.
         * @memberof namespace
         * @classdesc Represents a DeleteResponse.
         * @implements IDeleteResponse
         * @constructor
         * @param {namespace.IDeleteResponse=} [properties] Properties to set
         */
        function DeleteResponse(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * Encodes the specified DeleteResponse message. Does not implicitly {@link namespace.DeleteResponse.verify|verify} messages.
         * @function encode
         * @memberof namespace.DeleteResponse
         * @static
         * @param {namespace.DeleteResponse} message DeleteResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        DeleteResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            return writer;
        };

        /**
         * Decodes a DeleteResponse message from the specified reader or buffer.
         * @function decode
         * @memberof namespace.DeleteResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {namespace.DeleteResponse} DeleteResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        DeleteResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.namespace.DeleteResponse();
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

        return DeleteResponse;
    })();

    namespace.Namespace = (function() {

        /**
         * Constructs a new Namespace service.
         * @memberof namespace
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
         * Callback as used by {@link namespace.Namespace#all}.
         * @memberof namespace.Namespace
         * @typedef AllCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {namespace.AllResponse} [response] AllResponse
         */

        /**
         * Calls All.
         * @function all
         * @memberof namespace.Namespace
         * @instance
         * @param {namespace.AllRequest} request AllRequest message or plain object
         * @param {namespace.Namespace.AllCallback} callback Node-style callback called with the error, if any, and AllResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(Namespace.prototype.all = function all(request, callback) {
            return this.rpcCall(all, $root.namespace.AllRequest, $root.namespace.AllResponse, request, callback);
        }, "name", { value: "All" });

        /**
         * Calls All.
         * @function all
         * @memberof namespace.Namespace
         * @instance
         * @param {namespace.AllRequest} request AllRequest message or plain object
         * @returns {Promise<namespace.AllResponse>} Promise
         * @variation 2
         */

        /**
         * Callback as used by {@link namespace.Namespace#create}.
         * @memberof namespace.Namespace
         * @typedef CreateCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {namespace.CreateResponse} [response] CreateResponse
         */

        /**
         * Calls Create.
         * @function create
         * @memberof namespace.Namespace
         * @instance
         * @param {namespace.CreateRequest} request CreateRequest message or plain object
         * @param {namespace.Namespace.CreateCallback} callback Node-style callback called with the error, if any, and CreateResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(Namespace.prototype.create = function create(request, callback) {
            return this.rpcCall(create, $root.namespace.CreateRequest, $root.namespace.CreateResponse, request, callback);
        }, "name", { value: "Create" });

        /**
         * Calls Create.
         * @function create
         * @memberof namespace.Namespace
         * @instance
         * @param {namespace.CreateRequest} request CreateRequest message or plain object
         * @returns {Promise<namespace.CreateResponse>} Promise
         * @variation 2
         */

        /**
         * Callback as used by {@link namespace.Namespace#show}.
         * @memberof namespace.Namespace
         * @typedef ShowCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {namespace.ShowResponse} [response] ShowResponse
         */

        /**
         * Calls Show.
         * @function show
         * @memberof namespace.Namespace
         * @instance
         * @param {namespace.ShowRequest} request ShowRequest message or plain object
         * @param {namespace.Namespace.ShowCallback} callback Node-style callback called with the error, if any, and ShowResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(Namespace.prototype.show = function show(request, callback) {
            return this.rpcCall(show, $root.namespace.ShowRequest, $root.namespace.ShowResponse, request, callback);
        }, "name", { value: "Show" });

        /**
         * Calls Show.
         * @function show
         * @memberof namespace.Namespace
         * @instance
         * @param {namespace.ShowRequest} request ShowRequest message or plain object
         * @returns {Promise<namespace.ShowResponse>} Promise
         * @variation 2
         */

        /**
         * Callback as used by {@link namespace.Namespace#delete_}.
         * @memberof namespace.Namespace
         * @typedef DeleteCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {namespace.DeleteResponse} [response] DeleteResponse
         */

        /**
         * Calls Delete.
         * @function delete
         * @memberof namespace.Namespace
         * @instance
         * @param {namespace.DeleteRequest} request DeleteRequest message or plain object
         * @param {namespace.Namespace.DeleteCallback} callback Node-style callback called with the error, if any, and DeleteResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(Namespace.prototype["delete"] = function delete_(request, callback) {
            return this.rpcCall(delete_, $root.namespace.DeleteRequest, $root.namespace.DeleteResponse, request, callback);
        }, "name", { value: "Delete" });

        /**
         * Calls Delete.
         * @function delete
         * @memberof namespace.Namespace
         * @instance
         * @param {namespace.DeleteRequest} request DeleteRequest message or plain object
         * @returns {Promise<namespace.DeleteResponse>} Promise
         * @variation 2
         */

        /**
         * Callback as used by {@link namespace.Namespace#isExists}.
         * @memberof namespace.Namespace
         * @typedef IsExistsCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {namespace.IsExistsResponse} [response] IsExistsResponse
         */

        /**
         * Calls IsExists.
         * @function isExists
         * @memberof namespace.Namespace
         * @instance
         * @param {namespace.IsExistsRequest} request IsExistsRequest message or plain object
         * @param {namespace.Namespace.IsExistsCallback} callback Node-style callback called with the error, if any, and IsExistsResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(Namespace.prototype.isExists = function isExists(request, callback) {
            return this.rpcCall(isExists, $root.namespace.IsExistsRequest, $root.namespace.IsExistsResponse, request, callback);
        }, "name", { value: "IsExists" });

        /**
         * Calls IsExists.
         * @function isExists
         * @memberof namespace.Namespace
         * @instance
         * @param {namespace.IsExistsRequest} request IsExistsRequest message or plain object
         * @returns {Promise<namespace.IsExistsResponse>} Promise
         * @variation 2
         */

        return Namespace;
    })();

    return namespace;
})();

export const picture = $root.picture = (() => {

    /**
     * Namespace picture.
     * @exports picture
     * @namespace
     */
    const picture = {};

    picture.BackgroundRequest = (function() {

        /**
         * Properties of a BackgroundRequest.
         * @memberof picture
         * @interface IBackgroundRequest
         * @property {boolean|null} [random] BackgroundRequest random
         */

        /**
         * Constructs a new BackgroundRequest.
         * @memberof picture
         * @classdesc Represents a BackgroundRequest.
         * @implements IBackgroundRequest
         * @constructor
         * @param {picture.IBackgroundRequest=} [properties] Properties to set
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
         * @memberof picture.BackgroundRequest
         * @instance
         */
        BackgroundRequest.prototype.random = false;

        /**
         * Encodes the specified BackgroundRequest message. Does not implicitly {@link picture.BackgroundRequest.verify|verify} messages.
         * @function encode
         * @memberof picture.BackgroundRequest
         * @static
         * @param {picture.BackgroundRequest} message BackgroundRequest message or plain object to encode
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
         * @memberof picture.BackgroundRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {picture.BackgroundRequest} BackgroundRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        BackgroundRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.picture.BackgroundRequest();
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

    picture.BackgroundResponse = (function() {

        /**
         * Properties of a BackgroundResponse.
         * @memberof picture
         * @interface IBackgroundResponse
         * @property {string|null} [url] BackgroundResponse url
         * @property {string|null} [copyright] BackgroundResponse copyright
         */

        /**
         * Constructs a new BackgroundResponse.
         * @memberof picture
         * @classdesc Represents a BackgroundResponse.
         * @implements IBackgroundResponse
         * @constructor
         * @param {picture.IBackgroundResponse=} [properties] Properties to set
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
         * @memberof picture.BackgroundResponse
         * @instance
         */
        BackgroundResponse.prototype.url = "";

        /**
         * BackgroundResponse copyright.
         * @member {string} copyright
         * @memberof picture.BackgroundResponse
         * @instance
         */
        BackgroundResponse.prototype.copyright = "";

        /**
         * Encodes the specified BackgroundResponse message. Does not implicitly {@link picture.BackgroundResponse.verify|verify} messages.
         * @function encode
         * @memberof picture.BackgroundResponse
         * @static
         * @param {picture.BackgroundResponse} message BackgroundResponse message or plain object to encode
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
         * @memberof picture.BackgroundResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {picture.BackgroundResponse} BackgroundResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        BackgroundResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.picture.BackgroundResponse();
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

    picture.Picture = (function() {

        /**
         * Constructs a new Picture service.
         * @memberof picture
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
         * Callback as used by {@link picture.Picture#background}.
         * @memberof picture.Picture
         * @typedef BackgroundCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {picture.BackgroundResponse} [response] BackgroundResponse
         */

        /**
         * Calls Background.
         * @function background
         * @memberof picture.Picture
         * @instance
         * @param {picture.BackgroundRequest} request BackgroundRequest message or plain object
         * @param {picture.Picture.BackgroundCallback} callback Node-style callback called with the error, if any, and BackgroundResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(Picture.prototype.background = function background(request, callback) {
            return this.rpcCall(background, $root.picture.BackgroundRequest, $root.picture.BackgroundResponse, request, callback);
        }, "name", { value: "Background" });

        /**
         * Calls Background.
         * @function background
         * @memberof picture.Picture
         * @instance
         * @param {picture.BackgroundRequest} request BackgroundRequest message or plain object
         * @returns {Promise<picture.BackgroundResponse>} Promise
         * @variation 2
         */

        return Picture;
    })();

    return picture;
})();

export const project = $root.project = (() => {

    /**
     * Namespace project.
     * @exports project
     * @namespace
     */
    const project = {};

    project.DeleteRequest = (function() {

        /**
         * Properties of a DeleteRequest.
         * @memberof project
         * @interface IDeleteRequest
         * @property {number|null} [project_id] DeleteRequest project_id
         */

        /**
         * Constructs a new DeleteRequest.
         * @memberof project
         * @classdesc Represents a DeleteRequest.
         * @implements IDeleteRequest
         * @constructor
         * @param {project.IDeleteRequest=} [properties] Properties to set
         */
        function DeleteRequest(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * DeleteRequest project_id.
         * @member {number} project_id
         * @memberof project.DeleteRequest
         * @instance
         */
        DeleteRequest.prototype.project_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Encodes the specified DeleteRequest message. Does not implicitly {@link project.DeleteRequest.verify|verify} messages.
         * @function encode
         * @memberof project.DeleteRequest
         * @static
         * @param {project.DeleteRequest} message DeleteRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        DeleteRequest.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.project_id != null && Object.hasOwnProperty.call(message, "project_id"))
                writer.uint32(/* id 1, wireType 0 =*/8).int64(message.project_id);
            return writer;
        };

        /**
         * Decodes a DeleteRequest message from the specified reader or buffer.
         * @function decode
         * @memberof project.DeleteRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {project.DeleteRequest} DeleteRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        DeleteRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.project.DeleteRequest();
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

        return DeleteRequest;
    })();

    project.ShowRequest = (function() {

        /**
         * Properties of a ShowRequest.
         * @memberof project
         * @interface IShowRequest
         * @property {number|null} [project_id] ShowRequest project_id
         */

        /**
         * Constructs a new ShowRequest.
         * @memberof project
         * @classdesc Represents a ShowRequest.
         * @implements IShowRequest
         * @constructor
         * @param {project.IShowRequest=} [properties] Properties to set
         */
        function ShowRequest(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * ShowRequest project_id.
         * @member {number} project_id
         * @memberof project.ShowRequest
         * @instance
         */
        ShowRequest.prototype.project_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Encodes the specified ShowRequest message. Does not implicitly {@link project.ShowRequest.verify|verify} messages.
         * @function encode
         * @memberof project.ShowRequest
         * @static
         * @param {project.ShowRequest} message ShowRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ShowRequest.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.project_id != null && Object.hasOwnProperty.call(message, "project_id"))
                writer.uint32(/* id 1, wireType 0 =*/8).int64(message.project_id);
            return writer;
        };

        /**
         * Decodes a ShowRequest message from the specified reader or buffer.
         * @function decode
         * @memberof project.ShowRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {project.ShowRequest} ShowRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ShowRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.project.ShowRequest();
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

        return ShowRequest;
    })();

    project.ShowResponse = (function() {

        /**
         * Properties of a ShowResponse.
         * @memberof project
         * @interface IShowResponse
         * @property {types.ProjectModel|null} [project] ShowResponse project
         * @property {Array.<types.ServiceEndpoint>|null} [urls] ShowResponse urls
         * @property {string|null} [cpu] ShowResponse cpu
         * @property {string|null} [memory] ShowResponse memory
         * @property {Array.<mars.Element>|null} [elements] ShowResponse elements
         */

        /**
         * Constructs a new ShowResponse.
         * @memberof project
         * @classdesc Represents a ShowResponse.
         * @implements IShowResponse
         * @constructor
         * @param {project.IShowResponse=} [properties] Properties to set
         */
        function ShowResponse(properties) {
            this.urls = [];
            this.elements = [];
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * ShowResponse project.
         * @member {types.ProjectModel|null|undefined} project
         * @memberof project.ShowResponse
         * @instance
         */
        ShowResponse.prototype.project = null;

        /**
         * ShowResponse urls.
         * @member {Array.<types.ServiceEndpoint>} urls
         * @memberof project.ShowResponse
         * @instance
         */
        ShowResponse.prototype.urls = $util.emptyArray;

        /**
         * ShowResponse cpu.
         * @member {string} cpu
         * @memberof project.ShowResponse
         * @instance
         */
        ShowResponse.prototype.cpu = "";

        /**
         * ShowResponse memory.
         * @member {string} memory
         * @memberof project.ShowResponse
         * @instance
         */
        ShowResponse.prototype.memory = "";

        /**
         * ShowResponse elements.
         * @member {Array.<mars.Element>} elements
         * @memberof project.ShowResponse
         * @instance
         */
        ShowResponse.prototype.elements = $util.emptyArray;

        /**
         * Encodes the specified ShowResponse message. Does not implicitly {@link project.ShowResponse.verify|verify} messages.
         * @function encode
         * @memberof project.ShowResponse
         * @static
         * @param {project.ShowResponse} message ShowResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ShowResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.project != null && Object.hasOwnProperty.call(message, "project"))
                $root.types.ProjectModel.encode(message.project, writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
            if (message.urls != null && message.urls.length)
                for (let i = 0; i < message.urls.length; ++i)
                    $root.types.ServiceEndpoint.encode(message.urls[i], writer.uint32(/* id 13, wireType 2 =*/106).fork()).ldelim();
            if (message.cpu != null && Object.hasOwnProperty.call(message, "cpu"))
                writer.uint32(/* id 15, wireType 2 =*/122).string(message.cpu);
            if (message.memory != null && Object.hasOwnProperty.call(message, "memory"))
                writer.uint32(/* id 16, wireType 2 =*/130).string(message.memory);
            if (message.elements != null && message.elements.length)
                for (let i = 0; i < message.elements.length; ++i)
                    $root.mars.Element.encode(message.elements[i], writer.uint32(/* id 23, wireType 2 =*/186).fork()).ldelim();
            return writer;
        };

        /**
         * Decodes a ShowResponse message from the specified reader or buffer.
         * @function decode
         * @memberof project.ShowResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {project.ShowResponse} ShowResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ShowResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.project.ShowResponse();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1:
                    message.project = $root.types.ProjectModel.decode(reader, reader.uint32());
                    break;
                case 13:
                    if (!(message.urls && message.urls.length))
                        message.urls = [];
                    message.urls.push($root.types.ServiceEndpoint.decode(reader, reader.uint32()));
                    break;
                case 15:
                    message.cpu = reader.string();
                    break;
                case 16:
                    message.memory = reader.string();
                    break;
                case 23:
                    if (!(message.elements && message.elements.length))
                        message.elements = [];
                    message.elements.push($root.mars.Element.decode(reader, reader.uint32()));
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        return ShowResponse;
    })();

    project.AllContainersRequest = (function() {

        /**
         * Properties of an AllContainersRequest.
         * @memberof project
         * @interface IAllContainersRequest
         * @property {number|null} [project_id] AllContainersRequest project_id
         */

        /**
         * Constructs a new AllContainersRequest.
         * @memberof project
         * @classdesc Represents an AllContainersRequest.
         * @implements IAllContainersRequest
         * @constructor
         * @param {project.IAllContainersRequest=} [properties] Properties to set
         */
        function AllContainersRequest(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * AllContainersRequest project_id.
         * @member {number} project_id
         * @memberof project.AllContainersRequest
         * @instance
         */
        AllContainersRequest.prototype.project_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Encodes the specified AllContainersRequest message. Does not implicitly {@link project.AllContainersRequest.verify|verify} messages.
         * @function encode
         * @memberof project.AllContainersRequest
         * @static
         * @param {project.AllContainersRequest} message AllContainersRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        AllContainersRequest.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.project_id != null && Object.hasOwnProperty.call(message, "project_id"))
                writer.uint32(/* id 1, wireType 0 =*/8).int64(message.project_id);
            return writer;
        };

        /**
         * Decodes an AllContainersRequest message from the specified reader or buffer.
         * @function decode
         * @memberof project.AllContainersRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {project.AllContainersRequest} AllContainersRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        AllContainersRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.project.AllContainersRequest();
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

        return AllContainersRequest;
    })();

    project.AllContainersResponse = (function() {

        /**
         * Properties of an AllContainersResponse.
         * @memberof project
         * @interface IAllContainersResponse
         * @property {Array.<types.Container>|null} [items] AllContainersResponse items
         */

        /**
         * Constructs a new AllContainersResponse.
         * @memberof project
         * @classdesc Represents an AllContainersResponse.
         * @implements IAllContainersResponse
         * @constructor
         * @param {project.IAllContainersResponse=} [properties] Properties to set
         */
        function AllContainersResponse(properties) {
            this.items = [];
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * AllContainersResponse items.
         * @member {Array.<types.Container>} items
         * @memberof project.AllContainersResponse
         * @instance
         */
        AllContainersResponse.prototype.items = $util.emptyArray;

        /**
         * Encodes the specified AllContainersResponse message. Does not implicitly {@link project.AllContainersResponse.verify|verify} messages.
         * @function encode
         * @memberof project.AllContainersResponse
         * @static
         * @param {project.AllContainersResponse} message AllContainersResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        AllContainersResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.items != null && message.items.length)
                for (let i = 0; i < message.items.length; ++i)
                    $root.types.Container.encode(message.items[i], writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
            return writer;
        };

        /**
         * Decodes an AllContainersResponse message from the specified reader or buffer.
         * @function decode
         * @memberof project.AllContainersResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {project.AllContainersResponse} AllContainersResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        AllContainersResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.project.AllContainersResponse();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1:
                    if (!(message.items && message.items.length))
                        message.items = [];
                    message.items.push($root.types.Container.decode(reader, reader.uint32()));
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        return AllContainersResponse;
    })();

    project.ApplyResponse = (function() {

        /**
         * Properties of an ApplyResponse.
         * @memberof project
         * @interface IApplyResponse
         * @property {websocket.Metadata|null} [metadata] ApplyResponse metadata
         * @property {types.ProjectModel|null} [project] ApplyResponse project
         */

        /**
         * Constructs a new ApplyResponse.
         * @memberof project
         * @classdesc Represents an ApplyResponse.
         * @implements IApplyResponse
         * @constructor
         * @param {project.IApplyResponse=} [properties] Properties to set
         */
        function ApplyResponse(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * ApplyResponse metadata.
         * @member {websocket.Metadata|null|undefined} metadata
         * @memberof project.ApplyResponse
         * @instance
         */
        ApplyResponse.prototype.metadata = null;

        /**
         * ApplyResponse project.
         * @member {types.ProjectModel|null|undefined} project
         * @memberof project.ApplyResponse
         * @instance
         */
        ApplyResponse.prototype.project = null;

        /**
         * Encodes the specified ApplyResponse message. Does not implicitly {@link project.ApplyResponse.verify|verify} messages.
         * @function encode
         * @memberof project.ApplyResponse
         * @static
         * @param {project.ApplyResponse} message ApplyResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ApplyResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.metadata != null && Object.hasOwnProperty.call(message, "metadata"))
                $root.websocket.Metadata.encode(message.metadata, writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
            if (message.project != null && Object.hasOwnProperty.call(message, "project"))
                $root.types.ProjectModel.encode(message.project, writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim();
            return writer;
        };

        /**
         * Decodes an ApplyResponse message from the specified reader or buffer.
         * @function decode
         * @memberof project.ApplyResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {project.ApplyResponse} ApplyResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ApplyResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.project.ApplyResponse();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1:
                    message.metadata = $root.websocket.Metadata.decode(reader, reader.uint32());
                    break;
                case 2:
                    message.project = $root.types.ProjectModel.decode(reader, reader.uint32());
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        return ApplyResponse;
    })();

    project.DryRunApplyResponse = (function() {

        /**
         * Properties of a DryRunApplyResponse.
         * @memberof project
         * @interface IDryRunApplyResponse
         * @property {Array.<string>|null} [results] DryRunApplyResponse results
         */

        /**
         * Constructs a new DryRunApplyResponse.
         * @memberof project
         * @classdesc Represents a DryRunApplyResponse.
         * @implements IDryRunApplyResponse
         * @constructor
         * @param {project.IDryRunApplyResponse=} [properties] Properties to set
         */
        function DryRunApplyResponse(properties) {
            this.results = [];
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * DryRunApplyResponse results.
         * @member {Array.<string>} results
         * @memberof project.DryRunApplyResponse
         * @instance
         */
        DryRunApplyResponse.prototype.results = $util.emptyArray;

        /**
         * Encodes the specified DryRunApplyResponse message. Does not implicitly {@link project.DryRunApplyResponse.verify|verify} messages.
         * @function encode
         * @memberof project.DryRunApplyResponse
         * @static
         * @param {project.DryRunApplyResponse} message DryRunApplyResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        DryRunApplyResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.results != null && message.results.length)
                for (let i = 0; i < message.results.length; ++i)
                    writer.uint32(/* id 1, wireType 2 =*/10).string(message.results[i]);
            return writer;
        };

        /**
         * Decodes a DryRunApplyResponse message from the specified reader or buffer.
         * @function decode
         * @memberof project.DryRunApplyResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {project.DryRunApplyResponse} DryRunApplyResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        DryRunApplyResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.project.DryRunApplyResponse();
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

        return DryRunApplyResponse;
    })();

    project.ApplyRequest = (function() {

        /**
         * Properties of an ApplyRequest.
         * @memberof project
         * @interface IApplyRequest
         * @property {number|null} [namespace_id] ApplyRequest namespace_id
         * @property {string|null} [name] ApplyRequest name
         * @property {number|null} [git_project_id] ApplyRequest git_project_id
         * @property {string|null} [git_branch] ApplyRequest git_branch
         * @property {string|null} [git_commit] ApplyRequest git_commit
         * @property {string|null} [config] ApplyRequest config
         * @property {boolean|null} [atomic] ApplyRequest atomic
         * @property {boolean|null} [websocket_sync] ApplyRequest websocket_sync
         * @property {boolean|null} [send_percent] ApplyRequest send_percent
         * @property {Array.<types.ExtraValue>|null} [extra_values] ApplyRequest extra_values
         * @property {number|null} [install_timeout_seconds] ApplyRequest install_timeout_seconds
         */

        /**
         * Constructs a new ApplyRequest.
         * @memberof project
         * @classdesc Represents an ApplyRequest.
         * @implements IApplyRequest
         * @constructor
         * @param {project.IApplyRequest=} [properties] Properties to set
         */
        function ApplyRequest(properties) {
            this.extra_values = [];
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * ApplyRequest namespace_id.
         * @member {number} namespace_id
         * @memberof project.ApplyRequest
         * @instance
         */
        ApplyRequest.prototype.namespace_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * ApplyRequest name.
         * @member {string} name
         * @memberof project.ApplyRequest
         * @instance
         */
        ApplyRequest.prototype.name = "";

        /**
         * ApplyRequest git_project_id.
         * @member {number} git_project_id
         * @memberof project.ApplyRequest
         * @instance
         */
        ApplyRequest.prototype.git_project_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * ApplyRequest git_branch.
         * @member {string} git_branch
         * @memberof project.ApplyRequest
         * @instance
         */
        ApplyRequest.prototype.git_branch = "";

        /**
         * ApplyRequest git_commit.
         * @member {string} git_commit
         * @memberof project.ApplyRequest
         * @instance
         */
        ApplyRequest.prototype.git_commit = "";

        /**
         * ApplyRequest config.
         * @member {string} config
         * @memberof project.ApplyRequest
         * @instance
         */
        ApplyRequest.prototype.config = "";

        /**
         * ApplyRequest atomic.
         * @member {boolean} atomic
         * @memberof project.ApplyRequest
         * @instance
         */
        ApplyRequest.prototype.atomic = false;

        /**
         * ApplyRequest websocket_sync.
         * @member {boolean} websocket_sync
         * @memberof project.ApplyRequest
         * @instance
         */
        ApplyRequest.prototype.websocket_sync = false;

        /**
         * ApplyRequest send_percent.
         * @member {boolean} send_percent
         * @memberof project.ApplyRequest
         * @instance
         */
        ApplyRequest.prototype.send_percent = false;

        /**
         * ApplyRequest extra_values.
         * @member {Array.<types.ExtraValue>} extra_values
         * @memberof project.ApplyRequest
         * @instance
         */
        ApplyRequest.prototype.extra_values = $util.emptyArray;

        /**
         * ApplyRequest install_timeout_seconds.
         * @member {number} install_timeout_seconds
         * @memberof project.ApplyRequest
         * @instance
         */
        ApplyRequest.prototype.install_timeout_seconds = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Encodes the specified ApplyRequest message. Does not implicitly {@link project.ApplyRequest.verify|verify} messages.
         * @function encode
         * @memberof project.ApplyRequest
         * @static
         * @param {project.ApplyRequest} message ApplyRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ApplyRequest.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.namespace_id != null && Object.hasOwnProperty.call(message, "namespace_id"))
                writer.uint32(/* id 1, wireType 0 =*/8).int64(message.namespace_id);
            if (message.name != null && Object.hasOwnProperty.call(message, "name"))
                writer.uint32(/* id 2, wireType 2 =*/18).string(message.name);
            if (message.git_project_id != null && Object.hasOwnProperty.call(message, "git_project_id"))
                writer.uint32(/* id 3, wireType 0 =*/24).int64(message.git_project_id);
            if (message.git_branch != null && Object.hasOwnProperty.call(message, "git_branch"))
                writer.uint32(/* id 4, wireType 2 =*/34).string(message.git_branch);
            if (message.git_commit != null && Object.hasOwnProperty.call(message, "git_commit"))
                writer.uint32(/* id 5, wireType 2 =*/42).string(message.git_commit);
            if (message.config != null && Object.hasOwnProperty.call(message, "config"))
                writer.uint32(/* id 6, wireType 2 =*/50).string(message.config);
            if (message.atomic != null && Object.hasOwnProperty.call(message, "atomic"))
                writer.uint32(/* id 7, wireType 0 =*/56).bool(message.atomic);
            if (message.websocket_sync != null && Object.hasOwnProperty.call(message, "websocket_sync"))
                writer.uint32(/* id 8, wireType 0 =*/64).bool(message.websocket_sync);
            if (message.extra_values != null && message.extra_values.length)
                for (let i = 0; i < message.extra_values.length; ++i)
                    $root.types.ExtraValue.encode(message.extra_values[i], writer.uint32(/* id 9, wireType 2 =*/74).fork()).ldelim();
            if (message.install_timeout_seconds != null && Object.hasOwnProperty.call(message, "install_timeout_seconds"))
                writer.uint32(/* id 10, wireType 0 =*/80).int64(message.install_timeout_seconds);
            if (message.send_percent != null && Object.hasOwnProperty.call(message, "send_percent"))
                writer.uint32(/* id 11, wireType 0 =*/88).bool(message.send_percent);
            return writer;
        };

        /**
         * Decodes an ApplyRequest message from the specified reader or buffer.
         * @function decode
         * @memberof project.ApplyRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {project.ApplyRequest} ApplyRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ApplyRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.project.ApplyRequest();
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
                    message.git_project_id = reader.int64();
                    break;
                case 4:
                    message.git_branch = reader.string();
                    break;
                case 5:
                    message.git_commit = reader.string();
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
                case 11:
                    message.send_percent = reader.bool();
                    break;
                case 9:
                    if (!(message.extra_values && message.extra_values.length))
                        message.extra_values = [];
                    message.extra_values.push($root.types.ExtraValue.decode(reader, reader.uint32()));
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

        return ApplyRequest;
    })();

    project.DeleteResponse = (function() {

        /**
         * Properties of a DeleteResponse.
         * @memberof project
         * @interface IDeleteResponse
         */

        /**
         * Constructs a new DeleteResponse.
         * @memberof project
         * @classdesc Represents a DeleteResponse.
         * @implements IDeleteResponse
         * @constructor
         * @param {project.IDeleteResponse=} [properties] Properties to set
         */
        function DeleteResponse(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * Encodes the specified DeleteResponse message. Does not implicitly {@link project.DeleteResponse.verify|verify} messages.
         * @function encode
         * @memberof project.DeleteResponse
         * @static
         * @param {project.DeleteResponse} message DeleteResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        DeleteResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            return writer;
        };

        /**
         * Decodes a DeleteResponse message from the specified reader or buffer.
         * @function decode
         * @memberof project.DeleteResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {project.DeleteResponse} DeleteResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        DeleteResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.project.DeleteResponse();
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

        return DeleteResponse;
    })();

    project.ListRequest = (function() {

        /**
         * Properties of a ListRequest.
         * @memberof project
         * @interface IListRequest
         * @property {number|null} [page] ListRequest page
         * @property {number|null} [page_size] ListRequest page_size
         */

        /**
         * Constructs a new ListRequest.
         * @memberof project
         * @classdesc Represents a ListRequest.
         * @implements IListRequest
         * @constructor
         * @param {project.IListRequest=} [properties] Properties to set
         */
        function ListRequest(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * ListRequest page.
         * @member {number} page
         * @memberof project.ListRequest
         * @instance
         */
        ListRequest.prototype.page = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * ListRequest page_size.
         * @member {number} page_size
         * @memberof project.ListRequest
         * @instance
         */
        ListRequest.prototype.page_size = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Encodes the specified ListRequest message. Does not implicitly {@link project.ListRequest.verify|verify} messages.
         * @function encode
         * @memberof project.ListRequest
         * @static
         * @param {project.ListRequest} message ListRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ListRequest.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.page != null && Object.hasOwnProperty.call(message, "page"))
                writer.uint32(/* id 1, wireType 0 =*/8).int64(message.page);
            if (message.page_size != null && Object.hasOwnProperty.call(message, "page_size"))
                writer.uint32(/* id 2, wireType 0 =*/16).int64(message.page_size);
            return writer;
        };

        /**
         * Decodes a ListRequest message from the specified reader or buffer.
         * @function decode
         * @memberof project.ListRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {project.ListRequest} ListRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ListRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.project.ListRequest();
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

        return ListRequest;
    })();

    project.ListResponse = (function() {

        /**
         * Properties of a ListResponse.
         * @memberof project
         * @interface IListResponse
         * @property {number|null} [page] ListResponse page
         * @property {number|null} [page_size] ListResponse page_size
         * @property {number|null} [count] ListResponse count
         * @property {Array.<types.ProjectModel>|null} [items] ListResponse items
         */

        /**
         * Constructs a new ListResponse.
         * @memberof project
         * @classdesc Represents a ListResponse.
         * @implements IListResponse
         * @constructor
         * @param {project.IListResponse=} [properties] Properties to set
         */
        function ListResponse(properties) {
            this.items = [];
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * ListResponse page.
         * @member {number} page
         * @memberof project.ListResponse
         * @instance
         */
        ListResponse.prototype.page = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * ListResponse page_size.
         * @member {number} page_size
         * @memberof project.ListResponse
         * @instance
         */
        ListResponse.prototype.page_size = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * ListResponse count.
         * @member {number} count
         * @memberof project.ListResponse
         * @instance
         */
        ListResponse.prototype.count = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * ListResponse items.
         * @member {Array.<types.ProjectModel>} items
         * @memberof project.ListResponse
         * @instance
         */
        ListResponse.prototype.items = $util.emptyArray;

        /**
         * Encodes the specified ListResponse message. Does not implicitly {@link project.ListResponse.verify|verify} messages.
         * @function encode
         * @memberof project.ListResponse
         * @static
         * @param {project.ListResponse} message ListResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ListResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.page != null && Object.hasOwnProperty.call(message, "page"))
                writer.uint32(/* id 1, wireType 0 =*/8).int64(message.page);
            if (message.page_size != null && Object.hasOwnProperty.call(message, "page_size"))
                writer.uint32(/* id 2, wireType 0 =*/16).int64(message.page_size);
            if (message.count != null && Object.hasOwnProperty.call(message, "count"))
                writer.uint32(/* id 3, wireType 0 =*/24).int64(message.count);
            if (message.items != null && message.items.length)
                for (let i = 0; i < message.items.length; ++i)
                    $root.types.ProjectModel.encode(message.items[i], writer.uint32(/* id 4, wireType 2 =*/34).fork()).ldelim();
            return writer;
        };

        /**
         * Decodes a ListResponse message from the specified reader or buffer.
         * @function decode
         * @memberof project.ListResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {project.ListResponse} ListResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ListResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.project.ListResponse();
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
                    if (!(message.items && message.items.length))
                        message.items = [];
                    message.items.push($root.types.ProjectModel.decode(reader, reader.uint32()));
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        return ListResponse;
    })();

    project.Project = (function() {

        /**
         * Constructs a new Project service.
         * @memberof project
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
         * Callback as used by {@link project.Project#list}.
         * @memberof project.Project
         * @typedef ListCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {project.ListResponse} [response] ListResponse
         */

        /**
         * Calls List.
         * @function list
         * @memberof project.Project
         * @instance
         * @param {project.ListRequest} request ListRequest message or plain object
         * @param {project.Project.ListCallback} callback Node-style callback called with the error, if any, and ListResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(Project.prototype.list = function list(request, callback) {
            return this.rpcCall(list, $root.project.ListRequest, $root.project.ListResponse, request, callback);
        }, "name", { value: "List" });

        /**
         * Calls List.
         * @function list
         * @memberof project.Project
         * @instance
         * @param {project.ListRequest} request ListRequest message or plain object
         * @returns {Promise<project.ListResponse>} Promise
         * @variation 2
         */

        /**
         * Callback as used by {@link project.Project#apply}.
         * @memberof project.Project
         * @typedef ApplyCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {project.ApplyResponse} [response] ApplyResponse
         */

        /**
         * Calls Apply.
         * @function apply
         * @memberof project.Project
         * @instance
         * @param {project.ApplyRequest} request ApplyRequest message or plain object
         * @param {project.Project.ApplyCallback} callback Node-style callback called with the error, if any, and ApplyResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(Project.prototype.apply = function apply(request, callback) {
            return this.rpcCall(apply, $root.project.ApplyRequest, $root.project.ApplyResponse, request, callback);
        }, "name", { value: "Apply" });

        /**
         * Calls Apply.
         * @function apply
         * @memberof project.Project
         * @instance
         * @param {project.ApplyRequest} request ApplyRequest message or plain object
         * @returns {Promise<project.ApplyResponse>} Promise
         * @variation 2
         */

        /**
         * Callback as used by {@link project.Project#applyDryRun}.
         * @memberof project.Project
         * @typedef ApplyDryRunCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {project.DryRunApplyResponse} [response] DryRunApplyResponse
         */

        /**
         * Calls ApplyDryRun.
         * @function applyDryRun
         * @memberof project.Project
         * @instance
         * @param {project.ApplyRequest} request ApplyRequest message or plain object
         * @param {project.Project.ApplyDryRunCallback} callback Node-style callback called with the error, if any, and DryRunApplyResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(Project.prototype.applyDryRun = function applyDryRun(request, callback) {
            return this.rpcCall(applyDryRun, $root.project.ApplyRequest, $root.project.DryRunApplyResponse, request, callback);
        }, "name", { value: "ApplyDryRun" });

        /**
         * Calls ApplyDryRun.
         * @function applyDryRun
         * @memberof project.Project
         * @instance
         * @param {project.ApplyRequest} request ApplyRequest message or plain object
         * @returns {Promise<project.DryRunApplyResponse>} Promise
         * @variation 2
         */

        /**
         * Callback as used by {@link project.Project#show}.
         * @memberof project.Project
         * @typedef ShowCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {project.ShowResponse} [response] ShowResponse
         */

        /**
         * Calls Show.
         * @function show
         * @memberof project.Project
         * @instance
         * @param {project.ShowRequest} request ShowRequest message or plain object
         * @param {project.Project.ShowCallback} callback Node-style callback called with the error, if any, and ShowResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(Project.prototype.show = function show(request, callback) {
            return this.rpcCall(show, $root.project.ShowRequest, $root.project.ShowResponse, request, callback);
        }, "name", { value: "Show" });

        /**
         * Calls Show.
         * @function show
         * @memberof project.Project
         * @instance
         * @param {project.ShowRequest} request ShowRequest message or plain object
         * @returns {Promise<project.ShowResponse>} Promise
         * @variation 2
         */

        /**
         * Callback as used by {@link project.Project#delete_}.
         * @memberof project.Project
         * @typedef DeleteCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {project.DeleteResponse} [response] DeleteResponse
         */

        /**
         * Calls Delete.
         * @function delete
         * @memberof project.Project
         * @instance
         * @param {project.DeleteRequest} request DeleteRequest message or plain object
         * @param {project.Project.DeleteCallback} callback Node-style callback called with the error, if any, and DeleteResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(Project.prototype["delete"] = function delete_(request, callback) {
            return this.rpcCall(delete_, $root.project.DeleteRequest, $root.project.DeleteResponse, request, callback);
        }, "name", { value: "Delete" });

        /**
         * Calls Delete.
         * @function delete
         * @memberof project.Project
         * @instance
         * @param {project.DeleteRequest} request DeleteRequest message or plain object
         * @returns {Promise<project.DeleteResponse>} Promise
         * @variation 2
         */

        /**
         * Callback as used by {@link project.Project#allContainers}.
         * @memberof project.Project
         * @typedef AllContainersCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {project.AllContainersResponse} [response] AllContainersResponse
         */

        /**
         * Calls AllContainers.
         * @function allContainers
         * @memberof project.Project
         * @instance
         * @param {project.AllContainersRequest} request AllContainersRequest message or plain object
         * @param {project.Project.AllContainersCallback} callback Node-style callback called with the error, if any, and AllContainersResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(Project.prototype.allContainers = function allContainers(request, callback) {
            return this.rpcCall(allContainers, $root.project.AllContainersRequest, $root.project.AllContainersResponse, request, callback);
        }, "name", { value: "AllContainers" });

        /**
         * Calls AllContainers.
         * @function allContainers
         * @memberof project.Project
         * @instance
         * @param {project.AllContainersRequest} request AllContainersRequest message or plain object
         * @returns {Promise<project.AllContainersResponse>} Promise
         * @variation 2
         */

        return Project;
    })();

    return project;
})();

export const types = $root.types = (() => {

    /**
     * Namespace types.
     * @exports types
     * @namespace
     */
    const types = {};

    /**
     * EventActionType enum.
     * @name types.EventActionType
     * @enum {number}
     * @property {number} Unknown=0 Unknown value
     * @property {number} Create=1 Create value
     * @property {number} Update=2 Update value
     * @property {number} Delete=3 Delete value
     * @property {number} Upload=4 Upload value
     * @property {number} Download=5 Download value
     * @property {number} DryRun=6 DryRun value
     * @property {number} Shell=7 Shell value
     */
    types.EventActionType = (function() {
        const valuesById = {}, values = Object.create(valuesById);
        values[valuesById[0] = "Unknown"] = 0;
        values[valuesById[1] = "Create"] = 1;
        values[valuesById[2] = "Update"] = 2;
        values[valuesById[3] = "Delete"] = 3;
        values[valuesById[4] = "Upload"] = 4;
        values[valuesById[5] = "Download"] = 5;
        values[valuesById[6] = "DryRun"] = 6;
        values[valuesById[7] = "Shell"] = 7;
        return values;
    })();

    types.Pod = (function() {

        /**
         * Properties of a Pod.
         * @memberof types
         * @interface IPod
         * @property {string|null} [namespace] Pod namespace
         * @property {string|null} [pod] Pod pod
         */

        /**
         * Constructs a new Pod.
         * @memberof types
         * @classdesc Represents a Pod.
         * @implements IPod
         * @constructor
         * @param {types.IPod=} [properties] Properties to set
         */
        function Pod(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * Pod namespace.
         * @member {string} namespace
         * @memberof types.Pod
         * @instance
         */
        Pod.prototype.namespace = "";

        /**
         * Pod pod.
         * @member {string} pod
         * @memberof types.Pod
         * @instance
         */
        Pod.prototype.pod = "";

        /**
         * Encodes the specified Pod message. Does not implicitly {@link types.Pod.verify|verify} messages.
         * @function encode
         * @memberof types.Pod
         * @static
         * @param {types.Pod} message Pod message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        Pod.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.namespace != null && Object.hasOwnProperty.call(message, "namespace"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.namespace);
            if (message.pod != null && Object.hasOwnProperty.call(message, "pod"))
                writer.uint32(/* id 2, wireType 2 =*/18).string(message.pod);
            return writer;
        };

        /**
         * Decodes a Pod message from the specified reader or buffer.
         * @function decode
         * @memberof types.Pod
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {types.Pod} Pod
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Pod.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.types.Pod();
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

        return Pod;
    })();

    types.Container = (function() {

        /**
         * Properties of a Container.
         * @memberof types
         * @interface IContainer
         * @property {string|null} [namespace] Container namespace
         * @property {string|null} [pod] Container pod
         * @property {string|null} [container] Container container
         */

        /**
         * Constructs a new Container.
         * @memberof types
         * @classdesc Represents a Container.
         * @implements IContainer
         * @constructor
         * @param {types.IContainer=} [properties] Properties to set
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
         * @memberof types.Container
         * @instance
         */
        Container.prototype.namespace = "";

        /**
         * Container pod.
         * @member {string} pod
         * @memberof types.Container
         * @instance
         */
        Container.prototype.pod = "";

        /**
         * Container container.
         * @member {string} container
         * @memberof types.Container
         * @instance
         */
        Container.prototype.container = "";

        /**
         * Encodes the specified Container message. Does not implicitly {@link types.Container.verify|verify} messages.
         * @function encode
         * @memberof types.Container
         * @static
         * @param {types.Container} message Container message or plain object to encode
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
         * @memberof types.Container
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {types.Container} Container
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Container.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.types.Container();
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

    types.ExtraValue = (function() {

        /**
         * Properties of an ExtraValue.
         * @memberof types
         * @interface IExtraValue
         * @property {string|null} [path] ExtraValue path
         * @property {string|null} [value] ExtraValue value
         */

        /**
         * Constructs a new ExtraValue.
         * @memberof types
         * @classdesc Represents an ExtraValue.
         * @implements IExtraValue
         * @constructor
         * @param {types.IExtraValue=} [properties] Properties to set
         */
        function ExtraValue(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * ExtraValue path.
         * @member {string} path
         * @memberof types.ExtraValue
         * @instance
         */
        ExtraValue.prototype.path = "";

        /**
         * ExtraValue value.
         * @member {string} value
         * @memberof types.ExtraValue
         * @instance
         */
        ExtraValue.prototype.value = "";

        /**
         * Encodes the specified ExtraValue message. Does not implicitly {@link types.ExtraValue.verify|verify} messages.
         * @function encode
         * @memberof types.ExtraValue
         * @static
         * @param {types.ExtraValue} message ExtraValue message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ExtraValue.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.path != null && Object.hasOwnProperty.call(message, "path"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.path);
            if (message.value != null && Object.hasOwnProperty.call(message, "value"))
                writer.uint32(/* id 2, wireType 2 =*/18).string(message.value);
            return writer;
        };

        /**
         * Decodes an ExtraValue message from the specified reader or buffer.
         * @function decode
         * @memberof types.ExtraValue
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {types.ExtraValue} ExtraValue
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ExtraValue.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.types.ExtraValue();
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

        return ExtraValue;
    })();

    types.ServiceEndpoint = (function() {

        /**
         * Properties of a ServiceEndpoint.
         * @memberof types
         * @interface IServiceEndpoint
         * @property {string|null} [name] ServiceEndpoint name
         * @property {string|null} [url] ServiceEndpoint url
         * @property {string|null} [port_name] ServiceEndpoint port_name
         */

        /**
         * Constructs a new ServiceEndpoint.
         * @memberof types
         * @classdesc Represents a ServiceEndpoint.
         * @implements IServiceEndpoint
         * @constructor
         * @param {types.IServiceEndpoint=} [properties] Properties to set
         */
        function ServiceEndpoint(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * ServiceEndpoint name.
         * @member {string} name
         * @memberof types.ServiceEndpoint
         * @instance
         */
        ServiceEndpoint.prototype.name = "";

        /**
         * ServiceEndpoint url.
         * @member {string} url
         * @memberof types.ServiceEndpoint
         * @instance
         */
        ServiceEndpoint.prototype.url = "";

        /**
         * ServiceEndpoint port_name.
         * @member {string} port_name
         * @memberof types.ServiceEndpoint
         * @instance
         */
        ServiceEndpoint.prototype.port_name = "";

        /**
         * Encodes the specified ServiceEndpoint message. Does not implicitly {@link types.ServiceEndpoint.verify|verify} messages.
         * @function encode
         * @memberof types.ServiceEndpoint
         * @static
         * @param {types.ServiceEndpoint} message ServiceEndpoint message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ServiceEndpoint.encode = function encode(message, writer) {
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
         * Decodes a ServiceEndpoint message from the specified reader or buffer.
         * @function decode
         * @memberof types.ServiceEndpoint
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {types.ServiceEndpoint} ServiceEndpoint
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ServiceEndpoint.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.types.ServiceEndpoint();
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

        return ServiceEndpoint;
    })();

    types.ChangelogModel = (function() {

        /**
         * Properties of a ChangelogModel.
         * @memberof types
         * @interface IChangelogModel
         * @property {number|null} [id] ChangelogModel id
         * @property {number|null} [version] ChangelogModel version
         * @property {string|null} [username] ChangelogModel username
         * @property {string|null} [manifest] ChangelogModel manifest
         * @property {string|null} [config] ChangelogModel config
         * @property {boolean|null} [config_changed] ChangelogModel config_changed
         * @property {number|null} [project_id] ChangelogModel project_id
         * @property {number|null} [git_project_id] ChangelogModel git_project_id
         * @property {types.ProjectModel|null} [project] ChangelogModel project
         * @property {types.GitProjectModel|null} [git_project] ChangelogModel git_project
         * @property {string|null} [date] ChangelogModel date
         * @property {string|null} [created_at] ChangelogModel created_at
         * @property {string|null} [updated_at] ChangelogModel updated_at
         * @property {string|null} [deleted_at] ChangelogModel deleted_at
         */

        /**
         * Constructs a new ChangelogModel.
         * @memberof types
         * @classdesc Represents a ChangelogModel.
         * @implements IChangelogModel
         * @constructor
         * @param {types.IChangelogModel=} [properties] Properties to set
         */
        function ChangelogModel(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * ChangelogModel id.
         * @member {number} id
         * @memberof types.ChangelogModel
         * @instance
         */
        ChangelogModel.prototype.id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * ChangelogModel version.
         * @member {number} version
         * @memberof types.ChangelogModel
         * @instance
         */
        ChangelogModel.prototype.version = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * ChangelogModel username.
         * @member {string} username
         * @memberof types.ChangelogModel
         * @instance
         */
        ChangelogModel.prototype.username = "";

        /**
         * ChangelogModel manifest.
         * @member {string} manifest
         * @memberof types.ChangelogModel
         * @instance
         */
        ChangelogModel.prototype.manifest = "";

        /**
         * ChangelogModel config.
         * @member {string} config
         * @memberof types.ChangelogModel
         * @instance
         */
        ChangelogModel.prototype.config = "";

        /**
         * ChangelogModel config_changed.
         * @member {boolean} config_changed
         * @memberof types.ChangelogModel
         * @instance
         */
        ChangelogModel.prototype.config_changed = false;

        /**
         * ChangelogModel project_id.
         * @member {number} project_id
         * @memberof types.ChangelogModel
         * @instance
         */
        ChangelogModel.prototype.project_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * ChangelogModel git_project_id.
         * @member {number} git_project_id
         * @memberof types.ChangelogModel
         * @instance
         */
        ChangelogModel.prototype.git_project_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * ChangelogModel project.
         * @member {types.ProjectModel|null|undefined} project
         * @memberof types.ChangelogModel
         * @instance
         */
        ChangelogModel.prototype.project = null;

        /**
         * ChangelogModel git_project.
         * @member {types.GitProjectModel|null|undefined} git_project
         * @memberof types.ChangelogModel
         * @instance
         */
        ChangelogModel.prototype.git_project = null;

        /**
         * ChangelogModel date.
         * @member {string} date
         * @memberof types.ChangelogModel
         * @instance
         */
        ChangelogModel.prototype.date = "";

        /**
         * ChangelogModel created_at.
         * @member {string} created_at
         * @memberof types.ChangelogModel
         * @instance
         */
        ChangelogModel.prototype.created_at = "";

        /**
         * ChangelogModel updated_at.
         * @member {string} updated_at
         * @memberof types.ChangelogModel
         * @instance
         */
        ChangelogModel.prototype.updated_at = "";

        /**
         * ChangelogModel deleted_at.
         * @member {string} deleted_at
         * @memberof types.ChangelogModel
         * @instance
         */
        ChangelogModel.prototype.deleted_at = "";

        /**
         * Encodes the specified ChangelogModel message. Does not implicitly {@link types.ChangelogModel.verify|verify} messages.
         * @function encode
         * @memberof types.ChangelogModel
         * @static
         * @param {types.ChangelogModel} message ChangelogModel message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ChangelogModel.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.id != null && Object.hasOwnProperty.call(message, "id"))
                writer.uint32(/* id 1, wireType 0 =*/8).int64(message.id);
            if (message.version != null && Object.hasOwnProperty.call(message, "version"))
                writer.uint32(/* id 2, wireType 0 =*/16).int64(message.version);
            if (message.username != null && Object.hasOwnProperty.call(message, "username"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.username);
            if (message.manifest != null && Object.hasOwnProperty.call(message, "manifest"))
                writer.uint32(/* id 4, wireType 2 =*/34).string(message.manifest);
            if (message.config != null && Object.hasOwnProperty.call(message, "config"))
                writer.uint32(/* id 5, wireType 2 =*/42).string(message.config);
            if (message.config_changed != null && Object.hasOwnProperty.call(message, "config_changed"))
                writer.uint32(/* id 6, wireType 0 =*/48).bool(message.config_changed);
            if (message.project_id != null && Object.hasOwnProperty.call(message, "project_id"))
                writer.uint32(/* id 7, wireType 0 =*/56).int64(message.project_id);
            if (message.git_project_id != null && Object.hasOwnProperty.call(message, "git_project_id"))
                writer.uint32(/* id 8, wireType 0 =*/64).int64(message.git_project_id);
            if (message.project != null && Object.hasOwnProperty.call(message, "project"))
                $root.types.ProjectModel.encode(message.project, writer.uint32(/* id 9, wireType 2 =*/74).fork()).ldelim();
            if (message.git_project != null && Object.hasOwnProperty.call(message, "git_project"))
                $root.types.GitProjectModel.encode(message.git_project, writer.uint32(/* id 10, wireType 2 =*/82).fork()).ldelim();
            if (message.date != null && Object.hasOwnProperty.call(message, "date"))
                writer.uint32(/* id 11, wireType 2 =*/90).string(message.date);
            if (message.created_at != null && Object.hasOwnProperty.call(message, "created_at"))
                writer.uint32(/* id 100, wireType 2 =*/802).string(message.created_at);
            if (message.updated_at != null && Object.hasOwnProperty.call(message, "updated_at"))
                writer.uint32(/* id 101, wireType 2 =*/810).string(message.updated_at);
            if (message.deleted_at != null && Object.hasOwnProperty.call(message, "deleted_at"))
                writer.uint32(/* id 102, wireType 2 =*/818).string(message.deleted_at);
            return writer;
        };

        /**
         * Decodes a ChangelogModel message from the specified reader or buffer.
         * @function decode
         * @memberof types.ChangelogModel
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {types.ChangelogModel} ChangelogModel
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ChangelogModel.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.types.ChangelogModel();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1:
                    message.id = reader.int64();
                    break;
                case 2:
                    message.version = reader.int64();
                    break;
                case 3:
                    message.username = reader.string();
                    break;
                case 4:
                    message.manifest = reader.string();
                    break;
                case 5:
                    message.config = reader.string();
                    break;
                case 6:
                    message.config_changed = reader.bool();
                    break;
                case 7:
                    message.project_id = reader.int64();
                    break;
                case 8:
                    message.git_project_id = reader.int64();
                    break;
                case 9:
                    message.project = $root.types.ProjectModel.decode(reader, reader.uint32());
                    break;
                case 10:
                    message.git_project = $root.types.GitProjectModel.decode(reader, reader.uint32());
                    break;
                case 11:
                    message.date = reader.string();
                    break;
                case 100:
                    message.created_at = reader.string();
                    break;
                case 101:
                    message.updated_at = reader.string();
                    break;
                case 102:
                    message.deleted_at = reader.string();
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        return ChangelogModel;
    })();

    types.EventModel = (function() {

        /**
         * Properties of an EventModel.
         * @memberof types
         * @interface IEventModel
         * @property {number|null} [id] EventModel id
         * @property {types.EventActionType|null} [action] EventModel action
         * @property {string|null} [username] EventModel username
         * @property {string|null} [message] EventModel message
         * @property {string|null} [old] EventModel old
         * @property {string|null} ["new"] EventModel new
         * @property {string|null} [duration] EventModel duration
         * @property {number|null} [file_id] EventModel file_id
         * @property {types.FileModel|null} [file] EventModel file
         * @property {string|null} [event_at] EventModel event_at
         * @property {string|null} [created_at] EventModel created_at
         * @property {string|null} [updated_at] EventModel updated_at
         * @property {string|null} [deleted_at] EventModel deleted_at
         */

        /**
         * Constructs a new EventModel.
         * @memberof types
         * @classdesc Represents an EventModel.
         * @implements IEventModel
         * @constructor
         * @param {types.IEventModel=} [properties] Properties to set
         */
        function EventModel(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * EventModel id.
         * @member {number} id
         * @memberof types.EventModel
         * @instance
         */
        EventModel.prototype.id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * EventModel action.
         * @member {types.EventActionType} action
         * @memberof types.EventModel
         * @instance
         */
        EventModel.prototype.action = 0;

        /**
         * EventModel username.
         * @member {string} username
         * @memberof types.EventModel
         * @instance
         */
        EventModel.prototype.username = "";

        /**
         * EventModel message.
         * @member {string} message
         * @memberof types.EventModel
         * @instance
         */
        EventModel.prototype.message = "";

        /**
         * EventModel old.
         * @member {string} old
         * @memberof types.EventModel
         * @instance
         */
        EventModel.prototype.old = "";

        /**
         * EventModel new.
         * @member {string} new
         * @memberof types.EventModel
         * @instance
         */
        EventModel.prototype["new"] = "";

        /**
         * EventModel duration.
         * @member {string} duration
         * @memberof types.EventModel
         * @instance
         */
        EventModel.prototype.duration = "";

        /**
         * EventModel file_id.
         * @member {number} file_id
         * @memberof types.EventModel
         * @instance
         */
        EventModel.prototype.file_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * EventModel file.
         * @member {types.FileModel|null|undefined} file
         * @memberof types.EventModel
         * @instance
         */
        EventModel.prototype.file = null;

        /**
         * EventModel event_at.
         * @member {string} event_at
         * @memberof types.EventModel
         * @instance
         */
        EventModel.prototype.event_at = "";

        /**
         * EventModel created_at.
         * @member {string} created_at
         * @memberof types.EventModel
         * @instance
         */
        EventModel.prototype.created_at = "";

        /**
         * EventModel updated_at.
         * @member {string} updated_at
         * @memberof types.EventModel
         * @instance
         */
        EventModel.prototype.updated_at = "";

        /**
         * EventModel deleted_at.
         * @member {string} deleted_at
         * @memberof types.EventModel
         * @instance
         */
        EventModel.prototype.deleted_at = "";

        /**
         * Encodes the specified EventModel message. Does not implicitly {@link types.EventModel.verify|verify} messages.
         * @function encode
         * @memberof types.EventModel
         * @static
         * @param {types.EventModel} message EventModel message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        EventModel.encode = function encode(message, writer) {
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
            if (message.duration != null && Object.hasOwnProperty.call(message, "duration"))
                writer.uint32(/* id 7, wireType 2 =*/58).string(message.duration);
            if (message.file_id != null && Object.hasOwnProperty.call(message, "file_id"))
                writer.uint32(/* id 8, wireType 0 =*/64).int64(message.file_id);
            if (message.file != null && Object.hasOwnProperty.call(message, "file"))
                $root.types.FileModel.encode(message.file, writer.uint32(/* id 9, wireType 2 =*/74).fork()).ldelim();
            if (message.event_at != null && Object.hasOwnProperty.call(message, "event_at"))
                writer.uint32(/* id 10, wireType 2 =*/82).string(message.event_at);
            if (message.created_at != null && Object.hasOwnProperty.call(message, "created_at"))
                writer.uint32(/* id 100, wireType 2 =*/802).string(message.created_at);
            if (message.updated_at != null && Object.hasOwnProperty.call(message, "updated_at"))
                writer.uint32(/* id 101, wireType 2 =*/810).string(message.updated_at);
            if (message.deleted_at != null && Object.hasOwnProperty.call(message, "deleted_at"))
                writer.uint32(/* id 102, wireType 2 =*/818).string(message.deleted_at);
            return writer;
        };

        /**
         * Decodes an EventModel message from the specified reader or buffer.
         * @function decode
         * @memberof types.EventModel
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {types.EventModel} EventModel
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        EventModel.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.types.EventModel();
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
                    message.duration = reader.string();
                    break;
                case 8:
                    message.file_id = reader.int64();
                    break;
                case 9:
                    message.file = $root.types.FileModel.decode(reader, reader.uint32());
                    break;
                case 10:
                    message.event_at = reader.string();
                    break;
                case 100:
                    message.created_at = reader.string();
                    break;
                case 101:
                    message.updated_at = reader.string();
                    break;
                case 102:
                    message.deleted_at = reader.string();
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        return EventModel;
    })();

    types.FileModel = (function() {

        /**
         * Properties of a FileModel.
         * @memberof types
         * @interface IFileModel
         * @property {number|null} [id] FileModel id
         * @property {string|null} [path] FileModel path
         * @property {number|null} [size] FileModel size
         * @property {string|null} [username] FileModel username
         * @property {string|null} [namespace] FileModel namespace
         * @property {string|null} [pod] FileModel pod
         * @property {string|null} [container] FileModel container
         * @property {string|null} [container_Path] FileModel container_Path
         * @property {string|null} [humanize_size] FileModel humanize_size
         * @property {string|null} [created_at] FileModel created_at
         * @property {string|null} [updated_at] FileModel updated_at
         * @property {string|null} [deleted_at] FileModel deleted_at
         */

        /**
         * Constructs a new FileModel.
         * @memberof types
         * @classdesc Represents a FileModel.
         * @implements IFileModel
         * @constructor
         * @param {types.IFileModel=} [properties] Properties to set
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
         * @memberof types.FileModel
         * @instance
         */
        FileModel.prototype.id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * FileModel path.
         * @member {string} path
         * @memberof types.FileModel
         * @instance
         */
        FileModel.prototype.path = "";

        /**
         * FileModel size.
         * @member {number} size
         * @memberof types.FileModel
         * @instance
         */
        FileModel.prototype.size = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * FileModel username.
         * @member {string} username
         * @memberof types.FileModel
         * @instance
         */
        FileModel.prototype.username = "";

        /**
         * FileModel namespace.
         * @member {string} namespace
         * @memberof types.FileModel
         * @instance
         */
        FileModel.prototype.namespace = "";

        /**
         * FileModel pod.
         * @member {string} pod
         * @memberof types.FileModel
         * @instance
         */
        FileModel.prototype.pod = "";

        /**
         * FileModel container.
         * @member {string} container
         * @memberof types.FileModel
         * @instance
         */
        FileModel.prototype.container = "";

        /**
         * FileModel container_Path.
         * @member {string} container_Path
         * @memberof types.FileModel
         * @instance
         */
        FileModel.prototype.container_Path = "";

        /**
         * FileModel humanize_size.
         * @member {string} humanize_size
         * @memberof types.FileModel
         * @instance
         */
        FileModel.prototype.humanize_size = "";

        /**
         * FileModel created_at.
         * @member {string} created_at
         * @memberof types.FileModel
         * @instance
         */
        FileModel.prototype.created_at = "";

        /**
         * FileModel updated_at.
         * @member {string} updated_at
         * @memberof types.FileModel
         * @instance
         */
        FileModel.prototype.updated_at = "";

        /**
         * FileModel deleted_at.
         * @member {string} deleted_at
         * @memberof types.FileModel
         * @instance
         */
        FileModel.prototype.deleted_at = "";

        /**
         * Encodes the specified FileModel message. Does not implicitly {@link types.FileModel.verify|verify} messages.
         * @function encode
         * @memberof types.FileModel
         * @static
         * @param {types.FileModel} message FileModel message or plain object to encode
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
                writer.uint32(/* id 3, wireType 0 =*/24).int64(message.size);
            if (message.username != null && Object.hasOwnProperty.call(message, "username"))
                writer.uint32(/* id 4, wireType 2 =*/34).string(message.username);
            if (message.namespace != null && Object.hasOwnProperty.call(message, "namespace"))
                writer.uint32(/* id 5, wireType 2 =*/42).string(message.namespace);
            if (message.pod != null && Object.hasOwnProperty.call(message, "pod"))
                writer.uint32(/* id 6, wireType 2 =*/50).string(message.pod);
            if (message.container != null && Object.hasOwnProperty.call(message, "container"))
                writer.uint32(/* id 7, wireType 2 =*/58).string(message.container);
            if (message.container_Path != null && Object.hasOwnProperty.call(message, "container_Path"))
                writer.uint32(/* id 8, wireType 2 =*/66).string(message.container_Path);
            if (message.humanize_size != null && Object.hasOwnProperty.call(message, "humanize_size"))
                writer.uint32(/* id 9, wireType 2 =*/74).string(message.humanize_size);
            if (message.created_at != null && Object.hasOwnProperty.call(message, "created_at"))
                writer.uint32(/* id 100, wireType 2 =*/802).string(message.created_at);
            if (message.updated_at != null && Object.hasOwnProperty.call(message, "updated_at"))
                writer.uint32(/* id 101, wireType 2 =*/810).string(message.updated_at);
            if (message.deleted_at != null && Object.hasOwnProperty.call(message, "deleted_at"))
                writer.uint32(/* id 102, wireType 2 =*/818).string(message.deleted_at);
            return writer;
        };

        /**
         * Decodes a FileModel message from the specified reader or buffer.
         * @function decode
         * @memberof types.FileModel
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {types.FileModel} FileModel
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        FileModel.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.types.FileModel();
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
                    message.size = reader.int64();
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
                    message.container_Path = reader.string();
                    break;
                case 9:
                    message.humanize_size = reader.string();
                    break;
                case 100:
                    message.created_at = reader.string();
                    break;
                case 101:
                    message.updated_at = reader.string();
                    break;
                case 102:
                    message.deleted_at = reader.string();
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

    types.GitProjectModel = (function() {

        /**
         * Properties of a GitProjectModel.
         * @memberof types
         * @interface IGitProjectModel
         * @property {number|null} [id] GitProjectModel id
         * @property {string|null} [default_branch] GitProjectModel default_branch
         * @property {string|null} [name] GitProjectModel name
         * @property {number|null} [git_project_id] GitProjectModel git_project_id
         * @property {boolean|null} [enabled] GitProjectModel enabled
         * @property {boolean|null} [global_enabled] GitProjectModel global_enabled
         * @property {string|null} [global_config] GitProjectModel global_config
         * @property {string|null} [created_at] GitProjectModel created_at
         * @property {string|null} [updated_at] GitProjectModel updated_at
         * @property {string|null} [deleted_at] GitProjectModel deleted_at
         */

        /**
         * Constructs a new GitProjectModel.
         * @memberof types
         * @classdesc Represents a GitProjectModel.
         * @implements IGitProjectModel
         * @constructor
         * @param {types.IGitProjectModel=} [properties] Properties to set
         */
        function GitProjectModel(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * GitProjectModel id.
         * @member {number} id
         * @memberof types.GitProjectModel
         * @instance
         */
        GitProjectModel.prototype.id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * GitProjectModel default_branch.
         * @member {string} default_branch
         * @memberof types.GitProjectModel
         * @instance
         */
        GitProjectModel.prototype.default_branch = "";

        /**
         * GitProjectModel name.
         * @member {string} name
         * @memberof types.GitProjectModel
         * @instance
         */
        GitProjectModel.prototype.name = "";

        /**
         * GitProjectModel git_project_id.
         * @member {number} git_project_id
         * @memberof types.GitProjectModel
         * @instance
         */
        GitProjectModel.prototype.git_project_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * GitProjectModel enabled.
         * @member {boolean} enabled
         * @memberof types.GitProjectModel
         * @instance
         */
        GitProjectModel.prototype.enabled = false;

        /**
         * GitProjectModel global_enabled.
         * @member {boolean} global_enabled
         * @memberof types.GitProjectModel
         * @instance
         */
        GitProjectModel.prototype.global_enabled = false;

        /**
         * GitProjectModel global_config.
         * @member {string} global_config
         * @memberof types.GitProjectModel
         * @instance
         */
        GitProjectModel.prototype.global_config = "";

        /**
         * GitProjectModel created_at.
         * @member {string} created_at
         * @memberof types.GitProjectModel
         * @instance
         */
        GitProjectModel.prototype.created_at = "";

        /**
         * GitProjectModel updated_at.
         * @member {string} updated_at
         * @memberof types.GitProjectModel
         * @instance
         */
        GitProjectModel.prototype.updated_at = "";

        /**
         * GitProjectModel deleted_at.
         * @member {string} deleted_at
         * @memberof types.GitProjectModel
         * @instance
         */
        GitProjectModel.prototype.deleted_at = "";

        /**
         * Encodes the specified GitProjectModel message. Does not implicitly {@link types.GitProjectModel.verify|verify} messages.
         * @function encode
         * @memberof types.GitProjectModel
         * @static
         * @param {types.GitProjectModel} message GitProjectModel message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        GitProjectModel.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.id != null && Object.hasOwnProperty.call(message, "id"))
                writer.uint32(/* id 1, wireType 0 =*/8).int64(message.id);
            if (message.default_branch != null && Object.hasOwnProperty.call(message, "default_branch"))
                writer.uint32(/* id 2, wireType 2 =*/18).string(message.default_branch);
            if (message.name != null && Object.hasOwnProperty.call(message, "name"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.name);
            if (message.git_project_id != null && Object.hasOwnProperty.call(message, "git_project_id"))
                writer.uint32(/* id 4, wireType 0 =*/32).int64(message.git_project_id);
            if (message.enabled != null && Object.hasOwnProperty.call(message, "enabled"))
                writer.uint32(/* id 5, wireType 0 =*/40).bool(message.enabled);
            if (message.global_enabled != null && Object.hasOwnProperty.call(message, "global_enabled"))
                writer.uint32(/* id 6, wireType 0 =*/48).bool(message.global_enabled);
            if (message.global_config != null && Object.hasOwnProperty.call(message, "global_config"))
                writer.uint32(/* id 7, wireType 2 =*/58).string(message.global_config);
            if (message.created_at != null && Object.hasOwnProperty.call(message, "created_at"))
                writer.uint32(/* id 100, wireType 2 =*/802).string(message.created_at);
            if (message.updated_at != null && Object.hasOwnProperty.call(message, "updated_at"))
                writer.uint32(/* id 101, wireType 2 =*/810).string(message.updated_at);
            if (message.deleted_at != null && Object.hasOwnProperty.call(message, "deleted_at"))
                writer.uint32(/* id 102, wireType 2 =*/818).string(message.deleted_at);
            return writer;
        };

        /**
         * Decodes a GitProjectModel message from the specified reader or buffer.
         * @function decode
         * @memberof types.GitProjectModel
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {types.GitProjectModel} GitProjectModel
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        GitProjectModel.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.types.GitProjectModel();
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
                    message.git_project_id = reader.int64();
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
                case 100:
                    message.created_at = reader.string();
                    break;
                case 101:
                    message.updated_at = reader.string();
                    break;
                case 102:
                    message.deleted_at = reader.string();
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        return GitProjectModel;
    })();

    types.ImagePullSecret = (function() {

        /**
         * Properties of an ImagePullSecret.
         * @memberof types
         * @interface IImagePullSecret
         * @property {string|null} [name] ImagePullSecret name
         */

        /**
         * Constructs a new ImagePullSecret.
         * @memberof types
         * @classdesc Represents an ImagePullSecret.
         * @implements IImagePullSecret
         * @constructor
         * @param {types.IImagePullSecret=} [properties] Properties to set
         */
        function ImagePullSecret(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * ImagePullSecret name.
         * @member {string} name
         * @memberof types.ImagePullSecret
         * @instance
         */
        ImagePullSecret.prototype.name = "";

        /**
         * Encodes the specified ImagePullSecret message. Does not implicitly {@link types.ImagePullSecret.verify|verify} messages.
         * @function encode
         * @memberof types.ImagePullSecret
         * @static
         * @param {types.ImagePullSecret} message ImagePullSecret message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ImagePullSecret.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.name != null && Object.hasOwnProperty.call(message, "name"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.name);
            return writer;
        };

        /**
         * Decodes an ImagePullSecret message from the specified reader or buffer.
         * @function decode
         * @memberof types.ImagePullSecret
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {types.ImagePullSecret} ImagePullSecret
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ImagePullSecret.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.types.ImagePullSecret();
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

        return ImagePullSecret;
    })();

    types.NamespaceModel = (function() {

        /**
         * Properties of a NamespaceModel.
         * @memberof types
         * @interface INamespaceModel
         * @property {number|null} [id] NamespaceModel id
         * @property {string|null} [name] NamespaceModel name
         * @property {Array.<types.ImagePullSecret>|null} [ImagePullSecrets] NamespaceModel ImagePullSecrets
         * @property {Array.<types.ProjectModel>|null} [projects] NamespaceModel projects
         * @property {string|null} [created_at] NamespaceModel created_at
         * @property {string|null} [updated_at] NamespaceModel updated_at
         * @property {string|null} [deleted_at] NamespaceModel deleted_at
         */

        /**
         * Constructs a new NamespaceModel.
         * @memberof types
         * @classdesc Represents a NamespaceModel.
         * @implements INamespaceModel
         * @constructor
         * @param {types.INamespaceModel=} [properties] Properties to set
         */
        function NamespaceModel(properties) {
            this.ImagePullSecrets = [];
            this.projects = [];
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * NamespaceModel id.
         * @member {number} id
         * @memberof types.NamespaceModel
         * @instance
         */
        NamespaceModel.prototype.id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * NamespaceModel name.
         * @member {string} name
         * @memberof types.NamespaceModel
         * @instance
         */
        NamespaceModel.prototype.name = "";

        /**
         * NamespaceModel ImagePullSecrets.
         * @member {Array.<types.ImagePullSecret>} ImagePullSecrets
         * @memberof types.NamespaceModel
         * @instance
         */
        NamespaceModel.prototype.ImagePullSecrets = $util.emptyArray;

        /**
         * NamespaceModel projects.
         * @member {Array.<types.ProjectModel>} projects
         * @memberof types.NamespaceModel
         * @instance
         */
        NamespaceModel.prototype.projects = $util.emptyArray;

        /**
         * NamespaceModel created_at.
         * @member {string} created_at
         * @memberof types.NamespaceModel
         * @instance
         */
        NamespaceModel.prototype.created_at = "";

        /**
         * NamespaceModel updated_at.
         * @member {string} updated_at
         * @memberof types.NamespaceModel
         * @instance
         */
        NamespaceModel.prototype.updated_at = "";

        /**
         * NamespaceModel deleted_at.
         * @member {string} deleted_at
         * @memberof types.NamespaceModel
         * @instance
         */
        NamespaceModel.prototype.deleted_at = "";

        /**
         * Encodes the specified NamespaceModel message. Does not implicitly {@link types.NamespaceModel.verify|verify} messages.
         * @function encode
         * @memberof types.NamespaceModel
         * @static
         * @param {types.NamespaceModel} message NamespaceModel message or plain object to encode
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
            if (message.ImagePullSecrets != null && message.ImagePullSecrets.length)
                for (let i = 0; i < message.ImagePullSecrets.length; ++i)
                    $root.types.ImagePullSecret.encode(message.ImagePullSecrets[i], writer.uint32(/* id 3, wireType 2 =*/26).fork()).ldelim();
            if (message.projects != null && message.projects.length)
                for (let i = 0; i < message.projects.length; ++i)
                    $root.types.ProjectModel.encode(message.projects[i], writer.uint32(/* id 4, wireType 2 =*/34).fork()).ldelim();
            if (message.created_at != null && Object.hasOwnProperty.call(message, "created_at"))
                writer.uint32(/* id 100, wireType 2 =*/802).string(message.created_at);
            if (message.updated_at != null && Object.hasOwnProperty.call(message, "updated_at"))
                writer.uint32(/* id 101, wireType 2 =*/810).string(message.updated_at);
            if (message.deleted_at != null && Object.hasOwnProperty.call(message, "deleted_at"))
                writer.uint32(/* id 102, wireType 2 =*/818).string(message.deleted_at);
            return writer;
        };

        /**
         * Decodes a NamespaceModel message from the specified reader or buffer.
         * @function decode
         * @memberof types.NamespaceModel
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {types.NamespaceModel} NamespaceModel
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        NamespaceModel.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.types.NamespaceModel();
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
                    if (!(message.ImagePullSecrets && message.ImagePullSecrets.length))
                        message.ImagePullSecrets = [];
                    message.ImagePullSecrets.push($root.types.ImagePullSecret.decode(reader, reader.uint32()));
                    break;
                case 4:
                    if (!(message.projects && message.projects.length))
                        message.projects = [];
                    message.projects.push($root.types.ProjectModel.decode(reader, reader.uint32()));
                    break;
                case 100:
                    message.created_at = reader.string();
                    break;
                case 101:
                    message.updated_at = reader.string();
                    break;
                case 102:
                    message.deleted_at = reader.string();
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

    /**
     * Deploy enum.
     * @name types.Deploy
     * @enum {number}
     * @property {number} StatusUnknown=0 StatusUnknown value
     * @property {number} StatusDeploying=1 StatusDeploying value
     * @property {number} StatusDeployed=2 StatusDeployed value
     * @property {number} StatusFailed=3 StatusFailed value
     */
    types.Deploy = (function() {
        const valuesById = {}, values = Object.create(valuesById);
        values[valuesById[0] = "StatusUnknown"] = 0;
        values[valuesById[1] = "StatusDeploying"] = 1;
        values[valuesById[2] = "StatusDeployed"] = 2;
        values[valuesById[3] = "StatusFailed"] = 3;
        return values;
    })();

    types.ProjectModel = (function() {

        /**
         * Properties of a ProjectModel.
         * @memberof types
         * @interface IProjectModel
         * @property {number|null} [id] ProjectModel id
         * @property {string|null} [name] ProjectModel name
         * @property {number|null} [git_project_id] ProjectModel git_project_id
         * @property {string|null} [git_branch] ProjectModel git_branch
         * @property {string|null} [git_commit] ProjectModel git_commit
         * @property {string|null} [config] ProjectModel config
         * @property {string|null} [override_values] ProjectModel override_values
         * @property {string|null} [docker_image] ProjectModel docker_image
         * @property {string|null} [pod_selectors] ProjectModel pod_selectors
         * @property {number|null} [namespace_id] ProjectModel namespace_id
         * @property {boolean|null} [atomic] ProjectModel atomic
         * @property {string|null} [env_values] ProjectModel env_values
         * @property {Array.<types.ExtraValue>|null} [extra_values] ProjectModel extra_values
         * @property {string|null} [final_extra_values] ProjectModel final_extra_values
         * @property {types.Deploy|null} [deploy_status] ProjectModel deploy_status
         * @property {string|null} [humanize_created_at] ProjectModel humanize_created_at
         * @property {string|null} [humanize_updated_at] ProjectModel humanize_updated_at
         * @property {string|null} [config_type] ProjectModel config_type
         * @property {string|null} [git_commit_web_url] ProjectModel git_commit_web_url
         * @property {string|null} [git_commit_title] ProjectModel git_commit_title
         * @property {string|null} [git_commit_author] ProjectModel git_commit_author
         * @property {string|null} [git_commit_date] ProjectModel git_commit_date
         * @property {types.NamespaceModel|null} [namespace] ProjectModel namespace
         * @property {string|null} [created_at] ProjectModel created_at
         * @property {string|null} [updated_at] ProjectModel updated_at
         * @property {string|null} [deleted_at] ProjectModel deleted_at
         */

        /**
         * Constructs a new ProjectModel.
         * @memberof types
         * @classdesc Represents a ProjectModel.
         * @implements IProjectModel
         * @constructor
         * @param {types.IProjectModel=} [properties] Properties to set
         */
        function ProjectModel(properties) {
            this.extra_values = [];
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * ProjectModel id.
         * @member {number} id
         * @memberof types.ProjectModel
         * @instance
         */
        ProjectModel.prototype.id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * ProjectModel name.
         * @member {string} name
         * @memberof types.ProjectModel
         * @instance
         */
        ProjectModel.prototype.name = "";

        /**
         * ProjectModel git_project_id.
         * @member {number} git_project_id
         * @memberof types.ProjectModel
         * @instance
         */
        ProjectModel.prototype.git_project_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * ProjectModel git_branch.
         * @member {string} git_branch
         * @memberof types.ProjectModel
         * @instance
         */
        ProjectModel.prototype.git_branch = "";

        /**
         * ProjectModel git_commit.
         * @member {string} git_commit
         * @memberof types.ProjectModel
         * @instance
         */
        ProjectModel.prototype.git_commit = "";

        /**
         * ProjectModel config.
         * @member {string} config
         * @memberof types.ProjectModel
         * @instance
         */
        ProjectModel.prototype.config = "";

        /**
         * ProjectModel override_values.
         * @member {string} override_values
         * @memberof types.ProjectModel
         * @instance
         */
        ProjectModel.prototype.override_values = "";

        /**
         * ProjectModel docker_image.
         * @member {string} docker_image
         * @memberof types.ProjectModel
         * @instance
         */
        ProjectModel.prototype.docker_image = "";

        /**
         * ProjectModel pod_selectors.
         * @member {string} pod_selectors
         * @memberof types.ProjectModel
         * @instance
         */
        ProjectModel.prototype.pod_selectors = "";

        /**
         * ProjectModel namespace_id.
         * @member {number} namespace_id
         * @memberof types.ProjectModel
         * @instance
         */
        ProjectModel.prototype.namespace_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * ProjectModel atomic.
         * @member {boolean} atomic
         * @memberof types.ProjectModel
         * @instance
         */
        ProjectModel.prototype.atomic = false;

        /**
         * ProjectModel env_values.
         * @member {string} env_values
         * @memberof types.ProjectModel
         * @instance
         */
        ProjectModel.prototype.env_values = "";

        /**
         * ProjectModel extra_values.
         * @member {Array.<types.ExtraValue>} extra_values
         * @memberof types.ProjectModel
         * @instance
         */
        ProjectModel.prototype.extra_values = $util.emptyArray;

        /**
         * ProjectModel final_extra_values.
         * @member {string} final_extra_values
         * @memberof types.ProjectModel
         * @instance
         */
        ProjectModel.prototype.final_extra_values = "";

        /**
         * ProjectModel deploy_status.
         * @member {types.Deploy} deploy_status
         * @memberof types.ProjectModel
         * @instance
         */
        ProjectModel.prototype.deploy_status = 0;

        /**
         * ProjectModel humanize_created_at.
         * @member {string} humanize_created_at
         * @memberof types.ProjectModel
         * @instance
         */
        ProjectModel.prototype.humanize_created_at = "";

        /**
         * ProjectModel humanize_updated_at.
         * @member {string} humanize_updated_at
         * @memberof types.ProjectModel
         * @instance
         */
        ProjectModel.prototype.humanize_updated_at = "";

        /**
         * ProjectModel config_type.
         * @member {string} config_type
         * @memberof types.ProjectModel
         * @instance
         */
        ProjectModel.prototype.config_type = "";

        /**
         * ProjectModel git_commit_web_url.
         * @member {string} git_commit_web_url
         * @memberof types.ProjectModel
         * @instance
         */
        ProjectModel.prototype.git_commit_web_url = "";

        /**
         * ProjectModel git_commit_title.
         * @member {string} git_commit_title
         * @memberof types.ProjectModel
         * @instance
         */
        ProjectModel.prototype.git_commit_title = "";

        /**
         * ProjectModel git_commit_author.
         * @member {string} git_commit_author
         * @memberof types.ProjectModel
         * @instance
         */
        ProjectModel.prototype.git_commit_author = "";

        /**
         * ProjectModel git_commit_date.
         * @member {string} git_commit_date
         * @memberof types.ProjectModel
         * @instance
         */
        ProjectModel.prototype.git_commit_date = "";

        /**
         * ProjectModel namespace.
         * @member {types.NamespaceModel|null|undefined} namespace
         * @memberof types.ProjectModel
         * @instance
         */
        ProjectModel.prototype.namespace = null;

        /**
         * ProjectModel created_at.
         * @member {string} created_at
         * @memberof types.ProjectModel
         * @instance
         */
        ProjectModel.prototype.created_at = "";

        /**
         * ProjectModel updated_at.
         * @member {string} updated_at
         * @memberof types.ProjectModel
         * @instance
         */
        ProjectModel.prototype.updated_at = "";

        /**
         * ProjectModel deleted_at.
         * @member {string} deleted_at
         * @memberof types.ProjectModel
         * @instance
         */
        ProjectModel.prototype.deleted_at = "";

        /**
         * Encodes the specified ProjectModel message. Does not implicitly {@link types.ProjectModel.verify|verify} messages.
         * @function encode
         * @memberof types.ProjectModel
         * @static
         * @param {types.ProjectModel} message ProjectModel message or plain object to encode
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
            if (message.git_project_id != null && Object.hasOwnProperty.call(message, "git_project_id"))
                writer.uint32(/* id 3, wireType 0 =*/24).int64(message.git_project_id);
            if (message.git_branch != null && Object.hasOwnProperty.call(message, "git_branch"))
                writer.uint32(/* id 4, wireType 2 =*/34).string(message.git_branch);
            if (message.git_commit != null && Object.hasOwnProperty.call(message, "git_commit"))
                writer.uint32(/* id 5, wireType 2 =*/42).string(message.git_commit);
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
            if (message.env_values != null && Object.hasOwnProperty.call(message, "env_values"))
                writer.uint32(/* id 12, wireType 2 =*/98).string(message.env_values);
            if (message.extra_values != null && message.extra_values.length)
                for (let i = 0; i < message.extra_values.length; ++i)
                    $root.types.ExtraValue.encode(message.extra_values[i], writer.uint32(/* id 13, wireType 2 =*/106).fork()).ldelim();
            if (message.final_extra_values != null && Object.hasOwnProperty.call(message, "final_extra_values"))
                writer.uint32(/* id 14, wireType 2 =*/114).string(message.final_extra_values);
            if (message.deploy_status != null && Object.hasOwnProperty.call(message, "deploy_status"))
                writer.uint32(/* id 15, wireType 0 =*/120).int32(message.deploy_status);
            if (message.humanize_created_at != null && Object.hasOwnProperty.call(message, "humanize_created_at"))
                writer.uint32(/* id 16, wireType 2 =*/130).string(message.humanize_created_at);
            if (message.humanize_updated_at != null && Object.hasOwnProperty.call(message, "humanize_updated_at"))
                writer.uint32(/* id 17, wireType 2 =*/138).string(message.humanize_updated_at);
            if (message.config_type != null && Object.hasOwnProperty.call(message, "config_type"))
                writer.uint32(/* id 18, wireType 2 =*/146).string(message.config_type);
            if (message.git_commit_web_url != null && Object.hasOwnProperty.call(message, "git_commit_web_url"))
                writer.uint32(/* id 19, wireType 2 =*/154).string(message.git_commit_web_url);
            if (message.git_commit_title != null && Object.hasOwnProperty.call(message, "git_commit_title"))
                writer.uint32(/* id 20, wireType 2 =*/162).string(message.git_commit_title);
            if (message.git_commit_author != null && Object.hasOwnProperty.call(message, "git_commit_author"))
                writer.uint32(/* id 21, wireType 2 =*/170).string(message.git_commit_author);
            if (message.git_commit_date != null && Object.hasOwnProperty.call(message, "git_commit_date"))
                writer.uint32(/* id 22, wireType 2 =*/178).string(message.git_commit_date);
            if (message.namespace != null && Object.hasOwnProperty.call(message, "namespace"))
                $root.types.NamespaceModel.encode(message.namespace, writer.uint32(/* id 50, wireType 2 =*/402).fork()).ldelim();
            if (message.created_at != null && Object.hasOwnProperty.call(message, "created_at"))
                writer.uint32(/* id 100, wireType 2 =*/802).string(message.created_at);
            if (message.updated_at != null && Object.hasOwnProperty.call(message, "updated_at"))
                writer.uint32(/* id 101, wireType 2 =*/810).string(message.updated_at);
            if (message.deleted_at != null && Object.hasOwnProperty.call(message, "deleted_at"))
                writer.uint32(/* id 102, wireType 2 =*/818).string(message.deleted_at);
            return writer;
        };

        /**
         * Decodes a ProjectModel message from the specified reader or buffer.
         * @function decode
         * @memberof types.ProjectModel
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {types.ProjectModel} ProjectModel
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ProjectModel.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.types.ProjectModel();
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
                    message.git_project_id = reader.int64();
                    break;
                case 4:
                    message.git_branch = reader.string();
                    break;
                case 5:
                    message.git_commit = reader.string();
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
                    message.env_values = reader.string();
                    break;
                case 13:
                    if (!(message.extra_values && message.extra_values.length))
                        message.extra_values = [];
                    message.extra_values.push($root.types.ExtraValue.decode(reader, reader.uint32()));
                    break;
                case 14:
                    message.final_extra_values = reader.string();
                    break;
                case 15:
                    message.deploy_status = reader.int32();
                    break;
                case 16:
                    message.humanize_created_at = reader.string();
                    break;
                case 17:
                    message.humanize_updated_at = reader.string();
                    break;
                case 18:
                    message.config_type = reader.string();
                    break;
                case 19:
                    message.git_commit_web_url = reader.string();
                    break;
                case 20:
                    message.git_commit_title = reader.string();
                    break;
                case 21:
                    message.git_commit_author = reader.string();
                    break;
                case 22:
                    message.git_commit_date = reader.string();
                    break;
                case 50:
                    message.namespace = $root.types.NamespaceModel.decode(reader, reader.uint32());
                    break;
                case 100:
                    message.created_at = reader.string();
                    break;
                case 101:
                    message.updated_at = reader.string();
                    break;
                case 102:
                    message.deleted_at = reader.string();
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

    return types;
})();

export const version = $root.version = (() => {

    /**
     * Namespace version.
     * @exports version
     * @namespace
     */
    const version = {};

    version.Request = (function() {

        /**
         * Properties of a Request.
         * @memberof version
         * @interface IRequest
         */

        /**
         * Constructs a new Request.
         * @memberof version
         * @classdesc Represents a Request.
         * @implements IRequest
         * @constructor
         * @param {version.IRequest=} [properties] Properties to set
         */
        function Request(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * Encodes the specified Request message. Does not implicitly {@link version.Request.verify|verify} messages.
         * @function encode
         * @memberof version.Request
         * @static
         * @param {version.Request} message Request message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        Request.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            return writer;
        };

        /**
         * Decodes a Request message from the specified reader or buffer.
         * @function decode
         * @memberof version.Request
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {version.Request} Request
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Request.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.version.Request();
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

        return Request;
    })();

    version.Response = (function() {

        /**
         * Properties of a Response.
         * @memberof version
         * @interface IResponse
         * @property {string|null} [version] Response version
         * @property {string|null} [build_date] Response build_date
         * @property {string|null} [git_branch] Response git_branch
         * @property {string|null} [git_commit] Response git_commit
         * @property {string|null} [git_tag] Response git_tag
         * @property {string|null} [go_version] Response go_version
         * @property {string|null} [compiler] Response compiler
         * @property {string|null} [platform] Response platform
         * @property {string|null} [kubectl_version] Response kubectl_version
         * @property {string|null} [helm_version] Response helm_version
         * @property {string|null} [git_repo] Response git_repo
         */

        /**
         * Constructs a new Response.
         * @memberof version
         * @classdesc Represents a Response.
         * @implements IResponse
         * @constructor
         * @param {version.IResponse=} [properties] Properties to set
         */
        function Response(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * Response version.
         * @member {string} version
         * @memberof version.Response
         * @instance
         */
        Response.prototype.version = "";

        /**
         * Response build_date.
         * @member {string} build_date
         * @memberof version.Response
         * @instance
         */
        Response.prototype.build_date = "";

        /**
         * Response git_branch.
         * @member {string} git_branch
         * @memberof version.Response
         * @instance
         */
        Response.prototype.git_branch = "";

        /**
         * Response git_commit.
         * @member {string} git_commit
         * @memberof version.Response
         * @instance
         */
        Response.prototype.git_commit = "";

        /**
         * Response git_tag.
         * @member {string} git_tag
         * @memberof version.Response
         * @instance
         */
        Response.prototype.git_tag = "";

        /**
         * Response go_version.
         * @member {string} go_version
         * @memberof version.Response
         * @instance
         */
        Response.prototype.go_version = "";

        /**
         * Response compiler.
         * @member {string} compiler
         * @memberof version.Response
         * @instance
         */
        Response.prototype.compiler = "";

        /**
         * Response platform.
         * @member {string} platform
         * @memberof version.Response
         * @instance
         */
        Response.prototype.platform = "";

        /**
         * Response kubectl_version.
         * @member {string} kubectl_version
         * @memberof version.Response
         * @instance
         */
        Response.prototype.kubectl_version = "";

        /**
         * Response helm_version.
         * @member {string} helm_version
         * @memberof version.Response
         * @instance
         */
        Response.prototype.helm_version = "";

        /**
         * Response git_repo.
         * @member {string} git_repo
         * @memberof version.Response
         * @instance
         */
        Response.prototype.git_repo = "";

        /**
         * Encodes the specified Response message. Does not implicitly {@link version.Response.verify|verify} messages.
         * @function encode
         * @memberof version.Response
         * @static
         * @param {version.Response} message Response message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        Response.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.version != null && Object.hasOwnProperty.call(message, "version"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.version);
            if (message.build_date != null && Object.hasOwnProperty.call(message, "build_date"))
                writer.uint32(/* id 2, wireType 2 =*/18).string(message.build_date);
            if (message.git_branch != null && Object.hasOwnProperty.call(message, "git_branch"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.git_branch);
            if (message.git_commit != null && Object.hasOwnProperty.call(message, "git_commit"))
                writer.uint32(/* id 4, wireType 2 =*/34).string(message.git_commit);
            if (message.git_tag != null && Object.hasOwnProperty.call(message, "git_tag"))
                writer.uint32(/* id 5, wireType 2 =*/42).string(message.git_tag);
            if (message.go_version != null && Object.hasOwnProperty.call(message, "go_version"))
                writer.uint32(/* id 6, wireType 2 =*/50).string(message.go_version);
            if (message.compiler != null && Object.hasOwnProperty.call(message, "compiler"))
                writer.uint32(/* id 7, wireType 2 =*/58).string(message.compiler);
            if (message.platform != null && Object.hasOwnProperty.call(message, "platform"))
                writer.uint32(/* id 8, wireType 2 =*/66).string(message.platform);
            if (message.kubectl_version != null && Object.hasOwnProperty.call(message, "kubectl_version"))
                writer.uint32(/* id 9, wireType 2 =*/74).string(message.kubectl_version);
            if (message.helm_version != null && Object.hasOwnProperty.call(message, "helm_version"))
                writer.uint32(/* id 10, wireType 2 =*/82).string(message.helm_version);
            if (message.git_repo != null && Object.hasOwnProperty.call(message, "git_repo"))
                writer.uint32(/* id 11, wireType 2 =*/90).string(message.git_repo);
            return writer;
        };

        /**
         * Decodes a Response message from the specified reader or buffer.
         * @function decode
         * @memberof version.Response
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {version.Response} Response
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Response.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.version.Response();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1:
                    message.version = reader.string();
                    break;
                case 2:
                    message.build_date = reader.string();
                    break;
                case 3:
                    message.git_branch = reader.string();
                    break;
                case 4:
                    message.git_commit = reader.string();
                    break;
                case 5:
                    message.git_tag = reader.string();
                    break;
                case 6:
                    message.go_version = reader.string();
                    break;
                case 7:
                    message.compiler = reader.string();
                    break;
                case 8:
                    message.platform = reader.string();
                    break;
                case 9:
                    message.kubectl_version = reader.string();
                    break;
                case 10:
                    message.helm_version = reader.string();
                    break;
                case 11:
                    message.git_repo = reader.string();
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        return Response;
    })();

    version.Version = (function() {

        /**
         * Constructs a new Version service.
         * @memberof version
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
         * Callback as used by {@link version.Version#version}.
         * @memberof version.Version
         * @typedef VersionCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {version.Response} [response] Response
         */

        /**
         * Calls Version.
         * @function version
         * @memberof version.Version
         * @instance
         * @param {version.Request} request Request message or plain object
         * @param {version.Version.VersionCallback} callback Node-style callback called with the error, if any, and Response
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(Version.prototype.version = function version(request, callback) {
            return this.rpcCall(version, $root.version.Request, $root.version.Response, request, callback);
        }, "name", { value: "Version" });

        /**
         * Calls Version.
         * @function version
         * @memberof version.Version
         * @instance
         * @param {version.Request} request Request message or plain object
         * @returns {Promise<version.Response>} Promise
         * @variation 2
         */

        return Version;
    })();

    return version;
})();

export const websocket = $root.websocket = (() => {

    /**
     * Namespace websocket.
     * @exports websocket
     * @namespace
     */
    const websocket = {};

    /**
     * Type enum.
     * @name websocket.Type
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
    websocket.Type = (function() {
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
     * @name websocket.ResultType
     * @enum {number}
     * @property {number} ResultUnknown=0 ResultUnknown value
     * @property {number} Error=1 Error value
     * @property {number} Success=2 Success value
     * @property {number} Deployed=3 Deployed value
     * @property {number} DeployedFailed=4 DeployedFailed value
     * @property {number} DeployedCanceled=5 DeployedCanceled value
     */
    websocket.ResultType = (function() {
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
     * @name websocket.To
     * @enum {number}
     * @property {number} ToSelf=0 ToSelf value
     * @property {number} ToAll=1 ToAll value
     * @property {number} ToOthers=2 ToOthers value
     */
    websocket.To = (function() {
        const valuesById = {}, values = Object.create(valuesById);
        values[valuesById[0] = "ToSelf"] = 0;
        values[valuesById[1] = "ToAll"] = 1;
        values[valuesById[2] = "ToOthers"] = 2;
        return values;
    })();

    websocket.WsRequestMetadata = (function() {

        /**
         * Properties of a WsRequestMetadata.
         * @memberof websocket
         * @interface IWsRequestMetadata
         * @property {websocket.Type|null} [type] WsRequestMetadata type
         */

        /**
         * Constructs a new WsRequestMetadata.
         * @memberof websocket
         * @classdesc Represents a WsRequestMetadata.
         * @implements IWsRequestMetadata
         * @constructor
         * @param {websocket.IWsRequestMetadata=} [properties] Properties to set
         */
        function WsRequestMetadata(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * WsRequestMetadata type.
         * @member {websocket.Type} type
         * @memberof websocket.WsRequestMetadata
         * @instance
         */
        WsRequestMetadata.prototype.type = 0;

        /**
         * Encodes the specified WsRequestMetadata message. Does not implicitly {@link websocket.WsRequestMetadata.verify|verify} messages.
         * @function encode
         * @memberof websocket.WsRequestMetadata
         * @static
         * @param {websocket.WsRequestMetadata} message WsRequestMetadata message or plain object to encode
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
         * @memberof websocket.WsRequestMetadata
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {websocket.WsRequestMetadata} WsRequestMetadata
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        WsRequestMetadata.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.websocket.WsRequestMetadata();
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

    websocket.AuthorizeTokenInput = (function() {

        /**
         * Properties of an AuthorizeTokenInput.
         * @memberof websocket
         * @interface IAuthorizeTokenInput
         * @property {websocket.Type|null} [type] AuthorizeTokenInput type
         * @property {string|null} [token] AuthorizeTokenInput token
         */

        /**
         * Constructs a new AuthorizeTokenInput.
         * @memberof websocket
         * @classdesc Represents an AuthorizeTokenInput.
         * @implements IAuthorizeTokenInput
         * @constructor
         * @param {websocket.IAuthorizeTokenInput=} [properties] Properties to set
         */
        function AuthorizeTokenInput(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * AuthorizeTokenInput type.
         * @member {websocket.Type} type
         * @memberof websocket.AuthorizeTokenInput
         * @instance
         */
        AuthorizeTokenInput.prototype.type = 0;

        /**
         * AuthorizeTokenInput token.
         * @member {string} token
         * @memberof websocket.AuthorizeTokenInput
         * @instance
         */
        AuthorizeTokenInput.prototype.token = "";

        /**
         * Encodes the specified AuthorizeTokenInput message. Does not implicitly {@link websocket.AuthorizeTokenInput.verify|verify} messages.
         * @function encode
         * @memberof websocket.AuthorizeTokenInput
         * @static
         * @param {websocket.AuthorizeTokenInput} message AuthorizeTokenInput message or plain object to encode
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
         * @memberof websocket.AuthorizeTokenInput
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {websocket.AuthorizeTokenInput} AuthorizeTokenInput
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        AuthorizeTokenInput.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.websocket.AuthorizeTokenInput();
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

    websocket.TerminalMessage = (function() {

        /**
         * Properties of a TerminalMessage.
         * @memberof websocket
         * @interface ITerminalMessage
         * @property {string|null} [op] TerminalMessage op
         * @property {string|null} [data] TerminalMessage data
         * @property {string|null} [session_id] TerminalMessage session_id
         * @property {number|null} [rows] TerminalMessage rows
         * @property {number|null} [cols] TerminalMessage cols
         */

        /**
         * Constructs a new TerminalMessage.
         * @memberof websocket
         * @classdesc Represents a TerminalMessage.
         * @implements ITerminalMessage
         * @constructor
         * @param {websocket.ITerminalMessage=} [properties] Properties to set
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
         * @memberof websocket.TerminalMessage
         * @instance
         */
        TerminalMessage.prototype.op = "";

        /**
         * TerminalMessage data.
         * @member {string} data
         * @memberof websocket.TerminalMessage
         * @instance
         */
        TerminalMessage.prototype.data = "";

        /**
         * TerminalMessage session_id.
         * @member {string} session_id
         * @memberof websocket.TerminalMessage
         * @instance
         */
        TerminalMessage.prototype.session_id = "";

        /**
         * TerminalMessage rows.
         * @member {number} rows
         * @memberof websocket.TerminalMessage
         * @instance
         */
        TerminalMessage.prototype.rows = 0;

        /**
         * TerminalMessage cols.
         * @member {number} cols
         * @memberof websocket.TerminalMessage
         * @instance
         */
        TerminalMessage.prototype.cols = 0;

        /**
         * Encodes the specified TerminalMessage message. Does not implicitly {@link websocket.TerminalMessage.verify|verify} messages.
         * @function encode
         * @memberof websocket.TerminalMessage
         * @static
         * @param {websocket.TerminalMessage} message TerminalMessage message or plain object to encode
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
         * @memberof websocket.TerminalMessage
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {websocket.TerminalMessage} TerminalMessage
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        TerminalMessage.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.websocket.TerminalMessage();
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

    websocket.TerminalMessageInput = (function() {

        /**
         * Properties of a TerminalMessageInput.
         * @memberof websocket
         * @interface ITerminalMessageInput
         * @property {websocket.Type|null} [type] TerminalMessageInput type
         * @property {websocket.TerminalMessage|null} [message] TerminalMessageInput message
         */

        /**
         * Constructs a new TerminalMessageInput.
         * @memberof websocket
         * @classdesc Represents a TerminalMessageInput.
         * @implements ITerminalMessageInput
         * @constructor
         * @param {websocket.ITerminalMessageInput=} [properties] Properties to set
         */
        function TerminalMessageInput(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * TerminalMessageInput type.
         * @member {websocket.Type} type
         * @memberof websocket.TerminalMessageInput
         * @instance
         */
        TerminalMessageInput.prototype.type = 0;

        /**
         * TerminalMessageInput message.
         * @member {websocket.TerminalMessage|null|undefined} message
         * @memberof websocket.TerminalMessageInput
         * @instance
         */
        TerminalMessageInput.prototype.message = null;

        /**
         * Encodes the specified TerminalMessageInput message. Does not implicitly {@link websocket.TerminalMessageInput.verify|verify} messages.
         * @function encode
         * @memberof websocket.TerminalMessageInput
         * @static
         * @param {websocket.TerminalMessageInput} message TerminalMessageInput message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        TerminalMessageInput.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.type != null && Object.hasOwnProperty.call(message, "type"))
                writer.uint32(/* id 1, wireType 0 =*/8).int32(message.type);
            if (message.message != null && Object.hasOwnProperty.call(message, "message"))
                $root.websocket.TerminalMessage.encode(message.message, writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim();
            return writer;
        };

        /**
         * Decodes a TerminalMessageInput message from the specified reader or buffer.
         * @function decode
         * @memberof websocket.TerminalMessageInput
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {websocket.TerminalMessageInput} TerminalMessageInput
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        TerminalMessageInput.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.websocket.TerminalMessageInput();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1:
                    message.type = reader.int32();
                    break;
                case 2:
                    message.message = $root.websocket.TerminalMessage.decode(reader, reader.uint32());
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

    websocket.WsHandleExecShellInput = (function() {

        /**
         * Properties of a WsHandleExecShellInput.
         * @memberof websocket
         * @interface IWsHandleExecShellInput
         * @property {websocket.Type|null} [type] WsHandleExecShellInput type
         * @property {types.Container|null} [container] WsHandleExecShellInput container
         */

        /**
         * Constructs a new WsHandleExecShellInput.
         * @memberof websocket
         * @classdesc Represents a WsHandleExecShellInput.
         * @implements IWsHandleExecShellInput
         * @constructor
         * @param {websocket.IWsHandleExecShellInput=} [properties] Properties to set
         */
        function WsHandleExecShellInput(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * WsHandleExecShellInput type.
         * @member {websocket.Type} type
         * @memberof websocket.WsHandleExecShellInput
         * @instance
         */
        WsHandleExecShellInput.prototype.type = 0;

        /**
         * WsHandleExecShellInput container.
         * @member {types.Container|null|undefined} container
         * @memberof websocket.WsHandleExecShellInput
         * @instance
         */
        WsHandleExecShellInput.prototype.container = null;

        /**
         * Encodes the specified WsHandleExecShellInput message. Does not implicitly {@link websocket.WsHandleExecShellInput.verify|verify} messages.
         * @function encode
         * @memberof websocket.WsHandleExecShellInput
         * @static
         * @param {websocket.WsHandleExecShellInput} message WsHandleExecShellInput message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        WsHandleExecShellInput.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.type != null && Object.hasOwnProperty.call(message, "type"))
                writer.uint32(/* id 1, wireType 0 =*/8).int32(message.type);
            if (message.container != null && Object.hasOwnProperty.call(message, "container"))
                $root.types.Container.encode(message.container, writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim();
            return writer;
        };

        /**
         * Decodes a WsHandleExecShellInput message from the specified reader or buffer.
         * @function decode
         * @memberof websocket.WsHandleExecShellInput
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {websocket.WsHandleExecShellInput} WsHandleExecShellInput
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        WsHandleExecShellInput.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.websocket.WsHandleExecShellInput();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1:
                    message.type = reader.int32();
                    break;
                case 2:
                    message.container = $root.types.Container.decode(reader, reader.uint32());
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

    websocket.CancelInput = (function() {

        /**
         * Properties of a CancelInput.
         * @memberof websocket
         * @interface ICancelInput
         * @property {websocket.Type|null} [type] CancelInput type
         * @property {number|null} [namespace_id] CancelInput namespace_id
         * @property {string|null} [name] CancelInput name
         */

        /**
         * Constructs a new CancelInput.
         * @memberof websocket
         * @classdesc Represents a CancelInput.
         * @implements ICancelInput
         * @constructor
         * @param {websocket.ICancelInput=} [properties] Properties to set
         */
        function CancelInput(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * CancelInput type.
         * @member {websocket.Type} type
         * @memberof websocket.CancelInput
         * @instance
         */
        CancelInput.prototype.type = 0;

        /**
         * CancelInput namespace_id.
         * @member {number} namespace_id
         * @memberof websocket.CancelInput
         * @instance
         */
        CancelInput.prototype.namespace_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * CancelInput name.
         * @member {string} name
         * @memberof websocket.CancelInput
         * @instance
         */
        CancelInput.prototype.name = "";

        /**
         * Encodes the specified CancelInput message. Does not implicitly {@link websocket.CancelInput.verify|verify} messages.
         * @function encode
         * @memberof websocket.CancelInput
         * @static
         * @param {websocket.CancelInput} message CancelInput message or plain object to encode
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
         * @memberof websocket.CancelInput
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {websocket.CancelInput} CancelInput
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        CancelInput.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.websocket.CancelInput();
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

    websocket.CreateProjectInput = (function() {

        /**
         * Properties of a CreateProjectInput.
         * @memberof websocket
         * @interface ICreateProjectInput
         * @property {websocket.Type|null} [type] CreateProjectInput type
         * @property {number|null} [namespace_id] CreateProjectInput namespace_id
         * @property {string|null} [name] CreateProjectInput name
         * @property {number|null} [git_project_id] CreateProjectInput git_project_id
         * @property {string|null} [git_branch] CreateProjectInput git_branch
         * @property {string|null} [git_commit] CreateProjectInput git_commit
         * @property {string|null} [config] CreateProjectInput config
         * @property {boolean|null} [atomic] CreateProjectInput atomic
         * @property {Array.<types.ExtraValue>|null} [extra_values] CreateProjectInput extra_values
         */

        /**
         * Constructs a new CreateProjectInput.
         * @memberof websocket
         * @classdesc Represents a CreateProjectInput.
         * @implements ICreateProjectInput
         * @constructor
         * @param {websocket.ICreateProjectInput=} [properties] Properties to set
         */
        function CreateProjectInput(properties) {
            this.extra_values = [];
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * CreateProjectInput type.
         * @member {websocket.Type} type
         * @memberof websocket.CreateProjectInput
         * @instance
         */
        CreateProjectInput.prototype.type = 0;

        /**
         * CreateProjectInput namespace_id.
         * @member {number} namespace_id
         * @memberof websocket.CreateProjectInput
         * @instance
         */
        CreateProjectInput.prototype.namespace_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * CreateProjectInput name.
         * @member {string} name
         * @memberof websocket.CreateProjectInput
         * @instance
         */
        CreateProjectInput.prototype.name = "";

        /**
         * CreateProjectInput git_project_id.
         * @member {number} git_project_id
         * @memberof websocket.CreateProjectInput
         * @instance
         */
        CreateProjectInput.prototype.git_project_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * CreateProjectInput git_branch.
         * @member {string} git_branch
         * @memberof websocket.CreateProjectInput
         * @instance
         */
        CreateProjectInput.prototype.git_branch = "";

        /**
         * CreateProjectInput git_commit.
         * @member {string} git_commit
         * @memberof websocket.CreateProjectInput
         * @instance
         */
        CreateProjectInput.prototype.git_commit = "";

        /**
         * CreateProjectInput config.
         * @member {string} config
         * @memberof websocket.CreateProjectInput
         * @instance
         */
        CreateProjectInput.prototype.config = "";

        /**
         * CreateProjectInput atomic.
         * @member {boolean} atomic
         * @memberof websocket.CreateProjectInput
         * @instance
         */
        CreateProjectInput.prototype.atomic = false;

        /**
         * CreateProjectInput extra_values.
         * @member {Array.<types.ExtraValue>} extra_values
         * @memberof websocket.CreateProjectInput
         * @instance
         */
        CreateProjectInput.prototype.extra_values = $util.emptyArray;

        /**
         * Encodes the specified CreateProjectInput message. Does not implicitly {@link websocket.CreateProjectInput.verify|verify} messages.
         * @function encode
         * @memberof websocket.CreateProjectInput
         * @static
         * @param {websocket.CreateProjectInput} message CreateProjectInput message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        CreateProjectInput.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.type != null && Object.hasOwnProperty.call(message, "type"))
                writer.uint32(/* id 1, wireType 0 =*/8).int32(message.type);
            if (message.namespace_id != null && Object.hasOwnProperty.call(message, "namespace_id"))
                writer.uint32(/* id 2, wireType 0 =*/16).int64(message.namespace_id);
            if (message.name != null && Object.hasOwnProperty.call(message, "name"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.name);
            if (message.git_project_id != null && Object.hasOwnProperty.call(message, "git_project_id"))
                writer.uint32(/* id 4, wireType 0 =*/32).int64(message.git_project_id);
            if (message.git_branch != null && Object.hasOwnProperty.call(message, "git_branch"))
                writer.uint32(/* id 5, wireType 2 =*/42).string(message.git_branch);
            if (message.git_commit != null && Object.hasOwnProperty.call(message, "git_commit"))
                writer.uint32(/* id 6, wireType 2 =*/50).string(message.git_commit);
            if (message.config != null && Object.hasOwnProperty.call(message, "config"))
                writer.uint32(/* id 7, wireType 2 =*/58).string(message.config);
            if (message.atomic != null && Object.hasOwnProperty.call(message, "atomic"))
                writer.uint32(/* id 8, wireType 0 =*/64).bool(message.atomic);
            if (message.extra_values != null && message.extra_values.length)
                for (let i = 0; i < message.extra_values.length; ++i)
                    $root.types.ExtraValue.encode(message.extra_values[i], writer.uint32(/* id 9, wireType 2 =*/74).fork()).ldelim();
            return writer;
        };

        /**
         * Decodes a CreateProjectInput message from the specified reader or buffer.
         * @function decode
         * @memberof websocket.CreateProjectInput
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {websocket.CreateProjectInput} CreateProjectInput
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        CreateProjectInput.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.websocket.CreateProjectInput();
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
                    message.git_project_id = reader.int64();
                    break;
                case 5:
                    message.git_branch = reader.string();
                    break;
                case 6:
                    message.git_commit = reader.string();
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
                    message.extra_values.push($root.types.ExtraValue.decode(reader, reader.uint32()));
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        return CreateProjectInput;
    })();

    websocket.UpdateProjectInput = (function() {

        /**
         * Properties of an UpdateProjectInput.
         * @memberof websocket
         * @interface IUpdateProjectInput
         * @property {websocket.Type|null} [type] UpdateProjectInput type
         * @property {number|null} [project_id] UpdateProjectInput project_id
         * @property {string|null} [git_branch] UpdateProjectInput git_branch
         * @property {string|null} [git_commit] UpdateProjectInput git_commit
         * @property {string|null} [config] UpdateProjectInput config
         * @property {boolean|null} [atomic] UpdateProjectInput atomic
         * @property {Array.<types.ExtraValue>|null} [extra_values] UpdateProjectInput extra_values
         */

        /**
         * Constructs a new UpdateProjectInput.
         * @memberof websocket
         * @classdesc Represents an UpdateProjectInput.
         * @implements IUpdateProjectInput
         * @constructor
         * @param {websocket.IUpdateProjectInput=} [properties] Properties to set
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
         * @member {websocket.Type} type
         * @memberof websocket.UpdateProjectInput
         * @instance
         */
        UpdateProjectInput.prototype.type = 0;

        /**
         * UpdateProjectInput project_id.
         * @member {number} project_id
         * @memberof websocket.UpdateProjectInput
         * @instance
         */
        UpdateProjectInput.prototype.project_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * UpdateProjectInput git_branch.
         * @member {string} git_branch
         * @memberof websocket.UpdateProjectInput
         * @instance
         */
        UpdateProjectInput.prototype.git_branch = "";

        /**
         * UpdateProjectInput git_commit.
         * @member {string} git_commit
         * @memberof websocket.UpdateProjectInput
         * @instance
         */
        UpdateProjectInput.prototype.git_commit = "";

        /**
         * UpdateProjectInput config.
         * @member {string} config
         * @memberof websocket.UpdateProjectInput
         * @instance
         */
        UpdateProjectInput.prototype.config = "";

        /**
         * UpdateProjectInput atomic.
         * @member {boolean} atomic
         * @memberof websocket.UpdateProjectInput
         * @instance
         */
        UpdateProjectInput.prototype.atomic = false;

        /**
         * UpdateProjectInput extra_values.
         * @member {Array.<types.ExtraValue>} extra_values
         * @memberof websocket.UpdateProjectInput
         * @instance
         */
        UpdateProjectInput.prototype.extra_values = $util.emptyArray;

        /**
         * Encodes the specified UpdateProjectInput message. Does not implicitly {@link websocket.UpdateProjectInput.verify|verify} messages.
         * @function encode
         * @memberof websocket.UpdateProjectInput
         * @static
         * @param {websocket.UpdateProjectInput} message UpdateProjectInput message or plain object to encode
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
            if (message.git_branch != null && Object.hasOwnProperty.call(message, "git_branch"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.git_branch);
            if (message.git_commit != null && Object.hasOwnProperty.call(message, "git_commit"))
                writer.uint32(/* id 4, wireType 2 =*/34).string(message.git_commit);
            if (message.config != null && Object.hasOwnProperty.call(message, "config"))
                writer.uint32(/* id 5, wireType 2 =*/42).string(message.config);
            if (message.atomic != null && Object.hasOwnProperty.call(message, "atomic"))
                writer.uint32(/* id 6, wireType 0 =*/48).bool(message.atomic);
            if (message.extra_values != null && message.extra_values.length)
                for (let i = 0; i < message.extra_values.length; ++i)
                    $root.types.ExtraValue.encode(message.extra_values[i], writer.uint32(/* id 7, wireType 2 =*/58).fork()).ldelim();
            return writer;
        };

        /**
         * Decodes an UpdateProjectInput message from the specified reader or buffer.
         * @function decode
         * @memberof websocket.UpdateProjectInput
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {websocket.UpdateProjectInput} UpdateProjectInput
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        UpdateProjectInput.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.websocket.UpdateProjectInput();
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
                    message.git_branch = reader.string();
                    break;
                case 4:
                    message.git_commit = reader.string();
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
                    message.extra_values.push($root.types.ExtraValue.decode(reader, reader.uint32()));
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

    websocket.Metadata = (function() {

        /**
         * Properties of a Metadata.
         * @memberof websocket
         * @interface IMetadata
         * @property {string|null} [id] Metadata id
         * @property {string|null} [uid] Metadata uid
         * @property {string|null} [slug] Metadata slug
         * @property {websocket.Type|null} [type] Metadata type
         * @property {boolean|null} [end] Metadata end
         * @property {websocket.ResultType|null} [result] Metadata result
         * @property {websocket.To|null} [to] Metadata to
         * @property {string|null} [message] Metadata message
         */

        /**
         * Constructs a new Metadata.
         * @memberof websocket
         * @classdesc Represents a Metadata.
         * @implements IMetadata
         * @constructor
         * @param {websocket.IMetadata=} [properties] Properties to set
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
         * @memberof websocket.Metadata
         * @instance
         */
        Metadata.prototype.id = "";

        /**
         * Metadata uid.
         * @member {string} uid
         * @memberof websocket.Metadata
         * @instance
         */
        Metadata.prototype.uid = "";

        /**
         * Metadata slug.
         * @member {string} slug
         * @memberof websocket.Metadata
         * @instance
         */
        Metadata.prototype.slug = "";

        /**
         * Metadata type.
         * @member {websocket.Type} type
         * @memberof websocket.Metadata
         * @instance
         */
        Metadata.prototype.type = 0;

        /**
         * Metadata end.
         * @member {boolean} end
         * @memberof websocket.Metadata
         * @instance
         */
        Metadata.prototype.end = false;

        /**
         * Metadata result.
         * @member {websocket.ResultType} result
         * @memberof websocket.Metadata
         * @instance
         */
        Metadata.prototype.result = 0;

        /**
         * Metadata to.
         * @member {websocket.To} to
         * @memberof websocket.Metadata
         * @instance
         */
        Metadata.prototype.to = 0;

        /**
         * Metadata message.
         * @member {string} message
         * @memberof websocket.Metadata
         * @instance
         */
        Metadata.prototype.message = "";

        /**
         * Encodes the specified Metadata message. Does not implicitly {@link websocket.Metadata.verify|verify} messages.
         * @function encode
         * @memberof websocket.Metadata
         * @static
         * @param {websocket.Metadata} message Metadata message or plain object to encode
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
            if (message.message != null && Object.hasOwnProperty.call(message, "message"))
                writer.uint32(/* id 8, wireType 2 =*/66).string(message.message);
            return writer;
        };

        /**
         * Decodes a Metadata message from the specified reader or buffer.
         * @function decode
         * @memberof websocket.Metadata
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {websocket.Metadata} Metadata
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Metadata.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.websocket.Metadata();
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
                    message.message = reader.string();
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

    websocket.WsMetadataResponse = (function() {

        /**
         * Properties of a WsMetadataResponse.
         * @memberof websocket
         * @interface IWsMetadataResponse
         * @property {websocket.Metadata|null} [metadata] WsMetadataResponse metadata
         */

        /**
         * Constructs a new WsMetadataResponse.
         * @memberof websocket
         * @classdesc Represents a WsMetadataResponse.
         * @implements IWsMetadataResponse
         * @constructor
         * @param {websocket.IWsMetadataResponse=} [properties] Properties to set
         */
        function WsMetadataResponse(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * WsMetadataResponse metadata.
         * @member {websocket.Metadata|null|undefined} metadata
         * @memberof websocket.WsMetadataResponse
         * @instance
         */
        WsMetadataResponse.prototype.metadata = null;

        /**
         * Encodes the specified WsMetadataResponse message. Does not implicitly {@link websocket.WsMetadataResponse.verify|verify} messages.
         * @function encode
         * @memberof websocket.WsMetadataResponse
         * @static
         * @param {websocket.WsMetadataResponse} message WsMetadataResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        WsMetadataResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.metadata != null && Object.hasOwnProperty.call(message, "metadata"))
                $root.websocket.Metadata.encode(message.metadata, writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
            return writer;
        };

        /**
         * Decodes a WsMetadataResponse message from the specified reader or buffer.
         * @function decode
         * @memberof websocket.WsMetadataResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {websocket.WsMetadataResponse} WsMetadataResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        WsMetadataResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.websocket.WsMetadataResponse();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1:
                    message.metadata = $root.websocket.Metadata.decode(reader, reader.uint32());
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

    websocket.WsHandleShellResponse = (function() {

        /**
         * Properties of a WsHandleShellResponse.
         * @memberof websocket
         * @interface IWsHandleShellResponse
         * @property {websocket.Metadata|null} [metadata] WsHandleShellResponse metadata
         * @property {websocket.TerminalMessage|null} [terminal_message] WsHandleShellResponse terminal_message
         * @property {types.Container|null} [container] WsHandleShellResponse container
         */

        /**
         * Constructs a new WsHandleShellResponse.
         * @memberof websocket
         * @classdesc Represents a WsHandleShellResponse.
         * @implements IWsHandleShellResponse
         * @constructor
         * @param {websocket.IWsHandleShellResponse=} [properties] Properties to set
         */
        function WsHandleShellResponse(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * WsHandleShellResponse metadata.
         * @member {websocket.Metadata|null|undefined} metadata
         * @memberof websocket.WsHandleShellResponse
         * @instance
         */
        WsHandleShellResponse.prototype.metadata = null;

        /**
         * WsHandleShellResponse terminal_message.
         * @member {websocket.TerminalMessage|null|undefined} terminal_message
         * @memberof websocket.WsHandleShellResponse
         * @instance
         */
        WsHandleShellResponse.prototype.terminal_message = null;

        /**
         * WsHandleShellResponse container.
         * @member {types.Container|null|undefined} container
         * @memberof websocket.WsHandleShellResponse
         * @instance
         */
        WsHandleShellResponse.prototype.container = null;

        /**
         * Encodes the specified WsHandleShellResponse message. Does not implicitly {@link websocket.WsHandleShellResponse.verify|verify} messages.
         * @function encode
         * @memberof websocket.WsHandleShellResponse
         * @static
         * @param {websocket.WsHandleShellResponse} message WsHandleShellResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        WsHandleShellResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.metadata != null && Object.hasOwnProperty.call(message, "metadata"))
                $root.websocket.Metadata.encode(message.metadata, writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
            if (message.terminal_message != null && Object.hasOwnProperty.call(message, "terminal_message"))
                $root.websocket.TerminalMessage.encode(message.terminal_message, writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim();
            if (message.container != null && Object.hasOwnProperty.call(message, "container"))
                $root.types.Container.encode(message.container, writer.uint32(/* id 3, wireType 2 =*/26).fork()).ldelim();
            return writer;
        };

        /**
         * Decodes a WsHandleShellResponse message from the specified reader or buffer.
         * @function decode
         * @memberof websocket.WsHandleShellResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {websocket.WsHandleShellResponse} WsHandleShellResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        WsHandleShellResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.websocket.WsHandleShellResponse();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1:
                    message.metadata = $root.websocket.Metadata.decode(reader, reader.uint32());
                    break;
                case 2:
                    message.terminal_message = $root.websocket.TerminalMessage.decode(reader, reader.uint32());
                    break;
                case 3:
                    message.container = $root.types.Container.decode(reader, reader.uint32());
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

    websocket.WsHandleClusterResponse = (function() {

        /**
         * Properties of a WsHandleClusterResponse.
         * @memberof websocket
         * @interface IWsHandleClusterResponse
         * @property {websocket.Metadata|null} [metadata] WsHandleClusterResponse metadata
         * @property {cluster.InfoResponse|null} [info] WsHandleClusterResponse info
         */

        /**
         * Constructs a new WsHandleClusterResponse.
         * @memberof websocket
         * @classdesc Represents a WsHandleClusterResponse.
         * @implements IWsHandleClusterResponse
         * @constructor
         * @param {websocket.IWsHandleClusterResponse=} [properties] Properties to set
         */
        function WsHandleClusterResponse(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * WsHandleClusterResponse metadata.
         * @member {websocket.Metadata|null|undefined} metadata
         * @memberof websocket.WsHandleClusterResponse
         * @instance
         */
        WsHandleClusterResponse.prototype.metadata = null;

        /**
         * WsHandleClusterResponse info.
         * @member {cluster.InfoResponse|null|undefined} info
         * @memberof websocket.WsHandleClusterResponse
         * @instance
         */
        WsHandleClusterResponse.prototype.info = null;

        /**
         * Encodes the specified WsHandleClusterResponse message. Does not implicitly {@link websocket.WsHandleClusterResponse.verify|verify} messages.
         * @function encode
         * @memberof websocket.WsHandleClusterResponse
         * @static
         * @param {websocket.WsHandleClusterResponse} message WsHandleClusterResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        WsHandleClusterResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.metadata != null && Object.hasOwnProperty.call(message, "metadata"))
                $root.websocket.Metadata.encode(message.metadata, writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
            if (message.info != null && Object.hasOwnProperty.call(message, "info"))
                $root.cluster.InfoResponse.encode(message.info, writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim();
            return writer;
        };

        /**
         * Decodes a WsHandleClusterResponse message from the specified reader or buffer.
         * @function decode
         * @memberof websocket.WsHandleClusterResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {websocket.WsHandleClusterResponse} WsHandleClusterResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        WsHandleClusterResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.websocket.WsHandleClusterResponse();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1:
                    message.metadata = $root.websocket.Metadata.decode(reader, reader.uint32());
                    break;
                case 2:
                    message.info = $root.cluster.InfoResponse.decode(reader, reader.uint32());
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

    return websocket;
})();

export { $root as default };
