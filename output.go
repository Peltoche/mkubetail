package main

import (
	"bufio"
	"fmt"
	"log"
	"sync"
)

// PrintOutput read from  the "Out" readCloser inside each pod and write the
// raw content into Stdout.
func PrintOutput(pods []Pod, lineCfg *LineConfig) {
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
				}

				fmt.Print(formatLogLine(string(line), pod, lineCfg))
			}
		}(pod)
	}

	wg.Wait()
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
