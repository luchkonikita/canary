import * as React from 'react'
import { Pane, Text, Button, colors } from 'evergreen-ui'
import { connect } from 'react-redux'
import { values } from 'lodash'

import { loadCrawlings, showModal } from '../../store'
import Crawling from '../Crawling'

interface IProps {
  crawlings: ICrawling[]
  dispatch: (action: any) => void
}

export class CrawlingsList extends React.Component<IProps, {}> {
  componentDidMount() {
    this.props.dispatch(loadCrawlings())
  }

  render() {
    const crawlings = this.props.crawlings
      .sort((a, b) => new Date(a.createdAt) < new Date(b.createdAt) ? 1 : -1)

    return (
      <Pane
        display='flex'
        flexDirection='column'
        height='100%'
        padding={24}>
        <Pane
          width={660}
          marginX='auto' >
          <Pane
            display='flex'
            justifyContent='space-between'
            alignItems='center'
            padding={16}
            marginBottom={16}>
            <Text>Your Crawlings</Text>
            <Button onClick={this.requestCreate} appearance='green'>
              Create Crawling
            </Button>
          </Pane>

          <Pane
            backgroundColor={colors.neutral['3']}
            borderColor={colors.neutral['30']}
            borderWidth={1}
            borderStyle='solid'
            minHeight={72}>
            {crawlings.map((crawling, i) => (
              <Pane
                key={crawling.id}
                borderBottomColor={colors.neutral['30']}
                borderBottomStyle='solid'
                borderBottomWidth={i < crawlings.length - 1 ? 1 : 0}
                minHeight={72}>
                <Crawling crawlingId={crawling.id} />
              </Pane>
            ))}
          </Pane>
        </Pane>
      </Pane>
    )
  }

  private requestCreate = () => {
    const { dispatch } = this.props
    dispatch(showModal('CreateCrawling', {}))
  }
}

export default connect(
  (state: StoreState) => ({ crawlings: values(state.crawlings) }),
  (dispatch) => ({dispatch})
)(CrawlingsList)
