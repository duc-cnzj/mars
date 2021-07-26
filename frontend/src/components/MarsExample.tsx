import React, { memo } from "react";
import { shadesOfPurple } from "react-syntax-highlighter/dist/esm/styles/hljs";
import SyntaxHighlighter from "react-syntax-highlighter";
import { message, Popover } from "antd";
import { CopyOutlined, QuestionCircleOutlined } from "@ant-design/icons";
import { CopyToClipboard } from "react-copy-to-clipboard";

const MarsExample: React.FC = () => {
  const example = `# 项目默认的配置文件(可选)
config_file: config.yaml
# 配置文件的类型(如果有config_file，必填)
config_file_type: yaml
# config_file 对应到 helm values.yaml 中的哪个字段(如果有config_file，必填)
config_field: conf
# 镜像仓库(必填)
docker_repository: nginx
# tag 可以使用的变量有 {{.Commit}} {{.Branch}} {{.Pipeline}}(必填)
docker_tag_format: "{{.Branch}}-{{.Pipeline}}"
# charts 文件在项目中存放的目录(必填), 也可以是别的项目的文件，格式为 "pid|branch|path"
local_chart_path: charts
# 是不是单字段的配置(如果有config_file，必填)
is_simple_env: false
# values.yaml 会合并其他配置(可选)
default_values:
  service:
    type: ClusterIP
  ingess:
    enabled: false
# 若配置则只会显示配置的分支, 默认 "*"(可选)
branches:
  - dev
  - master
# 如果默认的ingress 规则不符合，你可以通过这个重写
# 可用变量 {{Host1}} {{TlsSecret1}} {{Host2}} {{TlsSecret2}} {{Host3}} {{TlsSecret3}} ... {{Host10}} {{TlsSecret10}}
ingress_overwrite_values:
  - ingress.hosts.hostone={{.Host1}}
  - ingress.hosts.hosttwo={{.Host2}}
  - ingress.tls[0].hosts[0]={{.Host1}}
  - ingress.tls[0].secretName={{.TlsSecret1}}
  - ingress.tls[1].hosts[0]={{.Host2}}
  - ingress.tls[1].secretName={{.TlsSecret2}}`;

  return (
    <Popover
      placement="bottomLeft"
      title={
        <div>
          example
          <CopyToClipboard
            text={example}
            onCopy={() => message.success("已复制！")}
          >
            <CopyOutlined />
          </CopyToClipboard>
        </div>
      }
      content={
        <div className="mars-example">
          <SyntaxHighlighter
            language="yaml"
            style={shadesOfPurple}
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
