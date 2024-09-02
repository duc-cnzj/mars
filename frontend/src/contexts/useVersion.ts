import { useEffect, useState } from "react";
import ajax from "../api/ajax";
import { components } from "../api/schema";

export default function useVersion() {
  const [version, setVersion] =
    useState<components["schemas"]["version.Response"]>();

  useEffect(() => {
    ajax.GET("/api/version").then(({ data }) => setVersion(data));
  }, []);
  return version;
}
