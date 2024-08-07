import React, { useEffect, createContext, useState, useContext } from "react";
import { Navigate, useLocation, useNavigate } from "react-router-dom";
import { message } from "antd";
import { setToken, getToken, removeToken } from "../utils/token";
import ajax from "../api/ajax";

export const authContext = createContext<any>(null);

const realAuth = {
  async signin(username: string, password: string) {
    return ajax
      .POST("/api/auth/login", { body: { username, password } })
      .then(({ data }) => {
        data && setToken(data.token);
        return this.info();
      });
  },
  async signout() {
    removeToken();
  },
  async info() {
    return ajax.GET("/api/auth/info");
  },
};

function ProvideAuth({ children }: { children: any }) {
  const auth = useProvideAuth();
  return <authContext.Provider value={auth}>{children}</authContext.Provider>;
}

export interface userInfo {
  id: number;
  avatar: string;
  name: string;
  email: string;
  logoutUrl: string;
  roles: string[];
}

function useAuth(): {
  login: (username: string, password: string, cb: () => void) => {};
  user: userInfo;
  setUser: (u: userInfo) => void;
  logout: (cb: () => void) => {};
  isAdmin: () => boolean;
} {
  return useContext(authContext);
}

function useProvideAuth() {
  const [user, setUser] = useState<userInfo>();

  const h = useNavigate();
  useEffect(() => {
    if (getToken() && !user) {
      ajax.GET("/api/auth/info").then(({ data, error }) => {
        if (error) {
          removeToken();
          h("/login");
          return;
        }
        setUser(data);
      });
    }
  }, [user, h]);

  const signin = (username: string, password: string, cb: any) => {
    realAuth.signin(username, password).then(({ data, error }) => {
      if (error) {
        console.log(error);
        message.error("用户名或者密码不正确");
        return;
      }
      setUser(data);
      cb();
      message.success("登录成功");
    });
  };

  const signout = (cb: any) => {
    realAuth.signout().then(() => {
      setUser(undefined);
      cb();
      message.success("登出成功");
    });
  };

  const isAdmin = () => {
    return user
      ? user.roles.filter((item) => item === "mars_admin").length > 0
      : false;
  };

  return {
    setUser,
    user,
    isAdmin,
    login: signin,
    logout: signout,
  };
}

function PrivateRoute({ children }: { children: JSX.Element }) {
  let location = useLocation();
  if (!getToken()) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  return children;
}

function GuestRoute({ children }: { children: JSX.Element }) {
  let location = useLocation();
  if (!!getToken()) {
    return <Navigate to="/" state={{ from: location }} />;
  }

  return children;
}

export { ProvideAuth, PrivateRoute, useAuth, useProvideAuth, GuestRoute };
