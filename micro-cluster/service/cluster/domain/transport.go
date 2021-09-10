package domain

import (
	ctx "context"
	"errors"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/pingcap-inc/tiem/library/client"
	"github.com/pingcap-inc/tiem/library/secondparty/libtiup"
	proto "github.com/pingcap-inc/tiem/micro-cluster/proto"
	db "github.com/pingcap-inc/tiem/micro-metadb/proto"
	"os"
	"strconv"
	"time"
)

type TransportType string

const (
	TransportTypeExport TransportType = "export"
	TransportTypeImport TransportType = "import"
)

const (
	TransportStatusRunning string = "Running"
	TransportStatusFailed  string = "Failed"
	TransportStatusSuccess string = "Success"
)

type ImportInfo struct {
	ClusterId 	string
	UserName  	string
	Password  	string
	FilePath  	string
	RecordId  	string
	StorageType string
	ConfigPath 	string
}

type ExportInfo struct {
	ClusterId string
	UserName  string
	Password  string
	FileType  string
	RecordId  string
	FilePath  string
	Filter    string
	StorageType string
}

type TransportInfo struct {
	ClusterId     string
	RecordId      string
	TransportType string
	Status        string
	FilePath      string
	StartTime     int64
	EndTime       int64
}

/*
	data import toml config for lightning
	https://docs.pingcap.com/zh/tidb/dev/tidb-lightning-configuration
*/
type DataImportConfig struct {
	Lightning    LightningCfg    `toml:"lightning"`
	TikvImporter TikvImporterCfg `toml:"tikv-importer"`
	MyDumper     MyDumperCfg     `toml:"mydumper"`
	Tidb         TidbCfg         `toml:"tidb"`
}

type LightningCfg struct {
	Level             string `toml:"level"`              //lightning log level
	File              string `toml:"file"`               //lightning log path
	CheckRequirements bool   `toml:"check-requirements"` //lightning pre check
}

/*
	tidb-lightning backend
	https://docs.pingcap.com/zh/tidb/stable/tidb-lightning-backends#tidb-lightning-backend
 */
const (
	BackendLocal  string = "local"
	BackendImport string = "importer"
	BackendTidb   string = "tidb"
)

const (
	NfsStorageType string = "nfs"
	S3StorageType string = "s3"
)

const (
	DefaultTidbPort int = 4000
	DefaultTidbStatusPort int = 10080
	DefaultPDClientPort int = 2379
)

type TikvImporterCfg struct {
	Backend     string `toml:"backend"`       //backend mode: local/normal
	SortedKvDir string `toml:"sorted-kv-dir"` //temp store path
}

type MyDumperCfg struct {
	DataSourceDir string `toml:"data-source-dir"` //import data filepath
}

type TidbCfg struct {
	Host       string `toml:"host"`
	Port       int    `toml:"port"`
	User       string `toml:"user"`
	Password   string `toml:"password"`
	StatusPort int    `toml:"status-port"` //table infomation from tidb status port
	PdAddr     string `toml:"pd-addr"`
}

var contextDataTransportKey = "dataTransportInfo"
var defaultTransportNfsDirPrefix = "/tmp/tiem/transport" //todo: move to config
var defaultTransportS3DirPrefix = "s3://nfs/tiem/transport?access-key=minioadmin&secret-access-key=minioadmin&endpoint=http://minio.pingcap.net:9000&force-path-style=true" //todo: test env ak sk

func ExportDataPreCheck(req *proto.DataExportRequest) error {
	if req.GetFilePath() == "" {
		req.FilePath = defaultTransportS3DirPrefix
	}
	if req.GetStorageType() == "" {
		req.StorageType = S3StorageType
	}

	if req.GetClusterId() == "" {
		return errors.New("invalid param clusterId")
	}
	if req.GetUserName() == "" {
		return errors.New("invalid param userName")
	}
	/*
	if req.GetPassword() == "" {
		return errors.New("invalid param password")
	}
	*/
	if S3StorageType != req.GetStorageType() && NfsStorageType != req.GetStorageType() {
		return errors.New("invalid param storageType")
	}

	return nil
}

