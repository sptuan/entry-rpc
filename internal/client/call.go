package client

type Call struct {
	Seq           uint64
	ServiceMethod string
	Args          interface{}
	Returns       interface{}
	Error         error
	Done          chan *Call
}

// to support async call
func (c *Call) done() {
	c.Done <- c
}

func (client *Client) RegisterCall(call *Call) (uint64, error) {
	client.mu.Lock()
	defer client.mu.Unlock()
	// check if shutdown
	if client.shutdown || client.closing {
		return 0, ErrShutdown
	}
	// set call seq
	call.Seq = client.seq
	client.pending[client.seq] = call
	client.seq++
	return call.Seq, nil
}

func (client *Client) RemoveCall(seq uint64) *Call {
	client.mu.Lock()
	defer client.mu.Unlock()
	call := client.pending[seq]
	delete(client.pending, seq)
	return call
}

func (client *Client) terminateCalls(err error) {
	client.sending.Lock()
	defer client.sending.Unlock()
	client.mu.Lock()
	defer client.mu.Unlock()
	client.shutdown = true
	for _, call := range client.pending {
		call.Error = err
		call.done()
	}
}
