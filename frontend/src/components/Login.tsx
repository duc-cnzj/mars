import React, { useState, useEffect, memo } from "react";
import { Form, Button, Input, Divider } from "antd";
import { useAuth } from "../contexts/auth";
import { useLocation, useNavigate } from "react-router-dom";
import { PushpinFilled, PushpinOutlined } from "@ant-design/icons";
import { setState, isRandomBg, toggleRandomBg } from "../utils/token";
import { css } from "@emotion/css";
import theme from "../styles/theme";
import styled from "@emotion/styled";
import ajax from "../api/ajax";
import { components } from "../api/schema";

const Login: React.FC = () => {
  const [bgInfo, setBgInfo] =
    useState<components["schemas"]["picture.BackgroundResponse"]>();
  const [settings, setSettings] =
    useState<components["schemas"]["auth.SettingsResponse"]>();
  const [random, setRandom] = useState(isRandomBg());

  useEffect(() => {
    ajax
      .GET("/api/picture/background", {
        params: { query: { random: isRandomBg() } },
      })
      .then((res) => setBgInfo(res.data));
    ajax.GET("/api/auth/settings").then(({ data }) => {
      data && setSettings(data);
    });
  }, []);

  const h = useNavigate();
  const auth = useAuth();
  const location = useLocation();

  return (
    <div
      className={css`
        width: 100vw;
        height: 100vh;
        display: flex;
        justify-content: center;
        align-items: center;
        position: relative;
        background-position: center center;
        background-size: cover;
      `}
      style={bgInfo?.url ? { backgroundImage: "url(" + bgInfo.url + ")" } : {}}
    >
      <div
        className={css`
          position: absolute;
          top: 15px;
          right: 15px;
          font-size: 16px;
          opacity: 0.5;
          transition: 0.5s;
          &:hover {
            cursor: pointer;
            opacity: 1;
          }
        `}
        onClick={() => {
          setRandom(toggleRandomBg());
        }}
        title={random ? "固定壁纸" : "取消固定"}
      >
        {random ? <PushpinOutlined /> : <PushpinFilled />}
      </div>
      <LoginCard>
        <LoginTitle>Mars Login</LoginTitle>
        <div>
          <Form
            name="basic"
            onFinish={(values: any) => {
              auth.login(values.username, values.password, () => {
                let to = "/";
                if (location.state && location.state.from.search) {
                  to += location.state?.from.search;
                }
                h(to);
              });
            }}
            autoComplete="off"
          >
            <Form.Item
              name="username"
              rules={[{ required: true, message: "请输入用户名" }]}
            >
              <Input autoFocus placeholder="用户名" />
            </Form.Item>

            <Form.Item
              name="password"
              rules={[{ required: true, message: "请输入密码" }]}
            >
              <Input.Password placeholder="密码" />
            </Form.Item>

            <Form.Item>
              <button
                className={css`
                  line-height: 1.5715;
                  position: relative;
                  display: inline-block;
                  cursor: pointer;
                  font-weight: 400;
                  white-space: nowrap;
                  text-align: center;
                  -webkit-user-select: none;
                  user-select: none;
                  touch-action: manipulation;
                  height: 32px;
                  padding: 4px 15px;
                  font-size: 14px;

                  border: none;
                  background-color: ${theme.mainColor};
                  transition: 0.5s;
                  color: white;
                  border-radius: 2px;

                  &:hover {
                    background-color: ${theme.deepColor};
                    color: #fff;
                    text-decoration: none;
                  }
                `}
                type="submit"
                style={{ width: "100%" }}
              >
                登录
              </button>
            </Form.Item>
            {settings?.items && settings?.items.length > 0 && (
              <Divider style={{ fontWeight: "normal", fontSize: 12 }}>
                或者
              </Divider>
            )}

            <div style={{ display: "flex", flexDirection: "column" }}>
              {settings?.items.map((item) => (
                <Button
                  type="primary"
                  key={item.name}
                  className={css`
                    background-color: ${theme.deepColor};
                    border: none;
                    color: white;
                  `}
                  href="javascript(0);"
                  style={{ marginBottom: 10 }}
                  onClick={(e) => {
                    e.preventDefault();
                    setState(item.state || "");
                    window.location.href = item.url || "/login";
                  }}
                >
                  <span
                    style={{
                      textTransform: "uppercase",
                      fontFamily: "monospace",
                    }}
                  >
                    {item.name}
                  </span>
                </Button>
              ))}
            </div>
          </Form>
        </div>
      </LoginCard>

      {bgInfo?.copyright && (
        <div
          className={css`
            font-size: 15px;
            font-weight: lighter;
            color: white;
            position: absolute;
            bottom: 20px;
            margin: 0 auto;
            padding: 3px 15px;
            backdrop-filter: blur(5px);
            border-radius: 5px;
            @media (min-width: 640px) {
              font-size: 20px;
              bottom: 50px;
              right: 50px;
            }
          `}
        >
          <div>{bgInfo.copyright}</div>
        </div>
      )}
    </div>
  );
};

export default memo(Login);

const LoginTitle = styled.div`
  font-family: "Comic Sans MS";
  text-align: center;
  background-image: linear-gradient(
    to bottom,
    ${theme.deepColor},
    ${theme.titleSecondColor}
  );
  -webkit-background-clip: text;
  color: transparent;
  margin: 10px 0 20px;
  font-size: 50px;
  line-height: 1.5;
`;

const LoginCard = styled.div`
  opacity: 0.8;
  margin: 20px;
  padding: 10px 50px;
  width: 400px;
  border-radius: 8px;
  transition: box-shadow 0.5s;
  &:hover {
    opacity: 1;
    backdrop-filter: blur(8px);
    box-shadow: 0px 0px 10px ${theme.titleSecondColor};
    ${LoginTitle} {
      transition: 0.5s;
      text-shadow: 0 0 3px ${theme.titleSecondColor};
    }
  }
  @media (min-width: 640px) {
    margin: 0;
  }
`;
