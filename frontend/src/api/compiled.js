/*eslint-disable block-scoped-var, id-length, no-control-regex, no-magic-numbers, no-prototype-builtins, no-redeclare, no-shadow, no-var, sort-vars*/
import * as $protobuf from "protobufjs/minimal";

// Common aliases
const $util = $protobuf.util;

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

    return LoginRequest;
})();

export const LoginResponse = $root.LoginResponse = (() => {

    /**
     * Properties of a LoginResponse.
     * @exports ILoginResponse
     * @interface ILoginResponse
     * @property {string|null} [token] LoginResponse token
     * @property {number|Long|null} [expires_in] LoginResponse expires_in
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
     * @member {number|Long} expires_in
     * @memberof LoginResponse
     * @instance
     */
    LoginResponse.prototype.expires_in = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

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

    return OidcSetting;
})();

export const SettingsResponse = $root.SettingsResponse = (() => {

    /**
     * Properties of a SettingsResponse.
     * @exports ISettingsResponse
     * @interface ISettingsResponse
     * @property {Array.<IOidcSetting>|null} [items] SettingsResponse items
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
     * @member {Array.<IOidcSetting>} items
     * @memberof SettingsResponse
     * @instance
     */
    SettingsResponse.prototype.items = $util.emptyArray;

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
     * @param {ILoginRequest} request LoginRequest message or plain object
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
     * @param {ILoginRequest} request LoginRequest message or plain object
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
     * @param {google.protobuf.IEmpty} request Empty message or plain object
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
     * @param {google.protobuf.IEmpty} request Empty message or plain object
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
     * @param {google.protobuf.IEmpty} request Empty message or plain object
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
     * @param {google.protobuf.IEmpty} request Empty message or plain object
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
     * @param {IExchangeRequest} request ExchangeRequest message or plain object
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
     * @param {IExchangeRequest} request ExchangeRequest message or plain object
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
     * @param {google.protobuf.IEmpty} request Empty message or plain object
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
     * @param {google.protobuf.IEmpty} request Empty message or plain object
     * @returns {Promise<ClusterInfoResponse>} Promise
     * @variation 2
     */

    return Cluster;
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

    return DisableProjectRequest;
})();

export const GitlabProjectInfo = $root.GitlabProjectInfo = (() => {

    /**
     * Properties of a GitlabProjectInfo.
     * @exports IGitlabProjectInfo
     * @interface IGitlabProjectInfo
     * @property {number|Long|null} [id] GitlabProjectInfo id
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
     * @member {number|Long} id
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

    return GitlabProjectInfo;
})();

export const ProjectListResponse = $root.ProjectListResponse = (() => {

    /**
     * Properties of a ProjectListResponse.
     * @exports IProjectListResponse
     * @interface IProjectListResponse
     * @property {Array.<IGitlabProjectInfo>|null} [data] ProjectListResponse data
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
     * @member {Array.<IGitlabProjectInfo>} data
     * @memberof ProjectListResponse
     * @instance
     */
    ProjectListResponse.prototype.data = $util.emptyArray;

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

    return Option;
})();

export const ProjectsResponse = $root.ProjectsResponse = (() => {

    /**
     * Properties of a ProjectsResponse.
     * @exports IProjectsResponse
     * @interface IProjectsResponse
     * @property {Array.<IOption>|null} [data] ProjectsResponse data
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
     * @member {Array.<IOption>} data
     * @memberof ProjectsResponse
     * @instance
     */
    ProjectsResponse.prototype.data = $util.emptyArray;

    return ProjectsResponse;
})();

export const BranchesRequest = $root.BranchesRequest = (() => {

    /**
     * Properties of a BranchesRequest.
     * @exports IBranchesRequest
     * @interface IBranchesRequest
     * @property {string|null} [project_id] BranchesRequest project_id
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

    return BranchesRequest;
})();

export const BranchesResponse = $root.BranchesResponse = (() => {

    /**
     * Properties of a BranchesResponse.
     * @exports IBranchesResponse
     * @interface IBranchesResponse
     * @property {Array.<IOption>|null} [data] BranchesResponse data
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
     * @member {Array.<IOption>} data
     * @memberof BranchesResponse
     * @instance
     */
    BranchesResponse.prototype.data = $util.emptyArray;

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

    return CommitsRequest;
})();

export const CommitsResponse = $root.CommitsResponse = (() => {

    /**
     * Properties of a CommitsResponse.
     * @exports ICommitsResponse
     * @interface ICommitsResponse
     * @property {Array.<IOption>|null} [data] CommitsResponse data
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
     * @member {Array.<IOption>} data
     * @memberof CommitsResponse
     * @instance
     */
    CommitsResponse.prototype.data = $util.emptyArray;

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

    return CommitRequest;
})();

export const CommitResponse = $root.CommitResponse = (() => {

    /**
     * Properties of a CommitResponse.
     * @exports ICommitResponse
     * @interface ICommitResponse
     * @property {IOption|null} [data] CommitResponse data
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
     * @member {IOption|null|undefined} data
     * @memberof CommitResponse
     * @instance
     */
    CommitResponse.prototype.data = null;

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
     * @param {IEnableProjectRequest} request EnableProjectRequest message or plain object
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
     * @param {IEnableProjectRequest} request EnableProjectRequest message or plain object
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
     * @param {IDisableProjectRequest} request DisableProjectRequest message or plain object
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
     * @param {IDisableProjectRequest} request DisableProjectRequest message or plain object
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
     * @param {google.protobuf.IEmpty} request Empty message or plain object
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
     * @param {google.protobuf.IEmpty} request Empty message or plain object
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
     * @param {google.protobuf.IEmpty} request Empty message or plain object
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
     * @param {google.protobuf.IEmpty} request Empty message or plain object
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
     * @param {IBranchesRequest} request BranchesRequest message or plain object
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
     * @param {IBranchesRequest} request BranchesRequest message or plain object
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
     * @param {ICommitsRequest} request CommitsRequest message or plain object
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
     * @param {ICommitsRequest} request CommitsRequest message or plain object
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
     * @param {ICommitRequest} request CommitRequest message or plain object
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
     * @param {ICommitRequest} request CommitRequest message or plain object
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
     * @param {IPipelineInfoRequest} request PipelineInfoRequest message or plain object
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
     * @param {IPipelineInfoRequest} request PipelineInfoRequest message or plain object
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
     * @param {IConfigFileRequest} request ConfigFileRequest message or plain object
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
     * @param {IConfigFileRequest} request ConfigFileRequest message or plain object
     * @returns {Promise<ConfigFileResponse>} Promise
     * @variation 2
     */

    return Gitlab;
})();

