export interface TodoPanelState {
  listId: string;
  hovered: boolean;
  focused: boolean;
  pinned: boolean;
}

export type TodoPanelAction =
  | { type: "list"; listId: string }
  | { type: "hover"; value: boolean }
  | { type: "focus"; value: boolean }
  | { type: "toggle-pin" }
  | { type: "close" };

export function createTodoPanelState(listId: string): TodoPanelState {
  return { listId, hovered: false, focused: false, pinned: false };
}

export function reduceTodoPanelState(state: TodoPanelState, action: TodoPanelAction): TodoPanelState {
  switch (action.type) {
    case "list":
      return action.listId === state.listId ? state : createTodoPanelState(action.listId);
    case "hover":
      return { ...state, hovered: action.value };
    case "focus":
      return { ...state, focused: action.value };
    case "toggle-pin":
      return { ...state, pinned: !state.pinned };
    case "close":
      return { ...state, hovered: false, focused: false, pinned: false };
  }
}

export function isTodoPanelOpen(state: TodoPanelState): boolean {
  return state.hovered || state.focused || state.pinned;
}
