import React from "react";
import { Layout } from "antd";
import { ProvideWebsocket } from "../contexts/useWebsocket";
import AppHeader from "./AppHeader";
import AppFooter from "./AppFooter";
import { GlobalScrollbar } from "mac-scrollbar";
import "mac-scrollbar/dist/mac-scrollbar.css";
import { css } from "@emotion/css";
import appTheme from "../styles/theme";
import { Outlet } from "react-router-dom";

const { Header, Content, Footer } = Layout;

const AppLayout: React.FC = () => {
  return (
    <ProvideWebsocket>
      <GlobalScrollbar />
      <Layout className="app">
        <Header
          className={css`
            position: "fixed";
            z-index: 1;
            width: "100%";
            overflow: "hidden";
            background-image: ${appTheme.mainLinear};
          `}
          style={{
            position: "fixed",
            zIndex: 1,
            width: "100%",
            overflow: "hidden",
          }}
        >
          <AppHeader />
        </Header>{" "}
        <Content
          className={css`
            margin-top: 64px;
            padding: 0 50px;
            min-height: calc(100vh - 122px) !important;
          `}
        >
          <Outlet />
        </Content>
        <Footer
          className={css`
            background-image: ${appTheme.footerLinear};
            padding: 8px 50px 2px !important;
            text-align: center;
            .copyright {
              font-family: "Gill Sans", "Gill Sans MT", Calibri, "Trebuchet MS",
                sans-serif;
              color: ${appTheme.mainFontColor};
            }
          `}
        >
          <AppFooter />
        </Footer>
      </Layout>
    </ProvideWebsocket>
  );
};

export default AppLayout;
