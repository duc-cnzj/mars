import React, { useState, useEffect } from "react";
import { bg } from "../api/background";
import pb from "../api/compiled";
import { Form, Button, Input } from "antd";
import { useAuth } from "../contexts/auth";
import { useHistory } from "react-router-dom";
import {
  GoogleOutlined,
  GithubOutlined,
  QqOutlined,
  WechatOutlined,
  PushpinFilled,
  PushpinOutlined,
} from "@ant-design/icons";
import { settings as settingsApi } from "../api/auth";
import { setState, isRandomBg, toggleRandomBg } from "../utils/token";

const Login: React.FC = () => {
  const [bgInfo, setBgInfo] = useState<pb.BackgroundResponse>();
  const [settings, setSettings] = useState<pb.SettingsResponse>();
  const [random, setRandom] = useState(isRandomBg());

  useEffect(() => {
    bg({ random: random }).then((res) => setBgInfo(res.data));
    settingsApi().then((res) => {
      setSettings(res.data);
    });
  }, []);

  const h = useHistory();
  const auth = useAuth();

  const renderOidcItem: (name: string) => React.ReactNode = (name: string) => {
    switch (name) {
      case "wechat":
        return (
          <div className="login__sso-icon-item">
            <WechatOutlined />
          </div>
        );
      case "qq":
        return (
          <div className="login__sso-icon-item">
            <QqOutlined />
          </div>
        );
      case "github":
        return (
          <div className="login__sso-icon-item">
            <GithubOutlined />
          </div>
        );
      case "google":
        return (
          <div className="login__sso-icon-item">
            <GoogleOutlined />
          </div>
        );
      default:
        return <div className="login__sso-item__name">{name}</div>;
    }
  };

  return (
    <div
      className="login__bg"
      style={bgInfo?.url ? { backgroundImage: "url(" + bgInfo.url + ")" } : {}}
    >
      <div
        className="login__pin"
        onClick={() => {
          setRandom(toggleRandomBg());
        }}
        title={random ? "固定壁纸" : "取消固定"}
      >
        {random ? <PushpinOutlined /> : <PushpinFilled />}
      </div>
      <div className="login__card">
        <div className="login__title">Mars Login</div>
        <div>
          <Form
            name="basic"
            onFinish={(values: any) => {
              console.log(values);
              auth.login(values.username, values.password, () => {
                h.push("/");
              });
            }}
            autoComplete="off"
          >
            <Form.Item
              name="username"
              rules={[{ required: true, message: "请输入用户名" }]}
            >
              <Input placeholder="用户名" />
            </Form.Item>

            <Form.Item
              name="password"
              rules={[{ required: true, message: "请输入密码" }]}
            >
              <Input.Password placeholder="密码" />
            </Form.Item>

            <Form.Item>
              <Button
                type="primary"
                htmlType="submit"
                style={{ width: "100%" }}
              >
                登录
              </Button>
            </Form.Item>

            <div className="login__sso-card">
              {settings?.items.map((item, index) => (
                <Form.Item key={index}>
                  <a
                    href="javascript(0);"
                    onClick={(e) => {
                      e.preventDefault();
                      setState(item.state || "");
                      window.location.href = item.url || "/login";
                    }}
                    className="login__sso-item"
                  >
                    {renderOidcItem(item.name || "")}
                  </a>
                </Form.Item>
              ))}
            </div>
          </Form>
        </div>
      </div>

      {bgInfo?.copyright ? (
        <div className="login__copyright">
          <div>{bgInfo.copyright}</div>
        </div>
      ) : (
        <></>
      )}
    </div>
  );
};

export default Login;
