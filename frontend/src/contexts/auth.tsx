import React, { useEffect, createContext, useState, useContext } from "react";
import { Navigate, useLocation, useNavigate } from "react-router-dom";
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
  return <authContext.Provider value={auth}>{children}</authContext.Provider>;
}

function useAuth(): {
  login: (username: string, password: string, cb: () => void) => {};
  user: pb.auth.InfoResponse;
  setUser: (u: pb.auth.InfoResponse) => void;
  logout: (cb: () => void) => {};
  isAdmin: () => boolean;
} {
  return useContext(authContext);
}

function useProvideAuth() {
  const [user, setUser] = useState<pb.auth.InfoResponse>();

  const h = useNavigate();
  useEffect(() => {
    if (getToken() && !user) {
      info()
        .then((res) => {
          setUser(res.data);
        })
        .catch((e) => {
          removeToken();
          h("/login");
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
        message.error("用户名或者密码不正确");
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
