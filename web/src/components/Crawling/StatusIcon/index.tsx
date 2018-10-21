import * as React from 'react'
import { Pane, CheckCircleIcon, DangerIcon, Spinner, colors } from 'evergreen-ui'

const iconProps = { iconHeight: 20, iconWidth: 20, iconSize: 20 }

export default function StatusIcon(props: { type: string }) {
  switch (props.type) {
    case 'success':
      return <CheckCircleIcon {...iconProps} color={colors.green['300']} />
    case 'error':
      return <DangerIcon {...iconProps} color={colors.red['300']} />
    case 'pending':
      return <Spinner size={20} />
    default:
      return <Pane height={20} width={20} />
  }
}
