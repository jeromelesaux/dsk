package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jeromelesaux/dsk/amsdos"
	"github.com/jeromelesaux/dsk/dsk"
	"github.com/jeromelesaux/dsk/sna"
	"github.com/jeromelesaux/dsk/utils"
)

var (
	help           = flag.Bool("help", false, "Display extended help.")
	list           = flag.Bool("list", false, "List the contents of a DSK file.")
	track          = flag.Int("track", 39, "Track number (for formatting).")
	heads          = flag.Int("head", 1, "Number of heads in the DSK (format)")
	sector         = flag.Int("sector", 9, "Number of sectors (format).")
	format         = flag.Bool("format", false, "Format the specified DSK or SNA file.")
	dskType        = flag.Int("dsktype", 0, "DSK Type: 0 = DSK, 1 = EDSK, 3 = SNA.")
	dskPath        = flag.String("dsk", "", "Path to the DSK file to handle.")
	hexa           = flag.String("hex", "", "\tDisplay an AMSDOS file in hexadecimal format.")
	info           = flag.String("info", "", "Retrieve information about an AMSDOS file (size, execution, and loading address) or an SNA file.")
	ascii          = flag.String("ascii", "", "Display an AMSDOS file in ASCII format.")
	disassemble    = flag.String("disassemble", "", "Disassemble an AMSDOS file.")
	get            = flag.String("get", "", "\tExtract a file from the DSK file.")
	remove         = flag.String("remove", "", "Remove the AMSDOS file from the DSK file.")
	basic          = flag.String("basic", "", "Display a basic AMSDOS file.")
	put            = flag.String("put", "", "\tInsert the AMSDOS file into the DSK file.")
	executeAddress = flag.String("exec", "", "Execution address for the inserted file (hexadecimal format, e.g., #170 allowed).")
	loadingAddress = flag.String("load", "", "Loading address for the inserted file (hexadecimal format, e.g., #170 allowed).")
	user           = flag.Int("user", 0, "User number for the inserted file.")
	force          = flag.Bool("force", false, "Force overwrite of an existing file in the DSK.")
	//fileType       = flag.String("type", "", "Type of the inserted file: 'ascii' or 'binary'.")
	snaPath      = flag.String("sna", "", "\tPath to the SNA file to handle.")
	analyse      = flag.Bool("analyze", false, "Analyze and display the DSK header.")
	cpcType      = flag.Int("cpctype", 2, "CPC type for SNA import: 0 = CPC464, 1 = CPC664, 2 = CPC6128, 3 = Unknown, 4 = CPCPlus6128, 5 = CPCPlus464, 6 = GX4000.")
	screenMode   = flag.Int("screenmode", 1, "Screen mode parameter for SNA files.")
	vendorFormat = flag.Bool("vendor", false, "Use vendor format for formatting (sector count = #09, last track = #27).")
	dataFormat   = flag.Bool("data", true, "Use data format for formatting (sector count = #09, last track = #27).")
	rawimport    = flag.Bool("rawimport", false, "Perform a raw import of an AMSDOS file. Requires '-dsk', '-track', and '-sector' options. \n\t\tCopies the file directly starting from the specified track and sector. e.g.: dsk -dsk mydskfile.dsk -amsdosfile file.bin -rawimport -track 1 -sector 0")
	rawexport    = flag.Bool("rawexport", false, "Perform a raw export of an AMSDOS file. Requires '-dsk', '-track', '-sector', and '-size' options. \n\t\tExtracts the file content from the specified track and sector up to the given size. e.g.: dsk -dsk mydskfile.dsk -amsdosfile file.bin -rawexport -track 1 -sector 0 -size 16384")
	size         = flag.Int("size", 0, "Size of data to extract for 'rawexport'. See 'rawexport' for details.")
	autotest     = flag.Bool("autotest", false, "Run all available tests.")
	autoextract  = flag.String("autoextract", "", "Extract all DSK files from a specified folder.")
	snaVersion   = flag.Int("snaversion", 1, "Specify the SNA version (1 or 2 available).")
	quiet        = flag.Bool("quiet", false, "Suppress unnecessary output (useful for scripting).")
	stdoutOpt    = flag.Bool("stdout", false, "To redirect to stdout when using get file")
	appVersion   = "0.27"
	version      = flag.Bool("version", false, "Display the application version and exit.")
)

