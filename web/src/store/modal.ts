type ShowModalAction = StoreAction<'SHOW_MODAL', { modalType: ModalState['modalType'], modalParams: ModalState['modalParams'] }>
type HideModalAction = StoreAction<'HIDE_MODAL', {}>
type SetModalLoadingAction = StoreAction<'SET_MODAL_LOADING', { loading: boolean }>

type ModalAction = ShowModalAction | HideModalAction | SetModalLoadingAction

export function showModal(type: StoreState['modal']['modalType'], params: any): ShowModalAction {
  return {
    type: 'SHOW_MODAL',
    payload: {
      modalType: type,
      modalParams: params
    }
  }
}

export function hideModal(): HideModalAction {
  return {
    type: 'HIDE_MODAL',
    payload: {}
  }
}

export function setModalLoading(loading: boolean): SetModalLoadingAction {
  return {
    type: 'SET_MODAL_LOADING',
    payload: {loading}
  }
}

export function reducer(state = { show: false, loading: false }, action: ModalAction): ModalState {
  switch (action.type) {
    case 'SHOW_MODAL':
      return {...action.payload, show: true, loading: false}
    case 'HIDE_MODAL':
      return {show: false, loading: false}
    case 'SET_MODAL_LOADING':
      return {...state, loading: action.payload.loading}
    default:
      return state
  }
}
