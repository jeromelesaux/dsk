package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/jeromelesaux/dsk"
)

var (
	help           = flag.Bool("help", false, "display extended help.")
	list           = flag.Bool("list", false, "List content of dsk.")
	track          = flag.Int("track", 39, "Track number (format).")
	heads          = flag.Int("head", 1, "Number of heads in the DSK (format)")
	sector         = flag.Int("sector", 9, "Sector number (format).")
	format         = flag.Bool("format", false, "Format the followed dsk.")
	dskType        = flag.Int("dsktype", 0, "DSK Type :\n\t0 : DSK\n\t1 : EDSK\n\t3 : SNA\n")
	dskPath        = flag.String("dsk", "", "Dsk path to handle.")
	fileInDsk      = flag.String("amsdosfile", "", "File to handle in (or to insert in) the dsk.")
	hexa           = flag.Bool("hex", false, "List the amsdosfile in hexadecimal.")
	info           = flag.Bool("info", false, "Get informations of the amsdosfile (size, execute and loading address). Or get sna informations.")
	ascii          = flag.Bool("ascii", false, "list the amsdosfile in ascii mode.")
	desassemble    = flag.Bool("desassemble", false, "list the amsdosfile desassembled.")
	get            = flag.Bool("get", false, "Get the file in the dsk.")
	remove         = flag.Bool("remove", false, "Remove the amsdosfile from the current dsk.")
	basic          = flag.Bool("basic", false, "List a basic amsdosfile.")
	put            = flag.Bool("put", false, "Put the amsdosfile in the current dsk.")
	executeAddress = flag.String("exec", "", "Execute address of the inserted file. (hexadecimal #170 allowed.)")
	loadingAddress = flag.String("load", "", "Loading address of the inserted file. (hexadecimal #170 allowed.)")
	user           = flag.Int("user", 0, "User number of the inserted file.")
	force          = flag.Bool("force", false, "Force overwriting of the inserted file.")
	fileType       = flag.String("type", "", "Type of the inserted file \n\tascii : type ascii\n\tbinary : type binary\n")
	snaPath        = flag.String("sna", "", "SNA file to handle")
	analyse        = flag.Bool("analyze", false, "Returns the DSK header")
	cpcType        = flag.Int("cpctype", 2, "CPC type (sna import feature): \n\tCPC464 : 0\n\tCPC664: 1\n\tCPC6128 : 2\n\tUnknown : 3\n\tCPCPlus6128 : 4\n\tCPCPlus464 : 5\n\tGX4000 : 6\n\t")
	screenMode     = flag.Int("screenmode", 1, "screenmode of the importing file in sna.")
	version        = "0.7rc"
)

