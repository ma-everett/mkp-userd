/* mkp-userd/service.go */

package main

import (
	"errors"
	"flag"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	protocol "./protocol"
)

func main() {

	var optServiceAddr = flag.String("s", "localhost:9998", "net for service address")
	var optClientAddr = flag.String("c", "localhost:9999", "net for client address")

	flag.Parse()

	/* Service Address */
	serviceAddr, err := parseAddress(*optServiceAddr)
	if err != nil {
		log.Fatalf("failed to parse service address - %v\n", err)
	}

	addr, err := net.ResolveTCPAddr("tcp4", serviceAddr)
	if err != nil {
		log.Fatalf("unable to resolve service address - %v\n", err)
	}

	sconn, err := net.ListenTCP("tcp4", addr)
	if err != nil {
		log.Fatalf("failed to listen on TCP - %v\n", err)
	}
	defer sconn.Close()

	/* Client Address */
	clientAddr, err := parseAddress(*optClientAddr)
	if err != nil {
		log.Fatalf("failed to parse client address - %v\n", err)
	}

	addr, err = net.ResolveTCPAddr("tcp4", clientAddr)
	if err != nil {
		log.Fatalf("unable to resolve client address - %v\n", err)
	}

	cconn, err := net.ListenTCP("tcp4", addr)
	if err != nil {
		log.Fatalf("failed to listen on TCP - %v\n", err)
	}
	defer cconn.Close()

	/* create a store for all the keys : */
	store := NewStore()
	storech := make(chan *Store, 1)

	storech <- store

	/* create a new work group : */
	ctrl := NewWorkGroup()

	/* service client connections : */
	go func() {

		quit := ctrl.Start()
		defer ctrl.Finish()

		infoch := NewInformation()

		for {
			select {
			case <-quit:
				return

			case conn := <-wrapAcceptTCP(sconn): /* wait on service connection from service port */

				if conn == nil {
					continue
				}

				go handle(conn, ctrl, storech, infoch)
				break
			}
		}
	}()

	go func() {

		quit := ctrl.Start()
		defer ctrl.Finish()

		infoch := NewInformation()

		for {
			select {
			case <-quit:
				return

			case conn := <-wrapAcceptTCP(cconn): /* wait on client connection from client port */

				if conn == nil {
					continue
				}

				/* start multiplexor */

				//m := NewMultiplexor(conn,storech)
				//m.Start(ctrl)
				//break

				 
				go handleClient(conn, ctrl, storech, infoch)
				break
			}
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, os.Interrupt)

	<-c

	log.Printf("user interrupt\n")
	signal.Stop(c)
	ctrl.DownToolsAndWait()
	os.Exit(1)
}

/* wrapAcceptTCP - wrap a TCP listener for use in a select */
func wrapAcceptTCP(listener *net.TCPListener) chan *net.TCPConn {

	ch := make(chan *net.TCPConn, 1)

	go func() {

		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Printf("error on acceptTCP - %v", err)
			ch <- nil
			return
		}
		ch <- conn
	}()

	return ch
}

type Wrapper struct {
	data []byte
	eof  bool
}

/* wrapConn - wrap a TCP connection for use in a select */
func wrapConn(conn *net.TCPConn) chan Wrapper {

	ch := make(chan Wrapper, 1)

	go func() {

		b := make([]byte, 1024)
		n, err := conn.Read(b)
		if err != nil {
			if err == io.EOF {
				ch <- Wrapper{nil, true}
				return
			}
			ch <- Wrapper{nil, false}
			return
		}

		var w Wrapper
		w.data = make([]byte, n)
		copy(w.data, b[:n])
		w.eof = false
		ch <- w
	}()

	return ch
}

func filter(key string) (bool,string) {

	ok := strings.Contains(key,"\n")
	str := key
	if ok {
		str = strings.Trim(key,"\n")
	}
	return ok,str
}

/* handle client connection */
func handleClient(conn *net.TCPConn, wg *WorkGroup, storech chan *Store, infoch chan *Information) {

	log.Printf("New Client Connection\n")

	quit := wg.Start()
	defer wg.Finish()

	defer conn.Close()

	addConnection(infoch)
	defer removeConnection(infoch)

	for {
		select {
		case <-quit:
			conn.Write([]byte("CLOSING\n"))
			return

		case w := <-wrapConn(conn):
			if w.eof {
				log.Printf("End of File")
				return
			}
			if w.data == nil {
				log.Printf("no data")
				return
			}

			/* filter for \n at the end */
			filtered,str := filter(string(w.data))

			parts := strings.Split(str, " ")

			if len(parts) < 2 {
				log.Printf("Invalid data")
				conn.Write(protocol.MakeRInvalid(filtered))
				continue
			}

			switch strings.ToLower(parts[0]) {
			case "check":
				s := <-storech
				if s.Check(parts[1]) {

					conn.Write(protocol.MakeRCheck(true,filtered))
				} else {
					conn.Write(protocol.MakeRCheck(false,filtered))
				}
				storech <- s
				break
			default:
				conn.Write(protocol.MakeRUnknown(filtered))
				break
			}
		}
	}
}

