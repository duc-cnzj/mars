import React, { useState, useEffect, useCallback } from "react";
import { getHighlightSyntax } from "../utils/highlight";
import ReactDiffViewer from "react-diff-viewer";
import {
  Card,
  Skeleton,
  Divider,
  List,
  Tag,
  Button,
  Modal,
  message,
} from "antd";
import pb from "../api/compiled";
import InfiniteScroll from "react-infinite-scroll-component";
import { events } from "../api/event";
import ErrorBoundary from "../components/ErrorBoundary";

const defaultPageSize = 15;

const EventList: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [paginate, setPaginate] = useState<{
    page: number;
    page_size: number;
    count: number;
  }>({ page: 0, page_size: defaultPageSize, count: 0 });
  const [data, setData] = useState<pb.EventList.item[]>([]);

  const loadMoreData = () => {
    if (loading) {
      return;
    }
    setLoading(true);
    events({ page: paginate.page + 1, page_size: paginate.page_size })
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

  useEffect(() => {
    events({ page: 1, page_size: defaultPageSize })
      .then(({ data: res }) => {
        setData((data) => [...data, ...res.items]);
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

  const [config, setConfig] = useState({ old: "", new: "", title: "" });

  const getActionStyle = useCallback((type: pb.ActionType): React.ReactNode => {
    let style = { fontSize: 12, marginLeft: 5 };
    switch (type) {
      case pb.ActionType.Create:
        return (
          <Tag color="#1890ff" style={style}>
            创建
          </Tag>
        );
      case pb.ActionType.Update:
        return (
          <Tag color="#52c41a" style={style}>
            更新
          </Tag>
        );
      case pb.ActionType.Delete:
        return (
          <Tag color="#f5222d" style={style}>
            删除
          </Tag>
        );
      default:
        return (
          <Tag color="#f1c40f" style={style}>
            俺也不知道这是啥操作
          </Tag>
        );
    }
  }, []);

  const highlightSyntax = (str: string) => (
    <code
      dangerouslySetInnerHTML={{
        __html: getHighlightSyntax(str, "yaml"),
      }}
    />
  );

  const [isModalVisible, setIsModalVisible] = useState(false);

  const showModal = useCallback(() => {
    setIsModalVisible(true);
  }, []);

  const handleOk = useCallback(() => {
    setIsModalVisible(false);
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
      title={<div>事件日志: {paginate.count} 条</div>}
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
        style={{ height: "100%", overflowY: "auto", padding: 24 }}
      >
        <InfiniteScroll
          dataLength={data.length}
          next={loadMoreData}
          hasMore={paginate.count > data.length}
          loader={<Skeleton avatar={false} paragraph={{ rows: 1 }} active />}
          endMessage={<Divider plain>老铁，别翻了，到底了！</Divider>}
          scrollableTarget="scrollableDiv"
        >
          <List
            dataSource={data}
            renderItem={(item: pb.EventList.item) => (
              <List.Item key={item.id}>
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
                {item.action === pb.ActionType.Update ? (
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
                    查看改动
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
        width={"100%"}
        title={config.title}
        visible={isModalVisible}
        okText={"确定"}
        cancelText={"取消"}
        onOk={handleOk}
        onCancel={handleCancel}
        bodyStyle={{ width: "100%" }}
      >
        <ErrorBoundary>
          <ReactDiffViewer
            disableWordDiff
            styles={{
              line: { fontSize: 12 },
              gutter: { padding: "0 5px", minWidth: 20 },
              marker: { padding: "0 6px" },
              diffContainer: {
                display: "block",
                width: "100%",
              },
            }}
            useDarkTheme
            showDiffOnly
            splitView={false}
            renderContent={highlightSyntax}
            oldValue={config.old}
            newValue={config.new}
          />
        </ErrorBoundary>
      </Modal>
    </Card>
  );
};

export default EventList;
