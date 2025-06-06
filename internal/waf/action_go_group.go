package waf

import (
	"github.com/dashenmiren/EdgeNode/internal/remotelogs"
	"github.com/dashenmiren/EdgeNode/internal/waf/requests"
	"github.com/iwind/TeaGo/types"
	"net/http"
)

type GoGroupAction struct {
	BaseAction

	GroupId string `yaml:"groupId" json:"groupId"`
}

func (this *GoGroupAction) Init(waf *WAF) error {
	return nil
}

func (this *GoGroupAction) Code() string {
	return ActionGoGroup
}

func (this *GoGroupAction) IsAttack() bool {
	return false
}

func (this *GoGroupAction) WillChange() bool {
	return true
}

func (this *GoGroupAction) Perform(waf *WAF, group *RuleGroup, set *RuleSet, request requests.Request, writer http.ResponseWriter) PerformResult {
	var nextGroup = waf.FindRuleGroup(types.Int64(this.GroupId))
	if nextGroup == nil || !nextGroup.IsOn {
		return PerformResult{
			ContinueRequest: true,
			GoNextSet:       true,
		}
	}

	b, _, nextSet, err := nextGroup.MatchRequest(request)
	if err != nil {
		remotelogs.Error("WAF", "GO_GROUP_ACTION: "+err.Error())
		return PerformResult{
			ContinueRequest: true,
			GoNextSet:       true,
		}
	}

	if !b {
		return PerformResult{
			ContinueRequest: true,
			GoNextSet:       true,
		}
	}

	return nextSet.PerformActions(waf, nextGroup, request, writer)
}
