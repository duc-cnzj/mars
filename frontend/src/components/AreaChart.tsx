import React, { memo } from "react";
import { AreaChart, Area, ResponsiveContainer } from "recharts";

const MyArea: React.FC<{
  data: any[];
  color?: string;
  ekey: string;
  dataKey: string;
  stroke: string;
  fill: {
    start: string;
    end: string;
  };
}> = ({ data, dataKey, stroke, fill, ekey }) => {
  return data.length > 1 ? (
    <ResponsiveContainer
      id={"ResponsiveContainer-" + ekey}
      height={"100%"}
      width="100%"
    >
      <AreaChart data={data}>
        <defs>
          <linearGradient id={`myGradient-${ekey}`} x1="0" x2="0" y1="0" y2="1">
            <stop offset="0%" stopColor={fill.start} />
            <stop offset="100%" stopColor={fill.end} />
          </linearGradient>
        </defs>
        <Area
          type="monotone"
          dataKey={dataKey}
          stroke={stroke}
          fill={`url(#myGradient-${ekey})`}
        />
      </AreaChart>
    </ResponsiveContainer>
  ) : (
    <></>
  );
};

const Memory: React.FC<{
  data: any[];
  dataKey: string;
  uniqueKey: string;
}> = ({ data, dataKey, uniqueKey }) => {
  return (
    <div style={{ position: "relative", width: "100%", height: "100%" }}>
      <div
        style={{
          position: "absolute",
          fontSize: 12,
          left: "50%",
          top: "50%",
          transform: "translate(-50%, -50%)",
        }}
      >
        {data.length > 1 ? "memory: " + data[data.length - 1].name : ""}
      </div>
      <MyArea
        dataKey={dataKey}
        stroke="#3B82F6"
        fill={{ end: "#DBEAFE", start: "#60A5FA" }}
        ekey={uniqueKey + "-memory"}
        data={data}
      />
    </div>
  );
};

const Cpu: React.FC<{
  data: any[];
  dataKey: string;
  uniqueKey: string;
}> = ({ data, dataKey, uniqueKey }) => {
  return (
    <div style={{ position: "relative", width: "100%", height: "100%" }}>
      <div
        style={{
          position: "absolute",
          fontSize: 12,
          left: "50%",
          top: "50%",
          transform: "translate(-50%, -50%)",
        }}
      >
        {data.length > 1 ? "cpu: " + data[data.length - 1].name : ""}
      </div>
      <MyArea
        dataKey={dataKey}
        stroke="#059669"
        fill={{ end: "#6EE7B7", start: "#10B981" }}
        ekey={uniqueKey + "-cpu"}
        data={data}
      />
    </div>
  );
};

const MarsArea = { CpuArea: memo(Cpu), MemoryArea: memo(Memory) };

export default MarsArea;