export const MarsShowRequest = $root.MarsShowRequest = (() => {

    /**
     * Properties of a MarsShowRequest.
     * @exports IMarsShowRequest
     * @interface IMarsShowRequest
     * @property {number|Long|null} [project_id] MarsShowRequest project_id
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
     * @member {number|Long} project_id
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

    return MarsShowRequest;
})();

export const MarsShowResponse = $root.MarsShowResponse = (() => {

    /**
     * Properties of a MarsShowResponse.
     * @exports IMarsShowResponse
     * @interface IMarsShowResponse
     * @property {string|null} [branch] MarsShowResponse branch
     * @property {string|null} [config] MarsShowResponse config
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
     * @member {string} config
     * @memberof MarsShowResponse
     * @instance
     */
    MarsShowResponse.prototype.config = "";

    return MarsShowResponse;
})();

export const GlobalConfigRequest = $root.GlobalConfigRequest = (() => {

    /**
     * Properties of a GlobalConfigRequest.
     * @exports IGlobalConfigRequest
     * @interface IGlobalConfigRequest
     * @property {number|Long|null} [project_id] GlobalConfigRequest project_id
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
     * @member {number|Long} project_id
     * @memberof GlobalConfigRequest
     * @instance
     */
    GlobalConfigRequest.prototype.project_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    return GlobalConfigRequest;
})();

export const GlobalConfigResponse = $root.GlobalConfigResponse = (() => {

    /**
     * Properties of a GlobalConfigResponse.
     * @exports IGlobalConfigResponse
     * @interface IGlobalConfigResponse
     * @property {boolean|null} [enabled] GlobalConfigResponse enabled
     * @property {string|null} [config] GlobalConfigResponse config
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
     * @member {string} config
     * @memberof GlobalConfigResponse
     * @instance
     */
    GlobalConfigResponse.prototype.config = "";

    return GlobalConfigResponse;
})();

export const MarsUpdateRequest = $root.MarsUpdateRequest = (() => {

    /**
     * Properties of a MarsUpdateRequest.
     * @exports IMarsUpdateRequest
     * @interface IMarsUpdateRequest
     * @property {number|Long|null} [project_id] MarsUpdateRequest project_id
     * @property {string|null} [config] MarsUpdateRequest config
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
     * @member {number|Long} project_id
     * @memberof MarsUpdateRequest
     * @instance
     */
    MarsUpdateRequest.prototype.project_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * MarsUpdateRequest config.
     * @member {string} config
     * @memberof MarsUpdateRequest
     * @instance
     */
    MarsUpdateRequest.prototype.config = "";

    return MarsUpdateRequest;
})();

export const MarsUpdateResponse = $root.MarsUpdateResponse = (() => {

    /**
     * Properties of a MarsUpdateResponse.
     * @exports IMarsUpdateResponse
     * @interface IMarsUpdateResponse
     * @property {IGitlabProjectModal|null} [data] MarsUpdateResponse data
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
     * MarsUpdateResponse data.
     * @member {IGitlabProjectModal|null|undefined} data
     * @memberof MarsUpdateResponse
     * @instance
     */
    MarsUpdateResponse.prototype.data = null;

    return MarsUpdateResponse;
})();

export const ToggleEnabledRequest = $root.ToggleEnabledRequest = (() => {

    /**
     * Properties of a ToggleEnabledRequest.
     * @exports IToggleEnabledRequest
     * @interface IToggleEnabledRequest
     * @property {number|Long|null} [project_id] ToggleEnabledRequest project_id
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
     * @member {number|Long} project_id
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

    return ToggleEnabledRequest;
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
     * @param {IMarsShowRequest} request MarsShowRequest message or plain object
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
     * @param {IMarsShowRequest} request MarsShowRequest message or plain object
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
     * @param {IGlobalConfigRequest} request GlobalConfigRequest message or plain object
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
     * @param {IGlobalConfigRequest} request GlobalConfigRequest message or plain object
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
     * @param {IToggleEnabledRequest} request ToggleEnabledRequest message or plain object
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
     * @param {IToggleEnabledRequest} request ToggleEnabledRequest message or plain object
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
     * @param {IMarsUpdateRequest} request MarsUpdateRequest message or plain object
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
     * @param {IMarsUpdateRequest} request MarsUpdateRequest message or plain object
     * @returns {Promise<MarsUpdateResponse>} Promise
     * @variation 2
     */

    return Mars;
})();

