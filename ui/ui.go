package ui

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
	_ "unsafe"

	"github.com/gdamore/tcell/v2"
	"github.com/gocolly/colly"
	"github.com/quantumsheep/go-nyaa/v2/nyaa"
	"github.com/quantumsheep/go-nyaa/v2/types"
	"github.com/quantumsheep/nyaa-cli/engine"
	"github.com/quantumsheep/nyaa-cli/utils"
	"github.com/rivo/tview"
)

//go:linkname colorPattern github.com/rivo/tview.colorPattern
var colorPattern *regexp.Regexp

func init() {
	// Shady patch to disable color pattern matching in tview
	colorPattern = regexp.MustCompile(`$^`)
}

var sortOptions = []string{
	"Date",
	"Downloads",
	"Size",
	"Seeders",
	"Leechers",
	"Comments",
}

var orderOptions = []string{
	"Desc",
	"Asc",
}

type Torrent struct {
	types.Torrent

	hasExpanded bool
	fileCount   int
}

type UI struct {
	options *UIOptions

	app   *tview.Application
	pages *tview.Pages

	query   string
	sortBy  int
	orderBy int

	searchForm *tview.Form
	table      *tview.Table
	shortcuts  *tview.Table

	torrents map[string]*Torrent

	engine            *engine.Engine
	tempDataDirectory string
}

type UIOptions struct {
	VideoPlayer     string
	Fullscreen      bool
	OutputDirectory string
}

func NewUI(options *UIOptions) *UI {
	ui := &UI{
		options: options,
		app:     tview.NewApplication(),
		query:   "",
		sortBy:  0,
		orderBy: 0,
		pages:   tview.NewPages(),
	}

	ui.app.
		SetRoot(ui.pages, true).
		EnableMouse(true)

	ui.GenerateTable()
	ui.GenerateSearchForm()
	ui.GenerateShortcuts()

	flex := tview.NewFlex().
		AddItem(ui.searchForm, 5, 0, true).SetDirection(tview.FlexRow).
		AddItem(ui.table, 0, 6, true).SetDirection(tview.FlexRow).
		AddItem(ui.shortcuts, 1, 0, false).SetDirection(tview.FlexRow)

	ui.pages.AddPage("search", flex, true, true)

	err := ui.Search(nyaa.SearchOptions{
		Provider: "nyaa",
		Query:    ui.query,
		Category: "anime-eng",
		SortBy:   strings.ToLower(sortOptions[ui.sortBy]),
		OrderBy:  strings.ToLower(orderOptions[ui.orderBy]),
		Filter:   "no-filter",
	})
	if err != nil {
		log.Fatal(err)
	}

	return ui
}

func (ui *UI) Run() error {
	return ui.app.Run()
}

func (ui *UI) GenerateSearchForm() {
	ui.searchForm = tview.NewForm().
		SetHorizontal(true)

	ui.searchForm.
		SetBorder(true).
		SetBackgroundColor(tcell.ColorReset)

	ui.searchForm.SetCancelFunc(func() {
		ui.app.SetFocus(ui.table)
	})

	ui.searchForm.
		AddInputField("Query", "", 24, nil, func(text string) {
			ui.query = text
		}).
		AddDropDown("Sort By", sortOptions, ui.sortBy, func(option string, optionIndex int) {
			ui.sortBy = optionIndex
		}).
		AddDropDown("Order By", orderOptions, ui.sortBy, func(option string, optionIndex int) {
			ui.orderBy = optionIndex
		}).
		AddButton("Search", func() {
			err := ui.Search(nyaa.SearchOptions{
				Provider: "nyaa",
				Query:    ui.query,
				Category: "anime-eng",
				SortBy:   strings.ToLower(sortOptions[ui.sortBy]),
				OrderBy:  strings.ToLower(orderOptions[ui.orderBy]),
				Filter:   "no-filter",
			})
			if err != nil {
				ui.Fatal(err)
			}

			ui.app.SetFocus(ui.table)
		})
}

