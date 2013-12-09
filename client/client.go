
/* mkp-userd/client/client.go */
package userd

import (
	"errors"
	"net"
	"time"

	protocol "../protocol"
)

var (
	NotConnected   = errors.New("Not Connected")
	LengthMismatch = errors.New("Length Mismatch")
	NotImplemented = errors.New("Not Implemented")
	InvalidData    = errors.New("Invalid Data")
	TimeOut        = errors.New("Timeout")
)

type Client struct {

	minimal time.Duration /* the minimal time a check should take */
	timeout time.Duration /* the maximum time a check should take */
	conn net.Conn
}

func (c *Client) Dial() error {

	conn,err := net.Dial("tcp4","localhost:9999")
	if err != nil {
		return err
	}

	c.conn = conn

	return nil
}

func (c *Client) Hangup() error {

	if c.conn == nil {
		return NotConnected
	}

	return c.conn.Close()
}

func (c *Client) Check(key string) (bool,error) {

	start := time.Now()

	if c.conn == nil {
		return minimal(false,NotConnected,start,c.minimal)
	}

	_,err := c.conn.Write(protocol.MakeTCheck(key))
	if err != nil {
		return minimal(false,err,start,c.minimal)
	}

	/* TODO: check the length 
	if n != len(str) {
		return minimal(false,LengthMismatch,start,c.minimal)
	}
        */

	/* wait for remote connection */
	select {
	case <- time.After(c.timeout): 
		return minimal(false,TimeOut,start,c.minimal)

	case w := <- wrapConn(c.conn):
		
		if w.err != nil {
			return minimal(false,w.err,start,c.minimal)
		}

		ok,err := protocol.IsCheckValid(w.data)
		if err != nil {
			return minimal(false,err,start,c.minimal)
		}

		return minimal(ok,nil,start,c.minimal)
		break
	}

	return minimal(false,nil,start,c.minimal)
}

/* this is to ensure that all _check_ calls are at least of a set time 
 * inorder to prevent gaming. 
 */
func minimal(status bool,err error,start time.Time,finish time.Duration) (bool,error) {

	d := time.Now().Sub(start)
	if d < finish {
		time.Sleep(finish - d)
	}
	return status,err
}

func NewClient(minimal, timeout time.Duration) *Client {

	c := new(Client)
	c.minimal = minimal
	c.timeout = timeout
	
	return c
}


type Control struct {

	conn net.Conn
	timeout time.Duration
}

func (c *Control) Dial() error {

	conn,err := net.Dial("tcp4","localhost:9998")
	if err != nil {
		return err
	}

	c.conn = conn

	return nil
}

func (c *Control) Hangup() error {

	if c.conn == nil {
		return NotConnected
	}

	return c.conn.Close()
}

func (c *Control) Set(key string) (bool,error) {

	if c.conn == nil {
		return false,NotConnected
	}

	_,err := c.conn.Write(protocol.MakeTSet(key))
	if err != nil {
		return false,err
	}

	/* TODO: check the length  */

	/* wait for remote connection */
	select {
	case <- time.After(c.timeout): 
		return false,TimeOut

	case w := <- wrapConn(c.conn):
		
		if w.err != nil {
			return false,w.err
		}

		ok,err := protocol.IsSetValid(w.data)
		if err != nil {
			return false,err
		}

		return ok,nil
		break
	}

	return false,NotImplemented
}

func (c *Control) Remove(key string) (bool,error) {

	if c.conn == nil {
		return false,NotConnected
	}

	_,err := c.conn.Write(protocol.MakeTRemove(key))
	if err != nil {
		return false,err
	}

	/* TODO: check the length */

	select {
	case <- time.After(c.timeout):
		return false,TimeOut
	
	case w := <- wrapConn(c.conn):

		if w.err != nil {
			return false,w.err
		}

		ok,err := protocol.IsRemoveValid(w.data)
		if err != nil {
			return false,err
		}
		
		return ok,nil
		break
	}

	return false,NotImplemented
}

func (c *Control) Check(key string) (bool,error) {

	if c.conn == nil {
		return false,NotConnected
	}

	_,err := c.conn.Write(protocol.MakeTCheck(key))
	if err != nil {
		return false,err
	}
	
	select {
	case <- time.After(c.timeout):
		return false,TimeOut
		
	case w := <- wrapConn(c.conn):

		if w.err != nil {
			return false,w.err
		}

		ok,err := protocol.IsCheckValid(w.data)
		if err != nil {
			return false,err
		}

		return ok,nil
		break
	}

	return false,NotImplemented
}

func (c *Control) Purge() (bool,error) {

	if c.conn == nil {
		return false,NotConnected
	}

	_,err := c.conn.Write(protocol.MakeTPurge())
	if err != nil {
		return false,err
	}

	select {
	case <- time.After(c.timeout):
		return false,TimeOut

	case w := <- wrapConn(c.conn):

		if w.err != nil {
			return false,w.err
		}

		ok,err := protocol.IsPurgeValid(w.data)
		if err != nil {
			return false,err
		}

		return ok,nil
		break
	}

	return false,NotImplemented
}

func NewControl(timeout time.Duration) *Control {

	c := new(Control)
	c.timeout = timeout

	return c
}



type wrapper struct {

	data []byte
	err error
}

func wrapConn(conn net.Conn) chan wrapper {

	ch := make(chan wrapper,1)
	go func() {
		
		b := make([]byte,512)
		n,err := conn.Read(b)
		
		if err != nil {
			ch <- wrapper{nil,err}
			return
		}

		var w wrapper
		w.data = make([]byte,n)
		copy(w.data,b[:n])
		w.err = nil

		ch <- w
		return
	}()

	return ch
}
