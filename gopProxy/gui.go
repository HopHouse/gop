package gopproxy

import (
	"log"

	"github.com/hophouse/gop/utils"
	"github.com/hophouse/gop/utils/logger"
	"github.com/jroimartin/gocui"
)

var G *gocui.Gui

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	horizontalSep := int(maxX/2) - 20
	horizontalSepHalf := int(horizontalSep / 2)

	if v, err := g.SetView("host", 0, 0, horizontalSepHalf-1, 2); err != nil {
		if err != gocui.ErrUnknownView {
			logger.Println(err)
			return err
		}
		v.Title = "Host"
		v.Wrap = true
	}

	if v, err := g.SetView("url", horizontalSepHalf+1, 0, horizontalSep-1, 2); err != nil {
		if err != gocui.ErrUnknownView {
			logger.Println(err)
			return err
		}
		v.Title = "URL"
		v.Wrap = true
	}

	if v, err := g.SetView("request", 0, 3, horizontalSep-1, maxY-4); err != nil {
		if err != gocui.ErrUnknownView {
			logger.Println(err)
			return err
		}
		v.Title = "Request"
		v.Wrap = true
	}

	if v, err := g.SetView("response-header", horizontalSep+1, 0, maxX-1, maxY/2); err != nil {
		if err != gocui.ErrUnknownView {
			logger.Println(err)
			return err
		}
		v.Title = "Response Header"
		v.Wrap = true
	}

	if v, err := g.SetView("response-body", horizontalSep+1, maxY/2+1, maxX-1, maxY-4); err != nil {
		if err != gocui.ErrUnknownView {
			logger.Println(err)
			return err
		}
		v.Title = "Response Body"
		// v.Autoscroll = true
		v.Wrap = true
	}

	if v, err := g.SetView("mode", 0, maxY-3, 18, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			logger.Println(err)
			return err
		}
		v.Title = "Mode"
		// v.Autoscroll = true
		v.Wrap = true
	}

	if v, err := g.SetView("cmd", 19, maxY-3, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			logger.Println(err)
			return err
		}
		v.Title = "Commands"
		logger.Fprintf(v, " Ctrl+n: Next view | Ctrl+i: Toggle interception | Ctrl+Space: Forward | Ctrl+c: Exit")
		// v.Autoscroll = true
		v.Wrap = true
	}

	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

// func RunGUI(server *http.Server) error {
func RunGUI() error {
	var err error
	G, err = gocui.NewGui(gocui.OutputNormal)
	utils.CheckError(err)
	defer G.Close()

	G.Mouse = true
	G.Cursor = true
	G.Highlight = true

	G.SetManagerFunc(layout)

	// Quit
	if err := G.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		logger.Println(err)
		// server.Close()
		return err
	}

	if err := initKeybindings(G); err != nil {
		logger.Fatal(err)
	}

	if err := G.MainLoop(); err != nil && err != gocui.ErrQuit {
		logger.Fatal(err)
	}

	return nil
}

func ClearAllGUIViews() {
	views := []string{
		"host",
		"url",
		"request",
		"response-header",
		"response-body",
	}

	for _, view := range views {
		v, err := G.View(view)
		if err != nil {
			logger.Println(err)
			break
		}
		v.Clear()
	}
}

func ClearGUIView(g *gocui.Gui, view string) *gocui.View {
	v, err := g.View(view)
	if err != nil {
		logger.Println(err)
		return nil
	}
	v.Clear()
	return v
}

// Get from:
// https://github.com/jroimartin/gocui/blob/master/_examples/stdin.go#L100-L109
func initKeybindings(g *gocui.Gui) error {
	if err := initKeybindingsGeneral(g); err != nil {
		return err
	}

	return nil
}

func initKeybindingsGeneral(g *gocui.Gui) error {
	// Click on a view
	if err := g.SetKeybinding("", gocui.MouseLeft, gocui.ModNone, selectViewOnClick); err != nil {
		logger.Println(err)
		return err
	}

	// Key Up
	if err := g.SetKeybinding("", gocui.KeyArrowUp, gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error {
			err := scrollView(v, -1)
			if err != nil {
				log.Println("Error in scollview during key up")
				return err
			}
			return nil
		}); err != nil {
		return err
	}

	// Key Down
	if err := g.SetKeybinding("", gocui.KeyArrowDown, gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error {
			err := scrollView(v, 1)
			if err != nil {
				log.Println("Error in scollview during key up")
				return err
			}
			return nil
		}); err != nil {
		return err
	}

	// Change view
	if err := g.SetKeybinding("", gocui.KeyCtrlN, gocui.ModNone, selectNextView); err != nil {
		logger.Println(err)
		return err
	}

	// Toggle intercept mode
	if err := g.SetKeybinding("", gocui.KeyCtrlI, gocui.ModNone, toggleInterceptorMode); err != nil {
		logger.Println(err)
		return err
	}

	// Forward request
	if err := g.SetKeybinding("", gocui.KeyCtrlSpace, gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error {
			if len(InterceptChan) < 1 {
				InterceptChan <- true
			}
			return nil
		}); err != nil {
		return err
	}

	// Autoscroll
	if err := g.SetKeybinding("response-body", 'a', gocui.ModNone, autoscroll); err != nil {
		logger.Println(err)
		return err
	}

	return nil
}

func autoscroll(g *gocui.Gui, v *gocui.View) error {
	v.Autoscroll = true
	return nil
}

func scrollView(v *gocui.View, dy int) error {
	logger.Println("Scroll view")
	if v != nil {
		v.Autoscroll = false
		ox, oy := v.Origin()
		if err := v.SetOrigin(ox, oy+dy); err != nil {
			return err
		}
	}
	return nil
}

func selectViewOnClick(g *gocui.Gui, v *gocui.View) error {
	logger.Println("Select View on click")
	if v != nil {
		if _, err := g.SetCurrentView(v.Name()); err != nil {
			return err
		}
	}
	err := v.SetCursor(0, 0)
	if err != nil {
		return err
	}
	return nil
}

func selectNextView(g *gocui.Gui, v *gocui.View) error {
	logger.Println("Select Next View")

	// remove url view from selection
	views := []string{
		"request",
		"response-header",
		"response-body",
	}

	// If no view is selected
	if g.CurrentView() == nil {
		if _, err := g.SetCurrentView(views[0]); err != nil {
			logger.Println(err)
			return err
		}
		return nil
	}

	// Get position of the current view
	var newPosition int
	for i, view := range views {
		// Get next one ie, current +1 mod len(views)
		if view == v.Name() {
			newPosition = (i + 1) % len(views)
			break
		}
	}

	// Set cursor to new view
	if _, err := g.SetCurrentView(views[newPosition]); err != nil {
		logger.Println(err)
		return err
	}
	return nil
}

func toggleInterceptorMode(g *gocui.Gui, _ *gocui.View) error {
	logger.Printf("Toggle Interceptor mode from %v to %v", InterceptMode, !InterceptMode)
	InterceptMode = !InterceptMode
	var message string

	if InterceptMode {
		message = "Intercept"
	}

	v := ClearGUIView(g, "mode")
	logger.Fprintf(v, "%s", message)

	return nil
}
