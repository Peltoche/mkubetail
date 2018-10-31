package main

import (
	"fmt"
	"os/exec"
)

// Cmd describes the command needed to be runned and the rules used to choose
// the selected contexts.
type Cmd struct {
	Contexts   []string
	Pods       []string
	LineConfig LineConfig
}

// LineConfig describe the configs choosed for the log line representation.
type LineConfig struct {
	ShowContextName bool
	ShowPodName     bool
}

// Tail the all the matchings pod inside the matchin contexts described inside
// the cmd argument.
func Tail(cmd *Cmd) error {
	contexts, err := SelectMatchingContexts(cmd.Contexts)
	if err != nil {
		return err
	}

	pods := SelectMatchingPods(contexts, cmd.Pods)

	err = startAsyncPodWatching(pods)
	if err != nil {
		return err
	}

	PrintOutput(pods, &cmd.LineConfig)
	return nil
}

func startAsyncPodWatching(pods []Pod) error {
	for idx, pod := range pods {
		// Little mock because I haven't access to a cluster the weekend.
		kubeCmd := exec.Command("kubectl", "--context="+pod.Context, "logs", "-f", pod.Name)
		//kubeCmd := exec.Command("ping", "google.com")

		// Can be used for debugging purpose with not kubernetes available.
		cmdOut, err := kubeCmd.StdoutPipe()
		if err != nil {
			return err
		}

		pods[idx].Out = cmdOut

		go func() {
			err := kubeCmd.Start()
			if err != nil {
				fmt.Println(err)
			}
		}()
	}

	return nil
}
