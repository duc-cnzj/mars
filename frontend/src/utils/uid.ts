export function getUid():string {
    return window.localStorage.getItem("uid") || ""
}
export function setUid(uid:string) {
    window.localStorage.setItem("uid", uid)
}