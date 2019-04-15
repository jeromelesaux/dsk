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
	get     = flag.Bool("get", false, "Get the file in the dsk.")
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

		for _, i := range dskFile.GetFilesIndices() {
			size := fmt.Sprintf("%.3d ko", dskFile.GetFilesize(dskFile.Catalogue[i]))
			filename := fmt.Sprintf("%s.%s", dskFile.Catalogue[i].Nom, dskFile.Catalogue[i].Ext)
			fmt.Fprintf(os.Stdout, "[%.2d] : %s : %d %s\n", i, filename, int(dskFile.Catalogue[i].User), size)
		}
	}

	if *hexa {
		if *file == "" {
			fmt.Fprintf(os.Stderr, "File option is empty, set it.")
			os.Exit(-1)
		}
		amsdosFile := dsk.GetNomDir(*file)
		indice := dskFile.FileExists(amsdosFile)
		if indice == dsk.NOT_FOUND {
			fmt.Fprintf(os.Stderr, "File %s does not exist\n", *file)
		} else {
			content, err := dskFile.GetFileIn(*file, indice)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error while getting file in dsk error :%v\n", err)
			}
			fmt.Println(dsk.DisplayHex(content, 16))
		}
	}
	if *get {
		if *file == "" {
			fmt.Fprintf(os.Stderr, "File option is empty, set it.")
			os.Exit(-1)
		}
		amsdosFile := dsk.GetNomDir(*file)
		indice := dskFile.FileExists(amsdosFile)
		if indice == dsk.NOT_FOUND {
			fmt.Fprintf(os.Stderr, "File %s does not exist\n", *file)
		} else {
			content, err := dskFile.GetFileIn(*file, indice)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error while getting file in dsk error :%v\n", err)
			}
			af, err := os.Create(*file)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error while creating file (%s) error %v\n", *file, err)
				os.Exit(-1)
			}
			defer af.Close()
			_, err = af.Write(content)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error while copying content in file (%s) error %v\n", *file, err)
				os.Exit(-1)
			}
		}
	}
	os.Exit(0)
}
