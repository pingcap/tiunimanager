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
	"archive/zip"
	"errors"
	"fmt"
	"github.com/labstack/gommon/bytes"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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
		getLogger().Errorf("upload goroutine reach max, %d", maxUploadNum)
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

func (mgr * FileManager) zipDir(dir string, zipFile string) error {
	getLogger().Infof("begin zipDir: dir[%s] to file[%s]", dir, zipFile)
	defer getLogger().Info("end zipDir")
	fz, err := os.Create(zipFile)
	if err != nil {
		return fmt.Errorf("Create zip file failed: %s", err.Error())
	}
	defer fz.Close()

	w := zip.NewWriter(fz)
	defer w.Close()

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			relPath := strings.TrimPrefix(path, filepath.Dir(path))
			fDest, err := w.Create(relPath)
			if err != nil {
				return fmt.Errorf("zip Create failed: %s", err.Error())
			}
			fSrc, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("zip Open failed: %s", err.Error())
			}
			defer fSrc.Close()
			_, err = io.Copy(fDest, fSrc)
			if err != nil {
				return fmt.Errorf("zip Copy failed: %s", err.Error())
			}
		}
		return nil
	})
	if err != nil {
		getLogger().Errorf("filepath walk failed, %s", err.Error())
		return err
	}

	return nil
}

func (mgr * FileManager) unzipDir(zipFile string, dir string) error {
	getLogger().Infof("begin unzipDir: file[%s] to dir[%s]", zipFile, dir)
	defer getLogger().Info("end unzipDir")
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return fmt.Errorf("Open zip file failed: %s", err.Error())
	}
	defer r.Close()

	for _, f := range r.File {
		func() {
			path := dir + string(filepath.Separator) + f.Name
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				getLogger().Errorf("make filepath failed: %s", err.Error())
				return
			}
			fDest, err := os.Create(path)
			if err != nil {
				getLogger().Errorf("unzip Create failed: %s", err.Error())
				return
			}
			defer fDest.Close()

			fSrc, err := f.Open()
			if err != nil {
				getLogger().Errorf("unzip Open failed: %s", err.Error())
				return
			}
			defer fSrc.Close()

			_, err = io.Copy(fDest, fSrc)
			if err != nil {
				getLogger().Errorf("unzip Copy failed: %s", err.Error())
				return
			}
		}()
	}
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