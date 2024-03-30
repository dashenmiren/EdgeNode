package nodes

type LingerConn interface {
	SetLinger(sec int) error
}