func main() {
	var cmdRunned bool = false
	var addHeader bool
	var execAddress, loadAddress uint16
	fileType := "ascii"

	flag.Usage = sampleUsage
	flag.Parse()

	if *help || len(flag.Args()) == 1 {
		sampleUsage()
		os.Exit(0)
	}

	if *version {
		fmt.Printf("%s", appVersion)
		os.Exit(0)
	}

	if *autotest {
		autotests()
		os.Exit(0)
	}

	if *put != "" {
		content, err := os.ReadFile(*put)
		if err != nil {
			exitOnError(err.Error(), fmt.Sprintf("file %s issue", *put))
		}
		hasHeader, headerInf := amsdos.CheckAmsdos(content)
		if hasHeader {
			fileType = "binary"
			if *executeAddress == "" {
				execAddress = headerInf.Exec
			}
			if *loadingAddress == "" {
				loadAddress = headerInf.Address
			}
		}
	}

	if *loadingAddress != "" || *executeAddress != "" {
		addHeader = true
	}

	if *autoextract != "" {
		files, err := fs.ReadDir(os.DirFS("/"), *autoextract)
		if err != nil {
			exitOnError(err.Error(), "Please check your folder path")
		}
		for _, file := range files {
			if !file.IsDir() {
				if strings.ToUpper(path.Ext(file.Name())) == ".DSK" {
					dskfolderPath := *autoextract + string(filepath.Separator) + strings.Replace(file.Name(), path.Ext(file.Name()), "", -1)
					dskFilepath := *autoextract + string(filepath.Separator) + file.Name()
					err = os.Mkdir(dskfolderPath, os.ModePerm)
					if err != nil && !errors.Is(err, os.ErrExist) {
						exitOnError(err.Error(), "Please check your folder path")
					}

					d, err, msg, _ := openDsk(dskFilepath)
					if err {
						fmt.Fprintf(os.Stderr, "Error while opening file %s error :%s\n", dskFilepath, msg)
					}
					err, msg, _ = getFileDsk(d, "*", dskFilepath, dskfolderPath)
					if err {
						fmt.Fprintf(os.Stderr, "Error while writing file %s in folder %s error :%s\n", dskFilepath, dskfolderPath, msg)
					}
				}
			}
		}
		os.Exit(0)
	}

	if *executeAddress != "" {
		execAddress = parseHexadecimal16bits(*executeAddress)
	}
	if *loadingAddress != "" {
		loadAddress = parseHexadecimal16bits(*loadingAddress)
	}
	if !*quiet {
		fmt.Fprintf(os.Stderr, "DSK cli version [%s]\nMade by Sid (ImpAct)\n", appVersion)
	}
	// gestion des SNAs
	if *snaPath != "" {
		if *info != "" {
			cmdRunned = true
			isError, msg, hint := infoSna(*snaPath)
			if isError {
				exitOnError(msg, hint)
			}
			os.Exit(0)
		}
		if *format {
			cmdRunned = true
			isError, msg, hint := formatSna(*snaPath, *snaVersion)
			if isError {
				exitOnError(msg, hint)
			}
		}
		if *hexa != "" {
			cmdRunned = true
			sna, err := sna.ReadSna(*snaPath)
			if err != nil {
				exitOnError(err.Error(), "Check your sna path")
			}
			content := sna.Hexadecimal()
			fmt.Println(dsk.DisplayHex([]byte(content), 16))
			os.Exit(0)
		}
		if *put != "" {
			cmdRunned = true
			if *put != "" {
				cpcTYPE := sna.CPCType(*cpcType)
				crtc := sna.UM6845R
				if *cpcType > 3 {
					crtc = sna.ASIC_6845
				}
				if err := sna.ImportInSna(*put, *snaPath, execAddress, uint8(*screenMode), cpcTYPE, crtc, *snaVersion); err != nil {
					fmt.Fprintf(os.Stderr, "Error while trying to import file (%s) in new sna (%s) error: %v\n",
						*put,
						*snaPath,
						err)
					os.Exit(1)
				}
				os.Exit(0)
			} else {
				fmt.Fprintf(os.Stderr, "Missing input (argument -amsdosfile) file to import in sna file (%s)\n", *snaPath)
			}
		}
		if *get != "" {
			cmdRunned = true
			if *get != "" {
				content, err := sna.ExportFromSna(*snaPath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error while trying to import file (%s) in new sna (%s) error: %v\n",
						*get,
						*snaPath,
						err)
					os.Exit(1)
				}
				f, err := os.Create(*get)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error while trying to import file (%s) in new sna (%s) error: %v\n",
						*get,
						*snaPath,
						err)
					os.Exit(1)
				}
				defer f.Close()
				_, err = f.Write(content)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error while trying to import file (%s) in new sna (%s) error: %v\n",
						*get,
						*snaPath,
						err)
					os.Exit(1)
				}
				os.Exit(0)
			} else {
				if *dskPath != "" {
					content, err := sna.ExportFromSna(*snaPath)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error while trying to import file (%s) in new sna (%s) error: %v\n",
							*get,
							*snaPath,
							err)
						os.Exit(1)
					}
					d, isError, msg, hint := openDsk(*dskPath)
					if isError {
						exitOnError(msg, hint)
					}
					isError, msg, hint = rawImportDataInDsk(d, *get, *dskPath, *track, *sector, content)
					if isError {
						exitOnError(msg, hint)
					}

				} else {
					fmt.Fprintf(os.Stderr, "Missing input file to import in sna file (%s)\n", *snaPath)
				}
			}
		}
	}

	// verification que le fichier DSK est present

	if *format {
		cmdRunned = true
		isError, msg, hint := formatDsk(*dskPath, *sector, *track, *heads, 0, *vendorFormat, *dataFormat)

		if isError {
			exitOnError(msg, hint)
		}
	}

	if *dskPath != "" {
		d, isError, msg, hint := openDsk(*dskPath)
		if isError {
			exitOnError(msg, hint)
		}
		if *list {
			isError, msg, hint := listDsk(d, *dskPath)
			if isError {
				exitOnError(msg, hint)
			}
		}

		if *analyse {
			cmdRunned = true
			isError, msg, hint := analyseDsk(d, *dskPath)
			if isError {
				exitOnError(msg, hint)
			}
		}

		if *info != "" {
			cmdRunned = true
			isError, msg, hint := fileinfoDsk(d, *info)
			if isError {
				exitOnError(msg, hint)
			}
		}

		if *hexa != "" {
			cmdRunned = true
			isError, msg, hint := hexaFileDsk(d, *hexa)
			if isError {
				exitOnError(msg, hint)
			}
		}

		if *disassemble != "" {
			cmdRunned = true
			isError, msg, hint := desassembleFileDsk(d, *disassemble)
			if isError {
				exitOnError(msg, hint)
			}
		}

		if *ascii != "" {
			cmdRunned = true
			isError, msg, hint := asciiFileDsk(d, *ascii, *stdoutOpt)
			if isError {
				exitOnError(msg, hint)
			}
		}

		if *basic != "" {
			cmdRunned = true
			amsdosFile := dsk.GetNomDir(*basic)
			indice := d.FileExists(amsdosFile)
			if indice == dsk.NOT_FOUND {
				fmt.Fprintf(os.Stderr, "File %s does not exist\n", *basic)
			} else {
				content, err := d.GetFileIn(*basic, indice)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error while getting file in dsk error :%v\n", err)
				}

				hasAmsdos, _ := amsdos.CheckAmsdos(content)
				if hasAmsdos {

					body, filesize, _ := d.ViewFile(indice)
					fmt.Fprintf(os.Stderr, "File %s filesize :%d octets\n", *basic, filesize)
					fmt.Fprintf(os.Stdout, "%s", utils.Basic(body, uint16(filesize), true))
				} else {
					fmt.Fprintf(os.Stderr, "File %s filesize :%d octets\n", *basic, len(content))
					fmt.Fprintf(os.Stdout, "%s", content)
				}
			}
		}

		if *get != "" {
			cmdRunned = true
			directory, err := os.Getwd()
			if err != nil {
				exitOnError(err.Error(), "Please use autoextract option")
			}
			if *stdoutOpt {
				amsdosFile := dsk.GetNomDir(*get)
				indice := d.FileExists(amsdosFile)
				if indice == dsk.NOT_FOUND {
					fmt.Fprintf(os.Stderr, "File %s does not exist\n", *get)
				} else {
					content, err := d.GetFileIn(*get, indice)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error while getting file in dsk error :%v\n", err)
					}

					hasAmsdos, _ := amsdos.CheckAmsdos(content)
					if hasAmsdos {
						body, _, _ := d.ViewFile(indice)
						os.Stdout.Write(body)
					} else {
						os.Stdout.Write(content)
					}
				}
			} else {
				isError, msg, hint := getFileDsk(d, *get, *dskPath, directory)
				if isError {
					exitOnError(msg, hint)
				}
			}
		}

		if *rawimport {
			cmdRunned = true
			isError, msg, hint := rawImportDsk(d, *put, *dskPath, *track, *sector)
			if isError {
				exitOnError(msg, hint)
			}
		}

		if *rawexport {
			cmdRunned = true
			isError, msg, hint := rawExportDsk(d, *put, *dskPath, *track, *sector, *size)
			if isError {
				exitOnError(msg, hint)
			}
		}

		if *put != "" {
			cmdRunned = true
			isError, msg, hint := putFileDsk(d, *put, *dskPath, fileType, loadAddress, execAddress, uint16(*user))
			if isError {
				exitOnError(msg, hint)
			}
		}

		if *remove != "" {
			cmdRunned = true
			isError, msg, hint := removeFileDsk(d, *dskPath, *remove)
			if isError {
				exitOnError(msg, hint)
			}
		}

		// no arguments commands
		if !cmdRunned {
			cmdRunned = true
			isError, msg, hint = listDsk(d, *dskPath)
			if isError {
				exitOnError(msg, hint)
			}
			return
		}
	} else {
		//
		// now dsk will work on an amsdosfile no dsk file set
		//

		if *basic != "" {
			cmdRunned = true
			content, err := os.ReadFile(*basic)
			if err != nil {
				exitOnError(fmt.Sprintf("Cannot read file %s error :%v\n", *basic, err), "cannot read the file content")
			}
			isAmsdos, _ := amsdos.CheckAmsdos(content)
			// remove amsdos header
			if isAmsdos {
				content = content[dsk.HeaderSize:]
			}

			fmt.Fprintf(os.Stderr, "File %s filesize :%d octets\n", *basic, len(content))
			fmt.Fprintf(os.Stdout, "%s", utils.Basic(content, uint16(len(content)), true))
		}
		if *disassemble != "" {
			cmdRunned = true
			var address uint16
			content, err := os.ReadFile(*disassemble)
			if err != nil {
				exitOnError(fmt.Sprintf("Cannot read file %s error :%v\n", *disassemble, err), "cannot read the file content")
			}
			isAmsdos, header := amsdos.CheckAmsdos(content)
			if isAmsdos {
				address = header.Address
				content = content[dsk.HeaderSize:]
			}
			fmt.Println(utils.Desass(content, uint16(len(content)), address))
		}
		if *hexa != "" {
			cmdRunned = true
			content, err := os.ReadFile(*hexa)
			if err != nil {
				exitOnError(fmt.Sprintf("Cannot read file %s error :%v\n", *hexa, err), "cannot read the file content")
			}
			isAmsdos, _ := amsdos.CheckAmsdos(content)
			// remove amsdos header
			if isAmsdos {
				content = content[dsk.HeaderSize:]
			}
			fmt.Println(dsk.DisplayHex(content, 16))
		}
		if *info != "" {
			cmdRunned = true
			content, err := os.ReadFile(*info)
			if err != nil {
				exitOnError(fmt.Sprintf("Cannot read file %s error :%v\n", *info, err), "cannot read the file content")
			}
			isAmsdos, header := amsdos.CheckAmsdos(content)
			if !isAmsdos {
				exitOnError(fmt.Sprintf("File (%s) does not contain amsdos header.\n", *info), "may be a ascii file")
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
		if addHeader {
			if *get == "" {
				exitOnError("Error no file is set\n", "set your amsdos file with option like dsk -get hello.bin")
			}
			content, err := os.ReadFile(*get)
			if err != nil {
				exitOnError(fmt.Sprintf("Cannot read file %s error :%v\n", *get, err), "cannot read the file content")
			}
			informations := fmt.Sprintf("execute address [#%.4x], loading address [#%.4x]\n", execAddress, loadAddress)
			isAmsdos, header := amsdos.CheckAmsdos(content)
			if isAmsdos {
				exitOnError("The file already contains an amsdos header", "Check your file")
			}
			filename := dsk.GetNomAmsdos(*get)
			header.Size = uint16(len(content))
			header.Size2 = uint16(len(content))
			copy(header.Filename[:], []byte(filename[0:12]))
			header.Address = loadAddress
			header.Exec = execAddress
			// Il faut recalculer le checksum en comptant es adresses !
			header.Checksum = header.ComputedChecksum16()
			var rbuff bytes.Buffer
			err = binary.Write(&rbuff, binary.LittleEndian, header)
			if err != nil {
				exitOnError("error while writing header : "+err.Error(), "Check your file")
			}
			err = binary.Write(&rbuff, binary.LittleEndian, content)
			if err != nil {
				exitOnError("error while writing header : "+err.Error(), "Check your file")
			}

			f, err := os.Create(*get)
			if err != nil {
				exitOnError(fmt.Sprintf("Error while creating file [%s] error:%v\n", *get, err), "Check your dsk file path")
			}
			defer f.Close()
			_, err = f.Write(rbuff.Bytes())
			if err != nil {
				exitOnError(fmt.Sprintf("Error while writing data in file [%s] error:%v\n", *get, err), "Check your dsk  with option -dsk yourdsk.dsk -analyze")
			}
			resumeAction("none", "add amsdos header", *get, informations, *quiet)
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
			os.Exit(0)
		}
	}

	if !cmdRunned {
		sampleUsage()
	}

	os.Exit(0)
}

func resumeAction(dskFilepath, action, amsdosfile, informations string, quiet bool) {
	if !quiet {
		fmt.Fprintf(os.Stderr, "DSK path [%s]\n", dskFilepath)
		fmt.Fprintf(os.Stderr, "ACTION: Action on DSK [%s] on amsdos file [%s]\n", action, amsdosfile)
		fmt.Fprintf(os.Stderr, "INFO:   %s\n", informations)
	}
}

func sampleUsage() {
	fmt.Fprintf(os.Stderr, "\nHere are some sample usages:\n"+
		"  dsk -dsk output.dsk -format                  # Create an empty simple DSK file.\n"+
		"  dsk -dsk output.dsk -format -sector 8 -track 42  # Create a simple empty DSK file with custom tracks and sectors.\n"+
		"  dsk -dsk output.dsk -format -sector 8 -track 42 -dsktype 1 -head 2  # Create an empty extended DSK file with custom heads, tracks, and sectors.\n"+
		"  dsk -sna output.sna                          # Create an empty SNA file.\n"+
		"  dsk -info hello.bin                          # Display header informations on the file hello.bin"+
		"  dsk -dsk output.dsk -list                    # List the contents of the DSK file.\n"+
		"  dsk -sna output.sna -info                    # Get information about the SNA file.\n"+
		"  dsk -dsk output.dsk -info hello.bin          # Get information about a file in the DSK.\n"+
		"  dsk -dsk output.dsk -hex hello.bin           # Display the file content in hexadecimal format from the DSK file.\n"+
		"  dsk -dsk output.dsk -put hello.bin -exec \"#1000\" -load \"500\"  # Insert a file into the DSK file.\n"+
		"  dsk -sna output.sna -put hello.bin -exec \"#1000\" -load 500 -screenmode 0 -cpctype 4  # Insert a file into the SNA file (for a CPC Plus system).\n\n")
	fmt.Printf(("Options:\n"))
	flag.VisitAll(func(f *flag.Flag) {
		fmt.Printf("  -%s \t%s (default: %q)\n", f.Name, f.Usage, f.DefValue)
	})
}

func exitOnError(errorMessage, hint string) {
	// sampleUsage()
	fmt.Fprintf(os.Stderr, "*************************************************************************\n")
	fmt.Fprintf(os.Stderr, "[ERROR] :\t%s\n", errorMessage)
	fmt.Fprintf(os.Stderr, "[HINT ] :\t%s\n", hint)
	fmt.Fprintf(os.Stderr, "*************************************************************************\n")
	os.Exit(-1)
}

func formatDsk(dskPath string, sector, track, heads, extendedDskType int, vendorFormat bool, dataFormat bool) (onError bool, message, hint string) {
	_, err := os.Stat(dskPath)
	if err == nil {
		if !*force {
			return true, fmt.Sprintf("Error file (%s) already exists", dskPath), "Use option -force to avoid this message"
		}
	}
	f, err := os.Create(dskPath)
	if err != nil {
		return true, fmt.Sprintf("Error while write file (%s) error %v", dskPath, err), "Check your dsk file path."
	}
	defer f.Close()
	fmt.Fprintf(os.Stderr, "Formating number of sectors (%d), tracks (%d), head number (%d)\n", sector, track, heads)
	var dskFile *dsk.DSK
	if dataFormat {
		dskFile = dsk.FormatDsk(uint8(sector), uint8(track), uint8(heads), dsk.DataFormat, (extendedDskType))
	} else {
		if vendorFormat {
			dskFile = dsk.FormatDsk(uint8(sector), uint8(track), uint8(heads), dsk.VendorFormat, (extendedDskType))
		}
	}
	if err := dskFile.Write(f); err != nil {
		return true, fmt.Sprintf("Error while write file (%s) error %v\n", dskPath, err), "Check your dsk file with option -dsk yourdsk.dsk -analyze"
	}
	return false, "", ""
}

func formatSna(snaPath string, snaVersion int) (onError bool, message, hint string) {
	if _, err := sna.CreateSna(snaPath, snaVersion); err != nil {
		return true, fmt.Sprintf("Cannot create Sna file (%s) error : %v\n", snaPath, err), ""
	}
	fmt.Fprintf(os.Stderr, "Sna file (%s) created.\n", snaPath)
	return false, "", ""
}

func infoSna(snaPath string) (onError bool, message, hint string) {
	f, err := os.Open(snaPath)
	if err != nil {
		exitOnError(fmt.Sprintf("Error while read sna file (%s) error %v", snaPath, err), "Check your sna file")
	}
	defer f.Close()
	sna := &sna.SNA{}
	if err := sna.Read(f); err != nil {
		return true, fmt.Sprintf("Error while reading sna file (%s) error %v", snaPath, err), "Check your sna file"
	}
	fmt.Fprintf(os.Stderr, "Sna (%s) description :\n\tCPC type:%s\n\tCRTC type:%s\n", snaPath, sna.CPCType(), sna.CRTCType())
	fmt.Fprintf(os.Stderr, "\tSna version:%d\n\tMemory size:%dKo\n", sna.Header.Version, sna.Header.MemoryDumpSize)
	fmt.Fprintf(os.Stderr, "%s\n", sna.Header.String())
	return false, "", ""
}

func listDsk(d dsk.DSK, dskPath string) (onError bool, message, hint string) {
	if err := d.GetCatalogue(); err != nil {
		return true, fmt.Sprintf("Error while getting catalogue in dsk file (%s) error %v\n", dskPath, err), "Check your dsk file with option -dsk yourdsk.dsk -analyze"
	}
	totalUsed := 0
	for _, i := range d.GetFilesIndices() {
		size := fmt.Sprintf("%.3d ko", d.GetFilesize(d.Catalogue[i]))
		totalUsed += d.GetFilesize(d.Catalogue[i])
		ext := d.Catalogue[i].Ext
		for i := range ext {
			if ext[i] == 0xA0 {
				ext[i] = ' '
			}
		}
		filename := fmt.Sprintf("%s.%s", d.Catalogue[i].Nom, ext)
		fmt.Fprintf(os.Stdout, "[%.2d] : %s : %d %s\n", i, filename, int(d.Catalogue[i].User), size)
	}
	fmt.Fprintf(os.Stdout, "Dsk %.3d Ko used\n", totalUsed)
	return false, "", ""
}

func openDsk(osFile string) (d dsk.DSK, onError bool, message, hint string) {
	if _, err := os.Stat(osFile); errors.Is(err, os.ErrNotExist) {
		onError, msgErr, hint := formatDsk(osFile, *sector, *track, *heads, 0, true, false)
		if onError {
			return dsk.DSK{}, onError, msgErr, hint
		}
	}
	f, err := os.Open(osFile)
	if err != nil {
		// format new dsk
		return d, true, fmt.Sprintf("Error while read file (%s) error %v\n", osFile, err), "Check your dsk file path"
	}
	defer f.Close()
	if err := d.Read(f); err != nil {
		return d, true, fmt.Sprintf("Error while read dsk file (%s) error %v\n", *dskPath, err), "Check your dsk file path or Check your dsk file with option -dsk yourdsk.dsk -analyze"
	}

	if err := d.CheckDsk(); err != nil {
		return d, true, fmt.Sprintf("Error while read dsk file (%s) error %v\n", osFile, err), "Check your dsk file path or Check your dsk file with option -dsk yourdsk.dsk -analyze"
	}
	if !*quiet {
		fmt.Fprintf(os.Stderr, "Dsk file (%s)\n", osFile)
	}
	return d, false, "", ""
}

func fileinfoDsk(d dsk.DSK, fileInDsk string) (onError bool, message, hint string) {
	if fileInDsk == "" {
		return true, "amsdosfile option is empty, set it.", "usage sample : dsk -dsk output.dsk -amsdosfile hello.bin -info "
	}
	amsdosFile := dsk.GetNomDir(fileInDsk)
	indice := d.FileExists(amsdosFile)
	if indice == dsk.NOT_FOUND {
		fmt.Fprintf(os.Stderr, "File %s does not exist\n", fileInDsk)
	} else {
		content, err := d.GetFileIn(fileInDsk, indice)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while getting file in dsk error :%v\n", err)
		}
		isAmsdos, header := amsdos.CheckAmsdos(content)
		if !isAmsdos {
			return true, fmt.Sprintf("File (%s) does not contain amsdos header.\n", fileInDsk), "add address of execution and loading like : dsk -dsk output.dsk -put -amsdosfile hello.bin -exec \"#1000\" -load 500"
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
	return false, "", ""
}

func parseHexadecimal16bits(address string) (value16 uint16) {
	switch address[0] {
	case '#':
		value := strings.Replace(address, "#", "", -1)
		v, err := strconv.ParseUint(value, 16, 16)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot get the hexadecimal value fom %s, error : %v\n", *executeAddress, err)
		} else {
			value16 = uint16(v)
		}
	case '0':
		value := strings.Replace(address, "0x", "", -1)
		v, err := strconv.ParseUint(value, 16, 16)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot get the hexadecimal value fom %s, error : %v\n", *executeAddress, err)
		} else {
			value16 = uint16(v)
		}
	default:
		v, err := strconv.ParseUint(address, 10, 16)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cannot get the hexadecimal value fom %s, error : %v\n", *executeAddress, err)
		} else {
			value16 = uint16(v)
		}
	}
	return
}

