import { combineReducers } from "redux";
import createProject from './createProject'
import namespace from './namespace'
import cluster from './cluster'
import shell from './shell'

export default combineReducers({ createProject, namespace, cluster, shell });
