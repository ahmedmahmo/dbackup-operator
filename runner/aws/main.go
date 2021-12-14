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
	"strconv"

	utils "github.com/ahmedmahmo/discovery-operator/runner/aws/utils"
	pg_commands "github.com/habx/pg-commands"

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
	POSTGRES_HOST 		  = utils.GetEnvVariable("POSTGRES_HOST", "localhost")
	POSTGRES_PORT		  = utils.GetEnvVariable("POSTGRES_PORT", "5432")
	POSTGRES_DATABASE     = utils.GetEnvVariable("POSTGRES_DATABASE", "dvdrental")
	POSTGRES_USERNAME     = utils.GetEnvVariable("POSTGRES_USERNAME", "postgres")
	POSTGRES_PASSWORD     = utils.GetEnvVariable("POSTGRES_PASSWORD", "1234")			
)

func main()  {
	fmt.Println("Runner is up...")
	fmt.Printf("Starting dump from %s\n", POSTGRES_HOST)

	port, err := strconv.Atoi(POSTGRES_PORT)
    if err != nil {
		panic(err)
	}

	dump := pg_commands.NewDump(
		&pg_commands.Postgres{
		Host    : POSTGRES_HOST,
		Port    : port,
		DB      : POSTGRES_DATABASE,
		Username: POSTGRES_USERNAME,
		Password: POSTGRES_PASSWORD,
		})


	exec := dump.Exec(pg_commands.ExecOptions{StreamPrint: false})
	if exec.Error != nil {
		fmt.Println(exec.Error.Err)
    	fmt.Println(exec.FullCommand)
	} else {
		fmt.Println("Dump success")
		fmt.Printf("File %s:", exec.File)
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
	
	f, err := os.Open(exec.File)
	if err != nil {
		panic(err)
	}

	result, err := uploadManger.Upload(&s3manager.UploadInput{
		Bucket: aws.String(AWS_S3_BUCKET),
		Key:    aws.String(exec.File),
		Body:   f,
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("file uploaded to, %s\n", result.Location)
}