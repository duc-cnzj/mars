import { SET_NAMESPACE_RELOAD } from "./../actionTypes";
interface namespaceState {
  reload: boolean;
  nsID: number;
}
const initialState: namespaceState = {
  reload: false,
  nsID: 0,
};

export const selectReload = (state: { namespace: namespaceState }) =>
  state.namespace.reload;
export const selectReloadNsID = (state: { namespace: namespaceState }) =>
  state.namespace.nsID;

export default function namespace(
  state = initialState,
  action: { type: string; data?: { reload: boolean; nsID: number } }
) {
  switch (action.type) {
    case SET_NAMESPACE_RELOAD:
      const obj = action.data
        ? { reload: action.data.reload, nsID: action.data.nsID }
        : {};
      return { ...state, ...obj };
    default:
      return state;
  }
}
