package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
        "github.com/aws/aws-sdk-go/service/sts"
	"github.com/lyraproj/dgo/dgo"
	"github.com/lyraproj/hierasdk/hiera"
	"github.com/lyraproj/hierasdk/plugin"
	"github.com/lyraproj/hierasdk/register"
)

type SSM struct {
  client ssmiface.SSMAPI
}

type Param struct {
  Name string
  WithDecryption bool
  ssmsvc *SSM
}

func (s *SSM) Param(name string, decryption bool) *Param {
  return &Param{
    Name: name,
    WithDecryption: decryption,
    ssmsvc: s,
  }
}

func (p *Param) GetValue() (string, error){
  ssmsvc := p.ssmsvc.client
  parameter, err := ssmsvc.GetParameter(&ssm.GetParameterInput{
    Name:           &p.Name,
    WithDecryption: &p.WithDecryption,
  })
  if err != nil {
    return "" , err
  }
  value := *parameter.Parameter.Value
  return value, nil
}

func checkAWSAuth(sess *session.Session, profile string) error {
	svc := sts.New(sess)
	_, err := svc.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		fmt.Println("AWS authentication failed:", err)
		fmt.Printf("You may need to run: aws sso login --profile %s\n", profile)

		// Automatically attempt to run "aws sso login"
		fmt.Println("Attempting to run 'aws sso login' now...")
		cmd := exec.Command("aws", "sso", "login", "--profile", profile)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("failed to authenticate via AWS SSO: %w", err)
		}

		// Retry authentication after login
		fmt.Println("Retrying AWS authentication...")
		_, err = svc.GetCallerIdentity(&sts.GetCallerIdentityInput{})
		if err != nil {
			return fmt.Errorf("authentication still failing after SSO login: %w", err)
		}
	}

	return nil
}

func AuthenticateAWSSession(region, profile string) (*session.Session, error) {
	sess, err := session.NewSessionWithOptions(session.Options{
		Config:  aws.Config{Region: aws.String(region)},
		Profile: profile,
	})

	if err != nil {
		panic(fmt.Errorf(`Failed to create AWS session: %v`, err))
	}

	if err := checkAWSAuth(sess, profile); err != nil {
		panic(fmt.Errorf("AWS authentication error: %v", err))
	}

	return sess, nil
}

func AWSGetParameter(hc hiera.ProviderContext, key string) dgo.Value {
	awsProfileName, ok := hc.StringOption(`aws_profile`)
	if !ok {
		panic(fmt.Errorf(`Missing hiera plugin option option 'aws_profile'`))
	}

	awsRegionName, ok := hc.StringOption(`aws_region`)
	if !ok {
		panic(fmt.Errorf(`Missing hiera plugin option 'aws_region'`))
	}

	sess, err := AuthenticateAWSSession(awsRegionName, awsProfileName)
	if err != nil {
		panic(fmt.Errorf("AWS authentication error: %v", err))
	}

	ssmsvc := ssm.New(sess)
	decrypt := true
	result, err := ssmsvc.GetParameter(&ssm.GetParameterInput{
		Name:           aws.String(key),
		WithDecryption: &decrypt,
	})
	if err != nil {
		panic(fmt.Errorf("Failed to retrieve parameter: %v", err))
	}

        value := *result.Parameter.Value
        if value == "" {
          return nil
        }
        return hc.ToData("test")
//	return hc.ToData(*result.Parameter.Value)
}

func main() {
	register.LookupKey(`aws_ssm_parameter`, AWSGetParameter)
	plugin.ServeAndExit()
}