func ImportDataPreCheck(req *proto.DataImportRequest) error {
	if req.GetFilePath() == "" {
		req.FilePath = defaultTransportS3DirPrefix
	}
	if req.GetStorageType() == "" {
		req.StorageType = S3StorageType
	}

	if req.GetClusterId() == "" {
		return errors.New("invalid param clusterId")
	}
	if req.GetUserName() == "" {
		return errors.New("invalid param userName")
	}
	/*
		if req.GetPassword() == "" {
			return errors.New("invalid param password")
		}
	*/
	if S3StorageType != req.GetStorageType() && NfsStorageType != req.GetStorageType() {
		return errors.New("invalid param storageType")
	}

	return nil
}

func ExportData(ope *proto.OperatorDTO, clusterId string, filePath string, storageType string, userName string, password string, fileType string, filter string) (string, error) {
	getLogger().Infof("begin exportdata clusterId: %s, storageType: %s, filePath: %s, userName: %s, password: %s, fileType: %s, filter: %s", clusterId, storageType, fileType, userName, password, fileType, filter)
	defer getLogger().Infof("end exportdata")
	//todo: check operator
	operator := parseOperatorFromDTO(ope)
	getLogger().Info(operator)
	clusterAggregation, err := ClusterRepo.Load(clusterId)
	if err != nil {
		getLogger().Errorf("load cluster %s aggregation from metadb failed", clusterId)
		return "", err
	}

	req := &db.DBCreateTransportRecordRequest{
		Record: &db.TransportRecordDTO{
			ClusterId:     clusterId,
			TenantId:      operator.TenantId,
			TransportType: string(TransportTypeExport),
			FilePath:      getDataTransportDir(clusterId, TransportTypeExport, filePath, storageType),
			Status:        TransportStatusRunning,
			StartTime:     time.Now().Unix(),
			EndTime:       time.Now().Unix(),
		},
	}
	resp, err := client.DBClient.CreateTransportRecord(ctx.Background(), req)
	if err != nil {
		return "", err
	}

	info := &ExportInfo{
		ClusterId: clusterId,
		UserName:  userName,
		Password:  password, //todo: need encrypt
		FileType:  fileType,
		RecordId:  resp.GetId(),
		FilePath:  getDataTransportDir(clusterId, TransportTypeExport, filePath, storageType),
		Filter:    filter,
		StorageType: storageType,
	}

	// Start the workflow
	flow, err := CreateFlowWork(clusterId, FlowExportData, operator)
	if err != nil {
		return "", err
	}
	flow.AddContext(contextClusterKey, clusterAggregation)
	flow.AddContext(contextDataTransportKey, info)
	flow.Start()

	clusterAggregation.CurrentWorkFlow = flow.FlowWork
	ClusterRepo.Persist(clusterAggregation)
	return info.RecordId, nil
}

func ImportData(ope *proto.OperatorDTO, clusterId string, userName string, password string, filepath string, storageType string) (string, error) {
	getLogger().Infof("begin importdata clusterId: %s, storageType: %s, userName: %s, password: %s, filePath: %s", clusterId, storageType, userName, password, filepath)
	defer getLogger().Infof("end importdata")
	//todo: check operator
	operator := parseOperatorFromDTO(ope)
	getLogger().Info(operator)

	clusterAggregation, err := ClusterRepo.Load(clusterId)
	if err != nil {
		getLogger().Errorf("load cluster %s aggregation from metadb failed", clusterId)
		return "", err
	}

	req := &db.DBCreateTransportRecordRequest{
		Record: &db.TransportRecordDTO{
			ClusterId:     clusterId,
			TenantId:      operator.TenantId,
			TransportType: string(TransportTypeImport),
			FilePath:      filepath,
			Status:        TransportStatusRunning,
			StartTime:     time.Now().Unix(),
			EndTime:       time.Now().Unix(),
		},
	}
	resp, err := client.DBClient.CreateTransportRecord(ctx.Background(), req)
	if err != nil {
		return "", err
	}
	info := &ImportInfo{
		ClusterId: clusterId,
		UserName:  userName,
		Password:  password, //todo: need encrypt
		FilePath:  filepath,
		RecordId:  resp.GetId(),
		StorageType: storageType,
		ConfigPath: getDataTransportDir(clusterId, TransportTypeImport, "", ""),
	}

	// Start the workflow
	flow, err := CreateFlowWork(clusterId, FlowImportData, operator)
	if err != nil {
		return "", err
	}
	flow.AddContext(contextClusterKey, clusterAggregation)
	flow.AddContext(contextDataTransportKey, info)
	flow.Start()

	clusterAggregation.CurrentWorkFlow = flow.FlowWork
	ClusterRepo.Persist(clusterAggregation)
	return info.RecordId, nil
}

