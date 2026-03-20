package domain

type AccessScope struct {
	TokenID    string
	ProjectIDs []string
}

func (s AccessScope) MayAccessProject(projectID string) bool {
	for _, id := range s.ProjectIDs {
		if id == projectID {
			return true
		}
	}
	return false
}
