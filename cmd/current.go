package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/gjbae1212/kubectl-cred/internal"
	"github.com/mattn/go-colorable"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var (
	currentCmd = &cobra.Command{
		Use:    "current",
		Short:  "Show your current kubernetes context information",
		Long:   "Show your current kubernetes context information",
		PreRun: currentPreRun(),
		Run:    currentRun(),
	}
)

func currentPreRun() commandRun {
	return func(cmd *cobra.Command, args []string) {}
}

func currentRun() commandRun {
	return func(cmd *cobra.Command, args []string) {
		contexts, err := kubeConfig.GetContexts()
		if err != nil {
			internal.PanicWithRed(fmt.Errorf("[err] failed getting current kubernetes context. %w", err))
		}

		currentContext, err := kubeConfig.GetCurrentContext()
		if err != nil {
			internal.PanicWithRed(fmt.Errorf("[err] failed getting current kubernetes context. %w", err))
		}

		// table writer
		table := tablewriter.NewWriter(colorable.NewColorableStdout())
		table.SetHeader([]string{"CONTEXT NAME", "NAMESPACE"})
		table.SetBorder(false)
		table.SetRowLine(true)
		table.SetAlignment(tablewriter.ALIGN_CENTER)
		table.SetRowLine(true)
		table.SetHeaderColor(
			tablewriter.Colors{tablewriter.Bold, tablewriter.BgGreenColor},
			tablewriter.Colors{tablewriter.Bold, tablewriter.BgMagentaColor})
		table.SetColumnColor(tablewriter.Colors{tablewriter.Bold}, tablewriter.Colors{tablewriter.Bold})

		data := [][]string{}
		for _, ctx := range contexts {
			if ctx.Name == currentContext.Name {
				data = append(data, []string{color.GreenString(ctx.Name), color.MagentaString(ctx.GetNamespace())})
			} else {
				data = append(data, []string{ctx.Name, ctx.GetNamespace()})
			}
		}
		table.AppendBulk(data)
		table.Render()
	}
}

func init() {
	rootCmd.AddCommand(currentCmd)
}
