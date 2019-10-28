package main

import (
	"flag"
	"fmt"
	"github.com/jeromelesaux/m4client/m4"
	"os"
)

var (
	host       = flag.String("host", "", "Ip V4 or hostname of the your CPC")
	infile     = flag.String("file", "", "file path of the file to get or to send")
	remotePath = flag.String("remotepath", "", "remote path where to get or send your file")
	ls         = flag.String("ls", "", "get the current remote path on your m4.")
	cd         = flag.String("cd", "", "change the current remote path on your m4.")
	resetCpc   = flag.Bool("resetcpc", false, "Reset the remote CPC")
	resetM4    = flag.Bool("resetm4", false, "Reset the remote M4")
	m4mkdir    = flag.String("m4mkdir", "", "Create remote directory on your M4.")
	upload     = flag.String("upload", "", "Upload the current file of the current directory on the m4.")
	download   = flag.String("download", "", "Download the remote file from you sdcard.")
	execute    = flag.String("execute", "", "Execute the remote file on your remote CPC")
	run        = flag.String("run", "", "Run remote file on your remote CPC")
	current    = flag.Bool("current", false, "Get the content of the current directory")
	start      = flag.Bool("start", false, "Start m4")
	remove     = flag.String("remove", "", "Remove remote directory or file")
	cmd        = flag.String("cmd", "", "Execute remote command on your CPC")
)

func main() {
	flag.Parse()

	if *host == "" {
		fmt.Fprintf(os.Stderr, "Cannot contact M4 without its hostname or IP\n")
		flag.Usage()
		os.Exit(-1)
	}
	client := &m4.M4Client{
		IPClient: *host,
	}
	if *cmd != "" {
		if *remotePath != "" {
			client.ExecuteCmd(*cmd, *remotePath)
		} else {
			fmt.Fprintf(os.Stderr, "Cannot execute the file (%s) without the remotepath, set it\n", *cmd)
		}
	}
	if *remove != "" {
		if err := client.Remove(*remove); err != nil {
			fmt.Fprintf(os.Stderr, "Cannot remove remote directory or file (%s) host (%s) error %v\n", *remove, *host, err)
		}
	}
	if *start {
		if err := client.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Cannot reset your remote cpc (%s) error %v\n", *host, err)
		}
	}
	if *ls != "" {
		rpath, err := client.Ls(*ls)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while getting remote path with host (%s) error :%v\n", *host, err)
		} else {
			fmt.Fprintf(os.Stdout, "Remote path (%s) host (%s)\n", rpath, *host)
		}
	}
	if *current {
		rpath, err := client.CurrentDirectory()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while getting remote path with host (%s) error :%v\n", *host, err)
		} else {
			fmt.Fprintf(os.Stdout, "Remote path (%s) host (%s)\n", rpath, *host)
		}
	}
	if *cd != "" {
		if err := client.ChangeDirectory(*cd); err != nil {
			fmt.Fprintf(os.Stderr, "Cannot change the directory on the M4 (%s) error :%v\n", *host, err)
		}
	}
	if *resetCpc {
		if err := client.ResetCpc(); err != nil {
			fmt.Fprintf(os.Stderr, "Cannot reset your remote cpc (%s) error %v\n", *host, err)
		}
	}
	if *resetM4 {
		if err := client.ResetM4(); err != nil {
			fmt.Fprintf(os.Stderr, "Cannot reset your remote M4 (%s) error %v\n", *host, err)

		}
	}
	if *m4mkdir != "" {
		if err := client.MakeDirectory(*m4mkdir); err != nil {
			fmt.Fprintf(os.Stderr, "Cannot create remote directory (%s) error %v\n", *host, err)
		}

	}
	if *upload != "" {
		if *remotePath == "" {
			fmt.Fprintf(os.Stderr, "Cannot send the file (%s) without the remotepath, set it\n", *upload)
		} else {
			if err := client.Upload(*remotePath, *upload); err != nil {
				fmt.Fprintf(os.Stderr, "Cannot send the file (%s) to remote path (%s) host (%s) error :%v\n", *upload, *remotePath, *host, err)
			}
		}
	}
	if *download != "" {
		if err := client.Download(*download); err != nil {
			fmt.Fprintf(os.Stderr, "Cannot get the file (%s) from the host (%s) error :%v\n", *download, *host, err)
		}
	}
	if *execute != "" {
		if err := client.Execute(*execute); err != nil {
			fmt.Fprintf(os.Stderr, "Cannot execute remote file (%s) on host (%s) error :%v\n", *execute, *host, err)
		}
	}
	if *run != "" {
		if err := client.Run(*run); err != nil {
			fmt.Fprintf(os.Stderr, "Cannot run remote file (%s) on host (%s) error :%v\n", *execute, *host, err)
		}
	}
}
