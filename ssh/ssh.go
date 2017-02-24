//
// Copyright (c) 2014 The heketi Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package ssh

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"time"

	"github.com/tmc/scp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

type SshExec struct {
	clientConfig *ssh.ClientConfig
}

func getKeyFile(file string) (key ssh.Signer, err error) {
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}
	key, err = ssh.ParsePrivateKey(buf)
	if err != nil {
		fmt.Print(err)
		return
	}
	return
}

func NewSshExecWithAuth(user string) (SshExecutor, error) {

	sshexec := &SshExec{}

	authSocket := os.Getenv("SSH_AUTH_SOCK")
	if authSocket == "" {
		return nil, fmt.Errorf("SSH_AUTH_SOCK not set")
	}

	agentUnixSock, err := net.Dial("unix", authSocket)
	if err != nil {
		return nil, fmt.Errorf("Cannot connect to SSH_AUTH_SOCK")
	}

	agent := agent.NewClient(agentUnixSock)
	signers, err := agent.Signers()
	if err != nil {
		return nil, fmt.Errorf("Could not get key signatures: %v", err)
	}

	sshexec.clientConfig = &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{ssh.PublicKeys(signers...)},
	}

	return sshexec, nil
}

func NewSshExecWithKeyFile(user string, file string) (SshExecutor, error) {

	var key ssh.Signer
	var err error

	sshexec := &SshExec{}

	// Now in the main function DO:
	if key, err = getKeyFile(file); err != nil {
		return nil, fmt.Errorf("Unable to get keyfile")
	}
	// Define the Client Config as :
	sshexec.clientConfig = &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
	}

	return sshexec, nil
}

func (s *SshExec) Copy(size int64,
	mode os.FileMode,
	fileName string,
	contents io.Reader,
	host, destinationPath string) error {

	// Create a connection to the server
	client, err := ssh.Dial("tcp", host, s.clientConfig)
	if err != nil {
		return err
	}
	defer client.Close()

	// Create a session
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	// Copy Data
	err = scp.Copy(size, mode, fileName, contents, destinationPath, session)
	if err != nil {
		return err
	}

	return nil
}

func (s *SshExec) CopyPath(sourcePath, host, destinationPath string) error {

	// Create a connection to the server
	client, err := ssh.Dial("tcp", host, s.clientConfig)
	if err != nil {
		return err
	}
	defer client.Close()

	// Create a session
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	// Copy Data
	err = scp.CopyPath(sourcePath, destinationPath, session)
	if err != nil {
		return err
	}

	return nil
}

// This function was based from https://github.com/coreos/etcd-manager/blob/master/main.go
func (s *SshExec) Exec(host string, commands []string, timeoutMinutes int, useSudo bool) ([]string, error) {

	buffers := make([]string, len(commands))

	// :TODO: Will need a timeout here in case the server does not respond
	client, err := ssh.Dial("tcp", host, s.clientConfig)
	if err != nil {
		return nil, fmt.Errorf("Failed to create SSH connection to %v: %v", host, err)
	}
	defer client.Close()

	// Execute each command
	for index, command := range commands {

		session, err := client.NewSession()
		if err != nil {
			return nil, fmt.Errorf("Unable to create SSH session: %v", err)
		}
		defer session.Close()

		// Create a buffer to trap session output
		var b bytes.Buffer
		var berr bytes.Buffer
		session.Stdout = &b
		session.Stderr = &berr

		if useSudo {
			command = "sudo " + command
		}

		err = session.Start(command)
		if err != nil {
			return nil, err
		}

		// Spawn function to wait for results
		errch := make(chan error)
		go func() {
			errch <- session.Wait()
		}()

		// Set the timeout
		timeout := time.After(time.Minute * time.Duration(timeoutMinutes))

		// Wait for either the command completion or timeout
		select {
		case err := <-errch:
			buffers[index] = b.String()
			if err != nil {
				return buffers, fmt.Errorf("Failed to run command [%v] on %v: Err[%v]: Stdout [%v]: Stderr [%v]",
					command, host, err, b.String(), berr.String())
			}
			//LOG("Host: %v Command: %v\nResult: %v", host, command, b.String())

		case <-timeout:
			err := session.Signal(ssh.SIGKILL)
			if err != nil {
				return nil, fmt.Errorf("Command timed out and unable to send kill signal to command [%v] on host [%v]: %v",
					command, host, err)
			}
			return nil, fmt.Errorf("Timeout on command [%v] on %v: Err[%v]: Stdout [%v]: Stderr [%v]",
				command, host, err, b.String(), berr.String())
		}
	}

	return buffers, nil
}
