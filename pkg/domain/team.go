package domain

import "git.ronaksoft.com/river/msg/msg"

/**
 * @created 01/09/2020 - 16:13
 * @project riversdk
 * @author reza
 */

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
