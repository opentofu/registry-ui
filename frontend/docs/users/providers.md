# Using a provider

Providers give OpenTofu the ability to call various APIs. The OpenTofu registry currently holds over 4000 such providers created by the community. They provide integrations for a vast range of services, from cloud providers, such as [AWS](https://search.opentofu.org/provider/hashicorp/aws/latest), [Azure](https://search.opentofu.org/provider/hashicorp/azurerm/latest), [GCP](https://search.opentofu.org/provider/hashicorp/google/latest) and more, version control systems, such as [GitHub](https://search.opentofu.org/provider/integrations/github/latest), [GitLab](https://search.opentofu.org/provider/gitlabhq/gitlab/latest), to password manages like [1Password](https://search.opentofu.org/provider/1password/onepassword/latest). You can explore the providers available [using the OpenTofu Registry Search](https://search.opentofu.org/providers/) or learn more about how providers work in OpenTofu using the [OpenTofu documentation](https://opentofu.org/docs/language/providers/).

~> Providers are binary programs created by their authors that run on your machine. The OpenTofu Registry does not perform security scanning on providers, and they may contain malicious code. There is also no guarantee that the provider binary was compiled from the code in the linked GitHub repository. Only use providers from authors you trust.

## Integrating a provider into your project

Once you found a provider for your needs, you can add it to your OpenTofu project in the `terraform` block:

```hcl2
terraform {
  required_providers {
    PROVIDER_NAME_HERE = {
      source = "NAMESPACE_HERE/PROVIDER_NAME_HERE"
      version = "PROVIDER_VERSION_HERE"
    }
  }
}
```

You can then configure your provider as follows:

```hcl2
provider "PROVIDER_NAME_HERE" {
  option1 = "value1"
  option2 = "value2"
}
```

For example, you can configure and use the GitHub provider like this:

```hcl2
terraform {
  required_providers {
    integrations = {
      source = "integrations/github"
      version = "v6.2.3"
    }
  }
}

variable "github_token" {}

provider "github" {
  token = var.github_token
}

resource "github_repository" "myrepo" {
  name        = "myrepo"
}
```

## Reporting provider issues

If you find a bug in a provider, please report the issue directly to the provider author. The OpenTofu team cannot fix provider issues.

-> Provider namespaces and names in the OpenTofu registry translate directly to GitHub URLs in the form of `github.com/NAMESPACE/terraform-provider-NAME`.

## Using HashiCorp-maintained providers (AWS, AzureRM, etc)

The OpenTofu project maintains a mirror of all HashiCorp-created providers under the MPL-2.0 license, such as [AWS](https://github.com/opentofu/terraform-provider-aws), [AzureRM](https://github.com/opentofu/terraform-provider-azurerm), etc. The OpenTofu project builds provider binaries from the source code and publishes them in the OpenTofu Registry under the `hashicorp/` namespace (e.g. `hashicorp/aws`), but doesn't otherwise apply fixes or changes to them.

You can use these providers the same way you are used to from Terraform. If you found a bug in a HashiCorp-maintained provider, please make sure to report it in the original repository under https://github.com/hashicorp as the OpenTofu project cannot fix the issue.

## Asking provider authors to submit their GPG keys

If you are a user of a specific provider, and you would like to ask a provider author to submit their GPG key, you can use the following template to open an issue with the provider:

```markdown
Hello team, thank you for your work on this provider.

I'm using this provider with OpenTofu and I would like to ask you to
submit your GPG public key to the OpenTofu registry in order to enable
the verification of the provider binaries. You can do it simply via a
[GitHub issues here](https://github.com/opentofu/registry/issues/new/choose).
It takes a few minutes and OpenTofu does not require any special permissions
to your GitHub repository.

Thank you for your help!
```
