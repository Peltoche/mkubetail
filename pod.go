package main

import (
	"fmt"
	"io"
	"os"
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
	outChan := make(chan Pod)
	errChan := make(chan error)

	for _, context := range contexts {
		wg.Add(1)
		go retrieveAllContextPods(context, &wg, outChan, errChan)
	}

	// Close the out chan one all the thread are finished.
	go func(outChan chan Pod) {
		wg.Wait()
		close(outChan)
	}(outChan)

	select {
	case err, isError := <-errChan:
		if isError {
			fmt.Println(err)
			os.Exit(1)
		}
	default:
	}

	res := make([]Pod, 0, len(contexts))
	for pod := range outChan {
		res = append(res, pod)
	}

	return res
}

func retrieveAllContextPods(context string, wg *sync.WaitGroup, outChan chan Pod, errChan chan error) {
	defer wg.Done()

	cmd := exec.Command("kubectl", "--context="+context, "get", "pod", "-o=jsonpath='{.items..metadata.name}'")
	rawOut, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("err1 : %s\n", err)
		errChan <- err
		return
	}

	// Need some parsing
	stringOut := strings.Trim(string(rawOut), "'")
	pods := strings.Split(stringOut, " ")

	for _, podName := range pods {
		outChan <- Pod{
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