export const GitlabProjectModal = $root.GitlabProjectModal = (() => {

    /**
     * Properties of a GitlabProjectModal.
     * @exports IGitlabProjectModal
     * @interface IGitlabProjectModal
     * @property {number|Long|null} [id] GitlabProjectModal id
     * @property {string|null} [default_branch] GitlabProjectModal default_branch
     * @property {string|null} [name] GitlabProjectModal name
     * @property {number|Long|null} [gitlab_project_id] GitlabProjectModal gitlab_project_id
     * @property {boolean|null} [enabled] GitlabProjectModal enabled
     * @property {boolean|null} [global_enabled] GitlabProjectModal global_enabled
     * @property {string|null} [global_config] GitlabProjectModal global_config
     * @property {google.protobuf.ITimestamp|null} [created_at] GitlabProjectModal created_at
     * @property {google.protobuf.ITimestamp|null} [updated_at] GitlabProjectModal updated_at
     * @property {google.protobuf.ITimestamp|null} [deleted_at] GitlabProjectModal deleted_at
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
     * @member {number|Long} id
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
     * @member {number|Long} gitlab_project_id
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
     * @member {google.protobuf.ITimestamp|null|undefined} created_at
     * @memberof GitlabProjectModal
     * @instance
     */
    GitlabProjectModal.prototype.created_at = null;

    /**
     * GitlabProjectModal updated_at.
     * @member {google.protobuf.ITimestamp|null|undefined} updated_at
     * @memberof GitlabProjectModal
     * @instance
     */
    GitlabProjectModal.prototype.updated_at = null;

    /**
     * GitlabProjectModal deleted_at.
     * @member {google.protobuf.ITimestamp|null|undefined} deleted_at
     * @memberof GitlabProjectModal
     * @instance
     */
    GitlabProjectModal.prototype.deleted_at = null;

    return GitlabProjectModal;
})();

export const NamespaceModal = $root.NamespaceModal = (() => {

    /**
     * Properties of a NamespaceModal.
     * @exports INamespaceModal
     * @interface INamespaceModal
     * @property {number|Long|null} [id] NamespaceModal id
     * @property {string|null} [name] NamespaceModal name
     * @property {Array.<string>|null} [image_pull_secrets] NamespaceModal image_pull_secrets
     * @property {google.protobuf.ITimestamp|null} [created_at] NamespaceModal created_at
     * @property {google.protobuf.ITimestamp|null} [updated_at] NamespaceModal updated_at
     * @property {google.protobuf.ITimestamp|null} [deleted_at] NamespaceModal deleted_at
     * @property {Array.<IProjectModal>|null} [projects] NamespaceModal projects
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
     * @member {number|Long} id
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
     * @member {google.protobuf.ITimestamp|null|undefined} created_at
     * @memberof NamespaceModal
     * @instance
     */
    NamespaceModal.prototype.created_at = null;

    /**
     * NamespaceModal updated_at.
     * @member {google.protobuf.ITimestamp|null|undefined} updated_at
     * @memberof NamespaceModal
     * @instance
     */
    NamespaceModal.prototype.updated_at = null;

    /**
     * NamespaceModal deleted_at.
     * @member {google.protobuf.ITimestamp|null|undefined} deleted_at
     * @memberof NamespaceModal
     * @instance
     */
    NamespaceModal.prototype.deleted_at = null;

    /**
     * NamespaceModal projects.
     * @member {Array.<IProjectModal>} projects
     * @memberof NamespaceModal
     * @instance
     */
    NamespaceModal.prototype.projects = $util.emptyArray;

    return NamespaceModal;
})();

