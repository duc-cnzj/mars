import React from "react";
import { FieldTimeOutlined } from "@ant-design/icons";

const TimeCost: React.FC<{ seconds: string }> = ({ seconds }) => {
  const getColor = (seconds: number): string => {
    if (seconds < 5) {
      return "#6EE7B7";
    }
    if (seconds < 10) {
      return "#34D399";
    }
    if (seconds < 15) {
      return "#93C5FD";
    }
    if (seconds < 20) {
      return "#3B82F6";
    }
    if (seconds < 25) {
      return "#C4B5FD";
    }
    if (seconds < 30) {
      return "#8B5CF6";
    }
    if (seconds < 40) {
      return "#EF4444";
    }
    if (seconds < 80) {
      return "#EF4444";
    }

    return "#991B1B";
  };

  return (
    <div style={{ paddingTop: 10, paddingBottom: 10 }}>
      <FieldTimeOutlined />
      &nbsp; 耗时：
      <span style={{ color: getColor(Number(seconds)) }}>{seconds}</span> s
    </div>
  );
};

export default TimeCost;
