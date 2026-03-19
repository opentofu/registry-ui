package repository

import "time"

type RepositoryMetadata struct {
	Owner       string
	Name        string
	Stars       int64
	Forks       int64
	Description string
	IsFork      bool
	ParentOwner string
	ParentName  string
	// Redirect tracking fields
	ActualOwner string
	ActualName  string
	IsRedirect  bool
	// Additional stats fields
	OpenIssues  int64
	Subscribers int64
	Topics      []string
	// Repository metadata fields
	Homepage        string
	Language        string
	Archived        bool
	DefaultBranch   string
	CreatedAtGitHub time.Time
	PushedAt        time.Time
	UpdatedAtGitHub time.Time
}

type RepoIdentifier struct {
	Owner string
	Name  string
}
