import * as React from 'react'
import { Dialog, Text } from 'evergreen-ui'
import { connect } from 'react-redux'
import { bindActionCreators } from 'redux'

import { deleteCrawling } from '../../../store'

type StateProps = {
  crawling: ICrawling
}

type DispatchProps = {
  deleteCrawling: any
}

type OwnProps = {
  id: number
  onClose: () => void
  loading: boolean
}

type Props = StateProps & DispatchProps & OwnProps

function DeleteCrawling(props: Props) {
  const { crawling, onClose, deleteCrawling, loading } = props
  if (!crawling) return null

  const onConfirm = () => {
    deleteCrawling(crawling.id)
  }

  return (
    <Dialog
      isShown
      title='Delete Crawling'
      type='danger'
      confirmLabel='Delete'
      isConfirmLoading={loading}
      onCloseComplete={onClose}
      onConfirm={onConfirm}>
      <Text>
        Are you sure you want to delete the crawling number {crawling.id}?
      </Text>
    </Dialog>
  )
}

export default connect(
  (state: StoreState, ownProps: OwnProps): StateProps => ({ crawling: state.crawlings[ownProps.id] }),
  (dispatch): DispatchProps => bindActionCreators({deleteCrawling}, dispatch)
)(DeleteCrawling)
