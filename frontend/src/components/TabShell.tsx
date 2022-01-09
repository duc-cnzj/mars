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
import { message, Radio, Tag, Upload, Button } from "antd";
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

const TabShell: React.FC<{
  detail: pb.ProjectShowResponse;
  resizeAt: number;
  updatedAt: any;
}> = ({ detail, resizeAt, updatedAt }) => {
  const [list, setList] = useState<pb.ProjectPodLog[]>([]);
  const [sessionId, setSessionId] = useState<string>("");
  const [value, setValue] = useState<string>("");
  const [term, setTerm] = useState<Terminal>();
  const [timestamp, setTimestamp] = useState(new Date().getTime());
  const fitAddon = useMemo(() => new FitAddon(), []);

  const ref = useRef<HTMLDivElement>(null);
  const sessions = useSelector(selectSessions);
  const ws = useWs();
  const wsReady = useWsReady();

  let sname = useMemo(
    () => detail.namespace?.name + "|" + value,
    [detail, value]
  );

  const listContainer = useCallback(
    () =>
      allPodContainers({
        project_id: detail.id,
      }).then((res) => {
        setList(res.data.data);
        return res;
      }),
    [detail.id]
  );

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
      let s = pb.TerminalMessageInput.encode({
        type: pb.Type.HandleExecShellMsg,
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
    (frame: pb.TerminalMessage, term: Terminal) => {
      if (!term) {
        return;
      }
      if (frame.op === "stdout") {
        term.write(frame.data);
      }

      if (frame.op === "toast") {
        message.error(frame.data);
        listContainer();
      }
    },
    [listContainer]
  );

  const onTerminalResize = useCallback((id: string, ws: WebSocket) => {
    return ({ cols, rows }: { cols: number; rows: number }) => {
      let s = pb.TerminalMessageInput.encode({
        type: pb.Type.HandleExecShellMsg,
        message: new pb.TerminalMessage({
          session_id: id,
          op: "resize",
          cols: cols,
          rows: rows,
        }),
      }).finish();
      ws?.send(s);
    };
  }, []);

  const handleCloseShell = useCallback(
    (id: string) => {
      if (id) {
        let s = pb.TerminalMessageInput.encode({
          type: pb.Type.HandleCloseShell,
          message: new pb.TerminalMessage({ session_id: id }),
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
      let first = res.data.data[0];
      setValue(first.pod_name + "|" + first.container_name);
    });
  }, [updatedAt, listContainer]);

  const getTerm = useCallback(
    (id: string, ws: WebSocket) => {
      let myterm = new Terminal({
        fontSize: 14,
        fontFamily: '"Fira code", "Fira Mono", monospace',
        bellStyle: "sound",
        cursorBlink: true,
        cols: 106,
      });
      myterm.loadAddon(fitAddon);
      myterm.onResize(onTerminalResize(id, ws));
      myterm.onData(onTerminalSendString(id, ws));
      myterm.onKey((e: { key: string; domEvent: KeyboardEvent }) => {
        console.log(e.key);
      });
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
    let ss = pb.WsHandleExecShellInput.encode({
      type: pb.Type.HandleExecShell,
      namespace: detail.namespace?.name || "",
      pod: s[0],
      container: s[1],
    }).finish();
    sendMsg(ss);
  }, [value, detail.namespace?.name, sendMsg]);

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
            namespace: detail.namespace?.name || "",
            pod: s[0],
          }).then((res) => {
            if (res.data.running) {
              initShell();
            } else {
              // message.error(`容器列表有更新，请重试！`);
              listContainer().then((res) => {
                let first = res.data.data[0];
                setValue(first.pod_name + "|" + first.container_name);
              });
            }
          });
        }
        return e.target.value;
      });
    },
    [detail.namespace?.name, initShell, listContainer]
  );

  const beforeUpload = useCallback((file: any) => {
    const isLt50M = file.size / 1024 / 1024 <= 50;
    if (!isLt50M) {
      message.error("文件最大不能超过 50MB!");
    }
    setLoading(true);

    return isLt50M;
  }, []);

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
      console.log(info);
      if (info.file.status !== "uploading") {
        console.log(info.file, info.fileList);
      }
      if (info.file.status === "done") {
        let [pod, container] = value.split("|");
        copyToPod({
          pod: pod,
          container: container,
          namespace: detail.namespace?.name || "",
          file_id: info.file.response.id,
        })
          .then((res) => {
            console.log(res);
            message.success(
              `文件 ${info.file.name} 已上传到容器 ${res.data.podFilePath} 下`,
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
      <Radio.Group value={value} style={{ marginBottom: 5 }}>
        {list.map((item) => (
          <Radio
            onClick={reconnect}
            key={item.pod_name + "|" + item.container_name}
            value={item.pod_name + "|" + item.container_name}
          >
            {item.container_name}
            <Tag color="magenta" style={{ marginLeft: 10 }}>
              {item.pod_name}
            </Tag>
          </Radio>
        ))}
      </Radio.Group>

      {value.length > 0 && term ? (
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
            namespace={detail.namespace?.name || ""}
            pod={value.split("|")[0]}
            timestamp={timestamp}
          />
        </div>
      ) : (
        <></>
      )}
      <div ref={ref} id="terminal" style={{ height: "100%" }}></div>
    </div>
  );
};

export default memo(TabShell);
