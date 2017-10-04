// cjdnsui - Graphical user interface for Cjdns
// Copyright (C) 2017  William Wennerström
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package gui

import (
	"errors"
	"os"

	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/widgets"
	"github.com/willeponken/cjdnsui/patterns"
)

const (
	statusTopic = iota
)

type statusData struct {
	cjdnsIp   string
	publicKey string
	port      int
}

//go:generate qtmoc
type statusWidget struct {
	widgets.QWidget

	_ func(data statusData) error `slot:"statusUpdate"`
}

func newStatusWidget() *statusWidget {
	widget := NewStatusWidget(nil, 0)
	mainHBox := widgets.NewQHBoxLayout2(widget)

	localInfoGroup := widgets.NewQGroupBox2("Peering information", nil)
	mainHBox.AddWidget(localInfoGroup, 0, core.Qt__AlignLeft)

	localInfoGrid := widgets.NewQGridLayout2()
	localInfoGroup.SetLayout(localInfoGrid)

	localInfoGrid.AddWidget(widgets.NewQLabel2("Cjdns IP:", nil, 0), 0, 0, core.Qt__AlignLeft)
	localInfoGrid.AddWidget(widgets.NewQLabel2("Public Key:", nil, 0), 1, 0, core.Qt__AlignLeft)
	localInfoGrid.AddWidget(widgets.NewQLabel2("Port:", nil, 0), 2, 0, core.Qt__AlignLeft)

	cjdnsIpLabel := widgets.NewQLabel2("Unknown", nil, 0)
	publicKeyLabel := widgets.NewQLabel2("Unknown", nil, 0)
	portLabel := widgets.NewQLabel2("Unknown", nil, 0)

	localInfoGrid.AddWidget(cjdnsIpLabel, 0, 1, core.Qt__AlignLeft)
	localInfoGrid.AddWidget(publicKeyLabel, 1, 1, core.Qt__AlignLeft)
	localInfoGrid.AddWidget(portLabel, 2, 1, core.Qt__AlignLeft)

	widget.ConnectStatusUpdate(func(data statusData) error {
		cjdnsIpLabel.SetText(data.cjdnsIp)
		publicKeyLabel.SetText(data.publicKey)
		portLabel.SetText(string(data.port))

		return nil
	})

	return widget
}

type View struct {
	patterns.Observable
}

func (view *View) Run() {
	widgets.NewQApplication(len(os.Args), os.Args)

	tabWidget := widgets.NewQTabWidget(nil)

	statusWidget := newStatusWidget()
	view.AddObserver(statusTopic, func(d interface{}) error {
		data, ok := d.(statusData)
		if !ok {
			return errors.New("cannot cast data to status data type")
		}
		return statusWidget.StatusUpdate(data)
	})

	tabWidget.AddTab(statusWidget, "Status")

	tabWidget.Show()

	widgets.QApplication_Exec()
}

func NewView() *View {
	view := &View{}

	return view
}
