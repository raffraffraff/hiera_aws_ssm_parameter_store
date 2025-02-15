package main

import (
	"fmt"
	"strings"

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
	Name           string
	WithDecryption bool
	ssmsvc         *SSM
}

func (s *SSM) Param(name string, decryption bool) *Param {
	return &Param{
		Name:           name,
		WithDecryption: decryption,
		ssmsvc:         s,
	}
}

func (p *Param) GetValue() (string, error) {
	ssmsvc := p.ssmsvc.client
	parameter, err := ssmsvc.GetParameter(&ssm.GetParameterInput{
		Name:           &p.Name,
		WithDecryption: &p.WithDecryption,
	})
	if err != nil {
		return "", err
	}
	value := *parameter.Parameter.Value
	return value, nil
}

func AWSSSMSession(hc hiera.ProviderContext) *SSM {

	awsProfileName, ok := hc.StringOption(`aws_profile`)
	if !ok && awsProfileName != "" {
		panic(fmt.Errorf(`Missing hiera plugin option option 'aws_profile'`))
	}

	awsRegionName, ok := hc.StringOption(`aws_region`)
	if !ok && awsRegionName != "" {
		panic(fmt.Errorf(`Missing hiera plugin option 'aws_region'`))
	}
	sess := AWSNewSession(awsRegionName, awsProfileName)
	ssmsvc := &SSM{ssm.New(sess)}
	return ssmsvc

}

func AWSGetParameter(hc hiera.ProviderContext, key string) dgo.Value {

	/*	This function is registered for lookup, and then the hiera plugin serves content
		from it. During Hiera lookups, this function can get called many times, depending
		on your hierarchies, interpolation, data etc. If you log this function you will
		likely see the same key getting looked up many times. To prevent multiple parameter
		lookups for each key, it implements a simple cache.
	*/

	if key == "lookup_options" {
		return nil
	}

	allowedPrefix, ok := hc.StringOption(`allowed_prefix`)
	if !ok {
		allowedPrefix = "/"
	}

	result, hit := cache.Get(key)
	if !hit {

		if !strings.HasPrefix(key, allowedPrefix) {
			return nil
		} else {
			ssmsvc := AWSSSMSession(hc)
			result, _ = ssmsvc.Param(key, true).GetValue()
		}
		cache.Put(key, result)
	}
	if result == "" {
		return nil
	}
	return hc.ToData(result)

}

var ssmsvc *SSM
var allowedPrefix string

func main() {
	register.LookupKey(`aws_ssm_parameter`, AWSGetParameter)
	plugin.ServeAndExit()
}
