import React from "react";
import { FieldTimeOutlined } from "@ant-design/icons";

const TimeCost: React.FC<{ seconds: string }> = ({ seconds }) => {
  const getColor = (seconds: number): string => {
    if (seconds < 10) {
      return "#059669";
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

  return (
    <div style={{ paddingTop: 10, paddingBottom: 10 }}>
      <FieldTimeOutlined />
      &nbsp; 耗时：
      <span style={{ color: getColor(Number(seconds)) }}>{seconds ? seconds : "0.0"}</span> s
    </div>
  );
};

export default TimeCost;
