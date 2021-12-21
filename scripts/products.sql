/*
 * Copyright (c)  2021 PingCAP, Inc.
 * Licensed under the Apache License, Version 2.0 (the 'License',datetime('now'),datetime('now'));
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an 'AS IS' BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

CREATE TABLE IF NOT EXISTS `vendors`
(
    `vendor_id`  char(32) primary key,
    `name`       char(32)   DEFAULT NULL,
    `comment`    char(1024) DEFAULT NULL,
    `created_at` DATETIME   DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME   DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO Vendors
VALUES ('Aliyun', 'Aliyun', 'Alibaba cloud', datetime('now'), datetime('now'));

CREATE TABLE IF NOT EXISTS `regions`
(
    `vendor_id`  char(32) not null,
    `region_id`  char(32) not null,
    `name`       char(32)   DEFAULT NULL,
    `comment`    char(1024) DEFAULT NULL,
    `created_at` DATETIME   DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME   DEFAULT CURRENT_TIMESTAMP,
    primary key (vendor_id, region_id)
);

INSERT INTO Regions
VALUES ('Aliyun','CN-HANGZHOU', 'East China(Hangzhou)', 'default region', datetime('now'), datetime('now'));
INSERT INTO Regions
VALUES ('Aliyun','CN-BEIJING', 'North China(Beijing)', 'default region', datetime('now'), datetime('now'));

CREATE TABLE IF NOT EXISTS `zones`
(
    `vendor_id`  char(32) not null,
    `region_id`  char(32) not null,
    `zone_id`    char(32) not null,
    `name`       char(32)   DEFAULT NULL,
    `comment`    char(1024) DEFAULT NULL,
    `created_at` DATETIME   DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME   DEFAULT CURRENT_TIMESTAMP,
    primary key (vendor_id, region_id, zone_id)
);

INSERT INTO Zones
VALUES ('Aliyun','CN-BEIJING', 'CN-BEIJING-H', 'ZONE(H)', 'ch-beijing-h', datetime('now'), datetime('now'));
INSERT INTO Zones
VALUES ('Aliyun','CN-BEIJING', 'CN-BEIJING-G', 'ZONE(G)', 'ch-beijing-g', datetime('now'), datetime('now'));
INSERT INTO Zones
VALUES ('Aliyun','CN-HANGZHOU', 'CN-HANGZHOU-H', 'ZONE(H)', 'ch-hangzhou-h', datetime('now'), datetime('now'));
INSERT INTO Zones
VALUES ('Aliyun','CN-HANGZHOU', 'CN-HANGZHOU-G', 'ZONE(G)', 'ch-hangzhou-h', datetime('now'), datetime('now'));

CREATE TABLE IF NOT EXISTS `resource_specs`
(
    `resource_spec_id` char(32) not null,
    `zone_id`          char(32) not null,
    `arch`             char(32) not null,
    `name`             char(32) not null,
    `cpu`              int      DEFAULT NULL,
    `memory`           int      DEFAULT NULL,
    `disk_type`        char(32) DEFAULT NULL,
    `purpose_type`     char(32) DEFAULT NULL,
    `status`           char(32) DEFAULT NULL,
    `created_at`       DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at`       DATETIME DEFAULT CURRENT_TIMESTAMP,
    primary key (resource_spec_id, zone_id, arch)
);

/**CN-BEIJING-H X86_64**/
INSERT INTO resource_specs
VALUES ('c2.g.large', 'CN-BEIJING-H', 'x86_64', '4C8G', 4, 8, 'SSD', 'Compute', 'Online', datetime('now'),
        datetime('now'));
INSERT INTO resource_specs
VALUES ('c2.g.xlarge', 'CN-BEIJING-H', 'x86_64', '8C16G', 8, 16, 'SSD', 'Compute', 'Online', datetime('now'),
        datetime('now'));
