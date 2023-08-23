import React, { memo, useEffect, useState } from "react";
import pb from "../api/compiled";
import { version as versionApi } from "../api/version";
import dayjs from "dayjs";
import { Button, Popover } from "antd";
import { CoffeeOutlined, GithubOutlined } from "@ant-design/icons";
import Coffee from "./Coffee";

require("dayjs/locale/zh-cn");

const AppFooter: React.FC = () => {
  const [version, setVersion] = useState<pb.version.Response>();

  useEffect(() => {
    versionApi().then((res) => setVersion(res.data));
  }, []);

  return (
    <div className="copyright">
      <div style={{ fontSize: 14 }}>created by duc@2021.</div>
      <div
        style={{
          fontSize: 12,
          display: "flex",
          justifyContent: "center",
          alignItems: "center",
        }}
      >
        version: {version?.version}, build at{" "}
        {dayjs(version?.build_date).format("YYYY-MM-DD HH:mm:ss")}
        <Button
          icon={<GithubOutlined />}
          target={"_blank"}
          href={version?.git_repo}
          type="link"
        ></Button>
        <Popover
          content={<Coffee />}
          overlayInnerStyle={{ padding: 0, margin: 0 }}
          trigger="click"
        >
          <svg
            style={{ width: 18, height: 18, cursor: "pointer" }}
            className="icon"
            aria-hidden="true"
          >
            <use xlinkHref="#icon-dashang"></use>
          </svg>
        </Popover>
      </div>
    </div>
  );
};

export default memo(AppFooter);
