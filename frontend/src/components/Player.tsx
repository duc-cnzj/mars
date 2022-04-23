import React, { memo, useEffect, useRef } from "react";
// @ts-ignore
import * as AsciinemaPlayerLibrary from "asciinema-player";
import "asciinema-player/dist/bundle/asciinema-player.css";

interface AsciinemaPlayerProps {
  src: string | { url: string; fetchOpts: RequestInit };
  // START asciinemaOptions
  cols?: number;
  rows?: number;
  autoPlay?: boolean;
  preload?: boolean;
  loop?: boolean | number;
  startAt?: number | string;
  speed?: number;
  idleTimeLimit?: number;
  theme?:
    | "asciinema"
    | "monokai"
    | "tango"
    | "solarized-dark"
    | "solarized-light";
  poster?: string;
  fit?: string;
  fontSize?: string;
  // END asciinemaOptions
}

const AsciinemaPlayer: React.FC<AsciinemaPlayerProps> = ({
  src,
  cols,
  rows,
  autoPlay,
  preload,
  loop,
  startAt,
  speed,
  idleTimeLimit,
  theme,
  poster,
  fit,
  fontSize,
}) => {
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const currentRef = ref.current;
    const player = AsciinemaPlayerLibrary.create(src, currentRef, {
      cols,
      rows,
      autoPlay,
      preload,
      loop,
      startAt,
      speed,
      idleTimeLimit,
      theme,
      poster,
      fit,
      fontSize,
    });

    return () => {
      player.dispose();
    };
  }, [
    src,
    cols,
    rows,
    autoPlay,
    preload,
    loop,
    startAt,
    speed,
    idleTimeLimit,
    theme,
    poster,
    fit,
    fontSize,
  ]);

  return <div ref={ref} />;
};

export default memo(AsciinemaPlayer);