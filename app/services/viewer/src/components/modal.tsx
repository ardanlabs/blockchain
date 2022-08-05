import React, { Component } from "react";

interface modalProps {
  show: boolean,
  children: JSX.Element[],
  classes?: string,
}

class Modal extends Component<modalProps> {
  render() {
    const { show, classes, children } = this.props
    if(!show) {
      return null
    }
    return (
      <div className={`my-modal ${classes}`}>
        {children}
      </div>
    )
  }
}

export default Modal