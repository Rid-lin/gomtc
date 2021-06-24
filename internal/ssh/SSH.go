package ssh

import (
	"bytes"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

func GetResponseOverSSHfMT(SSHHost, SSHPort, SSHUser, SSHPass, command string) bytes.Buffer {
	sshConfig := &ssh.ClientConfig{
		User: SSHUser,
		Auth: []ssh.AuthMethod{
			ssh.Password(SSHPass),
		},
		// Non-production only
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	// Established connection
	client, err := ssh.Dial("tcp", SSHHost+":"+SSHPort, sshConfig)
	if err != nil {
		log.Errorf("Failed to dial: %s", err)
	}
	defer client.Close()
	var b bytes.Buffer
	// crete ssh session

	session, err := client.NewSession()
	if err != nil {
		log.Errorf("Failed to ceate session: %s", err)
	}
	defer session.Close()
	// Once a Session is created, you can execute a single command on
	// the remote side using the Run method.
	session.Stdout = &b
	if err := session.Run(command); err != nil {
		log.Error("Failed to run: " + err.Error())
	}
	session.Close()
	return b
}

func GetResponseOverSSHfMTWithBuffer(SSHHost, SSHPort, SSHUser, SSHPass, command string, MaxSSHRetries, SSHRetryDelay int) (bytes.Buffer, error) {
	var b bytes.Buffer
	if MaxSSHRetries == 0 {
		return b, fmt.Errorf("The number of connection attempts was not specified")
	}
	var i int = 1

	for b.Len() == 0 {
		if (i - MaxSSHRetries) == 0 {
			return b, fmt.Errorf("Connection attempts ended")
		}
		b = GetResponseOverSSHfMT(SSHHost, SSHPort, SSHUser, SSHPass, command)
		if b.Len() == 0 {
			log.Warningf("\rThe connection attempt failed. Trying again(%d) ", i)
			dur, err := time.ParseDuration(fmt.Sprintf("%ds", SSHRetryDelay))
			if err != nil {
				dur = time.Second * 5
			}
			time.Sleep(dur)
		}
		i++
	}
	return b, nil
}
