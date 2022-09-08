import { PROJECT_POD_EVENT } from "../actionTypes";
const initialState = {
  projectID: "",
};

export const selectPodEventProjectID = (state:{podEventWatcher: {projectID:string}}) => state.podEventWatcher.projectID;

export default function podEventWatcher(
  state = initialState,
  action: { type: string; projectID: string }
) {
  switch (action.type) {
    case PROJECT_POD_EVENT:
      return { projectID: action.projectID };
    default:
      return state;
  }
}
