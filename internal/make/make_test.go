package make

import (
	"reflect"
	"testing"
)

func TestParseTargets_SimpleTarget(t *testing.T) {
	data := []byte("build:\n\techo hi\n")
	got := parseTargets(data)
	want := []string{"build"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseTargets() = %v, want %v", got, want)
	}
}

func TestParseTargets_MultipleTargets(t *testing.T) {
	data := []byte("build:\n\techo build\ntest:\n\techo test\nlint:\n\techo lint\n")
	got := parseTargets(data)
	want := []string{"build", "test", "lint"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseTargets() = %v, want %v", got, want)
	}
}

func TestParseTargets_FilterPhonyAndVariables(t *testing.T) {
	data := []byte(`.PHONY: build test
CC = gcc
CFLAGS = -Wall

build:
	$(CC) $(CFLAGS) -o app main.c

test:
	go test ./...
`)
	got := parseTargets(data)
	want := []string{"build", "test"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseTargets() = %v, want %v", got, want)
	}
}

func TestParseTargets_SkipRecipeLines(t *testing.T) {
	data := []byte("build:\n\techo building\n\tgo build ./...\n")
	got := parseTargets(data)
	want := []string{"build"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseTargets() = %v, want %v", got, want)
	}
}

func TestParseTargets_SkipSpaceIndentedLines(t *testing.T) {
	data := []byte("build:\n        echo building\n")
	got := parseTargets(data)
	want := []string{"build"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseTargets() = %v, want %v", got, want)
	}
}

func TestParseTargets_EmptyContent(t *testing.T) {
	data := []byte("")
	got := parseTargets(data)
	if got != nil {
		t.Errorf("parseTargets() = %v, want nil", got)
	}
}

func TestParseTargets_OnlyComments(t *testing.T) {
	data := []byte("# This is a Makefile\n# No targets here\n")
	got := parseTargets(data)
	if got != nil {
		t.Errorf("parseTargets() = %v, want nil", got)
	}
}

func TestParseTargets_TargetWithDependencies(t *testing.T) {
	data := []byte("build: clean\n\tgo build ./...\nclean:\n\trm -f app\n")
	got := parseTargets(data)
	want := []string{"build", "clean"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseTargets() = %v, want %v", got, want)
	}
}

func TestParseTargets_SkipVariableAssignments(t *testing.T) {
	data := []byte("APP_NAME := myapp\nVERSION ?= 1.0\nbuild:\n\techo hi\n")
	got := parseTargets(data)
	want := []string{"build"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseTargets() = %v, want %v", got, want)
	}
}

func TestParseTargets_DotPrefixFiltered(t *testing.T) {
	data := []byte(".DEFAULT: build\nbuild:\n\tgo build\n")
	got := parseTargets(data)
	want := []string{"build"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("parseTargets() = %v, want %v", got, want)
	}
}
