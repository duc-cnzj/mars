import md5 from 'blueimp-md5'

export function toSlug(namespaceId:number, name:string) {
    return md5(namespaceId+"-"+name)
}