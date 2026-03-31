package action

type DskAction string

var (
	ActionListDsk            DskAction = "list"
	ActionFormatDsk          DskAction = "format"
	ActionDisplayHexaFileDsk DskAction = "hexa"
	ActionDesassembleFileDsk DskAction = "desassemble"
	ActionListBasic          DskAction = "listbasic"
	ActionAnalyseDsk         DskAction = "analyze"
	ActionPutFileDsk         DskAction = "put"
	ActionRemoveFileDsk      DskAction = "remove"
	ActionGetFileDsk         DskAction = "get"
	ActionGetAllFileDsk      DskAction = "getall"
	ActionAsciiFileDsk       DskAction = "ascii"
	ActionRawExportDsk       DskAction = "rawexport"
	ActionRawImportDsk       DskAction = "rawimport"
	ActionFileinfoDsk        DskAction = "info"
)

type DskActionFile struct {
	File string
	a    DskAction
}

type DskActions struct {
	a []DskActionFile
}

func NewDskActions() DskActions {
	return DskActions{
		a: []DskActionFile{},
	}
}

func (a *DskActions) WithActionListDsk(path string) DskActions {
	a.a = append(a.a, DskActionFile{File: path, a: ActionListDsk})
	return *a
}

func (a *DskActions) WithActionFormatDsk(path string) DskActions {
	a.a = append(a.a, DskActionFile{File: path, a: ActionFormatDsk})
	return *a
}
func (a *DskActions) WithActionDisplayHexaFileDsk(path string) DskActions {
	a.a = append(a.a, DskActionFile{File: path, a: ActionDisplayHexaFileDsk})
	return *a
}
func (a *DskActions) WithActionDesassembleFileDsk(path string) DskActions {
	a.a = append(a.a, DskActionFile{File: path, a: ActionDesassembleFileDsk})
	return *a
}
func (a *DskActions) WithActionListBasic(path string) DskActions {
	a.a = append(a.a, DskActionFile{File: path, a: ActionListBasic})
	return *a
}
func (a *DskActions) WithActionAnalyseDsk(path string) DskActions {
	a.a = append(a.a, DskActionFile{File: path, a: ActionAnalyseDsk})
	return *a
}
func (a *DskActions) WithActionPutFileDsk(path string) DskActions {
	a.a = append(a.a, DskActionFile{File: path, a: ActionPutFileDsk})
	return *a
}
func (a *DskActions) WithActionRemoveFileDsk(path string) DskActions {
	a.a = append(a.a, DskActionFile{File: path, a: ActionRemoveFileDsk})
	return *a
}
func (a *DskActions) WithActionGetFileDsk(path string) DskActions {
	a.a = append(a.a, DskActionFile{File: path, a: ActionGetFileDsk})
	return *a
}
func (a *DskActions) WithActionAsciiFileDsk(path string) DskActions {
	a.a = append(a.a, DskActionFile{File: path, a: ActionAsciiFileDsk})
	return *a
}
func (a *DskActions) WithActionRawExportDsk(path string) DskActions {
	a.a = append(a.a, DskActionFile{File: path, a: ActionRawExportDsk})
	return *a
}
func (a *DskActions) WithActionRawImportDsk(path string) DskActions {
	a.a = append(a.a, DskActionFile{File: path, a: ActionRawImportDsk})
	return *a
}
func (a *DskActions) WithActionFileinfoDsk(path string) DskActions {
	a.a = append(a.a, DskActionFile{File: path, a: ActionFileinfoDsk})
	return *a
}