INSERT INTO resource_specs
VALUES ('c2.g.2xlarge', 'CN-BEIJING-H', 'x86_64', '16C32G', 16, 32, 'SSD', 'Compute', 'Online', datetime('now'),
        datetime('now'));
INSERT INTO resource_specs
VALUES ('c3.g.large', 'CN-BEIJING-H', 'x86_64', '8C32G', 8, 32, 'SSD', 'Compute', 'Online', datetime('now'),
        datetime('now'));
INSERT INTO resource_specs
VALUES ('c3.g.xlarge', 'CN-BEIJING-H', 'x86_64', '16C64G', 16, 64, 'SSD', 'Compute', 'Online', datetime('now'),
        datetime('now'));
INSERT INTO resource_specs
VALUES ('sd2.g.large', 'CN-BEIJING-H', 'x86_64', '4C8G', 4, 8, 'SSD', 'Schedule', 'Online', datetime('now'),
        datetime('now'));
INSERT INTO resource_specs
VALUES ('sd2.g.xlarge', 'CN-BEIJING-H', 'x86_64', '8C16G', 8, 16, 'SSD', 'Schedule', 'Online', datetime('now'),
        datetime('now'));
INSERT INTO resource_specs
VALUES ('s2.g.large', 'CN-BEIJING-H', 'x86_64', '8C64G', 8, 64, 'NVMeSSD', 'Storage', 'Online', datetime('now'),
        datetime('now'));
INSERT INTO resource_specs
VALUES ('s2.g.xlarge', 'CN-BEIJING-H', 'x86_64', '16C128G', 16, 128, 'NVMeSSD', 'Storage', 'Online', datetime('now'),
        datetime('now'));

/**CN-BEIJING-G X86_64**/
INSERT INTO resource_specs
VALUES ('c2.g.large', 'CN-BEIJING-G', 'x86_64', '4C8G', 4, 8, 'SSD', 'Compute', 'Online', datetime('now'),
        datetime('now'));
INSERT INTO resource_specs
VALUES ('c2.g.xlarge', 'CN-BEIJING-G', 'x86_64', '8C16G', 8, 16, 'SSD', 'Compute', 'Online', datetime('now'),
        datetime('now'));
INSERT INTO resource_specs
VALUES ('c2.g.2xlarge', 'CN-BEIJING-G', 'x86_64', '16C32G', 16, 32, 'SSD', 'Compute', 'Online', datetime('now'),
        datetime('now'));
INSERT INTO resource_specs
VALUES ('c3.g.large', 'CN-BEIJING-G', 'x86_64', '8C32G', 8, 32, 'SSD', 'Compute', 'Online', datetime('now'),
        datetime('now'));
INSERT INTO resource_specs
VALUES ('c3.g.xlarge', 'CN-BEIJING-G', 'x86_64', '16C64G', 16, 64, 'SSD', 'Compute', 'Online', datetime('now'),
        datetime('now'));

INSERT INTO resource_specs
VALUES ('sd2.g.large', 'CN-BEIJING-G', 'x86_64', '4C8G', 4, 8, 'SSD', 'Schedule', 'Online', datetime('now'),
        datetime('now'));
INSERT INTO resource_specs
VALUES ('sd2.g.xlarge', 'CN-BEIJING-G', 'x86_64', '8C16G', 8, 16, 'SSD', 'Schedule', 'Online', datetime('now'),
        datetime('now'));

INSERT INTO resource_specs
VALUES ('s2.g.large', 'CN-BEIJING-G', 'x86_64', '8C64G', 8, 64, 'NVMeSSD', 'Storage', 'Online', datetime('now'),
        datetime('now'));
INSERT INTO resource_specs
VALUES ('s2.g.xlarge', 'CN-BEIJING-G', 'x86_64', '16C128G', 16, 128, 'NVMeSSD', 'Storage', 'Online', datetime('now'),
        datetime('now'));

