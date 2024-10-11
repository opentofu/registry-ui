# Using a module

Modules provide reusable pieces of code for your OpenTofu project. The OpenTofu Registry contains references to over 20,000 modules on GitHub created by the community. You can [find a module for your use case using the OpenTofu Registry Search](https://search.opentofu.org/modules/). You can learn more about how modules work in OpenTofu from the [OpenTofu documentation](https://opentofu.org/docs/language/modules/). 

~> The OpenTofu Registry does not perform security scanning on modules, and they may contain malicious code. Inspect any module you intend to use and only use modules from authors you trust.

## Integrating a module in your project

Module addresses have three parts: namespaces, names, and target systems. You can include a module in your project by specifying its address and its version:

```hcl2
source "my_name_for_the_module" {
  source  = "NAMESPACE/NAME/TARGETSYSTEM"
  version = "v1.2.3"
  
  # Add parameters for the module here.
}
```

Specifying the version tells OpenTofu to fetch the module from the registry. Once added, you can run `tofu init` to download the module.

For more information about modules, see [the OpenTofu documentation](https://opentofu.org/docs/language/modules/sources/).

## Reporting provider issues

If you find a bug in a module, please report the issue directly to the provider author. The OpenTofu team cannot fix module issues.

-> Module namespaces, names and target systems in the OpenTofu registry translate directly to GitHub URLs in the form of `github.com/NAMESPACE/terraform-TARGETSYSTEM-NAME`.