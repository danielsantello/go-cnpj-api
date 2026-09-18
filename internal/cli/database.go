package cli

import "fmt"

func runDatabase(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("database command is required")
	}

	switch args[0] {
	case "create-indexes":
		return createIndexes()

	default:
		return fmt.Errorf("unknown database command: %s", args[0])
	}
}
