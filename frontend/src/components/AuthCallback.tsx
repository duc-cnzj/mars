import React, { useEffect } from "react";
import { useLocation, useHistory } from "react-router-dom";
import { exchange, info } from "../api/auth";
import { setToken, setLogoutUrl } from "../utils/token";
import { useAuth } from "../contexts/auth";

function useQuery() {
  return new URLSearchParams(useLocation().search);
}
const Callback: React.FC = () => {
  let query = useQuery();
  const h = useHistory();
  let code = query.get("code");
  let state = query.get("state");
  const auth = useAuth();
  console.log(query.get("code"), query);
  if (!code) {
    h.push("/login");
  }
  useEffect(() => {
    if (code) {
      console.log("do query");
      exchange({ code }).then((res) => {
        setToken(res.data.token);
        info().then((res) => {
            setLogoutUrl(res.data.logout_url)
            auth.setUser(res.data)
        });
        h.push("/");
      });
    }
  }, [code, h]);

  return <div>login....</div>;
};

export default Callback;
