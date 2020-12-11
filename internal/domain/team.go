package domain

import (
	"git.ronaksoft.com/river/sdk/internal/tools"
	"github.com/ronaksoft/rony"
)

/**
 * @created 01/09/2020 - 16:13
 * @project riversdk
 * @author reza
 */

var (
	_CurrTeamID         int64
	_CurrTeamAccessHash uint64
)

func TeamHeader(teamID int64, teamAccess uint64) []*rony.KeyValue {
	if teamID == 0 {
		return nil
	}
	kv := make([]*rony.KeyValue, 0, 2)
	kv = append(kv,
		&rony.KeyValue{
			Key:   "TeamID",
			Value: tools.Int64ToStr(teamID),
		},
		&rony.KeyValue{
			Key:   "TeamAccess",
			Value: tools.UInt64ToStr(teamAccess),
		},
	)
	return kv
}

func GetTeamID(e *rony.MessageEnvelope) int64 {
	return tools.StrToInt64(e.Get("TeamID", "0"))
}

func GetTeamAccess(e *rony.MessageEnvelope) uint64 {
	return tools.StrToUInt64(e.Get("TeamAccess", "0"))
}

func SetCurrentTeam(teamID int64, teamAccess uint64) {
	_CurrTeamID = teamID
	_CurrTeamAccessHash = teamAccess
}

func GetCurrTeamID() int64 {
	return _CurrTeamID
}

func GetCurrTeamAccess() uint64 {
	return _CurrTeamAccessHash
}
