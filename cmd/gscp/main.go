package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	scp "github.com/onopm/go-scp"
	"golang.org/x/crypto/ssh"
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: gscp ...\n")
		flag.PrintDefaults()
	}
}

func main() {
	var (
		host      string
		port      int
		username  string
		password  string
		preserves bool
	)
	flag.IntVar(&port, "P", 22, "port")
	flag.StringVar(&password, "password", "", "password")
	flag.BoolVar(&preserves, "p", false, "Preserves ...") //TODO
	flag.Parse()

	if len(flag.Args()) < 2 {
		flag.Usage()
		os.Exit(1)
	}
	src, err := scp.ArgParse(flag.Args()[0])
	if err != nil {
		flag.Usage()
		os.Exit(1)
	}
	dst, err := scp.ArgParse(flag.Args()[1])
	if err != nil {
		flag.Usage()
		os.Exit(1)
	}
	//fmt.Printf("%s %s\n", flag.Args()[0], flag.Args()[1])

	username = os.Getenv("USER")
	if src.Remote == true && dst.Remote == false {
		host = src.Host
		if len(src.User) > 0 {
			username = src.User
		}
	} else if src.Remote == false && dst.Remote == true {
		host = dst.Host
		if len(dst.User) > 0 {
			username = dst.User
		}
	} else {
		fmt.Fprintf(os.Stderr, "not support ...\n")
		os.Exit(1)
	}

	var clientConfig *ssh.ClientConfig
	toServer := fmt.Sprintf("%s:%d", host, port)
	if len(password) < 1 {
		signer, _ := loadKey()
		clientConfig = &ssh.ClientConfig{
			User: username,
			Auth: []ssh.AuthMethod{
				ssh.PublicKeys(signer),
			},
		}
	} else {
		clientConfig = &ssh.ClientConfig{
			User: username,
			Auth: []ssh.AuthMethod{
				ssh.Password(password),
			},
		}
	}

	client, err := ssh.Dial("tcp", toServer, clientConfig)
	if err != nil {
		fmt.Println("Failed to dial: " + err.Error())
		os.Exit(1)
	}
	session, err := client.NewSession()
	if err != nil {
		fmt.Println("Failed to create session: " + err.Error())
		os.Exit(1)
	}

	if src.Remote == true && dst.Remote == false {
		err := scp.Get(session, src, dst)
		if err != nil {
			fmt.Println(err)
		}
	} else if src.Remote == false && dst.Remote == true {
		err := scp.Put(session, src, dst)
		if err != nil {
			fmt.Println(err)
		}
	} else {
		fmt.Fprintf(os.Stderr, "not support ...\n")
		os.Exit(1)
	}

}

func loadKey() (ssh.Signer, error) {
	keyfile := fmt.Sprintf("%s/.ssh/id_rsa", os.Getenv("HOME"))
	key, err := ioutil.ReadFile(keyfile)
	if err != nil {
		return nil, fmt.Errorf("unable to read private key: %v\n", err)
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("unable to parse private key: %v\n", err)

	}
	return signer, nil
}
