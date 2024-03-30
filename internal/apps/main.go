package apps

import teaconst "github.com/TeaOSLab/EdgeNode/internal/const"

func RunMain(f func()) {
	if teaconst.IsMain {
		f()
	}
}
