import React, { memo, useEffect, useState, useCallback } from "react";
import { allPodContainers } from "../api/project";
import { Radio, Skeleton, Button, Tag } from "antd";
import pb from "../api/compiled";
import LazyLog from "../pkg/lazylog/components/LazyLog";
import { getToken } from "./../utils/token";

const ProjectContainerLogs: React.FC<{
  updatedAt: any;
  id: number;
  namespace: string;
}> = ({ id, namespace, updatedAt }) => {
  const [value, setValue] = useState<string>();
  const [list, setList] = useState<pb.ProjectPod[]>();

  const listContainer = useCallback(async () => {
    return allPodContainers({ project_id: id }).then((res) => {
      setList(res.data.items);
      return res;
    });
  }, [id]);

  useEffect(() => {
    listContainer().then((res) => {
      if (res.data.items.length > 0) {
        let first = res.data.items[0];
        setValue(first.pod_name + "|" + first.container_name);
      }
    });
  }, [setList, id, namespace, updatedAt, listContainer]);

  const [timestamp, setTimestamp] = useState(new Date().getTime());

  const getUrl = () => {
    let [pod, container] = (value as string).split("|");

    return `${process.env.REACT_APP_BASE_URL}/api/containers/namespaces/${namespace}/pods/${pod}/containers/${container}/stream_logs?timestamp=${timestamp}`;
  };

  const reloadLog = useCallback((e: any) => {
    setValue(e.target.value);
    setTimestamp(new Date().getTime());
  }, []);

  return (
    <div style={{ display: "flex", flexDirection: "column", height: "100%" }}>
      <Radio.Group value={value} style={{ marginBottom: 10 }}>
        {list?.map((item) => (
          <Radio
            onClick={reloadLog}
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
          fontFamily: '"Fira code", "Fira Mono", monospace',
          fontSize: 12,
          height: "100%",
        }}
      >
        {value ? (
          <LazyLog
            renderErrLineFunc={(e: any) => {
              return JSON.parse(e.body).error.message
            }}
            fetchOptions={{ headers: { Authorization: getToken() } }}
            enableSearch
            selectableLines
            captureHotkeys
            formatPart={(text: string) => {
              let res = JSON.parse(text);
              if (res.error) {
                return (
                  <span style={{ textAlign: "center" }}>
                    <Button
                      type="text"
                      style={{ color: "red", fontSize: 12 }}
                      size="small"
                      onClick={() => setTimestamp(new Date().getTime())}
                    >
                      点击重新加载
                    </Button>
                  </span>
                );
              }
              return res.result.log;
            }}
            stream
            onError={(e: any) => {
              if (e.status === 404) {
                listContainer().then((res) => {
                  if (res.data.items.length > 0) {
                    let first = res.data.items[0];
                    setValue(first.pod_name + "|" + first.container_name);
                  }
                });
              }
            }}
            follow={true}
            url={getUrl()}
          />
        ) : (
          <Skeleton active />
        )}
      </div>
    </div>
  );
};

export default memo(ProjectContainerLogs);
