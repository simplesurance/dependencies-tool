# Dependencies
## What

dependencies is a small command client tool to get (recursive) dependencies out of our applications.

## How

This tool is able to collect dependency information in 1 way.

  1. provide sisu directory

  If you provide the **-sisu** parameter, it will read the **.baur.toml** file to get the directories where to find the applications.
  Inside of the application directory it will look for a **.deps.toml** file where the dependencies of tha application is defined.

### deps.toml

Structure:
```
name = "a-service"
talks_to = [ "consul","postgres" ]
```

### What

  1. Give me an ordered list of services I need to deploy in order to do integration tests
     1.1  This list can be a comma separated list of services like this: **...  -service claim-service,pdfrender-service**


```
% ./dependencies -sisu ~/sandbox/git/work/sisu -service actionrequest-service -region eu -environment stg
deployment order for actionrequest-service service(s)
"consul"
"rabbitmq"
...
...
"actionrequest-service"
```


  2. Generate a visualization:

```
./dependencies -sisu ~/sandbox/git/work/sisu -service certificate-service -region eu -environment stg -format dot > output.dot
dot -Tsvg -o output.svg output.dot
```

## Version Management

The sisu-deploy Jenkins Job uses the dependency tool during deployment.
The Jenkins jobs needs to use the dependency-tool version that is compatible
with the current branch.
Incompatibles can be introduced by e.g.:
- changes in the baur configuration file and builds of the dependency-tool that
- use newer baur packages,
- syntax changes in `.deps.toml`,
- changes in the commandline parameter of the depdencies-tool

Compatibility between the sisu-deploy job, the branch that is deployed and the
tool is done by:

- Adding a version number to the artifact filename that is uploaded to S3
  in its `.app.toml` file.
- Ensuring in the `sisu-deploy.Jenkinsfile` that the compatible
  dependencies-tool version is used,
- Executing the `sisu-deploy.Jenkinsfile` from the branch that is deployed in the
  Jenkins job
