package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func getEnv(key string) (string, error) {
	v, exists := os.LookupEnv(key)
	if !exists {
		return "", fmt.Errorf("variable %q not set", key)
	}
	return v, nil
}

func run(args []string) error {
	flags := flag.NewFlagSet("go run ./cmd", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	inputPath := flags.String("i", "", "input dir")
	outputPath := flags.String("o", "", "output dir")
	if err := flags.Parse(args); err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}
	if inputPath != nil && *inputPath == "" || outputPath != nil && *outputPath == "" || flags.NArg() != 0 {
		return fmt.Errorf("usage: go run ./cmd -i <input-dir> -o <output-dir>")
	}

	err := os.MkdirAll(*outputPath, 0755)
	if err != nil {
		return fmt.Errorf("create output directory %q: %w", *outputPath, err)
	}

	inputDirContents, err := os.ReadDir(*inputPath)
	if err != nil {
		return fmt.Errorf("read template input dir %q: %w", *inputPath, err)
	}

	for _, entry := range inputDirContents {
		if entry.IsDir() {
			continue
		}

		templatePath := fmt.Sprintf("%s/%s", *inputPath, entry.Name())
		data, err := os.ReadFile(templatePath)
		if err != nil {
			return fmt.Errorf("read template %q: %w", templatePath, err)
		}

		rendered, err := template.New(templatePath).Funcs(template.FuncMap{
			"env":        getEnv,
			"lower":      strings.ToLower,
			"trimSuffix": strings.TrimSuffix,
		}).Option("missingkey=error").Parse(string(data))
		if err != nil {
			return fmt.Errorf("parse template %q: %w", templatePath, err)
		}

		var output bytes.Buffer
		if err := rendered.Execute(&output, nil); err != nil {
			return fmt.Errorf("render template %q: %w", templatePath, err)
		}

		outputFilePath := fmt.Sprintf("%s/%s", *outputPath, entry.Name())
		if err := os.WriteFile(outputFilePath, output.Bytes(), 0o644); err != nil {
			return fmt.Errorf("write output %q: %w", outputFilePath, err)
		}
	}

	return nil
}
