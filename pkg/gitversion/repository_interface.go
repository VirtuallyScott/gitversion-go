package gitversion

type Repository interface {
	GetSHA() (string, error)
	GetShortSHA() (string, error)
	GetCommitDate() (string, error)
	GetLatestTag() (string, error)
	GetCommitCountSinceTag(tag string) (int, error)
}