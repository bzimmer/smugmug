# smugmug

![build](https://github.com/bzimmer/smugmug/actions/workflows/build.yaml/badge.svg)
[![codecov](https://codecov.io/gh/bzimmer/smugmug/branch/main/graph/badge.svg?token=MP6A7Z0BDQ)](https://codecov.io/gh/bzimmer/smugmug)

A functional approach to the SmugMug API.

## Design Philosophy

The design of this library (`smugmug`) is to separate the interactions with the SmugMug API as much as possible
from the handling of the results. `smugmug` provides both low-level connectivity as well as higher-level
abstractions for building applications. `smugmug` ships with no executable but a feature-rich application,
[ma](https://github.com/bzimmer/ma), has been developed leveraging most of the library's APIs.

The APIs returning scalars and iterators are the most likely candidates for use but the lower-level functionality
required to build the iterators, pages, is available if the need arises.

### Expansions

A note on `expansions`. SmugMug's API provides an optimization to return some attributes of
a request in a single call versus multiple. `smugmug` needs to explicitly support expansions and at this time
not all are handled (though the primary use cases have been implemented). The expansions returning paged results
will not be supported because it complicates the design and the number of results is limited so paging is almost
always required anyway.

### Singles

The methods `User`, `Node`, `Album`, and `Image` all return a single object by the primary key.

### Pages

The plural methods of the above (eg, `Nodes` or `Images`) will return a page of results as well as the metadata
required to page through all possible responses. These methods are provided as a building block for the iterators.

### Iterators

The methods ending in `Iter` (eg, `NodesIter` or `ImagesIter`) use the paging methods to handle the iteration of
all results and accept a typed callback function to provide a flexible mechanism for application-specific logic.

In addition to the `Iter` functions, the `NodeService` also supports iteration of parent and children nodes as
well as providing `Walk` which allows the complete traversal of the node tree (`Folder`s and `Album`s).


