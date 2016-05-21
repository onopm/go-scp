package scp

import ()

func ArgParse(arg string) (*FilePath, error) {

	path := &FilePath{Remote: false}

	for i := 0; i < len(arg); i++ {
		switch arg[i] {
		case '@':
			if len(path.User) < 1 {
				path.User = arg[0:i]
			}
		case ':':
			if len(path.User) > 0 {
				path.Host = arg[len(path.User)+1 : i]
				path.Remote = true
			} else if len(path.Host) < 1 {
				path.Host = arg[0:i]
				path.Remote = true
			}
		}

	}
	if len(path.User) < 1 && len(path.Host) < 1 {
		path.Path = arg
	} else if len(path.User) < 1 {
		path.Path = arg[len(path.Host)+1:]
	} else {
		path.Path = arg[len(path.User)+len(path.Host)+2:]
	}

	return path, nil
}
