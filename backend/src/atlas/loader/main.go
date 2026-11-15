package main

import (
	"fmt"
	"os"

	"tex/src/internal/domain"

	"ariga.io/atlas-provider-gorm/gormschema"
)

func main() {
	stmts, err := gormschema.New("postgres").Load(
		&domain.User{},
		&domain.Submission{},
		&domain.Schedule{},
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load gorm schema: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(stmts)
}
