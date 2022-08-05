import React, { Component } from 'react'
// Types
import { block } from '../../types/index.d'
// Components
import Button from './button'
import Modal from './modal'
import TransactionTable from './transactionTable'
// Icons
import CloseIcon from './icons/closeIcon'
import ArrowDownIcon from './icons/arrowDownIcon'
import BlockDetails from './blockDetails'

interface BlockProps {
  nodeID: number,
  blockNumber: number,
  block: block,
  successfullNode: boolean,
}
export type State = {
  showTransactions: boolean,
  modalTransactions: JSX.Element[],
  showBlockDetails: boolean,
}
class Block extends Component<BlockProps, State> {
  constructor(props: any) {
    super(props)
    this.state = {
      showTransactions: false,
      modalTransactions: [],
      showBlockDetails: false,
    }
    this.triggerBlockDetailsModal = this.triggerBlockDetailsModal.bind(this)
  }
  clickHandler = (btnId: string) => {
    switch (btnId) {
      case 'block-details-modal-btn':
        this.triggerBlockDetailsModal()
        break;
      case 'transactions-modal-btn':
        this.openTransactionsModal(this.props.block)
        break;
      default:
        break;
    }
  }
  triggerBlockDetailsModal() {
    this.setState({
      showBlockDetails: !this.state.showBlockDetails,
    })
  }
  openTransactionsModal(event: block) {
    const elements: JSX.Element[] = []
    event.trans.forEach(element => {
      const isLast = event.trans.indexOf(element) === event.trans.length - 1
      elements.push(
        <div key={element.r}>
          <TransactionTable key={element.r} transaction={element} />
          <ArrowDownIcon key={`${element.r}-arrow`} isLast={ isLast } />
        </div>
      )
    })
    this.setState({
      modalTransactions: elements,
      showTransactions: !this.state.showTransactions,
    })
  }
  hideTransactionsModal = () => {
    this.setState({showTransactions: false})
  }
  render() {
    const { nodeID, blockNumber, block, successfullNode } = this.props
    const { clickHandler, hideTransactionsModal, triggerBlockDetailsModal } = this
    const { showTransactions, showBlockDetails, modalTransactions } = this.state
    const id = `Block-${nodeID}-${blockNumber}`
    let extraClass: string = ''
  
    if (successfullNode) {
      extraClass = ' mine'
    }
  
    const classes = `block-details d-flex align-items-center flex-column justify-content-center${extraClass}`
    return (
      <div id={id} className={classes}>
        <strong>Block Number: </strong>{ blockNumber }
        <Button {...{clickHandler: clickHandler, classes: 'btn-outline-primary my-2', id: 'block-details-modal-btn', text: 'Block details', disabled: false}} />
        <Button {...{clickHandler: clickHandler, classes: 'btn-outline-secondary my-2', id: 'transactions-modal-btn', text: 'Transaction details', disabled: false}} />
        {/* Block Details */}
        <Modal classes="block-details-modal" show={showBlockDetails}>
          <CloseIcon classes="close-modal-button" onClick={() => triggerBlockDetailsModal()} />
          <div id="block-details-content">
            <BlockDetails
              {...{
                nodeID,
                blockNumber: block.block.number,
                successfullNode,
                block,
              }}
            />
          </div>
        </Modal>
        {/* Transactions Details */}
        <Modal classes="transactions" show={showTransactions}>
          <CloseIcon classes="close-modal-button" onClick={() => hideTransactionsModal()} />
          <div id="transactions-content">
            {modalTransactions}
          </div>
        </Modal>
      </div>
    )
  }
}

export default Block
