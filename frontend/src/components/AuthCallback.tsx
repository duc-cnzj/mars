import React, { useEffect, memo } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { exchange, info } from "../api/auth";
import { setToken, setLogoutUrl } from "../utils/token";
import { useAuth } from "../contexts/auth";
import { getState, removeState } from "../utils/token";
import { message } from "antd";

function useQuery() {
  return new URLSearchParams(useLocation().search);
}
const Callback: React.FC = () => {
  let query = useQuery();
  const h = useNavigate();
  let code = query.get("code");
  let state = query.get("state");
  const auth = useAuth();
  if (!code) {
    h("/login");
  }
  useEffect(() => {
    if (code) {
      if (state === getState()) {
        exchange({ code }).then((res) => {
          setToken(res.data.token);
          info().then((res) => {
            setLogoutUrl(res.data.logout_url);
            auth.setUser(res.data);
          });
          h("/");
        });
      } else {
        message.error("state 不一致，请重新登录");
        removeState();
        h("/login");
      }
    }
  }, [code, h, auth, state]);

  return <div>登录中....</div>;
};

export default memo(Callback);