func analyseDsk(d dsk.DSK, dskPath string) (onError bool, message, hint string) {
	if err := d.CheckDsk(); err != nil {
		return true, fmt.Sprintf("Error while read dsk file (%s) error %v\n", dskPath, err), "Check your dsk file path or Check your dsk file with option -dsk yourdsk.dsk -analyze"
	}
	fmt.Fprintf(os.Stderr, "Dsk file (%s)\n", dskPath)
	entry := d.Entry
	fmt.Fprintf(os.Stderr, "Dsk entry %s\n", entry.ToString())
	return false, "", ""
}

func putFileDsk(d dsk.DSK, fileInDsk, dskPath string, fileType string, loadAddress, execAddress, user uint16) (onError bool, message, hint string) {
	if fileInDsk == "" {
		exitOnError("amsdosfile option is empty, set it.", "dsk -dsk output.dsk -put -amsdosfile hello.bin -exec \"#1000\" -load 500")
	}
	amsdosFile := dsk.GetNomDir(fileInDsk)
	indice := d.FileExists(amsdosFile)
	if indice != dsk.NOT_FOUND && !*force {
		exitOnError(fmt.Sprintf("File %s already exists\n", fileInDsk), "use -force to force file put")
	} else {
		if indice != dsk.NOT_FOUND && *force {
			// suppress file
			err := d.RemoveFile(uint8(indice))
			if err != nil {
				exitOnError(fmt.Sprintf("error while removing file %v", err), "check your dsk content")
			}
		}
		switch fileType {
		case "ascii":
			informations := fmt.Sprintf("execute address [#%.4x], loading address [#%.4x]\n", execAddress, loadAddress)
			if err := d.PutFile(fileInDsk, dsk.MODE_ASCII, 0, 0, user, false, false); err != nil {
				return true, fmt.Sprintf("Error while inserted file (%s) in dsk (%s) error :%v\n", fileInDsk, dskPath, err), "Check your dsk  with option -dsk yourdsk.dsk -analyze"
			}
			resumeAction(dskPath, "put ascii", fileInDsk, informations, *quiet)
		case "binary":
			informations := fmt.Sprintf("execute address [#%.4x], loading address [#%.4x]\n", execAddress, loadAddress)
			if err := d.PutFile(fileInDsk, dsk.MODE_BINAIRE, loadAddress, execAddress, user, false, false); err != nil {
				return true, fmt.Sprintf("Error while inserted file (%s) in dsk (%s) error :%v\n", fileInDsk, dskPath, err), "Check your dsk  with option -dsk yourdsk.dsk -analyze"
			}
			resumeAction(dskPath, "put binary", fileInDsk, informations, *quiet)
		default:
			fmt.Fprintf(os.Stderr, "File type option unknown please choose between ascii or binary.")
		}
		f, err := os.Create(dskPath)
		if err != nil {
			return true, fmt.Sprintf("Error while write file (%s) error %v\n", dskPath, err), "Check your dsk path file"
		}
		defer f.Close()

		if err := d.Write(f); err != nil {
			return true, fmt.Sprintf("Error while write file (%s) error %v\n", dskPath, err), "Check your dsk with option -dsk yourdsk.dsk -analyze"
		}
	}
	return false, "", ""
}

