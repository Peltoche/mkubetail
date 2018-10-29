package main

import (
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strings"
	"sync"
)

// Pod resource description.
type Pod struct {
	Context string
	Name    string
	Out     io.ReadCloser
}

// ID return an unique ID string used to identify a pod.
func (t *Pod) ID() string {
	return t.Context + "-" + t.Name
}

// SelectMatchingPods take the cli arguments, retrieves all the available pods
// on the specified contexts and return only the ones matchins the arguments.
//
// In case of failure, return nil and the corresponding error.
func SelectMatchingPods(contexts []string, args []string) []Pod {
	pods := RetrieveAllPods(contexts)

	if len(args) > 0 {
		pods = filterPodsWithArgs(pods, args)
	}

	return pods
}

// RetrieveAllPods query all the contexts and returns all the known pods.
func RetrieveAllPods(contexts []string) []Pod {
	var wg sync.WaitGroup
	out := make(chan Pod)

	for _, context := range contexts {
		wg.Add(1)
		go retrieveAllContextPods(context, &wg, out)
	}

	// Close the out chan one all the thread are finished.
	go func(out chan Pod) {
		wg.Wait()
		close(out)
	}(out)

	res := make([]Pod, 0, len(out))
	for pod := range out {
		res = append(res, pod)
	}

	return res
}

func retrieveAllContextPods(context string, wg *sync.WaitGroup, out chan Pod) {
	defer wg.Done()

	cmd := exec.Command("kubectl", "--context="+context, "get", "pod", "-o=jsonpath='{.items..metadata.name}'")
	rawOut, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(err)
	}

	// Need some parsing
	stringOut := strings.Trim(string(rawOut), "'")
	pods := strings.Split(stringOut, " ")

	for _, podName := range pods {
		out <- Pod{
			Context: context,
			Name:    podName,
		}
	}
}

func filterPodsWithArgs(pods []Pod, args []string) []Pod {
	// Use a map[string]struct{} in order to create a base Set data structure.
	// It allows to avoid ady duplicates.
	set := make(map[string]Pod, len(pods))

	for _, arg := range args {
		for _, pod := range pods {
			matched, err := regexp.MatchString(arg, pod.Name)
			if err != nil {
				fmt.Println(err)
				continue
			}

			if matched {
				set[pod.Context+pod.Name] = pod
			}
		}
	}

	res := make([]Pod, 0, len(pods))
	for _, pod := range set {
		res = append(res, pod)
	}

	return res
}
