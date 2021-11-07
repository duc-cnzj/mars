import React, { memo, useEffect, useState, useCallback } from "react";
import { containerList, containerLog } from "../api/project";
import { Radio, Skeleton, RadioChangeEvent, Tag, message, Affix } from "antd";
import { Prism as SyntaxHighlighter } from "react-syntax-highlighter";
import { dracula } from "react-syntax-highlighter/dist/esm/styles/prism";
import AutoScroll from "./AutoScroll";
import pb from "../api/compiled";

const ProjectContainerLogs: React.FC<{
  updatedAt: any;
  id: number;
  namespaceId: number;
  autoRefresh: boolean;
}> = ({ id, namespaceId, autoRefresh, updatedAt }) => {
  const [value, setValue] = useState<string>();
  const [list, setList] = useState<pb.PodLog[]>();
  const [log, setLog] = useState<string>();

  const listContainer = useCallback(async () => {
    return containerList({ namespace_id: namespaceId, project_id: id }).then(
      (res) => {
        setList(res.data.data);
        return res;
      }
    );
  }, [namespaceId, id]);

  useEffect(() => {
    listContainer().then((res) => {
      if (res.data.data.length > 0) {
        let first = res.data.data[0];
        setValue(first.pod_name + "|" + first.container_name);
        containerLog({
          namespace_id: namespaceId,
          pod: first.pod_name,
          container: first.container_name,
          project_id: id,
        })
          .then(({ data: { data } }) => {
            let log: string = "暂无日志";
            if (data.log) {
              log = data.log;
            }

            setLog(log);
          })
          .catch((e) => {
            message.error(e.response.data.message);
          });
      }
    });
  }, [setList, id, namespaceId, updatedAt, listContainer]);

  const onChange = (e: RadioChangeEvent) => {
    setValue(e.target.value);
    let [pod, container] = (e.target.value as string).split("|");
    containerLog({
      namespace_id: namespaceId,
      project_id: id,
      pod: pod,
      container: container,
    })
      .then((res) => {
        setLog(res.data.data.log);
      })
      .catch((e) => {
        message.error(e.response.data.message);
        listContainer();
      });
    console.log("on change", e.target);
  };

  useEffect(() => {
    let intervalId: any;
    if (autoRefresh) {
      let fn = () => {
        let [pod, container] = (value as string).split("|");

        if (pod && container) {
          containerLog({
            namespace_id: namespaceId,
            project_id: id,
            pod: pod,
            container: container,
          }).then((res) => {
            setLog(res.data.data.log);
          });
        }

        console.log("setInterval");
      };
      fn();
      intervalId = setInterval(fn, 2000);
    }

    return () => {
      if (intervalId) {
        clearInterval(intervalId);
        console.log("clearInterval");
      }
    };
  }, [autoRefresh, id, namespaceId, value]);
  const [container, setContainer] = useState<HTMLDivElement | null>(null);

  return (
    <div
      ref={setContainer}
      style={{
        height: "100%",
        overflowY: "auto",
        display: "flex",
        flexDirection: "column",
      }}
    >
      <Affix target={() => container}>
        <div style={{ width: "100%", background: "white", paddingBottom: 10 }}>
          <Radio.Group onChange={onChange} value={value}>
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
        </div>
      </Affix>
      <div
        className="project-container-logs"
        style={{
          fontFamily: '"Fira code", "Fira Mono", monospace',
          fontSize: 12,
        }}
      >
        {log ? (
          <AutoScroll height={600} className="auto-scroll">
            <SyntaxHighlighter
              wrapLongLines={false}
              showLineNumbers
              language="vim"
              style={dracula}
            >
              {log}
            </SyntaxHighlighter>
          </AutoScroll>
        ) : (
          <Skeleton active />
        )}
      </div>
    </div>
  );
};

export default memo(ProjectContainerLogs);
