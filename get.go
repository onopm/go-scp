package scp

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

func Get(session *ssh.Session, remoteFile *FilePath, localDir *FilePath) error {

	store := &storeConfig{}

	ldir, lfile := path.Split(localDir.Path)
	_, err := os.Stat(ldir)
	if err != nil {
		return fmt.Errorf("local %s %s", ldir, err)
	}
	if len(lfile) > 0 {
		store.Path = localDir.Path
	} else {
		_, rfile := path.Split(remoteFile.Path)
		store.Path = fmt.Sprintf("%s/%s", ldir, rfile)
	}

	go func() {
		w, _ := session.StdinPipe()
		defer w.Close()
		r, _ := session.StdoutPipe()
		rdr := bufio.NewReader(r)

		fmt.Fprint(w, "\x00")

		for {
			b, _ := rdr.Peek(1)
			//fmt.Println(string(b))
			switch {
			case string(b) == "T": //T<mtime> 0 <atime> 0
				//TODO
				buf, _ := rdr.ReadBytes('\n')
				fields := strings.Fields(string(buf))
				store.mtime = unixStringToTime(fields[0][1:])
				store.atime = unixStringToTime(fields[2])
				fmt.Fprint(w, "\x00")
			case string(b) == "C": //Cmmmm <length> <filename>
				buf, _ := rdr.ReadBytes('\n')
				line := strings.TrimRight(string(buf), "\n")
				fields := strings.Fields(line)
				//fMode := fields[0][1:]
				//fSize := fields[1]
				//fName := fields[2]
				fmode, _ := strconv.ParseUint(fields[0][1:], 10, 32)
				store.Perm = os.FileMode(fmode)
				store.size, _ = strconv.Atoi(fields[1])
				fmt.Fprint(w, "\x00")

				err := store.dataRecv(r)
				if err != nil {
					fmt.Println(err)
				}
				fmt.Fprint(w, "\x00")
				return
			case string(b) == "E":
				fmt.Printf("not support 'E'\n")
				fmt.Fprint(w, "\x00")
				return
			default:
				buf, _ := rdr.ReadBytes('\n')
				fmt.Printf("unknown: %s\n", string(buf))
				fmt.Fprint(w, "\x00")
				return
			}
		}
		// not use.
		fmt.Fprint(w, "\x00")
	}()

	cmd := fmt.Sprintf("scp -fp %s", remoteFile.Path)
	//fmt.Printf("%s\n", cmd)
	if err := session.Run(cmd); err != nil {
		return fmt.Errorf("remote scp error: %s", err)
	}
	return nil
}

type storeConfig struct {
	Path  string
	Perm  os.FileMode
	size  int
	mtime time.Time
	atime time.Time
}

func (s *storeConfig) dataRecv(r io.Reader) error {

	w, err := os.Create(s.Path)
	if err != nil {
		return fmt.Errorf("%s", err)
	}

	readSize := 0
	remainSize := s.size
	bufSize := 1024 * 256
	rbuf := make([]byte, bufSize)

	for {
		n, err := r.Read(rbuf)
		if err != nil {
			return fmt.Errorf("%s", err)
		}
		readSize += n
		remainSize -= n
		//fmt.Printf("%v/%v %v\n", readSize, dataSize, remainSize)
		w.Write(rbuf)
		if remainSize <= 0 {
			fmt.Println("recv complate")
			err := w.Chmod(s.Perm)
			if err != nil {
				fmt.Println(err)
			}
			w.Close()
			os.Chtimes(s.Path, s.atime, s.mtime)
			return nil
		}
		if remainSize < bufSize {
			rbuf = make([]byte, remainSize)
		}
	}
}

func unixStringToTime(s string) time.Time {
	i, _ := strconv.ParseInt(s, 10, 64)
	return time.Unix(i, 0)
}
