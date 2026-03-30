package action

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jeromelesaux/dsk/amsdos"
	"github.com/jeromelesaux/dsk/cli/msg"
	"github.com/jeromelesaux/dsk/dsk"
	"github.com/jeromelesaux/dsk/utils"
)

type AmsdosType string

var (
	AmsdosTypeAscii  AmsdosType = "ascii"
	AmsdosTypeBinary AmsdosType = "binary"
)

type DskDescriptor struct {
	Sector int
	Track  int
	Head   int
	Path   string
	Type   int
}

type AmsdosFileDescriptor struct {
	Path      string
	Exec      uint16
	Load      uint16
	User      uint16
	Type      AmsdosType
	AddHeader bool
}

func ListDsk(d dsk.DSK, dskPath string) (onError bool, message, hint string) {
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

func FormatDsk(desc DskDescriptor, vendorFormat bool, dataFormat, force bool) (onError bool, message, hint string) {
	_, err := os.Stat(desc.Path)
	if err == nil {
		if !force {
			return true, fmt.Sprintf("Error file (%s) already exists", desc.Path), "Use option -force to avoid this message"
		}
	}
	f, err := os.Create(desc.Path)
	if err != nil {
		return true, fmt.Sprintf("Error while write file (%s) error %v", desc.Path, err), "Check your dsk file path."
	}
	defer f.Close()
	fmt.Fprintf(os.Stderr, "Formating number of sectors (%d), tracks (%d), head number (%d)\n", desc.Sector, desc.Track, desc.Head)
	var dskFile *dsk.DSK
	if dataFormat {
		dskFile = dsk.FormatDsk(uint8(desc.Sector), uint8(desc.Track), uint8(desc.Head), dsk.DataFormat, desc.Type)
	} else {
		if vendorFormat {
			dskFile = dsk.FormatDsk(uint8(desc.Sector), uint8(desc.Track), uint8(desc.Head), dsk.VendorFormat, desc.Type)
		}
	}
	if err := dskFile.Write(f); err != nil {
		return true, fmt.Sprintf("Error while write file (%s) error %v\n", desc.Path, err), "Check your dsk file with option -dsk yourdsk.dsk -analyze"
	}
	return false, "", ""
}

func DisplayHexaFileDsk(d dsk.DSK, filepath string) (onError bool, message, hint string) {
	if filepath == "" {
		return true, "amsdosfile option is empty, set it.", "dsk -dsk output.dsk -hex hello.bin"
	}

	content, fileSize, err := GetContentDsk(d, filepath)
	if err != nil {
		return true, err.Error(), "Check your dsk file with option -dsk yourdsk.dsk -analyze"
	}

	fmt.Println(dsk.DisplayHex(content[0:fileSize], 16))
	return false, "", ""
}

func GetContentDsk(d dsk.DSK, filepath string) ([]byte, int, error) {
	if filepath == "" {
		return nil, 0, fmt.Errorf("amsdosfile option is empty, set it.")
	}
	amsdosFile := dsk.GetNomDir(filepath)
	indice := d.FileExists(amsdosFile)
	if indice == dsk.NOT_FOUND {
		return nil, 0, fmt.Errorf("File %s does not exist", filepath)
	}
	content, size, err := d.ViewFile(indice)
	if err != nil {
		return nil, size, fmt.Errorf("Error while getting file in dsk error :%v", err)
	}
	return content, size, nil
}

func DesassembleFileDsk(d dsk.DSK, filepath string) (onError bool, message, hint string) {
	if filepath == "" {
		return true, "amsdosfile option is empty, set it.", "dsk -dsk output.dsk -desassemble hello.bin"
	}

	content, filesize, err := GetContentDsk(d, filepath)
	if err != nil {
		return true, err.Error(), "Check your dsk file with option -dsk yourdsk.dsk -analyze"
	}
	var address uint16
	isAmsdos, header := amsdos.CheckAmsdos(content)
	if isAmsdos {
		address = header.Exec
	}

	fmt.Println(utils.Desass(content[0:filesize], uint16(filesize), address))
	return false, "", ""
}

func ListBasic(d dsk.DSK, filepath string) (onError bool, message, hint string) {
	content, filesize, err := GetContentDsk(d, filepath)
	if err != nil {
		return true, err.Error(), "Check your dsk file with option -dsk yourdsk.dsk -analyze"
	}
	hasAmsdos, _ := amsdos.CheckAmsdos(content)
	if hasAmsdos {
		fmt.Fprintf(os.Stderr, "File %s filesize :%d octets\n", filepath, filesize)
		fmt.Fprintf(os.Stdout, "%s", utils.Basic(content, uint16(filesize), true))
	} else {
		fmt.Fprintf(os.Stderr, "File %s filesize :%d octets\n", filepath, len(content))
		fmt.Fprintf(os.Stdout, "%s", content)
	}

	return false, "", ""
}

func AnalyseDsk(d dsk.DSK, dskPath string) (onError bool, message, hint string) {
	if err := d.CheckDsk(); err != nil {
		return true, fmt.Sprintf("Error while read dsk file (%s) error %v\n", dskPath, err), "Check your dsk file path or Check your dsk file with option -dsk yourdsk.dsk -analyze"
	}
	fmt.Fprintf(os.Stderr, "Dsk file (%s)\n", dskPath)
	entry := d.Entry
	fmt.Fprintf(os.Stderr, "Dsk entry %s\n", entry.ToString())
	return false, "", ""
}

func PutFileDsk(d dsk.DSK, dskPath string, desc AmsdosFileDescriptor, hide, force, quiet bool) (onError bool, message, hint string) {
	if desc.Path == "" {
		msg.ExitOnError("amsdosfile option is empty, set it.", "dsk -dsk output.dsk -put hello.bin -exec \"#1000\" -load 500")
	}
	amsdosFile := dsk.GetNomDir(desc.Path)
	indice := d.FileExists(amsdosFile)
	if indice != dsk.NOT_FOUND && !force {
		msg.ExitOnError(fmt.Sprintf("File %s already exists\n", desc.Path), "use -force to force file put")
	} else {
		if indice != dsk.NOT_FOUND && force {
			// suppress file
			err := d.RemoveFile(uint8(indice))
			if err != nil {
				msg.ExitOnError(fmt.Sprintf("error while removing file %v", err), "check your dsk content")
			}
		}
		switch desc.Type {
		case AmsdosTypeAscii:
			informations := fmt.Sprintf("execute address [#%.4x], loading address [#%.4x]\n", desc.Exec, desc.Load)
			if err := d.PutFile(desc.Path, dsk.MODE_ASCII, 0, 0, desc.User, false, false, hide); err != nil {
				return true, fmt.Sprintf("Error while inserted file (%s) in dsk (%s) error :%v\n", desc.Path, dskPath, err), "Check your dsk  with option -dsk yourdsk.dsk -analyze"
			}
			msg.ResumeAction(dskPath, "put ascii", desc.Path, informations, quiet)
		case AmsdosTypeBinary:
			informations := fmt.Sprintf("execute address [#%.4x], loading address [#%.4x]\n", desc.Exec, desc.Load)
			if err := d.PutFile(desc.Path, dsk.MODE_BINAIRE, desc.Load, desc.Exec, desc.User, false, false, hide); err != nil {
				return true, fmt.Sprintf("Error while inserted file (%s) in dsk (%s) error :%v\n", desc.Path, dskPath, err), "Check your dsk  with option -dsk yourdsk.dsk -analyze"
			}
			msg.ResumeAction(dskPath, "put binary", desc.Path, informations, quiet)
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

func RemoveFileDsk(d dsk.DSK, dskPath, fileInDsk string) (onError bool, message, hint string) {
	if fileInDsk == "" {
		return true, "amsdosfile option is empty, set it.", "dsk -dsk output.dsk -remove hello.bin"
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

func GetFileDsk(d dsk.DSK, fileInDsk, dskPath, directory string, removeHeader, quiet bool) (onError bool, message, hint string) {
	if fileInDsk == "" {
		return true, "amsdosfile option is empty, set it.", "dsk -dsk output.dsk -get hello.bin"
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
				if removeHeader {
					isAmsdos, _ := amsdos.CheckAmsdos(content)
					if isAmsdos {
						content = content[256:] // Remove the first 256 bytes (AMS/DOS header)
					}
				}
				_, err = af.Write(content)
				if err != nil {
					return true, fmt.Sprintf("Error while copying content in file (%s) error %v\n", filename, err), "Check your dsk  with option -dsk yourdsk.dsk -analyze"
				}
				af.Close()
				informations := fmt.Sprintf("Extract file [%s] Indice in DSK [%d] is saved\n", fPath, indice)
				msg.ResumeAction(dskPath, "get amsdosfile", fileInDsk, informations, quiet)
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
			if removeHeader {
				isAmsdos, _ := amsdos.CheckAmsdos(content)
				if isAmsdos {
					content = content[256:] // Remove the first 256 bytes (AMS/DOS header)
				}
			}
			_, err = af.Write(content)
			if err != nil {
				return true, fmt.Sprintf("Error while copying content in file (%s) error %v\n", filename, err), " Check your dsk  with option -dsk yourdsk.dsk -analyze"
			}
			informations := fmt.Sprintf("Extract file [%s] Indice in DSK [%d] is saved\n", filename, indice)
			msg.ResumeAction(dskPath, "get amsdosfile", filename, informations, quiet)
		}
	}
	return false, "", ""
}

func AsciiFileDsk(d dsk.DSK, fileInDsk string, isSdtout bool) (onError bool, message, hint string) {
	if fileInDsk == "" {
		return true, "amsdosfile option is empty, set it.", "dsk -dsk output.dsk -ascii hello.txt"
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

func RawExportDsk(d dsk.DSK, fileInDsk string, desc DskDescriptor, size int, quiet bool) (onError bool, message, hint string) {
	if fileInDsk == "" {
		return true, "amsdosfile option is empty, set it.", "dsk -dsk output.dsk -put hello.bin -rawimport -track 1 -sector 0"
	}

	if desc.Track == 39 {
		fmt.Fprintf(os.Stdout, "Warning the starting track is set as default : [%d]\n", desc.Track)
	}
	if desc.Sector == 9 {
		fmt.Fprintf(os.Stdout, "Warning the starting sector is set as default : [%d]\n", desc.Sector)
	}
	if size == 0 {
		fmt.Fprintf(os.Stdout, "Warning the size is set as default : [%d]\n", size)
	}

	fmt.Fprintf(os.Stdout, "Writing file content starting from track [%d] sector [%d] to file  [%s] in dsk [%s] size [%d]\n",
		desc.Track,
		desc.Sector,
		fileInDsk,
		desc.Path,
		size,
	)
	endedTrack, endedSector, content := d.ExtractRawFile(uint16(size), desc.Track, desc.Sector)

	if err := utils.Save(fileInDsk, content); err != nil {
		return true, fmt.Sprintf("Error while write file (%s) error %v\n", fileInDsk, err), "Check your dsk  with option -dsk yourdsk.dsk -analyze"
	}
	informations := fmt.Sprintf("raw extract to file [%s] size [%d] starting at track [%d] sector [%d] and ending at track [%d] sector [%d]",
		fileInDsk,
		size,
		desc.Track,
		desc.Sector,
		endedTrack,
		endedSector)
	msg.ResumeAction(desc.Path, "raw export ", fileInDsk, informations, quiet)
	return false, "", ""
}

func OpenDsk(osFile string, desc DskDescriptor, quiet bool) (d dsk.DSK, onError bool, message, hint string) {
	if _, err := os.Stat(osFile); errors.Is(err, os.ErrNotExist) {
		onError, msgErr, hint := FormatDsk(desc, true, false, false)
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
		return d, true, fmt.Sprintf("Error while read dsk file (%s) error %v\n", desc.Path, err), "Check your dsk file path or Check your dsk file with option -dsk yourdsk.dsk -analyze"
	}
	if err := d.CheckDsk(); err != nil {
		return d, true, fmt.Sprintf("Error while read dsk file (%s) error %v\n", osFile, err), "Check your dsk file path or Check your dsk file with option -dsk yourdsk.dsk -analyze"
	}
	if !quiet {
		fmt.Fprintf(os.Stderr, "Dsk file (%s)\n", osFile)
	}
	return d, false, "", ""
}

func FileinfoDsk(d dsk.DSK, fileInDsk string) (onError bool, message, hint string) {
	if fileInDsk == "" {
		return true, "amsdosfile option is empty, set it.", "usage sample : dsk -dsk output.dsk hello.bin -info "
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
			entry, err := d.GetInfoDirEntry(uint8(indice))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error while getting file entry in dsk error :%v\n", err)
			}
			fmt.Fprintf(os.Stdout, "Amsdos informations :\n\tAscii file\n\tFilename:%s\n\tPage Number:#%d\n\tUser:%d\n",
				string(entry.Nom[:])+"."+string(entry.Ext[:]),
				entry.NbPages,
				entry.User,
			)
		} else {
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
	return false, "", ""
}

func RawImportDataInDsk(d dsk.DSK, fileInDsk string, desc DskDescriptor, content []byte, quiet bool) (onError bool, message, hint string) {
	if desc.Track == 39 {
		fmt.Fprintf(os.Stdout, "Warning the starting track is set as default : [%d]\n", desc.Track)
	}
	if desc.Sector == 9 {
		fmt.Fprintf(os.Stdout, "Warning the starting sector is set as default : [%d]\n", desc.Sector)
	}
	endedTrack, endedSector, err := d.CopyRawFile(content, uint16(len(content)), desc.Track, desc.Sector)
	if err != nil {
		return true, fmt.Sprintf("Cannot write file %s error :%v\n", fileInDsk, err), "Check your file path"
	}
	f, err := os.Create(desc.Path)
	if err != nil {
		return true, fmt.Sprintf("Error while write file (%s) error %v\n", desc.Path, err), "Check your dsk path file"
	}
	defer f.Close()

	if err := d.Write(f); err != nil {
		return true, fmt.Sprintf("Error while write file (%s) error %v\n", desc.Path, err), "Check your dsk  with option -dsk yourdsk.dsk -analyze"
	}
	informations := fmt.Sprintf("raw copy file [%s] size [%d] starting at track [%d] sector [%d] and ending at track [%d] sector [%d]",
		fileInDsk,
		len(content),
		desc.Track,
		desc.Sector,
		endedTrack,
		endedSector)
	msg.ResumeAction(desc.Path, "raw import ", fileInDsk, informations, quiet)
	return false, "", ""
}

func RawImportDsk(d dsk.DSK, fileInDsk string, desc DskDescriptor, quiet bool) (onError bool, message, hint string) {
	if fileInDsk == "" {
		return true, "amsdosfile option is empty, set it.", "dsk -dsk output.dsk -put hello.bin -rawimport -track 1 -sector 0"
	}

	buf, err := os.ReadFile(fileInDsk)
	if err != nil {
		return true, fmt.Sprintf("Cannot open file %s error :%v\n", fileInDsk, err), "Check your file path"
	}
	fmt.Fprintf(os.Stdout, "Writing file content [%s] in dsk [%s] starting at track [%d] sector [%d]\n",
		fileInDsk,
		desc.Path,
		desc.Track,
		desc.Sector,
	)
	return RawImportDataInDsk(d, fileInDsk, desc, buf, quiet)
}
