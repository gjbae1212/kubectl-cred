package internal

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"

	"github.com/mitchellh/go-homedir"
	"github.com/tomwright/dasel"
	"gopkg.in/yaml.v2"
)

const (
	kubeConfigEnv = "KUBECONFIG"
)

// KubeConfig is a struct for kubernetes config file.
type KubeConfig struct {
	rootNode   *dasel.Node
	originPath string
	prevPath   string
}

// KubeContext is a struct for kubernetes context.
type KubeContext struct {
	Name      string
	User      string
	Cluster   string
	Namespace string
	Server    string
}

// GetNamespace returns namespace
func (k *KubeContext) GetNamespace() string {
	if k.Namespace == "" {
		return "default"
	}
	return k.Namespace
}

// GetCurrentContext returns current context.
func (k *KubeConfig) GetCurrentContext() (*KubeContext, error) {
	node, err := k.rootNode.Query(".current-context")
	if err != nil {
		return nil, err
	}

	currentCtxName, ok := node.InterfaceValue().(string)
	if !ok {
		return nil, ErrUnknownValue
	}

	ctxs, err := k.GetContexts()
	if err != nil {
		return nil, err
	}

	for _, ctx := range ctxs {
		if ctx.Name == currentCtxName {
			return ctx, nil
		}
	}
	return &KubeContext{}, nil
}

// SetCurrentContext sets current context by passing param.
func (k *KubeConfig) SetCurrentContext(ctxName string) error {
	if err := k.rootNode.Put(".current-context", ctxName); err != nil {
		return err
	}
	return nil
}

// SetNamespace sets namespace in context.
func (k *KubeConfig) SetNamespace(ctxName, namespace string) error {
	node, err := k.rootNode.Query(fmt.Sprintf(".contexts.(.name=%s)", ctxName))
	if err != nil {
		return err
	}
	if err := node.Put(".context.namespace", namespace); err != nil {
		return err
	}
	return nil
}

// ChangeContextName changes context name in context section.
func (k *KubeConfig) ChangeContextName(ctxName, changedCtxName string) error {
	node, err := k.rootNode.Query(fmt.Sprintf(".contexts.(.name=%s)", ctxName))
	if err != nil {
		return err
	}
	if err := node.Put(".name", changedCtxName); err != nil {
		return err
	}
	return nil
}

// GetContexts returns kubernetes context list.
func (k *KubeConfig) GetContexts() ([]*KubeContext, error) {
	// get cluster nodes.
	clusterNode, err := k.rootNode.Query(".clusters")

	clusterMap := make(map[string]string)
	if clusterList, ok := clusterNode.InterfaceValue().([]interface{}); ok {
		for _, ctx := range clusterList {
			var clusterName string
			var clusterServer string
			for e1, v1 := range ctx.(map[interface{}]interface{}) {
				k1, err := InterfaceToString(e1)
				if err != nil {
					return nil, err
				}
				switch k1 {
				case "cluster":
					m, err := InterfaceToMap(v1)
					if err != nil {
						return nil, err
					}
					for k2, v2 := range m {
						switch k2 {
						case "server":
							clusterServer = v2
						}
					}
				case "name":
					v2, err := InterfaceToString(v1)
					if err != nil {
						return nil, err
					}
					clusterName = v2
				}
			}
			clusterMap[clusterName] = clusterServer
		}
	}

	// get context nodes.
	ctxNode, err := k.rootNode.Query(".contexts")
	if err != nil {
		return nil, err
	}

	var kubectxs []*KubeContext
	if ctxList, ok := ctxNode.InterfaceValue().([]interface{}); ok {
		for _, ctx := range ctxList {
			kubectx := &KubeContext{}
			for e1, v1 := range ctx.(map[interface{}]interface{}) {
				k1, err := InterfaceToString(e1)
				if err != nil {
					return nil, err
				}
				switch k1 {
				case "context":
					m, err := InterfaceToMap(v1)
					if err != nil {
						return nil, err
					}
					for k2, v2 := range m {
						switch k2 {
						case "user":
							kubectx.User = v2
						case "cluster":
							kubectx.Cluster = v2
							kubectx.Server = clusterMap[v2]
						case "namespace":
							kubectx.Namespace = v2
						}
					}
				case "name":
					v2, err := InterfaceToString(v1)
					if err != nil {
						return nil, err
					}
					kubectx.Name = v2
				}
			}
			kubectxs = append(kubectxs, kubectx)
		}
	} else {
		return nil, ErrInvalidFileFormat
	}

	return kubectxs, nil
}

