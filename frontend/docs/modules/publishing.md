# Publishing an OpenTofu module

You can host OpenTofu modules in any git repository. However, if you would like to publish the module in the OpenTofu Registry, you will need to host it on [GitHub](https://docs.github.com/en/get-started/start-your-journey).

Once you pushed your code, make sure to [create a tag](https://git-scm.com/book/en/v2/Git-Basics-Tagging) following [semantic versioning](https://semver.org/). This tag will translate to a version in the OpenTofu registry.

Once you have pushed your tag, you can now [add the module to the Registry](/docs/modules/adding).
