import { describe, expect, test } from "vitest";
import {
  extensionsToStrip,
  stripExtension,
  shortToLongPath,
  reworkRelativePaths,
  prefixAnchors,
} from "./docsProcessor";

describe("stripExtension", () => {
  extensionsToStrip.forEach((extension) => {
    test(`should strip ${extension} extensions`, () => {
      expect(stripExtension(`/docs/overview${extension}`)).toBe(
        "/docs/overview",
      );
    });
    test(`should handle paths with ${extension} extensions and anchor links`, () => {
      expect(stripExtension(`/docs/overview${extension}#section`)).toBe(
        "/docs/overview#section",
      );
    });
  });

  test("should not affect paths without extensions", () => {
    expect(stripExtension("/docs/overview")).toBe("/docs/overview");
  });
});

describe("prefixAnchors", () => {
  test("should prefix anchors with 'user-content-'", () => {
    expect(prefixAnchors("/docs/overview#section")).toBe(
      "/docs/overview#user-content-section",
    );
  });
  test("should not modify paths without anchors", () => {
    expect(prefixAnchors("/docs/overview")).toBe("/docs/overview");
  });
});

describe("shortToLongPath", () => {
  test("should convert /r/ to /resources/", () => {
    expect(shortToLongPath("/docs/r/resource_name")).toBe(
      "/docs/resources/resource_name",
    );
  });

  test("should convert /d/ to /datasources/", () => {
    expect(shortToLongPath("/docs/d/data_source_name")).toBe(
      "/docs/datasources/data_source_name",
    );
  });

  test("should convert /f/ to /functions/", () => {
    expect(shortToLongPath("/docs/f/function_name")).toBe(
      "/docs/functions/function_name",
    );
  });

  test("should not affect paths without short forms", () => {
    expect(shortToLongPath("/docs/overview")).toBe("/docs/overview");
  });
});

describe("reworkRelativePaths", () => {
  test("should not modify provider links with different providers", () => {
    const input = `[GCP resource](/docs/providers/google/r/compute_instance.html)`;
    const output = reworkRelativePaths(input, "hashicorp", "aws", "v1.0.0");
    expect(output).toEqual(input);
  });

  test("should rework links", () => {
    const input = `
    - [\`aws_api_gateway_deployment\` resource](/docs/providers/aws/r/api_gateway_deployment.html)
    - [\`aws_api_gateway_rest_api\` resource](/docs/providers/aws/r/api_gateway_rest_api.html)
    - [\`aws_api_gateway_stage\` resource](/docs/providers/aws/r/api_gateway_stage.html)
    - [\`aws_apigatewayv2_api\` data source](/docs/providers/aws/d/apigatewayv2_api.html)`;

    const output = reworkRelativePaths(input, "hashicorp", "aws", "v1.0.0");

    const expectedLinks = [
      "/provider/hashicorp/aws/v1.0.0/docs/resources/api_gateway_deployment",
      "/provider/hashicorp/aws/v1.0.0/docs/resources/api_gateway_rest_api",
      "/provider/hashicorp/aws/v1.0.0/docs/resources/api_gateway_stage",
      "/provider/hashicorp/aws/v1.0.0/docs/datasources/apigatewayv2_api",
    ];
    expectedLinks.forEach((link) => {
      expect(output).toContain(link);
    });
  });

  test("should not modify non-docs links", () => {
    const input =
      "Check out [our website](https://example.com) for more information.";
    expect(reworkRelativePaths(input, "hashicorp", "aws", "v1.0.0")).toEqual(
      input,
    );
  });

  test("should process multiple links in the same content", () => {
    const input = `
      # Documentation
      
      Check out [this resource](/docs/providers/aws/r/resource_name.md) for resource information.
      
      More details in [this data source](/docs/providers/aws/d/data_source_name.html#usage).
      
      [External link](https://example.com) should remain unchanged.
    `;

    const output = reworkRelativePaths(input, "hashicorp", "aws", "v1.0.0");

    expect(output).toContain(
      "/provider/hashicorp/aws/v1.0.0/docs/resources/resource_name",
    );
    expect(output).toContain(
      "/provider/hashicorp/aws/v1.0.0/docs/datasources/data_source_name#user-content-usage",
    );
    expect(output).toContain("https://example.com");
    expect(output).not.toContain("/docs/providers/aws/r/resource_name.md");
    expect(output).not.toContain("/docs/providers/aws/d/data_source_name.html");
  });

  test("should handle anchor links correctly", () => {
    const input = `[EC2 Instance Scheduling](/docs/providers/aws/r/instance.html#scheduling)`;
    const output = reworkRelativePaths(input, "hashicorp", "aws", "v1.0.0");

    expect(output).toContain(
      "/provider/hashicorp/aws/v1.0.0/docs/resources/instance#user-content-scheduling",
    );
    expect(output).not.toContain(".html");
  });

  test("should ignore provider links when provider doesn't match", () => {
    const input = `
      # Mixed Provider Documentation
      
      [AWS Resource](/docs/providers/aws/r/instance.html)
      [GCP Resource](/docs/providers/google/r/compute_instance.html)
      [Azure Resource](/docs/providers/azurerm/r/virtual_machine.html)
    `;

    // When processing AWS provider docs
    const output = reworkRelativePaths(input, "hashicorp", "aws", "v1.0.0");

    // AWS resources should be transformed
    expect(output).toContain(
      "/provider/hashicorp/aws/v1.0.0/docs/resources/instance",
    );

    // Other provider resources should remain unchanged
    expect(output).toContain("/docs/providers/google/r/compute_instance.html");
    expect(output).toContain("/docs/providers/azurerm/r/virtual_machine.html");
  });

  test("should handle nested directory structures", () => {
    const input = `[Nested Guide](/docs/providers/aws/guides/iam/policies.html)`;
    const output = reworkRelativePaths(input, "hashicorp", "aws", "v1.0.0");

    expect(output).toContain(
      "/provider/hashicorp/aws/v1.0.0/docs/guides/iam/policies",
    );
    expect(output).not.toContain(".html");
  });

  test("should strip extensions from simple filename links", () => {
    const input = `
      # Documentation with filename links
      
      See [README](README.md) for instructions.
      Also check [CHANGELOG](CHANGELOG.html) for recent changes.
      And [LICENSE](LICENSE) doesn't need any changes.
      Note this [anchor test](README.md#installation) should keep the anchor.
    `;

    const output = reworkRelativePaths(input, "hashicorp", "aws", "v1.0.0");

    // Extensions should be stripped but filenames preserved
    expect(output).toContain("[README](README)");
    expect(output).toContain("[CHANGELOG](CHANGELOG)");
    expect(output).toContain("[LICENSE](LICENSE)");
    expect(output).toContain("[anchor test](README#user-content-installation)");

    // Original extensions should be removed
    expect(output).not.toContain("README.md");
    expect(output).not.toContain("CHANGELOG.html");
  });
});
