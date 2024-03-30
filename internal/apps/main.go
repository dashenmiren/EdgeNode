package apps

import teaconst "github.com/dashenmiren/EdgeNode/internal/const"

func RunMain(f func()) {
	if teaconst.IsMain {
		f()
	}
}
