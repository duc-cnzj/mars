import { SET_NAMESPACE_RELOAD } from "./../actionTypes";
const initialState = {
  reload: false,
};

export const selectReload = (state:{namespace: {reload:boolean}}) => state.namespace.reload;

export default function namespace(
  state = initialState,
  action: { type: string; data?: { reload: boolean } }
) {
  switch (action.type) {
    case SET_NAMESPACE_RELOAD:
      return { ...state, reload: action.data ? action.data.reload : false };
    default:
      return state;
  }
}
