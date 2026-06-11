package run

import "strings"

type Import struct {
	Alias string
	Path  string
}

type Config struct {
	ProjectModule string
	ProjectDir    string
	Code          string
	Imports       []Import
}

func ParseCode(input string) (code string, imports []Import) {
	var lines []string

	for _, line := range strings.Split(input, "\n") {
		t := strings.TrimSpace(line)

		if strings.HasPrefix(t, "// import ") {
			p := strings.Trim(strings.TrimPrefix(t, "// import "), `"`)
			imports = append(imports, Import{Path: p})
			continue
		}

		if strings.HasPrefix(t, "import ") {
			p := strings.Trim(strings.TrimPrefix(t, "import "), `"`)
			imports = append(imports, Import{Path: p})
			continue
		}

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n"), imports
}
