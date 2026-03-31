package action

import (
	"fmt"
	"os"

	"github.com/jeromelesaux/dsk/cli/msg"
	"github.com/jeromelesaux/dsk/dsk"
	"github.com/jeromelesaux/dsk/sna"
)

type SnaTask string

var (
	SnaInfoAction     SnaTask = "snainfo"
	SnaFormatAction   SnaTask = "snaformat"
	SnaHexaListAction SnaTask = "snahexa"
	SnaPutAction      SnaTask = "snaput"
	SnaGetAction      SnaTask = "snaget"
)

type SnaAction struct {
	Path       string
	File       string
	CPCType    int
	Version    int
	Screenmode int
	tasks      []SnaTask
}

func NewSnaAction(snapath string) *SnaAction {
	return &SnaAction{
		Path: snapath,
	}
}

func (s *SnaAction) SnaIsSet() bool {
	return s.Path != ""
}

func (s *SnaAction) WithFiles(str ...string) *SnaAction {
	for _, v := range str {
		if v != "" {
			s.File = v
		}
	}
	return s
}
func (s *SnaAction) WithCPCType(t int) *SnaAction {
	s.CPCType = t
	return s
}
func (s *SnaAction) WithVersion(v int) *SnaAction {
	s.Version = v
	return s
}
func (s *SnaAction) WithScreemode(m int) *SnaAction {
	s.Screenmode = m
	return s
}

func (s *SnaAction) WithSnaInfoAction(isSet bool) *SnaAction {
	s.tasks = append(s.tasks, SnaInfoAction)
	return s
}

func (s *SnaAction) WithSnaFormatAction(isSet bool) *SnaAction {
	if isSet {
		s.tasks = append(s.tasks, SnaFormatAction)
	}
	return s
}

func (s *SnaAction) WithSnaHexaListAction(isSet bool) *SnaAction {
	if isSet {
		s.tasks = append(s.tasks, SnaHexaListAction)
	}
	return s
}
func (s *SnaAction) WithSnaPutAction(isSet bool) *SnaAction {
	if isSet {
		s.tasks = append(s.tasks, SnaPutAction)
	}
	return s
}
func (s *SnaAction) WithSnaGetAction(isSet bool) *SnaAction {
	if isSet {
		s.tasks = append(s.tasks, SnaGetAction)
	}
	return s
}

func (s *SnaAction) DoSnaActions() (onError bool, message, hint string) {
	for _, task := range s.tasks {
		switch task {
		case SnaInfoAction:
			return InfoSna(s.Path)
		case SnaFormatAction:
			return FormatSna(s.Path, s.Version)
		case SnaHexaListAction:
			sna, err := sna.ReadSna(s.Path)
			if err != nil {
				msg.ExitOnError(err.Error(), "Check your sna path")
			}
			content := sna.Hexadecimal()
			fmt.Println(dsk.DisplayHex([]byte(content), 16))
		case SnaPutAction:
			crtc := sna.UM6845R
			if s.CPCType > 3 {
				crtc = sna.ASIC_6845
			}
			if err := sna.ImportInSna(s.File, s.Path, uint8(s.Screenmode), sna.CPCType(s.CPCType), crtc, s.Version); err != nil {
				return true, fmt.Sprintf("Error while trying to import file (%s) in new sna (%s) error: %v\n", s.File, s.Path, err), ""
			}
		case SnaGetAction:
			content, err := sna.ExportFromSna(s.Path)
			if err != nil {
				return true, fmt.Sprintf("Error while trying to import file (%s) in new sna (%s) error: %v\n",
					s.File,
					s.Path,
					err), ""
			}
			f, err := os.Create(s.File)
			if err != nil {
				return true, fmt.Sprintf("Error while trying to import file (%s) in new sna (%s) error: %v\n",
					s.File,
					s.Path,
					err), ""

			}
			defer f.Close()
			_, err = f.Write(content)
			if err != nil {
				return true, fmt.Sprintf("Error while trying to import file (%s) in new sna (%s) error: %v\n",
					s.File,
					s.Path,
					err), ""
			}
		default:
			return InfoSna(s.Path)
		}
	}
	return false, "", ""
}

func FormatSna(snaPath string, snaVersion int) (onError bool, message, hint string) {
	if _, err := sna.CreateSna(snaPath, snaVersion); err != nil {
		return true, fmt.Sprintf("Cannot create Sna file (%s) error : %v\n", snaPath, err), ""
	}
	fmt.Fprintf(os.Stderr, "Sna file (%s) created.\n", snaPath)
	return false, "", ""
}

func InfoSna(snaPath string) (onError bool, message, hint string) {
	f, err := os.Open(snaPath)
	if err != nil {
		msg.ExitOnError(fmt.Sprintf("Error while read sna file (%s) error %v", snaPath, err), "Check your sna file")
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
