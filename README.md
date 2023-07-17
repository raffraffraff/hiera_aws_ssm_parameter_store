## AWS SSM Parameter Store parameter lookup

This function allows you to look up single values stored as parameters in AWS SSM Parameter Store. It can optionally specify a KMS key alias for decrypting secrets.

## Installation
Build the plugin from the root directory of this module:
```
go build -o aws_ssm_parameter_store
```
Then make the plugin available to Hiera, you need to configure a `plugindir` and copy the go bin to it. See [Extending Hiera](https://github.com/lyraproj/hiera#Extending-Hiera) for more information.

#### A Note about debugging
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

If you don't have the Hiera `lookup` tool installed, then install it:

```
[user@box ~]$ go install github.com/lyraproj/hiera/lookup@latest
```

Now you can perform a lookup for `/dev/db_password`: 

```
[user@box ~]$ lookup --config=hiera.yaml "/dev/db_password"
VerySecretPassword
```

## Performing parameter store lookups in your existing Hiera yaml
You can use Hiera's internal `lookup` function to insert the value of a parameter anywhere. Consider the following example YAML which is used to define a list of databases for a PostgreSQL terraform module:

```
postgres:
  databases:
    monolithdb:
      encoding: "UTF8"
      lc_collate: "en_US.UTF-8"
      lc_ctype: "en_US.UTF-8"
      owner: pgadmin
  roles:
    frontend:
      database: monolithdb
      password: "%{lookup('/database/monolithdb/frontend')}"
```

We can execute a lookup for all `postgres.roles` using the `lookup` tool:

```
[user@box ~]$ lookup --config=hiera.yaml postgres.roles
frontend:
  database: monolithdb
  password: VerySecretPassword
```

As long as the AWS SSM Parameter Store has a value of type SecureString in that path, Hiera can interpolate it.
