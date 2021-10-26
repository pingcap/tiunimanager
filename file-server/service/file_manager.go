/******************************************************************************
 * Copyright (c)  2021 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 *  Unless required by applicable law or agreed to in writing, software       *
 *  distributed under the License is distributed on an "AS IS" BASIS,         *
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  *
 *  See the License for the specific language governing permissions and       *
 *  limitations under the License.                                            *
 ******************************************************************************/

package service

import (
	"errors"
	"fmt"
	"github.com/labstack/gommon/bytes"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

const maxUploadSize int64 = 1 * bytes.GB
const maxUploadNum int32 = 3
var uploadPath string = "/tmp"

var FileMgr FileManager

type FileManager struct {
	maxUploadSize 	int64
	uploadPath 		string
	uploadCount 	int32
	mutex 			sync.RWMutex
}

func InitFileManager() *FileManager {
	FileMgr = FileManager{
		uploadCount: 0,
	}
	return &FileMgr
}

func (mgr * FileManager) UploadFile(r *http.Request) error {
	if !mgr.checkUploadCnt() {
		return errors.New("upload goroutine reach max")
	}

	mgr.addUploadCnt()
	defer mgr.reduceUploadCnt()
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		return fmt.Errorf("could not parse multipart form: %s", err.Error())
	}

	// parse and validate file and post parameters
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		return err
	}
	defer file.Close()
	// Get and print out file size
	fileSize := fileHeader.Size
	fmt.Printf("File size bytes: %v\n", fileSize)
	// validate file size
	if fileSize > maxUploadSize {
		return errors.New("file size reach max upload file size")
	}
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	// check file type, detectcontenttype only needs the first 512 bytes
	detectedFileType := http.DetectContentType(fileBytes)
	switch detectedFileType {
	case "application/zip":
		break
	default:
		return errors.New("invalid file type")
	}
	newPath := filepath.Join(uploadPath, "data.zip")
	fmt.Printf("FileType: %s, File: %s\n", detectedFileType, newPath)

	// write file
	newFile, err := os.Create(newPath)
	if err != nil {
		return err
	}
	defer newFile.Close() // idempotent, okay to call twice
	if _, err := newFile.Write(fileBytes); err != nil || newFile.Close() != nil {
		return err
	}
	return nil
}

func (mgr * FileManager) DownloadFile(r *http.Request) error {
	//todo
	return nil
}

func (mgr *FileManager) addUploadCnt() {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()
	mgr.uploadCount++
}

func (mgr *FileManager) reduceUploadCnt() {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()
	if mgr.uploadCount > 0 {
		mgr.uploadCount--
	}
}

func (mgr *FileManager) checkUploadCnt() bool {
	return mgr.uploadCount < maxUploadNum
}