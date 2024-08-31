import React, { memo } from "react";
import { Link } from "react-router-dom";
import ClusterInfo from "./ClusterInfo";
import { useWsReady } from "../contexts/useWebsocket";
import { UserOutlined } from "@ant-design/icons";
import { useAuth } from "../contexts/auth";
import { removeToken } from "../utils/token";
import { useNavigate } from "react-router-dom";
import { Dropdown } from "antd";
import theme from "../styles/theme";
import { css } from "@emotion/css";
import {
  LogoutOutlined,
  SettingOutlined,
  ReadOutlined,
  KeyOutlined,
  NotificationOutlined,
} from "@ant-design/icons";
import { ItemType } from "rc-menu/lib/interface";
import useVersion from "../contexts/useVersion";
import logo from "../assets/marslogo.png";

const AppHeader: React.FC = () => {
  const h = useNavigate();
  const { user, isAdmin } = useAuth();
  const version = useVersion();

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
              h("/repos");
            }}
          >
            <SettingOutlined /> 仓库管理
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
              h("/events");
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
            h("/access_token_manager");
          }}
        >
          <KeyOutlined /> 令牌管理
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
            if (user.logoutUrl) {
              window.location.href = user.logoutUrl;
            } else {
              h("/login");
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
        className={css`
          color: ${theme.mainFontColor};
          font-size: 18px;
          display: flex;
          align-items: center;
        `}
        style={{ color: useWsReady() ? "white" : "red" }}
      >
        <img
          src={logo}
          style={{ width: 28, height: 28, marginRight: 5 }}
          alt="logo"
        />
        <div style={{ fontFamily: "dank mono" }}>Mars</div>
        <span style={{ fontSize: 10, marginLeft: "5px", marginTop: -9 }}>
          {version?.version}
        </span>
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
          <Dropdown
            overlayClassName="app-header-dropdown"
            menu={{ items }}
            trigger={["click"]}
          >
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
