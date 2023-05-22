/// 2>/dev/null ; gorun "$0" "$@" ; exit $?

package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/briandowns/spinner"
	lop "github.com/samber/lo/parallel"
	"github.com/urfave/cli/v2"
)

var BuildVersion string = "1.0.0"

var typeMime = map[string]string{
	"mp3":  "audio/mpeg",
	"flac": "audio/flac",
	"mp4":  "video/mp4",
	"mkv":  "video/x-matroska",
	"jpg":  "image/jpeg",
	"jpeg": "image/jpeg",
	"png":  "image/png",
	"gif":  "image/gif",
	"svg":  "image/svg+xml",
	"pdf":  "application/pdf",
	"epub": "application/epub+zip",
	"cbr":  "application/x-cbr",
	"cbz":  "application/x-cbz",
	"3ds":  "application/x-3ds",
	"cia":  "application/x-cia",
	"gb":   "application/x-gameboy-rom",
	"gba":  "application/x-gba-rom",
	"gbc":  "application/x-gbc-rom",
	"iso":  "application/x-iso9660-image",
	"nds":  "application/x-nintendo-ds-rom",
	"nes":  "application/x-nes-rom",
	"rvz":  "application/x-rvz",
	"sfc":  "application/x-sfc-rom",
	"smc":  "application/x-snes-rom",
	"wux":  "application/x-wux",
	"xci":  "application/x-xci-rom",
	"z64":  "application/x-n64-rom",
}

func backup(src string, dist string) {
	fileInfo, err := os.Stat(src)
	if err != nil {
		log.Fatal(err)
	}

	if fileInfo.IsDir() {
		files, err := GetDirFiles(src)
		if err != nil {
			log.Fatal(err)
		}

		lop.ForEach(files, func(file string, _ int) {
			fileData, err := GetFileData(file)
			if err != nil {
				log.Fatal(err)
			}

			uploadPath := fmt.Sprintf("%s/%s", dist, fileData.name)
			archiveUrl, uploadtob2Err := UploadToB2(fileData.data, uploadPath, typeMime[fileData.extension], false)
			if uploadtob2Err != nil {
				log.Fatal(uploadtob2Err)
			}

			log.Printf("[Public URL]: %s", archiveUrl)
		})
	} else {
		fileData, err := GetFileData(src)
		if err != nil {
			log.Fatal(err)
		}

		path := fmt.Sprintf("%s/%s", dist, fileData.name)
		archiveUrl, uploadtob2Err := UploadToB2(fileData.data, path, typeMime[fileData.extension], false)
		if uploadtob2Err != nil {
			log.Fatal(uploadtob2Err)
		}

		log.Printf("[Public URL]: %s", archiveUrl)
	}
}

func main() {
	var src string
	var dist string
	loader := spinner.New(spinner.CharSets[21], 100*time.Millisecond)

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
			loader.Start()
			backup(src, dist)
			loader.Stop()

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
