package scp

import (
	"testing"
)

func TestArg(t *testing.T) {

	path, err := argParse("/var/log/maillog")
	if err != nil {
		t.Error("%s", err)
	} else if len(path.User) < 0 || len(path.Host) > 0 || path.Path != "/var/log/maillog" || path.Remote == true {
		t.Errorf("arg[%s],user[%s],host[%s],path[%s]", "/var/log/maillog", path.User, path.Host, path.Path)
	}

	path, err = argParse("remote:/var/log/maillog")
	if err != nil {
		t.Error("%s", err)
	} else if len(path.User) < 0 || path.Host != "remote" || path.Path != "/var/log/maillog" || path.Remote == false {
		t.Errorf("arg[%s],user[%s],host[%s],path[%s]", "/var/log/maillog", path.User, path.Host, path.Path)
	}

	path, err = argParse("username@remote:/var/log/maillog")
	if err != nil {
		t.Error("%s", err)
	} else if path.User != "username" || path.Host != "remote" || path.Path != "/var/log/maillog" || path.Remote == false {
		t.Errorf("arg[%s],user[%s],host[%s],path[%s]", "/var/log/maillog", path.User, path.Host, path.Path)
	}
}
