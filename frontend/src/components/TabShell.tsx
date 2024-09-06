import React, {
  useMemo,
  useRef,
  useEffect,
  useCallback,
  useState,
  memo,
} from "react";
import { useSelector } from "react-redux";
import { message, Radio, Upload, Button, Empty, Input, Popconfirm } from "antd";
import { selectSessions } from "../store/reducers/shell";
import { debounce } from "lodash";
import { Terminal } from "xterm";
import { FitAddon } from "xterm-addon-fit";
import "xterm/css/xterm.css";
import { useWs, useWsReady } from "../contexts/useWebsocket";
import {
  CloseOutlined,
  MinusOutlined,
  UploadOutlined,
} from "@ant-design/icons";
import pb from "../api/websocket";
import PodMetrics from "./PodMetrics";
import { getToken } from "../utils/token";
import { selectPodEventProjectID } from "../store/reducers/podEventWatcher";
import PodStateTag from "./PodStateTag";
import { v4 as uuidv4 } from "uuid";
import { Allotment } from "allotment";
import "allotment/dist/style.css";
import { css } from "@emotion/css";
import "../styles/allotment-overwrite.css";
import _ from "lodash";
import ajax from "../api/ajax";
import { components } from "../api/schema";
import JsFileDownloader from "js-file-downloader";

const encoder = new TextEncoder();
const decoder = new TextDecoder();

const generateSessionID = (
  namespace: string,
  pod: string,
  container: string,
): string => {
  return `${namespace}-${pod}-${container}:${uuidv4()}`;
};

interface NPC {
  namespace: string;
  pod: string;
  container: string;
}

