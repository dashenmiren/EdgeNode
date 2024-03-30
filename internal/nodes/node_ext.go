//go:build !script
// +build !script

package nodes

func (this *Node) reloadCommonScripts() error {
	return nil
}

func (this *Node) reloadIPLibrary() {

}

func (this *Node) notifyPlusChange() error {
	return nil
}

func (this *Node) execTOAChangedTask() error {
	return nil
}
