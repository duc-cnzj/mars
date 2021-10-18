import React, { useEffect, useCallback, memo } from "react";
import classnames from "classnames";
import { clusterInfo } from "../api/cluster";
import { useDispatch, useSelector } from "react-redux";
import { setClusterInfo } from "../store/actions";
import { selectClusterInfo } from "../store/reducers/cluster";
import { Tooltip } from "antd";

const ClusterInfo: React.FC = () => {
  const dispatch = useDispatch();
  const info = useSelector(selectClusterInfo);

  useEffect(() => {
    clusterInfo().then((res) => {
      dispatch(setClusterInfo(res.data));
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
            {info.free_request_cpu}
          </div>
          <div>
            <span>memory 剩余可分配量: </span>
            {info.free_request_memory}
          </div>
          <div>
            <span>cpu 分配率: </span>
            {info.request_cpu_rate}
          </div>
          <div>
            <span>memory 分配率: </span>
            {info.request_memory_rate}
          </div>
          <div>
            <span>cpu 总量: </span>
            {info.total_cpu}
          </div>
          <div>
            <span>memory 总量: </span>
            {info.total_memory}
          </div>
        </div>
      }
    >
      <div
        className={classnames("dot", {
          "dot--health": info.status === "health",
          "dot--bad": info.status === "bad",
          "dot--not-good": info.status === "not good",
        })}
      ></div>
    </Tooltip>
  );
};

export default memo(ClusterInfo);
