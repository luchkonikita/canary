import * as React from 'react'
import { Dialog, Pane, IconButton, Label, TextInput, SegmentedControl, Button } from 'evergreen-ui'
import { connect } from 'react-redux'
import { uniqueId, set, cloneDeep, debounce } from 'lodash'

import { createCrawling } from '../../../store'

type Props = {
  dispatch: (action: any) => Promise<any>
  onClose: () => void
  loading: boolean
}

type State = {
  url: string
  concurrency: number
  headers: {id: string, name: string, value: string}[]
}

const DEBOUNCE_RATE = 300

const concurrencyOptions = [
  { label: '1', value: '1' },
  { label: '2', value: '2' },
  { label: '4', value: '4' },
  { label: '6', value: '6' },
  { label: '8', value: '8' },
  { label: '10', value: '10' }
]

class CreateCrawling extends React.Component<Props, State> {
  state: State = {
    url: '',
    concurrency: 1,
    headers: []
  }

  render() {
    const { onClose, dispatch, loading } = this.props
    const { headers } = this.state

    const onConfirm = () => {
      dispatch(createCrawling(this.state))
    }

    return (
      <Dialog
        isShown
        title='Start New Crawling'
        confirmLabel='Start'
        isConfirmLoading={loading}
        onCloseComplete={onClose}
        onConfirm={onConfirm}>

        <Pane>
          <Pane display='flex' justifyContent='space-between' marginBottom={16}>
            <Label size={400} display='block' marginRight={16}>Sitemap URL</Label>
            <TextInput
              flexGrow={1}
              required
              placeholder='Enter Sitemap URL'
              onChange={this.inputHandler('url')} />
          </Pane>

          <Pane display='flex' justifyContent='space-between' alignItems='center' marginBottom={16}>
            <Label size={400} display='block' marginRight={16}>Concurrency</Label>
            <SegmentedControl
              flexGrow={1}
              options={concurrencyOptions}
              onChange={this.updateConcurrency} />
          </Pane>

          {headers.map((header, i) => (
            <Pane display='flex' justifyContent='space-between' marginBottom={16} key={header.id}>
              <TextInput
                flexGrow={1}
                marginRight={16}
                required
                placeholder={`Header ${i} Name`}
                defaultValue={header.name}
                onChange={this.inputHandler(`headers.${i}.name`)} />
              <TextInput
                flexGrow={1}
                marginRight={16}
                required
                placeholder={`Header ${i} Value`}
                defaultValue={header.value}
                onChange={this.inputHandler(`headers.${i}.value`)} />
              <IconButton icon='close' onClick={() => this.removeHeader(header.id)} />
            </Pane>
          ))}

          <Pane display='flex' justifyContent='center'>
            <Button appearance='ghost' onClick={this.addHeader}>
              + Add Request Header
            </Button>
          </Pane>
        </Pane>
      </Dialog>
    )
  }

  private inputHandler = (inputName: string) => {
    const handler = debounce(this.updateValue, DEBOUNCE_RATE)
    return (e: React.SyntheticEvent) => {
      e.persist()
      handler(e, inputName)
    }
  }

  private updateValue = (e: React.SyntheticEvent, field: string) => {
    const newState = cloneDeep(this.state)
    set(newState, field, (e.target as HTMLInputElement).value)
    this.setState(newState)
  }

  private updateConcurrency = (value: string) => {
    this.setState({concurrency: parseInt(value, 10)})
  }

  private addHeader = () => {
    const { headers } = this.state

    this.setState({
      headers: headers.concat({id: uniqueId(), name: '', value: ''})
    })
  }

  private removeHeader = (id: string) => {
    const { headers } = this.state

    this.setState({
      headers: headers.filter(header => header.id !== id)
    })
  }
}

export default connect(
  null,
  (dispatch) => ({dispatch})
)(CreateCrawling)
