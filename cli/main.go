package main

import (
	"flag"
	"fmt"
	"github.com/jeromelesaux/dsk"
	"os"
)

var (
	list           = flag.Bool("list", false, "List content of dsk.")
	track          = flag.Int("track", 39, "Track number (format).")
	sector         = flag.Int("sector", 9, "Sector number (format).")
	format         = flag.Bool("format", false, "Format the followed dsk.")
	dskPath        = flag.String("dsk", "", "Dsk path to handle.")
	fileInDsk      = flag.String("amsdosfile", "", "File to handle in (or to insert in) the dsk.")
	hexa           = flag.Bool("hex", false, "List the amsdosfile in hexadecimal.")
	info           = flag.Bool("info", false, "Get informations of the amsdosfile (size, execute and loading address).")
	ascii          = flag.Bool("ascii", false, "list the amsdosfile in ascii mode.")
	desassemble    = flag.Bool("desassemble", false, "list the amsdosfile desassembled.")
	get            = flag.Bool("get", false, "Get the file in the dsk.")
	remove         = flag.Bool("remove", false, "Remove the amsdosfile from the current dsk.")
	basic          = flag.Bool("basic", false, "List a basic amsdosfile.")
	put            = flag.Bool("put", false, "Put the amsdosfile in the current dsk.")
	executeAddress = flag.Int("exec", -1, "Execute address of the inserted file.")
	loadingAddress = flag.Int("load", -1, "Loading address of the inserted file.")
	user           = flag.Int("user", 0, "User number of the inserted file.")
	force          = flag.Bool("force", false, "Force overwriting of the inserted file.")
	fileType       = flag.String("type", "", "Type of the inserted file \n\tascii : type ascii\n\tbinary : type binary\n")
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
		if *fileInDsk == "" {
			fmt.Fprintf(os.Stderr, "amsdosfile option is empty, set it.")
			os.Exit(-1)
		}
		amsdosFile := dsk.GetNomDir(*fileInDsk)
		indice := dskFile.FileExists(amsdosFile)
		if indice == dsk.NOT_FOUND {
			fmt.Fprintf(os.Stderr, "File %s does not exist\n", *fileInDsk)
		} else {
			content, err := dskFile.ViewFile(indice)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error while getting file in dsk error :%v\n", err)
			}
			fmt.Println(dsk.DisplayHex(content, 16))
		}
	}

	if *desassemble {
		if *fileInDsk == "" {
			fmt.Fprintf(os.Stderr, "amsdosfile option is empty, set it.")
			os.Exit(-1)
		}
		amsdosFile := dsk.GetNomDir(*fileInDsk)
		indice := dskFile.FileExists(amsdosFile)
		if indice == dsk.NOT_FOUND {
			fmt.Fprintf(os.Stderr, "File %s does not exist\n", *fileInDsk)
		} else {
			content, err := dskFile.ViewFile(indice)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error while getting file in dsk error :%v\n", err)
			}
			fmt.Println(dsk.Desass(content, uint16(len(content))))
		}
	}

	if *get {
		if *fileInDsk == "" {
			fmt.Fprintf(os.Stderr, "amsdosfile option is empty, set it.")
			os.Exit(-1)
		}
		amsdosFile := dsk.GetNomDir(*fileInDsk)
		indice := dskFile.FileExists(amsdosFile)
		if indice == dsk.NOT_FOUND {
			fmt.Fprintf(os.Stderr, "File %s does not exist\n", *fileInDsk)
		} else {
			content, err := dskFile.GetFileIn(*fileInDsk, indice)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error while getting file in dsk error :%v\n", err)
			}
			af, err := os.Create(*fileInDsk)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error while creating file (%s) error %v\n", *fileInDsk, err)
				os.Exit(-1)
			}
			defer af.Close()
			_, err = af.Write(content)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error while copying content in file (%s) error %v\n", *fileInDsk, err)
				os.Exit(-1)
			}
		}
	}
	if *put {
		if *fileInDsk == "" {
			fmt.Fprintf(os.Stderr, "amsdosfile option is empty, set it.")
			os.Exit(-1)
		}
		amsdosFile := dsk.GetNomDir(*fileInDsk)
		indice := dskFile.FileExists(amsdosFile)
		if indice != dsk.NOT_FOUND && ! *force{
			fmt.Fprintf(os.Stderr, "File %s already exists\n", *fileInDsk)
		} else {
			switch *fileType {
			case "ascii":
				informations := fmt.Sprintf("execute address [#%.4x], loading address [#%.4x]\n", *executeAddress, *loadingAddress)
				if err := dskFile.PutFile(*fileInDsk,dsk.MODE_ASCII,0,0,uint16(*user),false,false); err != nil {
					fmt.Fprintf(os.Stderr,"Error while inserted file (%s) in dsk (%s) error :%v\n",*fileInDsk,*dskPath,err)
					os.Exit(-1)
				}
				resumeAction(*dskPath, "put ascii", *fileInDsk, informations)
			case "binary":
				informations := fmt.Sprintf("execute address [#%.4x], loading address [#%.4x]\n", *executeAddress, *loadingAddress)
				if err := dskFile.PutFile(*fileInDsk,dsk.MODE_BINAIRE,uint16(*loadingAddress),uint16(*executeAddress),uint16(*user),false,false); err != nil {
					fmt.Fprintf(os.Stderr,"Error while inserted file (%s) in dsk (%s) error :%v\n",*fileInDsk,*dskPath,err)
					os.Exit(-1)
				}
				resumeAction(*dskPath, "put binary", *fileInDsk, informations)
			default: 
				fmt.Fprintf(os.Stderr,"File type option unknown please choose between ascii or binary.")
			}
			f, err := os.Create(*dskPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error while write file (%s) error %v\n", *dskPath, err)
				os.Exit(-1)
			}
			defer f.Close()
			
			if err := dskFile.Write(f); err != nil {
				fmt.Fprintf(os.Stderr, "Error while write file (%s) error %v\n", *dskPath, err)
				os.Exit(-1)
			}
		}
	}

	if *remove {
		if *fileInDsk == "" {
			fmt.Fprintf(os.Stderr, "amsdosfile option is empty, set it.")
			os.Exit(-1)
		}
		amsdosFile := dsk.GetNomDir(*fileInDsk)
		indice := dskFile.FileExists(amsdosFile)
		if indice != dsk.NOT_FOUND && ! *force{
			fmt.Fprintf(os.Stderr, "File %s already exists\n", *fileInDsk)
		} else {
			if err := dskFile.RemoveFile(uint8(indice)); err != nil {
				fmt.Fprintf(os.Stderr,"Error while removing file %s (indice:%d) error :%v\n",*fileInDsk,indice,err)
			} else {
				f, err := os.Create(*dskPath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error while write file (%s) error %v\n", *dskPath, err)
					os.Exit(-1)
				}
				defer f.Close()
				if err := dskFile.Write(f); err != nil {
					fmt.Fprintf(os.Stderr, "Error while write file (%s) error %v\n", *dskPath, err)
					os.Exit(-1)
				}
			}
		}
	
	}


	os.Exit(0)
}

func resumeAction(dskFilepath, action, amsdosfile, informations string) {
	fmt.Fprintf(os.Stderr, "DSK path [%s]\n", dskFilepath)
	fmt.Fprintf(os.Stderr, "Action on DSK [%s] on amsdos file [%s]\n", action, amsdosfile)
	fmt.Fprintf(os.Stderr, "%s\n", informations)
}
