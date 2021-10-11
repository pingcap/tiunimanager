package secondparty

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"runtime/debug"
)

func assert(b bool) {
	if b {
	} else {
		logger.Error("unexpected panic with stack trace:", string(debug.Stack()))
		panic("unexpected")
	}
}

func myPanic(v interface{}) {
	s := fmt.Sprint(v)
	logger.Errorf("panic: %s, with stack trace: %s", s, string(debug.Stack()))
	panic("unexpected" + s)
}

func newTmpFileWithContent(content []byte) (fileName string, err error) {
	tmpfile, err := ioutil.TempFile("", "tiem-topology-*.yaml")
	if err != nil {
		err = fmt.Errorf("fail to create temp file err: %s", err)
		return "", err
	}
	fileName = tmpfile.Name()
	var ct int
	ct, err = tmpfile.Write(content)
	if err != nil || ct != len(content) {
		tmpfile.Close()
		os.Remove(fileName)
		err = fmt.Errorf(fmt.Sprint("fail to write content to temp file ", fileName, "err:", err, "length of content:", "writed:", ct))
		return "", err
	}
	if err := tmpfile.Close(); err != nil {
		myPanic(fmt.Sprintln("fail to close temp file ", fileName))
	}
	return fileName, nil
}

func jsonMustMarshal(v interface{}) []byte {
	bs, err := json.Marshal(v)
	assert(err == nil)
	return bs
}