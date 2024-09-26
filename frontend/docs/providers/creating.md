# Creating an OpenTofu provider

OpenTofu, itself a fork of Terraform, uses the [Terraform plugin protocol](https://developer.hashicorp.com/terraform/plugin/terraform-plugin-protocol) in order to communicate with providers. Any provider implementing this protocol will also work with OpenTofu.

## Using the Terraform Plugin Framework (recommended)

As per the recommendation from HashiCorp for Terraform, the easiest option to write a provider is using the Terraform Plugin Framework. You can find detailed guides for this in the [Terraform documentation](https://developer.hashicorp.com/terraform/plugin/framework).

## Manually (only for the adventurous)

~> Support for programming languages other than Go and bypassing the official SDK/framework is very limited. For production use we recommend using Go and the [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework).

OpenTofu supports providers written in any language as long as they can be compiled to a static binary. Providers written in languages that don't compile to a static binary may or may not work, depending on the host operating system.

In order to start the provider OpenTofu uses [go-plugin](https://github.com/hashicorp/go-plugin). This process is as follows:

1. OpenTofu starts the plugin with the `PLUGIN_MIN_PORT`, `PLUGIN_MAX_PORT`, `PLUGIN_PROTOCOL_VERSIONS` and the `PLUGIN_CLIENT_CERT` environment variables set. While OpenTofu currently doesn't use them, your implementation should also implement handling the `PLUGIN_UNIX_SOCKET_DIR`, `PLUGIN_UNIX_SOCKET_GROUP`, and `PLUGIN_MULTIPLEX_GRPC` environment variables.
2. On Windows, the plugin will scan from `PLUGIN_MIN_PORT` and `PLUGIN_MAX_PORT` to find an open port and open a TCP listen socket on that port, binding to `127.0.0.1`.
3. On all other systems, the plugin will create a Unix socket in a directory of its choosing. (See the note about the additional environment variables here.)
4. The plugin chooses one of the `PLUGIN_PROTOCOL_VERSIONS` to support. This can be currently version `5` or `6`.
5. The plugin writes the following line to the stdout, terminated by a newline: `<CoreProtocolVersion>|<ProtocolVersion>|<SocketType>|<SocketAddr>|<Protocol>|<ServerCert>\n`. The `<CoreProtocolVersion>` is always `1` and references the go-plugin protocol version. The `<ProtocolVersion>` references the plugin protocol version must be `5` or `6` for OpenTofu. The `<SocketType>` can only be `unix` or `tcp`, with the corresponding `<SocketAddr>` set. The `<Protocol>` must always be `grpc` for OpenTofu. Finally, the `<ServerCert>` may contain a raw (not PEM-encoded) certificate with base64 encoding to secure the connection between the plugin and OpenTofu in conjunction with `PLUGIN_CLIENT_CERT`.
6. OpenTofu now sends GRPC requests to the plugin. You can find the protocol definitions for these requests [in the OpenTofu repository](https://github.com/opentofu/opentofu/tree/main/docs/plugin-protocol/). We also recommend reading the [Terraform Plugin Protocol documentation](https://developer.hashicorp.com/terraform/plugin/terraform-plugin-protocol).
7. The plugin may write logs to stderr, which OpenTofu records as plugin logs.

## OpenTofu-specific enhancements

### Configured functions

When calling provider-defined functions (introduced in OpenTofu 1.7 and Terraform 1.8), Terraform does not pass any configuration to the provider. This means, your functions cannot make use of provider configuration if you want to support Terraform. OpenTofu configures the provider, so your functions may make use of this configuration.

## Things that don't work in OpenTofu (yet)

### Moving resources between different types ([#1369](https://github.com/opentofu/opentofu/issues/1369))

As of OpenTofu 1.8, OpenTofu does not yet implement using the `moved` block between resources of different types. See issue [#1369](https://github.com/opentofu/opentofu/issues/1369) for details.

## Next steps

Once you have written your provider code, you can proceed to [write your documentation](/docs/providers/docs).