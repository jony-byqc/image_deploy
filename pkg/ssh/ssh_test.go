package ssh

import (
	"bytes"
	log "github.com/sirupsen/logrus"
	"strings"
	"testing"
)

const (
	username = "wjq"
	password = "123456"
	ip       = "192.168.60.157"
	port     = 22
	cmds     = "show clock;cd /data;pwd;ls;exit"
)

func Test_RunCmds(t *testing.T) {

	cli := NewSSHClient(username, password, ip, port)

	err := cli.Connect()
	if err != nil {
		log.Error("ssh client connect failed")
		return
	}
	session, err := cli.SessionRequestPty()
	if err != nil {
		return
	}
	if session == nil {
		return
	}
	defer session.Close()

	cmdlist := strings.Split(cmds, ";")
	stdinBuf, err := session.StdinPipe()
	if err != nil {
		t.Error(err)
		return
	}

	var outbt, errbt bytes.Buffer
	session.Stdout = &outbt

	session.Stderr = &errbt
	err = session.Shell()
	if err != nil {
		t.Error(err)
		return
	}
	for _, cmd := range cmdlist {
		cmd = cmd + "\n"
		stdinBuf.Write([]byte(cmd))
	}
	session.Wait()
	t.Log(outbt.String() + errbt.String())
	return

}
