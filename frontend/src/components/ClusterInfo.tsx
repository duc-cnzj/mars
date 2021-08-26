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
      dispatch(setClusterInfo(res.data.data));
    });
  }, [dispatch]);

  const isHealth = useCallback(() => {
    return info.status === "health";
  }, [info.status]);

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
            <span>cpu 剩余: </span>
            {info.free_cpu}
          </div>
          <div>
            <span>memory 剩余: </span>
            {info.free_memory}
          </div>
          <div>
            <span>cpu 使用率: </span>
            {info.usage_cpu_rate}
          </div>
          <div>
            <span>memory 使用率: </span>
            {info.usage_memory_rate}
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
          "dot--health": isHealth(),
          "dot--bad": !isHealth(),
        })}
      ></div>
    </Tooltip>
  );
};

export default memo(ClusterInfo);
