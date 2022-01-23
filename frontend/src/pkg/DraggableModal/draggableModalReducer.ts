import { getWindowSize } from './getWindowSize'
import { clamp } from './clamp'

const mapObject = <T>(o: { [key: string]: T }, f: (value: T) => T): { [key: string]: T } =>
    Object.assign({}, ...Object.keys(o).map(k => ({ [k]: f(o[k]) })))

// ID for a specific modal.
export type ModalID = string

// State for a specific modal.
export interface ModalState {
    previous?: {
        x: number
        y: number
        width: number
        height: number
        zIndex: number
    }
    x: number
    y: number
    width: number
    height: number
    zIndex: number
    visible: boolean
}

// State of all modals.
export interface ModalsState {
    maxZIndex: number
    initSize: {
        width: number
        height: number
    }
    windowSize: {
        width: number
        height: number
    }
    modals: {
        [key: string]: ModalState
    }
}

export const initialModalsState: ModalsState = {
    maxZIndex: 0,
    initSize: getWindowSize(),
    windowSize: getWindowSize(),
    modals: {},
}

export const initialModalState: ModalState = {
    x: 0,
    y: 0,
    width: 800,
    height: 800,
    zIndex: 0,
    visible: false,
}

const getInitialModalState = ({
    initialWidth = initialModalState.width,
    initialHeight = initialModalState.height,
}: {
    initialWidth?: number
    initialHeight?: number
}) => {
    return {
        ...initialModalState,
        width: initialWidth,
        height: initialHeight,
    }
}

export type Action =
    | { type: 'doubleClick'; id: ModalID }
    | { type: 'show'; id: ModalID }
    | { type: 'hide'; id: ModalID }
    | { type: 'focus'; id: ModalID }
    | { type: 'unmount'; id: ModalID }
    | { type: 'mount'; id: ModalID; intialState: { initialWidth?: number; initialHeight?: number } }
    | { type: 'windowResize'; size: { width: number; height: number } }
    | { type: 'drag'; id: ModalID; x: number; y: number }
    | {
          type: 'resize'
          id: ModalID
          x: number
          y: number
          width: number
          height: number
      }

export const getModalState = ({
    state,
    id,
    initialWidth,
    initialHeight,
}: {
    state: ModalsState
    id: ModalID
    initialWidth?: number
    initialHeight?: number
}): ModalState => state.modals[id] || getInitialModalState({ initialWidth, initialHeight })

const getNextZIndex = (state: ModalsState, id: string): number =>
    getModalState({ state, id }).zIndex === state.maxZIndex ? state.maxZIndex : state.maxZIndex + 1

const clampDrag = (
    windowWidth: number,
    windowHeight: number,
    x: number,
    y: number,
    width: number,
    height: number,
): { x: number; y: number } => {
    const maxX = windowWidth - width
    const maxY = windowHeight - height
    const clampedX = clamp(0, maxX, x)
    const clampedY = clamp(0, maxY, y)
    return { x: clampedX, y: clampedY }
}

const clampResize = (
    windowWidth: number,
    windowHeight: number,
    x: number,
    y: number,
    width: number,
    height: number,
): { width: number; height: number } => {
    const maxWidth = windowWidth - x
    const maxHeight = windowHeight - y
    const clampedWidth = clamp(200, maxWidth, width)
    const clampedHeight = clamp(200, maxHeight, height)
    return { width: clampedWidth, height: clampedHeight }
}