func getFileDsk(d dsk.DSK, fileInDsk, dskPath, directory string) (onError bool, message, hint string) {
	if fileInDsk == "" {
		return true, "amsdosfile option is empty, set it.", "dsk -dsk output.dsk -get -amsdosfile hello.bin"
	}
	if fileInDsk == "*" {
		err := d.GetCatalogue()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while getting the catalogue in dsk error :%v\n", err)
		}
		var lastFilename string
		for indice, v := range d.Catalogue {
			if v.User != dsk.USER_DELETED && v.NbPages != 0 {
				var nom, ext string
				nom = dsk.ToAscii(v.Nom[:])
				ext = dsk.ToAscii(v.Ext[:])
				filename := fmt.Sprintf("%s.%s", nom, ext)
				if lastFilename == filename {
					continue
				}
				lastFilename = filename
				fmt.Fprintf(os.Stderr, "Filename to get : %s\n", filename)
				content, err := d.GetFileIn(filename, indice)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error while getting file in dsk error :%v\n", err)
				}
				filename = strings.ReplaceAll(filename, " ", "")
				var af *os.File
				filename = strings.ReplaceAll(filename, " ", "")
				var fPath string
				if directory == "" {
					fPath = filename
					af, err = os.Create(filename)
				} else {
					fPath = directory + string(filepath.Separator) + filename
					af, err = os.Create(fPath)
				}
				if err != nil {
					return true, fmt.Sprintf("Error while creating file (%s) error %v\n", filename, err), "Check your dsk  with option -dsk yourdsk.dsk -analyze"
				}
				_, err = af.Write(content)
				if err != nil {
					return true, fmt.Sprintf("Error while copying content in file (%s) error %v\n", filename, err), "Check your dsk  with option -dsk yourdsk.dsk -analyze"
				}
				af.Close()
				informations := fmt.Sprintf("Extract file [%s] Indice in DSK [%d] is saved\n", fPath, indice)
				resumeAction(dskPath, "get amsdosfile", fileInDsk, informations, *quiet)
			}
		}
	} else {
		amsdosFile := dsk.GetNomDir(fileInDsk)
		indice := d.FileExists(amsdosFile)
		if indice == dsk.NOT_FOUND {
			fmt.Fprintf(os.Stderr, "File %s does not exist\n", fileInDsk)
		} else {
			content, err := d.GetFileIn(fileInDsk, indice)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error while getting file in dsk error :%v\n", err)
			}
			filename := strings.ReplaceAll(fileInDsk, " ", "")
			af, err := os.Create(filename)
			if err != nil {
				return true, fmt.Sprintf("Error while creating file (%s) error %v\n", filename, err), "Check your file path"
			}
			defer af.Close()
			_, err = af.Write(content)
			if err != nil {
				return true, fmt.Sprintf("Error while copying content in file (%s) error %v\n", filename, err), " Check your dsk  with option -dsk yourdsk.dsk -analyze"
			}
			informations := fmt.Sprintf("Extract file [%s] Indice in DSK [%d] is saved\n", filename, indice)
			resumeAction(dskPath, "get amsdosfile", filename, informations, *quiet)
		}
	}
	return false, "", ""
}

