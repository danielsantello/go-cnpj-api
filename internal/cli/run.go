package cli

import "fmt"

func Run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("command is required")
	}

	switch args[0] {
	case "serve":
		return serve()
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}
