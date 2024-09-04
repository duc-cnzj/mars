import React, { memo } from "react";
import dayjs from "dayjs";
import { Popover } from "antd";
import Coffee from "./Coffee";
import IconFont from "./Icon";
import useVersion from "../contexts/useVersion";
import { css } from "@emotion/css";

require("dayjs/locale/zh-cn");

const AppFooter: React.FC = () => {
  const version = useVersion();

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
            className={css`
              margin: 3px;
              width: 26px;
              height: 26px;
              cursor: pointer;
              margin-left: 10px;
              transform: scaleX(-1);
              transition: all 0.3s;
              &:hover {
                scale: 1.3;
              }
            `}
            name="#icon-naicha"
          />
        </Popover>
      </div>
    </div>
  );
};

export default memo(AppFooter);
