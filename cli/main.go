package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/jeromelesaux/dsk/amsdos"
	"github.com/jeromelesaux/dsk/cli/action"
	"github.com/jeromelesaux/dsk/cli/msg"
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
	rawimport    = flag.Bool("rawimport", false, "Perform a raw import of an AMSDOS file. Requires '-dsk', '-track', and '-sector' options. \n\t\tCopies the file directly starting from the specified track and sector. e.g.: dsk -dsk mydskfile.dsk -put file.bin -rawimport -track 1 -sector 0")
	rawexport    = flag.Bool("rawexport", false, "Perform a raw export of an AMSDOS file. Requires '-dsk', '-track', '-sector', and '-size' options. \n\t\tExtracts the file content from the specified track and sector up to the given size. e.g.: dsk -dsk mydskfile.dsk -get file.bin -rawexport -track 1 -sector 0 -size 16384")
	size         = flag.Int("size", 0, "Size of data to extract for 'rawexport'. See 'rawexport' for details.")
	autotest     = flag.Bool("autotest", false, "Run all available tests.")
	autoextract  = flag.String("autoextract", "", "Extract all DSK files from a specified folder.")
	snaVersion   = flag.Int("snaversion", 1, "Specify the SNA version (1 or 2 available).")
	quiet        = flag.Bool("quiet", false, "Suppress unnecessary output (useful for scripting).")
	stdoutOpt    = flag.Bool("stdout", false, "To redirect to stdout when using get file")
	hidden       = flag.Bool("hide", false, "hide the imported file")
	removeHeader = flag.Bool("removeheader", false, "remove amsdos header from exported file")
	appVersion   = "0.35"
	version      = flag.Bool("version", false, "Display the application version and exit.")
)

