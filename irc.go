package main

import (
	"crypto/tls"
	"fmt"

	irc "go.lindenii.runxiyu.org/lindenii-irc"
)

var (
	ircSendBuffered   chan string
	ircSendDirectChan chan errorBack[string]
)

type errorBack[T any] struct {
	content   T
	errorBack chan error
}

func ircBotSession() error {
	underlyingConn, err := tls.Dial(ircNet, ircAddr, nil)
	if err != nil {
		return err
	}
	defer underlyingConn.Close()

	conn := irc.NewConn(underlyingConn)

	logAndWriteLn := func(s string) (n int, err error) {
		return conn.WriteString(s + "\r\n")
	}

	_, err = logAndWriteLn("NICK :" + ircNick)
	if err != nil {
		return err
	}
	_, err = logAndWriteLn("USER " + ircUsername + " 0 * :" + ircRealname)
	if err != nil {
		return err
	}

	readLoopError := make(chan error)
	writeLoopAbort := make(chan struct{})
	go func() {
		for {
			select {
			case <-writeLoopAbort:
				return
			default:
			}

			msg, line, err := conn.ReadMessage()
			if err != nil {
				readLoopError <- err
				return
			}

			fmt.Println(line)

			switch msg.Command {
			case "001":
				_, err = logAndWriteLn("JOIN " + ircChannel)
				if err != nil {
					readLoopError <- err
					return
				}
			case "PING":
				_, err = logAndWriteLn("PONG :" + msg.Args[0])
				if err != nil {
					readLoopError <- err
					return
				}
			case "JOIN":
				c, ok := msg.Source.(irc.Client)
				if !ok {
				}
				if c.Nick != ircNick {
					continue
				}
			default:
			}
		}
	}()

	for {
		select {
		case err = <-readLoopError:
			return err
		case line := <-ircSendBuffered:
			_, err = logAndWriteLn(line)
			if err != nil {
				select {
				case ircSendBuffered <- line:
				default:
				}
				writeLoopAbort <- struct{}{}
				return err
			}
		case lineErrorBack := <-ircSendDirectChan:
			_, err = logAndWriteLn(lineErrorBack.content)
			lineErrorBack.errorBack <- err
			if err != nil {
				writeLoopAbort <- struct{}{}
				return err
			}
		}
	}
}

func ircSendDirect(s string) error {
	ech := make(chan error, 1)

	ircSendDirectChan <- errorBack[string]{
		content:   s,
		errorBack: ech,
	}

	return <-ech
}

func ircBotLoop() {
	ircSendBuffered = make(chan string, ircBuf)
	ircSendDirectChan = make(chan errorBack[string])

	for {
		_ = ircBotSession()
	}
}

func init() {
	go ircBotLoop()
}
