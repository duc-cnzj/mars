import { combineReducers } from "redux";
import createProject from './createProject'
import namespace from './namespace'
import namespaceWatcher from './namespaceWatcher'
import cluster from './cluster'
import shell from './shell'
import deployTimer from './deployTimer'

export default combineReducers({ createProject,  namespaceWatcher, namespace, cluster, shell, deployTimer });
