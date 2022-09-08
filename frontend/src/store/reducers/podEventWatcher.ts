import { PROJECT_POD_EVENT } from "../actionTypes";
const initialState = {
  projectIDWithTimestamp: "",
};

export const selectPodEventProjectID = (state: {
  podEventWatcher: { projectIDWithTimestamp: string };
}) => state.podEventWatcher.projectIDWithTimestamp;

export default function podEventWatcher(
  state = initialState,
  action: { type: string; projectIDWithTimestamp: string }
) {
  switch (action.type) {
    case PROJECT_POD_EVENT:
      return { projectIDWithTimestamp: action.projectIDWithTimestamp };
    default:
      return state;
  }
}