// DeleteContext deletes context.
func (k *KubeConfig) DeleteContext(ctxName string) error {
	contexts, err := k.GetContexts()
	if err != nil {
		return err
	}

	var deleteContext *KubeContext
	for _, c := range contexts {
		if c.Name == ctxName {
			deleteContext = c
			break
		}
	}
	if deleteContext == nil {
		return nil
	}

	unmarshalYaml := k.rootNode.OriginalValue
	if yamlMap, ok := unmarshalYaml.(map[interface{}]interface{}); ok {
		for name, value := range yamlMap {
			switch name.(string) {
			case "clusters":
				changed, err := k.deleteByName(value.([]interface{}), deleteContext.Cluster)
				if err != nil {
					return err
				}
				yamlMap[name] = changed
			case "contexts":
				changed, err := k.deleteByName(value.([]interface{}), deleteContext.Name)
				if err != nil {
					return err
				}
				yamlMap[name] = changed
			case "current-context":
				valueString, err := InterfaceToString(value)
				if err != nil {
					return err
				}
				if valueString == deleteContext.Name {
					yamlMap[name] = ""
				}
			case "users":
				changed, err := k.deleteByName(value.([]interface{}), deleteContext.User)
				if err != nil {
					return err
				}
				yamlMap[name] = changed
			}
		}
		k.rootNode.Value = reflect.ValueOf(yamlMap)
		k.rootNode.OriginalValue = yamlMap
	}
	return nil
}

// Sync syncs rootNode to file.
func (k *KubeConfig) Sync() error {
	// copy origin file to prev file.
	readOnlyOriginFile, err := os.Open(k.originPath)
	if err != nil {
		return err
	}
	defer readOnlyOriginFile.Close()

	prevFile, err := os.OpenFile(k.prevPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer prevFile.Close()

	if _, err := io.Copy(prevFile, readOnlyOriginFile); err != nil {
		return err
	}

	raw, err := k.RawBytes()
	if err != nil {
		return err
	}

	// write changed config to origin path.
	if err := ioutil.WriteFile(k.originPath, raw, 0600); err != nil {
		return err
	}
	return nil
}

// RawBytes returns kubernetes config raw bytes.
func (k *KubeConfig) RawBytes() ([]byte, error) {
	bys, err := yaml.Marshal(k.rootNode.InterfaceValue())
	if err != nil {
		return nil, err
	}
	return bys, err
}

func (k *KubeConfig) deleteByName(arr []interface{}, name string) ([]interface{}, error) {
	deleteIndex := -1
Loop:
	for index, data := range arr {
		if m, ok := data.(map[interface{}]interface{}); ok {
			for k, v := range m {
				if k.(string) == "name" {
					vStr, err := InterfaceToString(v)
					if err != nil {
						return nil, err
					}
					if vStr == name {
						deleteIndex = index
						break Loop
					}
				}
			}
		}
	}

	if deleteIndex != -1 {
		arr = append(arr[:deleteIndex], arr[deleteIndex+1:]...)
	}

	return arr, nil
}

// NewKubeConfig returns struct for kubernetes config file.
func NewKubeConfig(path string) (*KubeConfig, error) {
	if path == "" {
		return nil, ErrInvalidParams
	}

	// read kubernetes config file.
	f, err := os.OpenFile(path, os.O_RDONLY, 0600)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	bys, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(bys) == 0 {
		return nil, ErrInvalidFileFormat
	}

	var conf interface{}
	if err := yaml.Unmarshal(bys, &conf); err != nil {
		return nil, err
	}

	rootNode := dasel.New(conf)
	return &KubeConfig{
		rootNode:   rootNode,
		originPath: path,
		prevPath:   fmt.Sprintf("%s_prev.bak", path),
	}, nil
}

// KubeConfigPath returns a config path for kubernetes.
func KubeConfigPath() (string, error) {
	if v := os.Getenv(kubeConfigEnv); v != "" {
		// if kubeConfigEnv exists.
		if cpath, err := homedir.Expand(v); err == nil {
			if cpath, err = filepath.Abs(cpath); err == nil {
				if stat, err := os.Stat(cpath); err == nil && !stat.IsDir() {
					return cpath, nil
				}
			}
		}
	}

	// get home directory.
	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}

	cfgPath := filepath.Join(home, ".kube", "config")
	if stat, err := os.Stat(cfgPath); err == nil && !stat.IsDir() {
		return cfgPath, nil
	}

	if os.IsNotExist(err) {
		return "", fmt.Errorf("%w (%s)", ErrNotFoundPath, cfgPath)
	} else {
		return "", err
	}
}
