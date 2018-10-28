package main

import (
	"bufio"
	"fmt"
	"log"
	"sync"

	"github.com/jroimartin/gocui"
)

// PrintOutput read from  the "Out" readCloser inside each pod and write the
// raw content into a ncurse dashboard.
func PrintOutput(pods []Pod) error {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		return err
	}
	defer g.Close()

	g.SetManagerFunc(func(g *gocui.Gui) error {
		return onUpdate(pods, g)
	})

	// the SetKeybinding need to be set after SetManagerFunc
	err = g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit)
	if err != nil {
		return err
	}
	err = g.SetKeybinding("", gocui.KeyCtrlD, gocui.ModNone, quit)
	if err != nil {
		return err
	}
	err = g.SetKeybinding("", 'q', gocui.ModNone, quit)
	if err != nil {
		return err
	}

	for _, pod := range pods {
		go func(pod Pod) {
			reader := bufio.NewReader(pod.Out)

			for {
				var line []byte

				line, _, err = reader.ReadLine()
				if err != nil {
					log.Println(err)
				}

				g.Update(func(g *gocui.Gui) error {
					var view *gocui.View

					view, err = g.View(pod.ID())
					if err != nil {
						return err
					}

					fmt.Fprintln(view, string(line))

					return err
				})

			}
		}(pod)
	}

	err = g.MainLoop()
	if err == gocui.ErrQuit {
		return nil
	}

	return err
}

// PrintRawOutput read from  the "Out" readCloser inside each pod and write the
// raw content into Stdout.
func PrintRawOutput(pods []Pod, lineCfg *LineConfig) {
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

func onUpdate(pods []Pod, g *gocui.Gui) error {
	maxX, maxY := g.Size()

	for idx, pod := range pods {
		view, err := g.SetView(pod.ID(), maxX/len(pods)*idx, 0, maxX/len(pods)*(idx+1), maxY)
		if err != nil && err != gocui.ErrUnknownView {
			return err
		}

		view.Title = fmt.Sprintf("  %s - %s  ", pod.Context, pod.Name)
		view.Autoscroll = true
		view.Wrap = true
		view.Frame = true
	}

	return nil
}

func quit(_ *gocui.Gui, _ *gocui.View) error {
	return gocui.ErrQuit
}
