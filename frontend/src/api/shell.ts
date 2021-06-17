import ajax from "./ajax";

export function handleExecShell(namespace:string, pod:string, container: string = "") {
  return ajax.get<{data: {id: string}}>(`/api/pod/${namespace}/${pod}/shell?container=${container}`);
}
