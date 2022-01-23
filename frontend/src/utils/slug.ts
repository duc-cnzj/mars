import md5 from 'blueimp-md5'

export function toSlug(namespaceId:number, name:string) {
    console.log(md5(namespaceId+"-"+name), "slug");
    return md5(namespaceId+"-"+name)
}