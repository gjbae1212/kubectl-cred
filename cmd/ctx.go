package cmd

import (
	"context"
	"fmt"

	"github.com/fatih/color"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gjbae1212/kubectl-cred/internal"
	"github.com/spf13/cobra"
)

var (
	ctxCmd = &cobra.Command{
		Use:    "ctx",
		Short:  "Switch context using an interactive CLI.",
		Long:   "Switch context using an interactive CLI.",
		PreRun: ctxPreRun(),
		Run:    ctxRun(),
	}
)

func ctxPreRun() commandRun {
	return func(cmd *cobra.Command, args []string) {}
}

func ctxRun() commandRun {
	return func(cmd *cobra.Command, args []string) {
		contexts, err := kubeConfig.GetContexts()
		if err != nil {
			internal.PanicWithRed(fmt.Errorf("[err] failed getting current kubernetes context. %w", err))
		}

		// ask context.
		contextName, err := askContext(contexts)
		if err != nil {
			internal.PanicWithRed(fmt.Errorf("[err] failed context selecting"))
		}

		if err := kubeConfig.SetCurrentContext(contextName); err != nil {
			internal.PanicWithRed(fmt.Errorf("[err] context: %s invalid format", contextName))
		}

		// get kubernetes client
		client, err := internal.NewKubeClient(kubeConfig)
		if err != nil {
			internal.PanicWithRed(fmt.Errorf("[err] context: %s invalid endpoint or credentials or unauthorized", contextName))
		}

		var next string
		var namespaces []string
		for {
			list, err := client.CoreV1().Namespaces().List(context.Background(),
				metav1.ListOptions{Limit: 100, Continue: next})
			if err != nil {
				internal.PanicWithRed(fmt.Errorf("[err] context: %s invalid endpoint or credentials or unauthorized", contextName))
			}

			next = list.Continue
			for _, it := range list.Items {
				namespaces = append(namespaces, it.Name)
			}
			if next == "" {
				break
			}
		}

		// ask namespaces
		namespace, err := askNamespace(contextName, namespaces)
		if err != nil {
			internal.PanicWithRed(fmt.Errorf("[err] context: %s unknown error", contextName))
		}

		if err := kubeConfig.SetNamespace(contextName, namespace); err != nil {
			internal.PanicWithRed(fmt.Errorf("[err] context: %s, namespace %s failed set",
				contextName, namespace))
		}

		if err := kubeConfig.Sync(); err != nil {
			internal.PanicWithRed(fmt.Errorf("[err] config sync error"))
		}
		fmt.Printf("%s %s %s, %s %s\n",
			color.GreenString("[success]"),
			color.YellowString("set context:"),
			color.CyanString(contextName),
			color.YellowString("set namespace:"),
			color.MagentaString(namespace),
		)
	}
}

func init() {
	rootCmd.AddCommand(ctxCmd)
}
