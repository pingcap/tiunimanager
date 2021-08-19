package domain

import (
	"archive/zip"
	ctx "context"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/pingcap/tiem/library/secondparty/libtiup"
	proto "github.com/pingcap/tiem/micro-cluster/proto"
	"github.com/pingcap/tiem/micro-metadb/client"
	db "github.com/pingcap/tiem/micro-metadb/proto"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
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
	ClusterId string
	UserName  string
	Password  string
	FilePath  string
	RecordId  string
}

type ExportInfo struct {
	ClusterId string
	UserName  string
	Password  string
	FileType  string
	RecordId  string
	FilePath  string
	Filter    string
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
	data import toml config for lighting
	https://docs.pingcap.com/zh/tidb/dev/deploy-tidb-lightning
*/
type DataImportConfig struct {
	Lighting     LightingCfg     `toml:"lighting"`
	TikvImporter TikvImporterCfg `toml:"tikv-importer"`
	MyDumper     MyDumperCfg     `toml:"mydumper"`
	Tidb         TidbCfg         `toml:"tidb"`
}

type LightingCfg struct {
	Level 				string	`toml:"level"` //lighting log level
	File  				string	`toml:"file"`  //lighting log path
	CheckRequirements 	bool	`toml:"check-requirements"`	//lightning pre check
}

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
var defaultTransportDirPrefix = "/tmp/tiem/datatransport"

func ExportData(ope *proto.OperatorDTO, clusterId string, userName string, password string, fileType string, filter string) (string, error) {
	log.Infof("[domain] begin exportdata clusterId: %s, userName: %s, password: %s, fileType: %s, filter: %s", clusterId, userName, password, fileType, filter)
	defer log.Infof("[domain] end exportdata")
	//todo: check operator
	operator := parseOperatorFromDTO(ope)
	log.Info(operator)
	clusterAggregation, err := ClusterRepo.Load(clusterId)

	req := &db.DBCreateTransportRecordRequest{
		Record: &db.TransportRecordDTO{
			ClusterId:     clusterId,
			TenantId:      operator.TenantId,
			TransportType: string(TransportTypeExport),
			FilePath:      fmt.Sprintf("%s/data.zip", getDataTransportDir(clusterId, TransportTypeExport)),
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
		FilePath: getDataTransportDir(clusterId, TransportTypeExport),
		Filter: filter,
	}

	// Start the workflow
	flow, err := CreateFlowWork(clusterId, FlowExportData)
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

func ImportData(ope *proto.OperatorDTO, clusterId string, userName string, password string, filepath string) (string, error) {
	log.Infof("[domain] begin importdata clusterId: %s, userName: %s, password: %s, datadIR: %s", clusterId, userName, password, filepath)
	defer log.Infof("[domain] end importdata")
	//todo: check operator
	operator := parseOperatorFromDTO(ope)
	log.Info(operator)
	clusterAggregation, err := ClusterRepo.Load(clusterId)

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
	}

	// Start the workflow
	flow, err := CreateFlowWork(clusterId, FlowImportData)
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

func DescribeDataTransportRecord(ope *proto.OperatorDTO, recordId, clusterId string, page, pageSize int32) ([]*TransportInfo, error) {
	log.Infof("[domain] begin DescribeDataTransportRecord clusterId: %s, recordId: %s, page: %d, pageSize: %s", clusterId, recordId, page, pageSize)
	defer log.Info("[domain] end DescribeDataTransportRecord")
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
		return nil, err
	}

	records := resp.GetRecords()
	info := make([]*TransportInfo, len(records))
	for index := 0; index < len(info); index++ {
		info[index] = &TransportInfo{
			RecordId:      records[index].GetID(),
			TransportType: records[index].GetTransportType(),
			ClusterId:     records[index].GetClusterId(),
			Status:        records[index].GetStatus(),
			FilePath:      records[index].GetFilePath(),
			StartTime:     records[index].GetStartTime(),
			EndTime:       records[index].GetEndTime(),
		}
	}

	return info, nil
}

func convertTomlConfig(clusterAggregation *ClusterAggregation, info *ImportInfo) *DataImportConfig {
	log.Info("begin convertTomlConfig")
	defer log.Info("end convertTomlConfig")
	if clusterAggregation == nil || clusterAggregation.CurrentTiUPConfigRecord == nil {
		return nil
	}
	configModel := clusterAggregation.CurrentTiUPConfigRecord.ConfigModel
	if configModel == nil || configModel.TiDBServers == nil || configModel.PDServers == nil {
		return nil
	}
	tidbServer := configModel.TiDBServers[0]
	pdServer := configModel.PDServers[0]

	config := &DataImportConfig{
		Lighting: LightingCfg{
			Level: "info",
			File:  fmt.Sprintf("%s/tidb-lighting.log", info.FilePath),
			CheckRequirements: false, //todo: need check
		},
		TikvImporter: TikvImporterCfg{
			Backend:     "local",
			SortedKvDir: "/mnt/ssd/sorted-kv-dir", //todo: replace config item
		},
		MyDumper: MyDumperCfg{
			DataSourceDir: info.FilePath,
		},
		Tidb: TidbCfg{
			Host:       tidbServer.Host,
			Port:       tidbServer.Port,
			User:       info.UserName,
			Password:   info.Password,
			StatusPort: tidbServer.StatusPort,
			PdAddr:     fmt.Sprintf("%s:%d", pdServer.Host, pdServer.ClientPort),
		},
	}
	return config
}

/**
data import && export dir
└── filePath/[cluster-id]
	├── import
	|	├── data
	|	├── data.zip
	|	├── tidb-lighting.toml
	|	└── log
	└── export
		├── data
		├── data.zip
		└── log
*/
func getDataTransportDir(clusterId string, transportType TransportType) string {
	return fmt.Sprintf("%s/%s/%s", defaultTransportDirPrefix, clusterId, transportType)
}

func cleanDataTransportDir(filepath string) error {
	log.Infof("clean and re-mkdir data dir: %s", filepath)
	if err := os.RemoveAll(filepath); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath, os.ModePerm); err != nil {
		return err
	}
	return nil
}

func zipDir(dir string, zipFile string) error {
	log.Infof("begin zipDir: dir[%s] --> file[%s]", dir, zipFile)
	defer log.Info("end zipDir")
	fz, err := os.Create(zipFile)
	if err != nil {
		return fmt.Errorf("Create zip file failed: %s", err.Error())
	}
	defer fz.Close()

	w := zip.NewWriter(fz)
	defer w.Close()

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			fDest, err := w.Create(path)
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

	return nil
}

func unzipDir(zipFile string, dir string) error {
	log.Infof("begin unzipDir: file[%s] --> dir[%s]", zipFile, dir)
	defer log.Info("end unzipDir")
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return fmt.Errorf("Open zip file failed: %s", err.Error())
	}
	defer r.Close()

	for _, f := range r.File {
		func() {
			path := dir + string(filepath.Separator) + f.Name
			os.MkdirAll(filepath.Dir(path), 0755)
			fDest, err := os.Create(path)
			if err != nil {
				log.Errorf("unzip Create failed: %s", err.Error())
				return
			}
			defer fDest.Close()

			fSrc, err := f.Open()
			if err != nil {
				log.Errorf("unzip Open failed: %s", err.Error())
				return
			}
			defer fSrc.Close()

			_, err = io.Copy(fDest, fSrc)
			if err != nil {
				log.Errorf("unzip Copy failed: %s", err.Error())
				return
			}
		}()
	}
	return nil
}

