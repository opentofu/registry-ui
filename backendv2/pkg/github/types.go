package github

type RepositoryMetadata struct {
	Owner       string
	Name        string
	Stars       int64
	Forks       int64
	Description string
	IsFork      bool
	ParentOwner string
	ParentName  string
}

type RepoIdentifier struct {
	Owner string
	Name  string
}