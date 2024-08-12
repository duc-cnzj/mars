import React, { memo, useEffect, useState } from "react";
import dayjs from "dayjs";
import { Popover } from "antd";
import Coffee from "./Coffee";
import ajax from "../api/ajax";
import { components } from "../api/schema";
import IconFont from "./Icon";

require("dayjs/locale/zh-cn");

const AppFooter: React.FC = () => {
  const [version, setVersion] =
    useState<components["schemas"]["version.Response"]>();

  useEffect(() => {
    ajax.GET("/api/version").then(({ data }) => setVersion(data));
  }, []);

  return (
    <div className="copyright">
      <div style={{ fontSize: 14 }}>
        created by duc@2021~{new Date().getFullYear()}.
      </div>
      <div
        style={{
          fontSize: 12,
          display: "flex",
          justifyContent: "center",
          alignItems: "center",
        }}
      >
        version: {version?.version}, build at{" "}
        {dayjs(version?.buildDate).format("YYYY-MM-DD HH:mm:ss")}
        <Popover
          content={<Coffee />}
          overlayInnerStyle={{ padding: 0, margin: 0, borderRadius: 5 }}
          trigger="click"
        >
          <IconFont
            style={{
              margin: 7,
              width: 18,
              height: 18,
              cursor: "pointer",
              marginLeft: 10,
            }}
            name="#icon-dashang"
          />
        </Popover>
      </div>
    </div>
  );
};

export default memo(AppFooter);
