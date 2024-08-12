import React, { useState, useCallback, memo, useEffect } from "react";
import { useStream } from "../pkg/fetchStream";
import { Row, Col } from "antd";
import Area from "./AreaChart";
import { getToken } from "../utils/token";

interface CpuM {
  cpu: number;
  name: number;
  time: string;
}
interface MemoryM {
  memory: number;
  name: number;
  time: string;
}

const defaultMetricsLength = 15;
const PodMetrics: React.FC<{
  namespace: string;
  pod: string;
  timestamp: any;
}> = ({ namespace, pod, timestamp }) => {
  const [cpuMetrics, setCpuMetrics] = useState<CpuM[]>([]);
  const [memoryMetrics, setMemoryMetrics] = useState<MemoryM[]>([]);

  const onNext = useCallback(
    async (res: any) => {
      const data = await res.text();
      let r = JSON.parse(data);
      if (r.result) {
        setCpuMetrics((l) => {
          if (l.length > defaultMetricsLength) {
            l.shift();
          }
          let ll = [
            ...l,
            {
              cpu: r.result.cpu,
              name: r.result.humanizeCpu,
              time: r.result.time,
            },
          ];
          return ll;
        });
        setMemoryMetrics((l) => {
          if (l.length > defaultMetricsLength) {
            l.shift();
          }
          let ll = [
            ...l,
            {
              memory: r.result.memory,
              name: r.result.humanizeMemory,
              time: r.result.time,
            },
          ];
          return ll;
        });
      } else {
        setCpuMetrics([]);
        setMemoryMetrics([]);
      }
    },
    [setCpuMetrics, setMemoryMetrics]
  );
  useEffect(() => {
    setCpuMetrics([]);
    setMemoryMetrics([]);
  }, [namespace, pod]);
  const onError = useCallback((e: any) => {
    console.log(e);
  }, []);
  let { close } = useStream(
    `${process.env.REACT_APP_BASE_URL}/api/metrics/namespace/${namespace}/pods/${pod}/stream?time=${timestamp}`,
    {
      onNext,
      onError,
      fetchParams: { headers: { Authorization: getToken() } },
    }
  );

  useEffect(() => {
    return () => {
      close();
    };
  }, [close]);

  return (
    <Row gutter={0} style={{ width: "100%", height: "100%", display: "flex" }}>
      <Col
        span={12}
        style={{ width: "100%", height: "100%", overflow: "hidden" }}
      >
        <div style={{ width: "100%", height: "100%" }}>
          <Area.CpuArea
            uniqueKey={`${namespace}-${pod}`}
            dataKey={"cpu"}
            data={cpuMetrics}
          />
        </div>
      </Col>
      <Col span={12} style={{ height: "100%", width: "100%" }}>
        <div style={{ width: "100%", height: "100%", overflow: "hidden" }}>
          <Area.MemoryArea
            uniqueKey={`${namespace}-${pod}`}
            dataKey={"memory"}
            data={memoryMetrics}
          />
        </div>
      </Col>
    </Row>
  );
};

export default memo(PodMetrics);
