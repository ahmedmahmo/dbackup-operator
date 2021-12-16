/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"strconv"
	"time"

	utils "github.com/ahmedmahmo/discovery-operator/runner/aws/utils"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var (
	// AWS Variables
	AWS_S3_REGION         = utils.GetEnvVariable("AWS_S3_REGION", "eu-central-1")
	AWS_S3_BUCKET 	      = utils.GetEnvVariable("AWS_S3_BUCKET", "kubebucketforbackup")
	AWS_ACCESS_KEY_ID     = utils.GetEnvVariable("AWS_ACCESS_KEY_ID", "")
	AWS_SECRET_ACCESS_KEY = utils.GetEnvVariable("AWS_SECRET_ACCESS_KEY", "")

	// Postgres variables
	POSTGRES_HOST 		  = utils.GetEnvVariable("POSTGRES_HOST", "postgres.postgres.svc.cluster.local")
	POSTGRES_PORT		  = utils.GetEnvVariable("POSTGRES_PORT", "5432")
	POSTGRES_DATABASE     = utils.GetEnvVariable("POSTGRES_DATABASE", "dvdrental")
	POSTGRES_USERNAME     = utils.GetEnvVariable("POSTGRES_USERNAME", "postgres")
	POSTGRES_PASSWORD     = utils.GetEnvVariable("POSTGRES_PASSWORD", "1234")			
)

const (
	command = "pg_dump"
)

func main()  {
	fmt.Println("Runner is up...")
	fmt.Printf("Starting dump from %s\n", POSTGRES_HOST)

	f := strings.Join([]string{
		POSTGRES_DATABASE,
		"-",
		strconv.FormatInt(
		time.Now().Unix(), 10),
		".sql",
	},"")

	arguments := []string{}
	arguments = append(arguments, "--no-owner")
	arguments = append(arguments, "--verbose")
	arguments = append(arguments, strings.Join([]string{
		"--dbname=postgresql://",
		POSTGRES_USERNAME,
		":",
		POSTGRES_PASSWORD,
		"@",
		POSTGRES_HOST,
		":",
		POSTGRES_PORT,
		"/",
		POSTGRES_DATABASE,
	}, ""))

	arguments = append(arguments, "-f")
	arguments = append(arguments, f)


	cmd := exec.Command(command, arguments...)
	err := cmd.Run()
	
	if err != nil {
		fmt.Println(err)
		panic(err)
	}else {
		fmt.Printf("Dumped successfully to %s", f)
	}

	
	fmt.Println("Starting uploading dump ")
	s3Configration := &aws.Config{
		Region: aws.String(AWS_S3_REGION),
		Credentials: credentials.NewStaticCredentials(
			AWS_ACCESS_KEY_ID,
			AWS_SECRET_ACCESS_KEY,
			"",
		),
	}

	s3Session := session.New(s3Configration)
	
	uploadManger := s3manager.NewUploader(s3Session)
	
	opened, err := os.Open(f)
	if err != nil {
		panic(err)
	}

	result, err := uploadManger.Upload(&s3manager.UploadInput{
		Bucket: aws.String(AWS_S3_BUCKET),
		Key:    aws.String(f),
		Body:   opened,
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("file uploaded to, %s\n", result.Location)
}