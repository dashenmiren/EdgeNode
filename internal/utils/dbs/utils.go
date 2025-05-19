package dbs

func IsClosedErr(err error) bool {
	return err == errDBIsClosed
}
