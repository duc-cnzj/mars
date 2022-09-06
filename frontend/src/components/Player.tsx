import React, { memo, useEffect, useRef, useState } from "react";
// @ts-ignore
import * as AsciinemaPlayerLibrary from "asciinema-player";
import "asciinema-player/dist/bundle/asciinema-player.css";
import { Radio } from "antd";
interface AsciinemaPlayerProps {
  src: { data: string } | { url: string; fetchOpts: RequestInit };
  // START asciinemaOptions
  cols?: number;
  rows?: number;
  autoPlay?: boolean;
  preload?: boolean;
  loop?: boolean | number;
  startAt?: number | string;
  speed?: number;
  idleTimeLimit?: number;
  terminalLineHeight?: number;
  theme?:
    | "asciinema"
    | "monokai"
    | "tango"
    | "solarized-dark"
    | "solarized-light";
  poster?: string;
  fit?: string | boolean;
  terminalFontSize?: string;
  // END asciinemaOptions
}

const AsciinemaPlayer: React.FC<AsciinemaPlayerProps> = ({
  src,
  cols,
  rows,
  autoPlay: initAutoPlay,
  preload,
  loop,
  startAt: initStartAt,
  speed: initSpeed,
  idleTimeLimit,
  theme,
  fit,
  terminalLineHeight,
  terminalFontSize,
}) => {
  const ref = useRef<HTMLDivElement>(null);
  const [speed, setSpeed] = useState(initSpeed);
  const [player, setPlayer] = useState<any>();
  const [startAt, setStartAt] = useState(initStartAt);
  const [autoPlay, setAutoPlay] = useState(initAutoPlay);
  const [paused, setPaused] = useState(!initAutoPlay);

  useEffect(() => {
    const currentRef = ref.current;
    const p = AsciinemaPlayerLibrary.create(src, currentRef, {
      cols,
      rows,
      autoPlay,
      preload,
      loop,
      startAt,
      speed,
      idleTimeLimit,
      theme,
      fit,
      terminalFontSize,
      terminalLineHeight,
      poster: "npt:" + startAt,
    });
    setPlayer(p);
    p.addEventListener("play", () => {
      setPaused(false);
    });
    p.addEventListener("pause", () => {
      setPaused(true);
    });

    return () => {
      p.dispose();
    };
  }, [
    terminalLineHeight,
    src,
    cols,
    rows,
    autoPlay,
    preload,
    loop,
    speed,
    idleTimeLimit,
    theme,
    fit,
    terminalFontSize,
    startAt,
  ]);

  return (
    <div>
      <div
        style={{ display: "flex", alignItems: "center", marginBottom: 3 }}
        onKeyDown={(k) => {
          if (k.code === "Space") {
            if (paused) {
              player.play();
            } else {
              player.pause();
            }
          }
        }}
      >
        <span style={{ marginRight: 10 }}>速度:</span>
        <Radio.Group
          onChange={(e) => {
            setSpeed(e.target.value);
            if (player.getCurrentTime() > 0) {
              setStartAt(player.getCurrentTime());
              setAutoPlay(!paused);
            }
          }}
          value={speed}
        >
          <Radio value={0.5}>0.5x</Radio>
          <Radio value={0.75}>0.75x</Radio>
          <Radio value={1}>1x</Radio>
          <Radio value={1.5}>1.5x</Radio>
          <Radio value={2}>2x</Radio>
          <Radio value={2.5}>2.5x</Radio>
          <Radio value={3}>3x</Radio>
        </Radio.Group>
      </div>
      <div ref={ref} />
    </div>
  );
};

export default memo(AsciinemaPlayer);
