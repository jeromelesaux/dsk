package m4

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// M4HttpAction is struct for url complement according to the action
type M4HttpAction string

var userAgent = "cpcxfer"

// M4 Wifi card http possibles actions
const (
	M4Reset  M4HttpAction = "config.cgi?mres"
	CpcReset M4HttpAction = "config.cgi?cres"
	Start    M4HttpAction = "config.cgi?cctr"
	Mkdir    M4HttpAction = "config.cgi?mkdir="
	Ls       M4HttpAction = "config.cgi?ls="
	Cd       M4HttpAction = "config.cgi?cd="
	Rm       M4HttpAction = "config.cgi?rm="
	Execute  M4HttpAction = "config.cgi?run2="
	Run      M4HttpAction = "config.cgi?run="
	Pause    M4HttpAction = "config.cgi?chlt"
	Upload   M4HttpAction = "upload.html"
	Download M4HttpAction = "sd/"
	Rom      M4HttpAction = "roms.shtml"
)

type M4Node struct {
	Name        string
	IsDirectory bool
	Size        string
}

type M4Dir struct {
	CurrentPath string
	Nodes       []M4Node
}

func NewM4Dir(content string) *M4Dir {
	res := strings.Split(content, "\n")
	if len(res) == 0 {
		return &M4Dir{}
	}
	d := &M4Dir{CurrentPath: res[0], Nodes: make([]M4Node, 0)}
	for i := 1; i < len(res); i++ {
		l := res[i]
		var size, name, dir string
		var index int
		for j := len(l) - 1; j >= 0; j-- {
			if l[j] == ',' {
				break
			}
			size = string(l[j]) + size
			index++
		}
		for j := len(l) - index; j >= 0; j-- {
			if l[j] == ',' {
				break
			}
			dir = string(l[j]) + dir
			index++
		}

		name = string(l[0:(len(l) - (index + 2))])

		node := M4Node{Name: name, Size: size}
		if dir == "0" {
			node.IsDirectory = true
		}
		d.Nodes = append(d.Nodes, node)
	}

	return d
}

// M4Client M4 http client with action, address ip client
// and Cpc file path
type M4Client struct {
	action   M4HttpAction
	IPClient string
	//	CpcLocalFilePath  string
	//	CpcRemoteFilePath string
}

func (m *M4Client) Url() string {
	return "http://" + m.IPClient + "/" + string(m.action)
}

func PerformHttpAction(req *http.Request) error {
	client := &http.Client{}
	req.Header.Add("user-agent", userAgent)
	fmt.Fprintf(os.Stdout, "User-agent:%s\n", req.Header.Get("user-agent"))
	fmt.Fprintf(os.Stdout, "Query:%s\n", req.RemoteAddr)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	fmt.Fprintf(os.Stdout, "Response code :%d\n", resp.StatusCode)
	if resp.StatusCode != http.StatusOK {
		return errors.New("Response from cpc http server differs from 200")
	}
	//body, err := ioutil.ReadAll(resp.Body)
	//if err != nil {
	//	return err
	//}
	//fmt.Fprintf(os.Stdout, "Response body %s\n", body)
	return nil
}

func (m *M4Client) PauseCpc() error {
	m.action = Pause
	req, err := http.NewRequest("GET", m.Url(), nil)
	if err != nil {
		return err
	}
	return PerformHttpAction(req)
}

func (m *M4Client) Start() error {
	m.action = Start
	req, err := http.NewRequest("GET", m.Url(), nil)
	if err != nil {
		return err
	}
	return PerformHttpAction(req)
}

func (m *M4Client) ResetM4() error {
	m.action = M4Reset
	req, err := http.NewRequest("GET", m.Url(), nil)
	if err != nil {
		return err
	}
	return PerformHttpAction(req)
}

func (m *M4Client) ResetCpc() error {
	m.action = CpcReset
	req, err := http.NewRequest("GET", m.Url(), nil)
	if err != nil {
		return err
	}
	return PerformHttpAction(req)
}

func (m *M4Client) Download(remotePath string) error {
	m.action = Download
	fh, err := os.Create(UniversalBase(remotePath))
	if err != nil {
		return err
	}
	defer fh.Close()
	req, err := http.NewRequest("GET", m.Url()+remotePath, nil)
	req.Header.Add("user-agent", userAgent)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New("Not http status ok ")
	}
	_, err = io.Copy(fh, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

func (m *M4Client) UploadDirectoryContent(remotePath, localDirectoryPath string) error {
	files, err := ioutil.ReadDir(localDirectoryPath)
	if err != nil {
		return err
	}

	for _, file := range files {
		if !file.IsDir() {
			m.Upload(remotePath, localDirectoryPath+string(filepath.Separator)+file.Name())
		}
	}
	return nil
}

func UniversalBase(filePath string) string {

	if runtime.GOOS == "windows" {
		pos := strings.LastIndex(filePath, "\\")
		return filePath[pos+1 : len(filePath)]
	} else {
		return path.Base(filePath)
	}
}

func (m *M4Client) Upload(remotePath, localPath string) error {
	m.action = Upload
	remoteFilePath := remotePath + "/" + UniversalBase(localPath)
	fmt.Fprintf(os.Stdout, "M4 action :%s,input file:%s url:%s, parameter:%s\n", m.action, localPath, m.Url(), remoteFilePath)
	fh, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer fh.Close()
	/*if _, err := cpc.NewCpcHeader(fh); err != nil {
		return err
	}
	fh.Seek(0, 0)*/
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	fmt.Fprintf(os.Stdout, "remote file path (%s)\n", remoteFilePath)
	part, err := writer.CreateFormFile("upfile", remoteFilePath)
	if err != nil {
		return err
	}
	_, err = io.Copy(part, fh)
	if err != nil {
		return err
	}
	writer.Close()

	req, err := http.NewRequest("POST", m.Url(), body)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())
	req.Header.Add("Expires", "0")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New("Http response differs from 200")
	}
	return nil
}