func DescribeDataTransportRecord(ope *proto.OperatorDTO, recordId, clusterId string, page, pageSize int32) ([]*db.TransportRecordDTO, *db.DBPageDTO, error) {
	getLogger().Infof("begin DescribeDataTransportRecord clusterId: %s, recordId: %s, page: %d, pageSize: %d", clusterId, recordId, page, pageSize)
	defer getLogger().Info("end DescribeDataTransportRecord")
	req := &db.DBListTransportRecordRequest{
		Page: &db.DBPageDTO{
			Page:     page,
			PageSize: pageSize,
		},
		ClusterId: clusterId,
		RecordId:  recordId,
	}
	resp, err := client.DBClient.ListTrasnportRecord(ctx.Background(), req)
	if err != nil {
		return nil, nil, err
	}

	return resp.GetRecords(), resp.GetPage(), nil
}

func convertTomlConfig(clusterAggregation *ClusterAggregation, info *ImportInfo) *DataImportConfig {
	getLogger().Info("begin convertTomlConfig")
	defer getLogger().Info("end convertTomlConfig")
	if clusterAggregation == nil || clusterAggregation.CurrentTiUPConfigRecord == nil {
		return nil
	}
	configModel := clusterAggregation.CurrentTiUPConfigRecord.ConfigModel
	if configModel == nil || configModel.TiDBServers == nil || configModel.PDServers == nil {
		return nil
	}
	tidbServer := configModel.TiDBServers[0]
	pdServer := configModel.PDServers[0]

	tidbServerPort := tidbServer.Port
	if tidbServerPort == 0 {
		tidbServerPort = DefaultTidbPort
	}

	tidbStatusPort := tidbServer.StatusPort
	if tidbStatusPort == 0 {
		tidbStatusPort = DefaultTidbStatusPort
	}

	pdClientPort := pdServer.ClientPort
	if pdClientPort == 0 {
		pdClientPort = DefaultPDClientPort
	}

	/*
	 * todo: sorted-kv-dir and data-source-dir in the same disk, may slow down import performance,
	 *  and check-requirements = true can not pass lightning pre-check
	 *  in real environment, config data-source-dir = user nfs storage, sorted-kv-dir = other disk, turn on pre-check
	 */
	config := &DataImportConfig{
		Lightning: LightningCfg{
			Level:             "info",
			File:              fmt.Sprintf("%s/tidb-lightning.log", info.ConfigPath),
			CheckRequirements: false, //todo: TBD
		},
		TikvImporter: TikvImporterCfg{
			Backend:     BackendLocal,
			SortedKvDir: info.ConfigPath, //todo: TBD
		},
		MyDumper: MyDumperCfg{
			DataSourceDir: info.FilePath,
		},
		Tidb: TidbCfg{
			Host:       tidbServer.Host,
			Port:       tidbServerPort,
			User:       info.UserName,
			Password:   info.Password,
			StatusPort: tidbStatusPort,
			PdAddr:     fmt.Sprintf("%s:%d", pdServer.Host, pdClientPort),
		},
	}
	return config
}

/**
data import && export dir
└── filePath/[cluster-id]
	├── import
	|	├── data
	|	├── tidb-lightning.toml
	|	└── tidb-lightning.log
	└── export
		└── data
*/
func getDataTransportDir(clusterId string, transportType TransportType, filePath string, storageType string) string {
	if S3StorageType == storageType {
		if filePath != "" {
			return fmt.Sprintf("%s/%s/%s", filePath, clusterId, transportType)
		} else {
			return fmt.Sprintf("%s/%s/%s", defaultTransportS3DirPrefix, clusterId, transportType)
		}
	} else {
		if filePath != "" {
			return fmt.Sprintf("%s/%s/%s", filePath, clusterId, transportType)
		} else {
			return fmt.Sprintf("%s/%s/%s", defaultTransportNfsDirPrefix, clusterId, transportType)
		}
	}
}

func cleanDataTransportDir(filepath string) error {
	getLogger().Infof("clean and re-mkdir data dir: %s", filepath)
	if err := os.RemoveAll(filepath); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath, os.ModePerm); err != nil {
		return err
	}
	return nil
}

