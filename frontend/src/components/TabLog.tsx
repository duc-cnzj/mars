import React, { memo, useEffect, useState, useCallback } from "react";
import { containerList } from "../api/project";
import { Radio, Skeleton, Button, Tag } from "antd";
import pb from "../api/compiled";
import LazyLog from "../pkg/lazylog/components/LazyLog";
import { getToken } from "./../utils/token";

const ProjectContainerLogs: React.FC<{
  updatedAt: any;
  id: number;
  namespaceId: number;
}> = ({ id, namespaceId, updatedAt }) => {
  const [value, setValue] = useState<string>();
  const [list, setList] = useState<pb.PodLog[]>();

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
      }
    });
  }, [setList, id, namespaceId, updatedAt, listContainer]);

  const [timestamp, setTimestamp] = useState(new Date().getTime());

  const getUrl = () => {
    let [pod, container] = (value as string).split("|");

    return `${process.env.REACT_APP_BASE_URL}/api/namespaces/${namespaceId}/projects/${id}/pods/${pod}/containers/${container}/stream_logs?timestamp=${timestamp}`;
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
              return res.result.data.log;
            }}
            stream
            onError={(e: any) => {
              console.log(e);
              if (e.status === 404) {
                listContainer().then((res) => {
                  if (res.data.data.length > 0) {
                    let first = res.data.data[0];
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