const TabShell: React.FC<{
  namespaceID: number;
  namespace: string;
  id: number;
  resizeAt: number;
  updatedAt: any;
}> = ({ namespaceID, namespace, id, resizeAt, updatedAt }) => {
  console.log("render: TabShell");
  const [list, setList] = useState<
    components["schemas"]["types.StateContainer"][]
  >([]);
  const [value, setValue] = useState<NPC | null>();
  const [timestamp, setTimestamp] = useState(new Date().getTime());
  const [maxUploadInfo, setMaxUploadInfo] = useState({
    bytes: 0,
    humanizeSize: "",
  });
  const ws = useWs();
  const wsReady = useWsReady();
  const [termMap, setTermMap] = useState<
    { type: "vertical" | "horizontal" | undefined; id: string }[]
  >([{ type: undefined, id: uuidv4() }]);

  const [resizeTime, setResizeTime] = useState(resizeAt);
  useEffect(() => {
    setResizeTime(resizeAt);
  }, [resizeAt]);

  const projectIDStr = useSelector(selectPodEventProjectID);

  const resetTermMap = useCallback(() => {
    setTermMap([{ type: undefined, id: uuidv4() }]);
  }, []);

  const listContainer = useCallback(
    () =>
      ajax
        .GET("/api/projects/{id}/containers", { params: { path: { id: id } } })
        .then(({ data, error }) => {
          if (error) {
            return;
          }
          setList(data.items);
          return data;
        }),
    [id],
  );

  const setValuesByResult = useCallback(
    (items: components["schemas"]["types.StateContainer"][]) => {
      if (items.length > 0) {
        let first = items[0];
        setValue({
          namespace,
          pod: first.pod,
          container: first.container,
        });
      } else {
        setValue(null);
      }
    },
    [namespace, setValue],
  );

  useEffect(() => {
    let d = debounce(() => {
      listContainer();
    }, 2000);
    console.log("ns event: ", projectIDStr, id);
    if (projectIDStr.split("-").length === 2) {
      let pid = Number(projectIDStr.split("-")[1]);
      if (pid === Number(id)) {
        d();
      }
    }
    return () => {
      d.cancel();
    };
  }, [projectIDStr, listContainer, id]);

  useEffect(() => {
    if (list.length > 0) {
      if (!value) {
        const first = list[0];
        setValue({ namespace, pod: first.pod, container: first.container });
        return;
      }
      if (!list.map((v) => v.pod).includes(value.pod)) {
        const first = list[0];
        setValue({ namespace, pod: first.pod, container: first.container });
        return;
      }
    }
    if (list.length === 0 && !!value && value.container !== "") {
      setValue(null);
    }
  }, [list, value, namespace]);

  useEffect(() => {
    ajax.GET("/api/files/max_upload_size").then(({ data }) => {
      data &&
        setMaxUploadInfo({
          bytes: data.bytes,
          humanizeSize: data.humanizeSize,
        });
    });
  }, []);

  useEffect(() => {
    listContainer().then((data) => {
      data && setValuesByResult(data.items);
    });
  }, [listContainer, setValuesByResult, updatedAt]);

  const [forceRender, setForceRender] = useState<any>(null);
  const reconnect = useCallback(
    (e: any) => {
      resetTermMap();
      setTimestamp(new Date().getTime());
      setValue((v) => {
        let [pod, container] = (e.target.value as string).split("|");
        if (v?.pod === pod && v.container === container) {
          ajax
            .POST("/api/containers/pod_running_status", {
              body: {
                namespace,
                pod,
              },
            })
            .then(({ data, error }) => {
              if (error) {
                return;
              }
              if (data.running) {
                setForceRender(new Date().getTime());
              } else {
                listContainer().then((data) => {
                  data && setValuesByResult(data.items);
                });
              }
            });
        }
        return { pod, container, namespace };
      });
    },
    [namespace, listContainer, setValuesByResult, resetTermMap],
  );

  const beforeUpload = useCallback(
    (file: any) => {
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
    [maxUploadInfo],
  );

  const [loading, setLoading] = useState(false);

  const props = {
    name: "file",
    beforeUpload: beforeUpload,
    action: process.env.REACT_APP_BASE_URL + "/api/files",
    headers: {
      authorization: getToken(),
    },
    showUploadList: false,
    onChange(info: any) {
      if (info.file.status !== "uploading") {
        console.log(info.file, info.fileList);
      }
      if (!!value && info.file.status === "done") {
        let pod = value.pod;
        let container = value.container;
        ajax
          .POST("/api/containers/copy_to_pod", {
            body: {
              pod: pod,
              container: container,
              namespace: namespace,
              fileId: info.file.response.id,
            },
          })
          .then(({ data, error }) => {
            setLoading(false);
            if (error) {
              message.error(`文件 ${info.file.name} 上传到容器失败`);
              return;
            }
            console.log(data);
            message.success(
              `文件 ${info.file.name} 已上传到容器 ${data?.podFilePath} 下`,
              2,
            );
          });
      } else if (info.file.status === "error") {
        message.error(`文件 ${info.file.name} 上传失败`);
        setLoading(false);
      }
    },
  };

  const canAddTerm = useCallback(() => termMap.length >= 4, [termMap]);

  const resizeShellWindow = useCallback(() => {
    setResizeTime(new Date().getTime());
  }, []);

  useEffect(() => {
    if (termMap.length > 0) {
      console.log("resizeShellWindow");
      resizeShellWindow();
    }
  }, [termMap, resizeShellWindow]);

  const addWebTerm = useCallback(
    (type: "vertical" | "horizontal") => {
      console.log("add web term");
      setTermMap((tmap) => {
        if (canAddTerm()) {
          message.error("不能超过四个分屏");
          return tmap;
        }
        tmap.push({ type: type, id: uuidv4() });
        return [...tmap];
      });
    },
    [canAddTerm],
  );
  const subWebTerm = useCallback((id: string) => {
    console.log("sub web term");
    setTermMap((tmap) => {
      if (_.keys(tmap).length <= 1) {
        message.error("至少一个屏幕");
        return tmap;
      }

      let newTerms = tmap.filter((v) => v.id !== id);
      if (newTerms.length === 1) {
        return [{ id: newTerms[0].id, type: undefined }];
      }
      return [...newTerms];
    });
  }, []);

  const nestedAllotment = (
    items: { type: "vertical" | "horizontal" | undefined; id: string }[],
    ws: WebSocket,
    value: { pod: string; namespace: string; container: string },
  ): React.ReactNode[] => {
    let idx = 0;
    let group: React.ReactNode[] = [];
    let groups: React.ReactNode[] = [];
    let lastType: "vertical" | "horizontal" | undefined;
    for (let index = 0; index < items.length; index++) {
      let element = items[index];
      group.push(
        <ShellWindow
          id={element.id}
          key={element.id}
          ws={ws}
          canClose={termMap.length > 1}
          namespace={namespace}
          pod={value.pod}
          container={value.container}
          resizeAt={resizeTime}
          forceRender={forceRender}
          onClose={subWebTerm}
        />,
      );
      lastType = items[index].type;
      if (
        idx !== 0 &&
        index + 1 < items.length &&
        items[index].type !== items[index + 1].type
      ) {
        groups.push(
          <Allotment
            key={groups.length + 1}
            vertical={lastType === "vertical"}
            onDragEnd={resizeShellWindow}
          >
            {group.map((v) => v)}
          </Allotment>,
        );
        group = [];
        idx = 0;
        continue;
      }
      idx++;
    }
    if (group.length > 0) {
      if (group.length === 1 && items.length > 1) {
        groups.push(group[0]);
      } else {
        groups.push(
          <Allotment
            key={groups.length + 1}
            vertical={lastType === "vertical"}
            onDragEnd={resizeShellWindow}
          >
            {group.map((v) => v)}
          </Allotment>,
        );
      }
      group = [];
    }
    return [...groups];
  };

  const [isModalOpen, setIsModalOpen] = useState(false);
  const [podfilepath, setPodfilepath] = useState("");

  return (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        height: "100%",
        overflowY: "auto",
      }}
    >
      {list.length > 0 && value ? (
        <>
          <Radio.Group
            value={`${value.pod}|${value.container}`}
            style={{ marginBottom: 5 }}
          >
            {list.map((item) => (
              <Radio
                onClick={reconnect}
                key={item.pod + "|" + item.container}
                value={item.pod + "|" + item.container}
              >
                {item.container}
                <PodStateTag pod={item} />
              </Radio>
            ))}
          </Radio.Group>

          {!!value && (
            <div
              style={{
                display: "flex",
                alignItems: "center",
                justifyContent: "start",
              }}
            >
              <div
                style={{
                  display: "flex",
                  alignItems: "center",
                  marginBottom: 5,
                }}
              >
                <Upload {...props}>
                  <Button
                    disabled={loading}
                    loading={loading}
                    size="small"
                    style={{ fontSize: 12, marginRight: 2 }}
                    icon={<UploadOutlined />}
                  >
                    {loading ? "上传中" : "上传到容器"}
                  </Button>
                </Upload>
                <Popconfirm
                  overlayClassName="copyfrompod"
                  title="下载文件(绝对路径)"
                  open={isModalOpen}
                  description={
                    <Input
                      value={podfilepath}
                      onChange={(v) => {
                        setPodfilepath(v.target.value);
                      }}
                      style={{ width: "80%", fontSize: 12 }}
                      placeholder="绝对路径"
                    />
                  }
                  onConfirm={() => {
                    new JsFileDownloader({
                      url: `${process.env.REACT_APP_BASE_URL}/api/copy_from_pod`,
                      method: "POST",
                      headers: [
                        { name: "Authorization", value: getToken() },
                        { name: "Content-Type", value: "application/json" },
                      ],
                      body: JSON.stringify({
                        namespace: value.namespace,
                        pod: value.pod,
                        container: value.container,
                        filepath: podfilepath,
                      }),
                    }).then(function () {
                      message.success("下载成功");
                      setIsModalOpen(false);
                    });
                  }}
                  onCancel={() => {
                    setIsModalOpen(false);
                    setPodfilepath("");
                  }}
                  overlayInnerStyle={{ width: 500 }}
                  okText="下载"
                  cancelText="取消"
                >
                  <Button
                    onClick={() => setIsModalOpen(true)}
                    style={{ fontSize: 12, marginRight: 2 }}
                    size="small"
                  >
                    下载文件
                  </Button>
                </Popconfirm>
                <div
                  style={{
                    height: "100%",
                    display: "flex",
                    alignItems: "center",
                  }}
                >
                  <Button
                    style={{
                      borderWidth: 1,
                      borderTopRightRadius: 0,
                      borderEndEndRadius: 0,
                    }}
                    className={css`
                      :hover {
                        z-index: 9999;
                      }
                    `}
                    size="small"
                    disabled={canAddTerm()}
                    icon={
                      <MinusOutlined style={{ transform: "rotate(90deg)" }} />
                    }
                    onClick={() => addWebTerm("horizontal")}
                  />
                  <Button
                    style={{
                      borderWidth: 1,
                      marginLeft: -1,
                      borderTopLeftRadius: 0,
                      borderEndStartRadius: 0,
                    }}
                    size="small"
                    disabled={canAddTerm()}
                    icon={<MinusOutlined />}
                    onClick={() => addWebTerm("vertical")}
                  />
                </div>
              </div>
              <PodMetrics
                namespace={namespace}
                pod={value.pod}
                timestamp={timestamp}
              />
            </div>
          )}
        </>
      ) : (
        <div
          style={{
            height: "100%",
            width: "100%",
            display: "flex",
            justifyContent: "center",
            alignItems: "center",
          }}
        >
          <Empty description="列表还没有任何容器" />
        </div>
      )}

      {ws && wsReady && value && (
        <Allotment onDragEnd={resizeShellWindow} vertical={true}>
          {termMap.length > 1 && termMap[1].type === "vertical" ? (
            <Allotment vertical={false} onDragEnd={resizeShellWindow}>
              {nestedAllotment(termMap, ws, value).map((v) => v)}
            </Allotment>
          ) : (
            nestedAllotment(termMap, ws, value).map((v) => v)
          )}
        </Allotment>
      )}
    </div>
  );
};

