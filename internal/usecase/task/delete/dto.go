package delete

type Command struct {
	ID      string
	Version uint64
}

type Result struct {
	ID string
} //TODO:versionも返すべき
