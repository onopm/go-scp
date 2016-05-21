package scp

import (
	"golang.org/x/crypto/ssh"
)

type FilePath struct {
	Host   string
	User   string
	Path   string
	Remote bool
}

type Config struct {
	ScpPath string
	Session *ssh.Session
}
