import React, { useEffect, useState, useCallback, memo, useRef } from "react";
import { FieldTimeOutlined } from "@ant-design/icons";

const TimeCost: React.FC<{ start: boolean; startAt?: number }> = ({ start, startAt }) => {
  const [startTime, setStartTime] = useState(0);
  const [now, setNow] = useState(0);
  const intervalRef = useRef<NodeJS.Timer>();

  const handleStart = useCallback(() => {
    setStartTime(startAt && startAt > 0 ? startAt : Date.now());
    setNow(Date.now());

    intervalRef.current && clearInterval(intervalRef.current);
    intervalRef.current = setInterval(() => {
      setNow(Date.now());
    }, 10);
  }, [startAt]);

  const handleStop = useCallback(() => {
    intervalRef.current && clearInterval(intervalRef.current);
  }, []);

  useEffect(() => {
    if (start) {
      handleStart();
    }

    return () => {
      console.log("clear");
      handleStop();
    };
  }, [start, handleStart, handleStop]);

  let secondsPassed = 0;
  if (startTime != null && now != null) {
    secondsPassed = (now - startTime) / 1000;
  }

  const getColor = useCallback((seconds: number): string => {
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
  }, []);

  return (
    <div style={{ paddingTop: 10, paddingBottom: 10 }}>
      <FieldTimeOutlined />
      &nbsp; 耗时：
      <span style={{ color: getColor(secondsPassed) }}>
        {secondsPassed.toFixed(1)}
      </span>{" "}
      s
    </div>
  );
};

export default memo(TimeCost);
