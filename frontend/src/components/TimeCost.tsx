import React, { useEffect, useState, useCallback, memo, useRef } from "react";
import { FieldTimeOutlined } from "@ant-design/icons";

const TimeCost: React.FC<{ done: boolean }> = ({ done }) => {
  const [startTime, setStartTime] = useState<number | null>(null);
  const [now, setNow] = useState<number>(Date.now());
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const handleStart = useCallback(() => {
    setStartTime(Date.now());
    setNow(Date.now());

    if (intervalRef.current) {
      clearInterval(intervalRef.current);
    }
    intervalRef.current = setInterval(() => {
      setNow(Date.now());
    }, 10);
  }, []);

  const handleStop = useCallback(() => {
    if (intervalRef.current) {
      clearInterval(intervalRef.current);
    }
  }, []);

  useEffect(() => {
    if (!done) {
      handleStart();
    } else {
      handleStop();
    }

    return () => {
      handleStop();
    };
  }, [done, handleStart, handleStop]);

  let secondsPassed = 0;
  if (startTime != null) {
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
