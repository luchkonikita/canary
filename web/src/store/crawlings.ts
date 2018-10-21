import { Dispatch } from 'redux'
import { omit } from 'lodash'
import { stringify } from 'qs'
import { toaster } from 'evergreen-ui'

import { setModalLoading, hideModal } from './modal'

function denormalizeCrawlings(crawlings: ICrawling[]): CrawlingsState {
  return crawlings.reduce((memo: any, current: ICrawling) => {
    memo[current.id] = current
    return memo
  }, {})
}

type StoreCrawlingsAction = StoreAction<'STORE_CRAWLINGS', { crawlings: ICrawling[] }>
type StoreCrawlingAction = StoreAction<'STORE_CRAWLING', { crawling: ICrawling }>
type RemoveCrawlingAction = StoreAction<'REMOVE_CRAWLING', { crawlingId: number }>

type CrawlingsAction = StoreCrawlingsAction | StoreCrawlingAction | RemoveCrawlingAction

function storeCrawlings(crawlings: ICrawling[]): StoreCrawlingsAction {
  return { type: 'STORE_CRAWLINGS', payload: { crawlings } }
  }

function storeCrawling(crawling: ICrawling): StoreCrawlingAction {
  return { type: 'STORE_CRAWLING', payload: { crawling } }
  }

function removeCrawling(crawlingId: number): RemoveCrawlingAction {
  return { type: 'REMOVE_CRAWLING', payload: { crawlingId } }
}

export function loadCrawlings() {
  return async function (dispatch: Dispatch) {
    const result = await fetch('http://localhost:4000/crawlings')
    const crawlings = await result.json()
    dispatch(storeCrawlings(crawlings))
  }
}

export function loadCrawling(id: number) {
  return async function (dispatch: Dispatch) {
    const result = await fetch(`http://localhost:4000/crawlings/${id}`)
    const crawling = await result.json()
    return dispatch(storeCrawling(crawling))
  }
}

export function deleteCrawling(id: number) {
  return async function (dispatch: Dispatch) {
    dispatch(setModalLoading(true))
    const result = await fetch(`http://localhost:4000/crawlings/${id}`, { method: 'DELETE' })

    if (result.status === 200) {
      dispatch(removeCrawling(id))
      dispatch(hideModal())
    } else {
      toaster.danger('Something went wrong')
      dispatch(setModalLoading(false))
    }
  }
}

export function createCrawling(data: {url: string, concurrency: number, headers: {name: string, value: string}[]}) {
  return async function (dispatch: Dispatch) {
    dispatch(setModalLoading(true))

    const params: {[index: string]: any} = {
      url: data.url,
      concurrency: data.concurrency,
    }

    data.headers.forEach((header, index) => {
      params[`headers.${index}.name`] = header.name
      params[`headers.${index}.value`] = header.value
    })

    const body = new URLSearchParams(stringify(params))

    const result = await fetch('http://localhost:4000/crawlings', { method: 'POST', body })

    if (result.status === 201) {
      const crawling = await result.json()

      dispatch(storeCrawling(crawling))
      dispatch(hideModal())
    } else {
      const message = await result.text()
      toaster.danger(`Failed to start: ${message}`)
      dispatch(setModalLoading(false))
    }
  }
}

export function reducer(state: CrawlingsState = [], action: CrawlingsAction): CrawlingsState {
  switch (action.type) {
    case 'STORE_CRAWLINGS':
      return denormalizeCrawlings(action.payload.crawlings)
    case 'STORE_CRAWLING':
      return { ...state, ...denormalizeCrawlings([action.payload.crawling]) }
    case 'REMOVE_CRAWLING':
      return omit(state, action.payload.crawlingId)
    default:
      return state
  }
}
