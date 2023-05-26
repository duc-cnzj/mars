import { combineReducers } from "redux";
import createProject from "./createProject";
import namespace from "./namespace";
import podEventWatcher from "./podEventWatcher";
import cluster from "./cluster";
import shell from "./shell";
import deployTimer from "./deployTimer";
import openedModal from "./openedModal";

export default combineReducers({
  createProject,
  podEventWatcher,
  namespace,
  cluster,
  shell,
  deployTimer,
  openedModal,
});
