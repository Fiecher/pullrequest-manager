package api

import (
	"pullrequest-inator/internal/api/dtos"
)

func ToAPITeam(d dtos.Team) Team {
	members := make([]TeamMember, len(d.Members))
	for i, m := range d.Members {
		member := ToAPITeamMember(m)
		members[i] = member
	}

	return Team{
		TeamName: d.TeamName,
		Members:  members,
	}
}

func ToAPITeamMember(m dtos.TeamMember) TeamMember {
	return TeamMember{
		UserId:   m.UserId,
		Username: m.Username,
		IsActive: m.IsActive,
	}
}

func FromAPITeam(t Team) dtos.Team {
	members := make([]dtos.TeamMember, len(t.Members))
	for i, m := range t.Members {
		member := FromAPITeamMember(m)
		members[i] = member
	}

	return dtos.Team{
		TeamName: t.TeamName,
		Members:  members,
	}
}

func FromAPITeamMember(m TeamMember) dtos.TeamMember {
	return dtos.TeamMember{
		UserId:   m.UserId,
		Username: m.Username,
		IsActive: m.IsActive,
	}
}

func ToAPIUser(u dtos.User) User {
	return User{
		UserId:   u.UserId,
		Username: u.Username,
		TeamName: u.TeamName,
		IsActive: u.IsActive,
	}
}

func ToAPIPullRequest(d dtos.PullRequest) PullRequest {
	reviewers := make([]string, len(d.AssignedReviewers))

	return PullRequest{
		PullRequestId:     d.PullRequestId,
		PullRequestName:   d.PullRequestName,
		AuthorId:          d.AuthorId,
		Status:            PullRequestStatus(d.Status),
		AssignedReviewers: reviewers,
		CreatedAt:         d.CreatedAt,
		MergedAt:          d.MergedAt,
	}
}

func ToAPIPullRequestShort(d dtos.PullRequest) PullRequestShort {
	return PullRequestShort{
		PullRequestId:   d.PullRequestId,
		PullRequestName: d.PullRequestName,
		AuthorId:        d.AuthorId,
		Status:          PullRequestShortStatus(d.Status),
	}
}

func ToAPIPullRequestShortList(list []*dtos.PullRequest) []PullRequestShort {
	out := make([]PullRequestShort, len(list))
	for i, pr := range list {
		out[i] = ToAPIPullRequestShort(*pr)
	}
	return out
}
