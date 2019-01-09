package utils

import (
	"bytes"
	"os/exec"
	"os/user"
	"strconv"
	"syscall"
)

func RunCMD(command string, user *string) ([]byte, error) {
	var stdout bytes.Buffer

	cmd := exec.Command("/bin/sh", "-c", command)
	if user != nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
		uid, gid, err := findUserIDs(*user)
		if err != nil {
			return stdout.Bytes(), err
		}
		cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uid, Gid: gid}
	}
	return cmd.Output()
}

func findUserIDs(username string) (uint32, uint32, error) {
	u, err := user.Lookup(username)
	if err != nil {
		return 0, 0, err
	}

	uid, err := strconv.ParseInt(u.Uid, 10, 32)
	gid, err := strconv.ParseInt(u.Gid, 10, 32)
	return uint32(uid), uint32(gid), err
}
