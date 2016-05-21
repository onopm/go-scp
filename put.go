package scp

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/ssh"
)

func Put(session *ssh.Session, localFile *FilePath, remoteDir *FilePath) error {

	baseDir := "./"
	switch {
	case remoteDir.Path[len(remoteDir.Path)-1] == '/':
		baseDir = remoteDir.Path
	default:
		baseDir = fmt.Sprintf("%s/", remoteDir.Path)
	}

	finfo, err := os.Stat(localFile.Path)
	if err != nil {
		return fmt.Errorf("local stat %s", err)
	}
	if finfo.IsDir() {
		return fmt.Errorf("local stat is dir %s", localFile)
	}

	src, err := os.Open(localFile.Path)
	if err != nil {
		return fmt.Errorf("local stat %s", err)
	}

	defer session.Close()
	go func() {
		w, _ := session.StdinPipe()
		defer w.Close()

		fmt.Fprintln(w, "C0644", finfo.Size(), finfo.Name())
		io.Copy(w, src)
		fmt.Fprint(w, "\x00") // transfer end with \x00
	}()

	//cmd := fmt.Sprintf("/usr/bin/scp -tr %s", baseDir)
	cmd := fmt.Sprintf("scp -t %s", baseDir)
	if err := session.Run(cmd); err != nil {
		return fmt.Errorf("remote scp error: %s", err)
	}
	return nil
}