func removeFileDsk(d dsk.DSK, dskPath, fileInDsk string) (onError bool, message, hint string) {
	if fileInDsk == "" {
		return true, "amsdosfile option is empty, set it.", "dsk -dsk output.dsk -remove -amsdosfile hello.bin"
	}
	amsdosFile := dsk.GetNomDir(fileInDsk)
	indice := d.FileExists(amsdosFile)
	if indice == dsk.NOT_FOUND {
		return true, fmt.Sprintf("File (%s) not found in dsk (%s)\n", fileInDsk, dskPath), "Check you dsk"
	}
	if err := d.RemoveFile(uint8(indice)); err != nil {
		fmt.Fprintf(os.Stderr, "Error while removing file %s (indice:%d) error :%v\n", fileInDsk, indice, err)
	} else {
		fmt.Fprintf(os.Stderr, "File (%.8s.%.3s) deleted in dsk (%s)\n",
			amsdosFile.Nom,
			amsdosFile.Ext,
			dskPath)
		f, err := os.Create(dskPath)
		if err != nil {
			return true, fmt.Sprintf("Error while write file (%s) error %v\n", dskPath, err), "Check your dsk path file"
		}
		defer f.Close()
		if err := d.Write(f); err != nil {
			return true, fmt.Sprintf("Error while write file (%s) error %v\n", dskPath, err), "Check your dsk  with option -dsk yourdsk.dsk -analyze"
		}
	}
	return false, "", ""
}