func buildDataImportConfig(task *TaskEntity, context *FlowContext) bool {
	getLogger().Info("begin buildDataImportConfig")
	defer getLogger().Info("end buildDataImportConfig")

	clusterAggregation := context.value(contextClusterKey).(*ClusterAggregation)
	info := context.value(contextDataTransportKey).(*ImportInfo)

	if err := cleanDataTransportDir(info.ConfigPath); err != nil {
		getLogger().Errorf("clean import directory failed, %s", err.Error())
		return false
	}

	config := convertTomlConfig(clusterAggregation, info)
	if config == nil {
		getLogger().Errorf("convert toml config failed, cluster: %v", clusterAggregation)
		return false
	}
	filePath := fmt.Sprintf("%s/tidb-lightning.toml", info.ConfigPath)
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0766)
	if err != nil {
		getLogger().Errorf("create import toml config failed, %s", err.Error())
		return false
	}

	if err = toml.NewEncoder(file).Encode(config); err != nil {
		getLogger().Errorf("decode data import toml config failed, %s", err.Error())
		return false
	}
	getLogger().Infof("build lightning toml file sucess, %v", config)

	return true
}

func importDataToCluster(task *TaskEntity, context *FlowContext) bool {
	getLogger().Info("begin importDataToCluster")
	defer getLogger().Info("end importDataToCluster")

	info := context.value(contextDataTransportKey).(*ImportInfo)

	//tiup tidb-lightning -config tidb-lightning.toml
	//todo: tiupmgr not return failed err
	resp, err := libtiup.MicroSrvTiupLightning(0,
		[]string{"-config", fmt.Sprintf("%s/tidb-lightning.toml", info.ConfigPath)},
		uint64(task.Id))
	if err != nil {
		getLogger().Errorf("call tiup lightning api failed, %s", err.Error())
		return false
	}
	getLogger().Infof("call tiupmgr tidb-lightning api success, %v", resp)

	return true
}

func updateDataImportRecord(task *TaskEntity, context *FlowContext) bool {
	getLogger().Info("begin updateDataImportRecord")
	defer getLogger().Info("end updateDataImportRecord")
	clusterAggregation := context.value(contextClusterKey).(*ClusterAggregation)
	info := context.value(contextDataTransportKey).(*ImportInfo)
	cluster := clusterAggregation.Cluster

	req := &db.DBUpdateTransportRecordRequest{
		Record: &db.TransportRecordDTO{
			ID:        info.RecordId,
			ClusterId: cluster.Id,
			Status:    TransportStatusSuccess,
			EndTime:   time.Now().Unix(),
		},
	}
	resp, err := client.DBClient.UpdateTransportRecord(ctx.Background(), req)
	if err != nil {
		getLogger().Errorf("update data transport record failed, %s", err.Error())
		return false
	}
	getLogger().Infof("update data transport record success, %v", resp)
	return true
}

func exportDataFromCluster(task *TaskEntity, context *FlowContext) bool {
	getLogger().Info("begin exportDataFromCluster")
	defer getLogger().Info("end exportDataFromCluster")

	clusterAggregation := context.value(contextClusterKey).(*ClusterAggregation)
	info := context.value(contextDataTransportKey).(*ExportInfo)
	configModel := clusterAggregation.CurrentTiUPConfigRecord.ConfigModel
	tidbServer := configModel.TiDBServers[0]
	tidbServerPort := tidbServer.Port
	if tidbServerPort == 0 {
		tidbServerPort = DefaultTidbPort
	}

	if err := cleanDataTransportDir(info.FilePath); err != nil {
		getLogger().Errorf("clean export directory failed, %s", err.Error())
		return false
	}

	//tiup dumpling -u root -P 4000 --host 127.0.0.1 --filetype sql -t 8 -o /tmp/test -r 200000 -F 256MiB --filter "user*"
	//todo: admin root password
	//todo: tiupmgr not return failed err
	cmd := []string{"-u", info.UserName,
		"-p", info.Password,
		"-P", strconv.Itoa(tidbServerPort),
		"--host", tidbServer.Host,
		"--filetype", info.FileType,
		"-t", "8",
		"-o", fmt.Sprintf("%s", info.FilePath),
		"-r", "200000",
		"-F", "256MiB"}
	if info.Filter != "" {
		cmd = append(cmd, "--filter", fmt.Sprintf("\"%s\"", info.Filter))
	}
	getLogger().Infof("call tiupmgr dumpling api, cmd: %v", cmd)
	resp, err := libtiup.MicroSrvTiupDumpling(0, cmd, uint64(task.Id))
	if err != nil {
		getLogger().Errorf("call tiup dumpling api failed, %s", err.Error())
		return false
	}

	getLogger().Infof("call tiupmgr succee, resp: %v", resp)

	return true
}

