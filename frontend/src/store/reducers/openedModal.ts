import { SET_OPENED_MODALS } from "../actionTypes";

interface modalState {
  modals: {
    [key: number]: boolean;
  };
}
const initialState: modalState = {
  modals: {},
};

export const modals = (state: { openedModal: modalState }) =>
  state.openedModal.modals;

export default function openedModal(
  state = initialState,
  action: { type: string; data: { modals: Pick<modalState, "modals"> } },
) {
  switch (action.type) {
    case SET_OPENED_MODALS:
      return { modals: action.data.modals };
    default:
      return state;
  }
}
