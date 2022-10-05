package state

// Reorganize corrects an identified fork. No mining is allowed to take place
// while this process is running. New transactions can be placed into the mempool.
func (s *State) Reorganize() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Don't allow mining to continue.
	s.allowMining = false

	// Reset the state of the blockchain node.
	s.db.Reset()

	// Resync the state of the blockchain.
	s.resyncWG.Add(1)
	go func() {
		s.evHandler("state: Resync: started: *****************************")
		defer func() {
			s.turnMiningOn()
			s.evHandler("state: Resync: completed: *****************************")
			s.resyncWG.Done()
		}()

		s.Worker.Sync()
	}()

	return nil
}

// turnMiningOn sets the allowMining flag back to true.
func (s *State) turnMiningOn() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.allowMining = true
}
