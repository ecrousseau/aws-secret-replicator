package main

import (
    "context"
    "errors"
    "os"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    corev1 "k8s.io/api/core/v1"
    k8serrors "k8s.io/apimachinery/pkg/api/errors"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"
    "encoding/json"
)

type Config struct {
    Secrets []Secret `json:"secrets"`
}

type Secret struct {
    ARN string `json:"arn"`
    Name string `json:"name"`
    Type corev1.SecretType `json:"type,omitempty"`
}

var (
    clientset *kubernetes.Clientset
    namespace string
    configmapName string
)

func InitialiseKubernetesClient() (error) {
    cfg, err := rest.InClusterConfig()
    if err != nil {
        return err
    }
    clientset, err = kubernetes.NewForConfig(cfg)
    if err != nil {
        return err
    }
    namespace = os.Getenv("POD_NAMESPACE")
    if namespace == "" {
        return errors.New("Required environment variable POD_NAMESPACE is not set")
    }
    configmapName = os.Getenv("CONFIGMAP_NAME")
    if configmapName == "" {
        return errors.New("Required environment variable CONFIGMAP_NAME is not set")
    }
    return nil
}

func GetConfig(config *Config) (error) {
    configmap, err := clientset.CoreV1().ConfigMaps(namespace).Get(context.TODO(), configmapName, metav1.GetOptions{})
    if err != nil {
        return err
    }
    // configmap should contain a single key "config", containing a JSON data structure
    err = json.Unmarshal([]byte(configmap.Data["config"]), config)
    if err != nil {
        return err
    }
    return nil
}

func CreateOrUpdateSecret(secret Secret, secretValue string) (error) {
    // unmarshal JSON string
    var secretData map[string]string
    err := json.Unmarshal([]byte(secretValue), &secretData)
    if err != nil {
        return err
    }
    // assemble secret data structure
    k8sSecret := corev1.Secret{
        ObjectMeta: metav1.ObjectMeta{
            Name: secret.Name,
            Namespace: namespace,
        },
        StringData: secretData,
        Type: secret.Type,
    }
    // see if the secret exists already
    _, err = clientset.CoreV1().Secrets(namespace).Get(context.TODO(), secret.Name, metav1.GetOptions{})
    if err != nil {
        if k8serrors.IsNotFound(err) {
            // if it does not exist, create it
            _, err = clientset.CoreV1().Secrets(namespace).Create(context.TODO(), &k8sSecret, metav1.CreateOptions{})
            if err != nil {
                return err
            }
        } else {
            return err
        }
    } else {
        // if it does exist, update it
        _, err = clientset.CoreV1().Secrets(namespace).Update(context.TODO(), &k8sSecret, metav1.UpdateOptions{})
        if err != nil {
            return err
        }
    }
    return nil
}
