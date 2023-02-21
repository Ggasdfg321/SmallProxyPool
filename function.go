package main

import "golang.org/x/text/encoding/simplifiedchinese"

func sliceDelete(tslice any, val any) any {
	if _, ok := tslice.([]string); ok {
		slice := tslice.([]string)
		for i := 0; i < len(slice); i++ {
			if slice[i] == val {
				slice = append(slice[:i], slice[i+1:]...)
				i--
			}
		}
		return slice

	} else if _, ok := tslice.([]int64); ok {
		slice := tslice.([]int64)
		for i := 0; i < len(slice); i++ {
			if slice[i] == val {
				slice = append(slice[:i], slice[i+1:]...)
				i--
			}
		}
		return slice
	} else if _, ok := tslice.([][]string); ok {
		slice := tslice.([][]string)
		for i := 0; i < len(slice); i++ {
			if slice[i][0] == val {
				slice = append(slice[:i], slice[i+1:]...)
				i--
			}
		}
		return slice
	}
	panic("暂时不支持这种类型转换")

}

func slicesFind(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func silcesIndex(slice []string, val string) int {
	for n, item := range slice {
		if item == val {
			return n
		}
	}
	return -1
}

func ConvertByte2String(byte []byte, charset string) string {

	var str string
	switch charset {
	case "GB18030":
		var decodeBytes, _ = simplifiedchinese.GB18030.NewDecoder().Bytes(byte)
		str = string(decodeBytes)
	case "UTF8":
		fallthrough
	default:
		str = string(byte)
	}

	return str
}