/**CN-HANGZHOU-H X86_64**/
INSERT INTO resource_specs
VALUES ('c1.g.large', 'CN-HANGZHOU-H', 'x86_64', '2C2G', 2, 2, 'SSD', 'Compute', 'Online', datetime('now'),
        datetime('now'));
INSERT INTO resource_specs
VALUES ('c2.g.large', 'CN-HANGZHOU-H', 'x86_64', '4C8G', 4, 8, 'SSD', 'Compute', 'Online', datetime('now'),
        datetime('now'));
INSERT INTO resource_specs
VALUES ('c2.g.xlarge', 'CN-HANGZHOU-H', 'x86_64', '8C16G', 8, 16, 'SSD', 'Compute', 'Online', datetime('now'),
        datetime('now'));
INSERT INTO resource_specs
VALUES ('c2.g.2xlarge', 'CN-HANGZHOU-H', 'x86_64', '16C32G', 16, 32, 'SSD', 'Compute', 'Online', datetime('now'),
        datetime('now'));
INSERT INTO resource_specs
VALUES ('c3.g.large', 'CN-HANGZHOU-H', 'x86_64', '8C32G', 8, 32, 'SSD', 'Compute', 'Online', datetime('now'),
        datetime('now'));
INSERT INTO resource_specs
VALUES ('c3.g.xlarge', 'CN-HANGZHOU-H', 'x86_64', '16C64G', 16, 64, 'SSD', 'Compute', 'Online', datetime('now'),
        datetime('now'));

INSERT INTO resource_specs
VALUES ('sd1.g.large', 'CN-HANGZHOU-H', 'x86_64', '2C2G', 2, 2, 'SSD', 'Schedule', 'Online', datetime('now'),
        datetime('now'));
INSERT INTO resource_specs
VALUES ('sd2.g.large', 'CN-HANGZHOU-H', 'x86_64', '4C8G', 4, 8, 'SSD', 'Schedule', 'Online', datetime('now'),
        datetime('now'));
INSERT INTO resource_specs
VALUES ('sd2.g.xlarge', 'CN-HANGZHOU-H', 'x86_64', '8C16G', 8, 16, 'SSD', 'Schedule', 'Online', datetime('now'),
        datetime('now'));

INSERT INTO resource_specs
VALUES ('s1.g.large', 'CN-HANGZHOU-H', 'x86_64', '2C2G', 2, 2, 'NVMeSSD', 'Storage', 'Online', datetime('now'),
        datetime('now'));
INSERT INTO resource_specs
VALUES ('s2.g.large', 'CN-HANGZHOU-H', 'x86_64', '8C64G', 8, 64, 'NVMeSSD', 'Storage', 'Online', datetime('now'),
        datetime('now'));
INSERT INTO resource_specs
VALUES ('s2.g.xlarge', 'CN-HANGZHOU-H', 'x86_64', '16C128G', 16, 128, 'NVMeSSD', 'Storage', 'Online',
        datetime('now'), datetime('now'));
INSERT INTO resource_specs
VALUES ('cs1.g.large', 'CN-HANGZHOU-H', 'x86_64', '2C2G', 2, 2, 'SATA', 'Storage', 'Online', datetime('now'),
        datetime('now'));
INSERT INTO resource_specs
VALUES ('cs2.g.large', 'CN-HANGZHOU-H', 'x86_64', '8C64G', 8, 64, 'SATA', 'Storage', 'Online', datetime('now'),
        datetime('now'));
INSERT INTO resource_specs
VALUES ('cs2.g.xlarge', 'CN-HANGZHOU-H', 'x86_64', '16C128G', 16, 128, 'SATA', 'Storage', 'Online', datetime('now'),
        datetime('now'));

/**CN-HANGZHOU-H ARM64*/
INSERT INTO resource_specs
VALUES ('c1.g.large', 'CN-HANGZHOU-H', 'ARM64', '2C2G', 2, 2, 'SSD', 'Compute', 'Online', datetime('now'),
        datetime('now'));
