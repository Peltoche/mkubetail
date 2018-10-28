package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	// Set the flag variables.
	var matchsOpt []string
	var rawOpt bool
	var podNamePrefixOpt bool
	var contextNamePrefixOpt bool

	// Define the cli.
	rootCmd := cobra.Command{
		Use:   "mkubetail [pods...]",
		Short: "Mkubetail tails several pods on several contexts at the same time.",

		// Run only parse the options and trigger the Execute function which
		// contains all the stuff. It allows to isolate the cli logic.
		Run: func(cmd *cobra.Command, args []string) {
			err := Tail(&Cmd{
				Contexts: matchsOpt,
				Pods:     args,
				Raw:      rawOpt,
				LineConfig: LineConfig{
					ShowPodName:     podNamePrefixOpt,
					ShowContextName: contextNamePrefixOpt,
				},
			})
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}

	// Link the flag variables to the cli.
	rootCmd.PersistentFlags().StringArrayVarP(&matchsOpt, "context", "c", []string{}, "Context name or regex. All contexts are used if not specified.")
	rootCmd.PersistentFlags().BoolVarP(&rawOpt, "raw", "r", false, "Merge all the logs together and return an output similar to \"tail -f\"")
	rootCmd.PersistentFlags().BoolVarP(&podNamePrefixOpt, "pod-name", "p", false, "Prefix each line with the pod's name")
	rootCmd.PersistentFlags().BoolVarP(&contextNamePrefixOpt, "context-name", "C", false, "Prefix each line with the pod's context name")

	// Run the cli.
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
