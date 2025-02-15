package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

func AWSNewSession(region, profile string) *session.Session {
	newsess, err := session.NewSessionWithOptions(session.Options{
		Config:            aws.Config{Region: aws.String(region)},
		Profile:           profile,
		SharedConfigState: session.SharedConfigEnable,
	})
	session := session.Must(newsess, err)
	stssvc := sts.New(session)
	_, err = stssvc.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		panic(fmt.Errorf("AWSCheckAuth(sess, awsProfileName) failed: %v", err))
	}
	return session
}
