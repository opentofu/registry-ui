import { describe, expect, test } from "vitest";

import type { Docs } from "./docsSidebar";
import { NestedItem, transformStructure } from "./docsSidebar";

describe("transformStructure", () => {
  test("should handle empty input", () => {
    const data = {} as Docs;

    const result = transformStructure(data, "", "");

    expect(result).toEqual([]);
  });

  test("should handle basic input with just resources", () => {
    const data = {
      resources: [
        {
          name: "resource1",
          title: "Resource 1",
        },
      ],
    } as Docs;

    const result = transformStructure(data, "", "");

    expect(result).toEqual([
      {
        name: "resources",
        label: "Resources",
        open: false,
        items: [
          {
            name: "resource1",
            label: "Resource 1",
            type: "resources",
            active: false,
            open: false,
          },
        ],
      },
    ] as NestedItem[]);
  });

  test("should handle basic input with just resources and active item", () => {
    const data: Docs = {
      resources: [
        {
          name: "resource1",
          title: "Resource 1",
        },
      ],
    } as Docs;

    const result = transformStructure(data, "resources", "resource1");

    expect(result).toEqual([
      {
        name: "resources",
        label: "Resources",
        open: true,
        items: [
          {
            name: "resource1",
            label: "Resource 1",
            type: "resources",
            active: true,
            open: false,
          },
        ],
      },
    ] as NestedItem[]);
  });

  test("with one simple category and an active item", () => {
    const data: Docs = {
      resources: [
        {
          name: "resource1",
          title: "Resource 1",
          subcategory: "Category 1",
        },
      ],
      datasources: [],
      functions: [],
      guides: [],
    };

    const result = transformStructure(data, "resources", "resource1");

    expect(result).toEqual([
      {
        name: "Category 1",
        label: "Category 1",
        open: true,
        items: [
          {
            name: "resources",
            label: "Resources",
            type: "resources",
            open: true,
            items: [
              {
                name: "resource1",
                label: "Resource 1",
                type: "resources",
                active: true,
                open: false,
              },
            ],
          },
        ],
      },
    ] as NestedItem[]);
  });

  test("with one simple category containing 2 items, one non category item and an active item", () => {
    const data: Docs = {
      resources: [
        {
          name: "resource1",
          title: "Resource 1",
          subcategory: "Category 1",
        },
        {
          name: "resource2",
          title: "Resource 2",
        },
      ],
      datasources: [
        {
          name: "datasource1",
          title: "Data source 1",
          subcategory: "Category 1",
        },
      ],
      functions: [],
      guides: [],
    };

    const result = transformStructure(data, "resources", "resource1");

    expect(result).toEqual([
      {
        name: "resources",
        label: "Resources",
        open: false,
        items: [
          {
            name: "resource2",
            label: "Resource 2",
            type: "resources",
            active: false,
            open: false,
          },
        ],
      },
      {
        name: "Category 1",
        label: "Category 1",
        open: true,
        items: [
          {
            name: "datasources",
            label: "Data Sources",
            type: "datasources",
            open: false,
            items: [
              {
                name: "datasource1",
                label: "Data source 1",
                type: "datasources",
                active: false,
                open: false,
              },
            ],
          },
          {
            name: "resources",
            label: "Resources",
            type: "resources",
            open: true,
            items: [
              {
                name: "resource1",
                label: "Resource 1",
                type: "resources",
                active: true,
                open: false,
              },
            ],
          },
        ],
      },
    ] as NestedItem[]);
  });
});
