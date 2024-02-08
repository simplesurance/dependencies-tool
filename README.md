# dependencies-tool

*dependencies-tool* is a command client utility that generates a dependency order
(topological order) from dependency definition files. \
It can be used to generate the order in which application needs to be deployed.


## Dependency Definition Files

Dependency definition files are defined in YAML.
Usually each application has a definition file in its directory.
*dependencies-tool* discovers all definition files that are in a parent
directory or in any subdirectory.

The name of the configuration files or their path suffix (`subdir/deps.yaml`)
can be specified via a command line parameter.

### Format

The YAML format of the configurations files is the following:

```yaml
name: myapp
dependencies:
    prd: &prd
        billing-service: ~
        calc-service: { type: hard }
        letter-service: { type: soft }
    stg: 
      << : *prd
      auth-service: ~
    testing: ~
```

The file defines an app called `myapp`.
`myapp` is part of the distribution `prd`, `stg` and `testing`.
In the `prd` distribution it depends on the `billing-service`, `calc-service`
and `letter-service` applications.
The `billing-service` and `calc-service` are defined as hard dependencies.
If the dependency type is omitted, it defaults to *hard*.
The `letter-service` is a soft dependency, it must be deployed together with
`myapp` but can be deployed before or after `myapp`.
Hard dependencies can not contain loops, soft dependencies can.
For the `prd` also the YAML anchor `&prd` is defined.
The anchor is used to use the same dependencies in the `stg` 
distribution.
The `staging` distribution defines additionally `auth-service` as a hard
dependency.
In the `testing` distribution `myapp` does not have any dependencies.

For all applications that are listed as dependencies, a dependency
definition file must also exist.
The applications must also have distribution entries in their `dependencies`
dictionary, for which they were declared a dependency by `myapp`.

## Examples

1. Give me an dependency-ordered list of applications for the distribution `stg`.
   Discover and parse dependency definitions file in `/repo`:

    ```sh
    dependencies-tool order /repo stg
    ```

2. Export all dependency information found in `/repo` to `/tmp/export.deps`:

    ```sh
    dependencies-tool export /repo /tmp/out.deps
    ```

3. Give me an dependency-ordered list for the application `billing-service` of
   the distribution `prd`, read the dependency information from
   `/tmp/export.deps`, output the list as JSON:

    ```sh
    dependencies-tool order --format json billing-service /tmp/export prd
    ```


3. Generate a [DOT](https://en.wikipedia.org/wiki/DOT_(graph_description_language))
   graph of the dependencies, that can be visualized with
   [Graphviz](https://graphviz.org):

    ```sh
    dependencies-tool order --format dot /repo stg
    ```
