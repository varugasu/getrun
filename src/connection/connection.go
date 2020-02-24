package connection

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// Host groups the main attributes
type Host struct {
	prompt  string
	Client  *ssh.Client
	Session *ssh.Session
	Writer  io.WriteCloser
	Reader  io.Reader
}

// Connect to a given Host and it credentials
func Connect(host string, port int, username string, password string) (*Host, error) {
	var conn *ssh.Client
	var err error
	addr := fmt.Sprintf("%s:%d", host, port)
	sshConfig := &ssh.ClientConfig{
		User:            username,
		Auth:            []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

DIAL:
	conn, err = ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		m := regexp.MustCompile(`server offered: \[(.*)\]`)

		errString := err.Error()
		if strings.Contains(errString, "key exchange") {
			sshConfig.Config.KeyExchanges = strings.Split(m.FindStringSubmatch(errString)[1], " ")
			goto DIAL
		}
		if strings.Contains(errString, "server cipher") {
			sshConfig.Config.Ciphers = strings.Split(m.FindStringSubmatch(errString)[1], " ")
			goto DIAL
		}
		return nil, err
	}

	session, err := conn.NewSession()
	r, _ := session.StdoutPipe()
	w, _ := session.StdinPipe()

	if err != nil {
		return nil, err
	}
	if err := session.Shell(); err != nil {
		return nil, errors.New("failed to invoke shell: " + err.Error())
	}

	h := &Host{
		Client:  conn,
		Session: session,
		Reader:  r,
		Writer:  w,
	}

	// h.SendCommand("", "R1#")
	err = h.FindPrompt()
	if err != nil {
		return nil, err
	}
	w.Write([]byte("ter len 0\n\n"))
	return h, nil

}

func (h *Host) FindPrompt() error {
	h.Writer.Write([]byte("\n\n\n\n"))
	promptChan := make(chan string)
	defer close(promptChan)
	scanner := bufio.NewScanner(h.Reader)
	regex := regexp.MustCompile(`^.*[#|>]`)
	go func(s *bufio.Scanner, c chan string) {
		for scanner.Scan() {
			if regex.Match((scanner.Bytes())) {
				c <- scanner.Text()
				return
			}
		}

	}(scanner, promptChan)
	select {
	case prompt := <-promptChan:
		h.prompt = prompt
		return nil
	case <-time.After(30 * time.Second):
		return errors.New("Timeout")
	}

}

func (h *Host) SendCommand(cmd string) (string, error) {
	h.Writer.Write([]byte(cmd + "\n\n"))
	stop := make(chan string)
	defer close(stop)
	scanner := bufio.NewScanner(h.Reader)
	go func(s *bufio.Scanner, c chan string) {
		var output string
		capture := false
		for scanner.Scan() {
			hold := scanner.Text()
			if capture {
				if strings.Contains(hold, h.prompt) {
					c <- output
					return
				}
				output += hold + "\n"
			}
			if strings.Contains(hold, cmd) {
				capture = true
				continue
			}

		}

	}(scanner, stop)
	select {
	case output := <-stop:
		return output, nil
	case <-time.After(120 * time.Second):
		return "", errors.New("Timeout")
	}
}
