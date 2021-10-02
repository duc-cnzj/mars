import React, {useMemo, useRef, useEffect, useCallback, useState, memo } from "react";
import { useSelector } from "react-redux";
import { containerList, PodContainerItem, ProjectDetail } from "../api/project";
import { message, Radio, Tag , RadioChangeEvent} from "antd";
import { selectSessions } from "../store/reducers/shell";
import { debounce } from "lodash";
import { Terminal } from "xterm";
import { FitAddon } from "xterm-addon-fit";
import "xterm/css/xterm.css";
import { useWs, useWsReady } from "../contexts/useWebsocket";

// class Shell extends Component<{
//   detail: ProjectDetail;
//   resizeAt: number;
// }> {
//   private connecting_: boolean = false;
//   private connectionClosed_: boolean = false;
//   private conn_: WebSocket | null = null;
//   private connected_ = false;
//   private term: Terminal | null = null;
//   private debouncedFit_: Function | null = null;
//   private connSubject_ = new ReplaySubject(100);
//   private incommingMessage$_ = new Subject();
//   private readonly unsubscribe_ = new Subject<void>();
//   private readonly keyEvent$_ = new ReplaySubject<KeyboardEvent>(2);

//   myRef = React.createRef<HTMLDivElement>();

//   state: { value: string; list: PodContainerItem[]; sessionId: string } = {
//     value: "",
//     list: [],
//     sessionId: "",
//   };

//   componentDidMount() {
//     this.fetchData();
//   }

//   componentWillUnmount() {
//     console.log("unmount");
//     this.unsubscribe_.next();
//     this.unsubscribe_.complete();
//     if (this.conn_) {
//       this.conn_.close();
//     }
//     if (this.connSubject_) {
//       this.connSubject_.complete();
//     }
//     if (this.term) {
//       this.term.dispose();
//     }
//     this.incommingMessage$_.complete();
//   }

//   initTerm = () => {
//     if (this.connSubject_) {
//       this.connSubject_.complete();
//       this.connSubject_ = new ReplaySubject(100);
//     }
//     if (this.term) {
//       this.term.dispose();
//     }
//     this.term = new Terminal({
//       fontSize: 14,
//       fontFamily: '"Fira code", "Fira Mono", monospace',
//       bellStyle: "sound",
//       cursorBlink: true,
//     });
//     const fitAddon = new FitAddon();
//     this.myRef.current && this.term.open(this.myRef.current);
//     this.term.loadAddon(fitAddon);
//     this.debouncedFit_ = debounce(() => {
//       try {
//         fitAddon.fit();
//       } catch (e) {
//         console.log(e);
//       }
//     }, 300);
//     this.connSubject_.pipe(takeUntil(this.unsubscribe_)).subscribe((frame) => {
//       this.handleConnectionMessage(frame);
//     });
//     this.term.onData(this.onTerminalSendString.bind(this));
//     this.term.onResize(this.onTerminalResize.bind(this));
//     this.term.onKey((event: any) => {
//       console.log(event);
//     });
//   };

//   listContainer = async () => {
//     return containerList(
//       this.props.detail.namespace.id,
//       this.props.detail.id
//     ).then((res) => {
//       this.setState({ list: res.data.data });
//       return res;
//     });
//   };

//   handleConnectionMessage = (frame: any) => {
//     if (frame.Op === "stdout") {
//       this.term?.write(frame.Data);
//     }

//     if (frame.Op === "toast") {
//       message.error(frame.Data);
//       this.listContainer();
//     }

//     this.incommingMessage$_.next(frame);
//   };

//   onTerminalSendString = (str: string) => {
//     if (this.connected_) {
//       this.conn_?.send(
//         JSON.stringify({
//           Op: "stdin",
//           Data: str,
//           Cols: this.term?.cols,
//           Rows: this.term?.rows,
//         })
//       );
//     }
//   };

//   setupConnection = () => {
//     if (this.connecting_) {
//       return;
//     }
//     this.connecting_ = true;
//     this.connectionClosed_ = false;
//     let [pod, container] = this.state.value.split("|");
//     handleExecShell(this.props.detail.namespace.name, pod, container).then(
//       ({ data }) => {
//         this.setState({ sessionId: data.data.id });
//         let url = process.env.REACT_APP_BASE_URL;
//         if (url === "") {
//           url = window.location.origin;
//         }
//         this.conn_ = new SockJS(`${url}/api/sockjs?${data.data.id}`);
//         this.conn_.onopen = this.onConnectionOpen.bind(
//           this,
//           this.state.sessionId
//         );
//         this.conn_.onmessage = this.onConnectionMessage.bind(this);
//         this.conn_.onclose = this.onConnectionClose.bind(this);
//         this.conn_.onerror = this.onErrorMessage.bind(this);
//       }
//     ).catch((e) => {
//       this.connecting_ = false;
//       this.connectionClosed_ = false;
//       message.error(e.response.data.message);
//       this.listContainer()
//     });
//   };

