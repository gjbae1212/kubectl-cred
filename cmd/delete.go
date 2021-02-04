package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/gjbae1212/kubectl-cred/internal"
	"github.com/spf13/cobra"
)

var (
	deleteCmd = &cobra.Command{
		Use:    "delete",
		Short:  "Delete context in k8s config, using an interactive CLI.",
		Long:   "Delete context in k8s config, using an interactive CLI.",
		PreRun: deletePreRun(),
		Run:    deleteRun(),
	}
)

func deletePreRun() commandRun {
	return func(cmd *cobra.Command, args []string) {}
}

func deleteRun() commandRun {
	return func(cmd *cobra.Command, args []string) {
		contexts, err := kubeConfig.GetContexts()
		if err != nil {
			internal.PanicWithRed(fmt.Errorf("[err] failed getting current kubernetes context. %w", err))
		}

		deleteContextName, err := askContext(contexts)
		if err != nil {
			internal.PanicWithRed(fmt.Errorf("[err] failed context selecting"))
		}

		if ok := askConfirmDeletingContextName(deleteContextName); !ok {
			return
		}

		if err := kubeConfig.DeleteContext(deleteContextName); err != nil {
			internal.PanicWithRed(fmt.Errorf("[err] failed delete context"))
		}

		if err := kubeConfig.Sync(); err != nil {
			internal.PanicWithRed(fmt.Errorf("[err] config sync error"))
		}

		fmt.Printf("%s %s %s\n",
			color.GreenString("[success]"),
			color.YellowString("delete context:"),
			color.CyanString(deleteContextName),
		)
	}
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
