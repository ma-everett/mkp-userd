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

type Control struct {
	conn    net.Conn
	timeout time.Duration
}

func (c *Control) Dial() error {

	conn, err := net.Dial("tcp4", "localhost:9998")
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

func (c *Control) Set(key string) (bool, error) {

	if c.conn == nil {
		return false, NotConnected
	}

	_, err := c.conn.Write(protocol.MakeTSet(key))
	if err != nil {
		return false, err
	}

	/* TODO: check the length  */

	/* wait for remote connection */
	select {
	case <-time.After(c.timeout):
		return false, TimeOut

	case w := <-wrapConn(c.conn):

		if w.err != nil {
			return false, w.err
		}

		ok, err := protocol.IsSetValid(w.data)
		if err != nil {
			return false, err
		}

		return ok, nil
		break
	}

	return false, NotImplemented
}

func (c *Control) Remove(key string) (bool, error) {

	if c.conn == nil {
		return false, NotConnected
	}

	_, err := c.conn.Write(protocol.MakeTRemove(key))
	if err != nil {
		return false, err
	}

	/* TODO: check the length */

	select {
	case <-time.After(c.timeout):
		return false, TimeOut

	case w := <-wrapConn(c.conn):

		if w.err != nil {
			return false, w.err
		}

		ok, err := protocol.IsRemoveValid(w.data)
		if err != nil {
			return false, err
		}

		return ok, nil
		break
	}

	return false, NotImplemented
}

func (c *Control) Check(key string) (bool, error) {

	if c.conn == nil {
		return false, NotConnected
	}

	_, err := c.conn.Write(protocol.MakeTCheck(key))
	if err != nil {
		return false, err
	}

	select {
	case <-time.After(c.timeout):
		return false, TimeOut

	case w := <-wrapConn(c.conn):

		if w.err != nil {
			return false, w.err
		}

		ok, err := protocol.IsCheckValid(w.data)
		if err != nil {
			return false, err
		}

		return ok, nil
		break
	}

	return false, NotImplemented
}

func (c *Control) Purge() (bool, error) {

	if c.conn == nil {
		return false, NotConnected
	}

	_, err := c.conn.Write(protocol.MakeTPurge())
	if err != nil {
		return false, err
	}

	select {
	case <-time.After(c.timeout):
		return false, TimeOut

	case w := <-wrapConn(c.conn):

		if w.err != nil {
			return false, w.err
		}

		ok, err := protocol.IsPurgeValid(w.data)
		if err != nil {
			return false, err
		}

		return ok, nil
		break
	}

	return false, NotImplemented
}

func NewControl(timeout time.Duration) *Control {

	c := new(Control)
	c.timeout = timeout

	return c
}

type wrapper struct {
	data []byte
	err  error
}

func wrapConn(conn net.Conn) chan wrapper {

	ch := make(chan wrapper, 1)
	go func() {

		b := make([]byte, 512)
		n, err := conn.Read(b)

		if err != nil {
			ch <- wrapper{nil, err}
			return
		}

		var w wrapper
		w.data = make([]byte, n)
		copy(w.data, b[:n])
		w.err = nil

		ch <- w
		return
	}()

	return ch
}
