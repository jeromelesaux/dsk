package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"io/ioutil"
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
	screenMode     = flag.Int("screenmode", 1, "screen mode parameter for the sna.")
	addHeader      = flag.Bool("addheader", false, "Add header to the standalone file (must be set with exec, load and type options).")
	version        = "0.11"
)

func main() {
	var cmdRunned bool = false
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

	fmt.Fprintf(os.Stderr, "DSK cli version [%s]\nMade by Sid (ImpAct)\n", version)
	if *snaPath != "" {
		if *info {
			cmdRunned = true
			f, err := os.Open(*snaPath)
			if err != nil {
				exitOnError(fmt.Sprintf("Error while read sna file (%s) error %v", *snaPath, err), "Check your sna file")
			}
			defer f.Close()
			sna := &dsk.SNA{}
			if err := sna.Read(f); err != nil {
				exitOnError(fmt.Sprintf("Error while reading sna file (%s) error %v", *snaPath, err), "Check your sna file")
			}
			fmt.Fprintf(os.Stderr, "Sna (%s) description :\n\tCPC type:%s\n\tCRTC type:%s\n", *snaPath, sna.CPCType(), sna.CRTCType())
			fmt.Fprintf(os.Stderr, "\tSna version:%d\n\tMemory size:%dKo\n", sna.Header.Version, sna.Header.MemoryDumpSize)
			fmt.Fprintf(os.Stderr, "%s\n", sna.Header.String())
			os.Exit(0)
		}
		if *format {
			cmdRunned = true
			if _, err := dsk.CreateSna(*snaPath); err != nil {
				fmt.Fprintf(os.Stderr, "Cannot create Sna file (%s) error : %v\n", *snaPath, err)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "Sna file (%s) created.\n", *snaPath)
			os.Exit(0)
		}
		if *put {
			cmdRunned = true
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
	if *dskPath == "" && *fileInDsk == "" {
		exitOnError("No dsk set.", "")
	}
	if *format {
		cmdRunned = true
		_, err := os.Stat(*dskPath)
		if err == nil {
			exitOnError(fmt.Sprintf("Error file (%s) already exists", *dskPath), "Use option -force to avoid this message")
		}

		f, err := os.Create(*dskPath)
		if err != nil {
			exitOnError(fmt.Sprintf("Error while write file (%s) error %v", *dskPath, err), "Check your dsk file path.")
		}
		defer f.Close()
		fmt.Fprintf(os.Stderr, "Formating number of sectors (%d), tracks (%d), head number (%d)\n", *sector, *track, *heads)
		dskFile := dsk.FormatDsk(uint8(*sector), uint8(*track), uint8(*heads), (*dskType))
		if err := dskFile.Write(f); err != nil {
			exitOnError(fmt.Sprintf("Error while write file (%s) error %v\n", *dskPath, err), "Check your dsk file with option -dsk yourdsk.dsk -analyze")
		}
	}

	if *dskPath != "" {
		f, err := os.Open(*dskPath)
		if err != nil {
			exitOnError(fmt.Sprintf("Error while read file (%s) error %v\n", *dskPath, err), "Check your dsk file path")
		}
		defer f.Close()
		if err := dskFile.Read(f); err != nil {
			exitOnError(fmt.Sprintf("Error while read dsk file (%s) error %v\n", *dskPath, err), "Check your dsk file path or Check your dsk file with option -dsk yourdsk.dsk -analyze")
		}

		if err := dskFile.CheckDsk(); err != nil {
			exitOnError(fmt.Sprintf("Error while read dsk file (%s) error %v\n", *dskPath, err), "Check your dsk file path or Check your dsk file with option -dsk yourdsk.dsk -analyze")
		}
		fmt.Fprintf(os.Stderr, "Dsk file (%s)\n", *dskPath)
		if *list {
			cmdRunned = true
			if err := dskFile.GetCatalogue(); err != nil {
				exitOnError(fmt.Sprintf("Error while getting catalogue in dsk file (%s) error %v\n", *dskPath, err), "Check your dsk file with option -dsk yourdsk.dsk -analyze")
			}
			totalUsed := 0
			for _, i := range dskFile.GetFilesIndices() {
				size := fmt.Sprintf("%.3d ko", dskFile.GetFilesize(dskFile.Catalogue[i]))
				totalUsed += dskFile.GetFilesize(dskFile.Catalogue[i])
				filename := fmt.Sprintf("%s.%s", dskFile.Catalogue[i].Nom, dskFile.Catalogue[i].Ext)
				fmt.Fprintf(os.Stderr, "[%.2d] : %s : %d %s\n", i, filename, int(dskFile.Catalogue[i].User), size)
			}
			fmt.Fprintf(os.Stderr, "Dsk %.3d Ko used\n", totalUsed)
		}

		if *analyse {
			cmdRunned = true
			entry := dskFile.Entry
			fmt.Fprintf(os.Stderr, "Dsk entry %s\n", entry.ToString())
		}

		if *info {
			cmdRunned = true
			if *fileInDsk == "" {
				exitOnError("amsdosfile option is empty, set it.", "usage sample : dsk -dsk output.dsk -amsdosfile hello.bin -info ")
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
					exitOnError(fmt.Sprintf("File (%s) does not contain amsdos header.\n", *fileInDsk), "add address of execution and loading like : dsk -dsk output.dsk -put -amsdosfile hello.bin -exec #1000 -load 500")
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
			cmdRunned = true
			if *fileInDsk == "" {
				exitOnError("amsdosfile option is empty, set it.", "dsk -dsk output.dsk -hex -amsdosfile hello.bin")
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
			cmdRunned = true
			if *fileInDsk == "" {
				exitOnError("amsdosfile option is empty, set it.", "dsk -dsk output.dsk -desassemble -amsdosfile hello.bin")
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
				var address uint16
				raw, err := dskFile.GetFileIn(*fileInDsk, indice)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error while getting file in dsk error :%v\n", err)
				} else {
					isAmsdos, header := dsk.CheckAmsdos(raw)
					if isAmsdos {
						address = header.Exec
					}
				}

				fmt.Println(dsk.Desass(content[0:filesize], uint16(filesize), address))
			}
		}

		if *ascii {
			cmdRunned = true
			if *fileInDsk == "" {
				exitOnError("amsdosfile option is empty, set it.", "dsk -dsk output.dsk -ascii -amsdosfile hello.txt")
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
			cmdRunned = true
			if *fileInDsk == "" {
				exitOnError("amsdosfile option is empty, set it.", "dsk -dsk output.dsk -basic -amsdosfile hello.bin")
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
				fmt.Fprintf(os.Stderr, "File %s filesize :%d octets\n", *fileInDsk, filesize)
				fmt.Fprintf(os.Stdout, "%s", dsk.Basic(content, uint16(filesize), true))
			}
		}

		if *get {
			cmdRunned = true
			if *fileInDsk == "" {
				exitOnError("amsdosfile option is empty, set it.", "dsk -dsk output.dsk -get -amsdosfile hello.bin")
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
						fmt.Fprintf(os.Stderr, "Filename to get : %s\n", filename)
						content, err := dskFile.GetFileIn(filename, indice)
						if err != nil {
							fmt.Fprintf(os.Stderr, "Error while getting file in dsk error :%v\n", err)
						}
						filename = strings.ReplaceAll(filename, " ", "")
						af, err := os.Create(filename)
						if err != nil {
							exitOnError(fmt.Sprintf("Error while creating file (%s) error %v\n", filename, err), "Check your dsk  with option -dsk yourdsk.dsk -analyze")
						}
						defer af.Close()
						_, err = af.Write(content)
						if err != nil {
							exitOnError(fmt.Sprintf("Error while copying content in file (%s) error %v\n", filename, err), "Check your dsk  with option -dsk yourdsk.dsk -analyze")
						}
						informations := fmt.Sprintf("Extract file [%s] Indice in DSK [%d] is saved\n", filename, indice)
						resumeAction(*dskPath, "get amsdosfile", *fileInDsk, informations)
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
					filename := strings.ReplaceAll(*fileInDsk, " ", "")
					af, err := os.Create(filename)
					if err != nil {
						exitOnError(fmt.Sprintf("Error while creating file (%s) error %v\n", filename, err), "Check your file path")
					}
					defer af.Close()
					_, err = af.Write(content)
					if err != nil {
						exitOnError(fmt.Sprintf("Error while copying content in file (%s) error %v\n", filename, err), " Check your dsk  with option -dsk yourdsk.dsk -analyze")
					}
					informations := fmt.Sprintf("Extract file [%s] Indice in DSK [%d] is saved\n", filename, indice)
					resumeAction(*dskPath, "get amsdosfile", filename, informations)
				}
			}
		}
		if *put {
			cmdRunned = true
			if *fileInDsk == "" {
				exitOnError("amsdosfile option is empty, set it.", "dsk -dsk output.dsk -put -amsdosfile hello.bin -exec #1000 -load 500")
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
						exitOnError(fmt.Sprintf("Error while inserted file (%s) in dsk (%s) error :%v\n", *fileInDsk, *dskPath, err), "Check your dsk  with option -dsk yourdsk.dsk -analyze")
					}
					resumeAction(*dskPath, "put ascii", *fileInDsk, informations)
				case "binary":
					informations := fmt.Sprintf("execute address [#%.4x], loading address [#%.4x]\n", execAddress, loadAddress)
					if err := dskFile.PutFile(*fileInDsk, dsk.MODE_BINAIRE, loadAddress, execAddress, uint16(*user), false, false); err != nil {
						exitOnError(fmt.Sprintf("Error while inserted file (%s) in dsk (%s) error :%v\n", *fileInDsk, *dskPath, err), "Check your dsk  with option -dsk yourdsk.dsk -analyze")
					}
					resumeAction(*dskPath, "put binary", *fileInDsk, informations)
				default:
					fmt.Fprintf(os.Stderr, "File type option unknown please choose between ascii or binary.")
				}
				f, err := os.Create(*dskPath)
				if err != nil {
					exitOnError(fmt.Sprintf("Error while write file (%s) error %v\n", *dskPath, err), "Check your dsk path file")
				}
				defer f.Close()

				if err := dskFile.Write(f); err != nil {
					exitOnError(fmt.Sprintf("Error while write file (%s) error %v\n", *dskPath, err), "Check your dsk  with option -dsk yourdsk.dsk -analyze")
				}
			}
		}

		if *remove {
			cmdRunned = true
			if *fileInDsk == "" {
				exitOnError("amsdosfile option is empty, set it.", "dsk -dsk output.dsk -remove -amsdosfile hello.bin")
			}
			amsdosFile := dsk.GetNomDir(*fileInDsk)
			indice := dskFile.FileExists(amsdosFile)
			if indice == dsk.NOT_FOUND {
				exitOnError(fmt.Sprintf("File (%s) not found in dsk (%s)\n", *fileInDsk, *dskPath), "Check you dsk")
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
					exitOnError(fmt.Sprintf("Error while write file (%s) error %v\n", *dskPath, err), "Check your dsk path file")
				}
				defer f.Close()
				if err := dskFile.Write(f); err != nil {
					exitOnError(fmt.Sprintf("Error while write file (%s) error %v\n", *dskPath, err), "Check your dsk  with option -dsk yourdsk.dsk -analyze")
				}
			}
		}
	} else {
		if *fileInDsk == "" {
			exitOnError("Error no amsdos file is set\n", "set your amsdos file with option -amsdosfile like dsk -dsk output.dsk -put -amsdosfile hello.bin")
		}
		f, err := os.Open(*fileInDsk)
		if err != nil {
			exitOnError(fmt.Sprintf("Cannot open file %s error :%v\n", *fileInDsk, err), "Check your dsk file path")
		}
		content, err := ioutil.ReadAll(f)
		if err != nil {
			exitOnError(fmt.Sprintf("Cannot read file %s error :%v\n", *fileInDsk, err), "Check your dsk  with option -dsk yourdsk.dsk -analyze")
		}

		if *basic {
			cmdRunned = true
			isAmsdos, _ := dsk.CheckAmsdos(content)
			// remove amsdos header
			if isAmsdos {
				content = content[128:]
			}

			fmt.Fprintf(os.Stderr, "File %s filesize :%d octets\n", *fileInDsk, len(content))
			fmt.Fprintf(os.Stdout, "%s", dsk.Basic(content, uint16(len(content)), true))
		}
		if *desassemble {
			cmdRunned = true
			var address uint16
			isAmsdos, header := dsk.CheckAmsdos(content)
			if isAmsdos {
				address = header.Address
				content = content[128:]
			}
			fmt.Println(dsk.Desass(content, uint16(len(content)), address))
		}
		if *hexa {
			cmdRunned = true
			isAmsdos, _ := dsk.CheckAmsdos(content)
			// remove amsdos header
			if isAmsdos {
				content = content[128:]
			}
			fmt.Println(dsk.DisplayHex(content, 16))
		}
		if *info {
			cmdRunned = true
			isAmsdos, header := dsk.CheckAmsdos(content)
			if !isAmsdos {
				exitOnError(fmt.Sprintf("File (%s) does not contain amsdos header.\n", *fileInDsk), "may be a ascii file")
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
		if *addHeader {
			if *executeAddress == "" {
				exitOnError("When adding amsdos header executing address must be set.", "dsk -addheader -amsdosfile hello.bin -exec #1000 -load 500")
			}
			if *loadingAddress == "" {
				exitOnError("When adding amsdos header loading address must be set.", "dsk -addheader -amsdosfile hello.bin -exec #1000 -load 500")
			}
			switch *fileType {
			case "binary":
				informations := fmt.Sprintf("execute address [#%.4x], loading address [#%.4x]\n", execAddress, loadAddress)
				isAmsdos, header := dsk.CheckAmsdos(content)
				if !isAmsdos {
					exitOnError(fmt.Sprintf("File (%s) does not contain amsdos header.\n", *fileInDsk), "dsk -addheader -amsdosfile hello.bin -exec #1000 -load 500")
				} else {
					fmt.Fprintf(os.Stdout, "File (%s) removing current amsdos header.\n", *fileInDsk)
					content = content[128:]
				}
				var typeModeImport uint8
				filename := dsk.GetNomAmsdos(*fileInDsk)
				header.Size = uint16(len(content))
				header.Size2 = uint16(len(content))
				copy(header.Filename[:], []byte(filename[0:12]))
				header.Address = loadAddress
				if loadAddress != 0 {
					typeModeImport = dsk.MODE_BINAIRE
				}
				header.Exec = execAddress
				if execAddress != 0 {
					typeModeImport = dsk.MODE_BINAIRE
				}
				if typeModeImport == dsk.MODE_BINAIRE {
					header.Type = 1
				}
				// Il faut recalculer le checksum en comptant es adresses !
				header.Checksum = header.ComputedChecksum16()
				var rbuff bytes.Buffer
				binary.Write(&rbuff, binary.LittleEndian, header)
				binary.Write(&rbuff, binary.LittleEndian, content)

				f, err := os.Create(*fileInDsk)
				if err != nil {
					exitOnError(fmt.Sprintf("Error while creating file [%s] error:%v\n", *fileInDsk, err), "Check your dsk file path")
				}
				defer f.Close()
				_, err = f.Write(rbuff.Bytes())
				if err != nil {
					exitOnError(fmt.Sprintf("Error while writing data in file [%s] error:%v\n", *fileInDsk, err), "Check your dsk  with option -dsk yourdsk.dsk -analyze")
				}
				resumeAction("none", "add amsdos header", *fileInDsk, informations)
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
	}

	if !cmdRunned {
		sampleUsage()
	}

	os.Exit(0)
}

func resumeAction(dskFilepath, action, amsdosfile, informations string) {
	fmt.Fprintf(os.Stderr, "DSK path [%s]\n", dskFilepath)
	fmt.Fprintf(os.Stderr, "ACTION: Action on DSK [%s] on amsdos file [%s]\n", action, amsdosfile)
	fmt.Fprintf(os.Stderr, "INFO:   %s\n", informations)
}

func sampleUsage() {

	fmt.Fprintf(os.Stderr, "\nHere sample usages :\n"+
		"\t* Create empty simple dsk file : dsk -dsk output.dsk -format\n"+
		"\t* Create empty simple dsk file with custom tracks and sectors: dsk -dsk output.dsk -format -sector 8 -track 42\n"+
		"\t* Create empty extended dsk file with custom head, tracks and sectors: dsk -dsk output.dsk -format -sector 8 -track 42 -dsktype 1 -head 2\n"+
		"\t* Create empty sna file : dsk -sna output.sna\n"+
		"\t* List dsk content : dsk -dsk output.dsk -list\n"+
		"\t* Get information on Sna file : dsk -sna output.sna -info\n"+
		"\t* Get information on file in dsk  : dsk -dsk output.dsk -amsdosfile hello.bin -info\n"+
		"\t* List file content in hexadecimal in dsk file : dsk -dsk output.dsk -amsdosfile hello.bin -hex\n"+
		"\t* Put file in dsk file : dsk -dsk output.dsk -put -amsdosfile hello.bin -exec #1000 -load 500\n"+
		"\t* Put file in sna file (here for a cpc plus): dsk -sna output.sna -put -amsdosfile hello.bin -exec #1000 -load 500 -screenmode 0 -cpctype 4\n\n\n")
	flag.PrintDefaults()
}

func exitOnError(errorMessage, hint string) {
	//sampleUsage()
	fmt.Fprintf(os.Stderr, "*************************************************************************\n")
	fmt.Fprintf(os.Stderr, "[ERROR] :\t%s\n", errorMessage)
	fmt.Fprintf(os.Stderr, "[HINT ] :\t%s\n", hint)
	fmt.Fprintf(os.Stderr, "*************************************************************************\n")
	os.Exit(-1)
}
