import React, { useState, useEffect, memo } from "react";
import { bg } from "../api/background";
import pb from "../api/compiled";
import { Form, Button, Input, Divider } from "antd";
import { useAuth } from "../contexts/auth";
import { useHistory } from "react-router-dom";
import { PushpinFilled, PushpinOutlined } from "@ant-design/icons";
import { settings as settingsApi } from "../api/auth";
import { setState, isRandomBg, toggleRandomBg } from "../utils/token";

const Login: React.FC = () => {
  const [bgInfo, setBgInfo] = useState<pb.picture.BackgroundResponse>();
  const [settings, setSettings] = useState<pb.auth.SettingsResponse>();
  const [random, setRandom] = useState(isRandomBg());

  useEffect(() => {
    bg({ random: isRandomBg() }).then((res) => setBgInfo(res.data));
    settingsApi().then((res) => {
      setSettings(res.data);
    });
  }, []);

  const h = useHistory();
  const auth = useAuth();

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
              <button
                className="login__button"
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
                  className="login__odic-btn"
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
      </div>

      {bgInfo?.copyright && (
        <div className="login__copyright">
          <div>{bgInfo.copyright}</div>
        </div>
      )}
    </div>
  );
};

export default memo(Login);
