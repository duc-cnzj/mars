import { SET_TIMER_START, SET_TIMER_START_AT } from "../actionTypes";
import { set } from "lodash";

export interface List {
  [id: string]: { start: boolean; startAt: number };
}

const initialState: List = {};

export const selectTimer = (state: { deployTimer: List }) => state.deployTimer;

export default function deployTimer(
  state = initialState,
  action: {
    type: string;
    data: { id: string; start?: boolean; startAt?: number };
  },
) {
  switch (action.type) {
    case SET_TIMER_START:
      if (action.data.start !== undefined) {
        return { ...set(state, [action.data.id, "start"], action.data.start) };
      }
      return state;
    case SET_TIMER_START_AT:
      if (action.data.startAt !== undefined) {
        return {
          ...set(state, [action.data.id, "startAt"], action.data.startAt),
        };
      }

      return state;
    default:
      return state;
  }
}
