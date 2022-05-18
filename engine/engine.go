package engine

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	_ "unsafe"

	"github.com/anacrolix/torrent"
	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/quantumsheep/nyaa-cli/utils"
	range_parser "github.com/quantumsheep/range-parser"
)

var httpRegex = regexp.MustCompile(`^https?:\/\/`)

type Engine struct {
	DataDirectory string

	client  *torrent.Client
	torrent *torrent.Torrent

	server *http.Server

	runStatusLoop bool
}

func NewEngine(dataDirectory string) (*Engine, error) {
	var err error

	torrentConfig := torrent.NewDefaultClientConfig()
	torrentConfig.DataDir = dataDirectory
	torrentConfig.NoUpload = true
	torrentConfig.DisableTCP = false
	torrentConfig.ListenPort = 0
	// torrentConfig.IPBlocklist = blocklist

	engine := &Engine{
		DataDirectory: dataDirectory,
	}

	engine.client, err = torrent.NewClient(torrentConfig)
	if err != nil {
		return nil, err
	}

	return engine, nil
}

func (e *Engine) RunServer(port string, index int) error {
	handler := http.NewServeMux()
	e.server = &http.Server{
		Addr:    "localhost:" + port,
		Handler: handler,
	}

	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		headerWriter := w.Header()

		if r.Method == "OPTIONS" {
			accessControlRequestHeaders := r.Header.Get("Access-Control-Request-Headers")

			if accessControlRequestHeaders != "" {
				headerWriter.Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
				headerWriter.Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
				headerWriter.Set("Access-Control-Allow-Headers", accessControlRequestHeaders)
				headerWriter.Set("Access-Control-Max-Age", "1728000")

				w.WriteHeader(http.StatusOK)
				return
			}
		}

		if r.Header.Get("Origin") != "" {
			headerWriter.Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		}

		url := r.URL
		if url.Path == "/" || url.Path == "" {
			url.Path = fmt.Sprintf("/%d", index)
		}

		// Ignore favicon requests
		if url.Path == "/favicon.ico" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		host := r.Header.Get("Host")
		if host == "" {
			host = "localhost"
		}

		// if url.Path == "/.json" {
		// 	json := toJSON()
		// 	// response.setHeader('Content-Type', 'application/json; charset=utf-8')
		// 	// response.setHeader('Content-Length', Buffer.byteLength(json))
		// 	// response.end(json)
		// 	headerWriter.Set("Content-Type", "application/json; charset=utf-8")
		// 	headerWriter.Set("Content-Length", fmt.Sprintf("%d", len(json)))
		// 	w.Write(json)
		// 	return
		// }

		files := e.torrent.Files()

		if url.Path == "/.m3u" {
			playlist := "#EXTM3U\n"
			for i, file := range files {
				playlist += fmt.Sprintf("#EXTINF:-1,%s\nhttp://%s:%s/%d\n", file.Path(), host, port, i)
			}

			headerWriter.Set("Content-Type", "application/x-mpegurl; charset=utf-8")
			headerWriter.Set("Content-Length", fmt.Sprintf("%d", len(playlist)))
			_, err := io.WriteString(w, playlist)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		name := url.Path[1:]

		// If name is a file name, change it to the file's index
		for i, file := range files {
			if name == file.Torrent().Name() {
				name = fmt.Sprintf("/%d", i)
				break
			}
		}

		i, err := strconv.Atoi(name)
		if err != nil || i < 0 || i >= len(files) {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		file := files[i]

		var rang *range_parser.Range
		if rangeHeader := r.Header.Get("Range"); rangeHeader != "" {
			ranges, err := range_parser.Parse(file.Length(), rangeHeader)
			if err != nil {
				w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
				return
			}

			rang = ranges[0]
		}

		headerWriter.Set("Accept-Ranges", "bytes")
		headerWriter.Set("Content-Type", mime.TypeByExtension(filepath.Ext(file.Path())))
		headerWriter.Set("transferMode.dlna.org", "Streaming")
		headerWriter.Set("contentFeatures.dlna.org", "DLNA.ORG_OP=01;DLNA.ORG_CI=0;DLNA.ORG_FLAGS=01700000000000000000000000000000")

		if rang == nil {
			headerWriter.Set("Content-Length", strconv.FormatInt(file.Length(), 10))
			// if (request.method === 'HEAD') return response.end()
			// pump(file.createReadStream(), response)

			if r.Method == "HEAD" {
				w.WriteHeader(http.StatusOK)
				return
			}

			reader := file.NewReader()
			defer reader.Close()

			_, _ = io.Copy(w, reader)
			return
		}

		// response.statusCode = 206
		// response.setHeader('Content-Length', range.end - range.start + 1)
		// response.setHeader('Content-Range', 'bytes ' + range.start + '-' + range.end + '/' + file.length)
		// if (request.method === 'HEAD') return response.end()
		// pump(file.createReadStream(range), response)

		headerWriter.Set("Content-Length", strconv.FormatInt(rang.End-rang.Start+1, 10))
		headerWriter.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", rang.Start, rang.End, file.Length()))

		w.WriteHeader(http.StatusPartialContent)

		if r.Method == "HEAD" {
			return
		}

		reader := file.NewReader()
		defer reader.Close()
		_, err = reader.Seek(rang.Start, io.SeekStart)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, _ = io.CopyN(w, reader, rang.End-rang.Start+1)
	})

	if err := e.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (e *Engine) StopServer() error {
	e.DropCurrentTorrent()
	return e.server.Close()
}

//go:linkname downloadRate github.com/anacrolix/torrent.(*Peer).downloadRate
func downloadRate(self *torrent.Peer) float64

func (e *Engine) RunStatusLoop(index int) {
	e.runStatusLoop = true

	filename := e.GetFileName(index)

	max := e.torrent.Length()
	previousDownloadSpeed := 0.0
	previousDownloadSize := 0.0

	for e.runStatusLoop {
		current := e.torrent.BytesCompleted()

		fmt.Printf("\033[2J")
		fmt.Printf("\033[H")
		fmt.Printf("%s: %s / %s\n", color.CyanString("%s", filename), humanize.Bytes(uint64(current)), humanize.Bytes(uint64(max)))
		fmt.Printf("Peers: %d / %d\n", e.torrent.Stats().ActivePeers, e.torrent.Stats().TotalPeers)

		downloadSpeed := float64(current) - previousDownloadSize
		fmt.Printf("Download speed: %s/s\n\n", humanize.Bytes(uint64((downloadSpeed+previousDownloadSpeed)/2)))

		previousDownloadSpeed = downloadSpeed
		previousDownloadSize = float64(current)

		peers := e.torrent.PeerConns()
		sort.Slice(peers, func(i, j int) bool {
			return downloadRate(&peers[i].Peer) > downloadRate(&peers[j].Peer)
		})

		for i, peer := range peers {
			if i >= 10 {
				fmt.Printf("...\n")
				break
			}

			downloadRate := humanize.Bytes(uint64(downloadRate(&peer.Peer)))
			fmt.Printf("%s (%s)\n", color.MagentaString("%s", peer.RemoteAddr), color.GreenString("%s/s", downloadRate))
		}

		time.Sleep(time.Second)
	}
}

func (e *Engine) StopStatusLoop() {
	e.runStatusLoop = false
}

func (e *Engine) SetTorrentFromPath(torrentPath string) error {
	if strings.HasPrefix(torrentPath, "magnet:") {
		t, err := e.client.AddMagnet(torrentPath)
		if err != nil {
			return err
		}

		e.torrent = t
		return nil
	}

	if httpRegex.MatchString(torrentPath) {
		var err error

		f, err := os.CreateTemp("", "nyaa-cli")
		if err != nil {
			return err
		}
		defer f.Close()
		defer os.Remove(f.Name())

		torrentPath, err = utils.Download(torrentPath, f.Name())
		if err != nil {
			return err
		}
	}

	t, err := e.client.AddTorrentFromFile(torrentPath)
	if err != nil {
		return err
	}

	e.torrent = t
	return nil
}

func (e *Engine) DropCurrentTorrent() {
	e.torrent.Drop()
}

func (e *Engine) GetFileName(i int) string {
	if i == -1 {
		return e.torrent.Name()
	}

	files := e.torrent.Files()
	return files[i].DisplayPath()
}
