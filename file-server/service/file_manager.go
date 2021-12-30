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
	"github.com/pingcap-inc/tiem/common/constants"
	"github.com/pingcap-inc/tiem/file-server/common"
	"github.com/pingcap-inc/tiem/library/framework"
	"golang.org/x/net/context"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

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
		maxFileSize:   common.MaxFileSize,
	}
	return &FileMgr
}

func (mgr *FileManager) UploadFile(ctx context.Context, r *http.Request, uploadPath string) error {
	framework.LogWithContext(ctx).Infof("begin UploadFile: uploadPath %s", uploadPath)
	defer framework.LogWithContext(ctx).Info("end UploadFile")
	if !mgr.checkUploadCnt() {
		framework.LogWithContext(ctx).Errorf("upload goroutine reach max, %d", common.MaxUploadNum)
		return fmt.Errorf("upload goroutine reach max, %d", common.MaxUploadNum)
	}
	mgr.addUploadCnt()
	defer mgr.reduceUploadCnt()

	if err := r.ParseMultipartForm(mgr.maxFileSize); err != nil {
		framework.LogWithContext(ctx).Errorf("could not parse multipart form: %s", err.Error())
		return fmt.Errorf("could not parse multipart form: %s", err.Error())
	}

	// parse and validate file and post parameters
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		framework.LogWithContext(ctx).Errorf("form file failed, %s", err.Error())
		return err
	}
	defer file.Close()
	// Get and print out file size
	fileSize := fileHeader.Size
	framework.LogWithContext(ctx).Infof("File size bytes: %d", fileSize)
	// validate file size
	if fileSize > mgr.maxFileSize {
		framework.LogWithContext(ctx).Errorf("file size %d GB reach max upload file size %d GB", fileSize/bytes.GB, mgr.maxFileSize/bytes.GB)
		return fmt.Errorf("file size %d GB reach max upload file size %d GB", fileSize/bytes.GB, mgr.maxFileSize/bytes.GB)
	}
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("ioutil.ReadAll failed, %s", err.Error())
		return err
	}

	// check file type, detectcontenttype only needs the first 512 bytes
	detectedFileType := http.DetectContentType(fileBytes)
	switch detectedFileType {
	case "application/zip":
		break
	default:
		framework.LogWithContext(ctx).Errorf("invalid file type %s, not xxx.zip", detectedFileType)
		return errors.New("invalid file type, not .zip")
	}
	newPath := filepath.Join(uploadPath, constants.DefaultZipName)
	framework.LogWithContext(ctx).Infof("FileType: %s, File: %s", detectedFileType, newPath)

	// write file
	err = os.RemoveAll(uploadPath)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("remove dir %s failed, %s", uploadPath, err.Error())
		return err
	}
	err = os.MkdirAll(uploadPath, os.ModePerm)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("make dir %s failed, %s", uploadPath, err.Error())
		return err
	}
	newFile, err := os.Create(newPath)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("create new file %s failed, %s", newPath, err.Error())
		return err
	}
	defer newFile.Close() // idempotent, okay to call twice
	if _, err = newFile.Write(fileBytes); err != nil {
		framework.LogWithContext(ctx).Errorf("write new file failed %s", err.Error())
		return err
	}
	if err = newFile.Close(); err != nil {
		framework.LogWithContext(ctx).Errorf("close file failed %s", err.Error())
		return err
	}
	return nil
}

func (mgr *FileManager) DownloadFile(ctx context.Context, c *gin.Context, filePath string) error {
	framework.LogWithContext(ctx).Infof("begin DownloadFile: filePath %s", filePath)
	defer framework.LogWithContext(ctx).Info("end DownloadFile")
	if !mgr.checkDownloadCnt() {
		framework.LogWithContext(ctx).Errorf("download goroutine reach max, %d", common.MaxDownloadNum)
		return fmt.Errorf("download goroutine reach max, %d", common.MaxDownloadNum)
	}
	mgr.addDownloadCnt()
	defer mgr.reduceDownloadCnt()

	info, err := os.Stat(filePath)
	if err != nil && !os.IsExist(err) {
		framework.LogWithContext(ctx).Errorf("stat file failed, %s", err.Error())
		return err
	}
	if info.Size() > mgr.maxFileSize {
		framework.LogWithContext(ctx).Errorf("file size %d GB reach max download file size %d GB", info.Size()/bytes.GB, mgr.maxFileSize/bytes.GB)
		return fmt.Errorf("file size %d GB reach max download file size %d GB", info.Size()/bytes.GB, mgr.maxFileSize/bytes.GB)
	}

	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename="+filepath.Base(filePath))
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Cache-Control", "no-cache")

	c.File(filePath)
	return nil
}

func (mgr *FileManager) ZipDir(ctx context.Context, dir string, zipFile string) error {
	framework.LogWithContext(ctx).Infof("begin zipDir, dir: %s to file: %s", dir, zipFile)
	defer framework.LogWithContext(ctx).Info("end zipDir")
	fz, err := os.Create(zipFile)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("create zip file failed, %s", err.Error())
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
		framework.LogWithContext(ctx).Errorf("filepath walk failed, %s", err.Error())
		return err
	}

	return nil
}

func (mgr *FileManager) UnzipDir(ctx context.Context, zipFile string, dir string) (unzipErr error) {
	framework.LogWithContext(ctx).Infof("begin unzipDir, file: %s to dir: %s", zipFile, dir)
	defer framework.LogWithContext(ctx).Info("end unzipDir")
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		framework.LogWithContext(ctx).Errorf("open zip file failed: %s", err.Error())
		return fmt.Errorf("open zip file failed: %s", err.Error())
	}
	defer r.Close()

	for _, f := range r.File {
		func() {
			path := dir + string(filepath.Separator) + f.Name
			if !strings.HasPrefix(path, dir) {
				framework.LogWithContext(ctx).Errorf("file %s in zip file invalid, may cause path crossing", f.Name)
				unzipErr = fmt.Errorf("file %s in zip file invalid, may cause path crossing", f.Name)
				return
			}

			if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
				framework.LogWithContext(ctx).Errorf("make filepath failed: %s", err.Error())
				unzipErr = fmt.Errorf("make filepath failed: %s", err.Error())
				return
			}
			fDest, err := os.Create(path)
			if err != nil {
				framework.LogWithContext(ctx).Errorf("unzip create failed: %s", err.Error())
				unzipErr = fmt.Errorf("unzip create failed: %s", err.Error())
				return
			}
			defer fDest.Close()

			fSrc, err := f.Open()
			if err != nil {
				framework.LogWithContext(ctx).Errorf("unzip open failed: %s", err.Error())
				unzipErr = fmt.Errorf("unzip open failed: %s", err.Error())
				return
			}
			defer fSrc.Close()

			for {
				_, err := io.CopyN(fDest, fSrc, 1024)
				if err != nil {
					if err == io.EOF {
						break
					}
					framework.LogWithContext(ctx).Errorf("unzip copy failed: %s", err.Error())
					unzipErr = fmt.Errorf("unzip copy failed: %s", err.Error())
					return
				}
			}
		}()
	}
	return
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
	return mgr.downloadCount < common.MaxUploadNum
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
	return mgr.downloadCount < common.MaxDownloadNum
}
