Golang GraphQL
==============

The library is still very much a work in progress, but the parser is
fully functional and (to the best of my knowledge) standard-conforming.

Implementation
--------------

Evaluation occurs in four steps:

#### Parsing ####

The plain-text request is parsed into and Abstact Syntax Tree (AST).
Syntax errors will be found at this point, but no type-checking will be
done.

#### Processing ####

Next we traverse the AST and expand fragments, creating a tree of
response nodes, structs which carry information such as the arguments,
directives, and child fields, which will be used by the schema's resolve
callbacks to resolve the request. Each node in the response tree
corresponds a field in the AST and its matching field in the response.

#### Execution ####

Only once the query has been processed do any custom resolvers get
called. The query executor traverses through the response tree, calling
resolve functions as necessary. Unless serial execution is required, the
tree is processed in parallel from the root down to the leaves.

#### Serialization ####

Once all resolvers have run to completion, the response tree is
traversed once more. This time we serialize the data that has been
placed in each response node into json.
