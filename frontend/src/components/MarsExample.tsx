import React, { memo } from "react";
import { xonokai } from "react-syntax-highlighter/dist/esm/styles/prism";
import { Popover } from "antd";
import { CopyOutlined, QuestionCircleOutlined } from "@ant-design/icons";
import  CopyToClipboard  from "./CopyToClipboard";
import { PrismLight as SyntaxHighlighter } from "react-syntax-highlighter";
import pyaml from 'react-syntax-highlighter/dist/esm/languages/prism/yaml';

SyntaxHighlighter.registerLanguage('yaml', pyaml);

const MarsExample: React.FC = () => {
  const example = `# 项目默认的配置文件(可选)
config_file: config.yaml
# 默认配置, 必须用 '|', 全局配置文件，如果没有设置 config_file 则使用这个
config_file_values: |
  env: dev
  port: 8000
# 配置文件的类型(如果有config_file，必填)
config_file_type: yaml
# config_field 对应到 helm values.yaml 中的哪个字段(如果有config_file，必填)
# 可以使用 '->' 指向下一级, 比如：'config->app_name'， 会变成
# config:
#   app_name: xxxx
config_field: conf
# charts 文件在项目中存放的目录(必填), 也可以是别的项目的文件，格式为 "pid|branch|path"
local_chart_path: charts
# 是不是单字段的配置(如果有config_file，必填)
is_simple_env: false
# 若配置则只会显示配置的分支, 默认 "*"(可选)
branches:
  - dev
  - master
elements:
  - path: replicaCount
    type: 1
    default: "2" # 必须是字符串
    description: "描述"
  - path: web->enabled
    type: 5
    default: "true" # 必须是字符串
    description: "开启web服务"
# values_yaml 和 helm 的 values.yaml 用法一模一样，但是可以使用变量
# 目前支持的变量有，使用 \`<>\` 作为 Delim，避免和内置模板语法冲突
# \`<.ImagePullSecrets>\` \`<.Branch>\` \`<.Commit>\` \`<.Pipeline>\` \`<.ClusterIssuer>\`
# \`<.Host1>...<.Host10>\` \`<.TlsSecret1>...<.TlsSecret10>\`
values_yaml: |
  # Default values for charts.
  # This is a YAML-formatted file.
  # Declare variables to be passed into your templates.
  
  replicaCount: 1
  
  image:
    repository: xxx
    pullPolicy: IfNotPresent
    # Overrides the image tag whose default is the chart appVersion.
    tag: "<.Branch>-<.Pipeline>"
  
  imagePullSecrets: []
  nameOverride: ""
  fullnameOverride: ""

  ingress:
    enabled: false
    className: ""
    annotations: 
      kubernetes.io/ingress.class: nginx
      kubernetes.io/tls-acme: "true"
      cert-manager.io/cluster-issuer: "<.ClusterIssuer>"
    hosts:
      - host: <.Host1>
        paths:
          - path: /
            pathType: Prefix
    tls: 
      - secretName: <.TlsSecret1>
        hosts:
          - <.Host1>`;

  return (
    <Popover
      overlayInnerStyle={{ maxHeight: 600, overflowY: "scroll" }}
      placement="bottomLeft"
      title={
        <div>
          example
          <CopyToClipboard
            text={example}
            successText="已复制！"
          >
            <CopyOutlined />
          </CopyToClipboard>
        </div>
      }
      content={
        <div className="mars-example">
          <SyntaxHighlighter
            language="yaml"
            style={xonokai}
            customStyle={{
              lineHeight: 1.2,
              padding: "10px",
              fontFamily: '"Fira code", "Fira Mono", monospace',
              fontSize: 10,
            }}
          >
            {example}
          </SyntaxHighlighter>
        </div>
      }
    >
      <QuestionCircleOutlined />
    </Popover>
  );
};

export default memo(MarsExample);