func (ui *UI) Search(opts nyaa.SearchOptions) error {
	torrents, err := nyaa.Search(opts)
	if err != nil {
		return err
	}

	ui.torrents = make(map[string]*Torrent)
	ui.table.Clear()

	for i, torrent := range torrents {
		link := strings.Split(torrent.Link, "download/")
		id := strings.TrimSuffix(link[1], ".torrent")

		t, _ := time.Parse("Mon, 02 Jan 2006 15:04:05 -0700", torrent.Date)
		date := t.Format("2006-01-02 15:04")

		trusted := ""
		if torrent.IsTrusted == "Yes" {
			trusted = "✓"
		}

		ui.table.SetCell(i, 0, ui.GenerateCell(id, 8, 0, tcell.ColorBlue))
		ui.table.SetCell(i, 1, ui.GenerateCell("○", 4, 0, tcell.ColorWhite))
		ui.table.SetCell(i, 2, ui.GenerateCell(torrent.Size, 10, 0, tcell.ColorYellow))
		ui.table.SetCell(i, 3, ui.GenerateCell(date, 17, 0, tcell.ColorGray))
		ui.table.SetCell(i, 4, ui.GenerateCell(torrent.Seeders, 6, 0, tcell.ColorGreen))
		ui.table.SetCell(i, 5, ui.GenerateCell(torrent.Leechers, 6, 0, tcell.ColorRed))
		ui.table.SetCell(i, 6, ui.GenerateCell(trusted, 2, 0, tcell.ColorWhite))
		ui.table.SetCell(i, 7, ui.GenerateCell(torrent.Name, 0, 0, tcell.ColorWhite).SetAlign(tview.AlignLeft).SetExpansion(1))

		ui.torrents[id] = &Torrent{
			Torrent: torrent,
		}
	}

	ui.table.Select(0, 0)
	ui.table.ScrollToBeginning()

	return nil
}

func (ui *UI) GenerateCell(value string, leftPadding int, rightPadding int, color tcell.Color) *tview.TableCell {
	if leftPadding > 0 && len(value) < leftPadding {
		value = strings.Repeat(" ", leftPadding-len(value)) + value
	}

	if rightPadding > 0 && len(value) < rightPadding {
		value = value + strings.Repeat(" ", rightPadding-len(value))
	}

	return tview.NewTableCell(value).
		SetTextColor(color).
		SetAlign(tview.AlignRight)
}

