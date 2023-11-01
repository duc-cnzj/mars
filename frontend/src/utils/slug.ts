import sha256 from "crypto-js/sha256";

export function toSlug(namespaceId: number, name: string) {
  return sha256(namespaceId + "-" + name).toString();
}
