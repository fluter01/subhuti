package bot

type Command func(string) (string, error)

func VersionCommand(string) (string, error) {
	return Version(), nil
}

func SourceCommand(string) (string, error) {
	return Source(), nil
}
