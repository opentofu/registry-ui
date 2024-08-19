import searchIndex from "../data/provider_search_index.json";

import lunr from "lunr";

import { useState, useMemo } from "react";

function getProvider(namespace: string, name: string) {
  // @ts-expect-error
  var found = searchIndex[namespace].find((item) => item.name === name);
  if (found) {
    return {
      namespace: namespace,
      ...found,
    };
  }
  return null;
}

var start = performance.now();
var index = lunr(function () {
  this.field("namespace");
  this.field("name");
  this.field("identifier");
  this.field("content");

  Object.keys(searchIndex).forEach((namespace) => {
    // @ts-expect-error
    var providers = searchIndex[namespace].map((item) => ({
      namespace: namespace,
      name: item.name,
      content: ["datasources", "resources", "functions"]
        .map((key) => {
          const list = item[key] || [];
          return (
            list
              // @ts-expect-error
              .map((item) => {
                return item.name;
              })
              .join(" ")
          );
        })
        .join(" "),
    }));

    // @ts-expect-error
    providers.forEach((item) => {
      // TODO what is this?
      if (namespace === "opentffoundation") {
        return;
      }
      var indexItem = {
        namespace: namespace,
        name: item.name,
        identifier: `${namespace}/${item.name}`,
        content: item.content,
        id: `${namespace}/${item.name}`,
      };

      var boost = 1;
      if (namespace === "hashicorp") {
        boost = 10;
      }
      this.add(indexItem, {
        boost,
      });
    });
  });
});

var end = performance.now();
console.log(`Indexing took ${end - start}ms`);

export default function Search() {
  var [searchText, setSearchText] = useState("");

  var results = useMemo(() => {
    // TODO: Debounce this
    if (!searchText) {
      return [];
    }
    // time this function and spit it out
    var start = performance.now();
    var results = index.search(searchText);

    var end = performance.now();
    console.log(`Search for ${searchText} took ${end - start}ms`);
    return results;
  }, [searchText]);

  return (
    <div className="flex h-screen flex-col items-center">
      <div>
        <input
          className="rounded-md border px-4 py-2 text-black"
          type="text"
          placeholder="Search..."
          value={searchText}
          onChange={(e) => setSearchText(e.target.value)}
        />
      </div>
      {searchText != "" && (
        <div className="w-3/4">
          {results.map((result) => {
            var provider = getProvider(
              result.ref.split("/")[0],
              result.ref.split("/")[1],
            );
            return (
              <div className="flex w-full" key={result.ref}>
                <div className="my-2 w-full bg-blue-900 p-2 shadow-lg">
                  <div className="flex text-lg font-semibold">
                    <div>{provider.namespace}</div>
                    <div className="mx-2">/</div>
                    <div>{provider.name}</div>
                  </div>
                  <p>Some description of the provider</p>
                  <small className="text-gray-400">
                    Latest Version: {provider.latestVersion}
                  </small>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
