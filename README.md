## AWS SSM Parameter Store parameter plugin for Hiera 5 (go)

This function allows you to look up single values stored as parameters in AWS SSM Parameter Store. The intended use-case is to store secrets with KMS encryption in Parameter Store.

## Installation
Build the plugin from the root directory of this module:
```
go build -o aws_ssm_parameter_store
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
      aws_profile_name: %{aws_account}.AdministratorAccess
      aws_region: %{region}
```

NOTE: Since I already pass `region` and `aws_account` values to the Hiera provider, I'm using them to configure the 'aws_ssm_parameter' plugin, so it always performs lookups in the correct AWS account and region.

## Test lookup
To test the plugin, write a test parameter to AWS SSM Parameter Store using the AWS cli:

```
[user@box ~]$ aws ssm put-parameter \
  --region eu-west-1 \
  --profile dev.AdministratorAccess \
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
