import { combineReducers } from "redux";
import createProject from './createProject'
import namespace from './namespace'

export default combineReducers({ createProject, namespace });
