package main

import (
  "fmt"

  "github.com/aws/aws-sdk-go/aws"
  "github.com/aws/aws-sdk-go/aws/session"
  "github.com/aws/aws-sdk-go/service/ssm"
  "github.com/aws/aws-sdk-go/service/ssm/ssmiface"
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

func AWSSSMParameterStoreLookupKey(hc hiera.ProviderContext, key string) dgo.Value {

  if key == `lookup_options` {
    return nil
  }

  awsProfileName, ok := hc.StringOption(`aws_profile_name`)
  if !ok {
    panic(fmt.Errorf(`missing required provider option 'aws_profile_name'`))
  }

  awsRegionName, ok := hc.StringOption(`aws_region`)
  if !ok {
    panic(fmt.Errorf(`missing required provider option 'aws_region'`))
  }

  newsess, err := session.NewSessionWithOptions(session.Options{
    Config:            aws.Config{
                         Region:  aws.String(awsRegionName),
                       },
    Profile:           awsProfileName,
    SharedConfigState: session.SharedConfigEnable,
   })

  sess := session.Must(newsess, err)
  ssmsvc := &SSM{ssm.New(sess)}
  result,err := ssmsvc.Param(key, true).GetValue()

  return hc.ToData(result)
}

func main() {
  register.LookupKey(`aws_ssm_parameter`, AWSSSMParameterStoreLookupKey)
  plugin.ServeAndExit()
}
