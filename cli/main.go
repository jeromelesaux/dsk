package main

import (
	"flag"
	"fmt"
	"github.com/jeromelesaux/dsk"
	"os"
)

var (
	list    = flag.Bool("list", false, "List content of dsk.")
	track   = flag.Int("track", 39, "Track number (format).")
	sector  = flag.Int("sector", 9, "Sector number (format)")
	format  = flag.Bool("format", false, "Format the followed dsk.")
	dskPath = flag.String("dsk", "", "Dsk path to handle.")
	file    = flag.String("file", "", "File in th dsk")
	hexa    = flag.Bool("hex", false, "List the file in hexadecimal")
)

func main() {
	var dskFile dsk.DSK
	flag.Parse()
	if *dskPath == "" {
		fmt.Fprintf(os.Stdout, "No dsk set.\n")
		flag.PrintDefaults()
		os.Exit(-1)
	}
	if *format {
		_, err := os.Stat(*dskPath)
		if err == nil {
			fmt.Fprintf(os.Stderr, "Error file (%s) already exists\n", *dskPath)
			os.Exit(-1)
		}

		f, err := os.Create(*dskPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while write file (%s) error %v\n", *dskPath, err)
			os.Exit(-1)
		}
		defer f.Close()
		fmt.Fprintf(os.Stdout, "Formating number of sectors (%d), tracks (%d)\n", *sector, *track)
		dskFile := dsk.FormatDsk(uint8(*sector), uint8(*track))
		if err := dskFile.Write(f); err != nil {
			fmt.Fprintf(os.Stderr, "Error while write file (%s) error %v\n", *dskPath, err)
			os.Exit(-1)
		}
	}
	f, err := os.Open(*dskPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while read file (%s) error %v\n", *dskPath, err)
		os.Exit(-1)
	}
	defer f.Close()
	if err := dskFile.Read(f); err != nil {
		fmt.Fprintf(os.Stderr, "Error while read dsk file (%s) error %v\n", *dskPath, err)
		os.Exit(-1)
	}

	if err := dskFile.CheckDsk(); err != nil {
		fmt.Fprintf(os.Stderr, "Error while read dsk file (%s) error %v\n", *dskPath, err)
		os.Exit(-1)
	}
	fmt.Fprintf(os.Stdout, "Dsk file (%s)\n", *dskPath)
	if *list {
		if err := dskFile.GetCatalogue(); err != nil {
			fmt.Fprintf(os.Stderr, "Error while getting catalogue in dsk file (%s) error %v\n", *dskPath, err)
			os.Exit(-1)
		}
		var index int
		var lenght uint8
		for index < 64 {
			entry := dskFile.Catalogue[index]
			lenght += entry.NbPages

			if entry.User != dsk.USER_DELETED && entry.NumPage != 0 {
				for {
					index++
					if dskFile.Catalogue[index].NbPages == 0 || index >= 64 {
						break
					}
					next := dskFile.Catalogue[index]
					if next.User == entry.User {
						lenght += next.NbPages
					} else {
						break
					}
				}
				size := fmt.Sprintf("%d ko", (int(lenght)+7)>>3)
				filename := fmt.Sprintf("%s.%s", entry.Nom, entry.Ext)
				fmt.Fprintf(os.Stdout, "[%.2d] : %s : %s\n", index, filename, size)
			}
			index++
		}
	}

	if *hexa {
		if *file == "" {
			fmt.Fprintf(os.Stderr, "File option is empty, set it.")
			os.Exit(-1)
		}
		amsdosFile := dsk.GetNomDir(*file)
		indice := dskFile.FileExists(amsdosFile)
		if indice == -1 {
			fmt.Fprintf(os.Stderr, "File %s does not exist\n", *file)
		} else {
			content, err := dskFile.GetFileIn(*file, indice)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error while getting file in dsk error :%v\n", err)
			}
			fmt.Println(dsk.DisplayHex(content))
		}
	}
	os.Exit(0)
}
