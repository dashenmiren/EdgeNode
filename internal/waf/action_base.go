package waf

type BaseAction struct {
	currentActionId int64
}

// ActionId 读取ActionId
func (this *BaseAction) ActionId() int64 {
	return this.currentActionId
}

// SetActionId 设置Id
func (this *BaseAction) SetActionId(actionId int64) {
	this.currentActionId = actionId
}
