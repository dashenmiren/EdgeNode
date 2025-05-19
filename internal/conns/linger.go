package conns

type LingerConn interface {
	SetLinger(sec int) error
}
