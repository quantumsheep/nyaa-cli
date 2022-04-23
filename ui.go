package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	_ "unsafe"

	"github.com/gdamore/tcell/v2"
	"github.com/quantumsheep/go-nyaa/v2/nyaa"
	"github.com/quantumsheep/go-nyaa/v2/types"
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

type UI struct {
	options *UIOptions

	app   *tview.Application
	pages *tview.Pages

	query   string
	sortBy  int
	orderBy int

	searchForm *tview.Form
	table      *tview.Table
	torrents   []types.Torrent
}

type UIOptions struct {
	UsePeerflix        bool
	PeerflixFullscreen bool
	OutputDirectory    string
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

	flex := tview.NewFlex().
		AddItem(ui.searchForm, 5, 0, true).SetDirection(tview.FlexRow).
		AddItem(ui.table, 0, 6, true).SetDirection(tview.FlexRow)

	ui.pages.AddPage("search", flex, true, true)

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
			var err error

			opt := nyaa.SearchOptions{
				Provider: "nyaa",
				Query:    ui.query,
				Category: "anime-eng",
				SortBy:   strings.ToLower(sortOptions[ui.sortBy]),
				OrderBy:  strings.ToLower(orderOptions[ui.orderBy]),
				Filter:   "trusted-only",
			}

			ui.torrents, err = nyaa.Search(opt)
			if err != nil {
				log.Fatal(err)
			}

			ui.table.Clear()

			for i, torrent := range ui.torrents {
				t, _ := time.Parse("Mon, 02 Jan 2006 15:04:05 -0700", torrent.Date)
				date := t.Format("2006-01-02 15:04")

				ui.table.SetCell(i, 0, ui.GenerateCell("○", 4, 0, tcell.ColorWhite).SetAlign(tview.AlignRight))
				ui.table.SetCell(i, 1, ui.GenerateCell(torrent.Size, 10, 0, tcell.ColorYellow).SetAlign(tview.AlignRight))
				ui.table.SetCell(i, 2, ui.GenerateCell(date, 0, 0, tcell.ColorGray).SetAlign(tview.AlignRight))
				ui.table.SetCell(i, 3, ui.GenerateCell(torrent.Name, 0, 0, tcell.ColorWhite).SetExpansion(1))
				ui.table.SetCell(i, 4, ui.GenerateCell(torrent.Seeders, 6, 0, tcell.ColorGreen).SetAlign(tview.AlignRight))
				ui.table.SetCell(i, 5, ui.GenerateCell(torrent.Leechers, 6, 0, tcell.ColorRed).SetAlign(tview.AlignRight))
			}

			ui.app.SetFocus(ui.table)
			ui.table.Select(0, 0)
			ui.table.ScrollToBeginning()
		})
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
		SetAlign(tview.AlignLeft)
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

	ui.table.SetSelectedFunc(func(row int, column int) {
		if row < len(ui.torrents) {
			torrent := ui.torrents[row]

			if ui.options.UsePeerflix {
				ui.app.Suspend(func() {
					cmd := exec.Command("peerflix", torrent.Link, "--vlc")

					if ui.options.PeerflixFullscreen {
						cmd.Args = append(cmd.Args, "--", "--fullscreen")
					}

					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr
					err := cmd.Run()
					if err != nil {
						log.Fatal(err)
					}
				})
			} else {
				res, err := http.Get(torrent.Link)
				if err != nil {
					log.Fatal(err)
				}
				defer res.Body.Close()

				directory, err := filepath.Abs(ui.options.OutputDirectory)
				if err != nil {
					log.Fatal(err)
				}

				if err := os.MkdirAll(directory, os.ModePerm); err != nil {
					log.Fatal(err)
				}

				outputPath := filepath.Join(directory, torrent.Name+".torrent")
				file, err := os.Create(outputPath)
				if err != nil {
					log.Fatal(err)
				}
				defer file.Close()

				if _, err := io.Copy(file, res.Body); err != nil {
					log.Fatal(err)
				}

				ui.table.SetCell(row, 0, ui.GenerateCell("●", 4, 0, tcell.ColorGreen).SetAlign(tview.AlignRight))
			}
		}
	})
}
