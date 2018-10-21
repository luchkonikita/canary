import * as React from 'react'
import { Provider } from 'react-redux'

import store from '../../store'
import CrawlingsList from '../CrawlingsList'
import Modal from '../Modal'

export default class App extends React.Component<{}, {}> {
  render() {
    return (
      <Provider store={store}>
        <div>
          <CrawlingsList />
          <Modal />
        </div>
      </Provider>
    )
  }
}
