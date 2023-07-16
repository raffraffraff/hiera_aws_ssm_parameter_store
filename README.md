## AWS SSM Parameter Store parameter lookup

This function allows you to look up single values stored as parameters in AWS SSM Parameter Store. It can optionally specify a KMS key alias for decrypting secrets.

## Installation
Build the plugin from the root directory of this module:
```
go build -o aws_ssm_parameter_store
```
Then make the plugin available to Hiera. See
[Extending Hiera](https://github.com/lyraproj/hiera#Extending-Hiera) for info on how to do that.

#### A Note about debugging
When debugging remotely from an IDE like JetBrains goland, use `-gcflags 'all=N -l'` to ensure that all symbols are present in the
final binary.
```
go build -o aws_ssm_parameter_store -gcflags 'all=-N -l'
```

## Examples
To add the Parameter Store to Hiera's lookup hierarchy, update `hiera.yaml`:

```
---
version: 5

defaults:
  plugindir: ./plugins
  datadir: ./hiera
  data_hash: yaml_data

hierarchy:
  - name: secrets
    lookup_key: ssm_parameter_store
    options:
	aws_profile_name: %{aws_account}.admin
	region: %{region}

  - name: common
    path: common.yaml
```
