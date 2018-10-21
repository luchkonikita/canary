import { createStore, applyMiddleware, combineReducers } from 'redux'
import thunkMiddleware from 'redux-thunk'

import { reducer as crawlingsReducer } from './crawlings'
import { reducer as modalReducer } from './modal'

export { loadCrawlings, loadCrawling, createCrawling, deleteCrawling } from './crawlings'
export { showModal, hideModal } from './modal'

export default createStore(
  combineReducers<StoreState>({
    crawlings: crawlingsReducer,
    modal: modalReducer
  }),
  applyMiddleware(thunkMiddleware)
)
