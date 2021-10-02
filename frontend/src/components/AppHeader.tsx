import React, { memo } from "react";
import { Link } from "react-router-dom";
import ClusterInfo from "./ClusterInfo";
import { useWsReady } from "../contexts/useWebsocket";

const AppHeader: React.FC = () => {
  return (
    <div
      style={{
        display: "flex",
        alignItems: "center",
        justifyContent: "space-between",
      }}
    >
      <Link
        to="/"
        className="app-title"
        style={{ color: useWsReady() ? "white" : "red" }}
      >
        Mars
      </Link>
      <ClusterInfo />
    </div>
  );
};

export default memo(AppHeader);
