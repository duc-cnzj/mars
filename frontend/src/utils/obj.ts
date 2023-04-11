import { omit, isEqual } from "lodash";

export function OmitEqual<T extends object, K extends keyof T>(
  a: T,
  b: T,
  ...paths: K[]
): boolean {
  return isEqual(omit(a, paths), omit(b, paths));
}
