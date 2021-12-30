import React, { useEffect, memo } from "react";
import { useLocation, useHistory } from "react-router-dom";
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
  const h = useHistory();
  let code = query.get("code");
  let state = query.get("state");
  const auth = useAuth();
  if (!code) {
    h.push("/login");
  }
  useEffect(() => {
    if (code) {
      if (state === getState()) {
        console.log("do query");
        exchange({ code }).then((res) => {
          setToken(res.data.token);
          info().then((res) => {
            setLogoutUrl(res.data.logout_url);
            auth.setUser(res.data);
          });
          h.push("/");
        });
      } else {
        message.error("state 不一致，请重新登录");
        removeState();
        h.push("/login");
      }
    }
  }, [code, h, auth, state]);

  return <div>login....</div>;
};

export default memo(Callback);
