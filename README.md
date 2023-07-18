## AWS SSM Parameter Store parameter plugin for Hiera 5 (go)

This function allows you to look up single values stored as parameters in AWS SSM Parameter Store. The intended use-case is to store secrets with KMS encryption in Parameter Store.

## Installation
Build the plugin from the root directory of this module:
```
go build -o aws_ssm_parameter
```
To add theplugin to Hiera, you need to configure a `plugindir` in your `hiera.yaml` and copy the go bin to that directory. See [Extending Hiera](https://github.com/lyraproj/hiera#Extending-Hiera) for more information.

### A Note about debugging
When debugging remotely from an IDE like JetBrains goland, use `-gcflags 'all=N -l'` to ensure that all symbols are present in the
final binary.
```
go build -o aws_ssm_parameter -gcflags 'all=-N -l'
```

To install it, copy the binary to your Hiera plugins directory.

## Hiera config example
To add the Parameter Store to Hiera's lookup hierarchy, update `hiera.yaml`. The plugin currently requires two parameters: `aws_profile_name` and `aws_region`. This example `hiera.yaml` places the AWS Parameter Store lookup _last_, so Hiera won't query AWS SSM Parameter Store unless the other hierarchies fail to return data. (Of course, this depends on your hiera lookup configuration)

```
---
version: 5

defaults:
  plugindir: ./plugins
  datadir: ./hiera
  data_hash: yaml_data

hierarchy:
  - name: common
    path: common.yaml

  - name: region
    path: region/%{region}.yaml

  - name: aws_account
    path: region/%{aws_account}.yaml

  - name: secrets
    lookup_key: aws_ssm_parameter
    options:
      aws_profile_name: Management.ReadOnlyAccess
      aws_region: us-west-1
```

## Status: works, but not flexible
Even though I already pass my AWS `region` and `aws_account` to the Hiera, I can't use them in the `options:` section. Interpolation in those fields _fails_. I created an [issue](https://github.com/lyraproj/hiera/issues/96) on the hiera project, but I think that the project may be dead :(

That leaves two options:
1. Hard-code the lookup to use a specific AWS account and region
2. Remove the options and export `AWS_PROFILE` and `AWS_REGION` environment variables instead

#1 depends on your use case and interpretation of your company security policies
#2 requires external automation or manual steps, and may cause other unwanted effects but at least you're not hard-coding stuff into your hiera.yaml

## Test lookup
To test the plugin, you'll need to write a test parameter to AWS SSM Parameter Store using the AWS cli:

```
[user@box ~]$ aws ssm put-parameter \
  --region us-west-1 \
  --profile Management.ReadOnlyAccess \
  --type SecureString \
  --name "/dev/db_password" \
  --value "VerySecretPassword"
```

Execute a command-line Hiera lookup using the `lookup` tool (install it with: `go install github.com/lyraproj/hiera/lookup@latest`)

```
[user@box ~]$ lookup --config=hiera.yaml "/dev/db_password"
VerySecretPassword
```

## Parameter Store lookups anywhere in your existing yaml!
Using Hiera's internal lookup function, you can insert the value of a parameter _anywhere_. Consider the following example YAML which is used to define a list of databases and roles:

```
postgres:
  databases:
    dev:
      encoding: "UTF8"
      lc_collate: "en_US.UTF-8"
      lc_ctype: "en_US.UTF-8"
      owner: pgadmin
  roles:
    frontend:
      database: dev
      password: "%{lookup('/dev/db_password')}"
```

Given this YAML, we can execute a command-line lookup for `postgres.roles` like this:

```
[user@box ~]$ lookup --config=hiera.yaml postgres.roles
frontend:
  database: monolithdb
  password: VerySecretPassword
```
