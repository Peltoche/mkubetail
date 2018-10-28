package main

import (
	"fmt"
	"io/ioutil"
	"path"
	"regexp"

	"github.com/ghodss/yaml"
	homedir "github.com/mitchellh/go-homedir"
)

// SelectMatchingContexts take the cli arguments, retrieves all the available
// contexts and return only the ones matching the arguments.
//
// In case of failure, return nil and the corresponding error.
func SelectMatchingContexts(args []string) ([]string, error) {
	contexts, err := RetrieveAllContexts()
	if err != nil {
		return nil, err
	}

	if len(args) > 0 {
		contexts = filterContextsWithArgs(contexts, args)
	}

	return contexts, nil
}

// RetrieveAllContexts reads the k8s configuration file and returns all the
// known contexts
func RetrieveAllContexts() ([]string, error) {
	type kubeConfig struct {
		Contexts []struct {
			Name string
		}
	}

	configPath, err := getDefaultConfigPath()
	if err != nil {
		return nil, err
	}

	rawCfgFile, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var cfg kubeConfig
	err = yaml.Unmarshal(rawCfgFile, &cfg)
	if err != nil {
		return nil, err
	}

	contextNames := make([]string, len(cfg.Contexts))

	for idx, ctx := range cfg.Contexts {
		contextNames[idx] = ctx.Name
	}

	return contextNames, nil
}

func getDefaultConfigPath() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}

	configPath := path.Join(home, ".kube", "config")

	return configPath, nil
}

func filterContextsWithArgs(contexts []string, args []string) []string {
	// Use a map[string]struct{} in order to create a base Set data structure.
	// It allows to avoid ady duplicates.
	set := make(map[string]struct{}, len(contexts))

	for _, arg := range args {
		for _, context := range contexts {
			matched, err := regexp.MatchString(arg, context)
			if err != nil {
				fmt.Println(err)
				continue
			}

			if matched {
				set[context] = struct{}{}
			}
		}
	}

	res := make([]string, 0, len(contexts))
	for context := range set {
		res = append(res, context)
	}

	return res
}
