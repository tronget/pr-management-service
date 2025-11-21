package team

type CreateTeamRequest struct {
	TeamName string   `json:"team_name"`
	Members  []Member `json:"members"`
}

type CreateTeamResponse struct {
	Team Team `json:"team"`
}

type GetTeamResponse struct {
	TeamName string   `json:"team_name"`
	Members  []Member `json:"members"`
}