export const ProjectModal = $root.ProjectModal = (() => {

    /**
     * Properties of a ProjectModal.
     * @exports IProjectModal
     * @interface IProjectModal
     * @property {number|Long|null} [id] ProjectModal id
     * @property {string|null} [name] ProjectModal name
     * @property {number|Long|null} [gitlab_project_id] ProjectModal gitlab_project_id
     * @property {string|null} [gitlab_branch] ProjectModal gitlab_branch
     * @property {string|null} [gitlab_commit] ProjectModal gitlab_commit
     * @property {string|null} [config] ProjectModal config
     * @property {string|null} [override_values] ProjectModal override_values
     * @property {string|null} [docker_image] ProjectModal docker_image
     * @property {string|null} [pod_selectors] ProjectModal pod_selectors
     * @property {number|Long|null} [namespace_id] ProjectModal namespace_id
     * @property {boolean|null} [atomic] ProjectModal atomic
     * @property {google.protobuf.ITimestamp|null} [created_at] ProjectModal created_at
     * @property {google.protobuf.ITimestamp|null} [updated_at] ProjectModal updated_at
     * @property {google.protobuf.ITimestamp|null} [deleted_at] ProjectModal deleted_at
     * @property {INamespaceModal|null} [namespace] ProjectModal namespace
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
     * @member {number|Long} id
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
     * @member {number|Long} gitlab_project_id
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
     * @member {number|Long} namespace_id
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
     * @member {google.protobuf.ITimestamp|null|undefined} created_at
     * @memberof ProjectModal
     * @instance
     */
    ProjectModal.prototype.created_at = null;

    /**
     * ProjectModal updated_at.
     * @member {google.protobuf.ITimestamp|null|undefined} updated_at
     * @memberof ProjectModal
     * @instance
     */
    ProjectModal.prototype.updated_at = null;

    /**
     * ProjectModal deleted_at.
     * @member {google.protobuf.ITimestamp|null|undefined} deleted_at
     * @memberof ProjectModal
     * @instance
     */
    ProjectModal.prototype.deleted_at = null;

    /**
     * ProjectModal namespace.
     * @member {INamespaceModal|null|undefined} namespace
     * @memberof ProjectModal
     * @instance
     */
    ProjectModal.prototype.namespace = null;

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

            return Empty;
        })();

        protobuf.FileDescriptorSet = (function() {

            /**
             * Properties of a FileDescriptorSet.
             * @memberof google.protobuf
             * @interface IFileDescriptorSet
             * @property {Array.<google.protobuf.IFileDescriptorProto>|null} [file] FileDescriptorSet file
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
             * @member {Array.<google.protobuf.IFileDescriptorProto>} file
             * @memberof google.protobuf.FileDescriptorSet
             * @instance
             */
            FileDescriptorSet.prototype.file = $util.emptyArray;

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
             * @property {Array.<google.protobuf.IDescriptorProto>|null} [message_type] FileDescriptorProto message_type
             * @property {Array.<google.protobuf.IEnumDescriptorProto>|null} [enum_type] FileDescriptorProto enum_type
             * @property {Array.<google.protobuf.IServiceDescriptorProto>|null} [service] FileDescriptorProto service
             * @property {Array.<google.protobuf.IFieldDescriptorProto>|null} [extension] FileDescriptorProto extension
             * @property {google.protobuf.IFileOptions|null} [options] FileDescriptorProto options
             * @property {google.protobuf.ISourceCodeInfo|null} [source_code_info] FileDescriptorProto source_code_info
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
             * @member {Array.<google.protobuf.IDescriptorProto>} message_type
             * @memberof google.protobuf.FileDescriptorProto
             * @instance
             */
            FileDescriptorProto.prototype.message_type = $util.emptyArray;

            /**
             * FileDescriptorProto enum_type.
             * @member {Array.<google.protobuf.IEnumDescriptorProto>} enum_type
             * @memberof google.protobuf.FileDescriptorProto
             * @instance
             */
            FileDescriptorProto.prototype.enum_type = $util.emptyArray;

            /**
             * FileDescriptorProto service.
             * @member {Array.<google.protobuf.IServiceDescriptorProto>} service
             * @memberof google.protobuf.FileDescriptorProto
             * @instance
             */
            FileDescriptorProto.prototype.service = $util.emptyArray;

            /**
             * FileDescriptorProto extension.
             * @member {Array.<google.protobuf.IFieldDescriptorProto>} extension
             * @memberof google.protobuf.FileDescriptorProto
             * @instance
             */
            FileDescriptorProto.prototype.extension = $util.emptyArray;

            /**
             * FileDescriptorProto options.
             * @member {google.protobuf.IFileOptions|null|undefined} options
             * @memberof google.protobuf.FileDescriptorProto
             * @instance
             */
            FileDescriptorProto.prototype.options = null;

            /**
             * FileDescriptorProto source_code_info.
             * @member {google.protobuf.ISourceCodeInfo|null|undefined} source_code_info
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

            return FileDescriptorProto;
        })();

        protobuf.DescriptorProto = (function() {

            /**
             * Properties of a DescriptorProto.
             * @memberof google.protobuf
             * @interface IDescriptorProto
             * @property {string|null} [name] DescriptorProto name
             * @property {Array.<google.protobuf.IFieldDescriptorProto>|null} [field] DescriptorProto field
             * @property {Array.<google.protobuf.IFieldDescriptorProto>|null} [extension] DescriptorProto extension
             * @property {Array.<google.protobuf.IDescriptorProto>|null} [nested_type] DescriptorProto nested_type
             * @property {Array.<google.protobuf.IEnumDescriptorProto>|null} [enum_type] DescriptorProto enum_type
             * @property {Array.<google.protobuf.DescriptorProto.IExtensionRange>|null} [extension_range] DescriptorProto extension_range
             * @property {Array.<google.protobuf.IOneofDescriptorProto>|null} [oneof_decl] DescriptorProto oneof_decl
             * @property {google.protobuf.IMessageOptions|null} [options] DescriptorProto options
             * @property {Array.<google.protobuf.DescriptorProto.IReservedRange>|null} [reserved_range] DescriptorProto reserved_range
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
             * @member {Array.<google.protobuf.IFieldDescriptorProto>} field
             * @memberof google.protobuf.DescriptorProto
             * @instance
             */
            DescriptorProto.prototype.field = $util.emptyArray;

            /**
             * DescriptorProto extension.
             * @member {Array.<google.protobuf.IFieldDescriptorProto>} extension
             * @memberof google.protobuf.DescriptorProto
             * @instance
             */
            DescriptorProto.prototype.extension = $util.emptyArray;

            /**
             * DescriptorProto nested_type.
             * @member {Array.<google.protobuf.IDescriptorProto>} nested_type
             * @memberof google.protobuf.DescriptorProto
             * @instance
             */
            DescriptorProto.prototype.nested_type = $util.emptyArray;

            /**
             * DescriptorProto enum_type.
             * @member {Array.<google.protobuf.IEnumDescriptorProto>} enum_type
             * @memberof google.protobuf.DescriptorProto
             * @instance
             */
            DescriptorProto.prototype.enum_type = $util.emptyArray;

            /**
             * DescriptorProto extension_range.
             * @member {Array.<google.protobuf.DescriptorProto.IExtensionRange>} extension_range
             * @memberof google.protobuf.DescriptorProto
             * @instance
             */
            DescriptorProto.prototype.extension_range = $util.emptyArray;

            /**
             * DescriptorProto oneof_decl.
             * @member {Array.<google.protobuf.IOneofDescriptorProto>} oneof_decl
             * @memberof google.protobuf.DescriptorProto
             * @instance
             */
            DescriptorProto.prototype.oneof_decl = $util.emptyArray;

            /**
             * DescriptorProto options.
             * @member {google.protobuf.IMessageOptions|null|undefined} options
             * @memberof google.protobuf.DescriptorProto
             * @instance
             */
            DescriptorProto.prototype.options = null;

            /**
             * DescriptorProto reserved_range.
             * @member {Array.<google.protobuf.DescriptorProto.IReservedRange>} reserved_range
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
             * @property {google.protobuf.IFieldOptions|null} [options] FieldDescriptorProto options
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
             * @member {google.protobuf.IFieldOptions|null|undefined} options
             * @memberof google.protobuf.FieldDescriptorProto
             * @instance
             */
            FieldDescriptorProto.prototype.options = null;

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
             * @property {google.protobuf.IOneofOptions|null} [options] OneofDescriptorProto options
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
             * @member {google.protobuf.IOneofOptions|null|undefined} options
             * @memberof google.protobuf.OneofDescriptorProto
             * @instance
             */
            OneofDescriptorProto.prototype.options = null;

            return OneofDescriptorProto;
        })();

        protobuf.EnumDescriptorProto = (function() {

            /**
             * Properties of an EnumDescriptorProto.
             * @memberof google.protobuf
             * @interface IEnumDescriptorProto
             * @property {string|null} [name] EnumDescriptorProto name
             * @property {Array.<google.protobuf.IEnumValueDescriptorProto>|null} [value] EnumDescriptorProto value
             * @property {google.protobuf.IEnumOptions|null} [options] EnumDescriptorProto options
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
             * @member {Array.<google.protobuf.IEnumValueDescriptorProto>} value
             * @memberof google.protobuf.EnumDescriptorProto
             * @instance
             */
            EnumDescriptorProto.prototype.value = $util.emptyArray;

            /**
             * EnumDescriptorProto options.
             * @member {google.protobuf.IEnumOptions|null|undefined} options
             * @memberof google.protobuf.EnumDescriptorProto
             * @instance
             */
            EnumDescriptorProto.prototype.options = null;

            return EnumDescriptorProto;
        })();

        protobuf.EnumValueDescriptorProto = (function() {

            /**
             * Properties of an EnumValueDescriptorProto.
             * @memberof google.protobuf
             * @interface IEnumValueDescriptorProto
             * @property {string|null} [name] EnumValueDescriptorProto name
             * @property {number|null} [number] EnumValueDescriptorProto number
             * @property {google.protobuf.IEnumValueOptions|null} [options] EnumValueDescriptorProto options
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
             * @member {google.protobuf.IEnumValueOptions|null|undefined} options
             * @memberof google.protobuf.EnumValueDescriptorProto
             * @instance
             */
            EnumValueDescriptorProto.prototype.options = null;

            return EnumValueDescriptorProto;
        })();

        protobuf.ServiceDescriptorProto = (function() {

            /**
             * Properties of a ServiceDescriptorProto.
             * @memberof google.protobuf
             * @interface IServiceDescriptorProto
             * @property {string|null} [name] ServiceDescriptorProto name
             * @property {Array.<google.protobuf.IMethodDescriptorProto>|null} [method] ServiceDescriptorProto method
             * @property {google.protobuf.IServiceOptions|null} [options] ServiceDescriptorProto options
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
             * @member {Array.<google.protobuf.IMethodDescriptorProto>} method
             * @memberof google.protobuf.ServiceDescriptorProto
             * @instance
             */
            ServiceDescriptorProto.prototype.method = $util.emptyArray;

            /**
             * ServiceDescriptorProto options.
             * @member {google.protobuf.IServiceOptions|null|undefined} options
             * @memberof google.protobuf.ServiceDescriptorProto
             * @instance
             */
            ServiceDescriptorProto.prototype.options = null;

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
             * @property {google.protobuf.IMethodOptions|null} [options] MethodDescriptorProto options
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
             * @member {google.protobuf.IMethodOptions|null|undefined} options
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
             * @property {Array.<google.protobuf.IUninterpretedOption>|null} [uninterpreted_option] FileOptions uninterpreted_option
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
             * @member {Array.<google.protobuf.IUninterpretedOption>} uninterpreted_option
             * @memberof google.protobuf.FileOptions
             * @instance
             */
            FileOptions.prototype.uninterpreted_option = $util.emptyArray;

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
             * @property {Array.<google.protobuf.IUninterpretedOption>|null} [uninterpreted_option] MessageOptions uninterpreted_option
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
             * @member {Array.<google.protobuf.IUninterpretedOption>} uninterpreted_option
             * @memberof google.protobuf.MessageOptions
             * @instance
             */
            MessageOptions.prototype.uninterpreted_option = $util.emptyArray;

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
             * @property {Array.<google.protobuf.IUninterpretedOption>|null} [uninterpreted_option] FieldOptions uninterpreted_option
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
             * @member {Array.<google.protobuf.IUninterpretedOption>} uninterpreted_option
             * @memberof google.protobuf.FieldOptions
             * @instance
             */
            FieldOptions.prototype.uninterpreted_option = $util.emptyArray;

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
             * @property {Array.<google.protobuf.IUninterpretedOption>|null} [uninterpreted_option] OneofOptions uninterpreted_option
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
             * @member {Array.<google.protobuf.IUninterpretedOption>} uninterpreted_option
             * @memberof google.protobuf.OneofOptions
             * @instance
             */
            OneofOptions.prototype.uninterpreted_option = $util.emptyArray;

            return OneofOptions;
        })();

        protobuf.EnumOptions = (function() {

            /**
             * Properties of an EnumOptions.
             * @memberof google.protobuf
             * @interface IEnumOptions
             * @property {boolean|null} [allow_alias] EnumOptions allow_alias
             * @property {boolean|null} [deprecated] EnumOptions deprecated
             * @property {Array.<google.protobuf.IUninterpretedOption>|null} [uninterpreted_option] EnumOptions uninterpreted_option
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
             * @member {Array.<google.protobuf.IUninterpretedOption>} uninterpreted_option
             * @memberof google.protobuf.EnumOptions
             * @instance
             */
            EnumOptions.prototype.uninterpreted_option = $util.emptyArray;

            return EnumOptions;
        })();

        protobuf.EnumValueOptions = (function() {

            /**
             * Properties of an EnumValueOptions.
             * @memberof google.protobuf
             * @interface IEnumValueOptions
             * @property {boolean|null} [deprecated] EnumValueOptions deprecated
             * @property {Array.<google.protobuf.IUninterpretedOption>|null} [uninterpreted_option] EnumValueOptions uninterpreted_option
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
             * @member {Array.<google.protobuf.IUninterpretedOption>} uninterpreted_option
             * @memberof google.protobuf.EnumValueOptions
             * @instance
             */
            EnumValueOptions.prototype.uninterpreted_option = $util.emptyArray;

            return EnumValueOptions;
        })();

        protobuf.ServiceOptions = (function() {

            /**
             * Properties of a ServiceOptions.
             * @memberof google.protobuf
             * @interface IServiceOptions
             * @property {boolean|null} [deprecated] ServiceOptions deprecated
             * @property {Array.<google.protobuf.IUninterpretedOption>|null} [uninterpreted_option] ServiceOptions uninterpreted_option
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
             * @member {Array.<google.protobuf.IUninterpretedOption>} uninterpreted_option
             * @memberof google.protobuf.ServiceOptions
             * @instance
             */
            ServiceOptions.prototype.uninterpreted_option = $util.emptyArray;

            return ServiceOptions;
        })();

        protobuf.MethodOptions = (function() {

            /**
             * Properties of a MethodOptions.
             * @memberof google.protobuf
             * @interface IMethodOptions
             * @property {boolean|null} [deprecated] MethodOptions deprecated
             * @property {Array.<google.protobuf.IUninterpretedOption>|null} [uninterpreted_option] MethodOptions uninterpreted_option
             * @property {google.api.IHttpRule|null} [".google.api.http"] MethodOptions .google.api.http
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
             * @member {Array.<google.protobuf.IUninterpretedOption>} uninterpreted_option
             * @memberof google.protobuf.MethodOptions
             * @instance
             */
            MethodOptions.prototype.uninterpreted_option = $util.emptyArray;

            /**
             * MethodOptions .google.api.http.
             * @member {google.api.IHttpRule|null|undefined} .google.api.http
             * @memberof google.protobuf.MethodOptions
             * @instance
             */
            MethodOptions.prototype[".google.api.http"] = null;

            return MethodOptions;
        })();

        protobuf.UninterpretedOption = (function() {

            /**
             * Properties of an UninterpretedOption.
             * @memberof google.protobuf
             * @interface IUninterpretedOption
             * @property {Array.<google.protobuf.UninterpretedOption.INamePart>|null} [name] UninterpretedOption name
             * @property {string|null} [identifier_value] UninterpretedOption identifier_value
             * @property {number|Long|null} [positive_int_value] UninterpretedOption positive_int_value
             * @property {number|Long|null} [negative_int_value] UninterpretedOption negative_int_value
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
             * @member {Array.<google.protobuf.UninterpretedOption.INamePart>} name
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
             * @member {number|Long} positive_int_value
             * @memberof google.protobuf.UninterpretedOption
             * @instance
             */
            UninterpretedOption.prototype.positive_int_value = $util.Long ? $util.Long.fromBits(0,0,true) : 0;

            /**
             * UninterpretedOption negative_int_value.
             * @member {number|Long} negative_int_value
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

                return NamePart;
            })();

            return UninterpretedOption;
        })();

        protobuf.SourceCodeInfo = (function() {

            /**
             * Properties of a SourceCodeInfo.
             * @memberof google.protobuf
             * @interface ISourceCodeInfo
             * @property {Array.<google.protobuf.SourceCodeInfo.ILocation>|null} [location] SourceCodeInfo location
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
             * @member {Array.<google.protobuf.SourceCodeInfo.ILocation>} location
             * @memberof google.protobuf.SourceCodeInfo
             * @instance
             */
            SourceCodeInfo.prototype.location = $util.emptyArray;

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

                return Location;
            })();

            return SourceCodeInfo;
        })();

        protobuf.GeneratedCodeInfo = (function() {

            /**
             * Properties of a GeneratedCodeInfo.
             * @memberof google.protobuf
             * @interface IGeneratedCodeInfo
             * @property {Array.<google.protobuf.GeneratedCodeInfo.IAnnotation>|null} [annotation] GeneratedCodeInfo annotation
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
             * @member {Array.<google.protobuf.GeneratedCodeInfo.IAnnotation>} annotation
             * @memberof google.protobuf.GeneratedCodeInfo
             * @instance
             */
            GeneratedCodeInfo.prototype.annotation = $util.emptyArray;

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

                return Annotation;
            })();

            return GeneratedCodeInfo;
        })();

        protobuf.Timestamp = (function() {

            /**
             * Properties of a Timestamp.
             * @memberof google.protobuf
             * @interface ITimestamp
             * @property {number|Long|null} [seconds] Timestamp seconds
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
             * @member {number|Long} seconds
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
             * @property {Array.<google.api.IHttpRule>|null} [rules] Http rules
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
             * @member {Array.<google.api.IHttpRule>} rules
             * @memberof google.api.Http
             * @instance
             */
            Http.prototype.rules = $util.emptyArray;

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
             * @property {google.api.ICustomHttpPattern|null} [custom] HttpRule custom
             * @property {string|null} [selector] HttpRule selector
             * @property {string|null} [body] HttpRule body
             * @property {Array.<google.api.IHttpRule>|null} [additional_bindings] HttpRule additional_bindings
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
             * @member {google.api.ICustomHttpPattern|null|undefined} custom
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
             * @member {Array.<google.api.IHttpRule>} additional_bindings
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
     * @property {number|Long|null} [namespace_id] NamespaceID namespace_id
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
     * @member {number|Long} namespace_id
     * @memberof NamespaceID
     * @instance
     */
    NamespaceID.prototype.namespace_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    return NamespaceID;
})();

