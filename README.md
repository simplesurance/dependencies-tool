# Dependencies
## What

dependencies is a small command client tool to get (recursive) dependencies out of our applications.

## How

This tool is currently able to collect dependency informations in 2 ways.

  1. provide sisu directory

  If you provide the **-sisu** parameter, it will read the **.baur.toml** file to get the directories where to find the applications.
  Inside of the application directory it will look for a **.deps.toml** file where the dependencies of tha application is defined.

  2. provide the output of docker-compose as file

  The parameter **-docker-compose** will read the dependency information out of this file.

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

  2. Give me the service dependencies for actionrequest-service usable in docker-compose template


```
 % ./dependencies -sisu ~/sandbox/git/work/sisu -deps actionrequest-service -region eu -environment stg
[ "...-service","consul","...-service","...-service","postgres" ]
```

  3. Generate a visualization:

```
./dependencies -sisu ~/sandbox/git/work/sisu -service certificate-service -region eu -environment stg -format dot > output.dot
dot -Tsvg -o output.svg output.dot
```

  4. basically every service has a dependency to consul and if data is stored, there is a dependency to postgres as well. If you visualize the dependencies for all services, this will make the graphic a bit confusing. Considered that every service with persistant data has it own database you can give **dependencies** a parameter **-owndb**. This will produce a variation of the graph omitting consul and add appropiate **<servicename>-db**.

```
./dependencies -sisu ~/sandbox/git/work/sisu -owndb -region eu -environment stg  -format dot > output.dot
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
