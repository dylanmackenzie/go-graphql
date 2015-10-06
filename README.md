Golang GraphQL
==============

Evaluation occurs in four steps:

#### Parsing ####

The plain-text request is parsed into and Abstact Syntax Tree (AST).
Syntax errors will be found at this point, but no type-checking will be
done.

#### Validation (Optional) ####

If desired, we can validate each query against the schema at this point.

#### Processing ####

Next we traverse the AST and expand fragments, creating a tree of
response nodes, structs which carry information such as the arguments,
directives, and child fields, which will be used by the schema's resolve
callbacks to resolve the request. Each node in the response tree
corresponds a field in the AST and its matching field in the response.
Invalid queries result in memory allocation beyond what is required for
parsing.  However, this optimizes for the case where the query is valid,
as we don't need to traverse the AST twice.

#### Execution ####

Only once the query has been processed do any custom resolvers get
called. The query executor traverses through the response tree, calling
resolve functions as necessary. Unless serial execution is required, the
tree is processed in parallel from the root down to the leaves.

#### Serialization ####

Once all resolvers have run to completion, the response tree is
traversed once more. This time we serialize the data that has been
placed in each response node into json.
