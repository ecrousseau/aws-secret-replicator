# AWS Secret Replicator

[![Build Status](https://github.com/ecrousseau/aws-secret-replicator/workflows/master/badge.svg)](https://github.com/ecrousseau/aws-secret-replicator/actions)

_aws-secret-replicator_ allows your containerized applications to consume secrets from AWS Secrets Manager by copying their value to a Kubernetes secret. The solution consists of a  daemon that runs in your cluster and regularly replicates secret values, as configured by a configmap.

## Prerequsites 
- An AWS account
- IRSA configured on your Kubernetes cluster

## Installation

Grab the [latest release](https://github.com/ecrousseau/aws-secret-replicator/releases), create a values file to set your IAM role and list of secrets, and use [Helm](https://helm.sh/) to install it.

```yaml
configMap:
  data:
    config: |
      {
        "secrets": [
          {
            "arn": "arn:aws:secretsmanager:us-west-2:123456789012:secret:example-cert",
            "name": "example-cert",
            "type": "kubernetes.io/tls"
          }
        ]
      }

serviceAccount:
  annotations:
    eks.amazonaws.com/role-arn: arn:aws:iam::123456789012:role/example-role
```

```bash
$ kubectl create ns example-namespace
$ helm install --namespace example-namespace --values values.yaml aws-secret-replicator https://github.com/ecrousseau/aws-secret-replicator/releases/download/v1.0/aws-secret-replicator-v1.0.tgz
NAME: aws-secret-replicator
LAST DEPLOYED: Fri Feb  5 01:09:20 2021
NAMESPACE: example-namespace
STATUS: deployed
REVISION: 1
TEST SUITE: None
```

Alternatively, create your own manifests using the [helm chart](https://github.com/ecrousseau/aws-secret-replicator/tree/master/charts/aws-secret-replicator) as a guide. The container image is published in GHCR [here](https://github.com/ecrousseau?tab=packages&q=aws-secret-replicator).

## Usage

The daemon is designed to run inside a Kubernetes namespace alongside your existing application(s). For example, you might use it inside your `istio-system` namespace to create TLS secrets to use with your ingress gateways.

You will need to configure the daemon using a configmap. Here is the example default configuration from the Helm chart values.yaml. 

```yaml
configMap:
  create: true
  name: aws-secret-replicator-config
  data:
    config: |
      {
        "secrets": [
          {
            "arn": "arn:aws:secretsmanager:us-west-2:123456789012:secret:example-cert",
            "name": "example-cert",
            "type": "kubernetes.io/tls"
          }
        ]
      }
```

You should override this with your configuration. Note that the config itself must be valid JSON, and should have a single key "secrets", with the value being a list of objects. Each object should have keys "arn" and "name", and optionally "type". 

The value for "arn" must be the ARN of a secret that the daemon can retrieve - make sure your IAM role and/or resource policies are set up appropriately. The value for "name" must be a valid Kubernetes secret name. The secrets will be created in the namespace the daemon is running in. If "type" is supplied it must be a valid [Kubernetes secret type](https://kubernetes.io/docs/concepts/configuration/secret/#secret-types).

Each AWS secret must contain a map of key/value pairs encoded in JSON. These will become the keys and values in the resulting Kubernetes secret. If you need to store multi-line strings (for example, TLS certificates) then make sure you escape any newlines properly, as per the following example.

```json
{
    "tls.crt": "-----BEGIN CERTIFICATE-----\nMIIE9jCCAt6gAwIBAgIUODlRZ+IoXf+DDH7XwUtUWZiO4fYwDQYJKoZIhvcNAQEN\nBQAwEzERMA8GA1UEAxMIc25ha2VvaWwwHhcNMjAwNzIyMDk1NTAwWhcNMjUwNzIx\nMDk1NTAwWjATMREwDwYDVQQDEwhzbmFrZW9pbDCCAiIwDQYJKoZIhvcNAQEBBQAD\nggIPADCCAgoCggIBAN2zvT4CCJfvyvh7XW0tJ4lkoHpngOt6ss73RgTQG8oGFb+a\n6LLl264so0reBTapFvwQT808mGs5mgPYsB+cG8nW/nv1hoa7YgPBW7ICqQyrPU7v\nM6/qUWiLQWGudgVqiYiskXK5uQK5s2M6OOXgxQUij7wyTLvMOuFgDNJgJxU4KwY0\nlnubjdPeFumYd4UZ5HFrW4UegNvHuLh7Ep/HYaEE5jB8AKzDuJ4/imAFuzl5rnvU\nmizQmFh9JKnMOj1QrB4PsoidfqCx450Bq5VUSsOQP1t4PTiNBc5YP2xR1oBVkpo3\npbyWROPC/tbuvE4FIsm6kOqDBR5z011onbRW9VU6rXp9p0L+/L5sg3rEPUXwueCv\nqlnGfo8lnxiotQPGx4NRnLlHhP6Uu3dX1Y21+ZnY67Cdx9JcjMvfyPENLQE6t/2h\ngAASXL4NCOIl0cAww6l2vG+rhCiy0z6EZQ8yjxAv4SLEy69TFe8/kT/x2qwhpbr+\nNPps1j0mB/x0riJG6lIA/UzP59bzUZ8mHa5nFL2E8BSYMssYh1MG6FqhZQozFpxo\nAk9OzhTcoDG2ksGF4zSo89dblwQWN48+BPPNiHcb6+H5qvkgEoHnmNqzVt+yNXan\nsKnKwJGjoFKQxezhD7iGyx5oqXTztx4Q+Ue3sJabddvFxszrKMgNN39CLGtpAgMB\nAAGjQjBAMA4GA1UdDwEB/wQEAwIBBjAPBgNVHRMBAf8EBTADAQH/MB0GA1UdDgQW\nBBTmYQ/XUP8hMp8C31tERALUmaTz0jANBgkqhkiG9w0BAQ0FAAOCAgEAZAMEmd6l\n1NMEhk6On78H7ec5wde4pVgkHTdJOJWflV3qtC9CvtUSsWoxGtHfycf/COevmCEr\nTUAIEYtUQP6B16zuRTDRFstQQcTQGIKipUa3Eb+rn9b+z3XyGY32uIMNrzlZK5gR\n59UXAs/tEvHxVKtrufxB0onMs3q7sgyYTlyjJXFqvy7VTLKNtEi2drDpsBvUeCA7\nQ20HBwmHMnn2UAj7ZfrrHIwmGHsyonwAMwKZ9nCOoAXpJEomJojDuh87VehHSwOE\n72f7fMB9/9JLsvyiwcHDohPNe+HUnpkmtCNipyFRNwR67GjI3TphWCghubrfPRA5\n+EXL4uVUQ/ZXH+xTW0lfN/RhUZ27LEZi7PTEU72BRv3SQ0J/BDkYNcJTchUYz2Mk\n4umTPy1SP7+s7m/hUcHXcX3ko/2DImvtoyota1g7fUJXAlggemZ6Yh8xYk9meH3K\n05bjDXY3wAqCMknFIYrfxBQZnLh4MHoNActPwJhsIXTcjrBPU0h7T0so0KNuIDxX\nM0F0M+H/ha8wa1Ve+AH4d0dp6P5nVbmdqpSYCuEcwKzjVBCBZWCZ1YUY0L4x7mSw\nbzGkwIl/RdvrAxw/DwMYZ+eTP/ceL+z6bwE+xk7CE2jXCnLz+6IFHsVkcHYTN1kZ\nG5cIOTI3oUxCu0czPKfV5apwhoSX0woSvys=\n-----END CERTIFICATE-----\n",
    "tls.key": "-----BEGIN RSA PRIVATE KEY-----\n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\n-----END RSA PRIVATE KEY-----\n"
}
```
