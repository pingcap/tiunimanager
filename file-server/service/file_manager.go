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
	"github.com/gin-gonic/gin"
	"github.com/labstack/gommon/bytes"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const maxFileSize int64 = 1 * bytes.GB
const maxUploadNum int32 = 3
const maxDownloadNum int32 = 3

var FileMgr FileManager

type FileManager struct {
	maxFileSize   int64
	uploadCount   int32
	downloadCount int32
	upMutex       sync.Mutex
	downMutex     sync.Mutex
}

func InitFileManager() *FileManager {
	FileMgr = FileManager{
		uploadCount:   0,
		downloadCount: 0,
		maxFileSize:   maxFileSize,
	}
	return &FileMgr
}

func (mgr *FileManager) UploadFile(r *http.Request, uploadPath string) error {
	getLogger().Infof("begin UploadFile: uploadPath %s", uploadPath)
	defer getLogger().Info("end UploadFile")
	if !mgr.checkUploadCnt() {
		getLogger().Errorf("upload goroutine reach max, %d", maxUploadNum)
		return fmt.Errorf("upload goroutine reach max, %d", maxUploadNum)
	}
	mgr.addUploadCnt()
	defer mgr.reduceUploadCnt()

	if err := r.ParseMultipartForm(mgr.maxFileSize); err != nil {
		getLogger().Errorf("could not parse multipart form: %s", err.Error())
		return fmt.Errorf("could not parse multipart form: %s", err.Error())
	}

	// parse and validate file and post parameters
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		getLogger().Errorf("form file failed, %s", err.Error())
		return err
	}
	defer file.Close()
	// Get and print out file size
	fileSize := fileHeader.Size
	getLogger().Infof("File size bytes: %d", fileSize)
	// validate file size
	if fileSize > mgr.maxFileSize {
		getLogger().Errorf("file size %d GB reach max upload file size %d GB", fileSize/bytes.GB, mgr.maxFileSize/bytes.GB)
		return fmt.Errorf("file size %d GB reach max upload file size %d GB", fileSize/bytes.GB, mgr.maxFileSize/bytes.GB)
	}
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		getLogger().Errorf("ioutil.ReadAll failed, %s", err.Error())
		return err
	}

	// check file type, detectcontenttype only needs the first 512 bytes
	detectedFileType := http.DetectContentType(fileBytes)
	switch detectedFileType {
	case "application/zip":
		break
	default:
		getLogger().Errorf("invalid file type %s, not xxx.zip", detectedFileType)
		return errors.New("invalid file type, not .zip")
	}
	newPath := filepath.Join(uploadPath, "data.zip")
	getLogger().Infof("FileType: %s, File: %s", detectedFileType, newPath)

	// write file
	err = os.RemoveAll(uploadPath)
	if err != nil {
		getLogger().Errorf("remove dir %s failed, %s", uploadPath, err.Error())
		return err
	}
	err = os.MkdirAll(uploadPath, os.ModePerm)
	if err != nil {
		getLogger().Errorf("make dir %s failed, %s", uploadPath, err.Error())
		return err
	}
	newFile, err := os.Create(newPath)
	if err != nil {
		getLogger().Errorf("create new file %s failed, %s", newPath, err.Error())
		return err
	}
	defer newFile.Close() // idempotent, okay to call twice
	if _, err = newFile.Write(fileBytes); err != nil {
		getLogger().Errorf("write new file failed %s", err.Error())
		return err
	}
	if err = newFile.Close(); err != nil {
		getLogger().Errorf("close file failed %s", err.Error())
		return err
	}
	return nil
}

func (mgr *FileManager) DownloadFile(c *gin.Context, filePath string) error {
	getLogger().Infof("begin DownloadFile: filePath %s", filePath)
	defer getLogger().Info("end DownloadFile")
	if !mgr.checkUploadCnt() {
		getLogger().Errorf("download goroutine reach max, %d", maxDownloadNum)
		return fmt.Errorf("download goroutine reach max, %d", maxDownloadNum)
	}
	mgr.addDownloadCnt()
	defer mgr.reduceDownloadCnt()

	info, err := os.Stat(filePath)
	if err != nil && !os.IsExist(err) {
		getLogger().Errorf("stat file failed, %s", err.Error())
		return err
	}
	if info.Size() > mgr.maxFileSize {
		getLogger().Errorf("file size %d GB reach max download file size %d GB", info.Size()/bytes.GB, mgr.maxFileSize/bytes.GB)
		return fmt.Errorf("file size %d GB reach max download file size %d GB", info.Size()/bytes.GB, mgr.maxFileSize/bytes.GB)
	}

	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename="+DefaultDataFile)
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Cache-Control", "no-cache")

	c.File(filePath)
	return nil
}

func (mgr *FileManager) ZipDir(dir string, zipFile string) error {
	getLogger().Infof("begin zipDir: dir[%s] to file[%s]", dir, zipFile)
	defer getLogger().Info("end zipDir")
	fz, err := os.Create(zipFile)
	if err != nil {
		return fmt.Errorf("create zip file failed, %s", err.Error())
	}
	defer fz.Close()

	w := zip.NewWriter(fz)
	defer w.Close()

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			relPath := strings.TrimPrefix(path, filepath.Dir(path))
			fDest, err := w.Create(relPath)
			if err != nil {
				return fmt.Errorf("zip create failed: %s", err.Error())
			}
			fSrc, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("zip open failed: %s", err.Error())
			}
			defer fSrc.Close()
			_, err = io.Copy(fDest, fSrc)
			if err != nil {
				return fmt.Errorf("zip copy failed: %s", err.Error())
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

func (mgr *FileManager) UnzipDir(zipFile string, dir string) error {
	getLogger().Infof("begin unzipDir: file[%s] to dir[%s]", zipFile, dir)
	defer getLogger().Info("end unzipDir")
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return fmt.Errorf("open zip file failed: %s", err.Error())
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
				getLogger().Errorf("unzip create failed: %s", err.Error())
				return
			}
			defer fDest.Close()

			fSrc, err := f.Open()
			if err != nil {
				getLogger().Errorf("unzip open failed: %s", err.Error())
				return
			}
			defer fSrc.Close()

			_, err = io.Copy(fDest, fSrc)
			if err != nil {
				getLogger().Errorf("unzip copy failed: %s", err.Error())
				return
			}
		}()
	}
	return nil
}

func (mgr *FileManager) addUploadCnt() {
	mgr.upMutex.Lock()
	defer mgr.upMutex.Unlock()
	mgr.uploadCount++
}

func (mgr *FileManager) reduceUploadCnt() {
	mgr.upMutex.Lock()
	defer mgr.upMutex.Unlock()
	if mgr.uploadCount > 0 {
		mgr.uploadCount--
	}
}

func (mgr *FileManager) checkUploadCnt() bool {
	return mgr.downloadCount < maxUploadNum
}

func (mgr *FileManager) addDownloadCnt() {
	mgr.downMutex.Lock()
	defer mgr.downMutex.Unlock()
	mgr.downloadCount++
}

func (mgr *FileManager) reduceDownloadCnt() {
	mgr.downMutex.Lock()
	defer mgr.downMutex.Unlock()
	if mgr.downloadCount > 0 {
		mgr.downloadCount--
	}
}

func (mgr *FileManager) checkDownloadCnt() bool {
	return mgr.downloadCount < maxDownloadNum
}