const ShellWindow: React.FC<{
  id: string;
  ws: WebSocket;
  namespace: string;
  pod: string;
  container: string;
  canClose: boolean;
  onClose?: (id: string) => void;
  resizeAt: any;
  forceRender: any;
}> = memo(
  ({
    resizeAt,
    namespace,
    pod,
    container,
    forceRender,
    ws,
    onClose,
    id,
    canClose,
  }) => {
    console.log("render: ShellWindow", forceRender);
    const ref = useRef<HTMLDivElement>(null);
    const fitAddon = useMemo(() => new FitAddon(), []);
    const [term, setTerm] = useState<Terminal>();
    const sessionID = useMemo(() => {
      let sid = generateSessionID(namespace, pod, container);
      if (forceRender > 0) {
        console.log("forceRender:", sid);
      }
      return sid;
    }, [namespace, pod, container, forceRender]);
    const sendMsg = useCallback(
      (msg: any) => {
        try {
          ws?.send(msg);
        } catch (e) {
          console.log(e);
        }
      },
      [ws],
    );
    const initShell = useCallback(() => {
      let ss = pb.websocket.WsHandleExecShellInput.encode({
        type: pb.websocket.Type.HandleExecShell,
        container: {
          namespace: namespace,
          pod: pod,
          container: container,
        },
        sessionId: sessionID,
      }).finish();
      sendMsg(ss);
    }, [namespace, pod, container, sendMsg, sessionID]);

    const onTerminalResize = useCallback((id: string, ws: WebSocket) => {
      return debounce(({ cols, rows }: { cols: number; rows: number }) => {
        console.log("cols, rows. onTerminalResize");
        let s = pb.websocket.TerminalMessageInput.encode({
          type: pb.websocket.Type.HandleExecShellMsg,
          message: new pb.websocket.TerminalMessage({
            sessionId: id,
            op: "resize",
            cols: cols,
            rows: rows,
          }),
        }).finish();
        ws?.send(s);
      }, 200);
    }, []);

    const onTerminalSendString = useCallback((id: string, ws: WebSocket) => {
      return (str: string) => {
        let s = pb.websocket.TerminalMessageInput.encode({
          type: pb.websocket.Type.HandleExecShellMsg,
          message: {
            sessionId: id,
            op: "stdin",
            data: encoder.encode(str),
            cols: 0,
            rows: 0,
          },
        }).finish();
        ws?.send(s);
      };
    }, []);

    const debouncedFit_ = useMemo(
      () =>
        debounce(() => {
          try {
            fitAddon.fit();
          } catch (e) {
            console.log(e);
          }
        }, 300),
      [fitAddon],
    );

    const handleConnectionMessage = useCallback(
      (frame: pb.websocket.TerminalMessage, term: Terminal) => {
        if (!term) {
          return;
        }
        if (frame.op === "stdout") {
          term.write(frame.data);
        }

        if (frame.op === "toast") {
          message.error(decoder.decode(frame.data));
        }
      },
      [],
    );

    const sessions = useSelector(selectSessions);
    let logCount = useMemo(
      () => sessions[sessionID]?.logCount,
      [sessions, sessionID],
    );
    let log = useMemo(() => sessions[sessionID]?.log, [sessions, sessionID]);
    useEffect(() => {
      if (logCount && term) {
        handleConnectionMessage(log, term);
      }
    }, [logCount, log, handleConnectionMessage, term]);

    useEffect(() => {
      if (!!resizeAt) {
        debouncedFit_();
      }
    }, [debouncedFit_, resizeAt]);

    const getTerm = useCallback(
      (id: string, ws: WebSocket) => {
        let myterm = new Terminal({
          fontSize: 14,
          fontFamily: '"Fira code", "Fira Mono", monospace',
          cursorBlink: true,
          // cols: 106,
          rows: 25,
        });

        myterm.loadAddon(fitAddon);
        myterm.onResize(onTerminalResize(id, ws));
        myterm.onData(onTerminalSendString(id, ws));
        ref.current && myterm.open(ref.current);
        debouncedFit_();
        myterm.focus();

        return myterm;
      },
      [onTerminalResize, onTerminalSendString, fitAddon, debouncedFit_],
    );
    const handleCloseShell = useCallback(
      (id: string) => {
        if (id) {
          let s = pb.websocket.TerminalMessageInput.encode({
            type: pb.websocket.Type.HandleCloseShell,
            message: new pb.websocket.TerminalMessage({ sessionId: id }),
          }).finish();
          sendMsg(s);
        }
      },
      [sendMsg],
    );
    useEffect(() => {
      initShell();
      if (sessionID) {
        const t = getTerm(sessionID, ws);
        setTerm(t);

        return () => {
          t.dispose();
          handleCloseShell(sessionID);
          console.log("term close id: ", sessionID);
        };
      }
    }, [sessionID, initShell, setTerm, ws, getTerm, handleCloseShell]);

    return (
      <div
        ref={ref}
        id="terminal"
        className={css`
          .xterm {
            height: 100%;
          }
        `}
        style={{
          height: "100%",
          display: !!sessionID ? "block" : "none",
          position: "relative",
        }}
      >
        {canClose && (
          <Button
            size="small"
            type="text"
            icon={<CloseOutlined style={{ color: "gray" }} />}
            onClick={() => onClose?.(id)}
            className={css`
              :hover {
                background-color: #525252 !important;
              }
            `}
            style={{ position: "absolute", top: 5, right: 3, zIndex: 1000 }}
          />
        )}
      </div>
    );
  },
);

export default memo(TabShell);