INSERT INTO resource_specs
VALUES ('c2.g.large', 'CN-HANGZHOU-H', 'ARM64', '4C8G', 4, 8, 'SSD', 'Compute', 'Online', datetime('now'),
        datetime('now'));
INSERT INTO resource_specs
VALUES ('c2.g.xlarge', 'CN-HANGZHOU-H', 'ARM64', '8C16G', 8, 16, 'SSD', 'Compute', 'Online', datetime('now'),
        datetime('now'));
INSERT INTO resource_specs
VALUES ('c2.g.2xlarge', 'CN-HANGZHOU-H', 'ARM64', '16C32G', 16, 32, 'SSD', 'Compute', 'Online', datetime('now'),
        datetime('now'));
INSERT INTO resource_specs
VALUES ('c3.g.large', 'CN-HANGZHOU-H', 'ARM64', '8C32G', 8, 32, 'SSD', 'Compute', 'Online', datetime('now'),
        datetime('now'));
INSERT INTO resource_specs
VALUES ('c3.g.xlarge', 'CN-HANGZHOU-H', 'ARM64', '16C64G', 16, 64, 'SSD', 'Compute', 'Online', datetime('now'),
        datetime('now'));

INSERT INTO resource_specs
VALUES ('sd1.g.large', 'CN-HANGZHOU-H', 'ARM64', '2C2G', 2, 2, 'SSD', 'Schedule', 'Online', datetime('now'),
        datetime('now'));
INSERT INTO resource_specs
VALUES ('sd2.g.large', 'CN-HANGZHOU-H', 'ARM64', '4C8G', 4, 8, 'SSD', 'Schedule', 'Online', datetime('now'),
        datetime('now'));
INSERT INTO resource_specs
VALUES ('sd2.g.xlarge', 'CN-HANGZHOU-H', 'ARM64', '8C16G', 8, 16, 'SSD', 'Schedule', 'Online', datetime('now'),
        datetime('now'));

INSERT INTO resource_specs
VALUES ('s1.g.large', 'CN-HANGZHOU-H', 'ARM64', '2C2G', 2, 2, 'NVMeSSD', 'Storage', 'Online', datetime('now'),
        datetime('now'));
INSERT INTO resource_specs
VALUES ('s2.g.large', 'CN-HANGZHOU-H', 'ARM64', '8C64G', 8, 64, 'NVMeSSD', 'Storage', 'Online', datetime('now'),
        datetime('now'));
INSERT INTO resource_specs
VALUES ('s2.g.xlarge', 'CN-HANGZHOU-H', 'ARM64', '16C128G', 16, 128, 'NVMeSSD', 'Storage', 'Online', datetime('now'),
        datetime('now'));
INSERT INTO resource_specs
VALUES ('cs1.g.large', 'CN-HANGZHOU-H', 'ARM64', '2C2G', 2, 2, 'SATA', 'Storage', 'Online', datetime('now'),
        datetime('now'));
INSERT INTO resource_specs
VALUES ('cs2.g.large', 'CN-HANGZHOU-H', 'ARM64', '8C64G', 8, 64, 'SATA', 'Storage', 'Online', datetime('now'),
        datetime('now'));
INSERT INTO resource_specs
VALUES ('cs2.g.xlarge', 'CN-HANGZHOU-H', 'ARM64', '16C128G', 16, 128, 'SATA', 'Storage', 'Online', datetime('now'),
        datetime('now'));

/**CN-HANGZHOU-G X86_64*/
INSERT INTO resource_specs
VALUES ('c1.g.large', 'CN-HANGZHOU-G', 'x86_64', '2C2G', 2, 2, 'SSD', 'Compute', 'Online', datetime('now'),
        datetime('now'));
