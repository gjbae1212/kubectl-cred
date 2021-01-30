package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/gjbae1212/kubectl-cred/internal"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "kubectl-cred",
		Short: color.CyanString(`kubectl-cred is an interactive CLI to be able to switch kubernetes context.`),
		Long:  color.CyanString(`kubectl-cred is an interactive CLI to be able to switch kubernetes context.`),
		Example: `
# Switch context using an interactive CLI.
kubectl cred ctx

# Switch namespace using an interactive CLI.
kubectl cred ns

# Show your all of contexts information formatted tree.
kubectl cred ls

# Show your current context information.
kubectl cred current

# Rename context name in k8s config, using an interactive CLI.
kubectl cred rename
`,
	}
)

var (
	kubeConfig *internal.KubeConfig
)

type commandRun func(cmd *cobra.Command, args []string)

// Execute is an entry point for kubectl-cred.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		internal.PanicWithRed(err)
	}
}

func init() {
	cobra.OnInitialize(initCobraConfig)
	rootCmd.SetHelpCommand(&cobra.Command{Use: "no-help", Hidden: true})
}

func initCobraConfig() {
	var err error

	path, err := internal.KubeConfigPath()
	if err != nil {
		internal.PanicWithRed(fmt.Errorf("[err] failed getting kube config path %w", err))
	}

	kubeConfig, err = internal.NewKubeConfig(path)
	if err != nil {
		internal.PanicWithRed(fmt.Errorf("[err] failed getting kube config %w", err))
	}
}

func askContext(contexts []*internal.KubeContext) (string, error) {
	var sortedContextNames []string
	for _, ctx := range contexts {
		sortedContextNames = append(sortedContextNames, ctx.Name)
	}

	prompt := &survey.Select{
		Message: fmt.Sprintf("Choose a context in k8s:"),
		Options: sortedContextNames,
	}

	selectKey := ""
	if err := survey.AskOne(prompt, &selectKey,
		survey.WithIcons(func(icons *survey.IconSet) { icons.SelectFocus.Format = "green+hb" }),
		survey.WithPageSize(20)); err != nil {
		return "", err
	}

	return selectKey, nil
}

func askNamespace(contextName string, namespaces []string) (string, error) {
	var sortedNamespaces []string
	sort.Strings(namespaces)
	for _, namespace := range namespaces {
		if namespace == "default" {
			sortedNamespaces = append([]string{namespace}, sortedNamespaces...)
		} else {
			sortedNamespaces = append(sortedNamespaces, namespace)
		}
	}

	prompt := &survey.Select{
		Message: fmt.Sprintf("Choose a namespace in k8s context:"),
		Options: sortedNamespaces,
	}

	selectKey := ""
	if err := survey.AskOne(prompt, &selectKey,
		survey.WithIcons(func(icons *survey.IconSet) { icons.SelectFocus.Format = "green+hb" }), survey.WithPageSize(20)); err != nil {
		return "", err
	}

	return selectKey, nil
}

func askWantedContextName() string {
	prompt := &survey.Input{
		Message: "Type to your wanted to change context name:",
	}
	var contextName string
	survey.AskOne(prompt, &contextName)
	contextName = strings.TrimSpace(contextName)
	return contextName
}

func askConfirmChangingContextName(currentContextName, changeContextName string) bool {
	ok := false
	prompt := &survey.Confirm{
		Message: fmt.Sprintf("Do you really want to change %s to %s", currentContextName, changeContextName),
	}
	survey.AskOne(prompt, &ok)
	return ok
}