func main() {
	var dskFile dsk.DSK
	var execAddress, loadAddress uint16
	flag.Parse()

	if *help {
		sampleUsage()
		os.Exit(0)
	}

	if *executeAddress != "" {
		address := *executeAddress
		switch address[0] {
		case '#':
			value := strings.Replace(address, "#", "", -1)
			v, err := strconv.ParseUint(value, 16, 16)
			if err != nil {
				fmt.Fprintf(os.Stderr, "cannot get the hexadecimal value fom %s, error : %v\n", *executeAddress, err)
			} else {
				execAddress = uint16(v)
			}
		case '0':
			value := strings.Replace(address, "0x", "", -1)
			v, err := strconv.ParseUint(value, 16, 16)
			if err != nil {
				fmt.Fprintf(os.Stderr, "cannot get the hexadecimal value fom %s, error : %v\n", *executeAddress, err)
			} else {
				execAddress = uint16(v)
			}
		default:
			v, err := strconv.ParseUint(address, 10, 16)
			if err != nil {
				fmt.Fprintf(os.Stderr, "cannot get the hexadecimal value fom %s, error : %v\n", *executeAddress, err)
			} else {
				execAddress = uint16(v)
			}
		}
	}
	if *loadingAddress != "" {
		address := *loadingAddress
		switch address[0] {
		case '#':
			value := strings.Replace(address, "#", "", -1)
			v, err := strconv.ParseUint(value, 16, 16)
			if err != nil {
				fmt.Fprintf(os.Stderr, "cannot get the hexadecimal value fom %s, error : %v\n", *loadingAddress, err)
			} else {
				loadAddress = uint16(v)
			}
		case '0':
			value := strings.Replace(address, "0x", "", -1)
			v, err := strconv.ParseUint(value, 16, 16)
			if err != nil {
				fmt.Fprintf(os.Stderr, "cannot get the hexadecimal value fom %s, error : %v\n", *loadingAddress, err)
			} else {
				loadAddress = uint16(v)
			}
		default:
			v, err := strconv.ParseUint(address, 10, 16)
			if err != nil {
				fmt.Fprintf(os.Stderr, "cannot get the hexadecimal value fom %s, error : %v\n", *loadingAddress, err)
			} else {
				loadAddress = uint16(v)
			}
		}
	}

	fmt.Fprintf(os.Stdout, "DSK cli version [%s]\nMade by Sid (ImpAct)\n", version)
	if *snaPath != "" {
		if *info {
			f, err := os.Open(*snaPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error while read sna file (%s) error %v\n", *snaPath, err)
				os.Exit(-1)
			}
			defer f.Close()
			sna := &dsk.SNA{}
			if err := sna.Read(f); err != nil {
				fmt.Fprintf(os.Stderr, "Error while reading sna file (%s) error %v\n", *snaPath, err)
				os.Exit(-1)
			}
			fmt.Fprintf(os.Stdout, "Sna (%s) description :\n\tCPC type:%s\n\tCRTC type:%s\n", *snaPath, sna.CPCType(), sna.CRTCType())
			fmt.Fprintf(os.Stdout, "\tSna version:%d\n\tMemory size:%dKo\n", sna.Header.Version, sna.Header.MemoryDumpSize)
			fmt.Fprintf(os.Stdout, "%s\n", sna.Header.String())
			os.Exit(0)
		}
		if *format {
			if _, err := dsk.CreateSna(*snaPath); err != nil {
				fmt.Fprintf(os.Stderr, "Cannot create Sna file (%s) error : %v\n", *snaPath, err)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stdout, "Sna file (%s) created.\n", *snaPath)
			os.Exit(0)
		}
		if *put {
			if *fileInDsk != "" {
				cpcTYPE := dsk.CPCType(*cpcType)
				crtc := dsk.UM6845R
				if *cpcType > 3 {
					crtc = dsk.ASIC_6845
				}
				if err := dsk.ImportInSna(*fileInDsk, *snaPath, execAddress, uint8(*screenMode), cpcTYPE, crtc); err != nil {
					fmt.Fprintf(os.Stderr, "Error while trying to import file (%s) in new sna (%s) error: %v\n",
						*fileInDsk,
						*snaPath,
						err)
					os.Exit(1)
				}
				os.Exit(0)
			} else {
				fmt.Fprintf(os.Stderr, "Missing input file to import in sna file (%s)\n", *snaPath)
			}
		}
	}
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
		fmt.Fprintf(os.Stdout, "Formating number of sectors (%d), tracks (%d), head number (%d)\n", *sector, *track, *heads)
		dskFile := dsk.FormatDsk(uint8(*sector), uint8(*track), uint8(*heads), (*dskType))
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

	if *analyse {
		entry := dskFile.Entry
		fmt.Fprintf(os.Stdout, "Dsk entry %s\n", entry.ToString())
	}

	if *info {
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
			isAmsdos, header := dsk.CheckAmsdos(content)
			if !isAmsdos {
				fmt.Fprintf(os.Stderr, "File (%s) does not contain amsdos header.\n", *fileInDsk)
				os.Exit(-1)
			}
			fmt.Fprintf(os.Stdout, "Amsdos informations :\n\tFilename:%s\n\tSize:#%X (%.2f Ko)\n\tSize2:#%X (%.2f Ko)\n\tLogical Size:#%X (%.2f Ko)\n\tExecute Address:#%X\n\tLoading Address:#%X\n\tChecksum:#%X\n\tType:%d\n\tUser:%d\n",
				header.Filename,
				header.Size,
				float64(header.Size)/1024.,
				header.Size2,
				float64(header.Size2)/1024.,
				header.LogicalSize,
				float64(header.LogicalSize)/1024.,
				header.Exec,
				header.Address,
				header.Checksum,
				header.Type,
				header.User)
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
			content, fileSize, err := dskFile.ViewFile(indice)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error while getting file in dsk error :%v\n", err)
			}
			fmt.Println(dsk.DisplayHex(content[0:fileSize], 16))
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
			content, filesize, err := dskFile.ViewFile(indice)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error while getting file in dsk error :%v\n", err)
			}
			fmt.Println(dsk.Desass(content[0:filesize], uint16(filesize)))
		}
	}

	if *ascii {
		if *fileInDsk == "" {
			fmt.Fprintf(os.Stderr, "amsdosfile option is empty, set it.")
			os.Exit(-1)
		}
		amsdosFile := dsk.GetNomDir(*fileInDsk)
		indice := dskFile.FileExists(amsdosFile)
		if indice == dsk.NOT_FOUND {
			fmt.Fprintf(os.Stderr, "File %s does not exist\n", *fileInDsk)
		} else {
			content, filesize, err := dskFile.ViewFile(indice)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error while getting file in dsk error :%v\n", err)
			}
			fmt.Println(string(content[0:filesize]))
		}
	}

	if *basic {
		if *fileInDsk == "" {
			fmt.Fprintf(os.Stderr, "amsdosfile option is empty, set it.")
			os.Exit(-1)
		}
		amsdosFile := dsk.GetNomDir(*fileInDsk)
		indice := dskFile.FileExists(amsdosFile)
		if indice == dsk.NOT_FOUND {
			fmt.Fprintf(os.Stderr, "File %s does not exist\n", *fileInDsk)
		} else {
			content, filesize, err := dskFile.ViewFile(indice)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error while getting file in dsk error :%v\n", err)
			}
			fmt.Fprintf(os.Stdout, "File %s filesize :%d octets\n", *fileInDsk, filesize)
			fmt.Fprintf(os.Stdout, "%s", dsk.Basic(content, uint16(filesize), true))
		}
	}

	if *get {
		if *fileInDsk == "" {
			fmt.Fprintf(os.Stderr, "amsdosfile option is empty, set it.")
			os.Exit(-1)
		}
		if *fileInDsk == "*" {
			dskFile.GetCatalogue()
			var lastFilename string
			for indice, v := range dskFile.Catalogue {
				if v.User != dsk.USER_DELETED && v.NbPages != 0 {
					filename := fmt.Sprintf("%s.%s", v.Nom, v.Ext)
					if lastFilename == filename {
						continue
					}
					lastFilename = filename
					fmt.Fprintf(os.Stdout, "Filename to get : %s\n", filename)
					content, err := dskFile.GetFileIn(filename, indice)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error while getting file in dsk error :%v\n", err)
					}
					af, err := os.Create(filename)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error while creating file (%s) error %v\n", filename, err)
						os.Exit(-1)
					}
					defer af.Close()
					_, err = af.Write(content)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error while copying content in file (%s) error %v\n", filename, err)
						os.Exit(-1)
					}
				}
			}
		} else {
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
				informations := fmt.Sprintf("Extract file [%s]\nIndice in DSK [%d]\n", *fileInDsk, indice)
				resumeAction(*dskPath, "get amsdosfile", *fileInDsk, informations)
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
		if indice != dsk.NOT_FOUND && !*force {
			fmt.Fprintf(os.Stderr, "File %s already exists\n", *fileInDsk)
		} else {
			switch *fileType {
			case "ascii":
				informations := fmt.Sprintf("execute address [#%.4x], loading address [#%.4x]\n", execAddress, loadAddress)
				if err := dskFile.PutFile(*fileInDsk, dsk.MODE_ASCII, 0, 0, uint16(*user), false, false); err != nil {
					fmt.Fprintf(os.Stderr, "Error while inserted file (%s) in dsk (%s) error :%v\n", *fileInDsk, *dskPath, err)
					os.Exit(-1)
				}
				resumeAction(*dskPath, "put ascii", *fileInDsk, informations)
			case "binary":
				informations := fmt.Sprintf("execute address [#%.4x], loading address [#%.4x]\n", execAddress, loadAddress)
				if err := dskFile.PutFile(*fileInDsk, dsk.MODE_BINAIRE, loadAddress, execAddress, uint16(*user), false, false); err != nil {
					fmt.Fprintf(os.Stderr, "Error while inserted file (%s) in dsk (%s) error :%v\n", *fileInDsk, *dskPath, err)
					os.Exit(-1)
				}
				resumeAction(*dskPath, "put binary", *fileInDsk, informations)
			default:
				fmt.Fprintf(os.Stderr, "File type option unknown please choose between ascii or binary.")
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
		if indice == dsk.NOT_FOUND {
			fmt.Fprintf(os.Stderr, "File (%s) not found in dsk (%s)\n", *fileInDsk, *dskPath)
			os.Exit(-1)
		}
		if err := dskFile.RemoveFile(uint8(indice)); err != nil {
			fmt.Fprintf(os.Stderr, "Error while removing file %s (indice:%d) error :%v\n", *fileInDsk, indice, err)
		} else {
			fmt.Fprintf(os.Stderr, "File (%.8s.%.3s) deleted in dsk (%s)\n",
				amsdosFile.Nom,
				amsdosFile.Ext,
				*dskPath)
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

	os.Exit(0)
}

func resumeAction(dskFilepath, action, amsdosfile, informations string) {
	fmt.Fprintf(os.Stderr, "DSK path [%s]\n", dskFilepath)
	fmt.Fprintf(os.Stderr, "Action on DSK [%s] on amsdos file [%s]\n", action, amsdosfile)
	fmt.Fprintf(os.Stderr, "%s\n", informations)
}

func sampleUsage() {
	flag.PrintDefaults()
	fmt.Fprintf(os.Stdout, "\nHere sample usages :\n"+
		"\t* Create empty simple dsk file : dsk -dsk output.dsk -format\n"+
		"\t* Create empty simple dsk file with custom tracks and sectors: dsk -dsk output.dsk -format -sector 8 -track 42\n"+
		"\t* Create empty extended dsk file with custom head, tracks and sectors: dsk -dsk output.dsk -format -sector 8 -track 42 -dsktype 1 -head 2\n"+
		"\t* Create empty sna file : dsk -sna output.sna\n"+
		"\t* List dsk content : dsk -dsk output.dsk -list\n"+
		"\t* Get information on Sna file : dsk -sna output.sna -info\n"+
		"\t* Get information on file in dsk  : dsk -dsk output.dsk -amsdosfile hello.bin -info\n"+
		"\t* List file content in hexadecimal in dsk file : dsk -dsk output.dsk -amsdosfile hello.bin -hex\n"+
		"\t* Put file in dsk file : dsk -dsk output.dsk -put -amsdosfile hello.bin -exec #1000 -load 500\n"+
		"\t* Put file in sna file (here for a cpc plus): dsk -sna output.sna -put -amsdosfile hello.bin -exec #1000 -load 500 -screenmode 0 -cpctype 4\n")

}
