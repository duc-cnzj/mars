import { useCallback, useEffect, useRef } from "react";

/**
 * @typedef {object} StreamHook
 * @property {function()} close - Close the stream
 */

/**
 * React hook for the [Streams API](https://developer.mozilla.org/en-US/docs/Web/API/Streams_API).
 * Use this hook to stream data from a URL.
 * @param {string} url
 * @param {object} [params]
 * @param {function(Response)} [params.onNext]
 * @param {function(Error)} [params.onError]
 * @param {function()} [params.onDone]
 * @param {RequestInit} [params.fetchParams]
 *
 * @returns {StreamHook}
 */
export function useStream(url: string, params: any) {
  if (typeof params !== "object" || params === null) {
    params = {};
  }

  const streamRef = useRef<any>();
  const onNext = useRef(params.onNext);
  const onError = useRef(params.onError);
  const onDone = useRef(params.onDone);

  const close = useCallback(() => {
    if (streamRef.current) {
      streamRef.current.abort();
    }
  }, []);

  useEffect(() => {
    if (streamRef.current) {
      streamRef.current.abort();
    }
    streamRef.current = new AbortController();
    startStream(url, {
      onNext: onNext,
      onError: onError,
      onDone: onDone,
      fetchParams: {
        ...params.fetchParams,
        signal: streamRef.current.signal,
      },
    });
  }, [url]);

  useEffect(() => {
    onNext.current = params.onNext;
  }, [params.onNext]);

  useEffect(() => {
    onError.current = params.onError;
  }, [params.onError]);

  useEffect(() => {
    onDone.current = params.onDone;
  }, [params.onDone]);

  return { close };
}

/**
 * Use this function to start streaming data from an URL
 * @param {string} url
 * @param {object} params
 * @param {React.MutableRefObject<function(Response)>} params.onNext
 * @param {React.MutableRefObject<function(Error)>} params.onError
 * @param {React.MutableRefObject<function()>} params.onDone
 * @param {RequestInit} params.fetchParams
 */
async function startStream(
  url: string,
  { onNext, onError, onDone, fetchParams }: any
) {
  const errCb = (err: any) => {
    if (typeof onError.current === "function") {
      onError.current(err);
    }
  };

  try {
    const res = await fetch(url, {
      ...fetchParams,
      method: "GET",
    });

    const reader = res.body?.getReader();

    if (fetchParams.signal instanceof AbortSignal) {
      fetchParams.signal.addEventListener(
        "abort",
        (evt: any) => reader?.cancel(evt),
        {
          once: true,
          passive: true,
        }
      );
    }

    // eslint-disable-next-line no-constant-condition
    while (true) {
      try {
        const r: any = await reader?.read();
        let { done, value } = r;
        if (done) {
          if (typeof onDone.current === "function") {
            onDone.current();
          }
          return;
        }

        const res = new Response(value);
        if (typeof onNext.current === "function") {
          onNext.current(res);
        }
      } catch (e) {
        errCb(e);
        return;
      }
    }
  } catch (e) {
    errCb(e);
  }
}
