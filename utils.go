package main

import (
	"fmt"
)

func must(err error) {
	if err != nil {
		panic(fmt.Errorf("fatal error: %w", err))
	}
}