func buildDataImportConfig(task *TaskEntity, context *FlowContext) bool {
	log.Info("begin buildDataImportConfig")
	defer log.Info("end buildDataImportConfig")
	clusterAggregation := context.value(contextClusterKey).(*ClusterAggregation)
	info := context.value(contextDataTransportKey).(*ImportInfo)
	cluster := clusterAggregation.Cluster

	if err := cleanDataTransportDir(getDataTransportDir(cluster.Id, TransportTypeImport)); err != nil {
		log.Errorf("[domain] clean import directory failed, %s", err.Error())
		return false
	}

	config := convertTomlConfig(clusterAggregation, info)
	if config == nil {
		log.Errorf("[domain] convert toml config failed, cluster: %v", clusterAggregation)
		return false
	}
	filePath := fmt.Sprintf("%s/tidb-lighting.toml", info.FilePath)
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0766)
	if err != nil {
		log.Errorf("[domain] create import toml config failed, %s", err.Error())
		return false
	}

	if err = toml.NewEncoder(file).Encode(config); err != nil {
		log.Errorf("[domain] decode data import toml config failed, %s", err.Error())
		return false
	}
	log.Infof("build lighting toml file sucess, %v", config)
	return true
}

func importDataToCluster(task *TaskEntity, context *FlowContext) bool {
	log.Info("begin importDataToCluster")
	defer log.Info("end importDataToCluster")
	info := context.value(contextDataTransportKey).(*ImportInfo)

	//tiup tidb-lightning -config tidb-lightning.toml
	_, err := libtiup.MicroSrvTiupLightning(0,
		[]string{"-config", fmt.Sprintf("%s/tidb-lighting.toml", info.FilePath)},
		uint64(task.Id))
	if err != nil {
		log.Errorf("[domain] call tiup lighting api failed, %s", err.Error())
		return false
	}

	return true
}

