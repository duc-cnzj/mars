import React, { memo, useCallback } from "react";
import { Timeline } from "antd";
import { useSelector } from "react-redux";
import { selectList } from "../store/reducers/createProject";

const LogOutput: React.FC<{ slug: string }> = ({ slug }) => {
  const list = useSelector(selectList);
  const getResultColor = useCallback((data: string) => {
    switch (data) {
      case "部署已取消":
        return "#F59E0B";
      case "部署失败":
        return "red";
      default:
        return "blue";
    }
  }, []);

  return (
    <Timeline
      pending={list[slug]?.isLoading ? "loading..." : false}
      reverse={true}
      style={{ paddingLeft: 2 }}
    >
      {list[slug]?.output.map((data, index) => (
        <Timeline.Item key={index} color={getResultColor(data)}>
          {data}
        </Timeline.Item>
      ))}
    </Timeline>
  );
};

export default memo(LogOutput);
