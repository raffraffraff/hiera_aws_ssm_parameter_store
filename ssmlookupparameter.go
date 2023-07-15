package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
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

	// Create a new AWS session with the specified profile
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Profile:           awsProfileName,
	}))

	// Create a new AWS SSM client
	ssmSvc := ssm.New(sess)

	// Call the AWS SSM API to get the parameter value
	resp, err := ssmSvc.GetParameterWithContext(context.Background(), &ssm.GetParameterInput{
		Name: &parameterName,
	})
	if err != nil {
		panic(err)
	}

	// Get the parameter value
	parameterValue := *resp.Parameter.Value

	// Check if KMS encryption was used
	if resp.Parameter.KeyId != nil {
		kmsAlias, ok := hc.StringOption(`kms_key_alias`)
		if !ok {
			panic(fmt.Errorf(`missing required provider option 'kms_key_alias' for KMS-encrypted parameter`))
		}

		// Create a new AWS KMS client
		kmsSvc := kms.New(sess)

		// Call the AWS KMS API to decrypt the parameter value
		decryptResp, err := kmsSvc.DecryptWithContext(context.Background(), &kms.DecryptInput{
			CiphertextBlob: parameterValue,
			EncryptionContext: map[string]*string{
				"PARAMETER_NAME": &parameterName,
			},
		})
		if err != nil {
			panic(err)
		}

		// Get the decrypted value
		parameterValue = string(decryptResp.Plaintext)
	}

	// Return the parameter value
	return hc.ToData(parameterValue)
}

