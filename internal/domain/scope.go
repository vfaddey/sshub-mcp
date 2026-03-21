package domain

type AccessScope struct {
	TokenID    int64
	ProjectIDs []int64
}

func (s AccessScope) MayAccessProject(projectID int64) bool {
	for _, id := range s.ProjectIDs {
		if id == projectID {
			return true
		}
	}
	return false
}
