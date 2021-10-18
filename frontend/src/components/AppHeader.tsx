import React, { memo } from "react";
import { Link } from "react-router-dom";
import ClusterInfo from "./ClusterInfo";
import { useWsReady } from "../contexts/useWebsocket";
import { QuestionCircleOutlined } from "@ant-design/icons";

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
      <div
        style={{
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
        }}
      >
        <ClusterInfo />
        <a href="/docs/index.html" target="_blank">
          <QuestionCircleOutlined
            style={{ borderRadius: "50%", background: "white", marginLeft: 10 }}
          />
        </a>
      </div>
    </div>
  );
};

export default memo(AppHeader);
