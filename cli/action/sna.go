package action

import (
	"fmt"
	"os"

	"github.com/jeromelesaux/dsk/cli/msg"
	"github.com/jeromelesaux/dsk/sna"
)

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