//   fetchData = () => {
//     this.listContainer().then((res) => {
//       if (res.data.data.length > 0) {
//         let first = res.data.data[0];
//         this.setState({ value: first.pod_name + "|" + first.container_name });
//         if (this.conn_ && this.connected_) {
//           this.disconnect();
//         }
//         this.setupConnection();
//         this.initTerm();
//         this.debouncedFit_ && this.debouncedFit_();
//       }
//     });
//   };

//   componentDidUpdate(prevProps: any, prevState: any, snapshot: any) {
//     if (this.props.resizeAt !== prevProps.resizeAt) {
//       this.debouncedFit_ && this.debouncedFit_();
//       return false;
//     }

//     if (this.props.detail.updated_at !== prevProps.detail.updated_at) {
//       this.fetchData();
//     }

//     console.log("this.props, prevProps", this.props, prevProps);
//     if (this.state.value !== prevState.value && this.state.value !== "") {
//       console.log("value changed");
//       this.reconnect()
//     }
//   }

//   reconnect = () => {
//     console.log("reconnect")
//     this.disconnect();
//     this.setupConnection();
//     this.initTerm();
//   }

//   onConnectionClose() {
//     if (!this.connected_) {
//       return;
//     }
//     this.conn_?.close();
//     this.connected_ = false;
//     this.connecting_ = false;
//     this.connectionClosed_ = true;
//   }

//   onErrorMessage(evt: any) {
//     console.log("error", evt);
//   }

//   onConnectionMessage(evt: any) {
//     console.log("onConnectionMessage", evt);
//     const msg = JSON.parse(evt.data);
//     this.connSubject_.next(msg);
//   }
//   onConnectionOpen(id: string) {
//     console.log("onConnectionOpen: ", id);
//     let startData = {
//       Op: "bind",
//       SessionID: id,
//     };
//     this.connected_ = true;
//     this.conn_?.send(JSON.stringify(startData));
//     this.connSubject_.next(startData);
//     this.connected_ = true;
//     this.connecting_ = false;
//     this.connectionClosed_ = false;
//     // Make sure the terminal is with correct display size.
//     this.onTerminalResize();
//     // Focus on connection
//     this.term?.focus();
//   }

//   onTerminalResize = () => {
//     if (this.connected_) {
//       this.conn_?.send(
//         JSON.stringify({
//           Op: "resize",
//           Cols: this.term?.cols,
//           Rows: this.term?.rows,
//         })
//       );
//     }
//   };

//   disconnect = () => {
//     if (this.conn_) {
//       this.conn_.close();
//     }
//     if (this.connSubject_) {
//       this.connSubject_.complete();
//       this.connSubject_ = new ReplaySubject(100);
//     }
//     if (this.term) {
//       this.term.dispose();
//     }
//     this.incommingMessage$_.complete();
//     this.incommingMessage$_ = new Subject();
//   };

//   onChange = (e: RadioChangeEvent) => {
//     console.log("onchange");
//     this.setState({ value: e.target.value });
//   };
//   render() {
//     return (
//       <div>
//         <Radio.Group
//           onChange={this.onChange}
//           value={this.state.value}
//           style={{ marginBottom: 10 }}
//         >
//           {this.state.list.map((item) => (
//             <Radio
//               onClick={() => this.reconnect()}
//               key={item.pod_name + "|" + item.container_name}
//               value={item.pod_name + "|" + item.container_name}
//             >
//               {item.container_name}
//               <Tag color="magenta" style={{ marginLeft: 10 }}>
//                 {item.pod_name}
//               </Tag>
//             </Radio>
//           ))}
//         </Radio.Group>
//         <div>
//           <div style={{ maxHeight: 400 }}>
//             <div ref={this.myRef} id="terminal"></div>
//           </div>
//         </div>
//       </div>
//     );
//   }
// }

