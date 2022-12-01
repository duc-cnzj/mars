import React, { useState, useEffect, useCallback, memo, useRef } from "react";
import dayjs from "dayjs";
import {
  Tag,
  Card,
  Popconfirm,
  Select,
  InputNumber,
  Skeleton,
  Divider,
  List,
  Form,
  Button,
  Modal,
  message,
  Input,
} from "antd";
import pb from "../api/compiled";
import InfiniteScroll from "react-infinite-scroll-component";
import { List as tokenList, Grant, Revoke } from "../api/access_token";
import { useForm } from "antd/es/form/Form";
import classNames from "classnames";
import { copy } from "../utils/copy";

const defaultPageSize = 15;

const { Option } = Select;

type unitImp = "day" | "hour" | "minute" | "second";

const AccessTokenManager: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [paginate, setPaginate] = useState<{
    page: number;
    page_size: number;
    count: number;
  }>({ page: 0, page_size: defaultPageSize, count: 0 });
  const [data, setData] = useState<pb.types.AccessTokenModel[]>([]);

  const loadMoreData = () => {
    if (loading) {
      return;
    }
    setLoading(true);
    tokenList({
      page: paginate.page + 1,
      page_size: paginate.page_size,
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
  const fetch = useCallback(() => {
    if (scrollDiv.current) {
      scrollDiv.current.scrollTo(0, 0);
    }
    tokenList({
      page: 1,
      page_size: defaultPageSize,
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

  useEffect(() => {
    fetch();
  }, [fetch]);

  const [isModalVisible, setIsModalVisible] = useState(false);

  const getHeight = () => {
    let h = window.innerHeight - 260;
    if (h < 400) {
      return 400;
    }
    return h;
  };

  const [form] = useForm();

  const onFinish = (values: any) => {
    Grant({ usage: values.usage, expire_seconds: values.expire_seconds })
      .then((res) => {
        message.success("创建成功");
        setIsModalVisible(false);
        form.resetFields();
        setData([]);
        fetch();
      })
      .catch((e) => {
        message.error(e.response.data.message);
      });
  };
  const onFinishFailed = () => {
    console.log("values");
  };
  const [unit, setUnit] = useState<unitImp>("day");

  const selectAfter = (
    <Select
      defaultValue={unit}
      onChange={(v: unitImp) => setUnit(v)}
      style={{ width: 80 }}
    >
      <Option value="day">天</Option>
      <Option value="hour">小时</Option>
      <Option value="minute">分钟</Option>
      <Option value="second">秒</Option>
    </Select>
  );

  const getUnit = (unit: unitImp): string => {
    switch (unit) {
      case "day":
        return "天";
      case "hour":
        return "小时";
      case "minute":
        return "分钟";
      case "second":
        return "秒";
    }
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
            <div>access token 列表</div>
          </div>
          <Button
            onClick={() => setIsModalVisible(true)}
            style={{ fontSize: 12 }}
            size="small"
          >
            创建 access token
          </Button>
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
          endMessage={<Divider plain>老铁，别翻了，到底了！</Divider>}
          scrollableTarget="scrollableDiv"
        >
          <List
            dataSource={data}
            renderItem={(item: pb.types.AccessTokenModel) => (
              <List.Item key={item.token} className="access-token__list-item">
                <List.Item.Meta
                  title={
                    <div style={{ textDecoration: "null" }}>
                      {item.usage}
                      {item.is_deleted && (
                        <Tag style={{ marginLeft: 5 }} color="red">
                          已失效
                        </Tag>
                      )}
                    </div>
                  }
                  description={
                    <div
                      className={classNames({
                        "access-token__list-item--deleted": item.is_deleted,
                      })}
                    >
                      <span
                        onClick={() => copy(item.token, "已复制 token！")}
                        style={{ userSelect: "all" }}
                        className="access-token__list-item__token"
                      >
                        {item.token}
                      </span>
                      过期时间是{" "}
                      <span className="access-token__list-item__token-expire-date">
                        {dayjs(item.expired_at).format("YYYY-MM-DD HH:mm:ss")}
                      </span>
                      &nbsp;
                      {!!item.last_used_at
                        ? item.last_used_at + "使用过"
                        : "从未使用过"}
                    </div>
                  }
                />

                {!item.is_deleted && (
                  <>
                    {!item.is_expired && (
                      <Button
                        size="small"
                        type="primary"
                        style={{ marginRight: 3 }}
                      >
                        续租
                      </Button>
                    )}
                    <Popconfirm
                      title="确定要撤销 token ?"
                      okText="Yes"
                      cancelText="No"
                      onConfirm={() => {
                        Revoke({ token: item.token }).then((res) => {
                          message.success("撤销成功");
                          setData([]);
                          fetch();
                        });
                      }}
                    >
                      <Button size="small" type="dashed" danger>
                        撤销
                      </Button>
                    </Popconfirm>
                  </>
                )}
              </List.Item>
            )}
          />
        </InfiniteScroll>
      </div>
      <Modal
        width={"50%"}
        title={"创建 access token"}
        destroyOnClose
        open={isModalVisible}
        footer={null}
        onCancel={() => {
          setIsModalVisible(false);
        }}
      >
        <div style={{ width: "80%" }}>
          <Form
            name="basic"
            form={form}
            layout="horizontal"
            autoComplete="off"
            initialValues={{ expire_seconds: 7 }}
            labelCol={{ span: 4 }}
            onFinish={onFinish}
            onFinishFailed={onFinishFailed}
          >
            <Form.Item
              label="有效期"
              name="expire_seconds"
              rules={[
                {
                  required: true,
                  message: "有效期必填",
                },
                {
                  min: 1,
                  type: "number",
                  message: `有效期不能小于 1 ${getUnit(unit)}`,
                },
              ]}
            >
              <InputNumber addonAfter={selectAfter} />
            </Form.Item>
            <Form.Item
              label="用途"
              name="usage"
              rules={[
                {
                  required: true,
                  max: 30,
                  message: "请输入 token 的用途，不超过 30 字",
                },
              ]}
            >
              <Input.TextArea showCount />
            </Form.Item>

            <Form.Item wrapperCol={{ offset: 4, span: 16 }}>
              <Button type="primary" htmlType="submit">
                提交
              </Button>
            </Form.Item>
          </Form>
        </div>
      </Modal>
    </Card>
  );
};

export default memo(AccessTokenManager);