export const NamespaceResponse = $root.NamespaceResponse = (() => {

    /**
     * Properties of a NamespaceResponse.
     * @exports INamespaceResponse
     * @interface INamespaceResponse
     * @property {number|Long|null} [id] NamespaceResponse id
     * @property {string|null} [name] NamespaceResponse name
     * @property {Array.<string>|null} [image_pull_secrets] NamespaceResponse image_pull_secrets
     * @property {google.protobuf.ITimestamp|null} [created_at] NamespaceResponse created_at
     * @property {google.protobuf.ITimestamp|null} [updated_at] NamespaceResponse updated_at
     * @property {google.protobuf.ITimestamp|null} [deleted_at] NamespaceResponse deleted_at
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
     * @member {number|Long} id
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
     * @member {google.protobuf.ITimestamp|null|undefined} created_at
     * @memberof NamespaceResponse
     * @instance
     */
    NamespaceResponse.prototype.created_at = null;

    /**
     * NamespaceResponse updated_at.
     * @member {google.protobuf.ITimestamp|null|undefined} updated_at
     * @memberof NamespaceResponse
     * @instance
     */
    NamespaceResponse.prototype.updated_at = null;

    /**
     * NamespaceResponse deleted_at.
     * @member {google.protobuf.ITimestamp|null|undefined} deleted_at
     * @memberof NamespaceResponse
     * @instance
     */
    NamespaceResponse.prototype.deleted_at = null;

    return NamespaceResponse;
})();

