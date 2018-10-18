package main

//go:generate mockgen -package mocks -destination resource/mocks/autoscaling.go -source=vendor/github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface/interface.go
//go:generate mockgen -package mocks -destination resource/mocks/ec2.go -source=vendor/github.com/aws/aws-sdk-go/service/ec2/ec2iface/interface.go
//go:generate mockgen -package mocks -destination resource/mocks/sts.go -source=vendor/github.com/aws/aws-sdk-go/service/sts/stsiface/interface.go

import (
	"os"

	"github.com/cloudetc/awsweeper/command"
)

func main() {
	os.Exit(command.WrappedMain())
}