func asciiFileDsk(d dsk.DSK, fileInDsk string, isSdtout bool) (onError bool, message, hint string) {
	if fileInDsk == "" {
		return true, "amsdosfile option is empty, set it.", "dsk -dsk output.dsk -ascii -amsdosfile hello.txt"
	}
	amsdosFile := dsk.GetNomDir(fileInDsk)
	indice := d.FileExists(amsdosFile)
	if indice == dsk.NOT_FOUND {
		fmt.Fprintf(os.Stderr, "File %s does not exist\n", fileInDsk)
	} else {
		content, filesize, err := d.ViewFile(indice)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while getting file in dsk error :%v\n", err)
		}
		paddedFileSize := 0
		for i := filesize - 1; i != 0; i-- {
			if content[i] == 0x1A {
				paddedFileSize = i - 1
				break
			}

		}

		if paddedFileSize != 0 {
			filesize = paddedFileSize
		}
		if isSdtout {
			os.Stdout.Write(content[0:filesize])
		} else {
			fmt.Println(string(content[0:filesize]))
		}
	}
	return false, "", ""
}

func desassembleFileDsk(d dsk.DSK, fileInDsk string) (onError bool, message, hint string) {
	if fileInDsk == "" {
		return true, "amsdosfile option is empty, set it.", "dsk -dsk output.dsk -desassemble -amsdosfile hello.bin"
	}
	amsdosFile := dsk.GetNomDir(fileInDsk)
	indice := d.FileExists(amsdosFile)
	if indice == dsk.NOT_FOUND {
		fmt.Fprintf(os.Stderr, "File %s does not exist\n", fileInDsk)
	} else {
		content, filesize, err := d.ViewFile(indice)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while getting file in dsk error :%v\n", err)
		}
		var address uint16
		raw, err := d.GetFileIn(fileInDsk, indice)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while getting file in dsk error :%v\n", err)
		} else {
			isAmsdos, header := amsdos.CheckAmsdos(raw)
			if isAmsdos {
				address = header.Exec
			}
		}

		fmt.Println(utils.Desass(content[0:filesize], uint16(filesize), address))
	}
	return false, "", ""
}

func hexaFileDsk(d dsk.DSK, fileInDsk string) (onError bool, message, hint string) {
	if fileInDsk == "" {
		return true, "amsdosfile option is empty, set it.", "dsk -dsk output.dsk -hex -amsdosfile hello.bin"
	}
	amsdosFile := dsk.GetNomDir(fileInDsk)
	indice := d.FileExists(amsdosFile)
	if indice == dsk.NOT_FOUND {
		fmt.Fprintf(os.Stderr, "File %s does not exist\n", fileInDsk)
	} else {
		content, fileSize, err := d.ViewFile(indice)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while getting file in dsk error :%v\n", err)
		}
		fmt.Println(dsk.DisplayHex(content[0:fileSize], 16))
	}
	return false, "", ""
}

func rawImportDataInDsk(d dsk.DSK, fileInDsk, dskPath string, track, sector int, content []byte) (onError bool, message, hint string) {
	if track == 39 {
		fmt.Fprintf(os.Stdout, "Warning the starting track is set as default : [%d]\n", track)
	}
	if sector == 9 {
		fmt.Fprintf(os.Stdout, "Warning the starting sector is set as default : [%d]\n", sector)
	}
	endedTrack, endedSector, err := d.CopyRawFile(content, uint16(len(content)), track, sector)
	if err != nil {
		return true, fmt.Sprintf("Cannot write file %s error :%v\n", fileInDsk, err), "Check your file path"
	}
	f, err := os.Create(dskPath)
	if err != nil {
		return true, fmt.Sprintf("Error while write file (%s) error %v\n", dskPath, err), "Check your dsk path file"
	}
	defer f.Close()

	if err := d.Write(f); err != nil {
		return true, fmt.Sprintf("Error while write file (%s) error %v\n", dskPath, err), "Check your dsk  with option -dsk yourdsk.dsk -analyze"
	}
	informations := fmt.Sprintf("raw copy file [%s] size [%d] starting at track [%d] sector [%d] and ending at track [%d] sector [%d]",
		fileInDsk,
		len(content),
		track,
		sector,
		endedTrack,
		endedSector)
	resumeAction(dskPath, "raw import ", fileInDsk, informations, *quiet)
	return false, "", ""
}

func rawImportDsk(d dsk.DSK, fileInDsk, dskPath string, track, sector int) (onError bool, message, hint string) {
	if fileInDsk == "" {
		return true, "amsdosfile option is empty, set it.", "dsk -dsk output.dsk -put -amsdosfile hello.bin -rawimport -track 1 -sector 0"
	}

	fr, err := os.Open(fileInDsk)
	if err != nil {
		return true, fmt.Sprintf("Cannot open file %s error :%v\n", fileInDsk, err), "Check your file path"
	}
	defer fr.Close()
	buf, err := io.ReadAll(fr)
	if err != nil {
		return true, fmt.Sprintf("Cannot read file %s error :%v\n", fileInDsk, err), "Check your file path"
	}
	fmt.Fprintf(os.Stdout, "Writing file content [%s] in dsk [%s] starting at track [%d] sector [%d]\n",
		fileInDsk,
		dskPath,
		track,
		sector,
	)
	return rawImportDataInDsk(d, fileInDsk, dskPath, track, sector, buf)
}

func rawExportDsk(d dsk.DSK, fileInDsk, dskPath string, track, sector, size int) (onError bool, message, hint string) {
	if fileInDsk == "" {
		return true, "amsdosfile option is empty, set it.", "dsk -dsk output.dsk -put -amsdosfile hello.bin -rawimport -track 1 -sector 0"
	}

	if track == 39 {
		fmt.Fprintf(os.Stdout, "Warning the starting track is set as default : [%d]\n", track)
	}
	if sector == 9 {
		fmt.Fprintf(os.Stdout, "Warning the starting sector is set as default : [%d]\n", sector)
	}
	if size == 0 {
		fmt.Fprintf(os.Stdout, "Warning the size is set as default : [%d]\n", size)
	}

	fmt.Fprintf(os.Stdout, "Writing file content starting from track [%d] sector [%d] to file  [%s] in dsk [%s] size [%d]\n",
		track,
		sector,
		fileInDsk,
		dskPath,
		size,
	)
	endedTrack, endedSector, content := d.ExtractRawFile(uint16(size), track, sector)

	fw, err := os.Create(fileInDsk)
	if err != nil {
		return true, fmt.Sprintf("Error while write file (%s) error %v\n", fileInDsk, err), "Check your amsdos path file"
	}
	defer fw.Close()
	_, err = fw.Write(content)
	if err != nil {
		return true, fmt.Sprintf("Error while write file (%s) error %v\n", dskPath, err), "Check your dsk  with option -dsk yourdsk.dsk -analyze"
	}
	informations := fmt.Sprintf("raw extract to file [%s] size [%d] starting at track [%d] sector [%d] and ending at track [%d] sector [%d]",
		fileInDsk,
		size,
		track,
		sector,
		endedTrack,
		endedSector)
	resumeAction(dskPath, "raw export ", fileInDsk, informations, *quiet)
	return false, "", ""
}

