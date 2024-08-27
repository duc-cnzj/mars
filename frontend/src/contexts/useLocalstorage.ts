import { useEffect, useState } from "react";

export default function useLocalStorage(key: string, defaultValue?: string) {
  let initValue = localStorage.getItem(key) || defaultValue || "";
  const [store, setStore] = useState(initValue);
  useEffect(() => {
    localStorage.setItem(key, store);
  }, [store, key]);

  return { store, setStore };
}