INSERT INTO resource_specs
VALUES ('c2.g.large', 'CN-HANGZHOU-G', 'x86_64', '4C8G', 4, 8, 'SSD', 'Compute', 'Online', datetime('now'),
        datetime('now'));
INSERT INTO resource_specs
VALUES ('c2.g.xlarge', 'CN-HANGZHOU-G', 'x86_64', '8C16G', 8, 16, 'SSD', 'Compute', 'Online', datetime('now'),
        datetime('now'));
INSERT INTO resource_specs
VALUES ('c2.g.2xlarge', 'CN-HANGZHOU-G', 'x86_64', '16C32G', 16, 32, 'SSD', 'Compute', 'Online', datetime('now'),
        datetime('now'));
INSERT INTO resource_specs
VALUES ('c3.g.large', 'CN-HANGZHOU-G', 'x86_64', '8C32G', 8, 32, 'SSD', 'Compute', 'Online', datetime('now'),
        datetime('now'));
INSERT INTO resource_specs
VALUES ('c3.g.xlarge', 'CN-HANGZHOU-G', 'x86_64', '16C64G', 16, 64, 'SSD', 'Compute', 'Online', datetime('now'),
        datetime('now'));

INSERT INTO resource_specs
VALUES ('sd1.g.large', 'CN-HANGZHOU-G', 'x86_64', '2C2G', 2, 2, 'SSD', 'Schedule', 'Online', datetime('now'),
        datetime('now'));
INSERT INTO resource_specs
VALUES ('sd2.g.large', 'CN-HANGZHOU-G', 'x86_64', '4C8G', 4, 8, 'SSD', 'Schedule', 'Online', datetime('now'),
        datetime('now'));
INSERT INTO resource_specs
VALUES ('sd2.g.xlarge', 'CN-HANGZHOU-G', 'x86_64', '8C16G', 8, 16, 'SSD', 'Schedule', 'Online', datetime('now'),
        datetime('now'));

INSERT INTO resource_specs
VALUES ('s1.g.large', 'CN-HANGZHOU-G', 'x86_64', '2C2G', 2, 2, 'NVMeSSD', 'Storage', 'Online', datetime('now'),
        datetime('now'));
INSERT INTO resource_specs
VALUES ('s2.g.large', 'CN-HANGZHOU-G', 'x86_64', '8C64G', 8, 64, 'NVMeSSD', 'Storage', 'Online', datetime('now'),
        datetime('now'));
INSERT INTO resource_specs
VALUES ('s2.g.xlarge', 'CN-HANGZHOU-G', 'x86_64', '16C128G', 16, 128, 'NVMeSSD', 'Storage', 'Online',
        datetime('now'), datetime('now'));
INSERT INTO resource_specs
VALUES ('cs1.g.large', 'CN-HANGZHOU-G', 'x86_64', '2C2G', 2, 2, 'SATA', 'Storage', 'Online', datetime('now'),
        datetime('now'));
INSERT INTO resource_specs
VALUES ('cs2.g.large', 'CN-HANGZHOU-G', 'x86_64', '8C64G', 8, 64, 'SATA', 'Storage', 'Online', datetime('now'),
        datetime('now'));
INSERT INTO resource_specs
VALUES ('cs2.g.xlarge', 'CN-HANGZHOU-G', 'x86_64', '16C128G', 16, 128, 'SATA', 'Storage', 'Online', datetime('now'),
        datetime('now'));

CREATE TABLE IF NOT EXISTS `products`
(
    `vendor_id`  char(32) DEFAULT NULL,
    `region_id`  char(32) DEFAULT NULL,
    `product_id` char(32) DEFAULT NULL,
    `name`       char(32) DEFAULT NULL,
    `version`    char(32) DEFAULT NULL,
    `arch`       char(32) DEFAULT NULL,
    `status`     char(32) DEFAULT NULL,
    `internal`   int      DEFAULT 0,
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    primary key (vendor_id, region_id, product_id, version, arch)
);

