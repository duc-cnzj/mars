import React, { useEffect, createContext, useState, useContext } from "react";
import { Route, Redirect, useHistory } from "react-router-dom";
import { message } from "antd";
import { login, info } from "../api/auth";
import { setToken, getToken, removeToken } from "../utils/token";
import pb from "../api/compiled";

export const authContext = createContext<any>(null);

const realAuth = {
  async signin(username: string, password: string) {
    return login({ username, password }).then((res) => {
      setToken(res.data.token);
      return this.info();
    });
  },
  async signout() {
    removeToken();
  },
  async info() {
    console.log(getToken());
    return info();
  },
};

function ProvideAuth({ children }: { children: any }) {
  const auth = useProvideAuth();
  console.log("auth", auth);
  return <authContext.Provider value={auth}>{children}</authContext.Provider>;
}

function useAuth(): {
  login: (username: string, password: string, cb: () => void) => {};
  user: pb.InfoResponse;
  setUser: (u: pb.InfoResponse) => void;
  logout: (cb: () => void) => {};
} {
  return useContext(authContext);
}

function useProvideAuth() {
  const [user, setUser] = useState<any>(null);

  const h = useHistory();
  useEffect(() => {
    if (getToken() && !user) {
      console.log("set user");
      info()
        .then((res) => {
          setUser(res.data);
        })
        .catch((e) => {
          removeToken();
          h.push("/login");
        });
    }
  }, [user, h]);

  const signin = (username: string, password: string, cb: any) => {
    realAuth
      .signin(username, password)
      .then((res) => {
        setUser(res.data);
        cb();
        message.success("登录成功");
      })
      .catch((e) => {
        console.log(e);
        message.error("登录失败");
      });
  };

  const signout = (cb: any) => {
    realAuth.signout().then(() => {
      setUser(null);
      cb();
      message.success("登出成功");
    });
  };

  return {
    setUser,
    user,
    login: signin,
    logout: signout,
  };
}

function PrivateRoute({
  path,
  exact,
  children,
  ...rest
}: {
  exact?: boolean;
  path: string;
  children: any;
}) {
  return (
    <Route
      {...rest}
      exact={exact}
      render={({ location }) =>
        getToken() ? (
          children
        ) : (
          <Redirect
            to={{
              pathname: "/login",
              state: { from: location },
            }}
          />
        )
      }
    />
  );
}
function GuestRoute({
  path,
  children,
  ...rest
}: {
  path: string;
  children: any;
}) {
  return (
    <Route
      {...rest}
      render={({ location }) =>
        !getToken() ? (
          children
        ) : (
          <Redirect
            to={{
              pathname: "/",
              state: { from: location },
            }}
          />
        )
      }
    />
  );
}

export { ProvideAuth, PrivateRoute, useAuth, useProvideAuth, GuestRoute };
