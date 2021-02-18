package main

import (
    "context"
    "errors"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/aws/arn"
    "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

func GetSecretValue(secretArn string) (string, error) {
    parsedArn, err := arn.Parse(secretArn)
    if err != nil {
        return "", err
    }
    ctx := context.TODO()
    cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(parsedArn.Region))
    if err != nil {
        return "", err
    }
    client := secretsmanager.NewFromConfig(cfg)
    input := &secretsmanager.GetSecretValueInput{SecretId: &secretArn}
    result, err := client.GetSecretValue(ctx, input)
    if err != nil {
        return "", err
    }
    if result.SecretString == nil {
        return "", errors.New("Secret must contain a string value")
    }
    return *result.SecretString, nil
}
