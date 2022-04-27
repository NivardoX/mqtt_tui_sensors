package mqtt

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"math/rand"
	"sort"
	"strconv"
	"time"
)

type sensorData struct {
	Id       string `json:"id"`
	LastRead string `json:"last_read"`
}

var (
	app   *tview.Application // The tview application.
	pages *tview.Pages       // The application pages.

	form      *tview.Form     // The initial form
	list      *tview.Flex     // The flex containing the listing
	listText  *tview.TextView // The text containing current sensor info
	listTable *tview.Table    // The table containing other sensors info

	maxValue     int
	minValue     int
	currentValue int
	id           string
	selectedType string

	sensorsData map[string]sensorData
	ticker      *time.Ticker
	mqttHandler *Handler
)

func Start() {
	app = tview.NewApplication()

	pages = tview.NewPages()
	form = getInitialForm()
	framedForm := getFrame(form)
	list = getListing()
	framedList := getFrame(list)

	pages.AddPage("INITIAL_FORM", framedForm, true, true)
	pages.AddPage("LISTING", framedList, true, false)
	if err := app.SetRoot(pages, true).Run(); err != nil {
		panic(err)
	}
}

func getFrame(primitive tview.Primitive) *tview.Frame {
	frame := tview.NewFrame(primitive).
		AddText("MQTT Sensors", true, tview.AlignLeft, tcell.ColorWhite).
		AddText("IFCE - PPD", true, tview.AlignLeft, tcell.ColorRed).
		AddText("By: Nivardox", false, tview.AlignLeft, tcell.ColorGreen)
	return frame
}
func getInitialForm() *tview.Form {
	form := tview.NewForm().
		AddDropDown("Sensor", []string{"Temperature", "Humidity", "Velocity"}, 0, func(option string, optionIndex int) {
			selectedType = option

		}).
		AddInputField("ID", "", 20, nil, func(text string) {
			id = text
		}).
		AddInputField("Min. Value", "", 10, func(textToCheck string, lastChar rune) bool {
			value, err := strconv.Atoi(textToCheck)
			return err == nil && value >= 0
		}, func(text string) {
			minValue, _ = strconv.Atoi(text)
		}).
		AddInputField("Max. Value", "", 10, func(textToCheck string, lastChar rune) bool {
			_, err := strconv.Atoi(textToCheck)
			return err == nil
		}, func(text string) {
			maxValue, _ = strconv.Atoi(text)
		}).
		AddButton("Next", func() {
			setListingText()
			startTicker()
			mqttHandler = NewMqttHandler(selectedType, func(data sensorData) {
				sensorsData[data.Id] = data
				app.QueueUpdateDraw(setListingTableData)
			})
			mqttHandler.sub()

			pages.SwitchToPage("LISTING")

		}).
		AddButton("Quit", func() {
			app.Stop()
		})

	form.SetBorder(true).SetTitle("Enter you sensor data").SetTitleAlign(tview.AlignLeft)
	return form
}
func setListingText() {
	text := fmt.Sprintf("%s (%d - %d)\nCurrent Read: %d", id, minValue, maxValue, currentValue)
	listText.SetText(text)
}
func getListing() *tview.Flex {
	listText = tview.NewTextView()
	listTable = tview.NewTable().SetBorders(true)
	sensorsData = make(map[string]sensorData)

	setListingTableData()
	var flex = tview.NewFlex()

	flex.SetDirection(tview.FlexRow).
		AddItem(listText, 0, 1, true).
		AddItem(listTable, 0, 3, true)

	return flex
}
func setListingTableData() {
	listTable.SetBorder(true).SetTitle("Other Sensors")
	listTable.SetCell(0, 0, tview.NewTableCell("Sensor ID").SetAlign(tview.AlignCenter).SetTextColor(tcell.ColorYellow))
	listTable.SetCell(0, 1, tview.NewTableCell("Last Read").SetAlign(tview.AlignCenter).SetTextColor(tcell.ColorYellow))
	listTable.SetCell(0, 2, tview.NewTableCell("Under Alert").SetAlign(tview.AlignCenter).SetTextColor(tcell.ColorYellow))
	i := 1

	keys := make([]string, 0, len(sensorsData))
	for k := range sensorsData {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		data := sensorsData[k]
		if data.Id != id {
			lastReadAsInt, _ := strconv.Atoi(data.LastRead)
			alertStr := ""
			if lastReadAsInt > maxValue || lastReadAsInt < minValue {
				alertStr = "X"
			}
			listTable.SetCell(i, 0, tview.NewTableCell(data.Id).SetAlign(tview.AlignCenter))
			listTable.SetCell(i, 1, tview.NewTableCell(data.LastRead).SetAlign(tview.AlignCenter))
			listTable.SetCell(i, 2, tview.NewTableCell(alertStr).SetAlign(tview.AlignCenter))
			i += 1
		}

	}

}
func startTicker() {
	ticker = time.NewTicker(2000 * time.Millisecond)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				currentValue = rand.Intn((maxValue+10)-minValue) + (minValue - 10)
				mqttHandler.pub(sensorData{
					Id:       id,
					LastRead: strconv.Itoa(currentValue),
				})
				app.QueueUpdateDraw(setListingText)
			}
		}
	}()

}
