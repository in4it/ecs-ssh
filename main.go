package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jroimartin/gocui"
)

var e = newECS()

func previousView(g *gocui.Gui, v *gocui.View) error {
	if v == nil || v.Name() == "services" {
		v.Clear()
		v.Highlight = false
		v2, err := g.SetCurrentView("clusters")
		v2.Highlight = true
		return err
	}
	if v == nil || v.Name() == "tasks" {
		v.Clear()
		v.Highlight = false
		v2, err := g.SetCurrentView("services")
		v2.Highlight = true
		return err
	}
	return nil
}

func cursorDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy+1); err != nil {
			ox, oy := v.Origin()
			if err := v.SetOrigin(ox, oy+1); err != nil {
				return err
			}
		}
	}
	return nil
}

func cursorUp(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		ox, oy := v.Origin()
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy-1); err != nil && oy > 0 {
			if err := v.SetOrigin(ox, oy-1); err != nil {
				return err
			}
		}
	}
	return nil
}

func showError(g *gocui.Gui, v *gocui.View, errToDisplay error) error {
	g.Update(func(g2 *gocui.Gui) error {
		maxX, maxY := g2.Size()
		if v, err := g2.SetView("error", maxX/2-30, maxY/2, maxX/2+30, maxY/2+2); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			if _, err := g2.SetCurrentView("error"); err != nil {
				return err
			}
			fmt.Fprintln(v, errToDisplay.Error())
		}
		return nil
	})
	return nil
}
func hideError(g *gocui.Gui, v *gocui.View) error {
	err := g.DeleteView("error")
	v, _ = g.SetCurrentView("clusters")
	v.Highlight = true
	return err
}
func getServices(g *gocui.Gui, v *gocui.View) error {
	var err error
	var clusterName string

	_, cy := v.Cursor()
	if clusterName, err = v.Line(cy); err != nil {
		return showError(g, v, err)
	}
	v.Highlight = false

	v2, err := g.SetCurrentView("services")
	fmt.Fprintln(v2, "Loading...")
	g.Update(func(g *gocui.Gui) error {
		v2, _ := g.View("services")
		services, err := e.getServices(clusterName)
		if err != nil {
			return showError(g, v, err)
		}
		v2.Clear()
		v2.Highlight = true

		if len(services) == 0 {
			fmt.Fprintln(v2, errNoServiceFound())
			err = getTasks(g, v2)
			if err != nil {
				return showError(g, v2, err)
			}
		} else {
			for _, s := range services {
				fmt.Fprintln(v2, s)
			}
		}
		return nil
	})

	return err
}
func errNoServiceFound() string {
	return "No Services Found"
}
func getTasks(g *gocui.Gui, v *gocui.View) error {
	var err error
	var serviceName string

	_, cy := v.Cursor()
	if serviceName, err = v.Line(cy); err != nil {
		return showError(g, v, err)
	}
	v.Highlight = false

	v2, err := g.SetCurrentView("tasks")
	fmt.Fprintln(v2, "Loading...")
	g.Update(func(g *gocui.Gui) error {
		v2, _ := g.View("tasks")
		var containers []string
		var err error
		if serviceName == errNoServiceFound() {
			containers, err = e.getAllTasks()
		} else {
			containers, err = e.getTasks(serviceName)
		}
		if err != nil {
			return showError(g, v2, err)
		}

		v2.Clear()
		for _, c := range containers {
			fmt.Fprintln(v2, c)
		}

		v2.Highlight = true

		return nil
	})

	return err
}

func doSSH(g *gocui.Gui, v *gocui.View) error {
	var err error
	var taskName string
	_, cy := v.Cursor()
	if taskName, err = v.Line(cy); err != nil {
		return showError(g, v, err)
	}
	_, err = e.getContainerInstanceIP(taskName)
	if err != nil {
		return showError(g, v, err)
	}

	// exit and start ssh
	g.Close()

	return gocui.ErrQuit
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func keybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}
	if err := g.SetKeybinding("clusters", gocui.KeyArrowDown, gocui.ModNone, cursorDown); err != nil {
		return err
	}
	if err := g.SetKeybinding("clusters", gocui.KeyArrowUp, gocui.ModNone, cursorUp); err != nil {
		return err
	}
	if err := g.SetKeybinding("clusters", gocui.KeyEnter, gocui.ModNone, getServices); err != nil {
		return err
	}
	if err := g.SetKeybinding("services", gocui.KeyArrowLeft, gocui.ModNone, previousView); err != nil {
		return err
	}
	if err := g.SetKeybinding("services", gocui.KeyArrowDown, gocui.ModNone, cursorDown); err != nil {
		return err
	}
	if err := g.SetKeybinding("services", gocui.KeyArrowUp, gocui.ModNone, cursorUp); err != nil {
		return err
	}
	if err := g.SetKeybinding("services", gocui.KeyEnter, gocui.ModNone, getTasks); err != nil {
		return err
	}
	if err := g.SetKeybinding("tasks", gocui.KeyEnter, gocui.ModNone, doSSH); err != nil {
		return err
	}
	if err := g.SetKeybinding("tasks", gocui.KeyArrowLeft, gocui.ModNone, previousView); err != nil {
		return err
	}
	if err := g.SetKeybinding("tasks", gocui.KeyArrowDown, gocui.ModNone, cursorDown); err != nil {
		return err
	}
	if err := g.SetKeybinding("tasks", gocui.KeyArrowUp, gocui.ModNone, cursorUp); err != nil {
		return err
	}
	if err := g.SetKeybinding("error", gocui.KeyEsc, gocui.ModNone, hideError); err != nil {
		return err
	}
	if err := g.SetKeybinding("error", gocui.KeyEnter, gocui.ModNone, hideError); err != nil {
		return err
	}

	return nil
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("clusters", 0, 0, maxX/3, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Highlight = true
		v.Frame = true
		v.Title = "Clusters"
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack

		for _, c := range e.clusterNames {
			fmt.Fprintln(v, c)
		}
		if _, err := g.SetCurrentView("clusters"); err != nil {
			return err
		}
	}
	if v, err := g.SetView("services", maxX/3, 0, maxX/3*2, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Highlight = false
		v.Frame = true
		v.Title = "Services"
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack

	}
	if v, err := g.SetView("tasks", maxX/3*2, 0, maxX, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Highlight = false
		v.Frame = true
		v.Title = "Tasks"
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack

	}
	return nil
}

// for debug purposes:

func main() {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.Cursor = true

	g.SetManagerFunc(layout)

	if err := keybindings(g); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil {
		if err != gocui.ErrQuit {
			log.Panicln(err)
		}
		if e.ipAddr != nil {
			err = startSSH(*e.ipAddr, *e.keyName)
			if err != nil {
				fmt.Printf("Error: %v\n\n", err)
			}
			os.Exit(0)
		}
	}
}
