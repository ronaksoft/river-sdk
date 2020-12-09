package domain

import (
	"git.ronaksoft.com/river/msg/go/msg"
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

func GetTeam(teamID int64, teamAccessHash uint64) *msg.InputTeam {
	return &msg.InputTeam{
		ID:         teamID,
		AccessHash: teamAccessHash,
	}
}

func GetTeamID(team *msg.InputTeam) int64 {
	if team == nil {
		return 0
	}
	return team.ID
}

func SetCurrentTeam(t *msg.InputTeam) {
	if t == nil {
		_CurrTeamID = 0
		_CurrTeamAccessHash = 0
	} else {
		_CurrTeamID = t.ID
		_CurrTeamAccessHash = t.AccessHash
	}
}

func GetCurrTeamID() int64 {
	return _CurrTeamID
}

func GetCurrTeam() *msg.InputTeam {
	if _CurrTeamID == 0 {
		return &msg.InputTeam{
			ID:         0,
			AccessHash: 0,
		}
	}
	return &msg.InputTeam{
		ID:         _CurrTeamID,
		AccessHash: _CurrTeamAccessHash,
	}
}
