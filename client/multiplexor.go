
/* mkp-userd/client/multiplexor.go */
package userd

import (
	"net"
	"time"
	"errors"
	
	protocol "../protocol"
)

type task struct {

	Key string
	ch chan CheckOutput
}

type Client struct {
	
	minimal time.Duration
	timeout time.Duration

	queuech chan task
	quit chan bool
}

func (m *Client) Dial() error {

	if m.quit != nil {
		return errors.New("Already dialled")
	}

	m.quit = make(chan bool,1)

	conn,err := net.Dial("tcp4","localhost:9999")
	if err != nil {
		return err
	}

	/* start queue gorountine */
	go func() {

		quit := m.quit
		queue := m.queuech

		defer conn.Close()

		for {
			select {
			case <- quit:
				return
				
			case ntask := <- queue:
		
				_,err := conn.Write(protocol.MakeTCheck(ntask.Key))
				if err != nil {
					ntask.ch <- CheckOutput{false,err}
					break
				}
				
				select {
				case <- time.After(m.timeout):
					ntask.ch <- CheckOutput{false,TimeOut}
					
					break
				case w := <- wrapConn(conn):
					
					if w.err != nil {
						ntask.ch <- CheckOutput{false,w.err}
						break
					}
					
					ok,err := protocol.IsCheckValid(w.data)
					ntask.ch <- CheckOutput{ok,err}
					break
				}
				break
			}
		}
	}()

	return nil
}

func (m *Client) Hangup() error {

	if m.quit == nil {
		return NotConnected
	}

	close(m.quit)
	return nil
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


