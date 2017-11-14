package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xinerama"
	"github.com/BurntSushi/xgbutil/xrect"
	"github.com/BurntSushi/xgbutil/xwindow"

	"github.com/tw4452852/twwm/prompt"
)

var (
	findItems sync.Map

	selectActivate = "Control-Return"

	findWidget struct {
		x       *xgbutil.XUtil
		geom    xrect.Rect
		slct    *prompt.Select
		message *prompt.Message
	}
)

type findItem struct {
	group, name string
}

func (item *findItem) String() string {
	return fmt.Sprintf("[%s:%s]", item.group, item.name)
}

func (item *findItem) SelectText() string {
	return item.name
}

func (item *findItem) SelectHighlighted(_ interface{}) {}

func (item *findItem) SelectSelected(_ interface{}) {
	input := findWidget.slct.Input
	withPath := filepath.Join(item.group, item.name)
	cmd := strings.Fields(string(input.Text))
	for i, p := range cmd {
		if strings.HasPrefix(strings.ToLower(item.name), strings.ToLower(p)) {
			cmd[i] = withPath
			break
		}
	}
	cmdline := strings.Join(cmd, " ")

	cmdRun(cmdline)
}

var showLock sync.Mutex

func showResult(cmd string, stdoutStderr []byte, err error) {
	showLock.Lock()
	defer showLock.Unlock()

	buf := &bytes.Buffer{}
	fmt.Fprintf(buf, "cmdline: %q\n", cmd)
	if err != nil {
		fmt.Fprintf(buf, "error: %s\n", err)
	}
	fmt.Fprintf(buf, "stdout&stderr:\n%s\n", string(stdoutStderr))

	findWidget.message.Hide()
	findWidget.message.Show(findWidget.geom, buf.String(), 0, nil)
}

func doFind() bool {
	var showGroups = make(map[*prompt.SelectGroupItem][]*prompt.SelectItem)
	findItems.Range(func(k, v interface{}) bool {
		groupName := k.(string)
		groupItems := v.(sync.Map)
		group := findWidget.slct.AddGroup(findWidget.slct.NewStaticGroup(groupName))
		items := showGroups[group]

		groupItems.Range(func(k, v interface{}) bool {
			itemName := k.(string)
			item := findWidget.slct.AddChoice(
				&findItem{
					group: groupName,
					name:  itemName,
				})
			items = append(items, item)
			return true
		})

		showGroups[group] = items
		return true
	})

	keybind.KeyPressFun(func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
		findWidget.slct.Show(findWidget.geom, prompt.TabCompleteCustom, newShowGroups(showGroups), nil)
	}).Connect(findWidget.x, findWidget.x.RootWin(), selectActivate, true)
	xevent.Main(findWidget.x)
	return true

}

func headGeom(X *xgbutil.XUtil) (xrect.Rect, error) {
	if X.ExtInitialized("XINERAMA") {
		heads, err := xinerama.PhysicalHeads(X)
		if err == nil {
			return heads[0], nil
		}
	}
	return xwindow.New(X, X.RootWin()).Geometry()
}

func newShowGroups(m map[*prompt.SelectGroupItem][]*prompt.SelectItem) []*prompt.SelectShowGroup {
	showGroups := make([]*prompt.SelectShowGroup, 0)

	for k, v := range m {
		showGroups = append(showGroups, k.ShowGroup(v))
	}

	return showGroups
}

func initFind() error {
	populateFindItems()
	err := createWidget()
	if err != nil {
		return err
	}
	doFind()
	return nil
}

func createWidget() error {
	X, err := xgbutil.NewConn()
	if err != nil {
		return err
	}
	keybind.Initialize(X)

	geom, err := headGeom(X)
	if err != nil {
		return err
	}

	selectConfig := prompt.DefaultSelectConfig
	selectConfig.CompleteFn = cmdComplete
	selectConfig.ConfirmFn = cmdRun
	slct := prompt.NewSelect(X,
		prompt.DefaultSelectTheme, selectConfig)

	message := prompt.NewMessage(X,
		prompt.DefaultMessageTheme, prompt.DefaultMessageConfig)

	findWidget.x = X
	findWidget.geom = geom
	findWidget.slct = slct
	findWidget.message = message
	return nil
}

func cmdComplete(input, item string) bool {
	cmd := strings.Fields(strings.ToLower(input))
	if len(cmd) == 0 {
		return false
	}
	input = cmd[len(cmd)-1]
	item = strings.ToLower(item)
	return strings.HasPrefix(item, input)
}

func cmdRun(cmdline string) {
	if cmdline == "" {
		return
	}
	go func() {
		cmd := exec.Command("sh", "-c", cmdline)
		stdoutStderr, err := cmd.CombinedOutput()
		showResult(cmdline, stdoutStderr, err)
	}()
}

func populateFindItems() {
	for _, path := range strings.Split(os.Getenv("PATH"), ":") {
		var bins sync.Map
		for _, bin := range readFiles(path) {
			bins.Store(bin, struct{}{})
		}
		findItems.Store(path, bins)
	}
}

func readFiles(path string) []string {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	fis, err := f.Readdir(-1)
	if err != nil {
		return nil
	}
	var names []string
	for _, fi := range fis {
		if !fi.IsDir() {
			names = append(names, fi.Name())
		}
	}
	return names
}
