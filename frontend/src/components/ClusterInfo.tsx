import React, { useEffect, memo } from "react";
import { useDispatch, useSelector } from "react-redux";
import { setClusterInfo } from "../store/actions";
import { selectClusterInfo } from "../store/reducers/cluster";
import { Tooltip } from "antd";
import { css } from "@emotion/react";
import styled from "@emotion/styled";
import ajax from "../api/ajax";

const ClusterInfo: React.FC = () => {
  const dispatch = useDispatch();
  const info = useSelector(selectClusterInfo);

  useEffect(() => {
    ajax.GET("/api/cluster_info").then(({ data }) => {
      data && dispatch(setClusterInfo(data.item));
    });
  }, [dispatch]);

  return (
    <Tooltip
      placement="bottom"
      title={
        <div style={{ fontSize: 12 }}>
          <div>
            <span>状态: </span>
            {info.status}
          </div>
          <div>
            <span>cpu 剩余可分配量: </span>
            {info.freeRequestCpu}
          </div>
          <div>
            <span>memory 剩余可分配量: </span>
            {info.freeRequestMemory}
          </div>
          <div>
            <span>cpu 分配率: </span>
            {info.requestCpuRate}
          </div>
          <div>
            <span>memory 分配率: </span>
            {info.requestMemoryRate}
          </div>
          <div>
            <span>cpu 总量: </span>
            {info.totalCpu}
          </div>
          <div>
            <span>memory 总量: </span>
            {info.totalMemory}
          </div>
        </div>
      }
    >
      <DotDiv status={info.status} />
    </Tooltip>
  );
};

export default memo(ClusterInfo);

const Bad = css`
  background-color: #f87171;
  box-shadow: 0px 0px 5px #fca5a5;

  animation: fade 2s infinite;
`;

const NotGood = css`
  background-color: #f59e0b;
  box-shadow: 0px 0px 5px #fbbf24;

  animation: fade 3s infinite;
`;

const Health = css`
  background-color: #6ee7b7;
  box-shadow: 0px 0px 5px #a7f3d0;
`;

const DotDiv = styled.div<{ status: string }>`
  width: 10px;
  height: 10px;
  border-radius: 50%;
  ${({ status }) => status === "bad" && Bad}
  ${({ status }) => status === "health" && Health}
  ${({ status }) => status === "not good" && NotGood}

  @keyframes fade {
    from {
      opacity: 1;
    }
    50% {
      opacity: 0.5;
    }
    to {
      opacity: 1;
    }
  }
`;
