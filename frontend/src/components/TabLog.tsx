import React, { memo, useEffect, useState, useCallback } from "react";
import { containerList } from "../api/project";
import { Radio, Skeleton, RadioChangeEvent, Tag } from "antd";
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

  const onChange = (e: RadioChangeEvent) => {
    setValue(e.target.value);
  };

  const getUrl = () => {
    let [pod, container] = (value as string).split("|");

    return `${process.env.REACT_APP_BASE_URL}/api/namespaces/${namespaceId}/projects/${id}/pods/${pod}/containers/${container}/stream_logs`;
  };

  return (
    <div style={{ display: "flex", flexDirection: "column", height: "100%" }}>
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
            formatPart={(text: string) => {
              return JSON.parse(text).result.data.log;
            }}
            stream
            onError={(e: any)=>{
              if (e.status === 404) {
                listContainer()
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
