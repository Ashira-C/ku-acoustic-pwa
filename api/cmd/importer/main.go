package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"github.com/Ashira-C/ku-acoustic-pwa/api/internal/importer"
	"github.com/Ashira-C/ku-acoustic-pwa/api/internal/platform"
)

func main() {
	membersPath := flag.String("members", "", "members.csv exported from the club roster (required)")
	correctionsPath := flag.String("corrections", "", "corrections.csv with manual fixes (optional)")
	apply := flag.Bool("apply", false, "write to the database (default is a dry run that never touches the database)")
	flag.Parse()

	if err := run(*membersPath, *correctionsPath, *apply); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(membersPath, correctionsPath string, apply bool) error {
	if membersPath == "" {
		return errors.New("-members is required")
	}

	rows, err := readFile(membersPath, importer.ReadMembers)
	if err != nil {
		return err
	}
	var corrections []importer.Correction
	if correctionsPath != "" {
		if corrections, err = readFile(correctionsPath, importer.ReadCorrections); err != nil {
			return err
		}
	}

	res := importer.Build(rows, corrections)
	printReport(res)

	if len(res.Errors) > 0 {
		return fmt.Errorf("%d error(s): fix them before importing", len(res.Errors))
	}
	if !apply {
		fmt.Println("\nDry run OK: nothing was written. Re-run with -apply to import.")
		return nil
	}

	// .env is optional; variables already set in the environment always win.
	_ = godotenv.Load()
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return errors.New("DATABASE_URL is not set")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := platform.NewPostgresPool(ctx, databaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	if err := importer.Apply(ctx, pool, res.Members); err != nil {
		return fmt.Errorf("import failed, nothing was written: %w", err)
	}
	fmt.Printf("\nImported %d members.\n", len(res.Members))
	return nil
}

// readFile opens a file and hands it to a parser (works for both CSV types via generics).
func readFile[T any](path string, parse func(io.Reader) ([]T, error)) ([]T, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	items, err := parse(f)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return items, nil
}

func printReport(res importer.Result) {
	for _, w := range res.Warnings {
		fmt.Println("WARN ", w)
	}
	for _, e := range res.Errors {
		fmt.Println("ERROR", e)
	}

	roles := make(map[string]int)
	parts := make(map[string]int)
	for _, m := range res.Members {
		roles[m.Role]++
		for _, p := range m.Parts {
			parts[p]++
		}
	}
	fmt.Printf("\nmembers: %d | warnings: %d | errors: %d\n", len(res.Members), len(res.Warnings), len(res.Errors))
	fmt.Println("roles:", roles) // fmt prints maps sorted by key
	fmt.Println("parts:", parts)
}