import {
  SET_SHELL_SESSION_ID,
  SET_SHELL_LOG,
  REMOVE_SHELL,
} from "./../actionTypes";
import pb from "../../api/websocket";

const initialState: {
  [id: string]: {
    log: pb.websocket.TerminalMessage;
    logCount: number;
  };
} = {};

export const selectSessions = (state: {
  shell: {
    [id: string]: {
      log: pb.websocket.TerminalMessage;
      logCount: number;
    };
  };
}) => state.shell;

export default function shell(
  state = initialState,
  action: {
    type: string;
    data: { id: string; log: pb.websocket.TerminalMessage };
  }
) {
  switch (action.type) {
    case REMOVE_SHELL:
      delete state[action.data.id];
      return { ...state };
    case SET_SHELL_LOG:
      let count = 0;
      if (state[action.data.id] && state[action.data.id].logCount) {
        count = state[action.data.id].logCount;
      }
      return {
        ...state,
        [action.data.id]: {
          ...state[action.data.id],
          log: action.data.log,
          logCount: count + 1,
        },
      };
    case SET_SHELL_SESSION_ID:
      return {
        ...state,
        [action.data.id]: {},
      };
    default:
      return state;
  }
}
