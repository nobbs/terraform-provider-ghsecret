# GitHub Secrets Provider

This provider allows you to encrypt plain text to be stored as an pre-encrypted
GitHub secret via the [github](https://registry.terraform.io/providers/hashicorp/github) provider.
This way, no plain text secrets need to be stored in the Terraform state.

The data is encrypted using the [libsodium sealed box encryption scheme](https://libsodium.gitbook.io/doc/public-key_cryptography/sealed_boxes)
and the [public key of the GitHub repository](https://docs.github.com/en/rest/actions/secrets?apiVersion=2022-11-28#get-a-repository-public-key)
where the secret will be stored. The encrypted data is then base64 encoded and returned as a
string ready to be used in the GitHub provider.

The encryption is done using the [anonymous sealed box encryption](https://pkg.go.dev/golang.org/x/crypto/nacl/box#SealAnonymous)
provided by the Go implementation of libsodium. Part of this encryption scheme is the generation
of a ephemeral key pair that is used to encrypt the data with the public key of the GitHub
repository. The ephemeral public key is included in the encrypted data and is used by the
GitHub repository to decrypt the data. This way, the data can only be decrypted by the GitHub
repository that has the corresponding private key.

**Caution:** the ephemeral key pair requires a random number generator for secure key generation.
This provider uses a random number generator seeded with data derived from the hash of the cleartext
data to ensure that the ephemeral key pair is deterministic for a given cleartext. This is done,
since Terraform expects constant results for a given input and the encryption otherwise would
result in different outputs for the same input on each run. This is most likely not to be
considered secure in a **strict cryptographic sense**, but still an improvement over ending up with
plain text secrets in the Terraform state.

## Requirements

As provider functions are a fairly new feature in Terraform, you will need to be using Terraform v1.8 or later.

## Usage

This provider contains only one provider function:

- `encrypt` - Encrypts a plain text with the public key of a GitHub repository to be stored as a GitHub secret.

To make use of the provider, you will need to add the provider to your Terraform configuration:

```hcl
terraform {
  required_providers {
    ghsecret = {
      source = "nobbs/ghsecret"
      version = "~> 0.1.0"
    }
  }
}

# The provider has no configuration options
provider "ghsecret" {}
```

## Examples

This example can also be found in the [examples](./examples) directory.

Once the provider is configured, you can use the provider functions in your Terraform configuration, for example like this:

```hcl
data "github_actions_public_key" "repo" {
  repository = "terraform-provider-ghsecret"
}

resource "github_actions_secret" "this" {
  repository  = "terraform-provider-ghsecret"
  secret_name = "EXAMPLE_SECRET"
  encrypted_value = provider::ghsecret::encrypt(
    "Hello, World!",
    data.github_actions_public_key.repo.key
  )
}

output "raw" {
  value = nonsensitive(github_actions_secret.this.encrypted_value)
}

# data "github_actions_public_key" "repo" {
#     id         = (sensitive value)
#     key        = (sensitive value)
#     key_id     = (sensitive value)
#     repository = "terraform-provider-ghsecret"
# }
#
# resource "github_actions_secret" "this" {
#     created_at      = "2024-12-04 22:27:42 +0000 UTC"
#     encrypted_value = (sensitive value)
#     id              = "terraform-provider-ghsecret:EXAMPLE_SECRET"
#     plaintext_value = (sensitive value)
#     repository      = "terraform-provider-ghsecret"
#     secret_name     = "EXAMPLE_SECRET"
#     updated_at      = "2024-12-04 22:27:42 +0000 UTC"
# }
#
# Outputs:
#
# raw = "23k/JkPJ..."
```