export const NamespaceItem = $root.NamespaceItem = (() => {

    /**
     * Properties of a NamespaceItem.
     * @exports INamespaceItem
     * @interface INamespaceItem
     * @property {number|Long|null} [id] NamespaceItem id
     * @property {string|null} [name] NamespaceItem name
     * @property {google.protobuf.ITimestamp|null} [created_at] NamespaceItem created_at
     * @property {google.protobuf.ITimestamp|null} [updated_at] NamespaceItem updated_at
     * @property {Array.<NamespaceItem.ISimpleProjectItem>|null} [projects] NamespaceItem projects
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
     * @member {number|Long} id
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
     * @member {google.protobuf.ITimestamp|null|undefined} created_at
     * @memberof NamespaceItem
     * @instance
     */
    NamespaceItem.prototype.created_at = null;

    /**
     * NamespaceItem updated_at.
     * @member {google.protobuf.ITimestamp|null|undefined} updated_at
     * @memberof NamespaceItem
     * @instance
     */
    NamespaceItem.prototype.updated_at = null;

    /**
     * NamespaceItem projects.
     * @member {Array.<NamespaceItem.ISimpleProjectItem>} projects
     * @memberof NamespaceItem
     * @instance
     */
    NamespaceItem.prototype.projects = $util.emptyArray;

    NamespaceItem.SimpleProjectItem = (function() {

        /**
         * Properties of a SimpleProjectItem.
         * @memberof NamespaceItem
         * @interface ISimpleProjectItem
         * @property {number|Long|null} [id] SimpleProjectItem id
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
         * @member {number|Long} id
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

        return SimpleProjectItem;
    })();

    return NamespaceItem;
})();

export const NamespaceList = $root.NamespaceList = (() => {

    /**
     * Properties of a NamespaceList.
     * @exports INamespaceList
     * @interface INamespaceList
     * @property {Array.<INamespaceItem>|null} [data] NamespaceList data
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
     * @member {Array.<INamespaceItem>} data
     * @memberof NamespaceList
     * @instance
     */
    NamespaceList.prototype.data = $util.emptyArray;

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

    return NsStoreRequest;
})();

export const NsStoreResponse = $root.NsStoreResponse = (() => {

    /**
     * Properties of a NsStoreResponse.
     * @exports INsStoreResponse
     * @interface INsStoreResponse
     * @property {INamespaceResponse|null} [data] NsStoreResponse data
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
     * @member {INamespaceResponse|null|undefined} data
     * @memberof NsStoreResponse
     * @instance
     */
    NsStoreResponse.prototype.data = null;

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

    return CpuAndMemoryResponse;
})();

export const ServiceEndpointsResponse = $root.ServiceEndpointsResponse = (() => {

    /**
     * Properties of a ServiceEndpointsResponse.
     * @exports IServiceEndpointsResponse
     * @interface IServiceEndpointsResponse
     * @property {Array.<ServiceEndpointsResponse.Iitem>|null} [data] ServiceEndpointsResponse data
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
     * @member {Array.<ServiceEndpointsResponse.Iitem>} data
     * @memberof ServiceEndpointsResponse
     * @instance
     */
    ServiceEndpointsResponse.prototype.data = $util.emptyArray;

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

        return item;
    })();

    return ServiceEndpointsResponse;
})();

