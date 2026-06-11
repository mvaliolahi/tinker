package run

const mainTmpl = `package main

import (
	"context"
	"fmt"
	"log"
	"os"
	{{- range .Imports}}
	{{.Alias}} "{{.Path}}"
	{{- end}}
)

func main() {
	ctx := context.Background()
	_ = ctx

	{{.Code}}
}
`