func updateDataImportRecord(task *TaskEntity, context *FlowContext) bool {
	log.Info("begin updateDataImportRecord")
	defer log.Info("end updateDataImportRecord")
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
		log.Errorf("[domain] update data transport record failed, %s", err.Error())
		return false
	}
	log.Infof("[domain] update data transport record success, %v", resp)
	return true
}

func exportDataFromCluster(task *TaskEntity, context *FlowContext) bool {
	log.Info("begin exportDataFromCluster")
	defer log.Info("end exportDataFromCluster")
	clusterAggregation := context.value(contextClusterKey).(*ClusterAggregation)
	info := context.value(contextDataTransportKey).(*ExportInfo)
	configModel := clusterAggregation.CurrentTiUPConfigRecord.ConfigModel
	tidbServer := configModel.TiDBServers[0]

	if err := cleanDataTransportDir(info.FilePath); err != nil {
		log.Errorf("[domain] clean export directory failed, %s", err.Error())
		return false
	}

	//tiup dumpling -u root -P 4000 --host 127.0.0.1 --filetype sql -t 8 -o /tmp/test -r 200000 -F 256MiB --filter "user*"
	//todo: admin root password
	cmd :=  []string{"-u", info.UserName,
		"-p", info.Password,
		"-P", strconv.Itoa(tidbServer.Port),
		"--host", tidbServer.Host,
		"--filetype", info.FileType,
		"-t", "8",
		"-o", fmt.Sprintf("%s/data" ,info.FilePath),
		"-r", "200000",
		"-F", "256MiB"}
	if info.Filter != "" {
		cmd = append(cmd, "--filter", fmt.Sprintf("\"%s\"", info.Filter))
	}
	log.Infof("call tiupmgr dumpling api, cmd: %v", cmd)
	resp, err := libtiup.MicroSrvTiupDumpling(0, cmd, uint64(task.Id))
	if err != nil {
		log.Errorf("[domain] call tiup dumpling api failed, %s", err.Error())
		return false
	}
	log.Infof("call tiupmgr succee, resp: %v", resp)

	return true
}

func updateDataExportRecord(task *TaskEntity, context *FlowContext) bool {
	log.Info("begin updateDataExportRecord")
	defer log.Info("end updateDataExportRecord")
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
		log.Errorf("[domain] update data transport record failed, %s", err.Error())
		return false
	}
	log.Infof("[domain] update data transport record success, %v", resp)
	return true
}

func compressExportData(task *TaskEntity, context *FlowContext) bool {
	log.Info("begin compressExportData")
	defer log.Info("end compressExportData")
	info := context.value(contextDataTransportKey).(*ExportInfo)

	dataDir := fmt.Sprintf("%s/data", info.FilePath)
	dataZipDir := fmt.Sprintf("%s/data.zip", info.FilePath)
	if err := zipDir(dataDir, dataZipDir); err != nil {
		log.Errorf("[domain] compress export data failed, %s", err.Error())
		return false
	}

	return true
}

func deCompressImportData(task *TaskEntity, context *FlowContext) bool {
	log.Info("begin deCompressImportData")
	defer log.Info("end deCompressImportData")
	clusterAggregation := context.value(contextClusterKey).(*ClusterAggregation)
	info := context.value(contextDataTransportKey).(*ImportInfo)
	cluster := clusterAggregation.Cluster

	dataDir := fmt.Sprintf("%s", getDataTransportDir(cluster.Id, TransportTypeImport))
	dataZipDir := fmt.Sprintf("%s", info.FilePath)
	if err := unzipDir(dataZipDir, dataDir); err != nil {
		log.Errorf("[domain] deCompress import data failed, %s", err.Error())
		return false
	}

	return true
}

func importDataFailed(task *TaskEntity, context *FlowContext) bool {
	log.Info("begin importDataFailed")
	defer log.Info("end importDataFailed")
	clusterAggregation := context.value(contextClusterKey).(ClusterAggregation)
	info := context.value(contextDataTransportKey).(*ImportInfo)
	cluster := clusterAggregation.Cluster

	if err := updateTransportRecordFailed(info.RecordId, cluster.Id); err != nil {
		return false
	}

	return DefaultFail(task, context)
}

func exportDataFailed(task *TaskEntity, context *FlowContext) bool {
	log.Info("begin exportDataFailed")
	defer log.Info("end exportDataFailed")
	clusterAggregation := context.value(contextClusterKey).(ClusterAggregation)
	info := context.value(contextDataTransportKey).(*ExportInfo)
	cluster := clusterAggregation.Cluster

	if err := updateTransportRecordFailed(info.RecordId, cluster.Id); err != nil {
		return false
	}

	return DefaultFail(task, context)
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
		log.Errorf("[domain] update data transport record failed, %s", err.Error())
		return err
	}
	log.Infof("[domain] update data transport record success, %v", resp)
	return nil
}
