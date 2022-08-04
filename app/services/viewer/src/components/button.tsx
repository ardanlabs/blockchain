import React, { FC } from 'react'

interface ButtonProps {
  clickHandler: Function,
  classes?: string,
  id?: string,
  text: string,
  disabled?: boolean,
}

const Button: FC<ButtonProps> = (ButtonProps) => {
  const { clickHandler, classes, id, text, disabled } = ButtonProps
  return (
    <button type="button" id={id} className={`btn btn-block ${classes}`} disabled={disabled} onClick={() => clickHandler(id)}>
      { text }
    </button>
  )
}

export default Button
