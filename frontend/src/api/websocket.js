/*eslint-disable block-scoped-var, id-length, no-control-regex, no-magic-numbers, no-prototype-builtins, no-redeclare, no-shadow, no-var, sort-vars*/
import * as $protobuf from "protobufjs/minimal";

// Common aliases
const $Reader = $protobuf.Reader, $Writer = $protobuf.Writer, $util = $protobuf.util;

// Exported root namespace
const $root = $protobuf.roots["default"] || ($protobuf.roots["default"] = {});

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

    websocket.ClusterInfo = (function() {

        /**
         * Properties of a ClusterInfo.
         * @memberof websocket
         * @interface IClusterInfo
         * @property {string|null} [status] ClusterInfo status
         * @property {string|null} [freeMemory] ClusterInfo freeMemory
         * @property {string|null} [freeCpu] ClusterInfo freeCpu
         * @property {string|null} [freeRequestMemory] ClusterInfo freeRequestMemory
         * @property {string|null} [freeRequestCpu] ClusterInfo freeRequestCpu
         * @property {string|null} [totalMemory] ClusterInfo totalMemory
         * @property {string|null} [totalCpu] ClusterInfo totalCpu
         * @property {string|null} [usageMemoryRate] ClusterInfo usageMemoryRate
         * @property {string|null} [usageCpuRate] ClusterInfo usageCpuRate
         * @property {string|null} [requestMemoryRate] ClusterInfo requestMemoryRate
         * @property {string|null} [requestCpuRate] ClusterInfo requestCpuRate
         */

        /**
         * Constructs a new ClusterInfo.
         * @memberof websocket
         * @classdesc Represents a ClusterInfo.
         * @implements IClusterInfo
         * @constructor
         * @param {websocket.IClusterInfo=} [properties] Properties to set
         */
        function ClusterInfo(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * ClusterInfo status.
         * @member {string} status
         * @memberof websocket.ClusterInfo
         * @instance
         */
        ClusterInfo.prototype.status = "";

        /**
         * ClusterInfo freeMemory.
         * @member {string} freeMemory
         * @memberof websocket.ClusterInfo
         * @instance
         */
        ClusterInfo.prototype.freeMemory = "";

        /**
         * ClusterInfo freeCpu.
         * @member {string} freeCpu
         * @memberof websocket.ClusterInfo
         * @instance
         */
        ClusterInfo.prototype.freeCpu = "";

        /**
         * ClusterInfo freeRequestMemory.
         * @member {string} freeRequestMemory
         * @memberof websocket.ClusterInfo
         * @instance
         */
        ClusterInfo.prototype.freeRequestMemory = "";

        /**
         * ClusterInfo freeRequestCpu.
         * @member {string} freeRequestCpu
         * @memberof websocket.ClusterInfo
         * @instance
         */
        ClusterInfo.prototype.freeRequestCpu = "";

        /**
         * ClusterInfo totalMemory.
         * @member {string} totalMemory
         * @memberof websocket.ClusterInfo
         * @instance
         */
        ClusterInfo.prototype.totalMemory = "";

        /**
         * ClusterInfo totalCpu.
         * @member {string} totalCpu
         * @memberof websocket.ClusterInfo
         * @instance
         */
        ClusterInfo.prototype.totalCpu = "";

        /**
         * ClusterInfo usageMemoryRate.
         * @member {string} usageMemoryRate
         * @memberof websocket.ClusterInfo
         * @instance
         */
        ClusterInfo.prototype.usageMemoryRate = "";

        /**
         * ClusterInfo usageCpuRate.
         * @member {string} usageCpuRate
         * @memberof websocket.ClusterInfo
         * @instance
         */
        ClusterInfo.prototype.usageCpuRate = "";

        /**
         * ClusterInfo requestMemoryRate.
         * @member {string} requestMemoryRate
         * @memberof websocket.ClusterInfo
         * @instance
         */
        ClusterInfo.prototype.requestMemoryRate = "";

        /**
         * ClusterInfo requestCpuRate.
         * @member {string} requestCpuRate
         * @memberof websocket.ClusterInfo
         * @instance
         */
        ClusterInfo.prototype.requestCpuRate = "";

        /**
         * Encodes the specified ClusterInfo message. Does not implicitly {@link websocket.ClusterInfo.verify|verify} messages.
         * @function encode
         * @memberof websocket.ClusterInfo
         * @static
         * @param {websocket.ClusterInfo} message ClusterInfo message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ClusterInfo.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.status != null && Object.hasOwnProperty.call(message, "status"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.status);
            if (message.freeMemory != null && Object.hasOwnProperty.call(message, "freeMemory"))
                writer.uint32(/* id 2, wireType 2 =*/18).string(message.freeMemory);
            if (message.freeCpu != null && Object.hasOwnProperty.call(message, "freeCpu"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.freeCpu);
            if (message.freeRequestMemory != null && Object.hasOwnProperty.call(message, "freeRequestMemory"))
                writer.uint32(/* id 4, wireType 2 =*/34).string(message.freeRequestMemory);
            if (message.freeRequestCpu != null && Object.hasOwnProperty.call(message, "freeRequestCpu"))
                writer.uint32(/* id 5, wireType 2 =*/42).string(message.freeRequestCpu);
            if (message.totalMemory != null && Object.hasOwnProperty.call(message, "totalMemory"))
                writer.uint32(/* id 6, wireType 2 =*/50).string(message.totalMemory);
            if (message.totalCpu != null && Object.hasOwnProperty.call(message, "totalCpu"))
                writer.uint32(/* id 7, wireType 2 =*/58).string(message.totalCpu);
            if (message.usageMemoryRate != null && Object.hasOwnProperty.call(message, "usageMemoryRate"))
                writer.uint32(/* id 8, wireType 2 =*/66).string(message.usageMemoryRate);
            if (message.usageCpuRate != null && Object.hasOwnProperty.call(message, "usageCpuRate"))
                writer.uint32(/* id 9, wireType 2 =*/74).string(message.usageCpuRate);
            if (message.requestMemoryRate != null && Object.hasOwnProperty.call(message, "requestMemoryRate"))
                writer.uint32(/* id 10, wireType 2 =*/82).string(message.requestMemoryRate);
            if (message.requestCpuRate != null && Object.hasOwnProperty.call(message, "requestCpuRate"))
                writer.uint32(/* id 11, wireType 2 =*/90).string(message.requestCpuRate);
            return writer;
        };

        /**
         * Decodes a ClusterInfo message from the specified reader or buffer.
         * @function decode
         * @memberof websocket.ClusterInfo
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {websocket.ClusterInfo} ClusterInfo
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ClusterInfo.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.websocket.ClusterInfo();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1: {
                        message.status = reader.string();
                        break;
                    }
                case 2: {
                        message.freeMemory = reader.string();
                        break;
                    }
                case 3: {
                        message.freeCpu = reader.string();
                        break;
                    }
                case 4: {
                        message.freeRequestMemory = reader.string();
                        break;
                    }
                case 5: {
                        message.freeRequestCpu = reader.string();
                        break;
                    }
                case 6: {
                        message.totalMemory = reader.string();
                        break;
                    }
                case 7: {
                        message.totalCpu = reader.string();
                        break;
                    }
                case 8: {
                        message.usageMemoryRate = reader.string();
                        break;
                    }
                case 9: {
                        message.usageCpuRate = reader.string();
                        break;
                    }
                case 10: {
                        message.requestMemoryRate = reader.string();
                        break;
                    }
                case 11: {
                        message.requestCpuRate = reader.string();
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
         * Gets the default type url for ClusterInfo
         * @function getTypeUrl
         * @memberof websocket.ClusterInfo
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        ClusterInfo.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/websocket.ClusterInfo";
        };

        return ClusterInfo;
    })();

    websocket.ExtraValue = (function() {

        /**
         * Properties of an ExtraValue.
         * @memberof websocket
         * @interface IExtraValue
         * @property {string|null} [path] ExtraValue path
         * @property {string|null} [value] ExtraValue value
         */

        /**
         * Constructs a new ExtraValue.
         * @memberof websocket
         * @classdesc Represents an ExtraValue.
         * @implements IExtraValue
         * @constructor
         * @param {websocket.IExtraValue=} [properties] Properties to set
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
         * @memberof websocket.ExtraValue
         * @instance
         */
        ExtraValue.prototype.path = "";

        /**
         * ExtraValue value.
         * @member {string} value
         * @memberof websocket.ExtraValue
         * @instance
         */
        ExtraValue.prototype.value = "";

        /**
         * Encodes the specified ExtraValue message. Does not implicitly {@link websocket.ExtraValue.verify|verify} messages.
         * @function encode
         * @memberof websocket.ExtraValue
         * @static
         * @param {websocket.ExtraValue} message ExtraValue message or plain object to encode
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
         * @memberof websocket.ExtraValue
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {websocket.ExtraValue} ExtraValue
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ExtraValue.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.websocket.ExtraValue();
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
         * @memberof websocket.ExtraValue
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        ExtraValue.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/websocket.ExtraValue";
        };

        return ExtraValue;
    })();

    websocket.Container = (function() {

        /**
         * Properties of a Container.
         * @memberof websocket
         * @interface IContainer
         * @property {string|null} [namespace] Container namespace
         * @property {string|null} [pod] Container pod
         * @property {string|null} [container] Container container
         */

        /**
         * Constructs a new Container.
         * @memberof websocket
         * @classdesc Represents a Container.
         * @implements IContainer
         * @constructor
         * @param {websocket.IContainer=} [properties] Properties to set
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
         * @memberof websocket.Container
         * @instance
         */
        Container.prototype.namespace = "";

        /**
         * Container pod.
         * @member {string} pod
         * @memberof websocket.Container
         * @instance
         */
        Container.prototype.pod = "";

        /**
         * Container container.
         * @member {string} container
         * @memberof websocket.Container
         * @instance
         */
        Container.prototype.container = "";

        /**
         * Encodes the specified Container message. Does not implicitly {@link websocket.Container.verify|verify} messages.
         * @function encode
         * @memberof websocket.Container
         * @static
         * @param {websocket.Container} message Container message or plain object to encode
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
         * @memberof websocket.Container
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {websocket.Container} Container
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Container.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.websocket.Container();
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
         * @memberof websocket.Container
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        Container.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/websocket.Container";
        };

        return Container;
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
         * @property {string|null} [sessionId] TerminalMessage sessionId
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
         * TerminalMessage sessionId.
         * @member {string} sessionId
         * @memberof websocket.TerminalMessage
         * @instance
         */
        TerminalMessage.prototype.sessionId = "";

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
            if (message.sessionId != null && Object.hasOwnProperty.call(message, "sessionId"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.sessionId);
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
                        message.sessionId = reader.string();
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
         * @property {number|null} [projectId] ProjectPodEventJoinInput projectId
         * @property {number|null} [namespaceId] ProjectPodEventJoinInput namespaceId
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
         * ProjectPodEventJoinInput projectId.
         * @member {number} projectId
         * @memberof websocket.ProjectPodEventJoinInput
         * @instance
         */
        ProjectPodEventJoinInput.prototype.projectId = 0;

        /**
         * ProjectPodEventJoinInput namespaceId.
         * @member {number} namespaceId
         * @memberof websocket.ProjectPodEventJoinInput
         * @instance
         */
        ProjectPodEventJoinInput.prototype.namespaceId = 0;

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
            if (message.projectId != null && Object.hasOwnProperty.call(message, "projectId"))
                writer.uint32(/* id 3, wireType 0 =*/24).int32(message.projectId);
            if (message.namespaceId != null && Object.hasOwnProperty.call(message, "namespaceId"))
                writer.uint32(/* id 4, wireType 0 =*/32).int32(message.namespaceId);
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
                        message.projectId = reader.int32();
                        break;
                    }
                case 4: {
                        message.namespaceId = reader.int32();
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
         * @property {websocket.Container|null} [container] WsHandleExecShellInput container
         * @property {string|null} [sessionId] WsHandleExecShellInput sessionId
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
         * @member {websocket.Container|null|undefined} container
         * @memberof websocket.WsHandleExecShellInput
         * @instance
         */
        WsHandleExecShellInput.prototype.container = null;

        /**
         * WsHandleExecShellInput sessionId.
         * @member {string} sessionId
         * @memberof websocket.WsHandleExecShellInput
         * @instance
         */
        WsHandleExecShellInput.prototype.sessionId = "";

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
                $root.websocket.Container.encode(message.container, writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim();
            if (message.sessionId != null && Object.hasOwnProperty.call(message, "sessionId"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.sessionId);
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
                        message.container = $root.websocket.Container.decode(reader, reader.uint32());
                        break;
                    }
                case 3: {
                        message.sessionId = reader.string();
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
         * @property {number|null} [namespaceId] CancelInput namespaceId
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
         * CancelInput namespaceId.
         * @member {number} namespaceId
         * @memberof websocket.CancelInput
         * @instance
         */
        CancelInput.prototype.namespaceId = 0;

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
            if (message.namespaceId != null && Object.hasOwnProperty.call(message, "namespaceId"))
                writer.uint32(/* id 2, wireType 0 =*/16).int32(message.namespaceId);
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
                        message.namespaceId = reader.int32();
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
         * @property {number|null} [namespaceId] CreateProjectInput namespaceId
         * @property {string|null} [name] CreateProjectInput name
         * @property {number|null} [repoId] CreateProjectInput repoId
         * @property {string|null} [gitBranch] CreateProjectInput gitBranch
         * @property {string|null} [gitCommit] CreateProjectInput gitCommit
         * @property {string|null} [config] CreateProjectInput config
         * @property {Array.<websocket.ExtraValue>|null} [extraValues] CreateProjectInput extraValues
         * @property {boolean|null} [atomic] CreateProjectInput atomic
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
            this.extraValues = [];
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
         * CreateProjectInput namespaceId.
         * @member {number} namespaceId
         * @memberof websocket.CreateProjectInput
         * @instance
         */
        CreateProjectInput.prototype.namespaceId = 0;

        /**
         * CreateProjectInput name.
         * @member {string|null|undefined} name
         * @memberof websocket.CreateProjectInput
         * @instance
         */
        CreateProjectInput.prototype.name = null;

        /**
         * CreateProjectInput repoId.
         * @member {number} repoId
         * @memberof websocket.CreateProjectInput
         * @instance
         */
        CreateProjectInput.prototype.repoId = 0;

        /**
         * CreateProjectInput gitBranch.
         * @member {string} gitBranch
         * @memberof websocket.CreateProjectInput
         * @instance
         */
        CreateProjectInput.prototype.gitBranch = "";

        /**
         * CreateProjectInput gitCommit.
         * @member {string} gitCommit
         * @memberof websocket.CreateProjectInput
         * @instance
         */
        CreateProjectInput.prototype.gitCommit = "";

        /**
         * CreateProjectInput config.
         * @member {string} config
         * @memberof websocket.CreateProjectInput
         * @instance
         */
        CreateProjectInput.prototype.config = "";

        /**
         * CreateProjectInput extraValues.
         * @member {Array.<websocket.ExtraValue>} extraValues
         * @memberof websocket.CreateProjectInput
         * @instance
         */
        CreateProjectInput.prototype.extraValues = $util.emptyArray;

        /**
         * CreateProjectInput atomic.
         * @member {boolean|null|undefined} atomic
         * @memberof websocket.CreateProjectInput
         * @instance
         */
        CreateProjectInput.prototype.atomic = null;

        // OneOf field names bound to virtual getters and setters
        let $oneOfFields;

        // Virtual OneOf for proto3 optional field
        Object.defineProperty(CreateProjectInput.prototype, "_name", {
            get: $util.oneOfGetter($oneOfFields = ["name"]),
            set: $util.oneOfSetter($oneOfFields)
        });

        // Virtual OneOf for proto3 optional field
        Object.defineProperty(CreateProjectInput.prototype, "_atomic", {
            get: $util.oneOfGetter($oneOfFields = ["atomic"]),
            set: $util.oneOfSetter($oneOfFields)
        });

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
            if (message.namespaceId != null && Object.hasOwnProperty.call(message, "namespaceId"))
                writer.uint32(/* id 2, wireType 0 =*/16).int32(message.namespaceId);
            if (message.name != null && Object.hasOwnProperty.call(message, "name"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.name);
            if (message.repoId != null && Object.hasOwnProperty.call(message, "repoId"))
                writer.uint32(/* id 4, wireType 0 =*/32).int32(message.repoId);
            if (message.gitBranch != null && Object.hasOwnProperty.call(message, "gitBranch"))
                writer.uint32(/* id 5, wireType 2 =*/42).string(message.gitBranch);
            if (message.gitCommit != null && Object.hasOwnProperty.call(message, "gitCommit"))
                writer.uint32(/* id 6, wireType 2 =*/50).string(message.gitCommit);
            if (message.config != null && Object.hasOwnProperty.call(message, "config"))
                writer.uint32(/* id 7, wireType 2 =*/58).string(message.config);
            if (message.extraValues != null && message.extraValues.length)
                for (let i = 0; i < message.extraValues.length; ++i)
                    $root.websocket.ExtraValue.encode(message.extraValues[i], writer.uint32(/* id 8, wireType 2 =*/66).fork()).ldelim();
            if (message.atomic != null && Object.hasOwnProperty.call(message, "atomic"))
                writer.uint32(/* id 9, wireType 0 =*/72).bool(message.atomic);
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
                        message.namespaceId = reader.int32();
                        break;
                    }
                case 3: {
                        message.name = reader.string();
                        break;
                    }
                case 4: {
                        message.repoId = reader.int32();
                        break;
                    }
                case 5: {
                        message.gitBranch = reader.string();
                        break;
                    }
                case 6: {
                        message.gitCommit = reader.string();
                        break;
                    }
                case 7: {
                        message.config = reader.string();
                        break;
                    }
                case 8: {
                        if (!(message.extraValues && message.extraValues.length))
                            message.extraValues = [];
                        message.extraValues.push($root.websocket.ExtraValue.decode(reader, reader.uint32()));
                        break;
                    }
                case 9: {
                        message.atomic = reader.bool();
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
         * @property {number|null} [projectId] UpdateProjectInput projectId
         * @property {string|null} [gitBranch] UpdateProjectInput gitBranch
         * @property {string|null} [gitCommit] UpdateProjectInput gitCommit
         * @property {string|null} [config] UpdateProjectInput config
         * @property {Array.<websocket.ExtraValue>|null} [extraValues] UpdateProjectInput extraValues
         * @property {number|null} [version] UpdateProjectInput version
         * @property {boolean|null} [atomic] UpdateProjectInput atomic
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
            this.extraValues = [];
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
         * UpdateProjectInput projectId.
         * @member {number} projectId
         * @memberof websocket.UpdateProjectInput
         * @instance
         */
        UpdateProjectInput.prototype.projectId = 0;

        /**
         * UpdateProjectInput gitBranch.
         * @member {string} gitBranch
         * @memberof websocket.UpdateProjectInput
         * @instance
         */
        UpdateProjectInput.prototype.gitBranch = "";

        /**
         * UpdateProjectInput gitCommit.
         * @member {string} gitCommit
         * @memberof websocket.UpdateProjectInput
         * @instance
         */
        UpdateProjectInput.prototype.gitCommit = "";

        /**
         * UpdateProjectInput config.
         * @member {string} config
         * @memberof websocket.UpdateProjectInput
         * @instance
         */
        UpdateProjectInput.prototype.config = "";

        /**
         * UpdateProjectInput extraValues.
         * @member {Array.<websocket.ExtraValue>} extraValues
         * @memberof websocket.UpdateProjectInput
         * @instance
         */
        UpdateProjectInput.prototype.extraValues = $util.emptyArray;

        /**
         * UpdateProjectInput version.
         * @member {number} version
         * @memberof websocket.UpdateProjectInput
         * @instance
         */
        UpdateProjectInput.prototype.version = 0;

        /**
         * UpdateProjectInput atomic.
         * @member {boolean|null|undefined} atomic
         * @memberof websocket.UpdateProjectInput
         * @instance
         */
        UpdateProjectInput.prototype.atomic = null;

        // OneOf field names bound to virtual getters and setters
        let $oneOfFields;

        // Virtual OneOf for proto3 optional field
        Object.defineProperty(UpdateProjectInput.prototype, "_atomic", {
            get: $util.oneOfGetter($oneOfFields = ["atomic"]),
            set: $util.oneOfSetter($oneOfFields)
        });

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
            if (message.projectId != null && Object.hasOwnProperty.call(message, "projectId"))
                writer.uint32(/* id 2, wireType 0 =*/16).int32(message.projectId);
            if (message.gitBranch != null && Object.hasOwnProperty.call(message, "gitBranch"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.gitBranch);
            if (message.gitCommit != null && Object.hasOwnProperty.call(message, "gitCommit"))
                writer.uint32(/* id 4, wireType 2 =*/34).string(message.gitCommit);
            if (message.config != null && Object.hasOwnProperty.call(message, "config"))
                writer.uint32(/* id 5, wireType 2 =*/42).string(message.config);
            if (message.extraValues != null && message.extraValues.length)
                for (let i = 0; i < message.extraValues.length; ++i)
                    $root.websocket.ExtraValue.encode(message.extraValues[i], writer.uint32(/* id 6, wireType 2 =*/50).fork()).ldelim();
            if (message.version != null && Object.hasOwnProperty.call(message, "version"))
                writer.uint32(/* id 7, wireType 0 =*/56).int32(message.version);
            if (message.atomic != null && Object.hasOwnProperty.call(message, "atomic"))
                writer.uint32(/* id 8, wireType 0 =*/64).bool(message.atomic);
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
                        message.projectId = reader.int32();
                        break;
                    }
                case 3: {
                        message.gitBranch = reader.string();
                        break;
                    }
                case 4: {
                        message.gitCommit = reader.string();
                        break;
                    }
                case 5: {
                        message.config = reader.string();
                        break;
                    }
                case 6: {
                        if (!(message.extraValues && message.extraValues.length))
                            message.extraValues = [];
                        message.extraValues.push($root.websocket.ExtraValue.decode(reader, reader.uint32()));
                        break;
                    }
                case 7: {
                        message.version = reader.int32();
                        break;
                    }
                case 8: {
                        message.atomic = reader.bool();
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
        Metadata.prototype.percent = 0;

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
                writer.uint32(/* id 9, wireType 0 =*/72).int32(message.percent);
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
                        message.percent = reader.int32();
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
         * @property {websocket.TerminalMessage|null} [terminalMessage] WsHandleShellResponse terminalMessage
         * @property {websocket.Container|null} [container] WsHandleShellResponse container
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
         * WsHandleShellResponse terminalMessage.
         * @member {websocket.TerminalMessage|null|undefined} terminalMessage
         * @memberof websocket.WsHandleShellResponse
         * @instance
         */
        WsHandleShellResponse.prototype.terminalMessage = null;

        /**
         * WsHandleShellResponse container.
         * @member {websocket.Container|null|undefined} container
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
            if (message.terminalMessage != null && Object.hasOwnProperty.call(message, "terminalMessage"))
                $root.websocket.TerminalMessage.encode(message.terminalMessage, writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim();
            if (message.container != null && Object.hasOwnProperty.call(message, "container"))
                $root.websocket.Container.encode(message.container, writer.uint32(/* id 3, wireType 2 =*/26).fork()).ldelim();
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
                        message.terminalMessage = $root.websocket.TerminalMessage.decode(reader, reader.uint32());
                        break;
                    }
                case 3: {
                        message.container = $root.websocket.Container.decode(reader, reader.uint32());
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
         * @property {websocket.ClusterInfo|null} [info] WsHandleClusterResponse info
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
         * @member {websocket.ClusterInfo|null|undefined} info
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
                $root.websocket.ClusterInfo.encode(message.info, writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim();
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
                        message.info = $root.websocket.ClusterInfo.decode(reader, reader.uint32());
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
         * @property {Array.<websocket.Container>|null} [containers] WsWithContainerMessageResponse containers
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
         * @member {Array.<websocket.Container>} containers
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
                    $root.websocket.Container.encode(message.containers[i], writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim();
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
                        message.containers.push($root.websocket.Container.decode(reader, reader.uint32()));
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
         * @property {number|null} [projectId] WsProjectPodEventResponse projectId
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
         * WsProjectPodEventResponse projectId.
         * @member {number} projectId
         * @memberof websocket.WsProjectPodEventResponse
         * @instance
         */
        WsProjectPodEventResponse.prototype.projectId = 0;

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
            if (message.projectId != null && Object.hasOwnProperty.call(message, "projectId"))
                writer.uint32(/* id 2, wireType 0 =*/16).int32(message.projectId);
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
                        message.projectId = reader.int32();
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

    websocket.WsReloadProjectsResponse = (function() {

        /**
         * Properties of a WsReloadProjectsResponse.
         * @memberof websocket
         * @interface IWsReloadProjectsResponse
         * @property {websocket.Metadata|null} [metadata] WsReloadProjectsResponse metadata
         * @property {number|null} [namespaceId] WsReloadProjectsResponse namespaceId
         */

        /**
         * Constructs a new WsReloadProjectsResponse.
         * @memberof websocket
         * @classdesc Represents a WsReloadProjectsResponse.
         * @implements IWsReloadProjectsResponse
         * @constructor
         * @param {websocket.IWsReloadProjectsResponse=} [properties] Properties to set
         */
        function WsReloadProjectsResponse(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * WsReloadProjectsResponse metadata.
         * @member {websocket.Metadata|null|undefined} metadata
         * @memberof websocket.WsReloadProjectsResponse
         * @instance
         */
        WsReloadProjectsResponse.prototype.metadata = null;

        /**
         * WsReloadProjectsResponse namespaceId.
         * @member {number} namespaceId
         * @memberof websocket.WsReloadProjectsResponse
         * @instance
         */
        WsReloadProjectsResponse.prototype.namespaceId = 0;

        /**
         * Encodes the specified WsReloadProjectsResponse message. Does not implicitly {@link websocket.WsReloadProjectsResponse.verify|verify} messages.
         * @function encode
         * @memberof websocket.WsReloadProjectsResponse
         * @static
         * @param {websocket.WsReloadProjectsResponse} message WsReloadProjectsResponse message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        WsReloadProjectsResponse.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.metadata != null && Object.hasOwnProperty.call(message, "metadata"))
                $root.websocket.Metadata.encode(message.metadata, writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
            if (message.namespaceId != null && Object.hasOwnProperty.call(message, "namespaceId"))
                writer.uint32(/* id 2, wireType 0 =*/16).int32(message.namespaceId);
            return writer;
        };

        /**
         * Decodes a WsReloadProjectsResponse message from the specified reader or buffer.
         * @function decode
         * @memberof websocket.WsReloadProjectsResponse
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {websocket.WsReloadProjectsResponse} WsReloadProjectsResponse
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        WsReloadProjectsResponse.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.websocket.WsReloadProjectsResponse();
            while (reader.pos < end) {
                let tag = reader.uint32();
                switch (tag >>> 3) {
                case 1: {
                        message.metadata = $root.websocket.Metadata.decode(reader, reader.uint32());
                        break;
                    }
                case 2: {
                        message.namespaceId = reader.int32();
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
         * Gets the default type url for WsReloadProjectsResponse
         * @function getTypeUrl
         * @memberof websocket.WsReloadProjectsResponse
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        WsReloadProjectsResponse.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/websocket.WsReloadProjectsResponse";
        };

        return WsReloadProjectsResponse;
    })();

    return websocket;
})();

export { $root as default };
