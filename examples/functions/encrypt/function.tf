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
