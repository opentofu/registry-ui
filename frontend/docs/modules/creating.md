# Creating an OpenTofu module

Any OpenTofu project can be a reusable module others can use. By pushing it to GitHub and then publishing the link to your module in the OpenTofu Registry, anyone can easily find and use it.

In order to create a module for reuse, simply write your OpenTofu code the way you normally would and create [variables](https://opentofu.org/docs/language/values/variables/) for configuration.

Your module may also create resources that a user may want to use further. For example, you may create a virtual machine and want to supply your user with the ID of that virtual machine. To return values from your module, you can use [output](https://opentofu.org/docs/language/values/outputs/) blocks.

You can read more about how modules work [in the OpenTofu documentation](https://opentofu.org/docs/language/modules/).

## Readme

Once the code is done, make sure that you add a `README.md` file and explain to your users how to use your module. This file will show up in the Registry Search if the license allows for it.

## License

In order for your module to show up in the OpenTofu Registry Search, your module should be licensed under one of the [supported licenses](https://github.com/opentofu/registry-ui/blob/main/licenses.json). If your module is not under one of these licenses, your module will be findable in the Registry Search, but no other data will be displayed.

## Submodules

If you want to structure your modules further, you can use submodules you can place in subdirectories. To make the submodules show up in the Registry Search, place your submodules in the `modules/MODULENAME` directory in your module.

You can add a `README.md` file to your submodule to provide more information about your submodule.

## Examples

Similar in structure to submodules, examples provide your users with an easy way to get into using your module. To make your example show up in the OpenTofu Registry Search, place the example project into the `examples/EXAMPLENAME` folder and add a `README.md` file.

## Testing your module

Tests are a great way to ensure that your module stays working when community pull requests come in. The [`tofu test` command](https://opentofu.org/docs/cli/commands/test/) has a host of tools to let you write automated tests for your module so you can merge pull requests safely.

## Publishing your module

Once you are happy with your module, [proceed to publishing it](/docs/modules/publishing).