func (m *M4Client) Execute(cpcfile string) error {
	m.action = Execute
	req, err := http.NewRequest("GET", m.Url()+cpcfile, nil)
	if err != nil {
		return err
	}
	return PerformHttpAction(req)
}
func (m *M4Client) ExecuteCmd(cmd, cpcfile string) error {
	m.action = Execute
	req, err := http.NewRequest("GET", m.Url()+cmd+","+cpcfile, nil)
	if err != nil {
		return err
	}
	return PerformHttpAction(req)
}

func (m *M4Client) Remove(cpcfile string) error {
	m.action = Rm
	req, err := http.NewRequest("GET", m.Url()+cpcfile, nil)
	if err != nil {
		return err
	}
	return PerformHttpAction(req)
}

func (m *M4Client) Run(cpcfile string) error {
	m.action = Run
	req, err := http.NewRequest("GET", m.Url()+cpcfile, nil)
	if err != nil {
		return err
	}
	return PerformHttpAction(req)
}

func (m *M4Client) MakeDirectory(remotedirectory string) error {
	m.action = Mkdir
	fmt.Fprintf(os.Stdout, "M4 action :%s, url:%s\n", m.action, m.Url()+remotedirectory)
	req, err := http.NewRequest("GET", m.Url()+remotedirectory, nil)
	if err != nil {
		return err
	}
	return PerformHttpAction(req)
}

func (m *M4Client) ChangeDirectory(remotedirectory string) error {
	m.action = Cd
	req, err := http.NewRequest("GET", m.Url()+remotedirectory, nil)
	if err != nil {
		return err
	}
	return PerformHttpAction(req)
}

func (m *M4Client) DeleteRom(romNumber int) error {
	m.action = Rom
	req, err := http.NewRequest("GET", m.Url()+"?rmsl="+strconv.Itoa(romNumber), nil)
	if err != nil {
		return err
	}
	return PerformHttpAction(req)
}

func (m *M4Client) UploadRom(romFilpath, romName string, romId int) error {
	if romId < 0 || romId >= 32 {
		return errors.New("Rom id is not compliant.")
	}
	m.action = Rom

	fh, err := os.Open(romFilpath)
	if err != nil {
		return err
	}
	defer fh.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("uploadedfile", "rom.bin")
	if err != nil {
		return err
	}
	_, err = io.Copy(part, fh)
	if err != nil {
		return err
	}
	slotNumW, err := writer.CreateFormField("slotnum")
	if err != nil {
		return err
	}
	slotNumW.Write([]byte(fmt.Sprintf("%d", romId)))

	slotNameW, err := writer.CreateFormField("slotname")
	if err != nil {
		return err
	}
	slotNameW.Write([]byte(romName))

	writer.Close()

	req, err := http.NewRequest("POST", m.Url(), body)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())
	req.Header.Add("Expires", "0")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New("Http response differs from 200")
	}
	return nil
}

func (m *M4Client) GetCache(remotePath string) (string, error) {
	m.action = Download

	req, err := http.NewRequest("GET", m.Url()+remotePath, nil)
	req.Header.Add("user-agent", userAgent)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", errors.New("Not http status ok ")
	}
	fh := new(bytes.Buffer)
	_, err = io.Copy(fh, resp.Body)
	if err != nil {
		return "", err
	}
	return string(fh.Bytes()), nil
}

/*
./m4client -host 192.168.1.20 -ls Jeux
Remote path (Jeux/
Ishido.dsk,1,190K
Doomsday_Lost_Echoes_v1.0,0,0
GalacticTomb_128K,0,0
ImperialMahjong,0,0
Orion Prime (FR) (Cargosoft),0,0
The Shadows Of Sergoth v1.0 (F,UK,S) (128K) (Face A) (2018) [Original].dsk,1,190K
The Shadows Of Sergoth v1.0 (F,UK,S) (128K) (Face B) (2018) [Original].dsk,1,190K
Ishido,0,0
) host (192.168.1.20)
*/
func (m *M4Client) Ls(remoteDirectory string) (string, error) {
	m.action = Ls
	client := &http.Client{}
	req, err := http.NewRequest("GET", m.Url()+remoteDirectory, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("user-agent", userAgent)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("Response from cpc http server differs from 200")
	}
	content, err := m.GetCache("/m4/dir.txt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot get the dir.txt content file error :%v\n", err)
		return "", err
	}

	return content, nil
}

func (m *M4Client) CurrentDirectory() (string, error) {
	content, err := m.GetCache("/m4/dir.txt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot get the dir.txt content file error :%v\n", err)
		return "", err
	}

	return content, nil
}