INSERT INTO products
VALUES ('Aliyun','CN-HANGZHOU', 'TiDB', 'TiDB', '5.0.0', 'x86_64', 'Online', 0, datetime('now'), datetime('now'));
INSERT INTO products
VALUES ('Aliyun','CN-HANGZHOU', 'TiDB', 'TiDB', '5.1.0', 'ARM64', 'Online', 0, datetime('now'), datetime('now'));
INSERT INTO products
VALUES ('Aliyun','CN-BEIJING', 'TiDB', 'TiDB', '5.0.0', 'x86_64', 'Online', 0, datetime('now'), datetime('now'));
INSERT INTO products
VALUES ('Aliyun','CN-BEIJING', 'TiDB', 'TiDB', '5.1.0', 'x86_64', 'Online', 0, datetime('now'), datetime('now'));
INSERT INTO products
VALUES ('Aliyun','CN-HANGZHOU', 'EnterpriseManager', 'EnterpriseManager', '1.0.0', 'x86_64', 'Online', 1, datetime('now'),
        datetime('now'));
/**INSERT INTO products VALUES('CN-HANGZHOU','TiDB Data Migration','Data Migration','2.0.0','x86_64','UnOnline',datetime('now'),datetime('now'));*/


CREATE TABLE IF NOT EXISTS `product_components`
(
    `component_id`    char(32) NOT NULL,
    `product_id`      char(32) NOT NULL,
    `product_version` char(32) NOT NULL,
    `name`            char(32) NOT NULL,
    `status`          char(32) DEFAULT NULL,
    `purpose_type`    char(32) DEFAULT NULL,
    `start_port`      int,
    `end_port`        int,
    `max_port`        int,
    `min_instance`    int      DEFAULT 1,
    `max_instance`    int      DEFAULT 1,
    `created_at`      DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at`      DATETIME DEFAULT CURRENT_TIMESTAMP,
    primary key (component_id,product_id,product_version)
);

INSERT INTO product_components
VALUES ('tidb', 'TiDB', '5.0.0', 'Compute Engine', 'Online', 'Compute', 10000, 10020, 2, 1, 10240, datetime('now'),
        datetime('now'));
INSERT INTO product_components
VALUES ('tikv', 'TiDB', '5.0.0', 'Storage Engine', 'Online', 'Storage', 10020, 10040, 2, 1, 10240, datetime('now'),
        datetime('now'));
INSERT INTO product_components
VALUES ('tiflash', 'TiDB', '5.0.0', 'Column Storage Engine', 'Online', 'Storage', 10120, 10180, 6, 1, 10240,
        datetime('now'), datetime('now'));
INSERT INTO product_components
VALUES ('pd', 'TiDB', '5.0.0', 'Schedule Engine', 'Online', 'Schedule', 10040, 10120, 8, 1, 7, datetime('now'),
        datetime('now'));
INSERT INTO product_components
VALUES ('cdc', 'TiDB', '5.0.0', 'CDC', 'Online', 'Schedule', 10180, 10200, 2, 1, 512, datetime('now'),
        datetime('now'));
INSERT INTO product_components
VALUES ('grafana', 'TiDB', '5.0.0', 'Monitor GUI', 'Online', 'Schedule', 10040, 10120, 1, 1, 1, datetime('now'),
        datetime('now'));
INSERT INTO product_components
VALUES ('prometheus', 'TiDB', '5.0.0', 'Monitor', 'Online', 'Schedule', 10040, 10120, 1, 1, 1, datetime('now'),
        datetime('now'));
INSERT INTO product_components
VALUES ('alertmanager', 'TiDB', '5.0.0', 'Alert', 'Online', 'Schedule', 10040, 10120, 1, 1, 1, datetime('now'),
        datetime('now'));
INSERT INTO product_components
VALUES ('node_exporter', 'TiDB', '5.0.0', 'NodeExporter', 'Online', 'Schedule', 11000, 12000, 2, 1, 1,
        datetime('now'), datetime('now'));
INSERT INTO product_components
VALUES ('blackbox_exporter', 'TiDB', '5.0.0', 'BlackboxExporter', 'Online', 'Schedule', 11000, 12000, 1, 1, 1,
        datetime('now'), datetime('now'));

INSERT INTO product_components
VALUES ('tidb', 'TiDB', '5.1.0', 'Compute Engine', 'Online', 'Compute', 10000, 10020, 2, 1, 10240, datetime('now'),
        datetime('now'));
INSERT INTO product_components
VALUES ('tikv', 'TiDB', '5.1.0', 'Storage Engine', 'Online', 'Storage', 10020, 10040, 2, 1, 10240, datetime('now'),
        datetime('now'));
INSERT INTO product_components
VALUES ('tiflash', 'TiDB', '5.1.0', 'Column Storage Engine', 'Online', 'Storage', 10120, 10180, 6, 1, 10240,
        datetime('now'), datetime('now'));
INSERT INTO product_components
VALUES ('pd', 'TiDB', '5.1.0', 'Schedule Engine', 'Online', 'Schedule', 10040, 10120, 8, 1, 7, datetime('now'),
        datetime('now'));
INSERT INTO product_components
VALUES ('cdc', 'TiDB', '5.1.0', 'CDC', 'Online', 'Schedule', 10180, 10200, 2, 1, 512, datetime('now'),
        datetime('now'));
INSERT INTO product_components
VALUES ('grafana', 'TiDB', '5.1.0', 'Monitor GUI', 'Online', 'Schedule', 10040, 10120, 1, 1, 1, datetime('now'),
        datetime('now'));
INSERT INTO product_components
VALUES ('prometheus', 'TiDB', '5.1.0', 'Monitor', 'Online', 'Schedule', 10040, 10120, 1, 1, 1, datetime('now'),
        datetime('now'));
INSERT INTO product_components
VALUES ('alertmanager', 'TiDB', '5.1.0', 'Alert', 'Online', 'Schedule', 10040, 10120, 1, 1, 1, datetime('now'),
        datetime('now'));
INSERT INTO product_components
VALUES ('node_exporter', 'TiDB', '5.1.0', 'NodeExporter', 'Online', 'Schedule', 11000, 12000, 2, 1, 1,
        datetime('now'), datetime('now'));
INSERT INTO product_components
VALUES ('blackbox_exporter', 'TiDB', '5.1.0', 'BlackboxExporter', 'Online', 'Schedule', 11000, 12000, 1, 1, 1,
        datetime('now'), datetime('now'));

INSERT INTO product_components
VALUES ('cluster-server', 'EnterpriseManager', '1.0.0', 'cluster-server', 'Online', 'Schedule', 10000, 10400, 2, 1,
        1, datetime('now'), datetime('now'));
INSERT INTO product_components
VALUES ('openapi-server', 'EnterpriseManager', '1.0.0', 'openapi-server', 'Online', 'Storage', 10000, 10400, 2, 1,
        1, datetime('now'), datetime('now'));
INSERT INTO product_components
VALUES ('grafana', 'EnterpriseManager', '1.0.0', 'Monitor GUI', 'Online', 'Schedule', 10000, 10400, 2, 1, 1,
        datetime('now'), datetime('now'));
INSERT INTO product_components
VALUES ('prometheus', 'EnterpriseManager', '1.0.0', 'Monitor', 'Online', 'Schedule', 10000, 10400, 2, 1, 1,
        datetime('now'), datetime('now'));
INSERT INTO product_components
VALUES ('alertmanager', 'EnterpriseManager', '1.0.0', 'Alert GUI', 'Online', 'Schedule', 10000, 10400, 2, 1, 1,
        datetime('now'), datetime('now'));

