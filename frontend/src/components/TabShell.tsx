import React, {
  useMemo,
  useRef,
  useEffect,
  useCallback,
  useState,
  memo,
} from "react";
import { useSelector } from "react-redux";
import { containerList, isPodRunning } from "../api/project";
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

const TabShell: React.FC<{
  detail: pb.ProjectShowResponse;
  resizeAt: number;
  updatedAt: any;
}> = ({ detail, resizeAt, updatedAt }) => {
  const [list, setList] = useState<pb.PodLog[]>([]);
  const [sessionId, setSessionId] = useState<string>("");
  const [value, setValue] = useState<string>("");
  const [term, setTerm] = useState<Terminal>();
  const ref = useRef<HTMLDivElement>(null);
  const ws = useWs();
  const wsReady = useWsReady();
  const [fitAddon, _] = useState(new FitAddon());
  const sessions = useSelector(selectSessions);
  let sname = useMemo(
    () => detail.namespace?.name + "|" + value,
    [detail, value]
  );

  const listContainer = useCallback(async () => {
    return containerList({
      namespace_id: detail.namespace?.id,
      project_id: detail.id,
    }).then((res) => {
      setList(res.data.data);
      return res;
    });
  }, [detail.id, detail.namespace?.id]);

  const sendMsg = useCallback(
    (msg: string) => {
      try {
        ws?.send(msg);
      } catch (e) {
        console.log(e);
      }
    },
    [ws]
  );

  const onTerminalSendString = (str: string) => {
    let re = {
      type: "handle_exec_shell_msg",
      data: JSON.stringify({
        session_id: sessionId,
        op: "stdin",
        data: str,
        cols: term?.cols,
        rows: term?.rows,
      }),
    };

    sendMsg(JSON.stringify(re));
  };
  const debouncedFit_ = debounce(() => {
    try {
      fitAddon.fit();
    } catch (e) {
      console.log(e);
    }
  }, 300);
  const handleConnectionMessage = (frame: any) => {
    if (frame.op === "stdout") {
      term?.write(frame.data);
    }

    if (frame.op === "toast") {
      message.error(frame.data);
      listContainer();
    }
  };

  const onTerminalResize = ({ cols, rows }: { cols: number; rows: number }) => {
    let re = {
      type: "handle_exec_shell_msg",
      data: JSON.stringify({
        session_id: sessionId,
        op: "resize",
        cols: cols,
        rows: rows,
      }),
    };
    sendMsg(JSON.stringify(re));
  };

  const handleCloseShell = useCallback(() => {
    if (sessionId) {
      let re = {
        type: "handle_close_shell",
        data: JSON.stringify({
          session_id: sessionId,
        }),
      };
      sendMsg(JSON.stringify(re));
    }
    console.log("closed closed closed closed");
  }, [sessionId]);

  useEffect(() => {
    return () => {
      handleCloseShell();
    };
  }, [handleCloseShell]);

  useEffect(() => {
    // 关闭上一个连接如果有的话
    console.log("handle_close_shell");
    handleCloseShell();
    if (sessions[sname]) {
      setSessionId(sessions[sname].sessionID);
    }
  }, [sessions[sname]?.sessionID]);

  useEffect(() => {
    if (sessions[sname] && sessions[sname].log !== undefined) {
      handleConnectionMessage(JSON.parse(sessions[sname].log));
    }
  }, [sessions[sname]?.logCount]);

  useEffect(() => {
    listContainer().then((res) => {
      let first = res.data.data[0];
      setValue(first.pod_name + "|" + first.container_name);
    });
  }, [updatedAt, listContainer]);

  const initTerm = () => {
    if (term) {
      term.dispose();
    }
    let myterm = new Terminal({
      fontSize: 14,
      fontFamily: '"Fira code", "Fira Mono", monospace',
      bellStyle: "sound",
      cursorBlink: true,
    });
    setTerm(myterm);
    myterm.loadAddon(fitAddon);
    myterm.onResize(onTerminalResize);
    myterm.onData(onTerminalSendString);
    myterm.onKey((event: any) => {
      console.log(event);
    });

    ref.current !== null && myterm.open(ref.current);
    debouncedFit_();
    myterm.focus();
  };

  useEffect(() => {
    if (!wsReady || !sessionId) {
      console.log(ws, wsReady);
      return;
    }

    initTerm();
  }, [value, wsReady, sessionId]);

  useEffect(() => {
    debouncedFit_();
  }, [resizeAt, debouncedFit_]);

  const initShell = useCallback(() => {
    let s = value.split("|");
    let re = {
      type: "handle_exec_shell",
      data: JSON.stringify({
        namespace: detail.namespace?.name,
        pod: s[0],
        container: s[1],
      }),
    };
    sendMsg(JSON.stringify(re));
  }, [value, detail.namespace?.name, sendMsg]);
  useEffect(() => {
    if (value && wsReady) {
      initShell();
    }
  }, [initShell, value, wsReady]);

  const reconnect = (e: any) => {
    setValue(e.target.value);
    let s = (e.target.value as string).split("|");
    isPodRunning({ namespace: detail.namespace?.name || "", pod: s[0] }).then(
      (res) => {
        if (res.data.running) {
          initShell();
        } else {
          // message.error(`容器列表有更新，请重试！`);
          listContainer().then((res) => {
            let first = res.data.data[0];
            setValue(first.pod_name + "|" + first.container_name);
          });
        }
      }
    );
  };

  function beforeUpload(file: any) {
    const isLt50M = file.size / 1024 / 1024 <= 50;
    if (!isLt50M) {
      message.error("文件最大不能超过 50MB!");
    }
    setLoading(true);

    return isLt50M;
  }

  const [loading, setLoading] = useState(false);

  const props = {
    name: "file",
    beforeUpload: beforeUpload,
    action: process.env.REACT_APP_BASE_URL + "/api/files",
    // TODO: 需要认证
    headers: {
      // authorization: "authorization-text",
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
      <Radio.Group
        // onChange={onChange}
        value={value}
        style={{ marginBottom: 10 }}
      >
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
      <Upload {...props}>
        <Button
          disabled={loading}
          loading={loading}
          size="small"
          style={{ fontSize: 12, marginBottom: 5 }}
          icon={<UploadOutlined />}
        >
          {loading ? "上传中" : "上传到容器"}
        </Button>
      </Upload>
      <div ref={ref} id="terminal" style={{ height: "100%" }}></div>
    </div>
  );
};

export default memo(TabShell);