func (ui *UI) GenerateTable() {
	ui.table = tview.NewTable()
	ui.table.
		SetSelectable(true, false).
		SetBorder(true).
		SetBackgroundColor(tcell.ColorReset)

	ui.table.SetDoneFunc(func(key tcell.Key) {
		ui.app.SetFocus(ui.searchForm)
	})

	ui.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		row, _ := ui.table.GetSelection()

		if event.Key() == tcell.KeyF2 {
			id, _ := ui.GetTorrentId(row)
			torrent := ui.torrents[id]

			directory, err := filepath.Abs(ui.options.OutputDirectory)
			if err != nil {
				ui.Fatal(err)
			}

			if err := os.MkdirAll(directory, os.ModePerm); err != nil {
				ui.Fatal(err)
			}

			_, err = utils.Download(torrent.Link, filepath.Join(directory, torrent.Name+".torrent"))
			if err != nil {
				ui.Fatal(err)
			}

			ui.table.SetCell(row, 1, ui.GenerateCell("●", 4, 0, tcell.ColorGreen).SetAlign(tview.AlignRight))

			return nil
		}

		return event
	})

	ui.table.SetSelectedFunc(func(row int, column int) {
		id, index := ui.GetTorrentId(row)
		torrent := ui.torrents[id]

		if !torrent.hasExpanded {
			files, err := nyaaTorrentFiles(torrent.GUID)
			if err != nil {
				ui.Fatal(err)
			}

			torrent.fileCount = len(files)

			if len(files) > 1 {
				for i, file := range files {
					newRow := row + 1 + i

					ui.table.InsertRow(newRow)
					ui.table.SetCell(newRow, 0, ui.GenerateCell("", 8, 0, tcell.ColorWhite).SetAlign(tview.AlignLeft))
					ui.table.SetCell(newRow, 1, ui.GenerateCell("│", 4, 0, tcell.ColorWhite))
					ui.table.SetCell(newRow, 2, ui.GenerateCell(file.size, 10, 0, tcell.ColorYellow))
					ui.table.SetCell(newRow, 3, ui.GenerateCell("", 17, 0, tcell.ColorYellow))
					ui.table.SetCell(newRow, 4, ui.GenerateCell("", 6, 0, tcell.ColorYellow))
					ui.table.SetCell(newRow, 5, ui.GenerateCell("", 6, 0, tcell.ColorYellow))
					ui.table.SetCell(newRow, 6, ui.GenerateCell("", 2, 0, tcell.ColorYellow))
					ui.table.SetCell(newRow, 7, ui.GenerateCell(file.name, 0, 0, tcell.ColorDimGray).SetAlign(tview.AlignLeft).SetExpansion(1))
				}

				torrent.hasExpanded = true
				return
			} else {
				torrent.hasExpanded = true
			}
		}

		if torrent.fileCount > 1 && index == -1 {
			return
		}

		if ui.engine == nil {
			tempDir, err := os.MkdirTemp("", "nyaa-cli")
			if err != nil {
				ui.Fatal(err)
			}

			ui.engine, err = engine.NewEngine(tempDir)
			if err != nil {
				ui.Fatal(err)
			}
		}

		ui.app.Suspend(func() {
			if err := ui.engine.SetTorrentFromPath(torrent.Link); err != nil {
				ui.Fatal(err)
			}
			defer ui.engine.DropCurrentTorrent()

			port := "3001"
			url := fmt.Sprintf("http://localhost:%s", port)

			if index > -1 {
				url += fmt.Sprintf("/%d", index)
			}

			wg := sync.WaitGroup{}
			wg.Add(3)

			go func() {
				defer wg.Done()

				if err := ui.engine.RunServer(port, 0); err != nil {
					ui.Fatal(err)
				}
			}()

			go func() {
				defer wg.Done()
				ui.engine.RunStatusLoop(index)
			}()

			go func() {
				defer wg.Done()

				err := utils.RunVideoPlayer(utils.VideoPlayerConfig{
					VideoPlayer: ui.options.VideoPlayer,
					Url:         url,
					Name:        ui.engine.GetFileName(index),
					OnTop:       true,
					Fullscreen:  ui.options.Fullscreen,
				})
				if err != nil {
					ui.Fatal(err)
				}

				ui.engine.StopStatusLoop()

				err = ui.engine.StopServer()
				if err != nil {
					ui.Fatal(err)
				}
			}()

			wg.Wait()
		})
	})
}

func (ui *UI) GetTorrentId(row int) (string, int) {
	id := strings.TrimSpace(ui.table.GetCell(row, 0).Text)
	index := -1

	for id == "" {
		id = strings.TrimSpace(ui.table.GetCell(row-2-index, 0).Text)
		index++
	}

	return id, index
}

func (ui *UI) GenerateShortcuts() {
	ui.shortcuts = tview.NewTable().
		SetBorders(false).
		SetCell(0, 0, tview.NewTableCell("F1").
			SetTextColor(tcell.ColorWhite).
			SetAlign(tview.AlignCenter),
		).
		SetCell(0, 1, tview.NewTableCell("Download").
			SetTextColor(tcell.ColorBlack).
			SetBackgroundColor(tcell.ColorBlue).
			SetAlign(tview.AlignCenter),
		)
}

func (ui *UI) Fatal(err error) {
	ui.app.Stop()
	log.Fatal(err)
}

type torrentFile struct {
	name string
	size string
}

func nyaaTorrentFiles(viewURL string) ([]*torrentFile, error) {
	var files []*torrentFile

	c := colly.NewCollector()

	c.OnHTML(".torrent-file-list li", func(e *colly.HTMLElement) {
		if e.DOM.Has("ul").Length() == 0 {
			fileSize := e.ChildText(".file-size")

			files = append(files, &torrentFile{
				name: strings.TrimSuffix(e.Text, " "+fileSize),
				size: fileSize[1 : len(fileSize)-1], // remove "(" and ")"
			})
		}
	})

	var e error
	c.OnError(func(r *colly.Response, err error) {
		e = err
	})
	if e != nil {
		return nil, e
	}

	err := c.Visit(viewURL)
	if err != nil {
		return nil, err
	}

	return files, nil
}
