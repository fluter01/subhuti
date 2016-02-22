package bot

type Command interface {
	Run(string) (string, error)
}

type VersionCommand struct {
}

func (cmd *VersionCommand) Run(string) (string, error) {
	return Version(), nil
}

type SourceCommand struct {
}

func (cmd *SourceCommand) Run(string) (string, error) {
	return Source(), nil
}