export const ServiceEndpointsRequest = $root.ServiceEndpointsRequest = (() => {

    /**
     * Properties of a ServiceEndpointsRequest.
     * @exports IServiceEndpointsRequest
     * @interface IServiceEndpointsRequest
     * @property {number|Long|null} [namespace_id] ServiceEndpointsRequest namespace_id
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
     * @member {number|Long} namespace_id
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
     * @param {google.protobuf.IEmpty} request Empty message or plain object
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
     * @param {google.protobuf.IEmpty} request Empty message or plain object
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
     * @param {INsStoreRequest} request NsStoreRequest message or plain object
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
     * @param {INsStoreRequest} request NsStoreRequest message or plain object
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
     * @param {INamespaceID} request NamespaceID message or plain object
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
     * @param {INamespaceID} request NamespaceID message or plain object
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
     * @param {IServiceEndpointsRequest} request ServiceEndpointsRequest message or plain object
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
     * @param {IServiceEndpointsRequest} request ServiceEndpointsRequest message or plain object
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
     * @param {INamespaceID} request NamespaceID message or plain object
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
     * @param {INamespaceID} request NamespaceID message or plain object
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
     * @param {IBackgroundRequest} request BackgroundRequest message or plain object
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
     * @param {IBackgroundRequest} request BackgroundRequest message or plain object
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
     * @property {number|Long|null} [namespace_id] ProjectDestroyRequest namespace_id
     * @property {number|Long|null} [project_id] ProjectDestroyRequest project_id
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
     * @member {number|Long} namespace_id
     * @memberof ProjectDestroyRequest
     * @instance
     */
    ProjectDestroyRequest.prototype.namespace_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * ProjectDestroyRequest project_id.
     * @member {number|Long} project_id
     * @memberof ProjectDestroyRequest
     * @instance
     */
    ProjectDestroyRequest.prototype.project_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    return ProjectDestroyRequest;
})();

