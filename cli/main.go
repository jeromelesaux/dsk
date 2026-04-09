package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"

	"github.com/jeromelesaux/dsk/cli/action"
	"github.com/jeromelesaux/dsk/cli/msg"
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
	hidden       = flag.Bool("hide", false, "Hide the imported file")
	removeHeader = flag.Bool("removeheader", false, "Remove amsdos header from exported file")
	hfeFilepath  = flag.String("hfe", "", "Path to the HFE file to handle.")

	appVersion = "0.36"
	version    = flag.Bool("version", false, "Display the application version and exit.")
)

func main() {
	var cmdRunned bool

	flag.Usage = sampleUsage
	flag.Parse()

	fd := action.NewAmsdosFileDescriptor().
		WithUser(uint16(*user)).
		AddExec(*executeAddress).
		AddLoad(*loadingAddress).
		WithAddHeader(*executeAddress != "" || *loadingAddress != "").
		WithPaths(*put, *get, *basic, *hexa, *disassemble, *ascii, *remove, *info)

	opts := action.NewOptions().
		WithQuiet(*quiet).
		WithFormat(*format).
		WithForce(*force).
		WithAnalyze(*analyse).
		WithDataFormat(*dataFormat).
		WithVendorFormat(*vendorFormat).
		WithStdout(*stdoutOpt).
		WithHidden(*hidden).
		WithRemoveHeader(*removeHeader).
		WithRawImport(*rawimport).
		WithRawExport(*rawexport)

	acts := action.NewDskTasks().
		WithActionListDsk(*dskPath, true).
		WithActionFormatDsk(*dskPath, *format).
		WithActionDisplayHexaFileDsk(*dskPath, *hexa != "").
		WithActionDesassembleFileDsk(*dskPath, *disassemble != "").
		WithActionListBasic(*dskPath, *basic != "").
		WithActionAnalyseDsk(*dskPath, *analyse).
		WithActionPutFileDsk(*dskPath, *put != "").
		WithActionRemoveFileDsk(*dskPath, *remove != "").
		WithActionGetFileDsk(*dskPath, *get != "").
		WithActionAsciiFileDsk(*dskPath, *ascii != "").
		WithActionRawExportDsk(*dskPath, *rawexport).
		WithActionRawImportDsk(*dskPath, *rawimport).
		WithActionFileinfoDsk(*dskPath, *info != "").
		WithActionGetAllFileDsk(*autoextract, *autoextract != "").
		WithActionHFEFile(*hfeFilepath, *hfeFilepath != "")

	desc := action.NewDskDescriptor().
		WithSector(*sector).
		WithTrack(*track).
		WithHead(*heads).
		WithType(*dskType)

	dskAct := action.NewAction(*dskPath, *autoextract).
		WithOptions(*opts).
		WithAmsdosFileDescriptor(*fd).
		WithDskDescriptor(*desc).
		WithDskActions(acts)

	snaAct := action.NewSnaAction(*snaPath).
		WithCPCType(*cpcType).
		WithScreemode(*screenMode).
		WithVersion(*snaVersion).
		WithSnaFormatAction(*format).
		WithSnaInfoAction(*info != "").
		WithSnaGetAction(*get != "").
		WithSnaPutAction(*put != "").
		WithSnaHexaListAction(*hexa != "").
		WithFiles(*get, *put)

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

	if !*quiet {
		fmt.Fprintf(os.Stderr, "DSK cli version [%s]\nMade by Sid (ImpAct)\n", appVersion)
	}

	if snaAct.SnaIsSet() {
		onErr, message, hint := snaAct.DoSnaActions()
		if onErr {
			msg.ExitOnError(message, hint)
		}
		os.Exit(0)
	}

	if dskAct.DskIsSet() {
		onErr, message, hint := dskAct.DoDskActions()
		if onErr {
			msg.ExitOnError(message, hint)
		}
		os.Exit(0)
	}

	onErr, message, hint := dskAct.DoFileActions()
	if onErr {
		msg.ExitOnError(message, hint)
	}
	os.Exit(0)

	// gestion des SNAs

	// if *dskPath != "" {
	// 	content, err := sna.ExportFromSna(*snaPath)
	// 	if err != nil {
	// 		fmt.Fprintf(os.Stderr, "Error while trying to import file (%s) in new sna (%s) error: %v\n",
	// 			*get,
	// 			*snaPath,
	// 			err)
	// 		os.Exit(1)
	// 	}
	// 	d, isError, m, hint := action.OpenDsk(*dskPath, action.DskDescriptor{Path: *dskPath, Sector: *sector, Track: *track, Head: *heads}, *quiet)
	// 	if isError {
	// 		msg.ExitOnError(m, hint)
	// 	}
	// 	isError, m, hint = action.RawImportDataInDsk(d, *get, action.DskDescriptor{Path: *dskPath, Sector: *sector, Track: *track, Head: *heads}, content, *quiet)
	// 	if isError {
	// 		msg.ExitOnError(m, hint)
	// 	}

	// } else {
	// 	fmt.Fprintf(os.Stderr, "Missing input file to import in sna file (%s)\n", *snaPath)
	// }

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
	onError, _, _ := action.FormatDsk(dskFilepath, action.DskDescriptor{Path: dskFilepath, Sector: sector, Track: track, Head: heads}, true, false, false)
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
	onError, _, _ := action.FormatDsk(dskFilepath, action.DskDescriptor{Path: dskFilepath, Sector: sector, Track: track, Head: heads}, true, false, true)
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
	onError, _, _ := action.FormatDsk(dskFilepath, action.DskDescriptor{Path: dskFilepath, Sector: sector, Track: track, Head: heads, Type: dskType}, false, true, false)
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
	onError, _, _ := action.FormatDsk(dskFilepath, action.DskDescriptor{Path: dskFilepath, Sector: sector, Track: track, Head: heads, Type: dskType}, false, true, false)
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
	onError, _, _ := action.FormatDsk(dskFilepath, action.DskDescriptor{Path: dskFilepath, Sector: sector, Track: track, Head: heads, Type: dskType}, false, true, true)
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
