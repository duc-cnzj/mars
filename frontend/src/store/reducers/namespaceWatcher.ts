import { PROJECT_POD_EVENT } from "../actionTypes";
const initialState = {
  projectID: "",
};

export const selectPodEventProjectID = (state:{namespaceWatcher: {projectID:string}}) => state.namespaceWatcher.projectID;

export default function namespaceWatcher(
  state = initialState,
  action: { type: string; projectID: string }
) {
  switch (action.type) {
    case PROJECT_POD_EVENT:
      console.log("action.projectID", action.projectID)
      return { projectID: action.projectID };
    default:
      return state;
  }
}
