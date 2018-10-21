import * as React from 'react'
import { shallow } from 'enzyme'
import { CrawlingsList } from '.'

test('CrawlingsList renders', () => {
  const crawlings: ICrawling[] = [
    {
      id: 1,
      url: 'http://example.com/sitemap.xml',
      createdAt: '2018-10-22T21:06:21.765601129+04:00',
      processed: false,
      pageResults: []
    }
  ]

  const props = {
    crawlings,
    dispatch: (action: any) => {}
  }

  const rendered = shallow(<CrawlingsList {...props} />)
  expect(rendered).toMatchSnapshot()
})
