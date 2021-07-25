import React from "react";
import { FieldTimeOutlined } from "@ant-design/icons";

const TimeCost: React.FC<{ seconds: string }> = ({ seconds: s }) => {
  const getColor = (seconds: number): string => {
    if (seconds < 10) {
      return "#6EE7B7";
    }
    if (seconds < 30) {
      return "#FCD34D";
    }
    if (seconds < 50) {
      return "#EC4899";
    }
    if (seconds < 70) {
      return "#F87171";
    }

    return "#DC2626";
  };

  let seconds: number = Number(s ? s : "0.0");

  return (
    <div style={{ paddingTop: 10, paddingBottom: 10 }}>
      <FieldTimeOutlined />
      &nbsp; 耗时：
      <span style={{ color: getColor(seconds) }}>{seconds}</span> s
    </div>
  );
};

export default TimeCost;
