package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/mitchellh/go-homedir"
	"github.com/stretchr/testify/assert"
)

func TestNewKubeConfig(t *testing.T) {
	assert := assert.New(t)

	cfgPath, err := KubeConfigPath()
	if err != nil {
		return
	}

	tests := map[string]struct {
		path     string
		tempPath string
		prevPath string
		isErr    bool
	}{
		"error-1": {isErr: true},
		"error-2": {path: "empty", isErr: true},
		"success": {path: cfgPath, tempPath: fmt.Sprintf("%s_temp.bak", cfgPath),
			prevPath: fmt.Sprintf("%s_prev.bak", cfgPath), isErr: false},
	}

	for _, t := range tests {
		kube, err := NewKubeConfig(t.path)
		assert.Equal(t.isErr, err != nil)
		if err == nil {
			assert.Equal(t.path, kube.originPath)
			assert.Equal(t.prevPath, kube.prevPath)
			fmt.Println(kube.originPath)
			fmt.Println(kube.prevPath)
		}
	}
}

func TestKubeConfig_GetCurrentContext(t *testing.T) {
	assert := assert.New(t)

	cfgPath, err := KubeConfigPath()
	if err != nil {
		return
	}

	kubeConfig, err := NewKubeConfig(cfgPath)
	assert.NoError(err)

	tests := map[string]struct {
		cfg   *KubeConfig
		isErr bool
	}{
		"success": {cfg: kubeConfig},
	}

	for _, t := range tests {
		ctx, err := kubeConfig.GetCurrentContext()
		assert.Equal(t.isErr, err != nil)
		fmt.Println(ctx)
	}
}

func TestKubeConfig_SetCurrentContext(t *testing.T) {
	assert := assert.New(t)

	cfgPath, err := KubeConfigPath()
	if err != nil {
		return
	}

	kubeConfig, err := NewKubeConfig(cfgPath)
	assert.NoError(err)

	tests := map[string]struct {
		cfg   *KubeConfig
		input string
		isErr bool
	}{
		"success": {cfg: kubeConfig, input: "minikube"},
	}

	for _, t := range tests {
		err := kubeConfig.SetCurrentContext(t.input)
		assert.Equal(t.isErr, err != nil)
		if err == nil {
			ctx, err := kubeConfig.GetCurrentContext()
			assert.NoError(err)
			assert.Equal(t.input, ctx.Name)
		}

	}
}

func TestKubeConfig_SetNamespace(t *testing.T) {
	assert := assert.New(t)

	cfgPath, err := KubeConfigPath()
	if err != nil {
		return
	}

	kubeConfig, err := NewKubeConfig(cfgPath)
	assert.NoError(err)

	tests := map[string]struct {
		contextName string
		namespace   string
		isErr       bool
	}{
		"success-1": {
			contextName: "minikube",
			namespace:   "test",
		},
	}

	for _, t := range tests {
		err := kubeConfig.SetNamespace(t.contextName, t.namespace)
		assert.Equal(t.isErr, err != nil)
		if err == nil {
			contexts, err := kubeConfig.GetContexts()
			assert.NoError(err)
			for _, ctx := range contexts {
				if ctx.Name == t.contextName {
					assert.Equal(t.namespace, ctx.Namespace)
				}
			}
		}
	}
}

func TestKubeConfig_GetContexts(t *testing.T) {
	assert := assert.New(t)

	cfgPath, err := KubeConfigPath()
	if err != nil {
		return
	}

	kubeConfig, err := NewKubeConfig(cfgPath)
	assert.NoError(err)

	tests := map[string]struct {
		input *KubeConfig
		isErr bool
	}{
		"success": {input: kubeConfig},
	}

	for _, t := range tests {
		ctxs, err := kubeConfig.GetContexts()
		assert.Equal(t.isErr, err != nil)
		if err == nil {
			for _, ctx := range ctxs {
				assert.NotEmpty(ctx.Name)
				assert.NotEmpty(ctx.Cluster)
				assert.NotEmpty(ctx.User)
				assert.NotEmpty(ctx.Server)
			}
		}

	}
}

func TestKubeConfig_Sync(t *testing.T) {
	assert := assert.New(t)

	cfgPath, err := KubeConfigPath()
	if err != nil {
		return
	}

	kubeConfig, err := NewKubeConfig(cfgPath)
	assert.NoError(err)

	tests := map[string]struct {
		isErr bool
	}{
		"success": {},
	}

	for _, t := range tests {
		err := kubeConfig.Sync()
		assert.Equal(t.isErr, err != nil)
	}

}

func TestKubeConfigPath(t *testing.T) {
	assert := assert.New(t)

	home, err := homedir.Dir()
	assert.NoError(err)
	currentDir, err := homedir.Expand(".")
	assert.NoError(err)
	currentDir, err = filepath.Abs(currentDir)
	assert.NoError(err)

	tests := map[string]struct {
		envs   map[string]string
		output string
		isErr  bool
	}{
		"success-1": {
			envs:   map[string]string{kubeConfigEnv: "~/empty"},
			output: filepath.Join(home, ".kube", "config"),
		},
		"success-2": {
			envs:   map[string]string{kubeConfigEnv: ""},
			output: filepath.Join(home, ".kube", "config"),
		},
		"success-3": {
			envs:   map[string]string{kubeConfigEnv: "~/"},
			output: filepath.Join(home, ".kube", "config"),
		},
		"success-4": {
			envs:   map[string]string{kubeConfigEnv: "kubeconfig.go"},
			output: filepath.Join(currentDir, "kubeconfig.go"),
		},
	}

	for _, t := range tests {
		for k, v := range t.envs {
			os.Setenv(k, v)
		}
		cpath, err := KubeConfigPath()
		assert.Equal(t.isErr, err != nil)
		if err == nil {
			assert.Equal(t.output, cpath)
		}
	}
}
