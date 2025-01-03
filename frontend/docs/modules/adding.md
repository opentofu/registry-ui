# Adding your module to the OpenTofu registry

Once you have [published your module](/docs/modules/publishing), you can add your module to the OpenTofu Registry. You can do this by [creating an issue](https://github.com/opentofu/registry/issues/new/choose) on the OpenTofu Registry GitHub repository.

Here you will have to provide your username/organization name and repository name, which will translate to a module name. For example, consider the following repository:

```
YOURNAME/terraform-NAME-TARGETSYSTEM
```

This will translate to a module address your users can reference as `YOURNAME/NAME/TARGETSYSTEM`.
