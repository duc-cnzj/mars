import React, { useState, useEffect } from "react";
import { bg } from "../api/background";
import pb from "../api/compiled";
import { Form, Button, Input } from "antd";
import { useAuth } from "../contexts/auth";
import { useHistory } from "react-router-dom";
import { ArrowRightOutlined } from "@ant-design/icons";
import { settings as settingsApi } from "../api/auth";
import {setState} from '../utils/token'

const Login: React.FC = () => {
  const [bgInfo, setBgInfo] = useState<pb.BackgroundResponse>();
  const [settings, setSettings] = useState<pb.SettingsResponse>();
  useEffect(() => {
    bg({ random: true }).then((res) => setBgInfo(res.data));
    settingsApi().then((res) => {
      setSettings(res.data)
      setState(res.data.state)
    });
  }, []);

  const h = useHistory();
  const auth = useAuth();

  return (
    <div
      className="login__bg"
      style={{
        backgroundImage: "url(" + bgInfo?.url + ")",
      }}
    >
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

            {settings?.sso_enabled ? (
              <Form.Item>
                <a href={settings.url} className="login__use-sso">
                  单点登录
                  <ArrowRightOutlined />
                </a>
              </Form.Item>
            ) : (
              ""
            )}
          </Form>
        </div>
      </div>

      {bgInfo ? (
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
