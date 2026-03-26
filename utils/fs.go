package utils

import "os"

func SaveFile(b []byte, path string) bool {
	fw, err := os.Create(path)
	if err != nil {
		return true
	}
	defer fw.Close()
	_, err = fw.Write(b)
	return err != nil
}

func Save(path string, b []byte) error {
	return os.WriteFile(path, b, 0666)
}
