import React, {
  useState,
  useEffect,
  useCallback,
  memo,
  useRef,
  useMemo,
} from "react";
import { getHighlightSyntax } from "../utils/highlight";
import ReactDiffViewer from "react-diff-viewer";
import { debounce } from "lodash";
import {
  Card,
  Skeleton,
  Divider,
  Select,
  List,
  Tag,
  Button,
  Popconfirm,
  Modal,
  message,
  Input,
} from "antd";
import AsciinemaPlayer from "./Player";
import pb from "../api/compiled";
import InfiniteScroll from "react-infinite-scroll-component";
import { events } from "../api/event";
import {
  deleteFile,
  downloadFile,
  diskInfo as diskInfoApi,
  deleteUndocumentedFiles,
} from "../api/file";
import ErrorBoundary from "../components/ErrorBoundary";
import { getToken } from "../utils/token";

const defaultPageSize = 15;
const { Option } = Select;

const initQuery = { action_type: pb.types.EventActionType.Unknown, search: "" }

const EventList: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [diskInfo, setDiskInfo] = useState<pb.file.DiskInfoResponse>();
  const [paginate, setPaginate] = useState<{
    page: number;
    page_size: number;
    count: number;
  }>({ page: 0, page_size: defaultPageSize, count: 0 });
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
          count: Number(res.count),
        });
        setLoading(false);
      })
      .catch((e) => {
        message.error(e.response.data.message);
        setLoading(false);
      });
  };

  const scrollDiv = useRef<HTMLDivElement>(null);
  const fetch = useCallback((action_type, search) => {
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
          count: Number(res.count),
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
  }, [fetch])

  const [config, setConfig] = useState({ old: "", new: "", title: "" });
  const [shellModalVisible, setShellModalVisible] = useState(false);
  const [fileID, setFileID] = useState(0);

  const getActionStyle = useCallback(
    (type: pb.types.EventActionType): React.ReactNode => {
      let style = { fontSize: 12, marginLeft: 5 };
      switch (type) {
        case pb.types.EventActionType.Create:
          return (
            <Tag color="#1890ff" style={style}>
              ??????
            </Tag>
          );
        case pb.types.EventActionType.Shell:
          return (
            <Tag color="#1890ff" style={style}>
              ????????????
            </Tag>
          );
        case pb.types.EventActionType.Update:
          return (
            <Tag color="#52c41a" style={style}>
              ??????
            </Tag>
          );
        case pb.types.EventActionType.Delete:
          return (
            <Tag color="#f5222d" style={style}>
              ??????
            </Tag>
          );
        case pb.types.EventActionType.Upload:
          return (
            <Tag color="#fcd34d" style={style}>
              ????????????
            </Tag>
          );
        case pb.types.EventActionType.Download:
          return (
            <Tag color="#2dd4bf" style={style}>
              ????????????
            </Tag>
          );
        case pb.types.EventActionType.Login:
          return (
            <Tag color="#38bdf8" style={style}>
              ??????
            </Tag>
          );
        case pb.types.EventActionType.DryRun:
          return (
            <Tag color="#818cf8" style={style}>
              ?????????
            </Tag>
          );
        default:
          return (
            <Tag color="#f1c40f" style={style}>
              ??????????????????????????????
            </Tag>
          );
      }
    },
    []
  );

  const highlightSyntax = useCallback(
    (str: string) => (
      <code
        dangerouslySetInnerHTML={{
          __html: getHighlightSyntax(str, "yaml"),
        }}
      />
    ),
    []
  );
  const [isModalVisible, setIsModalVisible] = useState(false);

  const showModal = useCallback(() => {
    setIsModalVisible(true);
  }, []);

  const handleOk = useCallback(() => {
    setIsModalVisible(false);
  }, []);

  const [clearLoading, setClearLoading] = useState(false);
  const clearDisk = useCallback(() => {
    setClearLoading(true);
    deleteUndocumentedFiles().then((res) => {
      message.success("????????????");
      diskInfoApi()
        .then(({ data }) => {
          setDiskInfo(data);
        })
        .finally(() => {
          setClearLoading(false);
        });
    });
  }, []);

  const handleCancel = useCallback(() => {
    setIsModalVisible(false);
  }, []);
  const getHeight = () => {
    let h = window.innerHeight - 260;
    if (h < 400) {
      return 400;
    }
    return h;
  };

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
            <div>????????????: {paginate.count} ???</div>
            <Select
              defaultValue={pb.types.EventActionType.Unknown}
              size="small"
              style={{ width: 200, marginLeft: 10 }}
              onChange={(v) => {
                setQueries((q) => ({ ...q, action_type: v }));
                fetch(v, queries.search)
              }}
            >
              <Option value={pb.types.EventActionType.Unknown}>??????</Option>
              <Option value={pb.types.EventActionType.Create}>??????</Option>
              <Option value={pb.types.EventActionType.Delete}>??????</Option>
              <Option value={pb.types.EventActionType.Download}>
                ????????????
              </Option>
              <Option value={pb.types.EventActionType.DryRun}>?????????</Option>
              <Option value={pb.types.EventActionType.Shell}>????????????</Option>
              <Option value={pb.types.EventActionType.Update}>??????</Option>
              <Option value={pb.types.EventActionType.Upload}>????????????</Option>
              <Option value={pb.types.EventActionType.Login}>??????</Option>
            </Select>
            <Input
              size="small"
              placeholder="????????????"
              style={{ marginLeft: 10, zIndex: 0 }}
              allowClear
              onChange={(v) => {
                setQueries((q) => ({ ...q, search: v.target.value }));
                debounceFetch(queries.action_type, v.target.value);
              }}
            />
          </div>
          <div style={{ fontSize: 12, fontWeight: "normal" }}>
            ????????????:{" "}
            <Button
              loading={clearLoading}
              style={{ fontSize: 10 }}
              type="link"
              onClick={clearDisk}
            >
              {diskInfo?.humanize_usage} ????????????
            </Button>
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
          hasMore={paginate.count > data.length}
          loader={<Skeleton avatar={false} paragraph={{ rows: 1 }} active />}
          endMessage={<Divider plain>?????????????????????????????????</Divider>}
          scrollableTarget="scrollableDiv"
        >
          <List
            dataSource={data}
            renderItem={(item: pb.types.EventModel) => (
              <List.Item key={item.id} className="events__list-item">
                <List.Item.Meta
                  title={
                    <div>
                      {item.username}
                      {getActionStyle(item.action)}
                      <span
                        style={{
                          fontSize: 10,
                          fontWeight: "normal",
                        }}
                      >
                        {item.event_at}
                      </span>
                    </div>
                  }
                  description={`${item.message}`}
                />
                {item.file_id > 0 &&
                  item.action === pb.types.EventActionType.Shell && (
                    <>
                      <Button
                        type="dashed"
                        style={{ marginRight: 5 }}
                        onClick={() => {
                          setShellModalVisible(true);
                          setFileID(item.file_id);
                        }}
                      >
                        ??????????????????{" "}
                        {item.duration && (
                          <span style={{ fontSize: "10px", marginLeft: 5 }}>
                            (??????: {item.duration})
                          </span>
                        )}
                      </Button>
                      <DeleteFile
                        onDelete={() => {
                          deleteFile({ id: item.file_id })
                            .then((res) => {
                              setData(
                                data.map((v) =>
                                  v.id === item.id ? { ...v, file_id: 0 } : v
                                )
                              );
                              message.success("????????????");
                            })
                            .catch((e) =>
                              message.error(e.response.data.message)
                            );
                        }}
                      />
                    </>
                  )}
                {item.file_id > 0 &&
                  item.action === pb.types.EventActionType.Upload && (
                    <>
                      <Button
                        type="dashed"
                        style={{ marginRight: 5 }}
                        onClick={() => {
                          downloadFile(item.file_id);
                        }}
                      >
                        ????????????
                      </Button>
                      <DeleteFile
                        onDelete={() => {
                          deleteFile({ id: item.file_id })
                            .then((res) => {
                              setData(
                                data.map((v) =>
                                  v.id === item.id ? { ...v, file_id: 0 } : v
                                )
                              );
                              message.success("????????????");
                            })
                            .catch((e) =>
                              message.error(e.response.data.message)
                            );
                        }}
                      />
                    </>
                  )}
                {!!(item.old || item.new) ? (
                  <Button
                    type="dashed"
                    onClick={() => {
                      setConfig({
                        old: item.old,
                        new: item.new,
                        title: `[${item.username}]: ` + item.message,
                      });
                      showModal();
                    }}
                  >
                    ????????????
                  </Button>
                ) : (
                  <></>
                )}
              </List.Item>
            )}
          />
        </InfiniteScroll>
      </div>
      <Modal
        width={"80%"}
        title={config.title}
        visible={isModalVisible}
        okText={"??????"}
        cancelText={"??????"}
        onOk={handleOk}
        footer={null}
        onCancel={handleCancel}
      >
        <ErrorBoundary>
          <div style={{ maxHeight: 550, overflowY: "auto" }}>
            <ReactDiffViewer
              disableWordDiff
              styles={{
                line: { fontSize: 12, wordBreak: "break-word" },
              }}
              useDarkTheme
              showDiffOnly
              splitView={config.old !== ""}
              renderContent={highlightSyntax}
              oldValue={config.old}
              newValue={config.new}
            />
          </div>
        </ErrorBoundary>
      </Modal>
      <Modal
        width={"65%"}
        title={null}
        destroyOnClose
        visible={shellModalVisible}
        footer={null}
        onCancel={() => {
          setShellModalVisible(false);
          setFileID(0);
        }}
      >
        <div style={{ width: "100%" }}>
          {fileID > 0 && (
            <AsciinemaPlayer
              speed={1.5}
              src={{
                url: `${process.env.REACT_APP_BASE_URL}/api/raw_file/${fileID}`,
                fetchOpts: {
                  method: "GET",
                  headers: { Authorization: getToken() },
                },
              }}
              cols={120}
              rows={36}
              idleTimeLimit={3}
              fit={"width"}
              terminalLineHeight={1.2}
              preload
              theme="tango"
            />
          )}
        </div>
      </Modal>
    </Card>
  );
};

const DeleteFile: React.FC<{ onDelete: () => void }> = ({ onDelete }) => {
  return (
    <Popconfirm
      title="??????????????????????????????"
      onConfirm={onDelete}
      okText="Yes"
      cancelText="No"
    >
      <Button type="dashed" style={{ marginRight: 5 }} danger>
        ??????
      </Button>
    </Popconfirm>
  );
};
export default memo(EventList);
