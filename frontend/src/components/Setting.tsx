import React, {useState,useEffect, memo} from "react";
import { Affix, Button } from "antd";
import { SettingOutlined } from "@ant-design/icons";
import { CreateProjectItem, List, selectList } from "../store/reducers/createProject";
import { useSelector } from "react-redux";
import { useHistory } from "react-router-dom";

const Setting: React.FC = () => {
  const list = useSelector<any, List>(selectList);
  const [can, setCan] = useState(true);
  useEffect(() => {
    let flag: boolean = false;
    for (const key in list) {
      if ((list[key] as CreateProjectItem).isLoading) {
        flag = true;
        break;
      }
    }

    setCan(!flag);
  }, [list]);

  let h = useHistory();

  return (
    <>
      <Affix offsetTop={130} style={{ position: "absolute", right: "10px" }}>
        <Button
          disabled={!can}
          size="large"
          type="ghost"
          shape="circle"
          icon={<SettingOutlined />}
          onClick={() => h.push("/web/gitlab_project_manager")}
        />
      </Affix>
    </>
  );
};

export default memo(Setting);
