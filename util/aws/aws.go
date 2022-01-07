/******************************************************************************
 * Copyright (c)  2022 PingCAP, Inc.                                          *
 * Licensed under the Apache License, Version 2.0 (the "License");            *
 * you may not use this file except in compliance with the License.           *
 * You may obtain a copy of the License at                                    *
 *                                                                            *
 * http://www.apache.org/licenses/LICENSE-2.0                                 *
 *                                                                            *
 * Unless required by applicable law or agreed to in writing, software        *
 * distributed under the License is distributed on an "AS IS" BASIS,          *
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.   *
 * See the License for the specific language governing permissions and        *
 * limitations under the License.                                             *
 *                                                                            *
 ******************************************************************************/

/*******************************************************************************
 * @File: aws.go
 * @Description:
 * @Author: jiangxunyu@pingcap.com
 * @Version: 1.0.0
 * @Date: 2022/1/7 10:52
*******************************************************************************/

package aws

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pingcap-inc/tiem/micro-cluster/resourcemanager/management/structs"
	"github.com/pingcap-inc/tiem/util/convert"
	"github.com/pingcap-inc/tiem/util/uuidutil"
)

const (
	accessKey     string = "YOUR-ACCESS-KEY"
	secretKey     string = "YOUR-SECRET-KEY"
	defaultRegion string = "us-west-2"
	defaultUser   string = "ec2-user"
)

var sess *session.Session

func newSession() *session.Session {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(defaultRegion),
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
	})
	if err != nil {
		panic("aws sdk new session error = " + err.Error())
	}
	return sess
}

func getSession() *session.Session {
	if sess == nil {
		sess = newSession()
	}
	return sess
}

func DescribeCloudInstances(region string, instanceIds []*string) (instancesInfo InstancesInfo, err error) {
	// set region
	getSession().Config.Region = aws.String(region)
	svc := ec2.New(getSession())

	// Retrieves all regions/endpoints that work with EC2
	result, err := svc.DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: instanceIds,
	})
	if err != nil {
		return
	}

	err = convert.ConvertObj(result, &instancesInfo)
	if err != nil {
		return
	}
	return instancesInfo, nil
}

func CreateCloudInstances(region string, count int32) (instances []structs.Compute, err error) {
	// set region
	getSession().Config.Region = aws.String(region)
	svc := ec2.New(getSession())

	// Specify the details of the instance that you want to create.
	runResult, err := svc.RunInstances(&ec2.RunInstancesInput{
		ImageId:          aws.String("ami-00f7e5c52c0f43726"),
		InstanceType:     aws.String("t2.large"),
		MinCount:         aws.Int64(int64(count)),
		MaxCount:         aws.Int64(int64(count)),
		SecurityGroupIds: []*string{aws.String("sg-07a3e9cbdbd4c6e9c")},
		KeyName:          aws.String("testkey"),
		SubnetId:         aws.String("subnet-07c9b4ef33a0949bb"),
	})
	if err != nil {
		return
	}

	instances = make([]structs.Compute, count)
	for i := 0; i < int(count); i++ {
		instanceName := fmt.Sprintf("%s-%d-%s", region, i, uuidutil.GenerateID())

		// Add tags to the created instance
		_, err := svc.CreateTags(&ec2.CreateTagsInput{
			Resources: []*string{runResult.Instances[i].InstanceId},
			Tags: []*ec2.Tag{
				{
					Key:   aws.String("Name"),
					Value: aws.String(instanceName),
				},
			},
		})
		if err != nil {
			return instances, err
		}

		instances[i] = structs.Compute{
			HostId:   *runResult.Instances[i].InstanceId,
			HostName: instanceName,
			HostIp:   *runResult.Instances[i].PrivateIpAddress,
			UserName: defaultUser,
		}
	}

	// Polling for instance status
	instanceIds := make([]*string, count)
	for i, host := range instances {
		instanceIds[i] = aws.String(host.HostId)
	}
	for {
		time.Sleep(time.Second * 2)
		instances, err := DescribeCloudInstances(region, instanceIds)
		if err != nil {
			return nil, err
		}

		hasReady := true
		for _, reservation := range instances.Reservations {
			for _, instance := range reservation.Instances {
				if instance.State.Name != string(running) {
					hasReady = false
				}
			}
		}
		if hasReady {
			break
		}
	}

	// instance is running but os is not ready - may cause ssh timeout
	time.Sleep(time.Second * 30)
	return instances, nil
}
