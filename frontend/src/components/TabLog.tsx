import React, { memo, useEffect, useState, useCallback, useMemo } from "react";
import { allPodContainers } from "../api/project";
import { Radio, Skeleton, Button, Tag, message } from "antd";
import pb from "../api/compiled";
import LazyLog from "../pkg/lazylog/components/LazyLog";
import { getToken } from "./../utils/token";

const ProjectContainerLogs: React.FC<{
  updatedAt: any;
  id: number;
  namespace: string;
}> = ({ id, namespace, updatedAt }) => {
  const [value, setValue] = useState<string>("");
  const [list, setList] = useState<pb.types.StateContainer[]>();

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
        setValue(first.pod + "|" + first.container);
      }
    });
  }, [setList, id, namespace, updatedAt, listContainer]);

  const [timestamp, setTimestamp] = useState(new Date().getTime());
  let [pod, container] = useMemo(() => (value as string).split("|"), [value]);

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
            key={item.pod + "|" + item.container}
            value={item.pod + "|" + item.container}
          >
            {item.container}
            {item.is_old && (
              <span style={{ marginLeft: 2, fontSize: 10, color: "#ef4444" }}>
                (old)
              </span>
            )}
            <Tag color="magenta" style={{ marginLeft: 10 }}>
              {item.pod}
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
        <Skeleton active loading={!(pod && container)}>
          <MyLogUtil
            namespace={namespace}
            freshTime={timestamp}
            pod={pod}
            container={container}
            onError={{
              praseJsonError: () => {
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
              },
              on404Error: (e) => {
                listContainer().then((res) => {
                  if (res.data.items.length > 0) {
                    let first = res.data.items[0];
                    setValue(first.pod + "|" + first.container);
                  }
                });
              },
            }}
          />
        </Skeleton>
      </div>
    </div>
  );
};

const MyLogUtil: React.FC<{
  namespace: string;
  pod: string;
  container: string;
  freshTime: number;
  onError?: {
    praseJsonError: (e: any) => any;
    on404Error: (e: any) => void;
  };
}> = ({ namespace, pod, container, onError, freshTime }) => {
  const getUrl = useCallback(() => {
    return `${process.env.REACT_APP_BASE_URL}/api/containers/namespaces/${namespace}/pods/${pod}/containers/${container}/stream_logs?timestamp=${freshTime}`;
  }, [namespace, pod, container, freshTime]);

  return (
    <LazyLog
      renderErrLineFunc={(e: any) => {
        return JSON.parse(e.body).error.message;
      }}
      fetchOptions={{ headers: { Authorization: getToken() } }}
      enableSearch
      selectableLines
      captureHotkeys
      formatPart={(text: string) => {
        let res = JSON.parse(text);
        if (res.error) {
          if (onError?.praseJsonError) {
            return onError?.praseJsonError(res.error);
          }
          return ""
        }
        return res.result.log;
      }}
      stream
      onError={(e: any) => {
        if (e.status === 404) {
          message.error(JSON.parse(e.body).error.message);
          onError?.on404Error(e);
        }
      }}
      follow={true}
      url={getUrl()}
    />
  );
};

export const LogUtil = memo(MyLogUtil);
export default memo(ProjectContainerLogs);