// export default memo(Shell);
//   detail: ProjectDetail;
//   resizeAt: number;
//   state: { value: string; list: PodContainerItem[]; sessionId: string } = {
//     value: "",
//     list: [],
//     sessionId: "",
//   };
const TabShell: React.FC<{ detail: ProjectDetail; resizeAt: number }> = ({
  detail,
  resizeAt,
}) => {
  const [list, setList] = useState<PodContainerItem[]>([]);
  const [sessionId, setSessionId] = useState<string>("");
  const [value, setValue] = useState<string>("");
  const [term, setTerm] = useState<Terminal>();
  
  const listContainer = useCallback(async () => {
    return containerList(detail.namespace.id, detail.id).then((res) => {
      setList(res.data.data);
      return res;
    });
  }, [detail.id, detail.namespace.id]);
  const ref = useRef<HTMLDivElement>(null);

  const ws = useWs();
  const wsReady = useWsReady();

  const sendMsg = useCallback((msg: string) => {
    try {
      ws?.send(msg)
    } catch (e) {
      console.log(e)
    }
  }, [ws])

  const [fitAddon, _] = useState(new FitAddon());
  const onTerminalSendString = useCallback(
    (str: string) => {
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
      
      sendMsg(JSON.stringify(re))
    },
    [term, sessionId, sendMsg]
  );
  const debouncedFit_ = debounce(() => {
    try {
      fitAddon.fit();
    } catch (e) {
      console.log(e);
    }
  }, 300);
  const handleConnectionMessage = useCallback(
    (frame: any) => {
      if (frame.op === "stdout") {
        term?.write(frame.data);
      }

      if (frame.op === "toast") {
        message.error(frame.data);
        listContainer();
      }
    },
    [listContainer, term]
  );

  const onTerminalResize = useCallback(() => {
    let re = {
      type: "handle_exec_shell_msg",
      data: JSON.stringify({
        session_id: sessionId,
        op: "resize",
        cols: term?.cols,
        rows: term?.rows,
      }),
    };
    sendMsg(JSON.stringify(re))
  }, [term, sendMsg, sessionId]);

  const sessions = useSelector(selectSessions);
  let sname = useMemo(() => detail.namespace.name + "|" + value, [detail, value])

  const handleCloseShell = useCallback(
    () => {
      if (sessionId) {
        let re = {
          type: "handle_close_shell",
          data: JSON.stringify({
            session_id: sessionId,
          }),
        };
        sendMsg(JSON.stringify(re))
      }
      console.log("closed closed closed closed")
    },
    [sessionId],
  )

  useEffect(() => {
    return () => {
      handleCloseShell()
    }
  }, [handleCloseShell])
  useEffect(() => {
    // 关闭上一个连接如果有的话
    console.log("handle_close_shell")
    handleCloseShell()
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
  }, []);

  useEffect(() => {
    if (!wsReady || !sessionId) {
      console.log(ws, wsReady)
      return
    }

    if (term) {
      term.dispose();
    }
    let myterm = new Terminal({
      fontSize: 14,
      fontFamily: '"Fira code", "Fira Mono", monospace',
      bellStyle: "sound",
      cursorBlink: true,
      logLevel: "debug",
    });
    setTerm(myterm);
    myterm.loadAddon(fitAddon);
    myterm.onResize(onTerminalResize);
    myterm.onData(onTerminalSendString)
    myterm.onKey((event: any) => {
      console.log(event);
    });

    ref.current !== null && myterm.open(ref.current);
    debouncedFit_();
    myterm.focus();
  }, [value, wsReady, sessionId])

  useEffect(() => {
    debouncedFit_();
  }, [resizeAt, debouncedFit_]);

  useEffect(() => {
    if (value && wsReady) {
      let s = value.split("|");
      let re = {
        type: "handle_exec_shell",
        data: JSON.stringify({
          namespace: detail.namespace.name,
          pod: s[0],
          container: s[1],
        }),
      };
      sendMsg(JSON.stringify(re))
    }
  }, [value, detail.namespace.name, sendMsg, wsReady]);

  const reconnect = () => {};
  // useEffect(() => {
  //   return () => {
  //     // 关闭上一个连接如果有的话
  //     console.log("handle_close_shell")
  //     if (sessionId) {
  //       let re = {
  //         type: "handle_close_shell",
  //         data: JSON.stringify({
  //           session_id: sessionId,
  //         }),
  //       };
  //       sendMsg(JSON.stringify(re))
  //       setSessionId("")
  //     }
  //   }
  // }, [sessionId, sendMsg])
  const onChange = (e: RadioChangeEvent) => {
      // // 关闭上一个连接如果有的话
      // if (sessionId) {
      //   let re = {
      //     type: "handle_close_shell",
      //     data: JSON.stringify({
      //       session_id: sessionId,
      //     }),
      //   };
      //   sendMsg(JSON.stringify(re))
      //   setSessionId("")
      // }
    console.log(e.target.value)
    setValue(e.target.value)
  };

  return (
    <div>
      <Radio.Group
        onChange={onChange}
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
      <div>
        <div style={{ maxHeight: 400 }}>
          <div ref={ref} id="terminal"></div>
        </div>
      </div>
    </div>
  );
};

export default TabShell;