func main() {
	var cmdRunned bool

	flag.Usage = sampleUsage
	flag.Parse()

	fd := action.NewAmsdosFileDescriptor().WithUser(uint16(*user))

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
			msg.ExitOnError(err.Error(), fmt.Sprintf("file %s issue", *put))
		}
		hasHeader, headerInf := amsdos.CheckAmsdos(content)
		if hasHeader {
			fd.Type = action.AmsdosTypeBinary
			if *executeAddress == "" {
				fd.Exec = headerInf.Exec
			}
			if *loadingAddress == "" {
				fd.Load = headerInf.Address
			}
		}
	}

	if *loadingAddress != "" || *executeAddress != "" {
		fd.AddHeader = true
		fd.Type = action.AmsdosTypeBinary
	}

	if *autoextract != "" {
		files, err := fs.ReadDir(os.DirFS("/"), *autoextract)
		if err != nil {
			msg.ExitOnError(err.Error(), "Please check your folder path")
		}
		for _, file := range files {
			if !file.IsDir() {
				if strings.ToUpper(path.Ext(file.Name())) == ".DSK" {
					dskfolderPath := *autoextract + string(filepath.Separator) + strings.Replace(file.Name(), path.Ext(file.Name()), "", -1)
					dskFilepath := *autoextract + string(filepath.Separator) + file.Name()
					err = os.Mkdir(dskfolderPath, os.ModePerm)
					if err != nil && !errors.Is(err, os.ErrExist) {
						msg.ExitOnError(err.Error(), "Please check your folder path")
					}

					d, err, msg, _ := action.OpenDsk(dskFilepath, action.DskDescriptor{Path: dskFilepath, Sector: *sector, Track: *track, Head: *heads}, *quiet)
					if err {
						fmt.Fprintf(os.Stderr, "Error while opening file %s error :%s\n", dskFilepath, msg)
					}
					err, msg, _ = action.GetFileDsk(d, "*", dskFilepath, dskfolderPath, *removeHeader, *quiet)
					if err {
						fmt.Fprintf(os.Stderr, "Error while writing file %s in folder %s error :%s\n", dskFilepath, dskfolderPath, msg)
					}
				}
			}
		}
		os.Exit(0)
	}

	if *executeAddress != "" {
		var err error
		fd.Exec, err = utils.ParseHex16(*executeAddress)
		if err != nil {
			msg.ExitOnError(err.Error(), "Invalid execute address format")
		}
	}
	if *loadingAddress != "" {
		var err error
		fd.Load, err = utils.ParseHex16(*loadingAddress)
		if err != nil {
			msg.ExitOnError(err.Error(), "Invalid loading address format")
		}
	}
	if !*quiet {
		fmt.Fprintf(os.Stderr, "DSK cli version [%s]\nMade by Sid (ImpAct)\n", appVersion)
	}
	// gestion des SNAs
	if *snaPath != "" {
		if *info != "" {
			cmdRunned = true
			isError, m, hint := action.InfoSna(*snaPath)
			if isError {
				msg.ExitOnError(m, hint)
			}
			os.Exit(0)
		}
		if *format {
			cmdRunned = true
			isError, m, hint := action.FormatSna(*snaPath, *snaVersion)
			if isError {
				msg.ExitOnError(m, hint)
			}
		}
		if *hexa != "" {
			cmdRunned = true
			sna, err := sna.ReadSna(*snaPath)
			if err != nil {
				msg.ExitOnError(err.Error(), "Check your sna path")
			}
			content := sna.Hexadecimal()
			fmt.Println(dsk.DisplayHex([]byte(content), 16))
			os.Exit(0)
		}
		if *put != "" {
			cmdRunned = true
			if *put != "" {
				crtc := sna.UM6845R
				if *cpcType > 3 {
					crtc = sna.ASIC_6845
				}
				if err := sna.ImportInSna(*put, *snaPath, fd.Exec, uint8(*screenMode), sna.CPCType(*cpcType), crtc, *snaVersion); err != nil {
					fmt.Fprintf(os.Stderr, "Error while trying to import file (%s) in new sna (%s) error: %v\n",
						*put,
						*snaPath,
						err)
					os.Exit(1)
				}
				os.Exit(0)
			} else {
				fmt.Fprintf(os.Stderr, "Missing input (argument -put) file to import in sna file (%s)\n", *snaPath)
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
					d, isError, m, hint := action.OpenDsk(*dskPath, action.DskDescriptor{Path: *dskPath, Sector: *sector, Track: *track, Head: *heads}, *quiet)
					if isError {
						msg.ExitOnError(m, hint)
					}
					isError, m, hint = action.RawImportDataInDsk(d, *get, action.DskDescriptor{Path: *dskPath, Sector: *sector, Track: *track, Head: *heads}, content, *quiet)
					if isError {
						msg.ExitOnError(m, hint)
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
		isError, m, hint := action.FormatDsk(action.DskDescriptor{Path: *dskPath, Sector: *sector, Track: *track, Head: *heads}, *vendorFormat, *dataFormat, *force)
		if isError {
			msg.ExitOnError(m, hint)
		}
	}

	if *dskPath != "" {
		d, isError, m, hint := action.OpenDsk(*dskPath, action.DskDescriptor{Path: *dskPath, Sector: *sector, Track: *track, Head: *heads}, *quiet)
		if isError {
			msg.ExitOnError(m, hint)
		}
		if *list {
			isError, m, hint := action.ListDsk(d, *dskPath)
			if isError {
				msg.ExitOnError(m, hint)
			}
		}

		if *analyse {
			cmdRunned = true
			isError, m, hint := action.AnalyseDsk(d, *dskPath)
			if isError {
				msg.ExitOnError(m, hint)
			}
		}

		if *info != "" {
			cmdRunned = true
			isError, m, hint := action.FileinfoDsk(d, *info)
			if isError {
				msg.ExitOnError(m, hint)
			}
		}

		if *hexa != "" {
			cmdRunned = true
			isError, m, hint := action.DisplayHexaFileDsk(d, *hexa)
			if isError {
				msg.ExitOnError(m, hint)
			}
		}

		if *disassemble != "" {
			cmdRunned = true
			isError, m, hint := action.DesassembleFileDsk(d, *disassemble)
			if isError {
				msg.ExitOnError(m, hint)
			}
		}

		if *ascii != "" {
			cmdRunned = true
			isError, m, hint := action.AsciiFileDsk(d, *ascii, *stdoutOpt)
			if isError {
				msg.ExitOnError(m, hint)
			}
		}

		if *basic != "" {
			cmdRunned = true
			action.ListBasic(d, *basic)
		}

		if *get != "" {
			cmdRunned = true
			directory, err := os.Getwd()
			if err != nil {
				msg.ExitOnError(err.Error(), "Please use autoextract option")
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
				isError, m, hint := action.GetFileDsk(d, *get, *dskPath, directory, *removeHeader, *quiet)
				if isError {
					msg.ExitOnError(m, hint)
				}
			}
		}

		if *rawimport {
			cmdRunned = true
			isError, m, hint := action.RawImportDsk(d, *put, action.DskDescriptor{Path: *dskPath, Sector: *sector, Track: *track, Head: *heads}, *quiet)
			if isError {
				msg.ExitOnError(m, hint)
			}
		}

		if *rawexport {
			cmdRunned = true
			isError, m, hint := action.RawExportDsk(d, *put, action.DskDescriptor{Path: *dskPath, Sector: *sector, Head: *heads, Track: *track}, *size, *quiet)
			if isError {
				msg.ExitOnError(m, hint)
			}
		}

		if *put != "" {
			cmdRunned = true
			fd.Path = *put
			isError, m, hint := action.PutFileDsk(d, *dskPath, fd, *hidden, *force, *quiet)
			if isError {
				msg.ExitOnError(m, hint)
			}
		}

		if *remove != "" {
			cmdRunned = true
			isError, m, hint := action.RemoveFileDsk(d, *dskPath, *remove)
			if isError {
				msg.ExitOnError(m, hint)
			}
		}

		// no arguments commands
		if !cmdRunned {
			cmdRunned = true
			isError, m, hint = action.ListDsk(d, *dskPath)
			if isError {
				msg.ExitOnError(m, hint)
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
				msg.ExitOnError(fmt.Sprintf("Cannot read file %s error :%v\n", *basic, err), "cannot read the file content")
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
				msg.ExitOnError(fmt.Sprintf("Cannot read file %s error :%v\n", *disassemble, err), "cannot read the file content")
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
				msg.ExitOnError(fmt.Sprintf("Cannot read file %s error :%v\n", *hexa, err), "cannot read the file content")
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
				msg.ExitOnError(fmt.Sprintf("Cannot read file %s error :%v\n", *info, err), "cannot read the file content")
			}
			isAmsdos, header := amsdos.CheckAmsdos(content)
			if !isAmsdos {
				msg.ExitOnError(fmt.Sprintf("File (%s) does not contain amsdos header.\n", *info), "may be a ascii file")
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
		if fd.AddHeader {
			if *get == "" {
				msg.ExitOnError("Error no file is set\n", "set your amsdos file with option like dsk -get hello.bin")
			}
			content, err := os.ReadFile(*get)
			if err != nil {
				msg.ExitOnError(fmt.Sprintf("Cannot read file %s error :%v\n", *get, err), "cannot read the file content")
			}
			informations := fmt.Sprintf("execute address [#%.4x], loading address [#%.4x]\n", fd.Exec, fd.Load)
			isAmsdos, header := amsdos.CheckAmsdos(content)
			if isAmsdos {
				msg.ExitOnError("The file already contains an amsdos header", "Check your file")
			}
			filename := dsk.GetNomAmsdos(*get)
			header.Size = uint16(len(content))
			header.Size2 = uint16(len(content))
			copy(header.Filename[:], []byte(filename[0:12]))
			header.Address = fd.Load
			header.Exec = fd.Exec
			// Il faut recalculer le checksum en comptant es adresses !
			header.Checksum = header.ComputedChecksum16()
			var rbuff bytes.Buffer
			err = binary.Write(&rbuff, binary.LittleEndian, header)
			if err != nil {
				msg.ExitOnError("error while writing header : "+err.Error(), "Check your file")
			}
			err = binary.Write(&rbuff, binary.LittleEndian, content)
			if err != nil {
				msg.ExitOnError("error while writing header : "+err.Error(), "Check your file")
			}

			if err := utils.Save(*get, rbuff.Bytes()); err != nil {
				msg.ExitOnError(fmt.Sprintf("Error while writing data in file [%s] error:%v\n", *get, err), "Check your dsk  with option -dsk yourdsk.dsk -analyze")
			}

			msg.ResumeAction("none", "add amsdos header", *get, informations, *quiet)
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
	if utils.SaveFile(generateSliceByte(512), "test.bin") {
		fmt.Printf("KO\n")
		testsOnError++
	} else {
		fmt.Printf("OK\n")
	}

	testsDone++
	if putFileBinaryDataTest("test.bin", dskpath, false) {
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
	if utils.SaveFile(generateSliceByte(2048), "test3.bin") {
		fmt.Printf("KO\n")
		testsOnError++
	} else {
		fmt.Printf("OK\n")
	}

	testsDone++
	if putFileBinaryDataTest("test3.bin", dskpath, *hidden) {
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
	onError, _, _ := action.FormatDsk(action.DskDescriptor{Path: dskFilepath, Sector: sector, Track: track, Head: heads}, true, false, false)
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
	onError, _, _ := action.FormatDsk(action.DskDescriptor{Path: dskFilepath, Sector: sector, Track: track, Head: heads}, true, false, true)
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
	onError, _, _ := action.FormatDsk(action.DskDescriptor{Path: dskFilepath, Sector: sector, Track: track, Head: heads, Type: dskType}, false, true, false)
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
	onError, _, _ := action.FormatDsk(action.DskDescriptor{Path: dskFilepath, Sector: sector, Track: track, Head: heads, Type: dskType}, false, true, false)
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
	onError, _, _ := action.FormatDsk(action.DskDescriptor{Path: dskFilepath, Sector: sector, Track: track, Head: heads, Type: dskType}, false, true, true)
	return onError
}

func formatSnaTest(snafilepath string) bool {
	resetArguments()
	fmt.Printf("Format sna image file ")
	onError, _, _ := action.FormatSna(snafilepath, 2)
	return onError
}

func infoSnaTest(snafilepath string) bool {
	resetArguments()
	fmt.Printf("Get information from sna image file ")
	*info = snafilepath
	onError, _, _ := action.InfoSna(snafilepath)
	return onError
}

func listVendorTest(dskFilepath string) bool {
	resetArguments()
	fmt.Printf("list vendor format ")
	*vendorFormat = true
	*format = true
	d, onError, _, _ := action.OpenDsk(dskFilepath, action.DskDescriptor{Path: dskFilepath, Sector: *sector, Track: *track, Head: *heads}, *quiet)
	if onError {
		return onError
	}
	onError, _, _ = action.ListDsk(d, dskFilepath)
	return onError
}

func listDataTest(dskFilepath string) bool {
	resetArguments()
	fmt.Printf("list vendor format ")
	*format = true
	d, onError, _, _ := action.OpenDsk(dskFilepath, action.DskDescriptor{Path: dskFilepath, Sector: *sector, Track: *track, Head: *heads}, *quiet)
	if onError {
		return onError
	}
	onError, _, _ = action.ListDsk(d, dskFilepath)
	return onError
}

func analyseDataTest(dskFilepath string) bool {
	resetArguments()
	fmt.Printf("list vendor format ")
	*analyse = true
	d, onError, _, _ := action.OpenDsk(dskFilepath, action.DskDescriptor{Path: dskFilepath, Sector: *sector, Track: *track, Head: *heads}, *quiet)
	if onError {
		return onError
	}
	onError, _, _ = action.AnalyseDsk(d, dskFilepath)
	return onError
}

func parsing16bitsRasmAnnotation() bool {
	fmt.Printf("Parsing value rasm annotation #C000 ")
	v, err := utils.ParseHex16("#C000")
	if err != nil {
		return false
	}
	return v == 0xC000
}

func parsing16bitsCAnnotation() bool {
	fmt.Printf("Parsing value c annotation 0xC000 ")
	v, err := utils.ParseHex16("0xC000")
	if err != nil {
		return false
	}
	return v == 0xC000
}

func parsing16bitsIntegerAnnotation() bool {
	fmt.Printf("Parsing value integer annotation 49152 ")
	v, err := utils.ParseHex16("49152")
	if err != nil {
		return false
	}
	return v == 0xC000
}

func parsing8bitsRasmAnnotation() bool {
	fmt.Printf("Parsing value c annotation #D0 ")
	v, err := utils.ParseHex16("#D0")
	if err != nil {
		return false
	}
	return v == 0xD0
}

func putFileBinaryDataTest(filePath, dskFilepath string, hide bool) bool {
	*put = dskFilepath
	d, onError, _, _ := action.OpenDsk(dskFilepath, action.DskDescriptor{Path: dskFilepath, Sector: *sector, Track: *track, Head: *heads}, *quiet)
	if onError {
		return onError
	}
	fd := action.AmsdosFileDescriptor{
		Path: filePath,
		Exec: 0x800,
		Load: 0x800,
		User: uint16(*user),
		Type: action.AmsdosTypeBinary,
	}
	isError, _, _ := action.PutFileDsk(d, dskFilepath, fd, hide, false, *quiet)
	return isError
}

func getFileBinaryDataTest(filePath, dskFilepath string) bool {
	*get = dskFilepath
	// fileType := "binary"
	d, onError, _, _ := action.OpenDsk(dskFilepath, action.DskDescriptor{Path: dskFilepath, Sector: *sector, Track: *track, Head: *heads}, *quiet)
	if onError {
		return onError
	}
	isError, _, _ := action.GetFileDsk(d, filePath, dskFilepath, "", false, *quiet)
	return isError
}

func removeFileBinaryDataTest(fileInDsk, dskFilepath string) bool {
	*remove = fileInDsk
	d, onError, _, _ := action.OpenDsk(dskFilepath, action.DskDescriptor{Path: dskFilepath, Sector: *sector, Track: *track, Head: *heads}, *quiet)
	if onError {
		return onError
	}
	isError, _, _ := action.RemoveFileDsk(d, dskFilepath, fileInDsk)
	return isError
}

func asciiFileBinaryDataTest(fileInDsk, dskFilepath string) bool {
	*ascii = fileInDsk
	d, onError, _, _ := action.OpenDsk(dskFilepath, action.DskDescriptor{Path: dskFilepath, Sector: *sector, Track: *track, Head: *heads}, *quiet)
	if onError {
		return onError
	}
	isError, _, _ := action.AsciiFileDsk(d, fileInDsk, false)
	return isError
}

func desassembleFileBinaryDataTest(fileInDsk, dskFilepath string) bool {
	*disassemble = fileInDsk
	d, onError, _, _ := action.OpenDsk(dskFilepath, action.DskDescriptor{Path: dskFilepath, Sector: *sector, Track: *track, Head: *heads}, *quiet)
	if onError {
		return onError
	}
	isError, _, _ := action.DesassembleFileDsk(d, fileInDsk)
	return isError
}

func hexaFileBinaryDataTest(fileInDsk, dskFilepath string) bool {
	*hexa = fileInDsk
	d, onError, _, _ := action.OpenDsk(dskFilepath, action.DskDescriptor{Path: dskFilepath, Sector: *sector, Track: *track, Head: *heads}, *quiet)
	if onError {
		return onError
	}
	isError, _, _ := action.DisplayHexaFileDsk(d, fileInDsk)
	return isError
}

func rawImportFileBinaryDataTest(filePath, dskFilepath string) bool {
	*rawimport = true
	sector := 0
	track := 1
	d, onError, _, _ := action.OpenDsk(dskFilepath, action.DskDescriptor{Path: dskFilepath, Sector: sector, Track: track, Head: *heads}, *quiet)
	if onError {
		return onError
	}
	isError, _, _ := action.RawImportDsk(d, filePath, action.DskDescriptor{Path: dskFilepath, Sector: sector, Track: track, Head: *heads}, *quiet)
	return isError
}

func rawExportFileBinaryDataTest(filePath, dskFilepath string, length int) bool {
	*rawexport = true
	sector := 0
	track := 1
	d, onError, _, _ := action.OpenDsk(dskFilepath, action.DskDescriptor{Path: dskFilepath, Sector: sector, Track: track, Head: *heads}, *quiet)
	if onError {
		return onError
	}
	isError, _, _ := action.RawExportDsk(d, filePath, action.DskDescriptor{Path: dskFilepath, Sector: sector, Track: track, Head: *heads}, length, *quiet)
	return isError
}
