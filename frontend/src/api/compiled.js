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
                case 1: {
                        message.username = reader.string();
                        break;
                    }
                case 2: {
                        message.password = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for LoginRequest
         * @function getTypeUrl
         * @memberof auth.LoginRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        LoginRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/auth.LoginRequest";
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
                case 1: {
                        message.token = reader.string();
                        break;
                    }
                case 2: {
                        message.expires_in = reader.int64();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for LoginResponse
         * @function getTypeUrl
         * @memberof auth.LoginResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        LoginResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/auth.LoginResponse";
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
                case 1: {
                        message.code = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for ExchangeRequest
         * @function getTypeUrl
         * @memberof auth.ExchangeRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        ExchangeRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/auth.ExchangeRequest";
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
                case 1: {
                        message.token = reader.string();
                        break;
                    }
                case 2: {
                        message.expires_in = reader.int64();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for ExchangeResponse
         * @function getTypeUrl
         * @memberof auth.ExchangeResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        ExchangeResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/auth.ExchangeResponse";
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

        /**
         * Gets the default type url for InfoRequest
         * @function getTypeUrl
         * @memberof auth.InfoRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        InfoRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/auth.InfoRequest";
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
                case 1: {
                        message.id = reader.string();
                        break;
                    }
                case 2: {
                        message.avatar = reader.string();
                        break;
                    }
                case 3: {
                        message.name = reader.string();
                        break;
                    }
                case 4: {
                        message.email = reader.string();
                        break;
                    }
                case 5: {
                        message.logout_url = reader.string();
                        break;
                    }
                case 6: {
                        if (!(message.roles && message.roles.length))
                            message.roles = [];
                        message.roles.push(reader.string());
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for InfoResponse
         * @function getTypeUrl
         * @memberof auth.InfoResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        InfoResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/auth.InfoResponse";
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

        /**
         * Gets the default type url for SettingsRequest
         * @function getTypeUrl
         * @memberof auth.SettingsRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        SettingsRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/auth.SettingsRequest";
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
                case 1: {
                        if (!(message.items && message.items.length))
                            message.items = [];
                        message.items.push($root.auth.SettingsResponse.OidcSetting.decode(reader, reader.uint32()));
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for SettingsResponse
         * @function getTypeUrl
         * @memberof auth.SettingsResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        SettingsResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/auth.SettingsResponse";
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
                    case 1: {
                            message.enabled = reader.bool();
                            break;
                        }
                    case 2: {
                            message.name = reader.string();
                            break;
                        }
                    case 3: {
                            message.url = reader.string();
                            break;
                        }
                    case 4: {
                            message.end_session_endpoint = reader.string();
                            break;
                        }
                    case 5: {
                            message.state = reader.string();
                            break;
                        }
                    default:
                        reader.skipType(tag & 7);
                        break;
                    }
                }
                return message;
            };

            /**
             * Gets the default type url for OidcSetting
             * @function getTypeUrl
             * @memberof auth.SettingsResponse.OidcSetting
             * @static
             * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
             * @returns {string} The default type url
             */
            OidcSetting.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
                if (typeUrlPrefix === undefined) {
                    typeUrlPrefix = "type.googleapis.com";
                }
                return typeUrlPrefix + "/auth.SettingsResponse.OidcSetting";
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
                case 1: {
                        message.project_id = reader.int64();
                        break;
                    }
                case 2: {
                        message.only_changed = reader.bool();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for ShowRequest
         * @function getTypeUrl
         * @memberof changelog.ShowRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        ShowRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/changelog.ShowRequest";
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
                case 1: {
                        if (!(message.items && message.items.length))
                            message.items = [];
                        message.items.push($root.types.ChangelogModel.decode(reader, reader.uint32()));
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for ShowResponse
         * @function getTypeUrl
         * @memberof changelog.ShowResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        ShowResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/changelog.ShowResponse";
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
                case 1: {
                        message.status = reader.string();
                        break;
                    }
                case 2: {
                        message.free_memory = reader.string();
                        break;
                    }
                case 3: {
                        message.free_cpu = reader.string();
                        break;
                    }
                case 4: {
                        message.free_request_memory = reader.string();
                        break;
                    }
                case 5: {
                        message.free_request_cpu = reader.string();
                        break;
                    }
                case 6: {
                        message.total_memory = reader.string();
                        break;
                    }
                case 7: {
                        message.total_cpu = reader.string();
                        break;
                    }
                case 8: {
                        message.usage_memory_rate = reader.string();
                        break;
                    }
                case 9: {
                        message.usage_cpu_rate = reader.string();
                        break;
                    }
                case 10: {
                        message.request_memory_rate = reader.string();
                        break;
                    }
                case 11: {
                        message.request_cpu_rate = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for InfoResponse
         * @function getTypeUrl
         * @memberof cluster.InfoResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        InfoResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/cluster.InfoResponse";
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

        /**
         * Gets the default type url for InfoRequest
         * @function getTypeUrl
         * @memberof cluster.InfoRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        InfoRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/cluster.InfoRequest";
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
                case 1: {
                        message.file_id = reader.int64();
                        break;
                    }
                case 2: {
                        message.namespace = reader.string();
                        break;
                    }
                case 3: {
                        message.pod = reader.string();
                        break;
                    }
                case 4: {
                        message.container = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for CopyToPodRequest
         * @function getTypeUrl
         * @memberof container.CopyToPodRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        CopyToPodRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/container.CopyToPodRequest";
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
                case 1: {
                        message.pod_file_path = reader.string();
                        break;
                    }
                case 2: {
                        message.output = reader.string();
                        break;
                    }
                case 3: {
                        message.file_name = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for CopyToPodResponse
         * @function getTypeUrl
         * @memberof container.CopyToPodResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        CopyToPodResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/container.CopyToPodResponse";
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
                case 1: {
                        message.namespace = reader.string();
                        break;
                    }
                case 2: {
                        message.pod = reader.string();
                        break;
                    }
                case 3: {
                        message.container = reader.string();
                        break;
                    }
                case 4: {
                        if (!(message.command && message.command.length))
                            message.command = [];
                        message.command.push(reader.string());
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for ExecRequest
         * @function getTypeUrl
         * @memberof container.ExecRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        ExecRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/container.ExecRequest";
        };

        return ExecRequest;
    })();

    container.ExecError = (function() {

        /**
         * Properties of an ExecError.
         * @memberof container
         * @interface IExecError
         * @property {number|null} [code] ExecError code
         * @property {string|null} [message] ExecError message
         */

        /**
         * Constructs a new ExecError.
         * @memberof container
         * @classdesc Represents an ExecError.
         * @implements IExecError
         * @constructor
         * @param {container.IExecError=} [properties] Properties to set
         */
        function ExecError(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * ExecError code.
         * @member {number} code
         * @memberof container.ExecError
         * @instance
         */
        ExecError.prototype.code = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * ExecError message.
         * @member {string} message
         * @memberof container.ExecError
         * @instance
         */
        ExecError.prototype.message = "";

        /**
         * Encodes the specified ExecError message. Does not implicitly {@link container.ExecError.verify|verify} messages.
         * @function encode
         * @memberof container.ExecError
         * @static
         * @param {container.ExecError} message ExecError message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ExecError.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.code != null && Object.hasOwnProperty.call(message, "code"))
                writer.uint32(/* id 1, wireType 0 =*/8).int64(message.code);
            if (message.message != null && Object.hasOwnProperty.call(message, "message"))
                writer.uint32(/* id 2, wireType 2 =*/18).string(message.message);
            return writer;
        };

        /**
         * Decodes an ExecError message from the specified reader or buffer.
         * @function decode
         * @memberof container.ExecError
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {container.ExecError} ExecError
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ExecError.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.container.ExecError();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1: {
                        message.code = reader.int64();
                        break;
                    }
                case 2: {
                        message.message = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for ExecError
         * @function getTypeUrl
         * @memberof container.ExecError
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        ExecError.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/container.ExecError";
        };

        return ExecError;
    })();

    container.ExecResponse = (function() {

        /**
         * Properties of an ExecResponse.
         * @memberof container
         * @interface IExecResponse
         * @property {string|null} [message] ExecResponse message
         * @property {container.ExecError|null} [error] ExecResponse error
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
         * ExecResponse error.
         * @member {container.ExecError|null|undefined} error
         * @memberof container.ExecResponse
         * @instance
         */
        ExecResponse.prototype.error = null;

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
            if (message.error != null && Object.hasOwnProperty.call(message, "error"))
                $root.container.ExecError.encode(message.error, writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim();
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
                case 1: {
                        message.message = reader.string();
                        break;
                    }
                case 2: {
                        message.error = $root.container.ExecError.decode(reader, reader.uint32());
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for ExecResponse
         * @function getTypeUrl
         * @memberof container.ExecResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        ExecResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/container.ExecResponse";
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
                case 1: {
                        message.file_name = reader.string();
                        break;
                    }
                case 2: {
                        message.data = reader.bytes();
                        break;
                    }
                case 3: {
                        message.namespace = reader.string();
                        break;
                    }
                case 4: {
                        message.pod = reader.string();
                        break;
                    }
                case 5: {
                        message.container = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for StreamCopyToPodRequest
         * @function getTypeUrl
         * @memberof container.StreamCopyToPodRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        StreamCopyToPodRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/container.StreamCopyToPodRequest";
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
                case 1: {
                        message.size = reader.int64();
                        break;
                    }
                case 2: {
                        message.pod_file_path = reader.string();
                        break;
                    }
                case 3: {
                        message.output = reader.string();
                        break;
                    }
                case 4: {
                        message.pod = reader.string();
                        break;
                    }
                case 5: {
                        message.namespace = reader.string();
                        break;
                    }
                case 6: {
                        message.container = reader.string();
                        break;
                    }
                case 7: {
                        message.filename = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for StreamCopyToPodResponse
         * @function getTypeUrl
         * @memberof container.StreamCopyToPodResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        StreamCopyToPodResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/container.StreamCopyToPodResponse";
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
                case 1: {
                        message.namespace = reader.string();
                        break;
                    }
                case 2: {
                        message.pod = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for IsPodRunningRequest
         * @function getTypeUrl
         * @memberof container.IsPodRunningRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        IsPodRunningRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/container.IsPodRunningRequest";
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
                case 1: {
                        message.running = reader.bool();
                        break;
                    }
                case 2: {
                        message.reason = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for IsPodRunningResponse
         * @function getTypeUrl
         * @memberof container.IsPodRunningResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        IsPodRunningResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/container.IsPodRunningResponse";
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
                case 1: {
                        message.namespace = reader.string();
                        break;
                    }
                case 2: {
                        message.pod = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for IsPodExistsRequest
         * @function getTypeUrl
         * @memberof container.IsPodExistsRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        IsPodExistsRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/container.IsPodExistsRequest";
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
                case 1: {
                        message.exists = reader.bool();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for IsPodExistsResponse
         * @function getTypeUrl
         * @memberof container.IsPodExistsResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        IsPodExistsResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/container.IsPodExistsResponse";
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
                case 1: {
                        message.namespace = reader.string();
                        break;
                    }
                case 2: {
                        message.pod = reader.string();
                        break;
                    }
                case 3: {
                        message.container = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for LogRequest
         * @function getTypeUrl
         * @memberof container.LogRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        LogRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/container.LogRequest";
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
                case 1: {
                        message.namespace = reader.string();
                        break;
                    }
                case 2: {
                        message.pod_name = reader.string();
                        break;
                    }
                case 3: {
                        message.container_name = reader.string();
                        break;
                    }
                case 4: {
                        message.log = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for LogResponse
         * @function getTypeUrl
         * @memberof container.LogResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        LogResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/container.LogResponse";
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
                case 1: {
                        message.namespace_id = reader.int64();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for InNamespaceRequest
         * @function getTypeUrl
         * @memberof endpoint.InNamespaceRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        InNamespaceRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/endpoint.InNamespaceRequest";
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
                case 1: {
                        if (!(message.items && message.items.length))
                            message.items = [];
                        message.items.push($root.types.ServiceEndpoint.decode(reader, reader.uint32()));
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for InNamespaceResponse
         * @function getTypeUrl
         * @memberof endpoint.InNamespaceResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        InNamespaceResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/endpoint.InNamespaceResponse";
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
                case 1: {
                        message.project_id = reader.int64();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for InProjectRequest
         * @function getTypeUrl
         * @memberof endpoint.InProjectRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        InProjectRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/endpoint.InProjectRequest";
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
                case 1: {
                        if (!(message.items && message.items.length))
                            message.items = [];
                        message.items.push($root.types.ServiceEndpoint.decode(reader, reader.uint32()));
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for InProjectResponse
         * @function getTypeUrl
         * @memberof endpoint.InProjectResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        InProjectResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/endpoint.InProjectResponse";
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
         * @property {types.EventActionType|null} [action_type] ListRequest action_type
         * @property {string|null} [search] ListRequest search
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
         * ListRequest action_type.
         * @member {types.EventActionType} action_type
         * @memberof event.ListRequest
         * @instance
         */
        ListRequest.prototype.action_type = 0;

        /**
         * ListRequest search.
         * @member {string} search
         * @memberof event.ListRequest
         * @instance
         */
        ListRequest.prototype.search = "";

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
            if (message.action_type != null && Object.hasOwnProperty.call(message, "action_type"))
                writer.uint32(/* id 3, wireType 0 =*/24).int32(message.action_type);
            if (message.search != null && Object.hasOwnProperty.call(message, "search"))
                writer.uint32(/* id 4, wireType 2 =*/34).string(message.search);
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
                case 1: {
                        message.page = reader.int64();
                        break;
                    }
                case 2: {
                        message.page_size = reader.int64();
                        break;
                    }
                case 3: {
                        message.action_type = reader.int32();
                        break;
                    }
                case 4: {
                        message.search = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for ListRequest
         * @function getTypeUrl
         * @memberof event.ListRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        ListRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/event.ListRequest";
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
                case 1: {
                        message.page = reader.int64();
                        break;
                    }
                case 2: {
                        message.page_size = reader.int64();
                        break;
                    }
                case 3: {
                        if (!(message.items && message.items.length))
                            message.items = [];
                        message.items.push($root.types.EventModel.decode(reader, reader.uint32()));
                        break;
                    }
                case 4: {
                        message.count = reader.int64();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for ListResponse
         * @function getTypeUrl
         * @memberof event.ListResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        ListResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/event.ListResponse";
        };

        return ListResponse;
    })();

    event.ShowRequest = (function() {

        /**
         * Properties of a ShowRequest.
         * @memberof event
         * @interface IShowRequest
         * @property {number|null} [id] ShowRequest id
         */

        /**
         * Constructs a new ShowRequest.
         * @memberof event
         * @classdesc Represents a ShowRequest.
         * @implements IShowRequest
         * @constructor
         * @param {event.IShowRequest=} [properties] Properties to set
         */
        function ShowRequest(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * ShowRequest id.
         * @member {number} id
         * @memberof event.ShowRequest
         * @instance
         */
        ShowRequest.prototype.id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Encodes the specified ShowRequest message. Does not implicitly {@link event.ShowRequest.verify|verify} messages.
         * @function encode
         * @memberof event.ShowRequest
         * @static
         * @param {event.ShowRequest} message ShowRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ShowRequest.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.id != null && Object.hasOwnProperty.call(message, "id"))
                writer.uint32(/* id 1, wireType 0 =*/8).int64(message.id);
            return writer;
        };

        /**
         * Decodes a ShowRequest message from the specified reader or buffer.
         * @function decode
         * @memberof event.ShowRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {event.ShowRequest} ShowRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ShowRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.event.ShowRequest();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1: {
                        message.id = reader.int64();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for ShowRequest
         * @function getTypeUrl
         * @memberof event.ShowRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        ShowRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/event.ShowRequest";
        };

        return ShowRequest;
    })();

    event.ShowResponse = (function() {

        /**
         * Properties of a ShowResponse.
         * @memberof event
         * @interface IShowResponse
         * @property {types.EventModel|null} [event] ShowResponse event
         */

        /**
         * Constructs a new ShowResponse.
         * @memberof event
         * @classdesc Represents a ShowResponse.
         * @implements IShowResponse
         * @constructor
         * @param {event.IShowResponse=} [properties] Properties to set
         */
        function ShowResponse(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * ShowResponse event.
         * @member {types.EventModel|null|undefined} event
         * @memberof event.ShowResponse
         * @instance
         */
        ShowResponse.prototype.event = null;

        /**
         * Encodes the specified ShowResponse message. Does not implicitly {@link event.ShowResponse.verify|verify} messages.
         * @function encode
         * @memberof event.ShowResponse
         * @static
         * @param {event.ShowResponse} message ShowResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ShowResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.event != null && Object.hasOwnProperty.call(message, "event"))
                $root.types.EventModel.encode(message.event, writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
            return writer;
        };

        /**
         * Decodes a ShowResponse message from the specified reader or buffer.
         * @function decode
         * @memberof event.ShowResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {event.ShowResponse} ShowResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ShowResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.event.ShowResponse();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1: {
                        message.event = $root.types.EventModel.decode(reader, reader.uint32());
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for ShowResponse
         * @function getTypeUrl
         * @memberof event.ShowResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        ShowResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/event.ShowResponse";
        };

        return ShowResponse;
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

        /**
         * Callback as used by {@link event.Event#show}.
         * @memberof event.Event
         * @typedef ShowCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {event.ShowResponse} [response] ShowResponse
         */

        /**
         * Calls Show.
         * @function show
         * @memberof event.Event
         * @instance
         * @param {event.ShowRequest} request ShowRequest message or plain object
         * @param {event.Event.ShowCallback} callback Node-style callback called with the error, if any, and ShowResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(Event.prototype.show = function show(request, callback) {
            return this.rpcCall(show, $root.event.ShowRequest, $root.event.ShowResponse, request, callback);
        }, "name", { value: "Show" });

        /**
         * Calls Show.
         * @function show
         * @memberof event.Event
         * @instance
         * @param {event.ShowRequest} request ShowRequest message or plain object
         * @returns {Promise<event.ShowResponse>} Promise
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
                case 1: {
                        message.id = reader.int64();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for DeleteRequest
         * @function getTypeUrl
         * @memberof file.DeleteRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        DeleteRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/file.DeleteRequest";
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
                case 1: {
                        message.file = $root.types.FileModel.decode(reader, reader.uint32());
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for DeleteResponse
         * @function getTypeUrl
         * @memberof file.DeleteResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        DeleteResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/file.DeleteResponse";
        };

        return DeleteResponse;
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

        /**
         * Gets the default type url for DiskInfoRequest
         * @function getTypeUrl
         * @memberof file.DiskInfoRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        DiskInfoRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/file.DiskInfoRequest";
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
                case 1: {
                        message.usage = reader.int64();
                        break;
                    }
                case 2: {
                        message.humanize_usage = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for DiskInfoResponse
         * @function getTypeUrl
         * @memberof file.DiskInfoResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        DiskInfoResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/file.DiskInfoResponse";
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
                case 1: {
                        message.page = reader.int64();
                        break;
                    }
                case 2: {
                        message.page_size = reader.int64();
                        break;
                    }
                case 3: {
                        message.without_deleted = reader.bool();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for ListRequest
         * @function getTypeUrl
         * @memberof file.ListRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        ListRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/file.ListRequest";
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
                case 1: {
                        message.page = reader.int64();
                        break;
                    }
                case 2: {
                        message.page_size = reader.int64();
                        break;
                    }
                case 3: {
                        if (!(message.items && message.items.length))
                            message.items = [];
                        message.items.push($root.types.FileModel.decode(reader, reader.uint32()));
                        break;
                    }
                case 4: {
                        message.count = reader.int64();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for ListResponse
         * @function getTypeUrl
         * @memberof file.ListResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        ListResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/file.ListResponse";
        };

        return ListResponse;
    })();

    file.MaxUploadSizeRequest = (function() {

        /**
         * Properties of a MaxUploadSizeRequest.
         * @memberof file
         * @interface IMaxUploadSizeRequest
         */

        /**
         * Constructs a new MaxUploadSizeRequest.
         * @memberof file
         * @classdesc Represents a MaxUploadSizeRequest.
         * @implements IMaxUploadSizeRequest
         * @constructor
         * @param {file.IMaxUploadSizeRequest=} [properties] Properties to set
         */
        function MaxUploadSizeRequest(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * Encodes the specified MaxUploadSizeRequest message. Does not implicitly {@link file.MaxUploadSizeRequest.verify|verify} messages.
         * @function encode
         * @memberof file.MaxUploadSizeRequest
         * @static
         * @param {file.MaxUploadSizeRequest} message MaxUploadSizeRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        MaxUploadSizeRequest.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            return writer;
        };

        /**
         * Decodes a MaxUploadSizeRequest message from the specified reader or buffer.
         * @function decode
         * @memberof file.MaxUploadSizeRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {file.MaxUploadSizeRequest} MaxUploadSizeRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        MaxUploadSizeRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.file.MaxUploadSizeRequest();
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

        /**
         * Gets the default type url for MaxUploadSizeRequest
         * @function getTypeUrl
         * @memberof file.MaxUploadSizeRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        MaxUploadSizeRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/file.MaxUploadSizeRequest";
        };

        return MaxUploadSizeRequest;
    })();

    file.MaxUploadSizeResponse = (function() {

        /**
         * Properties of a MaxUploadSizeResponse.
         * @memberof file
         * @interface IMaxUploadSizeResponse
         * @property {string|null} [humanize_size] MaxUploadSizeResponse humanize_size
         * @property {number|null} [bytes] MaxUploadSizeResponse bytes
         */

        /**
         * Constructs a new MaxUploadSizeResponse.
         * @memberof file
         * @classdesc Represents a MaxUploadSizeResponse.
         * @implements IMaxUploadSizeResponse
         * @constructor
         * @param {file.IMaxUploadSizeResponse=} [properties] Properties to set
         */
        function MaxUploadSizeResponse(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * MaxUploadSizeResponse humanize_size.
         * @member {string} humanize_size
         * @memberof file.MaxUploadSizeResponse
         * @instance
         */
        MaxUploadSizeResponse.prototype.humanize_size = "";

        /**
         * MaxUploadSizeResponse bytes.
         * @member {number} bytes
         * @memberof file.MaxUploadSizeResponse
         * @instance
         */
        MaxUploadSizeResponse.prototype.bytes = $util.Long ? $util.Long.fromBits(0,0,true) : 0;

        /**
         * Encodes the specified MaxUploadSizeResponse message. Does not implicitly {@link file.MaxUploadSizeResponse.verify|verify} messages.
         * @function encode
         * @memberof file.MaxUploadSizeResponse
         * @static
         * @param {file.MaxUploadSizeResponse} message MaxUploadSizeResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        MaxUploadSizeResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.humanize_size != null && Object.hasOwnProperty.call(message, "humanize_size"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.humanize_size);
            if (message.bytes != null && Object.hasOwnProperty.call(message, "bytes"))
                writer.uint32(/* id 2, wireType 0 =*/16).uint64(message.bytes);
            return writer;
        };

        /**
         * Decodes a MaxUploadSizeResponse message from the specified reader or buffer.
         * @function decode
         * @memberof file.MaxUploadSizeResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {file.MaxUploadSizeResponse} MaxUploadSizeResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        MaxUploadSizeResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.file.MaxUploadSizeResponse();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1: {
                        message.humanize_size = reader.string();
                        break;
                    }
                case 2: {
                        message.bytes = reader.uint64();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for MaxUploadSizeResponse
         * @function getTypeUrl
         * @memberof file.MaxUploadSizeResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        MaxUploadSizeResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/file.MaxUploadSizeResponse";
        };

        return MaxUploadSizeResponse;
    })();

    file.ShowRecordsRequest = (function() {

        /**
         * Properties of a ShowRecordsRequest.
         * @memberof file
         * @interface IShowRecordsRequest
         * @property {number|null} [id] ShowRecordsRequest id
         */

        /**
         * Constructs a new ShowRecordsRequest.
         * @memberof file
         * @classdesc Represents a ShowRecordsRequest.
         * @implements IShowRecordsRequest
         * @constructor
         * @param {file.IShowRecordsRequest=} [properties] Properties to set
         */
        function ShowRecordsRequest(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * ShowRecordsRequest id.
         * @member {number} id
         * @memberof file.ShowRecordsRequest
         * @instance
         */
        ShowRecordsRequest.prototype.id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Encodes the specified ShowRecordsRequest message. Does not implicitly {@link file.ShowRecordsRequest.verify|verify} messages.
         * @function encode
         * @memberof file.ShowRecordsRequest
         * @static
         * @param {file.ShowRecordsRequest} message ShowRecordsRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ShowRecordsRequest.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.id != null && Object.hasOwnProperty.call(message, "id"))
                writer.uint32(/* id 1, wireType 0 =*/8).int64(message.id);
            return writer;
        };

        /**
         * Decodes a ShowRecordsRequest message from the specified reader or buffer.
         * @function decode
         * @memberof file.ShowRecordsRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {file.ShowRecordsRequest} ShowRecordsRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ShowRecordsRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.file.ShowRecordsRequest();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1: {
                        message.id = reader.int64();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for ShowRecordsRequest
         * @function getTypeUrl
         * @memberof file.ShowRecordsRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        ShowRecordsRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/file.ShowRecordsRequest";
        };

        return ShowRecordsRequest;
    })();

    file.ShowRecordsResponse = (function() {

        /**
         * Properties of a ShowRecordsResponse.
         * @memberof file
         * @interface IShowRecordsResponse
         * @property {Array.<string>|null} [items] ShowRecordsResponse items
         */

        /**
         * Constructs a new ShowRecordsResponse.
         * @memberof file
         * @classdesc Represents a ShowRecordsResponse.
         * @implements IShowRecordsResponse
         * @constructor
         * @param {file.IShowRecordsResponse=} [properties] Properties to set
         */
        function ShowRecordsResponse(properties) {
            this.items = [];
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * ShowRecordsResponse items.
         * @member {Array.<string>} items
         * @memberof file.ShowRecordsResponse
         * @instance
         */
        ShowRecordsResponse.prototype.items = $util.emptyArray;

        /**
         * Encodes the specified ShowRecordsResponse message. Does not implicitly {@link file.ShowRecordsResponse.verify|verify} messages.
         * @function encode
         * @memberof file.ShowRecordsResponse
         * @static
         * @param {file.ShowRecordsResponse} message ShowRecordsResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ShowRecordsResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.items != null && message.items.length)
                for (let i = 0; i < message.items.length; ++i)
                    writer.uint32(/* id 1, wireType 2 =*/10).string(message.items[i]);
            return writer;
        };

        /**
         * Decodes a ShowRecordsResponse message from the specified reader or buffer.
         * @function decode
         * @memberof file.ShowRecordsResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {file.ShowRecordsResponse} ShowRecordsResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ShowRecordsResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.file.ShowRecordsResponse();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1: {
                        if (!(message.items && message.items.length))
                            message.items = [];
                        message.items.push(reader.string());
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for ShowRecordsResponse
         * @function getTypeUrl
         * @memberof file.ShowRecordsResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        ShowRecordsResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/file.ShowRecordsResponse";
        };

        return ShowRecordsResponse;
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
         * Callback as used by {@link file.File#showRecords}.
         * @memberof file.File
         * @typedef ShowRecordsCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {file.ShowRecordsResponse} [response] ShowRecordsResponse
         */

        /**
         * Calls ShowRecords.
         * @function showRecords
         * @memberof file.File
         * @instance
         * @param {file.ShowRecordsRequest} request ShowRecordsRequest message or plain object
         * @param {file.File.ShowRecordsCallback} callback Node-style callback called with the error, if any, and ShowRecordsResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(File.prototype.showRecords = function showRecords(request, callback) {
            return this.rpcCall(showRecords, $root.file.ShowRecordsRequest, $root.file.ShowRecordsResponse, request, callback);
        }, "name", { value: "ShowRecords" });

        /**
         * Calls ShowRecords.
         * @function showRecords
         * @memberof file.File
         * @instance
         * @param {file.ShowRecordsRequest} request ShowRecordsRequest message or plain object
         * @returns {Promise<file.ShowRecordsResponse>} Promise
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

        /**
         * Callback as used by {@link file.File#maxUploadSize}.
         * @memberof file.File
         * @typedef MaxUploadSizeCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {file.MaxUploadSizeResponse} [response] MaxUploadSizeResponse
         */

        /**
         * Calls MaxUploadSize.
         * @function maxUploadSize
         * @memberof file.File
         * @instance
         * @param {file.MaxUploadSizeRequest} request MaxUploadSizeRequest message or plain object
         * @param {file.File.MaxUploadSizeCallback} callback Node-style callback called with the error, if any, and MaxUploadSizeResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(File.prototype.maxUploadSize = function maxUploadSize(request, callback) {
            return this.rpcCall(maxUploadSize, $root.file.MaxUploadSizeRequest, $root.file.MaxUploadSizeResponse, request, callback);
        }, "name", { value: "MaxUploadSize" });

        /**
         * Calls MaxUploadSize.
         * @function maxUploadSize
         * @memberof file.File
         * @instance
         * @param {file.MaxUploadSizeRequest} request MaxUploadSizeRequest message or plain object
         * @returns {Promise<file.MaxUploadSizeResponse>} Promise
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
                case 1: {
                        message.git_project_id = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for EnableProjectRequest
         * @function getTypeUrl
         * @memberof git.EnableProjectRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        EnableProjectRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/git.EnableProjectRequest";
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
                case 1: {
                        message.git_project_id = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for DisableProjectRequest
         * @function getTypeUrl
         * @memberof git.DisableProjectRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        DisableProjectRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/git.DisableProjectRequest";
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
                case 1: {
                        message.id = reader.int64();
                        break;
                    }
                case 2: {
                        message.name = reader.string();
                        break;
                    }
                case 3: {
                        message.path = reader.string();
                        break;
                    }
                case 4: {
                        message.web_url = reader.string();
                        break;
                    }
                case 5: {
                        message.avatar_url = reader.string();
                        break;
                    }
                case 6: {
                        message.description = reader.string();
                        break;
                    }
                case 7: {
                        message.enabled = reader.bool();
                        break;
                    }
                case 8: {
                        message.global_enabled = reader.bool();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for ProjectItem
         * @function getTypeUrl
         * @memberof git.ProjectItem
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        ProjectItem.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/git.ProjectItem";
        };

        return ProjectItem;
    })();

    git.AllResponse = (function() {

        /**
         * Properties of an AllResponse.
         * @memberof git
         * @interface IAllResponse
         * @property {Array.<git.ProjectItem>|null} [items] AllResponse items
         */

        /**
         * Constructs a new AllResponse.
         * @memberof git
         * @classdesc Represents an AllResponse.
         * @implements IAllResponse
         * @constructor
         * @param {git.IAllResponse=} [properties] Properties to set
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
         * @member {Array.<git.ProjectItem>} items
         * @memberof git.AllResponse
         * @instance
         */
        AllResponse.prototype.items = $util.emptyArray;

        /**
         * Encodes the specified AllResponse message. Does not implicitly {@link git.AllResponse.verify|verify} messages.
         * @function encode
         * @memberof git.AllResponse
         * @static
         * @param {git.AllResponse} message AllResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        AllResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.items != null && message.items.length)
                for (let i = 0; i < message.items.length; ++i)
                    $root.git.ProjectItem.encode(message.items[i], writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
            return writer;
        };

        /**
         * Decodes an AllResponse message from the specified reader or buffer.
         * @function decode
         * @memberof git.AllResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {git.AllResponse} AllResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        AllResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.git.AllResponse();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1: {
                        if (!(message.items && message.items.length))
                            message.items = [];
                        message.items.push($root.git.ProjectItem.decode(reader, reader.uint32()));
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for AllResponse
         * @function getTypeUrl
         * @memberof git.AllResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        AllResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/git.AllResponse";
        };

        return AllResponse;
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
         * @property {string|null} [display_name] Option display_name
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
         * Option display_name.
         * @member {string} display_name
         * @memberof git.Option
         * @instance
         */
        Option.prototype.display_name = "";

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
            if (message.display_name != null && Object.hasOwnProperty.call(message, "display_name"))
                writer.uint32(/* id 7, wireType 2 =*/58).string(message.display_name);
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
                case 1: {
                        message.value = reader.string();
                        break;
                    }
                case 2: {
                        message.label = reader.string();
                        break;
                    }
                case 3: {
                        message.type = reader.string();
                        break;
                    }
                case 4: {
                        message.isLeaf = reader.bool();
                        break;
                    }
                case 5: {
                        message.gitProjectId = reader.string();
                        break;
                    }
                case 6: {
                        message.branch = reader.string();
                        break;
                    }
                case 7: {
                        message.display_name = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for Option
         * @function getTypeUrl
         * @memberof git.Option
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        Option.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/git.Option";
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
                case 1: {
                        if (!(message.items && message.items.length))
                            message.items = [];
                        message.items.push($root.git.Option.decode(reader, reader.uint32()));
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for ProjectOptionsResponse
         * @function getTypeUrl
         * @memberof git.ProjectOptionsResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        ProjectOptionsResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/git.ProjectOptionsResponse";
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
                case 1: {
                        message.git_project_id = reader.string();
                        break;
                    }
                case 2: {
                        message.all = reader.bool();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for BranchOptionsRequest
         * @function getTypeUrl
         * @memberof git.BranchOptionsRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        BranchOptionsRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/git.BranchOptionsRequest";
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
                case 1: {
                        if (!(message.items && message.items.length))
                            message.items = [];
                        message.items.push($root.git.Option.decode(reader, reader.uint32()));
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for BranchOptionsResponse
         * @function getTypeUrl
         * @memberof git.BranchOptionsResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        BranchOptionsResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/git.BranchOptionsResponse";
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
                case 1: {
                        message.git_project_id = reader.string();
                        break;
                    }
                case 2: {
                        message.branch = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for CommitOptionsRequest
         * @function getTypeUrl
         * @memberof git.CommitOptionsRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        CommitOptionsRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/git.CommitOptionsRequest";
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
                case 1: {
                        if (!(message.items && message.items.length))
                            message.items = [];
                        message.items.push($root.git.Option.decode(reader, reader.uint32()));
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for CommitOptionsResponse
         * @function getTypeUrl
         * @memberof git.CommitOptionsResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        CommitOptionsResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/git.CommitOptionsResponse";
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
                case 1: {
                        message.git_project_id = reader.string();
                        break;
                    }
                case 2: {
                        message.branch = reader.string();
                        break;
                    }
                case 3: {
                        message.commit = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for CommitRequest
         * @function getTypeUrl
         * @memberof git.CommitRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        CommitRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/git.CommitRequest";
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
                case 1: {
                        message.id = reader.string();
                        break;
                    }
                case 2: {
                        message.short_id = reader.string();
                        break;
                    }
                case 3: {
                        message.git_project_id = reader.string();
                        break;
                    }
                case 4: {
                        message.label = reader.string();
                        break;
                    }
                case 5: {
                        message.title = reader.string();
                        break;
                    }
                case 6: {
                        message.branch = reader.string();
                        break;
                    }
                case 7: {
                        message.author_name = reader.string();
                        break;
                    }
                case 8: {
                        message.author_email = reader.string();
                        break;
                    }
                case 9: {
                        message.committer_name = reader.string();
                        break;
                    }
                case 10: {
                        message.committer_email = reader.string();
                        break;
                    }
                case 11: {
                        message.web_url = reader.string();
                        break;
                    }
                case 12: {
                        message.message = reader.string();
                        break;
                    }
                case 13: {
                        message.committed_date = reader.string();
                        break;
                    }
                case 14: {
                        message.created_at = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for CommitResponse
         * @function getTypeUrl
         * @memberof git.CommitResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        CommitResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/git.CommitResponse";
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
                case 1: {
                        message.git_project_id = reader.string();
                        break;
                    }
                case 2: {
                        message.branch = reader.string();
                        break;
                    }
                case 3: {
                        message.commit = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for PipelineInfoRequest
         * @function getTypeUrl
         * @memberof git.PipelineInfoRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        PipelineInfoRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/git.PipelineInfoRequest";
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
                case 1: {
                        message.status = reader.string();
                        break;
                    }
                case 2: {
                        message.web_url = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for PipelineInfoResponse
         * @function getTypeUrl
         * @memberof git.PipelineInfoResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        PipelineInfoResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/git.PipelineInfoResponse";
        };

        return PipelineInfoResponse;
    })();

    git.MarsConfigFileRequest = (function() {

        /**
         * Properties of a MarsConfigFileRequest.
         * @memberof git
         * @interface IMarsConfigFileRequest
         * @property {string|null} [git_project_id] MarsConfigFileRequest git_project_id
         * @property {string|null} [branch] MarsConfigFileRequest branch
         */

        /**
         * Constructs a new MarsConfigFileRequest.
         * @memberof git
         * @classdesc Represents a MarsConfigFileRequest.
         * @implements IMarsConfigFileRequest
         * @constructor
         * @param {git.IMarsConfigFileRequest=} [properties] Properties to set
         */
        function MarsConfigFileRequest(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * MarsConfigFileRequest git_project_id.
         * @member {string} git_project_id
         * @memberof git.MarsConfigFileRequest
         * @instance
         */
        MarsConfigFileRequest.prototype.git_project_id = "";

        /**
         * MarsConfigFileRequest branch.
         * @member {string} branch
         * @memberof git.MarsConfigFileRequest
         * @instance
         */
        MarsConfigFileRequest.prototype.branch = "";

        /**
         * Encodes the specified MarsConfigFileRequest message. Does not implicitly {@link git.MarsConfigFileRequest.verify|verify} messages.
         * @function encode
         * @memberof git.MarsConfigFileRequest
         * @static
         * @param {git.MarsConfigFileRequest} message MarsConfigFileRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        MarsConfigFileRequest.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.git_project_id != null && Object.hasOwnProperty.call(message, "git_project_id"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.git_project_id);
            if (message.branch != null && Object.hasOwnProperty.call(message, "branch"))
                writer.uint32(/* id 2, wireType 2 =*/18).string(message.branch);
            return writer;
        };

        /**
         * Decodes a MarsConfigFileRequest message from the specified reader or buffer.
         * @function decode
         * @memberof git.MarsConfigFileRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {git.MarsConfigFileRequest} MarsConfigFileRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        MarsConfigFileRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.git.MarsConfigFileRequest();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1: {
                        message.git_project_id = reader.string();
                        break;
                    }
                case 2: {
                        message.branch = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for MarsConfigFileRequest
         * @function getTypeUrl
         * @memberof git.MarsConfigFileRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        MarsConfigFileRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/git.MarsConfigFileRequest";
        };

        return MarsConfigFileRequest;
    })();

    git.MarsConfigFileResponse = (function() {

        /**
         * Properties of a MarsConfigFileResponse.
         * @memberof git
         * @interface IMarsConfigFileResponse
         * @property {string|null} [data] MarsConfigFileResponse data
         * @property {string|null} [type] MarsConfigFileResponse type
         * @property {Array.<mars.Element>|null} [elements] MarsConfigFileResponse elements
         */

        /**
         * Constructs a new MarsConfigFileResponse.
         * @memberof git
         * @classdesc Represents a MarsConfigFileResponse.
         * @implements IMarsConfigFileResponse
         * @constructor
         * @param {git.IMarsConfigFileResponse=} [properties] Properties to set
         */
        function MarsConfigFileResponse(properties) {
            this.elements = [];
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * MarsConfigFileResponse data.
         * @member {string} data
         * @memberof git.MarsConfigFileResponse
         * @instance
         */
        MarsConfigFileResponse.prototype.data = "";

        /**
         * MarsConfigFileResponse type.
         * @member {string} type
         * @memberof git.MarsConfigFileResponse
         * @instance
         */
        MarsConfigFileResponse.prototype.type = "";

        /**
         * MarsConfigFileResponse elements.
         * @member {Array.<mars.Element>} elements
         * @memberof git.MarsConfigFileResponse
         * @instance
         */
        MarsConfigFileResponse.prototype.elements = $util.emptyArray;

        /**
         * Encodes the specified MarsConfigFileResponse message. Does not implicitly {@link git.MarsConfigFileResponse.verify|verify} messages.
         * @function encode
         * @memberof git.MarsConfigFileResponse
         * @static
         * @param {git.MarsConfigFileResponse} message MarsConfigFileResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        MarsConfigFileResponse.encode = function encode(message, writer) {
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
         * Decodes a MarsConfigFileResponse message from the specified reader or buffer.
         * @function decode
         * @memberof git.MarsConfigFileResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {git.MarsConfigFileResponse} MarsConfigFileResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        MarsConfigFileResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.git.MarsConfigFileResponse();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1: {
                        message.data = reader.string();
                        break;
                    }
                case 2: {
                        message.type = reader.string();
                        break;
                    }
                case 3: {
                        if (!(message.elements && message.elements.length))
                            message.elements = [];
                        message.elements.push($root.mars.Element.decode(reader, reader.uint32()));
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for MarsConfigFileResponse
         * @function getTypeUrl
         * @memberof git.MarsConfigFileResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        MarsConfigFileResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/git.MarsConfigFileResponse";
        };

        return MarsConfigFileResponse;
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

        /**
         * Gets the default type url for EnableProjectResponse
         * @function getTypeUrl
         * @memberof git.EnableProjectResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        EnableProjectResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/git.EnableProjectResponse";
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

        /**
         * Gets the default type url for DisableProjectResponse
         * @function getTypeUrl
         * @memberof git.DisableProjectResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        DisableProjectResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/git.DisableProjectResponse";
        };

        return DisableProjectResponse;
    })();

    git.AllRequest = (function() {

        /**
         * Properties of an AllRequest.
         * @memberof git
         * @interface IAllRequest
         */

        /**
         * Constructs a new AllRequest.
         * @memberof git
         * @classdesc Represents an AllRequest.
         * @implements IAllRequest
         * @constructor
         * @param {git.IAllRequest=} [properties] Properties to set
         */
        function AllRequest(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * Encodes the specified AllRequest message. Does not implicitly {@link git.AllRequest.verify|verify} messages.
         * @function encode
         * @memberof git.AllRequest
         * @static
         * @param {git.AllRequest} message AllRequest message or plain object to encode
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
         * @memberof git.AllRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {git.AllRequest} AllRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        AllRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.git.AllRequest();
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

        /**
         * Gets the default type url for AllRequest
         * @function getTypeUrl
         * @memberof git.AllRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        AllRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/git.AllRequest";
        };

        return AllRequest;
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

        /**
         * Gets the default type url for ProjectOptionsRequest
         * @function getTypeUrl
         * @memberof git.ProjectOptionsRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        ProjectOptionsRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/git.ProjectOptionsRequest";
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
         * @param {git.AllResponse} [response] AllResponse
         */

        /**
         * Calls All.
         * @function all
         * @memberof git.Git
         * @instance
         * @param {git.AllRequest} request AllRequest message or plain object
         * @param {git.Git.AllCallback} callback Node-style callback called with the error, if any, and AllResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(Git.prototype.all = function all(request, callback) {
            return this.rpcCall(all, $root.git.AllRequest, $root.git.AllResponse, request, callback);
        }, "name", { value: "All" });

        /**
         * Calls All.
         * @function all
         * @memberof git.Git
         * @instance
         * @param {git.AllRequest} request AllRequest message or plain object
         * @returns {Promise<git.AllResponse>} Promise
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
         * @param {git.MarsConfigFileResponse} [response] MarsConfigFileResponse
         */

        /**
         * Calls MarsConfigFile.
         * @function marsConfigFile
         * @memberof git.Git
         * @instance
         * @param {git.MarsConfigFileRequest} request MarsConfigFileRequest message or plain object
         * @param {git.Git.MarsConfigFileCallback} callback Node-style callback called with the error, if any, and MarsConfigFileResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(Git.prototype.marsConfigFile = function marsConfigFile(request, callback) {
            return this.rpcCall(marsConfigFile, $root.git.MarsConfigFileRequest, $root.git.MarsConfigFileResponse, request, callback);
        }, "name", { value: "MarsConfigFile" });

        /**
         * Calls MarsConfigFile.
         * @function marsConfigFile
         * @memberof git.Git
         * @instance
         * @param {git.MarsConfigFileRequest} request MarsConfigFileRequest message or plain object
         * @returns {Promise<git.MarsConfigFileResponse>} Promise
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
                case 1: {
                        message.git_project_id = reader.string();
                        break;
                    }
                case 2: {
                        message.branch = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for FileRequest
         * @function getTypeUrl
         * @memberof gitconfig.FileRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        FileRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/gitconfig.FileRequest";
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
                case 1: {
                        message.data = reader.string();
                        break;
                    }
                case 2: {
                        message.type = reader.string();
                        break;
                    }
                case 3: {
                        if (!(message.elements && message.elements.length))
                            message.elements = [];
                        message.elements.push($root.mars.Element.decode(reader, reader.uint32()));
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for FileResponse
         * @function getTypeUrl
         * @memberof gitconfig.FileResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        FileResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/gitconfig.FileResponse";
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
                case 1: {
                        message.git_project_id = reader.int64();
                        break;
                    }
                case 2: {
                        message.branch = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for ShowRequest
         * @function getTypeUrl
         * @memberof gitconfig.ShowRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        ShowRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/gitconfig.ShowRequest";
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
                case 1: {
                        message.branch = reader.string();
                        break;
                    }
                case 2: {
                        message.config = $root.mars.Config.decode(reader, reader.uint32());
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for ShowResponse
         * @function getTypeUrl
         * @memberof gitconfig.ShowResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        ShowResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/gitconfig.ShowResponse";
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
                case 1: {
                        message.git_project_id = reader.int64();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for GlobalConfigRequest
         * @function getTypeUrl
         * @memberof gitconfig.GlobalConfigRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        GlobalConfigRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/gitconfig.GlobalConfigRequest";
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
                case 1: {
                        message.enabled = reader.bool();
                        break;
                    }
                case 2: {
                        message.config = $root.mars.Config.decode(reader, reader.uint32());
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for GlobalConfigResponse
         * @function getTypeUrl
         * @memberof gitconfig.GlobalConfigResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        GlobalConfigResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/gitconfig.GlobalConfigResponse";
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
                case 1: {
                        message.git_project_id = reader.int64();
                        break;
                    }
                case 2: {
                        message.config = $root.mars.Config.decode(reader, reader.uint32());
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for UpdateRequest
         * @function getTypeUrl
         * @memberof gitconfig.UpdateRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        UpdateRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/gitconfig.UpdateRequest";
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
                case 1: {
                        message.config = $root.mars.Config.decode(reader, reader.uint32());
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for UpdateResponse
         * @function getTypeUrl
         * @memberof gitconfig.UpdateResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        UpdateResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/gitconfig.UpdateResponse";
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
                case 1: {
                        message.git_project_id = reader.int64();
                        break;
                    }
                case 2: {
                        message.enabled = reader.bool();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for ToggleGlobalStatusRequest
         * @function getTypeUrl
         * @memberof gitconfig.ToggleGlobalStatusRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        ToggleGlobalStatusRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/gitconfig.ToggleGlobalStatusRequest";
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
                case 1: {
                        message.git_project_id = reader.int64();
                        break;
                    }
                case 2: {
                        message.branch = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for DefaultChartValuesRequest
         * @function getTypeUrl
         * @memberof gitconfig.DefaultChartValuesRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        DefaultChartValuesRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/gitconfig.DefaultChartValuesRequest";
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
                case 1: {
                        message.value = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for DefaultChartValuesResponse
         * @function getTypeUrl
         * @memberof gitconfig.DefaultChartValuesResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        DefaultChartValuesResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/gitconfig.DefaultChartValuesResponse";
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

        /**
         * Gets the default type url for ToggleGlobalStatusResponse
         * @function getTypeUrl
         * @memberof gitconfig.ToggleGlobalStatusResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        ToggleGlobalStatusResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/gitconfig.ToggleGlobalStatusResponse";
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
         * @property {string|null} [display_name] Config display_name
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
         * Config display_name.
         * @member {string} display_name
         * @memberof mars.Config
         * @instance
         */
        Config.prototype.display_name = "";

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
            if (message.display_name != null && Object.hasOwnProperty.call(message, "display_name"))
                writer.uint32(/* id 10, wireType 2 =*/82).string(message.display_name);
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
                case 1: {
                        message.config_file = reader.string();
                        break;
                    }
                case 2: {
                        message.config_file_values = reader.string();
                        break;
                    }
                case 3: {
                        message.config_field = reader.string();
                        break;
                    }
                case 4: {
                        message.is_simple_env = reader.bool();
                        break;
                    }
                case 5: {
                        message.config_file_type = reader.string();
                        break;
                    }
                case 6: {
                        message.local_chart_path = reader.string();
                        break;
                    }
                case 7: {
                        if (!(message.branches && message.branches.length))
                            message.branches = [];
                        message.branches.push(reader.string());
                        break;
                    }
                case 8: {
                        message.values_yaml = reader.string();
                        break;
                    }
                case 9: {
                        if (!(message.elements && message.elements.length))
                            message.elements = [];
                        message.elements.push($root.mars.Element.decode(reader, reader.uint32()));
                        break;
                    }
                case 10: {
                        message.display_name = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for Config
         * @function getTypeUrl
         * @memberof mars.Config
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        Config.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/mars.Config";
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
     * @property {number} ElementTypeTextArea=6 ElementTypeTextArea value
     */
    mars.ElementType = (function() {
        const valuesById = {}, values = Object.create(valuesById);
        values[valuesById[0] = "ElementTypeUnknown"] = 0;
        values[valuesById[1] = "ElementTypeInput"] = 1;
        values[valuesById[2] = "ElementTypeInputNumber"] = 2;
        values[valuesById[3] = "ElementTypeSelect"] = 3;
        values[valuesById[4] = "ElementTypeRadio"] = 4;
        values[valuesById[5] = "ElementTypeSwitch"] = 5;
        values[valuesById[6] = "ElementTypeTextArea"] = 6;
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
                case 1: {
                        message.path = reader.string();
                        break;
                    }
                case 2: {
                        message.type = reader.int32();
                        break;
                    }
                case 3: {
                        message["default"] = reader.string();
                        break;
                    }
                case 4: {
                        message.description = reader.string();
                        break;
                    }
                case 6: {
                        if (!(message.select_values && message.select_values.length))
                            message.select_values = [];
                        message.select_values.push(reader.string());
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for Element
         * @function getTypeUrl
         * @memberof mars.Element
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        Element.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/mars.Element";
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
                case 1: {
                        message.namespace = reader.string();
                        break;
                    }
                case 2: {
                        message.pod = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for TopPodRequest
         * @function getTypeUrl
         * @memberof metrics.TopPodRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        TopPodRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/metrics.TopPodRequest";
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
                case 1: {
                        message.cpu = reader.double();
                        break;
                    }
                case 2: {
                        message.memory = reader.double();
                        break;
                    }
                case 3: {
                        message.humanize_cpu = reader.string();
                        break;
                    }
                case 4: {
                        message.humanize_memory = reader.string();
                        break;
                    }
                case 5: {
                        message.time = reader.string();
                        break;
                    }
                case 6: {
                        message.length = reader.int64();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for TopPodResponse
         * @function getTypeUrl
         * @memberof metrics.TopPodResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        TopPodResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/metrics.TopPodResponse";
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
                case 1: {
                        message.namespace_id = reader.int64();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for CpuMemoryInNamespaceRequest
         * @function getTypeUrl
         * @memberof metrics.CpuMemoryInNamespaceRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        CpuMemoryInNamespaceRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/metrics.CpuMemoryInNamespaceRequest";
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
                case 1: {
                        message.cpu = reader.string();
                        break;
                    }
                case 2: {
                        message.memory = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for CpuMemoryInNamespaceResponse
         * @function getTypeUrl
         * @memberof metrics.CpuMemoryInNamespaceResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        CpuMemoryInNamespaceResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/metrics.CpuMemoryInNamespaceResponse";
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
                case 1: {
                        message.project_id = reader.int64();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for CpuMemoryInProjectRequest
         * @function getTypeUrl
         * @memberof metrics.CpuMemoryInProjectRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        CpuMemoryInProjectRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/metrics.CpuMemoryInProjectRequest";
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
                case 1: {
                        message.cpu = reader.string();
                        break;
                    }
                case 2: {
                        message.memory = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for CpuMemoryInProjectResponse
         * @function getTypeUrl
         * @memberof metrics.CpuMemoryInProjectResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        CpuMemoryInProjectResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/metrics.CpuMemoryInProjectResponse";
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
                case 1: {
                        message.namespace = reader.string();
                        break;
                    }
                case 2: {
                        message.ignore_if_exists = reader.bool();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for CreateRequest
         * @function getTypeUrl
         * @memberof namespace.CreateRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        CreateRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/namespace.CreateRequest";
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
                case 1: {
                        message.namespace_id = reader.int64();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for ShowRequest
         * @function getTypeUrl
         * @memberof namespace.ShowRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        ShowRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/namespace.ShowRequest";
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
                case 1: {
                        message.namespace_id = reader.int64();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for DeleteRequest
         * @function getTypeUrl
         * @memberof namespace.DeleteRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        DeleteRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/namespace.DeleteRequest";
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
                case 1: {
                        message.name = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for IsExistsRequest
         * @function getTypeUrl
         * @memberof namespace.IsExistsRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        IsExistsRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/namespace.IsExistsRequest";
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
                case 1: {
                        if (!(message.items && message.items.length))
                            message.items = [];
                        message.items.push($root.types.NamespaceModel.decode(reader, reader.uint32()));
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for AllResponse
         * @function getTypeUrl
         * @memberof namespace.AllResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        AllResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/namespace.AllResponse";
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
                case 1: {
                        message.namespace = $root.types.NamespaceModel.decode(reader, reader.uint32());
                        break;
                    }
                case 2: {
                        message.exists = reader.bool();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for CreateResponse
         * @function getTypeUrl
         * @memberof namespace.CreateResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        CreateResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/namespace.CreateResponse";
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
                case 1: {
                        message.namespace = $root.types.NamespaceModel.decode(reader, reader.uint32());
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for ShowResponse
         * @function getTypeUrl
         * @memberof namespace.ShowResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        ShowResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/namespace.ShowResponse";
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
                case 1: {
                        message.exists = reader.bool();
                        break;
                    }
                case 2: {
                        message.id = reader.int64();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for IsExistsResponse
         * @function getTypeUrl
         * @memberof namespace.IsExistsResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        IsExistsResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/namespace.IsExistsResponse";
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

        /**
         * Gets the default type url for AllRequest
         * @function getTypeUrl
         * @memberof namespace.AllRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        AllRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/namespace.AllRequest";
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

        /**
         * Gets the default type url for DeleteResponse
         * @function getTypeUrl
         * @memberof namespace.DeleteResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        DeleteResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/namespace.DeleteResponse";
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
                case 1: {
                        message.random = reader.bool();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for BackgroundRequest
         * @function getTypeUrl
         * @memberof picture.BackgroundRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        BackgroundRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/picture.BackgroundRequest";
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
                case 1: {
                        message.url = reader.string();
                        break;
                    }
                case 2: {
                        message.copyright = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for BackgroundResponse
         * @function getTypeUrl
         * @memberof picture.BackgroundResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        BackgroundResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/picture.BackgroundResponse";
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
                case 1: {
                        message.project_id = reader.int64();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for DeleteRequest
         * @function getTypeUrl
         * @memberof project.DeleteRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        DeleteRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/project.DeleteRequest";
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
                case 1: {
                        message.project_id = reader.int64();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for ShowRequest
         * @function getTypeUrl
         * @memberof project.ShowRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        ShowRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/project.ShowRequest";
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
                case 1: {
                        message.project = $root.types.ProjectModel.decode(reader, reader.uint32());
                        break;
                    }
                case 13: {
                        if (!(message.urls && message.urls.length))
                            message.urls = [];
                        message.urls.push($root.types.ServiceEndpoint.decode(reader, reader.uint32()));
                        break;
                    }
                case 15: {
                        message.cpu = reader.string();
                        break;
                    }
                case 16: {
                        message.memory = reader.string();
                        break;
                    }
                case 23: {
                        if (!(message.elements && message.elements.length))
                            message.elements = [];
                        message.elements.push($root.mars.Element.decode(reader, reader.uint32()));
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for ShowResponse
         * @function getTypeUrl
         * @memberof project.ShowResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        ShowResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/project.ShowResponse";
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
                case 1: {
                        message.project_id = reader.int64();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for AllContainersRequest
         * @function getTypeUrl
         * @memberof project.AllContainersRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        AllContainersRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/project.AllContainersRequest";
        };

        return AllContainersRequest;
    })();

    project.AllContainersResponse = (function() {

        /**
         * Properties of an AllContainersResponse.
         * @memberof project
         * @interface IAllContainersResponse
         * @property {Array.<types.StateContainer>|null} [items] AllContainersResponse items
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
         * @member {Array.<types.StateContainer>} items
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
                    $root.types.StateContainer.encode(message.items[i], writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
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
                case 1: {
                        if (!(message.items && message.items.length))
                            message.items = [];
                        message.items.push($root.types.StateContainer.decode(reader, reader.uint32()));
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for AllContainersResponse
         * @function getTypeUrl
         * @memberof project.AllContainersResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        AllContainersResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/project.AllContainersResponse";
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
                case 1: {
                        message.metadata = $root.websocket.Metadata.decode(reader, reader.uint32());
                        break;
                    }
                case 2: {
                        message.project = $root.types.ProjectModel.decode(reader, reader.uint32());
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for ApplyResponse
         * @function getTypeUrl
         * @memberof project.ApplyResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        ApplyResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/project.ApplyResponse";
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
                case 1: {
                        if (!(message.results && message.results.length))
                            message.results = [];
                        message.results.push(reader.string());
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for DryRunApplyResponse
         * @function getTypeUrl
         * @memberof project.DryRunApplyResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        DryRunApplyResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/project.DryRunApplyResponse";
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
                case 1: {
                        message.namespace_id = reader.int64();
                        break;
                    }
                case 2: {
                        message.name = reader.string();
                        break;
                    }
                case 3: {
                        message.git_project_id = reader.int64();
                        break;
                    }
                case 4: {
                        message.git_branch = reader.string();
                        break;
                    }
                case 5: {
                        message.git_commit = reader.string();
                        break;
                    }
                case 6: {
                        message.config = reader.string();
                        break;
                    }
                case 7: {
                        message.atomic = reader.bool();
                        break;
                    }
                case 8: {
                        message.websocket_sync = reader.bool();
                        break;
                    }
                case 11: {
                        message.send_percent = reader.bool();
                        break;
                    }
                case 9: {
                        if (!(message.extra_values && message.extra_values.length))
                            message.extra_values = [];
                        message.extra_values.push($root.types.ExtraValue.decode(reader, reader.uint32()));
                        break;
                    }
                case 10: {
                        message.install_timeout_seconds = reader.int64();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for ApplyRequest
         * @function getTypeUrl
         * @memberof project.ApplyRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        ApplyRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/project.ApplyRequest";
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

        /**
         * Gets the default type url for DeleteResponse
         * @function getTypeUrl
         * @memberof project.DeleteResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        DeleteResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/project.DeleteResponse";
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
                case 1: {
                        message.page = reader.int64();
                        break;
                    }
                case 2: {
                        message.page_size = reader.int64();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for ListRequest
         * @function getTypeUrl
         * @memberof project.ListRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        ListRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/project.ListRequest";
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
                case 1: {
                        message.page = reader.int64();
                        break;
                    }
                case 2: {
                        message.page_size = reader.int64();
                        break;
                    }
                case 3: {
                        message.count = reader.int64();
                        break;
                    }
                case 4: {
                        if (!(message.items && message.items.length))
                            message.items = [];
                        message.items.push($root.types.ProjectModel.decode(reader, reader.uint32()));
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for ListResponse
         * @function getTypeUrl
         * @memberof project.ListResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        ListResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/project.ListResponse";
        };

        return ListResponse;
    })();

    project.HostVariablesRequest = (function() {

        /**
         * Properties of a HostVariablesRequest.
         * @memberof project
         * @interface IHostVariablesRequest
         * @property {string|null} [project_name] HostVariablesRequest project_name
         * @property {string|null} [namespace] HostVariablesRequest namespace
         * @property {number|null} [git_project_id] HostVariablesRequest git_project_id
         * @property {string|null} [git_branch] HostVariablesRequest git_branch
         */

        /**
         * Constructs a new HostVariablesRequest.
         * @memberof project
         * @classdesc Represents a HostVariablesRequest.
         * @implements IHostVariablesRequest
         * @constructor
         * @param {project.IHostVariablesRequest=} [properties] Properties to set
         */
        function HostVariablesRequest(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * HostVariablesRequest project_name.
         * @member {string} project_name
         * @memberof project.HostVariablesRequest
         * @instance
         */
        HostVariablesRequest.prototype.project_name = "";

        /**
         * HostVariablesRequest namespace.
         * @member {string} namespace
         * @memberof project.HostVariablesRequest
         * @instance
         */
        HostVariablesRequest.prototype.namespace = "";

        /**
         * HostVariablesRequest git_project_id.
         * @member {number} git_project_id
         * @memberof project.HostVariablesRequest
         * @instance
         */
        HostVariablesRequest.prototype.git_project_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * HostVariablesRequest git_branch.
         * @member {string} git_branch
         * @memberof project.HostVariablesRequest
         * @instance
         */
        HostVariablesRequest.prototype.git_branch = "";

        /**
         * Encodes the specified HostVariablesRequest message. Does not implicitly {@link project.HostVariablesRequest.verify|verify} messages.
         * @function encode
         * @memberof project.HostVariablesRequest
         * @static
         * @param {project.HostVariablesRequest} message HostVariablesRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        HostVariablesRequest.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.project_name != null && Object.hasOwnProperty.call(message, "project_name"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.project_name);
            if (message.namespace != null && Object.hasOwnProperty.call(message, "namespace"))
                writer.uint32(/* id 2, wireType 2 =*/18).string(message.namespace);
            if (message.git_project_id != null && Object.hasOwnProperty.call(message, "git_project_id"))
                writer.uint32(/* id 3, wireType 0 =*/24).int64(message.git_project_id);
            if (message.git_branch != null && Object.hasOwnProperty.call(message, "git_branch"))
                writer.uint32(/* id 4, wireType 2 =*/34).string(message.git_branch);
            return writer;
        };

        /**
         * Decodes a HostVariablesRequest message from the specified reader or buffer.
         * @function decode
         * @memberof project.HostVariablesRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {project.HostVariablesRequest} HostVariablesRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        HostVariablesRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.project.HostVariablesRequest();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1: {
                        message.project_name = reader.string();
                        break;
                    }
                case 2: {
                        message.namespace = reader.string();
                        break;
                    }
                case 3: {
                        message.git_project_id = reader.int64();
                        break;
                    }
                case 4: {
                        message.git_branch = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for HostVariablesRequest
         * @function getTypeUrl
         * @memberof project.HostVariablesRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        HostVariablesRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/project.HostVariablesRequest";
        };

        return HostVariablesRequest;
    })();

    project.HostVariablesResponse = (function() {

        /**
         * Properties of a HostVariablesResponse.
         * @memberof project
         * @interface IHostVariablesResponse
         * @property {Object.<string,string>|null} [hosts] HostVariablesResponse hosts
         */

        /**
         * Constructs a new HostVariablesResponse.
         * @memberof project
         * @classdesc Represents a HostVariablesResponse.
         * @implements IHostVariablesResponse
         * @constructor
         * @param {project.IHostVariablesResponse=} [properties] Properties to set
         */
        function HostVariablesResponse(properties) {
            this.hosts = {};
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * HostVariablesResponse hosts.
         * @member {Object.<string,string>} hosts
         * @memberof project.HostVariablesResponse
         * @instance
         */
        HostVariablesResponse.prototype.hosts = $util.emptyObject;

        /**
         * Encodes the specified HostVariablesResponse message. Does not implicitly {@link project.HostVariablesResponse.verify|verify} messages.
         * @function encode
         * @memberof project.HostVariablesResponse
         * @static
         * @param {project.HostVariablesResponse} message HostVariablesResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        HostVariablesResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.hosts != null && Object.hasOwnProperty.call(message, "hosts"))
                for (let keys = Object.keys(message.hosts), i = 0; i < keys.length; ++i)
                    writer.uint32(/* id 1, wireType 2 =*/10).fork().uint32(/* id 1, wireType 2 =*/10).string(keys[i]).uint32(/* id 2, wireType 2 =*/18).string(message.hosts[keys[i]]).ldelim();
            return writer;
        };

        /**
         * Decodes a HostVariablesResponse message from the specified reader or buffer.
         * @function decode
         * @memberof project.HostVariablesResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {project.HostVariablesResponse} HostVariablesResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        HostVariablesResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.project.HostVariablesResponse(), key, value;
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1: {
                        if (message.hosts === $util.emptyObject)
                            message.hosts = {};
                        let end2 = reader.uint32() + reader.pos;
                        key = "";
                        value = "";
                        while (reader.pos < end2) {
                            let tag2 = reader.uint32();
                            switch (tag2 >>> 3) {
                            case 1:
                                key = reader.string();
                                break;
                            case 2:
                                value = reader.string();
                                break;
                            default:
                                reader.skipType(tag2 & 7);
                                break;
                            }
                        }
                        message.hosts[key] = value;
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for HostVariablesResponse
         * @function getTypeUrl
         * @memberof project.HostVariablesResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        HostVariablesResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/project.HostVariablesResponse";
        };

        return HostVariablesResponse;
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

        /**
         * Callback as used by {@link project.Project#hostVariables}.
         * @memberof project.Project
         * @typedef HostVariablesCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {project.HostVariablesResponse} [response] HostVariablesResponse
         */

        /**
         * Calls HostVariables.
         * @function hostVariables
         * @memberof project.Project
         * @instance
         * @param {project.HostVariablesRequest} request HostVariablesRequest message or plain object
         * @param {project.Project.HostVariablesCallback} callback Node-style callback called with the error, if any, and HostVariablesResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(Project.prototype.hostVariables = function hostVariables(request, callback) {
            return this.rpcCall(hostVariables, $root.project.HostVariablesRequest, $root.project.HostVariablesResponse, request, callback);
        }, "name", { value: "HostVariables" });

        /**
         * Calls HostVariables.
         * @function hostVariables
         * @memberof project.Project
         * @instance
         * @param {project.HostVariablesRequest} request HostVariablesRequest message or plain object
         * @returns {Promise<project.HostVariablesResponse>} Promise
         * @variation 2
         */

        return Project;
    })();

    return project;
})();

export const token = $root.token = (() => {

    /**
     * Namespace token.
     * @exports token
     * @namespace
     */
    const token = {};

    token.AllRequest = (function() {

        /**
         * Properties of an AllRequest.
         * @memberof token
         * @interface IAllRequest
         */

        /**
         * Constructs a new AllRequest.
         * @memberof token
         * @classdesc Represents an AllRequest.
         * @implements IAllRequest
         * @constructor
         * @param {token.IAllRequest=} [properties] Properties to set
         */
        function AllRequest(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * Encodes the specified AllRequest message. Does not implicitly {@link token.AllRequest.verify|verify} messages.
         * @function encode
         * @memberof token.AllRequest
         * @static
         * @param {token.AllRequest} message AllRequest message or plain object to encode
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
         * @memberof token.AllRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {token.AllRequest} AllRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        AllRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.token.AllRequest();
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

        /**
         * Gets the default type url for AllRequest
         * @function getTypeUrl
         * @memberof token.AllRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        AllRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/token.AllRequest";
        };

        return AllRequest;
    })();

    token.AllResponse = (function() {

        /**
         * Properties of an AllResponse.
         * @memberof token
         * @interface IAllResponse
         * @property {Array.<types.AccessTokenModel>|null} [items] AllResponse items
         */

        /**
         * Constructs a new AllResponse.
         * @memberof token
         * @classdesc Represents an AllResponse.
         * @implements IAllResponse
         * @constructor
         * @param {token.IAllResponse=} [properties] Properties to set
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
         * @member {Array.<types.AccessTokenModel>} items
         * @memberof token.AllResponse
         * @instance
         */
        AllResponse.prototype.items = $util.emptyArray;

        /**
         * Encodes the specified AllResponse message. Does not implicitly {@link token.AllResponse.verify|verify} messages.
         * @function encode
         * @memberof token.AllResponse
         * @static
         * @param {token.AllResponse} message AllResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        AllResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.items != null && message.items.length)
                for (let i = 0; i < message.items.length; ++i)
                    $root.types.AccessTokenModel.encode(message.items[i], writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
            return writer;
        };

        /**
         * Decodes an AllResponse message from the specified reader or buffer.
         * @function decode
         * @memberof token.AllResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {token.AllResponse} AllResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        AllResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.token.AllResponse();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1: {
                        if (!(message.items && message.items.length))
                            message.items = [];
                        message.items.push($root.types.AccessTokenModel.decode(reader, reader.uint32()));
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for AllResponse
         * @function getTypeUrl
         * @memberof token.AllResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        AllResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/token.AllResponse";
        };

        return AllResponse;
    })();

    token.GrantRequest = (function() {

        /**
         * Properties of a GrantRequest.
         * @memberof token
         * @interface IGrantRequest
         * @property {number|null} [expire_seconds] GrantRequest expire_seconds
         * @property {string|null} [usage] GrantRequest usage
         */

        /**
         * Constructs a new GrantRequest.
         * @memberof token
         * @classdesc Represents a GrantRequest.
         * @implements IGrantRequest
         * @constructor
         * @param {token.IGrantRequest=} [properties] Properties to set
         */
        function GrantRequest(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * GrantRequest expire_seconds.
         * @member {number} expire_seconds
         * @memberof token.GrantRequest
         * @instance
         */
        GrantRequest.prototype.expire_seconds = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * GrantRequest usage.
         * @member {string} usage
         * @memberof token.GrantRequest
         * @instance
         */
        GrantRequest.prototype.usage = "";

        /**
         * Encodes the specified GrantRequest message. Does not implicitly {@link token.GrantRequest.verify|verify} messages.
         * @function encode
         * @memberof token.GrantRequest
         * @static
         * @param {token.GrantRequest} message GrantRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        GrantRequest.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.expire_seconds != null && Object.hasOwnProperty.call(message, "expire_seconds"))
                writer.uint32(/* id 1, wireType 0 =*/8).int64(message.expire_seconds);
            if (message.usage != null && Object.hasOwnProperty.call(message, "usage"))
                writer.uint32(/* id 2, wireType 2 =*/18).string(message.usage);
            return writer;
        };

        /**
         * Decodes a GrantRequest message from the specified reader or buffer.
         * @function decode
         * @memberof token.GrantRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {token.GrantRequest} GrantRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        GrantRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.token.GrantRequest();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1: {
                        message.expire_seconds = reader.int64();
                        break;
                    }
                case 2: {
                        message.usage = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for GrantRequest
         * @function getTypeUrl
         * @memberof token.GrantRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        GrantRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/token.GrantRequest";
        };

        return GrantRequest;
    })();

    token.GrantResponse = (function() {

        /**
         * Properties of a GrantResponse.
         * @memberof token
         * @interface IGrantResponse
         * @property {types.AccessTokenModel|null} [token] GrantResponse token
         */

        /**
         * Constructs a new GrantResponse.
         * @memberof token
         * @classdesc Represents a GrantResponse.
         * @implements IGrantResponse
         * @constructor
         * @param {token.IGrantResponse=} [properties] Properties to set
         */
        function GrantResponse(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * GrantResponse token.
         * @member {types.AccessTokenModel|null|undefined} token
         * @memberof token.GrantResponse
         * @instance
         */
        GrantResponse.prototype.token = null;

        /**
         * Encodes the specified GrantResponse message. Does not implicitly {@link token.GrantResponse.verify|verify} messages.
         * @function encode
         * @memberof token.GrantResponse
         * @static
         * @param {token.GrantResponse} message GrantResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        GrantResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.token != null && Object.hasOwnProperty.call(message, "token"))
                $root.types.AccessTokenModel.encode(message.token, writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
            return writer;
        };

        /**
         * Decodes a GrantResponse message from the specified reader or buffer.
         * @function decode
         * @memberof token.GrantResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {token.GrantResponse} GrantResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        GrantResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.token.GrantResponse();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1: {
                        message.token = $root.types.AccessTokenModel.decode(reader, reader.uint32());
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for GrantResponse
         * @function getTypeUrl
         * @memberof token.GrantResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        GrantResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/token.GrantResponse";
        };

        return GrantResponse;
    })();

    token.LeaseRequest = (function() {

        /**
         * Properties of a LeaseRequest.
         * @memberof token
         * @interface ILeaseRequest
         * @property {string|null} [token] LeaseRequest token
         * @property {number|null} [expire_seconds] LeaseRequest expire_seconds
         */

        /**
         * Constructs a new LeaseRequest.
         * @memberof token
         * @classdesc Represents a LeaseRequest.
         * @implements ILeaseRequest
         * @constructor
         * @param {token.ILeaseRequest=} [properties] Properties to set
         */
        function LeaseRequest(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * LeaseRequest token.
         * @member {string} token
         * @memberof token.LeaseRequest
         * @instance
         */
        LeaseRequest.prototype.token = "";

        /**
         * LeaseRequest expire_seconds.
         * @member {number} expire_seconds
         * @memberof token.LeaseRequest
         * @instance
         */
        LeaseRequest.prototype.expire_seconds = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Encodes the specified LeaseRequest message. Does not implicitly {@link token.LeaseRequest.verify|verify} messages.
         * @function encode
         * @memberof token.LeaseRequest
         * @static
         * @param {token.LeaseRequest} message LeaseRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        LeaseRequest.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.token != null && Object.hasOwnProperty.call(message, "token"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.token);
            if (message.expire_seconds != null && Object.hasOwnProperty.call(message, "expire_seconds"))
                writer.uint32(/* id 2, wireType 0 =*/16).int64(message.expire_seconds);
            return writer;
        };

        /**
         * Decodes a LeaseRequest message from the specified reader or buffer.
         * @function decode
         * @memberof token.LeaseRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {token.LeaseRequest} LeaseRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        LeaseRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.token.LeaseRequest();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1: {
                        message.token = reader.string();
                        break;
                    }
                case 2: {
                        message.expire_seconds = reader.int64();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for LeaseRequest
         * @function getTypeUrl
         * @memberof token.LeaseRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        LeaseRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/token.LeaseRequest";
        };

        return LeaseRequest;
    })();

    token.LeaseResponse = (function() {

        /**
         * Properties of a LeaseResponse.
         * @memberof token
         * @interface ILeaseResponse
         * @property {types.AccessTokenModel|null} [token] LeaseResponse token
         */

        /**
         * Constructs a new LeaseResponse.
         * @memberof token
         * @classdesc Represents a LeaseResponse.
         * @implements ILeaseResponse
         * @constructor
         * @param {token.ILeaseResponse=} [properties] Properties to set
         */
        function LeaseResponse(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * LeaseResponse token.
         * @member {types.AccessTokenModel|null|undefined} token
         * @memberof token.LeaseResponse
         * @instance
         */
        LeaseResponse.prototype.token = null;

        /**
         * Encodes the specified LeaseResponse message. Does not implicitly {@link token.LeaseResponse.verify|verify} messages.
         * @function encode
         * @memberof token.LeaseResponse
         * @static
         * @param {token.LeaseResponse} message LeaseResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        LeaseResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.token != null && Object.hasOwnProperty.call(message, "token"))
                $root.types.AccessTokenModel.encode(message.token, writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
            return writer;
        };

        /**
         * Decodes a LeaseResponse message from the specified reader or buffer.
         * @function decode
         * @memberof token.LeaseResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {token.LeaseResponse} LeaseResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        LeaseResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.token.LeaseResponse();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1: {
                        message.token = $root.types.AccessTokenModel.decode(reader, reader.uint32());
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for LeaseResponse
         * @function getTypeUrl
         * @memberof token.LeaseResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        LeaseResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/token.LeaseResponse";
        };

        return LeaseResponse;
    })();

    token.RevokeRequest = (function() {

        /**
         * Properties of a RevokeRequest.
         * @memberof token
         * @interface IRevokeRequest
         * @property {string|null} [token] RevokeRequest token
         */

        /**
         * Constructs a new RevokeRequest.
         * @memberof token
         * @classdesc Represents a RevokeRequest.
         * @implements IRevokeRequest
         * @constructor
         * @param {token.IRevokeRequest=} [properties] Properties to set
         */
        function RevokeRequest(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * RevokeRequest token.
         * @member {string} token
         * @memberof token.RevokeRequest
         * @instance
         */
        RevokeRequest.prototype.token = "";

        /**
         * Encodes the specified RevokeRequest message. Does not implicitly {@link token.RevokeRequest.verify|verify} messages.
         * @function encode
         * @memberof token.RevokeRequest
         * @static
         * @param {token.RevokeRequest} message RevokeRequest message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        RevokeRequest.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.token != null && Object.hasOwnProperty.call(message, "token"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.token);
            return writer;
        };

        /**
         * Decodes a RevokeRequest message from the specified reader or buffer.
         * @function decode
         * @memberof token.RevokeRequest
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {token.RevokeRequest} RevokeRequest
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        RevokeRequest.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.token.RevokeRequest();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1: {
                        message.token = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for RevokeRequest
         * @function getTypeUrl
         * @memberof token.RevokeRequest
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        RevokeRequest.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/token.RevokeRequest";
        };

        return RevokeRequest;
    })();

    token.RevokeResponse = (function() {

        /**
         * Properties of a RevokeResponse.
         * @memberof token
         * @interface IRevokeResponse
         */

        /**
         * Constructs a new RevokeResponse.
         * @memberof token
         * @classdesc Represents a RevokeResponse.
         * @implements IRevokeResponse
         * @constructor
         * @param {token.IRevokeResponse=} [properties] Properties to set
         */
        function RevokeResponse(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * Encodes the specified RevokeResponse message. Does not implicitly {@link token.RevokeResponse.verify|verify} messages.
         * @function encode
         * @memberof token.RevokeResponse
         * @static
         * @param {token.RevokeResponse} message RevokeResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        RevokeResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            return writer;
        };

        /**
         * Decodes a RevokeResponse message from the specified reader or buffer.
         * @function decode
         * @memberof token.RevokeResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {token.RevokeResponse} RevokeResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        RevokeResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.token.RevokeResponse();
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

        /**
         * Gets the default type url for RevokeResponse
         * @function getTypeUrl
         * @memberof token.RevokeResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        RevokeResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/token.RevokeResponse";
        };

        return RevokeResponse;
    })();

    token.AccessToken = (function() {

        /**
         * Constructs a new AccessToken service.
         * @memberof token
         * @classdesc Represents an AccessToken
         * @extends $protobuf.rpc.Service
         * @constructor
         * @param {$protobuf.RPCImpl} rpcImpl RPC implementation
         * @param {boolean} [requestDelimited=false] Whether requests are length-delimited
         * @param {boolean} [responseDelimited=false] Whether responses are length-delimited
         */
        function AccessToken(rpcImpl, requestDelimited, responseDelimited) {
            $protobuf.rpc.Service.call(this, rpcImpl, requestDelimited, responseDelimited);
        }

        (AccessToken.prototype = Object.create($protobuf.rpc.Service.prototype)).constructor = AccessToken;

        /**
         * Callback as used by {@link token.AccessToken#all}.
         * @memberof token.AccessToken
         * @typedef AllCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {token.AllResponse} [response] AllResponse
         */

        /**
         * Calls All.
         * @function all
         * @memberof token.AccessToken
         * @instance
         * @param {token.AllRequest} request AllRequest message or plain object
         * @param {token.AccessToken.AllCallback} callback Node-style callback called with the error, if any, and AllResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(AccessToken.prototype.all = function all(request, callback) {
            return this.rpcCall(all, $root.token.AllRequest, $root.token.AllResponse, request, callback);
        }, "name", { value: "All" });

        /**
         * Calls All.
         * @function all
         * @memberof token.AccessToken
         * @instance
         * @param {token.AllRequest} request AllRequest message or plain object
         * @returns {Promise<token.AllResponse>} Promise
         * @variation 2
         */

        /**
         * Callback as used by {@link token.AccessToken#grant}.
         * @memberof token.AccessToken
         * @typedef GrantCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {token.GrantResponse} [response] GrantResponse
         */

        /**
         * Calls Grant.
         * @function grant
         * @memberof token.AccessToken
         * @instance
         * @param {token.GrantRequest} request GrantRequest message or plain object
         * @param {token.AccessToken.GrantCallback} callback Node-style callback called with the error, if any, and GrantResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(AccessToken.prototype.grant = function grant(request, callback) {
            return this.rpcCall(grant, $root.token.GrantRequest, $root.token.GrantResponse, request, callback);
        }, "name", { value: "Grant" });

        /**
         * Calls Grant.
         * @function grant
         * @memberof token.AccessToken
         * @instance
         * @param {token.GrantRequest} request GrantRequest message or plain object
         * @returns {Promise<token.GrantResponse>} Promise
         * @variation 2
         */

        /**
         * Callback as used by {@link token.AccessToken#lease}.
         * @memberof token.AccessToken
         * @typedef LeaseCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {token.LeaseResponse} [response] LeaseResponse
         */

        /**
         * Calls Lease.
         * @function lease
         * @memberof token.AccessToken
         * @instance
         * @param {token.LeaseRequest} request LeaseRequest message or plain object
         * @param {token.AccessToken.LeaseCallback} callback Node-style callback called with the error, if any, and LeaseResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(AccessToken.prototype.lease = function lease(request, callback) {
            return this.rpcCall(lease, $root.token.LeaseRequest, $root.token.LeaseResponse, request, callback);
        }, "name", { value: "Lease" });

        /**
         * Calls Lease.
         * @function lease
         * @memberof token.AccessToken
         * @instance
         * @param {token.LeaseRequest} request LeaseRequest message or plain object
         * @returns {Promise<token.LeaseResponse>} Promise
         * @variation 2
         */

        /**
         * Callback as used by {@link token.AccessToken#revoke}.
         * @memberof token.AccessToken
         * @typedef RevokeCallback
         * @type {function}
         * @param {Error|null} error Error, if any
         * @param {token.RevokeResponse} [response] RevokeResponse
         */

        /**
         * Calls Revoke.
         * @function revoke
         * @memberof token.AccessToken
         * @instance
         * @param {token.RevokeRequest} request RevokeRequest message or plain object
         * @param {token.AccessToken.RevokeCallback} callback Node-style callback called with the error, if any, and RevokeResponse
         * @returns {undefined}
         * @variation 1
         */
        Object.defineProperty(AccessToken.prototype.revoke = function revoke(request, callback) {
            return this.rpcCall(revoke, $root.token.RevokeRequest, $root.token.RevokeResponse, request, callback);
        }, "name", { value: "Revoke" });

        /**
         * Calls Revoke.
         * @function revoke
         * @memberof token.AccessToken
         * @instance
         * @param {token.RevokeRequest} request RevokeRequest message or plain object
         * @returns {Promise<token.RevokeResponse>} Promise
         * @variation 2
         */

        return AccessToken;
    })();

    return token;
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
     * @property {number} Login=8 Login value
     * @property {number} CancelDeploy=9 CancelDeploy value
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
        values[valuesById[8] = "Login"] = 8;
        values[valuesById[9] = "CancelDeploy"] = 9;
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
                case 1: {
                        message.namespace = reader.string();
                        break;
                    }
                case 2: {
                        message.pod = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for Pod
         * @function getTypeUrl
         * @memberof types.Pod
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        Pod.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/types.Pod";
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
                case 1: {
                        message.namespace = reader.string();
                        break;
                    }
                case 2: {
                        message.pod = reader.string();
                        break;
                    }
                case 3: {
                        message.container = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for Container
         * @function getTypeUrl
         * @memberof types.Container
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        Container.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/types.Container";
        };

        return Container;
    })();

    types.StateContainer = (function() {

        /**
         * Properties of a StateContainer.
         * @memberof types
         * @interface IStateContainer
         * @property {string|null} [namespace] StateContainer namespace
         * @property {string|null} [pod] StateContainer pod
         * @property {string|null} [container] StateContainer container
         * @property {boolean|null} [is_old] StateContainer is_old
         * @property {boolean|null} [terminating] StateContainer terminating
         * @property {boolean|null} [pending] StateContainer pending
         */

        /**
         * Constructs a new StateContainer.
         * @memberof types
         * @classdesc Represents a StateContainer.
         * @implements IStateContainer
         * @constructor
         * @param {types.IStateContainer=} [properties] Properties to set
         */
        function StateContainer(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * StateContainer namespace.
         * @member {string} namespace
         * @memberof types.StateContainer
         * @instance
         */
        StateContainer.prototype.namespace = "";

        /**
         * StateContainer pod.
         * @member {string} pod
         * @memberof types.StateContainer
         * @instance
         */
        StateContainer.prototype.pod = "";

        /**
         * StateContainer container.
         * @member {string} container
         * @memberof types.StateContainer
         * @instance
         */
        StateContainer.prototype.container = "";

        /**
         * StateContainer is_old.
         * @member {boolean} is_old
         * @memberof types.StateContainer
         * @instance
         */
        StateContainer.prototype.is_old = false;

        /**
         * StateContainer terminating.
         * @member {boolean} terminating
         * @memberof types.StateContainer
         * @instance
         */
        StateContainer.prototype.terminating = false;

        /**
         * StateContainer pending.
         * @member {boolean} pending
         * @memberof types.StateContainer
         * @instance
         */
        StateContainer.prototype.pending = false;

        /**
         * Encodes the specified StateContainer message. Does not implicitly {@link types.StateContainer.verify|verify} messages.
         * @function encode
         * @memberof types.StateContainer
         * @static
         * @param {types.StateContainer} message StateContainer message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        StateContainer.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.namespace != null && Object.hasOwnProperty.call(message, "namespace"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.namespace);
            if (message.pod != null && Object.hasOwnProperty.call(message, "pod"))
                writer.uint32(/* id 2, wireType 2 =*/18).string(message.pod);
            if (message.container != null && Object.hasOwnProperty.call(message, "container"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.container);
            if (message.is_old != null && Object.hasOwnProperty.call(message, "is_old"))
                writer.uint32(/* id 4, wireType 0 =*/32).bool(message.is_old);
            if (message.terminating != null && Object.hasOwnProperty.call(message, "terminating"))
                writer.uint32(/* id 5, wireType 0 =*/40).bool(message.terminating);
            if (message.pending != null && Object.hasOwnProperty.call(message, "pending"))
                writer.uint32(/* id 6, wireType 0 =*/48).bool(message.pending);
            return writer;
        };

        /**
         * Decodes a StateContainer message from the specified reader or buffer.
         * @function decode
         * @memberof types.StateContainer
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {types.StateContainer} StateContainer
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        StateContainer.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.types.StateContainer();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1: {
                        message.namespace = reader.string();
                        break;
                    }
                case 2: {
                        message.pod = reader.string();
                        break;
                    }
                case 3: {
                        message.container = reader.string();
                        break;
                    }
                case 4: {
                        message.is_old = reader.bool();
                        break;
                    }
                case 5: {
                        message.terminating = reader.bool();
                        break;
                    }
                case 6: {
                        message.pending = reader.bool();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for StateContainer
         * @function getTypeUrl
         * @memberof types.StateContainer
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        StateContainer.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/types.StateContainer";
        };

        return StateContainer;
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
                case 1: {
                        message.path = reader.string();
                        break;
                    }
                case 2: {
                        message.value = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for ExtraValue
         * @function getTypeUrl
         * @memberof types.ExtraValue
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        ExtraValue.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/types.ExtraValue";
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
                case 1: {
                        message.name = reader.string();
                        break;
                    }
                case 2: {
                        message.url = reader.string();
                        break;
                    }
                case 3: {
                        message.port_name = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for ServiceEndpoint
         * @function getTypeUrl
         * @memberof types.ServiceEndpoint
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        ServiceEndpoint.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/types.ServiceEndpoint";
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
         * @property {string|null} [config_type] ChangelogModel config_type
         * @property {string|null} [git_branch] ChangelogModel git_branch
         * @property {string|null} [git_commit] ChangelogModel git_commit
         * @property {string|null} [docker_image] ChangelogModel docker_image
         * @property {string|null} [env_values] ChangelogModel env_values
         * @property {string|null} [extra_values] ChangelogModel extra_values
         * @property {string|null} [final_extra_values] ChangelogModel final_extra_values
         * @property {string|null} [git_commit_web_url] ChangelogModel git_commit_web_url
         * @property {string|null} [git_commit_title] ChangelogModel git_commit_title
         * @property {string|null} [git_commit_author] ChangelogModel git_commit_author
         * @property {string|null} [git_commit_date] ChangelogModel git_commit_date
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
         * ChangelogModel config_type.
         * @member {string} config_type
         * @memberof types.ChangelogModel
         * @instance
         */
        ChangelogModel.prototype.config_type = "";

        /**
         * ChangelogModel git_branch.
         * @member {string} git_branch
         * @memberof types.ChangelogModel
         * @instance
         */
        ChangelogModel.prototype.git_branch = "";

        /**
         * ChangelogModel git_commit.
         * @member {string} git_commit
         * @memberof types.ChangelogModel
         * @instance
         */
        ChangelogModel.prototype.git_commit = "";

        /**
         * ChangelogModel docker_image.
         * @member {string} docker_image
         * @memberof types.ChangelogModel
         * @instance
         */
        ChangelogModel.prototype.docker_image = "";

        /**
         * ChangelogModel env_values.
         * @member {string} env_values
         * @memberof types.ChangelogModel
         * @instance
         */
        ChangelogModel.prototype.env_values = "";

        /**
         * ChangelogModel extra_values.
         * @member {string} extra_values
         * @memberof types.ChangelogModel
         * @instance
         */
        ChangelogModel.prototype.extra_values = "";

        /**
         * ChangelogModel final_extra_values.
         * @member {string} final_extra_values
         * @memberof types.ChangelogModel
         * @instance
         */
        ChangelogModel.prototype.final_extra_values = "";

        /**
         * ChangelogModel git_commit_web_url.
         * @member {string} git_commit_web_url
         * @memberof types.ChangelogModel
         * @instance
         */
        ChangelogModel.prototype.git_commit_web_url = "";

        /**
         * ChangelogModel git_commit_title.
         * @member {string} git_commit_title
         * @memberof types.ChangelogModel
         * @instance
         */
        ChangelogModel.prototype.git_commit_title = "";

        /**
         * ChangelogModel git_commit_author.
         * @member {string} git_commit_author
         * @memberof types.ChangelogModel
         * @instance
         */
        ChangelogModel.prototype.git_commit_author = "";

        /**
         * ChangelogModel git_commit_date.
         * @member {string} git_commit_date
         * @memberof types.ChangelogModel
         * @instance
         */
        ChangelogModel.prototype.git_commit_date = "";

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
            if (message.config_type != null && Object.hasOwnProperty.call(message, "config_type"))
                writer.uint32(/* id 12, wireType 2 =*/98).string(message.config_type);
            if (message.git_branch != null && Object.hasOwnProperty.call(message, "git_branch"))
                writer.uint32(/* id 13, wireType 2 =*/106).string(message.git_branch);
            if (message.git_commit != null && Object.hasOwnProperty.call(message, "git_commit"))
                writer.uint32(/* id 14, wireType 2 =*/114).string(message.git_commit);
            if (message.docker_image != null && Object.hasOwnProperty.call(message, "docker_image"))
                writer.uint32(/* id 15, wireType 2 =*/122).string(message.docker_image);
            if (message.env_values != null && Object.hasOwnProperty.call(message, "env_values"))
                writer.uint32(/* id 16, wireType 2 =*/130).string(message.env_values);
            if (message.extra_values != null && Object.hasOwnProperty.call(message, "extra_values"))
                writer.uint32(/* id 17, wireType 2 =*/138).string(message.extra_values);
            if (message.final_extra_values != null && Object.hasOwnProperty.call(message, "final_extra_values"))
                writer.uint32(/* id 18, wireType 2 =*/146).string(message.final_extra_values);
            if (message.git_commit_web_url != null && Object.hasOwnProperty.call(message, "git_commit_web_url"))
                writer.uint32(/* id 19, wireType 2 =*/154).string(message.git_commit_web_url);
            if (message.git_commit_title != null && Object.hasOwnProperty.call(message, "git_commit_title"))
                writer.uint32(/* id 20, wireType 2 =*/162).string(message.git_commit_title);
            if (message.git_commit_author != null && Object.hasOwnProperty.call(message, "git_commit_author"))
                writer.uint32(/* id 21, wireType 2 =*/170).string(message.git_commit_author);
            if (message.git_commit_date != null && Object.hasOwnProperty.call(message, "git_commit_date"))
                writer.uint32(/* id 22, wireType 2 =*/178).string(message.git_commit_date);
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
                case 1: {
                        message.id = reader.int64();
                        break;
                    }
                case 2: {
                        message.version = reader.int64();
                        break;
                    }
                case 3: {
                        message.username = reader.string();
                        break;
                    }
                case 4: {
                        message.manifest = reader.string();
                        break;
                    }
                case 5: {
                        message.config = reader.string();
                        break;
                    }
                case 6: {
                        message.config_changed = reader.bool();
                        break;
                    }
                case 7: {
                        message.project_id = reader.int64();
                        break;
                    }
                case 8: {
                        message.git_project_id = reader.int64();
                        break;
                    }
                case 9: {
                        message.project = $root.types.ProjectModel.decode(reader, reader.uint32());
                        break;
                    }
                case 10: {
                        message.git_project = $root.types.GitProjectModel.decode(reader, reader.uint32());
                        break;
                    }
                case 11: {
                        message.date = reader.string();
                        break;
                    }
                case 12: {
                        message.config_type = reader.string();
                        break;
                    }
                case 13: {
                        message.git_branch = reader.string();
                        break;
                    }
                case 14: {
                        message.git_commit = reader.string();
                        break;
                    }
                case 15: {
                        message.docker_image = reader.string();
                        break;
                    }
                case 16: {
                        message.env_values = reader.string();
                        break;
                    }
                case 17: {
                        message.extra_values = reader.string();
                        break;
                    }
                case 18: {
                        message.final_extra_values = reader.string();
                        break;
                    }
                case 19: {
                        message.git_commit_web_url = reader.string();
                        break;
                    }
                case 20: {
                        message.git_commit_title = reader.string();
                        break;
                    }
                case 21: {
                        message.git_commit_author = reader.string();
                        break;
                    }
                case 22: {
                        message.git_commit_date = reader.string();
                        break;
                    }
                case 100: {
                        message.created_at = reader.string();
                        break;
                    }
                case 101: {
                        message.updated_at = reader.string();
                        break;
                    }
                case 102: {
                        message.deleted_at = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for ChangelogModel
         * @function getTypeUrl
         * @memberof types.ChangelogModel
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        ChangelogModel.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/types.ChangelogModel";
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
         * @property {boolean|null} [has_diff] EventModel has_diff
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
         * EventModel has_diff.
         * @member {boolean} has_diff
         * @memberof types.EventModel
         * @instance
         */
        EventModel.prototype.has_diff = false;

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
            if (message.has_diff != null && Object.hasOwnProperty.call(message, "has_diff"))
                writer.uint32(/* id 11, wireType 0 =*/88).bool(message.has_diff);
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
                case 1: {
                        message.id = reader.int64();
                        break;
                    }
                case 2: {
                        message.action = reader.int32();
                        break;
                    }
                case 3: {
                        message.username = reader.string();
                        break;
                    }
                case 4: {
                        message.message = reader.string();
                        break;
                    }
                case 5: {
                        message.old = reader.string();
                        break;
                    }
                case 6: {
                        message["new"] = reader.string();
                        break;
                    }
                case 7: {
                        message.duration = reader.string();
                        break;
                    }
                case 8: {
                        message.file_id = reader.int64();
                        break;
                    }
                case 9: {
                        message.file = $root.types.FileModel.decode(reader, reader.uint32());
                        break;
                    }
                case 10: {
                        message.event_at = reader.string();
                        break;
                    }
                case 11: {
                        message.has_diff = reader.bool();
                        break;
                    }
                case 100: {
                        message.created_at = reader.string();
                        break;
                    }
                case 101: {
                        message.updated_at = reader.string();
                        break;
                    }
                case 102: {
                        message.deleted_at = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for EventModel
         * @function getTypeUrl
         * @memberof types.EventModel
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        EventModel.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/types.EventModel";
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
                case 1: {
                        message.id = reader.int64();
                        break;
                    }
                case 2: {
                        message.path = reader.string();
                        break;
                    }
                case 3: {
                        message.size = reader.int64();
                        break;
                    }
                case 4: {
                        message.username = reader.string();
                        break;
                    }
                case 5: {
                        message.namespace = reader.string();
                        break;
                    }
                case 6: {
                        message.pod = reader.string();
                        break;
                    }
                case 7: {
                        message.container = reader.string();
                        break;
                    }
                case 8: {
                        message.container_Path = reader.string();
                        break;
                    }
                case 9: {
                        message.humanize_size = reader.string();
                        break;
                    }
                case 100: {
                        message.created_at = reader.string();
                        break;
                    }
                case 101: {
                        message.updated_at = reader.string();
                        break;
                    }
                case 102: {
                        message.deleted_at = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for FileModel
         * @function getTypeUrl
         * @memberof types.FileModel
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        FileModel.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/types.FileModel";
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
                case 1: {
                        message.id = reader.int64();
                        break;
                    }
                case 2: {
                        message.default_branch = reader.string();
                        break;
                    }
                case 3: {
                        message.name = reader.string();
                        break;
                    }
                case 4: {
                        message.git_project_id = reader.int64();
                        break;
                    }
                case 5: {
                        message.enabled = reader.bool();
                        break;
                    }
                case 6: {
                        message.global_enabled = reader.bool();
                        break;
                    }
                case 7: {
                        message.global_config = reader.string();
                        break;
                    }
                case 100: {
                        message.created_at = reader.string();
                        break;
                    }
                case 101: {
                        message.updated_at = reader.string();
                        break;
                    }
                case 102: {
                        message.deleted_at = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for GitProjectModel
         * @function getTypeUrl
         * @memberof types.GitProjectModel
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        GitProjectModel.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/types.GitProjectModel";
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
                case 1: {
                        message.name = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for ImagePullSecret
         * @function getTypeUrl
         * @memberof types.ImagePullSecret
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        ImagePullSecret.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/types.ImagePullSecret";
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
                case 1: {
                        message.id = reader.int64();
                        break;
                    }
                case 2: {
                        message.name = reader.string();
                        break;
                    }
                case 3: {
                        if (!(message.ImagePullSecrets && message.ImagePullSecrets.length))
                            message.ImagePullSecrets = [];
                        message.ImagePullSecrets.push($root.types.ImagePullSecret.decode(reader, reader.uint32()));
                        break;
                    }
                case 4: {
                        if (!(message.projects && message.projects.length))
                            message.projects = [];
                        message.projects.push($root.types.ProjectModel.decode(reader, reader.uint32()));
                        break;
                    }
                case 100: {
                        message.created_at = reader.string();
                        break;
                    }
                case 101: {
                        message.updated_at = reader.string();
                        break;
                    }
                case 102: {
                        message.deleted_at = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for NamespaceModel
         * @function getTypeUrl
         * @memberof types.NamespaceModel
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        NamespaceModel.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/types.NamespaceModel";
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
                case 1: {
                        message.id = reader.int64();
                        break;
                    }
                case 2: {
                        message.name = reader.string();
                        break;
                    }
                case 3: {
                        message.git_project_id = reader.int64();
                        break;
                    }
                case 4: {
                        message.git_branch = reader.string();
                        break;
                    }
                case 5: {
                        message.git_commit = reader.string();
                        break;
                    }
                case 6: {
                        message.config = reader.string();
                        break;
                    }
                case 7: {
                        message.override_values = reader.string();
                        break;
                    }
                case 8: {
                        message.docker_image = reader.string();
                        break;
                    }
                case 9: {
                        message.pod_selectors = reader.string();
                        break;
                    }
                case 10: {
                        message.namespace_id = reader.int64();
                        break;
                    }
                case 11: {
                        message.atomic = reader.bool();
                        break;
                    }
                case 12: {
                        message.env_values = reader.string();
                        break;
                    }
                case 13: {
                        if (!(message.extra_values && message.extra_values.length))
                            message.extra_values = [];
                        message.extra_values.push($root.types.ExtraValue.decode(reader, reader.uint32()));
                        break;
                    }
                case 14: {
                        message.final_extra_values = reader.string();
                        break;
                    }
                case 15: {
                        message.deploy_status = reader.int32();
                        break;
                    }
                case 16: {
                        message.humanize_created_at = reader.string();
                        break;
                    }
                case 17: {
                        message.humanize_updated_at = reader.string();
                        break;
                    }
                case 18: {
                        message.config_type = reader.string();
                        break;
                    }
                case 19: {
                        message.git_commit_web_url = reader.string();
                        break;
                    }
                case 20: {
                        message.git_commit_title = reader.string();
                        break;
                    }
                case 21: {
                        message.git_commit_author = reader.string();
                        break;
                    }
                case 22: {
                        message.git_commit_date = reader.string();
                        break;
                    }
                case 50: {
                        message.namespace = $root.types.NamespaceModel.decode(reader, reader.uint32());
                        break;
                    }
                case 100: {
                        message.created_at = reader.string();
                        break;
                    }
                case 101: {
                        message.updated_at = reader.string();
                        break;
                    }
                case 102: {
                        message.deleted_at = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for ProjectModel
         * @function getTypeUrl
         * @memberof types.ProjectModel
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        ProjectModel.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/types.ProjectModel";
        };

        return ProjectModel;
    })();

    types.AccessTokenModel = (function() {

        /**
         * Properties of an AccessTokenModel.
         * @memberof types
         * @interface IAccessTokenModel
         * @property {string|null} [token] AccessTokenModel token
         * @property {string|null} [email] AccessTokenModel email
         * @property {string|null} [expired_at] AccessTokenModel expired_at
         * @property {string|null} [usage] AccessTokenModel usage
         * @property {string|null} [last_used_at] AccessTokenModel last_used_at
         * @property {string|null} [created_at] AccessTokenModel created_at
         * @property {string|null} [updated_at] AccessTokenModel updated_at
         * @property {string|null} [deleted_at] AccessTokenModel deleted_at
         */

        /**
         * Constructs a new AccessTokenModel.
         * @memberof types
         * @classdesc Represents an AccessTokenModel.
         * @implements IAccessTokenModel
         * @constructor
         * @param {types.IAccessTokenModel=} [properties] Properties to set
         */
        function AccessTokenModel(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * AccessTokenModel token.
         * @member {string} token
         * @memberof types.AccessTokenModel
         * @instance
         */
        AccessTokenModel.prototype.token = "";

        /**
         * AccessTokenModel email.
         * @member {string} email
         * @memberof types.AccessTokenModel
         * @instance
         */
        AccessTokenModel.prototype.email = "";

        /**
         * AccessTokenModel expired_at.
         * @member {string} expired_at
         * @memberof types.AccessTokenModel
         * @instance
         */
        AccessTokenModel.prototype.expired_at = "";

        /**
         * AccessTokenModel usage.
         * @member {string} usage
         * @memberof types.AccessTokenModel
         * @instance
         */
        AccessTokenModel.prototype.usage = "";

        /**
         * AccessTokenModel last_used_at.
         * @member {string} last_used_at
         * @memberof types.AccessTokenModel
         * @instance
         */
        AccessTokenModel.prototype.last_used_at = "";

        /**
         * AccessTokenModel created_at.
         * @member {string} created_at
         * @memberof types.AccessTokenModel
         * @instance
         */
        AccessTokenModel.prototype.created_at = "";

        /**
         * AccessTokenModel updated_at.
         * @member {string} updated_at
         * @memberof types.AccessTokenModel
         * @instance
         */
        AccessTokenModel.prototype.updated_at = "";

        /**
         * AccessTokenModel deleted_at.
         * @member {string} deleted_at
         * @memberof types.AccessTokenModel
         * @instance
         */
        AccessTokenModel.prototype.deleted_at = "";

        /**
         * Encodes the specified AccessTokenModel message. Does not implicitly {@link types.AccessTokenModel.verify|verify} messages.
         * @function encode
         * @memberof types.AccessTokenModel
         * @static
         * @param {types.AccessTokenModel} message AccessTokenModel message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        AccessTokenModel.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.token != null && Object.hasOwnProperty.call(message, "token"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.token);
            if (message.email != null && Object.hasOwnProperty.call(message, "email"))
                writer.uint32(/* id 2, wireType 2 =*/18).string(message.email);
            if (message.expired_at != null && Object.hasOwnProperty.call(message, "expired_at"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.expired_at);
            if (message.usage != null && Object.hasOwnProperty.call(message, "usage"))
                writer.uint32(/* id 4, wireType 2 =*/34).string(message.usage);
            if (message.last_used_at != null && Object.hasOwnProperty.call(message, "last_used_at"))
                writer.uint32(/* id 5, wireType 2 =*/42).string(message.last_used_at);
            if (message.created_at != null && Object.hasOwnProperty.call(message, "created_at"))
                writer.uint32(/* id 100, wireType 2 =*/802).string(message.created_at);
            if (message.updated_at != null && Object.hasOwnProperty.call(message, "updated_at"))
                writer.uint32(/* id 101, wireType 2 =*/810).string(message.updated_at);
            if (message.deleted_at != null && Object.hasOwnProperty.call(message, "deleted_at"))
                writer.uint32(/* id 102, wireType 2 =*/818).string(message.deleted_at);
            return writer;
        };

        /**
         * Decodes an AccessTokenModel message from the specified reader or buffer.
         * @function decode
         * @memberof types.AccessTokenModel
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {types.AccessTokenModel} AccessTokenModel
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        AccessTokenModel.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.types.AccessTokenModel();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1: {
                        message.token = reader.string();
                        break;
                    }
                case 2: {
                        message.email = reader.string();
                        break;
                    }
                case 3: {
                        message.expired_at = reader.string();
                        break;
                    }
                case 4: {
                        message.usage = reader.string();
                        break;
                    }
                case 5: {
                        message.last_used_at = reader.string();
                        break;
                    }
                case 100: {
                        message.created_at = reader.string();
                        break;
                    }
                case 101: {
                        message.updated_at = reader.string();
                        break;
                    }
                case 102: {
                        message.deleted_at = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for AccessTokenModel
         * @function getTypeUrl
         * @memberof types.AccessTokenModel
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        AccessTokenModel.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/types.AccessTokenModel";
        };

        return AccessTokenModel;
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

        /**
         * Gets the default type url for Request
         * @function getTypeUrl
         * @memberof version.Request
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        Request.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/version.Request";
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
                case 1: {
                        message.version = reader.string();
                        break;
                    }
                case 2: {
                        message.build_date = reader.string();
                        break;
                    }
                case 3: {
                        message.git_branch = reader.string();
                        break;
                    }
                case 4: {
                        message.git_commit = reader.string();
                        break;
                    }
                case 5: {
                        message.git_tag = reader.string();
                        break;
                    }
                case 6: {
                        message.go_version = reader.string();
                        break;
                    }
                case 7: {
                        message.compiler = reader.string();
                        break;
                    }
                case 8: {
                        message.platform = reader.string();
                        break;
                    }
                case 9: {
                        message.kubectl_version = reader.string();
                        break;
                    }
                case 10: {
                        message.helm_version = reader.string();
                        break;
                    }
                case 11: {
                        message.git_repo = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for Response
         * @function getTypeUrl
         * @memberof version.Response
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        Response.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/version.Response";
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
     * @property {number} ProjectPodEvent=10 ProjectPodEvent value
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
        values[valuesById[10] = "ProjectPodEvent"] = 10;
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
     * @property {number} LogWithContainers=6 LogWithContainers value
     */
    websocket.ResultType = (function() {
        const valuesById = {}, values = Object.create(valuesById);
        values[valuesById[0] = "ResultUnknown"] = 0;
        values[valuesById[1] = "Error"] = 1;
        values[valuesById[2] = "Success"] = 2;
        values[valuesById[3] = "Deployed"] = 3;
        values[valuesById[4] = "DeployedFailed"] = 4;
        values[valuesById[5] = "DeployedCanceled"] = 5;
        values[valuesById[6] = "LogWithContainers"] = 6;
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
                case 1: {
                        message.type = reader.int32();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for WsRequestMetadata
         * @function getTypeUrl
         * @memberof websocket.WsRequestMetadata
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        WsRequestMetadata.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/websocket.WsRequestMetadata";
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
                case 1: {
                        message.type = reader.int32();
                        break;
                    }
                case 2: {
                        message.token = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for AuthorizeTokenInput
         * @function getTypeUrl
         * @memberof websocket.AuthorizeTokenInput
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        AuthorizeTokenInput.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/websocket.AuthorizeTokenInput";
        };

        return AuthorizeTokenInput;
    })();

    websocket.TerminalMessage = (function() {

        /**
         * Properties of a TerminalMessage.
         * @memberof websocket
         * @interface ITerminalMessage
         * @property {string|null} [op] TerminalMessage op
         * @property {Uint8Array|null} [data] TerminalMessage data
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
         * @member {Uint8Array} data
         * @memberof websocket.TerminalMessage
         * @instance
         */
        TerminalMessage.prototype.data = $util.newBuffer([]);

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
                writer.uint32(/* id 2, wireType 2 =*/18).bytes(message.data);
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
                case 1: {
                        message.op = reader.string();
                        break;
                    }
                case 2: {
                        message.data = reader.bytes();
                        break;
                    }
                case 3: {
                        message.session_id = reader.string();
                        break;
                    }
                case 4: {
                        message.rows = reader.uint32();
                        break;
                    }
                case 5: {
                        message.cols = reader.uint32();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for TerminalMessage
         * @function getTypeUrl
         * @memberof websocket.TerminalMessage
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        TerminalMessage.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/websocket.TerminalMessage";
        };

        return TerminalMessage;
    })();

    websocket.ProjectPodEventJoinInput = (function() {

        /**
         * Properties of a ProjectPodEventJoinInput.
         * @memberof websocket
         * @interface IProjectPodEventJoinInput
         * @property {websocket.Type|null} [type] ProjectPodEventJoinInput type
         * @property {boolean|null} [join] ProjectPodEventJoinInput join
         * @property {number|null} [project_id] ProjectPodEventJoinInput project_id
         * @property {number|null} [namespace_id] ProjectPodEventJoinInput namespace_id
         */

        /**
         * Constructs a new ProjectPodEventJoinInput.
         * @memberof websocket
         * @classdesc Represents a ProjectPodEventJoinInput.
         * @implements IProjectPodEventJoinInput
         * @constructor
         * @param {websocket.IProjectPodEventJoinInput=} [properties] Properties to set
         */
        function ProjectPodEventJoinInput(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * ProjectPodEventJoinInput type.
         * @member {websocket.Type} type
         * @memberof websocket.ProjectPodEventJoinInput
         * @instance
         */
        ProjectPodEventJoinInput.prototype.type = 0;

        /**
         * ProjectPodEventJoinInput join.
         * @member {boolean} join
         * @memberof websocket.ProjectPodEventJoinInput
         * @instance
         */
        ProjectPodEventJoinInput.prototype.join = false;

        /**
         * ProjectPodEventJoinInput project_id.
         * @member {number} project_id
         * @memberof websocket.ProjectPodEventJoinInput
         * @instance
         */
        ProjectPodEventJoinInput.prototype.project_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * ProjectPodEventJoinInput namespace_id.
         * @member {number} namespace_id
         * @memberof websocket.ProjectPodEventJoinInput
         * @instance
         */
        ProjectPodEventJoinInput.prototype.namespace_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Encodes the specified ProjectPodEventJoinInput message. Does not implicitly {@link websocket.ProjectPodEventJoinInput.verify|verify} messages.
         * @function encode
         * @memberof websocket.ProjectPodEventJoinInput
         * @static
         * @param {websocket.ProjectPodEventJoinInput} message ProjectPodEventJoinInput message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ProjectPodEventJoinInput.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.type != null && Object.hasOwnProperty.call(message, "type"))
                writer.uint32(/* id 1, wireType 0 =*/8).int32(message.type);
            if (message.join != null && Object.hasOwnProperty.call(message, "join"))
                writer.uint32(/* id 2, wireType 0 =*/16).bool(message.join);
            if (message.project_id != null && Object.hasOwnProperty.call(message, "project_id"))
                writer.uint32(/* id 3, wireType 0 =*/24).int64(message.project_id);
            if (message.namespace_id != null && Object.hasOwnProperty.call(message, "namespace_id"))
                writer.uint32(/* id 4, wireType 0 =*/32).int64(message.namespace_id);
            return writer;
        };

        /**
         * Decodes a ProjectPodEventJoinInput message from the specified reader or buffer.
         * @function decode
         * @memberof websocket.ProjectPodEventJoinInput
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {websocket.ProjectPodEventJoinInput} ProjectPodEventJoinInput
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ProjectPodEventJoinInput.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.websocket.ProjectPodEventJoinInput();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1: {
                        message.type = reader.int32();
                        break;
                    }
                case 2: {
                        message.join = reader.bool();
                        break;
                    }
                case 3: {
                        message.project_id = reader.int64();
                        break;
                    }
                case 4: {
                        message.namespace_id = reader.int64();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for ProjectPodEventJoinInput
         * @function getTypeUrl
         * @memberof websocket.ProjectPodEventJoinInput
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        ProjectPodEventJoinInput.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/websocket.ProjectPodEventJoinInput";
        };

        return ProjectPodEventJoinInput;
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
                case 1: {
                        message.type = reader.int32();
                        break;
                    }
                case 2: {
                        message.message = $root.websocket.TerminalMessage.decode(reader, reader.uint32());
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for TerminalMessageInput
         * @function getTypeUrl
         * @memberof websocket.TerminalMessageInput
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        TerminalMessageInput.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/websocket.TerminalMessageInput";
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
                case 1: {
                        message.type = reader.int32();
                        break;
                    }
                case 2: {
                        message.container = $root.types.Container.decode(reader, reader.uint32());
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for WsHandleExecShellInput
         * @function getTypeUrl
         * @memberof websocket.WsHandleExecShellInput
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        WsHandleExecShellInput.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/websocket.WsHandleExecShellInput";
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
                case 1: {
                        message.type = reader.int32();
                        break;
                    }
                case 2: {
                        message.namespace_id = reader.int64();
                        break;
                    }
                case 3: {
                        message.name = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for CancelInput
         * @function getTypeUrl
         * @memberof websocket.CancelInput
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        CancelInput.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/websocket.CancelInput";
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
                case 1: {
                        message.type = reader.int32();
                        break;
                    }
                case 2: {
                        message.namespace_id = reader.int64();
                        break;
                    }
                case 3: {
                        message.name = reader.string();
                        break;
                    }
                case 4: {
                        message.git_project_id = reader.int64();
                        break;
                    }
                case 5: {
                        message.git_branch = reader.string();
                        break;
                    }
                case 6: {
                        message.git_commit = reader.string();
                        break;
                    }
                case 7: {
                        message.config = reader.string();
                        break;
                    }
                case 8: {
                        message.atomic = reader.bool();
                        break;
                    }
                case 9: {
                        if (!(message.extra_values && message.extra_values.length))
                            message.extra_values = [];
                        message.extra_values.push($root.types.ExtraValue.decode(reader, reader.uint32()));
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for CreateProjectInput
         * @function getTypeUrl
         * @memberof websocket.CreateProjectInput
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        CreateProjectInput.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/websocket.CreateProjectInput";
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
                case 1: {
                        message.type = reader.int32();
                        break;
                    }
                case 2: {
                        message.project_id = reader.int64();
                        break;
                    }
                case 3: {
                        message.git_branch = reader.string();
                        break;
                    }
                case 4: {
                        message.git_commit = reader.string();
                        break;
                    }
                case 5: {
                        message.config = reader.string();
                        break;
                    }
                case 6: {
                        message.atomic = reader.bool();
                        break;
                    }
                case 7: {
                        if (!(message.extra_values && message.extra_values.length))
                            message.extra_values = [];
                        message.extra_values.push($root.types.ExtraValue.decode(reader, reader.uint32()));
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for UpdateProjectInput
         * @function getTypeUrl
         * @memberof websocket.UpdateProjectInput
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        UpdateProjectInput.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/websocket.UpdateProjectInput";
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
         * @property {number|null} [percent] Metadata percent
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
         * Metadata percent.
         * @member {number} percent
         * @memberof websocket.Metadata
         * @instance
         */
        Metadata.prototype.percent = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

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
            if (message.percent != null && Object.hasOwnProperty.call(message, "percent"))
                writer.uint32(/* id 9, wireType 0 =*/72).int64(message.percent);
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
                case 1: {
                        message.id = reader.string();
                        break;
                    }
                case 2: {
                        message.uid = reader.string();
                        break;
                    }
                case 3: {
                        message.slug = reader.string();
                        break;
                    }
                case 4: {
                        message.type = reader.int32();
                        break;
                    }
                case 5: {
                        message.end = reader.bool();
                        break;
                    }
                case 6: {
                        message.result = reader.int32();
                        break;
                    }
                case 7: {
                        message.to = reader.int32();
                        break;
                    }
                case 8: {
                        message.message = reader.string();
                        break;
                    }
                case 9: {
                        message.percent = reader.int64();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for Metadata
         * @function getTypeUrl
         * @memberof websocket.Metadata
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        Metadata.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/websocket.Metadata";
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
                case 1: {
                        message.metadata = $root.websocket.Metadata.decode(reader, reader.uint32());
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for WsMetadataResponse
         * @function getTypeUrl
         * @memberof websocket.WsMetadataResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        WsMetadataResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/websocket.WsMetadataResponse";
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
                case 1: {
                        message.metadata = $root.websocket.Metadata.decode(reader, reader.uint32());
                        break;
                    }
                case 2: {
                        message.terminal_message = $root.websocket.TerminalMessage.decode(reader, reader.uint32());
                        break;
                    }
                case 3: {
                        message.container = $root.types.Container.decode(reader, reader.uint32());
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for WsHandleShellResponse
         * @function getTypeUrl
         * @memberof websocket.WsHandleShellResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        WsHandleShellResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/websocket.WsHandleShellResponse";
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
                case 1: {
                        message.metadata = $root.websocket.Metadata.decode(reader, reader.uint32());
                        break;
                    }
                case 2: {
                        message.info = $root.cluster.InfoResponse.decode(reader, reader.uint32());
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for WsHandleClusterResponse
         * @function getTypeUrl
         * @memberof websocket.WsHandleClusterResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        WsHandleClusterResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/websocket.WsHandleClusterResponse";
        };

        return WsHandleClusterResponse;
    })();

    websocket.WsWithContainerMessageResponse = (function() {

        /**
         * Properties of a WsWithContainerMessageResponse.
         * @memberof websocket
         * @interface IWsWithContainerMessageResponse
         * @property {websocket.Metadata|null} [metadata] WsWithContainerMessageResponse metadata
         * @property {Array.<types.Container>|null} [containers] WsWithContainerMessageResponse containers
         */

        /**
         * Constructs a new WsWithContainerMessageResponse.
         * @memberof websocket
         * @classdesc Represents a WsWithContainerMessageResponse.
         * @implements IWsWithContainerMessageResponse
         * @constructor
         * @param {websocket.IWsWithContainerMessageResponse=} [properties] Properties to set
         */
        function WsWithContainerMessageResponse(properties) {
            this.containers = [];
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * WsWithContainerMessageResponse metadata.
         * @member {websocket.Metadata|null|undefined} metadata
         * @memberof websocket.WsWithContainerMessageResponse
         * @instance
         */
        WsWithContainerMessageResponse.prototype.metadata = null;

        /**
         * WsWithContainerMessageResponse containers.
         * @member {Array.<types.Container>} containers
         * @memberof websocket.WsWithContainerMessageResponse
         * @instance
         */
        WsWithContainerMessageResponse.prototype.containers = $util.emptyArray;

        /**
         * Encodes the specified WsWithContainerMessageResponse message. Does not implicitly {@link websocket.WsWithContainerMessageResponse.verify|verify} messages.
         * @function encode
         * @memberof websocket.WsWithContainerMessageResponse
         * @static
         * @param {websocket.WsWithContainerMessageResponse} message WsWithContainerMessageResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        WsWithContainerMessageResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.metadata != null && Object.hasOwnProperty.call(message, "metadata"))
                $root.websocket.Metadata.encode(message.metadata, writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
            if (message.containers != null && message.containers.length)
                for (let i = 0; i < message.containers.length; ++i)
                    $root.types.Container.encode(message.containers[i], writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim();
            return writer;
        };

        /**
         * Decodes a WsWithContainerMessageResponse message from the specified reader or buffer.
         * @function decode
         * @memberof websocket.WsWithContainerMessageResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {websocket.WsWithContainerMessageResponse} WsWithContainerMessageResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        WsWithContainerMessageResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.websocket.WsWithContainerMessageResponse();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1: {
                        message.metadata = $root.websocket.Metadata.decode(reader, reader.uint32());
                        break;
                    }
                case 2: {
                        if (!(message.containers && message.containers.length))
                            message.containers = [];
                        message.containers.push($root.types.Container.decode(reader, reader.uint32()));
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for WsWithContainerMessageResponse
         * @function getTypeUrl
         * @memberof websocket.WsWithContainerMessageResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        WsWithContainerMessageResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/websocket.WsWithContainerMessageResponse";
        };

        return WsWithContainerMessageResponse;
    })();

    websocket.WsProjectPodEventResponse = (function() {

        /**
         * Properties of a WsProjectPodEventResponse.
         * @memberof websocket
         * @interface IWsProjectPodEventResponse
         * @property {websocket.Metadata|null} [metadata] WsProjectPodEventResponse metadata
         * @property {number|null} [project_id] WsProjectPodEventResponse project_id
         */

        /**
         * Constructs a new WsProjectPodEventResponse.
         * @memberof websocket
         * @classdesc Represents a WsProjectPodEventResponse.
         * @implements IWsProjectPodEventResponse
         * @constructor
         * @param {websocket.IWsProjectPodEventResponse=} [properties] Properties to set
         */
        function WsProjectPodEventResponse(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * WsProjectPodEventResponse metadata.
         * @member {websocket.Metadata|null|undefined} metadata
         * @memberof websocket.WsProjectPodEventResponse
         * @instance
         */
        WsProjectPodEventResponse.prototype.metadata = null;

        /**
         * WsProjectPodEventResponse project_id.
         * @member {number} project_id
         * @memberof websocket.WsProjectPodEventResponse
         * @instance
         */
        WsProjectPodEventResponse.prototype.project_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Encodes the specified WsProjectPodEventResponse message. Does not implicitly {@link websocket.WsProjectPodEventResponse.verify|verify} messages.
         * @function encode
         * @memberof websocket.WsProjectPodEventResponse
         * @static
         * @param {websocket.WsProjectPodEventResponse} message WsProjectPodEventResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        WsProjectPodEventResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.metadata != null && Object.hasOwnProperty.call(message, "metadata"))
                $root.websocket.Metadata.encode(message.metadata, writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
            if (message.project_id != null && Object.hasOwnProperty.call(message, "project_id"))
                writer.uint32(/* id 2, wireType 0 =*/16).int64(message.project_id);
            return writer;
        };

        /**
         * Decodes a WsProjectPodEventResponse message from the specified reader or buffer.
         * @function decode
         * @memberof websocket.WsProjectPodEventResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {websocket.WsProjectPodEventResponse} WsProjectPodEventResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        WsProjectPodEventResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.websocket.WsProjectPodEventResponse();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1: {
                        message.metadata = $root.websocket.Metadata.decode(reader, reader.uint32());
                        break;
                    }
                case 2: {
                        message.project_id = reader.int64();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Gets the default type url for WsProjectPodEventResponse
         * @function getTypeUrl
         * @memberof websocket.WsProjectPodEventResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        WsProjectPodEventResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/websocket.WsProjectPodEventResponse";
        };

        return WsProjectPodEventResponse;
    })();

    return websocket;
})();

export { $root as default };
