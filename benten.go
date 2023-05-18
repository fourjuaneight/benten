/// 2>/dev/null ; gorun "$0" "$@" ; exit $?

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

var BuildVersion string = "1.0.0"

type FileData struct {
	data      []byte
	name      string
	extension string
}

var typeMime = map[string]string{
	"mp3":  "audio/mpeg",
	"flac": "audio/flac",
	"mp4":  "video/mp4",
	"mkv":  "video/x-matroska",
}

func getFileData(src string) (FileData, error) {
	data, err := os.ReadFile(src)
	if err != nil {
		return FileData{}, fmt.Errorf("[readFile][ioutil.ReadFile]: %w", err)
	}

	name := filepath.Base(src)
	extension := filepath.Ext(src)
	extension = extension[1:]

	return FileData{
		data:      data,
		name:      name,
		extension: extension,
	}, nil
}

func backup(src string, dist string) {
	fileInfo, err := os.Stat(src)
	if err != nil {
		log.Fatal(err)
	}

	if fileInfo.IsDir() {
		filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				log.Fatal(err)
			}

			if info.IsDir() {
				return nil
			}

			fileData, err := getFileData(path)
			if err != nil {
				log.Fatal(err)
			}

			path = fmt.Sprintf("%s/%s", dist, fileData.name)
			archiveUrl, uploadtob2Err := UploadToB2(fileData.data, path, typeMime[fileData.extension])
			if uploadtob2Err != nil {
				log.Fatal(uploadtob2Err)
			}

			log.Printf("[Public URL]: %s", archiveUrl)

			return nil
		})
	} else {

		fileData, err := getFileData(src)
		if err != nil {
			log.Fatal(err)
		}

		path := fmt.Sprintf("%s/%s", dist, fileData.name)
		archiveUrl, uploadtob2Err := UploadToB2(fileData.data, path, typeMime[fileData.extension])
		if uploadtob2Err != nil {
			log.Fatal(uploadtob2Err)
		}

		log.Printf("[Public URL]: %s", archiveUrl)
	}
}

func main() {
	var src string
	var dist string

	// versioning
	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"v"},
		Usage:   "print app version",
	}

	// help
	cli.AppHelpTemplate = `NAME:
	{{.Name}} - {{.Usage}}

VERSION:
	{{.Version}}

USAGE:
	{{.HelpName}} [optional options]

OPTIONS:
{{range .VisibleFlags}}	{{.}}{{ "\n" }}{{end}}	
	`
	cli.HelpFlag = &cli.BoolFlag{
		Name:    "help",
		Aliases: []string{"h"},
		Usage:   "show help",
	}

	// execute app
	app := &cli.App{
		Name:    "benten",
		Usage:   "Save media to B2 via rsync",
		Version: BuildVersion,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "src",
				Aliases:     []string{"s"},
				Usage:       "source folder name",
				Destination: &src,
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "dist",
				Aliases:     []string{"d"},
				Usage:       "destination folder name",
				Destination: &dist,
				Required:    true,
			},
		},
		Action: func(*cli.Context) error {
			backup(src, dist)
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
