package scp

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

type Config struct {
	ScpPath string
	Session *ssh.Session
}

func Put(session *ssh.Session, localFile string, remoteDir string) error {

	baseDir := "./"
	switch {
	case remoteDir[len(remoteDir)-1] == '/':
		baseDir = remoteDir
	default:
		baseDir = fmt.Sprintf("%s/", remoteDir)
	}

	finfo, err := os.Stat(localFile)
	if err != nil {
		return fmt.Errorf("local stat %s", err)
	}
	if finfo.IsDir() {
		return fmt.Errorf("local stat is dir %s", localFile)
	}

	src, err := os.Open(localFile)
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

func Get(session *ssh.Session, remoteFile string, localDir string) error {

	//TODO
	//var storeFile string

	//_, rfile := path.Split(remoteFile)
	//if len(rfile) > 0 {
	//	storeFile = fmt.Sprintf("%s/%s", localDir, rfile)
	//	fmt.Println(storeFile)
	//}

	//stat, err := os.Stat(localDir)
	//if err != nil {
	//	storeFile = localDir
	//} else {
	//	if stat.IsDir() {
	//		storeFile = localDir
	//	} else {
	//		storeFile = localDir
	//	}
	//}

	defer session.Close()
	go func() {
		w, _ := session.StdinPipe()
		defer w.Close()
		r, _ := session.StdoutPipe()
		rdr := bufio.NewReader(r)

		fmt.Fprint(w, "\x00")

		var mtime time.Time
		var atime time.Time
		for {
			b, _ := rdr.Peek(1)
			//fmt.Println(string(b))
			switch {
			case string(b) == "T": //T<mtime> 0 <atime> 0
				//TODO
				buf, _ := rdr.ReadBytes('\n')
				fields := strings.Fields(string(buf))
				mtime = unixStringToTime(fields[0][1:])
				atime = unixStringToTime(fields[2])
				fmt.Fprint(w, "\x00")
			case string(b) == "C": //Cmmmm <length> <filename>
				buf, _ := rdr.ReadBytes('\n')
				line := strings.TrimRight(string(buf), "\n")
				fields := strings.Fields(line)
				fMode := fields[0][1:]
				fSize := fields[1]
				fName := fields[2]
				fmt.Fprint(w, "\x00")

				rsize, _ := strconv.Atoi(fSize)
				err := dataRecv(r, rsize)
				fmt.Printf("TODO: save file[%s] mode[%s]\n", fName, fMode)
				fmt.Printf("TODO: set mtime[%s],atime[%s]\n", mtime, atime)
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

	cmd := fmt.Sprintf("scp -fp %s", remoteFile)
	if err := session.Run(cmd); err != nil {
		return fmt.Errorf("remote scp error: %s", err)
	}
	return nil
}

func dataRecv(r io.Reader, dataSize int) error {

	readSize := 0
	remainSize := dataSize
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
		if remainSize <= 0 {
			fmt.Println("recv complate")
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
