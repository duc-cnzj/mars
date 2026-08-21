/** 容器名显示：截掉 "{项目名}-" 前缀，只留尾段 hash（hz-uco-9d48959bf-r2f54 → 9d48959bf-r2f54）。
 *  仅当名字以 "{项目名}-" 开头才截断（符合规则才截），否则原样返回。 */
export function shortContainerName(name: string, projectName: string): string {
  const prefix = `${projectName}-`
  return name.startsWith(prefix) ? name.slice(prefix.length) : name
}