export const ProjectShowRequest = $root.ProjectShowRequest = (() => {

    /**
     * Properties of a ProjectShowRequest.
     * @exports IProjectShowRequest
     * @interface IProjectShowRequest
     * @property {number|Long|null} [namespace_id] ProjectShowRequest namespace_id
     * @property {number|Long|null} [project_id] ProjectShowRequest project_id
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
     * @member {number|Long} namespace_id
     * @memberof ProjectShowRequest
     * @instance
     */
    ProjectShowRequest.prototype.namespace_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * ProjectShowRequest project_id.
     * @member {number|Long} project_id
     * @memberof ProjectShowRequest
     * @instance
     */
    ProjectShowRequest.prototype.project_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    return ProjectShowRequest;
})();

export const ProjectShowResponse = $root.ProjectShowResponse = (() => {

    /**
     * Properties of a ProjectShowResponse.
     * @exports IProjectShowResponse
     * @interface IProjectShowResponse
     * @property {number|Long|null} [id] ProjectShowResponse id
     * @property {string|null} [name] ProjectShowResponse name
     * @property {number|Long|null} [gitlab_project_id] ProjectShowResponse gitlab_project_id
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
     * @property {ProjectShowResponse.INamespace|null} [namespace] ProjectShowResponse namespace
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
     * @member {number|Long} id
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
     * @member {number|Long} gitlab_project_id
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
     * @member {ProjectShowResponse.INamespace|null|undefined} namespace
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

    ProjectShowResponse.Namespace = (function() {

        /**
         * Properties of a Namespace.
         * @memberof ProjectShowResponse
         * @interface INamespace
         * @property {number|Long|null} [id] Namespace id
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
         * @member {number|Long} id
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

        return Namespace;
    })();

    return ProjectShowResponse;
})();

export const AllPodContainersRequest = $root.AllPodContainersRequest = (() => {

    /**
     * Properties of an AllPodContainersRequest.
     * @exports IAllPodContainersRequest
     * @interface IAllPodContainersRequest
     * @property {number|Long|null} [namespace_id] AllPodContainersRequest namespace_id
     * @property {number|Long|null} [project_id] AllPodContainersRequest project_id
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
     * @member {number|Long} namespace_id
     * @memberof AllPodContainersRequest
     * @instance
     */
    AllPodContainersRequest.prototype.namespace_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * AllPodContainersRequest project_id.
     * @member {number|Long} project_id
     * @memberof AllPodContainersRequest
     * @instance
     */
    AllPodContainersRequest.prototype.project_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

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

    return PodLog;
})();

export const AllPodContainersResponse = $root.AllPodContainersResponse = (() => {

    /**
     * Properties of an AllPodContainersResponse.
     * @exports IAllPodContainersResponse
     * @interface IAllPodContainersResponse
     * @property {Array.<IPodLog>|null} [data] AllPodContainersResponse data
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
     * @member {Array.<IPodLog>} data
     * @memberof AllPodContainersResponse
     * @instance
     */
    AllPodContainersResponse.prototype.data = $util.emptyArray;

    return AllPodContainersResponse;
})();

export const PodContainerLogRequest = $root.PodContainerLogRequest = (() => {

    /**
     * Properties of a PodContainerLogRequest.
     * @exports IPodContainerLogRequest
     * @interface IPodContainerLogRequest
     * @property {number|Long|null} [namespace_id] PodContainerLogRequest namespace_id
     * @property {number|Long|null} [project_id] PodContainerLogRequest project_id
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
     * @member {number|Long} namespace_id
     * @memberof PodContainerLogRequest
     * @instance
     */
    PodContainerLogRequest.prototype.namespace_id = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

    /**
     * PodContainerLogRequest project_id.
     * @member {number|Long} project_id
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

    return PodContainerLogRequest;
})();

export const PodContainerLogResponse = $root.PodContainerLogResponse = (() => {

    /**
     * Properties of a PodContainerLogResponse.
     * @exports IPodContainerLogResponse
     * @interface IPodContainerLogResponse
     * @property {IPodLog|null} [data] PodContainerLogResponse data
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
     * @member {IPodLog|null|undefined} data
     * @memberof PodContainerLogResponse
     * @instance
     */
    PodContainerLogResponse.prototype.data = null;

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
     * @param {IProjectDestroyRequest} request ProjectDestroyRequest message or plain object
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
     * @param {IProjectDestroyRequest} request ProjectDestroyRequest message or plain object
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
     * @param {IProjectShowRequest} request ProjectShowRequest message or plain object
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
     * @param {IProjectShowRequest} request ProjectShowRequest message or plain object
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
     * @param {IIsPodRunningRequest} request IsPodRunningRequest message or plain object
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
     * @param {IIsPodRunningRequest} request IsPodRunningRequest message or plain object
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
     * @param {IAllPodContainersRequest} request AllPodContainersRequest message or plain object
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
     * @param {IAllPodContainersRequest} request AllPodContainersRequest message or plain object
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
     * @param {IPodContainerLogRequest} request PodContainerLogRequest message or plain object
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
     * @param {IPodContainerLogRequest} request PodContainerLogRequest message or plain object
     * @returns {Promise<PodContainerLogResponse>} Promise
     * @variation 2
     */

    return Project;
})();

export { $root as default };