/*************************/
/* auto tests functions  */
/*************************/

func autotests() {
	var testsOnError int
	var testsDone int
	testsDone++
	if parsing16bitsCAnnotation() {
		fmt.Printf("OK\n")
	} else {
		fmt.Printf("OK\n")
		testsOnError++
	}
	testsDone++
	if parsing16bitsRasmAnnotation() {
		fmt.Printf("OK\n")
	} else {
		fmt.Printf("OK\n")
		testsOnError++
	}
	testsDone++
	if parsing8bitsRasmAnnotation() {
		fmt.Printf("OK\n")
	} else {
		fmt.Printf("OK\n")
		testsOnError++
	}
	testsDone++
	if parsing16bitsIntegerAnnotation() {
		fmt.Printf("OK\n")
	} else {
		fmt.Printf("OK\n")
		testsOnError++
	}

	dskpath := "vendor-format.dsk"
	os.Remove(dskpath)
	testsDone++
	if formatVendorTest(dskpath) {
		fmt.Printf("KO\n")
		testsOnError++
	} else {
		fmt.Printf("OK\n")
	}
	testsDone++
	if formatVendorForceTest(dskpath) {
		fmt.Printf("KO\n")
		testsOnError++
	} else {
		fmt.Printf("OK\n")
	}
	testsDone++
	if listVendorTest(dskpath) {
		fmt.Printf("KO\n")
		testsOnError++
	} else {
		fmt.Printf("OK\n")
	}
	os.Remove(dskpath)
	dskpath = "data-format-double.dsk"
	testsDone++
	if formatDoubleHeadDataTest(dskpath) {
		fmt.Printf("KO\n")
		testsOnError++
	} else {
		fmt.Printf("OK\n")
	}
	os.Remove(dskpath)
	dskpath = "data-format.dsk"
	testsDone++
	if formatDataTest(dskpath) {
		fmt.Printf("KO\n")
		testsOnError++
	} else {
		fmt.Printf("OK\n")
	}
	testsDone++
	if formatDataForceTest(dskpath) {
		fmt.Printf("KO\n")
		testsOnError++
	} else {
		fmt.Printf("OK\n")
	}
	testsDone++
	if listDataTest(dskpath) {
		fmt.Printf("KO\n")
		testsOnError++
	} else {
		fmt.Printf("OK\n")
	}
	testsDone++
	if analyseDataTest(dskpath) {
		fmt.Printf("KO\n")
		testsOnError++
	} else {
		fmt.Printf("OK\n")
	}
	testsDone++
	if saveToFile(generateSliceByte(512), "test.bin") {
		fmt.Printf("KO\n")
		testsOnError++
	} else {
		fmt.Printf("OK\n")
	}

	testsDone++
	if putFileBinaryDataTest("test.bin", dskpath) {
		fmt.Printf("KO\n")
		testsOnError++
	} else {
		fmt.Printf("OK\n")
	}
	testsDone++
	if err := os.Rename("test.bin", "test2.bin"); err != nil {
		testsOnError++
	}
	testsDone++
	if getFileBinaryDataTest("test.bin", dskpath) {
		fmt.Printf("KO\n")
		testsOnError++
	} else {
		fmt.Printf("OK\n")
	}
	testsDone++
	if saveToFile(generateSliceByte(2048), "test3.bin") {
		fmt.Printf("KO\n")
		testsOnError++
	} else {
		fmt.Printf("OK\n")
	}

	testsDone++
	if putFileBinaryDataTest("test3.bin", dskpath) {
		fmt.Printf("KO\n")
		testsOnError++
	} else {
		fmt.Printf("OK\n")
	}
	testsDone++
	if err := os.Rename("test3.bin", "test4.bin"); err != nil {
		testsOnError++
	}
	testsDone++
	if getFileBinaryDataTest("test3.bin", dskpath) {
		fmt.Printf("KO\n")
		testsOnError++
	} else {
		fmt.Printf("OK\n")
	}
	testsDone++
	if asciiFileBinaryDataTest("test3.bin", dskpath) {
		fmt.Printf("KO\n")
		testsOnError++
	} else {
		fmt.Printf("OK\n")
	}
	testsDone++
	if desassembleFileBinaryDataTest("test3.bin", dskpath) {
		fmt.Printf("KO\n")
		testsOnError++
	} else {
		fmt.Printf("OK\n")
	}
	testsDone++
	if hexaFileBinaryDataTest("test3.bin", dskpath) {
		fmt.Printf("KO\n")
		testsOnError++
	} else {
		fmt.Printf("OK\n")
	}
	testsDone++
	if rawImportFileBinaryDataTest("test3.bin", dskpath) {
		fmt.Printf("KO\n")
		testsOnError++
	} else {
		fmt.Printf("OK\n")
	}
	testsDone++
	if rawExportFileBinaryDataTest("test3.bin", dskpath, 2048) {
		fmt.Printf("KO\n")
		testsOnError++
	} else {
		fmt.Printf("OK\n")
	}
	testsDone++
	if removeFileBinaryDataTest("test3.bin", dskpath) {
		fmt.Printf("KO\n")
		testsOnError++
	} else {
		fmt.Printf("OK\n")
	}
	testsDone++
	os.Remove("test.bin")
	testsDone++
	os.Remove("test2.bin")
	testsDone++
	os.Remove("test3.bin")
	testsDone++
	os.Remove("test4.bin")
	testsDone++
	os.Remove(dskpath)

	snapath := "test.sna"
	testsDone++
	if formatSnaTest(snapath) {
		fmt.Printf("KO\n")
		testsOnError++
	} else {
		fmt.Printf("OK\n")
	}
	testsDone++
	if infoSnaTest(snapath) {
		fmt.Printf("KO\n")
		testsOnError++
	} else {
		fmt.Printf("OK\n")
	}
	testsDone++
	os.Remove(snapath)

	fmt.Printf("Tests done [%d] on failure [%d]\n", testsDone, testsOnError)
}

func resetArguments() {
	*help = false
	*list = false
	*track = 39
	*heads = 1
	*sector = 9
	*format = false
	*dskType = 0
	*dskPath = ""
	*hexa = ""
	*info = ""
	*ascii = ""
	*disassemble = ""
	*get = ""
	*remove = ""
	*basic = ""
	*put = ""
	*executeAddress = ""
	*loadingAddress = ""
	*user = 0
	*force = false
	*snaPath = ""
	*analyse = false
	*cpcType = 2
	*screenMode = 1
	*vendorFormat = false
	*dataFormat = true
	*rawimport = false
	*rawexport = false
	*size = 0
	*autotest = false
}

