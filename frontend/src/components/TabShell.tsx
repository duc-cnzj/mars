import React, {
  useMemo,
  useRef,
  useEffect,
  useCallback,
  useState,
  memo,
} from "react";
import { useSelector } from "react-redux";
import { allPodContainers, isPodRunning } from "../api/project";
import { message, Radio, Upload, Button, Empty } from "antd";
import { selectSessions } from "../store/reducers/shell";
import { debounce } from "lodash";
import { Terminal } from "xterm";
import { FitAddon } from "xterm-addon-fit";
import "xterm/css/xterm.css";
import { useWs, useWsReady } from "../contexts/useWebsocket";
import { UploadOutlined } from "@ant-design/icons";
import pb from "../api/compiled";
import { copyToPod } from "../api/cp";
import PodMetrics from "./PodMetrics";
import { getToken } from "../utils/token";
import { maxUploadSize } from "../api/file";
import { selectPodEventProjectID } from "../store/reducers/podEventWatcher";
import PodStateTag from "./PodStateTag";

const TabShell: React.FC<{
  namespaceID: number;
  namespace: string;
  id: number;
  resizeAt: number;
  updatedAt: any;
}> = ({ namespaceID, namespace, id, resizeAt, updatedAt }) => {
  const [list, setList] = useState<pb.types.StateContainer[]>([]);
  const [sessionId, setSessionId] = useState<string>("");
  const [value, setValue] = useState<string>("");
  const [term, setTerm] = useState<Terminal>();
  const [timestamp, setTimestamp] = useState(new Date().getTime());
  const fitAddon = useMemo(() => new FitAddon(), []);
  const [maxUploadInfo, setMaxUploadInfo] = useState({
    bytes: 0,
    humanizeSize: "",
  });

  const ref = useRef<HTMLDivElement>(null);
  const sessions = useSelector(selectSessions);
  const ws = useWs();
  const wsReady = useWsReady();

  let sname = useMemo(() => namespace + "|" + value, [namespace, value]);

  const listContainer = useCallback(
    () =>
      allPodContainers({
        project_id: id,
      }).then((res) => {
        setList(res.data.items);
        return res;
      }),
    [id]
  );

  const projectIDStr = useSelector(selectPodEventProjectID);

  useEffect(() => {
    let d = debounce(() => {
      listContainer();
    }, 2000);
    console.log("ns event: ", projectIDStr);
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
    if (
      list.length > 0 &&
      !list.map((v) => v.pod + "|" + v.container).includes(value)
    ) {
      setValue(list[0].pod + "|" + list[0].container);
    }
  }, [list, value]);

  useEffect(() => {
    maxUploadSize().then(({ data }) => {
      setMaxUploadInfo({
        bytes: data.bytes,
        humanizeSize: data.humanize_size,
      });
    });
  }, []);

  const sendMsg = useCallback(
    (msg: any) => {
      try {
        ws?.send(msg);
      } catch (e) {
        console.log(e);
      }
    },
    [ws]
  );

  const onTerminalSendString = useCallback((id: string, ws: WebSocket) => {
    return (str: string) => {
      let s = pb.websocket.TerminalMessageInput.encode({
        type: pb.websocket.Type.HandleExecShellMsg,
        message: {
          session_id: id,
          op: "stdin",
          data: str,
          cols: 0,
          rows: 0,
        },
      }).finish();
      ws?.send(s);
    };
  }, []);

  const debouncedFit_ = useCallback(
    () =>
      debounce(() => {
        try {
          fitAddon.fit();
        } catch (e) {
          console.log(e);
        }
      }, 300)(),
    [fitAddon]
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
        message.error(frame.data);
        listContainer().then((res) => {
          if (res.data.items.length> 0) {
            let first = res.data.items[0];
            setValue(first.pod + "|" + first.container);
          }
        });
      }
    },
    [listContainer]
  );

  const onTerminalResize = useCallback((id: string, ws: WebSocket) => {
    return debounce(({ cols, rows }: { cols: number; rows: number }) => {
      console.log("cols, rows. onTerminalResize");
      let s = pb.websocket.TerminalMessageInput.encode({
        type: pb.websocket.Type.HandleExecShellMsg,
        message: new pb.websocket.TerminalMessage({
          session_id: id,
          op: "resize",
          cols: cols,
          rows: rows,
        }),
      }).finish();
      ws?.send(s);
    }, 200);
  }, []);

  const handleCloseShell = useCallback(
    (id: string) => {
      if (id) {
        let s = pb.websocket.TerminalMessageInput.encode({
          type: pb.websocket.Type.HandleCloseShell,
          message: new pb.websocket.TerminalMessage({ session_id: id }),
        }).finish();
        sendMsg(s);
      }
    },
    [sendMsg]
  );

  let logCount = useMemo(() => sessions[sname]?.logCount, [sessions, sname]);
  let log = useMemo(() => sessions[sname]?.log, [sessions, sname]);
  useEffect(() => {
    if (logCount && term) {
      handleConnectionMessage(log, term);
    }
  }, [logCount, log, handleConnectionMessage, term]);

  useEffect(() => {
    listContainer().then((res) => {
      let first = res.data.items[0];
      setValue(first.pod + "|" + first.container);
    });
  }, [updatedAt, listContainer]);

  const getTerm = useCallback(
    (id: string, ws: WebSocket) => {
      let myterm = new Terminal({
        fontSize: 14,
        fontFamily: '"Fira code", "Fira Mono", monospace',
        bellStyle: "sound",
        cursorBlink: true,
        // cols: 106,
        rows: 25,
      });

      myterm.loadAddon(fitAddon);
      myterm.onResize(onTerminalResize(id, ws));
      myterm.onData(onTerminalSendString(id, ws));
      ref.current !== null && myterm.open(ref.current);
      debouncedFit_();
      myterm.focus();

      return myterm;
    },
    [onTerminalResize, onTerminalSendString, debouncedFit_, fitAddon]
  );

  let sid = useMemo(() => sessions[sname]?.sessionID, [sessions, sname]);
  useEffect(() => {
    if (sid) {
      setSessionId(sid);
    }
  }, [sid]);

  useEffect(() => {
    if (wsReady && sessionId && ws) {
      const t = getTerm(sessionId, ws);
      setTerm(t);

      return () => {
        t.dispose();
        handleCloseShell(sessionId);
        console.log("close id: ", sessionId);
      };
    }
  }, [wsReady, sessionId, handleCloseShell, setTerm, ws, getTerm]);

  useEffect(() => {
    debouncedFit_();
  }, [debouncedFit_, resizeAt]);

  const initShell = useCallback(() => {
    let s = value.split("|");
    let ss = pb.websocket.WsHandleExecShellInput.encode({
      type: pb.websocket.Type.HandleExecShell,
      container: {
        namespace: namespace,
        pod: s[0],
        container: s[1],
      },
    }).finish();
    sendMsg(ss);
  }, [value, namespace, sendMsg]);

  useEffect(() => {
    if (value && wsReady) {
      initShell();
    }
  }, [initShell, value, wsReady]);

  const reconnect = useCallback(
    (e: any) => {
      setTimestamp(new Date().getTime());
      setValue((v) => {
        if (v === e.target.value) {
          let s = (e.target.value as string).split("|");
          isPodRunning({
            namespace: namespace,
            pod: s[0],
          }).then((res) => {
            if (res.data.running) {
              initShell();
            } else {
              // message.error(res.data.reason);
              listContainer().then((res) => {
                let first = res.data.items[0];
                setValue(first.pod + "|" + first.container);
              });
            }
          });
        }
        return e.target.value;
      });
    },
    [namespace, initShell, listContainer]
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
    [maxUploadInfo]
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
      if (info.file.status === "done") {
        let [pod, container] = value.split("|");
        copyToPod({
          pod: pod,
          container: container,
          namespace: namespace,
          file_id: info.file.response.id,
        })
          .then((res) => {
            console.log(res);
            message.success(
              `文件 ${info.file.name} 已上传到容器 ${res.data.pod_file_path} 下`,
              2
            );
          })
          .catch((e) => {
            message.error(`文件 ${info.file.name} 上传到容器失败`);
          })
          .finally(() => {
            setLoading(false);
          });
      } else if (info.file.status === "error") {
        message.error(`文件 ${info.file.name} 上传失败`);
        setLoading(false);
      }
    },
  };

  return (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        height: "100%",
        overflowY: "auto",
      }}
    >
      {list.length > 0 ? (
        <>
          <Radio.Group value={value} style={{ marginBottom: 5 }}>
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

          {value.length > 0 && term && (
            <>
              <div style={{ display: "flex", justifyContent: "start" }}>
                <Upload {...props}>
                  <Button
                    disabled={loading}
                    loading={loading}
                    size="small"
                    style={{ fontSize: 12, marginRight: 5, margin: "5px 0" }}
                    icon={<UploadOutlined />}
                  >
                    {loading ? "上传中" : "上传到容器"}
                  </Button>
                </Upload>
                <PodMetrics
                  namespace={namespace}
                  pod={value.split("|")[0]}
                  timestamp={timestamp}
                />
              </div>
            </>
          )}

          <div ref={ref} id="terminal" style={{ height: "100%" }} />
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
    </div>
  );
};

export default memo(TabShell);
