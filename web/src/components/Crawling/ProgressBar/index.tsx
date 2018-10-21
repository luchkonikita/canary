import * as React from 'react'
import { parse, addSeconds, differenceInSeconds, distanceInWords } from 'date-fns'
import { Pane, Text, colors } from 'evergreen-ui'

type Props = {
  done: number,
  total: number,
  startTime: string
}

export default function ProgressBar(props: Props) {
  const { done, total } = props

  let timeLeft: string
  const doneRate = (done / total) * 100
  const leftRate = 100 - doneRate

  if (done > 0) {
    const startTime = parse(props.startTime)
    const currentTime = new Date()
    const diffInSeconds = differenceInSeconds(currentTime, startTime)
    const approximateLeftSeconds = (diffInSeconds / doneRate) * leftRate
    const endTime = addSeconds(currentTime, approximateLeftSeconds)
    timeLeft = distanceInWords(currentTime, endTime)
  } else {
    timeLeft = 'Calculating'
  }

  return (
    <Pane>
      <Pane display='flex' justifyContent='space-between' marginBottom={4}>
        <Text size={400}>
          Time left: {timeLeft}
        </Text>
        <Text size={400}>
          Crawled: {done} / Total: {total}
        </Text>
      </Pane>

      <Pane height={6} borderRadius={3} backgroundColor={colors.neutral[100]}>
        <Pane
          width={doneRate + '%'}
          height={6}
          borderRadius={3}
          backgroundColor={colors.neutral[500]}
          transition={'width 2000ms'} />
      </Pane>
    </Pane>
  )
}
