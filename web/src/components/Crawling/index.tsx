import * as React from 'react'
import { Pane, Button, Text, Link, Strong, toaster, colors } from 'evergreen-ui'
import { parse, format } from 'date-fns'
import { connect } from 'react-redux'
import { bindActionCreators } from 'redux'

import { loadCrawling, showModal } from '../../store'
import StatusIcon from './StatusIcon'
import ProgressBar from './ProgressBar'

type OwnProps = {
  crawlingId: number
}

type StateProps = {
  crawling: ICrawling
}

type DispatchProps = {
  showModal: typeof showModal
  loadCrawling: any
}

type Props = OwnProps & StateProps & DispatchProps

type State = {
  expanded: boolean
}

export class Crawling extends React.Component<Props, State> {
  timeout: any
  state = {
    expanded: false
  }

  componentDidMount() {
    this.loadCrawling()
  }

  componentWillUnmount() {
    if (this.timeout) clearTimeout(this.timeout)
  }

  private async loadCrawling() {
    await this.props.loadCrawling(this.props.crawlingId)

    this.timeout = setTimeout(() => {
      const { crawling: { pageResults } } = this.props
      if (pageResults && pageResults.find(pr => pr.status === 0)) {
        this.loadCrawling()
      }
    }, 3000)
  }

  shouldComponentUpdate(nextProps: Props, nextState: State) {
    return nextProps.crawling !== this.props.crawling || nextState.expanded !== this.state.expanded
  }

  render() {
    const { crawling } = this.props
    const pageResults = crawling.pageResults || []

    const { pending, errored, successful, status } = aggregatePageResults(pageResults)
    const totalPageResults = pending.length + successful.length + errored.length
    const donePageResults = successful.length + errored.length

    return (
      <Pane>
        <Pane display='flex' alignItems='center' justifyContent='space-between' padding={16}>
          <Pane display='flex' alignItems='center'>
            <Pane display='flex' alignItems='center' justifyContent='center' width={20} height={20}>
              <StatusIcon type={status} />
            </Pane>

            <Pane display='flex' flexDirection='column' marginLeft={16}>
              <Text
                fontWeight={500}
                cursor='pointer'
                title='Click to copy the URL'
                onClick={() => this.copyUrl(crawling.url)}>
                {crawling.url}
              </Text>

              <Text size={300} marginTop={2}>
                {format(parse(crawling.createdAt), 'HH:mm / MMMM DD')}
              </Text>
            </Pane>
          </Pane>

          <Pane display='flex'>
            <Button
              onClick={() => this.setState({ expanded: !this.state.expanded })}>
              {this.state.expanded ? 'Hide Details' : 'View Details'}
            </Button>
            <Button
              marginLeft={16}
              onClick={this.requestDelete}>
              Delete
            </Button>
          </Pane>
        </Pane>

        {this.state.expanded && (
          <Pane padding={16} backgroundColor={colors.neutral[10]}>
            {pending.length > 0 && (
              <Pane marginBottom={16}>
                <ProgressBar
                  startTime={crawling.createdAt}
                  done={donePageResults}
                  total={totalPageResults} />
              </Pane>
            )}

            <ReportRow>
              Total pages: {totalPageResults}
            </ReportRow>

            <ReportRow>
              Pages with success status codes: {successful.length}
            </ReportRow>

            <ReportRow>
              Pages with error status codes: {errored.length}
            </ReportRow>

            {errored.length > 0 && (
              <Pane marginTop={16} overflowX='auto'>
                <Strong size={400}>
                  Problems found:
                </Strong>

                {errored.map(pageResult => <ErroredResult key={pageResult.url} pageResult={pageResult} />)}
              </Pane>
            )}
          </Pane>
        )}
      </Pane>
    )
  }

  private copyUrl(url: string) {
    (window.navigator as any).clipboard
      .writeText(url)
      .then(() => toaster.success('URL successfully copied to the clipboard', {duration: 1}))
      .catch(() => toaster.danger('Cannot copy the URL', {duration: 1}))
  }

  private requestDelete = () => {
    const { crawling, showModal } = this.props
    showModal('DeleteCrawling', {id: crawling.id})
  }
}

function ReportRow(props: { children: any }) {
  return (
    <Pane>
      <Text size={400}>
        {props.children}
      </Text>
    </Pane>
  )
}

function ErroredResult(props: { pageResult: IPageResult }) {
  const { pageResult } = props
  return (
    <ReportRow>
      ({pageResult.status})&nbsp;
      <Link size={400} appearance='blue' href={pageResult.url} target='_blank'>
        {pageResult.url}
      </Link>
    </ReportRow>
  )
}

function aggregatePageResults(pageResults: IPageResult[]) {
  let successful: IPageResult[] = []
  let errored: IPageResult[] = []
  let pending: IPageResult[] = []

  pageResults.forEach((pageResult: IPageResult) => {
    if (pageResult.status === 200) {
      successful.push(pageResult)
    } else if (pageResult.status === 0) {
      pending.push(pageResult)
    } else {
      errored.push(pageResult)
    }
  })

  let status: 'success' | 'error' | 'pending' | 'none'

  if (pending.length) {
    status = 'pending'
  } else if (errored.length) {
    status = 'error'
  } else if (successful.length) {
    status = 'success'
  } else {
    status = 'none'
  }

  return { pending, errored, successful, status }
}

export default connect(
  (state: StoreState, ownProps: { crawlingId: number }): StateProps => {
    return { crawling: state.crawlings[ownProps.crawlingId] }
  },
  (dispatch): DispatchProps => {
    return bindActionCreators({ loadCrawling, showModal }, dispatch)
  }
)(Crawling)
