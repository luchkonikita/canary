import * as React from 'react'
import { connect } from 'react-redux'
import { bindActionCreators } from 'redux'

import { hideModal } from '../../store'

import CreateCrawling from './CreateCrawling'
import DeleteCrawling from './DeleteCrawling'

const MODALS: {
  [P in StoreState['modal']['modalType']]: React.ComponentType<any>
} = {
  CreateCrawling,
  DeleteCrawling
}

type StateProps = StoreState['modal']

type DispatchProps = {
  hideModal: typeof hideModal
}

type Props =  StateProps & DispatchProps

function Modal(props: Props) {
  const { show, modalType, modalParams, hideModal, loading } = props
  if (!show) return null

  const ModalComponent = MODALS[modalType]

  return (
    <ModalComponent {...modalParams} loading={loading} onClose={hideModal} />
  )
}

export default connect(
  (state: StoreState): StateProps => ({ ...state.modal }),
  (dispatch): DispatchProps => bindActionCreators({hideModal}, dispatch)
)(Modal)
