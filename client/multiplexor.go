
/* mkp-userd/client/multiplexor.go */
package userd

import (
	"net"
	"time"
	
	protocol "../protocol"
)

type task struct {

	Key string
	ch chan CheckOutput
}

type Client struct {

	
	minimal time.Duration
	timeout time.Duration
	conn    net.Conn

	queuech chan task
}

func (m *Client) Dial() error {

	conn,err := net.Dial("tcp4","localhost:9999")
	if err != nil {
		return err
	}

	m.conn = conn

	/* start queue gorountine */
	/* TODO: the goroutine needs a WorkGroup to close properly */
	go func() {

		for {
			ntask := <- m.queuech
			_,err := m.conn.Write(protocol.MakeTCheck(ntask.Key))
			if err != nil {
				ntask.ch <- CheckOutput{false,err}
				continue
			}

			select {
			case <- time.After(m.timeout):
				ntask.ch <- CheckOutput{false,TimeOut}
				
				break
			case w := <- wrapConn(m.conn):
				
				if w.err != nil {
					ntask.ch <- CheckOutput{false,w.err}
					break
				}
				
				ok,err := protocol.IsCheckValid(w.data)
				ntask.ch <- CheckOutput{ok,err}
				break
			}
		}
	}()

	return nil
}

func (m *Client) Hangup() error {

	if m.conn == nil {
		return NotConnected
	}

	return m.conn.Close()
}

func (m *Client) Check(key string) chan CheckOutput {

	outch := make(chan CheckOutput,1)
	inch := make(chan CheckOutput,1)

	m.queuech <- task{key,inch}
	
	go func() {
		
		start := time.Now()
		
		select {
		case <- time.After(m.timeout):
			
			outch <- CheckOutput{false,TimeOut}
			break

		case data := <- inch:
			
			/* wait the minimal time */
			d := time.Now().Sub(start)
			if d < m.minimal {
			
				time.Sleep(m.minimal - d)
			}
			
			outch <- data
			break
		}
	}()

	return outch
}

func (m *Client) Checked(key string) (bool,error) {

	d := <- m.Check(key)
	return d.Checked,d.Error
}

func NewClient(minimal,timeout time.Duration) *Client {

	m := new(Client)
	m.minimal = minimal
	m.timeout = timeout
	m.queuech = make(chan task,100)

	return m
}

type CheckOutput struct {

	Checked bool
	Error error
}


