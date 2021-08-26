import React, { useEffect, useCallback, memo } from "react";
import classnames from "classnames";
import { clusterInfo } from "../api/cluster";
import { useDispatch, useSelector } from "react-redux";
import { setClusterInfo } from "../store/actions";
import { selectClusterInfo } from "../store/reducers/cluster";

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
    <div
      className={classnames("dot", {
        "dot--health": isHealth(),
        "dot--bad": !isHealth(),
      })}
    ></div>
  );
};

export default memo(ClusterInfo);
