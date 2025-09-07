package common

func ParseURLPath(p string) (up URLPath, err error) {
	// TODO Add some validation here
	up = URLPath(p)
	return up, err
}

func ParseFilepath(f string) (fp Filepath, err error) {
	// TODO Add some validation here
	fp = Filepath(f)
	return fp, err
}

func ParseDirPath(f string) (dp DirPath, err error) {
	// TODO Add some validation here
	dp = DirPath(f)
	return dp, err
}

func ParseIdentifier(s string) (id Identifier, err error) {
	// TODO Add some validation here
	id = Identifier(s)
	return id, err
}
