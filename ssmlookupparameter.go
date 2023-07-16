package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/lyraproj/dgo/dgo"
	"github.com/lyraproj/hierasdk/hiera"
	"github.com/lyraproj/hierasdk/plugin"
	"github.com/lyraproj/hierasdk/register"
)

func main() {
	register.LookupKey(`aws_ssm_parameter_store`, AWSSSMParameterStoreLookupKey)
	plugin.ServeAndExit()
}

// AWSSSMParameterStoreLookupKey looks up a single value from AWS SSM Parameter Store
func AWSSSMParameterStoreLookupKey(hc hiera.ProviderContext, key string) dgo.Value {
	if key == `lookup_options` {
		return nil
	}

	parameterName, ok := hc.StringOption(`parameter_name`)
	if !ok {
		panic(fmt.Errorf(`missing required provider option 'parameter_name'`))
	}
	awsProfileName, _ := hc.StringOption(`aws_profile_name`)
	if !ok {
		panic(fmt.Errorf(`missing required provider option 'aws_profile_name'`))
	}
	awsRegionName, _ := hc.StringOption(`aws_region`)
	if !ok {
		panic(fmt.Errorf(`missing required provider option 'aws_region'`))
	}

	// Create a new AWS session with the specified profile
	sess, err := session.NewSessionWithOptions(session.Options{
		Profile: awsProfileName,
		Config: aws.Config{
			Region: aws.String(awsRegionName),
		},
        })

	ssmSvc := ssm.New(sess)

	res, err := ssmSvc.GetParameter(&ssm.GetParameterInput{
		Name:           aws.String(parameterName),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return nil
	}
	decryptedValue := aws.StringValue(res.Parameter.Value)

	return hc.ToData(decryptedValue)
}