func updateDataExportRecord(task *TaskEntity, context *FlowContext) bool {
	getLogger().Info("begin updateDataExportRecord")
	defer getLogger().Info("end updateDataExportRecord")
	clusterAggregation := context.value(contextClusterKey).(*ClusterAggregation)
	info := context.value(contextDataTransportKey).(*ExportInfo)
	cluster := clusterAggregation.Cluster

	req := &db.DBUpdateTransportRecordRequest{
		Record: &db.TransportRecordDTO{
			ID:        info.RecordId,
			ClusterId: cluster.Id,
			Status:    TransportStatusSuccess,
			EndTime:   time.Now().Unix(),
		},
	}
	resp, err := client.DBClient.UpdateTransportRecord(ctx.Background(), req)
	if err != nil {
		getLogger().Errorf("update data transport record failed, %s", err.Error())
		return false
	}
	getLogger().Infof("update data transport record success, %v", resp)
	return true
}

/*
func compressExportData(task *TaskEntity, context *FlowContext) bool {
	getLogger().Info("begin compressExportData")
	defer getLogger().Info("end compressExportData")

	info := context.value(contextDataTransportKey).(*ExportInfo)

	dataDir := fmt.Sprintf("%s/data", info.FilePath)
	dataZipDir := fmt.Sprintf("%s/data.zip", info.FilePath)
	if err := zipDir(dataDir, dataZipDir); err != nil {
		getLogger().Errorf("compress export data failed, %s", err.Error())
		return false
	}

	return true
}

func deCompressImportData(task *TaskEntity, context *FlowContext) bool {
	getLogger().Info("begin deCompressImportData")
	defer getLogger().Info("end deCompressImportData")

	info := context.value(contextDataTransportKey).(*ImportInfo)

	dataDir := fmt.Sprintf("%s/data", info.ConfigPath)
	dataZipDir := info.FilePath
	if err := unzipDir(dataZipDir, dataDir); err != nil {
		getLogger().Errorf("deCompress import data failed, %s", err.Error())
		return false
	}

	return true
}
*/

func importDataFailed(task *TaskEntity, context *FlowContext) bool {
	getLogger().Info("begin importDataFailed")
	defer getLogger().Info("end importDataFailed")
	clusterAggregation := context.value(contextClusterKey).(ClusterAggregation)
	info := context.value(contextDataTransportKey).(*ImportInfo)
	cluster := clusterAggregation.Cluster

	if err := updateTransportRecordFailed(info.RecordId, cluster.Id); err != nil {
		return false
	}

	return ClusterFail(task, context)
}

func exportDataFailed(task *TaskEntity, context *FlowContext) bool {
	getLogger().Info("begin exportDataFailed")
	defer getLogger().Info("end exportDataFailed")
	clusterAggregation := context.value(contextClusterKey).(ClusterAggregation)
	info := context.value(contextDataTransportKey).(*ExportInfo)
	cluster := clusterAggregation.Cluster

	if err := updateTransportRecordFailed(info.RecordId, cluster.Id); err != nil {
		return false
	}

	return ClusterFail(task, context)
}

func updateTransportRecordFailed(recordId, clusterId string) error {
	req := &db.DBUpdateTransportRecordRequest{
		Record: &db.TransportRecordDTO{
			ID:        recordId,
			ClusterId: clusterId,
			Status:    TransportStatusFailed,
			EndTime:   time.Now().Unix(),
		},
	}
	resp, err := client.DBClient.UpdateTransportRecord(ctx.Background(), req)
	if err != nil {
		getLogger().Errorf("update data transport record failed, %s", err.Error())
		return err
	}
	getLogger().Infof("update data transport record success, %v", resp)
	return nil
}
/*
func zipDir(dir string, zipFile string) error {
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

func unzipDir(zipFile string, dir string) error {
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
*/