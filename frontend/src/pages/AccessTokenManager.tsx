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
import InfiniteScroll from "react-infinite-scroll-component";
import { useForm } from "antd/es/form/Form";
import { copy } from "../utils/copy";
import styled from "@emotion/styled";
import theme from "../styles/theme";
import ajax from "../api/ajax";
import { components } from "../api/schema";

const defaultPageSize = 15;

const { Option } = Select;

type unitImp = "day" | "hour" | "minute" | "second" | "month";

const AccessTokenManager: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [paginate, setPaginate] = useState<{
    page: number;
    page_size: number;
    count: number;
  }>({ page: 0, page_size: defaultPageSize, count: 0 });
  const [data, setData] = useState<
    components["schemas"]["types.AccessTokenModel"][]
  >([]);

  const loadMoreData = () => {
    if (loading) {
      return;
    }
    setLoading(true);
    ajax
      .GET("/api/access_tokens", {
        params: {
          query: {
            page: paginate.page + 1,
            pageSize: paginate.page_size,
          },
        },
      })
      .then(({ data: res }) => {
        res && setData((data) => [...data, ...res.items]);
        res &&
          setPaginate({
            page: res.page,
            page_size: res.pageSize,
            count: res.count,
          });
        setLoading(false);
      })
      .catch((e) => {
        message.error(e.repo);
        setLoading(false);
      });
  };

  const scrollDiv = useRef<HTMLDivElement>(null);
  const fetch = useCallback(() => {
    if (scrollDiv.current) {
      scrollDiv.current.scrollTo(0, 0);
    }
    ajax
      .GET("/api/access_tokens", {
        params: { query: { page: 1, pageSize: defaultPageSize } },
      })

      .then(({ data }) => {
        data && setData(data.items);
        data &&
          setPaginate({
            page: data.page,
            page_size: data.pageSize,
            count: data.count,
          });
      })
      .catch((e) => {
        console.log(e);
        // message.error(e.reponse);
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
  const getSeconds = (num: number): number => {
    switch (unit) {
      case "month":
        const now = new Date();
        return dayjs(now).add(num, "months").diff(now, "seconds");
      case "day":
        return 24 * 60 * 60 * num;
      case "hour":
        return 60 * 60 * num;
      case "minute":
        return 60 * num;
      case "second":
        return num;
    }
  };

  const onFinish = (values: any) => {
    ajax
      .POST("/api/access_tokens", {
        body: {
          usage: values.usage,
          expireSeconds: getSeconds(values.unit_number),
        },
      })
      .then((res) => {
        message.success("创建成功");
        setIsModalVisible(false);
        form.resetFields();
        setData([]);
        fetch();
      })
      .catch((e) => {
        console.log(e);
        // message.error(e.response.data.message);
      });
  };
  const onLease = (values: any) => {
    ajax
      .PUT("/api/access_tokens/{token}", {
        body: {
          expireSeconds: getSeconds(values.unit_number),
          token: currToken,
        },
        params: { path: { token: currToken } },
      })
      .then((res) => {
        message.success("续租成功");
        setIsModalVisible(false);
        form.resetFields();
        setCurrToken("");
        setData([]);
        fetch();
      })
      .catch((e) => {
        message.error(e.response.data.message);
      });
  };
  const [unit, setUnit] = useState<unitImp>("day");

  const selectAfter = (
    <Select
      defaultValue={unit}
      onChange={(v: unitImp) => setUnit(v)}
      style={{ width: 80 }}
    >
      <Option value="month">月</Option>
      <Option value="day">天</Option>
      <Option value="hour">小时</Option>
      <Option value="minute">分钟</Option>
      <Option value="second">秒</Option>
    </Select>
  );

  const getUnit = (unit: unitImp): string => {
    switch (unit) {
      case "month":
        return "月";
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

  const [currToken, setCurrToken] = useState("");
  const getTags = (
    item: components["schemas"]["types.AccessTokenModel"]
  ): React.ReactNode => {
    return (
      <>
        {!item.isDeleted && (
          <>
            {dayjs(item.expiredAt).isBefore(dayjs(new Date()).add(1, "day")) &&
              dayjs(item.expiredAt).isAfter(dayjs(new Date())) && (
                <Tag style={{ marginLeft: 5 }} color="#facc15">
                  即将过期
                </Tag>
              )}
            {item.isExpired && (
              <Tag style={{ marginLeft: 5 }} color="#a1a1aa">
                已过期
              </Tag>
            )}
          </>
        )}
        {item.isDeleted && (
          <Tag style={{ marginLeft: 5 }} color="red">
            已撤销
          </Tag>
        )}
      </>
    );
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
            <div>访问令牌列表</div>
          </div>
          <Button
            onClick={() => setIsModalVisible(true)}
            style={{ fontSize: 12 }}
            size="small"
          >
            创建访问令牌
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
            renderItem={(
              item: components["schemas"]["types.AccessTokenModel"]
            ) => (
              <AccessTokenListItem key={item.token}>
                <List.Item.Meta
                  title={
                    <div style={{ textDecoration: "null" }}>
                      {item.usage}
                      {getTags(item)}
                    </div>
                  }
                  description={
                    <ItemWrapper lineThrough={item.isDeleted || item.isExpired}>
                      <TokenSpan
                        onClick={() => copy(item.token, "已复制 token！")}
                        style={{ userSelect: "all" }}
                      >
                        {item.token}
                      </TokenSpan>
                      过期时间是{" "}
                      <TokenExpireDateSpan>
                        {dayjs(item.expiredAt).format("YYYY-MM-DD HH:mm:ss")}
                      </TokenExpireDateSpan>
                      &nbsp;
                      {!!item.lastUsedAt
                        ? item.lastUsedAt + "使用过"
                        : "从未使用过"}
                    </ItemWrapper>
                  }
                />

                {!item.isDeleted && !item.isExpired && (
                  <>
                    <Button
                      size="small"
                      type="primary"
                      style={{ marginRight: 3 }}
                      onClick={() => {
                        setCurrToken(item.token);
                        setIsModalVisible(true);
                      }}
                    >
                      续租
                    </Button>
                    <Popconfirm
                      title="确定要撤销 token ?"
                      okText="Yes"
                      cancelText="No"
                      onConfirm={() => {
                        ajax
                          .DELETE("/api/access_tokens/{token}", {
                            params: {
                              path: {
                                token: item.token,
                              },
                            },
                          })
                          .then((res) => {
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
              </AccessTokenListItem>
            )}
          />
        </InfiniteScroll>
      </div>
      <Modal
        width={"50%"}
        title={`${!currToken ? "创建" : "续租"}访问令牌`}
        destroyOnClose
        open={isModalVisible}
        footer={null}
        onCancel={() => {
          setIsModalVisible(false);
          setUnit("day");
          form.resetFields();
        }}
      >
        <div style={{ width: "80%" }}>
          <Form
            name="basic"
            form={form}
            layout="horizontal"
            autoComplete="off"
            initialValues={{ unit_number: 7 }}
            labelCol={{ span: 4 }}
            onFinish={(values: any) => {
              !currToken ? onFinish(values) : onLease(values);
            }}
          >
            <Form.Item
              label="有效期"
              name="unit_number"
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
            {!currToken && (
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
            )}

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

const TokenSpan = styled.span`
  background-color: #1b1f230d;
  padding: 3px;
  color: #fb7185;
  font-family: "Monaco", monospace;
  border-radius: 5px;
  margin-right: 5px;
`;

const TokenExpireDateSpan = styled.span`
  background-color: #1b1f230d;
  padding: 3px;
  color: #fb7185;
  font-family: "Monaco", monospace;
  border-radius: 5px;
  margin-right: 5px;
`;

const ItemWrapper = styled.div<{ lineThrough: boolean }>`
  text-decoration: ${({ lineThrough }) =>
    lineThrough ? "line-through" : "none"};
  text-decoration-thickness: 2px;
  text-decoration-color: black;
`;

const AccessTokenListItem = styled(List.Item)`
  padding: 14px 24px !important;
  &:hover {
    background-image: linear-gradient(to right, ${theme.lightColor}, #ffffff);
  }
`;
