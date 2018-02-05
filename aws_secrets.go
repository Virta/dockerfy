package main

import (
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
)

//
// return a map of AWS secrets (from AWS System Manager Parameter Store)
//
// If 'describe parameters' succeeds, returns a map of ENVIRONMENT variables with secrets overwritten from ssm.
// Otherwise, returns a map of ENVIRONMENT variables only.
//

func getAWS_Secrets() map[string]string {

	sess := session.Must(session.NewSession())
	svc := ssm.New(sess)

	parameterNames,err := describeAWS_ParameterNames(svc)
        if err != nil {
          return GetEnvMap()
        }
        prefix := string_template_eval(awsSecretsPrefixFlag)
        filtered := filterNames(parameterNames, prefix)
        if len(filtered) == 0 {
          return GetEnvMap()
        }
   	secrets := fetchAWS_Secrets(svc,filtered)
	return asMap(filtered, secrets)
}

func asMap(names []string, values []string) map[string]string {
		secrets := GetEnvMap() 
	    prefix := string_template_eval(awsSecretsPrefixFlag)

		for i := 0; i < len(names); i++ {
			name := names[i]
	                name = strings.Replace(name, prefix, "", 1)
			secrets[name] = values[i]
		}
		return secrets
}

func fetchAWS_Secrets(svc *ssm.SSM, parameterNames []string) []string {
	size := len(parameterNames)
	values := make([]string, size)
	for i := 0; i < size; i++ {
		params := &ssm.GetParameterInput{
			Name:           aws.String(parameterNames[i]),
			WithDecryption: aws.Bool(true),
		}

		resp, err := svc.GetParameter(params)

		if err != nil {
			log.Fatalf("cannot fetch AWS System Manager Parameters %s", err.Error())
		}

		values[i] = *resp.Parameter.Value
	}

	return values
}


func filterNames(input []string, prefix string) []string {
	size := len(input)
        var output []string
        for i := 0; i < size; i++ {
                if strings.HasPrefix(input[i], prefix) {
                        output = append(output, input[i])
                }
        }
        return output
}

func describeAWS_ParameterNames(svc *ssm.SSM) ([]string,error) {
	criteria := &ssm.DescribeParametersInput{
		MaxResults: aws.Int64(45), // limited by API call GetParametersInput
	}
	resp, err := svc.DescribeParameters(criteria)
	if err != nil {
		log.Printf("cannot describe AWS Parameter Names %s", err.Error())
                return nil,err
	}

	size := len(resp.Parameters)
	names := make([]string, size)

	for i := 0; i < size; i++ {
		names[i] = *resp.Parameters[i].Name
	}

	return names,nil
}
