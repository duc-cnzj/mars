import { CLEAR_CREATE_PROJECT_LOG, SET_DEPLOY_STATUS, SET_ELAPSED_TIME, SET_PROCESS_PERCENT } from "./../actionTypes";
import {
  CREATE_PROJECT_LOADING,
  CREATE_PROJECT_LOADING_DONE,
  APPEND_CREATE_PROJECT_LOG,
  SET_CREATE_PROJECT_LOADING,
} from "../actionTypes";

import { set, get } from "lodash";

export enum DeployStatus {
  DeployUnknown = "unknown",
  DeployFailed = "failed",
  DeployCanceled = "canceled",
  DeploySuccess = "success",

  DeployUpdateSuccess = "update_success"
}

export interface CreateProjectItem {
  isLoading: boolean;
  deployStatus: DeployStatus;
  output: string[];
  processPercent: number;
  ElapsedTime: string;
}

export const selectList = (state: { createProject: List }):List =>
  state.createProject;

export interface List {
  [id: string]: CreateProjectItem;
}

const initialState: List = {};

export default function createProject(
  state = initialState,
  action: {
    type: string;
    data?: {
      id: string;
      isLoading: boolean;
      output: string;
      deployStatus: string;
      processPercent: number;
      elapsedTime: string;
    };
  }
) {
  switch (action.type) {
    case SET_PROCESS_PERCENT:
      if (action.data) {
        return {
          ...set(
            state,
            [action.data.id, "processPercent"],
            action.data.processPercent
          ),
        };
      }

      return state;
    case SET_DEPLOY_STATUS:
      if (action.data) {
        return {
          ...set(
            state,
            [action.data.id, "deployStatus"],
            action.data.deployStatus
          ),
        };
      }

      return state;
    case SET_ELAPSED_TIME:
      if (action.data) {
        return {
          ...set(
            state,
            [action.data.id, "ElapsedTime"],
            action.data.elapsedTime
          ),
        };
      }

      return state;
    case SET_CREATE_PROJECT_LOADING:
      if (action.data) {
        return {
          ...set(state, [action.data.id, "isLoading"], action.data.isLoading),
        };
      }

      return state;
    case CREATE_PROJECT_LOADING:
      if (action.data) {
        return { ...set(state, [action.data.id, "isLoading"], true) };
      }

      return state;
    case CREATE_PROJECT_LOADING_DONE:
      if (action.data) {
        return { ...set(state, [action.data.id, "isLoading"], false) };
      }

      return state;
    case CLEAR_CREATE_PROJECT_LOG:
      if (action.data) {
        return { ...set(state, [action.data.id, "output"], []) };
      }

      return state;
    case APPEND_CREATE_PROJECT_LOG:
      if (action.data) {
        let g = get(state, [action.data.id, "output"], []);
        return {
          ...set(state, [action.data.id, "output"], [...g, action.data.output]),
        };
      }

      return state;
    default:
      return state;
  }
}
