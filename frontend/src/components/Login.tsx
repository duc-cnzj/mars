import React, { useState, useEffect, memo } from "react";
import { Form, Input } from "antd";
import { useAuth } from "../contexts/auth";
import { useLocation, useNavigate } from "react-router-dom";
import { PushpinFilled, PushpinOutlined } from "@ant-design/icons";
import { setState, isRandomBg, toggleRandomBg } from "../utils/token";
import theme from "../styles/theme";
import styled from "@emotion/styled";
import ajax from "../api/ajax";
import { components } from "../api/schema";
import marsLogo from "../assets/marslogo.png";

const LOGIN_TYPE_KEY = "mars_login_type";

const getSavedLoginType = (): "sso" | "password" => {
  const saved = localStorage.getItem(LOGIN_TYPE_KEY);
  if (saved === "password" || saved === "sso") {
    return saved;
  }
  return "sso";
};

const saveLoginType = (type: "sso" | "password") => {
  localStorage.setItem(LOGIN_TYPE_KEY, type);
};

const Login: React.FC = () => {
  const [bgInfo, setBgInfo] =
    useState<components["schemas"]["picture.BackgroundResponse"]>();
  const [settings, setSettings] =
    useState<components["schemas"]["auth.SettingsResponse"]>();
  const [random, setRandom] = useState(isRandomBg());
  const [loginType, setLoginType] = useState<"sso" | "password">("sso");

  useEffect(() => {
    ajax
      .GET("/api/picture/background", {
        params: { query: { random: isRandomBg() } },
      })
      .then((res) => setBgInfo(res.data));
    ajax.GET("/api/auth/settings").then(({ data }) => {
      data && setSettings(data);
      // 如果没有 SSO 选项，默认切换到密码登录
      if (!data?.items?.length) {
        setLoginType("password");
      } else {
        // 有 SSO 选项时，读取用户偏好
        setLoginType(getSavedLoginType());
      }
    });
  }, []);

  const h = useNavigate();
  const auth = useAuth();
  const location = useLocation();

  const handleLoginTypeChange = (type: "sso" | "password") => {
    setLoginType(type);
    saveLoginType(type);
  };

  const handleSSO = (item: { name?: string; state?: string; url?: string }) => {
    setState(item.state || "");
    window.location.href = item.url || "/login";
  };

  const hasSSO = settings?.items && settings?.items.length > 0;

  return (
    <Background bg={bgInfo?.url}>
      <PinButton
        onClick={() => setRandom(toggleRandomBg())}
        title={random ? "固定壁纸" : "取消固定"}
      >
        {random ? <PushpinOutlined /> : <PushpinFilled />}
      </PinButton>

      <CenterHint>点击或悬停登录</CenterHint>

      <LoginCard>
        <LogoWrap>
          <Logo src={marsLogo} alt="Mars" />
          <LogoText>Mars</LogoText>
        </LogoWrap>

        {hasSSO && (
          <TabSwitch>
            <TabBtn
              $active={loginType === "sso"}
              onClick={() => handleLoginTypeChange("sso")}
            >
              SSO 登录
            </TabBtn>
            <TabBtn
              $active={loginType === "password"}
              onClick={() => handleLoginTypeChange("password")}
            >
              账号密码
            </TabBtn>
          </TabSwitch>
        )}

        {loginType === "sso" && hasSSO ? (
          <SSOPanel>
            <SSOList>
              {settings?.items?.map((item) => (
                <SSOBtn key={item.name} onClick={() => handleSSO(item)}>
                  <SSOIcon>
                    <svg viewBox="0 0 24 24" fill="currentColor">
                      <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-1 17.93c-3.95-.49-7-3.85-7-7.93 0-.62.08-1.21.21-1.79L9 15v1c0 1.1.9 2 2 2v1.93zm6.9-2.54c-.26-.81-1-1.39-1.9-1.39h-1v-3c0-.55-.45-1-1-1H8v-2h2c.55 0 1-.45 1-1V7h2c1.1 0 2-.9 2-2v-.41c2.93 1.19 5 4.06 5 7.41 0 2.08-.8 3.97-2.1 5.39z" />
                    </svg>
                  </SSOIcon>
                  <SSOContent>
                    <SSOTitle>{item.name}</SSOTitle>
                    <SSODesc>使用 {item.name} 账号登录</SSODesc>
                  </SSOContent>
                  <SSOArrow>→</SSOArrow>
                </SSOBtn>
              ))}
            </SSOList>
          </SSOPanel>
        ) : (
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
              rules={[{ required: true, message: "" }]}
            >
              <InputField placeholder="用户名" autoFocus />
            </Form.Item>

            <Form.Item
              name="password"
              rules={[{ required: true, message: "" }]}
            >
              <InputField placeholder="密码" type="password" />
            </Form.Item>

            <Form.Item style={{ marginBottom: 0 }}>
              <SubmitBtn type="submit">登录</SubmitBtn>
            </Form.Item>
          </Form>
        )}
      </LoginCard>

      {bgInfo?.copyright && <Copyright>{bgInfo.copyright}</Copyright>}
    </Background>
  );
};

export default memo(Login);

// ========== Styled Components (顺序很重要) ==========

