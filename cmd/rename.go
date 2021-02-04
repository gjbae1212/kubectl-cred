package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/gjbae1212/kubectl-cred/internal"
	"github.com/spf13/cobra"
)

var (
	renameCmd = &cobra.Command{
		Use:    "rename",
		Short:  "Rename context name in k8s config, using an interactive CLI.",
		Long:   "Rename context name in k8s config, using an interactive CLI.",
		PreRun: renamePreRun(),
		Run:    renameRun(),
	}
)

func renamePreRun() commandRun {
	return func(cmd *cobra.Command, args []string) {}
}

func renameRun() commandRun {
	return func(cmd *cobra.Command, args []string) {
		contexts, err := kubeConfig.GetContexts()
		if err != nil {
			internal.PanicWithRed(fmt.Errorf("[err] failed getting current kubernetes context. %w", err))
		}

		currentContext, err := kubeConfig.GetCurrentContext()
		if err != nil {
			internal.PanicWithRed(fmt.Errorf("[err] failed getting current kubernetes context. %w", err))
		}

		outdatedContextName, err := askContext(contexts)
		if err != nil {
			internal.PanicWithRed(fmt.Errorf("[err] failed context selecting"))
		}

		var needCurrentContext bool
		if currentContext.Name == outdatedContextName {
			needCurrentContext = true
		}

		wantedContextName := askWantedContextName()
		if wantedContextName == "" {
			internal.PanicWithRed(fmt.Errorf("[err] failed invalid context name"))
		} else {
			if err := kubeConfig.ChangeContextName(outdatedContextName, wantedContextName); err != nil {
				internal.PanicWithRed(fmt.Errorf("[err] failed changing context name"))
			}
		}

		if needCurrentContext {
			if err := kubeConfig.SetCurrentContext(wantedContextName); err != nil {
				internal.PanicWithRed(fmt.Errorf("[err] failed changing context name"))
			}
		}
		if ok := askConfirmChangingContextName(outdatedContextName, wantedContextName); !ok {
			return
		}

		if err := kubeConfig.Sync(); err != nil {
			internal.PanicWithRed(fmt.Errorf("[err] config sync error"))
		}
		fmt.Printf("%s %s %s, %s %s\n",
			color.GreenString("[success]"),
			color.YellowString("change context:"),
			color.CyanString(outdatedContextName),
			color.YellowString("to context:"),
			color.CyanString(wantedContextName),
		)
	}
}

func init() {
	rootCmd.AddCommand(renameCmd)
}
