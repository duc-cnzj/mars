import React, { useEffect, useState } from "react";
import { FieldTimeOutlined } from "@ant-design/icons";

const TimeCost: React.FC<{ start: boolean }> = ({ start }) => {
  const [seconds, setSeconds] = useState<number>(0.0);
  useEffect(() => {
    if (start) {
      setSeconds(0.0);
      let id = setInterval(() => {
        setSeconds((c) => (c += 0.1));
      }, 100);
      return () => {
        clearInterval(id);
        console.log("clearInterval(id)");
      };
    }
  }, [start, setSeconds]);

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

  return (
    <div style={{ paddingTop: 10, paddingBottom: 10 }}>
      <FieldTimeOutlined />
      &nbsp; 耗时：
      <span style={{ color: getColor(seconds) }}>{seconds.toFixed(1)}</span> s
    </div>
  );
};

export default TimeCost;