func saveToFile(b []byte, filePath string) bool {
	fw, err := os.Create(filePath)
	if err != nil {
		return true
	}
	defer fw.Close()
	_, err = fw.Write(b)
	return err != nil
}

func generateSliceByte(size int) []byte {
	b := make([]byte, size)
	for i := 0; i < size; i++ {
		v := rand.Intn(255 - 0)
		b[i] = byte(v)
	}

	return b
}

func formatVendorTest(dskFilepath string) bool {
	resetArguments()
	fmt.Printf("Format vendor format ")
	*format = true
	sector := *sector
	track := *track
	heads := *heads
	onError, _, _ := formatDsk(dskFilepath, sector, track, heads, 0, true, false)
	return onError
}

func formatVendorForceTest(dskFilepath string) bool {
	resetArguments()
	fmt.Printf("Format vendor format force option ")
	*format = true
	sector := *sector
	track := *track
	heads := *heads
	*force = true
	onError, _, _ := formatDsk(dskFilepath, sector, track, heads, 0, true, false)
	return onError
}

func formatDoubleHeadDataTest(dskFilepath string) bool {
	resetArguments()
	fmt.Printf("Format data format ")
	*format = true
	sector := *sector
	track := *track
	heads := 2
	dskType := *dskType
	onError, _, _ := formatDsk(dskFilepath, sector, track, heads, dskType, false, true)
	return onError
}

func formatDataTest(dskFilepath string) bool {
	resetArguments()
	fmt.Printf("Format data format ")
	*format = true
	sector := *sector
	track := *track
	heads := *heads
	dskType := *dskType
	onError, _, _ := formatDsk(dskFilepath, sector, track, heads, dskType, false, true)
	return onError
}

func formatDataForceTest(dskFilepath string) bool {
	resetArguments()
	fmt.Printf("Format data format force option")
	*format = true
	*force = true
	sector := *sector
	track := *track
	heads := *heads
	dskType := *dskType
	onError, _, _ := formatDsk(dskFilepath, sector, track, heads, dskType, false, true)
	return onError
}

func formatSnaTest(snafilepath string) bool {
	resetArguments()
	fmt.Printf("Format sna image file ")
	onError, _, _ := formatSna(snafilepath, 2)
	return onError
}

func infoSnaTest(snafilepath string) bool {
	resetArguments()
	fmt.Printf("Get information from sna image file ")
	*info = snafilepath
	onError, _, _ := infoSna(snafilepath)
	return onError
}

func listVendorTest(dskFilepath string) bool {
	resetArguments()
	fmt.Printf("list vendor format ")
	*vendorFormat = true
	*format = true
	d, onError, _, _ := openDsk(dskFilepath)
	if onError {
		return onError
	}
	onError, _, _ = listDsk(d, dskFilepath)
	return onError
}

func listDataTest(dskFilepath string) bool {
	resetArguments()
	fmt.Printf("list vendor format ")
	*format = true
	d, onError, _, _ := openDsk(dskFilepath)
	if onError {
		return onError
	}
	onError, _, _ = listDsk(d, dskFilepath)
	return onError
}

func analyseDataTest(dskFilepath string) bool {
	resetArguments()
	fmt.Printf("list vendor format ")
	*analyse = true
	d, onError, _, _ := openDsk(dskFilepath)
	if onError {
		return onError
	}
	onError, _, _ = analyseDsk(d, dskFilepath)
	return onError
}

func parsing16bitsRasmAnnotation() bool {
	fmt.Printf("Parsing value rasm annotation #C000 ")
	v := parseHexadecimal16bits("#C000")
	return v == 0xC000
}

func parsing16bitsCAnnotation() bool {
	fmt.Printf("Parsing value c annotation 0xC000 ")
	v := parseHexadecimal16bits("0xC000")
	return v == 0xC000
}

func parsing16bitsIntegerAnnotation() bool {
	fmt.Printf("Parsing value integer annotation 49152 ")
	v := parseHexadecimal16bits("49152")
	return v == 0xC000
}

func parsing8bitsRasmAnnotation() bool {
	fmt.Printf("Parsing value c annotation #D0 ")
	v := parseHexadecimal16bits("#D0")
	return v == 0xD0
}

func putFileBinaryDataTest(filePath, dskFilepath string) bool {
	*put = dskFilepath
	fileType := "binary"
	d, onError, _, _ := openDsk(dskFilepath)
	if onError {
		return onError
	}
	isError, _, _ := putFileDsk(d, filePath, dskFilepath, fileType, 0x800, 0x800, uint16(*user))
	return isError
}

func getFileBinaryDataTest(filePath, dskFilepath string) bool {
	*get = dskFilepath
	// fileType := "binary"
	d, onError, _, _ := openDsk(dskFilepath)
	if onError {
		return onError
	}
	isError, _, _ := getFileDsk(d, filePath, dskFilepath, "")
	return isError
}

func removeFileBinaryDataTest(fileInDsk, dskFilepath string) bool {
	*remove = fileInDsk
	d, onError, _, _ := openDsk(dskFilepath)
	if onError {
		return onError
	}
	isError, _, _ := removeFileDsk(d, dskFilepath, fileInDsk)
	return isError
}

func asciiFileBinaryDataTest(fileInDsk, dskFilepath string) bool {
	*ascii = fileInDsk
	d, onError, _, _ := openDsk(dskFilepath)
	if onError {
		return onError
	}
	isError, _, _ := asciiFileDsk(d, fileInDsk, false)
	return isError
}

func desassembleFileBinaryDataTest(fileInDsk, dskFilepath string) bool {
	*disassemble = fileInDsk
	d, onError, _, _ := openDsk(dskFilepath)
	if onError {
		return onError
	}
	isError, _, _ := desassembleFileDsk(d, fileInDsk)
	return isError
}

func hexaFileBinaryDataTest(fileInDsk, dskFilepath string) bool {
	*hexa = fileInDsk
	d, onError, _, _ := openDsk(dskFilepath)
	if onError {
		return onError
	}
	isError, _, _ := hexaFileDsk(d, fileInDsk)
	return isError
}

func rawImportFileBinaryDataTest(filePath, dskFilepath string) bool {
	*rawimport = true
	sector := 0
	track := 1
	d, onError, _, _ := openDsk(dskFilepath)
	if onError {
		return onError
	}
	isError, _, _ := rawImportDsk(d, filePath, dskFilepath, track, sector)
	return isError
}

func rawExportFileBinaryDataTest(filePath, dskFilepath string, length int) bool {
	*rawexport = true
	sector := 0
	track := 1
	d, onError, _, _ := openDsk(dskFilepath)
	if onError {
		return onError
	}
	isError, _, _ := rawExportDsk(d, filePath, dskFilepath, track, sector, length)
	return isError
}
