package main

import (
	"encoding/binary"
	"github.com/docker/docker/pkg/term"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"

	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func startSSH() error {
	width := 80
	height := 24
	sshConfig := &ssh.ClientConfig{
		User: "ec2-user",
		Auth: []ssh.AuthMethod{
			SSHAgent(),
			PublicKeyFile("/keys/" + *e.keyName),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}
	fmt.Printf("Opening connection to %v:22 with key %v", *e.ipAddr, "/keys/"+*e.keyName)
	connection, err := ssh.Dial("tcp", *e.ipAddr+":22", sshConfig)
	if err != nil {
		return fmt.Errorf("Failed to dial: %s", err)
	}
	session, err := connection.NewSession()

	if err != nil {
		return fmt.Errorf("Failed to create session: %s", err)
	}

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	modes := ssh.TerminalModes{
		ssh.ECHO: 1,
	}

	fd := os.Stdin.Fd()

	if term.IsTerminal(fd) {
		oldState, err := term.MakeRaw(fd)
		if err != nil {
			return err
		}

		defer term.RestoreTerminal(fd, oldState)

		winsize, err := term.GetWinsize(fd)
		if err == nil {
			width = int(winsize.Width)
			height = int(winsize.Height)
		}
	}

	if err := session.RequestPty("xterm", width, height, modes); err != nil {
		session.Close()
		return fmt.Errorf("request for pseudo terminal failed: %s", err)
	}

	// start shell
	if err := session.Shell(); err != nil {
		return fmt.Errorf("Couldn't start shell: %v", err)
	}
	go monitorChanges(session, os.Stdout.Fd())

	session.Wait()

	return nil
}

func SSHAgent() ssh.AuthMethod {
	if sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		return ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers)
	}
	return nil
}
func PublicKeyFile(file string) ssh.AuthMethod {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		return nil
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil
	}
	return ssh.PublicKeys(key)
}

// Function from: https://github.com/nanobox-io/golang-ssh (Apache 2.0 licensed)
func monitorChanges(session *ssh.Session, fd uintptr) {
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGWINCH)
	defer signal.Stop(sigs)

	for range sigs {
		session.SendRequest("window-change", false, termSize(fd))
	}
}

// Function from: https://github.com/nanobox-io/golang-ssh (Apache 2.0 licensed)
func termSize(fd uintptr) []byte {
	size := make([]byte, 16)

	winsize, err := term.GetWinsize(fd)
	if err != nil {
		binary.BigEndian.PutUint32(size, uint32(80))
		binary.BigEndian.PutUint32(size[4:], uint32(24))
		return size
	}

	binary.BigEndian.PutUint32(size, uint32(winsize.Width))
	binary.BigEndian.PutUint32(size[4:], uint32(winsize.Height))

	return size
}
