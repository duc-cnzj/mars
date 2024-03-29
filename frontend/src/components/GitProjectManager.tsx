import React, { useState, useEffect, useCallback, memo } from "react";
import { disabledProject, enabledProject, allProjects } from "../api/git";
import { CopyOutlined } from "@ant-design/icons";
import { copy } from "../utils/copy";
import { maxUploadSize } from "../api/file";
import type { UploadProps } from "antd";
import {
  List,
  Avatar,
  Card,
  Button,
  Select,
  message,
  Tooltip,
  Divider,
  Upload,
} from "antd";
import ConfigModal from "./ConfigModal";
import {
  GlobalOutlined,
  UploadOutlined,
  CloudDownloadOutlined,
} from "@ant-design/icons";
import pb from "../api/compiled";
import { downloadConfig } from "../api/file";
import { getToken } from "../utils/token";
import { RcFile } from "antd/lib/upload";
import { css } from "@emotion/css";
import theme from "../styles/theme";

const { Option } = Select;
const GitProjectManager: React.FC = () => {
  const [list, setList] = useState<pb.git.ProjectItem[]>([]);
  const [initLoading, setInitLoading] = useState(true);
  const [loadingList, setLoadingList] = useState<{ [name: number]: boolean }>();
  const [maxUploadInfo, setMaxUploadInfo] = useState({
    bytes: 0,
    humanizeSize: "",
  });
  const fetchList = useCallback(() => {
    return allProjects()
      .then((res) => {
        setList(res.data.items);
      })
      .catch((e) => message.error(e.response.data.message));
  }, [setList]);

  useEffect(() => {
    maxUploadSize().then(({ data }) => {
      setMaxUploadInfo({
        bytes: data.bytes,
        humanizeSize: data.humanize_size,
      });
    });
  }, []);

  useEffect(() => {
    fetchList().then(() => {
      setInitLoading(false);
    });
  }, [fetchList, setInitLoading]);

  const toggleStatus = async (item: pb.git.ProjectItem) => {
    setLoadingList((l) => ({ ...l, [item.id]: true }));
    try {
      if (item.enabled) {
        await disabledProject({ git_project_id: String(item.id) });
      } else {
        await enabledProject({ git_project_id: String(item.id) });
      }
    } catch (e: any) {
      message.error(e.response.data.message);
      setLoadingList((l) => ({ ...l, [item.id]: false }));
      return;
    }

    fetchList().then((res) => {
      setLoadingList((l) => ({ ...l, [item.id]: false }));
      message.success("操作成功");
    });
  };

  const [currentItem, setCurrentItem] = useState<pb.git.ProjectItem>();
  const [configVisible, setConfigVisible] = useState(false);
  const [selected, setSelected] = useState<pb.git.ProjectItem>();

  const onChange = useCallback(
    (v: Pick<UploadProps, "onChange">) => {
      if (!v) {
        setSelected(undefined);
        return;
      }
      let item = list.find((item) => item.id === v);
      if (item) {
        setSelected(item);
      }
    },
    [list]
  );

  const [loading, setLoading] = useState(false);

  const beforeUpload = useCallback(
    (file: RcFile, FileList: RcFile[]) => {
      if (maxUploadInfo.bytes === 0) {
        return true;
      }
      const isLtMaxSize = file.size <= maxUploadInfo.bytes;
      if (!isLtMaxSize) {
        message.error(`文件最大不能超过 ${maxUploadInfo.humanizeSize}!`);
      }
      setLoading(isLtMaxSize);

      return isLtMaxSize;
    },
    [maxUploadInfo]
  );

  let props: UploadProps = {
    name: "file",
    beforeUpload: beforeUpload,
    action: process.env.REACT_APP_BASE_URL + "/api/config/import",
    headers: {
      authorization: getToken(),
    },
    showUploadList: false,
    onChange(info) {
      if (info.file.status !== "uploading") {
        console.log(info.file, info.fileList);
      }
      if (info.file.status === "done") {
        message.success("导入成功");
        setLoading(false);
      } else if (info.file.status === "error") {
        message.error(`文件 ${info.file.name} 导入失败`);
        setLoading(false);
      }
    },
  };

  const projectNameFunc = useCallback(
    (item: pb.git.ProjectItem) =>
      `${item.name}${!!item.display_name ? `(${item.display_name})` : ""}`,
    []
  );

  return (
    <>
      <Card
        className="git"
        title={
          <div style={{ display: "flex", justifyContent: "space-between" }}>
            <span>git项目管理</span>
            <div>
              <Button type="link" size="small" onClick={() => downloadConfig()}>
                下载配置
              </Button>
              <Upload {...props}>
                <Button
                  disabled={loading}
                  loading={loading}
                  size="small"
                  style={{ fontSize: 12, marginRight: 5, margin: "5px 0" }}
                  icon={<UploadOutlined />}
                >
                  {loading ? "导入中" : "导入配置"}
                </Button>
              </Upload>
            </div>
          </div>
        }
        style={{ marginTop: 20, marginBottom: 30 }}
        bodyStyle={{ padding: 0 }}
      >
        <div style={{ padding: "24px 24px 0 24px" }}>
          <Select
            showSearch
            allowClear
            style={{ width: 500 }}
            placeholder="搜索项目"
            optionFilterProp="children"
            onChange={onChange}
            filterOption={(input, option: any) =>
              option.children.toLowerCase().indexOf(input.toLowerCase()) >= 0
            }
          >
            {list &&
              list.map((item, key) => (
                <Option value={item.id} key={key}>
                  {projectNameFunc(item)}
                </Option>
              ))}
          </Select>
        </div>
        <Divider />
        <List
          itemLayout="horizontal"
          loading={initLoading}
          dataSource={list.filter((item) =>
            selected ? item.id === selected.id : true
          )}
          renderItem={(item: pb.git.ProjectItem) => (
            <List.Item
              className={css`
                padding: 14px 24px !important;
                &:hover {
                  background-image: ${theme.lightLinear};
                }
              `}
              key={item.id}
              actions={[
                item.enabled && (
                  <Button
                    onClick={() => {
                      setCurrentItem(item);
                      setConfigVisible(true);
                    }}
                  >
                    查看配置
                  </Button>
                ),
                <Button
                  danger={item.enabled}
                  loading={loadingList && loadingList[item.id]}
                  type={!item.enabled ? "primary" : "dashed"}
                  onClick={() => toggleStatus(item)}
                >
                  {item.enabled ? "关闭" : "开启"}
                </Button>,
              ]}
            >
              <List.Item.Meta
                key={item.id}
                avatar={<Avatar src={item.avatar_url} />}
                title={
                  <div style={{ fontSize: 16 }}>
                    {projectNameFunc(item)}
                    <div
                      style={{
                        display: "inline-block",
                        fontSize: 10,
                        marginLeft: 3,
                        marginRight: 1,
                        fontWeight: "normal",
                      }}
                    >
                      (id: <span style={{ marginRight: 1 }}>{item.id}</span>
                      <span
                        style={{ cursor: "pointer" }}
                        onClick={() => copy(String(item.id), "已复制项目id！")}
                      >
                        <CopyOutlined />
                      </span>
                      )
                    </div>
                    {item.global_enabled && (
                      <>
                        <Tooltip
                          placement="top"
                          title="已使用全局配置"
                          overlayStyle={{ fontSize: "10px" }}
                        >
                          <GlobalOutlined
                            style={{
                              color: item.enabled ? "green" : "red",
                              marginLeft: 3,
                            }}
                          />
                        </Tooltip>
                        <Tooltip
                          placement="top"
                          title="下载项目配置"
                          overlayStyle={{ fontSize: "10px" }}
                        >
                          <CloudDownloadOutlined
                            onClick={() => downloadConfig(item.id)}
                            style={{ marginLeft: 3, cursor: "pointer" }}
                          />
                        </Tooltip>
                      </>
                    )}
                  </div>
                }
                description={
                  item.description ? item.description : "该项目还没有描述信息哦"
                }
              />
            </List.Item>
          )}
        />
        {configVisible && currentItem && (
          <ConfigModal
            visible={configVisible}
            item={currentItem}
            onCancel={() => setConfigVisible(false)}
          />
        )}
      </Card>
    </>
  );
};

export default memo(GitProjectManager);
