package bot

type Command func(*IRC, string) (string, error)

func VersionCommand(*IRC, string) (string, error) {
	return Version(), nil
}

func SourceCommand(*IRC, string) (string, error) {
	return Source(), nil
}