const Copyright = styled.div`
  position: absolute;
  bottom: 24px;
  left: 50%;
  transform: translateX(-50%);
  font-size: 13px;
  color: rgba(255, 255, 255, 0.8);
  padding: 8px 20px;
  background: rgba(0, 0, 0, 0.3);
  backdrop-filter: blur(10px);
  border-radius: 20px;

  @media (min-width: 640px) {
    left: auto;
    right: 40px;
    transform: none;
  }
`;

const SubmitBtn = styled.button`
  width: 100%;
  height: 48px;
  border: none;
  border-radius: 12px;
  background: linear-gradient(135deg, ${theme.mainColor}, ${theme.deepColor});
  color: #fff;
  font-size: 15px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  margin-top: 8px;
  box-shadow: 0 4px 15px rgba(79, 70, 229, 0.3);

  &:hover {
    transform: translateY(-1px);
    box-shadow: 0 6px 20px rgba(79, 70, 229, 0.4);
  }

  &:active {
    transform: translateY(0);
  }
`;

const InputField = styled(Input)`
  height: 48px;
  border-radius: 12px;
  border: 2px solid #f0f0f0;
  background: #fafafa;
  font-size: 15px;
  padding: 0 18px;
  transition: all 0.3s;

  &:hover,
  &:focus {
    border-color: ${theme.mainColor};
    box-shadow: none;
    background: #fff;
  }

  &::placeholder {
    color: #aaa;
  }
`;

const SSOIcon = styled.div`
  width: 44px;
  height: 44px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: ${theme.mainColor};
  background: #f0eeff;
  border-radius: 12px;
  flex-shrink: 0;
  transition: all 0.25s;

  svg {
    width: 22px;
    height: 22px;
  }
`;

const SSOContent = styled.div`
  flex: 1;
  margin-left: 16px;
`;

const SSOTitle = styled.div`
  font-size: 15px;
  font-weight: 600;
  color: #222;
`;

const SSODesc = styled.div`
  font-size: 12px;
  color: #888;
  margin-top: 2px;
`;

const SSOArrow = styled.div`
  font-size: 18px;
  color: #ccc;
  transition: all 0.25s;
`;

const SSOBtn = styled.button`
  display: flex;
  align-items: center;
  padding: 16px 20px;
  border: 1px solid #eee;
  border-radius: 14px;
  background: #fafafa;
  cursor: pointer;
  transition: all 0.25s ease;
  width: 100%;
  text-align: left;

  &:hover {
    background: #fff;
    border-color: ${theme.mainColor};
    box-shadow: 0 4px 20px rgba(79, 70, 229, 0.1);
    transform: translateX(4px);

    ${SSOIcon} {
      background: ${theme.mainColor};
      color: #fff;
    }

    ${SSOArrow} {
      color: ${theme.mainColor};
      transform: translateX(4px);
    }
  }
`;

const SSOList = styled.div`
  display: flex;
  flex-direction: column;
  gap: 12px;
`;

const SSOPanel = styled.div`
  min-height: 120px;
`;

const TabBtn = styled.button<{ $active: boolean }>`
  flex: 1;
  padding: 12px 16px;
  border: none;
  border-radius: 10px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s;
  background: ${({ $active }) => ($active ? "#fff" : "transparent")};
  color: ${({ $active }) => ($active ? "#1a1a2e" : "#666")};
  box-shadow: ${({ $active }) =>
    $active ? "0 2px 8px rgba(0,0,0,0.08)" : "none"};

  &:hover {
    color: #1a1a2e;
  }
`;

const TabSwitch = styled.div`
  display: flex;
  background: #f5f5f7;
  border-radius: 12px;
  padding: 4px;
  margin-bottom: 28px;
`;

const LogoText = styled.span`
  font-size: 28px;
  font-weight: 700;
  color: #1a1a2e;
  font-family: "Comic Sans MS", cursive;
`;

const Logo = styled.img`
  width: 44px;
  height: 44px;
  border-radius: 12px;
`;

const LogoWrap = styled.div`
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  margin-bottom: 32px;
`;

const Background = styled.div<{ bg?: string }>`
  width: 100vw;
  height: 100vh;
  display: flex;
  justify-content: center;
  align-items: center;
  position: relative;
  background-position: center center;
  background-size: cover;
  background-image: ${({ bg }) => (bg ? `url(${bg})` : "none")};
  background-color: #1a1a2e;
`;

const PinButton = styled.div`
  position: absolute;
  top: 20px;
  right: 20px;
  font-size: 18px;
  color: rgba(255, 255, 255, 0.7);
  cursor: pointer;
  transition: all 0.3s;
  padding: 8px;
  border-radius: 50%;
  background: rgba(0, 0, 0, 0.2);
  backdrop-filter: blur(10px);

  &:hover {
    color: #fff;
    background: rgba(0, 0, 0, 0.3);
  }
`;

const CenterHint = styled.div`
  display: none;
`;

const LoginCard = styled.div`
  background: #fff;
  margin: 20px;
  padding: 40px;
  width: 400px;
  max-width: calc(100vw - 40px);
  border-radius: 20px;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
  animation: fadeInUp 0.6s ease-out forwards;
  opacity: 0;
  transform: translateY(20px);

  @keyframes fadeInUp {
    from {
      opacity: 0;
      transform: translateY(20px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }
`;
