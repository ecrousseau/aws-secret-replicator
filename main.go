package main

import (
    "time"
    "k8s.io/klog/v2"
)

var (
    cfg Config
)

func main() {
    klog.Info("Starting up")
    klog.Info("Initialising kubernetes client")
    err := InitialiseKubernetesClient()
    if err != nil {
        klog.Fatal(err)
    }
    OnTick()
    ticker := time.NewTicker(30 * time.Second)
    for _ = range ticker.C {
		OnTick()
    }
}

func OnTick() {
    klog.Info("Retrieving current config from configmap")
    err := GetConfig(&cfg)
    if err != nil {
        klog.Fatal(err)
    }
    klog.Info("Refreshing secrets")
    for _, secret := range cfg.Secrets {
        value, err := GetSecretValue(secret.ARN)
        if err != nil {
            klog.Errorf("Error while retrieving AWS secret %v: %v", secret.ARN, err)
            continue
        }
        klog.Infof("Retrieved value for AWS secret %v", secret.ARN)
        klog.Infof("DEBUG value: %#v", value)
        err = CreateOrUpdateSecret(secret, value)
        if err != nil {
            klog.Errorf("Error while creating kubernetes secret %v: %v", secret.Name, err)
            continue
        }
        klog.Infof("Created or updated kubernetes secret %v", secret.Name)
    }
    klog.Info("Done refreshing secrets")
}
