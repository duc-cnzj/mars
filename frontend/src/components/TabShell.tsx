import React, { Component, memo } from "react";
import { containerList, PodContainerItem } from "../api/project";
import { message, Radio, RadioChangeEvent, Tag } from "antd";
import { debounce } from "lodash";
import { Terminal } from "xterm";
import { FitAddon } from "xterm-addon-fit";
import { takeUntil } from "rxjs/operators";
import { ReplaySubject, Subject } from "rxjs";
import SockJS from "sockjs-client";
import "xterm/css/xterm.css";
import { handleExecShell } from "../api/shell";

class Shell extends Component<{
  updatedAt: string;
  namespaceId: number;
  id: number;
  namespace: string;
}> {
  private connecting_: boolean = false;
  private connectionClosed_: boolean = false;
  private conn_: WebSocket | null = null;
  private connected_ = false;
  private term: Terminal | null = null;
  private debouncedFit_: Function | null = null;
  private connSubject_ = new ReplaySubject(100);
  private incommingMessage$_ = new Subject();
  private readonly unsubscribe_ = new Subject<void>();
  private readonly keyEvent$_ = new ReplaySubject<KeyboardEvent>(2);

  myRef = React.createRef<HTMLDivElement>();

  state: { value: string; list: PodContainerItem[]; sessionId: string } = {
    value: "",
    list: [],
    sessionId: "",
  };

  componentDidMount() {
    this.fetchData();
  }

  componentWillUnmount() {
    console.log("unmount");
    this.unsubscribe_.next();
    this.unsubscribe_.complete();
    if (this.conn_) {
      this.conn_.close();
    }
    if (this.connSubject_) {
      this.connSubject_.complete();
    }
    if (this.term) {
      this.term.dispose();
    }
    this.incommingMessage$_.complete();
  }

  initTerm = () => {
    if (this.connSubject_) {
      this.connSubject_.complete();
      this.connSubject_ = new ReplaySubject(100);
    }
    if (this.term) {
      this.term.dispose();
    }
    this.term = new Terminal({
      fontSize: 14,
      fontFamily: '"Fira code", "Fira Mono", monospace',
      bellStyle: "sound",
      cursorBlink: true,
    });
    const fitAddon = new FitAddon();
    this.myRef.current && this.term.open(this.myRef.current);
    this.term.loadAddon(fitAddon);
    this.debouncedFit_ = debounce(() => {
      try {
        fitAddon.fit();
      } catch (e) {
        console.log(e);
      }
    }, 300);
    window.addEventListener(
      "resize",
      () => this.debouncedFit_ && this.debouncedFit_()
    );
    this.connSubject_.pipe(takeUntil(this.unsubscribe_)).subscribe((frame) => {
      this.handleConnectionMessage(frame);
    });
    this.term.onData(this.onTerminalSendString.bind(this));
    this.term.onResize(this.onTerminalResize.bind(this));
    this.term.onKey((event: any) => {
      console.log(event);
    });
  };

  handleConnectionMessage = (frame: any) => {
    if (frame.Op === "stdout") {
      this.term?.write(frame.Data);
    }

    if (frame.Op === "toast") {
      message.error(frame.Data);
    }

    this.incommingMessage$_.next(frame);
  };

  onTerminalSendString = (str: string) => {
    if (this.connected_) {
      this.conn_?.send(
        JSON.stringify({
          Op: "stdin",
          Data: str,
          Cols: this.term?.cols,
          Rows: this.term?.rows,
        })
      );
    }
  };

  setupConnection = () => {
    if (this.connecting_) {
      return;
    }
    this.connecting_ = true;
    this.connectionClosed_ = false;
    let [pod, container] = this.state.value.split("|");
    handleExecShell(this.props.namespace, pod, container).then(({ data }) => {
      this.setState({ sessionId: data.data.id });
      let url = process.env.REACT_APP_BASE_URL;
      if (url === "") {
        url = window.location.origin;
      }
      this.conn_ = new SockJS(`${url}/api/sockjs?${data.data.id}`);
      this.conn_.onopen = this.onConnectionOpen.bind(
        this,
        this.state.sessionId
      );
      this.conn_.onmessage = this.onConnectionMessage.bind(this);
      this.conn_.onclose = this.onConnectionClose.bind(this);
      this.conn_.onerror = this.onErrorMessage.bind(this);
    });
  };

  fetchData = () => {
    containerList(this.props.namespaceId, this.props.id).then((res) => {
      this.setState({ list: res.data.data });
      if (res.data.data.length > 0) {
        let first = res.data.data[0];
        this.setState({ value: first.pod_name + "|" + first.container_name });
        if (this.conn_ && this.connected_) {
          this.disconnect();
        }
        this.setupConnection();
        this.initTerm();
        this.debouncedFit_ && this.debouncedFit_();
      }
    });
  };

  componentDidUpdate(prevProps: any, prevState: any, snapshot: any) {
    if (this.props.updatedAt !== prevProps.updatedAt) {
      this.fetchData();
    }
    if (this.state.value !== prevState.value && this.state.value !== "") {
      this.disconnect();
      this.setupConnection();
      this.initTerm();
    }
  }

  onConnectionClose() {
    if (!this.connected_) {
      return;
    }
    this.conn_?.close();
    this.connected_ = false;
    this.connecting_ = false;
    this.connectionClosed_ = true;
  }

  onErrorMessage(evt: any) {
    console.log("error", evt);
  }

  onConnectionMessage(evt: any) {
    console.log("onConnectionMessage", evt);
    const msg = JSON.parse(evt.data);
    this.connSubject_.next(msg);
  }
  onConnectionOpen(id: string) {
    console.log("onConnectionOpen: ", id);
    let startData = {
      Op: "bind",
      SessionID: id,
    };
    this.connected_ = true;
    this.conn_?.send(JSON.stringify(startData));
    this.connSubject_.next(startData);
    this.connected_ = true;
    this.connecting_ = false;
    this.connectionClosed_ = false;
    // Make sure the terminal is with correct display size.
    this.onTerminalResize();
    // Focus on connection
    this.term?.focus();
  }

  onTerminalResize = () => {
    if (this.connected_) {
      this.conn_?.send(
        JSON.stringify({
          Op: "resize",
          Cols: this.term?.cols,
          Rows: this.term?.rows,
        })
      );
    }
  };

  disconnect = () => {
    if (this.conn_) {
      this.conn_.close();
    }
    if (this.connSubject_) {
      this.connSubject_.complete();
      this.connSubject_ = new ReplaySubject(100);
    }
    if (this.term) {
      this.term.dispose();
    }
    this.incommingMessage$_.complete();
    this.incommingMessage$_ = new Subject();
  };

  onChange = (e: RadioChangeEvent) => {
    this.setState({ value: e.target.value });
  };

  render() {
    return (
      <div>
        <Radio.Group
          onChange={this.onChange}
          value={this.state.value}
          style={{ marginBottom: 10 }}
        >
          {this.state.list.map((item) => (
            <Radio
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
            <div ref={this.myRef} id="terminal"></div>
          </div>
        </div>
      </div>
    );
  }
}

export default memo(Shell);
