import React, { memo } from "react";
import { Tooltip, Switch, Form } from "antd";
import { InfoCircleOutlined } from "@ant-design/icons";

const DebugModeSwitch: React.FC<{
  value?: boolean;
  onChange?: (v: boolean) => void;
}> = ({ value, onChange }) => {
  return (
    <div
      style={{
        display: "flex",
        alignItems: "center",
        paddingBottom: 10,
        justifyContent: "center",
      }}
    >
      <div
        style={{
          paddingRight: 10,
          display: "flex",
          alignItems: "center",
        }}
      >
        <Tooltip
          placement="top"
          title={
            <div style={{ fontSize: 12 }}>
              <div>
                debug=true，在部署失败的时候，可以选择开启，开启之后能看到容器日志以及错误原因，并不能保证能成功访问页面。
              </div>
              <div>debug=false，部署成功即可访问页面。</div>
            </div>
          }
        >
          <InfoCircleOutlined />
        </Tooltip>
        <div style={{ paddingLeft: 3 }}>
          <span>debug 模式:</span>
        </div>
      </div>

      <Switch checked={value} onChange={onChange} />
    </div>
  );
};

export default memo(DebugModeSwitch);
