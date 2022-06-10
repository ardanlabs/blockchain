package main

import (
	"crypto/sha256"
	"encoding/json"
	"hash/fnv"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

const cycleDuration = 5 * time.Second

// block represents a block of data that is mined.
type block struct {
	Number        uint64
	PrevBlockHash string
	TimeStamp     uint64
}

// newBlock constructs a block from the previous block.
func newBlock(prevBlock block) block {
	return block{
		Number:        prevBlock.Number + 1,
		PrevBlockHash: prevBlock.hash(),
		TimeStamp:     uint64(time.Now().UTC().UnixMilli()),
	}
}

// hash generates the hash for this block.
func (b block) hash() string {
	const ZeroHash string = "0x0000000000000000000000000000000000000000000000000000000000000000"

	data, err := json.Marshal(b)
	if err != nil {
		return ZeroHash
	}

	hash := sha256.Sum256(data)
	return hexutil.Encode(hash[:])
}

// =============================================================================

// registry maintains a list of nodes participating in selection.
type registry struct {
	name  string
	mu    sync.RWMutex
	nodes map[string]*node
}

// newRegistry constructs a registry for use.
func newRegistry(n *node) *registry {
	return &registry{
		name: n.name,
		nodes: map[string]*node{
			n.name: n,
		},
	}
}

// add a new node to the registry.
func (r *registry) add(node *node) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.nodes[node.name] = node
}

// nodeList returns a slice of the current registry.
func (r *registry) nodeList() []*node {
	r.mu.RLock()
	defer r.mu.RUnlock()

	nodes := make([]*node, 0, len(r.nodes))
	for _, node := range r.nodes {
		nodes = append(nodes, node)
	}
	return nodes
}

// sendBlock sends the block to all registered nodes.
func (r *registry) sendBlock(b block) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, node := range r.nodes {
		if node.name != r.name {
			node.sendBlock <- b
		}
	}
}

// registerNode sends the node to all registered nodes.
func (r *registry) registerNode(n *node) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, node := range r.nodes {
		if node.name != r.name {
			node.registerNode <- n
		}
	}
}

// =============================================================================

// node represents a node running as a process.
type node struct {
	name         string
	latestBlock  block
	wg           sync.WaitGroup
	ticker       *time.Ticker
	shut         chan struct{}
	registry     *registry
	sendBlock    chan block
	registerNode chan *node
	getNodes     chan chan []*node
}

// newNode gets the node up and running. The origin node should be the
// first node we start so other nodes can find other peers.
func newNode(name string, originNode *node) *node {
	n := node{
		name:         name,
		ticker:       time.NewTicker(cycleDuration),
		shut:         make(chan struct{}),
		sendBlock:    make(chan block, 10),
		registerNode: make(chan *node),
		getNodes:     make(chan chan []*node),
	}

	// Register this node in the local registry.
	n.registry = newRegistry(&n)

	// If this is not the origin node, ask the list of registered
	// nodes for their registry.
	if originNode != nil {
		log.Println(n.name, ":ask", originNode.name, "for registry")

		ch := make(chan []*node)
		originNode.getNodes <- ch
		nodes := <-ch

		for _, node := range nodes {
			log.Println(n.name, ":register", node.name, "on", n.name)
			n.registry.add(node)
		}

		// Register this node to the registry of the other nodes.
		n.registry.registerNode(&n)
	}

	return &n
}

// run gets the node up and running.
func (n *node) run() {
	n.wg.Add(1)
	go func() {
		defer n.wg.Done()

		log.Println(n.name, ":starting node")

		// Start this on the 5 second marks.
		n.resetTicker(5 * time.Second)

		for {

			select {
			case <-n.ticker.C:
				n.performWork()

			case b := <-n.sendBlock:
				log.Println(n.name, ":node received block", b.hash())
				n.latestBlock = b

			case otherNode := <-n.registerNode:
				log.Println(n.name, ":register node", otherNode.name, "on", n.name)
				n.registry.add(otherNode)

			case ch := <-n.getNodes:
				log.Println(n.name, ":send registery")
				ch <- n.registry.nodeList()

			case <-n.shut:
				log.Println(n.name, ":node shutting down")
				return
			}

			n.resetTicker(0)
		}
	}()
}

// resetTicker makes sure the next tick happens on the described cadence.
func (n *node) resetTicker(waitOnSecond time.Duration) {
	nextTick := time.Now().Add(cycleDuration).Round(waitOnSecond)
	diff := time.Until(nextTick)
	n.ticker.Reset(diff)
}

// shutdown terminates the node from existence.
func (n *node) shutdown() {
	close(n.shut)
	n.wg.Wait()
}

// performWork represents the work to perform on each 12 second cycle.
func (n *node) performWork() {
	log.Println(n.name, "******************* CYCLE *******************")

	selectedNode := n.selection()
	switch selectedNode {
	case n.name:
		log.Println(n.name, ":SELECTED")
		n.mineNewBlock()
	default:
	}
}

// selection selects an index from 0 to 2.
func (n *node) selection() string {

	// Sort the current list of registered names.
	nodes := n.registry.nodeList()
	names := make([]string, len(nodes))
	for i, node := range nodes {
		names[i] = node.name
	}
	sort.Strings(names)

	// Based on the latest block, pick an index number from
	// the registry.
	h := fnv.New32a()
	h.Write([]byte(n.latestBlock.hash()))
	integerHash := h.Sum32()
	i := integerHash % uint32(len(names))

	// Return the name of the node selected.
	return names[i]
}

// mineNewBlock creates a new block and sends that to the p2p network.
func (n *node) mineNewBlock() {
	b := newBlock(n.latestBlock)
	n.latestBlock = b

	// Send block to all other nodes in the registry.
	n.registry.sendBlock(b)
}

// =============================================================================

// simulation represents a set of nodes talking to each other.
type simulation struct {
	nodes map[string]*node
}

// newSimulation starts 3 nodes for the simulation.
func newSimulation() *simulation {
	nodeA := newNode("nodeA", nil)
	nodeA.run()

	time.Sleep(2 * time.Second)

	nodeB := newNode("nodeB", nodeA)
	nodeB.run()

	time.Sleep(2 * time.Second)

	nodeC := newNode("nodeC", nodeA)
	nodeC.run()

	return &simulation{
		nodes: map[string]*node{
			nodeA.name: nodeA,
			nodeB.name: nodeB,
			nodeC.name: nodeC,
		},
	}
}

// shutdown turns all the nodes off.
func (s *simulation) shutdown() {
	for _, n := range s.nodes {
		n.shutdown()
	}
}
