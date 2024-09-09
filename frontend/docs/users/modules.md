# Using a module

Modules provide reusable pieces of code for your OpenTofu project. The OpenTofu Registry contains references to over 20.000 modules on GitHub created by the community. You can [find a module for your use case using the OpenTofu Registry Search](https://search.opentofu.org/modules/). You can learn more about how modules work in OpenTofu from the [OpenTofu documentation](https://opentofu.org/docs/language/modules/). 

~> The OpenTofu Registry does not perform security scanning on modules, and they may contain malicious code. Inspect any module you intend to use and only use modules from authors you trust.

## Integrating a module in your project

Module addresses have three parts: namespaces, names, and target systems. For example, the 