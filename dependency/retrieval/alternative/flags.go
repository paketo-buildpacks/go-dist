package alternative

import "flag"

func ParseFlags(args []string) (string, string, error) {
	var b, o string
	set := flag.NewFlagSet("", flag.ContinueOnError)
	set.StringVar(&b, "buildpack-toml-path", "", "path to the buildpack.toml file")
	set.StringVar(&o, "output", "", "path to the output file")
	err := set.Parse(args)
	if err != nil {
		return "", "", err
	}

	return b, o, nil
}
