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
const PodMetrics: React.FC<{ namespace: string; pod: string; timestamp: any }> = ({
  namespace,
  pod,
  timestamp,
}) => {
  const [cpuMetrics, setCpuMetrics] = useState<CpuM[]>([]);
  const [memoryMetrics, setMemoryMetrics] = useState<MemoryM[]>([]);

  const onNext = useCallback(
    async (res: any) => {
      const data = await res.text();
      let r = JSON.parse(data);
      if (r.result) {
        setCpuMetrics((l) => {
          if (l.length > r.result.length) {
            l.shift();
          }
          let ll = [
            ...l,
            {
              cpu: r.result.cpu,
              name: r.result.humanize_cpu,
              time: r.result.time,
            },
          ];
          console.log(ll);
          return ll;
        });
        setMemoryMetrics((l) => {
          if (l.length > r.result.length) {
            l.shift();
          }
          let ll = [
            ...l,
            {
              memory: r.result.memory,
              name: r.result.humanize_memory,
              time: r.result.time,
            },
          ];
          console.log(ll);
          return ll;
        });
      } else {
        setCpuMetrics([])
        setMemoryMetrics([])
      }
    },
    [setCpuMetrics, setMemoryMetrics]
  );
  const onError = (e: any) => {
    console.log(e);
  };
  let { close } = useStream(
    `${process.env.REACT_APP_BASE_URL}/api/metrics/namespace/${namespace}/pods/${pod}?time=${timestamp}`,
    { onNext, onError, fetchParams: { headers: { Authorization: getToken() } } }
  );

  useEffect(() => {
    return () => {
      close();
    };
  }, [close]);

  return (
    <Row gutter={10} style={{ width: "100%", height: "100%", display: "flex" }}>
      <Col span={12} style={{ width: "100%", height: "100%" }}>
        <div style={{ width: "100%", height: "100%" }}>
          <Area.CpuArea
            uniqueKey={`${namespace}-${pod}`}
            dataKey={"cpu"}
            data={cpuMetrics}
          />
        </div>
      </Col>
      <Col span={12} style={{ height: "100%", width: "100%" }}>
        <div style={{ width: "100%", height: "100%" }}>
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
