package main

import (
	"bufio"
	"fmt"
	"log"
	"os/exec"
	"sync"
	"time"
)

// Cmd describes the command needed to be runned and the rules used to choose
// the selected contexts.
type Cmd struct {
	Contexts   []string
	Pods       []string
	Duration   time.Duration
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

	err = startAsyncPodWatching(pods, cmd)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	for _, pod := range pods {
		wg.Add(1)

		go func(pod Pod) {
			defer wg.Done()

			reader := bufio.NewReader(pod.Out)

			for {
				line, _, err := reader.ReadLine()
				if err != nil {
					log.Println(err)
					return
				}

				fmt.Print(formatLogLine(string(line), pod, &cmd.LineConfig))
			}
		}(pod)
	}

	wg.Wait()

	return nil
}

func startAsyncPodWatching(pods []Pod, cmd *Cmd) error {
	for idx, pod := range pods {
		opts := []string{"--context=" + pod.Context, "logs", "-f", pod.Name}

		if cmd.Duration != 0 {
			opts = append(opts, "--since="+cmd.Duration.String())
		}

		// Little mock because I haven't access to a cluster the weekend.
		kubeCmd := exec.Command("kubectl", opts...)
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

func formatLogLine(content string, pod Pod, cfg *LineConfig) string {
	var prefix string

	if cfg.ShowContextName {
		prefix = fmt.Sprintf("[%s]", pod.Context)
	}

	if cfg.ShowPodName {
		prefix = fmt.Sprintf("%s[%s]", prefix, pod.Name)
	}

	if cfg.ShowContextName || cfg.ShowPodName {
		prefix = prefix + " "
	}

	return prefix + content + "\n"
}
