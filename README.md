# dependencies-tool

## What

dependencies-tool is a command client tool that generates ordered lists from
dependency trees.

## How

This tool is able to generate dependency trees from the following inputs:

1. A directory tree containing `.deps*.toml` files in child directories.
   It reads the `Discover.application_dirs` list of the
   [`.baur.toml`](https://github.com/simplesurance/baur/) in the root directory.
   In all sub-directories of the `Discover.application_dirs`
   directories, it searches looks for files matching the following names, the
   first found in the following preference order is read:

   - `.deps-<ENV>-<REGION>.toml`,
   - `.deps-<REGION>.toml`,
   - `.deps-<ENVIRONMENT>.toml`,
   - `.deps.toml`

   The values for `<ENV>` and `<REGION`> can be specified as command line
   arguments.

2. A marshalled dependency tree that was created by `dependency-tool export`.

### `deps.toml` File Format

Format:

```toml
name = "a-service"
talks_to = ["consul", "postgres"]
```

## Examples

1. Give me an ordered list of application names to deploy:

    ```sh
    dependencies-tool deploy-order --apps actionrequest-service ~/git/sisu eu staging
    ```

2. Generate a [DOT](https://en.wikipedia.org/wiki/DOT_(graph_description_language))
   graph, that can be visualized with [Graphviz](https://graphviz.org)

    ```sh
    dependencies-tool deploy-order --format dot --apps actionrequest-service ~/git/sisu eu staging
    ```
