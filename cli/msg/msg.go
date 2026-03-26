package msg

import (
	"fmt"
	"os"
)

func ResumeAction(dskFilepath, action, amsdosfile, informations string, quiet bool) {
	if !quiet {
		fmt.Fprintf(os.Stderr, "DSK path [%s]\n", dskFilepath)
		fmt.Fprintf(os.Stderr, "ACTION: Action on DSK [%s] on amsdos file [%s]\n", action, amsdosfile)
		fmt.Fprintf(os.Stderr, "INFO:   %s\n", informations)
	}
}

func ExitOnError(errorMessage, hint string) {
	fmt.Fprintf(os.Stderr, "*************************************************************************\n")
	fmt.Fprintf(os.Stderr, "[ERROR] :\t%s\n", errorMessage)
	fmt.Fprintf(os.Stderr, "[HINT ] :\t%s\n", hint)
	fmt.Fprintf(os.Stderr, "*************************************************************************\n")
	os.Exit(-1)
}
