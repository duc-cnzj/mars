import { combineReducers } from "redux";
import createProject from './createProject'
import namespace from './namespace'
import cluster from './cluster'

export default combineReducers({ createProject, namespace, cluster });
