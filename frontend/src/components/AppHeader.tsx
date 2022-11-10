import React, { memo } from "react";
import { Link } from "react-router-dom";
import ClusterInfo from "./ClusterInfo";
import { useWsReady } from "../contexts/useWebsocket";
import { UserOutlined } from "@ant-design/icons";
import { useAuth } from "../contexts/auth";
import { getToken, removeToken } from "../utils/token";
import { useHistory } from "react-router-dom";
import { Dropdown } from "antd";
import { copy } from "../utils/copy";
import {
  LogoutOutlined,
  SettingOutlined,
  ReadOutlined,
  KeyOutlined,
  NotificationOutlined,
} from "@ant-design/icons";
import { ItemType } from "rc-menu/lib/interface";

const AppHeader: React.FC = () => {
  const h = useHistory();
  const { user, isAdmin } = useAuth();

  let items: ItemType[] = [
    {
      label: (
        <a href="/docs/index.html" target="_blank">
          <ReadOutlined /> 接口文档
        </a>
      ),
      key: "0",
    },
  ];
  if (isAdmin()) {
    items = [
      ...items,
      {
        label: (
          <a
            href="javascript(0);"
            onClick={(e) => {
              e.preventDefault();
              h.push("/git_project_manager");
            }}
          >
            <SettingOutlined /> 项目配置
          </a>
        ),
        key: "1",
      },
      {
        label: (
          <a
            href="javascript(0);"
            onClick={(e) => {
              e.preventDefault();
              h.push("/events");
            }}
          >
            <NotificationOutlined /> 查看事件
          </a>
        ),
        key: "2",
      },
    ];
  }
  items = [
    ...items,
    {
      label: (
        <a
          href="javascript(0);"
          onClick={(e) => {
            e.preventDefault();
            copy(getToken());
          }}
        >
          <KeyOutlined /> 获取令牌
        </a>
      ),
      key: "3",
    },
    {
      type: "divider",
    },
    {
      label: (
        <a
          href="javascript(0);"
          onClick={(e) => {
            e.preventDefault();
            removeToken();
            if (user.logout_url) {
              window.location.href = user.logout_url;
            } else {
              h.push("/login");
            }
          }}
        >
          <LogoutOutlined />
          登出
        </a>
      ),
      key: "4",
    },
  ];

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
        {user && (
          <Dropdown menu={{ items }} trigger={["click"]}>
            <a
              href="javascript(0);"
              style={{ marginLeft: 20, color: "white" }}
              className="ant-dropdown-link"
              onClick={(e) => e.preventDefault()}
            >
              {user.avatar ? (
                <img
                  className="avatar"
                  style={{ borderRadius: "50%", width: 20, height: 20 }}
                  src={user.avatar}
                  alt="avatar"
                />
              ) : (
                <UserOutlined />
              )}

              <span style={{ fontSize: 12, marginLeft: 5 }}>{user.name}</span>
            </a>
          </Dropdown>
        )}
      </div>
    </div>
  );
};

export default memo(AppHeader);
