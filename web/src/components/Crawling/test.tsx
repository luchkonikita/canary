import * as React from 'react'
import { shallow } from 'enzyme'
import { Crawling } from '.'
import { loadCrawling, showModal } from '../../store'

test('Crawling renders', () => {
  const crawling: ICrawling = {
    id: 1,
    url: 'http://example.com/sitemap.xml',
    createdAt: '2018-10-22T21:06:21.765601129+04:00',
    processed: false,
    pageResults: []
  }

  const props = {
    crawlingId: 1,
    loadCrawling,
    showModal,
    crawling
  }

  const rendered = shallow(<Crawling {...props} />)
  expect(rendered).toMatchSnapshot()
})
