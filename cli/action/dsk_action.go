package action

type DskTask string

var (
	ActionListDsk            DskTask = "list"
	ActionFormatDsk          DskTask = "format"
	ActionDisplayHexaFileDsk DskTask = "hexa"
	ActionDesassembleFileDsk DskTask = "desassemble"
	ActionListBasic          DskTask = "listbasic"
	ActionAnalyseDsk         DskTask = "analyze"
	ActionPutFileDsk         DskTask = "put"
	ActionRemoveFileDsk      DskTask = "remove"
	ActionGetFileDsk         DskTask = "get"
	ActionGetAllFileDsk      DskTask = "getall"
	ActionAsciiFileDsk       DskTask = "ascii"
	ActionRawExportDsk       DskTask = "rawexport"
	ActionRawImportDsk       DskTask = "rawimport"
	ActionFileinfoDsk        DskTask = "info"
)

type DskTaskFile struct {
	File   string
	Folder string
	a      DskTask
}

type DskTasks struct {
	a []DskTaskFile
}

func NewDskTasks() *DskTasks {
	return &DskTasks{
		a: []DskTaskFile{},
	}
}

func (a *DskTasks) WithActionListDsk(path string, isSet bool) *DskTasks {
	if isSet {
		a.a = append(a.a, DskTaskFile{File: path, a: ActionListDsk})
	}
	return a
}

func (a *DskTasks) WithActionFormatDsk(path string, isSet bool) *DskTasks {
	if isSet {
		a.a = append(a.a, DskTaskFile{File: path, a: ActionFormatDsk})
	}
	return a
}
func (a *DskTasks) WithActionDisplayHexaFileDsk(path string, isSet bool) *DskTasks {
	if isSet {
		a.a = append(a.a, DskTaskFile{File: path, a: ActionDisplayHexaFileDsk})
	}
	return a
}
func (a *DskTasks) WithActionDesassembleFileDsk(path string, isSet bool) *DskTasks {
	if isSet {
		a.a = append(a.a, DskTaskFile{File: path, a: ActionDesassembleFileDsk})
	}
	return a
}
func (a *DskTasks) WithActionListBasic(path string, isSet bool) *DskTasks {
	if isSet {
		a.a = append(a.a, DskTaskFile{File: path, a: ActionListBasic})
	}
	return a
}
func (a *DskTasks) WithActionAnalyseDsk(path string, isSet bool) *DskTasks {
	if isSet {
		a.a = append(a.a, DskTaskFile{File: path, a: ActionAnalyseDsk})
	}
	return a
}
func (a *DskTasks) WithActionPutFileDsk(path string, isSet bool) *DskTasks {
	if isSet {
		a.a = append(a.a, DskTaskFile{File: path, a: ActionPutFileDsk})
	}
	return a
}
func (a *DskTasks) WithActionRemoveFileDsk(path string, isSet bool) *DskTasks {
	if isSet {
		a.a = append(a.a, DskTaskFile{File: path, a: ActionRemoveFileDsk})
	}
	return a
}
func (a *DskTasks) WithActionGetFileDsk(path string, isSet bool) *DskTasks {
	if isSet {
		a.a = append(a.a, DskTaskFile{File: path, a: ActionGetFileDsk})
	}
	return a
}
func (a *DskTasks) WithActionAsciiFileDsk(path string, isSet bool) *DskTasks {
	if isSet {
		a.a = append(a.a, DskTaskFile{File: path, a: ActionAsciiFileDsk})
	}
	return a
}
func (a *DskTasks) WithActionRawExportDsk(path string, isSet bool) *DskTasks {
	if isSet {
		a.a = append(a.a, DskTaskFile{File: path, a: ActionRawExportDsk})
	}
	return a
}
func (a *DskTasks) WithActionRawImportDsk(path string, isSet bool) *DskTasks {
	if isSet {
		a.a = append(a.a, DskTaskFile{File: path, a: ActionRawImportDsk})
	}
	return a
}
func (a *DskTasks) WithActionFileinfoDsk(path string, isSet bool) *DskTasks {
	if isSet {
		a.a = append(a.a, DskTaskFile{File: path, a: ActionFileinfoDsk})
	}
	return a
}

func (a *DskTasks) WithActionGetAllFileDsk(path string, isSet bool) *DskTasks {
	if isSet {
		a.a = append(a.a, DskTaskFile{Folder: path, File: path, a: ActionGetAllFileDsk})
	}
	return a
}
