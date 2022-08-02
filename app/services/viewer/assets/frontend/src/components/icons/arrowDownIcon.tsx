import React, { FC } from 'react'

interface arrowDownIconProps {
  isLast?: boolean
}

const ArrowDownIcon: FC<arrowDownIconProps> = (props: {
  isLast?: boolean
}) => {
  const { isLast } = props
  if (isLast) {
    return null
  }
  return (
    <svg style={{width: '24px', height: '24px'}} viewBox="0 0 24 24">
      <path fill="currentColor" d="M7.03 13.92H11.03V5L13.04 4.97V13.92H17.03L12.03 18.92Z" />
    </svg>
  )
}

export default ArrowDownIcon