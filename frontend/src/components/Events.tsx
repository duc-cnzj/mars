import React, {
  useState,
  useEffect,
  useCallback,
  memo,
  useRef,
  useMemo,
} from "react";
import { debounce } from "lodash";
import {
  Card,
  Spin,
  Divider,
  Select,
  List,
  Tag,
  Button,
  Popconfirm,
  Drawer,
  message,
  Input,
  Radio,
  RadioChangeEvent,
} from "antd";
import theme from "../styles/theme";
import AsciinemaPlayer from "./Player";
import pb from "../api/compiled";
import InfiniteScroll from "react-infinite-scroll-component";
import { events, showEvent } from "../api/event";
import { showRecords } from "../api/file";
import { deleteFile, downloadFile, diskInfo as diskInfoApi } from "../api/file";
import ErrorBoundary from "../components/ErrorBoundary";
import DiffViewer from "./DiffViewer";
import { css } from "@emotion/css";
import {
  ClockCircleOutlined,
  PlayCircleOutlined,
  UserOutlined,
} from "@ant-design/icons";
import dayjs from "dayjs";
import styled from "@emotion/styled";

const defaultPageSize = 15;
const { Option } = Select;

const initQuery = { action_type: pb.types.EventActionType.Unknown, search: "" };

const EventList: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [diskInfo, setDiskInfo] = useState<pb.file.DiskInfoResponse>();
  const [paginate, setPaginate] = useState<{
    page: number;
    page_size: number;
  }>({ page: 0, page_size: defaultPageSize });
  const [data, setData] = useState<pb.types.EventModel[]>([]);
  const [queries, setQueries] = useState<{
    action_type: pb.types.EventActionType;
    search: string;
  }>(initQuery);

  useEffect(() => {
    diskInfoApi().then(({ data }) => setDiskInfo(data));
  }, []);

  const loadMoreData = () => {
    if (loading) {
      return;
    }
    setLoading(true);
    events({
      page: paginate.page + 1,
      page_size: paginate.page_size,
      action_type: queries.action_type,
      search: queries.search,
    })
      .then(({ data: res }) => {
        setData((data) => [...data, ...res.items]);
        setPaginate({
          page: Number(res.page),
          page_size: Number(res.page_size),
        });
        setLoading(false);
      })
      .catch((e) => {
        message.error(e.response.data.message);
        setLoading(false);
      });
  };

  const scrollDiv = useRef<HTMLDivElement>(null);
  const fetch = useCallback((action_type: any, search: any) => {
    if (scrollDiv.current) {
      scrollDiv.current.scrollTo(0, 0);
    }
    events({
      page: 1,
      page_size: defaultPageSize,
      action_type: action_type,
      search: search,
    })
      .then(({ data: res }) => {
        setData(res.items);
        setPaginate({
          page: Number(res.page),
          page_size: Number(res.page_size),
        });
      })
      .catch((e) => {
        message.error(e.response.data.message);
      });
  }, []);

  const debounceFetch = useMemo(
    () =>
      debounce((action_type, search) => {
        fetch(action_type, search);
      }, 500),
    [fetch]
  );

  useEffect(() => {
    fetch(initQuery.action_type, initQuery.search);
  }, [fetch]);

  const [config, setConfig] = useState<{
    old: string;
    new: string;
    title: React.ReactNode;
  }>({ old: "", new: "", title: "" });
  const [recordWindow, setRecordWindow] = useState<{
    title: React.ReactNode;
    visible: boolean;
  }>({ title: "", visible: false });

  const detail = useCallback(
    (username: string, message: string): React.ReactNode => {
      return (
        <div>
          <span style={{ color: "red", marginRight: 5 }}>
            <UserOutlined />
            {username}
          </span>
          <span style={{ fontWeight: "normal", fontSize: 13 }}>{message}</span>
        </div>
      );
    },
    []
  );

  const getActionStyle = useCallback(
    (type: pb.types.EventActionType): React.ReactNode => {
      let style = { fontSize: 12, marginLeft: 5 };
      switch (type) {
        case pb.types.EventActionType.Create:
          return (
            <Tag color="#1890ff" style={style}>
              创建
            </Tag>
          );
        case pb.types.EventActionType.Shell:
          return (
            <Tag color="#1890ff" style={style}>
              执行命令
            </Tag>
          );
        case pb.types.EventActionType.Exec:
          return (
            <Tag color="#a78bfa" style={style}>
              SDK 执行命令
            </Tag>
          );
        case pb.types.EventActionType.Update:
          return (
            <Tag color="#52c41a" style={style}>
              更新
            </Tag>
          );
        case pb.types.EventActionType.Delete:
          return (
            <Tag color="#f5222d" style={style}>
              删除
            </Tag>
          );
        case pb.types.EventActionType.Upload:
          return (
            <Tag color="#fb7185" style={style}>
              上传文件
            </Tag>
          );
        case pb.types.EventActionType.Download:
          return (
            <Tag color="#2dd4bf" style={style}>
              下载文件
            </Tag>
          );
        case pb.types.EventActionType.Login:
          return (
            <Tag color="#38bdf8" style={style}>
              登录
            </Tag>
          );
        case pb.types.EventActionType.CancelDeploy:
          return (
            <Tag color="#facc15" style={style}>
              取消部署
            </Tag>
          );
        case pb.types.EventActionType.DryRun:
          return (
            <Tag color="#818cf8" style={style}>
              试运行
            </Tag>
          );
        default:
          return (
            <Tag color="#f1c40f" style={style}>
              俺也不知道这是啥操作
            </Tag>
          );
      }
    },
    []
  );

  const [isWindowVisible, setIsWindowVisible] = useState(false);

  const showWindow = useCallback(
    (id: number) => {
      showEvent(id).then(({ data }) => {
        data.event &&
          setConfig({
            old: data.event.old,
            new: data.event.new,
            title: detail(data.event.username, data.event.message),
          });
        setIsWindowVisible(true);
      });
    },
    [detail]
  );

  const handleCancel = useCallback(() => {
    setIsWindowVisible(false);
    setConfig({ new: "", old: "", title: "" });
  }, []);
  const getHeight = () => {
    let h = window.innerHeight - 260;
    if (h < 400) {
      return 400;
    }
    return h;
  };

  const [records, setRecords] = useState<string[]>([]);
  const [key, setKey] = useState(0);

  const fetchFileRaw = useCallback((id: number) => {
    showRecords(id).then(({ data }) => {
      setRecords(data.items);
    });
  }, []);

  return (
    <Card
      title={
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            alignItems: "center",
          }}
        >
          <div style={{ display: "flex", alignItems: "center" }}>
            <div>事件日志</div>
            <Select
              defaultValue={pb.types.EventActionType.Unknown}
              size="small"
              style={{ width: 300, marginLeft: 10 }}
              onChange={(v) => {
                setQueries((q) => ({ ...q, action_type: v }));
                fetch(v, queries.search);
              }}
            >
              <Option value={pb.types.EventActionType.Unknown}>全部</Option>
              <Option value={pb.types.EventActionType.Create}>创建</Option>
              <Option value={pb.types.EventActionType.Delete}>删除</Option>
              <Option value={pb.types.EventActionType.Download}>
                下载文件
              </Option>
              <Option value={pb.types.EventActionType.DryRun}>试运行</Option>
              <Option value={pb.types.EventActionType.Shell}>执行命令</Option>
              <Option value={pb.types.EventActionType.Exec}>
                SDK 执行命令
              </Option>
              <Option value={pb.types.EventActionType.Update}>更新</Option>
              <Option value={pb.types.EventActionType.Upload}>上传文件</Option>
              <Option value={pb.types.EventActionType.Login}>登录</Option>
              <Option value={pb.types.EventActionType.CancelDeploy}>
                取消部署
              </Option>
            </Select>
            <Input
              size="small"
              placeholder="搜索内容"
              style={{ marginLeft: 10, zIndex: 0 }}
              allowClear
              onChange={(v) => {
                setQueries((q) => ({ ...q, search: v.target.value }));
                debounceFetch(queries.action_type, v.target.value);
              }}
            />
          </div>
          <div style={{ fontSize: 12, fontWeight: "normal" }}>
            文件占用:{" "}
            <span style={{ color: "blue" }}>{diskInfo?.humanize_usage}</span>
          </div>
        </div>
      }
      bordered={false}
      bodyStyle={{
        height: getHeight(),
        overflowY: "auto",
        padding: 0,
      }}
      style={{
        marginTop: 20,
        marginBottom: 30,
      }}
    >
      <div
        id="scrollableDiv"
        ref={scrollDiv}
        style={{ height: "100%", overflowY: "auto" }}
      >
        <InfiniteScroll
          dataLength={data.length}
          next={loadMoreData}
          hasMore={data.length !== 0 && data.length % paginate.page_size === 0}
          loader={
            <Spin
              style={{
                margin: "20px 0",
                display: "flex",
                justifyContent: "center",
              }}
              size="large"
            />
          }
          endMessage={<Divider plain>老铁，别翻了，到底了！</Divider>}
          scrollableTarget="scrollableDiv"
        >
          <List
            dataSource={data}
            renderItem={(item: pb.types.EventModel) => (
              <ListItem key={item.id}>
                <List.Item.Meta
                  title={
                    <div>
                      {item.username}
                      {getActionStyle(item.action)}
                      <div
                        className={css`
                          display: inline;
                          font-size: 10px;
                          font-weight: normal;
                        `}
                      >
                        {item.event_at}
                        <DateSpan>
                          {dayjs(item.created_at).format("YYYY-MM-DD HH:mm:ss")}
                        </DateSpan>
                      </div>
                    </div>
                  }
                  description={`${item.message}`}
                />
                {!!item.file &&
                  (item.action === pb.types.EventActionType.Shell ||
                    item.action === pb.types.EventActionType.Exec) && (
                    <>
                      <Button
                        type="dashed"
                        style={{ marginRight: 5 }}
                        onClick={() => {
                          setRecordWindow({
                            visible: true,
                            title: detail(item.username, item.message),
                          });
                          fetchFileRaw(item.file_id);
                        }}
                      >
                        查看操作记录{" "}
                        {item.duration && (
                          <span style={{ fontSize: "10px", marginLeft: 5 }}>
                            (时长: {item.duration}, 大小:{" "}
                            {item.file?.humanize_size})
                          </span>
                        )}
                      </Button>
                      <DeleteFile
                        onDelete={() => {
                          deleteFile({ id: item.file_id })
                            .then((res) => {
                              setData(
                                data.map((v) =>
                                  v.id === item.id ? { ...v, file: null } : v
                                )
                              );
                              message.success("删除成功");
                            })
                            .catch((e) =>
                              message.error(e.response.data.message)
                            );
                        }}
                      />
                    </>
                  )}
                {!!item.file &&
                  item.action === pb.types.EventActionType.Upload && (
                    <>
                      <Button
                        type="dashed"
                        style={{ marginRight: 5 }}
                        onClick={() => {
                          downloadFile(item.file_id);
                        }}
                      >
                        下载文件
                      </Button>
                      <DeleteFile
                        onDelete={() => {
                          deleteFile({ id: item.file_id })
                            .then((res) => {
                              setData(
                                data.map((v) =>
                                  v.id === item.id ? { ...v, file: null } : v
                                )
                              );
                              message.success("删除成功");
                            })
                            .catch((e) =>
                              message.error(e.response.data.message)
                            );
                        }}
                      />
                    </>
                  )}
                {item.has_diff && (
                  <Button
                    type="dashed"
                    onClick={() => {
                      showWindow(item.id);
                    }}
                  >
                    查看改动
                  </Button>
                )}
              </ListItem>
            )}
          />
        </InfiniteScroll>
      </div>
      <Drawer
        destroyOnClose
        width={"100%"}
        title={config.title}
        open={isWindowVisible}
        footer={null}
        onClose={handleCancel}
      >
        <ErrorBoundary>
          <div style={{ maxHeight: "100%", overflowY: "auto" }}>
            <DiffViewer
              showCopyButton
              styles={{
                line: { fontSize: 12, wordBreak: "break-word" },
              }}
              showDiffOnly
              splitView={config.old !== ""}
              mode="yaml"
              oldValue={config.old}
              newValue={config.new}
            />
          </div>
        </ErrorBoundary>
      </Drawer>
      <Drawer
        width={"100%"}
        title={recordWindow.title}
        destroyOnClose
        open={recordWindow.visible}
        footer={null}
        onClose={() => {
          setRecordWindow({ visible: false, title: "" });
          setRecords([]);
          setKey(0);
        }}
      >
        <div style={{ width: "100%" }}>
          {records.length > 1 && (
            <>
              <Radio.Group
                onChange={(e: RadioChangeEvent) => setKey(e.target.value)}
                value={key}
              >
                {records.map((_, index) => (
                  <Radio value={index} key={index}>
                    <Tag
                      color={key === index ? "success" : "default"}
                      icon={
                        key === index ? (
                          <PlayCircleOutlined />
                        ) : (
                          <ClockCircleOutlined />
                        )
                      }
                    >
                      片段 {index + 1}
                    </Tag>
                  </Radio>
                ))}
              </Radio.Group>
              <Divider plain />
            </>
          )}
          {records.map((v, index) => (
            <div
              key={index}
              style={{ display: index === key ? "block" : "none" }}
            >
              <AsciinemaPlayer
                speed={1.5}
                src={{ data: records[key] }}
                idleTimeLimit={3}
                fit={false}
                terminalLineHeight={1.5}
                preload
                theme="tango"
              />
            </div>
          ))}
        </div>
      </Drawer>
    </Card>
  );
};

const DeleteFile: React.FC<{ onDelete: () => void }> = ({ onDelete }) => {
  return (
    <Popconfirm
      title="你确定要删除该文件吗"
      onConfirm={onDelete}
      okText="Yes"
      cancelText="No"
    >
      <Button type="dashed" style={{ marginRight: 5 }} danger>
        删除
      </Button>
    </Popconfirm>
  );
};

export default memo(EventList);

const DateSpan = styled.span`
  opacity: 0;
  transition: 0.3s opacity ease-in;
  margin-left: 5px;
`;

const ListItem = styled(List.Item)`
  padding: 14px 24px !important;
  &:hover {
    background-image: ${theme.lightLinear} !important;
    ${DateSpan} {
      opacity: 1 !important;
    }
  }
`;
