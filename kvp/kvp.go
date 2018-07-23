package kvp

import (
	"encoding/json"

	"os"
	"strings"

	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type envvars map[string]string

var cfg aws.Config

func EnvPairs(secretsNames []string) ([]string, error) {
	initialVars := envVarsFromOS(os.Environ())

	awsVars, err := envVarsFromAWS(secretsNames)
	if err != nil {
		return nil, err
	}

	for k, v := range awsVars {
		initialVars[k] = v
	}

	var out []string
	for k, v := range initialVars {
		out = append(out, fmt.Sprintf("%s=%v", k, v))
	}

	return out, nil
}

func envVarsFromOS(pairs []string) envvars {

	e := make(map[string]string)

	for _, v := range pairs {
		kvpair := strings.Split(v, "=")
		if len(kvpair) > 1 {
			e[kvpair[0]] = strings.Join(kvpair[1:], "")
		}
	}

	return envvars(e)
}

func envVarsFromAWS(secretsNames []string) (envvars, error) {

	var err error
	cfg, err = external.LoadDefaultAWSConfig()
	if err != nil {
		return nil, err
	}

	if region := os.Getenv("SECRETLY_REGION"); region != ""{
		cfg.Region = region
	}

	smSvc := secretsmanager.New(cfg)

	outData := make(map[string]string)

	for _, secretName := range secretsNames {

		req := smSvc.GetSecretValueRequest(&secretsmanager.GetSecretValueInput{
			SecretId: aws.String(secretName),
		})

		resp, err := req.Send()
		if err != nil {
			return nil, err
		}

		var respData map[string]string

		err = json.Unmarshal([]byte(*resp.SecretString), &respData)
		if err != nil {
			return nil, err
		}

		for k, v := range respData {
			outData[k] = v
		}
	}

	return envvars(outData), nil
}