/* handle service connection */
func handle(conn *net.TCPConn, wg *WorkGroup, storech chan *Store, infoch chan *Information) {

	log.Printf("New Control Connection\n")

	quit := wg.Start()
	defer wg.Finish()

	defer conn.Close()

	addConnection(infoch)
	defer removeConnection(infoch)

	for {
		select {
		case <-quit:
			conn.Write([]byte("CLOSING\n"))
			return

		case w := <-wrapConn(conn):

			if w.eof {
				log.Printf("End of File")
				return
			}

			if w.data == nil {
				log.Printf("no data")
				return
			}

			filtered,str := filter(string(w.data))

			parts := strings.Split(str, " ")

			if len(parts) < 1 {
				log.Printf("Invalid data")
				conn.Write(protocol.MakeRInvalid(filtered))
				continue
			}

			switch strings.ToLower(parts[0]) {
			case "set": /* set operation */

				if len(parts) < 2 {
					conn.Write(protocol.MakeRInvalid(filtered))
					break
				}

				s := <-storech
				if s.Set(parts[1]) {

					conn.Write(protocol.MakeRSet(true,filtered))
				} else {
					conn.Write(protocol.MakeRSet(false,filtered))
				}
				storech <- s
				break
			case "remove": /* remove operation */

				if len(parts) < 2 {
					conn.Write(protocol.MakeRInvalid(filtered))
					break
				}

				s := <-storech
				if s.Remove(parts[1]) {

					conn.Write(protocol.MakeRRemove(true,filtered))
				} else {
					conn.Write(protocol.MakeRRemove(false,filtered))
				}
				storech <- s
				break
			case "check": /* check operation */

				if len(parts) < 2 {
					conn.Write(protocol.MakeRInvalid(filtered))
					break
				}

				s := <-storech
				if s.Check(parts[1]) {

					conn.Write(protocol.MakeRCheck(true,filtered))
				} else {
					conn.Write(protocol.MakeRCheck(false,filtered))
				}
				storech <- s
				break
			case "purge":

				<-storech /* original store should get garbage collected */

				ns := NewStore() /* create new store and replace */

				storech <- ns

				conn.Write(protocol.MakeRPurge(true,filtered))
				break

			default: /* default unknown operation */
				log.Printf("UNKNOWN")

				conn.Write(protocol.MakeRUnknown(filtered))
				break
			}

		}
	}
}

func parseAddress(a string) (string, error) {

	parts := strings.Split(a, ":")
	if len(parts) != 2 {
		return "", errors.New("address is incorrect format")
	}

	addr := a
	port := parts[1]

	if ip := net.ParseIP(parts[0]); ip == nil {

		revl, err := net.ResolveIPAddr("ip4", parts[0])
		if err != nil {
			return "", err
		}
		addr = revl.String() + ":" + port
	}

	return addr, nil
}

type WorkGroup struct {
	wg   sync.WaitGroup
	quit chan bool
}

func (wg *WorkGroup) Start() chan bool {

	wg.wg.Add(1)
	return wg.quit
}

func (wg *WorkGroup) Finish() {

	wg.wg.Done()
}

func (wg *WorkGroup) DownToolsAndWait() {

	close(wg.quit)
	wg.wg.Wait()
}

func NewWorkGroup() *WorkGroup {

	wg := new(WorkGroup)
	wg.quit = make(chan bool, 1)
	return wg
}

type Store struct {
	Entries map[string]time.Time
}

func (s *Store) Check(key string) bool {

	/* TODO: add a filter to remove /n and other end-of-lines */

	if len(key) >= 8 {
		log.Printf("check %s..\n",key[:8])
	} else {
		log.Printf("check %s\n",key)
	}

	if _, exists := s.Entries[key]; exists {
		return true
	}

	return false
}

func (s *Store) Set(key string) bool {

	/* TODO: add a filter */

	if _, exists := s.Entries[key]; exists {
		return false
	}

	if len(key) >= 8 {
		log.Printf("key %s.. set\n",key[:8])
	} else {
		log.Printf("key %s set\n",key)
	}

	s.Entries[key] = time.Now()
	return true
}

func (s *Store) Remove(key string) bool {

	/* TODO: add a filter */

	if _, exists := s.Entries[key]; exists {

		delete(s.Entries, key)
		return true
	}

	return false
}

func NewStore() *Store {

	s := new(Store)
	s.Entries = make(map[string]time.Time, 0)
	return s
}

type Information struct {
	Connections int
}

func NewInformation() chan *Information {

	info := new(Information)
	info.Connections = 0

	ch := make(chan *Information, 1)
	ch <- info
	return ch
}

func addConnection(infoch chan *Information) {

	info := <-infoch
	info.Connections++
	infoch <- info
}

func removeConnection(infoch chan *Information) {

	info := <-infoch
	info.Connections--
	infoch <- info
}
