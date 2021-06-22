import React, { memo, useEffect, useState } from "react";
import { containerList, containerLog, PodContainerItem } from "../api/project";
import { Radio, Skeleton, RadioChangeEvent, Tag } from "antd";
import SyntaxHighlighter from "react-syntax-highlighter";
import { xt256 } from "react-syntax-highlighter/dist/esm/styles/hljs";

const ProjectContainerLogs: React.FC<{
  updatedAt:string;
  id: number;
  namespaceId: number;
  autoRefresh: boolean;
}> = ({ id, namespaceId, autoRefresh, updatedAt }) => {
  const [value, setValue] = useState<string>();
  const [list, setList] = useState<PodContainerItem[]>();
  const [log, setLog] = useState<string>();

  useEffect(() => {
    containerList(namespaceId, id).then((res) => {
      setList(res.data.data);
      if (res.data.data.length > 0) {
        let first = res.data.data[0];
        setValue(first.pod_name + "|" + first.container_name);
        containerLog(
          namespaceId,
          id,
          first.pod_name,
          first.container_name
        ).then((res) => {
          setLog(res.data.data.log);
        });
      }
    });
  }, [setList, id, namespaceId, updatedAt]);

  const onChange = (e: RadioChangeEvent) => {
    setValue(e.target.value);
    let [pod, container] = (e.target.value as string).split("|");

    containerLog(namespaceId, id, pod, container).then((res) => {
      setLog(res.data.data.log);
    });
    console.log("on change", e.target);
  };

  useEffect(() => {
    let intervalId: any;
    if (autoRefresh) {
      let fn = () => {
        let [pod, container] = (value as string).split("|");

        if (pod && container) {
          containerLog(namespaceId, id, pod, container).then((res) => {
            setLog(res.data.data.log);
          });
        }

        console.log("setInterval");
      }
      fn()
      intervalId = setInterval(fn, 5000);
    }

    return () => {
      if (intervalId) {
        clearInterval(intervalId);
        console.log("clearInterval");
      }
    };
  }, [autoRefresh, id, namespaceId, value]);

  return (
    <>
      <Radio.Group
        onChange={onChange}
        value={value}
        style={{ marginBottom: 10 }}
      >
        {list?.map((item) => (
          <Radio
            key={item.pod_name + "|" + item.container_name}
            value={item.pod_name + "|" + item.container_name}
          >
            {item.container_name}
            <Tag color="magenta" style={{ marginLeft: 10 }}>
              {item.pod_name}
            </Tag>
          </Radio>
        ))}
      </Radio.Group>

      <div
        className="project-container-logs"
        style={{
          maxHeight: 800,
          overflow: "auto",
          fontFamily: '"Fira code", "Fira Mono", monospace',
          fontSize: 12,
        }}
      >
        {log ? (
          <SyntaxHighlighter
            wrapLongLines={true}
            language="shell"
            style={xt256}
          >
            {log}
          </SyntaxHighlighter>
        ) : (
          <Skeleton active />
        )}
      </div>
    </>
  );
};

export default memo(ProjectContainerLogs);
