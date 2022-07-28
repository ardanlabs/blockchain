import React from 'react'

type SpanProps = {
  id?: string
  classes?: string
  text?: string
}

const Span = ({ id, classes, text }: SpanProps) => (
  <span id={id} className={classes}>
    {text}
  </span>
)

export default Span
