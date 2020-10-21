---
title: Teleport API Reference
description: The detailed guide to Teleport API
---

# Teleport API Reference

Teleport is currently working on documenting our API.

!!! warning

        We are currently working on this project. If you have an API suggestion, [please complete our survey](https://docs.google.com/forms/d/1HPQu5Asg3lR0cu5crnLDhlvovGpFVIIbDMRvqclPhQg/edit).

## Authentication
In order to interact with the API, you will need to provision appropriate
TLS certificates. In order to provision certificates, you will need to create a
user with appropriate permissions. You should only give the api user permissions for what it actually needs. 

To quickly get started with the api, you can use this api-admin user, but in real usage make sure to have stringent permissions in place.

```yaml
# Copy and Paste the below on the Teleport Auth server.
$ cat > api-admin.yaml <<EOF
{!examples/go-client/api-admin.yaml!}
EOF
```

Create the role, user, and TLS certificates. This should result in three PEM encoded files being generated: `api-admin.crt`, `api-admin.key`, and `api-admin.cas` in the `/certs` directory (certificate, private key, and CA certs respectively).

```bash
$ tctl create -f api-admin.yaml
$ mkdir -p certs
$ tctl auth sign --format=tls --user=api-admin --out=certs/api-admin 
```

!!! Note
    By default, `tctl auth sign` produces certificates with a relatively short lifetime.
    For production deployments, the `--ttl` flag can be used to ensure a more practical
    certificate lifetime.

# gRPC APIs

In most cases, you can interact with teleport using our cli tools, [tsh](cli-docs/#tsh) and [tctl](cli-docs/#tctl). However, there are some scenarios where you may need to interact with teleport programmatically. For this purpose, you can use the same gRPC auth API that `tctl` and `tsh` use directly.

## GoLang Examples
Below are some code examples that can be used with Teleport to perform a few key tasks.

Before you begin: 

- Install [Golang](https://golang.org/doc/install) 1.13+ and Setup Go Dev Environment
- Have access to a Teleport Auth server, follow our [quickstart](quickstart) if needed.

### Setup Go Client Authentication 

Follow the authentication section above to create a role, user, and TLS certificates. Put the generated PEM files in a folder called `/certs` if they aren't already, and provide this folder in the same directory as the following go file.

```go
{!examples/go-client/client.go!}
```

Create this go file to create a `client` for your auth server, which can be used for the following example. Just create a main.go file and use the client from `connectClient` to experiment. Your folder structure should look like this:

```
.
|-- api-admin.yaml
|-- certs
|   +-- api-admin.cas
|   +-- api-admin.crt
|   +-- api-admin.key
+-- client.go
```

### Roles

[Roles](enterprise/ssh-rbac/#roles) are used to define what resources and actions a user is allowed/denied access to. It's best to make your roles as stringent with permissions as possible.

#### Create a new role

```go
import github.com/gravitational/teleport/lib/services

// create a new auditor role which has very limited permissions
role, err := services.NewRole("auditor", services.RoleSpecV3{
  Options: services.RoleOptions{
    MaxSessionTTL: services.Duration(time.Hour),
  },
  Allow: services.RoleConditions{
    Logins: []string{"auditor"},
    Rules: []services.Rule{
      // - resources: ['session']
      //   verbs: ['list', 'read']
      services.NewRule(services.KindSession, services.RO()),
    },
  },
  Deny: services.RoleConditions{
    // node_labels: '*': '*'
    NodeLabels: services.Labels{"*": []string{"*"}},
  },
})
if err != nil {
  return err
}

err := client.UpsertRole(ctx, role)
```

#### Retrieve roles

```go
roles, err := client.GetRoles()
```

#### Update role

```go
role, err := client.GetRole("auditor")
if err != nil {
  return err
}
  
// update the auditor role's read rule to not provide access to secrets
role.SetRules(services.Allow, []services.Rule{
	services.NewRule(services.KindSession, services.ReadNoSecrets()),
})
err = client.UpsertRole(ctx, role)
```

#### Delete role

```go
err := client.DeleteRole(ctx, "auditor")
```

### Tokens

[Tokens](admin-guide/#adding-nodes-to-the-cluster) are used to add nodes to a cluster.

#### Create a new token

```go
tokenString, err := client.GenerateToken(ctx, auth.GenerateTokenRequest{
  // You can provide 'Token' for a non-random tokenString
  Roles: teleport.Roles{teleport.RoleProxy},
  TTL:   time.Hour,
})
```

#### Retrieve tokens

```go
tokens, err := client.GetTokens()
```

#### Update token

```go
provisionToken, err := client.GetToken(tokenString)
if err != nil {
  return err
}

// updates the token to be a proxy token
provisionToken.SetRoles(teleport.Roles{teleport.RoleProxy})
err = client.UpsertToken(provisionToken)
```

#### Delete token

```go
err := client.DeleteToken(tokenString)
```

// TODO: Decide how to section this part off

#### Create a remote cluster with Labels

New remote clusters added with a token will inherit any labels the token has.

```go
tokenString, err := client.GenerateToken(ctx, auth.GenerateTokenRequest{
  // You can provide 'Token' for a static token name
  Roles: teleport.Roles{teleport.RoleTrustedCluster},
  TTL:   time.Hour,
  Labels: map[string]string{
    "env": "staging",
  },
})
```

#### Update remote cluster's labels

```go
rc, err := services.NewRemoteCluster("remote")
if err != nil {
  return fmt.Errorf("Failed to make new remote cluster: %v", err)
}

md := rc.GetMetadata()
md.Labels = map[string]string{"env": "prod"}
rc.SetMetadata(md)
```
