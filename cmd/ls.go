package cmd

import (
	"context"
	"fmt"
	"sort"

	"github.com/disiqueira/gotree"
	"github.com/fatih/color"
	"github.com/gjbae1212/kubectl-cred/internal"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	lsCmd = &cobra.Command{
		Use:    "ls",
		Short:  "Show your all of k8s cluster namespaces formatted tree.",
		Long:   "Show your all of k8s cluster namespaces formatted tree.",
		PreRun: lsPreRun(),
		Run:    lsRun(),
	}
)

func lsPreRun() commandRun {
	return func(cmd *cobra.Command, args []string) {}
}

func lsRun() commandRun {
	return func(cmd *cobra.Command, args []string) {
		contexts, err := kubeConfig.GetContexts()
		if err != nil {
			internal.PanicWithRed(fmt.Errorf("[err] failed getting current kubernetes context. %w", err))
		}

		var errors []string
		tree := make(map[string][]string)
		for _, ctx := range contexts {
			// change kubernetes context.
			if err := kubeConfig.SetCurrentContext(ctx.Name); err != nil {
				fmt.Println(color.RedString("[err] context: %s invalid format", ctx.Name))
				continue
			}

			// get kubernetes client
			client, err := internal.NewKubeClient(kubeConfig)
			if err != nil {
				errors = append(errors, color.RedString("[err] context: %s invalid endpoint or credentials or unauthorized", ctx.Name))
				continue
			}

			var next string
			var namespaces []string
			var success bool
			for {
				list, err := client.CoreV1().Namespaces().List(context.Background(),
					metav1.ListOptions{Limit: 100, Continue: next})
				if err != nil {
					success = false
					errors = append(errors, color.RedString("[err] context: %s invalid endpoint or credentials or unauthorized", ctx.Name))
					break
				}

				next = list.Continue
				for _, it := range list.Items {
					namespaces = append(namespaces, it.Name)
				}
				if next == "" {
					success = true
					break
				}
			}
			if success {
				sort.Strings(namespaces)
				tree[ctx.Name] = namespaces
			}
		}

		root := gotree.New(color.CyanString("cluster"))
		for k, v := range tree {
			child := root.Add(color.GreenString(k))
			for _, vv := range v {
				child.Add(color.MagentaString(vv))
			}
		}
		fmt.Println(root.Print())
		for _, e := range errors {
			fmt.Println(e)
		}
	}
}

func init() {
	rootCmd.AddCommand(lsCmd)
}
