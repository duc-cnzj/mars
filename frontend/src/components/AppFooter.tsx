import React, { memo, useEffect, useState } from "react";
import { VersionResponse } from "../api/compiled";
import { version as versionApi } from "../api/version";
import dayjs from "dayjs";
import { Button } from "antd";
import { GithubOutlined } from "@ant-design/icons";

require("dayjs/locale/zh-cn");

const AppFooter: React.FC = () => {
  const [version, setVersion] = useState<VersionResponse>();

  useEffect(() => {
    versionApi().then((res) => setVersion(res.data));
  }, []);

  return (
    <div className="copyright">
      <div style={{ fontSize: 14 }}>created by duc@2021.</div>
      <div style={{ fontSize: 12 }}>
        version: {version?.version}, build at{" "}
        {dayjs(version?.build_date).format("YYYY-MM-DD HH:mm:ss")}
        <Button
          icon={<GithubOutlined />}
          target={"_blank"}
          href={version?.git_repo}
          type="link"
        ></Button>
      </div>
    </div>
  );
};

export default memo(AppFooter);
