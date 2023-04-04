package ostool

import (
	"errors"
	"golang.org/x/text/encoding/simplifiedchinese"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"unicode/utf8"
)

type ByModTime []os.FileInfo

func ConvertGBK2Str(gbkStr string) string {
	//将GBK编码的字符串转换为utf-8编码
	ret, _ := simplifiedchinese.GBK.NewDecoder().String(gbkStr)
	return ret //如果转换失败返回空字符串

	//如果是[]byte格式的字符串，可以使用Bytes方法
	b, _ := simplifiedchinese.GBK.NewDecoder().Bytes([]byte(gbkStr))
	return string(b)
}

func (fis ByModTime) Len() int {
	return len(fis)
}

func (fis ByModTime) Swap(i, j int) {
	fis[i], fis[j] = fis[j], fis[i]
}

func (fis ByModTime) Less(i, j int) bool {
	return fis[i].ModTime().After(fis[j].ModTime())
}

func Exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		//fmt.Println(err)
		return false
	}
	return true
}

func DirSize(path string) (string, error) {
	if Exists(path) != true {
		return "", errors.New(path + " is not a valid path!")
	}
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	rsize := strconv.FormatInt(size/1024/1024, 10)
	return rsize, err
}

func GetFilelst(path string, filter *regexp.Regexp) ([]os.FileInfo, error) {
	var flist []os.FileInfo

	fileInfo, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, errors.New("read" + path + "failed! error:" + err.Error())
	}

	for _, v := range fileInfo {
		if v.IsDir() {
			continue
		} else {
			fname := filter.FindString(v.Name())
			if fname != "" {
				flist = append(flist, v)
			}
		}
	}
	return flist, nil
}

func Tail(path string, offset int64) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}

	lineStr := make([]byte, offset)
	buff := ""
	fs, err := file.Stat()
	if err != nil {
		return "", err
	}
	fileSize := fs.Size()
	if offset > fileSize {
		_, err = file.Seek(0, io.SeekStart)
	} else {
		_, err = file.Seek(-offset, io.SeekEnd)
	}
	if err != nil {
		return "", err
	}
	_, err = file.Read(lineStr)
	if err != nil {
		return "", err
	}

	buff = string(lineStr)
	mylen := utf8.RuneCountInString(buff)
	//start := strings.Index(buff, "\n")
	start := 0
	nbuff := substring(buff, start, mylen)
	err = file.Close()
	if err != nil {
		return "", err
	}
	return nbuff, nil
}

func GetSortFileByModTime(path string, filter *regexp.Regexp) (ByModTime, error) {
	filelst, err := GetFilelst(path, filter)
	if err != nil {
		return nil, errors.New("read" + path + "failed! error:" + err.Error())
	}
	sort.Sort(ByModTime(filelst))
	return filelst, nil
}

func Grep(content string, filter *regexp.Regexp) ([]string, error) {
	mystr := filter.FindAllString(content, -1)
	return mystr, nil
}

func substring(source string, start int, end int) string {
	var r = []rune(source)
	length := len(r)

	if start < 0 || end > length || start > end {
		return ""
	}

	if start == 0 && end == length {
		return source
	}

	return string(r[start:end])
}

func GetOwner(path string) (string, error) {
	file_info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	file_sys := file_info.Sys().(*syscall.Stat_t)
	uid := file_sys.Uid
	u := strconv.FormatUint(uint64(uid), 10)
	usr, err := user.LookupId(u)
	if err != nil {
		return "", err
	}
	return usr.Username, nil
}

func GetGroup(path string) (string, error) {
	file_info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	file_sys := file_info.Sys().(*syscall.Stat_t)
	gid := file_sys.Gid
	g := strconv.FormatUint(uint64(gid), 10)
	group, err := user.LookupGroupId(g)
	if err != nil {
		return "", err
	}
	return group.Name, nil
}

func GetPerm(path string) (string, error) {
	file_info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	return file_info.Mode().String(), nil
}

func GetFSInfo(catalog string, path string) (map[string]interface{}, error) {
	mymap := make(map[string]interface{})
	if strings.Contains(path, "+") == false {
		owner, _ := GetOwner(path)
		group, _ := GetGroup(path)
		perm, _ := GetPerm(path)
		mymap["类别"] = catalog
		mymap["路径"] = path
		mymap["属主"] = owner
		mymap["属组"] = group
		mymap["权限"] = perm
	} else {
		mymap["类别"] = catalog
		mymap["路径"] = path
		mymap["属主"] = "+ASM"
		mymap["属组"] = "+ASM"
		mymap["权限"] = "+ASM"
	}
	return mymap, nil
}
