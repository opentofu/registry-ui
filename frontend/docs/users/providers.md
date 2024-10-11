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

variable "github_token {}

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