export const draggableModalReducer = (state: ModalsState, action: Action): ModalsState => {
    const removeOverflow = (exceptID: string) => {
        let hidden = true
            for (const key in state.modals) {
                if (Object.prototype.hasOwnProperty.call(state.modals, key)) {
                    const element = state.modals[key];
                    if (element.visible && exceptID != key) {
                        hidden = false
                        break
                    }
                }
            }
            if (hidden) {
                document.body.classList.remove("ant-design-draggable-modal-body-overflow")
            }
    }
    switch (action.type) {
        case 'doubleClick':
            let curr = state.modals[action.id]
            let prev = state.modals[action.id].previous

            if (state.modals[action.id].width === state.windowSize.width && state.modals[action.id].height === state.windowSize.height) {
                if (!prev || (prev.width === state.windowSize.width && prev.height === state.windowSize.height)) {
                    prev = {
                        x: 0,
                        y: 0,
                        width: state.initSize.width,
                        height: state.initSize.height,
                        zIndex: curr.zIndex,
                    }
                }

                return {
                    ...state,
                    maxZIndex: getNextZIndex(state, action.id),
                    modals: {
                        ...state.modals,
                        [action.id]: {
                            ...state.modals[action.id],
                            previous: {
                                x: 0,
                                y: 0,
                                width: state.windowSize.width,
                                height: state.windowSize.height,
                                zIndex: curr.zIndex,
                            },
                            x: prev.x,
                            y: prev.y,
                            zIndex: state.maxZIndex + 1,
                            width: prev.width,
                            height: prev.height,
                        },
                    },
                }
            }

            return {
                ...state,
                maxZIndex: getNextZIndex(state, action.id),
                modals: {
                    ...state.modals,
                    [action.id]: {
                        ...state.modals[action.id],
                        previous: {
                            x: curr.x,
                            y: curr.y,
                            width: curr.width,
                            height: curr.height,
                            zIndex: curr.zIndex,
                        },
                        x: 0,
                        y: 0,
                        zIndex: state.maxZIndex + 1,
                        width: state.windowSize.width,
                        height: state.windowSize.height,
                    },
                },
            }
        case 'resize':
            const size = clampResize(
                state.windowSize.width,
                state.windowSize.height,
                action.x,
                action.y,
                action.width,
                action.height,
            )
            return {
                ...state,
                maxZIndex: getNextZIndex(state, action.id),
                modals: {
                    ...state.modals,
                    [action.id]: {
                        ...state.modals[action.id],
                        ...size,
                        zIndex: getNextZIndex(state, action.id),
                    },
                },
            }
        case 'drag':
            return {
                ...state,
                maxZIndex: getNextZIndex(state, action.id),
                modals: {
                    ...state.modals,
                    [action.id]: {
                        ...state.modals[action.id],
                        ...clampDrag(
                            state.windowSize.width,
                            state.windowSize.height,
                            action.x,
                            action.y,
                            state.modals[action.id].width,
                            state.modals[action.id].height,
                        ),
                        zIndex: getNextZIndex(state, action.id),
                    },
                },
            }
        case 'show': {
            document.body.classList.add("ant-design-draggable-modal-body-overflow")
            const modalState = state.modals[action.id]
            const centerX = state.windowSize.width / 2 - modalState.width / 2
            const centerY = state.windowSize.height / 2 - modalState.height / 2
            const position = clampDrag(
                state.windowSize.width,
                state.windowSize.height,
                centerX,
                centerY,
                modalState.width,
                modalState.height,
            )
            const size = clampResize(
                state.windowSize.width,
                state.windowSize.height,
                position.x,
                position.y,
                modalState.width,
                modalState.height,
            )
            return {
                ...state,
                maxZIndex: state.maxZIndex + 1,
                modals: {
                    ...state.modals,
                    [action.id]: {
                        ...modalState,
                        ...position,
                        ...size,
                        zIndex: state.maxZIndex + 1,
                        visible: true,
                    },
                },
            }
        }
        case 'focus':
            const modalState = state.modals[action.id]
            return {
                ...state,
                maxZIndex: state.maxZIndex + 1,
                modals: {
                    ...state.modals,
                    [action.id]: {
                        ...modalState,
                        zIndex: state.maxZIndex + 1,
                    },
                },
            }
        case 'hide': {
            removeOverflow(action.id)
            const modalState = state.modals[action.id]
            return {
                ...state,
                modals: {
                    ...state.modals,
                    [action.id]: {
                        ...modalState,
                        visible: false,
                    },
                },
            }
        }
        case 'mount':
            const initialState = getInitialModalState(action.intialState)
            return {
                ...state,
                initSize: initialState,
                maxZIndex: state.maxZIndex + 1,
                modals: {
                    ...state.modals,
                    [action.id]: {
                        ...initialState,
                        x: state.windowSize.width / 2 - initialState.width / 2,
                        y: state.windowSize.height / 2 - initialState.height / 2,
                        zIndex: state.maxZIndex + 1,
                    },
                },
            }
        case 'unmount':
            removeOverflow(action.id)
            const modalsClone = { ...state.modals }
            delete modalsClone[action.id]
            return {
                ...state,
                modals: modalsClone,
            }
        case 'windowResize':
            return {
                ...state,
                windowSize: action.size,
                modals: mapObject(state.modals, (modalState: ModalState) => {
                    if (!modalState.visible) {
                        return modalState
                    }
                    const position = clampDrag(
                        state.windowSize.width,
                        state.windowSize.height,
                        modalState.x,
                        modalState.y,
                        modalState.width,
                        modalState.height,
                    )
                    const size = clampResize(
                        state.windowSize.width,
                        state.windowSize.height,
                        position.x,
                        position.y,
                        modalState.width,
                        modalState.height,
                    )
                    return {
                        ...modalState,
                        ...position,
                        ...size,
                    }
                }),
            }
        default:
            throw new Error()
    }
}
