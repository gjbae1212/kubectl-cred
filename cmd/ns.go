package cmd

import (
	"context"
	"fmt"

	"github.com/fatih/color"
	"github.com/gjbae1212/kubectl-cred/internal"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	nsCmd = &cobra.Command{
		Use:    "ns",
		Short:  "Switch namespace using an interactive CLI.",
		Long:   "Switch namespace using an interactive CLI.",
		PreRun: nsPreRun(),
		Run:    nsRun(),
	}
)

func nsPreRun() commandRun {
	return func(cmd *cobra.Command, args []string) {}
}

func nsRun() commandRun {
	return func(cmd *cobra.Command, args []string) {
		currentContext, err := kubeConfig.GetCurrentContext()
		if err != nil {
			internal.PanicWithRed(fmt.Errorf("[err] failed getting current kubernetes context. %w", err))
		}

		client, err := internal.NewKubeClient(kubeConfig)
		if err != nil {
			internal.PanicWithRed(fmt.Errorf("[err] context: %s invalid endpoint or credentials or unauthorized", currentContext.Name))
		}

		var next string
		var namespaces []string
		for {
			list, err := client.CoreV1().Namespaces().List(context.Background(),
				metav1.ListOptions{Limit: 100, Continue: next})
			if err != nil {
				internal.PanicWithRed(fmt.Errorf("[err] context: %s invalid endpoint or credentials or unauthorized", currentContext.Name))
			}

			next = list.Continue
			for _, it := range list.Items {
				namespaces = append(namespaces, it.Name)
			}
			if next == "" {
				break
			}
		}

		namespace, err := askNamespace(currentContext.Name, namespaces)
		if err != nil {
			internal.PanicWithRed(fmt.Errorf("[err] context: %s unknown error", currentContext.Name))
		}

		if err := kubeConfig.SetNamespace(currentContext.Name, namespace); err != nil {
			internal.PanicWithRed(fmt.Errorf("[err] context: %s, namespace %s failed set",
				currentContext.Name, namespace))
		}

		if err := kubeConfig.Sync(); err != nil {
			internal.PanicWithRed(fmt.Errorf("[err] config sync error"))
		}
		fmt.Printf("%s %s %s, %s %s\n",
			color.GreenString("[success]"),
			color.YellowString("set context:"),
			color.CyanString(currentContext.Name),
			color.YellowString("set namespace:"),
			color.MagentaString(namespace),
		)
	}
}

func init() {
	rootCmd.AddCommand(nsCmd)
}
